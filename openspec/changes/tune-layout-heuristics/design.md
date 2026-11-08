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

The first probes (Findings, M3) show that neither candidate alone matches the reference: its ties follow the path of its network simplex. The initial tree of section 4.3 matches it on two forks, but on no reduced case. The work on M3 continues with the experiments that the Findings list: the initial tree, the edges that leave and enter, the order of the balancing and its rounding.

### D3. Label room (M1)

The probes of M1 (Findings) split it into forms with different causes, and each goes with its mechanism:

- a labeled edge in a cycle is a tie of the positioning (D2);
- a label that the reference puts at another point of its curve comes from a route through the box of the label's rank. The route of the label of a flat edge with a port (D6) is the same kind, and both are probed together: the boxes, and the points of them that the path touches;
- for edges between the same nodes with labels, measure the order of the edges, the room of each label and the effect of `nodesep`, then extend the extents of `layout.go` to them;
- for two components next to a node with `rank=source` on a rank of its own, measure the separation of the components.

A label in one of two components needs no change. The case of a single labeled edge keeps its extents, which match the reference today.

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
- [The mechanisms interact] → A changed order changes positions. The mechanisms go in the order of the tasks: flat edges, flat labels, ties in the positioning, routes through the boxes of labels, ties in the order, edges between the same nodes, the separation next to a rank of its own. The ties in the positioning come first among the rest, since they explain the most graphs (Findings, M1). Each step's checks see the earlier steps.
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

### M3: the path of the network simplex (task 4.1, not concluded)

The bench:

- 400 random graphs of 4 to 9 nodes, a tenth of their edges with `minlen=2`, laid out by the reference in one run;
- the gopuml corpora (296 cases) and the curated cases (297), compared in `plain`: the node positions exactly, and the `plain` part of the proxy.

Before any change, gogv lays out 263 of the random graphs as the reference does, and exactly 268 gopuml and 237 curated cases; 274 gopuml cases pass the `plain` part of the proxy.

The reference does not center in general:

- in `a -> b; c -> d; c -> e`, `c` lies 18 points from `d` and 54 from `e` (above);
- in `state-diagram/history-style-1` itself (nodes 22 and 50 points wide, `nodesep` 35, `searchsize` 500), `sh0006` lies 39 and 18 points from its tails, and `sh0007` 17 and 40.

Both are reproduced exactly by the initial tree of section 4.3, followed by gogv's pivots and balancing, with these choices:

- **The edge `f` between two ranks:** the edge from the leftmost node of the upper rank that has edges down, to the leftmost of its lower ends. In `dot-layout/edge-label-lr`, it goes to the label of `a -> c`, which lies left of the label of `a -> b`, the first edge. In `a -> b; b -> d; a -> d`, it is `b -> d`. A tie for the leftmost end of another edge goes to its lower end.
- **The order of the tree's edges:** the edges of the `f`s first, then the separations, then the others. An entering edge takes the place of the leaving one. The balancing takes the tree edges with a zero cut value in this order:
  - in `history-style-1`, only the order "separation `sh0010`-`sh0009`, separation `sh0006`-`sh0007`, then the edge to `sh0007` of `sh0009 -> sh0007`" of the six gives the reference's layout;
  - `vega-nonreg/group2917-smetana_hide_state_plain_edges-1` needs the edges of the `f`s before the separations.
- **The order of the auxiliary graph's edges:** by the rank and the order of their upper end, so that a tie between entering edges goes to the edge further left. The reduced `history-style-1` and `x -> a; x -> b; y -> c` need it.
- **The rounding:** truncation. On the random graphs, rounding down gives 235 matches, up 186, away from zero 174, and truncation 263.

This candidate, C1, keeps the current path for the graphs with clusters, and for those with two adjacent ranks without an edge between them, where section 4.3 does not apply. Its results:

- 270 gopuml cases are exact, and 276 pass the `plain` part of the proxy;
- in the report, 275 of the 296 gopuml cases pass the proxy, against 273. `history-style-1`, `group2681-history_style_all_forms-1` and their reduced cases pass, and `simple-A0004-1` goes from 8.0% to 4.8%;
- 239 curated cases are exact, and none fails;
- 267 random graphs match.

C1 does not explain:

- **The leaving edge among equal cut values.**
  - `b -> e; b -> c; c -> e; d -> a` needs the separation of rank 0 to leave before the edge of the `f` from `b`.
  - `group2917` needs the edge of the `f` from rank 1 to leave before the separation of rank 1.
  - `a -> b; c -> d; c -> e; f -> g` needs the second separation of rank 0 to leave before the first.

  No order of the edges by kind, rank and position satisfies the three.
- **The form of the Chen graphs**, `a -> b; c -> d; c -> e [minlen=2, label="L"]`, the scenario "Tail between a node and a labeled edge". Before the balancing, the reference has `c` over the chain of the label, and then moves it left by 31, half of 63 truncated, and `v6` by 16. C1 has `c` over `d`.

  An order of the edges by rank, with the separations of each rank before its edges down and reversed within the rank, lays out this graph and `reduced-ChenMovie_Test-1` (7.2% to 1.1%) as the reference does. It breaks `group2917` (1.0% to 20.2%).
- **The reduced `entities-relationship-and-attributes`, `ChenMovieAlias_Test` and `state2` cases.** No variant changes them.

C1 is implemented as a first step of task 4.2. After it:

- 607 cases: 522 passed, 46 known differences, 39 not supported yet, 0 failed;
- 275 of 296 gopuml cases pass the proxy;
- `history-style-1`, `group2681-history_style_all_forms-1` and their two reduced cases leave `known-differences.yaml`;
- no other case changes its outcome;
- the outputs are the same on amd64, arm64 and WebAssembly.

On top of C1, every choice among tied leaving and entering edges was enumerated on the probes, with C1's balancing:

- the leaving edge that the reference's layout needs differs from graph to graph:
  - in `b -> e; b -> c; c -> e; d -> a`, the separation `b`-`d`;
  - in `a -> b; c -> d; c -> e; f -> g`, the second separation of rank 0;
  - in `a -> b; c -> d; c -> e [minlen=2, label="L"]`, twice a separation before an edge to an edge node;
  - in `history-style-1`, the edge `sh0008 -> sh0006` before the separation `sh0006`-`sh0007`;
  - in `group2917`, the edges of the `f`s before the separations;
- in `a -> b [label="LLLLLLLLLLLL"]; b -> c; c -> a`, the probe with three ranks of `minlen=2`, and the three probes of the Chen form with `s10`, no path gives the reference's layout. In `digraph { s10 -> s6; s10 -> s8 [minlen=2, label="L"] }`, one path has the reference's positions before the balancing, and its first step puts `s10` where the reference does. Its second step then moves the part that holds `s6` and `s8`, where the reference's result moves only the virtual node between `s10` and `s6`. The reference thus balances another tree at the same positions;
- with the weight 2 of section 4.2 for an edge between a real and a virtual node, the ties of the first graph disappear, but only 142 random graphs match, against 267: the weight 1 holds.

C1's model of the balancing is thus incomplete for these graphs, and a rule for the leaving edge cannot be found from it.

Variants rejected:

- the last of equal edges, on the current tree: `history-style-1` goes from 6.8% to 12.1% and `group2917` to 20.2%, and `chen-eer-diagram/element-styles-1` fails;
- a tree grown from the top left node: it lays out `history-style-1` as the reference does, and puts 11 other gopuml cases further from it;
- a search of the leaving edge in postorder: it breaks `history-style-1`.

### Routes through the boxes of labels (task 5.1, not concluded)

Section 5.3 of Gansner et al. represents the label of an edge between ranks by an off-center virtual node, so that it overlaps no node, edge or label: the edge passes on one side of the node, and the label takes the other.

The probes:

- **`reduced-choice-fork-join-1`.** The reference's curve of `sh0006 -> sh0010` bends at the top left corner of its label, (59.9, 144), in two segments. gogv's label node lies at 57.9, and its straight curve passes at 67.5 at that height, so that its label lies 7.3 points right of the reference's. In both, the label's left side lies on the rightmost point of the curve across the label. The reference's bend leaves its label half `nodesep` left of x = 111.9; gogv puts the virtual node of `sh0006 -> sh0012`, the next node on the rank, at 124.9.
- **`dot-layout/edge-label-fan`.** The reference's three curves are straight, and the left side of each label lies 0.5 points left of the rightmost point of its curve across the label. The label of `a -> b` lies 9.3 points left of the curve of `a -> c`, where gogv's label node keeps 43 points from the next one.

`router.bounds` has a switch, `labelMode`, that keeps the curve left of the label node when it is 1; it is 2. Variants of the gates on the rank of a label:

- **The curve left of the label node (`labelMode` 1).**
  - Better: `state-diagram/history-state-1` and its variant with clusters (18.2% to 1.6%), `state-diagram/junction-1` (21.2% to 2.0%), the two `junction-pseudostate-1` (12.3% to 3.0%), `reduced-choice-fork-join-1` (4.2% to 1.2%), `reduced-state2-1` (6.3% to 1.4%) and its `-worst` (4.6% to 1.0%), `dot-layout/table-edge-not-fixed` (13.4% to 0.7%), `dot-layout/edge-label-long` (7.5% to 0.5%).
  - Worse: `chen-eer-diagram/bold-only-from-the-left-1` (3.9% to 19.4%), `weak-entity-and-identifying-relationship-1` (0.2% to 7.3%), `legacy-activity-diagram/branch-two-outcomes-1` (0.6% to 6.6%), `group2917` (1.0% to 10.1%), `edge-label-fan` (0.5% to 13.9%), `table-edge-small` (0.8% to 8.7%), `output-formats/tooltips-ids` (1.9% to 10.5%), and others by less.

  In `edge-label-fan`, the curve of `a -> b` bends to gogv's label node, 28 points left of where the reference's straight curve passes.
- **Half `nodesep` between a label and the next virtual node, instead of a quarter.**
  - Better: `history-state-1` (18.2% to 13.3%), `junction-1` (21.2% to 17.4%), `simple-ComponentExtraArrows_0001-1` (4.1% to 2.2%), `reduced-state2-1` (6.3% to 4.6%), `edge-label-long` (7.5% to 3.7%).
  - Worse: `reduced-simple-ComponentExtraArrows_0001-1` (7.2% to 10.8%) and `tooltips-ids` (1.9% to 3.2%).

The reduced `state2` cases are thus cases of this mechanism, not of M3.

Conclusion so far: where the reference bends a route, and where it puts the labels, depends on the positions of the virtual nodes: the label nodes and the nodes of the chains. They are ties of the positioning, as in M3, and gogv places them elsewhere. No rule of the gates alone fits the probes.

### M2: ties in the order (tasks 6.1 and 6.2)

Section 3 of Gansner et al. flips the nodes with equal values while sorting by medians and transposing, on every other forward and backward traversal. gogv has the switch, `tieReverse`, and kept it off.

In `digraph { a -> b [label="LLLLLL"]; a -> c; d -> b }`, the scenario of the spec, the initial order has one crossing:

- rank 0: `a`, `d`;
- rank 1: the label node of `a -> b`, the virtual nodes of `a -> c` and `d -> b`;
- rank 2: `b`, `c`.

On the first pass down, the label node and the node of `a -> c` have the same median, that of `a`. Flipped, they put `c` before `b` on rank 2 by their medians, with no crossing: the reference's order. Without the flip, transpose reaches the mirror, as gogv did. The three reduced cases of M2 are the mirrors of the reference's orders in the same way.

A bench compares the order of the real nodes on each rank with the reference's: the 400 random graphs, the gopuml corpora and the curated cases. Of the 48 settings of the four options of `mincross`, flipping the ties on the passes down (`tieReverse: 1`) gives the most matches:

- 360 random graphs against 357, 288 gopuml cases as before, 289 curated cases against 286;
- the three reduced cases of M2 get the reference's order, and no curated or gopuml case loses it;
- the number of crossings goes down in 7 graphs, and up in 1, the random graph `g236` (`digraph { b -> d [minlen=2]; d -> e; c -> g; b -> d; a -> d; a -> c; f -> c; b -> g; c -> g }`): from 0 to 1, where the reference has none.

The second run that section 3 also suggests, from the nodes of the last rank and searching up, kept when it has fewer crossings, gives `g236` the reference's order. With it, no graph gets more crossings, and 8 fewer. It costs 18% to 25% more time on the graphs of `BenchmarkRenderClasses`, above the threshold of `make perf`, and is not taken.

After `tieReverse: 1`:

- the reduced cases of M2 pass the proxy and leave `known-differences.yaml`;
- `state-diagram/choice-fork-join-1` goes from 24.7% to 1.2% in `plain`, and fails the proxy on its SSIM, 0.852, through the labels of task 5. Its entry stays, with the reason updated;
- 607 cases: 525 passed, 43 known differences, 39 not supported yet, 0 failed; 275 of 296 gopuml cases pass the proxy;
- the outputs are the same on amd64, arm64 and WebAssembly.

### Routes through the boxes of labels: the routed curves (tasks 5.1 and 5.2)

The positions of the virtual nodes are not the cause by themselves. Two probes rule out the extents and the gates:

- **The separations.** Two edges pulled under one node by `weight=100`, with ends of `shape=point`, put their virtual nodes against each other. gogv's distances match the reference's within a point, for a label node then a virtual node, a virtual node then a label node, two virtual nodes on a rank of labels or of nodes, and two label nodes, with `nodesep` 18, 36 and 72 and labels of one to eight letters.
- **The gate.** A labeled edge `a -> b` left of `a -> c [weight=100]`, whose virtual node lies under `a`, bends so that its label keeps `floor(nodesep/4)` from that node, 4, 9 and 18 points, in the reference and in gogv alike.

Section 5.1.1 of Gansner et al. gives the rest: after a curve is computed, the boxes of its virtual nodes shrink to the curve, so that the curves computed afterward keep clear of it; and the shorter curves are computed first.

- In `reduced-choice-fork-join-1`, the reference's curve of `sh0006 -> sh0012` reaches 106.3 at the least across the label's height, and 106.3 - 4 - 43 = 59.3 matches the bend at 59.9 within a point.
- In `junction-1`, `edge-label-parallel` and `choice-fork-join`, the reference bends the left edge around the right one, which it thus routes first.
- In `group2917`, the long edge goes around the labels of two shorter edges, routed before it.

The rule implemented:

- the edges are routed by the number of their ranks, the shorter first, and of two as long, the one whose first virtual node lies further right first;
- once a chain is routed, each of its virtual nodes holds how far its curve reaches across the band of its rank;
- the gates of a chain routed later stand off this reach instead of the node's position, with the same room: `floor(nodesep/4)` for a label, half `nodesep` and the extent of the node otherwise;
- the curves of the edges between the same two nodes do not narrow each other's gates, as the reference's `edge-label-parallel` shows.

Results:

- `reduced-choice-fork-join-1` passes the proxy (0.3%) and leaves `known-differences.yaml`, with `table-edge-not-fixed` and `tooltips-ids`, which now pass;
- `history-state-1` and its variant with clusters go from 18.2% to 0.9% and 0.4%, the two `junction-pseudostate-1` from 12.3% to 0.6%, `junction-1` from 21.2% to 4.0%;
- 607 cases: 527 passed, 41 known differences, 39 not supported yet, 0 failed; 279 of 296 gopuml cases pass the proxy;
- `reduced-simple-A0004-1-worst` goes from 6.4% to 7.7%, and `table-edge-small` from 0.8% to 2.5%, within its contract;
- in `docs/example.gv`, the label of `parser -> canon` moves 2.5 points, to 1.9 from the reference's, against 0.6 before;
- the time of `BenchmarkRenderClasses` stays within 3%.

In `x -> l [label="LLLLLLLL"]; x -> r` and its variants of label width and `nodesep`, the label's left side still lies 2 to 3 points right of the reference's (10 before): the curve of `x -> r`, routed first, is not the reference's.

### The order of the label nodes (task 6, continued)

The order bench above compares real nodes only. With the labels taken as nodes of their ranks, 286 gopuml cases match the reference's order, against 288. Ten gopuml cases differ in their order, and eight of them are graphs of #28: the four Chen graphs, `simple-A0004-1`, `simple-ComponentExtraArrows_0001-1`, `state3-1` and its variant with clusters, and `interactive-SVG0006_Svek-1`. In `entities-relationship-and-attributes-1`, the reference puts the labels of `sh0010 -> sh0006` and `sh0010 -> sh0008` outside `sh0007` and `sh0009`, and gogv between them: both orders have no crossing.

A second bench of 400 random graphs with edge labels matches in 332 cases. In most mismatches of both benches, gogv lays out the mirror of the reference's order:

- in 38 of 800 random graphs, the initial order already has no crossing and differs from the reference's;
- in `digraph { f -> g; c -> f; b -> g }`, the reference puts `b` left of `f`, gogv right; in the five other orders of the statements, and with the nodes declared first, gogv matches the reference but when the nodes come as `f, g, c, b`.

Variants of the initial order, on the benches with and without labels:

- the depth-first and breadth-first searches of section 3, from the nodes of the lowest rank, along the out-edges or all edges;
- the sources taken in reverse order, the neighbors in reverse order, or both.

All are worse than gogv's breadth-first search from the sources in node order: 126 to 328 random graphs against 360. The settings of `mincross` do not change the gopuml cases.

A simulation of 60 initial orders on 84 small graphs without crossings, each in up to 12 statement orders, explains at best 59 of them, gogv's 58. The reference's initial order thus follows another rule, or `mincross` changes it from an order with crossings.

The rule found next: the breadth-first search takes the sources in the order of a depth-first search of the component from its first node, along the edges down before those up, instead of in node order.

- In `digraph { f -> g; c -> f; b -> g }`, with the nodes as `f, g, c, b`, the search visits `f, g, b, c`, so that `b` comes first; with `g, f, c, b`, it visits `g, f, c, b`, and `c` comes first. The rule gives the reference's order for the six statement orders and the six declarations of the nodes.
- In the simulation, it gives the reference's order for 66 of the 84 small graphs, and for none of the others: their initial orders have crossings, which `mincross` then changes.
- In gogv, the orders match for 373 random graphs without labels, against 360, and 342 with labels, against 332; the layouts match exactly for 279 random graphs, against 268.
- No gopuml or curated case changes its order or its outputs.

The order of the label nodes in the graphs of #28 is thus not explained by the initial order.

The form of `entities-relationship-and-attributes-1` reduced further: in `digraph { a -> b; c -> d; e -> f; f -> b [minlen=2, label="L"]; f -> d [minlen=2, label="L"] }`, the reference puts the labels outside `a` and `c`, and gogv between them, both without crossing. With `e` declared first, the reference puts them between, and gogv lays out the mirror. Variants that do not explain it:

- the breadth-first search from the last source in search order, placed from the right: 183 to 223 random graphs match, against 373;
- `mincross` iterating on at zero crossings, with the flip of equal medians on the passes down, up or both: 343 to 374 random graphs match, 308 to 340 with labels, and no gopuml case gains the reference's order. By hand, a flip on the first pass up gives the reference's order of this graph, but the next flips undo it, and the paper keeps only an order with fewer crossings.


### Separation next to a rank of its own (task 8.1, not concluded)

`digraph { a -> b; c -> d [label="L"]; subgraph { rank=source; s } }` shows the form of `reduced-state3-1-worst`: the reference puts `c` 59 points right of `a`, gogv 72. Measured with the default sizes:

- the gap between `a` and `c`, sides to sides, is 5 points for `nodesep` up to 21.6, then 9, 19, 37, 54, 82 and 127 points for `nodesep` 25.2, 28.8, 36, 43.2, 54 and 72: about 2.5 `nodesep` less 53;
- it does not depend on the widths of the nodes nor on the width of the label, and a third component keeps the same gap;
- without the label, with `minlen=2` instead of it, with `rank=sink`, `rank=min` or `rank=max`, or with an edge from `s`, the layouts are the same as gogv's.

No mechanism explains these numbers yet.

### Edges between the same nodes with labels (task 7.1, not concluded)

In `digraph { a -> b [label="LLLLLLLLLLLL"]; a -> b }`, the reference's unlabeled curve crosses the height of the label 2.0, 6.8, 11.2 and 16.0 points right of the label, for `nodesep` 18, 36, 54 and 72: about a quarter of `nodesep` less 2.5. With a third edge to another node, the gaps grow by 3.4 to 6.6 points. These are curves, not the virtual nodes they pass by: the rule depends on the extents and positions of the label nodes, as in task 5.

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

### M1: the room of labels (probes before task 4, revised map)

M1 is not one mechanism. Its forms, probed one by one:

- **A labeled edge in a cycle** (`digraph { a -> b [label="LLLLLLLLLLLL"]; b -> c; c -> a }`). The nodes keep the reference's positions relative to each other, and all lie 6.2 points left of the reference's. `a` is pulled equally by the label node and by the virtual node of the reversed edge: a tie of the x positioning (M3).
- **A label in one of two components** (`digraph { x -> y; a -> b [label="LLLLLL"] }`, in both orders, with `nodesep=1`, with an isolated node). gogv lays it out as the reference does.
- **The same with a node of `rank=source` on a rank of its own** (`reduced-state3-1-worst`). The reference puts the two components 13 points closer, so that `sh0038` and `sh0025` are 5.2 points apart, less than `nodesep`. Without that node, or with it on the rank of the others, the layouts are the same.
- **A label not where the reference puts it** (`reduced-choice-fork-join-1`). The nodes are in place, and the reference's curve goes through the top left corner of the label, in two segments: the route passes through the box of the label's rank. gogv draws one curve, and puts the label at its rightmost point.
- **Edges between the same nodes with labels** (`digraph { a -> b [label="LLLLLLLLLLLL"]; a -> b }` and the two edges in the other order, with `nodesep` from 0.25 to 1, measured where the curves cross the height of the label):
  - a labeled edge has a chain of its own, as in gogv;
  - an unlabeled edge after the labeled one runs 107.1 to 122.1 points right of it in the reference, and 118.9 to 160.6 in gogv;
  - an unlabeled edge before the labeled one runs 2.2 to 17.2 points left of it in the reference, and 20.2 to 83.5 in gogv.

The tasks follow this map:

1. the ties of M3, which also cover the cycle;
2. the routes through the boxes of labels, which also cover the route of M5;
3. the ties of M2;
4. the edges between the same nodes;
5. the separation next to a rank of its own.

