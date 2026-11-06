# Proposal

## Why

gogv has no page of its own: a visitor learns what it is only from the README, and cannot try it without installing Go. gopuml, its first consumer, has a landing page at `https://ekalinin.github.io/gopuml/` with a playground that renders diagrams in the browser. gogv runs in WebAssembly without a server, so a page with a playground shows what it does in the most direct way: a visitor types DOT and sees the SVG.

## What Changes

- A landing page of gogv at `https://ekalinin.github.io/gogv/`, static files in `site/`, with the structure and the look of gopuml's page:
  - what gogv is, with the example of the README as source and as SVG;
  - the three ways to run it: a Go library, a `dot` command line, a WebAssembly module;
  - its use with `go tool pprof`;
  - the output formats, with a link to the Status lists of the README;
  - the installation of the release the page names;
  - the conformance results of that release, from its notes;
  - the clean-room sources and the license.
- A playground on the page: an editor whose DOT source is rendered as SVG in the browser. It offers example graphs, shows the warnings, and shows the syntax errors with their line. It lets the visitor download the SVG, and loads the module only when it is reached.
- `internal/playground`: the WebAssembly entry point of the playground and its JavaScript loader, adapted from gopuml's `cmd/plantuml-wasm`. It is not a public API: the library and the command line do not change.
- `tools/site` and `make site`: assemble the site from `site/`, the example and a WebAssembly build of the playground:
  - for a release, from the sources of the tag the page names;
  - for a local preview, from the working tree.
- A `pages` workflow that deploys the site to GitHub Pages:
  - on a push to `main` that changes the page;
  - on the push of a release tag;
  - by hand.
- Tests keep the page consistent: the page must name the newest release that has notes in `docs/releases/`, and its conformance results must match those notes. A release therefore updates the page in the pull request of its notes, which the Releasing section of the README states.

Out of scope:
- the other output formats in the playground, and rendering in a Web Worker;
- release assets: the release process still publishes no binaries;
- third-party assets or services on the page: web fonts, CDNs, analytics.

## Capabilities

### New Capabilities

- `landing-page`: the published page, its content and its consistency with the release, the playground and the loading of its module, the deployment, the local preview, and the license files of the module.

### Modified Capabilities

None.

## Impact

- New files:
  - `site/`: `index.html`, `style.css`, `page.mjs`, `playground.mjs`;
  - `internal/playground`: `main.go` and `gogv.mjs`;
  - `tools/site`;
  - `.github/workflows/pages.yml`.
- Tests: the page and its release in the root package, the playground under Node.js, and `tools/site`.
- `Makefile`: the `site` target. `README.md`: a link to the page, and the step of the Releasing section that names the release on the page.
- The repository settings: GitHub Pages with "GitHub Actions" as the source, and the page as the homepage of the repository. The workflow cannot set these.
- The first deployment comes with the first release that contains `internal/playground`. Until then, the workflow deploys nothing, with a notice.
- No change to the library API, the command line or the outputs.
