# Design

## Context

See proposal.md - Why. The clean-room rules exist today in four places, written for the maintainer:

- decision D1 of `openspec/changes/archive/2026-10-31-add-graphviz-dot-foundation/design.md`: the allowed sources, the sources that are not allowed, and the `Provenance:` header of every Go file; and, among its risks, the rule for AI assistants (every session gets the list of allowed sources, no code of Graphviz or its ports is fetched or quoted)
- `NOTICE`: the kinds of `Provenance:` lines and the notices of fdlibm and the Liberation fonts
- `README.md`, "Clean-room implementation"
- `tools/provenance`: it accepts `original` (alone), `paper <citation>`, `doc https://graphviz.org/...`, `observed dot 16.1.0`, `gopuml <path>.go`, `fdlibm 5.3 <file>.c` (with Sun's notice) and `liberation-fonts 2.1.5`, in the first 15 lines of every non-test Go file outside `.cache`, `openspec` and `testdata`

The repository has no `CONTRIBUTING.md` and no pull request template; `.github/` holds only `workflows/ci.yml`. The history follows Conventional Commits (`type(scope): subject`) on branches prefixed with the type.

## Goals / Non-Goals

**Goals:**

- A contributor finds every rule of D1 in one file, in contributor terms
- Every pull request carries its author's statement that it follows them

**Non-Goals:**

- Enforcing the statement in CI, DCO sign-off, a CLA
- Issue templates, `SECURITY.md`, a code of conduct
- New clean-room rules: the file restates D1 and the existing conventions

## Decisions

### D1. CONTRIBUTING restates the foundation's D1

`CONTRIBUTING.md` lists the allowed and the forbidden sources as D1 does, with the same examples (Smetana in PlantUML, graphviz-pure, goccy/go-graphviz and its forks; gopuml's own code allowed), and the `Provenance:` kinds that `tools/provenance` accepts, with one example line each. It names D1 as the origin of the rules, so that a later change of the rules has one place to start from.

Alternative: a summary that refers to NOTICE and the archived design, which would send contributors to three files.

### D2. The statement covers the change, not the contributor's past

The pull request template asks the author to confirm that they wrote the change without reading or copying the source code of Graphviz or of its ports and bindings, which is what D1 says ("without reading the source code"). The maintainer judges doubtful cases in review.

Alternative: "I have never read the source code of Graphviz", stricter than D1, which would exclude anyone who has ever looked at it.

### D3. AI assistants follow the foundation's mitigation

A contributor who uses an AI assistant gives it the list of allowed sources and does not let it fetch or quote the code of Graphviz or of its ports, as the foundation design does for its own sessions. The template's first checkbox includes the assistant.

### D4. OpenSpec only for behavior in the specs

A change of behavior that `openspec/specs` describes (a shape, an attribute, a format, the command line, the conformance contract) comes with an OpenSpec change in `openspec/changes` (proposal, delta specs, design when needed, tasks), in the same pull request. Fixes that keep the specs true, tests, tools and documentation come as a plain pull request. The maintainer archives the OpenSpec change after the merge, as for the earlier changes.

### D5. Checks, commits, bug reports, license

- Checks: `make lint test provenance` before every pull request, and `make conformance` after `tools/reference/fetch-tools.sh` and `go generate ./internal/textmetrics` for a change of the layout or the output. New conformance cases need their references, recorded with `make reference-image` and `make reference` (Docker). CI runs the same checks.
- Commits: `type(scope): subject` with the types `feat`, `fix`, `docs`, `refactor`, `perf`, `chore`; branches prefixed `feat/`, `fix/`, `refactor/`, `docs/` or `chore/`.
- Bug reports: the DOT input, the output of `dot -V`, gogv's output, and Graphviz 16.1.0's output when they differ. The output of Graphviz is observed behavior, an allowed source.
- License: contributions are under the MIT License of the project; no CLA.

### D6. Files

`CONTRIBUTING.md` at the repository root, which GitHub links from new issues and pull requests; `.github/pull_request_template.md`, which GitHub puts into every new pull request; a Contributing section in the README before License.

## Risks / Trade-offs

- [The checkboxes are a statement of the author, not a proof] → they make the rule explicit for each pull request; the `Provenance:` lines and the review remain the checks
- [The rules drift between CONTRIBUTING, NOTICE and `tools/provenance`] → a new `Provenance:` kind changes all three in the same pull request; CONTRIBUTING says so
- [A contributor has read the Graphviz sources in the past] → D2 scopes the statement to the change; the maintainer decides in review
