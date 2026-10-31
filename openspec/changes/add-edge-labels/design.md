# Design

## Context

See proposal.md for the motivation and the scope of this slice, and the delta specs for the behavior.

Current state:

- `attrs.Check` warns for every edge `label`, `xlabel`, `headlabel`, `taillabel`, `sametail` and edge font attribute, and for every HTML label; the layout ignores them.
- The layout ranks the nodes, adds one chain of virtual nodes per pair of ranked ends (parallel edges share it), orders the ranks, places x by network simplex with symmetric separations of half widths plus `nodesep`, places y rank by rank, and routes splines through the corridors of the virtual nodes.
- Text labels are measured by `makeText` and drawn by `writeText`; nodes and clusters give a text label 16 by 8 points of margins.

Facts observed on the reference that shape the approach:

- A graph with an edge label has a rank between every two of its ranks: in `digraph { a -> b [label="x"] }`, `a` and `b` are 88.8 points apart instead of 72, the label of 16.8 points lies on the middle rank, half `ranksep` (18) from both nodes, and the unlabeled edges of the graph are stretched the same way. With `minlen=2`, the label lies on the middle one of four ranks.
- An edge label sits right of its edge, its left side on the edge: `x` is centered 3.5 points right of the edge, and the drawing widens to hold a long label. In `rankdir=LR` it sits above the edge. Two labeled edges between the same nodes get two curves, one label on each side.
- A flat edge with a label pushes its ends apart, and its label is centered between them, above the edge. A loop's label is right of the loop.
- External labels do not move nodes or edges. A node's `xlabel` touches the upper left corner of the node's box from outside. A `headlabel` without `labelangle` and `labeldistance` touches the edge's end at the head from the left and above, a `taillabel` its start at the tail from the left and below, and an edge's `xlabel` lies left of the edge near its middle.
- With `sametail`, the edges share one tail point on the tail's outline, between the directions of their heads.
- A single-cell table of `WIDTH` 57 and `HEIGHT` 17 is a 57 by 17 points label, with or without `FIXEDSIZE`, and one of `WIDTH` 30 and `HEIGHT` 10 is 30 by 10 points although its cell needs 12 (the reference warns that the table is too small): as an edge label it has no margins; a node and a cluster give it the 16 by 8 points of a text label. The SVG draws it as a polygon filled with `BGCOLOR` without a stroke, a polygon for the table's border and a 6 by 6 points polygon for the empty cell 3 points inside the upper left corner, the borders in the object's pen color. `plain` writes a node's HTML label as written and an edge's as `"<TABLE>"`.
- gopuml's single-cell tables have no other attributes, and none of its graphs sets `labelangle`, `labeldistance` or `sametail`; 4 of the 77 graphs that this slice serves have head or tail labels, 9 have loops, 2 are `rankdir=LR`.

## Goals / Non-Goals

**Goals:**

- Edge labels laid out as the reference lays them out, so that gopuml's graphs with labels pass the similarity contract.
- Table labels drawn so that gopuml finds them in the SVG by their fill color, as it finds them in the reference's output.

**Non-Goals:**

- `labelangle` and `labeldistance`: they place head and tail labels in polar coordinates, which gopuml does not use; without them, the reference places those labels as external labels, which this slice does.
- Tables with several cells, cell content, `PORT`, borders, spacing or padding other than the defaults: slice 4.
- The reference's warning about fixed-size tables smaller than their content.

## Decisions

### D1. Sources

The foundation's clean-room rule holds. Sources added for this slice:

- Gansner, Koutsofios, North and Vo, "A Technique for Drawing Directed Graphs" (IEEE TSE, 1993), section 5.3, edge labels as off-center virtual nodes
- the Graphviz documentation: the `dot(1)` manual page (edge labels as nodes with space in the layout), the attributes `label`, `xlabel`, `headlabel`, `taillabel`, `sametail`, `labelangle`, `labeldistance` and `forcelabels`, the HTML-like labels of "Node Shapes" (elements, attributes and their defaults: `BORDER` 1, `CELLBORDER` 1, `CELLSPACING` 2, `CELLPADDING` 2), and the `plain` output format
- black-box observation of the reference

### D2. Table labels are read once, into a size and a color

A reader in `internal/label` recognizes the single-cell table of the dot-language spec and returns its size, its `FIXEDSIZE` and its `BGCOLOR`; any other HTML label is not a table label and keeps its "not supported yet" warning. `makeText`'s callers get a label that is either text or a table, with one size, so that nodes, clusters, edge labels and external labels size them the same way.

Alternative considered: a general HTML-like label parser now. Rejected: it belongs to slice 4, and gopuml needs only this form here.

### D3. Ranks: doubled when a graph has an edge label

When any edge of the graph has a `label` (text or table), loops and flat edges included, every `minlen` is doubled before ranking and y positions use half `ranksep`; each labeled edge between ranks then has a chain of at least one virtual node, and the label belongs to the virtual node on the middle rank of the chain. Labeled edges get chains of their own instead of sharing them with parallel edges.

Alternative considered: adding label nodes without doubling the ranks. Rejected: the reference doubles them, and the rank distances, and so the sizes of the drawings, would differ by about a fifth.

### D4. Positioning: off-center label nodes, labels placed after routing

The ranked graph gets a left and a right extent per node instead of one width, so that the separations of the auxiliary graph can be off-center. The extents of a label node, relative to its x, are calibrated on the reference: on a rank of labels (an odd rank), half `nodesep` plus 2.5 points to the left and the label's width plus 2.5 points minus half `nodesep` to the right; on a rank of nodes (the middle of an edge with `minlen` 2 or more), `nodesep` to the left and the label's width to the right. The other virtual nodes on a rank of labels are 7 points wide. In `rankdir=LR` the label's height lies along the ranks, and its width across them.

As the reference routes edges in all the free space next to them (Gansner et al., section 5.1.1), the corridor of an edge at its label node is bounded by what comes next on the rank, not by the label node itself: the label must fit between the edge and the next node, a quarter of `nodesep` from the edge of a virtual node, half `nodesep` from the side of a real one. Once the edge is routed, its label is placed with its left side on the rightmost point of the curve across the label's height, as observed on the reference.

### D5. Flat edges and loops

A labeled flat edge between neighbors raises their separation by the widest of their labels, and its label is centered between them, 3 points above the edge; further labeled flat edges between the same neighbors bend away from the first one, alternately below and above, and their labels stack the same way. As in the reference, the labels of flat edges do not grow the drawing.

A node with loops reaches further right, for its neighbors and for the border of its cluster: 18 points, the widths of the loops' labels, and `nodesep` for each loop but the first, as observed on the reference whatever the size of the node and `nodesep`. A loop around others reaches past their labels, and a loop's label lies beyond its far side, on a rank at least as high as the label.

### D6. External labels: placed after the layout

After the layout and the routes, `xlabel`, `headlabel` and `taillabel` are placed at the first free corner around their anchor, in the order upper left, upper right, lower left, lower right: the corner overlaps no node outline, edge label or external label placed before. The anchors are the upper left corner of a node's box, the middle of an edge along its length, and where the edge meets its head and its tail. The drawing grows to hold them.

### D7. sametail

The edges of a `sametail` group share one tail point: the point of the tail's outline in the mean direction from the tail to their heads; each curve keeps the shape it had, moved to start there.

### D8. Output

`svg` writes an edge's labels in its group after the arrowheads, a node's `xlabel` in its group after its label, and table labels as the three polygons of the context. `plain` writes the edge label and its center after the control points, and HTML labels as the reference writes them.

### D9. Conformance

Curated cases are added for every scenario, including tables on nodes, edges and clusters and labeled edges in clusters, recorded with `make reference`. The gopuml corpus is judged as before; the 77 graphs that need nothing beyond this slice are judged instead of being reported as not supported yet. Remaining differences are listed under `tune-layout-heuristics`, as in the earlier slices.

## Risks / Trade-offs

- [The placement of external labels in the reference is a search whose order is not documented] → Calibrate the first candidates on small graphs; gopuml has head or tail labels in 4 graphs only.
- [The `dot(1)` manual page calls flat edge labels "slightly broken" and edge labels between clusters "totally broken" in the reference] → Follow the reference's observed output where it is stable, and list the cases that are not as known differences.
- [Doubling the ranks changes the layout of every graph with an edge label] → This is the reference's behavior; graphs without edge labels keep today's path, which the existing conformance cases guard.

## Migration Plan

No migration: the change only adds behavior. When it is archived, the Purpose of the `dot-layout` main spec stops excluding edge labels, and excludes HTML-like labels other than single-cell tables and node ports instead.
