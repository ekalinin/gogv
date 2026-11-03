# Proposal

## Why

`gogv.Formats` is an exported `[]string` variable, and `Render` decides which formats it accepts by looking it up. Any caller can change it, for the whole process: after `gogv.Formats = append(gogv.Formats, "png")`, `Render` returns 0 bytes and no error for `png`, and after `gogv.Formats[3] = "pdf"` it rejects `svg` and lists `pdf` among the formats it writes (both checked on `main`). A write while other goroutines render is also a data race, against the promise that `Render` is safe for concurrent use. Next to it, `FormatError.NotSupportedYet` is always false: `laterFormats` has been empty since the last slice of v1, and the cli spec reports every format outside v1 as not recognized. No module imports gogv yet and no version is tagged, so the API can still change without breaking anyone; once gopuml depends on a tagged version, it cannot.

## What Changes

- **BREAKING** `gogv.Formats` becomes a function, `func Formats() []string`, that returns a new slice with the 14 formats of v1 in the current order on every call. `Render` and `FormatError.Error()` use an unexported list, so changing a returned slice changes nothing else.
- **BREAKING** `FormatError.NotSupportedYet` and the unexported `laterFormats` are removed. A field can be added back later without breaking callers; removing one after a tag cannot.
- The command line, its messages and its exit statuses stay the same; `internal/cli` no longer tests the removed field.
- `internal/corpushash`, the tests and the README use `gogv.Formats()`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `library-api`: a new requirement for the list of output formats, which a caller can read but cannot change.

## Impact

- `gogv.go`: `Formats`, `FormatError`, `Render`.
- `internal/cli/cli.go`, `internal/corpushash/main.go`, `gogv_test.go`, `README.md`.
- Callers of `gogv.Formats` write `gogv.Formats()`; callers of `FormatError.NotSupportedYet` drop the test. No module outside this repository uses either today.
- Outputs are unchanged: the order of the list, and with it the hashes of `make determinism`, stay the same.
