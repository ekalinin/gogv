// SPDX-License-Identifier: MIT
// Provenance: doc https://graphviz.org/docs/attrs/xlabel/
// Provenance: doc https://graphviz.org/docs/attrs/headlabel/
// Provenance: doc https://graphviz.org/docs/attrs/forcelabels/
// Provenance: observed dot 16.1.0

package layout

import (
	"math"
	"slices"

	"github.com/ekalinin/gogv/internal/graph"
	"github.com/ekalinin/gogv/internal/label"
)

// rect is a box in the final orientation.
type rect struct{ x0, y0, x1, y1 float64 }

// overlaps reports whether two boxes share more than their sides, up to a
// hundredth of a point.
func (r rect) overlaps(o rect) bool {
	const eps = 0.01
	return r.x0 < o.x1-eps && o.x0 < r.x1-eps && r.y0 < o.y1-eps && o.y0 < r.y1-eps
}

// placeExternalLabels places the xlabel of nodes and edges and the
// headlabel and taillabel of edges once the layout is done, without moving
// anything, as the reference places them when labelangle and labeldistance
// are not set. Each label has an anchor: the upper left corner of its
// node's box, the middle of its edge, or where the edge meets its head or
// its tail. It takes the first of the corners around its anchor, upper
// left, upper right, lower left, lower right, where it overlaps no node,
// edge label or label placed before; when every corner does, the first one.
// The drawing grows to hold the labels.
func placeExternalLabels(g *graph.Graph, res *Result) {
	var taken []rect
	for _, e := range res.Edges {
		if e.Label.Present() {
			taken = append(taken, centered(e.LabelPos, e.Label))
		}
	}
	var placed []rect
	place := func(t Text, anchor Point) Point {
		w, h := t.Width, t.Height
		corners := []Point{
			{anchor.X - half(w), anchor.Y + half(h)},
			{anchor.X + half(w), anchor.Y + half(h)},
			{anchor.X - half(w), anchor.Y - half(h)},
			{anchor.X + half(w), anchor.Y - half(h)},
		}
		pick := corners[0]
		for _, c := range corners {
			b := centered(c, t)
			if !slices.ContainsFunc(taken, b.overlaps) && !overlapsNode(res, b) {
				pick = c
				break
			}
		}
		taken = append(taken, centered(pick, t))
		placed = append(placed, centered(pick, t))
		return pick
	}

	for i := range res.Nodes {
		n := &res.Nodes[i]
		obj := label.Object{Graph: g.Name, Node: n.Node.Name}
		t := labelOf(nodeValue(g, n.Node, "xlabel"), "", obj, nodeValue(g, n.Node, "fontname"), nodeValue(g, n.Node, "fontcolor"), nodeValue(g, n.Node, "fontsize"))
		if t.Present() {
			n.XLabel = t
			n.XLabelPos = place(t, Point{n.X - half(n.W), n.Y + half(n.H)})
		}
	}
	for i := range res.Edges {
		e := &res.Edges[i]
		if e.Invisible || len(e.Points) == 0 {
			continue
		}
		ed := e.Edge
		obj := label.Object{Graph: g.Name, Edge: true, Tail: ed.Tail.Name, Head: ed.Head.Name, DirectedEdgeName: g.Directed()}
		text := func(name string) Text {
			return labelOf(edgeValue(g, ed, name), "", obj, edgeValue(g, ed, "fontname"), edgeValue(g, ed, "fontcolor"), edgeValue(g, ed, "fontsize"))
		}
		if t := text("xlabel"); t.Present() {
			e.XLabel, e.XLabelPos = t, place(t, e.middle)
		}
		if t := text("headlabel"); t.Present() {
			e.HeadLabel, e.HeadLabelPos = t, place(t, e.headEnd)
		}
		if t := text("taillabel"); t.Present() {
			e.TailLabel, e.TailLabelPos = t, place(t, e.tailEnd)
		}
	}

	// Grow the drawing to hold the labels.
	if len(placed) == 0 {
		return
	}
	lo := Point{0, 0}
	hi := Point{res.Width, res.Height}
	for _, b := range placed {
		lo = Point{math.Min(lo.X, b.x0), math.Min(lo.Y, b.y0)}
		hi = Point{math.Max(hi.X, b.x1), math.Max(hi.Y, b.y1)}
	}
	res.translate(Point{-lo.X, -lo.Y})
	res.Width, res.Height = hi.X-lo.X, hi.Y-lo.Y
}

// overlapsNode reports whether a box overlaps the outline of a node: some
// point of the box, on a grid of its corners, sides and middle, lies inside
// the outline, or the node's center lies inside the box.
func overlapsNode(res *Result, b rect) bool {
	const steps = 4
	for i := range res.Nodes {
		n := &res.Nodes[i]
		if b.x1 <= n.X-half(n.W) || b.x0 >= n.X+half(n.W) || b.y1 <= n.Y-half(n.H) || b.y0 >= n.Y+half(n.H) {
			continue
		}
		if n.X > b.x0 && n.X < b.x1 && n.Y > b.y0 && n.Y < b.y1 {
			return true
		}
		for i := 0; i <= steps; i++ {
			for j := 0; j <= steps; j++ {
				q := Point{lerp(Point{b.x0, 0}, Point{b.x1, 0}, float64(i)/steps).X, lerp(Point{0, b.y0}, Point{0, b.y1}, float64(j)/steps).Y}
				// Strictly inside: points on the outline do not count.
				d := sub(q, Point{n.X, n.Y})
				if insideShape(n, scale(d, 1.01)) {
					return true
				}
			}
		}
	}
	return false
}

// centered returns the box of a label centered on c.
func centered(c Point, t Text) rect {
	return rect{c.X - half(t.Width), c.Y - half(t.Height), c.X + half(t.Width), c.Y + half(t.Height)}
}
