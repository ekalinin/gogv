// SPDX-License-Identifier: MIT
// Provenance: original

package attrs

import (
	"regexp"
	"slices"
	"strings"

	"github.com/ekalinin/gogv/internal/graph"
	"github.com/ekalinin/gogv/internal/label"
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
var shortValue = regexp.MustCompile(`^[A-Za-z0-9_.#:+,-]{1,40}$`)

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
		if slices.Contains(Styles(style.S), "rounded") && (shape.S == "diamond" || shape.S == "hexagon" || shape.S == "octagon") {
			c.add(Warning{Object: NodeName(n), Feature: "style=rounded with shape=" + shape.S, Default: "solid"})
		}
	}
	for _, e := range root.Edges() {
		for _, s := range root.Symbols(graph.EdgeKind) {
			if s.Name == "tailport" || s.Name == "headport" {
				c.port(root, e, s.Name, e.AttrAt(s))
				continue
			}
			c.value(graph.EdgeKind, EdgeName(root, e), s.Name, e.AttrAt(s))
		}
	}
	return c.warnings
}

// nodeTable returns the table label of a node, or nil when its label is
// not a table label.
func nodeTable(g *graph.Graph, n *graph.Node) *label.Table {
	v, _ := n.Attr(g, "label")
	if !v.HTML {
		return nil
	}
	t, ok := label.ParseTable(v.S)
	if !ok {
		return nil
	}
	return &t
}

// port checks a tailport or a headport: a port on a loop and a compass
// point are not supported yet; the edge attaches to the center of the
// node, or to the port before the compass point.
func (c *checker) port(g *graph.Graph, e *graph.Edge, name string, v graph.Value) {
	if v.S == "" {
		return
	}
	n := e.Tail
	if name == "headport" {
		n = e.Head
	}
	port, compass := label.SplitPort(nodeTable(g, n), v.S)
	switch {
	case e.Tail == e.Head:
		c.add(Warning{Object: EdgeName(g, e), Feature: feature(name, v.S), Default: "center"})
	case compass != "" && port != "":
		c.add(Warning{Object: EdgeName(g, e), Feature: feature(name, v.S), Default: port})
	case compass != "":
		c.add(Warning{Object: EdgeName(g, e), Feature: feature(name, v.S), Default: "center"})
	}
}

// UnknownPorts returns the reference's warnings, without their "Warning: "
// prefix, for the ports that edges name and their nodes do not define: one
// for each edge end, the edges of one tail node after those of the
// previous one, in node order, and the tail end before the head end.
func UnknownPorts(g *graph.Graph) []string {
	root := g.Root()
	tables := map[*graph.Node]*label.Table{}
	tableOf := func(n *graph.Node) *label.Table {
		t, ok := tables[n]
		if !ok {
			t = nodeTable(root, n)
			tables[n] = t
		}
		return t
	}
	byTail := map[*graph.Node][]*graph.Edge{}
	for _, e := range root.Edges() {
		byTail[e.Tail] = append(byTail[e.Tail], e)
	}
	var out []string
	for _, n := range root.Nodes() {
		for _, e := range byTail[n] {
			for _, end := range []struct {
				attr string
				node *graph.Node
			}{{"tailport", e.Tail}, {"headport", e.Head}} {
				v, _ := e.Attr(root, end.attr)
				if v.S == "" {
					continue
				}
				t := tableOf(end.node)
				port, _ := label.SplitPort(t, v.S)
				if port == "" {
					continue
				}
				if t != nil {
					if _, ok := t.PortBox(port); ok {
						continue
					}
				}
				out = append(out, "node "+end.node.Name+", port "+port+" unrecognized")
			}
		}
	}
	return out
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

// isLabel reports whether an attribute holds a label.
func isLabel(name string) bool {
	return name == "label" || name == "xlabel" || name == "headlabel" || name == "taillabel"
}

// tableAllowed reports whether a single-cell table label is supported as
// that attribute of that object: the label and the xlabel of a node, the
// labels of an edge, and the label of a cluster, whose warnings name it as a
// subgraph, but not the label of the root graph.
func tableAllowed(k graph.Kind, name, object string) bool {
	switch k {
	case graph.NodeKind:
		return name == "label" || name == "xlabel"
	case graph.EdgeKind:
		return true
	}
	return name == "label" && strings.HasPrefix(object, "subgraph ")
}

// value checks one attribute value of a graph, a cluster, a node or an edge.
func (c *checker) value(k graph.Kind, object, name string, v graph.Value) {
	if v.S == "" && !v.HTML {
		return
	}
	switch {
	case isLabel(name) && v.HTML:
		if _, ok := label.ParseTable(v.S); ok && tableAllowed(k, name, object) {
			return
		}
		def := ""
		if k == graph.NodeKind && name == "label" {
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
