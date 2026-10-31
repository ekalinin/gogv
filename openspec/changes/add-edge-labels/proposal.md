# Proposal

## Why

Slices 1 and 2 ignore labels on edges and every HTML label, with a "not supported yet" warning. gopuml labels edges and clusters, and sizes some nodes, with HTML tables of one empty cell of a fixed size and a unique background color, which it finds again in the SVG. In its harvested corpus, 77 of the 260 graphs need nothing else: 51 have edge labels and these tables, 26 have these tables on clusters. Plain-text edge labels alone would serve none of them, so this slice takes these tables as well. This is slice 3 of the v1 roadmap.

## What Changes

- Edge labels: `label` on an edge is drawn beside the edge, halfway between its ends, in the edge's `fontname`, `fontsize` and `fontcolor`. As the reference does, a graph with an edge label is laid out with twice as many ranks, and the label takes room in the layout, so it overlaps no node, edge or other label. Labels of flat edges and loops are drawn above the flat edge and beside the loop.
- External labels: `xlabel` on nodes and edges, and `headlabel` and `taillabel` on edges, are placed after the layout next to their object, as the reference places them when `labelangle` and `labeldistance` are not set.
- `sametail`: edges with the same tail and the same `sametail` value leave their tail from one point.
- Single-cell table labels: an HTML label that is a `TABLE` with one row of one empty cell, with `BGCOLOR`, `FIXEDSIZE`, `WIDTH` and `HEIGHT` on the table, is read as a table label of that size, on nodes, edges (`label`, `xlabel`, `headlabel`, `taillabel`) and clusters, and drawn as the reference draws it: its background, its border and the border of its cell.
- `plain` output: edge lines carry the label and its position.
- Conformance: curated cases for every scenario; the gopuml graphs with edge labels and single-cell tables are judged instead of being reported as not supported yet.

Still not supported yet, with a warning: other HTML labels (several cells, cell content, `PORT`, other attributes), HTML labels of the root graph, `labelangle`, `labeldistance`, `labelfontname`, `labelfontsize`, `labelfontcolor`, `labelfloat`, `decorate`, `samehead`, and node ports. They stay in slice 4 or later.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `dot-language`: single-cell table labels; `xlabel`, `headlabel` and `taillabel` as plain-text labels; the new supported attributes; HTML labels other than single-cell tables stay not supported yet.
- `dot-layout`: edge labels in the layout, labels of flat edges and loops, external labels, `sametail`, and the size of table labels on nodes, edges and clusters.
- `output-formats`: edge labels, external labels and table labels in `svg`, and edge labels in `plain`.
- `upstream-compatibility`: the example of a feature of a later slice moves from edge labels to node ports.

## Impact

- Code: `internal/attrs` (supported attributes and table labels), `internal/label` or a new package for reading single-cell tables, `internal/layout` (ranks for edge labels, label nodes, external label placement, `sametail`), `internal/output` (`svg` and `plain`).
- Conformance: new curated cases and references; the gopuml corpus and its cluster variant get 77 graphs that are judged; known differences are updated.
- `README.md`: the list of what the current slice supports.
- No change to the library API, the command line or `canon`.
