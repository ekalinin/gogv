// SPDX-License-Identifier: MIT
// Provenance: original

// Command clustervariant writes a variant of the harvested gopuml graphs
// that have clusters: each HTML label that gopuml puts on a cluster is
// replaced by an empty label. gopuml puts such a label on every cluster, so
// the variant lets the conformance run check the layout of gopuml's clusters
// before HTML labels are supported. The variant goes to one capability
// directory, gopuml-clusters, as <capability>-<case>.gv, and stays in .cache
// like the corpus it comes from.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// clusterRE finds a graph with a cluster.
var clusterRE = regexp.MustCompile(`(?i)subgraph\s+cluster`)

// htmlLabelRE finds an HTML table label set in the statement list of a
// cluster, on the line that opens the cluster, as gopuml writes it.
var htmlLabelRE = regexp.MustCompile(`(subgraph cluster\w*\s*\{[^\n]*?)label=<<TABLE.*?</TABLE>>;`)

// variant returns src with the HTML labels of its clusters made empty.
func variant(src string) string {
	return htmlLabelRE.ReplaceAllString(src, `${1}label="";`)
}

func main() {
	in := flag.String("in", ".cache/gopuml-dot", "harvested corpus")
	out := flag.String("out", ".cache/gopuml-cluster-dot", "variant corpus")
	flag.Parse()
	if err := run(*in, *out); err != nil {
		fmt.Fprintln(os.Stderr, "clustervariant:", err)
		os.Exit(1)
	}
}

func run(in, out string) error {
	if err := os.RemoveAll(out); err != nil {
		return err
	}
	dir := filepath.Join(out, "gopuml-clusters")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	n := 0
	err := filepath.WalkDir(in, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".gv" {
			return err
		}
		src, err := os.ReadFile(path)
		if err != nil || !clusterRE.Match(src) {
			return err
		}
		rel, err := filepath.Rel(in, path)
		if err != nil {
			return err
		}
		name := strings.ReplaceAll(filepath.ToSlash(rel), "/", "-")
		n++
		return os.WriteFile(filepath.Join(dir, name), []byte(variant(string(src))), 0o644)
	})
	if err != nil {
		return err
	}
	// The variant comes from the same gopuml as the corpus.
	source, err := os.ReadFile(filepath.Join(in, "SOURCE"))
	if err == nil {
		err = os.WriteFile(filepath.Join(out, "SOURCE"), source, 0o644)
	} else if errors.Is(err, fs.ErrNotExist) {
		err = nil
	}
	if err != nil {
		return err
	}
	fmt.Printf("clustervariant: %d graph(s) written to %s\n", n, dir)
	return nil
}
