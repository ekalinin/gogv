# Design

## Context

See proposal.md for the motivation, the delta spec of `dot-layout` for the behavior, and the Findings of `openspec/changes/archive/2026-11-07-add-layout-reduction/design.md` for the five mechanisms. The current state that shapes the approach:

- **Positioning:**
  - `placeX` builds the auxiliary graph of section 4.2 of Gansner et al.: Ω is 1, 1 and 8, as observed;
  - `solveConnected` solves each component with `balanceLR`. It slides the part of the tree attached by each tree edge with a zero cut value to the middle of its free range;
  - this centers `x` in `digraph { a -> x; b -> x; c -> y; d -> y }`, but not `y` when the same edges come interleaved. The result thus depends on the tree that the pivots leave.
- **Order:** `order.go` builds an initial order, then runs 24 iterations of medians and transpositions. In `digraph { a -> b [label="LLLLLL"]; a -> c; d -> b }`, both `a, d` and `d, a` have no crossing on the first rank; gogv keeps `d, a`, and the reference `a, d`.
- **Label room:** in `layout.go`, a label node reaches half `nodesep` plus 2.5 points left of the edge on a rank of labels, and its width plus 2.5 points minus half `nodesep` right of it. This was observed on a single labeled edge. A labeled edge gets a chain of its own, as an edge with ports does.
- **Flat edges:**
  - `route.go` draws a flat edge between non-neighbors as one arc just over the rank, which rises 4 points above the nodes in the case of M4;
  - the reference's curve has three Bezier segments and rises 26 points;
  - `layout.go` places the label of a flat edge `flatLabelLift` (3 points) above it, the same with and without ports.
- **Measure:** the proxy of gopuml's contract (`internal/conformance`), and `tools/reduce` with `-on`. The 20 reduced cases are curated cases with references, and the 13 original graphs are in the harvested corpora.
- **Clean-room:** the reference is observed only as a black box. Its rules come from probes, as in the previous changes.

## Goals / Non-Goals

**Goals:**

- The 13 graphs of #28 and their 20 reduced cases pass the proxy, so that gopuml's check passes after the release.
- Each mechanism is fixed by a rule that holds on variants of its reduced cases, not only on the cases themselves.
- No case that passes today fails, and the number of gopuml cases that pass the proxy does not go down.

**Non-Goals:**

- Byte-identical layouts with the reference.
- The known differences that these mechanisms do not explain.

## Decisions

### D1. One method for each mechanism

Each mechanism follows the same steps, and the tasks follow them in order:

1. **Probe.** Lay out its reduced cases and variants of them with the reference, changing one thing at a time: the order of statements, the number of nodes, the sizes, the attributes. Stop when a rule predicts every variant.
2. **Record.** Write the rule, with the probes that support it, in the Findings section of this design before writing code.
3. **Implement** the rule in `internal/layout`, with the scenarios of the delta spec as tests.
4. **Accept** with three checks:
   - its reduced cases pass the proxy against their references;
   - its original graphs pass it in the report;
   - their entries leave `known-differences.yaml`.
5. **Check regressions:**
   - `make conformance` with the gopuml corpora reports no failed case;
   - no case changes its outcome but those that now pass;
   - the count of gopuml cases that pass the proxy does not go down.

Probing before coding keeps the fix a rule of the reference's behavior rather than a fit to five graphs. The reduced cases make each probe a graph of a few nodes.

### D2. Ties in the x positioning (M3)

A node or a part of the drawing that the edges pull equally both ways lies halfway across the range where its cost does not change. Today the balancing works on the parts of the spanning tree that the pivots leave, so it depends on the order of the edges.

The probes answer two questions:

- does the reference's result depend on the order of the statements? The cases of the Context suggest it does not;
- does the reference center only single nodes, or also parts of several nodes? The Chen case moves one node, and `history-style` one node in each component.

The candidates for the implementation are:

- balancing that does not depend on the tree: each node whose pulls are equal moves to the middle of its free range, repeated until nothing moves;
- the same balancing on the tree, but on a tree built in an order that does not depend on the pivots.

The probes choose between them. In both, the cost stays optimal: only positions of equal cost are chosen among.

### D3. Label room (M1)

Probe the room that the reference gives a label:

- two edges between the same nodes, one of them or both labeled;
- a labeled edge reversed by a cycle;
- a label in the second of two components.

For each, measure the extents of the label node left and right of its edge, and the separations on its rank. Then extend the extents of `layout.go` to these cases. The case of a single labeled edge keeps its extents, which match the reference today.

### D4. Ties in the order (M2)

Probe the three reduced cases of M2 and their variants:

- statements in other orders;
- a component in front of the other;
- labels removed.

They tell where the reference's tie-break lies: the initial order (a search from the nodes of the first rank, in input order), or the transpositions (which exchange two nodes only when it removes crossings). The fix goes into that step of `order.go`. It must not change an order that has fewer crossings.

### D5. Flat edges over nodes (M4)

Probe flat edges between non-neighbors:

- one, two and three nodes between them;
- with and without a rank above;
- `ranksep`;
- a second flat edge over the same nodes.

Measure the control points of the reference's curve, and whether the rank above moves away from it. Then route such edges in `route.go` with the reference's curve, and include its highest point in the bounding box and, when the reference does, in the gap between the ranks.

### D6. Labels of flat edges with ports (M5)

Probe a labeled flat edge with a port:

- at its tail, at its head, at both;
- for ports at the top, the middle and the bottom of their table.

Measure the label's center. `vega-svg/interactive-SVG0006_Svek-1` has its label below the rank and its reduced `-worst` case above it, so the rule depends on where the ports are. Implement it where `layout.go` places the labels of flat edges.

### D7. The proxy as a test of the reduced cases

`TestReducedCasesPassProxy`, in `internal/conformance` under the `conformance` build tag:

- every `dot-layout/reduced-*` case that is not listed in `known-differences.yaml` passes the proxy against its reference;
- it is written first, while all 20 are listed, and enforces each mechanism as its cases leave the list.

The report's proxy section keeps covering the gopuml corpora only, as the spec says.

### D8. The other known differences

At the end:

- every entry of `known-differences.yaml` is judged again;
- an entry whose case now passes the contract leaves the file;
- the others name `refine-layout-heuristics` instead of `tune-layout-heuristics`, which this change archives.

### D9. Release

The layouts change, so the release is a minor one: v0.4.0. Its notes and the landing page go in one pull request, as the Releasing section of the README asks: the changed outputs and the conformance results. The tag follows after the merge and a green CI, with the maintainer's approval.

## Risks / Trade-offs

- [Breaking ties changes the layouts of many graphs] → Any graph with a tie can move. The regression checks of D1 run after each mechanism, and an outcome that gets worse stops the mechanism until it is understood.
- [A rule found on a few probes does not generalize] → The probes vary each case, and the gopuml corpora, 296 graphs, check each rule at scale.
- [The mechanisms interact] → A changed order changes positions. The mechanisms go in the order of the tasks: positioning ties, label room, order ties, flat edges, flat labels. Each step's checks see the earlier steps.
- [An original graph still fails after its mechanisms are fixed] → It is reduced again with `tools/reduce -on` on its remaining largest deviation. A new mechanism is recorded in the Findings and raised before it is fixed, since it widens the change.
- [A fix changes the cost of the layout] → Ties are chosen among, not costs: the network simplex keeps its optimum, and the order keeps its number of crossings. `make perf` checks the time.
- [Determinism] → New arithmetic follows CONTRIBUTING: no fused multiply-add, and `internal/fdlibm`.

## Migration Plan

1. Merge the pull request: the layouts change, and the references do not.
2. Release v0.4.0, with its notes and the landing page at v0.4.0, in their own pull request. The push of the tag deploys the page.
3. gopuml updates gogv to v0.4.0 and removes its 12 entries; #28 is closed then.

Rollback: revert the merge, or stay on v0.3.1.

## Findings

To be filled during the implementation: for each mechanism, the rule found and the probes that support it.
