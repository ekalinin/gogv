// SPDX-License-Identifier: MIT
// Provenance: original

// Command reduce reduces a DOT input whose layout by gogv does not pass the
// proxy of gopuml's contract against the reference: it removes statements,
// attributes and subgraphs while the layout of the smaller input still
// fails the proxy on a node or an edge label that fails it in the original
// input, and writes the smallest input it finds.
//
// The reference runs in the image -image; its plain outputs are cached in
// -cache by the SHA-256 of the input, so that a second run of the same
// reduction runs no container.
//
// With -on, the reduction keeps failing on the objects it names, as the
// proxy names them ("node a", "label of a->b#0"), instead of any object
// that fails in the input.
//
// Usage: go run ./tools/reduce -in case.gv -out reduced.gv [-on objects] [-image name] [-cache dir]
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ekalinin/gogv/internal/dotlang"
)

func main() {
	log.SetFlags(0)
	if err := run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	fl := flag.NewFlagSet("reduce", flag.ContinueOnError)
	in := fl.String("in", "", "DOT input to reduce")
	out := fl.String("out", "", "file for the reduced input")
	image := fl.String("image", "gogv-reference", "image of the reference")
	cache := fl.String("cache", ".cache/reduce", "directory of the cached reference outputs")
	on := fl.String("on", "", "comma-separated objects to keep failing, as the proxy names them")
	if err := fl.Parse(args); err != nil {
		return err
	}
	if *in == "" || *out == "" || fl.NArg() > 0 {
		return errors.New("usage: reduce -in case.gv -out reduced.gv [-on objects] [-image name] [-cache dir]")
	}
	src, err := os.ReadFile(*in)
	if err != nil {
		return err
	}
	ref := &referenceRunner{image: *image, cache: *cache}
	var objects []string
	if *on != "" {
		objects = strings.Split(*on, ",")
	}
	test, failing, err := predicate(string(src), ref.layout, gogvLayout, objects)
	if err != nil {
		return fmt.Errorf("%s: %w", *in, err)
	}
	fmt.Printf("%s: reducing while it fails the proxy on %s\n", *in, strings.Join(failing, ", "))
	// The search tries some candidates more than once.
	seen := map[string]bool{}
	memo := func(candidate string) bool {
		v, ok := seen[candidate]
		if !ok {
			v = test(candidate)
			seen[candidate] = v
		}
		return v
	}
	reduced, err := reduce(string(src), memo)
	if err != nil {
		return err
	}
	if err := os.WriteFile(*out, []byte(reduced), 0o644); err != nil {
		return err
	}
	fmt.Printf("%d statements of %d kept in %s; %d reference runs, %d from the cache\n",
		countStmts(reduced), countStmts(string(src)), *out, ref.runs, ref.cached)
	return nil
}

// countStmts returns the number of statements of a DOT input.
func countStmts(src string) int {
	f, _, err := dotlang.Parse(src)
	if err != nil {
		return 0
	}
	return len(stmtRefs(f))
}
