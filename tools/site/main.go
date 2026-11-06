// SPDX-License-Identifier: MIT
// Provenance: gopuml tools/site/main.go

// Command site assembles the landing page and the WebAssembly module of its
// playground into a directory that GitHub Pages or any static file server
// can serve.
//
// It copies the page (site/ of the repository -root), the picture of its
// example (docs/example.svg) and, next to the module in wasm/, the LICENSE
// and NOTICE of the repository. It builds the module from internal/playground
// with the wasm_exec.js of the toolchain that built it: from the working tree
// by default, or with -release from the sources of the release tag that the
// page names in its gogv-release meta tag, checked out in a temporary git
// worktree. When the repository has no such tag, or the sources of the tag
// have no playground, it prints a notice, writes nothing and succeeds.
// Nothing is written when a step fails.
//
// When GITHUB_OUTPUT is set, it appends deploy=true, or deploy=false when it
// wrote nothing, to the file it names.
//
// Usage: go run ./tools/site -out dir [-release] [-root dir]
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// licenseFiles are the files of the repository root that go next to the
// module.
var licenseFiles = []string{"LICENSE", "NOTICE"}

// playgroundMain is the file whose presence in the sources of a release
// tells that the release has the playground.
const playgroundMain = "internal/playground/main.go"

var releaseMeta = regexp.MustCompile(`<meta name="gogv-release" content="(v\d+\.\d+\.\d+)">`)

// errNoModule reports a release whose module cannot be built: its tag is
// not in the repository yet, or its sources have no playground.
var errNoModule = errors.New("no module to build")

func main() {
	log.SetFlags(0)
	if err := run(os.Args[1:], os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func run(args []string, stdout io.Writer) error {
	fl := flag.NewFlagSet("site", flag.ContinueOnError)
	root := fl.String("root", ".", "repository root, with the page in site/")
	out := fl.String("out", "", "output directory, replaced by the site")
	release := fl.Bool("release", false, "build the module from the release tag that the page names")
	if err := fl.Parse(args); err != nil {
		return err
	}
	if *out == "" || fl.NArg() > 0 {
		return errors.New("usage: site -out dir [-release] [-root dir]")
	}

	files, err := readFiles(*root, map[string]string{"docs/example.svg": "example.svg"})
	if err != nil {
		return err
	}
	module, err := readFiles(*root, map[string]string{"LICENSE": "wasm/LICENSE", "NOTICE": "wasm/NOTICE"})
	if err != nil {
		return err
	}
	maps.Copy(files, module)
	what := "the working tree"
	if *release {
		var tag string
		if tag, err = pageRelease(*root); err == nil {
			what = tag
			module, err = buildRelease(*root, tag)
		}
	} else {
		module, err = buildModule(*root, "devel")
	}
	if errors.Is(err, errNoModule) {
		fmt.Fprintf(stdout, "::notice::%v; nothing to deploy\n", err)
		return setOutput(false)
	}
	if err != nil {
		return err
	}
	maps.Copy(files, module)
	if err := write(filepath.Join(*root, "site"), *out, files); err != nil {
		os.RemoveAll(*out)
		return err
	}
	fmt.Fprintf(stdout, "site of %s written to %s\n", what, *out)
	return setOutput(true)
}

// pageRelease returns the release that the page of the repository in root
// names.
func pageRelease(root string) (string, error) {
	name := filepath.Join(root, "site", "index.html")
	page, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}
	m := releaseMeta.FindAllSubmatch(page, -1)
	if len(m) != 1 {
		return "", fmt.Errorf("%s: %d gogv-release meta tags, want 1", name, len(m))
	}
	return string(m[0][1]), nil
}

// git runs git in the repository root and returns its standard output.
func git(root string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", root}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		var stderr []byte
		if ee, ok := errors.AsType[*exec.ExitError](err); ok {
			stderr = ee.Stderr
		}
		return "", fmt.Errorf("git %s: %v\n%s", strings.Join(args, " "), err, stderr)
	}
	return string(out), nil
}

// buildRelease builds the module from the sources of the release tag, in a
// temporary worktree of the repository in root, with tag as its version.
func buildRelease(root, tag string) (map[string][]byte, error) {
	if _, err := git(root, "rev-parse", "--verify", "--quiet", "refs/tags/"+tag+"^{commit}"); err != nil {
		return nil, fmt.Errorf("%w: the repository has no tag %s yet", errNoModule, tag)
	}
	if _, err := git(root, "cat-file", "-e", tag+":"+playgroundMain); err != nil {
		return nil, fmt.Errorf("%w: the sources of %s have no %s", errNoModule, tag, playgroundMain)
	}
	tmp, err := os.MkdirTemp("", "site-release-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	src := filepath.Join(tmp, "src")
	if _, err := git(root, "worktree", "add", "--detach", src, tag); err != nil {
		return nil, err
	}
	defer git(root, "worktree", "remove", "--force", src)
	return buildModule(src, tag)
}

// readFiles reads files of dir, given by their path in dir and their path in
// the site.
func readFiles(dir string, names map[string]string) (map[string][]byte, error) {
	files := map[string][]byte{}
	for name, dest := range names {
		data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(name)))
		if err != nil {
			return nil, err
		}
		files[dest] = data
	}
	return files, nil
}

// buildModule builds internal/playground of the module in dir for
// WebAssembly, with version as the version it names, and returns it with
// its loader and the wasm_exec.js of the toolchain that built it, by their
// paths in the site.
func buildModule(dir, version string) (map[string][]byte, error) {
	tmp, err := os.MkdirTemp("", "site-module-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	wasm := filepath.Join(tmp, "gogv.wasm")
	build := exec.Command("go", "build", "-trimpath", "-ldflags=-X main.version="+version, "-o", wasm, "./internal/playground")
	build.Dir = dir
	build.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	if out, err := build.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("building the module: %v\n%s", err, out)
	}
	env := exec.Command("go", "env", "GOROOT")
	env.Dir = dir
	goroot, err := env.Output()
	if err != nil {
		return nil, fmt.Errorf("go env GOROOT: %v", err)
	}
	module, err := readFiles(dir, map[string]string{"internal/playground/gogv.mjs": "wasm/gogv.mjs"})
	if err != nil {
		return nil, err
	}
	for src, dest := range map[string]string{
		wasm: "wasm/gogv.wasm",
		filepath.Join(strings.TrimSpace(string(goroot)), "lib", "wasm", "wasm_exec.js"): "wasm/wasm_exec.js",
	} {
		if module[dest], err = os.ReadFile(src); err != nil {
			return nil, err
		}
	}
	return module, nil
}

// write replaces out with the files of src and the given files, by their
// paths in the site.
func write(src, out string, files map[string][]byte) error {
	if err := os.RemoveAll(out); err != nil {
		return err
	}
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(out, rel), 0o755)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(out, rel), data, 0o644)
	})
	if err != nil {
		return err
	}
	for name, data := range files {
		dest := filepath.Join(out, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dest, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}

// setOutput appends the deploy output of the step to the file that
// GITHUB_OUTPUT names, when it is set.
func setOutput(deploy bool) error {
	name := os.Getenv("GITHUB_OUTPUT")
	if name == "" {
		return nil
	}
	f, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(f, "deploy=%t\n", deploy); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
