// SPDX-License-Identifier: MIT
// Provenance: doc https://graphviz.org/doc/info/command.html
// Provenance: observed dot 16.1.0

// Package cli is the dot command line: it reads DOT files or standard input
// and writes their layouts in the format chosen with -T.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ekalinin/gogv"
)

// GraphvizVersion is the Graphviz version that -V names, so that programs
// that read the Graphviz version from it find the reference's.
const GraphvizVersion = "16.1.0"

const usage = "Usage: dot [-V] [-T<format>] [-o<file>] [-Kdot] <dot files>"

// Run runs the command with args, without the program name, and returns the
// exit status.
func Run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer, version string) int {
	format := "dot"
	outFile := ""
	var files []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if !strings.HasPrefix(a, "-") {
			files = append(files, a)
			continue
		}
		if len(a) < 2 {
			fmt.Fprintf(stderr, "Error: dot: option %s is not supported yet\n", a)
			return 1
		}
		flag, value := a[1], a[2:]
		needsValue := flag == 'T' || flag == 'o' || flag == 'K'
		if needsValue && value == "" {
			if i+1 >= len(args) {
				fmt.Fprintf(stderr, "Missing argument for -%c flag\n%s\n", flag, usage)
				return 1
			}
			i++
			value = args[i]
		}
		switch flag {
		case 'V':
			fmt.Fprintf(stderr, "dot - graphviz version %s (gogv %s)\n", GraphvizVersion, version)
			return 0
		case 'T':
			format = value
		case 'o':
			outFile = value
		case 'K':
			if value != "dot" {
				fmt.Fprintf(stderr, "Error: dot: layout engine %q is not supported yet\n", value)
				return 1
			}
		default:
			fmt.Fprintf(stderr, "Error: dot: option -%c is not supported yet\n", flag)
			return 1
		}
	}

	// Check the format before reading any input.
	if _, _, err := gogv.Render(ctx, "", format); err != nil {
		var fe *gogv.FormatError
		if errors.As(err, &fe) {
			fmt.Fprintln(stderr, fe.Error())
		} else {
			fmt.Fprintf(stderr, "Error: dot: %s\n", err)
		}
		return 1
	}

	out := stdout
	if outFile != "" {
		f, err := os.Create(outFile)
		if err != nil {
			fmt.Fprintf(stderr, "Error: dot: can't open %s: %s\n", outFile, reason(err))
			return 2
		}
		defer f.Close()
		out = f
	}

	status := 0
	render := func(name string, src []byte) {
		data, warnings, err := gogv.Render(ctx, string(src), format)
		for _, w := range warnings {
			text := w.String()
			if name != "<stdin>" {
				text = strings.ReplaceAll(text, " of <stdin> ", " of "+name+" ")
			}
			fmt.Fprintf(stderr, "Warning: %s\n", text)
		}
		out.Write(data)
		if err != nil {
			fmt.Fprintf(stderr, "Error: %s: %s\n", name, err)
			status = 1
		}
	}
	if len(files) == 0 {
		src, err := io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintf(stderr, "Error: dot: %s\n", err)
			return 2
		}
		render("<stdin>", src)
		return status
	}
	for _, name := range files {
		src, err := os.ReadFile(name)
		if err != nil {
			fmt.Fprintf(stderr, "Error: dot: can't open %s: %s\n", name, reason(err))
			status = 2
			continue
		}
		render(name, src)
	}
	return status
}

// reason returns the system's reason for a file error, as the reference
// writes it.
func reason(err error) string {
	switch {
	case errors.Is(err, os.ErrNotExist):
		return "No such file or directory"
	case errors.Is(err, os.ErrPermission):
		return "Permission denied"
	}
	var pe *os.PathError
	if errors.As(err, &pe) {
		return pe.Err.Error()
	}
	return err.Error()
}
