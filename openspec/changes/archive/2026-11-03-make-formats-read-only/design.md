# Design

## Context

See proposal.md - Why. Today `gogv.go` has `var Formats = []string{...}` (14 formats) and `var laterFormats []string` (empty). `Render` rejects a format with `slices.Contains(Formats, format)` and fills `FormatError.NotSupportedYet` from `laterFormats`; `FormatError.Error()` lists `Formats`. Its readers in the repository are `gogv_test.go`, `internal/corpushash` (which hashes every format in the order of the list) and, through the error, `internal/cli`, which checks the format with `Render(ctx, "", format)` and prints `FormatError` as is unless `NotSupportedYet` is set. `Render`'s `switch` on the format has no default branch, so a format that passes the check but has no case writes nothing. `internal/conformance` keeps its own list and is not affected. gopuml does not import gogv, and gogv has no tag.

## Goals / Non-Goals

**Goals:**

- No caller can change which formats `Render` accepts or how it reports one it rejects
- The exported API shrinks to what is used

**Non-Goals:**

- Typed format names or constants
- The version of the first tag, a separate decision
- Renaming the cli scenario "Format of a later slice", whose text already describes a supported `xdot`

## Decisions

### D1. A function that returns a copy

`func Formats() []string` returns `slices.Clone(formats)`, where `formats` is the unexported list in today's order. `Render` and `FormatError.Error()` read `formats`. The order stays, so the message of an unknown format and the hashes of `internal/corpushash` stay byte for byte.

Alternatives:

- Keep the variable and give `Render` a private copy: two sources of the list, and `Formats` still lies to whoever reads it after a caller changed it.
- `func Formats() iter.Seq[string]`: callers need `slices.Collect` for `len` and `slices.Contains`, the two things they do with it.
- `type Format string` with 14 constants: 15 more names to keep stable, while the command line passes a string from a flag anyway.

### D2. Remove FormatError.NotSupportedYet

The field is always false since `laterFormats` is empty, and the cli spec reports every format outside v1 as not recognized. `laterFormats` and the field go; `FormatError.Error()` keeps only the reference's wording; `internal/cli` prints any `FormatError` as is and wraps other errors as before. Adding a field to `FormatError` later, for formats of a later version, does not break callers; removing it after a tag would.

## Risks / Trade-offs

- [A caller of `gogv.Formats` or `FormatError.NotSupportedYet` breaks] → none exists outside this repository; the change lands before the first tag
- [The order of the list changes by mistake] → the test of the unknown-format message pins the order, and the `make determinism` hash of all outputs must stay `0915e7dbdad0afba7d0d868f4ffd30de6085772be21fcbbce57d67fdabcb23bc`
