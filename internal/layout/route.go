// SPDX-License-Identifier: MIT
// Provenance: paper Dobkin, Gansner, Koutsofios, North 1997, Implementing a General-Purpose Edge Router
// Provenance: paper Gansner, Koutsofios, North, Vo 1993, A Technique for Drawing Directed Graphs, section 5
// Provenance: observed dot 16.1.0

package layout

import (
	"math"
	"slices"

	"github.com/ekalinin/gogv/internal/attrs"
	"github.com/ekalinin/gogv/internal/graph"
)

// The arrowheads of the arrow shapes documentation, in points at
// arrowsize 1.
const (
	arrowLength    = 10
	arrowHalfWidth = 3.5
)

// arrowMiter is how far the reference keeps the tip of an arrowhead from
// the end of its edge, so that the stroke of the tip does not reach into
// the node: half the pen width over the sine of half the tip angle.
var arrowMiter = 0.5 * hypot(arrowLength, arrowHalfWidth) / arrowHalfWidth

// PolygonVertices returns the outline of a polygonal shape relative to its
// center, y growing upward, in the order the reference writes them.
func PolygonVertices(shape string, w, h float64) []Point {
	hw, hh := w/2, h/2
	switch shape {
	case "box", "rect", "rectangle", "plaintext":
		return []Point{{hw, hh}, {-hw, hh}, {-hw, -hh}, {hw, -hh}}
	case "box3d":
		// A box whose upper left and lower right corners are cut by its
		// depth.
		d := float64(box3dDepth)
		hw, hh := half(w), half(h)
		return []Point{{hw, hh}, {-hw + d, hh}, {-hw, hh - d}, {-hw, -hh}, {hw - d, -hh}, {hw, -hh + d}}
	case "diamond":
		return []Point{{0, hh}, {-hw, 0}, {0, -hh}, {hw, 0}}
	case "hexagon":
		return []Point{{hw, 0}, {hw / 2, hh}, {-hw / 2, hh}, {-hw, 0}, {-hw / 2, -hh}, {hw / 2, -hh}}
	case "octagon":
		k := math.Sqrt2 - 1 // tan(pi/8)
		kw, kh := float64(k*hw), float64(k*hh)
		return []Point{{hw, -kh}, {hw, kh}, {kw, hh}, {-kw, hh}, {-hw, kh}, {-hw, -kh}, {-kw, -hh}, {kw, -hh}}
	}
	return nil
}

// box3dDepth is the depth of a box3d node, in points, whatever its size.
const box3dDepth = 4

// Box3DLines returns the three lines that draw the depth of a box3d node,
// relative to its center, y growing upward: from the upper right corner of
// its front face to the left side, to the bottom side and to the upper right
// corner of the node.
func Box3DLines(w, h float64) [3][2]Point {
	hw, hh := half(w), half(h)
	corner := Point{hw - box3dDepth, hh - box3dDepth}
	return [3][2]Point{{corner, {-hw, hh - box3dDepth}}, {corner, {hw - box3dDepth, -hh}}, {corner, {hw, hh}}}
}

// CornerRadius is the radius of the corners of a rounded rectangle: a
// third of its smaller side, at most 12 points.
func CornerRadius(w, h float64) float64 { return min(12, min(w, h)/3) }

// RoundedOutline returns the outline of a rounded rectangle relative to its
// center, y growing upward, as a piecewise cubic Bezier curve: from the right
// end of the top side, counterclockwise, with straight sides as degenerate
// segments and each corner as one segment whose control points are half a
// radius from its ends towards the corner.
func RoundedOutline(w, h float64) []Point {
	hw, hh := half(w), half(h)
	r := CornerRadius(w, h)
	k := half(r)
	pts := []Point{{hw - r, hh}}
	side := func(to Point) {
		from := pts[len(pts)-1]
		pts = append(pts, from, to, to)
	}
	corner := func(c1, c2, to Point) { pts = append(pts, c1, c2, to) }
	side(Point{-hw + r, hh})
	corner(Point{-hw + k, hh}, Point{-hw, hh - k}, Point{-hw, hh - r})
	side(Point{-hw, -hh + r})
	corner(Point{-hw, -hh + k}, Point{-hw + k, -hh}, Point{-hw + r, -hh})
	side(Point{hw - r, -hh})
	corner(Point{hw - k, -hh}, Point{hw, -hh + k}, Point{hw, -hh + r})
	side(Point{hw, hh - r})
	corner(Point{hw, hh - k}, Point{hw - k, hh}, Point{hw - r, hh})
	return pts
}

// insideShape reports whether a point, relative to the center of a node in
// the final orientation, is inside the node's outline.
func insideShape(n *NodeResult, d Point) bool {
	hw, hh := half(n.W), half(n.H)
	if hw <= 0 || hh <= 0 {
		return false
	}
	switch n.Shape {
	case "ellipse", "oval", "circle", "point":
		x, y := d.X/hw, d.Y/hh
		return float64(x*x)+float64(y*y) <= 1
	case "box", "rect", "rectangle", "plaintext":
		if math.Abs(d.X) > hw || math.Abs(d.Y) > hh {
			return false
		}
		if !n.Rounded {
			return true
		}
		// Outside the quarter circle of a rounded corner.
		r := CornerRadius(n.W, n.H)
		cx, cy := math.Abs(d.X)-(hw-r), math.Abs(d.Y)-(hh-r)
		return cx <= 0 || cy <= 0 || float64(cx*cx)+float64(cy*cy) <= float64(r*r)
	case "diamond":
		return math.Abs(d.X)/hw+math.Abs(d.Y)/hh <= 1
	}
	vs := PolygonVertices(n.Shape, n.W, n.H)
	for i := range vs {
		// Counterclockwise outline: the point must not be right of any side.
		if cross(vs[i], vs[(i+1)%len(vs)], d) < 0 {
			return false
		}
	}
	return true
}

// router routes the edges in the layout orientation.
type router struct {
	g   *ranked
	pos []Point // ranked node positions
	// left and right are how far each ranked node reaches from its x.
	left, right []float64
	bandUp      []float64 // by rank: how far the rank's band reaches above its y
	bandDn      []float64 // and below
	ys          []float64
	nodesep     float64
	lr          bool
	nodes       []NodeResult
	// tree holds the clusters, when the graph has them.
	tree *clusterTree
	// labelAlong is, for the ranked nodes that hold an edge label, the
	// label's extent along the rank.
	labelAlong map[int]float64
	// loopLabels is, by edge, the extent along the rank of the label of a
	// loop.
	loopLabels []float64
	// flatLabels numbers the labeled flat edges between two neighbors, by
	// edge, and flatAcross is the height of their labels.
	flatLabels map[int]int
	flatAcross map[int]float64
	// sameTail lists, by edge, the edges of its sametail group.
	sameTail map[int][]int
	// ports holds, by edge, where its tail and its head attach when they
	// have ports.
	ports [][2]portEnd
}

// end returns where end j (0 for the tail, 1 for the head) of edge i
// attaches to the ranked node v: the center of its port's box, or the
// center of the node.
func (r *router) end(i, j, v int) Point {
	if pe := r.ports[i][j]; pe.set {
		return add(r.pos[v], pe.off)
	}
	return r.pos[v]
}

// insideEnd returns the test that clips end j of edge i at its node: the
// box of its port, or the node's outline.
func (r *router) insideEnd(i, j, v int) func(Point) bool {
	if pe := r.ports[i][j]; pe.set {
		c := add(r.pos[v], pe.off)
		return func(q Point) bool {
			return math.Abs(q.X-c.X) <= pe.half.X && math.Abs(q.Y-c.Y) <= pe.half.Y
		}
	}
	c := r.pos[v]
	return func(q Point) bool { return r.inside(v, c, q) }
}

// sameTailPoint returns the point of the tail's outline that the edges of
// edge i's sametail group share: in the mean direction from the tail to
// their heads, as documented for sametail.
func (r *router) sameTailPoint(i int, ends [][2]int) (Point, bool) {
	group := r.sameTail[i]
	if len(group) < 2 {
		return Point{}, false
	}
	t := ends[i][0]
	c := r.pos[t]
	var d Point
	for _, j := range group {
		d = add(d, normalize(sub(r.pos[ends[j][1]], c)))
	}
	if d == (Point{}) {
		return Point{}, false
	}
	d = normalize(d)
	// Out along d until the point leaves the outline.
	far := max(r.left[t], r.right[t], r.nodes[r.g.nodes[t].real].W, r.nodes[r.g.nodes[t].real].H) * 2
	seg := [4]Point{c, c, add(c, scale(d, far)), add(c, scale(d, far))}
	in := func(q Point) bool { return r.inside(t, c, q) }
	return cubic(seg, boundary(seg, in, 0, 1)), true
}

// toFinal turns an offset in the layout orientation into one in the final
// orientation.
func (r *router) toFinal(d Point) Point {
	if r.lr {
		return Point{-d.Y, d.X}
	}
	return d
}

// inside reports whether a layout point is inside a real node.
func (r *router) inside(real int, center, p Point) bool {
	return insideShape(&r.nodes[real], r.toFinal(sub(p, center)))
}

// bounds returns the free horizontal range around a ranked node: from the
// right side of its left neighbor plus half nodesep to the left side of its
// right neighbor minus half nodesep. With clusters, the box of a cluster
// that has the neighbor and not the node bounds the range instead, half a
// margin away, and so does the box of the innermost cluster of the node.
// The vertical range [down, up] of the bound widens to the boxes that bound
// it, unless the edge enters them: ends are the ends of the edge.
func (r *router) bounds(v int, ends [2]int) (lo, hi, down, up float64) {
	rn := r.g.nodes[v]
	rank := r.g.ranks[rn.rank]
	lo, hi = math.Inf(-1), math.Inf(1)
	down, up = math.Inf(1), math.Inf(-1)
	box := func(c int) [2]Point {
		b := r.tree.clusters[c].box
		if !r.tree.contains(c, r.g.clust[ends[0]]) && !r.tree.contains(c, r.g.clust[ends[1]]) {
			down, up = min(down, b[0].Y), max(up, b[1].Y)
		}
		return b
	}
	var cv int
	if r.tree != nil {
		cv = r.g.clust[v]
		if cv > 0 {
			b := r.tree.clusters[cv].box
			lo, hi = b[0].X+half(clusterMargin), b[1].X-half(clusterMargin)
		}
	}
	if rn.order > 0 {
		u := rank[rn.order-1]
		if cu := r.clusterOf(u); cu >= 0 && r.tree.common(cv, cu) != cu {
			a := r.tree.childToward(r.tree.common(cv, cu), cu)
			lo = max(lo, box(a)[1].X+half(clusterMargin))
		} else {
			lo = max(lo, r.pos[u].X+r.right[u]+half(r.nodesep))
		}
	}
	if rn.order+1 < len(rank) {
		u := rank[rn.order+1]
		if cu := r.clusterOf(u); cu >= 0 && r.tree.common(cv, cu) != cu {
			a := r.tree.childToward(r.tree.common(cv, cu), cu)
			hi = min(hi, box(a)[0].X-half(clusterMargin))
		} else if along, ok := r.labelAlong[v]; ok {
			// Observed on the reference: the label, right of the edge,
			// keeps a quarter of nodesep from the edge of a virtual node
			// that comes next, and half nodesep from the side of a real
			// one.
			next := r.pos[u].X - math.Floor(r.nodesep/4)
			if r.g.nodes[u].real >= 0 {
				next = r.pos[u].X - r.left[u] - half(r.nodesep)
			}
			hi = min(hi, next-along)
		} else {
			hi = min(hi, r.pos[u].X-r.left[u]-half(r.nodesep))
		}
	}
	if _, ok := r.labelAlong[v]; ok && labelMode == 1 {
		hi = min(hi, r.pos[v].X)
	}
	return lo, hi, down, up
}

var labelMode = 2

// clusterOf returns the innermost cluster of a ranked node, or -1 when the
// graph has no clusters.
func (r *router) clusterOf(v int) int {
	if r.tree == nil {
		return -1
	}
	return r.g.clust[v]
}

// gate is a horizontal opening the path crosses, at height y.
type gate struct {
	y, lo, hi float64
}

// chainPath returns the shortest path from src, on a chain's upper node, to
// dst, on its lower node, that stays within the free space around its
// virtual nodes: the corridor of section 5, with the funnel algorithm for
// the path.
func (r *router) chainPath(chain []int, src, dst Point) []Point {
	var gates []gate
	for _, v := range chain[1 : len(chain)-1] {
		lo, hi, down, up := r.bounds(v, [2]int{chain[0], chain[len(chain)-1]})
		rank := r.g.nodes[v].rank
		// A box widens the gates up to the bands of the neighboring ranks.
		up = min(up, r.ys[rank-1]-r.bandDn[rank-1])
		down = max(down, r.ys[rank+1]+r.bandUp[rank+1])
		gates = append(gates,
			gate{y: max(r.ys[rank]+r.bandUp[rank], up), lo: lo, hi: hi},
			gate{y: min(r.ys[rank]-r.bandDn[rank], down), lo: lo, hi: hi})
	}
	return funnel(src, dst, gates)
}

// funnel finds the shortest path from src down to dst through horizontal
// gates of decreasing y, by the funnel (string pulling) method.
func funnel(src, dst Point, gates []gate) []Point {
	// Walking down, the left hand points to +x.
	type portal struct{ left, right Point }
	var portals []portal
	for _, g := range gates {
		lo, hi := g.lo, g.hi
		if math.IsInf(lo, -1) {
			lo = min(src.X, dst.X, hi) - 1e6
		}
		if math.IsInf(hi, 1) {
			hi = max(src.X, dst.X, lo) + 1e6
		}
		portals = append(portals, portal{left: Point{hi, g.y}, right: Point{lo, g.y}})
	}
	portals = append(portals, portal{left: dst, right: dst})

	path := []Point{src}
	apex, left, right := src, src, src
	apexIdx, leftIdx, rightIdx := -1, -1, -1
	for i := 0; i < len(portals); i++ {
		pl, pr := portals[i].left, portals[i].right
		// Tighten the right side.
		if cross(apex, right, pr) >= 0 {
			if apex == right || cross(apex, left, pr) < 0 {
				right, rightIdx = pr, i
			} else {
				path = append(path, left)
				apex, apexIdx = left, leftIdx
				left, right = apex, apex
				leftIdx, rightIdx = apexIdx, apexIdx
				i = apexIdx
				continue
			}
		}
		// Tighten the left side.
		if cross(apex, left, pl) <= 0 {
			if apex == left || cross(apex, right, pl) > 0 {
				left, leftIdx = pl, i
			} else {
				path = append(path, right)
				apex, apexIdx = right, rightIdx
				left, right = apex, apex
				leftIdx, rightIdx = apexIdx, apexIdx
				i = apexIdx
				continue
			}
		}
	}
	if path[len(path)-1] != dst {
		path = append(path, dst)
	}
	return path
}

// splineThrough returns a piecewise cubic Bezier curve through the points
// of a path: the tangent at an inner point follows the direction from the
// point before it to the point after it, with a length of a third of each
// segment; at both ends the curve leaves and arrives without a tangent, as
// the reference's curves between node centers do.
func splineThrough(path []Point) []Point {
	n := len(path)
	tangent := make([]Point, n)
	for i := 1; i+1 < n; i++ {
		tangent[i] = normalize(sub(path[i+1], path[i-1]))
	}
	out := []Point{path[0]}
	for i := 0; i+1 < n; i++ {
		l := dist(path[i], path[i+1]) / 3
		out = append(out,
			add(path[i], scale(tangent[i], l)),
			sub(path[i+1], scale(tangent[i+1], l)),
			path[i+1])
	}
	return out
}

// curveMiddle returns the point halfway along a piecewise curve.
func curveMiddle(segs [][4]Point) Point {
	const steps = 32
	var pts []Point
	for _, sg := range segs {
		for k := 0; k <= steps; k++ {
			pts = append(pts, cubic(sg, float64(k)/steps))
		}
	}
	var total float64
	for i := 1; i < len(pts); i++ {
		total += dist(pts[i-1], pts[i])
	}
	var run float64
	for i := 1; i < len(pts); i++ {
		d := dist(pts[i-1], pts[i])
		if run+d >= half(total) && d > 0 {
			return lerp(pts[i-1], pts[i], (half(total)-run)/d)
		}
		run += d
	}
	return pts[0]
}

// segments splits a piecewise curve into its cubic segments.
func segments(pts []Point) [][4]Point {
	var segs [][4]Point
	for i := 0; i+3 < len(pts); i += 3 {
		segs = append(segs, [4]Point{pts[i], pts[i+1], pts[i+2], pts[i+3]})
	}
	return segs
}

func join(segs [][4]Point) []Point {
	if len(segs) == 0 {
		return nil
	}
	out := []Point{segs[0][0]}
	for _, s := range segs {
		out = append(out, s[1], s[2], s[3])
	}
	return out
}

// boundary finds, by bisection, where a segment crosses from inside (at t0)
// to outside (at t1) of a node, or the reverse.
func boundary(seg [4]Point, in func(Point) bool, t0, t1 float64) float64 {
	for range 50 {
		m := (t0 + t1) / 2
		if in(cubic(seg, m)) == in(cubic(seg, t0)) {
			t0 = m
		} else {
			t1 = m
		}
	}
	return (t0 + t1) / 2
}

// clipStart removes the part of a curve that lies inside the node it starts
// in.
func clipStart(segs [][4]Point, in func(Point) bool) [][4]Point {
	for i, s := range segs {
		if !in(s[0]) {
			return segs[i:]
		}
		if !in(s[3]) {
			t := boundary(s, in, 0, 1)
			_, right := splitCubic(s, t)
			return append([][4]Point{right}, segs[i+1:]...)
		}
	}
	return segs
}

// clipEnd removes the part of a curve that lies inside the node it ends in.
func clipEnd(segs [][4]Point, in func(Point) bool) [][4]Point {
	for i := len(segs) - 1; i >= 0; i-- {
		s := segs[i]
		if !in(s[3]) {
			return segs[:i+1]
		}
		if !in(s[0]) {
			t := boundary(s, in, 1, 0)
			left, _ := splitCubic(s, t)
			return append(slices.Clone(segs[:i]), left)
		}
	}
	return segs
}

// reverseCurve reverses the direction of a piecewise curve.
func reverseCurve(pts []Point) []Point {
	out := slices.Clone(pts)
	slices.Reverse(out)
	return out
}

// arrowAtEnd shortens a curve for an arrowhead at its end: the tip is
// arrowMiter before the end along the curve, the base length further, and
// the curve stops at the base. It returns the shortened curve and the arrow.
func arrowAtEnd(segs [][4]Point, shape string, size float64) ([][4]Point, Arrow) {
	end := segs[len(segs)-1][3]
	length := float64(arrowLength * size)
	reach := length + arrowMiter
	// The last point of the curve at least reach away from its end.
	for i := len(segs) - 1; i >= 0; i-- {
		s := segs[i]
		if dist(s[0], end) < reach && i > 0 {
			continue
		}
		if dist(s[0], end) < reach {
			break
		}
		t0, t1 := 0.0, 1.0
		for range 50 {
			m := (t0 + t1) / 2
			if dist(cubic(s, m), end) > reach {
				t0 = m
			} else {
				t1 = m
			}
		}
		left, _ := splitCubic(s, t0)
		base := left[3]
		tip := add(base, scale(normalize(sub(end, base)), length))
		return append(slices.Clone(segs[:i]), left), Arrow{Shape: shape, Size: size, Base: base, Tip: tip}
	}
	// The curve is shorter than the arrow: draw the arrow over it.
	start := segs[0][0]
	tip := end
	base := sub(tip, scale(normalize(sub(end, start)), length))
	return nil, Arrow{Shape: shape, Size: size, Base: base, Tip: tip}
}

// ArrowPolygon returns the outline of an arrowhead: the two corners of its
// base and its tip.
func ArrowPolygon(a Arrow) []Point {
	dir := normalize(sub(a.Tip, a.Base))
	perp := Point{-dir.Y, dir.X}
	hw := arrowHalfWidth * a.Size
	return []Point{add(a.Base, scale(perp, hw)), a.Tip, sub(a.Base, scale(perp, hw))}
}

// edgeAttrs are the attributes that decide how an edge is drawn.
type edgeAttrs struct {
	invisible  bool
	style      string
	dotted     bool
	color      string
	head, tail string // arrowhead shapes, "" for none
	size       float64
	penWidth   float64
}

func readEdgeAttrs(g *graph.Graph, e *graph.Edge) edgeAttrs {
	style := supported(graph.EdgeKind, "style", edgeValue(g, e, "style"), "solid")
	styles := attrs.Styles(style)
	a := edgeAttrs{
		invisible: slices.Contains(styles, "invis"),
		style:     style,
		dotted:    slices.Contains(styles, "dotted"),
		color:     supported(graph.EdgeKind, "color", edgeValue(g, e, "color"), "black"),
		size:      number(edgeValue(g, e, "arrowsize"), 1, 0),
		penWidth:  1,
	}
	if v := supported(graph.EdgeKind, "penwidth", edgeValue(g, e, "penwidth"), ""); v != "" {
		a.penWidth, _ = parseNumber(v)
	}
	defDir := "forward"
	if !g.Directed() {
		defDir = "none"
	}
	dir := supported(graph.EdgeKind, "dir", edgeValue(g, e, "dir"), defDir)
	if !g.Directed() && !isSet(edgeValue(g, e, "dir")) {
		dir = "none"
	}
	head := supported(graph.EdgeKind, "arrowhead", edgeValue(g, e, "arrowhead"), "normal")
	tail := supported(graph.EdgeKind, "arrowtail", edgeValue(g, e, "arrowtail"), "normal")
	if (dir == "forward" || dir == "both") && head != "none" {
		a.head = head
	}
	if (dir == "back" || dir == "both") && tail != "none" {
		a.tail = tail
	}
	return a
}

// routeEdges routes every edge of the graph, in edge order, in the layout
// orientation.
func (r *router) routeEdges(g *graph.Graph, ends [][2]int, chains [][]int, polyline bool) []EdgeResult {
	edges := g.Edges()
	out := make([]EdgeResult, len(edges))
	// Parallel edges: their index among the edges between the same nodes.
	type pairKey struct{ a, b int }
	groups := map[pairKey][]int{}
	keyOf := func(i int) pairKey {
		t, h := ends[i][0], ends[i][1]
		if t > h {
			t, h = h, t
		}
		return pairKey{t, h}
	}
	for i := range edges {
		k := keyOf(i)
		groups[k] = append(groups[k], i)
	}
	// Among parallel edges, those that go down the ranks come first, then
	// the reversed ones, each in edge order.
	reversed := func(i int) bool {
		return r.g.nodes[ends[i][0]].rank > r.g.nodes[ends[i][1]].rank
	}
	for _, group := range groups {
		slices.SortStableFunc(group, func(a, b int) int {
			ra, rb := reversed(a), reversed(b)
			switch {
			case ra == rb:
				return 0
			case !ra:
				return -1
			}
			return 1
		})
	}
	for i, e := range edges {
		a := readEdgeAttrs(g, e)
		res := EdgeResult{Edge: e, Invisible: a.invisible, Color: a.color, Style: a.style, Dotted: a.dotted, PenWidth: a.penWidth}
		group := groups[keyOf(i)]
		idx := slices.Index(group, i)
		off := float64(idx) - half(float64(len(group)-1))
		t, h := ends[i][0], ends[i][1]
		var pts []Point
		switch {
		case t == h:
			// Nested loops reach past the loops inside them: nodesep
			// beyond an unlabeled one, past the label of a labeled one.
			var inner float64
			for _, j := range group[:idx] {
				if r.loopLabels[j] > 0 {
					inner += r.loopLabels[j]
				} else {
					inner += r.nodesep
				}
			}
			pts = r.loop(t, inner)
		case chains[i] == nil:
			if k, ok := r.flatLabels[i]; ok {
				pts = r.labeledFlat(t, h, k, r.flatAcross[i])
			} else {
				pts = r.flat(t, h, off)
			}
			// From port to port: the ends move with their control points.
			if n := len(pts); n == 4 {
				dt, dh := sub(r.end(i, 0, t), r.pos[t]), sub(r.end(i, 1, h), r.pos[h])
				pts[0], pts[1] = add(pts[0], dt), add(pts[1], dt)
				pts[2], pts[3] = add(pts[2], dh), add(pts[3], dh)
			}
		default:
			// The ends at the top and at the bottom of the chain. An end
			// at the port of a cell leaves its box from the middle of the
			// side that faces the other end, straight across the ranks, as
			// the port boxes of section 5.1.1 lead it.
			tj, bj, tv, bv := 0, 1, t, h
			if r.g.nodes[t].rank > r.g.nodes[h].rank {
				tj, bj, tv, bv = 1, 0, h, t
			}
			top, bottom := r.end(i, tj, tv), r.end(i, bj, bv)
			cellT, cellB := r.ports[i][tj].set && !r.ports[i][tj].table, r.ports[i][bj].set && !r.ports[i][bj].table
			if cellT {
				top.Y -= r.ports[i][tj].half.Y
			}
			if cellB {
				bottom.Y += r.ports[i][bj].half.Y
			}
			pts = r.chain(chains[i], top, bottom, float64(off*r.nodesep), polyline)
			if n := len(pts); n >= 4 {
				if cellT {
					straightAcross(pts, 0, 1, 3)
				}
				if cellB {
					straightAcross(pts, n-1, n-2, n-4)
				}
			}
			if r.g.nodes[t].rank > r.g.nodes[h].rank {
				pts = reverseCurve(pts)
			}
		}
		if p, ok := r.sameTailPoint(i, ends); ok && t != h && len(pts) > 1 && !r.ports[i][0].set {
			// The curve leaves from the shared point, its first control
			// point moved with it.
			pts[1] = add(pts[1], sub(p, pts[0]))
			pts[0] = p
		}
		segs := segments(pts)
		segs = clipStart(segs, r.insideEnd(i, 0, t))
		segs = clipEnd(segs, r.insideEnd(i, 1, h))
		if len(segs) > 0 {
			// Where the edge meets its nodes, before the arrowheads, and
			// its middle, for the external labels.
			res.tailEnd, res.headEnd = segs[0][0], segs[len(segs)-1][3]
			res.middle = curveMiddle(segs)
		}
		if a.head != "" && len(segs) > 0 {
			segs, res.Head = arrowAtEnd(segs, a.head, a.size)
		}
		if a.tail != "" && len(segs) > 0 {
			rev := segments(reverseCurve(join(segs)))
			rev, res.Tail = arrowAtEnd(rev, a.tail, a.size)
			segs = segments(reverseCurve(join(rev)))
		}
		res.Points = join(segs)
		out[i] = res
	}
	return out
}

// straightAcross puts the control point next of the end at straight across
// the ranks from it: at its x, and a third of the way to the point other
// when it is level with the end.
func straightAcross(pts []Point, at, next, other int) {
	if pts[next].Y == pts[at].Y {
		pts[next].Y = pts[at].Y + (pts[other].Y-pts[at].Y)/3
	}
	pts[next].X = pts[at].X
}

// chain returns the curve of an edge through its chain of ranked nodes, from
// top, where it attaches to the upper node, to bottom, where it attaches to
// the lower one, moved sideways by off for parallel edges.
func (r *router) chain(chain []int, top, bottom Point, off float64, polyline bool) []Point {
	if len(chain) == 2 {
		if polyline {
			return []Point{top, lerp(top, bottom, 1.0/3), lerp(top, bottom, 2.0/3), bottom}
		}
		return []Point{top, add(top, Point{off, 0}), add(bottom, Point{off, 0}), bottom}
	}
	path := r.chainPath(chain, top, bottom)
	if len(path) == 2 && !polyline {
		// Nothing in the way: one curve, as between neighboring ranks.
		return []Point{top, add(top, Point{off, 0}), add(bottom, Point{off, 0}), bottom}
	}
	for j := 1; j+1 < len(path); j++ {
		path[j].X += off
	}
	if polyline {
		var pts []Point
		for j := 0; j+1 < len(path); j++ {
			if j == 0 {
				pts = append(pts, path[0])
			}
			pts = append(pts, lerp(path[j], path[j+1], 1.0/3), lerp(path[j], path[j+1], 2.0/3), path[j+1])
		}
		return pts
	}
	return splineThrough(path)
}

// flat returns the curve of an edge between two nodes of the same rank: a
// straight curve when they are neighbors, an arc above the rank otherwise.
func (r *router) flat(t, h int, off float64) []Point {
	a, b := r.pos[t], r.pos[h]
	ot, oh := r.g.nodes[t].order, r.g.nodes[h].order
	if ot-oh == 1 || oh-ot == 1 {
		d := Point{0, float64(off * r.nodesep)}
		return []Point{a, add(a, d), add(b, d), b}
	}
	lift := r.bandUp[r.g.nodes[t].rank] + half(r.nodesep) + math.Abs(float64(off*r.nodesep))
	return []Point{a, add(a, Point{0, lift}), add(b, Point{0, lift}), b}
}

// labeledFlat returns the curve of the k-th labeled flat edge between two
// neighbors, whose labels are across high: as observed on the reference,
// the first is straight, and the next ones bend away from it, alternately
// below and above, a label's height plus 4.3 points further at each level.
func (r *router) labeledFlat(t, h, k int, across float64) []Point {
	a, b := r.pos[t], r.pos[h]
	if k == 0 {
		return []Point{a, a, b, b}
	}
	sign, level := flatLevel(k)
	// A cubic reaches three quarters of the displacement of its control
	// points.
	d := Point{0, float64(sign*float64(level)*(across+4.3)) / 0.75}
	return []Point{a, add(a, d), add(b, d), b}
}

// flatLevel returns the side (1 above, -1 below) and the level of the k-th
// labeled flat edge between two neighbors.
func flatLevel(k int) (sign float64, level int) {
	sign = 1
	if k%2 == 1 {
		sign = -1
	}
	return sign, (k + 1) / 2
}

// loop returns a loop from a node to itself beyond its right side in the
// layout orientation, as the reference draws it: right of the node with
// rankdir=TB, above it with rankdir=LR. It leaves a third of the node's
// half height off the center, reaches nodesep beyond the node, and comes
// back on the other side. A loop around others reaches inner further.
func (r *router) loop(v int, inner float64) []Point {
	n := &r.nodes[r.g.nodes[v].real]
	hw, hh := half(n.W), half(n.H)
	if r.lr {
		hw, hh = hh, hw
	}
	dy := hh / 3
	reach := hw + r.nodesep + inner
	local := []Point{
		{0, dy}, {hw + half(r.nodesep), dy}, {reach, float64(dy * 0.75)}, {reach, 0},
		{reach, float64(-dy * 0.75)}, {hw + half(r.nodesep), -dy}, {0, -dy},
	}
	c := r.pos[v]
	out := make([]Point, len(local))
	for i, q := range local {
		out[i] = add(c, q)
	}
	return out
}
