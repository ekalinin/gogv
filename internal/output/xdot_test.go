// SPDX-License-Identifier: MIT
// Provenance: original

package output

import (
	"regexp"
	"strings"
	"testing"

	"github.com/ekalinin/gogv/internal/layout"
)

func xdotOf(version string) func(*layout.Result) []byte {
	return func(res *layout.Result) []byte { return Xdot(res, version) }
}

func TestXdotScenarios(t *testing.T) {
	// Node in xdot: the reference's output, byte for byte.
	want := `digraph {
	graph [_draw_="c 9 -#fffffe00 C 7 -#ffffff P 4 0 0 0 36 54 36 54 0 ",
		bb="0,0,54,36",
		xdotversion=1.7
	];
	node [label="\N"];
	a	[_draw_="c 7 -#000000 e 27 18 27 18 ",
		_ldraw_="F 14 11 -Times-Roman c 7 -#000000 T 27 13.8 0 6.21 1 -a ",
		height=0.5,
		pos="27,18",
		width=0.75];
}
`
	if got := render(t, `digraph { a }`, xdotOf("1.7")); got != want {
		t.Errorf("node in xdot:\n%s\nwant\n%s", got, want)
	}

	// Edge in xdot.
	x := render(t, `digraph { a -> b [label="x"] }`, xdotOf("1.7"))
	edge := attrsOf(t, x, "a -> b")
	for _, want := range []string{`_draw_="c 7 -#000000 B 4 `, `_hdraw_="S 5 -solid c 7 -#000000 C 7 -#000000 P 3 `, `_ldraw_="F 14 11 -Times-Roman c 7 -#000000 T `, ` 1 -x "`} {
		if !strings.Contains(edge, want) {
			t.Errorf("edge in xdot lacks %q: %s", want, edge)
		}
	}

	// Table label in xdot: the table's border, the first cell's fill and
	// border, the second cell's border.
	x = render(t, `digraph { a [shape=plaintext, label=<<TABLE><TR><TD PORT="p" BGCOLOR="#000001"></TD><TD WIDTH="20"></TD></TR></TABLE>>] }`, xdotOf("1.7"))
	// Long strings are broken with a backslash and a newline.
	node := strings.ReplaceAll(attrsOf(t, x, "a"), "\\\n", "")
	ops := []string{"S 5 -solid c 7 -#000000 p 4 ", "S 5 -solid c 9 -#fffffe00 C 7 -#000001 P 4 ", "S 5 -solid c 7 -#000000 p 4 ", "S 5 -solid c 7 -#000000 p 4 "}
	rest := node
	for _, op := range ops {
		i := strings.Index(rest, op)
		if i < 0 {
			t.Fatalf("table label in xdot lacks %q in order: %s", op, node)
		}
		rest = rest[i+len(op):]
	}
	if strings.Contains(node, "_draw_") {
		t.Errorf("a plaintext node has an outline: %s", node)
	}

	// Versions.
	src := `digraph { a -> b [label="x"]; subgraph cluster_x { c } }`
	full := render(t, src, xdotOf("1.7"))
	for _, v := range []string{"1.2", "1.4"} {
		if got := render(t, src, xdotOf(v)); got != strings.Replace(full, "xdotversion=1.7", "xdotversion="+v, 1) {
			t.Errorf("xdot%s differs from xdot beyond its version:\n%s", v, got)
		}
	}

	// Pen widths and explicit styles.
	x = render(t, `digraph { a [shape=plaintext, label=<<TABLE BORDER="2" CELLBORDER="3"><TR><TD></TD><TD BORDER="1"></TD></TR></TABLE>>]; b [style=solid] }`, xdotOf("1.7"))
	if node := strings.ReplaceAll(attrsOf(t, x, "a"), "\\\n", ""); !strings.Contains(node, "S 15 -setlinewidth(2) S 5 -solid c 7 -#000000 p 4 ") ||
		!strings.Contains(node, "S 15 -setlinewidth(3) S 5 -solid") || !strings.Contains(node, "S 15 -setlinewidth(1) S 5 -solid") {
		t.Errorf("pen widths: %s", node)
	}
	if node := attrsOf(t, x, "b"); !strings.Contains(node, `_draw_="S 5 -solid c 7 -#000000 e `) {
		t.Errorf("explicit style: %s", node)
	}
}

func TestXdotPprofScenarios(t *testing.T) {
	// Filled node in xdot.
	if x := render(t, `digraph { a [style=filled] }`, xdotOf("1.7")); !strings.Contains(x, `_draw_="c 7 -#000000 C 7 -#d3d3d3 E 27 18 27 18 "`) {
		t.Errorf("filled node in xdot:\n%s", x)
	}
	// Dotted and wide edges in xdot.
	x := render(t, `digraph { a -> b [style=dotted]; c -> d [penwidth=3] }`, xdotOf("1.7"))
	dotted, wide := attrsOf(t, x, "a -> b"), attrsOf(t, x, "c -> d")
	for _, tt := range []struct{ attrs, want string }{
		{dotted, `_draw_="S 6 -dotted c 7 -#000000 B 4 `},
		{dotted, `_hdraw_="S 5 -solid c 7 -#000000 C 7 -#000000 P 3 `},
		{wide, `_draw_="S 15 -setlinewidth(3) c 7 -#000000 B 4 `},
		{wide, `_hdraw_="S 15 -setlinewidth(3) S 5 -solid c 7 -#000000 C 7 -#000000 P 3 `},
	} {
		if !strings.Contains(tt.attrs, tt.want) {
			t.Errorf("dotted and wide edges in xdot: no %s in %s", tt.want, tt.attrs)
		}
	}
	// box3d draws its depth as three lines, each with its pen.
	x = strings.ReplaceAll(render(t, `digraph { a [shape=box3d, style=filled] }`, xdotOf("1.7")), "\\\n", "")
	if !regexp.MustCompile(`_draw_="c 7 -#000000 C 7 -#d3d3d3 P 6 [-\d. ]+(c 7 -#000000 L 2 [-\d. ]+){3}"`).MatchString(x) {
		t.Errorf("box3d in xdot:\n%s", x)
	}
}

func TestPlainStyles(t *testing.T) {
	plain := render(t, `digraph { a [style=filled, fillcolor="#00FF00"]; a -> b [style=dotted] }`, Plain)
	if !regexp.MustCompile(`(?m)^node a .* filled ellipse black #00FF00$`).MatchString(plain) || !regexp.MustCompile(`(?m)^edge a b .* dotted black$`).MatchString(plain) {
		t.Errorf("styles in plain:\n%s", plain)
	}
}
