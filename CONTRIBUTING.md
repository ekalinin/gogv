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

New cases in `testdata/conformance/cases` need their references, recorded
from the reference with Docker:

```sh
make reference-image
make reference
```

CI runs the same checks on every pull request.

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
