// SPDX-License-Identifier: MIT
// Provenance: doc https://graphviz.org/docs/outputs/canon/
// Provenance: doc https://graphviz.org/docs/attrs/_draw_/
// Provenance: observed dot 16.1.0

package output

import (
	"strconv"
	"strings"

	"github.com/ekalinin/gogv/internal/graph"
	"github.com/ekalinin/gogv/internal/label"
	"github.com/ekalinin/gogv/internal/layout"
)

// Xdot writes the laid-out graph as the reference's -Txdot, -Txdot1.2 and
// -Txdot1.4 do: the dot output with the drawing of each object as xdot
// drawing operations, and version as xdotversion.
func Xdot(res *layout.Result, version string) []byte {
	setLayoutAttrs(res)
	setDrawAttrs(res, version)
	return writeDOT(res.Graph, true)
}

// xdraw builds the drawing operations of one attribute, in the order of
// the svg drawing.
type xdraw struct {
	b strings.Builder
	// width is the pen width that the operations set last.
	width int
}

func newXdraw() *xdraw { return &xdraw{width: 1} }

func (d *xdraw) String() string { return d.b.String() }

// op writes an operation and its fields.
func (d *xdraw) op(fields ...string) {
	for _, f := range fields {
		d.b.WriteString(f)
		d.b.WriteByte(' ')
	}
}

// text writes a string field: its length in bytes, a dash and the string,
// its backslashes doubled as the reference writes them.
func text(s string) string { return strconv.Itoa(len(s)) + " -" + strings.ReplaceAll(s, `\`, `\\`) }

// xdotColor writes a color as the drawing operations do.
func xdotColor(c string) string {
	if c == "" || c == "black" {
		return "#000000"
	}
	return strings.ToLower(c)
}

func (d *xdraw) pen(c string)  { d.op("c", text(xdotColor(c))) }
func (d *xdraw) fill(c string) { d.op("C", text(xdotColor(c))) }
func (d *xdraw) style(s string) {
	d.op("S", text(s))
}

// points writes an operation on a list of points.
func (d *xdraw) points(op string, pts []layout.Point) {
	fields := []string{op, strconv.Itoa(len(pts))}
	for _, p := range pts {
		fields = append(fields, svgNumber(p.X), svgNumber(p.Y))
	}
	d.op(fields...)
}

// box writes the polygon of a box from its lower left corner upward and
// around, as svg draws it.
func (d *xdraw) box(op string, x0, y0, x1, y1 float64) {
	d.points(op, []layout.Point{{X: x0, Y: y0}, {X: x0, Y: y1}, {X: x1, Y: y1}, {X: x1, Y: y0}})
}

// label writes a label centered on c: its lines, each with its font and
// color, or the polygons of its table.
func (d *xdraw) label(t layout.Text, c layout.Point, pen string) {
	if t.Table == nil {
		for _, l := range textLines(t, c) {
			just := "0"
			switch l.just {
			case label.Left:
				just = "-1"
			case label.Right:
				just = "1"
			}
			d.op("F", svgNumber(t.FontSize), text(t.FontName))
			d.pen(t.FontColor)
			d.op("T", svgNumber(l.x), svgNumber(l.base), just, svgNumber(t.Face.Width(l.text, t.FontSize)), text(l.text))
		}
		return
	}
	for _, bx := range tableBoxes(t, c, pen) {
		if bx.stroke == "" {
			d.style("solid")
			d.op("c", text("#fffffe00"))
			d.fill(bx.fill)
			d.box("P", bx.x0, bx.y0, bx.x1, bx.y1)
			continue
		}
		if bx.width != d.width {
			d.style("setlinewidth(" + strconv.Itoa(bx.width) + ")")
			d.width = bx.width
		}
		d.style("solid")
		d.pen(bx.stroke)
		d.box("p", bx.x0, bx.y0, bx.x1, bx.y1)
	}
}

// solid writes the solid style of an object that sets it explicitly, as the
// reference writes it before the object's outline.
func solid(d *xdraw, v graph.Value) {
	if v.S == "solid" {
		d.style("solid")
	}
}

// setDrawAttrs sets the drawing operations of the graph, its clusters, nodes
// and edges as attributes, and xdotversion on the graph.
func setDrawAttrs(res *layout.Result, version string) {
	g := res.Graph
	set := func(d *xdraw, assign func(string)) {
		if s := d.String(); s != "" {
			assign(s)
		}
	}

	bg := newXdraw()
	bg.op("c", text("#fffffe00"))
	bg.fill("#ffffff")
	bg.box("P", 0, 0, res.Width, res.Height)
	g.SetAttr("_draw_", graph.Str(bg.String()))
	if res.Label.Present() {
		d := newXdraw()
		d.label(res.Label, res.LabelPos, "black")
		g.SetAttr("_ldraw_", graph.Str(d.String()))
	}
	g.SetAttr("xdotversion", graph.Str(version))

	for _, c := range res.Clusters {
		d := newXdraw()
		v, _ := c.Graph.Attr("style")
		solid(d, v)
		d.pen(c.Color)
		d.box("p", c.X0, c.Y0, c.X1, c.Y1)
		c.Graph.SetAttr("_draw_", graph.Str(d.String()))
		if c.Label.Present() {
			d := newXdraw()
			d.label(c.Label, c.LabelPos, svgColor(c.Color))
			c.Graph.SetAttr("_ldraw_", graph.Str(d.String()))
		}
	}

	for i := range res.Nodes {
		n := &res.Nodes[i]
		d := newXdraw()
		v, _ := n.Node.Attr(g, "style")
		switch n.Shape {
		case "plaintext":
		case "ellipse", "oval", "circle":
			solid(d, v)
			d.pen(n.Color)
			d.op("e", svgNumber(n.X), svgNumber(n.Y), svgNumber(n.W/2), svgNumber(n.H/2))
		case "point":
			solid(d, v)
			d.pen(n.Color)
			d.fill(n.Color)
			d.op("E", svgNumber(n.X), svgNumber(n.Y), svgNumber(n.W/2), svgNumber(n.H/2))
		default:
			solid(d, v)
			d.pen(n.Color)
			if n.Rounded {
				var pts []layout.Point
				for _, p := range layout.RoundedOutline(n.W, n.H) {
					pts = append(pts, layout.Point{X: n.X + p.X, Y: n.Y + p.Y})
				}
				d.points("B", pts)
				break
			}
			var pts []layout.Point
			for _, p := range layout.PolygonVertices(n.Shape, n.W, n.H) {
				pts = append(pts, layout.Point{X: n.X + p.X, Y: n.Y + p.Y})
			}
			d.points("p", pts)
		}
		set(d, func(s string) { g.SetNodeAttr(n.Node, "_draw_", graph.Str(s)) })
		l := newXdraw()
		if n.Shape != "point" {
			l.label(n.Label, layout.Point{X: n.X, Y: n.Y}, svgColor(n.Color))
		}
		if n.XLabel.Present() {
			l.label(n.XLabel, n.XLabelPos, svgColor(n.Color))
		}
		set(l, func(s string) { g.SetNodeAttr(n.Node, "_ldraw_", graph.Str(s)) })
	}

	for i := range res.Edges {
		e := &res.Edges[i]
		if e.Invisible {
			continue
		}
		setEdge := func(name string, d *xdraw) {
			set(d, func(s string) { g.SetEdgeAttr(e.Edge, name, graph.Str(s)) })
		}
		d := newXdraw()
		if len(e.Points) > 0 {
			v, _ := e.Edge.Attr(g, "style")
			solid(d, v)
			d.pen(e.Color)
			d.points("B", e.Points)
		}
		setEdge("_draw_", d)
		for _, a := range []struct {
			name  string
			arrow layout.Arrow
		}{{"_tdraw_", e.Tail}, {"_hdraw_", e.Head}} {
			if a.arrow.Shape == "" {
				continue
			}
			d := newXdraw()
			d.style("solid")
			d.pen(e.Color)
			op := "P"
			if a.arrow.Shape == "empty" {
				op = "p"
			} else {
				d.fill(e.Color)
			}
			d.points(op, layout.ArrowPolygon(a.arrow))
			setEdge(a.name, d)
		}
		pen := svgColor(e.Color)
		l := newXdraw()
		for _, lb := range []struct {
			t layout.Text
			c layout.Point
		}{{e.Label, e.LabelPos}, {e.XLabel, e.XLabelPos}} {
			if lb.t.Present() {
				l.label(lb.t, lb.c, pen)
			}
		}
		setEdge("_ldraw_", l)
		for _, lb := range []struct {
			name string
			t    layout.Text
			c    layout.Point
		}{{"_hldraw_", e.HeadLabel, e.HeadLabelPos}, {"_tldraw_", e.TailLabel, e.TailLabelPos}} {
			if lb.t.Present() {
				d := newXdraw()
				d.label(lb.t, lb.c, pen)
				setEdge(lb.name, d)
			}
		}
	}
}
