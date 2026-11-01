// SPDX-License-Identifier: MIT
// Provenance: doc https://graphviz.org/docs/outputs/canon/
// Provenance: doc https://graphviz.org/docs/attrs/pos/
// Provenance: observed dot 16.1.0

package output

import (
	"strconv"
	"strings"

	"github.com/ekalinin/gogv/internal/graph"
	"github.com/ekalinin/gogv/internal/layout"
)

// Dot writes the laid-out graph back in DOT with the attributes of the
// layout, as the reference's -Tdot and -Tgv do: the canon output of the
// graph once the layout has set them.
func Dot(res *layout.Result) []byte {
	setLayoutAttrs(res)
	return writeDOT(res.Graph, true)
}

// laidOutGraphAttr are the graph attributes that the layout and the drawing
// set: a graph has them only where they are set, not inherited.
var laidOutGraphAttr = map[string]bool{"bb": true, "lheight": true, "lp": true, "lwidth": true, "_draw_": true, "_ldraw_": true, "xdotversion": true}

// layoutPoint writes a point as "x,y", with the numbers of plain output, as
// the reference writes them in its attributes.
func layoutPoint(p layout.Point) string { return plainNumber(p.X) + "," + plainNumber(p.Y) }

// inches writes a size in inches, as the reference writes the sizes of
// nodes.
func inches(points float64) graph.Value { return graph.Str(plainNumber(points / pointsPerInch)) }

// labelInches writes the size of a label in inches, with at most two
// decimals, as the reference writes lwidth and lheight.
func labelInches(points float64) graph.Value {
	s := strconv.FormatFloat(points/pointsPerInch, 'f', 2, 64)
	s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	return graph.Str(s)
}

const pointsPerInch = 72

// setLayoutAttrs sets the attributes of the layout on the graph of res, its
// clusters, nodes and edges.
func setLayoutAttrs(res *layout.Result) {
	g := res.Graph
	setLabel := func(sg *graph.Graph, t layout.Text, at layout.Point) {
		sg.SetAttr("lheight", labelInches(t.Height))
		sg.SetAttr("lp", graph.Str(layoutPoint(at)))
		sg.SetAttr("lwidth", labelInches(t.Width))
	}
	g.SetAttr("bb", graph.Str("0,0,"+plainNumber(res.Width)+","+plainNumber(res.Height)))
	if res.Label.Present() {
		setLabel(g, res.Label, res.LabelPos)
	}
	for _, c := range res.Clusters {
		c.Graph.SetAttr("bb", graph.Str(layoutPoint(layout.Point{X: c.X0, Y: c.Y0})+","+layoutPoint(layout.Point{X: c.X1, Y: c.Y1})))
		if c.Label.Present() {
			setLabel(c.Graph, c.Label, c.LabelPos)
		}
	}
	for _, n := range res.Nodes {
		g.SetNodeAttr(n.Node, "height", inches(n.H))
		g.SetNodeAttr(n.Node, "pos", graph.Str(layoutPoint(layout.Point{X: n.X, Y: n.Y})))
		g.SetNodeAttr(n.Node, "width", inches(n.W))
		if n.XLabel.Present() {
			g.SetNodeAttr(n.Node, "xlp", graph.Str(layoutPoint(n.XLabelPos)))
		}
	}
	for _, e := range res.Edges {
		var parts []string
		if e.Tail.Shape != "" {
			parts = append(parts, "s,"+layoutPoint(e.Tail.Tip))
		}
		if e.Head.Shape != "" {
			parts = append(parts, "e,"+layoutPoint(e.Head.Tip))
		}
		for _, p := range e.Points {
			parts = append(parts, layoutPoint(p))
		}
		if len(e.Points) > 0 {
			g.SetEdgeAttr(e.Edge, "pos", graph.Str(strings.Join(parts, " ")))
		}
		for _, l := range []struct {
			name string
			t    layout.Text
			at   layout.Point
		}{{"lp", e.Label, e.LabelPos}, {"xlp", e.XLabel, e.XLabelPos}, {"head_lp", e.HeadLabel, e.HeadLabelPos}, {"tail_lp", e.TailLabel, e.TailLabelPos}} {
			if l.t.Present() {
				g.SetEdgeAttr(e.Edge, l.name, graph.Str(layoutPoint(l.at)))
			}
		}
	}
}
