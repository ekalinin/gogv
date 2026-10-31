// SPDX-License-Identifier: MIT
// Provenance: doc https://graphviz.org/doc/info/shapes.html
// Provenance: observed dot 16.1.0

package label

import (
	"regexp"
	"strconv"
	"strings"
)

// Table is a single-cell table label: an HTML label that is a TABLE with one
// row of one empty cell, as gopuml writes them to reserve room of a given
// size and find it again by its color.
type Table struct {
	// Width and Height are the size the table asks for, in points; Fixed is
	// set when FIXEDSIZE makes it that size exactly.
	Width, Height float64
	Fixed         bool
	// BgColor is the #rrggbb fill color, or "" when the table has none.
	BgColor string
	// hasWidth and hasHeight are set when the table gives WIDTH and HEIGHT.
	hasWidth, hasHeight bool
}

// The default border, cell spacing, cell border and cell padding of the
// documentation, in points.
const (
	tableBorder      = 1
	tableCellSpacing = 2
	cellBorder       = 1
	cellPadding      = 2
)

// EmptyCell is the side of an empty cell with the default cell border and
// padding.
const EmptyCell = 2*cellBorder + 2*cellPadding

// CellInset is how far an empty cell lies inside the table's upper left
// corner: the table's border and the cell spacing.
const CellInset = tableBorder + tableCellSpacing

// minTable is the side of a table of one empty cell with the default border
// and spacing.
const minTable = EmptyCell + 2*CellInset

var (
	tagRE   = regexp.MustCompile(`^<\s*(/?)\s*([A-Za-z]+)((?:\s+[A-Za-z]+\s*=\s*(?:"[^"]*"|'[^']*'))*)\s*>`)
	attrRE  = regexp.MustCompile(`([A-Za-z]+)\s*=\s*(?:"([^"]*)"|'([^']*)')`)
	colorRE = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
)

// ParseTable reads the text of an HTML label, without its outer angle
// brackets, as a single-cell table label. Element and attribute names are
// compared without regard to case, and white space may separate the
// elements. ok is false for any other HTML label.
func ParseTable(html string) (t Table, ok bool) {
	want := []string{"TABLE", "TR", "TD", "/TD", "/TR", "/TABLE"}
	s := strings.TrimSpace(html)
	for i, w := range want {
		m := tagRE.FindStringSubmatch(s)
		if m == nil {
			return Table{}, false
		}
		if strings.ToUpper(m[1]+m[2]) != w {
			return Table{}, false
		}
		attrs := attrRE.FindAllStringSubmatch(m[3], -1)
		if i > 0 && len(attrs) > 0 {
			return Table{}, false
		}
		for _, a := range attrs {
			v := a[2] + a[3]
			switch strings.ToUpper(a[1]) {
			case "BGCOLOR":
				if !colorRE.MatchString(v) {
					return Table{}, false
				}
				t.BgColor = v
			case "FIXEDSIZE":
				switch strings.ToUpper(v) {
				case "TRUE":
					t.Fixed = true
				case "FALSE":
				default:
					return Table{}, false
				}
			case "WIDTH", "HEIGHT":
				n, err := strconv.ParseFloat(v, 64)
				if err != nil || n < 0 {
					return Table{}, false
				}
				if strings.ToUpper(a[1]) == "WIDTH" {
					t.Width, t.hasWidth = n, true
				} else {
					t.Height, t.hasHeight = n, true
				}
			default:
				return Table{}, false
			}
		}
		s = strings.TrimSpace(s[len(m[0]):])
	}
	return t, s == ""
}

// Size returns the size of the table's label: the size it asks for, fixed
// or not, even below the size of its cell, as the reference draws it; a
// side that it does not give is the side of a table of one empty cell.
func (t Table) Size() (w, h float64) {
	w, h = t.Width, t.Height
	if !t.hasWidth {
		w = minTable
	}
	if !t.hasHeight {
		h = minTable
	}
	return w, h
}
