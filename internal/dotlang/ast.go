// SPDX-License-Identifier: MIT
// Provenance: doc https://graphviz.org/doc/info/lang.html

package dotlang

// File is the syntax tree of a DOT input: its graphs in input order.
type File struct {
	Graphs []*GraphDecl
}

// GraphDecl is a top-level graph.
type GraphDecl struct {
	Strict   bool
	Directed bool
	ID       *IDValue
	Stmts    []Stmt
}

// IDValue is an ID of the language with how it was written.
type IDValue struct {
	Text string
	// HTML is set for an HTML string.
	HTML bool
}

// Stmt is a statement: one of *NodeStmt, *EdgeStmt, *AttrStmt, *AssignStmt
// and *Subgraph.
type Stmt interface{ stmt() }

// Attr is one ID=ID pair of an attribute list.
type Attr struct {
	Key, Value IDValue
}

// NodeID is a node with an optional port: Port holds the port name and the
// compass point, joined by ':' when both are given, as written after the
// node name.
type NodeID struct {
	ID   IDValue
	Port string
}

// NodeStmt is a node statement: one node, or a list of nodes separated by
// commas, with the attributes they all get.
type NodeStmt struct {
	Nodes []NodeID
	Attrs []Attr
}

// EdgeOperand is either a list of nodes separated by commas or a subgraph.
type EdgeOperand struct {
	Nodes    []NodeID
	Subgraph *Subgraph
}

// EdgeStmt is an edge statement: two or more operands joined by edge
// operators.
type EdgeStmt struct {
	Operands []EdgeOperand
	Attrs    []Attr
}

// AttrTarget is the kind of objects an attribute statement applies to.
type AttrTarget int

// The targets of attribute statements.
const (
	TargetGraph AttrTarget = iota
	TargetNode
	TargetEdge
)

// AttrStmt is `graph [...]`, `node [...]` or `edge [...]`.
type AttrStmt struct {
	Target AttrTarget
	Attrs  []Attr
}

// AssignStmt is an `ID = ID` statement: a graph attribute.
type AssignStmt struct {
	Attr Attr
}

// Subgraph is a subgraph, named or anonymous.
type Subgraph struct {
	ID    *IDValue
	Stmts []Stmt
}

func (*NodeStmt) stmt()   {}
func (*EdgeStmt) stmt()   {}
func (*AttrStmt) stmt()   {}
func (*AssignStmt) stmt() {}
func (*Subgraph) stmt()   {}
