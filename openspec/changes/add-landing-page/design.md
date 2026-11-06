# Design

## Context

See proposal.md for the motivation, and specs/landing-page/spec.md for the behavior. The current state that shapes the approach:

- gopuml's landing page:
  - static files in `site/` (`index.html`, `style.css`, `page.mjs`, `playground.mjs`), assembled by `tools/site` and deployed by a `pages` workflow;
  - its playground runs `cmd/plantuml-wasm`, a WebAssembly entry point that defines a global `render` returning a Promise, loaded by `plantuml.mjs`;
  - the decisions behind it are in gopuml's archived change `add-landing-page`.

  The code of gopuml is an allowed source.
- gogv releases publish no binaries: a release is a tag on `main` with notes in `docs/releases/`, made by `make release` from a local checkout. A tag pushed with the maintainer's credentials starts workflows, unlike gopuml's GoReleaser release made with `GITHUB_TOKEN`.
- No tag has a WebAssembly entry point with a JavaScript API. `internal/wasmcheck` only writes SVG to standard output for `TestWebAssembly`.
- A WebAssembly build of the library is 4.6 MB, 1.3 MB with gzip. gopuml found that GitHub Pages serves `.wasm` files with gzip.
- The public API is enough for a playground:
  - `gogv.Render` returns the output, the warnings and an error;
  - `Warning.String` words a warning as the command line does;
  - `SyntaxError.Line` is the line of a syntax error.
- The README installs with `@latest`, and names no release. The notes of each release state the conformance results in one sentence: `N cases: P passed, K known differences, S not supported yet, F failed`.
- The module has no dependencies, so its tests cannot parse YAML. CI runners have Node.js, which `TestWebAssembly` already uses.
- The repository is public, and GitHub Pages is not enabled yet.

## Goals / Non-Goals

**Goals:**

- One release on the page, the same in its text and in its playground, kept consistent with the notes by tests.
- The files in `site/` are the published files: no build step for the page itself.
- The playground uses only the public API of the library, so it shows what a Go program gets.

**Non-Goals:**

- A JavaScript API of gogv for other programs: the entry point of the playground is internal.
- Running lint and tests in the `pages` workflow: what it deploys comes from `main`, which CI checks.

## Decisions

### D1. Files of the page

`site/` holds the published files:

- `index.html`: every section of the page, in static HTML. The playground has a `<noscript>` message. The example graph is written inline as source, and shown as `example.svg`.
- `style.css`: adapted from gopuml's, with the system font stack and light and dark color schemes.
- `playground.mjs`: the logic of the playground without the DOM, so that Node.js can test it (D6).
- `page.mjs`: the DOM glue: the editor, the example selector, the result, the warnings, the error, the download link, the lazy loading.

`tools/site` copies `docs/example.svg` into the site as `example.svg`, so that the picture stays the one that `TestReadmeExample` checks. The HTML, JavaScript and CSS files start with SPDX and `Provenance` comments, as Go files do. A file adapted from gopuml names its file, as in `Provenance: gopuml site/style.css`. `tools/provenance` checks Go files only.

Alternative considered: a site generator such as Hugo or Jekyll. Rejected: it adds a tool and a build step for one page.

### D2. One release named by the page

`index.html` names its release in `<meta name="gogv-release" content="v0.3.0">`, in its installation commands and with its conformance results. `landing_test.go` (root package) checks:

- that the meta tag names the newest release of `docs/releases/`, by semantic version;
- that the installation commands name it;
- that the conformance results equal those of the sentence in its notes, read with whitespace normalized.

Mutations of the page must make the checks fail: another tag, a changed number.

The pull request of the notes of a release therefore updates the page. The Releasing section of the README says so. CI fails on that pull request otherwise, and `make release` requires CI to pass.

Alternative considered: a WebAssembly build of `main`. Rejected: it would run unreleased code under the name of a release, and the playground would disagree with the installation section.

### D3. Module of the playground

`internal/playground` holds the entry point (package `main`, built for `js && wasm`) and `gogv.mjs`:

- `main.go` defines the global `gogv` with:
  - `version`, set with `-ldflags=-X main.version=<tag>`, else `devel`;
  - `render(source)`, which returns a Promise of `{ svg, warnings, error }`:
    - `svg` is a `Uint8Array`, or `null` after an error;
    - `warnings` is the list of the texts of the warnings;
    - `error` is `null`, or `{ message, line }`.

  It renders in a goroutine, as gopuml's entry point does, and is adapted from it (`Provenance: gopuml cmd/plantuml-wasm/main.go`).
- `gogv.mjs` exports `load(wasmURL)`, which loads `wasm_exec.js` next to the module, runs it, and resolves to the `gogv` object. It is adapted from gopuml's `plantuml.mjs`.

Version 0.3.0 has no such entry point. The first release that contains this change is the first one the playground can run.

Alternative considered: `cmd/gogv-wasm` as a public command. Rejected: a JavaScript API is a public interface with its own compatibility promise, out of the scope of a landing page.

### D4. Site assembly tool

`tools/site` is an original Go command, using the standard library only, that writes the site into `-out`. It copies:

- the files of `site/`;
- `docs/example.svg`, as `example.svg`;
- `LICENSE` and `NOTICE`, into `wasm/`;
- the module, into `wasm/`.

The module comes from one of two builds:

- **Local (default):** it builds `internal/playground` of the working tree with `GOOS=js GOARCH=wasm`, and copies `gogv.mjs` and the `wasm_exec.js` of the toolchain that built it (`go env GOROOT`). `make site` runs it with `-out .cache/site`.
- **`-release`:** it reads the tag from the meta tag, and checks that the tag has the playground:
  - When the repository has no tag `<tag>`, or `<tag>:internal/playground/main.go` does not exist, it prints a `::notice::` line and writes no output. It writes `deploy=false` to the file that `GITHUB_OUTPUT` names, when set, and exits with status 0.
  - Otherwise, it checks out the tag in a temporary `git worktree`. It builds there with `-X main.version=<tag>`, using the toolchain that the tag's `go.mod` selects, and takes `gogv.mjs` and `wasm_exec.js` from that worktree and that toolchain. Then it removes the worktree and writes `deploy=true`.

Its tests run against temporary git repositories:

- a missing tag;
- a tag without the playground;
- a tag with a minimal playground module;
- a local build.

Alternative considered: shell steps in the workflow. Rejected: they would be tested only by running them on GitHub.

### D5. Pages workflow

`.github/workflows/pages.yml`:

- Triggers:
  - `push` to `main` with the paths `site/**`, `tools/site/**`, `docs/example.svg`, `LICENSE`, `NOTICE` and the workflow file;
  - `push` of tags `v*`, for which GitHub does not evaluate path filters;
  - `workflow_dispatch`.
- Permissions: `contents: read`, `pages: write`, `id-token: write`. Concurrency group `pages`, without cancelling a deployment in progress.
- One job in the `github-pages` environment:
  1. `actions/checkout` with `ref: main` and `fetch-depth: 0`, so that the tags are there. The page always comes from `main`, also on a tag push.
  2. `actions/setup-go` from `go.mod`.
  3. `go run ./tools/site -release -out _site`.
  4. `actions/configure-pages`, `actions/upload-pages-artifact` (path `_site`) and `actions/deploy-pages`, these three only when the tool wrote `deploy=true`.

`TestPagesWorkflow` checks the lines of these triggers, permissions and conditions as text, since the module has no YAML parser. A mutation of each must make it fail.

### D6. Playground

- **Lazy loading:** `page.mjs` calls `load()` only when an `IntersectionObserver` reports the playground as visible, or when the editor gets the focus. While the module loads, the result area says so. When `load()` rejects, it says that the playground could not be loaded.
- **Rendering:** the source is rendered 300 ms after the last edit, and a result that arrives after a newer edit is dropped.
- **Display:** the SVG is shown in an `<img>` from a `blob:` URL of type `image/svg+xml`, not inserted into the page, so that nothing in it runs in the page.
- **Warnings:** each warning is shown under the result, prefixed with `Warning: `, as the command line writes it.
- **Errors:** the message of the error, which names its line, as `syntax error in line 3 near '->'` does. No picture is shown.
- **Examples:**
  - the example of the README, whose source `landing_test.go` checks against `docs/example.gv`;
  - clusters;
  - `rankdir=LR` with edge labels;
  - a table label with ports;
  - a call graph in the style of `go tool pprof`.

  The first is the initial source.
- **Version:** the playground names `gogv.version`.
- **Download:** a link with the `download` attribute (`graph.svg`) on the `blob:` URL of the last successful SVG.
- **Tests:** `playground.mjs` exports a controller created with a `load` function, so that the tests can pass the real loader or a fake one.

### D7. Tests

- `landing_test.go` (root package):
  - the presentation: the sections of the page in its static HTML, the `<noscript>` of the playground, the output formats equal to `Formats()` and in its order, the example, the links;
  - the release checks of D2;
  - `TestPagesWorkflow`.
- `internal/playground/playground_test.go` builds the module, as `TestWebAssembly` does. It runs a Node.js script on `site/playground.mjs`:
  - with the real loader: the simple graph, the warning of `shape=record`, the syntax error on line 3, every example without a warning, the download;
  - with fake loaders: no load before activation, and a rejected load.

  It skips without `node`, or with `-short`, as `TestWebAssembly` does.
- `tools/site`: the tests of D4, including the license files.

The DOM glue of `page.mjs` (the observer, the delay, the `<img>`) is checked by hand in a browser, on the output of `make site`.

## Risks / Trade-offs

- [The first deployment waits for a release] → Version 0.3.0 has no playground module, so the workflow deploys nothing, with a notice, until the next release. That release can follow the merge of this change, since this change adds a feature.
- [A long rendering blocks the page] → WebAssembly runs on the main thread, so a large graph freezes the tab while it is laid out. The examples are small, and moving the module to a Web Worker later changes no requirement.
- [Size of the module] → 1.3 MB to download with gzip, and only when the playground is reached.
- [Page behind after a release] → the tests of D2 fail the pull request of the notes until the page names the new release.
- [`wasm_exec.js` from another toolchain] → it would not run the module. Both come from the toolchain that built the module, in the worktree of the tag.
- [Node.js not installed locally] → the playground tests skip. CI runs them, and the skip shows in `go test -v`.
- [Repository settings] → the workflow cannot enable GitHub Pages, allow the release tags to deploy to the `github-pages` environment, or set the homepage. They are tasks for the maintainer.

## Migration Plan

1. Merge the pull request of this change. The page names v0.3.0, which has no playground module, so the workflow deploys nothing, with a notice.
2. The maintainer enables GitHub Pages with "GitHub Actions" as the source (`gh api -X POST repos/ekalinin/gogv/pages -f build_type=workflow`). The `github-pages` environment that this creates lets only `main` deploy, so the maintainer also allows the release tags (`gh api -X POST repos/ekalinin/gogv/environments/github-pages/deployment-branch-policies -f name='v*' -f type=tag`).
3. Release the next version: its notes pull request names it on the page, and `make release` pushes its tag, which deploys the site.
4. Check the page and the playground at `https://ekalinin.github.io/gogv/`, and set the page as the homepage of the repository (`gh repo edit --homepage`).

Rollback: disable GitHub Pages, or revert the commit.

## Open Questions

- Does the push of a tag start the workflow despite the path filters of the same `push` trigger? GitHub's documentation says that path filters are not evaluated for tag pushes. It is checked on the first release after the merge. If it does not start, a run by hand deploys, and a follow-up moves the tag trigger to the `create` event, without changing the spec.
