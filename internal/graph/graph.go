// SPDX-License-Identifier: MIT
// Provenance: doc https://graphviz.org/doc/info/lang.html
// Provenance: observed dot 16.1.0

// Package graph is the graph model that the DOT reader builds and that the
// layout and the outputs read: a root graph with nested subgraphs, nodes and
// edges, and attributes with the semantics observed on the reference:
//
//   - an attribute is declared once per kind (graph, node, edge), and the
//     declarations keep their order;
//   - when an attribute is declared, the objects of its kind that already
//     exist get the empty value;
//   - a node or an edge takes, when it is created, the defaults of the
//     subgraph it is created in; a subgraph's node and edge defaults are those
//     set in it or else its parent's, looked up when the object is created;
//   - a subgraph copies its parent's graph attributes when it is created;
//   - the node label is declared before anything else, with the default \N.
package graph

import "slices"

// Kind is the kind of object an attribute applies to.
type Kind int

// The kinds of objects.
const (
	GraphKind Kind = iota
	NodeKind
	EdgeKind
)

// Value is an attribute value. HTML marks values written as HTML strings.
type Value struct {
	S    string
	HTML bool
}

// Str returns a plain string value.
func Str(s string) Value { return Value{S: s} }

// Symbol is a declared attribute.
type Symbol struct {
	Name  string
	Index int
	// Default is the root graph's default for nodes and edges; for graph
	// attributes the root graph's value is in its own values.
	Default Value
	// SetAtRoot is set when the root graph sets the attribute: its default
	// for nodes and edges, its value for graphs.
	SetAtRoot bool
}

// root holds what the whole graph tree shares.
type root struct {
	strict, directed bool
	symbols          [3][]*Symbol
	byName           [3]map[string]*Symbol
	nodes            []*Node
	nodeByName       map[string]*Node
	edges            []*Edge
	graphs           []*Graph // the root and every subgraph, in creation order
	anonymous        int
}

// Graph is the root graph or a subgraph.
type Graph struct {
	r      *root
	Parent *Graph
	Name   string
	// Anonymous is set for subgraphs written without a name.
	Anonymous bool
	Subgraphs []*Graph
	values    []Value
	// defaults holds the node and edge defaults set in this subgraph, by
	// symbol index.
	defaults [3]map[int]Value
	// attrSet marks the graph attributes set in this graph, by symbol index.
	attrSet map[int]bool
	nodes   []*Node
	nodeIn  map[*Node]bool
	edges   []*Edge
	edgeIn  map[*Edge]bool
}

// Node is a node of the graph.
type Node struct {
	Name   string
	Seq    int
	values []Value
}

// Edge is an edge of the graph.
type Edge struct {
	Tail, Head *Node
	Seq        int
	values     []Value
}

// LabelDefault is the default label of nodes: their name.
const LabelDefault = `\N`

// New returns a root graph.
func New(name string, directed, strict bool) *Graph {
	r := &root{strict: strict, directed: directed, nodeByName: map[string]*Node{}}
	for k := range r.byName {
		r.byName[k] = map[string]*Symbol{}
	}
	g := &Graph{r: r, Name: name, nodeIn: map[*Node]bool{}, edgeIn: map[*Edge]bool{}}
	r.graphs = append(r.graphs, g)
	g.SetDefault(NodeKind, "label", Str(LabelDefault))
	return g
}

// Root returns the root graph.
func (g *Graph) Root() *Graph { return g.r.graphs[0] }

// Directed reports whether the graph is a digraph.
func (g *Graph) Directed() bool { return g.r.directed }

// Strict reports whether the graph is strict.
func (g *Graph) Strict() bool { return g.r.strict }

// IsRoot reports whether g is the root graph.
func (g *Graph) IsRoot() bool { return g.Parent == nil }

// Symbols returns the declared attributes of a kind, in declaration order.
func (g *Graph) Symbols(k Kind) []*Symbol { return g.r.symbols[k] }

// Symbol returns the declared attribute of a kind with that name, or nil.
func (g *Graph) Symbol(k Kind, name string) *Symbol { return g.r.byName[k][name] }

// Declare declares an attribute of a kind if it is not declared yet. The
// objects of that kind that exist get the empty value.
func (g *Graph) Declare(k Kind, name string) *Symbol {
	r := g.r
	if s := r.byName[k][name]; s != nil {
		return s
	}
	s := &Symbol{Name: name, Index: len(r.symbols[k])}
	r.symbols[k] = append(r.symbols[k], s)
	r.byName[k][name] = s
	switch k {
	case GraphKind:
		for _, sg := range r.graphs {
			sg.values = append(sg.values, Value{})
		}
	case NodeKind:
		for _, n := range r.nodes {
			n.values = append(n.values, Value{})
		}
	case EdgeKind:
		for _, e := range r.edges {
			e.values = append(e.values, Value{})
		}
	}
	return s
}

// SetDefault sets, in g, the default of a node or edge attribute for the
// objects created afterwards.
func (g *Graph) SetDefault(k Kind, name string, v Value) {
	s := g.Declare(k, name)
	if g.IsRoot() {
		s.Default = v
		s.SetAtRoot = true
		return
	}
	if g.defaults[k] == nil {
		g.defaults[k] = map[int]Value{}
	}
	g.defaults[k][s.Index] = v
}

// Default returns the default of a node or edge attribute in g: the one set
// in g, else in its nearest ancestor that sets it, else the root's.
func (g *Graph) Default(s *Symbol, k Kind) Value {
	for sg := g; sg != nil; sg = sg.Parent {
		if v, ok := sg.defaults[k][s.Index]; ok {
			return v
		}
	}
	return s.Default
}

// LocalDefault returns the default of a node or edge attribute set in g
// itself, if any.
func (g *Graph) LocalDefault(s *Symbol, k Kind) (Value, bool) {
	v, ok := g.defaults[k][s.Index]
	return v, ok
}

// SetAttr sets a graph attribute of g.
func (g *Graph) SetAttr(name string, v Value) {
	s := g.Declare(GraphKind, name)
	g.values[s.Index] = v
	if g.IsRoot() {
		s.SetAtRoot = true
	}
	if g.attrSet == nil {
		g.attrSet = map[int]bool{}
	}
	g.attrSet[s.Index] = true
}

// AttrSetHere reports whether g itself sets the graph attribute.
func (g *Graph) AttrSetHere(s *Symbol) bool { return g.attrSet[s.Index] }

// Attr returns a graph attribute of g, and whether it is declared.
func (g *Graph) Attr(name string) (Value, bool) {
	s := g.r.byName[GraphKind][name]
	if s == nil {
		return Value{}, false
	}
	return g.values[s.Index], true
}

// AttrAt returns the graph attribute of g for a declared symbol.
func (g *Graph) AttrAt(s *Symbol) Value { return g.values[s.Index] }

// Subgraph returns the subgraph of g with that name, creating it when it
// does not exist. An empty name creates a new anonymous subgraph.
func (g *Graph) Subgraph(name string) *Graph {
	if name != "" {
		for _, sg := range g.Subgraphs {
			if sg.Name == name {
				return sg
			}
		}
	}
	sg := &Graph{
		r:      g.r,
		Parent: g,
		Name:   name,
		values: slices.Clone(g.values),
		nodeIn: map[*Node]bool{},
		edgeIn: map[*Edge]bool{},
	}
	if name == "" {
		g.r.anonymous++
		sg.Anonymous = true
	}
	g.Subgraphs = append(g.Subgraphs, sg)
	g.r.graphs = append(g.r.graphs, sg)
	return sg
}

// Graphs returns the root and every subgraph, in creation order.
func (g *Graph) Graphs() []*Graph { return g.r.graphs }

// Node returns the node with that name, creating it in g when it does not
// exist, and makes it a member of g and of its ancestors.
func (g *Graph) Node(name string) (n *Node, created bool) {
	n = g.r.nodeByName[name]
	if n == nil {
		created = true
		n = &Node{Name: name, Seq: len(g.r.nodes)}
		for _, s := range g.r.symbols[NodeKind] {
			n.values = append(n.values, g.Default(s, NodeKind))
		}
		g.r.nodes = append(g.r.nodes, n)
		g.r.nodeByName[name] = n
	}
	for sg := g; sg != nil; sg = sg.Parent {
		sg.addNode(n)
	}
	return n, created
}

// FindNode returns the node with that name, or nil.
func (g *Graph) FindNode(name string) *Node { return g.r.nodeByName[name] }

func (g *Graph) addNode(n *Node) {
	if g.nodeIn[n] {
		return
	}
	g.nodeIn[n] = true
	i, _ := slices.BinarySearchFunc(g.nodes, n.Seq, func(m *Node, seq int) int { return m.Seq - seq })
	g.nodes = slices.Insert(g.nodes, i, n)
}

// RemoveNode removes n from g and from its subgraphs, with the edges of g
// that are incident to it.
func (g *Graph) RemoveNode(n *Node) {
	if !g.nodeIn[n] {
		return
	}
	for _, sg := range g.Subgraphs {
		sg.RemoveNode(n)
	}
	delete(g.nodeIn, n)
	g.nodes = slices.DeleteFunc(g.nodes, func(m *Node) bool { return m == n })
	g.edges = slices.DeleteFunc(g.edges, func(e *Edge) bool {
		if e.Tail == n || e.Head == n {
			delete(g.edgeIn, e)
			return true
		}
		return false
	})
}

// Nodes returns the nodes of g, in creation order.
func (g *Graph) Nodes() []*Node { return g.nodes }

// HasNode reports whether n is a member of g.
func (g *Graph) HasNode(n *Node) bool { return g.nodeIn[n] }

// Edge creates an edge from tail to head in g, or, in a strict graph,
// returns the existing edge between them, and makes it a member of g and of
// its ancestors.
func (g *Graph) Edge(tail, head *Node) (e *Edge, created bool) {
	if g.r.strict {
		e = g.r.findEdge(tail, head)
	}
	if e == nil {
		created = true
		e = &Edge{Tail: tail, Head: head, Seq: len(g.r.edges)}
		for _, s := range g.r.symbols[EdgeKind] {
			e.values = append(e.values, g.Default(s, EdgeKind))
		}
		g.r.edges = append(g.r.edges, e)
	}
	for sg := g; sg != nil; sg = sg.Parent {
		sg.addEdge(e)
	}
	return e, created
}

func (r *root) findEdge(tail, head *Node) *Edge {
	for _, e := range r.edges {
		if e.Tail == tail && e.Head == head || !r.directed && e.Tail == head && e.Head == tail {
			return e
		}
	}
	return nil
}

func (g *Graph) addEdge(e *Edge) {
	if g.edgeIn[e] {
		return
	}
	g.edgeIn[e] = true
	i, _ := slices.BinarySearchFunc(g.edges, e.Seq, func(f *Edge, seq int) int { return f.Seq - seq })
	g.edges = slices.Insert(g.edges, i, e)
}

// Edges returns the edges of g, in creation order.
func (g *Graph) Edges() []*Edge { return g.edges }

// HasEdge reports whether e is a member of g.
func (g *Graph) HasEdge(e *Edge) bool { return g.edgeIn[e] }

// SetNodeAttr sets an attribute of a node.
func (g *Graph) SetNodeAttr(n *Node, name string, v Value) {
	s := g.Declare(NodeKind, name)
	n.values[s.Index] = v
}

// SetEdgeAttr sets an attribute of an edge.
func (g *Graph) SetEdgeAttr(e *Edge, name string, v Value) {
	s := g.Declare(EdgeKind, name)
	e.values[s.Index] = v
}

// Attr returns an attribute of the node, and whether it is declared.
func (n *Node) Attr(g *Graph, name string) (Value, bool) {
	s := g.r.byName[NodeKind][name]
	if s == nil {
		return Value{}, false
	}
	return n.values[s.Index], true
}

// AttrAt returns the node's value of a declared symbol.
func (n *Node) AttrAt(s *Symbol) Value { return n.values[s.Index] }

// Attr returns an attribute of the edge, and whether it is declared.
func (e *Edge) Attr(g *Graph, name string) (Value, bool) {
	s := g.r.byName[EdgeKind][name]
	if s == nil {
		return Value{}, false
	}
	return e.values[s.Index], true
}

// AttrAt returns the edge's value of a declared symbol.
func (e *Edge) AttrAt(s *Symbol) Value { return e.values[s.Index] }
