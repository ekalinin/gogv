// SPDX-License-Identifier: MIT
// Provenance: original

// Command corpushash renders every case of a corpus directory as canon,
// plain and svg, and prints a SHA-256 of each output and of all of them, so
// that builds for different platforms can be compared.
package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ekalinin/gogv"
)

func main() {
	root := os.Args[1]
	var files []string
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(path, ".gv") {
			files = append(files, path)
		}
		return err
	})
	slices.Sort(files)
	all := sha256.New()
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for _, format := range gogv.Formats {
			out, _, _ := gogv.Render(context.Background(), string(src), format)
			sum := sha256.Sum256(out)
			all.Write(sum[:])
			rel, _ := filepath.Rel(root, f)
			fmt.Printf("%x %s %s\n", sum[:8], format, filepath.ToSlash(rel))
		}
	}
	fmt.Printf("%x all\n", all.Sum(nil))
}
