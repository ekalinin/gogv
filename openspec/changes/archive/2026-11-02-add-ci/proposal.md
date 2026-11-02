# Proposal

## Why

The repository has no CI: every check runs only on the maintainer's machine, and `make determinism` needs an arm64 Mac with Rosetta. Before the repository is made public, pull requests from anyone need the same checks to run on their own. A dry run on a clean clone also showed a hole in the conformance harness: without the Liberation fonts in `.cache/fonts/liberation`, `make conformance` passes, but `resvg` draws no text on either side, so the SVG comparison is weaker than the contract says (one known difference of the curated corpus passes only because its text is not drawn), and the README does not name the step that fetches the fonts.

## What Changes

- A GitHub Actions workflow on every pull request and every push to `main`, with no Docker and no Graphviz:
  - lint: `go vet` and `gofmt`, `tools/provenance`, and `internal/textmetrics/tables.go` regenerated from the fonts and compared with the committed file
  - tests: `go test -race ./...` on linux/amd64 and macOS/arm64, which also runs the WebAssembly test under Node.js
  - conformance: the curated corpus and the command line cases against their committed references, with the pinned `resvg` and the Liberation fonts
  - determinism: the hashes of the curated corpus rendered on linux/amd64, macOS/arm64 and in the WebAssembly build must be equal
- The conformance run fails when the Liberation fonts are missing, naming the command that fetches them, as it already fails when `resvg` is missing.
- The README names that command among the conformance steps.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `upstream-compatibility`: the `svg` comparison of the layered similarity contract rasterizes both documents with the Liberation 2.1.5 fonts as the only fonts, and a run without them fails.

## Impact

- New `.github/workflows/ci.yml`.
- `internal/conformance/conformance_test.go`: the check for the fonts.
- `README.md`: the Conformance section.
- The workflow downloads `resvg` 0.48.1 and the Liberation 2.1.5 fonts from GitHub releases, both pinned by checksum. The harvested gopuml corpus and the recording of references stay local, as they need Docker and the reference image.
- While the repository is private, the jobs use the Actions minutes of the account, and macOS minutes cost more than Linux minutes; public repositories run standard runners for free.
