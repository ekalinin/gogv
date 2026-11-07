// SPDX-License-Identifier: MIT
// Provenance: original

package main

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ekalinin/gogv"
	"github.com/ekalinin/gogv/internal/conformance"
	"github.com/ekalinin/gogv/internal/dotlang"
)

// repoRoot is the root of the repository, from the directory of the test.
const repoRoot = "../.."

func TestPrintRoundTrip(t *testing.T) {
	// The printed input of every curated case, and of every gopuml case when
	// the corpus is collected, gives the same canon output as the original.
	var files []string
	for _, dir := range []string{"testdata/conformance/cases", ".cache/gopuml-dot", ".cache/gopuml-cluster-dot"} {
		filepath.WalkDir(filepath.Join(repoRoot, dir), func(p string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && filepath.Ext(p) == ".gv" {
				files = append(files, p)
			}
			return nil
		})
	}
	if len(files) < 250 {
		t.Fatalf("%d cases found, want the curated cases at least", len(files))
	}
	checked := 0
	for _, name := range files {
		src, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		want, _, err := gogv.Render(context.Background(), string(src), "canon")
		if err != nil {
			continue // a case of a syntax error
		}
		f, _, err := dotlang.Parse(string(src))
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		printed := printFile(f)
		got, _, err := gogv.Render(context.Background(), printed, "canon")
		if err != nil || string(got) != string(want) {
			t.Errorf("%s: the printed input gives another canon output (error %v):\n%s", name, err, printed)
		}
		checked++
	}
	t.Logf("%d of %d inputs checked", checked, len(files))
}

// nodesOf returns the names of the nodes of a DOT input, in order.
func nodesOf(t *testing.T, src string) []string {
	t.Helper()
	f, _, err := dotlang.Parse(src)
	if err != nil {
		t.Fatalf("%v\n%s", err, src)
	}
	var names []string
	seen := map[string]bool{}
	add := func(ids []dotlang.NodeID) {
		for _, n := range ids {
			if !seen[n.ID.Text] {
				seen[n.ID.Text] = true
				names = append(names, n.ID.Text)
			}
		}
	}
	var walk func([]dotlang.Stmt)
	walk = func(stmts []dotlang.Stmt) {
		for _, s := range stmts {
			switch s := s.(type) {
			case *dotlang.NodeStmt:
				add(s.Nodes)
			case *dotlang.EdgeStmt:
				for _, o := range s.Operands {
					add(o.Nodes)
				}
			case *dotlang.Subgraph:
				walk(s.Stmts)
			}
		}
	}
	for _, g := range f.Graphs {
		walk(g.Stmts)
	}
	return names
}

// hasEdge reports whether a DOT input has an edge statement from tail to
// head.
func hasEdge(t *testing.T, src, tail, head string) bool {
	t.Helper()
	f, _, err := dotlang.Parse(src)
	if err != nil {
		t.Fatalf("%v\n%s", err, src)
	}
	found := false
	var walk func([]dotlang.Stmt)
	walk = func(stmts []dotlang.Stmt) {
		for _, s := range stmts {
			switch s := s.(type) {
			case *dotlang.EdgeStmt:
				if len(s.Operands) == 2 && len(s.Operands[0].Nodes) == 1 && len(s.Operands[1].Nodes) == 1 &&
					s.Operands[0].Nodes[0].ID.Text == tail && s.Operands[1].Nodes[0].ID.Text == head {
					found = true
				}
			case *dotlang.Subgraph:
				walk(s.Stmts)
			}
		}
	}
	for _, g := range f.Graphs {
		walk(g.Stmts)
	}
	return found
}

// fakeLayouts returns a reference that puts every node at the origin, and a
// gogv layout that moves the nodes that moved names for an input, by half
// of the graph's size.
func fakeLayouts(t *testing.T, moved func(src string) []string) (layout, layout) {
	place := func(src string, shifted []string) conformance.Plain {
		p := conformance.Plain{Width: 1, Height: 1, NodeCenters: map[string]conformance.Point{}, LabelCenters: map[string]conformance.Point{}}
		for _, n := range nodesOf(t, src) {
			p.Nodes = append(p.Nodes, n)
			p.NodeCenters[n] = conformance.Point{}
		}
		for _, n := range shifted {
			if _, ok := p.NodeCenters[n]; ok {
				p.NodeCenters[n] = conformance.Point{X: 0.5}
			}
		}
		return p
	}
	ref := func(src string) (conformance.Plain, error) { return place(src, nil), nil }
	got := func(src string) (conformance.Plain, error) { return place(src, moved(src)), nil }
	return ref, got
}

func TestReduceScenarios(t *testing.T) {
	// Spec scenarios "Unrelated statements removed" and "Another difference"
	// of the requirement "Reduction of a layout difference" of
	// upstream-compatibility, with fake references.
	unrelated := `digraph {
	x; y;
	n1 -> n2; n3 -> n4; n5 -> n6;
	subgraph cluster_a { n7; n8 -> n9 [label="l"] }
	n10 [shape=box];
	x -> y [color=red];
}
`
	t.Run("Unrelated statements removed", func(t *testing.T) {
		// The difference lies in x and the edge from x to y.
		ref, got := fakeLayouts(t, func(src string) []string {
			if hasEdge(t, src, "x", "y") {
				return []string{"x"}
			}
			return nil
		})
		test, failing, err := predicate(unrelated, ref, got)
		if err != nil || strings.Join(failing, ",") != "node x" {
			t.Fatalf("failing objects %q, error %v", failing, err)
		}
		reduced, err := reduce(unrelated, test)
		if err != nil {
			t.Fatal(err)
		}
		if names := strings.Join(nodesOf(t, reduced), ","); names != "x,y" {
			t.Errorf("reduced input has the nodes %s, want x,y:\n%s", names, reduced)
		}
		if strings.Contains(reduced, "color") || strings.Contains(reduced, "subgraph") {
			t.Errorf("reduced input keeps an attribute or a subgraph:\n%s", reduced)
		}
	})

	t.Run("Another difference", func(t *testing.T) {
		// With the edge s -> t, node a differs; without it, node c differs
		// instead, which does not differ in the original.
		src := "digraph { a; c; s -> t; n1 -> n2 }\n"
		ref, got := fakeLayouts(t, func(src string) []string {
			if hasEdge(t, src, "s", "t") {
				return []string{"a"}
			}
			return []string{"c"}
		})
		test, failing, err := predicate(src, ref, got)
		if err != nil || strings.Join(failing, ",") != "node a" {
			t.Fatalf("failing objects %q, error %v", failing, err)
		}
		reduced, err := reduce(src, test)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(reduced, `"s" -> "t"`) || strings.Contains(reduced, `"n1"`) {
			t.Errorf("reduced input should keep s -> t and drop n1 -> n2:\n%s", reduced)
		}
	})

	t.Run("Input that passes the proxy", func(t *testing.T) {
		ref, got := fakeLayouts(t, func(string) []string { return nil })
		if _, _, err := predicate(unrelated, ref, got); err == nil {
			t.Error("an input that passes the proxy gives no error")
		}
	})
}

func TestDdmin(t *testing.T) {
	// The subset that contains 3 and 7 is found, and nothing else is kept.
	test := func(s []int) bool {
		has := map[int]bool{}
		for _, i := range s {
			has[i] = true
		}
		return has[3] && has[7]
	}
	if got := ddmin(indexes(20), test); len(got) != 2 || got[0] != 3 || got[1] != 7 {
		t.Errorf("ddmin = %v, want [3 7]", got)
	}
}
