// SPDX-License-Identifier: MIT
// Provenance: doc https://graphviz.org/doc/info/lang.html
// Provenance: observed dot 16.1.0

// Package dotlang reads graphs written in the DOT language: a lexer, a parser
// that produces a syntax tree, and the construction of the graph model with
// the semantics of attribute defaults, subgraphs, edge chains and strict
// graphs.
package dotlang

import (
	"fmt"
	"strings"
)

// Kind is the kind of a token.
type Kind int

// The token kinds.
const (
	EOF        Kind = iota
	ID              // identifier or numeral
	QuotedID        // double-quoted string, with \" unescaped
	HTMLID          // HTML string, without its outer angle brackets
	LBrace          // {
	RBrace          // }
	LBracket        // [
	RBracket        // ]
	Semi            // ;
	Comma           // ,
	Equal           // =
	Colon           // :
	Plus            // +
	DirectedOp      // ->
	UndirectOp      // --
	KwStrict        // strict
	KwGraph         // graph
	KwDigraph       // digraph
	KwSubgraph      // subgraph
	KwNode          // node
	KwEdge          // edge
	Bad             // a character that starts no token
)

// Token is a lexical token of DOT.
type Token struct {
	Kind Kind
	// Text is the value of an ID: the identifier, the numeral, the string
	// without its quotes, or the HTML string without its angle brackets. For
	// other tokens it is the source text.
	Text string
	// Line is the line where the token starts, from 1.
	Line int
}

var keywords = map[string]Kind{
	"strict":   KwStrict,
	"graph":    KwGraph,
	"digraph":  KwDigraph,
	"subgraph": KwSubgraph,
	"node":     KwNode,
	"edge":     KwEdge,
}

// IsKeyword reports whether s is a DOT keyword, without regard to case.
func IsKeyword(s string) bool {
	_, ok := keywords[strings.ToLower(s)]
	return ok
}

// Lexer splits DOT source into tokens.
type Lexer struct {
	src  string
	pos  int
	line int
	// Input names the input in warnings, "<stdin>" by default.
	Input string
	// Warnings collects the lexical warnings, such as badly delimited
	// numbers.
	Warnings []string
}

// NewLexer returns a lexer for src.
func NewLexer(src string) *Lexer {
	return &Lexer{src: src, line: 1, Input: "<stdin>"}
}

func isIDStart(c byte) bool {
	return c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= 0x80
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

// skip skips blanks and comments.
func (l *Lexer) skip() error {
	for l.pos < len(l.src) {
		c := l.src[l.pos]
		switch {
		case c == '\n':
			l.line++
			l.pos++
		case c == ' ' || c == '\t' || c == '\r' || c == '\f' || c == '\v':
			l.pos++
		case c == '#':
			// The reference discards the rest of the line after '#', not
			// only on lines that start with it.
			for l.pos < len(l.src) && l.src[l.pos] != '\n' {
				l.pos++
			}
		case c == '/' && strings.HasPrefix(l.src[l.pos:], "//"):
			for l.pos < len(l.src) && l.src[l.pos] != '\n' {
				l.pos++
			}
		case c == '/' && strings.HasPrefix(l.src[l.pos:], "/*"):
			end := strings.Index(l.src[l.pos+2:], "*/")
			if end < 0 {
				l.line += strings.Count(l.src[l.pos:], "\n")
				l.pos = len(l.src)
				return nil
			}
			l.line += strings.Count(l.src[l.pos:l.pos+2+end], "\n")
			l.pos += 2 + end + 2
		default:
			return nil
		}
	}
	return nil
}

// Next returns the next token. At the end of the input it returns EOF
// tokens.
func (l *Lexer) Next() (Token, error) {
	if err := l.skip(); err != nil {
		return Token{}, err
	}
	if l.pos >= len(l.src) {
		return Token{Kind: EOF, Line: l.line}, nil
	}
	start, line := l.pos, l.line
	c := l.src[l.pos]
	single := map[byte]Kind{
		'{': LBrace, '}': RBrace, '[': LBracket, ']': RBracket,
		';': Semi, ',': Comma, '=': Equal, ':': Colon, '+': Plus,
	}
	switch {
	case single[c] != 0:
		l.pos++
		return Token{Kind: single[c], Text: string(c), Line: line}, nil
	case strings.HasPrefix(l.src[l.pos:], "->"):
		l.pos += 2
		return Token{Kind: DirectedOp, Text: "->", Line: line}, nil
	case strings.HasPrefix(l.src[l.pos:], "--"):
		l.pos += 2
		return Token{Kind: UndirectOp, Text: "--", Line: line}, nil
	case c == '"':
		return l.quoted()
	case c == '<':
		return l.html()
	case isIDStart(c):
		for l.pos < len(l.src) && (isIDStart(l.src[l.pos]) || isDigit(l.src[l.pos])) {
			l.pos++
		}
		text := l.src[start:l.pos]
		if k, ok := keywords[strings.ToLower(text)]; ok {
			return Token{Kind: k, Text: text, Line: line}, nil
		}
		return Token{Kind: ID, Text: text, Line: line}, nil
	case isDigit(c) || c == '.' || c == '-' && l.pos+1 < len(l.src) && (isDigit(l.src[l.pos+1]) || l.src[l.pos+1] == '.'):
		return l.numeral()
	}
	l.pos++
	return Token{Kind: Bad, Text: string(c), Line: line}, nil
}

// numeral reads [-]?(.[0-9]+ | [0-9]+(.[0-9]*)?). A numeral directly
// followed by an identifier character is split, with a warning, as the
// reference does.
func (l *Lexer) numeral() (Token, error) {
	start, line := l.pos, l.line
	if l.src[l.pos] == '-' {
		l.pos++
	}
	digits := 0
	for l.pos < len(l.src) && isDigit(l.src[l.pos]) {
		l.pos++
		digits++
	}
	if l.pos < len(l.src) && l.src[l.pos] == '.' {
		l.pos++
		for l.pos < len(l.src) && isDigit(l.src[l.pos]) {
			l.pos++
			digits++
		}
	}
	if digits == 0 {
		l.pos = start + 1
		return Token{Kind: Bad, Text: l.src[start:l.pos], Line: line}, nil
	}
	if l.pos < len(l.src) && (isIDStart(l.src[l.pos]) || l.src[l.pos] == '.') {
		end := l.pos
		for end < len(l.src) && (isIDStart(l.src[end]) || isDigit(l.src[end]) || l.src[end] == '.') {
			end++
		}
		l.Warnings = append(l.Warnings, fmt.Sprintf(
			"syntax ambiguity - badly delimited number '%s' in line %d of %s splits into two tokens",
			l.src[start:l.pos+1], line, l.Input))
	}
	return Token{Kind: ID, Text: l.src[start:l.pos], Line: line}, nil
}

// quoted reads a double-quoted string. \" becomes ", a backslash before a
// newline joins the lines, and every other character is kept as it is.
func (l *Lexer) quoted() (Token, error) {
	start, line := l.pos, l.line
	l.pos++ // opening quote
	var b strings.Builder
	for l.pos < len(l.src) {
		c := l.src[l.pos]
		switch {
		case c == '"':
			l.pos++
			return Token{Kind: QuotedID, Text: b.String(), Line: line}, nil
		case c == '\\' && l.pos+1 < len(l.src) && l.src[l.pos+1] == '"':
			b.WriteByte('"')
			l.pos += 2
		case c == '\\' && l.pos+1 < len(l.src) && l.src[l.pos+1] == '\n':
			l.line++
			l.pos += 2
		case c == '\\' && strings.HasPrefix(l.src[l.pos+1:], "\r\n"):
			l.line++
			l.pos += 3
		default:
			if c == '\n' {
				l.line++
			}
			b.WriteByte(c)
			l.pos++
		}
	}
	return Token{}, &SyntaxError{Line: l.line, Scanning: "a quoted string (missing endquote? longer than 16384?)", Start: l.src[start:]}
}

// html reads an HTML string: angle brackets must balance.
func (l *Lexer) html() (Token, error) {
	start, line := l.pos, l.line
	depth := 0
	for l.pos < len(l.src) {
		c := l.src[l.pos]
		switch c {
		case '<':
			depth++
		case '>':
			depth--
		case '\n':
			l.line++
		}
		l.pos++
		if depth == 0 {
			return Token{Kind: HTMLID, Text: l.src[start+1 : l.pos-1], Line: line}, nil
		}
	}
	return Token{}, &SyntaxError{Line: l.line, Scanning: "a HTML string (missing '>'? bad nesting? longer than 16384?)", Start: l.src[start:]}
}

// SyntaxError is a syntax error, worded as the reference words it.
type SyntaxError struct {
	Line int
	// Near is the text of the token where the error was found; AtEOF is set
	// instead when the input ended.
	Near  string
	AtEOF bool
	// Scanning describes an unterminated string, and Start is the input from
	// its beginning.
	Scanning string
	Start    string
}

// startLimit is how much of an unterminated string the reference shows.
const startLimit = 81

func (e *SyntaxError) Error() string {
	switch {
	case e.Scanning != "":
		start := e.Start
		if len(start) > startLimit {
			start = start[:startLimit]
		}
		return fmt.Sprintf("syntax error in line %d scanning %s\nString starting:%s", e.Line, e.Scanning, start)
	case e.AtEOF:
		return fmt.Sprintf("syntax error in line %d", e.Line)
	}
	return fmt.Sprintf("syntax error in line %d near '%s'", e.Line, e.Near)
}

// endGraph is called after the closing brace of a graph: as the reference
// does, it discards the rest of that line and numbers the next line 1.
func (l *Lexer) endGraph() {
	for l.pos < len(l.src) && l.src[l.pos] != '\n' {
		l.pos++
	}
	if l.pos < len(l.src) {
		l.pos++
	}
	l.line = 1
}
