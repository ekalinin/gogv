// SPDX-License-Identifier: MIT
// Provenance: doc https://graphviz.org/docs/outputs/plain/
// Provenance: observed dot 16.1.0

package output

import (
	"bytes"
	"strconv"

	"github.com/ekalinin/gogv/internal/layout"
)

// Plain writes a laid-out graph in the plain format: sizes and coordinates
// in inches, y growing upward.
func Plain(res *layout.Result) []byte {
	in := func(v float64) string { return plainNumber(v / 72) }
	var b bytes.Buffer
	b.WriteString("graph 1 " + in(res.Width) + " " + in(res.Height) + "\n")
	for _, n := range res.Nodes {
		fill := "lightgrey"
		if n.ColorSet {
			fill = n.Color
		}
		b.WriteString("node " + quoteID(n.Node.Name, false) + " " + in(n.X) + " " + in(n.Y) + " " + in(n.W) + " " + in(n.H) +
			" " + quoteID(n.LabelText, n.Label.Table != nil) + " " + n.Style + " " + n.Shape + " " + n.Color + " " + fill + "\n")
	}
	for _, e := range res.Edges {
		b.WriteString("edge " + quoteID(e.Edge.Tail.Name, false) + " " + quoteID(e.Edge.Head.Name, false) + " " + strconv.Itoa(len(e.Points)))
		for _, p := range e.Points {
			b.WriteString(" " + in(p.X) + " " + in(p.Y))
		}
		if e.Label.Present() {
			text := e.LabelText
			if e.Label.Table != nil {
				// The reference writes an HTML edge label so.
				text = "<TABLE>"
			}
			b.WriteString(" " + quoteID(text, false) + " " + in(e.LabelPos.X) + " " + in(e.LabelPos.Y))
		}
		style := "solid"
		if e.Invisible {
			style = "invis"
		}
		b.WriteString(" " + style + " " + e.Color + "\n")
	}
	b.WriteString("stop\n")
	return b.Bytes()
}
