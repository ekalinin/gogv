# Design

## Context

See proposal.md for the motivation, and the delta spec of `upstream-compatibility` for the behavior. The current state that shapes the approach:

- `internal/conformance`:
  - `ParsePlain` reads the graph size and the names of nodes and edges, without their positions;
  - `Judge` records the SSIM of the `svg` comparison in `CaseResult.SSIM`;
  - `TestConformance` runs the corpora one by one, so it knows the corpus of each case;
  - `Report.Markdown` writes the counts per capability and per format, then the cases that do not pass.
- The references of the harvested corpora are in `.cache/reference/gopuml-dot` and `.cache/reference/gopuml-cluster-dot`, with `canon`, `plain` and `svg`. `tools/refgen` runs the reference with `docker run` on the `gogv-reference` image; one run takes about a second.
- `internal/dotlang` parses DOT into an AST: graphs with their statements (`NodeStmt`, `EdgeStmt`, `AttrStmt`, `AssignStmt`, `Subgraph`), each with its attributes. `internal/graph` has no removal of edges. `canon` breaks attribute lists over several lines, so the lines of its output are not statements.
- The proxy, measured with a script during the exploration of #28 on the `plain` outputs of v0.3.1, against the recorded references, and then by the implementation with the SSIM:
  - 273 of the 296 gopuml cases pass it, and none differs in its nodes or labels;
  - the 23 others are the 12 graphs of #28, the variant with clusters of `vega-scxml/state3-1`, and 10 graphs that gopuml passes. The largest deviation of these 10 goes up to 21%, and one of them, `vega-nonreg/group2917-smetana_hide_state_plain_edges-1`, fails the proxy only by its SSIM of 0.87, with a deviation of 1%.
- Reduced by hand, `vega-scxml/state2-1` differs by 30 points only with its isolated nodes and its labeled flat edge together. Without its isolated nodes, a difference of 1 point remains, from another cause.

## Goals / Non-Goals

**Goals:**

- A measure of #28 that is strict enough to stand in for gopuml's check: a case that passes it passes gopuml's.
- Minimal inputs for each of the 13 graphs, as curated cases, so that `tune-layout-heuristics` starts from causes rather than whole graphs.

**Non-Goals:**

- Changes to the layout, or to the outcome of any case.
- Running gopuml's own conformance: gopuml is checked after the release, when it updates gogv.
- The SSIM of the proxy in the reduction: rasterizing both SVGs for each candidate input would multiply its time, and the differences lie in the positions that `plain` gives.

## Decisions

### D1. Positions in the parsed plain output

`Plain` gets the centers of the nodes, by name, and the centers of the edge labels, by edge. An edge is identified by its tail, its head and its place among the edges between the same nodes. `ComparePlain` keeps comparing names and sizes only, so the contract does not change.

Alternative considered: a second parser for the proxy. Rejected: the statements, the quoting and the continuation lines are already handled once.

### D2. The proxy

`ProxyDeviation(ref, got Plain)` returns the largest deviation and the object that has it. The deviation of the size is its relative difference. The deviation of a center is its distance to the reference's on each axis, divided by the reference's graph width or height. A node or a label that is missing on one side gives a deviation of 1 and names the object. A case passes the proxy when the deviation is at most 0.02 and its SSIM is at least 0.90.

`Run` computes the deviation of every case after `Judge`, from the `plain` outputs it already has, and records it in `CaseResult` next to the SSIM. `TestConformance` passes the results of the corpora marked `Proxy` in `Corpora`, `gopuml` and `gopuml-clusters`, to the report. The proxy does not enter `Judge`, so no outcome depends on it.

### D3. The report

`Report` gets the proxy results of the harvested cases. `Markdown` writes a section `gopuml contract proxy` after the counts per format:

- the number of cases that pass;
- a table of the others, sorted by decreasing deviation, with the case, the deviation in percent and the object that has it.

The section is written only when the run includes a harvested corpus, so the report of a run with the curated corpus alone does not change.

### D4. The reduction tool

`tools/reduce` is an original Go command: `go run ./tools/reduce -in <case.gv> -out <reduced.gv>`, with `make reduce IN=... OUT=...`.

- **Units:** it parses the input with `internal/dotlang`. The units are the statements of the graph and of its subgraphs, the attributes of each statement, and the subgraphs themselves. A subgraph is either removed or replaced by its statements.
- **Printing:** the tool prints an AST back as DOT, with every ID and value quoted, and HTML values between `<` and `>`. A round-trip test checks that the printed input of every curated and gopuml case gives the same `canon` output as the original.
- **Interesting input:** the original's failing objects are the nodes and labels whose deviation exceeds 0.02. A candidate input is interesting when:
  - gogv renders it without an error or a warning;
  - the reference renders it without an error;
  - its `plain` deviation, measured on the objects it still has, exceeds 0.02 on at least one of the original's failing objects.

  The last condition keeps the reduction from drifting to another difference. With `-on`, the objects to keep failing are the ones it names, which must fail in the original, instead of all its failing objects: a reduction anchored on any failing object can keep a minor one and lose the largest deviation, as the first reductions of the 13 graphs did.
- **Search:** delta debugging (Zeller's ddmin) over the statements, then over the attributes of the remaining statements, then the replacement of each subgraph by its statements, repeated until no step removes anything.
- **Reference:** `docker run -i` on the `gogv-reference` image, as `tools/refgen` runs it, for `-Tplain`. The outputs are cached under `.cache/reduce`, by the SHA-256 of the input. gogv runs in the process, through the library.
- **Testing:** the predicate is a function, so the search and the printing are tested with fake references.

Alternatives considered:
- Reducing the lines of the `canon` output. Rejected: its attribute lists span lines, and removing a line can leave a broken statement.
- Removing objects from the `internal/graph` model. Rejected: it has no removal of edges, and the tool would change the library for its own needs.

### D5. Reduced cases and their record

Each of the 13 graphs is reduced twice: anchored on any of its failing objects, into `testdata/conformance/cases/dot-layout/reduced-<case>.gv`, and anchored with `-on` on its node or label of largest deviation, into `reduced-<case>-worst.gv`, where `<case>` is the name of the gopuml case. Their references are recorded with `make reference`. Two graphs that reduce to the same input share one case, named after the first of them in #28:

- the two `history-style` graphs, which are the same graph;
- `vega-scxml/state3-1` and its variant with clusters;
- `vega-migration/ChenMovieAlias_Test-1` and `vega-migration/ChenMovieExtended_Test-1`.

Each reduced case gets an entry in `known-differences.yaml`, with the precise cause of its difference and `change: tune-layout-heuristics`. The entries of the 13 original graphs stay, and name their reduced cases.

The causes are classified in the Findings section of this design, which the implementation fills: the mechanism, the reduced cases that show it, and the original graphs that it explains. This is the input of the proposal of `tune-layout-heuristics`.

## Risks / Trade-offs

- [The reduction drifts to another difference] → The interesting input must keep failing on an object that failed in the original. The reduced cases are read by hand before they are committed. A graph whose reduction hides one of its causes, as `state2` did by hand, is reduced again from a variant without the first cause.
- [The run of the reduction is long] → About a second per reference run, and about n·log n runs for n units: some minutes for the largest of the 13 graphs. The cache keeps a second run cheap.
- [The proxy is stricter than gopuml's check] → Intended for #28: it lists 10 graphs that gopuml passes. It changes no outcome, so it costs no entry in `known-differences.yaml`.
- [The DOT printer changes the meaning of an input] → The round-trip test on every curated and gopuml case.

## Findings

To be filled during the implementation: for each mechanism, the reduced cases that show it and the original graphs that it explains.
