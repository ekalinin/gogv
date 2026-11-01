// SPDX-License-Identifier: MIT
// Provenance: original

package output

import (
	"regexp"
	"strings"
	"testing"

	"github.com/ekalinin/gogv/internal/layout"
)

// attrsOf returns the attribute list of the statement of an object in DOT
// output: the text from its name to the end of its statement.
func attrsOf(t *testing.T, dot, object string) string {
	t.Helper()
	m := regexp.MustCompile(`(?s)\n\t*` + regexp.QuoteMeta(object) + `\t\[(.*?)\];`).FindStringSubmatch(dot)
	if m == nil {
		t.Fatalf("no statement for %s in\n%s", object, dot)
	}
	return m[1]
}

func TestDotScenarios(t *testing.T) {
	// Layout attributes.
	dot := render(t, `digraph G { a -> b [label="x"]; a [shape=rect, color="#FF0000"] }`, Dot)
	if !strings.Contains(dot, "\tgraph [bb=\"0,0,") {
		t.Errorf("no bb on the graph:\n%s", dot)
	}
	for _, n := range []string{"a", "b"} {
		as := attrsOf(t, dot, n)
		for _, want := range []string{"height=", "pos=\"", "width="} {
			if !strings.Contains(as, want) {
				t.Errorf("node %s lacks %s: %s", n, want, as)
			}
		}
	}
	if as := attrsOf(t, dot, "a -> b"); !strings.Contains(as, "lp=\"") || !strings.Contains(as, "pos=\"e,") {
		t.Errorf("edge attributes: %s", as)
	}

	// Clusters and external labels.
	dot = render(t, `digraph G { label="Title"; subgraph cluster_x { label="C"; a } a -> b [headlabel="H"] }`, Dot)
	graphAttrs := regexp.MustCompile(`(?s)^digraph G \{\n\tgraph \[(.*?)\];`).FindStringSubmatch(dot)
	clusterAttrs := regexp.MustCompile(`(?s)subgraph cluster_x \{\n\t\tgraph \[(.*?)\];`).FindStringSubmatch(dot)
	if graphAttrs == nil || clusterAttrs == nil {
		t.Fatalf("no graph attributes:\n%s", dot)
	}
	for _, as := range []string{graphAttrs[1], clusterAttrs[1]} {
		for _, want := range []string{"bb=", "lheight=", "lp=", "lwidth="} {
			if !strings.Contains(as, want) {
				t.Errorf("graph or cluster lacks %s: %s", want, as)
			}
		}
	}
	if as := attrsOf(t, dot, "a -> b"); !strings.Contains(as, "head_lp=") {
		t.Errorf("edge lacks head_lp: %s", as)
	}

	// Numbers.
	for _, tt := range []struct {
		v    float64
		want string
	}{{88.65123, "88.651"}, {124.8, "124.8"}, {36.41, "36.41"}, {0, "0"}, {1.08528, "1.0853"}} {
		if got := plainNumber(tt.v); got != tt.want {
			t.Errorf("plainNumber(%v) = %q, want %q", tt.v, got, tt.want)
		}
	}
	for _, tt := range []struct {
		points float64
		want   string
	}{{16.8, "0.23"}, {26.43, "0.37"}, {36, "0.5"}, {72.72, "1.01"}} {
		if got := labelInches(tt.points).S; got != tt.want {
			t.Errorf("labelInches(%v) = %q, want %q", tt.points, got, tt.want)
		}
	}
}

func TestDotKeepsAnonymousSubgraphs(t *testing.T) {
	// As the reference writes them after the layout: those of the root and
	// of clusters, not those nested in other subgraphs.
	dot := render(t, `digraph { a; {} { {b} } subgraph cluster_x { {c} } }`, Dot)
	if n := strings.Count(dot, "{\n"); n != 5 {
		t.Errorf("%d graphs and subgraphs written, want 5:\n%s", n, dot)
	}
	canon := render(t, `digraph { a; {} { {b} } }`, func(res *layout.Result) []byte { return Canon(res.Graph) })
	if strings.Count(canon, "{\n") != 1 {
		t.Errorf("canon writes anonymous subgraphs:\n%s", canon)
	}
}
