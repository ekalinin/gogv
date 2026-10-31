// SPDX-License-Identifier: MIT
// Provenance: original

package gogv

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/ekalinin/gogv/internal/attrs"
	"github.com/ekalinin/gogv/internal/dotlang"
	"github.com/ekalinin/gogv/internal/layout"
	"github.com/ekalinin/gogv/internal/output"
)

// Formats are the output formats Render writes.
var Formats = []string{"canon", "plain", "svg"}

// laterFormats are the output formats of version 1 that are not supported
// yet.
var laterFormats = []string{"svg_inline", "plain-ext", "dot", "gv", "xdot", "xdot1.2", "xdot1.4", "json", "json0", "dot_json", "xdot_json"}

// Warning reports something the caller may want to know about an input:
// mostly a Graphviz feature that is not supported yet, with the default
// used instead.
type Warning struct {
	// Object names the graph, subgraph, node or edge, as "node a"; empty
	// for warnings about the input text.
	Object string
	// Feature names what is not supported yet, as "shape=record".
	Feature string
	// Default is what is used instead, or "".
	Default string
	text    string
}

// String returns the warning as the command line writes it, without its
// "Warning: " prefix.
func (w Warning) String() string { return w.text }

// NotSupportedYet reports whether the warning is about a feature that is not
// supported yet.
func (w Warning) NotSupportedYet() bool { return w.Feature != "" }

// SyntaxError is a syntax error in the input, worded as the reference words
// it.
type SyntaxError struct {
	Line    int
	Near    string
	message string
}

func (e *SyntaxError) Error() string { return e.message }

// FormatError is an output format that Render does not write.
type FormatError struct {
	Format string
	// NotSupportedYet is set for the formats of a later version.
	NotSupportedYet bool
}

func (e *FormatError) Error() string {
	if e.NotSupportedYet {
		return fmt.Sprintf("format %q is not supported yet", e.Format)
	}
	return fmt.Sprintf("Format: %q not recognized. Use one of: %s", e.Format, strings.Join(Formats, " "))
}

// Render lays out every graph of DOT source text and writes it in format, as
// `dot -T<format>` writes it to standard output. It returns the output, the
// warnings, and an error for a format it does not write or a syntax error.
// After a syntax error, the output holds the graphs read before the one that
// contains the error. Render is safe for concurrent use.
func Render(ctx context.Context, src string, format string) ([]byte, []Warning, error) {
	if !slices.Contains(Formats, format) {
		return nil, nil, &FormatError{Format: format, NotSupportedYet: slices.Contains(laterFormats, format)}
	}
	file, lexWarnings, parseErr := dotlang.Parse(src)
	var warnings []Warning
	for _, w := range lexWarnings {
		warnings = append(warnings, Warning{text: w})
	}
	var out bytes.Buffer
	for _, decl := range file.Graphs {
		g := dotlang.Build(decl)
		for _, w := range g.ResolveClusters() {
			warnings = append(warnings, Warning{text: w})
		}
		for _, w := range attrs.Check(g) {
			warnings = append(warnings, Warning{Object: w.Object, Feature: w.Feature, Default: w.Default, text: w.String()})
		}
		if format == "canon" {
			out.Write(output.Canon(g))
			continue
		}
		res, err := layout.Layout(ctx, g)
		if err != nil {
			return out.Bytes(), warnings, err
		}
		switch format {
		case "plain":
			out.Write(output.Plain(res))
		case "svg":
			out.Write(output.SVG(res))
		}
	}
	if parseErr != nil {
		var se *dotlang.SyntaxError
		if errors.As(parseErr, &se) {
			return out.Bytes(), warnings, &SyntaxError{Line: se.Line, Near: se.Near, message: se.Error()}
		}
		return out.Bytes(), warnings, parseErr
	}
	return out.Bytes(), warnings, nil
}
