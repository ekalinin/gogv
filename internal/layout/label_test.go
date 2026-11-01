// SPDX-License-Identifier: MIT
// Provenance: original

package layout

import "testing"

func edgeOf(t *testing.T, res *Result, tail, head string) EdgeResult {
	t.Helper()
	for _, e := range res.Edges {
		if e.Edge.Tail.Name == tail && e.Edge.Head.Name == head {
			return e
		}
	}
	t.Fatalf("no edge %s -> %s", tail, head)
	return EdgeResult{}
}

const gopumlTable = `<<TABLE BGCOLOR="#000009" FIXEDSIZE="TRUE" WIDTH="57" HEIGHT="17"><TR><TD></TD></TR></TABLE>>`

func TestEscapesInAnEdgeLabel(t *testing.T) {
	res := layoutResolved(t, `digraph G { a -> b [label="\T to \H"]; b -> c [label="\E in \G"] }`)
	for _, tt := range []struct{ tail, head, want string }{{"a", "b", "a to b"}, {"b", "c", "b->c in G"}} {
		e := edgeOf(t, res, tt.tail, tt.head)
		if len(e.Label.Lines) != 1 || e.Label.Lines[0].Text != tt.want || e.LabelText != tt.want {
			t.Errorf("%s -> %s: label %+v, text %q, want %q", tt.tail, tt.head, e.Label.Lines, e.LabelText, tt.want)
		}
	}
}

func TestTableLabelSizes(t *testing.T) {
	res := layoutResolved(t, `digraph { z [shape=rect, width=.01, height=.01, label=<<TABLE BGCOLOR="#000007" FIXEDSIZE="TRUE" WIDTH="18" HEIGHT="11"><TR><TD></TD></TR></TABLE>>] }`)
	if n := nodesOf(res)["z"]; !approx(n.W, 34) || !approx(n.H, 19) {
		t.Errorf("node sized by its table: %v x %v, want 34 x 19", n.W, n.H)
	}
	res = layoutResolved(t, "digraph { a -> b [label="+gopumlTable+"] }")
	if e := edgeOf(t, res, "a", "b"); e.Label.Table == nil || !approx(e.Label.Width, 57) || !approx(e.Label.Height, 17) || !approx(e.LabelPos.X-e.Label.Width/2, nodesOf(res)["a"].X) {
		t.Errorf("edge label table: %+v at %v", e.Label, e.LabelPos)
	}
	res = layoutResolved(t, `digraph { a -> b [label=<<TABLE WIDTH="30" HEIGHT="10"><TR><TD></TD></TR></TABLE>>]; a -> c [label=<<TABLE FIXEDSIZE="TRUE" WIDTH="30" HEIGHT="10"><TR><TD></TD></TR></TABLE>>] }`)
	if e := edgeOf(t, res, "a", "b"); !approx(e.Label.Width, 30) || !approx(e.Label.Height, 12) {
		t.Errorf("table smaller than its cell: %v x %v, want 30 x 12", e.Label.Width, e.Label.Height)
	}
	if e := edgeOf(t, res, "a", "c"); !approx(e.Label.Width, 30) || !approx(e.Label.Height, 10) {
		t.Errorf("fixed table smaller than its cell: %v x %v, want 30 x 10", e.Label.Width, e.Label.Height)
	}
	res = layoutResolved(t, `digraph { subgraph cluster6 { label=<<TABLE BGCOLOR="#000007" FIXEDSIZE="TRUE" WIDTH="51" HEIGHT="11"><TR><TD></TD></TR></TABLE>>; a [shape=rect, label="", width=1, height=0.5] } }`)
	if c := clusterOf(t, res, "cluster6"); !approx(c.X1-c.X0, 88) || !approx(c.Y1-c.Y0, 71) || !approx(c.LabelPos.X, (c.X0+c.X1)/2) {
		t.Errorf("table label of a cluster: box %v x %v, label at %v", c.X1-c.X0, c.Y1-c.Y0, c.LabelPos)
	}
}

// The edge label scenarios of the dot-layout spec.
func TestEdgeLabelScenarios(t *testing.T) {
	res := layoutResolved(t, `digraph { a -> b [label="x"] }`)
	n, e := nodesOf(res), edgeOf(t, res, "a", "b")
	plain := nodesOf(layoutResolved(t, "digraph { a -> b }"))
	if left := e.LabelPos.X - e.Label.Width/2; !approx(left, n["a"].X) || !(e.LabelPos.Y < n["a"].Y && e.LabelPos.Y > n["b"].Y) ||
		!(n["a"].Y-n["b"].Y > plain["a"].Y-plain["b"].Y) {
		t.Errorf("label beside the edge: label at %v, a %v, b %v", e.LabelPos, n["a"].Y, n["b"].Y)
	}

	n = nodesOf(layoutResolved(t, `digraph { a -> b [label="x"]; c -> d }`))
	if !approx(n["a"].Y, n["c"].Y) || !approx(n["b"].Y, n["d"].Y) {
		t.Errorf("ranks of the other edges: a %v c %v, b %v d %v", n["a"].Y, n["c"].Y, n["b"].Y, n["d"].Y)
	}

	e = edgeOf(t, layoutResolved(t, `digraph { a -> b [label="x", fontname="Helvetica", fontsize=20, fontcolor="#FF0000"] }`), "a", "b")
	if e.Label.FontName != "Helvetica" || e.Label.FontSize != 20 || e.Label.FontColor != "#FF0000" {
		t.Errorf("font of an edge label: %+v", e.Label)
	}

	res = layoutResolved(t, `digraph { rankdir=LR; a -> b [label="x"] }`)
	n, e = nodesOf(res), edgeOf(t, res, "a", "b")
	if !(e.LabelPos.Y-e.Label.Height/2 >= n["a"].Y-0.5) || !(e.LabelPos.X > n["a"].X && e.LabelPos.X < n["b"].X) {
		t.Errorf("left to right: label at %v, a %v, b %v", e.LabelPos, n["a"], n["b"])
	}

	res = layoutResolved(t, `digraph { a -> b [label="x"]; a -> b [label="y"] }`)
	x, y := res.Edges[0], res.Edges[1]
	if approx(x.LabelPos.X, y.LabelPos.X) || len(x.Points) == len(y.Points) && approx(x.Points[1].X, y.Points[1].X) {
		t.Errorf("two labeled edges: labels at %v and %v", x.LabelPos, y.LabelPos)
	}
}

func TestFlatAndLoopLabels(t *testing.T) {
	res := layoutResolved(t, `digraph { a -> b [minlen=0, label="flat"] }`)
	n, e := nodesOf(res), edgeOf(t, res, "a", "b")
	plain := nodesOf(layoutResolved(t, `digraph { a -> b [minlen=0] }`))
	if !(e.LabelPos.Y > n["a"].Y) || !(e.LabelPos.X > n["a"].X && e.LabelPos.X < n["b"].X) || !(n["b"].X-n["a"].X > plain["b"].X-plain["a"].X) {
		t.Errorf("flat edge: label at %v, a %v, b %v", e.LabelPos, n["a"], n["b"])
	}
	res = layoutResolved(t, `digraph { a -> a [label="loop"] }`)
	e = edgeOf(t, res, "a", "a")
	var far float64
	for _, p := range e.Points {
		far = max(far, p.X)
	}
	if !approx(e.LabelPos.X-e.Label.Width/2, far) {
		t.Errorf("loop: label at %v, loop reaches %v", e.LabelPos, far)
	}
}

func TestExternalLabelScenarios(t *testing.T) {
	res := layoutResolved(t, `digraph { a [xlabel="X"] }`)
	a := nodesOf(res)["a"]
	if !(a.XLabelPos.X < a.X-a.W/2) || !(a.XLabelPos.Y > a.Y+a.H/2) || !(res.Width > a.W) {
		t.Errorf("node xlabel: at %v, node %v, drawing %v", a.XLabelPos, a, res.Width)
	}

	res = layoutResolved(t, `digraph { a -> b [headlabel="H", taillabel="T"] }`)
	n, e := nodesOf(res), edgeOf(t, res, "a", "b")
	plain := nodesOf(layoutResolved(t, "digraph { a -> b }"))
	if !approx(n["a"].Y, plain["a"].Y) || !approx(n["b"].Y, plain["b"].Y) || !(e.HeadLabelPos.Y < e.TailLabelPos.Y) ||
		!(e.HeadLabelPos.Y-e.HeadLabel.Height/2 >= n["b"].Y+n["b"].H/2-0.5) || !(e.TailLabelPos.Y+e.TailLabel.Height/2 <= n["a"].Y-n["a"].H/2+0.5) {
		t.Errorf("head and tail labels: H at %v, T at %v, a %v, b %v", e.HeadLabelPos, e.TailLabelPos, n["a"], n["b"])
	}

	res = layoutResolved(t, `digraph { a -> b [xlabel="X"] }`)
	n, e = nodesOf(res), edgeOf(t, res, "a", "b")
	if !(e.XLabelPos.Y < n["a"].Y && e.XLabelPos.Y > n["b"].Y) || !approx(n["a"].Y, plain["a"].Y) {
		t.Errorf("edge xlabel: at %v, a %v, b %v", e.XLabelPos, n["a"], n["b"])
	}
}

func TestSameTail(t *testing.T) {
	res := layoutResolved(t, `digraph { a -> b [sametail=t]; a -> c [sametail=t]; a -> d }`)
	b, c, d := edgeOf(t, res, "a", "b").Points[0], edgeOf(t, res, "a", "c").Points[0], edgeOf(t, res, "a", "d").Points[0]
	if !approx(b.X, c.X) || !approx(b.Y, c.Y) || approx(b.X, d.X) && approx(b.Y, d.Y) {
		t.Errorf("shared tail point: %v %v %v", b, c, d)
	}
}
