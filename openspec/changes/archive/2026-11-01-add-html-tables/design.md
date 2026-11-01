# Design

## Context

See proposal.md for the motivation and the scope of this slice, and the delta specs for the behavior.

Current state:

- `label.ParseTable` reads only the single-cell table; `label.Table` holds its size, `FIXEDSIZE` and `BGCOLOR`, as floats. `writeLabel` in `internal/output/svg.go` draws a table as three polygons with the default border, spacing and padding (`label.EmptyCell`, `label.CellInset`).
- `dotlang.Build` already stores `a:p` as the edge's `tailport` or `headport`, and `canon` writes them back. `attrs.Check` warns `tailport=p is not supported yet` for every port, and the layout ignores them.
- The layout gives parallel edges one chain unless they have chains of their own (`own`, as labeled edges do). A `redge` has no ports. The median and the crossing counts compare integer orders. `placeX` pulls the ends of each edge together with auxiliary edges of `minlen` 0. Routing starts at the node centers and clips at the node outlines.

Facts observed on the reference that shape the approach:

- Tables:
  - The defaults of the documentation hold: an empty cell is 6 by 6 points.
  - A grid of two cells over one cell with `COLSPAN="2"` is 20 by 20 points, with a 14 points wide spanning cell.
  - Columns of 10 and 20 points under a 60 points spanning cell become 25 and 35 points wide. The extra width is divided exactly: three columns of 10 points under a 35 points cell become 11.67 points wide, and the spacing between the columns counts as width. Spanning cells are taken in the order of the cells: with a 25 points cell over the first two of those columns in the next row, they become 12.5, 12.5 and 11.67 points wide, and with the rows the other way round 12.5, 12.5 and 10.
  - A cell fills its row and its columns: a fixed 40 by 20 points cell next to a 60 by 30 one is drawn 40 by 30.
  - A table larger than its content, fixed or not, keeps its cells at its upper left. A table without `FIXEDSIZE` is at least as large as its content: `WIDTH="30" HEIGHT="10"` around one empty cell gives 30 by 12 points, not the 30 by 10 that slice 3's scenario says; only `FIXEDSIZE` makes a table smaller than its content.
  - `WIDTH` and `HEIGHT` drop their fractions: 41.9 gives 41, and 50.9 gives 50.
  - The drawing order is: the table's fill, the table's border, then each cell's fill and border. Borders take the cell's `COLOR`, else the table's, else the object's pen color. A border wider than 1 gets a `stroke-width` and is inset by half its width, and a border of 0 draws no polygon.
- Ports:
  - With `a:r -> b`, `b` lies under the center of the cell `r` (20 points right of `a`'s center). The edge starts at the cell's center x, at the bottom of its box.
  - With `a:r -> b` and `a:l -> c`, `c` is left of `b`, and `a` moves to a compromise between its two ports.
  - In `rankdir=LR`, an edge leaves the right side of the cell at the cell's vertical center.
  - A name that the table defines wins over a compass point of the same name.
  - An unknown port gives `node a, port P unrecognized` once per edge end. The edges come by tail node in node order, the tail end before the head end.
  - The `svg` edge title has the ports (`a:P->b`), while the comment before the group does not (`<!-- a->b -->`).
- Flat edges between ports go straight from cell to cell in gopuml's grids, where the port cell is the middle cell of three rows of three. When the port cell spans the height of its table, or sits in its upper left corner, the reference routes around: below the rank, or out of the top of the cell.
- gopuml's tables of several cells:
  - All are on `plaintext` nodes, with empty cells.
  - The `TABLE` has `BORDER`, `CELLBORDER`, `CELLSPACING` and `CELLPADDING` set to 0, and `BGCOLOR` in 2 cases.
  - The `TD` elements have `BGCOLOR`, `COLSPAN`, `FIXEDSIZE`, `WIDTH` and `HEIGHT` (with fractions, and `0.0`) and `PORT`.
- gopuml's ports, in the 17 graphs:
  - Ports are `:h` or `:P`. 4 graphs name a port that the node does not define.
  - 3 graphs have flat edges between ports, and 1 is `rankdir=LR`.
  - None has ports on loops, compass points or `sametail`.

## Goals / Non-Goals

**Goals:**

- Tables laid out and drawn as the reference does, so that gopuml finds its cells by their fill colors in the SVG as it finds them in the reference's output.
- Ports that move nodes, orders and edge ends as in the reference, so that gopuml's qualified associations and entry and exit points pass the similarity contract.
- Single-cell tables keep the output of slice 3, except for the fractions of `WIDTH` and `HEIGHT` and for tables without `FIXEDSIZE` smaller than their content, which now follow the reference.

**Non-Goals:**

- The reference's warnings about table sizes: "fixed cell size with unspecified width or height", and the warning for a table smaller than its content.
- The reference's routes around port cells for flat edges, beyond the straight edge between neighbors.
- Ports in `plain`, which the reference does not write either.

## Decisions

### D1. Sources

The foundation's clean-room rule holds. Sources added for this slice:

- Gansner, Koutsofios, North and Vo, "A Technique for Drawing Directed Graphs" (IEEE TSE, 1993):
  - section 4: node ports as offsets δ on the edges of the auxiliary graph;
  - section 5.1.1: port boxes at the ends of an edge's region.
- the Graphviz documentation:
  - the HTML-like labels of "Node Shapes": elements, attributes and their defaults;
  - the attributes `headport` and `tailport`, and the `portPos` type: port names and compass points.
- black-box observation of the reference

### D2. Table labels become a grid

`label.Table` keeps its name, `ParseTable`, `Size`, and the fields `Width`, `Height`, `Fixed` and `BgColor`. It gains:

- the table's `BORDER`, `CELLBORDER`, `CELLSPACING`, `CELLPADDING`, `COLOR` and `PORT`;
- its rows of cells, each with its own attributes.

`ParseTable` reads the whole form of the dot-language spec and checks each value: whole numbers for borders, spacing, padding and `COLSPAN`, `#rrggbb` colors, `TRUE` or `FALSE` for `FIXEDSIZE`, and non-negative numbers for `WIDTH` and `HEIGHT`, with their fractions dropped. Any other form is not a table label, and keeps its "not supported yet" warning. The single-cell table is the case of one row of one cell, so slice 3's callers are unchanged.

Alternative considered: a second type for grids next to the single-cell table. Rejected: the reference lays both out by the same rules, and two types would give two drawings to keep in step.

### D3. Tables are laid out once, when they are read

`ParseTable` also lays the table out, by the rules of the dot-layout spec:

1. The cells that span one column set the column widths, and every cell sets its row's height.
2. A spanning cell that is wider than its columns widens them by equal parts, in floats, the spanning cells taken in the order of the cells.
3. The table's size follows from the cells, the spacing and the borders, and then from `WIDTH`, `HEIGHT` and `FIXEDSIZE`.

The result is the table's size and each cell's box relative to the table's upper left. `Size` returns that size, and the outputs and the ports read the boxes, so the layout and the drawing cannot disagree. The constants `EmptyCell` and `CellInset` go away with the three-polygon drawing.

Alternative considered: laying out the table in `internal/layout`. Rejected: the size is needed wherever a label is measured (nodes, clusters, edge labels, external labels), and `label` is where text labels are measured too.

### D4. Ports are resolved once per edge end

A table can name its ports: its own `PORT` names the whole table, and each cell's `PORT` names that cell. The layout resolves each edge end once: when the `tailport` or `headport` value names a port of the node's `label` table, the end gets the port's box, centered on the node as the label is. In the layout orientation, the end then has:

- the offset of the box's center from the node's center, along the rank and across it;
- the box, which clips the curve.

Any other value gives the end no port:

- a compass point, alone or after a port;
- an unknown name;
- a port on a loop.

Warnings stay at the language level, as they are today:

- `attrs.Check` treats `tailport` and `headport` as supported. It warns `not supported yet`, using `center`, for a compass point alone and for a port on a loop, and using the port for a compass point after a port.
- A new `attrs` function returns the reference's `node a, port P unrecognized` texts, in the reference's order: edges by tail node in node order, the tail end before the head end. `gogv.go` appends them after the warnings of `attrs.Check`, as plain texts like those of `ResolveClusters`, so they do not count as "not supported yet".

### D5. Ordering: positions inside a node

An edge with a port gets a chain of its own, as a labeled edge does, and the `redge`s at its ends carry their port offsets along the rank. Where the median and the crossing counts read a neighbor's position, they read its order plus a fraction smaller than one half. The fraction grows with the port's offset across the node's width, so ports order the edges of one node without passing its neighbors. Edges without ports read the order alone, so graphs without ports keep their orders.

Alternative considered: ordering by the nodes alone and sorting the edges of each node afterwards. Rejected: the reference puts `c` left of `b` for `a:l -> c` and `a:r -> b`, which only an order that sees the ports gives.

### D6. Positioning: port offsets as δ

As section 4 of the paper describes, the auxiliary edges of an edge whose ends have the offsets `pu` (upper) and `pv` (lower) along the rank get `minlen` `m - pu` and `m - pv`, with `m = max(pu, pv)`. Both stay non-negative, and the cost becomes `|(xu + pu) - (xv + pv)|`, plus a constant. The offsets are rounded to whole points, as the x positions are. Flat edges get the same treatment for their pull. In `rankdir=LR`, the offset along the rank comes from the cell's vertical offset in the final orientation, turned into the layout orientation as node sizes are.

### D7. Routing: from the port's box

An end at the port of a cell starts from the middle of the side of the cell's box that faces the other end, and its next control point lies straight across the ranks from it, as the reference's edges leave their cells (`M88,-75C88,-65.77 ...` under the cell `r`). An end at the table's own `PORT` aims at the center, as at a node, and the curve is clipped by the table's box instead of the node's outline. For a chain, this moves its first and last points; the corridor between them is unchanged. A flat edge between neighbors goes straight from one port point to the other, and a flat arc starts at the port points. For a tail with a port, `sametail` does not move the start, because the port gives it.

### D8. Output

`writeLabel` draws a table from its boxes, in the order of the context:

- the table's fill when it has a `BGCOLOR`, and the table's border when its border is not 0;
- each cell's fill and border, likewise.

A border wider than 1 is written with `stroke-width` and inset by half its width. The halves are kept apart from the sums, as for slice 3's tables, against fused multiply-adds. The edge group's `title` adds `:port` after each end that has a `tailport` or `headport`, as written; the comment before the group stays `tail->head`. `plain` is unchanged.

### D9. Conformance

Curated cases are added for every scenario, recorded with `make reference`:

- tables with spans, fills, fixed sizes, fractions, borders, spacing, padding and colors;
- ports in `TB` and `LR`, two ports of one node, flat edges between ports, unknown ports, a port named like a compass point, a port on a loop, and compass points.

The "not supported yet" fixture of the conformance tests moves from a port to a compass point. The 17 gopuml graphs are judged instead of being reported as not supported yet. Remaining differences are listed under `tune-layout-heuristics`, as in the earlier slices, and slice 3's entries are re-checked now that table sizes drop their fractions.

## Risks / Trade-offs

- [The reference routes some flat edges between ports around the cells, by a rule that the documentation does not give, and raises the label of a labeled flat edge between ports above the rank, routing the edge through it] → The straight edge matches the 3 gopuml graphs where the edge has no label, whose port cells are the middle cells of their grids; other cases become known differences, as `vega-svg/interactive-SVG0006_Svek-1` with its labeled flat edges.
- [The fraction that ports add to positions in the ordering is this design's own model] → Curated cases with two and three ports on one node check the orders against the reference, and the gopuml graphs check them at scale.
- [A node pulled by two ports settles where the auxiliary graph's weights and separations put it] → δ goes into the same auxiliary graph with the same ω as the reference's layout. The curated case `a:r -> b; a:l -> c` checks the compromise (`a` at 68 points, `b` at 99, `c` at 27).
- [Generalizing the table, dropping fractions and growing tables without `FIXEDSIZE` to their content change some of slice 3's tables] → The change moves them towards the reference. The existing curated references guard the single-cell drawing, and the gopuml corpora are re-run.
- [Rounding the port offsets to whole points] → The rest is at most half a point, within the tolerance of the similarity contract.

## Migration Plan

No migration: the change only adds behavior. When it is archived, the Purpose of the `dot-layout` main spec stops excluding node ports and HTML-like labels other than single-cell tables. It excludes compass points, table cells with content and the other HTML-like labels instead.
