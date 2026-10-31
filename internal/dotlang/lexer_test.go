// SPDX-License-Identifier: MIT
// Provenance: doc https://graphviz.org/doc/info/lang.html

package dotlang

import (
	"strings"
	"testing"
)

func lexAll(t *testing.T, src string) []Token {
	t.Helper()
	l := NewLexer(src)
	var toks []Token
	for {
		tok, err := l.Next()
		if err != nil {
			t.Fatalf("lex %q: %v", src, err)
		}
		if tok.Kind == EOF {
			return toks
		}
		toks = append(toks, tok)
	}
}

func TestLexerTokenKinds(t *testing.T) {
	tests := []struct {
		src  string
		kind Kind
		text string
	}{
		{"abc_9", ID, "abc_9"},
		{"_x", ID, "_x"},
		{"\xc3\xa9t\xc3\xa9", ID, "\xc3\xa9t\xc3\xa9"},
		{"12", ID, "12"},
		{"-3.5", ID, "-3.5"},
		{".5", ID, ".5"},
		{"-.5", ID, "-.5"},
		{"12.", ID, "12."},
		{`"a b"`, QuotedID, "a b"},
		{`"c\"d"`, QuotedID, `c"d`},
		{`"a\\b"`, QuotedID, `a\\b`},
		{`"a\nb"`, QuotedID, `a\nb`},
		{"\"long\\\nline\"", QuotedID, "longline"},
		{"<<b>hi</b>>", HTMLID, "<b>hi</b>"},
		{"<x>", HTMLID, "x"},
		{"{", LBrace, "{"},
		{"}", RBrace, "}"},
		{"[", LBracket, "["},
		{"]", RBracket, "]"},
		{";", Semi, ";"},
		{",", Comma, ","},
		{"=", Equal, "="},
		{":", Colon, ":"},
		{"+", Plus, "+"},
		{"->", DirectedOp, "->"},
		{"--", UndirectOp, "--"},
		{"strict", KwStrict, "strict"},
		{"Graph", KwGraph, "Graph"},
		{"DIGRAPH", KwDigraph, "DIGRAPH"},
		{"subgraph", KwSubgraph, "subgraph"},
		{"Node", KwNode, "Node"},
		{"edge", KwEdge, "edge"},
		{"@", Bad, "@"},
	}
	for _, tt := range tests {
		toks := lexAll(t, tt.src)
		if len(toks) != 1 || toks[0].Kind != tt.kind || toks[0].Text != tt.text {
			t.Errorf("lex %q = %+v, want one %v %q", tt.src, toks, tt.kind, tt.text)
		}
	}
}

func TestLexerComments(t *testing.T) {
	src := "a // line comment\n/* block\ncomment */ b\n# preprocessor line\nc # rest of the line\nd#e\nf"
	toks := lexAll(t, src)
	var texts []string
	for _, tok := range toks {
		texts = append(texts, tok.Text)
	}
	if got := strings.Join(texts, " "); got != "a b c d f" {
		t.Fatalf("tokens %q, want a b c d f", texts)
	}
	if toks[1].Line != 3 || toks[2].Line != 5 || toks[4].Line != 7 {
		t.Errorf("lines of b, c and f = %d, %d, %d; want 3, 5, 7", toks[1].Line, toks[2].Line, toks[4].Line)
	}
}

func TestLexerLines(t *testing.T) {
	toks := lexAll(t, "a\n\"x\ny\"\nb\n<p\n>\nc")
	want := []int{1, 2, 4, 5, 7}
	for i, w := range want {
		if toks[i].Line != w {
			t.Errorf("token %d (%q) on line %d, want %d", i, toks[i].Text, toks[i].Line, w)
		}
	}
}

func TestLexerBadlyDelimitedNumber(t *testing.T) {
	l := NewLexer("2abc")
	tok, _ := l.Next()
	if tok.Kind != ID || tok.Text != "2" {
		t.Fatalf("first token %+v, want ID 2", tok)
	}
	tok, _ = l.Next()
	if tok.Kind != ID || tok.Text != "abc" {
		t.Fatalf("second token %+v, want ID abc", tok)
	}
	if len(l.Warnings) != 1 || l.Warnings[0] != "syntax ambiguity - badly delimited number '2a' in line 1 of <stdin> splits into two tokens" {
		t.Errorf("warnings %q", l.Warnings)
	}
}

func TestLexerUnterminated(t *testing.T) {
	tests := []struct{ src, want string }{
		{"\"abc }\n", "syntax error in line 2 scanning a quoted string (missing endquote? longer than 16384?)\nString starting:\"abc }\n"},
		{"<a<b>] }\n", "syntax error in line 2 scanning a HTML string (missing '>'? bad nesting? longer than 16384?)\nString starting:<a<b>] }\n"},
		{"\"" + strings.Repeat("x", 300), "syntax error in line 1 scanning a quoted string (missing endquote? longer than 16384?)\nString starting:\"" + strings.Repeat("x", 80)},
	}
	for _, tt := range tests {
		_, err := NewLexer(tt.src).Next()
		if err == nil || err.Error() != tt.want {
			t.Errorf("lex %q: error %v, want %q", tt.src, err, tt.want)
		}
	}
}
