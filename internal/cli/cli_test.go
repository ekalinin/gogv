// SPDX-License-Identifier: MIT
// Provenance: original

package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func run(t *testing.T, stdin string, args ...string) (string, string, int) {
	t.Helper()
	var out, errb bytes.Buffer
	status := Run(context.Background(), args, strings.NewReader(stdin), &out, &errb, "1.2.3")
	return out.String(), errb.String(), status
}

func TestStandardInput(t *testing.T) {
	out, errs, status := run(t, "digraph { a -> b }", "-Tsvg")
	if status != 0 || errs != "" || !strings.Contains(out, "<svg ") {
		t.Fatalf("status %d, stderr %q", status, errs)
	}
	out, _, _ = run(t, "digraph { a }", "-T", "plain")
	if !strings.HasPrefix(out, "graph 1 ") {
		t.Errorf("-T plain: %q", out)
	}
}

func TestMissingFile(t *testing.T) {
	_, errs, status := run(t, "", "-Tsvg", "missing.gv")
	if status == 0 || errs != "Error: dot: can't open missing.gv: No such file or directory\n" {
		t.Fatalf("status %d, stderr %q", status, errs)
	}
}

func TestFormats(t *testing.T) {
	if out, errs, status := run(t, "digraph { a }", "-Txdot"); status != 0 || errs != "" || !strings.Contains(out, "xdotversion=1.7") {
		t.Errorf("-Txdot: %d %q %q", status, errs, out)
	}
	if _, errs, status := run(t, "digraph {}", "-Tfoo"); status != 1 ||
		errs != "Format: \"foo\" not recognized. Use one of: canon plain plain-ext svg svg_inline dot gv xdot xdot1.2 xdot1.4 json json0 dot_json xdot_json\n" {
		t.Errorf("-Tfoo: %d %q", status, errs)
	}
	if out, errs, status := run(t, "digraph G { a }"); status != 0 || errs != "" || !strings.HasPrefix(out, "digraph G {\n\tgraph [bb=\"0,0,54,36\"];\n") {
		t.Errorf("default format: %d %q %q", status, errs, out)
	}
}

func TestOutputFile(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.gv")
	if err := os.WriteFile(in, []byte("digraph { a -> b }"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"-Tsvg", "-o" + filepath.Join(dir, "a.svg"), in}, {"-Tsvg", "-o", filepath.Join(dir, "b.svg"), in}} {
		out, errs, status := run(t, "", args...)
		if status != 0 || out != "" || errs != "" {
			t.Fatalf("%v: status %d, stdout %q, stderr %q", args, status, out, errs)
		}
	}
	for _, f := range []string{"a.svg", "b.svg"} {
		data, err := os.ReadFile(filepath.Join(dir, f))
		if err != nil || !bytes.Contains(data, []byte("<svg ")) {
			t.Errorf("%s: %v", f, err)
		}
	}
	if _, errs, status := run(t, "", "-Tsvg", "-o"); status != 1 || !strings.HasPrefix(errs, "Missing argument for -o flag\n") {
		t.Errorf("-o without a file: %d %q", status, errs)
	}
}

func TestVersion(t *testing.T) {
	out, errs, status := run(t, "this is not read", "-V")
	if status != 0 || out != "" || errs != "dot - graphviz version 16.1.0 (gogv 1.2.3)\n" {
		t.Fatalf("status %d, stdout %q, stderr %q", status, out, errs)
	}
	// gopuml reads the Graphviz version with these patterns
	// (internal/dot/graphvizversion.go).
	if m := regexp.MustCompile(`(\d+)\.(\d\d?)`).FindStringSubmatch(errs); m == nil || m[1] != "16" || m[2] != "1" {
		t.Errorf("NewGraphvizVersion pattern: %v", m)
	}
	if m := regexp.MustCompile(`\s(\d+)\.(\d\d?)\D`).FindStringSubmatch(errs); m == nil || m[1] != "16" || m[2] != "1" {
		t.Errorf("RetrieveVersion pattern: %v", m)
	}
}

func TestEngineAndOtherFlags(t *testing.T) {
	if _, errs, status := run(t, "digraph { a }", "-Kdot", "-Tcanon"); status != 0 || errs != "" {
		t.Errorf("-Kdot: %d %q", status, errs)
	}
	if _, errs, status := run(t, "digraph { a }", "-Kneato", "-Tsvg"); status != 1 || !strings.Contains(errs, `"neato" is not supported yet`) {
		t.Errorf("-Kneato: %d %q", status, errs)
	}
	for _, flag := range []string{"-Gnodesep=1", "-N", "-E", "-O", "-n", "-s", "-y", "-v"} {
		if _, errs, status := run(t, "digraph { a }", flag, "-Tsvg"); status != 1 || !strings.Contains(errs, "not supported yet") {
			t.Errorf("%s: %d %q", flag, status, errs)
		}
	}
}

func TestMessagesAndStatus(t *testing.T) {
	out, errs, status := run(t, "digraph { a [shape=record] }", "-Tsvg")
	if status != 0 || !strings.Contains(out, "<svg ") || errs != "Warning: node a: shape=record is not supported yet, using ellipse\n" {
		t.Errorf("warning: %d %q", status, errs)
	}
	_, errs, status = run(t, "digraph { a -> }", "-Tsvg")
	if status != 1 || errs != "Error: <stdin>: syntax error in line 1 near '}'\n" {
		t.Errorf("syntax error: %d %q", status, errs)
	}
}
