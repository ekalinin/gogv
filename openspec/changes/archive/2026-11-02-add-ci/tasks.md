# Tasks

## 1. Fonts in the conformance harness

- [x] 1.1 In `TestConformance`, right after the check for `resvg`, fail with `Liberation fonts not fetched: run go generate ./internal/textmetrics` when `.cache/fonts/liberation` holds no `.ttf` file (design D6). Verify on a clean clone with `resvg` fetched and no fonts that `make conformance` fails with that message and leaves `conformance-report.md` unchanged (`git diff --exit-code conformance-report.md`), and that after `go generate ./internal/textmetrics` it passes (the "Fonts not fetched" scenario).
- [x] 1.2 In the Conformance section of `README.md`, add `go generate ./internal/textmetrics` with the comment "fetch the Liberation fonts" after `tools/reference/fetch-tools.sh`. Verify that `tools/reference/fetch-tools.sh`, `go generate ./internal/textmetrics` and `make conformance` pass in that order on a clean clone.

## 2. Workflow

- [x] 2.1 Resolve the commit SHAs of the latest releases of `actions/checkout`, `actions/setup-go`, `actions/upload-artifact` and `actions/download-artifact` with `gh api`, and write `.github/workflows/ci.yml` with the `pull_request` and `push` to `main` triggers, `permissions: contents: read`, a `concurrency` group per ref that cancels a superseded run, and `actions/setup-go` with `go-version-file: go.mod` in every job (D1, D2, D7). Verify that every `uses:` names a 40-character SHA followed by its version in a comment.
- [x] 2.2 Add the `lint` job on `ubuntu-latest`: `make lint`, `make provenance`, `go generate ./internal/textmetrics` and `git diff --exit-code` (D5). Verified by 3.2.
- [x] 2.3 Add the `test` job on `ubuntu-latest` and `macos-latest`: `node --version`, `go test -race ./...`, then the hashes of `testdata/conformance/cases` from `internal/corpushash` built natively, and on Linux also for `GOOS=js GOARCH=wasm` run under `go_js_wasm_exec` with a clean environment as in `make determinism`; upload the hash files as an artifact per runner (D3, D4). Verified by 3.2.
- [x] 2.4 Add the `conformance` job on `ubuntu-latest`: `tools/reference/fetch-tools.sh`, `go generate ./internal/textmetrics`, `make conformance`, and upload `conformance-report.md` as an artifact, also when the run fails (D5). Verified by 3.2.
- [x] 2.5 Add the `determinism` job, needing `test`: download the hash artifacts and `cmp` every file against the linux/amd64 native one, printing the last line (the hash of all outputs) of each (D3). Verified by 3.2.

## 3. Verification

- [x] 3.1 Verify that `go run github.com/rhysd/actionlint/cmd/actionlint@latest .github/workflows/ci.yml` reports no problem, and that `make lint test provenance` passes.
- [x] 3.2 With the maintainer's go-ahead, push the branch and open the pull request. Verify that the `lint`, `test` (both runners), `conformance` and `determinism` jobs pass, that `node --version` prints a version on both runners, and that the hash of all outputs printed by `determinism` is the same for every file and equals the local `make determinism` result at the same commit.
- [x] 3.3 Verify that `openspec validate add-ci --strict` passes.
