# Tasks

## 1. Branch and inputs

- [x] 1.1 Create the branch `chore/add-performance-check` from an up-to-date `main`, taking the untracked files of this change along. Verify that `git log main..HEAD` is empty and that `git status` shows `openspec/changes/add-performance-check/` among the untracked files.
- [x] 1.2 Add `testdata/bench/classes-60.gv`, `classes-100.gv` and `classes-150.gv`: the DOT that gopuml at #81 or later writes for the class diagrams of issue #29, with the generator and the `save-dot.sh` script of the issue (design D1). Verify that the files have 119, 199 and 299 edge statements, and that gogv's `dot -Tsvg` renders each with exit status 0 and no warnings.

## 2. Benchmark

- [x] 2.1 Add `bench_test.go` at the repository root, in `package gogv_test`, with the SPDX line and `Provenance: original`: `BenchmarkRenderClasses` with the sub-benchmarks `60`, `100` and `150`, each reading its file of `testdata/bench` once and calling `gogv.Render(ctx, src, "svg")` in the loop, failing on an error, with `b.ReportAllocs` (D1). A comment says how the inputs were written. Verify that `go test -run '^$' -bench BenchmarkRenderClasses -benchtime 1x .` runs the three sub-benchmarks and reports `B/op` and `allocs/op`, and that `make lint test provenance` passes.

## 3. Script, make target, CI job and CONTRIBUTING

- [x] 3.1 Write `tools/perf/compare.sh <base>` (D2 to D5, D7): `set -euo pipefail`; a git worktree of the base in a temporary directory, removed on exit; the head's `bench_test.go` and `testdata/bench/` copied into it; both sides built with `go test -c`, and when the base does not build, a message that no comparison was possible and exit status 0; COUNT rounds (default 10), each running the base and then the head binary from its own tree; the text table of benchstat at the pinned version; and the check on its `-format csv` output, which exits 1 naming each benchmark whose `sec/op` is significantly slower than the base by more than THRESHOLD percent (default 20), or says that none is. Verify on a scratch branch: `COUNT=3 tools/perf/compare.sh HEAD` passes and prints the table; with a commit that makes `Render` lay out every graph twice, `COUNT=5 tools/perf/compare.sh <the commit before>` exits 1 and names the three benchmarks; `tools/perf/compare.sh 1602477`, the first commit, which has no `Render`, exits 0 and says that no comparison was possible; and no worktree is left behind (`git worktree list`). Delete the scratch branch afterwards.
- [x] 3.2 Add the target `perf` to the `Makefile`, with a comment like those of the other targets, `BASE ?= main`, COUNT and THRESHOLD passed to the script when set, and `perf` in `.PHONY` (D7). Verify that `make perf COUNT=3` compares the branch with `main` and passes.
- [x] 3.3 Add the job `performance` to `.github/workflows/ci.yml` as design D6 describes, with the actions pinned at the commits the other jobs use. Verify that `git diff .github/workflows/ci.yml` adds the job and changes nothing else, and that the file still parses as YAML (`ruby -ryaml -e 'YAML.load_file(ARGV[0])' .github/workflows/ci.yml`, or another parser at hand).
- [x] 3.4 In the Checks section of `CONTRIBUTING.md`, name `make perf` for a change of the layout or the output, and say that CI runs the same comparison on every pull request and that a red `performance` job means a significant slowdown of more than 20% (D7). Verify that the command runs as written and that no em dash appears in the changed files.

## 4. Verification

- [ ] 4.1 Verify that `make lint test provenance` passes and that `openspec validate add-performance-check --strict` passes.
- [ ] 4.2 With the maintainer's go-ahead, commit, push the branch and open the pull request with the sections of the pull request template. Verify that the `performance` job runs on it, writes the benchstat table into its summary and passes, that the other jobs pass, and record in the pull request the job's duration and the spread (the `±` values) that benchstat reports for the three benchmarks. If a spread comes near the 20% of the threshold, report it to the maintainer before the merge (D5).
