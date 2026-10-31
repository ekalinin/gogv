// SPDX-License-Identifier: MIT
// Provenance: doc https://graphviz.org/doc/info/lang.html
// Provenance: observed dot 16.1.0

package dotlang

import "github.com/ekalinin/gogv/internal/graph"

// Build turns a parsed graph into the graph model, executing its statements
// in order.
func Build(decl *GraphDecl) *graph.Graph {
	name := ""
	if decl.ID != nil {
		name = decl.ID.Text
	}
	g := graph.New(name, decl.Directed, decl.Strict)
	b := builder{}
	b.stmts(g, decl.Stmts)
	return g
}

type builder struct{}

func value(id IDValue) graph.Value { return graph.Value{S: id.Text, HTML: id.HTML} }

func (b builder) stmts(g *graph.Graph, stmts []Stmt) {
	for _, s := range stmts {
		switch s := s.(type) {
		case *AttrStmt:
			for _, a := range s.Attrs {
				switch s.Target {
				case TargetGraph:
					g.SetAttr(a.Key.Text, value(a.Value))
				case TargetNode:
					g.SetDefault(graph.NodeKind, a.Key.Text, value(a.Value))
				case TargetEdge:
					g.SetDefault(graph.EdgeKind, a.Key.Text, value(a.Value))
				}
			}
		case *AssignStmt:
			g.SetAttr(s.Attr.Key.Text, value(s.Attr.Value))
		case *NodeStmt:
			for _, id := range s.Nodes {
				n, _ := g.Node(id.ID.Text)
				for _, a := range s.Attrs {
					g.SetNodeAttr(n, a.Key.Text, value(a.Value))
				}
			}
		case *Subgraph:
			b.subgraph(g, s)
		case *EdgeStmt:
			b.edges(g, s)
		}
	}
}

func (b builder) subgraph(g *graph.Graph, s *Subgraph) *graph.Graph {
	name := ""
	if s.ID != nil {
		name = s.ID.Text
	}
	sg := g.Subgraph(name)
	b.stmts(sg, s.Stmts)
	return sg
}

// endpoint is a node of an edge operand, with the port written after it.
type endpoint struct {
	node *graph.Node
	port string
}

// edges creates the edges of an edge statement: every node of an operand to
// every node of the next one, in order. Operands are evaluated first, from
// left to right.
func (b builder) edges(g *graph.Graph, s *EdgeStmt) {
	operands := make([][]endpoint, len(s.Operands))
	for i, op := range s.Operands {
		if op.Subgraph == nil {
			for _, id := range op.Nodes {
				n, _ := g.Node(id.ID.Text)
				operands[i] = append(operands[i], endpoint{node: n, port: id.Port})
			}
			continue
		}
		sg := b.subgraph(g, op.Subgraph)
		for _, n := range sg.Nodes() {
			operands[i] = append(operands[i], endpoint{node: n})
		}
	}
	for i := 0; i+1 < len(operands); i++ {
		for _, t := range operands[i] {
			for _, h := range operands[i+1] {
				e, _ := g.Edge(t.node, h.node)
				if t.port != "" {
					g.SetEdgeAttr(e, "tailport", graph.Str(t.port))
				}
				if h.port != "" {
					g.SetEdgeAttr(e, "headport", graph.Str(h.port))
				}
				for _, a := range s.Attrs {
					g.SetEdgeAttr(e, a.Key.Text, value(a.Value))
				}
			}
		}
	}
}
