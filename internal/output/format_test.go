// SPDX-License-Identifier: MIT
// Provenance: observed dot 16.1.0

package output

import "testing"

func TestSVGNumber(t *testing.T) {
	for _, tt := range []struct {
		v    float64
		want string
	}{
		{27, "27"}, {-85.8, "-85.8"}, {-37.62, "-37.62"}, {0.004, "0"}, {-0.004, "0"},
		{136.125, "136.12"}, {1.005, "1"}, {2.5, "2.5"}, {-0.5, "-0.5"},
	} {
		if got := svgNumber(tt.v); got != tt.want {
			t.Errorf("svgNumber(%v) = %q, want %q", tt.v, got, tt.want)
		}
	}
}

func TestPlainNumber(t *testing.T) {
	for _, tt := range []struct {
		v    float64
		want string
	}{
		{0.75, "0.75"}, {2.5, "2.5"}, {0.59444444, "0.59444"}, {0.0398047, "0.039805"},
		{13.6458, "13.646"}, {2, "2"}, {1e6, "1e+06"}, {123456, "1.2346e+05"}, {0.00001234, "1.234e-05"},
	} {
		if got := plainNumber(tt.v); got != tt.want {
			t.Errorf("plainNumber(%v) = %q, want %q", tt.v, got, tt.want)
		}
	}
}

func TestXMLText(t *testing.T) {
	for _, tt := range []struct{ in, want string }{
		{"a-b", "a&#45;b"},
		{"c<d>", "c&lt;d&gt;"},
		{`x & y " z ' w`, "x &amp; y &quot; z &#39; w"},
		{"  two  spaces", " &#160;two &#160;spaces"},
		{"a&amp;b", "a&amp;b"},
		{"é ü", "é ü"},
	} {
		if got := xmlText(tt.in); got != tt.want {
			t.Errorf("xmlText(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
