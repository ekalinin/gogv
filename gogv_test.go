// SPDX-License-Identifier: MIT
// Provenance: original

package gogv

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
)

func TestRenderSVG(t *testing.T) {
	out, warnings, err := Render(context.Background(), "digraph { a -> b }", "svg")
	if err != nil || len(warnings) != 0 || !bytes.Contains(out, []byte("<title>a&#45;&gt;b</title>")) {
		t.Fatalf("out %s, warnings %v, err %v", out, warnings, err)
	}
}

func TestRenderSyntaxError(t *testing.T) {
	_, _, err := Render(context.Background(), "digraph { a -> }", "svg")
	var se *SyntaxError
	if !errors.As(err, &se) || se.Line != 1 || se.Near != "}" || err.Error() != "syntax error in line 1 near '}'" {
		t.Fatalf("err %#v", err)
	}
}

func TestRenderFormats(t *testing.T) {
	var fe *FormatError
	if _, _, err := Render(context.Background(), "digraph {}", "xdot"); !errors.As(err, &fe) || !fe.NotSupportedYet {
		t.Errorf("xdot: %v", err)
	}
	if _, _, err := Render(context.Background(), "digraph {}", "foo"); !errors.As(err, &fe) || fe.NotSupportedYet ||
		err.Error() != `Format: "foo" not recognized. Use one of: canon plain svg` {
		t.Errorf("foo: %v", err)
	}
}

func TestRenderWarnings(t *testing.T) {
	out, warnings, err := Render(context.Background(), "digraph { a [shape=record] }", "svg")
	if err != nil || len(out) == 0 {
		t.Fatal(err)
	}
	if len(warnings) != 1 || warnings[0].Object != "node a" || warnings[0].Feature != "shape=record" ||
		warnings[0].Default != "ellipse" || !warnings[0].NotSupportedYet() {
		t.Fatalf("warnings %+v", warnings)
	}
}

func TestParallelRenders(t *testing.T) {
	src := "digraph { a -> b -> c; a -> c; b -> d; d -> a; c -> e; e -> f; a -> f }"
	want, _, err := Render(context.Background(), src, "svg")
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var bad int
	for range 100 {
		wg.Go(func() {
			got, _, err := Render(context.Background(), src, "svg")
			if err != nil || !bytes.Equal(got, want) {
				mu.Lock()
				bad++
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	if bad > 0 {
		t.Fatalf("%d of 100 parallel renders differ", bad)
	}
}

func TestCanceledRender(t *testing.T) {
	var b strings.Builder
	b.WriteString("digraph {\n")
	for i := range 300 {
		fmt.Fprintf(&b, "n%d -> n%d; n%d -> n%d;\n", i, i+1, (i*7)%300, (i*13)%300)
	}
	b.WriteString("}\n")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	out, _, err := Render(ctx, b.String(), "svg")
	if !errors.Is(err, context.Canceled) || len(out) != 0 {
		t.Fatalf("out %d bytes, err %v", len(out), err)
	}
}

// TestSelfContained checks that the packages of the library import nothing
// that starts processes, opens files, reads the environment or uses the
// network.
func TestSelfContained(t *testing.T) {
	cmd := exec.Command("go", "list", "-deps", "-json", ".")
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	forbidden := []string{"os", "os/exec", "net", "net/http", "syscall", "io/fs", "path/filepath"}
	dec := json.NewDecoder(bytes.NewReader(out))
	for {
		var p struct {
			ImportPath string
			Imports    []string
			Standard   bool
		}
		if err := dec.Decode(&p); err != nil {
			break
		}
		if p.Standard {
			continue
		}
		for _, imp := range p.Imports {
			if slices.Contains(forbidden, imp) {
				t.Errorf("%s imports %s", p.ImportPath, imp)
			}
		}
	}
}

// TestWebAssembly renders a graph in the GOOS=js GOARCH=wasm build under
// Node.js and compares it with the native output.
func TestWebAssembly(t *testing.T) {
	if testing.Short() {
		t.Skip("builds and runs a WebAssembly module")
	}
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node not found")
	}
	goroot, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		t.Fatal(err)
	}
	runner := filepath.Join(strings.TrimSpace(string(goroot)), "lib", "wasm", "go_js_wasm_exec")
	wasm := filepath.Join(t.TempDir(), "check.wasm")
	build := exec.Command("go", "build", "-o", wasm, "./internal/wasmcheck")
	build.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	src := "digraph { a -> b -> c; a -> c; c -> a }"
	run := exec.Command(runner, wasm, src)
	// The WebAssembly runtime passes the environment to the module, which
	// limits its size: keep it small.
	run.Env = []string{"PATH=" + os.Getenv("PATH")}
	got, err := run.Output()
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	want, _, err := Render(context.Background(), src, "svg")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("WebAssembly output differs:\n%s\nnative:\n%s", got, want)
	}
}
