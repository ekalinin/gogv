# Proposal

## Why

`make harvest` collects the DOT that gopuml generates from a gopuml checkout at a sibling path that only exists on the maintainer's machine. In a public repository nobody else has that directory, and even with a checkout of their own they would harvest whatever branch and uncommitted changes it holds: the current corpus comes from commit `becdd08`, which was only on a feature branch of gopuml. The committed `conformance-report.md` cannot be reproduced from the repository alone. gopuml v0.3.0 is now released and contains that commit, with the same cases and the same Vega tag, so the corpus can be pinned to it.

## What Changes

- `make harvest` takes gopuml by reference: the module `github.com/ekalinin/gopuml` at a version pinned in the `Makefile` (`GOPUML_VERSION`, v0.3.0), fetched through the Go module proxy and verified by the checksum database. No gopuml checkout is needed.
- The Vega cases are fetched by gopuml's own `tools/corpus/fetch-vega.sh` of that version, into gogv's cache, so the Vega tag is pinned through the gopuml version.
- A local gopuml checkout stays available by explicit choice (`make harvest GOPUML=<dir>`), for work on both projects at once. It is no longer the default.
- The harvested corpus records its source: the pinned version, or the commit of the local checkout marked as local, and dirty when the checkout has uncommitted changes.
- `conformance-report.md` names the corpora of the run and the source of each harvested corpus.
- gopuml is not added to `go.mod`: gopuml depends on gogv, and the requirement would reach the module graph of the library's users.
- The README describes both ways of harvesting.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `upstream-compatibility`: the harvested gopuml corpus comes from a pinned gopuml version unless a local checkout is named, and records its source; the conformance report names the source of each corpus.

## Impact

- `Makefile`: `GOPUML` empty by default, new `GOPUML_VERSION ?= v0.3.0`, the `harvest` target passes both.
- `tools/harvest`: resolves gopuml's directory from the module cache or from `-gopuml`, fetches the Vega cases for a pinned version, writes the new `SOURCE`.
- `tools/clustervariant`: carries `SOURCE` over to the variant corpus.
- `internal/conformance`: the report lists the corpora and their sources.
- `README.md`: the Conformance section.
- Harvesting needs network access to the Go module proxy and GitHub, Docker and the reference image, as before. The curated corpus and `go test ./...` are unaffected.
