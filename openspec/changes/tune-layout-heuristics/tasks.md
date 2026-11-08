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

The scope follows design D10: the rule found is implemented, and the forms it does not reproduce move to `refine-layout-heuristics`. Cases: `reduced-history-style-1` and its `-worst`; the Chen and `state2` cases move.

- [x] 4.1 Find the path of the reference's network simplex for the x positioning (design D2): the initial tree, the edges that leave and enter, the order of the balancing and its rounding. Compare candidates on the probes of the Findings and on the corpora, and record the one chosen, C1, with its probes and what it does not reproduce. Verify that C1 lays out `history-style-1` and `group2917` as the reference does, and that no case fails.
- [x] 4.2 Implement C1. Add the test of the "Head between two tails" scenario. Remove the entries of the cases that pass. Verify:
  - the scenario test and `TestReducedCasesPassProxy` pass;
  - `state-diagram/history-style-1` and `vega-nonreg/group2681-history_style_all_forms-1` pass the proxy in the report;
  - the regression checks pass.

## 5. Routes through the boxes of labels

Case: `reduced-choice-fork-join-1`; the route of a flat edge through its label, `reduced-interactive-SVG0006_Svek-1`, moves (D10).

- [x] 5.1 Probe the separations of label nodes and virtual nodes, the gates next to a virtual node, and the routes of labeled edges between ranks (Findings, routed curves). Record the rule, the routing order and the gates clear of the curves routed before, with its probes. Verify that it predicts the bend of `reduced-choice-fork-join-1` within a point.
- [x] 5.2 Implement it. Add the test of the "Label clear of an edge routed before" scenario. Remove the entries of the cases that pass. Verify:
  - the scenario test and `TestReducedCasesPassProxy` pass;
  - `reduced-choice-fork-join-1` passes the proxy;
  - the regression checks pass.

## 6. Ties in the order (M2)

Cases: `reduced-ChenMovie_Test-1-worst`, `reduced-choice-fork-join-1-worst`, `reduced-simple-A0004-1`; the order of the label nodes in the originals moves (D10).

- [x] 6.1 Probe the tie-break of design D4 on these cases, their variants and the random benches, and record the flip of equal medians on the passes down and the sources in search order, with their probes and the graphs where they add a crossing. Verify that they predict the three reduced cases and the twelve orders of `f -> g; c -> f; b -> g`.
- [x] 6.2 Implement them. Add the tests of the "Orders with the same crossings" and "Sources in the order of a search" scenarios. Remove the entries of the cases that pass. Verify:
  - the scenario tests and `TestReducedCasesPassProxy` pass;
  - the regression checks pass.

## 7. Edges between the same nodes with labels

Cases: `reduced-simple-ComponentExtraArrows_0001-1` and its `-worst`, which move (D10).

- [x] 7.1 Probe how the reference lays out the edges between the same two nodes when some have labels, and record the probes in the Findings. Verify that the Findings give the gaps measured for `nodesep` 18 to 72.

## 8. Separation next to a rank of its own

Case: `reduced-state3-1-worst`, which moves (D10).

- [x] 8.1 Probe the separation of two components when a node with `rank=source` lies on a rank of its own, and record the probes in the Findings. Verify that the Findings give the gap measured for `nodesep` 1.4 to 72.

## 9. Acceptance and checks

- [x] 9.1 Check which of the 13 graphs of #28 pass the proxy in the report, and that their entries left `known-differences.yaml`. Record the others with the mechanisms they still need, for `refine-layout-heuristics` (D10). Verify that the Findings list each of the 13 with its outcome.
- [x] 9.2 Judge the other entries of `known-differences.yaml` (design D8): those whose cases pass leave it, and the others name `refine-layout-heuristics`. Verify that `make conformance` reports no failed case, and that no entry names `tune-layout-heuristics`.
- [x] 9.3 Run the checks:
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
