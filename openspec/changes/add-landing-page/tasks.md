# Tasks

## 1. Playground module

- [x] 1.1 Add `internal/playground` (design D3): `main.go`, for `js && wasm`, with `gogv.version` and `gogv.render`, and `gogv.mjs` with `load(wasmURL)`. Verify:
  - `GOOS=js GOARCH=wasm go build ./internal/playground` succeeds;
  - `go run ./tools/provenance` accepts `main.go`;
  - a Node.js run of `load` renders `digraph { a -> b }` as SVG, with no warning and no error.

## 2. Page

- [x] 2.1 Write `site/index.html` and `site/style.css` (design D1, D2):
  - the sections of the "Project presentation" requirement;
  - the `gogv-release` meta tag, the installation at `v0.3.0` with the line of `dot -V`, and the conformance results of `docs/releases/v0.3.0.md`;
  - the example of `docs/example.gv`, inline, and `example.svg`;
  - the playground markup with its `<noscript>` message;
  - the links to the repository, the releases, pkg.go.dev and the Status lists of the README.

  Add the presentation tests of `landing_test.go` for "Visit the page", "JavaScript disabled", "Output formats", "Example graph" and "Installation". Verify that they pass, and that the page, opened from the file with JavaScript disabled, shows every section except the playground.
- [x] 2.2 Add the release tests of `landing_test.go` (design D2) for "Page behind the notes" and "Conformance results out of date". Mutations of the page must make the checks fail: another tag, a changed number. Verify that the tests pass and that each mutation makes them fail.
- [x] 2.3 Write `site/playground.mjs` (design D6): the controller created with a `load` function, the examples, the warnings, the error, the version and the download. Add `internal/playground/playground_test.go`, which builds the module and runs Node.js on it:
  - with the real loader: "Simple graph", "Feature not supported yet", "Syntax error", "Example graph" for every example, and "Download";
  - with fake loaders: "Reading the top of the page", where nothing is loaded before activation, and "Module not available".

  It skips without `node` or with `-short`. Check in `landing_test.go` that the first example is `docs/example.gv`. Verify that the tests pass with `node` installed, and that `go test -v ./internal/playground` runs them rather than skipping them.

## 3. Site assembly

- [x] 3.1 Add `tools/site` with the local build (design D4), and the `site` target of the Makefile, which writes `.cache/site`. Add tests for "Preview" and "License files on the site". Verify that they pass, and that `make site` writes the following into `.cache/site`:
  - `index.html` and `example.svg`;
  - `wasm/` with `gogv.wasm`, `wasm_exec.js`, `gogv.mjs`, `LICENSE` and `NOTICE`.
- [x] 3.2 Add the `-release` build (design D4), and its tests on temporary git repositories:
  - a missing tag: a notice, no output, `deploy=false`, exit status 0 (the first half of "Release tag pushed after the notes");
  - a tag without the playground: the same ("Release without the playground");
  - a tag with a minimal playground module: the site with the module, its version set to the tag, and `deploy=true` ("Change of the page");
  - the license files with a release build.

  Verify that the tests pass. Also verify that a run on this repository, whose page names `v0.3.0`, prints the notice and writes nothing.

## 4. Playground in the browser

- [x] 4.1 Write `site/page.mjs` (design D6): the lazy loading, the loading and failure messages, the rendering 300 ms after the last edit, the SVG in an `<img>` from a `blob:` URL, the warnings, the error, the example selector, the version and the download link. Verify by hand in a browser, on the output of `make site` served with `python3 -m http.server -d .cache/site`:
  - the network panel shows no WebAssembly request before the playground is reached;
  - the result follows the edits;
  - `shape=record` shows its warning;
  - a syntax error shows its message;
  - an example replaces the source;
  - the download saves `graph.svg`;
  - the playground names its version;
  - with `wasm/gogv.wasm` removed, the playground says that it could not be loaded, and the rest of the page works;
  - the page reads well in the light and the dark color schemes.

## 5. Deployment

- [x] 5.1 Add `.github/workflows/pages.yml` (design D5), and `TestPagesWorkflow` in the root package, for the tag trigger of "Release tag pushed after the notes" and for "Change of the notices". Each of these mutations must make it fail:
  - no trigger on tags;
  - a `pull_request` trigger;
  - no `id-token: write`;
  - a deployment step not conditional on `deploy=true`;
  - a checkout without `ref: main`;
  - `NOTICE` missing from the paths.

  Verify that the test passes and that each mutation makes it fail.

## 6. Documentation

- [x] 6.1 Update `README.md`:
  - add a link to the page and its playground near the top;
  - in step 1 of the Releasing section, name the release on the page in the pull request of its notes.

  Verify that `TestReadmeExample` and the tests of `landing_test.go` pass.

## 7. Checks and publication

- [x] 7.1 Run the checks, with `node` installed:
  - `make lint`;
  - `go test -race ./...`;
  - `go run ./tools/provenance`;
  - `make determinism`;
  - `openspec validate add-landing-page --strict`.

  Verify that all pass, that the playground tests ran rather than skipped, and that the hash of `make determinism` is the same as on `main`.
- [x] 7.2 Push the branch and open a pull request with the template, with the user's approval. Verify that the CI jobs pass.
- [ ] 7.3 After the merge, the user enables GitHub Pages with "GitHub Actions" as the source, and allows the tags `v*` to deploy to the `github-pages` environment (design, Migration Plan). The first release that contains this change then pushes its tag. Verify:
  - that the tag push starts the workflow and deploys, and record the answer in the Open Questions of the design;
  - that `https://ekalinin.github.io/gogv/` shows the page, and that its playground renders `digraph { a -> b }`;
  - whether `gogv.wasm` is served with `content-encoding: gzip`.

  Then set the page as the homepage of the repository, with the user's approval.
