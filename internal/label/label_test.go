// SPDX-License-Identifier: MIT
// Provenance: observed dot 16.1.0

package label

import (
	"reflect"
	"testing"
)

func text(lines []Line) []string {
	var out []string
	for _, l := range lines {
		out = append(out, l.Text)
	}
	return out
}

// The expected texts are those the reference draws for the same labels.
func TestExpandMatchesReference(t *testing.T) {
	node := Object{Graph: "G", Node: "a"}
	edge := Object{Graph: "G", Edge: true, Tail: "a", Head: "b", DirectedEdgeName: true}
	graph := Object{Graph: "G"}
	tests := []struct {
		obj  Object
		src  string
		want string
	}{
		{node, `\N in \G`, "a in G"},
		{node, `\L|\E|\T|\H`, "L||T|H"},
		{edge, `\E \T \H \G \N`, "a->b a b G N"},
		{Object{Graph: "U", Edge: true, Tail: "a", Head: "b"}, `\E`, "a--b"},
		{edge, `\L|\N`, "L|N"},
		{graph, `\N|\G|\E|\T|\H|\L|end\\`, `N|G||T|H|L|end\`},
		{Object{Graph: "cluster_s"}, `\G`, "cluster_s"},
		{node, `x\\y \x \q end`, `x\y x q end`},
		{node, `x\\`, `x\`},
		{node, `\\N`, `\N`},
	}
	for _, tt := range tests {
		lines := Lines(Expand(tt.src, tt.obj))
		var got string
		for i, l := range lines {
			if i > 0 {
				got += "/"
			}
			got += l.Text
		}
		if got != tt.want {
			t.Errorf("%q for %+v: %q, want %q", tt.src, tt.obj, got, tt.want)
		}
	}
}

func TestLines(t *testing.T) {
	tests := []struct {
		src  string
		want []Line
	}{
		{`left\lright\rcenter\n`, []Line{{"left", Left}, {"right", Right}, {"center", Center}}},
		{`two\nlines`, []Line{{"two", Center}, {"lines", Center}}},
		{`tail\l`, []Line{{"tail", Left}}},
		{`trail\n\n`, []Line{{"trail", Center}, {"", Center}}},
		{"real\nnewline", []Line{{"real", Center}, {"newline", Center}}},
		{``, nil},
	}
	for _, tt := range tests {
		if got := Lines(tt.src); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Lines(%q) = %+v, want %+v", tt.src, got, tt.want)
		}
	}
}

func TestDefaultNodeLabel(t *testing.T) {
	got := text(Lines(Expand(`\N`, Object{Graph: "G", Node: "a"})))
	if !reflect.DeepEqual(got, []string{"a"}) {
		t.Errorf("default label of a: %q", got)
	}
}
