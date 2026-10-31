// SPDX-License-Identifier: MIT
// Provenance: original

package layout

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/ekalinin/gogv/internal/attrs"
	"github.com/ekalinin/gogv/internal/conformance"
	"github.com/ekalinin/gogv/internal/graph"
)

// layoutResolved lays out src after resolving its clusters, as Render does.
func layoutResolved(t *testing.T, src string) *Result {
	t.Helper()
	g := buildGraph(t, src)
	g.ResolveClusters()
	res, err := Layout(context.Background(), g)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func nodesOf(res *Result) map[string]NodeResult {
	out := map[string]NodeResult{}
	for _, n := range res.Nodes {
		out[n.Node.Name] = n
	}
	return out
}

func clusterOf(t *testing.T, res *Result, name string) ClusterResult {
	t.Helper()
	for _, c := range res.Clusters {
		if c.Graph.Name == name {
			return c
		}
	}
	t.Fatalf("no cluster %q", name)
	return ClusterResult{}
}

// rankOf numbers the distinct y of the nodes from the top.
func rankOf(res *Result) map[string]int {
	var ys []float64
	for _, n := range res.Nodes {
		found := false
		for _, y := range ys {
			if approx(y, n.Y) {
				found = true
			}
		}
		if !found {
			ys = append(ys, n.Y)
		}
	}
	out := map[string]int{}
	for _, n := range res.Nodes {
		for _, y := range ys {
			if y > n.Y+0.01 {
				out[n.Node.Name]++
			}
		}
	}
	return out
}

// The rank constraint scenarios of the dot-layout spec.
func TestRankConstraintScenarios(t *testing.T) {
	tests := []struct {
		src  string
		want map[string]int
	}{
		{"digraph { a -> b -> c; { rank=same; a; c } }", map[string]int{"a": 0, "c": 0, "b": 1}},
		{"digraph { a -> b -> c; d -> e; { rank=min; d } }", map[string]int{"a": 0, "d": 0, "e": 1, "c": 2}},
		{"digraph { a -> b -> c; d -> e; { rank=source; d } }", map[string]int{"d": 0, "a": 1, "e": 1, "c": 3}},
		{"digraph { a -> b -> c; d; { rank=max; d } }", map[string]int{"a": 0, "c": 2, "d": 2}},
		{"digraph { a -> b -> c; d; { rank=sink; d } }", map[string]int{"a": 0, "c": 2, "d": 3}},
		{"digraph { x -> a; subgraph cluster_c { a -> b -> c; d; { rank=source; d } } }", map[string]int{"x": 0, "d": 0, "a": 1, "c": 3}},
		{"digraph { subgraph cluster_x { rank=same; a -> b } }", map[string]int{"a": 0, "b": 1}},
		{"digraph { rank=same; a -> b }", map[string]int{"a": 0, "b": 1}},
		{"digraph { rank=same; subgraph s { a -> b } }", map[string]int{"a": 0, "b": 0}},
	}
	for _, tt := range tests {
		got := rankOf(layoutResolved(t, tt.src))
		for name, r := range tt.want {
			if got[name] != r {
				t.Errorf("%s: %s on rank %d, want %d (%v)", tt.src, name, got[name], r, got)
			}
		}
	}
}

func TestClusterRankedAsABlock(t *testing.T) {
	got := rankOf(layoutResolved(t, "digraph { subgraph cluster_a { a; b } a -> c -> d -> b }"))
	if got["a"] != 0 || got["b"] != 0 || got["c"] != 1 || got["d"] != 2 {
		t.Errorf("ranks %v", got)
	}
	// A node that a rank subgraph takes out of a cluster is ranked as if it
	// were still in it, as on the reference.
	got = rankOf(layoutResolved(t, "digraph { subgraph cluster_x { a; b } a -> c; c -> d; { rank=same; b; d } }"))
	if got["a"] != 0 || got["b"] != 0 || got["d"] != 0 || got["c"] != 1 {
		t.Errorf("rank set across a cluster: ranks %v", got)
	}
}

func inside(c ClusterResult, x, y float64) bool {
	return x > c.X0 && x < c.X1 && y > c.Y0 && y < c.Y1
}

func contains(r, s ClusterResult) bool {
	return s.X0 >= r.X0 && s.X1 <= r.X1 && s.Y0 >= r.Y0 && s.Y1 <= r.Y1
}

func TestClusterBoxScenarios(t *testing.T) {
	res := layoutResolved(t, "digraph { subgraph cluster_a { a } }")
	if c := clusterOf(t, res, "cluster_a"); !approx(c.X1-c.X0, 70) || !approx(c.Y1-c.Y0, 52) {
		t.Errorf("margin: box %v x %v", c.X1-c.X0, c.Y1-c.Y0)
	}

	res = layoutResolved(t, "digraph { subgraph cluster_a { subgraph cluster_b { a } c } }")
	a, b := clusterOf(t, res, "cluster_a"), clusterOf(t, res, "cluster_b")
	n := nodesOf(res)
	if !contains(a, b) || inside(b, n["c"].X, n["c"].Y) {
		t.Errorf("nested: a %v, b %v, c at %v,%v", a, b, n["c"].X, n["c"].Y)
	}

	res = layoutResolved(t, "digraph { subgraph cluster_a { a } subgraph cluster_b { b } }")
	a, b = clusterOf(t, res, "cluster_a"), clusterOf(t, res, "cluster_b")
	if a.X1 > b.X0 && b.X1 > a.X0 && a.Y1 > b.Y0 && b.Y1 > a.Y0 {
		t.Errorf("siblings overlap: %v %v", a, b)
	}

	res = layoutResolved(t, `digraph { subgraph cluster_a { color="#FF0000"; a } }`)
	if c := clusterOf(t, res, "cluster_a"); c.Color != "#FF0000" {
		t.Errorf("color %q", c.Color)
	}
	res = layoutResolved(t, "digraph { subgraph cluster_a { a } }")
	if c := clusterOf(t, res, "cluster_a"); c.Color != "black" {
		t.Errorf("default color %q", c.Color)
	}
}

func TestClusterLabelScenarios(t *testing.T) {
	res := layoutResolved(t, `digraph { subgraph cluster_a { label="Group"; a } }`)
	c, n := clusterOf(t, res, "cluster_a"), nodesOf(res)
	if len(c.Label.Lines) != 1 || c.Label.Lines[0].Text != "Group" || !approx(c.LabelPos.X, (c.X0+c.X1)/2) ||
		!(c.LabelPos.Y > n["a"].Y+n["a"].H/2) || !inside(c, c.LabelPos.X, c.LabelPos.Y) {
		t.Errorf("label at the top: %+v, a at %v", c, n["a"].Y)
	}

	res = layoutResolved(t, `digraph { subgraph cluster_a { label="T"; labeljust=l; a } }`)
	c = clusterOf(t, res, "cluster_a")
	if !approx(c.LabelPos.X-c.X0, (c.Label.Width+labelMarginX)/2) {
		t.Errorf("label at the left: at %v in [%v, %v]", c.LabelPos.X, c.X0, c.X1)
	}
	res = layoutResolved(t, `digraph { subgraph cluster_a { label="T"; labeljust=r; a } }`)
	c = clusterOf(t, res, "cluster_a")
	if !approx(c.X1-c.LabelPos.X, (c.Label.Width+labelMarginX)/2) {
		t.Errorf("label at the right: at %v in [%v, %v]", c.LabelPos.X, c.X0, c.X1)
	}

	res = layoutResolved(t, `digraph { subgraph cluster_a { label="Long cluster title"; a } }`)
	if c = clusterOf(t, res, "cluster_a"); c.X1-c.X0 < c.Label.Width {
		t.Errorf("wide label: box %v, label %v", c.X1-c.X0, c.Label.Width)
	}

	res = layoutResolved(t, `digraph { rankdir=LR; subgraph cluster_a { label="T"; a -> b } }`)
	c, n = clusterOf(t, res, "cluster_a"), nodesOf(res)
	if !(c.LabelPos.Y > n["a"].Y+n["a"].H/2) || !inside(c, c.LabelPos.X, c.LabelPos.Y) {
		t.Errorf("label in rankdir=LR: %+v, a at %v", c, n["a"].Y)
	}

	// An HTML label is not supported yet: the cluster has none.
	res = layoutResolved(t, "digraph { subgraph cluster_x { label=<<TABLE><TR><TD>x</TD></TR></TABLE>>; a } }")
	if c = clusterOf(t, res, "cluster_x"); len(c.Label.Lines) != 0 {
		t.Errorf("HTML label drawn: %+v", c.Label)
	}
}

func TestGraphLabelJust(t *testing.T) {
	for just, want := range map[string]func(res *Result) float64{
		"c": func(res *Result) float64 { return res.Width / 2 },
		"l": func(res *Result) float64 { return (res.Label.Width + labelMarginX) / 2 },
		"r": func(res *Result) float64 { return res.Width - (res.Label.Width+labelMarginX)/2 },
	} {
		res := layoutResolved(t, `digraph { label="Title"; labeljust=`+just+`; a -> b; a -> c }`)
		if !approx(res.LabelPos.X, want(res)) {
			t.Errorf("labeljust=%s: label at %v, want %v", just, res.LabelPos.X, want(res))
		}
	}
}

// curveInside reports whether a sampled point of an edge lies inside a box.
func curveInside(e EdgeResult, c ClusterResult) bool {
	for _, seg := range segments(e.Points) {
		for i := 0; i <= 20; i++ {
			p := cubic(seg, float64(i)/20)
			if inside(c, p.X, p.Y) {
				return true
			}
		}
	}
	return false
}

func TestEdgeAroundCluster(t *testing.T) {
	res := layoutResolved(t, "digraph { a -> d; a -> b; b -> d; subgraph cluster_x { b } }")
	c := clusterOf(t, res, "cluster_x")
	for _, e := range res.Edges {
		if e.Edge.Tail.Name == "a" && e.Edge.Head.Name == "d" && curveInside(e, c) {
			t.Errorf("a -> d crosses the box of cluster_x: %v", e.Points)
		}
	}
}

// flat reports whether an edge joins two nodes at the same height.
func flat(res *Result, e EdgeResult) bool {
	n := nodesOf(res)
	return approx(n[e.Edge.Tail.Name].Y, n[e.Edge.Head.Name].Y)
}

// memberOf reports whether the drawn cluster c has node n.
func memberOf(c ClusterResult, n *graph.Node) bool { return c.Graph.HasNode(n) }

// clusterCases are the cases with clusters and no feature that is not
// supported yet that the boxes are checked on: the curated ones and, when
// they have been collected, the variant of gopuml's.
func clusterCases(t *testing.T) []conformance.Case {
	var out []conformance.Case
	for _, dir := range []string{"../../testdata/conformance/cases", "../../.cache/gopuml-cluster-dot"} {
		cases, err := conformance.LoadCases(dir)
		if err != nil {
			continue
		}
		for _, c := range cases {
			if !regexp.MustCompile(`(?i)subgraph\s+cluster`).Match(c.Source) {
				continue
			}
			if len(attrs.Check(buildGraph(t, string(c.Source)))) == 0 {
				out = append(out, c)
			}
		}
	}
	return out
}

// No node outside a cluster lies in its box, so the nodes of each cluster
// are next to each other on each of its ranks; no edge between two ranks
// with no end in a cluster crosses its box. A flat edge can: the reference
// draws one across the box of a cluster that holds its rank, too.
func TestClustersKeepOthersOut(t *testing.T) {
	for _, cs := range clusterCases(t) {
		res := layoutResolved(t, string(cs.Source))
		for _, c := range res.Clusters {
			for _, n := range res.Nodes {
				if !memberOf(c, n.Node) && inside(c, n.X, n.Y) {
					t.Errorf("%s: node %s inside the box of %s", cs.ID, n.Node.Name, c.Graph.Name)
				}
			}
			for _, e := range res.Edges {
				if e.Invisible || memberOf(c, e.Edge.Tail) || memberOf(c, e.Edge.Head) || flat(res, e) {
					continue
				}
				if curveInside(e, c) {
					t.Errorf("%s: edge %s->%s crosses the box of %s", cs.ID, e.Edge.Tail.Name, e.Edge.Head.Name, c.Graph.Name)
				}
			}
		}
	}
}

// The box heights of the curated cluster cases are within 5% of the
// reference's.
func TestClusterBoxHeights(t *testing.T) {
	poly := regexp.MustCompile(`<g id="clust\d+" class="cluster">\n<title>(.*?)</title>\n<polygon fill="none" stroke="[^"]*" points="([^"]+)"`)
	for _, cs := range clusterCases(t) {
		if !strings.HasPrefix(cs.ID, "dot-") {
			continue
		}
		ref, err := os.ReadFile(filepath.Join("../../testdata/conformance/reference", cs.ID+".svg"))
		if err != nil {
			t.Fatal(err)
		}
		res := layoutResolved(t, string(cs.Source))
		for _, m := range poly.FindAllStringSubmatch(string(ref), -1) {
			pts := strings.Fields(m[2])
			_, y0, _ := strings.Cut(pts[0], ",")
			_, y1, _ := strings.Cut(pts[1], ",")
			a, _ := strconv.ParseFloat(y0, 64)
			b, _ := strconv.ParseFloat(y1, 64)
			want := a - b
			c := clusterOf(t, res, m[1])
			if got := c.Y1 - c.Y0; got < want*0.95 || got > want*1.05 {
				t.Errorf("%s: %s is %v high, the reference's %v", cs.ID, m[1], got, want)
			}
		}
	}
}
