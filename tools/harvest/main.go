// SPDX-License-Identifier: MIT
// Provenance: original

// Command harvest collects the DOT that gopuml generates for its conformance
// cases, so that gogv is checked against what its first consumer sends.
//
// It builds gopuml's command line, renders every .puml case of gopuml's
// curated corpus and of its Vega cache as SVG, with GRAPHVIZ_DOT pointing to
// a wrapper that saves each DOT input under the output directory before
// running the reference dot in its image. Inputs are named
// <capability>/<case>-<n>.gv, where <n> counts the dot runs of the case; the
// corpus stays in .cache and is never committed (it derives from upstream
// PlantUML inputs that gopuml does not commit either).
package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"
)

// wrapper is the dot that gopuml runs: it saves its stdin as the next
// <case>-<n>.gv and runs the reference dot on it with the same arguments.
// Runs without input (dot -V) save nothing.
const wrapper = `#!/bin/sh
set -u
in=$(mktemp)
cat > "$in"
if [ -s "$in" ]; then
	n=1
	while [ -e "$HARVEST_DIR/$HARVEST_CASE-$n.gv" ]; do n=$((n+1)); done
	mkdir -p "$(dirname "$HARVEST_DIR/$HARVEST_CASE")"
	cp "$in" "$HARVEST_DIR/$HARVEST_CASE-$n.gv"
fi
docker run --rm -i --platform linux/amd64 "$HARVEST_IMAGE" dot "$@" < "$in"
status=$?
rm -f "$in"
exit $status
`

func main() {
	gopuml := flag.String("gopuml", "../plantuml-in-go", "gopuml checkout")
	out := flag.String("out", ".cache/gopuml-dot", "corpus directory")
	image := flag.String("image", "gogv-reference", "reference image")
	jobs := flag.Int("j", min(runtime.NumCPU(), 4), "cases rendered at the same time")
	flag.Parse()
	if err := run(*gopuml, *out, *image, *jobs); err != nil {
		fmt.Fprintln(os.Stderr, "harvest:", err)
		os.Exit(1)
	}
}

type input struct{ id, path string }

func run(gopuml, out, image string, jobs int) error {
	tools, err := filepath.Abs(filepath.Join(".cache", "tools"))
	if err != nil {
		return err
	}
	if err := os.MkdirAll(tools, 0o755); err != nil {
		return err
	}
	bin := filepath.Join(tools, "gopuml-plantuml")
	build := exec.Command("go", "build", "-o", bin, "./cmd/plantuml")
	build.Dir = gopuml
	build.Stdout, build.Stderr = os.Stdout, os.Stderr
	if err := build.Run(); err != nil {
		return fmt.Errorf("build gopuml: %w", err)
	}
	dot := filepath.Join(tools, "dot-harvest")
	if err := os.WriteFile(dot, []byte(wrapper), 0o755); err != nil {
		return err
	}

	inputs, err := collect(gopuml)
	if err != nil {
		return err
	}
	absOut, err := filepath.Abs(out)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(absOut); err != nil {
		return err
	}
	if err := os.MkdirAll(absOut, 0o755); err != nil {
		return err
	}
	commit, _ := exec.Command("git", "-C", gopuml, "rev-parse", "HEAD").Output()
	if err := os.WriteFile(filepath.Join(absOut, "SOURCE"), []byte("gopuml "+strings.TrimSpace(string(commit))+"\n"), 0o644); err != nil {
		return err
	}

	work := make(chan input)
	var wg sync.WaitGroup
	for range max(jobs, 1) {
		wg.Go(func() {
			for in := range work {
				render(bin, dot, absOut, image, in)
			}
		})
	}
	for _, in := range inputs {
		work <- in
	}
	close(work)
	wg.Wait()

	var count int
	groups := map[string]int{}
	err = filepath.WalkDir(absOut, func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(path, ".gv") {
			count++
			rel, _ := filepath.Rel(absOut, path)
			groups[strings.Split(filepath.ToSlash(rel), "/")[0]]++
		}
		return err
	})
	if err != nil {
		return err
	}
	for _, g := range slices.Sorted(maps.Keys(groups)) {
		fmt.Printf("%-28s %d\n", g, groups[g])
	}
	fmt.Printf("harvest: %d DOT input(s) from %d case(s) in %s\n", count, len(inputs), out)
	return nil
}

// collect lists the .puml cases of gopuml's curated corpus, by capability,
// and of its Vega cache, as vega-<group>.
func collect(gopuml string) ([]input, error) {
	var inputs []input
	add := func(root, prefix string) error {
		if _, err := os.Stat(root); err != nil {
			return nil // the Vega cache is optional
		}
		return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || filepath.Ext(path) != ".puml" {
				return err
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(strings.TrimSuffix(rel, ".puml"))
			group, name, ok := strings.Cut(rel, "/")
			if !ok {
				group, name = "root", rel
			}
			name = strings.ReplaceAll(name, "/", "-")
			inputs = append(inputs, input{id: prefix + group + "/" + name, path: path})
			return nil
		})
	}
	if err := add(filepath.Join(gopuml, "testdata", "conformance", "cases"), ""); err != nil {
		return nil, err
	}
	if err := add(filepath.Join(gopuml, ".cache", "vega"), "vega-"); err != nil {
		return nil, err
	}
	return inputs, nil
}

// render runs gopuml on one case. Its output does not matter: the wrapper
// saves the DOT inputs as a side effect.
func render(bin, dot, out, image string, in input) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	src, err := os.Open(in.path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "harvest:", err)
		return
	}
	defer src.Close()
	cmd := exec.CommandContext(ctx, bin, "-tsvg", "-pipe")
	cmd.Dir = filepath.Dir(in.path) // relative !include paths
	cmd.Stdin = src
	cmd.Env = append(os.Environ(),
		"GRAPHVIZ_DOT="+dot,
		"HARVEST_DIR="+out,
		"HARVEST_CASE="+in.id,
		"HARVEST_IMAGE="+image,
	)
	_ = cmd.Run() // failing cases still leave their DOT behind
}
