// SPDX-License-Identifier: MIT
// Provenance: original

package layout

import "testing"

// The scenarios of the dot-layout spec that the pprof parity change adds.
func TestPprofLayoutScenarios(t *testing.T) {
	// Box with depth: sized as a box, its front face at its lower left.
	n := nodesOf(layoutResolved(t, `digraph { a [shape=box3d, width=2, height=1, label=""] }`))["a"]
	if !approx(n.W, 144) || !approx(n.H, 72) {
		t.Errorf("box3d node %v x %v, want 144 x 72", n.W, n.H)
	}
	lines := Box3DLines(144, 72)
	corner := lines[0][0]
	if !approx(corner.X+72, 140) || !approx(corner.Y+36, 68) || lines[0][1] != (Point{-72, 32}) || lines[1][1] != (Point{68, -36}) || lines[2][1] != (Point{72, 36}) {
		t.Errorf("box3d depth lines %v", lines)
	}
	if vs := PolygonVertices("box3d", 144, 72); len(vs) != 6 || vs[1] != (Point{-68, 36}) || vs[4] != (Point{68, -36}) {
		t.Errorf("box3d outline %v", vs)
	}

	// Heavy edge kept short: e moves to the rank right below a.
	n2 := nodesOf(layoutResolved(t, `digraph { a -> b -> c -> d; a -> e [weight=5]; e -> d }`))
	if !approx(n2["e"].Y, n2["b"].Y) {
		t.Errorf("heavy edge kept short: e at %v, b at %v", n2["e"].Y, n2["b"].Y)
	}
	n2 = nodesOf(layoutResolved(t, `digraph { a -> b -> c -> d; a -> e; e -> d }`))
	if !approx(n2["e"].Y, n2["c"].Y) {
		t.Errorf("without the weight: e at %v, c at %v", n2["e"].Y, n2["c"].Y)
	}
	// Heavy edge kept straight.
	n2 = nodesOf(layoutResolved(t, `digraph { a -> b; a -> c [weight=10] }`))
	if !approx(n2["a"].X, n2["c"].X) || !(n2["b"].X < n2["a"].X) {
		t.Errorf("heavy edge kept straight: a %v, b %v, c %v", n2["a"].X, n2["b"].X, n2["c"].X)
	}
	// Heavy edge across ranks.
	n2 = nodesOf(layoutResolved(t, `digraph { a -> b -> c; a -> c [weight=5] }`))
	if !approx(n2["a"].X, n2["c"].X) || !(n2["b"].X < n2["a"].X) {
		t.Errorf("heavy edge across ranks: a %v, b %v, c %v", n2["a"].X, n2["b"].X, n2["c"].X)
	}

	// Wider pen: the tip stops 3 points further from b with a 3 point pen,
	// and the arrowhead keeps its size up to 4 points, then widens.
	gap := func(pen string) (float64, Arrow) {
		res := layoutResolved(t, `digraph { a -> b [penwidth=`+pen+`] }`)
		b, e := nodesOf(res)["b"], res.Edges[0]
		return e.Head.Tip.Y - (b.Y + b.H/2), e.Head
	}
	g1, _ := gap("1")
	g3, a3 := gap("3")
	if !approx(float64(int((g3-g1)*10+0.5))/10, 3) {
		t.Errorf("wider pen: tip %v points further, want 3", g3-g1)
	}
	if pts := ArrowPolygon(a3); !approx(dist(a3.Base, a3.Tip), 10) || !approx(dist(pts[0], pts[2]), 7) {
		t.Errorf("arrowhead of a 3 point pen: %v", pts)
	}
	_, a5 := gap("5")
	if pts := ArrowPolygon(a5); !approx(dist(pts[0], pts[2]), 8.75) {
		t.Errorf("arrowhead of a 5 point pen is %v wide, want 8.75", dist(pts[0], pts[2]))
	}
}
