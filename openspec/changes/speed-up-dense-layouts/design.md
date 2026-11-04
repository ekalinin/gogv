# Design

## Context

See proposal.md - Why. The measurements below were made on the DOT that gopuml writes for the class diagrams of issue #29.

**Where the time goes.** A CPU profile of gogv on linux/arm64, for 300 classes:

| Part | Time |
|---|---|
| Layout, total | 104 s |
| network simplex for x coordinates | 101.5 s |
| - `exchange` | 90.8 s |
| - - `buildTree` | 58.0 s |
| - - `initCutValues` | 30.7 s |
| - `enterEdge` | 6.7 s |
| - `balanceLR` | 2.1 s |
| mincross | 2.3 s |
| garbage collection, outside Layout | about 16 s, mostly the slices `buildTree` allocates |

The ranking simplex has no pivots on these graphs. The profile for 150 classes has the same shape.

**Same problem, same pivot count, different cost per pivot.** For 150 classes, `dot -v` reports the size of the x-coordinate problem and its iterations:

| | Nodes | Edges | Pivots | Time per pivot |
|---|---|---|---|---|
| reference, Graphviz 16.1.0 | 13466 | 20122 | 7204 | (runs under emulation) |
| Graphviz 15.0.0, Homebrew | 13466 | 20122 | 7204 | 0.2 ms |
| gogv | 13466 | 20122 | 7393 | 0.8 ms |

gogv builds the same auxiliary graph as the reference and makes about as many pivots. The difference is the cost of a pivot.

**The current pivot.** `exchange` in `internal/layout/simplex.go` shifts the ranks of the component that the leaving edge cuts off, swaps the two edges in `inTree`, then calls `buildTree`, which allocates new `treeOut`, `treeIn`, `parent`, `low` and `lim` and redoes the postorder from node 0, and `initCutValues`, which recomputes the cut value of every tree edge. Each pivot therefore costs O(V + E) plus the allocations.

**What the paper says.** Section 2.4 of Gansner et al.: if f = (w,x) is the entering edge, "the only edges whose cut values must be adjusted are those in the path connecting w and x in the tree", found by following the parent edges from w and x to their least common ancestor l, and the postorder numbers "must also be adjusted when exchanging tree edges, but only for nodes below l". Section 4.3 calls these optimizations "crucial" for the x-coordinate pass. `simplex.go` already names section 2.4 in its `Provenance:` line.

**What the conformance suite can show.** On these graphs gogv's x coordinates already differ from the reference's: for 60 classes the ranks and y coordinates match and the x coordinates do not, and the pivot counts differ (7393 and 7204 for 150 classes). Comparing with the reference cannot show that gogv's own output stayed the same.

## Goals / Non-Goals

**Goals:**

- Every output of gogv stays byte for byte as it is, on every input
- A pivot costs work proportional to the part of the tree it changes, without allocations
- The graphs of issue #29 are laid out in about the time Graphviz takes, or less

**Non-Goals:**

- Different pivots: the initial feasible tree that section 4.3 builds from the structure of the auxiliary graph, other choices of leaving or entering edges, or the reference's pivot count. Each of these changes the output and belongs in its own change.
- The differences between gogv's x coordinates and the reference's on dense graphs
- Mincross, routing, and the parts of the simplex that decision D4 leaves as they are

## Decisions

### D1. The output stays the same byte for byte

Every choice the simplex makes reads only these: the set of tree edges, the ranks, the cut values, `parent`, whether a node is in a subtree, and the order of the edge indices. This holds for `leaveEdge`, `enterEdge`, the side that `exchange` moves, and the moves of `balanceLR` and `balanceTB`. Cut values and subtree membership follow from the tree and its root, node 0. They do not depend on the postorder numbers or on the order of the tree adjacency lists, and those are the only things the update computes differently from a full rebuild. `s.order` is read only by the first `initCutValues`. So an update that keeps the tree, the ranks, `parent`, membership and the cut values correct makes the same pivots in the same order and ends with the same ranks.

Alternative: accept changed outputs in exchange for more speed, for example with the initial tree of section 4.3. Rejected for this change because a changed output has to be judged against the reference case by case, and that is a different kind of work from a speed fix.

### D2. The pivot updates the tree as section 2.4 describes

With f the leaving edge, e the entering edge and `sub` the endpoint of f away from the root:

1. Find l, the least common ancestor of e's ends, with the old numbering: from one end, follow `parent` up to the first node whose `low`..`lim` range contains the `lim` of both ends. The cycle that e closes runs through f and lies below l. When f is removed and e added, the subtree of l keeps the same nodes and l keeps its parent edge, so nothing outside l's subtree changes.
2. Shift the ranks of `sub`'s component by the slack of e, as now, using the old numbering.
3. Update `treeOut` and `treeIn` in place: remove f from its two lists and append e.
4. Redo the postorder of l's subtree only, starting from `low[l]`. The numbers stay within l's old range, so the numbers of every other node stay valid. This step sets the new `parent` edges.
5. Recompute, with `localCut`, the cut values of the edges of the cycle other than f. In the new tree they form the path between f's ends, through e and l, so these are the parent edges of the nodes on the new tree paths from each end of f up to l, taken from the ends upward. The paths from e's ends would miss part of the cycle, because e is now a tree edge itself. A node off the cycle keeps its subtree and therefore its cut value. A node on a path gets its child on the path recomputed first, which is the order `localCut` needs.

A pivot then costs the size of l's subtree for the postorder plus the degrees of the nodes on the paths for the cut values. When l is the root, the postorder is still O(V), but it allocates nothing.

Alternative: adjust each cut value on the cycle by plus or minus the cut value of f, the update of the textbook network simplex. Rejected because it needs the orientation of every cycle edge relative to f, while `localCut` already exists and the tests already check it against the definition (`bruteCut`).

### D3. Sameness is checked by comparing gogv with itself

- `internal/corpushash` renders every `.gv` of a directory in all 14 formats and prints a hash per output. It runs on the current code and on the change over `testdata/conformance/cases` (255 cases), `.cache/gopuml-dot` (260), `.cache/gopuml-cluster-dot` (36), `.cache/random` (150), and the graphs of issue #29 for 60, 100 and 150 classes saved as `.gv`. The outputs must match line for line. The graph of 300 classes is compared in svg only, because the current code needs about 125 s for each layout of it, and 11 of the 14 formats lay it out.
- A test in `simplex_test.go` solves each problem twice: once with the update, and once with `buildTree` and `initCutValues` after every pivot, as the code does now. It checks that every pivot picks the same leaving and entering edges, that after every pivot the cut values equal `bruteCut` and subtree membership matches, and that the final ranks are equal. It covers random problems and the balance modes of ranking and x coordinates.
- `make conformance` must leave `conformance-report.md` unchanged.

Alternative: rely on the conformance report alone. Rejected because, as Context shows, the report compares gogv with the reference, and on the dense graphs that differ already it cannot tell whether gogv's output changed.

### D4. `enterEdge`, `leaveEdge` and `balanceLR` stay as they are

After D2 the profile leaves `enterEdge` (6.7 s at 300 classes) and `balanceLR` (2.1 s) as the largest parts. Both scan all edges, `balanceLR` once for every tree edge with a zero cut value. Both could look only at the edges of `sub`'s component and give the same results, with ties broken by (slack, edge index). They are left out until the measurements after D2 show that they are needed.

### D5. Cancellation stays

The loop keeps checking the context every 64 pivots, so the Cancellation requirement of `library-api` stays true.

### D6. The speed is measured as the issue measures it

Wall time of `dot -Tsvg` for 60, 100, 150 and 300 classes, gogv against Graphviz 15.0.0 from Homebrew on the same machine, as in issue #29. The reference image runs under emulation on arm64 and is used only for its `dot -v` pivot counts. From the profiles, about 1 s at 150 classes and about 15 s at 300 are expected, against 1.3 s and 55 s for Graphviz. These are estimates, not measurements.

## Risks / Trade-offs

- [A wrong update leaves a stale cut value, which changes the pivots and the output without failing] → the test of D3 compares with a full rebuild after every pivot, and `corpushash` compares 704 inputs in all formats
- [l is often the root, so the postorder still covers the whole tree] → it allocates nothing and is a fraction of today's cost; D6 measures the result
- [`s.order` is stale after the first pivot] → only the first `initCutValues` reads it; the field's comment says so
- [The gain is an estimate from profiles] → D6 measures before and after; if the graph of 300 classes stays slower than Graphviz, D4 names the next step
