# Proposal

## Why

gopuml is switching its graph layout to gogv (its change `switch-to-gogv`). Its `--check-graphviz` must print the line that gogv's `dot -V` prints, `dot - graphviz version 16.1.0 (gogv <version>)`, and its design D4 takes the Graphviz version from gogv rather than from a gopuml constant that would go stale. gogv v0.1.0 does not export that version: it is a constant of `internal/cli`, and the SVG writer writes the same number again in its own string, so the two can drift apart. gopuml needs a tagged gogv that exports it before it can require gogv in `go.mod`.

## What Changes

- The package exports the Graphviz version it follows, `gogv.GraphvizVersion` (`"16.1.0"`).
- The version is written in one place, `internal/output`, used by the comment of the SVG output and re-exported by the package; `dot -V` prints the exported constant. Outputs, messages and determinism hashes stay byte for byte the same.
- The release notes of v0.2.0, `docs/releases/v0.2.0.md`, and the release with `make release VERSION=v0.2.0` once this change is merged and CI has passed.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `library-api`: a new requirement for the Graphviz version that the package follows.

## Impact

- `internal/output/svg.go`, `gogv.go`, `internal/cli/cli.go`, tests, `docs/releases/v0.2.0.md`.
- Additive: no exported name changes or goes. The version after v0.1.0 is v0.2.0, a minor version for new API.
- gopuml requires `github.com/ekalinin/gogv v0.2.0` in its own change; the fixes of its design D9 for the 12 differing cases come in later gogv tags.
