# Design

## Context

See proposal.md - Why. Checked on a clean clone of `main` (`9d3f8f4`) before writing this design:

- `go vet ./...`, `go test ./...` and `go test -race ./...` (about 22 s) pass without `.cache`. `TestWebAssembly` builds `internal/wasmcheck` for `GOOS=js GOARCH=wasm` and runs it under `go_js_wasm_exec`; it skips itself when `node` is not found.
- After `tools/reference/fetch-tools.sh` (pinned `resvg` 0.48.1 for `Linux-x86_64`, `Darwin-arm64` and `Darwin-x86_64`), `make conformance` passes in about 3 s on the curated corpus and the command line cases; the harvested corpora are optional and skipped.
- Without `.cache/fonts/liberation`, `Resvg` still passes `--skip-system-fonts --use-fonts-dir <missing dir>`, so no text is drawn: the run passes with `svg` 209 passed and 10 known differences, against 208 and 11 with the fonts.
- `go generate ./internal/textmetrics` runs `tools/fontmetrics`, which downloads the Liberation 2.1.5 archive by checksum, unpacks the twelve fonts into `.cache/fonts/liberation` and writes `tables.go`; the result is identical to the committed `tables.go`.
- `make determinism` passes: 3571 hashes of the curated corpus, the same natively, for amd64 and in the WebAssembly build. It runs `internal/corpushash` on `testdata/conformance/cases`, and the WebAssembly build under `$(go env GOROOT)/lib/wasm/go_js_wasm_exec` with a clean environment.
- The repository has no `.github` directory; Actions are enabled and allow all actions.

## Goals / Non-Goals

**Goals:**

- Every pull request and push to `main` runs the checks that need no Docker
- The determinism of the output is checked across operating systems and architectures, not only under Rosetta
- A conformance run can no longer pass without the fonts it rasterizes with

**Non-Goals:**

- The harvested gopuml corpus, `make reference`, `make harvest` and the reference image: they need Docker and a Graphviz build
- Releases, a README badge, Dependabot, and the branch protection rules that would make the checks required (a repository setting)
- Linux on arm64 (`ubuntu-24.04-arm`), which is free for public repositories only

## Decisions

### D1. One workflow, four jobs

`.github/workflows/ci.yml` runs on `pull_request` and on `push` to `main`, with `permissions: contents: read` and a `concurrency` group per ref that cancels a superseded run.

```
lint (ubuntu)         make lint, make provenance,
                      go generate ./internal/textmetrics + git diff --exit-code
test (ubuntu, macos)  node --version, go test -race ./...,
                      corpus hashes: native, and wasm on ubuntu  -> artifact
conformance (ubuntu)  fetch-tools.sh, go generate (fonts), make conformance
                      -> conformance-report.md as an artifact
determinism           needs test: every hash file equals the linux/amd64 one
```

Alternative: one job running everything, which would give up the macOS run and the parallelism.

### D2. Go from go.mod

`actions/setup-go` takes the version from `go.mod` (`go-version-file`), so CI builds with the oldest Go the module supports (1.26), while the maintainer builds with a newer one.

Alternative: `stable`, which would not notice code that needs a newer Go than `go.mod` says.

### D3. Determinism from the test jobs

The `test` jobs already run on linux/amd64 and macOS/arm64, so they also build and run `internal/corpushash` as `make determinism` does; the Linux job adds the WebAssembly build under `go_js_wasm_exec`. Each uploads its hash files, and the `determinism` job compares them with `cmp` against the linux/amd64 file. Together they cover the platforms of `make determinism` (arm64, amd64, wasm) and add Linux. `make determinism` stays for local use.

Alternative: a separate matrix for the hashes, which would add a macOS job per run.

### D4. Node.js is required, not optional

`TestWebAssembly` skips itself when `node` is missing. The `test` jobs run `node --version` first, so that a runner image without Node.js fails instead of silently skipping the WebAssembly check. Both runner images ship Node.js.

### D5. Fonts for conformance through go generate

The `conformance` job fetches the fonts with `go generate ./internal/textmetrics`, the documented way, rather than a second download script. The `lint` job runs the same command and fails on `git diff --exit-code`, so that `tables.go` stays what `tools/fontmetrics` generates.

### D6. The harness requires the fonts

`TestConformance` already fails with `resvg not fetched: run tools/reference/fetch-tools.sh`. Right after that check it fails when `.cache/fonts/liberation` holds no `.ttf` file, with `Liberation fonts not fetched: run go generate ./internal/textmetrics`, before any case is compared or the report written. The README lists the command next to `tools/reference/fetch-tools.sh`.

Alternative: falling back to system fonts, which would make the outcome depend on the host.

### D7. Actions pinned by commit

The actions are pinned by full commit SHA with the version in a comment, as a public repository runs the workflow for pull requests from anyone. The SHAs are resolved when the workflow is written.

## Risks / Trade-offs

- [A download from GitHub releases fails (`resvg`, the fonts)] → the job fails and can be rerun; the checksums keep a changed file from passing
- [linux/amd64 and macOS/arm64 hash differently] → a real breach of the determinism the specs require: the `determinism` job names the first differing line, to be fixed before merging
- [The report written in CI covers only the curated corpus] → it is not compared with the committed report, only uploaded for inspection
- [macOS minutes while the repository is private] → one macOS job per run
- [Pinned SHAs age] → updated by hand; Dependabot is a Non-Goal

## Migration Plan

The pull request that adds the workflow runs it. Once the four jobs pass there, making them required is a repository setting outside this change.
