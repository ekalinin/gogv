// SPDX-License-Identifier: MIT
// Provenance: paper Gansner, Koutsofios, North, Vo 1993, A Technique for Drawing Directed Graphs, sections 4 and 5.1.1
// Provenance: observed dot 16.1.0

package layout

import (
	"github.com/ekalinin/gogv/internal/graph"
	"github.com/ekalinin/gogv/internal/label"
)

// portEnd is where an edge end with a port attaches, in the layout
// orientation: the center of the port's box, off from the center of its
// node, and the half extents of the box. set is false for an end without a
// port; table is set for the table's own PORT, whose box the edge meets as
// it meets a node's outline, aimed at the center.
type portEnd struct {
	off, half Point
	set       bool
	table     bool
}

// edgePorts resolves the tailport and the headport of each edge, by
// [tail, head], into the boxes of its nodes' table labels: the table's own
// box for its PORT, or a cell's. A loop, a name that the table does not
// define and a compass point leave the end without a port.
func edgePorts(g *graph.Graph, nodes []NodeResult, ends [][2]int, lr bool) [][2]portEnd {
	out := make([][2]portEnd, len(ends))
	for i, e := range g.Edges() {
		if ends[i][0] == ends[i][1] {
			continue
		}
		for j, attr := range [2]string{"tailport", "headport"} {
			v := edgeValue(g, e, attr)
			n := &nodes[ends[i][j]]
			if v.S == "" || n.Label.Table == nil {
				continue
			}
			port, _ := label.SplitPort(n.Label.Table, v.S)
			box, ok := n.Label.Table.PortBox(port)
			if !ok {
				continue
			}
			// The table is centered on its node, and y grows upward.
			dx := box.X + half(box.W) - half(n.Label.Width)
			dy := half(n.Label.Height) - box.Y - half(box.H)
			pe := portEnd{off: Point{dx, dy}, half: Point{half(box.W), half(box.H)}, set: true, table: port == n.Label.Table.Port}
			if lr {
				pe.off, pe.half = Point{dy, -dx}, Point{pe.half.Y, pe.half.X}
			}
			out[i][j] = pe
		}
	}
	return out
}

// hasPort reports whether an edge has a port at either end.
func hasPort(p [2]portEnd) bool { return p[0].set || p[1].set }

// setPorts gives the ranked edges at the ends of the edges with ports the
// offsets of their ports along the rank, for the positioning, and the same
// as a fraction of the width of their node, for the order; a mirrored order
// runs against the positions, so it takes the fractions the other way. Each
// edge with a port has a chain of its own; width is the extent of the real
// nodes along the rank.
func (g *ranked) setPorts(chains [][]int, ends [][2]int, ports [][2]portEnd, width []float64, mirror bool) {
	at := func(pe portEnd, w float64) float64 {
		if !pe.set || w <= 0 {
			return 0
		}
		f := pe.off.X / w
		if mirror {
			f = -f
		}
		return f
	}
	used := map[int]bool{}
	find := func(a, b int) int {
		for _, e := range g.nodes[a].down {
			if g.edges[e].lower == b && !used[e] {
				return e
			}
		}
		return -1
	}
	for i, c := range chains {
		if c == nil || !hasPort(ports[i]) {
			continue
		}
		upper, lower := ports[i][0], ports[i][1]
		if c[0] != ends[i][0] {
			upper, lower = lower, upper
		}
		first := find(c[0], c[1])
		last := first
		if len(c) > 2 {
			used[first] = true
			last = find(c[len(c)-2], c[len(c)-1])
		}
		if first < 0 || last < 0 {
			continue
		}
		used[first], used[last] = true, true
		g.edges[first].upperPort, g.edges[first].upperAt = upper.off.X, at(upper, width[c[0]])
		g.edges[last].lowerPort, g.edges[last].lowerAt = lower.off.X, at(lower, width[c[len(c)-1]])
	}
	for fi, f := range g.flats {
		p := ports[f.edge]
		g.flats[fi].leftPort, g.flats[fi].rightPort = p[0].off.X, p[1].off.X
	}
}
