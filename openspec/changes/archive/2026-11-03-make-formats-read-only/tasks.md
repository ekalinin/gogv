# Tasks

## 1. API

- [x] 1.1 In `gogv.go`, rename the list to the unexported `formats`, keeping its order, add `func Formats() []string` that returns `slices.Clone(formats)` with a doc comment saying that each call returns a new slice, make `Render` and `FormatError.Error()` read `formats`, and remove `laterFormats` and `FormatError.NotSupportedYet` (design D1, D2). Verified by 1.2.
- [x] 1.2 Update the callers: `internal/cli` prints every `FormatError` as is; `internal/corpushash` and `TestRenderFormats` call `gogv.Formats()`, and the test no longer reads the removed field; the Library section of `README.md` names `gogv.Formats()`. Verify that `make lint test provenance` passes and that `git grep -n 'NotSupportedYet' -- '*.go'` finds only the `Warning` method, the tests of warnings (`gogv_test.go`, `internal/attrs`) and `internal/conformance`.
- [x] 1.3 Add `TestFormatsScenarios` to `gogv_test.go` for the "List of formats" and "Changed list" scenarios of the delta spec: the 14 formats each render `digraph { a }` without an error; after appending `png` and replacing `svg` with `pdf` in a returned list, `png` gives a `*FormatError`, `svg` renders, the error lists `svg` and neither `pdf` nor `png`, and the next `Formats()` equals the first. Verify that it passes with `go test -race -run TestFormatsScenarios .`.

## 2. Verification

- [x] 2.1 Verify that the command line is unchanged: `go test ./internal/cli` passes with its pinned unknown-format message, and `make conformance` passes, including the `cli/unknown-format` case.
- [x] 2.2 Verify that `make determinism` passes and prints `0915e7dbdad0afba7d0d868f4ffd30de6085772be21fcbbce57d67fdabcb23bc all`.
- [x] 2.3 Verify that `openspec validate make-formats-read-only --strict` passes; with the maintainer's go-ahead, push the branch, open the pull request with the sections of the pull request template, and verify that CI passes.
