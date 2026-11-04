# Tasks

## 1. Baseline

- [x] 1.1 Create the branch `fix/speed-up-dense-layouts` from an up-to-date `main`, taking the untracked files of this change along. Verify that `git log main..HEAD` is empty and that `git status` shows only `openspec/changes/speed-up-dense-layouts/`.
- [x] 1.2 Write the DOT of issue #29 for 60, 100, 150 and 300 classes, with gopuml at #81 or later and the `save-dot.sh` script as the issue describes, as `.gv` files in a scratch directory outside the repository. Verify that each file has 2n-1 edge statements (119, 199, 299 and 599).
- [x] 1.3 On the current code, build `internal/corpushash` to a scratch binary and save its output for `testdata/conformance/cases`, `.cache/gopuml-dot`, `.cache/gopuml-cluster-dot` (both from `make harvest`), `.cache/random` where present (no make target writes it), and the scratch directory of 1.2 without the graph of 300 classes. Save the `-Tsvg` output of gogv for the graph of 300 classes, and the `all` line that `make determinism` prints (design D3). Verify that every saved file is non-empty and ends with an `all` line, where one is expected.
- [x] 1.4 On the current code, time `dot -Tsvg` for the four graphs of 1.2, gogv against Graphviz 15.0.0 from Homebrew on the same machine, as the median of 3 runs (1 run for 300 classes), as in issue #29 (D6). Verify that the ratios are of the order of the issue's (2 to 4).

## 2. The incremental update

- [x] 2.1 Add a test to `internal/layout/simplex_test.go` that solves each problem twice and steps both solvers pivot by pivot: one with `exchange`, one with a test helper that does what `exchange` does today (shift the ranks of `sub`'s component, swap the edges in `inTree`, `buildTree`, `initCutValues`). After every pivot it checks that both picked the same leaving and entering edges, that every cut value equals `bruteCut`, and that subtree membership agrees; at the end, that both give the same ranks with each of `balanceNone`, `balanceTopBottom` and `balanceLeftRight` (D3). Use problems from `randomDAG`, some with edges of weight 0 as x coordinates have, and check that the problems make pivots. Verify that the test passes on the current code.
- [x] 2.2 Rewrite `exchange` as design D2 describes: find l with the old numbering, shift the ranks as now, remove f from and append e to `treeOut` and `treeIn` in place, redo the postorder of l's subtree from `low[l]` without appending to `s.order`, and recompute with `localCut` the cut values on the paths from f's ends up to l, from the ends upward. Keep `buildTree` and `initCutValues` for the initial tree, keep the context check every 64 pivots (D5), and say in the comment of `s.order` that it holds the order of the initial tree only. Verify that `go test ./internal/layout` passes, including the test of 2.1, and that the test of 2.1 fails when the cut values of step 5 of D2 are not recomputed.

## 3. Verification

- [x] 3.1 Verify that the outputs are unchanged (D3): `corpushash` on the change gives output identical to the files saved in 1.3 (`cmp`), the `-Tsvg` output for 300 classes is identical to the one saved in 1.3, and `make determinism` passes and prints the same `all` line as in 1.3.
- [x] 3.2 Verify that `make lint test provenance` passes, and that `make conformance`, run with the harvested gopuml corpora, passes and leaves `conformance-report.md` unchanged (`git diff --exit-code conformance-report.md`).
- [x] 3.3 Time the four graphs again as in 1.4 (D6), and write the table of before and after. Verify that gogv lays out each graph in about the time Graphviz takes, or less. If the graph of 300 classes is still clearly slower, stop and report the profile instead of widening the change (D4).
- [x] 3.4 Verify that `openspec validate speed-up-dense-layouts --strict` passes. With the maintainer's go-ahead, push the branch and open the pull request with the sections of the pull request template, the table of 3.3, and a reference to issue #29. Verify that CI passes.
