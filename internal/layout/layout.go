// SPDX-License-Identifier: MIT
// Provenance: paper Gansner, Koutsofios, North, Vo 1993, A Technique for Drawing Directed Graphs
// Provenance: observed dot 16.1.0

package layout

import (
	"context"
	"math"

	"github.com/ekalinin/gogv/internal/graph"
	"github.com/ekalinin/gogv/internal/label"
)

// Point is a point in points, y growing upward.
type Point struct{ X, Y float64 }

// NodeResult is a laid-out node.
type NodeResult struct {
	Node  *graph.Node
	X, Y  float64 // center
	W, H  float64
	Shape string
	Label Text
	// LabelText is the label after the escapes that name objects are
	// replaced, before it is split into lines.
	LabelText string
	Color     string
	// ColorSet is set when the color attribute gives the color.
	ColorSet bool
	// Style is the supported style of the node: "solid" or "rounded".
	Style string
	// Rounded is set for a rectangle with rounded corners.
	Rounded bool
}

// EdgeResult is a laid-out edge.
type EdgeResult struct {
	Edge *graph.Edge
	// Points are the control points of the edge's cubic Bezier curve: the
	// start point, then three points per segment. Empty for invisible
	// edges.
	Points    []Point
	Invisible bool
	Color     string
	// Head and Tail are the arrowheads at the ends: Shape is "" when the
	// end has none; Tip is where the arrow points, Base where the curve
	// ends.
	Head, Tail Arrow
}

// Arrow is an arrowhead.
type Arrow struct {
	Shape     string
	Size      float64
	Base, Tip Point
}

// Result is a laid-out graph.
type Result struct {
	Graph *graph.Graph
	// Width and Height are the size of the drawing, in points; it spans
	// from (0, 0) to (Width, Height).
	Width, Height float64
	Nodes         []NodeResult
	Edges         []EdgeResult
	// Label is the graph label, centered on LabelPos; Label.Lines is empty
	// when there is none.
	Label    Text
	LabelPos Point
	// Clusters are the drawn clusters, outer clusters before the clusters
	// nested in them.
	Clusters []ClusterResult
}

// ClusterResult is a laid-out cluster: its box from (X0, Y0) to (X1, Y1),
// and its label, centered on LabelPos.
type ClusterResult struct {
	Graph          *graph.Graph
	X0, Y0, X1, Y1 float64
	Color          string
	Label          Text
	LabelPos       Point
}

// params are the graph attributes the layout reads.
type params struct {
	nodesep, ranksep float64
	lr               bool
	polyline         bool
	search           int
}

func readParams(g *graph.Graph) params {
	return params{
		nodesep:  number(graphValue(g, "nodesep"), 0.25, 0.02) * pointsPerInch,
		ranksep:  number(graphValue(g, "ranksep"), 0.5, 0.02) * pointsPerInch,
		lr:       supported(graph.GraphKind, "rankdir", graphValue(g, "rankdir"), "TB") == "LR",
		polyline: supported(graph.GraphKind, "splines", graphValue(g, "splines"), "spline") == "polyline",
		search:   int(number(graphValue(g, "searchsize"), 30, 0)),
	}
}

// Layout lays out a graph.
func Layout(ctx context.Context, g *graph.Graph) (*Result, error) {
	g = g.Root()
	p := readParams(g)
	res := &Result{Graph: g}

	nodes := g.Nodes()
	index := make(map[*graph.Node]int, len(nodes))
	lw := make([]float64, len(nodes)) // sizes in the layout orientation
	lh := make([]float64, len(nodes))
	for i, n := range nodes {
		index[n] = i
		shape := nodeShape(nodeValue(g, n, "shape"))
		obj := label.Object{Graph: g.Name, Node: n.Name}
		labelText := nodeLabelText(nodeValue(g, n, "label"))
		text := makeText(labelText, obj, nodeValue(g, n, "fontname"), nodeValue(g, n, "fontcolor"), nodeValue(g, n, "fontsize"))
		if shape == "point" {
			text.Lines = nil
		}
		w, h := nodeSize(shape, text, nodeValue(g, n, "width"), nodeValue(g, n, "height"))
		color := nodeColor(nodeValue(g, n, "color"))
		style := supported(graph.NodeKind, "style", nodeValue(g, n, "style"), "solid")
		res.Nodes = append(res.Nodes, NodeResult{
			Node: n, W: w, H: h, Shape: shape, Label: text, LabelText: label.Expand(labelText, obj),
			Color: color, ColorSet: color == nodeValue(g, n, "color").S,
			Style:   style,
			Rounded: isRectangular(shape) && shape != "plaintext" && style == "rounded",
		})
		lw[i], lh[i] = w, h
		if p.lr {
			lw[i], lh[i] = h, w
		}
	}

	var rankEdges []rankEdge
	ends := make([][2]int, 0, len(g.Edges()))
	for _, e := range g.Edges() {
		t, h := index[e.Tail], index[e.Head]
		ends = append(ends, [2]int{t, h})
		if t == h || !boolean(edgeValue(g, e, "constraint"), true) {
			continue
		}
		minlen := int(number(edgeValue(g, e, "minlen"), 1, 0))
		rankEdges = append(rankEdges, rankEdge{tail: t, head: h, minlen: minlen, weight: 1})
	}
	tree := buildClusterTree(g, index)
	var ranks []int
	var err error
	if tree.drawn() || tree.hasRankSets() {
		ranks, err = tree.rank(ctx, rankEdges, p.search)
	} else {
		ranks, err = assignRanks(ctx, len(nodes), rankEdges, p.search)
	}
	if err != nil {
		return nil, err
	}

	rg, chains := buildRanked(ranks, ends)
	rg.joins = tree.joins()
	if tree.drawn() {
		err = rg.orderClusters(ctx, tree, 24, p.lr)
	} else {
		err = rg.order(ctx, 24, p.lr)
	}
	if err != nil {
		return nil, err
	}
	width := make([]float64, len(rg.nodes))
	height := make([]float64, len(rg.nodes))
	for v, rn := range rg.nodes {
		if rn.real >= 0 {
			width[v], height[v] = lw[rn.real], lh[rn.real]
		} else {
			// A virtual node keeps nodesep free for each edge it carries,
			// plus 2 points; an empty rank is 1 point high.
			width[v] = float64(float64(rn.count)*p.nodesep) + 2
			height[v] = 1
		}
	}
	var extra func(next int) ([]nsEdge, int)
	if tree.drawn() {
		tree.readLabels()
		extra = func(next int) ([]nsEdge, int) { return tree.xEdges(rg, width, next, p.lr) }
	}
	xs, err := rg.placeX(ctx, width, p.nodesep, p.search, extra)
	if err != nil {
		return nil, err
	}
	var gaps []float64
	if tree.drawn() {
		gaps = tree.yGaps(rg, height, p.lr)
	}
	ys := rg.placeY(height, p.ranksep, gaps)
	if tree.drawn() {
		tree.boxes(xs, ys, p.lr)
	}

	// The drawing in the layout orientation: positions of the ranked nodes.
	pos := make([]Point, len(rg.nodes))
	for v, rn := range rg.nodes {
		pos[v] = Point{xs[v], ys[rn.rank]}
	}
	rt := &router{g: rg, pos: pos, width: width, ys: ys, nodesep: p.nodesep, lr: p.lr, nodes: res.Nodes}
	if tree.drawn() {
		rt.tree = tree
	}
	rt.bandUp = make([]float64, len(rg.ranks))
	rt.bandDn = make([]float64, len(rg.ranks))
	for v, rn := range rg.nodes {
		rt.bandUp[rn.rank] = max(rt.bandUp[rn.rank], height[v]/2)
		rt.bandDn[rn.rank] = max(rt.bandDn[rn.rank], height[v]/2)
	}
	res.Edges = rt.routeEdges(g, ends, chains, p.polyline)

	// The bounding box of nodes, edges and arrowheads, in the layout
	// orientation.
	minP := Point{math.Inf(1), math.Inf(1)}
	maxP := Point{math.Inf(-1), math.Inf(-1)}
	grow := func(lo, hi Point) {
		minP = Point{min(minP.X, lo.X), min(minP.Y, lo.Y)}
		maxP = Point{max(maxP.X, hi.X), max(maxP.Y, hi.Y)}
	}
	for i := range nodes {
		c := pos[i]
		grow(Point{c.X - half(lw[i]), c.Y - half(lh[i])}, Point{c.X + half(lw[i]), c.Y + half(lh[i])})
	}
	for _, e := range res.Edges {
		for _, seg := range segments(e.Points) {
			grow(cubicExtent(seg))
		}
		for _, a := range []Arrow{e.Head, e.Tail} {
			if a.Shape != "" {
				for _, q := range ArrowPolygon(a) {
					grow(q, q)
				}
			}
		}
	}
	for _, cl := range tree.clusters[1:] {
		// The drawing keeps clusterMargin around the boxes.
		m := Point{clusterMargin, clusterMargin}
		grow(sub(cl.box[0], m), add(cl.box[1], m))
	}
	if len(nodes) == 0 {
		minP, maxP = Point{}, Point{}
	}
	layoutW, layoutH := maxP.X-minP.X, maxP.Y-minP.Y
	final := func(q Point) Point {
		q = sub(q, minP)
		if p.lr {
			return Point{layoutH - q.Y, q.X}
		}
		return q
	}
	for i := range nodes {
		c := final(pos[i])
		res.Nodes[i].X, res.Nodes[i].Y = c.X, c.Y
	}
	for i := range res.Edges {
		e := &res.Edges[i]
		for j := range e.Points {
			e.Points[j] = final(e.Points[j])
		}
		for _, a := range []*Arrow{&e.Head, &e.Tail} {
			if a.Shape != "" {
				a.Base, a.Tip = final(a.Base), final(a.Tip)
			}
		}
	}
	for _, cl := range tree.clusters[1:] {
		a, b := final(cl.box[0]), final(cl.box[1])
		cr := ClusterResult{
			Graph: cl.g, Color: cl.color, Label: cl.label,
			X0: min(a.X, b.X), Y0: min(a.Y, b.Y), X1: max(a.X, b.X), Y1: max(a.Y, b.Y),
		}
		if len(cl.label.Lines) > 0 {
			cr.LabelPos = Point{half(cr.X0 + cr.X1), cr.Y1 - half(cl.labelH)}
			switch cl.labelJust {
			case "l":
				cr.LabelPos.X = cr.X0 + half(cl.labelW)
			case "r":
				cr.LabelPos.X = cr.X1 - half(cl.labelW)
			}
		}
		res.Clusters = append(res.Clusters, cr)
	}
	res.Width, res.Height = layoutW, layoutH
	if p.lr {
		res.Width, res.Height = layoutH, layoutW
	}
	placeGraphLabel(g, res)
	return res, nil
}

// placeGraphLabel puts the graph label below the drawing, centered or at
// the side that labeljust names; the drawing grows by the label's height and
// margins, and widens when the label is wider.
func placeGraphLabel(g *graph.Graph, res *Result) {
	v := graphValue(g, "label")
	if v.S == "" || v.HTML {
		return
	}
	text := makeText(v.S, label.Object{Graph: g.Name}, graphValue(g, "fontname"), graphValue(g, "fontcolor"), graphValue(g, "fontsize"))
	if len(text.Lines) == 0 {
		return
	}
	lw, lh := text.Width+labelMarginX, text.Height+labelMarginY
	shift := Point{0, lh}
	if lw > res.Width {
		shift.X = (lw - res.Width) / 2
		res.Width = lw
	}
	res.translate(shift)
	res.Height += lh
	res.Label = text
	res.LabelPos = Point{res.Width / 2, lh / 2}
	switch supported(graph.GraphKind, "labeljust", graphValue(g, "labeljust"), "c") {
	case "l":
		res.LabelPos.X = half(lw)
	case "r":
		res.LabelPos.X = res.Width - half(lw)
	}
}

// translate moves every node, edge and cluster of the drawing by d.
func (res *Result) translate(d Point) {
	for i := range res.Clusters {
		c := &res.Clusters[i]
		c.X0, c.Y0, c.X1, c.Y1 = c.X0+d.X, c.Y0+d.Y, c.X1+d.X, c.Y1+d.Y
		c.LabelPos = add(c.LabelPos, d)
	}
	for i := range res.Nodes {
		res.Nodes[i].X += d.X
		res.Nodes[i].Y += d.Y
	}
	for i := range res.Edges {
		e := &res.Edges[i]
		for j := range e.Points {
			e.Points[j] = add(e.Points[j], d)
		}
		for _, a := range []*Arrow{&e.Head, &e.Tail} {
			if a.Shape != "" {
				a.Base, a.Tip = add(a.Base, d), add(a.Tip, d)
			}
		}
	}
}
