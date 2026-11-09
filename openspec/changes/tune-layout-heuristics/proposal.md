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
- M4. A flat edge over a node: in `digraph { a -> c [minlen=0]; b -> c [minlen=0] }`, the reference's curve of `a -> c` rises 24 points above the nodes, and gogv's 2 points.
- M5. The label of a flat edge from a port of a table: the reference raises it 8.5 points higher than for the same edge without the port.

## What Changes

- The x positioning starts from the initial feasible tree of section 4.3 of Gansner et al., and balances the parts of the tree with a zero cut value in the order of its edges, as the reference does in the forms probed.
- The order within ranks takes the sources of a component in the order of a search of it, and flips the nodes with equal medians on the passes down, as the reference does in the forms probed.
- The shorter edges are routed first, and a routed edge keeps the edges routed after it clear of its curve, as section 5.1.1 of Gansner et al. describes.
- A flat edge between nodes that are not neighbors goes over the nodes between them as the reference's curve does, and the drawing's box includes it.
- The label of a flat edge with a port lies where the reference puts it.
- The cases that pass the contract leave `known-differences.yaml`, among them `history-style-1`, `group2681-history_style_all_forms-1` and the reduced cases of these forms. The other known differences name a new follow-up change, `refine-layout-heuristics`.
- A release with these layouts, so that gopuml can require it and remove the entries of the graphs that pass.

Out of scope, moved to `refine-layout-heuristics` (design D10):
- the other graphs of #28, whose remaining forms the probes of this change do not explain: the order of the label nodes (the Chen graphs, `simple-A0004-1`, `simple-ComponentExtraArrows_0001-1`, `state3-1`, `interactive-SVG0006_Svek-1`), of which the probes isolate one form, two labeled chains from one node; the ties of the positioning that the rule found does not reproduce; the room of labels between the same nodes and in a cycle; and the separation next to a rank of its own;
- any change of the proxy or of the layered similarity contract.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `dot-layout`: how ties are broken in the order within ranks and in the positioning; the order in which edges are routed; the curve of a flat edge over nodes; the label of a flat edge with a port.

## Impact

- Code: `internal/layout`:
  - the order (`order.go`);
  - the positioning and its balancing (`position.go`, `simplex.go`);
  - the order of the routes, and the routes of flat edges and the places of their labels (`route.go`, `layout.go`).
- Tests:
  - the scenarios of the delta spec;
  - a conformance test that requires every reduced case not listed in `known-differences.yaml` to pass the proxy.
- Conformance: entries leave `known-differences.yaml`, and the references do not change.
- Outputs: the layouts of graphs with these ties, labels and flat edges change. Graphs without them keep their outputs, byte for byte.
- Release: notes and the landing page at the new version, then the tag.
