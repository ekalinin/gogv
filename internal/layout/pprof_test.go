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
}
