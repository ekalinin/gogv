// SPDX-License-Identifier: MIT
// Provenance: original

package layout

import (
	"context"
	"testing"
)

// twoPorts is the table of the "Node in line with a port" scenario: cells l
// and r side by side; stacked has cells t and b one above the other.
const (
	twoPorts = `<<TABLE BORDER="0" CELLBORDER="0" CELLSPACING="0" CELLPADDING="0"><TR><TD FIXEDSIZE="TRUE" WIDTH="40" HEIGHT="20" PORT="l"></TD><TD FIXEDSIZE="TRUE" WIDTH="60" HEIGHT="30" PORT="r"></TD></TR></TABLE>>`
	stacked  = `<<TABLE BORDER="0" CELLBORDER="0" CELLSPACING="0" CELLPADDING="0"><TR><TD FIXEDSIZE="TRUE" WIDTH="40" HEIGHT="20" PORT="t"></TD></TR><TR><TD FIXEDSIZE="TRUE" WIDTH="40" HEIGHT="30" PORT="b"></TD></TR></TABLE>>`
)

func TestEdgePortOffsets(t *testing.T) {
	for _, tt := range []struct {
		name string
		src  string
		lr   bool
		want []portEnd
	}{
		{"TB", `digraph { a [shape=plaintext, label=` + twoPorts + `]; a:r -> b; a:l -> c; a:x -> d; a:n -> e; a:r -> a }`, false, []portEnd{
			{off: Point{20, 0}, half: Point{30, 15}, set: true},
			{off: Point{-30, 0}, half: Point{20, 15}, set: true},
			{}, {}, {},
		}},
		{"LR", `digraph { rankdir=LR; a [shape=plaintext, label=` + stacked + `]; a:t -> b; a:b -> c }`, true, []portEnd{
			{off: Point{15, 0}, half: Point{10, 20}, set: true},
			{off: Point{-10, 0}, half: Point{15, 20}, set: true},
		}},
	} {
		g := buildGraph(t, tt.src)
		res, err := Layout(context.Background(), g)
		if err != nil {
			t.Fatal(err)
		}
		index := map[string]int{}
		for i, n := range g.Nodes() {
			index[n.Name] = i
		}
		var ends [][2]int
		for _, e := range g.Edges() {
			ends = append(ends, [2]int{index[e.Tail.Name], index[e.Head.Name]})
		}
		ports := edgePorts(g, res.Nodes, ends, tt.lr)
		for i, want := range tt.want {
			if got := ports[i][0]; got != want {
				t.Errorf("%s: tail port of edge %d: %+v, want %+v", tt.name, i, got, want)
			}
		}
	}
}

func TestPortScenarios(t *testing.T) {
	// Node in line with a port: b lies under the center of the cell r, and
	// the edge starts at the bottom of that cell.
	res := layoutResolved(t, `digraph { a [shape=plaintext, label=`+twoPorts+`]; a:r -> b }`)
	n, e := nodesOf(res), edgeOf(t, res, "a", "b")
	if !approx(n["b"].X-n["a"].X, 20) || !approx(e.Points[0].X, n["a"].X+20) || !approx(e.Points[0].Y, n["a"].Y-15) {
		t.Errorf("node in line with a port: a %v, b %v, edge from %v", n["a"], n["b"], e.Points[0])
	}

	// Order of the ports, and the compromise of the reference: a at 68
	// points, b at 99 and c at 27.
	res = layoutResolved(t, `digraph { a [shape=plaintext, label=`+twoPorts+`]; a:r -> b; a:l -> c }`)
	n = nodesOf(res)
	if !(n["c"].X < n["b"].X) || !approx(n["a"].X, 68) || !approx(n["b"].X, 99) || !approx(n["c"].X, 27) {
		t.Errorf("order of the ports: a %v, b %v, c %v", n["a"].X, n["b"].X, n["c"].X)
	}

	// A flat edge between ports goes straight from the side of one cell to
	// the side of the other.
	grid := `<<TABLE BORDER="0" CELLBORDER="0" CELLSPACING="0" CELLPADDING="0"><TR><TD></TD><TD FIXEDSIZE="TRUE" WIDTH="1" HEIGHT="5"></TD><TD></TD></TR><TR><TD FIXEDSIZE="TRUE" WIDTH="20" HEIGHT="1"></TD><TD FIXEDSIZE="TRUE" WIDTH="40" HEIGHT="20" PORT="h"></TD><TD FIXEDSIZE="TRUE" WIDTH="20" HEIGHT="1"></TD></TR><TR><TD></TD><TD FIXEDSIZE="TRUE" WIDTH="1" HEIGHT="5"></TD><TD></TD></TR></TABLE>>`
	res = layoutResolved(t, `digraph { a [shape=plaintext, label=`+grid+`]; b [shape=plaintext, label=`+grid+`]; a:h -> b:h [minlen=0, arrowhead=none] }`)
	n, e = nodesOf(res), edgeOf(t, res, "a", "b")
	first, last := e.Points[0], e.Points[len(e.Points)-1]
	if !approx(n["a"].Y, n["b"].Y) || !approx(first.X, n["a"].X+20) || !approx(last.X, n["b"].X-20) {
		t.Errorf("flat edge between ports: a %v, b %v, edge %v", n["a"], n["b"], e.Points)
	}
	for _, p := range e.Points {
		if !approx(p.Y, n["a"].Y) {
			t.Errorf("flat edge between ports is not straight: %v", e.Points)
			break
		}
	}

	// In rankdir=LR, the edge leaves the right side of the cell at its
	// vertical center.
	res = layoutResolved(t, `digraph { rankdir=LR; a [shape=plaintext, label=`+stacked+`]; a:b -> b }`)
	n, e = nodesOf(res), edgeOf(t, res, "a", "b")
	if !approx(e.Points[0].X, n["a"].X+20) || !approx(e.Points[0].Y, n["a"].Y-10) {
		t.Errorf("left to right: a %v, edge from %v", n["a"], e.Points[0])
	}
}
