// SPDX-License-Identifier: MIT
// Provenance: observed dot 16.1.0

package textmetrics

import "testing"

func TestMonospacedWidths(t *testing.T) {
	courier := Face{Family: Mono}
	if a, b := courier.Width("iiii", 14), courier.Width("WWWW", 14); a != b {
		t.Errorf("Courier: iiii is %g wide, WWWW %g", a, b)
	}
}

func TestFontSizeScales(t *testing.T) {
	times := Face{Family: Serif}
	if w14, w28 := times.Width("Hello world", 14), times.Width("Hello world", 28); w28 != 2*w14 {
		t.Errorf("width at 28 is %g, at 14 %g", w28, w14)
	}
}

// The widths below are those the reference reports for the same labels
// (dot -Tplain, minus the node margins).
func TestReferenceWidths(t *testing.T) {
	times := Face{Family: Serif}
	tests := []struct {
		text string
		want float64 // points, to 0.01
	}{
		{"Hello world", 67.27},
		{"x", 7.00},
		{" ", 3.50},
		{"éééééééééééééééééééé", 140.00}, // two bytes, each a space
		{"中中中中中中中中中中中中中中中中中中中中", 210.00}, // three bytes, each a space
		{"tab\tx", 24.10},             // a control character has no width
		{"xxxxxxxxxx\x7f\x7f", 70.00}, // nor has DEL
	}
	for _, tt := range tests {
		got := times.Width(tt.text, 14)
		if diff := got - tt.want; diff > 0.005 || diff < -0.005 {
			t.Errorf("Width(%q) = %.3f, want %.2f", tt.text, got, tt.want)
		}
	}
	if got := (Face{Family: Mono}).Width("WWWWWWWWWWWWWWWWWWWW", 14); got < 168.02 || got > 168.03 {
		t.Errorf("Courier W x 20 = %.3f, want 168.025", got)
	}
}

func TestMeasureLines(t *testing.T) {
	times := Face{Family: Serif}
	w, h := times.Measure([]string{"x", "Hello world", ""}, 14)
	if w != times.Width("Hello world", 14) {
		t.Errorf("width %g is not the widest line's", w)
	}
	if want := 3 * LineHeight(14); h != want {
		t.Errorf("height %g, want %g", h, want)
	}
	if h1 := LineHeight(14); h1 < 16.79 || h1 > 16.81 {
		t.Errorf("line height %g, want 16.8", h1)
	}
}
