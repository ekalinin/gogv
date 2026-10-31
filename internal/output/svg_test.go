// SPDX-License-Identifier: MIT
// Provenance: original

package output

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/ekalinin/gogv/internal/dotlang"
	"github.com/ekalinin/gogv/internal/layout"
)

func render(t *testing.T, src string, write func(*layout.Result) []byte) string {
	t.Helper()
	f, _, err := dotlang.Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	for _, d := range f.Graphs {
		res, err := layout.Layout(context.Background(), dotlang.Build(d))
		if err != nil {
			t.Fatal(err)
		}
		out.Write(write(res))
	}
	return out.String()
}

func TestSVGDocument(t *testing.T) {
	svg := render(t, "digraph { a -> b }", SVG)
	for _, want := range []string{
		`<svg width="62pt" height="116pt"`,
		`viewBox="0.00 0.00 62.00 116.00"`,
		`<g id="graph0" class="graph" transform="scale(1 1) rotate(0) translate(4 112)">`,
		`<g id="node1" class="node">` + "\n<title>a</title>",
		`<g id="node2" class="node">` + "\n<title>b</title>",
		`<g id="edge1" class="edge">` + "\n<title>a&#45;&gt;b</title>",
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("SVG lacks %q:\n%s", want, svg)
		}
	}
}

// TestSVGForGopuml reads the output with the patterns of gopuml's
// SvgResult: the size in whole points, node titles, polygon points, path
// data, and the colors that identify objects.
func TestSVGForGopuml(t *testing.T) {
	src := `digraph { a [shape=rect, label="", width=1, height=0.5, color="#00000A"]; b [shape=rect, label="", color="#00000B"]; a -> b [color="#00000C", arrowhead=none] }`
	svg := render(t, src, SVG)
	if !regexp.MustCompile(`(?m)<svg\s+width="(\d+)pt"\s+height="(\d+)pt"`).MatchString(svg) {
		t.Error("no size in whole points")
	}
	for _, want := range []string{"<title>a</title>", "<title>b</title>", `points="`, `d="M`, `stroke="#00000a"`, `stroke="#00000b"`, `stroke="#00000c"`} {
		if !strings.Contains(svg, want) {
			t.Errorf("SVG lacks %q", want)
		}
	}
	path := regexp.MustCompile(`d="M[-0-9.]+,[-0-9.]+C[-0-9., ]+"`)
	if !path.MatchString(svg) {
		t.Error("edge path is not M x,y C ...")
	}
}

func TestSVGDrawing(t *testing.T) {
	svg := render(t, `digraph { a [color="#FF0000", fontcolor="#0000FF"]; b [shape=diamond]; a -> b [arrowhead=empty] }`, SVG)
	for _, want := range []string{
		`<ellipse fill="none" stroke="#ff0000"`,
		`fill="#0000ff">a</text>`,
		`<polygon fill="none" stroke="black" points=`,
		`<path fill="none" stroke="black" d="M`,
	} {
		if !strings.Contains(svg, want) {
			t.Errorf("SVG lacks %q:\n%s", want, svg)
		}
	}
	if strings.Count(svg, `<polygon fill="none"`) != 2 {
		t.Errorf("want the diamond and the empty arrowhead as unfilled polygons:\n%s", svg)
	}
}

func TestPlainLines(t *testing.T) {
	plain := render(t, "digraph { a -> b }", Plain)
	lines := strings.Split(strings.TrimSpace(plain), "\n")
	if len(lines) != 5 || !strings.HasPrefix(lines[0], "graph 1 ") || !strings.HasPrefix(lines[1], "node a ") ||
		!strings.HasPrefix(lines[2], "node b ") || !strings.HasPrefix(lines[3], "edge a b 4 ") || lines[4] != "stop" {
		t.Errorf("plain output:\n%s", plain)
	}
}

func TestSeveralGraphs(t *testing.T) {
	src := "digraph A { a }\ndigraph B { b }\n"
	svg := render(t, src, SVG)
	if strings.Count(svg, "<svg ") != 2 || strings.Index(svg, "<title>A</title>") > strings.Index(svg, "<title>B</title>") {
		t.Errorf("svg of two graphs:\n%s", svg)
	}
	plain := render(t, src, Plain)
	if strings.Count(plain, "stop\n") != 2 || strings.Index(plain, "node a") > strings.Index(plain, "node b") {
		t.Errorf("plain of two graphs:\n%s", plain)
	}
	f, _, _ := dotlang.Parse(src)
	var canon strings.Builder
	for _, d := range f.Graphs {
		canon.Write(Canon(dotlang.Build(d)))
	}
	if !strings.HasPrefix(canon.String(), "digraph A {") || !strings.Contains(canon.String(), "}\ndigraph B {") {
		t.Errorf("canon of two graphs:\n%s", canon.String())
	}
}

func TestSVGClusters(t *testing.T) {
	svg := render(t, "digraph { subgraph cluster_a { subgraph cluster_b { a } } }", SVG)
	ia := strings.Index(svg, `<g id="clust1" class="cluster">`+"\n<title>cluster_a</title>")
	ib := strings.Index(svg, `<g id="clust2" class="cluster">`+"\n<title>cluster_b</title>")
	in := strings.Index(svg, `<g id="node1" class="node">`)
	if ia < 0 || ib < ia || in < ib {
		t.Errorf("cluster groups are not before the node, outer first:\n%s", svg)
	}

	svg = render(t, `digraph { subgraph cluster_a { color="#FF0000"; label="A"; a } }`, SVG)
	box := regexp.MustCompile(`<g id="clust1" class="cluster">\n<title>cluster_a</title>\n<polygon fill="none" stroke="#ff0000" points="[^"]+"/>\n<text [^>]*>A</text>\n</g>`)
	if !box.MatchString(svg) {
		t.Errorf("cluster box:\n%s", svg)
	}
}
