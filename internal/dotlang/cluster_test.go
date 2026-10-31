// SPDX-License-Identifier: MIT
// Provenance: doc https://graphviz.org/doc/info/lang.html
// Provenance: observed dot 16.1.0

package dotlang

import (
	"slices"
	"testing"

	"github.com/ekalinin/gogv/internal/graph"
)

// resolved builds src and resolves its clusters.
func resolved(t *testing.T, src string) (*graph.Graph, []string) {
	t.Helper()
	g := build(t, src)
	return g, g.ResolveClusters()
}

func subgraph(t *testing.T, g *graph.Graph, name string) *graph.Graph {
	t.Helper()
	for _, sg := range g.Graphs() {
		if sg.Name == name {
			return sg
		}
	}
	t.Fatalf("no subgraph %q", name)
	return nil
}

func nodeNames(g *graph.Graph) []string {
	var names []string
	for _, n := range g.Nodes() {
		names = append(names, n.Name)
	}
	return names
}

func TestClusterNames(t *testing.T) {
	g := build(t, "digraph { subgraph Cluster_x { a } subgraph CLUSTERy { b } subgraph clu { c } subgraph s { d } }")
	for name, want := range map[string]bool{"Cluster_x": true, "CLUSTERy": true, "clu": false, "s": false} {
		if got := subgraph(t, g, name).IsCluster(); got != want {
			t.Errorf("%s: IsCluster() = %v, want %v", name, got, want)
		}
	}
	if g.IsCluster() {
		t.Error("the root graph is a cluster")
	}
}

func TestNodeInTwoSiblingClusters(t *testing.T) {
	g, w := resolved(t, "digraph { subgraph cluster_a { a } subgraph cluster_b { a b } }")
	if got := nodeNames(subgraph(t, g, "cluster_a")); !slices.Equal(got, []string{"a"}) {
		t.Errorf("cluster_a: %q", got)
	}
	if got := nodeNames(subgraph(t, g, "cluster_b")); !slices.Equal(got, []string{"b"}) {
		t.Errorf("cluster_b: %q", got)
	}
	if len(w) != 0 {
		t.Errorf("warnings %q", w)
	}
}

func TestNodeLeavesTheSubgraphsOfALaterCluster(t *testing.T) {
	g, _ := resolved(t, "digraph { subgraph cluster_a { a } subgraph cluster_b { subgraph cluster_c { a } subgraph s { a } b } subgraph t { subgraph cluster_d { a } } }")
	for name, want := range map[string][]string{"cluster_a": {"a"}, "cluster_b": {"b"}, "cluster_c": nil, "s": nil, "cluster_d": nil, "t": {"a"}} {
		if got := nodeNames(subgraph(t, g, name)); !slices.Equal(got, want) {
			t.Errorf("%s: %q, want %q", name, got, want)
		}
	}
}

func TestNodeInNestedClusters(t *testing.T) {
	g, _ := resolved(t, "digraph { subgraph cluster_a { subgraph cluster_b { a } a c } }")
	if got := nodeNames(subgraph(t, g, "cluster_b")); !slices.Equal(got, []string{"a"}) {
		t.Errorf("cluster_b: %q", got)
	}
	if got := nodeNames(subgraph(t, g, "cluster_a")); !slices.Equal(got, []string{"a", "c"}) {
		t.Errorf("cluster_a: %q", got)
	}
}

func TestClusterWithoutNodes(t *testing.T) {
	g, _ := resolved(t, "digraph { subgraph cluster_e { } a }")
	if got := subgraph(t, g, "cluster_e").Nodes(); len(got) != 0 {
		t.Errorf("cluster_e has %d nodes", len(got))
	}
}

func TestRankSetAcrossACluster(t *testing.T) {
	g, w := resolved(t, "digraph { subgraph cluster_x { a -> b; b -> c } c; { rank=same; b; c } }")
	want := []string{
		"b was already in a rankset, deleted from cluster cluster_x",
		"c was already in a rankset, deleted from cluster cluster_x",
	}
	if !slices.Equal(w, want) {
		t.Errorf("warnings %q, want %q", w, want)
	}
	x := subgraph(t, g, "cluster_x")
	if got := nodeNames(x); !slices.Equal(got, []string{"a"}) {
		t.Errorf("cluster_x: %q", got)
	}
	if got := edgeList(x); got != "" {
		t.Errorf("cluster_x edges: %q", got)
	}
}

func TestRankSetInsideItsCluster(t *testing.T) {
	g, w := resolved(t, "digraph { subgraph cluster_x { a; b; { rank=same; a; b } } }")
	if len(w) != 0 {
		t.Errorf("warnings %q", w)
	}
	if got := nodeNames(subgraph(t, g, "cluster_x")); !slices.Equal(got, []string{"a", "b"}) {
		t.Errorf("cluster_x: %q", got)
	}
}

func TestEdgeInItsInnermostCluster(t *testing.T) {
	g, _ := resolved(t, "digraph { subgraph cluster_a { subgraph s { subgraph cluster_b { a b } } c; subgraph t { d } } a -> b; a -> c; c -> d; c -> e }")
	for name, want := range map[string]string{"cluster_b": "a->b", "s": "a->b", "cluster_a": "a->b a->c c->d", "t": ""} {
		if got := edgeList(subgraph(t, g, name)); got != want {
			t.Errorf("%s: edges %q, want %q", name, got, want)
		}
	}
}

func TestRankValues(t *testing.T) {
	g := build(t, "digraph { rank=same; subgraph s { a } subgraph cluster_x { b } { rank=sink; c } { rank=foo; d } }")
	for name, want := range map[string]string{"s": "same", "cluster_x": "", "": ""} {
		sg := g
		if name != "" {
			sg = subgraph(t, g, name)
		}
		if got := sg.Rank(); got != want {
			t.Errorf("%q: Rank() = %q, want %q", name, got, want)
		}
	}
	var ranks []string
	for _, sg := range g.Graphs()[1:] {
		if sg.Anonymous {
			ranks = append(ranks, sg.Rank())
		}
	}
	if !slices.Equal(ranks, []string{"sink", ""}) {
		t.Errorf("anonymous subgraphs: %q", ranks)
	}
}

func TestGraphLabelReceivedByACluster(t *testing.T) {
	g := build(t, `digraph { label="G"; subgraph cluster_a { a } subgraph cluster_b { label="B"; b } }`)
	for name, want := range map[string]string{"cluster_a": "G", "cluster_b": "B"} {
		if v, _ := subgraph(t, g, name).Attr("label"); v.S != want {
			t.Errorf("%s: label %q, want %q", name, v.S, want)
		}
	}
}
