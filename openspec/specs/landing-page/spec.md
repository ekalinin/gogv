# landing-page Specification

## Purpose

Publishes a landing page of gogv on GitHub Pages that presents the project and lets visitors lay out DOT graphs in the browser with a WebAssembly build of a release.

## Requirements

### Requirement: Published landing page

The project SHALL publish a landing page at `https://ekalinin.github.io/gogv/` as static files, without a server-side component and without third-party assets or services. The content of the page other than the playground SHALL be readable without JavaScript.

#### Scenario: Visit the page

- **WHEN** a visitor opens `https://ekalinin.github.io/gogv/`
- **THEN** the page shows the name of the project, what it is, and links to the repository and the releases

#### Scenario: JavaScript disabled

- **WHEN** a visitor opens the page with JavaScript disabled
- **THEN** every section except the playground is readable, and the playground says that it needs JavaScript

### Requirement: Project presentation

The page SHALL describe:

- what gogv is, with the example graph of the README, as its DOT source and as the SVG that gogv writes for it;
- the three ways to run it: a Go library, a `dot` command line and a WebAssembly module;
- its use as the `dot` of `go tool pprof`;
- the output formats that the library writes, with a link to the Status and "Not supported yet" lists of the README;
- the installation of the release it names, with `go install` and `go get`;
- the conformance results of that release;
- the sources of the clean-room implementation;
- the license, and that gogv is not affiliated with the Graphviz project.

It SHALL link to the repository, the releases and the Go package documentation.

#### Scenario: Output formats

- **WHEN** a visitor reads the output formats on the page
- **THEN** they are the formats that the library writes, in the same order

#### Scenario: Example graph

- **WHEN** a visitor reads the example at the top of the page
- **THEN** its source is the example graph of the README, and its picture is the SVG that gogv writes for that source

#### Scenario: Installation

- **WHEN** a visitor reads the installation section
- **THEN** it names the release that the page names, shows how to install the command line with `go install` and the library with `go get` at that release, and the line that `dot -V` then prints

### Requirement: Consistency with the release

The page SHALL name exactly one release, the newest release that has notes in `docs/releases/`. Its conformance results SHALL be those of the notes of that release: the number of cases, passed, known differences, not supported yet and failed. The tests of the repository SHALL fail when the page names another release, or when its conformance results differ from those of the notes.

#### Scenario: Page behind the notes

- **WHEN** `docs/releases/v0.4.0.md` exists and the page still names `v0.3.0`
- **THEN** the tests of the repository fail

#### Scenario: Conformance results out of date

- **WHEN** a number of the conformance results on the page differs from the notes of the release that the page names
- **THEN** the tests of the repository fail

### Requirement: Playground

The page SHALL provide a playground: an editor whose DOT source is rendered as SVG in the browser by a WebAssembly build of the release that the page names, without a server. The rendering SHALL follow the edits of the source. The playground SHALL:

- offer example graphs, all of which render without a warning;
- show the warnings of a rendering, as the command line words them;
- show the message of a syntax error, with its line;
- name the release that it runs;
- let the visitor download the rendered SVG.

#### Scenario: Simple graph

- **WHEN** a visitor types `digraph { a -> b }` in the editor
- **THEN** the playground shows the rendered graph as SVG, with `b` below `a` and an arrow from `a` to `b`

#### Scenario: Feature not supported yet

- **WHEN** the source is `digraph { a [shape=record] }`
- **THEN** the playground shows the graph, and the warning that `shape=record` of node `a` is not supported yet

#### Scenario: Syntax error

- **WHEN** the source has a syntax error on its third line
- **THEN** the playground shows the message of the error, which names line 3

#### Scenario: Example graph

- **WHEN** a visitor selects an example graph
- **THEN** its source replaces the source of the editor and is rendered, without a warning

#### Scenario: Download

- **WHEN** a visitor downloads the result after a successful rendering
- **THEN** the browser saves the rendered SVG document

### Requirement: Loading of the module

The page SHALL NOT download the WebAssembly module until the playground becomes visible or the visitor interacts with it. While the module loads, the playground SHALL show that it is loading. When the module cannot be loaded, the playground SHALL show an error, and the rest of the page SHALL stay usable.

#### Scenario: Reading the top of the page

- **WHEN** a visitor reads the top of the page and does not reach the playground
- **THEN** the browser has not requested the WebAssembly module

#### Scenario: Module not available

- **WHEN** the request for the WebAssembly module fails
- **THEN** the playground shows that it could not be loaded, and the other sections of the page are unaffected

### Requirement: Deployment

A workflow SHALL assemble the site from the page on `main` and a WebAssembly build of the sources of the release tag that the page names, and deploy it to GitHub Pages. It SHALL run on a push to `main` that changes the page, the files it is assembled from or the workflow, on the push of a release tag, and by hand. When the tag that the page names is not in the repository yet, or when its sources have no playground module, it SHALL deploy nothing and finish without failing, with a notice.

#### Scenario: Change of the page

- **WHEN** a commit that changes the page is pushed to `main`, and the tag that the page names is in the repository
- **THEN** the site is deployed with that commit's page and a WebAssembly build of that tag

#### Scenario: Release tag pushed after the notes

- **WHEN** the pull request of the notes of `v0.4.0` makes the page name `v0.4.0`, and is merged before the tag `v0.4.0` is pushed
- **THEN** the workflow of the merge deploys nothing and succeeds with a notice, and the site is deployed when the tag `v0.4.0` is pushed

#### Scenario: Release without the playground

- **WHEN** the page names a release whose sources have no playground module
- **THEN** the workflow deploys nothing and succeeds with a notice

### Requirement: Local preview

A build target SHALL assemble the same site locally from the page and a WebAssembly build of the working tree, without publishing anything, so that it can be served by any static file server.

#### Scenario: Preview

- **WHEN** the target runs and its output directory is served by a static file server
- **THEN** the page and its playground work as on GitHub Pages, with the build of the working tree

### Requirement: License files of the module

The site SHALL serve the `LICENSE` and `NOTICE` of the repository next to the WebAssembly module, as `wasm/LICENSE` and `wasm/NOTICE`, both in the deployed site and in the local preview. A push to `main` that changes them SHALL deploy the site again.

#### Scenario: License files on the site

- **WHEN** the site is assembled, for a release or for a local preview
- **THEN** its `wasm` directory contains `LICENSE` and `NOTICE`, identical to those of the repository

#### Scenario: Change of the notices

- **WHEN** a commit that changes `NOTICE` is pushed to `main`
- **THEN** the site is deployed again, with that `NOTICE` in its `wasm` directory
