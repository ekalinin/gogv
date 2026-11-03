# Tasks

## 1. CONTRIBUTING

- [x] 1.1 In Allowed sources, add how to observe the reference, `docker run --rm -i --platform linux/amd64 gogv-reference dot -T<format> < in.gv`, and why it is preferred to a system Graphviz (another version, and Pango) (design D4). Verify that the command, with `-Tplain` and `digraph { a -> b }` on standard input, prints the reference's `plain` output.
- [x] 1.2 Add a Determinism section after Provenance lines (D4): `float64(a*b) + c` and `TestNoFusedMultiplyAdd`, elementary functions from `internal/fdlibm` with `math.Sqrt` allowed, text widths only from the tables of `internal/textmetrics`, and the reason (byte-identical output on every platform, checked by CI). Verify on a clean clone that rewriting one `float64(x*x) + float64(y*y)` of `internal/layout/geom.go` as `x*x + y*y` makes `go test -run TestNoFusedMultiplyAdd ./internal/layout` fail, and that it passes without the rewrite.
- [x] 1.3 Add to Checks (D4): `conformance-report.md` committed only from a run with the harvested corpora; a known difference with its `reason` and the `change` that resolves it; every spec scenario covered by a test, as the `Test...Scenarios` tests cover the scenarios of one requirement each; a newly supported attribute or value updates `internal/attrs/supported.go` and the Status and "Not supported yet" lists of the README. Verify that every file, field and list it names exists, and that `make conformance` on a clean clone without the harvested corpora writes `- Corpora: curated 255, cli 14`, the rewrite the rule guards against.
- [x] 1.4 Add a Style section before Commits and branches (D5): headers in sentence case, a spaced hyphen ` - ` where a dash is needed, minimal diffs that keep the existing names, and that a renamed heading of `CONTRIBUTING.md` updates the pointers of `AGENTS.md` in the same pull request. Verified by 3.1.

## 2. Agent files

- [x] 2.1 Write `AGENTS.md` (D2, D3): a one-line description of gogv, the clean-room rule inline, and one pointer per row of the table of D3, each naming its situation and its anchor. Verify that every anchor resolves to a heading of `CONTRIBUTING.md` (GitHub's slug of the heading), and that the file has at most 25 lines.
- [x] 2.2 Write `CLAUDE.md` with the single line `@AGENTS.md` (D1). Verify that it is the same as gopuml's `CLAUDE.md`.

## 3. Verification

- [x] 3.1 Verify that no em dash appears in the new or changed files, that every heading of them is in sentence case, that `make lint test provenance` passes, and that `openspec validate add-agents-md --strict` passes.
- [x] 3.2 With the maintainer's go-ahead, push the branch and open the pull request with the sections of the pull request template. Verify that CI passes.
