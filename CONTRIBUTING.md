# Contributing to gogv

gogv is a clean-room implementation of the Graphviz `dot` program. The rules
below keep it one: a contribution that does not follow them cannot be merged.

## Clean-room rules

The rules come from decision D1 of the foundation change,
`openspec/changes/archive/2026-10-31-add-graphviz-dot-foundation/design.md`.

### Allowed sources

- published papers, first of all Gansner, Koutsofios, North and Vo, "A
  Technique for Drawing Directed Graphs" (IEEE TSE, 1993), for ranking,
  crossing minimization and positioning, and Dobkin, Gansner, Koutsofios and
  North, "Implementing a General-Purpose Edge Router" (Graph Drawing 1997),
  for edge routing
- the Graphviz documentation (https://graphviz.org/documentation/): the DOT
  language, attributes, node shapes, arrow shapes, output formats
- the observed behavior of the reference, Graphviz 16.1.0 built without
  Pango: its outputs for chosen inputs, its error messages and its exit
  codes. `make reference-image` builds it in a container.
- the code of gopuml (https://github.com/ekalinin/gopuml), which is a port of
  PlantUML, not of Graphviz

Observe the reference with the image that `make reference-image` builds, with
any `-T` format:

```sh
docker run --rm -i --platform linux/amd64 gogv-reference dot -Tsvg < in.gv
```

A Graphviz installed on the host is another version, usually with Pango, and
its outputs are not the reference's.

### Sources that are not allowed

The source code of Graphviz and of its ports and bindings, such as Smetana in
PlantUML, graphviz-pure, and goccy/go-graphviz and its forks. Do not read it
while you work on a change, and do not copy from it.

### AI assistants

If you use an AI assistant, give it the allowed sources above, and do not let
it fetch or quote the code of Graphviz or of its ports.

### Provenance lines

Every Go file starts with the SPDX line and names its sources in
`Provenance:` lines within its first 15 lines, one line per source:

```go
// SPDX-License-Identifier: MIT
// Provenance: paper Gansner, Koutsofios, North, Vo 1993, A Technique for Drawing Directed Graphs, section 2
// Provenance: observed dot 16.1.0
```

| Line | Source |
|---|---|
| `// Provenance: original` | original work of this project; it cannot be combined with another line |
| `// Provenance: paper <citation>` | a published paper and its section |
| `// Provenance: doc https://graphviz.org/<page>` | a page of the Graphviz documentation |
| `// Provenance: observed dot 16.1.0` | the observed behavior of the reference |
| `// Provenance: gopuml <path>.go` | a file of gopuml |
| `// Provenance: fdlibm 5.3 <file>.c` | a translation of fdlibm, which keeps Sun's notice line, as in `internal/fdlibm` |
| `// Provenance: liberation-fonts 2.1.5` | tables generated from the Liberation fonts |

`make provenance` checks the files that are not tests; test files carry the
same header. A new kind of source changes `CONTRIBUTING.md`, `NOTICE` and
`tools/provenance` in the same pull request.

The pull request template asks you to confirm that your change follows these
rules.

## Determinism

gogv writes the same bytes on every platform, and CI compares the outputs of
linux/amd64, macOS/arm64 and the WebAssembly build. Three conventions keep it
so:

- Write a product that is added to or subtracted from as `float64(a*b) + c`.
  The conversion keeps the Go compiler from fusing it into one multiply-add,
  as it does on arm64 and other architectures; `TestNoFusedMultiplyAdd` fails
  on a fused one.
- Take `Sin`, `Cos`, `Atan` and `Atan2` from `internal/fdlibm`; `math.Sqrt` is
  exact and stays.
- Measure text only with the tables of `internal/textmetrics`, generated from
  the Liberation fonts by `go generate ./internal/textmetrics`; fonts installed
  on the host play no part.

## Changes and the specs

`openspec/specs` describes what gogv does. A change of that behavior, such as
a new shape, attribute or output format, or a change of the command line or
of the conformance contract, comes with an
[OpenSpec](https://github.com/Fission-AI/OpenSpec) change in
`openspec/changes` in the same pull request: a proposal, delta specs, a design
when the choices need one, and tasks. Fixes that keep the specs true, tests,
tools and documentation need no OpenSpec change. The maintainer archives the
OpenSpec change after the merge.

## Checks

Go 1.26 or later is needed. Before opening a pull request, run:

```sh
make lint test provenance
```

For a change of the layout or the output, also run the conformance suite:

```sh
tools/reference/fetch-tools.sh
go generate ./internal/textmetrics  # fetch the Liberation fonts
make conformance
```

and compare its speed with `main`:

```sh
make perf
```

CI runs the same comparison on every pull request, against its base. A red
`performance` job means that a benchmark is significantly slower than the
base, by more than 20%.

New cases in `testdata/conformance/cases` need their references, recorded
from the reference with Docker:

```sh
make reference-image
make reference
```

A change also keeps these in step:

- `conformance-report.md`: commit it from a run that includes the harvested
  gopuml corpora (`make harvest`, see the README). A run without them
  rewrites the report with the curated corpus alone; restore the file with
  `git checkout conformance-report.md` before committing.
- `testdata/conformance/known-differences.yaml`: an entry names the case, the
  `reason` of the difference and the `change` that resolves it.
- tests: every scenario of `openspec/specs` has a test, as the
  `Test...Scenarios` tests of `internal/layout` cover the scenarios of one
  requirement each.
- a newly supported attribute or value: `internal/attrs/supported.go`, and
  the Status and "Not supported yet" lists of the README.

CI runs the same checks on every pull request.

## Style

- Headers in sentence case.
- A spaced hyphen ` - ` where a dash is needed.
- Minimal diffs that keep the existing names, so that a change can be
  reviewed and undone without touching what it does not change.
- A renamed heading of this file updates the pointers of `AGENTS.md` in the
  same pull request.

## Commits and branches

Commit messages follow [Conventional Commits](https://conventionalcommits.org):
`type(scope): subject`, as in `feat(layout): place, order and route edges at
node ports`, with the types `feat`, `fix`, `docs`, `refactor`, `perf` and
`chore`. Branch names start with the type: `feat/`, `fix/`, `refactor/`,
`docs/` or `chore/`. Keep one topic per branch.

## Bug reports

Before reporting a missing feature, check the "Not supported yet" list of the
README. A bug report needs:

- the DOT input, as small as you can make it
- the output of `dot -V`
- the command line, or the call of `gogv.Render`, and gogv's output
- when gogv differs from Graphviz, the output of Graphviz 16.1.0 for the same
  input: observed behavior is an allowed source

## License

Contributions are made under the MIT License of the project (`LICENSE`).
There is no contributor license agreement.
