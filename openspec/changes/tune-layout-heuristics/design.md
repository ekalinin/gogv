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

The first probes (Findings, M3) show that neither candidate alone matches the reference: its ties follow the path of its network simplex. The initial tree of section 4.3 matches it on two forks, but on no reduced case. The work on M3 therefore comes after the mechanisms of geometry, and continues with the experiments that the Findings list.

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
- [The mechanisms interact] → A changed order changes positions. The mechanisms go in the order of the tasks: flat edges, flat labels, label room, positioning ties, order ties. The rules of geometry come first; the ties, which depend on the path of the reference's network simplex (Findings, M3), come last. Each step's checks see the earlier steps.
- [An original graph still fails after its mechanisms are fixed] → It is reduced again with `tools/reduce -on` on its remaining largest deviation. A new mechanism is recorded in the Findings and raised before it is fixed, since it widens the change.
- [A fix changes the cost of the layout] → Ties are chosen among, not costs: the network simplex keeps its optimum, and the order keeps its number of crossings. `make perf` checks the time.
- [Determinism] → New arithmetic follows CONTRIBUTING: no fused multiply-add, and `internal/fdlibm`.

## Migration Plan

1. Merge the pull request: the layouts change, and the references do not.
2. Release v0.4.0, with its notes and the landing page at v0.4.0, in their own pull request. The push of the tag deploys the page.
3. gopuml updates gogv to v0.4.0 and removes its 12 entries; #28 is closed then.

Rollback: revert the merge, or stay on v0.3.1.

## Findings

### Baseline

Before any mechanism, on `conformance-report.md` of the change's first commit:

- 607 cases: 516 passed, 52 known differences, 39 not supported yet, 0 failed;
- 273 of the 296 gopuml cases pass the proxy;
- the 52 known differences all name `tune-layout-heuristics`: the 13 graphs of #28, their 20 reduced cases, and 19 others.

Each mechanism below records the rule found, the probes that support it, and the counts after it.

### M3: ties in the x positioning (probes of task 2.1, not concluded)

The reference does not depend on the order of the statements: `digraph { a -> x; b -> x; c -> y; d -> y }` and its five orders tested all center `x` between `a` and `b`, and `y` between `c` and `d`. gogv centers only `x` when the edges of the two components come interleaved. Centering under a larger part matches: a chain below `x` or `y`, four tails into one head.

The ties are decided by the spanning tree that the network simplex ends with, and by the balancing of its tree edges with a zero cut value. Section 4.3 of Gansner et al. describes the initial feasible tree for the auxiliary graph:

- every separation edge;
- for each pair of adjacent ranks, both edges of one ranked edge between them;
- for every other ranked edge, the edge to its leftmost end.

gogv builds it instead from longest paths and a search of tight edges from node 0.

Probes, as x in points (reference / gogv):

| Graph | Reference | gogv | Section 4.3 tree, ties to the lower end |
|---|---|---|---|
| `a -> b; c -> d; c -> e` | c 135, d 117, e 189 | c 135, d 99, e 171 | as the reference |
| `a -> b; c -> d; c -> e [minlen=2]` | c 126, d 112, e 167 | c 126, d 99, e 154 | as the reference |
| `a -> x; c -> y; b -> x; d -> y` | y 207 | y 171 | y 171 |
| `a -> b; c -> d; c -> e [minlen=2, label="L"]` | c 131 | c 146 | c 146 |

In the first graph, the reference's result is that of a tree in which the edge node of `c -> d` hangs on `d`, balanced in the order of the edges: `c` first (36 points), then `{d, e}` (18 points). The section 4.3 tree gives this tree when a tie between the two ends goes to the lower one.

In the last graph:

- gogv's balancing moves `c` by 31, half of a range of 63 truncated, where the reference's result needs 32;
- gogv then moves a part that holds `c` by 16 more, which the reference does not move.

On the corpora, the section 4.3 tree changes the layout of few graphs:

- no reduced case gets closer to the reference;
- `vega-nonreg/group2917-smetana_hide_state_plain_edges-1` goes from a deviation of 1.0% to 13.5%.

Conclusion so far: the ties depend on the whole path of the reference's network simplex, which a black box shows only through its results: the initial tree, the choice of the edges that leave and enter, the order of the balancing and its rounding. The section 4.3 tree is part of it, but not enough alone.

### M4: flat edges over nodes (task 2.1)

The reference reorders a rank so that a flat edge joins neighbors when it can: `{rank=same; a; b; c} a -> c` puts `c` next to `a`. An edge goes over nodes when the order forbids it, as with two flat edges into the same head: `a -> c [minlen=0]; b -> c [minlen=0]`.

Rule, in points, with `top` the top of the rank (its center plus half the height of its highest node) and `ranksep` the separation between ranks as laid out (half of `ranksep` in a graph with edge labels):

- **Path:** the center of the tail, `L`, `R`, and the center of the head. For the k-th of the n flat edges between the same two nodes, in input order:
  - `L` is `k·ranksep/(n+1)` above `top`, at `nodesep/2 − k·nodesep/(n+1)` left of the left side of the first node between them;
  - `R` is at the same height, as far right of the right side of the last node between them.

  With one edge, `L` and `R` are half `ranksep` above `top`, at the sides of the nodes between.
- **Tangents:** at `L` and `R`, the bisector of the angle that the path turns; none at the two centers (section "Spline Fitting" of Dobkin et al.: the tangent at a split point bisects the angle turned).
- **Control points:** each at 4/9 of the length of its segment of the path, along the tangent of its end. That makes three cubic segments: center to `L`, `L` to `R`, `R` to center.
- **Ends:** the curve is clipped at the outline of the tail, and at the arrowhead of the head, as other edges are.
- **Box:** the drawing's box includes the curve, so the drawing is higher than its nodes. The rank above does not move: the gap between the ranks stays `ranksep`.

Probes, the reference's control points against the rule's before the arrowhead (largest distance):

| Graph | Largest distance |
|---|---|
| `a -> c [minlen=0]; b -> c [minlen=0]` | 0.27 |
| the same with `node [width=2]` | 0.72 |
| the same with `node [height=1]` | 0.29 |
| the same with `nodesep=1` | 0.28 |
| the same with `ranksep=1` and a rank above | 0.46 |
| `b` of the same graph 1 inch high | 0.46 |
| two edges `a -> c` over `b` | 0.32 |
| three edges `a -> c` over `b` | 0.27 |
| `ranksep=1`, two edges | 0.51 |

The control points after the start of the arrowhead differ by up to 1.9 points in the script that checked the rule, which shortens the curve by the length of the arrowhead only. gogv's arrowheads also stop short by the miter of their pen, as the reference's do.

Cases the rule does not cover, not met in the corpora:

- with `ranksep=0.1` and edges from the rank above to the tail and to the node between, the reference's path also goes around those edges, in four segments;
- the second of two edges with `nodesep=1` has two segments instead of three.

After M4 (task 2.2):

- gogv draws the probes within 0.4 points of the reference, arrowheads included;
- `reduced-state3-1` passes the proxy and leaves `known-differences.yaml`;
- 607 cases: 517 passed, 51 known differences, 39 not supported yet, 0 failed;
- 273 of 296 gopuml cases pass the proxy, as before: `vega-scxml/state3-1` and its variant also need M1;
- no other case changes its outcome.

### M5: labels of flat edges with ports (tasks 3.1 and 3.2, partly done)

Probes on a table of three cells (ports at the top, the middle and the bottom), and on a cell between two fillers that moves the port from 20 points below the center to 20 above, with a label of 14 and 28 points:

- the label lies below the rank when a port of the edge lies below the center of its node: its top 14 points below the center of the rank;
- otherwise above it, its bottom 11.6 points above the center of the rank, or on the edge where the edge passes higher;
- `ranksep`, a second rank and doubled ranks do not change these heights;
- the drawing's box holds the label.

An end at a port leaves the port's box from the middle of the side that faces the other end, level, its control point 4/9 of the way along; an end without a port gets its control point a sixth of the way towards the other end. A fit of `a:t -> b` and of `a -> b:t` gives 0.3 and 0.8 points.

With these, gogv places the labels of the 32 probes within a point of the reference's, but for the labels below the rank in 28 points, 2.7 to 3.1 points off.

The reference also routes the edge through its label, and gogv does not yet:

- below the rank, the curve dips under the label, so the drawing reaches 4.6 points lower in `vega-svg/interactive-SVG0006_Svek-1` reduced again;
- the path passes through points of the label's bottom side, with bisected tangents and control points at 4/9, as M4's. For text labels, the points are 4.5 points left of the label's center and a label's width right of that, which matches the two probes within 0.5 points. For the table label of SVG0006, the path touches only the bottom right corner.

The points that the path touches thus follow the reference's boxes for the route, which are not modeled yet.

After this step:

- `reduced-interactive-SVG0006_Svek-1-worst` passes the proxy and leaves `known-differences.yaml`;
- `reduced-interactive-SVG0006_Svek-1` passes it in `plain` (0.3%), but its SVG reaches an SSIM of 0.898;
- `vega-svg/interactive-SVG0006_Svek-1` goes from a deviation of 17.4% to 12.6%;
- 607 cases: 518 passed, 50 known differences, 39 not supported yet, 0 failed; 273 of 296 gopuml cases pass the proxy.

