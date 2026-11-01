// SPDX-License-Identifier: MIT
// Provenance: original

package output

import (
	"context"
	"regexp"
	"strconv"
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

func TestSVGEdgeLabels(t *testing.T) {
	svg := render(t, `digraph { a -> b [label="x"] }`, SVG)
	if !regexp.MustCompile(`(?s)<g id="edge1" class="edge">.*<polygon fill="black"[^>]*/>\n<text [^>]*>x</text>\n</g>`).MatchString(svg) {
		t.Errorf("edge label after the arrowhead:\n%s", svg)
	}
	svg = render(t, `digraph { a -> b [label=<<TABLE BGCOLOR="#000009" FIXEDSIZE="TRUE" WIDTH="57" HEIGHT="17"><TR><TD></TD></TR></TABLE>>] }`, SVG)
	m := regexp.MustCompile(`<polygon fill="#000009" stroke="none" points="([-\d.]+),([-\d.]+) [-\d.]+,([-\d.]+) ([-\d.]+),[-\d.]+ [^"]*"/>\n<polygon fill="none" stroke="black" points="[^"]+"/>\n<polygon fill="none" stroke="black" points="[^"]+"/>`).FindStringSubmatch(svg)
	if m == nil {
		t.Fatalf("table label polygons:\n%s", svg)
	}
	num := func(s string) float64 { v, _ := strconv.ParseFloat(s, 64); return v }
	if w, h := num(m[4])-num(m[1]), num(m[2])-num(m[3]); w != 57 || h != 17 {
		t.Errorf("table box %v x %v, want 57 x 17", w, h)
	}
}

func TestPlainEdgeLabel(t *testing.T) {
	plain := render(t, `digraph { a -> b [label="x"] }`, Plain)
	if !regexp.MustCompile(`(?m)^edge a b 4( [-\d.]+){8} x [-\d.]+ [-\d.]+ solid black$`).MatchString(plain) {
		t.Errorf("edge line:\n%s", plain)
	}
}

// TestSVGTables checks the table drawings against the reference's: the
// order of the polygons, their colors, stroke widths and insets.
func TestSVGTables(t *testing.T) {
	for _, tt := range []struct {
		src  string
		want []string
	}{
		{`digraph { a [shape=plaintext, label=<<TABLE BGCOLOR="#000009" COLOR="#00000F"><TR><TD BGCOLOR="#000001" COLOR="#00000E"></TD><TD BGCOLOR="#000002"></TD></TR></TABLE>>] }`, []string{
			`<polygon fill="#000009" stroke="none" points="17,-12 17,-24 37,-24 37,-12 17,-12"/>`,
			`<polygon fill="none" stroke="#00000f" points="17,-12 17,-24 37,-24 37,-12 17,-12"/>`,
			`<polygon fill="#000001" stroke="none" points="20,-15 20,-21 26,-21 26,-15 20,-15"/>`,
			`<polygon fill="none" stroke="#00000e" points="20,-15 20,-21 26,-21 26,-15 20,-15"/>`,
			`<polygon fill="#000002" stroke="none" points="28,-15 28,-21 34,-21 34,-15 28,-15"/>`,
			`<polygon fill="none" stroke="#00000f" points="28,-15 28,-21 34,-21 34,-15 28,-15"/>`,
		}},
		{`digraph { a [shape=plaintext, label=<<TABLE BORDER="2" CELLBORDER="3" CELLSPACING="4" CELLPADDING="5"><TR><TD></TD><TD WIDTH="30"></TD></TR></TABLE>>] }`, []string{
			`<polygon fill="none" stroke="black" stroke-width="2" points="9,-5 9,-31 69,-31 69,-5 9,-5"/>`,
			`<polygon fill="none" stroke="black" stroke-width="3" points="15.5,-11.5 15.5,-24.5 28.5,-24.5 28.5,-11.5 15.5,-11.5"/>`,
			`<polygon fill="none" stroke="black" stroke-width="3" points="35.5,-11.5 35.5,-24.5 62.5,-24.5 62.5,-11.5 35.5,-11.5"/>`,
		}},
		{`digraph { a [shape=plaintext, label=<<TABLE BORDER="0" CELLBORDER="0"><TR><TD></TD><TD BGCOLOR="#000006"></TD><TD></TD></TR></TABLE>>] }`, []string{
			`<polygon fill="#000006" stroke="none" points="25,-16 25,-20 29,-20 29,-16 25,-16"/>`,
		}},
	} {
		svg := render(t, tt.src, SVG)
		node := svg[strings.Index(svg, `<g id="node1"`):]
		var got []string
		for _, line := range strings.Split(node, "\n") {
			if strings.HasPrefix(line, "<polygon") {
				got = append(got, line)
			}
		}
		if strings.Join(got, "\n") != strings.Join(tt.want, "\n") {
			t.Errorf("%s: polygons\n%s\nwant\n%s", tt.src, strings.Join(got, "\n"), strings.Join(tt.want, "\n"))
		}
	}
}

func TestSVGEdgeTitleWithPorts(t *testing.T) {
	svg := render(t, `digraph { a [shape=rect]; a:P -> b }`, SVG)
	if !strings.Contains(svg, "<!-- a&#45;&gt;b -->\n<g id=\"edge1\" class=\"edge\">\n<title>a:P&#45;&gt;b</title>") {
		t.Errorf("edge title with a port:\n%s", svg)
	}
	svg = render(t, `digraph { a [shape=plaintext, label=<<TABLE><TR><TD PORT="p"></TD></TR></TABLE>>]; b -> a [headport=p] }`, SVG)
	if !strings.Contains(svg, "<title>b&#45;&gt;a:p</title>") {
		t.Errorf("edge title with a head port:\n%s", svg)
	}
}

func TestPlainExtAndInlineSVG(t *testing.T) {
	// Port on an edge line.
	src := `digraph { a [shape=plaintext, label=<<TABLE><TR><TD PORT="p"></TD></TR></TABLE>>]; a:p -> b }`
	plain, ext := render(t, src, Plain), render(t, src, PlainExt)
	if !strings.Contains(ext, "\nedge a:p b ") || strings.Replace(ext, "edge a:p b ", "edge a b ", 1) != plain {
		t.Errorf("plain-ext:\n%s\nplain:\n%s", ext, plain)
	}

	// No prolog.
	svg, inline := render(t, `digraph G { a }`, SVG), render(t, `digraph G { a }`, SVGInline)
	if !strings.HasPrefix(inline, "<!-- Generated by graphviz version") || !strings.Contains(inline, `<svg width="62pt" height="44pt"`) || strings.Contains(inline, "xmlns") {
		t.Errorf("svg_inline:\n%s", inline)
	}
	_, rest, _ := strings.Cut(svg, "<!-- Generated by graphviz version")
	rest = strings.Replace(rest, ` xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"`, "", 1)
	if inline != "<!-- Generated by graphviz version"+rest {
		t.Errorf("svg_inline differs from svg beyond its prolog:\n%s", inline)
	}
}
