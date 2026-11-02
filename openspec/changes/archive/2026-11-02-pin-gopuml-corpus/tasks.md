# Tasks

## 1. Harvest from a pinned version

- [x] 1.1 In `tools/harvest`, drop the default of `-gopuml` and add `-gopuml-version` without a default (design D4). With `-gopuml`, use the checkout; otherwise run `go mod download -json github.com/ekalinin/gopuml@<version>` in an empty temporary directory and use its `Dir` (D1); fail with a usage message when neither flag is given. Keep building `./cmd/plantuml` in the resolved directory into `.cache/tools/gopuml-plantuml` (D2), print the source and the directory before the build, and update the package comment. Verify that `go run ./tools/harvest` without flags prints the usage message and exits with status 2 (reported by `go run` as `exit status 2`), that a run with `-gopuml-version v0.3.0` prints a directory under `go env GOMODCACHE` and builds the command line, and that `git diff --exit-code go.mod` passes and no `go.sum` appears (the module has no dependencies).
- [x] 1.2 For a pinned version, fetch the Vega cases with `bash <dir>/tools/corpus/fetch-vega.sh` and `VEGA_CACHE` set to the absolute path of gogv's `.cache/vega`, collect them from there, and fail the harvest when the script fails (D5); leave the local checkout's `<checkout>/.cache/vega` as it is. Verify that after a run `.cache/vega/.complete` reads `425 v1.2026.9`, and that the counts printed for the `vega-*` groups are those of the current `.cache/harvest.log`.
- [x] 1.3 Write `SOURCE` as `gopuml <version>` for a pinned version, and as `gopuml <commit> local`, with ` dirty` when `git status --porcelain` prints anything and `unknown` when git fails, for a checkout (D6). Verify a new `tools/harvest/main_test.go` (with its `SPDX-License-Identifier` and `Provenance: original` lines) for the four forms of the line, and `make provenance`.
- [x] 1.4 In `Makefile`, set `GOPUML ?=` empty, add `GOPUML_VERSION ?= v0.3.0`, pass `-gopuml-version $(GOPUML_VERSION)` always and `-gopuml $(GOPUML)` only when `GOPUML` is set, and update the comment of the `harvest` target (D4). Verify that `make -n harvest` shows `-gopuml-version v0.3.0` and no `-gopuml ` flag, and that `make -n harvest GOPUML=../plantuml-in-go` shows both.
- [x] 1.5 Make `tools/clustervariant` copy `SOURCE` from its input to its output directory when the file exists (D6). Verify a test in `tools/clustervariant/main_test.go` that the output `SOURCE` equals the input's, and that the run still succeeds without one.

## 2. Corpora in the report

- [x] 2.1 Add the corpora of the run to `Report` (name, number of cases, source) and write them in `Markdown` as one `- Corpora:` line after `- Cases:`, a corpus without a source written without parentheses (D7). Verify unit tests in `internal/conformance` for the "Source of the harvested corpus" and "Harvested corpus not collected" scenarios.
- [x] 2.2 In `TestConformance`, add each corpus that runs, with the content of `<cases>/SOURCE` when it exists, and the command line cases as `cli`; skip the corpora that are not collected. Verified by 4.4.

## 3. Documentation

- [x] 3.1 Update the Conformance section of `README.md`: `make harvest` fetches gopuml at `GOPUML_VERSION` of the `Makefile` through the Go module proxy and the Vega cases through gopuml's script, needs network access, Docker and the reference image, and `make harvest GOPUML=<dir>` harvests from a local checkout instead; the report names the source of each corpus. Verify that every command in the section runs as written in 4.2 to 4.4.

## 4. Verification

- [x] 4.1 Copy the current `.cache/gopuml-dot` and `.cache/gopuml-cluster-dot` (harvested from `becdd08`) outside `.cache`. Verify that both copies exist and that the first has 260 `.gv` files.
- [x] 4.2 Run `make harvest GOPUML=../plantuml-in-go`. Verify that `.cache/gopuml-dot/SOURCE` and `.cache/gopuml-cluster-dot/SOURCE` read `gopuml <HEAD of the checkout> local`, with ` dirty` when the checkout has changes or untracked files (the "Local checkout with changes" scenario).
- [x] 4.3 Run `make harvest` with `GOPUML` unset. Verify that `diff -r -x SOURCE` against both copies of 4.1 reports no difference, that both `SOURCE` files read `gopuml v0.3.0`, and that the gopuml directory printed by the harvest is in the module cache (the "Pinned gopuml version" and "Same corpus from the release" scenarios).
- [x] 4.4 Run `make conformance`. Verify that it passes and that `git diff conformance-report.md` shows only the new line `- Corpora: curated 255, gopuml 260 (gopuml v0.3.0), gopuml-clusters 36 (gopuml v0.3.0), cli 14`.
- [x] 4.5 Verify that `make lint test provenance` passes and that `git diff --exit-code go.mod` passes and no `go.sum` appears (the module has no dependencies).
