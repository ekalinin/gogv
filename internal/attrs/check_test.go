// SPDX-License-Identifier: MIT
// Provenance: original

package attrs

import (
	"strings"
	"testing"

	"github.com/ekalinin/gogv/internal/dotlang"
	"github.com/ekalinin/gogv/internal/graph"
)

func check(t *testing.T, src string) []string {
	t.Helper()
	f, _, err := dotlang.Parse(src)
	if err != nil {
		t.Fatalf("parse %q: %v", src, err)
	}
	var out []string
	for _, w := range Check(dotlang.Build(f.Graphs[0])) {
		out = append(out, w.String())
	}
	return out
}

func TestDocumentedTable(t *testing.T) {
	for _, tt := range []struct {
		kind graph.Kind
		name string
		want bool
	}{
		{graph.NodeKind, "shape", true},
		{graph.EdgeKind, "shape", false},
		{graph.EdgeKind, "arrowhead", true},
		{graph.GraphKind, "rankdir", true},
		{graph.GraphKind, "rank", true}, // used by subgraphs
		{graph.NodeKind, "foo", false},
	} {
		if got := IsDocumented(tt.kind, tt.name); got != tt.want {
			t.Errorf("IsDocumented(%v, %q) = %v", tt.kind, tt.name, got)
		}
	}
}

func TestSupportedValues(t *testing.T) {
	for _, tt := range []struct {
		kind        graph.Kind
		name, value string
		want        bool
	}{
		{graph.NodeKind, "color", "#00FF00", true},
		{graph.NodeKind, "color", "#00ff00", true},
		{graph.NodeKind, "color", "red", false},
		{graph.NodeKind, "color", "#00ff0080", false},
		{graph.NodeKind, "shape", "box", true},
		{graph.NodeKind, "shape", "record", false},
		{graph.GraphKind, "rankdir", "LR", true},
		{graph.GraphKind, "rankdir", "BT", false},
		{graph.GraphKind, "splines", "polyline", true},
		{graph.GraphKind, "splines", "ortho", false},
		{graph.EdgeKind, "arrowhead", "empty", true},
		{graph.EdgeKind, "arrowhead", "vee", false},
		{graph.EdgeKind, "style", "invis", true},
		{graph.EdgeKind, "style", "dashed", false},
		{graph.EdgeKind, "label", "x", true},
		{graph.EdgeKind, "labelangle", "30", false},
	} {
		if got := IsSupported(tt.kind, tt.name, tt.value); got != tt.want {
			t.Errorf("IsSupported(%v, %q, %q) = %v", tt.kind, tt.name, tt.value, got)
		}
	}
}

func TestSupportedAttributesGiveNoWarning(t *testing.T) {
	src := `digraph { rankdir=LR; nodesep=1; ranksep=0.4; splines=polyline; remincross=true; searchsize=500;
		label="title"; fontname="Courier"; fontsize=12; fontcolor="#123456";
		a [shape=diamond, color="#00FF00", width=1, height=2, label="x\ny", fontname="Helvetica", fontsize=10, fontcolor="#000000"];
		r [shape=rect, style=rounded]; s [style=solid];
		a -> b [arrowhead=empty, arrowtail=none, dir=both, arrowsize=2, minlen=2, color="#ff0000", style=invis, constraint=false] }`
	if w := check(t, src); len(w) != 0 {
		t.Errorf("warnings %q", w)
	}
}

func TestClusterAttributesGiveNoWarning(t *testing.T) {
	for _, src := range []string{
		`digraph { subgraph cluster_a { label="A"; labeljust=l; style=solid; color="#FF0000"; a; { rank=source; b } } }`,
		`digraph { rank=same; labeljust=r; color="#00ff00"; subgraph cluster_a { labeljust=c; fontname="Helvetica"; fontsize=10; fontcolor="#0000ff"; a } }`,
		"digraph { { rank=same; a } { rank=min; b } { rank=max; c } { rank=source; d } { rank=sink; e } }",
	} {
		if w := check(t, src); len(w) != 0 {
			t.Errorf("%s: warnings %q", src, w)
		}
	}
}

func TestLabelAttributesGiveNoWarning(t *testing.T) {
	table := `<<TABLE BGCOLOR="#000009" FIXEDSIZE="TRUE" WIDTH="57" HEIGHT="17"><TR><TD></TD></TR></TABLE>>`
	for _, src := range []string{
		`digraph { a [xlabel="X"]; a -> b [label="L", xlabel="X", headlabel="H", taillabel="T", fontname="Helvetica", fontsize=10, fontcolor="#0000FF", sametail=t] }`,
		"digraph { a [label=" + table + ", xlabel=" + table + "]; a -> b [label=" + table + ", xlabel=" + table + ", headlabel=" + table + ", taillabel=" + table + "] }",
		"digraph { subgraph cluster_a { label=" + table + "; a } }",
	} {
		if w := check(t, src); len(w) != 0 {
			t.Errorf("%s: warnings %q", src, w)
		}
	}
}

func TestUnknownAttributeGivesNoWarning(t *testing.T) {
	if w := check(t, "digraph { a [foo=bar]; a -> b [baz=1] }"); len(w) != 0 {
		t.Errorf("warnings %q", w)
	}
}

func TestNotSupportedYetWarnings(t *testing.T) {
	tests := []struct {
		src  string
		want string
	}{
		{"digraph { a [shape=record] }", "node a: shape=record is not supported yet, using ellipse"},
		{"digraph { subgraph cluster_x { style=filled; a } }", "subgraph cluster_x: style=filled is not supported yet, using solid"},
		{"digraph { subgraph cluster_x { label=<<b>x</b>>; a } }", "subgraph cluster_x: HTML labels are not supported yet"},
		{`digraph { subgraph s { pencolor="#ff0000"; subgraph cluster_x { a } } }`, "subgraph cluster_x: pencolor=#ff0000 is not supported yet"},
		{"digraph { { rank=first; a b } }", "subgraph: rank=first is not supported yet"},
		{"digraph { a:p1:n -> b }", "edge a->b: tailport=p1:n is not supported yet, using p1"},
		{"digraph { a:n -> b }", "edge a->b: tailport=n is not supported yet, using center"},
		{"digraph { a -> b:sw }", "edge a->b: headport=sw is not supported yet, using center"},
		{`digraph { a [shape=plaintext, label=<<TABLE><TR><TD PORT="p"></TD></TR></TABLE>>]; a:p -> a }`, "edge a->a: tailport=p is not supported yet, using center"},
		{"digraph { a [label=<<TABLE><TR><TD ROWSPAN=\"2\"></TD></TR><TR><TD></TD></TR></TABLE>>] }", `node a: HTML labels are not supported yet, using \N`},
		{"digraph { a [label=<<b>x</b>>] }", `node a: HTML labels are not supported yet, using \N`},
		{"graph G { a -- b }", "graph G: undirected graphs are not supported yet, using dir=none"},
		{`digraph { a [fontname="Verdana"] }`, "node a: fontname=Verdana is not supported yet, using Times"},
		{"digraph { a [color=red] }", "node a: color=red is not supported yet, using black"},
		{"digraph { splines=ortho; a }", "graph: splines=ortho is not supported yet, using spline"},
		{`digraph { a -> b [labelangle=30, headlabel="H"] }`, "edge a->b: labelangle=30 is not supported yet"},
		{"digraph { a -> b [label=<<TABLE><TR><TD>x</TD></TR></TABLE>>] }", "edge a->b: HTML labels are not supported yet"},
		{`digraph { label=<<TABLE WIDTH="10"><TR><TD></TD></TR></TABLE>>; a }`, "graph: HTML labels are not supported yet"},
		{"digraph { a [style=filled] }", "node a: style=filled is not supported yet"},
		{"digraph { a [shape=diamond, style=rounded] }", "node a: style=rounded with shape=diamond is not supported yet, using solid"},
	}
	for _, tt := range tests {
		got := check(t, tt.src)
		if !strings.Contains(strings.Join(got, "\n"), tt.want) {
			t.Errorf("%s: warnings %q, want %q", tt.src, got, tt.want)
		}
	}
}

func TestWarningsAreNotRepeated(t *testing.T) {
	got := check(t, "digraph { node [shape=record]; a; b }")
	if len(got) != 2 || got[0] != "node a: shape=record is not supported yet, using ellipse" || !strings.HasPrefix(got[1], "node b:") {
		t.Errorf("warnings %q", got)
	}
}

func TestPortsGiveNoWarning(t *testing.T) {
	for _, src := range []string{
		`digraph { a [shape=plaintext, label=<<TABLE><TR><TD PORT="p"></TD></TR></TABLE>>]; a -> b [tailport=p] }`,
		`digraph { a [shape=plaintext, label=<<TABLE><TR><TD PORT="l"></TD><TD PORT="r"></TD></TR></TABLE>>]; a:r -> b }`,
		`digraph { a [shape=plaintext, label=<<TABLE><TR><TD PORT="n"></TD><TD WIDTH="40"></TD></TR></TABLE>>]; a:n -> b }`,
		`digraph { a [shape=plaintext, label=<<TABLE PORT="t"><TR><TD></TD></TR></TABLE>>]; b -> a:t }`,
		`digraph { a [shape=plaintext, label=<<TABLE BORDER="0" CELLBORDER="0" CELLSPACING="0" CELLPADDING="0"><TR><TD></TD><TD FIXEDSIZE="TRUE" WIDTH="1.0" HEIGHT="0.0"></TD><TD></TD></TR><TR><TD FIXEDSIZE="TRUE" WIDTH="0.0" HEIGHT="1.0"></TD><TD BGCOLOR="#00000A" FIXEDSIZE="TRUE" WIDTH="50.484375" HEIGHT="48.0" PORT="h"></TD><TD FIXEDSIZE="TRUE" WIDTH="12.873046875" HEIGHT="1.0"></TD></TR><TR><TD></TD><TD FIXEDSIZE="TRUE" WIDTH="1.0" HEIGHT="0.0"></TD><TD></TD></TR></TABLE>>]; a:h -> b }`,
		// An unknown port is not a feature that is not supported yet.
		`digraph { a [shape=rect]; a:P -> b }`,
	} {
		if w := check(t, src); len(w) != 0 {
			t.Errorf("%s: warnings %q", src, w)
		}
	}
}

func TestUnknownPorts(t *testing.T) {
	build := func(src string) *graph.Graph {
		file, _, err := dotlang.Parse(src)
		if err != nil {
			t.Fatal(err)
		}
		return dotlang.Build(file.Graphs[0])
	}
	table := `[shape=plaintext, label=<<TABLE><TR><TD PORT="p"></TD><TD PORT="n"></TD></TR></TABLE>>]`
	for _, tt := range []struct {
		src  string
		want []string
	}{
		{"digraph { a [shape=rect]; a:P -> b }", []string{"node a, port P unrecognized"}},
		// One warning per edge end, by tail node in node order, the tail
		// end before the head end.
		{"digraph { b:X -> a; a -> b:Y; a:Z -> c; c -> b:W }", []string{
			"node b, port X unrecognized", "node b, port Y unrecognized", "node a, port Z unrecognized", "node b, port W unrecognized"}},
		{"digraph { a:P -> b; a:P -> c; b -> a:Q; a:n -> d; a:h:n -> e }", []string{
			"node a, port P unrecognized", "node a, port P unrecognized", "node a, port h unrecognized", "node a, port Q unrecognized"}},
		{"digraph { a " + table + "; a:p -> b; a:n -> c; a:p:s -> d; b -> a [headport=x] }", []string{"node a, port x unrecognized"}},
		{"digraph { a -> b [label=<<TABLE><TR><TD PORT=\"p\"></TD></TR></TABLE>>, tailport=p] }", []string{"node a, port p unrecognized"}},
	} {
		got := UnknownPorts(build(tt.src))
		if strings.Join(got, "\n") != strings.Join(tt.want, "\n") {
			t.Errorf("%s: warnings %q, want %q", tt.src, got, tt.want)
		}
	}
}
