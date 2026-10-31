// SPDX-License-Identifier: MIT
// Provenance: observed dot 16.1.0

// Package output writes graphs in the output formats of dot.
package output

import (
	"bytes"
	"cmp"
	"regexp"
	"slices"
	"strings"

	"github.com/ekalinin/gogv/internal/dotlang"
	"github.com/ekalinin/gogv/internal/graph"
)

// Canon writes g back in DOT, without layout, as the reference's -Tcanon
// does.
func Canon(g *graph.Graph) []byte {
	c := &canon{root: g.Root()}
	var b bytes.Buffer
	if g.Strict() {
		b.WriteString("strict ")
	}
	if g.Directed() {
		b.WriteString("digraph ")
	} else {
		b.WriteString("graph ")
	}
	if g.Name != "" {
		b.WriteString(c.id(g.Name, false))
		b.WriteByte(' ')
	}
	b.WriteString("{\n")
	c.body(&b, g, "\t")
	b.WriteString("}\n")
	return b.Bytes()
}

type canon struct {
	root *graph.Graph
}

// written collects what a subgraph wrote, so that its ancestors do not write
// it again.
type written struct {
	nodes map[*graph.Node]bool
	edges map[*graph.Edge]bool
}

func newWritten() written {
	return written{nodes: map[*graph.Node]bool{}, edges: map[*graph.Edge]bool{}}
}

func (w written) merge(o written) {
	for n := range o.nodes {
		w.nodes[n] = true
	}
	for e := range o.edges {
		w.edges[e] = true
	}
}

type attr struct{ name, value string }

func sortAttrs(as []attr) []attr {
	slices.SortStableFunc(as, func(a, b attr) int { return cmp.Compare(a.name, b.name) })
	return as
}

// header returns the graph attributes, node defaults and edge defaults that g
// writes before its subgraphs.
func (c *canon) header(g *graph.Graph) (graphAttrs, nodeDefaults, edgeDefaults []attr) {
	for _, s := range c.root.Symbols(graph.GraphKind) {
		v := g.AttrAt(s)
		if g.IsRoot() && s.SetAtRoot || !g.IsRoot() && (g.AttrSetHere(s) || v != g.Parent.AttrAt(s)) {
			graphAttrs = append(graphAttrs, attr{s.Name, c.value(v)})
		}
	}
	defaults := func(k graph.Kind) []attr {
		var as []attr
		for _, s := range c.root.Symbols(k) {
			if g.IsRoot() {
				if s.SetAtRoot {
					as = append(as, attr{s.Name, c.value(s.Default)})
				}
			} else if v, ok := g.LocalDefault(s, k); ok {
				as = append(as, attr{s.Name, c.value(v)})
			}
		}
		return sortAttrs(as)
	}
	return sortAttrs(graphAttrs), defaults(graph.NodeKind), defaults(graph.EdgeKind)
}

// writeDefaults writes `graph [...]`, `node [...]` or `edge [...]`.
func writeDefaults(b *bytes.Buffer, indent, keyword string, as []attr) {
	if len(as) == 0 {
		return
	}
	b.WriteString(indent + keyword + " [")
	writeList(b, indent, as)
	if len(as) > 1 {
		b.WriteString("\n" + indent)
	}
	b.WriteString("];\n")
}

func writeList(b *bytes.Buffer, indent string, as []attr) {
	for i, a := range as {
		if i > 0 {
			b.WriteString(",\n" + indent + "\t")
		}
		b.WriteString(a.name + "=" + a.value)
	}
}

// body writes the contents of g and returns what it wrote.
func (c *canon) body(b *bytes.Buffer, g *graph.Graph, indent string) written {
	graphAttrs, nodeDefaults, edgeDefaults := c.header(g)
	writeDefaults(b, indent, "graph", graphAttrs)
	writeDefaults(b, indent, "node", nodeDefaults)
	writeDefaults(b, indent, "edge", edgeDefaults)

	desc := newWritten()
	for _, sg := range g.Subgraphs {
		desc.merge(c.subgraph(b, sg, indent))
	}

	// Nodes come in creation order, each followed by its out-edges ordered
	// by head. A node that needs a statement gets it before the first edge
	// that mentions it, so that reading the output back gives it its
	// attributes.
	done := newWritten()
	incident := map[*graph.Node]bool{}
	for _, e := range g.Edges() {
		incident[e.Tail] = true
		incident[e.Head] = true
	}
	writeNode := func(n *graph.Node) {
		if !done.nodes[n] && !desc.nodes[n] && (!incident[n] || c.differsFromRoot(n)) {
			c.node(b, g, n, indent)
			done.nodes[n] = true
		}
	}
	for _, n := range g.Nodes() {
		writeNode(n)
		var out []*graph.Edge
		for _, e := range g.Edges() {
			if e.Tail == n && !desc.edges[e] {
				out = append(out, e)
			}
		}
		slices.SortStableFunc(out, func(a, b *graph.Edge) int { return cmp.Compare(a.Head.Seq, b.Head.Seq) })
		for _, e := range out {
			writeNode(e.Head)
			c.edge(b, g, e, indent)
			done.edges[e] = true
		}
	}
	done.merge(desc)
	return done
}

// subgraph writes a subgraph, unless it is anonymous with nothing of its own
// to write: then its contents belong to its parent.
func (c *canon) subgraph(b *bytes.Buffer, sg *graph.Graph, indent string) written {
	graphAttrs, nodeDefaults, edgeDefaults := c.header(sg)
	if sg.Anonymous && len(graphAttrs)+len(nodeDefaults)+len(edgeDefaults) == 0 {
		w := newWritten()
		for _, child := range sg.Subgraphs {
			w.merge(c.subgraph(b, child, indent))
		}
		return w
	}
	if sg.Anonymous {
		b.WriteString(indent + "{\n")
	} else {
		b.WriteString(indent + "subgraph " + c.id(sg.Name, false) + " {\n")
	}
	w := c.body(b, sg, indent+"\t")
	b.WriteString(indent + "}\n")
	return w
}

// differsFromRoot reports whether a node has a value other than the root's
// default for some attribute: such a node is written even when it has edges.
func (c *canon) differsFromRoot(n *graph.Node) bool {
	for _, s := range c.root.Symbols(graph.NodeKind) {
		if n.AttrAt(s) != s.Default {
			return true
		}
	}
	return false
}

func (c *canon) node(b *bytes.Buffer, g *graph.Graph, n *graph.Node, indent string) {
	var as []attr
	for _, s := range c.root.Symbols(graph.NodeKind) {
		if v := n.AttrAt(s); v != g.Default(s, graph.NodeKind) {
			as = append(as, attr{s.Name, c.value(v)})
		}
	}
	b.WriteString(indent + c.id(n.Name, false))
	c.attrList(b, indent, sortAttrs(as))
}

func (c *canon) edge(b *bytes.Buffer, g *graph.Graph, e *graph.Edge, indent string) {
	var as []attr
	var tailport, headport string
	for _, s := range c.root.Symbols(graph.EdgeKind) {
		v := e.AttrAt(s)
		switch s.Name {
		case "tailport":
			tailport = v.S
			continue
		case "headport":
			headport = v.S
			continue
		}
		if v != g.Default(s, graph.EdgeKind) {
			as = append(as, attr{s.Name, c.value(v)})
		}
	}
	op := " -- "
	if g.Directed() {
		op = " -> "
	}
	b.WriteString(indent + c.id(e.Tail.Name, false) + c.port(tailport) + op + c.id(e.Head.Name, false) + c.port(headport))
	c.attrList(b, indent, sortAttrs(as))
}

func (c *canon) port(p string) string {
	if p == "" {
		return ""
	}
	var b strings.Builder
	for part := range strings.SplitSeq(p, ":") {
		b.WriteByte(':')
		b.WriteString(c.id(part, false))
	}
	return b.String()
}

// attrList ends a node or edge statement, with its attributes when it has
// some.
func (c *canon) attrList(b *bytes.Buffer, indent string, as []attr) {
	if len(as) > 0 {
		b.WriteString("\t[")
		writeList(b, indent, as)
		b.WriteString("]")
	}
	b.WriteString(";\n")
}

func (c *canon) value(v graph.Value) string { return c.id(v.S, v.HTML) }

var (
	identRE   = regexp.MustCompile(`^[A-Za-z_\x80-\xff][A-Za-z0-9_\x80-\xff]*$`)
	numeralRE = regexp.MustCompile(`^-?(\.[0-9]+|[0-9]+(\.[0-9]*)?)$`)
)

// lineBreakAt is how many characters of a quoted string the reference writes
// before it breaks the string, after a space, with a backslash and a newline.
const lineBreakAt = 128

func (c *canon) id(s string, html bool) string { return quoteID(s, html) }

// quoteID writes a string as a DOT ID: as it is when it is an identifier
// that is not a keyword or a numeral, between angle brackets when it is an
// HTML string, and quoted otherwise.
func quoteID(s string, html bool) string {
	switch {
	case html:
		return "<" + s + ">"
	case identRE.MatchString(s) && !dotlang.IsKeyword(s), numeralRE.MatchString(s):
		return s
	}
	var b strings.Builder
	b.WriteByte('"')
	run := 0
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '"' {
			b.WriteByte('\\')
		}
		b.WriteByte(ch)
		run++
		if ch == ' ' && run >= lineBreakAt && i+1 < len(s) {
			b.WriteString("\\\n")
			run = 0
		}
	}
	b.WriteByte('"')
	return b.String()
}
