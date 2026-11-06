# Proposal

## Why

Issue #29 found gogv 2 to 5 times slower than Graphviz on dense graphs only when gopuml's 30-second layout timeout was hit: nothing in CI measures speed, so a slowdown passes review unseen. #30 fixed that case. A comparison of the speed of every pull request with its base would show the next slowdown while the pull request is still open.

## What Changes

- A benchmark of `gogv.Render` on the DOT that gopuml writes for the class diagrams of issue #29, with 60, 100 and 150 classes, committed as test data.
- A script, also run by `make perf`, that builds the benchmark on the base commit and on the head, runs the two in turns, compares them with benchstat, and fails when a benchmark is slower than the base by more than a threshold, significantly.
- A `performance` job in CI for pull requests that runs the script and writes the comparison into the job summary. Like the other jobs, it is not required for merging.
- `CONTRIBUTING.md` names `make perf` among the checks for a change of the layout or the output.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None. The change adds a benchmark, a script and a CI job; no behavior of the library, the command line or the conformance harness changes, so the change declares `skip_specs`.

## Impact

- New: `bench_test.go`, `testdata/bench/`, `tools/perf/compare.sh`.
- `Makefile`: a `perf` target. `.github/workflows/ci.yml`: a `performance` job. `CONTRIBUTING.md`: the Checks section.
- CI fetches benchstat from `golang.org/x/perf` at a pinned version through the Go module proxy. `go.mod` keeps no dependencies.
- Each pull request gets one more job, of a few minutes.
