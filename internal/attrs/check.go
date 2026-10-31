// SPDX-License-Identifier: MIT
// Provenance: original

package attrs

import (
	"regexp"

	"github.com/ekalinin/gogv/internal/graph"
	"github.com/ekalinin/gogv/internal/textmetrics"
)

// Warning reports a Graphviz feature that gogv does not support yet and the
// default it uses instead.
type Warning struct {
	// Object names the graph, subgraph, node or edge, as "node a".
	Object string
	// Feature names what is not supported: "shape=record", "HTML labels".
	Feature string
	// Plural is set when Feature is a plural noun ("clusters are ...").
	Plural bool
	// Default is what gogv uses instead, or "" when there is nothing to name.
	Default string
}

// String returns the warning without its "Warning: " prefix.
func (w Warning) String() string {
	verb := " is"
	if w.Plural {
		verb = " are"
	}
	s := w.Object + ": " + w.Feature + verb + " not supported yet"
	if w.Default != "" {
		s += ", using " + w.Default
	}
	return s
}

// shortValue matches values printed in a warning with their attribute;
// other values (free text) are left out.
var shortValue = regexp.MustCompile(`^[A-Za-z0-9_.#:+-]{1,40}$`)

func feature(name, value string) string {
	if shortValue.MatchString(value) {
		return name + "=" + value
	}
	return name
}

// GraphName names a graph or subgraph in warnings.
func GraphName(g *graph.Graph) string {
	kind := "subgraph"
	if g.IsRoot() {
		kind = "graph"
	}
	if g.Name == "" {
		return kind
	}
	return kind + " " + g.Name
}

// NodeName names a node in warnings.
func NodeName(n *graph.Node) string { return "node " + n.Name }

// EdgeName names an edge in warnings: "edge a->b", or "edge a--b" in an
// undirected graph.
func EdgeName(g *graph.Graph, e *graph.Edge) string {
	op := "--"
	if g.Directed() {
		op = "->"
	}
	return "edge " + e.Tail.Name + op + e.Head.Name
}

// IsCluster reports whether a subgraph is a cluster.
func IsCluster(g *graph.Graph) bool { return g.IsCluster() }

// Check returns the warnings for the features of g that gogv does not
// support yet, in a stable order: the graph, its subgraphs, the nodes, then
// the edges, each with its attributes in declaration order. Attributes that
// the Graphviz documentation does not describe give no warning.
func Check(g *graph.Graph) []Warning {
	c := checker{seen: map[[2]string]bool{}}
	root := g.Root()
	if !root.Directed() {
		c.add(Warning{Object: GraphName(root), Feature: "undirected graphs", Plural: true, Default: "dir=none"})
	}
	for _, s := range root.Symbols(graph.GraphKind) {
		c.value(graph.GraphKind, GraphName(root), s.Name, root.AttrAt(s))
	}
	for _, sg := range root.Graphs()[1:] {
		if IsCluster(sg) {
			// A cluster draws the graph attributes that differ from those of
			// the cluster or graph it is drawn in.
			outer := sg.Parent
			for !outer.IsRoot() && !IsCluster(outer) {
				outer = outer.Parent
			}
			for _, s := range root.Symbols(graph.GraphKind) {
				if v := sg.AttrAt(s); sg.AttrSetHere(s) || v != outer.AttrAt(s) {
					c.value(graph.GraphKind, GraphName(sg), s.Name, v)
				}
			}
		} else if s := root.Symbol(graph.GraphKind, "rank"); s != nil {
			v := sg.AttrAt(s)
			if v.S != "" && v != sg.Parent.AttrAt(s) {
				c.value(graph.GraphKind, GraphName(sg), "rank", v)
			}
		}
	}
	for _, n := range root.Nodes() {
		for _, s := range root.Symbols(graph.NodeKind) {
			c.value(graph.NodeKind, NodeName(n), s.Name, n.AttrAt(s))
		}
		// Rounded corners are drawn on rectangles only.
		style, _ := n.Attr(root, "style")
		shape, _ := n.Attr(root, "shape")
		if style.S == "rounded" && (shape.S == "diamond" || shape.S == "hexagon" || shape.S == "octagon") {
			c.add(Warning{Object: NodeName(n), Feature: "style=rounded with shape=" + shape.S, Default: "solid"})
		}
	}
	for _, e := range root.Edges() {
		for _, s := range root.Symbols(graph.EdgeKind) {
			c.value(graph.EdgeKind, EdgeName(root, e), s.Name, e.AttrAt(s))
		}
	}
	return c.warnings
}

type checker struct {
	warnings []Warning
	seen     map[[2]string]bool
}

func (c *checker) add(w Warning) {
	key := [2]string{w.Object, w.Feature}
	if c.seen[key] {
		return
	}
	c.seen[key] = true
	c.warnings = append(c.warnings, w)
}

// value checks one attribute value of a graph, a cluster, a node or an edge.
func (c *checker) value(k graph.Kind, object, name string, v graph.Value) {
	if v.S == "" && !v.HTML {
		return
	}
	switch {
	case name == "label" && v.HTML && (k == graph.GraphKind || k == graph.NodeKind):
		def := ""
		if k == graph.NodeKind {
			def = graph.LabelDefault
		}
		c.add(Warning{Object: object, Feature: "HTML labels", Plural: true, Default: def})
	case name == "fontname" && IsSupported(k, name, v.S):
		if _, ok := textmetrics.Lookup(v.S); !ok {
			c.add(Warning{Object: object, Feature: feature(name, v.S), Default: "Times"})
		}
	case !IsSupported(k, name, v.S) && IsDocumented(k, name):
		c.add(Warning{Object: object, Feature: feature(name, v.S), Default: defaultOf(k, name)})
	}
}
