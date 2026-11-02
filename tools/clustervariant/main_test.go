// SPDX-License-Identifier: MIT
// Provenance: original

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunSource(t *testing.T) {
	in := t.TempDir()
	if err := os.MkdirAll(filepath.Join(in, "class-diagram"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(in, "class-diagram", "a-1.gv"), []byte("digraph { subgraph cluster1 { a } }"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Without SOURCE, the run succeeds and writes none.
	out := filepath.Join(t.TempDir(), "out")
	if err := run(in, out); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(out, "SOURCE")); !os.IsNotExist(err) {
		t.Errorf("SOURCE written without an input SOURCE: %v", err)
	}

	// With SOURCE, the variant records the same source.
	const source = "gopuml v0.3.0\n"
	if err := os.WriteFile(filepath.Join(in, "SOURCE"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run(in, out); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(out, "SOURCE"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != source {
		t.Errorf("SOURCE = %q, want %q", got, source)
	}
	if _, err := os.Stat(filepath.Join(out, "gopuml-clusters", "class-diagram-a-1.gv")); err != nil {
		t.Error(err)
	}
}

func TestVariant(t *testing.T) {
	src := `digraph unix {
subgraph cluster6p0 {label="";subgraph cluster6 {style=solid;color="#000006";labeljust="c";label=<<TABLE BGCOLOR="#000007" FIXEDSIZE="TRUE" WIDTH="51" HEIGHT="11"><TR><TD></TD></TR></TABLE>>;
a [shape=rect,label=<<TABLE><TR><TD></TD></TR></TABLE>>];
}}
}`
	want := `digraph unix {
subgraph cluster6p0 {label="";subgraph cluster6 {style=solid;color="#000006";labeljust="c";label="";
a [shape=rect,label=<<TABLE><TR><TD></TD></TR></TABLE>>];
}}
}`
	if got := variant(src); got != want {
		t.Errorf("variant:\n%s\nwant:\n%s", got, want)
	}
}
