// SPDX-License-Identifier: MIT
// Provenance: original

package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// repoRoot is the root of the repository, from the directory of the test.
const repoRoot = "../.."

// readOrFail returns the content of a file.
func readOrFail(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestSiteLocalScenarios(t *testing.T) {
	// Spec scenarios "Preview" and "License files on the site" of
	// landing-page, for a local preview: the site of the working tree.
	if testing.Short() {
		t.Skip("builds the WebAssembly module")
	}
	out := filepath.Join(t.TempDir(), "site")
	t.Setenv("GITHUB_OUTPUT", filepath.Join(t.TempDir(), "output"))
	if err := run([]string{"-root", repoRoot, "-out", out}, io.Discard); err != nil {
		t.Fatal(err)
	}

	t.Run("Preview", func(t *testing.T) {
		pages, err := os.ReadDir(filepath.Join(repoRoot, "site"))
		if err != nil || len(pages) == 0 {
			t.Fatalf("site/: %v", err)
		}
		for _, e := range pages {
			if !bytes.Equal(readOrFail(t, filepath.Join(out, e.Name())), readOrFail(t, filepath.Join(repoRoot, "site", e.Name()))) {
				t.Errorf("%s differs from site/%s", e.Name(), e.Name())
			}
		}
		if !bytes.Equal(readOrFail(t, filepath.Join(out, "example.svg")), readOrFail(t, filepath.Join(repoRoot, "docs", "example.svg"))) {
			t.Error("example.svg differs from docs/example.svg")
		}
		if !bytes.Equal(readOrFail(t, filepath.Join(out, "wasm", "gogv.mjs")), readOrFail(t, filepath.Join(repoRoot, "internal", "playground", "gogv.mjs"))) {
			t.Error("wasm/gogv.mjs differs from internal/playground/gogv.mjs")
		}
		if wasm := readOrFail(t, filepath.Join(out, "wasm", "gogv.wasm")); !bytes.HasPrefix(wasm, []byte("\x00asm")) {
			t.Error("wasm/gogv.wasm is not a WebAssembly module")
		} else if !bytes.Contains(wasm, []byte("devel")) {
			t.Error("the module of a local build does not name the version devel")
		}
		if !bytes.Contains(readOrFail(t, filepath.Join(out, "wasm", "wasm_exec.js")), []byte("globalThis.Go")) {
			t.Error("wasm/wasm_exec.js is not the runtime of Go")
		}
		if got := string(readOrFail(t, os.Getenv("GITHUB_OUTPUT"))); got != "deploy=true\n" {
			t.Errorf("GITHUB_OUTPUT %q, want deploy=true", got)
		}
	})

	t.Run("License files on the site", func(t *testing.T) {
		for _, name := range licenseFiles {
			if !bytes.Equal(readOrFail(t, filepath.Join(out, "wasm", name)), readOrFail(t, filepath.Join(repoRoot, name))) {
				t.Errorf("wasm/%s differs from %s of the repository", name, name)
			}
		}
	})
}
