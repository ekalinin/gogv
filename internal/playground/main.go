// SPDX-License-Identifier: MIT
// Provenance: gopuml cmd/plantuml-wasm/main.go

//go:build js && wasm

// Command playground is the WebAssembly entry point of the playground of the
// landing page. It defines the global gogv, with the version of the build
// and render(source), which returns a Promise of the SVG, the warnings and
// the error of the source; gogv.mjs loads it in browsers and Node.js.
package main

import (
	"context"
	"errors"
	"syscall/js"

	"github.com/ekalinin/gogv"
)

// version is the release that the module is built from, set with
// -ldflags=-X main.version=<tag>.
var version = "devel"

func main() {
	js.Global().Set("gogv", js.ValueOf(map[string]any{"version": version, "render": js.FuncOf(render)}))
	select {}
}

// result converts a rendering to a JavaScript object: svg (a Uint8Array,
// or null after an error), warnings (their texts, as the command line words
// them) and error (null, or an object with message and, for a syntax error,
// line).
func result(svg []byte, warnings []gogv.Warning, err error) js.Value {
	texts := make([]any, len(warnings))
	for i, w := range warnings {
		texts[i] = w.String()
	}
	out := map[string]any{"svg": nil, "warnings": texts, "error": nil}
	if err != nil {
		e := map[string]any{"message": err.Error(), "line": nil}
		if se, ok := errors.AsType[*gogv.SyntaxError](err); ok {
			e["line"] = se.Line
		}
		out["error"] = e
		return js.ValueOf(out)
	}
	data := js.Global().Get("Uint8Array").New(len(svg))
	js.CopyBytesToJS(data, svg)
	out["svg"] = data
	return js.ValueOf(out)
}

func render(_ js.Value, args []js.Value) any {
	source := ""
	if len(args) > 0 {
		source = args[0].String()
	}
	var executor js.Func
	executor = js.FuncOf(func(_ js.Value, pa []js.Value) any {
		defer executor.Release()
		resolve := pa[0]
		go func() {
			svg, warnings, err := gogv.Render(context.Background(), source, "svg")
			resolve.Invoke(result(svg, warnings, err))
		}()
		return nil
	})
	return js.Global().Get("Promise").New(executor)
}
