# Design

## Context

See proposal.md - Why. What there is today:

- No benchmark in the repository. The CI jobs of `.github/workflows/ci.yml` are lint, test (ubuntu and macOS), conformance and determinism; none measures time. The workflow has `permissions: contents: read`, pins every action by commit, and runs `setup-go` without a cache, since the module has no dependencies.
- An external tool is already fetched for a check: `tools/reference/fetch-tools.sh` downloads resvg 0.48.1 for the conformance comparison, pinned by version and checksum, into `.cache/tools`.
- `gogv.Render(ctx, src, format)` has had the same signature since v0.1.0. The tests of the root package are in `package gogv`.
- On the graphs of issue #29, written by gopuml at #81, gogv takes 0.05 s, 0.28 s and 1.24 s for 60, 100 and 150 classes on the machine of #30, and 20.6 s for 300.
- benchstat from `golang.org/x/perf` at `v0.0.0-20261009192801-be2c69fb417e`, the latest version, prints per benchmark the delta with its p-value, or `~` when the difference is not significant at its default alpha of 0.05; `-format csv` gives the same table as comma-separated values, with the delta in the column `vs base`.

## Goals / Non-Goals

**Goals:**

- Every pull request shows how fast its head lays out the benchmark graphs compared with its base
- A significant slowdown above the threshold turns the job red
- The same comparison runs locally with one command

**Non-Goals:**

- Making the job a required check: like the other jobs, it stays optional for merging
- Comparing with Graphviz in CI, or absolute time limits
- A history of results across commits, or comments on the pull request
- Benchmarks of single layout phases

## Decisions

### D1. The benchmark renders the graphs of issue #29

`BenchmarkRenderClasses` in `bench_test.go` at the repository root runs `gogv.Render(ctx, src, "svg")`, with one sub-benchmark per graph (`60`, `100`, `150`) on `testdata/bench/classes-60.gv`, `classes-100.gv` and `classes-150.gv`, the DOT that gopuml writes for the class diagrams of issue #29. `svg` is what gopuml asks for, and `Render` covers parsing, layout and output as gopuml uses them. These graphs reproduce the slowdown of #29, and 150 classes is the largest that keeps the job short; 300 classes take about 20 s per layout. The benchmark reports allocations as well (`b.ReportAllocs`).

Alternative: time the `dot` binary on these graphs as the issue did. Rejected because process start adds noise, and the timings would carry no allocations and need their own statistics.

### D2. Base and head run the same benchmark

The script checks out the base commit in a temporary git worktree and copies the head's `bench_test.go` and `testdata/bench/` into it. The benchmark is in `package gogv_test` and uses only `Render`, so it builds on any base since v0.1.0, including the base of the pull request that adds it. If the base does not build it, after a later change of the API, the script says that no comparison was possible and exits 0.

Alternative: each side runs its own benchmark. Rejected because the base of the first pull request has none, and a pull request that changes the benchmark would compare different work.

### D3. The two run in turns on one machine

Both sides are built as test binaries with `go test -c`. Then the script runs COUNT rounds (10 by default), each running the base binary and then the head binary once (`-test.run '^$' -test.bench BenchmarkRenderClasses -test.count 1`), each from its own tree so that it reads its `testdata`. Running both on the same machine, in turns, keeps the drift of a shared runner from looking like a difference.

Alternative: compare with results stored from earlier runs on `main`. Rejected because those come from other machines, whose differences can exceed the threshold.

### D4. benchstat compares, at a pinned version

`go run golang.org/x/perf/cmd/benchstat@v0.0.0-20261009192801-be2c69fb417e base.txt head.txt` prints the table. The go command checks the module against the checksum database, as `fetch-tools.sh` checks resvg against its checksum. `go.mod` stays without dependencies. The script prints the text table, and checks the threshold on `-format csv`.

Alternative: a comparator of our own in `tools/`. Rejected by the maintainer: its statistics would have to be written and tested.

### D5. The threshold: significant and more than 20% slower

The script fails when a benchmark's `sec/op` is slower than the base by more than THRESHOLD percent (20 by default) and benchstat finds the difference significant, that is, when its `vs base` column holds a percentage above the threshold rather than `~`. Faster benchmarks never fail. `B/op` and `allocs/op` are in the table but are not checked, because the allocations change with harmless changes too. 20% stays above the run-to-run noise expected on a shared runner, and the slowdown of #29 was 3 to 5 times. The first runs in CI show the actual noise (task 4.2).

Alternative: fail on any significant slowdown. Rejected because small significant differences come from the runner as well as from the code.

### D6. The CI job runs for pull requests only

A `performance` job in `ci.yml`, on `ubuntu-latest`, with `if: github.event_name == 'pull_request'`:

- `actions/checkout` at the pinned commit of the other jobs, with `fetch-depth: 0` so that the base commit is there for the worktree, and `persist-credentials: false`
- `actions/setup-go` as in the other jobs
- `tools/perf/compare.sh "$BASE_SHA"` with `BASE_SHA` set from `github.event.pull_request.base.sha` in `env`, and its output appended to `$GITHUB_STEP_SUMMARY` in a code block

For a pull request, `actions/checkout` checks out the merge commit, so the comparison shows what the pull request changes on its base. The job writes nothing but its summary, so it needs no permission beyond `contents: read`, also for pull requests from forks.

### D7. One script, for CI and locally

`tools/perf/compare.sh <base>` does D2 to D5, with COUNT and THRESHOLD from the environment, and removes its worktree when it exits. `make perf` runs it with `BASE ?= main`. `CONTRIBUTING.md` names `make perf` in its Checks, for a change of the layout or the output, next to `make conformance`.

## Risks / Trade-offs

- [Runner noise above 20% turns the job red without a slowdown] → turns, 10 rounds and the significance test keep it rare; a re-run tells noise from a slowdown; the job is optional
- [A slowdown below 20% passes] → the table shows every significant difference to the reviewer, red or not
- [Every pull request waits a few minutes longer for its checks] → two builds and 10 rounds of about 5 s per side on the machine of #30; task 4.2 records the time on CI
- [The base does not build the head's benchmark after a change of the API] → the summary says so, and the job passes without a comparison
- [benchstat comes from the network] → a pinned version checked against the checksum database; a failed fetch fails only this job
- [Three class diagrams stand for every input] → they cover the layout, where the time of #29 went; more graphs can join `testdata/bench` later
