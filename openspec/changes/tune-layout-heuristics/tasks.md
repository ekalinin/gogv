# Tasks

The reduced cases are assigned to mechanisms as the Findings of `add-layout-reduction` and the probes of the proposal place them. A case that the probes show to belong to another mechanism moves to that mechanism's task, and the move is recorded in the Findings of this design.

Each mechanism ends with the regression checks of design D1:

- `make conformance` with the gopuml corpora reports no failed case;
- no case changes its outcome but those that now pass;
- the count of gopuml cases that pass the proxy does not go down from the previous step.

## 1. Acceptance test

- [ ] 1.1 Add `TestReducedCasesPassProxy` (design D7). Record the baseline in the Findings: 273 of 296 gopuml cases pass the proxy, and the outcomes of `conformance-report.md`. Verify:
  - the test passes while the 20 reduced cases are listed;
  - it fails when the entry of a reduced case that still differs is removed.

## 2. Ties in the x positioning (M3)

Cases: `reduced-history-style-1` and its `-worst`, `reduced-entities-relationship-and-attributes-1` and its `-worst`, `reduced-ChenMovie_Test-1`, `reduced-ChenMovieAlias_Test-1`.

- [ ] 2.1 Probe the rule of design D2 on these cases and their variants, and record it with its probes in the Findings. Verify that the rule predicts every probe.
- [ ] 2.2 Implement the rule. Add the tests of the "Head between two tails" and "Tail between a node and a labeled edge" scenarios. Remove the entries of the cases that pass the proxy. Verify:
  - the scenario tests and `TestReducedCasesPassProxy` pass;
  - `state-diagram/history-style-1`, `vega-nonreg/group2681-history_style_all_forms-1` and `chen-eer-diagram/entities-relationship-and-attributes-1`, which M3 alone explains, pass the proxy in the report;
  - the regression checks pass.

## 3. Label room (M1)

Cases: `reduced-choice-fork-join-1`, `reduced-simple-ComponentExtraArrows_0001-1` and its `-worst`, `reduced-state2-1` and its `-worst`, `reduced-state3-1-worst`, `reduced-simple-A0004-1-worst`, `reduced-ChenMovieAlias_Test-1-worst`.

- [ ] 3.1 Probe the room of labels of design D3 on these cases and their variants, and record the extents with their probes in the Findings. Verify that the extents predict every probe, and still predict a single labeled edge.
- [ ] 3.2 Implement the extents. Add the tests of the "Labeled and unlabeled edges between the same nodes" and "Labeled edge in a cycle" scenarios. Remove the entries of the cases that pass the proxy. Verify:
  - the scenario tests and `TestReducedCasesPassProxy` pass;
  - `vega-nonreg/simple-ComponentExtraArrows_0001-1` and `vega-scxml/state2-1`, which M1 alone explains, and `vega-migration/ChenMovieAlias_Test-1` and `vega-migration/ChenMovieExtended_Test-1`, which M3 and M1 explain, pass the proxy in the report;
  - the regression checks pass.

## 4. Ties in the order (M2)

Cases: `reduced-ChenMovie_Test-1-worst`, `reduced-choice-fork-join-1-worst`, `reduced-simple-A0004-1`.

- [ ] 4.1 Probe the tie-break of design D4 on these cases and their variants, and record it with its probes in the Findings. Verify that it predicts every probe, and that it never chooses an order with more crossings.
- [ ] 4.2 Implement the tie-break. Add the test of the "Orders with the same crossings" scenario. Remove the entries of the cases that pass the proxy. Verify:
  - the scenario test and `TestReducedCasesPassProxy` pass;
  - `vega-migration/ChenMovie_Test-1`, `state-diagram/choice-fork-join-1` and `vega-nonreg/simple-A0004-1` pass the proxy in the report;
  - the regression checks pass.

## 5. Flat edges over nodes (M4)

Case: `reduced-state3-1`.

- [ ] 5.1 Probe the curves of design D5, and record the curve and the room it takes with their probes in the Findings. Verify that the curve predicts the control points of every probe within a point.
- [ ] 5.2 Implement the curve and its room. Add the test of the "Flat edge over a node" scenario. Remove the entry of the case if it passes the proxy. Verify:
  - the scenario test and `TestReducedCasesPassProxy` pass;
  - `vega-scxml/state3-1` and `gopuml-clusters/vega-scxml-state3-1` pass the proxy in the report;
  - the regression checks pass.

## 6. Labels of flat edges with ports (M5)

Cases: `reduced-interactive-SVG0006_Svek-1` and its `-worst`.

- [ ] 6.1 Probe the labels of design D6, and record the rule with its probes in the Findings. Verify that the rule predicts every probe within a point.
- [ ] 6.2 Implement the rule. Add the test of the "Flat edge from a port" scenario. Remove the entries of the cases that pass the proxy. Verify:
  - the scenario test and `TestReducedCasesPassProxy` pass;
  - `vega-svg/interactive-SVG0006_Svek-1` passes the proxy in the report;
  - the regression checks pass.

## 7. Acceptance and checks

- [ ] 7.1 Check that the 13 graphs of #28 pass the proxy in the report and that their entries left `known-differences.yaml`. A graph that still fails is reduced again with `tools/reduce -on` on its largest deviation, and its new mechanism is recorded in the Findings and raised before it is fixed. Verify that the proxy section of the report lists none of the 13.
- [ ] 7.2 Judge the other entries of `known-differences.yaml` (design D8): those whose cases pass leave it, and the others name `refine-layout-heuristics`. Verify that `make conformance` reports no failed case, and that no entry names `tune-layout-heuristics`.
- [ ] 7.3 Run the checks:
  - `make lint`;
  - `go test -race ./...`;
  - `go run ./tools/provenance`;
  - the fused multiply-add tests;
  - `make determinism`;
  - `make perf` against `main`;
  - `openspec validate tune-layout-heuristics --strict`.

  Verify that all pass, with no benchmark significantly slower than on `main`.
- [ ] 7.4 Push the branch and open a pull request with the template, with the user's approval. Verify that the CI jobs pass.

## 8. Release

- [ ] 8.1 After the merge, write `docs/releases/v0.4.0.md` and name v0.4.0 on the landing page, in a pull request, with the user's approval (design D9). The notes give the changed layouts and the conformance results. Verify that the tests of `landing_test.go` pass.
- [ ] 8.2 After its merge and a green CI, run `make release VERSION=v0.4.0` with the user's approval. Verify:
  - `go install github.com/ekalinin/gogv/cmd/dot@v0.4.0` prints `gogv 0.4.0` for `dot -V`;
  - the landing page names v0.4.0 and its playground runs it.
