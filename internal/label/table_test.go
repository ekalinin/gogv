// SPDX-License-Identifier: MIT
// Provenance: original

package label

import "testing"

func TestParseTable(t *testing.T) {
	tests := []struct {
		html string
		want Table
		ok   bool
	}{
		{`<TABLE BGCOLOR="#000009" FIXEDSIZE="TRUE" WIDTH="57" HEIGHT="17"><TR><TD></TD></TR></TABLE>`, Table{Width: 57, Height: 17, Fixed: true, BgColor: "#000009", hasWidth: true, hasHeight: true}, true},
		{`<table bgcolor="#00000a" fixedsize="true" width="40" height="20"> <tr> <td></td> </tr> </table>`, Table{Width: 40, Height: 20, Fixed: true, BgColor: "#00000a", hasWidth: true, hasHeight: true}, true},
		{`<TABLE WIDTH='30' HEIGHT='10'><TR><TD></TD></TR></TABLE>`, Table{Width: 30, Height: 10, hasWidth: true, hasHeight: true}, true},
		{`<TABLE FIXEDSIZE="FALSE"><TR><TD></TD></TR></TABLE>`, Table{}, true},
		{`<TABLE><TR><TD>x</TD></TR></TABLE>`, Table{}, false},
		{`<TABLE><TR><TD PORT="p"></TD></TR></TABLE>`, Table{}, false},
		{`<TABLE><TR><TD></TD><TD></TD></TR></TABLE>`, Table{}, false},
		{`<TABLE><TR><TD></TD></TR><TR><TD></TD></TR></TABLE>`, Table{}, false},
		{`<TABLE BORDER="0"><TR><TD></TD></TR></TABLE>`, Table{}, false},
		{`<TABLE BGCOLOR="red"><TR><TD></TD></TR></TABLE>`, Table{}, false},
		{`<TABLE FIXEDSIZE="YES"><TR><TD></TD></TR></TABLE>`, Table{}, false},
		{`<TABLE WIDTH="-1"><TR><TD></TD></TR></TABLE>`, Table{}, false},
		{`<TABLE><TR><TD></TD></TR></TABLE>x`, Table{}, false},
		{`<B>x</B>`, Table{}, false},
		{`x`, Table{}, false},
	}
	for _, tt := range tests {
		got, ok := ParseTable(tt.html)
		if ok != tt.ok || ok && got != tt.want {
			t.Errorf("ParseTable(%q) = %+v, %v; want %+v, %v", tt.html, got, ok, tt.want, tt.ok)
		}
	}
}

func TestTableSize(t *testing.T) {
	for _, tt := range []struct {
		t    Table
		w, h float64
	}{
		{Table{Width: 18, Height: 11, Fixed: true, hasWidth: true, hasHeight: true}, 18, 11},
		{Table{Width: 57, Height: 17, hasWidth: true, hasHeight: true}, 57, 17},
		{Table{Width: 30, Height: 10, hasWidth: true, hasHeight: true}, 30, 10},
		{Table{Width: 40, hasWidth: true}, 40, 12},
		{Table{}, 12, 12},
	} {
		if w, h := tt.t.Size(); w != tt.w || h != tt.h {
			t.Errorf("%+v: size %v x %v, want %v x %v", tt.t, w, h, tt.w, tt.h)
		}
	}
}
