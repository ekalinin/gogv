// SPDX-License-Identifier: MIT
// Provenance: original

// Command dot lays out graphs written in the DOT language, as Graphviz's dot
// does.
package main

import (
	"context"
	"os"
	"runtime/debug"
	"strings"

	"github.com/ekalinin/gogv/internal/cli"
)

func main() {
	os.Exit(cli.Run(context.Background(), os.Args[1:], os.Stdin, os.Stdout, os.Stderr, version()))
}

// version returns the version of gogv this command was built from.
func version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return "devel"
	}
	return strings.TrimPrefix(info.Main.Version, "v")
}
