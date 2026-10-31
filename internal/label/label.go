// SPDX-License-Identifier: MIT
// Provenance: doc https://graphviz.org/docs/attr-types/escString/
// Provenance: observed dot 16.1.0

// Package label turns plain-text label attributes into lines: it expands the
// escapes that name objects and splits the text at the line breaks, as the
// reference does.
package label

import "strings"

// Object is what a label belongs to, for the escapes that name objects.
type Object struct {
	// Graph is the name \G stands for: the root graph's for nodes, edges and
	// the root graph, a cluster's own for a cluster.
	Graph string
	// Node is the name \N stands for, set only for nodes.
	Node string
	// Edge is set for edges: \E stands for "tail->head" ("tail--head" in an
	// undirected graph), \T for the tail and \H for the head.
	Edge             bool
	Tail, Head       string
	DirectedEdgeName bool
}

// Expand replaces \G, \N, \E, \T and \H in s for obj. As in the reference,
// \E is empty for objects other than edges, \N is replaced only for nodes,
// \T and \H only for edges; other escapes, including \\, are kept for Lines.
func Expand(s string, obj Object) string {
	if !strings.Contains(s, `\`) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != '\\' || i+1 >= len(s) {
			b.WriteByte(c)
			continue
		}
		next := s[i+1]
		switch {
		case next == 'G':
			b.WriteString(obj.Graph)
		case next == 'N' && obj.Node != "":
			b.WriteString(obj.Node)
		case next == 'E':
			if obj.Edge {
				op := "--"
				if obj.DirectedEdgeName {
					op = "->"
				}
				b.WriteString(obj.Tail)
				b.WriteString(op)
				b.WriteString(obj.Head)
			}
		case next == 'T' && obj.Edge:
			b.WriteString(obj.Tail)
		case next == 'H' && obj.Edge:
			b.WriteString(obj.Head)
		default:
			b.WriteByte(c)
			b.WriteByte(next)
		}
		i++
	}
	return b.String()
}

// Justification is how a line is placed in its label.
type Justification byte

// The justifications of a line.
const (
	Center Justification = 'n'
	Left   Justification = 'l'
	Right  Justification = 'r'
)

// Line is one line of a label.
type Line struct {
	Text string
	Just Justification
}

// Lines splits expanded label text into lines. \n, \l and \r end a line that
// is centered, left-justified or right-justified, and so does a newline
// character (centered); the text after the last of them forms a centered line
// when it is not empty. Any other backslash is
// dropped and the character after it kept, so \\ gives a backslash.
func Lines(s string) []Line {
	var lines []Line
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\n' {
			// A newline in the text ends a centered line, as \n does.
			lines = append(lines, Line{Text: b.String(), Just: Center})
			b.Reset()
			continue
		}
		if c != '\\' || i+1 >= len(s) {
			b.WriteByte(c)
			continue
		}
		i++
		switch next := s[i]; next {
		case 'n', 'l', 'r':
			lines = append(lines, Line{Text: b.String(), Just: Justification(next)})
			b.Reset()
		default:
			b.WriteByte(next)
		}
	}
	if b.Len() > 0 {
		lines = append(lines, Line{Text: b.String(), Just: Center})
	}
	return lines
}
