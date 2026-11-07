// SPDX-License-Identifier: MIT
// Provenance: doc https://graphviz.org/doc/info/lang.html

package main

import (
	"strings"

	"github.com/ekalinin/gogv/internal/dotlang"
)

// printFile writes the graphs of a syntax tree as DOT, one statement per
// line. Every ID is quoted, so that keywords and numbers keep their meaning,
// and HTML strings are written between angle brackets.
func printFile(f *dotlang.File) string {
	var b strings.Builder
	for _, g := range f.Graphs {
		if g.Strict {
			b.WriteString("strict ")
		}
		if g.Directed {
			b.WriteString("digraph ")
		} else {
			b.WriteString("graph ")
		}
		if g.ID != nil {
			b.WriteString(id(*g.ID))
			b.WriteByte(' ')
		}
		b.WriteString("{\n")
		printStmts(&b, g.Stmts, g.Directed, "\t")
		b.WriteString("}\n")
	}
	return b.String()
}

// id writes an ID: an HTML string between angle brackets, any other between
// double quotes, with its double quotes escaped as the lexer reads them.
func id(v dotlang.IDValue) string {
	if v.HTML {
		return "<" + v.Text + ">"
	}
	return `"` + strings.ReplaceAll(v.Text, `"`, `\"`) + `"`
}

func printStmts(b *strings.Builder, stmts []dotlang.Stmt, directed bool, indent string) {
	for _, s := range stmts {
		b.WriteString(indent)
		switch s := s.(type) {
		case *dotlang.NodeStmt:
			b.WriteString(nodeList(s.Nodes))
			b.WriteString(attrList(s.Attrs))
		case *dotlang.EdgeStmt:
			op := " -- "
			if directed {
				op = " -> "
			}
			for i, o := range s.Operands {
				if i > 0 {
					b.WriteString(op)
				}
				if o.Subgraph != nil {
					printSubgraph(b, o.Subgraph, directed, indent)
				} else {
					b.WriteString(nodeList(o.Nodes))
				}
			}
			b.WriteString(attrList(s.Attrs))
		case *dotlang.AttrStmt:
			b.WriteString([...]string{"graph", "node", "edge"}[s.Target])
			b.WriteString(attrList(s.Attrs))
		case *dotlang.AssignStmt:
			b.WriteString(id(s.Attr.Key))
			b.WriteByte('=')
			b.WriteString(id(s.Attr.Value))
		case *dotlang.Subgraph:
			printSubgraph(b, s, directed, indent)
		}
		b.WriteString(";\n")
	}
}

func printSubgraph(b *strings.Builder, s *dotlang.Subgraph, directed bool, indent string) {
	b.WriteString("subgraph ")
	if s.ID != nil {
		b.WriteString(id(*s.ID))
		b.WriteByte(' ')
	}
	b.WriteString("{\n")
	printStmts(b, s.Stmts, directed, indent+"\t")
	b.WriteString(indent)
	b.WriteByte('}')
}

// nodeList writes nodes separated by commas, each with its port and compass
// point, as the parser joins them with ':'.
func nodeList(nodes []dotlang.NodeID) string {
	parts := make([]string, len(nodes))
	for i, n := range nodes {
		parts[i] = id(n.ID)
		if n.Port != "" {
			for p := range strings.SplitSeq(n.Port, ":") {
				parts[i] += ":" + id(dotlang.IDValue{Text: p})
			}
		}
	}
	return strings.Join(parts, ", ")
}

func attrList(attrs []dotlang.Attr) string {
	if len(attrs) == 0 {
		return ""
	}
	parts := make([]string, len(attrs))
	for i, a := range attrs {
		parts[i] = id(a.Key) + "=" + id(a.Value)
	}
	return " [" + strings.Join(parts, ", ") + "]"
}
