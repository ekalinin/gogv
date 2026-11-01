// SPDX-License-Identifier: MIT
// Provenance: original

package label

import "testing"

// grid is a table without borders, spacing or padding.
const grid = `BORDER="0" CELLBORDER="0" CELLSPACING="0" CELLPADDING="0"`

func TestParseTable(t *testing.T) {
	tests := []struct {
		html string
		ok   bool
	}{
		{`<TABLE BGCOLOR="#000009" FIXEDSIZE="TRUE" WIDTH="57" HEIGHT="17"><TR><TD></TD></TR></TABLE>`, true},
		{`<table bgcolor="#00000a" fixedsize="true" width="40" height="20"> <tr> <td></td> </tr> </table>`, true},
		{`<TABLE WIDTH='30' HEIGHT='10'><TR><TD></TD></TR></TABLE>`, true},
		{`<TABLE FIXEDSIZE="FALSE"><TR><TD></TD></TR></TABLE>`, true},
		{`<TABLE><TR><TD PORT="p"></TD></TR></TABLE>`, true},
		{`<TABLE><TR><TD></TD><TD></TD></TR></TABLE>`, true},
		{`<TABLE><TR><TD></TD></TR><TR><TD></TD></TR></TABLE>`, true},
		{`<TABLE BORDER="0" CELLBORDER="2" CELLSPACING="3" CELLPADDING="4" COLOR="#000001" PORT="t"><TR><TD></TD></TR></TABLE>`, true},
		{`<TABLE><TR><TD BORDER="0" CELLSPACING="1" CELLPADDING="0" BGCOLOR="#000001" COLOR="#000002" COLSPAN="2" FIXEDSIZE="TRUE" WIDTH="1.5" HEIGHT="0.0" PORT="p"></TD></TR></TABLE>`, true},
		{`<TABLE><TR><TD>x</TD></TR></TABLE>`, false},
		{`<TABLE><TR><TD><TABLE><TR><TD></TD></TR></TABLE></TD></TR></TABLE>`, false},
		{`<TABLE><TR><TD ROWSPAN="2"></TD></TR><TR><TD></TD></TR></TABLE>`, false},
		{`<TABLE><TR ALIGN="left"><TD></TD></TR></TABLE>`, false},
		{`<TABLE ALIGN="left"><TR><TD></TD></TR></TABLE>`, false},
		{`<TABLE><TR><TD ALIGN="left"></TD></TR></TABLE>`, false},
		{`<TABLE><TR></TR></TABLE>`, false},
		{`<TABLE></TABLE>`, false},
		{`<TABLE BORDER="x"><TR><TD></TD></TR></TABLE>`, false},
		{`<TABLE CELLSPACING="200"><TR><TD></TD></TR></TABLE>`, false},
		{`<TABLE><TR><TD COLSPAN="0"></TD></TR></TABLE>`, false},
		{`<TABLE BGCOLOR="red"><TR><TD></TD></TR></TABLE>`, false},
		{`<TABLE><TR><TD COLOR="red"></TD></TR></TABLE>`, false},
		{`<TABLE FIXEDSIZE="YES"><TR><TD></TD></TR></TABLE>`, false},
		{`<TABLE WIDTH="-1"><TR><TD></TD></TR></TABLE>`, false},
		{`<TABLE><TR><TD></TD></TR></TABLE>x`, false},
		{`<B>x</B>`, false},
		{`x`, false},
	}
	for _, tt := range tests {
		if _, ok := ParseTable(tt.html); ok != tt.ok {
			t.Errorf("ParseTable(%q): ok %v, want %v", tt.html, ok, tt.ok)
		}
	}
}

func TestTableAttributes(t *testing.T) {
	tb, ok := ParseTable(`<table border="0" cellborder="2" cellspacing="3" cellpadding="4" bgcolor="#00000A" color="#00000B" fixedsize="true" width="57.9" height="17" port="t">
		<tr> <td bgcolor="#000001" color="#000002" border="1" cellpadding="0" colspan="2" fixedsize="true" width="41.5771484375" height="0.0" port="p"></td> </tr>
		<tr> <td></td> <td cellspacing="9"></td> </tr>
	</table>`)
	if !ok {
		t.Fatal("not a table label")
	}
	if tb.Border != 0 || tb.CellBorder != 2 || tb.CellSpacing != 3 || tb.CellPadding != 4 || tb.BgColor != "#00000A" || tb.Color != "#00000B" || !tb.Fixed || tb.Width != 57 || tb.Height != 17 || tb.Port != "t" {
		t.Errorf("table attributes: %+v", tb)
	}
	c := tb.Rows[0][0]
	if c.BgColor != "#000001" || c.Color != "#000002" || c.Border != 1 || c.Padding != 0 || c.ColSpan != 2 || !c.Fixed || c.Width != 41 || c.Height != 0 || c.Port != "p" {
		t.Errorf("cell attributes: %+v", c)
	}
	if d := tb.Rows[1][0]; d.Border != 2 || d.Padding != 4 || d.ColSpan != 1 {
		t.Errorf("cell defaults from the table: %+v", d)
	}
}

func TestTableSize(t *testing.T) {
	for _, tt := range []struct {
		html string
		w, h float64
	}{
		// Single-cell tables, as gopuml writes them.
		{`<TABLE FIXEDSIZE="TRUE" WIDTH="18" HEIGHT="11"><TR><TD></TD></TR></TABLE>`, 18, 11},
		{`<TABLE WIDTH="57" HEIGHT="17"><TR><TD></TD></TR></TABLE>`, 57, 17},
		{`<TABLE FIXEDSIZE="TRUE" WIDTH="30" HEIGHT="10"><TR><TD></TD></TR></TABLE>`, 30, 10},
		// Without FIXEDSIZE, a table is at least as large as its content.
		{`<TABLE WIDTH="30" HEIGHT="10"><TR><TD></TD></TR></TABLE>`, 30, 12},
		{`<TABLE WIDTH="10" HEIGHT="5"><TR><TD></TD><TD></TD></TR></TABLE>`, 20, 12},
		{`<TABLE WIDTH="40"><TR><TD></TD></TR></TABLE>`, 40, 12},
		{`<TABLE FIXEDSIZE="TRUE" WIDTH="40"><TR><TD></TD></TR></TABLE>`, 40, 12},
		{`<TABLE><TR><TD></TD></TR></TABLE>`, 12, 12},
		// Default borders, spacing and padding.
		{`<TABLE><TR><TD></TD><TD></TD></TR><TR><TD COLSPAN="2"></TD></TR></TABLE>`, 20, 20},
		// Borders, spacing and padding.
		{`<TABLE BORDER="2" CELLBORDER="3" CELLSPACING="4" CELLPADDING="5"><TR><TD></TD><TD WIDTH="30"></TD></TR></TABLE>`, 62, 28},
		// A cell's own border and padding, and its spacing, which changes
		// nothing.
		{`<TABLE CELLPADDING="4"><TR><TD BORDER="3" CELLPADDING="1"></TD><TD CELLSPACING="6"></TD></TR></TABLE>`, 26, 16},
		// Fractions of a point are dropped.
		{`<TABLE FIXEDSIZE="TRUE" WIDTH="41.9" HEIGHT="17.9"><TR><TD></TD></TR></TABLE>`, 41, 17},
		{`<TABLE ` + grid + `><TR><TD FIXEDSIZE="TRUE" WIDTH="50.9" HEIGHT="20.5"></TD><TD WIDTH="30.7" HEIGHT="10.2"></TD></TR></TABLE>`, 80, 20},
	} {
		tb, ok := ParseTable(tt.html)
		if !ok {
			t.Errorf("%s: not a table label", tt.html)
			continue
		}
		if w, h := tb.Size(); w != tt.w || h != tt.h {
			t.Errorf("%s: size %v x %v, want %v x %v", tt.html, w, h, tt.w, tt.h)
		}
	}
}

func TestCellBoxes(t *testing.T) {
	boxes := func(html string) [][]Box {
		tb, ok := ParseTable(html)
		if !ok {
			t.Fatalf("%s: not a table label", html)
		}
		var out [][]Box
		for _, row := range tb.Rows {
			var r []Box
			for _, c := range row {
				r = append(r, c.Box)
			}
			out = append(out, r)
		}
		return out
	}
	near := func(a, b float64) bool { return a-b < 0.01 && b-a < 0.01 }
	check := func(name string, got [][]Box, want [][]Box) {
		t.Helper()
		for r := range want {
			for i, w := range want[r] {
				g := got[r][i]
				if !near(g.X, w.X) || !near(g.Y, w.Y) || !near(g.W, w.W) || !near(g.H, w.H) {
					t.Errorf("%s: cell %d,%d at %+v, want %+v", name, r, i, g, w)
				}
			}
		}
	}

	// A single empty cell lies 3 points inside the upper left corner.
	check("single cell", boxes(`<TABLE FIXEDSIZE="TRUE" WIDTH="57" HEIGHT="17"><TR><TD></TD></TR></TABLE>`),
		[][]Box{{{3, 3, 6, 6}}})
	check("default grid", boxes(`<TABLE><TR><TD></TD><TD></TD></TR><TR><TD COLSPAN="2"></TD></TR></TABLE>`),
		[][]Box{{{3, 3, 6, 6}, {11, 3, 6, 6}}, {{3, 11, 14, 6}}})
	// A spanning cell widens its columns by equal parts.
	check("spanning cell", boxes(`<TABLE `+grid+`><TR><TD WIDTH="10"></TD><TD WIDTH="20"></TD></TR><TR><TD COLSPAN="2" WIDTH="60"></TD></TR></TABLE>`),
		[][]Box{{{0, 0, 25, 0}, {25, 0, 35, 0}}, {{0, 0, 60, 0}}})
	// In fractions of a point, the spanning cells taken in order.
	uneven := `<TR><TD WIDTH="10" HEIGHT="8"></TD><TD WIDTH="10" HEIGHT="8"></TD><TD WIDTH="10" HEIGHT="8"></TD></TR>`
	three := `<TR><TD COLSPAN="3" WIDTH="35" HEIGHT="8"></TD></TR>`
	two := `<TR><TD COLSPAN="2" WIDTH="25" HEIGHT="8"></TD><TD></TD></TR>`
	check("uneven span", boxes(`<TABLE `+grid+`>`+uneven+three+two+`</TABLE>`),
		[][]Box{{{0, 0, 12.5, 8}, {12.5, 0, 12.5, 8}, {25, 0, 11.667, 8}}, {{0, 8, 36.667, 8}}, {{0, 16, 25, 8}, {25, 16, 11.667, 8}}})
	check("uneven span, rows the other way round", boxes(`<TABLE `+grid+`>`+uneven+two+three+`</TABLE>`),
		[][]Box{{{0, 0, 12.5, 8}, {12.5, 0, 12.5, 8}, {25, 0, 10, 8}}, {{0, 8, 25, 8}, {25, 8, 10, 8}}, {{0, 16, 35, 8}}})
	// The spacing between the columns counts as width.
	check("span with spacing", boxes(`<TABLE BORDER="0" CELLBORDER="0" CELLSPACING="3" CELLPADDING="0"><TR><TD WIDTH="10" HEIGHT="8"></TD><TD WIDTH="10" HEIGHT="8"></TD></TR><TR><TD COLSPAN="2" WIDTH="30" HEIGHT="8"></TD></TR></TABLE>`),
		[][]Box{{{3, 3, 13.5, 8}, {19.5, 3, 13.5, 8}}, {{3, 14, 30, 8}}})
	// A cell fills its row.
	check("cell filling its row", boxes(`<TABLE `+grid+`><TR><TD FIXEDSIZE="TRUE" WIDTH="40" HEIGHT="20"></TD><TD FIXEDSIZE="TRUE" WIDTH="60" HEIGHT="30"></TD></TR></TABLE>`),
		[][]Box{{{0, 0, 40, 30}, {40, 0, 60, 30}}})
	// The cells stay at the upper left of a larger table.
	check("larger table", boxes(`<TABLE WIDTH="100" HEIGHT="40"><TR><TD></TD><TD WIDTH="20"></TD></TR></TABLE>`),
		[][]Box{{{3, 3, 6, 6}, {11, 3, 20, 6}}})
	// A fixed cell that does not give a side takes its content's.
	check("fixed cell without a height", boxes(`<TABLE><TR><TD FIXEDSIZE="TRUE" WIDTH="20"></TD><TD FIXEDSIZE="TRUE"></TD></TR></TABLE>`),
		[][]Box{{{3, 3, 20, 6}, {25, 3, 6, 6}}})
}

func TestPortBox(t *testing.T) {
	tb, _ := ParseTable(`<TABLE PORT="t" ` + grid + `><TR><TD FIXEDSIZE="TRUE" WIDTH="40" HEIGHT="20" PORT="l"></TD><TD FIXEDSIZE="TRUE" WIDTH="60" HEIGHT="30" PORT="r"></TD></TR></TABLE>`)
	for _, tt := range []struct {
		port string
		box  Box
		ok   bool
	}{
		{"t", Box{0, 0, 100, 30}, true},
		{"l", Box{0, 0, 40, 30}, true},
		{"r", Box{40, 0, 60, 30}, true},
		{"x", Box{}, false},
		{"", Box{}, false},
	} {
		if box, ok := tb.PortBox(tt.port); box != tt.box || ok != tt.ok {
			t.Errorf("PortBox(%q) = %+v, %v; want %+v, %v", tt.port, box, ok, tt.box, tt.ok)
		}
	}
}
