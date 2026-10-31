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
		{graph.EdgeKind, "label", "x", false},
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
		{"digraph { subgraph cluster_x { a b } }", "subgraph cluster_x: clusters are not supported yet"},
		{"digraph { { rank=same; a b } }", "subgraph: rank=same is not supported yet"},
		{"digraph { a:p1:n -> b }", "edge a->b: tailport=p1:n is not supported yet"},
		{"digraph { a [label=<<b>x</b>>] }", `node a: HTML labels are not supported yet, using \N`},
		{"graph G { a -- b }", "graph G: undirected graphs are not supported yet, using dir=none"},
		{`digraph { a [fontname="Verdana"] }`, "node a: fontname=Verdana is not supported yet, using Times"},
		{"digraph { a [color=red] }", "node a: color=red is not supported yet, using black"},
		{"digraph { splines=ortho; a }", "graph: splines=ortho is not supported yet, using spline"},
		{`digraph { a -> b [label="some text"] }`, "edge a->b: label is not supported yet"},
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
