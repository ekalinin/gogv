# Proposal

## Why

Issue #28: 12 of gopuml's conformance cases fail because gogv lays out their graphs differently from Graphviz. gopuml lists them as known differences until a release of gogv lays them out within 5% of Graphviz's size and with the same arrangement.

`add-layout-reduction` reduced these graphs, and the variant with clusters of `vega-scxml/state3-1`, to 20 minimal cases, `dot-layout/reduced-*`, and classified them in five mechanisms (its design, Findings). Small graphs show each one against the reference:

- M3. Ties in the x positioning:
  - in `digraph { a -> x; c -> y; b -> x; d -> y }`, the reference puts `y` halfway between `c` and `d`, and gogv under `c`;
  - in `digraph { a -> b; c -> d; c -> e [minlen=2, label="L"] }`, the reference puts `c` halfway between `d` and `e`, and gogv 15 points right of that.

  The second graph is the form of M1 that the four Chen graphs show. It is a tie too: every position of `c` between `d` and `e` costs the same. M3 thus explains `history-style` and the four Chen graphs.
- M1. In its other forms, the room that edge labels take:
  - two edges between the same nodes, one of them labeled;
  - a labeled edge in a cycle;
  - a label in one of two components.
- M2. Ties in the order within ranks: in `digraph { a -> b [label="LLLLLL"]; a -> c; d -> b }`, gogv lays out the mirror of the reference's order.
- M4. A flat edge over a node: in `digraph { a -> c [minlen=0]; b -> c [minlen=0] }`, the reference's curve of `a -> c` rises 26 points above the nodes, and gogv's 4 points.
- M5. The label of a flat edge from a port of a table: the reference raises it 8.5 points higher than for the same edge without the port.

## What Changes

- The x positioning breaks its ties as the reference does: a node, or a part of the drawing, that edges pull equally both ways lies halfway across the range where its cost does not change.
- Edge labels take the room that the reference gives them when a labeled edge shares its nodes with another edge, lies in a cycle, or is in one of several components.
- The order within ranks breaks its ties as the reference does, so that gogv does not lay out the mirror of the reference's order.
- A flat edge between nodes that are not neighbors goes over the nodes between them as the reference's curve does, and the drawing's box includes it.
- The label of a flat edge with a port lies where the reference puts it.
- The 13 graphs and their 20 reduced cases pass the proxy of gopuml's contract, and leave `known-differences.yaml`. The other known differences are judged again:
  - those that pass leave it;
  - the others name a new follow-up change, `refine-layout-heuristics`.
- A release with these layouts, so that gopuml can require it and remove its own entries.

Out of scope:
- the known differences that these mechanisms do not explain, which stay listed with `refine-layout-heuristics`;
- any change of the proxy or of the layered similarity contract.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `dot-layout`: how ties are broken in the order within ranks and in the positioning; the room of edge labels; the curve of a flat edge over nodes; the label of a flat edge with a port.

## Impact

- Code: `internal/layout`:
  - the order (`order.go`);
  - the positioning and its balancing (`position.go`, `simplex.go`);
  - the room of label nodes (`layout.go`);
  - the routes of flat edges and the places of their labels (`route.go`, `layout.go`).
- Tests:
  - the scenarios of the delta spec;
  - a conformance test that requires every reduced case not listed in `known-differences.yaml` to pass the proxy.
- Conformance: entries leave `known-differences.yaml`, and the references do not change.
- Outputs: the layouts of graphs with these ties, labels and flat edges change. Graphs without them keep their outputs, byte for byte.
- Release: notes and the landing page at the new version, then the tag.
