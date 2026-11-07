# Tasks

The reduced cases are assigned to mechanisms as the Findings of `add-layout-reduction` and the probes of the proposal place them. A case that the probes show to belong to another mechanism moves to that mechanism's task, and the move is recorded in the Findings of this design.

Each mechanism ends with the regression checks of design D1:

- `make conformance` with the gopuml corpora reports no failed case;
- no case changes its outcome but those that now pass;
- the count of gopuml cases that pass the proxy does not go down from the previous step.

## 1. Acceptance test

- [x] 1.1 Add `TestReducedCasesPassProxy` (design D7). Record the baseline in the Findings: 273 of 296 gopuml cases pass the proxy, and the outcomes of `conformance-report.md`. Verify:
  - the test passes while the 20 reduced cases are listed;
  - it fails when the entry of a reduced case that still differs is removed.

## 2. Flat edges over nodes (M4)

Case: `reduced-state3-1`.

- [x] 2.1 Probe the curves of design D5, and record the curve and the room it takes with their probes in the Findings. Verify that the curve predicts the control points of every probe within a point.
- [x] 2.2 Implement the curve and its room. Add the test of the "Flat edge over a node" scenario. Remove the entry of the case if it passes the proxy. Verify:
  - the scenario test and `TestReducedCasesPassProxy` pass;
  - the regression checks pass.

## 3. Labels of flat edges with ports (M5)

Cases: `reduced-interactive-SVG0006_Svek-1` and its `-worst`.

- [x] 3.1 Probe the labels of design D6, and record the rule with its probes in the Findings. Verify that the rule predicts every probe within a point, but for the labels below the rank whose curve the reference routes under them (task 5).
- [x] 3.2 Implement the rule. Add the test of the "Flat edge from a port" scenario. Remove the entries of the cases that pass the proxy. Verify:
  - the scenario test and `TestReducedCasesPassProxy` pass;
  - the regression checks pass.

  `vega-svg/interactive-SVG0006_Svek-1` also needs the route of task 5.

## 4. Ties in the x positioning (M3)

The probes of M1 (Findings) show that its form in a cycle is a tie as well. Cases: `reduced-history-style-1` and its `-worst`, `reduced-entities-relationship-and-attributes-1` and its `-worst`, `reduced-ChenMovie_Test-1`, `reduced-ChenMovieAlias_Test-1`, `reduced-state2-1` and its `-worst`. `reduced-simple-A0004-1-worst` and `reduced-ChenMovieAlias_Test-1-worst` are checked after this step, and move to the mechanism that they show.

- [ ] 4.1 Find the path of the reference's network simplex for the x positioning (design D2): the initial tree, the edges that leave and enter, the order of the balancing and its rounding. Compare candidates on the probes of the Findings and on these cases, and record the one chosen with its probes. Verify that it predicts every probe, and that it changes no layout of the curated and gopuml cases that has no tie.
- [ ] 4.2 Implement it. Add the tests of the "Head between two tails", "Tail between a node and a labeled edge" and "Labeled edge in a cycle" scenarios. Remove the entries of the cases that pass the proxy. Verify:
  - the scenario tests and `TestReducedCasesPassProxy` pass;
  - `state-diagram/history-style-1`, `vega-nonreg/group2681-history_style_all_forms-1`, `chen-eer-diagram/entities-relationship-and-attributes-1` and `vega-scxml/state2-1` pass the proxy in the report;
  - the regression checks pass.

## 5. Routes through the boxes of labels

The reference routes an edge with a label through the box of its label's rank: the path touches the corners of the boxes, and the curve follows it with bisected tangents and control points at 4/9, as M4's. Cases: `reduced-choice-fork-join-1` and `reduced-interactive-SVG0006_Svek-1`.

- [ ] 5.1 Probe the routes of labeled edges between ranks and of labeled flat edges with ports (Findings, M1 and M5): the boxes, and the points of them that the path touches. Record the rule with its probes in the Findings. Verify that it predicts the control points of every probe within a point.
- [ ] 5.2 Implement it. Remove the entries of the cases that pass the proxy. Verify:
  - `TestReducedCasesPassProxy` passes;
  - `vega-svg/interactive-SVG0006_Svek-1` passes the proxy in the report;
  - the regression checks pass.

## 6. Ties in the order (M2)

Cases: `reduced-ChenMovie_Test-1-worst`, `reduced-choice-fork-join-1-worst`, `reduced-simple-A0004-1`.

- [ ] 6.1 Probe the tie-break of design D4 on these cases and their variants, and record it with its probes in the Findings. Verify that it predicts every probe, and that it never chooses an order with more crossings.
- [ ] 6.2 Implement the tie-break. Add the test of the "Orders with the same crossings" scenario. Remove the entries of the cases that pass the proxy. Verify:
  - the scenario test and `TestReducedCasesPassProxy` pass;
  - `vega-migration/ChenMovie_Test-1`, `state-diagram/choice-fork-join-1` and `vega-nonreg/simple-A0004-1` pass the proxy in the report, with the earlier steps;
  - the regression checks pass.

## 7. Edges between the same nodes with labels

Cases: `reduced-simple-ComponentExtraArrows_0001-1` and its `-worst`.

- [ ] 7.1 Probe how the reference lays out the edges between the same two nodes when some have labels (Findings, M1): their order, the room of each label, the effect of `nodesep`. Record the rule with its probes in the Findings, then implement it. Add the test of the "Labeled and unlabeled edges between the same nodes" scenario. Verify:
  - the rule predicts every probe within a point;
  - the scenario test and `TestReducedCasesPassProxy` pass;
  - `vega-nonreg/simple-ComponentExtraArrows_0001-1` passes the proxy in the report;
  - the regression checks pass.

## 8. Separation next to a rank of its own

Case: `reduced-state3-1-worst`.

- [ ] 8.1 Probe the separation of two components when a node with `rank=source` lies on a rank of its own (Findings, M1): the reference puts the components closer than `nodesep`. Record the rule with its probes in the Findings, then implement it. Verify:
  - the rule predicts every probe within a point;
  - `TestReducedCasesPassProxy` passes;
  - `vega-scxml/state3-1` and `gopuml-clusters/vega-scxml-state3-1` pass the proxy in the report;
  - the regression checks pass.

## 9. Acceptance and checks

- [ ] 9.1 Check that the 13 graphs of #28 pass the proxy in the report and that their entries left `known-differences.yaml`. A graph that still fails is reduced again with `tools/reduce -on` on its largest deviation, and its new mechanism is recorded in the Findings and raised before it is fixed. Verify that the proxy section of the report lists none of the 13.
- [ ] 9.2 Judge the other entries of `known-differences.yaml` (design D8): those whose cases pass leave it, and the others name `refine-layout-heuristics`. Verify that `make conformance` reports no failed case, and that no entry names `tune-layout-heuristics`.
- [ ] 9.3 Run the checks:
  - `make lint`;
  - `go test -race ./...`;
  - `go run ./tools/provenance`;
  - the fused multiply-add tests;
  - `make determinism`;
  - `make perf` against `main`;
  - `openspec validate tune-layout-heuristics --strict`.

  Verify that all pass, with no benchmark significantly slower than on `main`.
- [ ] 9.4 Push the branch and open a pull request with the template, with the user's approval. Verify that the CI jobs pass.

## 10. Release

- [ ] 10.1 After the merge, write `docs/releases/v0.4.0.md` and name v0.4.0 on the landing page, in a pull request, with the user's approval (design D9). The notes give the changed layouts and the conformance results. Verify that the tests of `landing_test.go` pass.
- [ ] 10.2 After its merge and a green CI, run `make release VERSION=v0.4.0` with the user's approval. Verify:
  - `go install github.com/ekalinin/gogv/cmd/dot@v0.4.0` prints `gogv 0.4.0` for `dot -V`;
  - the landing page names v0.4.0 and its playground runs it.
