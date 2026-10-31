// SPDX-License-Identifier: MIT
// Provenance: observed dot 16.1.0

package textmetrics

// Family is a family of metrics: Liberation Serif for Times, Liberation Sans
// for Helvetica and Arial, Liberation Mono for Courier.
type Family uint8

// The metric families.
const (
	Serif Family = iota
	Sans
	Mono
)

// Style is the weight and slant of a face.
type Style uint8

// The styles of a family.
const (
	Regular Style = iota
	Bold
	Italic
	BoldItalic
)

// Face selects one advance-width table.
type Face struct {
	Family Family
	Style  Style
}

// LineSpacing is the height of a line of text as a multiple of the font
// size, as the reference lays out labels.
const LineSpacing = 1.2

func (f Face) table() *faceTable {
	return &faceTables[int(f.Family)*4+int(f.Style)]
}

// units returns the advance width of s in font units. As in the reference,
// ASCII control characters (and DEL) have no width, and every byte of a
// non-ASCII character is as wide as a space.
func (f Face) units(s string) int {
	t := f.table()
	ascii := t.ranges[0]
	space := int(ascii.widths[' '-ascii.first])
	var u int
	for i := 0; i < len(s); i++ {
		c := rune(s[i])
		switch {
		case c >= 0x80:
			u += space
		case c >= ascii.first && int(c-ascii.first) < len(ascii.widths):
			u += int(ascii.widths[c-ascii.first])
		}
	}
	return u
}

// Width returns the width of one line of text in points, set in this face
// at size points.
func (f Face) Width(s string, size float64) float64 {
	return float64(f.units(s)) * size / float64(f.table().unitsPerEm)
}

// LineHeight returns the height of one line of text set at size points.
func LineHeight(size float64) float64 {
	return size * LineSpacing
}

// Measure returns the width of the widest line and the total height of the
// lines, in points.
func (f Face) Measure(lines []string, size float64) (width, height float64) {
	for _, l := range lines {
		width = max(width, f.Width(l, size))
	}
	return width, float64(len(lines)) * LineHeight(size)
}
