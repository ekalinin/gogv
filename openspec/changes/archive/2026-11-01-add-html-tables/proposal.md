# Proposal

## Why

After slice 3, 17 of the 260 graphs of gopuml's harvested corpus are still reported as not supported yet. 15 of them have nodes labeled with HTML tables of several empty cells, and 14 attach edges to a cell of such a table by its `PORT`; gopuml draws qualified associations and entry and exit points this way. 4 of them also name a port that the node does not define, which the reference only warns about. This is slice 4 of the v1 roadmap: HTML tables and node ports. It is the last slice that gopuml's graphs need: gopuml writes only `svg`.

## What Changes

- Table labels: an HTML label that is a `TABLE` of `TR` rows of empty `TD` cells, with the attributes of the v1 language: on the `TABLE`, `BORDER`, `CELLBORDER`, `CELLSPACING`, `CELLPADDING`, `BGCOLOR`, `COLOR`, `FIXEDSIZE`, `WIDTH` and `HEIGHT`; on a `TD`, `BORDER`, `CELLPADDING`, `CELLSPACING`, `BGCOLOR`, `COLOR`, `COLSPAN`, `FIXEDSIZE`, `WIDTH`, `HEIGHT` and `PORT`. The single-cell tables of slice 3 become a case of these. Tables are laid out and drawn as the reference does: rows and columns as large as their largest cells, cells as large as their row and column, spanning cells widening their columns equally, borders, spacing and padding, backgrounds and border colors.
- Node ports: an edge end written `node:port`, or set by `tailport` and `headport`, that names the `PORT` of a cell attaches to that cell. As in the reference and as section 4 of Gansner et al. describes, x positions keep the ports' offsets from the node centers, the order within ranks follows the ports, and the edge leaves or reaches the cell's box.
- Unknown ports: a port that the node does not define gives the reference's warning (`node a, port P unrecognized`), and the edge attaches to the node itself.
- `svg` output: the title of an edge names its ports, as the reference writes it (`a:p->b`).
- Conformance: curated cases for every scenario; the 17 gopuml graphs are judged instead of being reported as not supported yet.

Still not supported yet, with a warning: compass points on node IDs, ports on loops, cells with content (text, images, nested tables), `ROWSPAN`, the other attributes of HTML labels, HTML labels of the root graph, and HTML text markup. Flat edges attach to their ports like edges between ranks: 3 of the 17 gopuml graphs have them.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `dot-language`: table labels of several empty cells, with their attributes; ports that name table cells, and the warning for unknown ports; compass points stay not supported yet.
- `dot-layout`: the size and layout of tables, and node ports in positioning, ordering and routing.
- `output-formats`: the drawing of tables of several cells, and edge titles with ports in `svg`.
- `upstream-compatibility`: the example of a feature of a later slice moves from node ports to compass points.

## Impact

- Code: `internal/label` (table labels of several cells), `internal/attrs` (supported tables and ports), `internal/layout` (table layout, port offsets in positioning, ordering and routing), `internal/output` (`svg` tables and edge titles).
- Conformance: new curated cases and references; the 17 remaining gopuml graphs are judged; known differences are updated.
- `README.md`: the list of what the current slice supports.
- No change to the library API, the command line, `canon` or `plain`.
