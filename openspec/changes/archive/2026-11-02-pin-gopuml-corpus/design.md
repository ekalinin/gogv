# Design

## Context

See proposal.md - Why. The current flow of `make harvest`:

```
tools/harvest -gopuml ../plantuml-in-go
  go build ./cmd/plantuml (in the checkout)  -> .cache/tools/gopuml-plantuml
  .puml of <checkout>/testdata/conformance/cases and <checkout>/.cache/vega
  gopuml-plantuml -tsvg -pipe, GRAPHVIZ_DOT = wrapper that saves each DOT
  -> .cache/gopuml-dot/<capability>/<case>-<n>.gv, SOURCE = "gopuml <HEAD>"
tools/refgen         -> .cache/reference/gopuml-dot
tools/clustervariant -> .cache/gopuml-cluster-dot (SOURCE is not copied)
tools/refgen         -> .cache/reference/gopuml-cluster-dot
```

The Vega cases are not part of gopuml's repository: gopuml's `tools/corpus/fetch-vega.sh` fetches them from `github.com/plantuml/plantuml` at a pinned tag (`v1.2026.9`) into `$VEGA_CACHE`, by default its own `.cache/vega`. `LoadCases` reads only `.gv` files, so a `SOURCE` file in a corpus directory is not a case.

Checked on gopuml v0.3.0 before writing this design:

- `go mod download -json github.com/ekalinin/gopuml@v0.3.0`, run in an empty directory, returns the module directory in the module cache (origin `6a2bb64`, `refs/tags/v0.3.0`), with `testdata/conformance/cases` (477 `.puml`) and `tools/corpus/fetch-vega.sh`. Files in the module cache are read-only and not executable.
- `go build ./cmd/plantuml` run in that directory builds a working command line; gopuml's `go.mod` has no `replace`.
- `becdd08`, the source of the current corpus, is an ancestor of v0.3.0; `testdata/conformance/cases` and `tools/corpus` are the same in both, and the only Go files that differ are `landing_test.go` and `tools/bench/main.go`, which do not generate DOT.

## Goals / Non-Goals

**Goals:**

- Harvest without any gopuml checkout, from one pinned reference
- Keep a local checkout as an explicit override, with the rest of the flow unchanged
- Reproduce the current corpus byte for byte from v0.3.0

**Non-Goals:**

- Committing the harvested corpus: the Vega cases derive from upstream PlantUML inputs that gopuml does not commit either
- Running the harvest in CI
- Any change to the curated corpus, the comparisons or the known differences
- The names `plantuml-in-go` and `graphviz-dot-in-go` in the archived changes: a separate `chore/` branch

## Decisions

### D1. gopuml comes from the Go module proxy

`tools/harvest` runs `go mod download -json github.com/ekalinin/gopuml@<version>` with an empty temporary directory as its working directory and takes `Dir` from the output. The empty directory keeps gogv's `go.mod` untouched and keeps a `go.sum` from appearing (gogv has no dependencies and no `go.sum`). The module proxy and the checksum database verify what is downloaded, and Go is already required.

Alternatives: a `git clone --branch <version>` into `.cache/gopuml`, as `fetch-vega.sh` does for PlantUML, which keeps file modes but needs its own cache invalidation and has no checksum check; committing the harvested corpus, rejected under Non-Goals.

### D2. One build path for both sources

Whatever the source, `tools/harvest` resolves one gopuml directory (the module directory or the `-gopuml` checkout) and keeps building `./cmd/plantuml` in it into `.cache/tools/gopuml-plantuml`, as today. Only the resolution of the directory is new.

Alternative: `GOBIN=.cache/tools go install github.com/ekalinin/gopuml/cmd/plantuml@<version>`, which would need a second build path and a different binary name for the local checkout.

### D3. gopuml stays out of go.mod

Neither a `require` nor a `tool` directive: gopuml depends on gogv, so the two modules would require each other, and gopuml would enter the module graph of every user of the library.

### D4. The pin lives in the Makefile

`Makefile` gets `GOPUML_VERSION ?= v0.3.0`, and `GOPUML ?=` becomes empty. The `harvest` target always passes `-gopuml-version $(GOPUML_VERSION)` and passes `-gopuml $(GOPUML)` only when `GOPUML` is set. `tools/harvest` keeps its `-gopuml` flag, now with no default, and gets `-gopuml-version` with no default: `-gopuml` wins when both are given, and the command fails with a usage message when neither is. The version is written in one place, which `make harvest` always goes through.

Alternative: the version as the default of the flag in `tools/harvest`, which hides the pin from the `Makefile` or writes it in two places.

### D5. Vega cases through gopuml's script, into gogv's cache

For a pinned version, `tools/harvest` runs `bash <dir>/tools/corpus/fetch-vega.sh` with `VEGA_CACHE` set to the absolute path of gogv's `.cache/vega`, and collects the Vega cases from there. `bash` is needed because the script is not executable in the module cache, and `VEGA_CACHE` because its default destination is inside the read-only module directory. The script records its tag in `.complete` and does nothing on later runs. A failure of the script fails the harvest, since the corpus of a pinned version must be complete. The Vega tag is thereby pinned by the gopuml version.

For a local checkout nothing changes: the Vega cases are taken from `<checkout>/.cache/vega` when it exists.

### D6. SOURCE

`SOURCE` stays one line in `.cache/gopuml-dot`:

- pinned version: `gopuml v0.3.0`
- local checkout: `gopuml <git rev-parse HEAD> local`, followed by ` dirty` when `git status --porcelain` prints anything, and with `unknown` in place of the commit when git fails. Untracked files count, because an untracked `.puml` case changes the corpus.

`tools/clustervariant` copies `SOURCE` from its input to its output directory, so that the variant records the same source.

### D7. Corpora in the report

`Report` gets the list of the corpora of the run: name, number of cases and source. `TestConformance` adds each corpus it runs, with the content of `<cases>/SOURCE` when the file exists, and the command line cases. `Markdown` writes the list after the number of cases:

```
- Cases: 565
- Corpora: curated 255, gopuml 260 (gopuml v0.3.0), gopuml-clusters 36 (gopuml v0.3.0), cli 14
```

A corpus that was not collected is not listed.

## Risks / Trade-offs

- [The module proxy or GitHub is unreachable] → the harvest fails with the error of `go mod download` or of `fetch-vega.sh`; the curated corpus and `go test ./...` do not need the network; `GOPROXY=direct` or `GOPUML=<dir>` are fallbacks
- [A later gopuml version moves its cases or `fetch-vega.sh`] → bumping `GOPUML_VERSION` is a deliberate edit, done together with a harvest run, the conformance run, and the update of `conformance-report.md` and the known differences
- [`dirty` is shown for files that do not affect the corpus, such as notes] → the marker is informational and errs on the side of reporting
- [The module cache is shared with other projects] → the module directory is only read; the Vega cases and the built command line go to gogv's `.cache`

## Migration Plan

1. Keep a copy of the current `.cache/gopuml-dot` and `.cache/gopuml-cluster-dot`.
2. Run `make harvest`, which now harvests from v0.3.0, and compare the `.gv` files with the copy byte for byte (Goals: the same corpus).
3. Run `make conformance`; `conformance-report.md` differs from the committed one only by the line of the corpora.

Rollback: `make harvest GOPUML=../plantuml-in-go` harvests from the local checkout as before.
