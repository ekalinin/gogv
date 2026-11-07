# Tasks

## 1. Proxy of gopuml's contract

- [x] 1.1 Read the centers of nodes and edge labels in `ParsePlain`, and add `ProxyDeviation` (design D1, D2). Add tests for the "Close layout" and "Label moved" scenarios, a missing node and a missing label. Verify:
  - the tests pass;
  - `ComparePlain` gives the same results on every curated case.
- [x] 1.2 Compute the proxy in `TestConformance` for the harvested corpora, and write its section in the report (design D3). Add a test for the "Proxy in the report" scenario on a report of fake results, and one showing that a report without a harvested corpus has no such section. Verify, by running `make conformance` with the gopuml corpora:
  - 273 of the 296 cases pass the proxy, and the 23 others are those listed in the proposal;
  - no outcome changes;
  - the report of the curated corpus alone is unchanged.

## 2. Reduction tool

- [x] 2.1 Add `tools/reduce` with the parsing into units and the printing of an AST as DOT (design D4). Add the round-trip test on every curated case, and on every gopuml case when the corpus is collected. Verify that the printed input of each gives the same `canon` output as the original.
- [x] 2.2 Add the search: ddmin over statements, then attributes, then the replacement of subgraphs, with the predicate as a function. Add a test for the "Unrelated statements removed" scenario and one for the "Another difference" scenario, with fake references. Verify that the tests pass.
- [x] 2.3 Add the reference run with its cache, the predicate of design D4, and the `reduce` target of the `Makefile`. Run it on `vega-scxml/state2-1`. Verify:
  - it writes a smaller input that gogv renders without a warning;
  - that input still fails the proxy on one of the original's failing objects;
  - a second run uses the cache and runs no container.

## 3. The 13 graphs

- [x] 3.1 Reduce the 13 graphs, read each reduced input, and reduce again from a variant when a reduction hides a cause. Add the distinct reduced graphs as curated cases under `dot-layout`, and record their references with `make reference` (design D5). Verify:
  - each reduced case fails the proxy against its reference;
  - a second run of `make reference` leaves the references unchanged.
- [x] 3.2 List each reduced case in `known-differences.yaml` with the precise cause of its difference, and name the reduced cases in the entries of the 13 original graphs. Verify that `make conformance` with the gopuml corpora reports no failed case, and that no other case changes its outcome.
- [x] 3.3 Fill the Findings section of the design: each mechanism, the reduced cases that show it, and the original graphs it explains. Verify that every one of the 13 graphs is explained by at least one mechanism.

## 4. Documentation and checks

- [x] 4.1 Describe the proxy and `make reduce` in the Conformance section of `README.md`. Verify that the README tests pass.
- [ ] 4.2 Run the checks:
  - `make lint`;
  - `go test -race ./...`;
  - `go run ./tools/provenance`;
  - `make determinism`;
  - `openspec validate add-layout-reduction --strict`.

  Verify that all pass, and that the hash of `make determinism` is the same as on `main`.
