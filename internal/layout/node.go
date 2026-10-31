// SPDX-License-Identifier: MIT
// Provenance: doc https://graphviz.org/doc/info/shapes.html
// Provenance: doc https://graphviz.org/docs/attrs/regular/
// Provenance: observed dot 16.1.0

package layout

import (
	"math"

	"github.com/ekalinin/gogv/internal/fdlibm"
	"github.com/ekalinin/gogv/internal/graph"
	"github.com/ekalinin/gogv/internal/label"
	"github.com/ekalinin/gogv/internal/textmetrics"
)

// Text is a label: its lines and how they are set.
type Text struct {
	Lines     []label.Line
	FontName  string
	FontSize  float64
	FontColor string
	Face      textmetrics.Face
	// Width and Height are the size of the lines, without margins.
	Width, Height float64
}

// The margins around a node label, in points, as the reference applies
// them: 8 points on the left and right, 4 above and below.
const (
	labelMarginX = 16
	labelMarginY = 8
)

// The default sizes of the attribute documentation, in points.
const (
	defaultWidth     = 0.75 * pointsPerInch
	defaultHeight    = 0.5 * pointsPerInch
	defaultPointSize = 0.05 * pointsPerInch
	minWidth         = 0.01 * pointsPerInch
	minHeight        = 0.02 * pointsPerInch
	defaultFontSize  = 14
)

// canonicalShape maps a supported shape to the name the outputs use.
var canonicalShape = map[string]string{
	"box": "box", "rect": "rect", "rectangle": "rectangle",
	"ellipse": "ellipse", "oval": "oval", "circle": "circle",
	"diamond": "diamond", "hexagon": "hexagon", "octagon": "octagon",
	"plaintext": "plaintext", "point": "point",
}

// polygonSides is the number of sides of the regular polygon of a shape, or
// 0 for ellipses, 4 for rectangles drawn as such.
func polygonSides(shape string) int {
	switch shape {
	case "diamond":
		return 4
	case "hexagon":
		return 6
	case "octagon":
		return 8
	}
	return 0
}

// isRectangular reports whether a shape is drawn as an upright rectangle.
func isRectangular(shape string) bool {
	switch shape {
	case "box", "rect", "rectangle", "plaintext":
		return true
	}
	return false
}

// isElliptic reports whether a shape is drawn as an ellipse.
func isElliptic(shape string) bool {
	switch shape {
	case "ellipse", "oval", "circle", "point":
		return true
	}
	return false
}

// makeText resolves the lines and font of a label.
func makeText(text string, obj label.Object, fontname, fontcolor graph.Value, fontsize graph.Value) Text {
	t := Text{
		Lines:     label.Lines(label.Expand(text, obj)),
		FontName:  "Times-Roman",
		FontSize:  number(fontsize, defaultFontSize, 1),
		FontColor: supported(graph.NodeKind, "fontcolor", fontcolor, "black"),
	}
	if fontname.S != "" {
		t.FontName = fontname.S
	}
	t.Face, _ = textmetrics.Lookup(t.FontName)
	var lines []string
	for _, l := range t.Lines {
		lines = append(lines, l.Text)
	}
	t.Width, t.Height = t.Face.Measure(lines, t.FontSize)
	return t
}

// nodeLabelText returns the text of a node's label; an HTML label, which is
// not supported yet, gives the default label.
func nodeLabelText(v graph.Value) string {
	if v.HTML {
		return graph.LabelDefault
	}
	return v.S
}

// nodeSize returns the width and height of a node, in points, from its
// shape, its label and its width and height attributes, as the reference
// sizes them.
func nodeSize(shape string, text Text, width, height graph.Value) (w, h float64) {
	if shape == "point" {
		// A point is sized only by width and height: the smaller of the two
		// when both are set.
		switch {
		case isSet(width) && isSet(height):
			s := min(number(width, 0.05, 0), number(height, 0.05, 0)) * pointsPerInch
			return s, s
		case isSet(width):
			s := number(width, 0.05, 0) * pointsPerInch
			return s, s
		case isSet(height):
			s := number(height, 0.05, 0) * pointsPerInch
			return s, s
		}
		return defaultPointSize, defaultPointSize
	}

	minW := number(width, 0.75, 0.01) * pointsPerInch
	minH := number(height, 0.5, 0.02) * pointsPerInch
	if shape == "circle" {
		// A regular shape takes the smaller default size when neither is
		// set, the one that is set, or the larger of the two.
		switch {
		case isSet(width) && isSet(height):
			minW = max(minW, minH)
		case isSet(height):
			minW = minH
		case !isSet(width):
			minW = min(minW, minH)
		}
		minH = minW
	}

	// The label with its margins.
	var lw, lh float64
	if len(text.Lines) > 0 {
		lw, lh = text.Width+labelMarginX, text.Height+labelMarginY
	}

	switch {
	case isRectangular(shape):
		w, h = lw, lh
	default:
		// The ellipse that encloses the label box: with the minimum height
		// when that leaves room, else the label box scaled by sqrt 2. A
		// regular polygon encloses that ellipse.
		f := 1.0
		if sides := polygonSides(shape); sides > 0 {
			f = 1 / fdlibm.Cos(math.Pi/float64(sides))
		}
		if lh*math.Sqrt2 < minH {
			r := lh / minH
			h = minH
			w = lw / math.Sqrt(1-float64(r*r)) * f
		} else {
			w = lw * math.Sqrt2 * f
			h = lh * math.Sqrt2 * f
		}
	}
	w, h = max(w, minW), max(h, minH)
	if shape == "circle" {
		w = max(w, h)
		h = w
	}
	return w, h
}

// nodeShape returns the supported shape of a node, ellipse by default.
func nodeShape(v graph.Value) string {
	return canonicalShape[supported(graph.NodeKind, "shape", v, "ellipse")]
}

// nodeColor returns the supported color of a node, black by default.
func nodeColor(v graph.Value) string {
	return supported(graph.NodeKind, "color", v, "black")
}
