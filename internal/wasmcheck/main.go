// SPDX-License-Identifier: MIT
// Provenance: original

// Command wasmcheck renders the DOT text of its first argument as SVG on
// standard output; the library tests run it in the WebAssembly build.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ekalinin/gogv"
)

func main() {
	out, _, err := gogv.Render(context.Background(), os.Args[1], "svg")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Stdout.Write(out)
}
