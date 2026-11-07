# Proposal

## Why

Issue #28: 12 of gopuml's conformance cases fail because gogv lays out their graphs differently from Graphviz. gopuml lists them as known differences until gogv lays them out within 5% of Graphviz's size and with the same arrangement. All 32 known differences of gogv name one follow-up change, `tune-layout-heuristics`, whose design cannot be written yet:

- the reasons in `testdata/conformance/known-differences.yaml` describe whole graphs, as observed;
- the first graph reduced by hand, `vega-scxml/state2-1`, split into two independent differences. Its 30 points come from a labeled flat edge together with isolated nodes, not from the label alone.

gogv's own contract also does not tell when gopuml passes. `vega-nonreg/simple-ComponentExtraArrows_0001-1` is 3.9% narrower than the reference in `plain`, within gogv's 5%, but 7.2% narrower in gopuml's image. gopuml draws labels itself, at the positions of the layout.

This change gives the work on #28 a measure and minimal inputs, before any heuristic changes.

## What Changes

- A proxy of gopuml's contract, computed for the cases of the harvested gopuml corpora against the reference:
  - `plain`: the graph width and height each within 2% of the reference's;
  - the center of each node and of each edge label within 2% of the reference's graph size, on each axis;
  - `svg`: an SSIM of at least 0.90.

  The report shows how many cases pass it and lists the others with their largest deviation. It changes no outcome and never fails the run. Today 273 of the 296 gopuml cases pass it. The 23 others are the 12 graphs of #28, the variant with clusters of `vega-scxml/state3-1`, and 10 graphs that gopuml passes.
- `tools/reduce` and `make reduce`: a tool that reduces a DOT input whose layout fails the proxy against the reference. It removes statements, attributes and subgraphs while gogv's layout still fails the proxy on an object that fails in the original, and writes the smallest input it finds.
- The 13 graphs (the 12 of #28 and the variant of `vega-scxml/state3-1`) reduced with it:
  - committed as curated cases with their references, one per distinct reduced graph;
  - each listed as a known difference with the precise reason of its difference;
  - their differences classified by mechanism in the design. This is the input of `tune-layout-heuristics`.

Out of scope:
- changes to the layout: fixing the differences is `tune-layout-heuristics`, which closes #28;
- making the proxy a requirement for passing: it stays a measure in the report;
- the 10 graphs that fail the proxy and that gopuml passes.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `upstream-compatibility`: the proxy of gopuml's contract and its place in the report; the tool that reduces a layout difference.

## Impact

- Code:
  - `internal/conformance`: positions in the parsed `plain` output, the proxy, and its section in the report;
  - `tools/reduce`, with a `reduce` target in the `Makefile`.
- Cases: the reduced graphs under `testdata/conformance/cases/dot-layout`, with their references and their entries in `known-differences.yaml`.
- `conformance-report.md`: the section of the proxy.
- `README.md`: the reduction tool in the Conformance section.
- No change to the library API, the command line or the outputs.
