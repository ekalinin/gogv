// SPDX-License-Identifier: MIT
// Provenance: original

package layout

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"strings"
	"sync"
	"testing"

	"github.com/ekalinin/gogv/internal/dotlang"
	"github.com/ekalinin/gogv/internal/graph"
)

func buildGraph(t *testing.T, src string) *graph.Graph {
	t.Helper()
	f, _, err := dotlang.Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	return dotlang.Build(f.Graphs[0])
}

// bigGraph returns a random graph with n nodes, with cycles.
func bigGraph(n int) string {
	r := rand.New(rand.NewPCG(5, 6))
	var b strings.Builder
	b.WriteString("digraph {\n")
	for v := 1; v < n; v++ {
		fmt.Fprintf(&b, "n%d -> n%d;\n", r.IntN(v), v)
		fmt.Fprintf(&b, "n%d -> n%d;\n", r.IntN(n), r.IntN(n))
	}
	b.WriteString("}\n")
	return b.String()
}

func signature(res *Result) string {
	var b strings.Builder
	for _, n := range res.Nodes {
		fmt.Fprintf(&b, "%s %v %v;", n.Node.Name, n.X, n.Y)
	}
	return b.String()
}

// TestParallelLayouts lays out the same graphs from many goroutines; run
// with -race, it also checks that the layout keeps no shared mutable state.
func TestParallelLayouts(t *testing.T) {
	src := bigGraph(60)
	want := signature(must(Layout(context.Background(), buildGraph(t, src))))
	var wg sync.WaitGroup
	errs := make(chan string, 16)
	for range 16 {
		wg.Go(func() {
			res, err := Layout(context.Background(), buildGraph(t, src))
			if err != nil {
				errs <- err.Error()
				return
			}
			if got := signature(res); got != want {
				errs <- "different layout"
			}
		})
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Error(e)
	}
}

func must(res *Result, err error) *Result {
	if err != nil {
		panic(err)
	}
	return res
}

// TestCanceledLayout checks that a canceled context stops a large layout.
func TestCanceledLayout(t *testing.T) {
	g := buildGraph(t, bigGraph(400))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Layout(ctx, g); !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func layoutOf(t *testing.T, src string) map[string]NodeResult {
	t.Helper()
	res, err := Layout(context.Background(), buildGraph(t, src))
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]NodeResult{}
	for _, n := range res.Nodes {
		out[n.Node.Name] = n
	}
	return out
}

func approx(a, b float64) bool { return a-b < 0.01 && b-a < 0.01 }

// The ranking scenarios of the dot-layout spec.
func TestRankingScenarios(t *testing.T) {
	n := layoutOf(t, "digraph { a -> b -> c; a -> c }")
	if !(n["a"].Y > n["b"].Y && n["b"].Y > n["c"].Y) || !approx(n["a"].Y-n["b"].Y, n["b"].Y-n["c"].Y) {
		t.Errorf("shortest ranking: y %v %v %v", n["a"].Y, n["b"].Y, n["c"].Y)
	}
	n = layoutOf(t, "digraph { a -> b [minlen=2]; c -> d -> e }")
	if !approx(n["a"].Y, n["c"].Y) || !approx(n["b"].Y, n["e"].Y) {
		t.Errorf("minlen=2: a %v, b %v; ranks %v %v %v", n["a"].Y, n["b"].Y, n["c"].Y, n["d"].Y, n["e"].Y)
	}
	for _, src := range []string{"digraph { a -> b [minlen=0] }", "digraph { a -> b [constraint=false] }"} {
		n = layoutOf(t, src)
		if !approx(n["a"].Y, n["b"].Y) {
			t.Errorf("%s: a and b on different ranks", src)
		}
	}
	n = layoutOf(t, "digraph { a -> b -> a }")
	if !(n["a"].Y > n["b"].Y) {
		t.Errorf("cycle: a %v not above b %v", n["a"].Y, n["b"].Y)
	}
}

func TestRankDirectionAndOrder(t *testing.T) {
	n := layoutOf(t, "digraph { rankdir=LR; a -> b }")
	if !(n["b"].X > n["a"].X) || !approx(n["a"].Y, n["b"].Y) {
		t.Errorf("rankdir=LR: a %v,%v b %v,%v", n["a"].X, n["a"].Y, n["b"].X, n["b"].Y)
	}
	n = layoutOf(t, "digraph { a; b; c }")
	if !(n["a"].X < n["b"].X && n["b"].X < n["c"].X) {
		t.Errorf("input order: %v %v %v", n["a"].X, n["b"].X, n["c"].X)
	}
	n = layoutOf(t, "digraph { a; b; c; d; a -> d; b -> c }")
	if (n["a"].X < n["b"].X) != (n["d"].X < n["c"].X) {
		t.Errorf("crossing not removed: a %v b %v c %v d %v", n["a"].X, n["b"].X, n["c"].X, n["d"].X)
	}
	n = layoutOf(t, `digraph { a -> b [label="LLLLLL"]; a -> c; d -> b }`)
	if !(n["a"].X < n["d"].X && n["c"].X < n["b"].X) {
		t.Errorf("orders with the same crossings: a %v d %v, c %v b %v", n["a"].X, n["d"].X, n["c"].X, n["b"].X)
	}
	// The sources in the order of a search from f: b before c.
	n = layoutOf(t, "digraph { f -> g; c -> f; b -> g }")
	if !(n["b"].X < n["f"].X) {
		t.Errorf("sources in search order: b %v f %v", n["b"].X, n["f"].X)
	}
}

func TestSeparationAndPositioning(t *testing.T) {
	n := layoutOf(t, "digraph { nodesep=1; a; b }")
	if gap := n["b"].X - n["b"].W/2 - (n["a"].X + n["a"].W/2); gap < 72-0.5 {
		t.Errorf("nodesep=1: gap %v", gap)
	}
	n = layoutOf(t, "digraph { ranksep=1; a -> b }")
	if gap := n["a"].Y - n["a"].H/2 - (n["b"].Y + n["b"].H/2); !approx(gap, 72) {
		t.Errorf("ranksep=1: gap %v", gap)
	}
	n = layoutOf(t, "digraph { a -> b -> c }")
	if !approx(n["a"].X, n["b"].X) || !approx(n["b"].X, n["c"].X) {
		t.Errorf("straight chain: %v %v %v", n["a"].X, n["b"].X, n["c"].X)
	}
	n = layoutOf(t, "digraph { a -> x; c -> y; b -> x; d -> y }")
	if !approx(n["x"].X, (n["a"].X+n["b"].X)/2) || !approx(n["y"].X, (n["c"].X+n["d"].X)/2) {
		t.Errorf("heads between their tails: x %v of %v %v, y %v of %v %v",
			n["x"].X, n["a"].X, n["b"].X, n["y"].X, n["c"].X, n["d"].X)
	}
}

// Every rank between the ends of a long edge holds one virtual node of it.
func TestVirtualNodes(t *testing.T) {
	ranks := []int{0, 1, 2, 3}
	g, chains := buildRanked(ranks, [][2]int{{0, 3}, {0, 1}, {3, 0}}, nil, nil)
	if len(chains[0]) != 4 || len(chains[1]) != 2 {
		t.Fatalf("chains %v", chains)
	}
	for i, v := range chains[0] {
		if g.nodes[v].rank != i {
			t.Errorf("node %d of the chain on rank %d", i, g.nodes[v].rank)
		}
		if i > 0 && i < 3 && g.nodes[v].real >= 0 {
			t.Errorf("inner node %d is not virtual", i)
		}
	}
	// A reversed edge between the same nodes shares the chain.
	if &chains[2][0] != &chains[0][0] || g.nodes[chains[0][1]].count != 2 {
		t.Errorf("parallel edges do not share their chain")
	}
}

func TestFunnel(t *testing.T) {
	src, dst := Point{0, 100}, Point{0, 0}
	tests := []struct {
		gates []gate
		want  []Point
	}{
		{[]gate{{60, -10, 10}, {40, -10, 10}}, []Point{src, dst}},
		{[]gate{{60, 10, 20}, {40, 10, 20}}, []Point{src, {10, 60}, {10, 40}, dst}},
		{[]gate{{60, -20, -10}, {40, -20, -10}}, []Point{src, {-10, 60}, {-10, 40}, dst}},
		{[]gate{{75, 10, 30}, {25, -30, -10}}, []Point{src, {10, 75}, {-10, 25}, dst}},
	}
	for _, tt := range tests {
		got := funnel(src, dst, tt.gates)
		if fmt.Sprint(got) != fmt.Sprint(tt.want) {
			t.Errorf("funnel(%v) = %v, want %v", tt.gates, got, tt.want)
		}
	}
}

func edgesOf(t *testing.T, src string) (*Result, []EdgeResult) {
	t.Helper()
	res, err := Layout(context.Background(), buildGraph(t, src))
	if err != nil {
		t.Fatal(err)
	}
	return res, res.Edges
}

func nodeNamed(res *Result, name string) *NodeResult {
	for i := range res.Nodes {
		if res.Nodes[i].Node.Name == name {
			return &res.Nodes[i]
		}
	}
	return nil
}

// sample returns points along a curve.
func sample(pts []Point) []Point {
	var out []Point
	for _, s := range segments(pts) {
		for k := 0; k <= 20; k++ {
			out = append(out, cubic(s, float64(k)/20))
		}
	}
	return out
}

func TestRoutingScenarios(t *testing.T) {
	res, es := edgesOf(t, "digraph { a -> b -> c; a -> c }")
	b := nodeNamed(res, "b")
	for _, q := range sample(es[2].Points) {
		if insideShape(b, Point{q.X - b.X, q.Y - b.Y}) {
			t.Fatalf("edge a->c passes through b at %v", q)
		}
	}

	_, es = edgesOf(t, "digraph { a -> b; a -> b }")
	if fmt.Sprint(es[0].Points) == fmt.Sprint(es[1].Points) {
		t.Error("two edges between the same nodes share one curve")
	}

	res, es = edgesOf(t, "digraph { a -> a }")
	a := nodeNamed(res, "a")
	start := es[0].Points[0]
	if d := (start.X-a.X)*(start.X-a.X)/(a.W*a.W/4) + (start.Y-a.Y)*(start.Y-a.Y)/(a.H*a.H/4); d < 0.95 || d > 1.05 {
		t.Errorf("loop does not start on the outline of a: %v", start)
	}
	if es[0].Head.Shape == "" || !(maxX(es[0].Points) > a.X+a.W/2) {
		t.Errorf("loop is not beside the node: %+v", es[0])
	}

	// Spec scenario "Flat edge over a node": the curve of a -> c goes over
	// b, and the drawing is 60 points high, within 2%, where its nodes are 36.
	res, es = edgesOf(t, "digraph { a -> c [minlen=0]; b -> c [minlen=0] }")
	b = nodeNamed(res, "b")
	over := false
	for _, q := range sample(es[0].Points) {
		if math.Abs(q.X-b.X) < b.W/2 && q.Y > b.Y+b.H/2 {
			over = true
		}
	}
	if !over || math.Abs(res.Height-60) > 1.2 || b.H != 36 {
		t.Errorf("flat edge over a node: over b %v, drawing %g points high, nodes %g", over, res.Height, b.H)
	}

	// Spec scenario "Label clear of an edge routed before": the label of
	// x -> l lies 28 points right of l, within 2 points.
	res, es = edgesOf(t, `digraph { l -> m; x -> r; x -> l [label="LLLLLL"] }`)
	l := nodeNamed(res, "l")
	for _, e := range es {
		if d := e.LabelPos.X - l.X; e.Edge.Head.Name == "l" && math.Abs(d-28) > 2 {
			t.Errorf("label clear of an edge routed before: %g points right of l", d)
		}
	}

	_, es = edgesOf(t, "digraph { splines=polyline; a -> b -> c; a -> c }")
	for _, e := range es {
		for _, s := range segments(e.Points) {
			if c1, c2 := cross(s[0], s[3], s[1]), cross(s[0], s[3], s[2]); c1*c1 > 1e-6 || c2*c2 > 1e-6 {
				t.Errorf("polyline edge %s->%s has a curved segment %v", e.Edge.Tail.Name, e.Edge.Head.Name, s)
			}
		}
	}
}

func maxX(pts []Point) float64 {
	m := pts[0].X
	for _, p := range pts {
		m = max(m, p.X)
	}
	return m
}

func TestArrowScenarios(t *testing.T) {
	res, es := edgesOf(t, "digraph { a -> b [dir=back] }")
	if es[0].Head.Shape != "" || es[0].Tail.Shape != "normal" {
		t.Fatalf("dir=back: head %q, tail %q", es[0].Head.Shape, es[0].Tail.Shape)
	}
	a := nodeNamed(res, "a")
	if es[0].Tail.Tip.Y < a.Y-a.H/2-3 {
		t.Errorf("tail arrow does not point at a: tip %v", es[0].Tail.Tip)
	}
	_, es = edgesOf(t, "digraph { a -> b [arrowhead=empty]; c -> d [arrowhead=none]; e -> f [arrowsize=2] }")
	if es[0].Head.Shape != "empty" || es[1].Head.Shape != "" {
		t.Errorf("arrowheads %q %q", es[0].Head.Shape, es[1].Head.Shape)
	}
	if l := dist(es[2].Head.Base, es[2].Head.Tip); !approx(l, 2*arrowLength) {
		t.Errorf("arrowsize=2: length %v", l)
	}
	// The curve ends where its arrowhead begins.
	if last := es[0].Points[len(es[0].Points)-1]; dist(last, es[0].Head.Base) > 0.01 {
		t.Errorf("curve ends at %v, arrow base %v", last, es[0].Head.Base)
	}
}

func TestInvisibleEdgeAndGraphLabel(t *testing.T) {
	res, es := edgesOf(t, "digraph { a -> b [style=invis] }")
	if !es[0].Invisible || !(nodeNamed(res, "a").Y > nodeNamed(res, "b").Y) {
		t.Errorf("invisible edge: %+v", es[0])
	}
	res, _ = edgesOf(t, `digraph { label="Title"; a }`)
	a := nodeNamed(res, "a")
	if len(res.Label.Lines) != 1 || !(res.LabelPos.Y < a.Y-a.H/2) || !approx(res.LabelPos.X, res.Width/2) {
		t.Errorf("graph label at %v, node a at %v,%v", res.LabelPos, a.X, a.Y)
	}
}

func TestNodeSizeScenarios(t *testing.T) {
	n := layoutOf(t, `digraph { a [label="a fairly long label"]; b [shape=rect, label="", width=2, height=1]; c [shape=point, label="x"]; d [shape=circle] }`)
	if !(n["a"].W > 54) || n["a"].Label.Width+labelMarginX > n["a"].W {
		t.Errorf("wide label: node %v wide for a label %v wide", n["a"].W, n["a"].Label.Width)
	}
	if !approx(n["b"].W, 144) || !approx(n["b"].H, 72) {
		t.Errorf("rect 2x1: %v x %v", n["b"].W, n["b"].H)
	}
	if n["c"].Shape != "point" || len(n["c"].Label.Lines) != 0 {
		t.Errorf("point: %+v", n["c"])
	}
	if !approx(n["d"].W, n["d"].H) {
		t.Errorf("circle: %v x %v", n["d"].W, n["d"].H)
	}
}

func TestOutlines(t *testing.T) {
	for _, tt := range []struct {
		shape string
		in    []Point
		out   []Point
	}{
		{"ellipse", []Point{{26, 0}, {0, 17}}, []Point{{27.5, 0}, {20, 14}}},
		{"rect", []Point{{26, 17}}, []Point{{28, 0}, {0, 19}}},
		{"diamond", []Point{{13, 8}}, []Point{{14, 10}}},
		{"hexagon", []Point{{26, 0}, {13, 17}}, []Point{{25, 15}}},
		{"octagon", []Point{{26, 7}, {11, 17}}, []Point{{26, 15}}},
	} {
		n := &NodeResult{Shape: tt.shape, W: 54, H: 36}
		for _, p := range tt.in {
			if !insideShape(n, p) {
				t.Errorf("%s: %v should be inside", tt.shape, p)
			}
		}
		for _, p := range tt.out {
			if insideShape(n, p) {
				t.Errorf("%s: %v should be outside", tt.shape, p)
			}
		}
	}
}

func TestRoundedRectangle(t *testing.T) {
	n := layoutOf(t, `digraph { a [shape=rect, style=rounded, label="", width=1, height=0.5]; b [shape=diamond, style=rounded] }`)
	a := n["a"]
	if !a.Rounded || !approx(a.W, 72) || !approx(a.H, 36) || !approx(CornerRadius(a.W, a.H), 12) {
		t.Errorf("rounded rect: %+v", a)
	}
	if n["b"].Rounded {
		t.Error("a diamond got rounded corners")
	}
	// Inside the bounding box but outside the rounded corner.
	if insideShape(&a, Point{35.5, 17.5}) || !insideShape(&a, Point{35.5, 0}) {
		t.Error("corner of the rounded rectangle")
	}
	if got := CornerRadius(14.4, 36); !approx(got, 4.8) {
		t.Errorf("radius of a 14.4 x 36 rectangle: %v", got)
	}
}
