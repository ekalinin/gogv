// SPDX-License-Identifier: MIT
// Provenance: gopuml cmd/plantuml-wasm/playground_test.go

//go:build !js

package main

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// needNode skips the test in short mode or when Node.js is not installed.
func needNode(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("runs Node.js")
	}
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node not found")
	}
}

// build builds the module into a temporary directory, next to gogv.mjs and
// the wasm_exec.js of the toolchain that built it, as the site serves them,
// and returns the directory.
func build(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cmd := exec.Command("go", "build", "-o", filepath.Join(dir, "gogv.wasm"), ".")
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	goroot, err := exec.Command("go", "env", "GOROOT").Output()
	if err != nil {
		t.Fatal(err)
	}
	for src, name := range map[string]string{
		"gogv.mjs": "gogv.mjs",
		filepath.Join(strings.TrimSpace(string(goroot)), "lib", "wasm", "wasm_exec.js"): "wasm_exec.js",
	} {
		data, err := os.ReadFile(src)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// runNode runs an ES module script in Node.js from dir and decodes the JSON
// it prints into v. The script imports the playground of the landing page
// as PLAYGROUND.
func runNode(t *testing.T, dir, script string, v any) {
	t.Helper()
	playground, err := filepath.Abs(filepath.Join("..", "..", "site", "playground.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	script = strings.ReplaceAll(script, "PLAYGROUND", "file://"+filepath.ToSlash(playground))
	name := filepath.Join(dir, "playground-run.mjs")
	if err := os.WriteFile(name, []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("node", name)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		var stderr string
		if ee, ok := errors.AsType[*exec.ExitError](err); ok {
			stderr = string(ee.Stderr)
		}
		t.Fatalf("node: %v\n%s", err, stderr)
	}
	if err := json.Unmarshal(out, v); err != nil {
		t.Fatalf("output %q: %v", out, err)
	}
}

// playgroundResult is a render result of the playground, with its SVG.
type playgroundResult struct {
	SVG      string   `json:"svg"`
	Type     string   `json:"type"`
	Warnings []string `json:"warnings"`
	Error    *struct {
		Message string `json:"message"`
		Line    int    `json:"line"`
		Text    string `json:"text"`
	} `json:"error"`
}

// textY returns the y coordinate of the text of the node titled name in
// an SVG of gogv, and whether the node is there.
func textY(svg, name string) (float64, bool) {
	m := regexp.MustCompile(`(?s)<g id="node\d+" class="node">\s*<title>` + regexp.QuoteMeta(name) + `</title>.*?<text [^>]*y="([-0-9.]+)"`).FindStringSubmatch(svg)
	if m == nil {
		return 0, false
	}
	y, err := strconv.ParseFloat(m[1], 64)
	return y, err == nil
}

func TestPlaygroundScenarios(t *testing.T) {
	// Spec scenarios "Simple graph", "Feature not supported yet", "Syntax
	// error", "Example graph" and "Download" of landing-page, with the
	// module of this build. The DOM part is checked by hand in a browser.
	needNode(t)
	dir := build(t)
	script := `import { load } from './gogv.mjs';
import { createPlayground, examples, download } from 'PLAYGROUND';
const pg = createPlayground(load);
const show = async (r) => ({
  svg: r.svg ? await r.svg.text() : '',
  type: r.svg ? r.svg.type : '',
  warnings: r.warnings,
  error: r.error,
});
const out = { examples: {} };
for (const e of examples) {
  out.examples[e.name] = await show(await pg.render(e.source));
}
const simple = await pg.render('digraph { a -> b }');
out.simple = await show(simple);
out.record = await show(await pg.render('digraph { a [shape=record] }'));
const syntax = await pg.render('digraph {\n  a -> b\n  c -> -> d\n}\n');
out.syntax = await show(syntax);
const file = download(simple);
out.download = { filename: file.filename, type: file.blob.type, svg: (await file.blob.text()).slice(0, 5) };
out.downloadError = download(syntax);
out.status = pg.status;
out.version = pg.version;
console.log(JSON.stringify(out));
process.exit(0);
`
	var out struct {
		Examples      map[string]playgroundResult `json:"examples"`
		Simple        playgroundResult            `json:"simple"`
		Record        playgroundResult            `json:"record"`
		Syntax        playgroundResult            `json:"syntax"`
		Download      struct{ Filename, Type, SVG string }
		DownloadError any    `json:"downloadError"`
		Status        string `json:"status"`
		Version       string `json:"version"`
	}
	runNode(t, dir, script, &out)

	t.Run("Simple graph", func(t *testing.T) {
		r := out.Simple
		if r.Error != nil || len(r.Warnings) != 0 || r.Type != "image/svg+xml" || !strings.HasPrefix(r.SVG, "<?xml") {
			t.Fatalf("error %+v, warnings %q, type %q, svg %.40q", r.Error, r.Warnings, r.Type, r.SVG)
		}
		ya, okA := textY(r.SVG, "a")
		yb, okB := textY(r.SVG, "b")
		if !okA || !okB || yb <= ya {
			t.Errorf("node a at y %v (%v), node b at y %v (%v): want b below a", ya, okA, yb, okB)
		}
		edge := regexp.MustCompile(`(?s)<g id="edge1" class="edge">\s*<title>a&#45;&gt;b</title>.*?<path .*?<polygon `).MatchString(r.SVG)
		if !edge {
			t.Error("no edge a->b with a path and an arrowhead")
		}
	})

	t.Run("Feature not supported yet", func(t *testing.T) {
		r := out.Record
		want := []string{"Warning: node a: shape=record is not supported yet, using ellipse"}
		if r.Error != nil || !strings.HasPrefix(r.SVG, "<?xml") || strings.Join(r.Warnings, "\n") != strings.Join(want, "\n") {
			t.Errorf("error %+v, warnings %q, svg %.40q; want the graph and %q", r.Error, r.Warnings, r.SVG, want)
		}
	})

	t.Run("Syntax error", func(t *testing.T) {
		r := out.Syntax
		if r.Error == nil || r.Error.Line != 3 || !strings.Contains(r.Error.Text, "line 3") || r.SVG != "" {
			t.Errorf("error %+v, svg %.40q; want the message of line 3 and no picture", r.Error, r.SVG)
		}
	})

	t.Run("Example graph", func(t *testing.T) {
		if len(out.Examples) != 5 {
			t.Errorf("%d examples, want 5", len(out.Examples))
		}
		for name, r := range out.Examples {
			if r.Error != nil || len(r.Warnings) != 0 || !strings.HasPrefix(r.SVG, "<?xml") {
				t.Errorf("example %q: error %+v, warnings %q, svg %.40q", name, r.Error, r.Warnings, r.SVG)
			}
		}
	})

	t.Run("Download", func(t *testing.T) {
		if d := out.Download; d.Filename != "graph.svg" || d.Type != "image/svg+xml" || d.SVG != "<?xml" {
			t.Errorf("download: %+v", d)
		}
		if out.DownloadError != nil {
			t.Errorf("download of an error: %v, want nothing", out.DownloadError)
		}
	})

	if out.Status != "ready" || out.Version != "devel" {
		t.Errorf("status %q, version %q; want ready and devel", out.Status, out.Version)
	}
}

func TestPlaygroundLoadingScenarios(t *testing.T) {
	// Spec scenarios "Reading the top of the page" and "Module not
	// available" of landing-page, with fake loaders: nothing is loaded
	// before the playground is activated, a failed load is reported, and
	// only the latest render is kept.
	needNode(t)
	script := `import { createPlayground, describe } from 'PLAYGROUND';
const out = {};
let loads = 0;
const fake = {
  version: 'v9.9.9',
  render: (source) => new Promise((resolve) => setTimeout(() => resolve({
    svg: new TextEncoder().encode('<svg>' + source + '</svg>'), warnings: [], error: null,
  }), source === 'slow' ? 50 : 0)),
};
const pg = createPlayground(async () => { loads++; return fake; });
out.before = { loads, status: pg.status, version: pg.version };
const first = pg.activate();
out.loading = pg.status;
await Promise.all([first, pg.activate()]);
out.after = { loads, status: pg.status, version: pg.version };
const [slow, fast] = await Promise.all([pg.render('slow'), pg.render('fast')]);
out.slow = slow;
out.fast = fast && await fast.svg.text();

const failing = createPlayground(async () => { throw new Error('404 gogv.wasm'); });
out.failed = await failing.activate().then(() => 'loaded', (e) => e.message);
out.failedStatus = failing.status;
out.failedRender = await failing.render('x').then(() => 'rendered', (e) => e.message);

out.describe = [
  describe({ message: "syntax error in line 3 near '->'", line: 3 }),
  describe({ message: 'syntax error', line: 2 }),
];
console.log(JSON.stringify(out));
process.exit(0);
`
	var out struct {
		Before struct {
			Loads   int
			Status  string
			Version *string
		}
		Loading string
		After   struct {
			Loads   int
			Status  string
			Version *string
		}
		Slow         any
		Fast         string
		Failed       string
		FailedStatus string   `json:"failedStatus"`
		FailedRender string   `json:"failedRender"`
		Describe     []string `json:"describe"`
	}
	runNode(t, t.TempDir(), script, &out)

	t.Run("Reading the top of the page", func(t *testing.T) {
		if out.Before.Loads != 0 || out.Before.Status != "idle" || out.Before.Version != nil {
			t.Errorf("before activation: %+v, want no load", out.Before)
		}
		if out.Loading != "loading" {
			t.Errorf("status while loading: %q", out.Loading)
		}
		if out.After.Loads != 1 || out.After.Status != "ready" || out.After.Version == nil || *out.After.Version != "v9.9.9" {
			t.Errorf("after two activations: %+v, want one load and the version of the module", out.After)
		}
		if out.Slow != nil || out.Fast != "<svg>fast</svg>" {
			t.Errorf("out-of-order renders: slow %v, fast %q; want only the latest", out.Slow, out.Fast)
		}
	})

	t.Run("Module not available", func(t *testing.T) {
		if out.Failed != "404 gogv.wasm" || out.FailedStatus != "failed" || out.FailedRender != "404 gogv.wasm" {
			t.Errorf("failed load: %q, status %q, render %q", out.Failed, out.FailedStatus, out.FailedRender)
		}
	})

	want := []string{"syntax error in line 3 near '->'", "Line 2: syntax error"}
	if strings.Join(out.Describe, "\n") != strings.Join(want, "\n") {
		t.Errorf("describe: %q, want %q", out.Describe, want)
	}
}
