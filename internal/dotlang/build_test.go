// SPDX-License-Identifier: MIT
// Provenance: doc https://graphviz.org/doc/info/lang.html
// Provenance: observed dot 16.1.0

package dotlang

import (
	"strings"
	"testing"

	"github.com/ekalinin/gogv/internal/graph"
)

func build(t *testing.T, src string) *graph.Graph {
	t.Helper()
	f, _, err := Parse(src)
	if err != nil {
		t.Fatalf("parse %q: %v", src, err)
	}
	if len(f.Graphs) != 1 {
		t.Fatalf("parse %q: %d graphs", src, len(f.Graphs))
	}
	return Build(f.Graphs[0])
}

func edgeList(g *graph.Graph) string {
	var parts []string
	for _, e := range g.Edges() {
		parts = append(parts, e.Tail.Name+"->"+e.Head.Name)
	}
	return strings.Join(parts, " ")
}

func nodeAttr(t *testing.T, g *graph.Graph, node, name string) graph.Value {
	t.Helper()
	n := g.FindNode(node)
	if n == nil {
		t.Fatalf("no node %q", node)
	}
	v, _ := n.Attr(g, name)
	return v
}

func TestEdgeChainWithSubgraphOperand(t *testing.T) {
	g := build(t, "digraph { a -> { b c } -> d }")
	if got := edgeList(g); got != "a->b a->c b->d c->d" {
		t.Errorf("edges %q", got)
	}
}

func TestCrossProductOfSubgraphs(t *testing.T) {
	g := build(t, "digraph { {a b} -> {c d} }")
	if got := edgeList(g); got != "a->c a->d b->c b->d" {
		t.Errorf("edges %q", got)
	}
}

func TestConcatenatedString(t *testing.T) {
	g := build(t, `digraph { a [label="Hel" + "lo"] }`)
	if v := nodeAttr(t, g, "a", "label"); v.S != "Hello" || v.HTML {
		t.Errorf("label %+v", v)
	}
}

func TestHTMLString(t *testing.T) {
	g := build(t, `digraph { a [label=<<TABLE><TR><TD>x</TD></TR></TABLE>>] }`)
	if v := nodeAttr(t, g, "a", "label"); !v.HTML || v.S != "<TABLE><TR><TD>x</TD></TR></TABLE>" {
		t.Errorf("label %+v", v)
	}
}

func TestTwoGraphsOnTwoLines(t *testing.T) {
	f, _, err := Parse("digraph A { a }\ndigraph B { b }\n")
	if err != nil || len(f.Graphs) != 2 || f.Graphs[0].ID.Text != "A" || f.Graphs[1].ID.Text != "B" {
		t.Fatalf("graphs %+v, err %v", f.Graphs, err)
	}
	// As in the reference, the rest of the line after a graph is ignored.
	f, _, err = Parse("digraph A { a } digraph B { b }\n")
	if err != nil || len(f.Graphs) != 1 {
		t.Fatalf("one line: %d graphs, err %v", len(f.Graphs), err)
	}
	// Line numbers start again after a graph.
	_, _, err = Parse("digraph A { a }\n\ndigraph B {\nb -> }\n")
	if err == nil || err.Error() != "syntax error in line 3 near '}'" {
		t.Fatalf("error in the second graph: %v", err)
	}
}

func TestDefaultsInScope(t *testing.T) {
	g := build(t, "digraph { node [shape=rect]; a; subgraph s { node [shape=circle]; b } c }")
	for node, want := range map[string]string{"a": "rect", "b": "circle", "c": "rect"} {
		if v := nodeAttr(t, g, node, "shape"); v.S != want {
			t.Errorf("%s: shape %q, want %q", node, v.S, want)
		}
	}
}

func TestDefaultsDoNotChangeExistingObjects(t *testing.T) {
	g := build(t, "digraph { a; node [shape=box]; b; a -> b; edge [color=red]; b -> c }")
	if v := nodeAttr(t, g, "a", "shape"); v.S != "" {
		t.Errorf("a: shape %q, want empty", v.S)
	}
	if v := nodeAttr(t, g, "b", "shape"); v.S != "box" {
		t.Errorf("b: shape %q, want box", v.S)
	}
	colors := []string{}
	for _, e := range g.Edges() {
		v, _ := e.Attr(g, "color")
		colors = append(colors, v.S)
	}
	if strings.Join(colors, ",") != ",red" {
		t.Errorf("edge colors %q", colors)
	}
}

func TestSubgraphCopiesGraphAttributes(t *testing.T) {
	g := build(t, "digraph { subgraph s { x } label=L; subgraph t { y } }")
	s, tt := g.Subgraphs[0], g.Subgraphs[1]
	if v, _ := s.Attr("label"); v.S != "" {
		t.Errorf("s: label %q, want empty", v.S)
	}
	if v, _ := tt.Attr("label"); v.S != "L" {
		t.Errorf("t: label %q, want L", v.S)
	}
}

func TestStrictGraphMergesEdges(t *testing.T) {
	g := build(t, `strict digraph { a -> b [color="#ff0000"]; a -> b [minlen=2] }`)
	if len(g.Edges()) != 1 {
		t.Fatalf("edges %q", edgeList(g))
	}
	e := g.Edges()[0]
	color, _ := e.Attr(g, "color")
	minlen, _ := e.Attr(g, "minlen")
	if color.S != "#ff0000" || minlen.S != "2" {
		t.Errorf("color %q, minlen %q", color.S, minlen.S)
	}
	g = build(t, "strict graph { a -- b; b -- a; a -- a; a -- a }")
	if got := edgeList(g); got != "a->b a->a" {
		t.Errorf("strict undirected edges %q", got)
	}
}

func TestMembership(t *testing.T) {
	g := build(t, "digraph { a; subgraph s { a; b } subgraph s { c } }")
	if len(g.Subgraphs) != 1 {
		t.Fatalf("%d subgraphs, want s reopened", len(g.Subgraphs))
	}
	var names []string
	for _, n := range g.Subgraphs[0].Nodes() {
		names = append(names, n.Name)
	}
	if strings.Join(names, " ") != "a b c" {
		t.Errorf("members of s: %q", names)
	}
}

func TestPorts(t *testing.T) {
	g := build(t, "digraph { a:p1:n -> b:s; e:n -> f [tailport=s] }")
	tp, _ := g.Edges()[0].Attr(g, "tailport")
	hp, _ := g.Edges()[0].Attr(g, "headport")
	if tp.S != "p1:n" || hp.S != "s" {
		t.Errorf("ports %q %q", tp.S, hp.S)
	}
	tp, _ = g.Edges()[1].Attr(g, "tailport")
	if tp.S != "s" {
		t.Errorf("tailport %q, want the attribute's s", tp.S)
	}
}

func TestLabelDeclaredFirst(t *testing.T) {
	g := build(t, `digraph { a [label="x"]; b }`)
	if v := nodeAttr(t, g, "b", "label"); v.S != `\N` {
		t.Errorf("b: label %q, want \\N", v.S)
	}
	if syms := g.Symbols(graph.NodeKind); syms[0].Name != "label" {
		t.Errorf("first node symbol %q", syms[0].Name)
	}
}
