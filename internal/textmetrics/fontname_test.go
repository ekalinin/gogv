// SPDX-License-Identifier: MIT
// Provenance: observed dot 16.1.0

package textmetrics

import "testing"

// The expected widths are those the reference gives the label
// "abcdefghijklmnopqrstuvwxyz" at 14 points (dot -Tplain, minus the node
// margins), for names observed on Graphviz 16.1.0 built without Pango.
// -Tplain writes sizes in inches with five significant digits, so the
// comparison allows 0.01 points.
func TestLookupMatchesReference(t *testing.T) {
	const (
		times     = 167.190
		helvetica = 178.191
		courier   = 218.432
	)
	tests := []struct {
		name      string
		want      float64
		supported bool
	}{
		{"Times-Roman", times, true},
		{"times", times, true},
		{"Times New Roman", times, true},
		{"Helvetica", helvetica, true},
		{"ARIAL", helvetica, true},
		{"Liberation Sans", helvetica, true},
		{"FreeSans", helvetica, true},
		{"Courier", courier, true},
		{"Courier New", courier, true},
		{"Nimbus Mono", courier, true},
		{"Helvetica-Bold", times, true},
		{"Courier-Oblique", times, true},
		{"Times-Italic", times, true},
		{"Sans", times, true},
		{"Serif", times, true},
		{"Monospace", times, true},
		{"Unknownfont", times, true},
		{"Verdana", times, false},
		{"dejavu sans", times, false},
		{"Georgia", times, false},
		{"Consolas", times, false},
		{"Calibri", times, false},
		{"Trebuchet MS", times, false},
		{"Open Sans", times, false},
	}
	for _, tt := range tests {
		face, supported := Lookup(tt.name)
		got := face.Width("abcdefghijklmnopqrstuvwxyz", 14)
		if diff := got - tt.want; diff > 0.01 || diff < -0.01 || supported != tt.supported {
			t.Errorf("Lookup(%q): width %.3f supported %v, want %.3f %v", tt.name, got, supported, tt.want, tt.supported)
		}
	}
}
