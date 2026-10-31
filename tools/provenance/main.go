// SPDX-License-Identifier: MIT
// Provenance: gopuml tools/provenance/main.go

// Command provenance checks that every non-test Go source file declares where
// its code comes from, so that the project stays clean-room: written from
// papers, the Graphviz documentation and the observed behavior of dot, never
// from the code of Graphviz or of its ports.
//
// Each checked file must start with
//
//	// SPDX-License-Identifier: MIT
//
// and contain, within its first lines, one or more provenance lines:
//
//	// Provenance: original
//	// Provenance: paper Gansner, Koutsofios, North, Vo 1993, section 2
//	// Provenance: doc https://graphviz.org/doc/info/lang.html
//	// Provenance: observed dot 16.1.0
//	// Provenance: gopuml internal/conformance/compare.go
//	// Provenance: fdlibm 5.3 e_atan2.c
//	// Provenance: liberation-fonts 2.1.5
//
// The last one marks tables generated from the Liberation fonts. "original"
// cannot be combined with another provenance. Code translated from
// fdlibm keeps Sun's notice:
//
//	// Derived from fdlibm, Copyright (C) 1993 by Sun Microsystems, Inc. Permission to use, copy, modify, and distribute this software is freely granted, provided that this notice is preserved.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const (
	spdxLine         = "// SPDX-License-Identifier: MIT"
	provenancePrefix = "// Provenance: "
	docPrefix        = "https://graphviz.org/"
	observedValue    = "observed dot 16.1.0"
	fdlibmPrefix     = "fdlibm 5.3 "
	liberationValue  = "liberation-fonts 2.1.5"
	fdlibmNotice     = "// Derived from fdlibm, Copyright (C) 1993 by Sun Microsystems, Inc. Permission to use, copy, modify, and distribute this software is freely granted, provided that this notice is preserved."
	headerLines      = 15
)

// skippedDirs are not walked: they hold no project Go code.
var skippedDirs = map[string]bool{
	".git":     true,
	".cache":   true,
	".claude":  true,
	"openspec": true,
	"testdata": true,
}

// Problem is a single provenance violation.
type Problem struct {
	File    string
	Message string
}

func (p Problem) String() string { return p.File + ": " + p.Message }

// checkFile validates the header of one Go file.
func checkFile(path, content string) []Problem {
	var problems []Problem
	add := func(format string, args ...any) {
		problems = append(problems, Problem{File: path, Message: fmt.Sprintf(format, args...)})
	}

	var original bool
	var sources, fdlibm int
	var hasFdlibmNotice bool
	sc := bufio.NewScanner(strings.NewReader(content))
	for i := 0; i < headerLines && sc.Scan(); i++ {
		line := strings.TrimRight(sc.Text(), " \t\r")
		if i == 0 && line != spdxLine {
			add("first line is not %q", spdxLine)
		}
		if line == fdlibmNotice {
			hasFdlibmNotice = true
			continue
		}
		value, ok := strings.CutPrefix(line, provenancePrefix)
		if !ok {
			continue
		}
		kind, detail, _ := strings.Cut(value, " ")
		detail = strings.TrimSpace(detail)
		switch {
		case value == "original":
			original = true
		case kind == "paper" && detail != "":
			sources++
		case kind == "doc" && strings.HasPrefix(detail, docPrefix):
			sources++
		case value == observedValue, value == liberationValue:
			sources++
		case kind == "gopuml" && strings.HasSuffix(detail, ".go"):
			sources++
		case strings.HasPrefix(value, fdlibmPrefix) && strings.HasSuffix(value, ".c"):
			fdlibm++
		default:
			add("unrecognized provenance %q", value)
		}
	}

	switch {
	case original && sources+fdlibm > 0:
		add("declares both original and another provenance")
	case !original && sources+fdlibm == 0 && len(problems) == 0:
		add("missing provenance header")
	case fdlibm > 0 && !hasFdlibmNotice:
		add("fdlibm file lacks Sun's notice %q", fdlibmNotice)
	}
	return problems
}

// checkTree walks the given roots and checks every non-test .go file.
func checkTree(roots []string) ([]Problem, int, error) {
	var problems []Problem
	var count int
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if path != root && skippedDirs[d.Name()] {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			count++
			problems = append(problems, checkFile(filepath.ToSlash(path), string(data))...)
			return nil
		})
		if err != nil {
			return nil, count, err
		}
	}
	return problems, count, nil
}

func main() {
	flag.Parse()
	roots := flag.Args()
	if len(roots) == 0 {
		roots = []string{"."}
	}

	problems, count, err := checkTree(roots)
	if err != nil {
		fmt.Fprintln(os.Stderr, "provenance:", err)
		os.Exit(2)
	}
	for _, p := range problems {
		fmt.Fprintln(os.Stderr, p)
	}
	if len(problems) > 0 {
		fmt.Fprintf(os.Stderr, "provenance: %d problem(s) in %d file(s) checked\n", len(problems), count)
		os.Exit(1)
	}
	fmt.Printf("provenance: %d file(s) OK\n", count)
}
