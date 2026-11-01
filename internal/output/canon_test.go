// SPDX-License-Identifier: MIT
// Provenance: observed dot 16.1.0

package output

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/ekalinin/gogv/internal/conformance"
	"github.com/ekalinin/gogv/internal/dotlang"
)

// TestCanonMatchesReference writes every curated case that the reference
// reads without an error as canon and compares it with the reference's
// output, byte for byte.
func TestCanonMatchesReference(t *testing.T) {
	const casesDir = "../../testdata/conformance/cases"
	const refDir = "../../testdata/conformance/reference"
	cases, err := conformance.LoadCases(casesDir)
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, c := range cases {
		ref, err := conformance.ReadReference(refDir, c.ID)
		if err != nil {
			t.Fatalf("%s: %v", c.ID, err)
		}
		if ref["canon"].Status != 0 {
			continue
		}
		file, _, err := dotlang.Parse(string(c.Source))
		if err != nil {
			t.Errorf("%s: %v", c.ID, err)
			continue
		}
		var got bytes.Buffer
		for _, decl := range file.Graphs {
			// As Render builds the graph: clusters resolved first.
			g := dotlang.Build(decl)
			g.ResolveClusters()
			got.Write(Canon(g))
		}
		checked++
		if err := conformance.CompareCanon(ref["canon"].Stdout, got.Bytes()); err != nil {
			t.Errorf("%s: %v", c.ID, err)
		}
	}
	if checked == 0 {
		t.Fatal("no case checked")
	}
	t.Logf("%d cases checked", checked)
}

func TestCanonQuoting(t *testing.T) {
	c := &canon{}
	for _, tt := range []struct{ in, want string }{
		{"a", "a"},
		{"_a", "_a"},
		{"A1", "A1"},
		{"\xc3\xa9", "\xc3\xa9"},
		{"12", "12"},
		{"-3.5", "-3.5"},
		{".5", ".5"},
		{"1e5", `"1e5"`},
		{"1.5.6", `"1.5.6"`},
		{"a b", `"a b"`},
		{`c"d`, `"c\"d"`},
		{"", `""`},
		{"node", `"node"`},
		{"Graph", `"Graph"`},
		{`\N`, `"\N"`},
	} {
		if got := c.id(tt.in, false); got != tt.want {
			t.Errorf("id(%q) = %s, want %s", tt.in, got, tt.want)
		}
	}
	if got := c.id("<b>x</b>", true); got != "<<b>x</b>>" {
		t.Errorf("HTML id = %s", got)
	}
}

func TestMain(m *testing.M) {
	// The tests read the corpus relative to this package.
	if _, err := os.Stat(filepath.Join("..", "..", "go.mod")); err != nil {
		panic("tests must run in internal/output")
	}
	os.Exit(m.Run())
}
