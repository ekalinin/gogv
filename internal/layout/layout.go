// SPDX-License-Identifier: MIT
// Provenance: paper Gansner, Koutsofios, North, Vo 1993, A Technique for Drawing Directed Graphs
// Provenance: observed dot 16.1.0

package layout

import (
	"context"
	"math"
	"slices"
	"strconv"

	"github.com/ekalinin/gogv/internal/attrs"
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
	// Style is the supported style of the node as it is written: "solid",
	// or a list of "solid", "filled" and "rounded".
	Style string
	// Rounded is set for a rectangle with rounded corners.
	Rounded bool
	// Filled is set for a node with style=filled. FillColor is its
	// fillcolor, or "" when it has none.
	Filled    bool
	FillColor string
	// XLabel is the node's external label, centered on XLabelPos; it has
	// no lines and no table when the node has none.
	XLabel    Text
	XLabelPos Point
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
	// Style is the supported style of the edge as it is written: "solid",
	// or a list of "solid", "invis" and "dotted"; Dotted is set when it
	// holds "dotted". PenWidth is the width of its pen, in points.
	Style    string
	Dotted   bool
	PenWidth float64
	// Head and Tail are the arrowheads at the ends: Shape is "" when the
	// end has none; Tip is where the arrow points, Base where the curve
	// ends.
	Head, Tail Arrow
	// Label is the edge's label, centered on LabelPos; it has no lines
	// and no table when the edge has none. LabelText is its text after the
	// escapes that name objects are replaced.
	Label     Text
	LabelPos  Point
	LabelText string
	// XLabel, HeadLabel and TailLabel are the external labels, centered
	// on their positions; they have no lines and no table when the edge
	// has none.
	XLabel, HeadLabel, TailLabel          Text
	XLabelPos, HeadLabelPos, TailLabelPos Point
	// tailEnd and headEnd are where the edge meets its nodes, and middle
	// is halfway along it, in the layout orientation until the end of the
	// layout.
	tailEnd, headEnd, middle Point
}

// Fill returns the color that fills the node, as the reference chooses it
// and plain writes it: its fillcolor, or else its color when set, or else
// lightgrey.
func (n *NodeResult) Fill() string {
	switch {
	case n.FillColor != "":
		return n.FillColor
	case n.ColorSet:
		return n.Color
	}
	return "lightgrey"
}

// Arrow is an arrowhead.
type Arrow struct {
	Shape string
	Size  float64
	// PenWidth is the width of the pen of the edge.
	PenWidth  float64
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

// Observed on the reference: in a graph with edge labels, how far a label
// node reaches beyond the edge and the label, and the width of the other
// virtual nodes on the ranks of labels; and the room loops keep.
const (
	labelPad         = 2.5
	labelRankVirtual = 7
	// loopGap is the room that loops keep beyond their reach, whatever the
	// size of the node and nodesep.
	loopGap = 18
	// flatLabelLift is how far the label of a flat edge is above it.
	flatLabelLift = 3
)

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
		text := labelOf(nodeValue(g, n, "label"), graph.LabelDefault, obj, nodeValue(g, n, "fontname"), nodeValue(g, n, "fontcolor"), nodeValue(g, n, "fontsize"))
		if shape == "point" {
			text.Lines, text.Table = nil, nil
		}
		w, h := nodeSize(shape, text, nodeValue(g, n, "width"), nodeValue(g, n, "height"))
		color := nodeColor(nodeValue(g, n, "color"))
		style := supported(graph.NodeKind, "style", nodeValue(g, n, "style"), "solid")
		lt := label.Expand(labelText, obj)
		if text.Table != nil {
			// plain writes a table label as it is written.
			lt = nodeValue(g, n, "label").S
		}
		colorSet := color == nodeValue(g, n, "color").S
		fill := supported(graph.NodeKind, "fillcolor", nodeValue(g, n, "fillcolor"), "")
		styles := attrs.Styles(style)
		res.Nodes = append(res.Nodes, NodeResult{
			Node: n, W: w, H: h, Shape: shape, Label: text, LabelText: lt,
			Color: color, ColorSet: colorSet,
			Style:     style,
			Rounded:   isRectangular(shape) && shape != "plaintext" && shape != "box3d" && slices.Contains(styles, "rounded"),
			Filled:    slices.Contains(styles, "filled"),
			FillColor: fill,
		})
		lw[i], lh[i] = w, h
		if p.lr {
			lw[i], lh[i] = h, w
		}
	}

	// Edge labels. As in the reference, a graph with an edge label gets a
	// rank between every two of its ranks, half ranksep apart, for the
	// labels to lie on.
	elabels := make([]Text, len(g.Edges()))
	own := make([]bool, len(g.Edges()))
	doubled := false
	for i, e := range g.Edges() {
		obj := label.Object{Graph: g.Name, Edge: true, Tail: e.Tail.Name, Head: e.Head.Name, DirectedEdgeName: g.Directed()}
		elabels[i] = labelOf(edgeValue(g, e, "label"), "", obj, edgeValue(g, e, "fontname"), edgeValue(g, e, "fontcolor"), edgeValue(g, e, "fontsize"))
		own[i] = elabels[i].Present()
		doubled = doubled || own[i]
	}
	ranksep := p.ranksep
	if doubled {
		ranksep = half(ranksep)
	}

	var rankEdges []rankEdge
	ends := make([][2]int, 0, len(g.Edges()))
	// Each edge weighs its weight in the ranking and in the positioning, as
	// the weight of Gansner et al.
	weights := make([]int, 0, len(g.Edges()))
	for _, e := range g.Edges() {
		t, h := index[e.Tail], index[e.Head]
		ends = append(ends, [2]int{t, h})
		weight, err := strconv.Atoi(supported(graph.EdgeKind, "weight", edgeValue(g, e, "weight"), "1"))
		if err != nil {
			weight = 1
		}
		weights = append(weights, weight)
		if t == h || !boolean(edgeValue(g, e, "constraint"), true) {
			continue
		}
		minlen := int(number(edgeValue(g, e, "minlen"), 1, 0))
		if doubled {
			minlen *= 2
		}
		rankEdges = append(rankEdges, rankEdge{tail: t, head: h, minlen: minlen, weight: weight})
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

	// Edges with ports get chains of their own, as labeled edges do, whose
	// ends carry the ports.
	ports := edgePorts(g, res.Nodes, ends, p.lr)
	ownChain := make([]bool, len(own))
	for i := range own {
		ownChain[i] = own[i] || hasPort(ports[i])
	}
	rg, chains := buildRanked(ranks, ends, ownChain, weights)
	rg.setPorts(chains, ends, ports, lw, p.lr)
	// The label of an edge belongs to the virtual node halfway along its
	// chain.
	labelNode := make([]int, len(g.Edges()))
	labelEdge := map[int]int{}
	for i := range labelNode {
		labelNode[i] = -1
		if own[i] && len(chains[i]) >= 3 {
			labelNode[i] = chains[i][len(chains[i])/2]
			labelEdge[labelNode[i]] = i
		}
	}
	rg.joins = tree.joins()
	if tree.drawn() {
		err = rg.orderClusters(ctx, tree, 24, p.lr)
	} else {
		err = rg.order(ctx, 24, p.lr)
	}
	if err != nil {
		return nil, err
	}
	// How far each ranked node reaches left and right of its x, and its
	// height, in the layout orientation.
	left := make([]float64, len(rg.nodes))
	right := make([]float64, len(rg.nodes))
	height := make([]float64, len(rg.nodes))
	labelSize := func(t Text) (along, across float64) {
		// The extent of a label along the ranks' order and across it.
		if p.lr {
			return t.Height, t.Width
		}
		return t.Width, t.Height
	}
	for v, rn := range rg.nodes {
		i, isLabel := labelEdge[v]
		switch {
		case rn.real >= 0:
			left[v], right[v], height[v] = half(lw[rn.real]), half(lw[rn.real]), lh[rn.real]
		case isLabel:
			// Observed on the reference: on a rank of labels, the edge
			// passes half nodesep plus 2.5 points right of the node's left
			// side, and the label reaches its width plus 2.5 points minus
			// half nodesep beyond; on a rank of nodes, nodesep and the
			// label's width.
			along, across := labelSize(elabels[i])
			if rn.rank%2 == 1 {
				left[v] = half(p.nodesep) + labelPad
				right[v] = along + labelPad - half(p.nodesep)
			} else {
				left[v], right[v] = p.nodesep, along
			}
			height[v] = across
		case doubled && rn.rank%2 == 1:
			// A virtual node on a rank of labels is 7 points wide.
			left[v], right[v], height[v] = labelRankVirtual/2, labelRankVirtual/2, 1
		default:
			// A virtual node keeps nodesep free for each edge it carries,
			// plus 2 points; an empty rank is 1 point high.
			w := float64(float64(rn.count)*p.nodesep) + 2
			left[v], right[v], height[v] = half(w), half(w), 1
		}
	}
	// Loops take room right of their node, as the reference reserves it,
	// from its neighbors and from the border of its cluster: loopGap, the
	// widths of their labels, and nodesep for each loop but the first.
	loopRoom := make([]float64, len(nodes))
	loopLabels := make([]float64, len(ends))
	for i, e := range ends {
		if e[0] != e[1] {
			continue
		}
		if own[i] {
			loopLabels[i], _ = labelSize(elabels[i])
		}
		if loopRoom[e[0]] == 0 {
			loopRoom[e[0]] = loopGap - p.nodesep
		}
		loopRoom[e[0]] += p.nodesep + loopLabels[i]
	}
	for v := range nodes {
		right[v] += loopRoom[v]
	}
	// The labels of loops are as high as their rank.
	for i, e := range ends {
		if e[0] == e[1] && own[i] {
			_, across := labelSize(elabels[i])
			height[e[0]] = max(height[e[0]], across)
		}
	}
	// A labeled flat edge between neighbors keeps them apart by its label.
	for i, e := range ends {
		if !own[i] || chains[i] != nil || e[0] == e[1] {
			continue
		}
		l, r := e[0], e[1]
		if rg.nodes[l].order > rg.nodes[r].order {
			l, r = r, l
		}
		if rg.nodes[r].order-rg.nodes[l].order == 1 {
			along, _ := labelSize(elabels[i])
			if rg.labelSep == nil {
				rg.labelSep = map[[2]int]float64{}
			}
			rg.labelSep[[2]int{l, r}] = max(rg.labelSep[[2]int{l, r}], along)
		}
	}
	var extra func(next int) ([]nsEdge, int)
	if tree.drawn() {
		tree.readLabels()
		extra = func(next int) ([]nsEdge, int) { return tree.xEdges(rg, left, right, next, p.lr) }
	}
	xs, err := rg.placeX(ctx, left, right, p.nodesep, p.search, extra)
	if err != nil {
		return nil, err
	}
	var gaps []float64
	if tree.drawn() {
		gaps = tree.yGaps(rg, height, p.lr)
	}
	ys := rg.placeY(height, ranksep, gaps)
	if tree.drawn() {
		tree.boxes(xs, ys, p.lr)
	}

	// The drawing in the layout orientation: positions of the ranked nodes.
	pos := make([]Point, len(rg.nodes))
	for v, rn := range rg.nodes {
		pos[v] = Point{xs[v], ys[rn.rank]}
	}
	rt := &router{g: rg, pos: pos, left: left, right: right, ys: ys, nodesep: p.nodesep, ranksep: ranksep, lr: p.lr, nodes: res.Nodes, labelAlong: map[int]float64{}, ports: ports}
	for v, i := range labelEdge {
		rt.labelAlong[v], _ = labelSize(elabels[i])
	}
	rt.loopLabels = loopLabels
	// The sametail groups: edges with the same tail and sametail value.
	rt.sameTail = map[int][]int{}
	sameTails := map[[2]string][]int{}
	for i, e := range g.Edges() {
		if v := edgeValue(g, e, "sametail"); v.S != "" && ends[i][0] != ends[i][1] {
			k := [2]string{e.Tail.Name, v.S}
			sameTails[k] = append(sameTails[k], i)
		}
	}
	for _, group := range sameTails {
		for _, i := range group {
			rt.sameTail[i] = group
		}
	}
	// The labeled flat edges between neighbors, numbered by pair.
	rt.flatLabels, rt.flatAcross = map[int]int{}, map[int]float64{}
	flatCount := map[[2]int]int{}
	for i, e := range ends {
		if !own[i] || chains[i] != nil || e[0] == e[1] {
			continue
		}
		l, r := e[0], e[1]
		if rg.nodes[l].order > rg.nodes[r].order {
			l, r = r, l
		}
		if rg.nodes[r].order-rg.nodes[l].order == 1 {
			rt.flatLabels[i] = flatCount[[2]int{l, r}]
			flatCount[[2]int{l, r}]++
			_, rt.flatAcross[i] = labelSize(elabels[i])
		}
	}
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

	// Edge labels, in the layout orientation. As observed on the reference,
	// a label lies on the rank of its node, its left side on the rightmost
	// point of the edge across the label's height.
	labelPos := make([]Point, len(res.Edges))
	// The labels of flat edges stay out of the drawing's box, as in the
	// reference.
	outside := make([]bool, len(res.Edges))
	flatLabels := map[[2]int]int{}
	for i, v := range labelNode {
		res.Edges[i].Label = elabels[i]
		if lv := edgeValue(g, g.Edges()[i], "label"); !lv.HTML {
			e := g.Edges()[i]
			res.Edges[i].LabelText = label.Expand(lv.S, label.Object{Graph: g.Name, Edge: true, Tail: e.Tail.Name, Head: e.Head.Name, DirectedEdgeName: g.Directed()})
		}
		if !own[i] {
			continue
		}
		along, across := labelSize(elabels[i])
		t, h := ends[i][0], ends[i][1]
		switch {
		case v >= 0:
			y := pos[v].Y
			x := rightmostAt(res.Edges[i].Points, y-half(across), y+half(across), pos[v].X)
			labelPos[i] = Point{x + half(along), y}
		case t == h:
			// Beyond the far side of the loop.
			x := rightmostAt(res.Edges[i].Points, math.Inf(-1), math.Inf(1), pos[t].X+right[t])
			labelPos[i] = Point{x + half(along), pos[t].Y}
		default:
			// A flat edge: centered between its ends, above the edge; the
			// label of a second flat edge between them goes below it.
			l, r := t, h
			if rg.nodes[l].order > rg.nodes[r].order {
				l, r = r, l
			}
			x := half(pos[l].X + right[l] + pos[r].X - left[r])
			// The first label is above the edge; the next ones alternate
			// below and above, a label's height plus 6 points further at
			// each level.
			dy := half(across) + flatLabelLift
			if k := flatLabels[[2]int{l, r}]; k > 0 {
				sign, level := flatLevel(k)
				if sign < 0 {
					level--
				}
				dy = sign * (dy + float64(float64(level)*(across+6)))
			}
			flatLabels[[2]int{l, r}]++
			labelPos[i] = Point{x, pos[l].Y + dy}
			outside[i] = true
		}
	}

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
	for i, e := range res.Edges {
		if e.Label.Present() && !outside[i] {
			along, across := labelSize(e.Label)
			d := Point{half(along), half(across)}
			grow(sub(labelPos[i], d), add(labelPos[i], d))
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
		e.LabelPos = final(labelPos[i])
		e.tailEnd, e.headEnd, e.middle = final(e.tailEnd), final(e.headEnd), final(e.middle)
	}
	for _, cl := range tree.clusters[1:] {
		a, b := final(cl.box[0]), final(cl.box[1])
		cr := ClusterResult{
			Graph: cl.g, Color: cl.color, Label: cl.label,
			X0: min(a.X, b.X), Y0: min(a.Y, b.Y), X1: max(a.X, b.X), Y1: max(a.Y, b.Y),
		}
		if cl.label.Present() {
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
	placeExternalLabels(g, res)
	placeGraphLabel(g, res)
	return res, nil
}

// rightmostAt returns the largest x of a piecewise cubic curve between
// heights lo and hi, or def when the curve does not reach them.
func rightmostAt(pts []Point, lo, hi, def float64) float64 {
	x, found := math.Inf(-1), false
	for _, seg := range segments(pts) {
		const steps = 64
		for k := 0; k <= steps; k++ {
			q := cubic(seg, float64(k)/steps)
			if q.Y >= lo && q.Y <= hi {
				x, found = max(x, q.X), true
			}
		}
	}
	if !found {
		return def
	}
	return x
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
		e.LabelPos = add(e.LabelPos, d)
		e.XLabelPos, e.HeadLabelPos, e.TailLabelPos = add(e.XLabelPos, d), add(e.HeadLabelPos, d), add(e.TailLabelPos, d)
	}
	for i := range res.Nodes {
		res.Nodes[i].XLabelPos = add(res.Nodes[i].XLabelPos, d)
	}
}
