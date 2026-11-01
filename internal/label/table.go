// SPDX-License-Identifier: MIT
// Provenance: doc https://graphviz.org/doc/info/shapes.html
// Provenance: observed dot 16.1.0

package label

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Table is a table label: an HTML label that is a TABLE of rows of empty
// cells, as gopuml writes them to reserve room of a given size, find it
// again by its colors and attach edges to its cells by their ports.
type Table struct {
	// Width and Height are the size the table asks for, in whole points;
	// Fixed is set when FIXEDSIZE makes it that size exactly.
	Width, Height float64
	Fixed         bool
	// BgColor is the #rrggbb fill color, or "" when the table has none;
	// Color is the color of its borders, or "" for the pen color.
	BgColor, Color string
	// Border is the width of the table's border; CellBorder and
	// CellPadding are the defaults of its cells, and CellSpacing the space
	// between them and around them.
	Border, CellBorder, CellSpacing, CellPadding int
	// Port names the whole table, or is "".
	Port string
	// Rows holds the cells, row by row.
	Rows [][]Cell
	// hasWidth and hasHeight are set when the table gives WIDTH and HEIGHT.
	hasWidth, hasHeight bool
	// w and h are the size of the laid out table.
	w, h float64
}

// Cell is an empty cell of a table label.
type Cell struct {
	// Width and Height are the size the cell asks for, in whole points;
	// Fixed is set when FIXEDSIZE makes it that size exactly.
	Width, Height float64
	Fixed         bool
	// BgColor is the #rrggbb fill color, or "" when the cell has none;
	// Color is the color of its border, or "" for the table's.
	BgColor, Color string
	// Border and Padding are the cell's own, or the table's.
	Border, Padding int
	// ColSpan is the number of columns the cell spans.
	ColSpan int
	// Port names the cell, or is "".
	Port string
	// Box is where the laid out cell lies.
	Box Box
	// hasWidth and hasHeight are set when the cell gives WIDTH and HEIGHT.
	hasWidth, hasHeight bool
}

// Box is a rectangle relative to the upper left corner of its table, y
// growing downward.
type Box struct{ X, Y, W, H float64 }

// The default border, cell spacing, cell border and cell padding of the
// documentation, in points.
const (
	tableBorder      = 1
	tableCellSpacing = 2
	cellBorder       = 1
	cellPadding      = 2
)

var (
	tagRE   = regexp.MustCompile(`^<\s*(/?)\s*([A-Za-z]+)((?:\s+[A-Za-z]+\s*=\s*(?:"[^"]*"|'[^']*'))*)\s*>`)
	attrRE  = regexp.MustCompile(`([A-Za-z]+)\s*=\s*(?:"([^"]*)"|'([^']*)')`)
	colorRE = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
)

// tags reads the elements of an HTML label one tag at a time.
type tags struct{ s string }

// next returns the next tag, its name in upper case and after a "/" for a
// closing tag, with its attributes, names in upper case; ok is false when
// the text does not go on with a tag.
func (p *tags) next() (name string, attrs [][2]string, ok bool) {
	m := tagRE.FindStringSubmatch(p.s)
	if m == nil {
		return "", nil, false
	}
	for _, a := range attrRE.FindAllStringSubmatch(m[3], -1) {
		attrs = append(attrs, [2]string{strings.ToUpper(a[1]), a[2] + a[3]})
	}
	p.s = strings.TrimSpace(p.s[len(m[0]):])
	return strings.ToUpper(m[1] + m[2]), attrs, true
}

// whole reads a whole number of points from 0 to most.
func whole(v string, most int) (int, bool) {
	n, err := strconv.Atoi(v)
	return n, err == nil && n >= 0 && n <= most
}

// size reads a WIDTH or a HEIGHT, which the reference counts in whole
// points: the fraction is dropped.
func size(v string) (float64, bool) {
	n, err := strconv.ParseFloat(v, 64)
	if err != nil || n < 0 || n > 65535 {
		return 0, false
	}
	return math.Trunc(n), true
}

// fixed reads a FIXEDSIZE value.
func fixed(v string) (bool, bool) {
	switch strings.ToUpper(v) {
	case "TRUE":
		return true, true
	case "FALSE":
		return false, true
	}
	return false, false
}

// ParseTable reads the text of an HTML label, without its outer angle
// brackets, as a table label: a TABLE of one or more TR rows of one or more
// empty TD cells, with the attributes of the dot-language spec. Element
// and attribute names are compared without regard to case, and white space
// may separate the elements. ok is false for any other HTML label. The
// table is laid out as it is read.
func ParseTable(html string) (t Table, ok bool) {
	p := tags{s: strings.TrimSpace(html)}
	name, attrs, ok := p.next()
	if !ok || name != "TABLE" {
		return Table{}, false
	}
	t = Table{Border: tableBorder, CellBorder: cellBorder, CellSpacing: tableCellSpacing, CellPadding: cellPadding}
	if !t.set(attrs) {
		return Table{}, false
	}
	for {
		name, attrs, ok := p.next()
		switch {
		case !ok || len(attrs) > 0:
			return Table{}, false
		case name == "/TABLE" && len(t.Rows) > 0:
			if p.s != "" {
				return Table{}, false
			}
			t.layout()
			return t, true
		case name != "TR":
			return Table{}, false
		}
		var row []Cell
		for {
			name, attrs, ok := p.next()
			if !ok {
				return Table{}, false
			}
			if name == "/TR" && len(attrs) == 0 && len(row) > 0 {
				break
			}
			if name != "TD" {
				return Table{}, false
			}
			c := Cell{Border: t.CellBorder, Padding: t.CellPadding, ColSpan: 1}
			if !c.set(attrs) {
				return Table{}, false
			}
			if name, attrs, ok := p.next(); !ok || name != "/TD" || len(attrs) > 0 {
				return Table{}, false
			}
			row = append(row, c)
		}
		t.Rows = append(t.Rows, row)
	}
}

// set reads the attributes of the TABLE element.
func (t *Table) set(attrs [][2]string) bool {
	for _, a := range attrs {
		ok := true
		switch v := a[1]; a[0] {
		case "BORDER":
			t.Border, ok = whole(v, 255)
		case "CELLBORDER":
			t.CellBorder, ok = whole(v, 127)
		case "CELLSPACING":
			t.CellSpacing, ok = whole(v, 127)
		case "CELLPADDING":
			t.CellPadding, ok = whole(v, 255)
		case "BGCOLOR":
			t.BgColor, ok = v, colorRE.MatchString(v)
		case "COLOR":
			t.Color, ok = v, colorRE.MatchString(v)
		case "FIXEDSIZE":
			t.Fixed, ok = fixed(v)
		case "WIDTH":
			t.Width, ok = size(v)
			t.hasWidth = true
		case "HEIGHT":
			t.Height, ok = size(v)
			t.hasHeight = true
		case "PORT":
			t.Port = v
		default:
			ok = false
		}
		if !ok {
			return false
		}
	}
	return true
}

// set reads the attributes of a TD element. Its CELLSPACING changes
// nothing, as in the reference.
func (c *Cell) set(attrs [][2]string) bool {
	for _, a := range attrs {
		ok := true
		switch v := a[1]; a[0] {
		case "BORDER":
			c.Border, ok = whole(v, 127)
		case "CELLSPACING":
			_, ok = whole(v, 127)
		case "CELLPADDING":
			c.Padding, ok = whole(v, 255)
		case "BGCOLOR":
			c.BgColor, ok = v, colorRE.MatchString(v)
		case "COLOR":
			c.Color, ok = v, colorRE.MatchString(v)
		case "COLSPAN":
			c.ColSpan, ok = whole(v, 65535)
			ok = ok && c.ColSpan > 0
		case "FIXEDSIZE":
			c.Fixed, ok = fixed(v)
		case "WIDTH":
			c.Width, ok = size(v)
			c.hasWidth = true
		case "HEIGHT":
			c.Height, ok = size(v)
			c.hasHeight = true
		case "PORT":
			c.Port = v
		default:
			ok = false
		}
		if !ok {
			return false
		}
	}
	return true
}

// size returns the size that the cell needs: its border and its padding
// on both sides, its WIDTH and HEIGHT when they are larger, or exactly
// those when it is fixed.
func (c *Cell) size() (w, h float64) {
	least := float64(2*c.Border + 2*c.Padding)
	w, h = least, least
	if c.hasWidth {
		w = c.Width
		if !c.Fixed {
			w = max(w, least)
		}
	}
	if c.hasHeight {
		h = c.Height
		if !c.Fixed {
			h = max(h, least)
		}
	}
	return w, h
}

// layout gives the table its size and its cells their boxes. A column is
// as wide as its widest cell and a row as high as its highest; a spanning
// cell wider than its columns, with the spacing between them, widens them
// by equal parts, in the order of the cells; each cell fills its row and
// its columns, and the cells stay at the upper left of a table larger than
// they are.
func (t *Table) layout() {
	cols := 0
	for _, row := range t.Rows {
		n := 0
		for _, c := range row {
			n += c.ColSpan
		}
		cols = max(cols, n)
	}
	colW := make([]float64, cols)
	rowH := make([]float64, len(t.Rows))
	for r, row := range t.Rows {
		col := 0
		for i := range row {
			w, h := row[i].size()
			rowH[r] = max(rowH[r], h)
			if row[i].ColSpan == 1 {
				colW[col] = max(colW[col], w)
			}
			col += row[i].ColSpan
		}
	}
	sp := float64(t.CellSpacing)
	for _, row := range t.Rows {
		col := 0
		for i := range row {
			if n := row[i].ColSpan; n > 1 {
				w, _ := row[i].size()
				// The product held apart from the sum, so that it is not
				// fused into a multiply-add.
				have := float64(sp * float64(n-1))
				for j := col; j < col+n; j++ {
					have += colW[j]
				}
				if w > have {
					d := (w - have) / float64(n)
					for j := col; j < col+n; j++ {
						colW[j] += d
					}
				}
			}
			col += row[i].ColSpan
		}
	}
	b := float64(t.Border)
	x := make([]float64, cols+1)
	x[0] = b + sp
	for j := range cols {
		x[j+1] = x[j] + colW[j] + sp
	}
	y := b + sp
	for r, row := range t.Rows {
		col := 0
		for i := range row {
			end := col + row[i].ColSpan
			row[i].Box = Box{X: x[col], Y: y, W: x[end] - sp - x[col], H: rowH[r]}
			col = end
		}
		y += rowH[r] + sp
	}
	t.w, t.h = x[cols]+b, y+b
	if t.hasWidth {
		t.w = t.Width
		if !t.Fixed {
			t.w = max(t.w, x[cols]+b)
		}
	}
	if t.hasHeight {
		t.h = t.Height
		if !t.Fixed {
			t.h = max(t.h, y+b)
		}
	}
}

// Size returns the size of the laid out table.
func (t Table) Size() (w, h float64) { return t.w, t.h }

// PortBox returns the box of the port called name: the whole table for the
// table's own PORT, or the first cell with that PORT.
func (t Table) PortBox(name string) (Box, bool) {
	if t.Port == name && name != "" {
		return Box{W: t.w, H: t.h}, true
	}
	for _, row := range t.Rows {
		for _, c := range row {
			if c.Port == name && name != "" {
				return c.Box, true
			}
		}
	}
	return Box{}, false
}

// compassPoints are the compass points of the documentation's portPos.
var compassPoints = map[string]bool{"n": true, "ne": true, "e": true, "se": true, "s": true, "sw": true, "w": true, "nw": true, "c": true, "_": true}

// SplitPort splits the value of a tailport or headport into a port name
// and a compass point, either of which may be "": a name that the table
// defines is a port even when it is also a compass point, and so is a name
// that it does not define. t is nil for a node without a table label.
func SplitPort(t *Table, v string) (port, compass string) {
	if t != nil {
		if _, ok := t.PortBox(v); ok {
			return v, ""
		}
	}
	if compassPoints[v] {
		return "", v
	}
	if i := strings.LastIndexByte(v, ':'); i >= 0 && compassPoints[v[i+1:]] {
		return v[:i], v[i+1:]
	}
	return v, ""
}
