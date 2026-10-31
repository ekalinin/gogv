// SPDX-License-Identifier: MIT
// Provenance: observed dot 16.1.0

package dotlang

import "testing"

// syntaxErrorCases pairs inputs with the first line the reference writes to
// stderr for them, without its "Error: <stdin>: " prefix; an empty want means
// that the reference reads the input without an error.
var syntaxErrorCases = []struct{ src, want string }{
	{"digraph { a -> }\n", "syntax error in line 1 near '}'"},
	{"digraph { a -- b }\n", "syntax error in line 1 near '--'"},
	{"graph { a -> b }\n", "syntax error in line 1 near '->'"},
	{"digraph { node }\n", "syntax error in line 1 near '}'"},
	{"digraph { node; a }\n", "syntax error in line 1 near ';'"},
	{"digraph { a [label] }\n", "syntax error in line 1 near ']'"},
	{"digraph { a [label=] }\n", "syntax error in line 1 near ']'"},
	{"digraph { a [=x] }\n", "syntax error in line 1 near '='"},
	{"digraph { a = }\n", "syntax error in line 1 near '}'"},
	{"digraph { = b }\n", "syntax error in line 1 near '='"},
	{"digraph { a -> b -> }\n", "syntax error in line 1 near '}'"},
	{"digraph { a : }\n", "syntax error in line 1 near '}'"},
	{"digraph { a:b:c:d }\n", "syntax error in line 1 near ':'"},
	{"digraph { subgraph }\n", "syntax error in line 1 near '}'"},
	{"digraph { subgraph s }\n", "syntax error in line 1 near '}'"},
	{"digraph { subgraph s; }\n", "syntax error in line 1 near ';'"},
	{"digraph { { a }\n", "syntax error in line 2"},
	{"digraph { a } }\n", ""},
	{"digraph\n", "syntax error in line 2"},
	{"digraph {\n", "syntax error in line 2"},
	{"strict { a }\n", "syntax error in line 1 near '{'"},
	{"strict strict digraph { a }\n", "syntax error in line 1 near 'strict'"},
	{"digraph a b { c }\n", "syntax error in line 1 near 'b'"},
	{"digraph { \"a\" + b }\n", "syntax error in line 1 near 'b'"},
	{"digraph { \"a\" + }\n", "syntax error in line 1 near '}'"},
	{"digraph { a [color=\"r\" + \"ed\"] }\n", ""},
	{"digraph { a ; ; b }\n", "syntax error in line 1 near ';'"},
	{"digraph { a [x=1,,y=2] }\n", "syntax error in line 1 near ','"},
	{"digraph { a [x=1 y=2] }\n", ""},
	{"digraph { a [x=1][y=2] }\n", ""},
	{"digraph { edge [] }\n", ""},
	{"digraph { a -> {} }\n", ""},
	{"digraph { graph [x=1] y=2 }\n", ""},
	{"digraph { -> b }\n", "syntax error in line 1 near '->'"},
	{"digraph { a\n->\n}\n", "syntax error in line 3 near '}'"},
	{"digraph { \"x\" -> <y> }\n", ""},
	{"digraph { a [label=<<b>x</b>>, k=\"v\"] }\n", ""},
	{"digraph {\n\n a ->\n\n}\n", "syntax error in line 5 near '}'"},
	{"{ a }\n", "syntax error in line 1 near '{'"},
	{"digraph { a -> subgraph }\n", "syntax error in line 1 near '}'"},
	{"digraph { a, b }\n", ""},
	{"digraph { a,, b }\n", "syntax error in line 1 near ','"},
	{"digraph { a; , b }\n", "syntax error in line 1 near ','"},
	{"digraph { a , }\n", "syntax error in line 1 near '}'"},
	{"digraph { , a }\n", "syntax error in line 1 near ','"},
	{"digraph { a -> b, c }\n", ""},
	{"digraph { x=1, y=2 }\n", "syntax error in line 1 near ','"},
	{"digraph { node [shape=box], a }\n", "syntax error in line 1 near ','"},
	{"digraph { subgraph s { a }, b }\n", "syntax error in line 1 near ','"},
}

func TestSyntaxErrorsMatchReference(t *testing.T) {
	for _, tt := range syntaxErrorCases {
		_, _, err := Parse(tt.src)
		got := ""
		if err != nil {
			got = err.Error()
		}
		if got != tt.want {
			t.Errorf("Parse(%q): error %q, want %q", tt.src, got, tt.want)
		}
	}
}
