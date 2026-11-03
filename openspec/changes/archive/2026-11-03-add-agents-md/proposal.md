# Proposal

## Why

The foundation design meets the risk of clean-room with an AI assistant by giving every implementation session the list of allowed sources. Nothing in the repository does that today: it rests on the maintainer's prompts, and a contributor's agent gets nothing. The agent of the sibling project gopuml reads an `AGENTS.md` whose first rule is the opposite of gogv's, "Read the Java before writing Go". Agents also miss conventions that only show up as a failing test or a review comment: products written so that the compiler cannot fuse them, elementary functions from `internal/fdlibm`, a conformance report rewritten by a run without the harvested corpus, the maintainer's rules of style.

## What Changes

- `AGENTS.md`: the clean-room rule in the file, stated as the sources to work from, with the source code of Graphviz and of its ports as the hard limit; every other rule is a pointer to a section of `CONTRIBUTING.md`, worded with the situation in which to read it.
- `CLAUDE.md`: `@AGENTS.md`, so that Claude Code reads the same file as the other agents, as gopuml does.
- `CONTRIBUTING.md`, for people and agents alike:
  - Allowed sources: how to run the reference, `docker run --rm -i --platform linux/amd64 gogv-reference dot -T<format> < in.gv`, rather than a system Graphviz of another version and with Pango
  - a new Determinism section: `float64(a*b) + c` against fused multiply-adds, elementary functions from `internal/fdlibm` (`math.Sqrt` is exact and allowed), text widths only from the generated tables of `internal/textmetrics`
  - Checks: commit `conformance-report.md` only from a run with the harvested corpora; a known difference names its reason and the change that resolves it; spec scenarios get `Test...Scenarios` tests; a newly supported attribute changes `internal/attrs/supported.go` and the lists of the README
  - a new Style section: headers in sentence case, a spaced hyphen ` - ` where a dash is needed, minimal diffs that keep the existing names

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None. The change adds and extends documentation; no behavior of the library, the command line or the conformance harness changes, so the change declares `skip_specs`.

## Impact

- New `AGENTS.md` and `CLAUDE.md`.
- `CONTRIBUTING.md`: two new sections and additions to two others.
- No code, test or workflow changes.
