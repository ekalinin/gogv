// SPDX-License-Identifier: MIT
// Provenance: original

package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// gitOrFail runs git in dir, without the signing and the hooks of the
// user's configuration.
func gitOrFail(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=site", "-c", "user.email=site@example.com", "-c", "commit.gpgsign=false", "-c", "tag.gpgsign=false", "-c", "core.hooksPath=/dev/null"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

// writeFiles writes files into dir, by their slash-separated paths.
func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, data := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// page returns a page that names release.
func page(release, text string) string {
	return `<meta name="gogv-release" content="` + release + `">` + text + "\n"
}

// fakeRepo creates a git repository with a page that names v1.0.0, the
// files that the site takes from the repository, and, when playground is
// set, a minimal playground module. It commits them, and tags the commit
// v1.0.0 when tag is set.
func fakeRepo(t *testing.T, playground, tag bool) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"site/index.html":  page("v1.0.0", "tagged page"),
		"docs/example.svg": "<svg/>\n",
		"LICENSE":          "license of the repository\n",
		"NOTICE":           "notice of the repository\n",
		"go.mod":           "module example.com/fake\n\ngo 1.26\n",
	}
	if playground {
		files["internal/playground/main.go"] = "//go:build js && wasm\n\npackage main\n\nvar version = \"devel\"\n\nfunc main() { println(\"module \" + version) }\n"
		files["internal/playground/gogv.mjs"] = "// loader\n"
	}
	writeFiles(t, dir, files)
	gitOrFail(t, dir, "init", "-q", "-b", "main")
	gitOrFail(t, dir, "add", ".")
	gitOrFail(t, dir, "commit", "-q", "-m", "fake")
	if tag {
		gitOrFail(t, dir, "tag", "-a", "v1.0.0", "-m", "v1.0.0")
	}
	return dir
}

// runRelease runs the tool with -release on the repository in dir, and
// returns its output directory, its standard output and its GITHUB_OUTPUT.
func runRelease(t *testing.T, dir string) (string, string, string) {
	t.Helper()
	out := filepath.Join(t.TempDir(), "site")
	output := filepath.Join(t.TempDir(), "output")
	t.Setenv("GITHUB_OUTPUT", output)
	var stdout bytes.Buffer
	if err := run([]string{"-release", "-root", dir, "-out", out}, &stdout); err != nil {
		t.Fatal(err)
	}
	return out, stdout.String(), string(readOrFail(t, output))
}

func TestSiteReleaseScenarios(t *testing.T) {
	// Spec scenarios "Release tag pushed after the notes" (its first half,
	// before the tag is pushed), "Release without the playground", "Change
	// of the page" and "License files on the site" of landing-page, for the
	// build of a release.
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found")
	}

	t.Run("Release tag pushed after the notes", func(t *testing.T) {
		out, stdout, output := runRelease(t, fakeRepo(t, true, false))
		if !strings.HasPrefix(stdout, "::notice::") || !strings.Contains(stdout, "no tag v1.0.0") {
			t.Errorf("stdout %q, want a notice that the tag is missing", stdout)
		}
		if output != "deploy=false\n" {
			t.Errorf("GITHUB_OUTPUT %q, want deploy=false", output)
		}
		if _, err := os.Stat(out); !os.IsNotExist(err) {
			t.Errorf("%s was written: %v", out, err)
		}
	})

	t.Run("Release without the playground", func(t *testing.T) {
		out, stdout, output := runRelease(t, fakeRepo(t, false, true))
		if !strings.HasPrefix(stdout, "::notice::") || !strings.Contains(stdout, "have no "+playgroundMain) {
			t.Errorf("stdout %q, want a notice that the release has no playground", stdout)
		}
		if output != "deploy=false\n" {
			t.Errorf("GITHUB_OUTPUT %q, want deploy=false", output)
		}
		if _, err := os.Stat(out); !os.IsNotExist(err) {
			t.Errorf("%s was written: %v", out, err)
		}
	})

	if testing.Short() {
		t.Skip("builds a WebAssembly module")
	}
	dir := fakeRepo(t, true, true)
	// A commit after the tag changes the page and the notices: the site takes
	// them from the working tree, and only the module from the tag.
	writeFiles(t, dir, map[string]string{
		"site/index.html":              page("v1.0.0", "page of main"),
		"NOTICE":                       "notice of main\n",
		"internal/playground/gogv.mjs": "// loader of main\n",
	})
	gitOrFail(t, dir, "commit", "-q", "-am", "after the tag")
	out, stdout, output := runRelease(t, dir)

	t.Run("Change of the page", func(t *testing.T) {
		if output != "deploy=true\n" || !strings.Contains(stdout, "site of v1.0.0 written") {
			t.Errorf("stdout %q, GITHUB_OUTPUT %q; want the site of v1.0.0 and deploy=true", stdout, output)
		}
		if got := string(readOrFail(t, filepath.Join(out, "index.html"))); !strings.Contains(got, "page of main") {
			t.Errorf("index.html %q, want the page of the working tree", got)
		}
		if wasm := readOrFail(t, filepath.Join(out, "wasm", "gogv.wasm")); !bytes.HasPrefix(wasm, []byte("\x00asm")) || !bytes.Contains(wasm, []byte("v1.0.0")) {
			t.Error("wasm/gogv.wasm is not a module that names the version v1.0.0")
		}
		if got := string(readOrFail(t, filepath.Join(out, "wasm", "gogv.mjs"))); got != "// loader\n" {
			t.Errorf("wasm/gogv.mjs %q, want the loader of the tag", got)
		}
		if !bytes.Contains(readOrFail(t, filepath.Join(out, "wasm", "wasm_exec.js")), []byte("globalThis.Go")) {
			t.Error("wasm/wasm_exec.js is not the runtime of Go")
		}
		if worktrees := gitOrFail(t, dir, "worktree", "list"); strings.Count(worktrees, "\n") != 1 {
			t.Errorf("worktrees left behind:\n%s", worktrees)
		}
	})

	t.Run("License files on the site", func(t *testing.T) {
		for _, name := range licenseFiles {
			if !bytes.Equal(readOrFail(t, filepath.Join(out, "wasm", name)), readOrFail(t, filepath.Join(dir, name))) {
				t.Errorf("wasm/%s differs from %s of the repository", name, name)
			}
		}
	})
}
