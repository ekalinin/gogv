## ADDED Requirements

### Requirement: Node ports

An edge between ranks or a flat edge attached to a port SHALL leave or reach the box of the port's cell, aimed at the cell's center; a flat edge between neighbors SHALL go straight from one port to the other. The length of an edge in the x positioning SHALL include the offsets of its ports from the centers of their nodes along the rank, as section 4 of Gansner et al. describes, so that, when nothing else pulls, the node at the other end lies in line with the port. The order of the nodes within a rank SHALL follow the order of the ports that their edges attach to. In `rankdir=LR`, where the nodes of a rank lie from top to bottom, the offset along the rank is the port's vertical offset.

#### Scenario: Node in line with a port

- **WHEN** a node `a` has `shape=plaintext, label=<<TABLE BORDER="0" CELLBORDER="0" CELLSPACING="0" CELLPADDING="0"><TR><TD FIXEDSIZE="TRUE" WIDTH="40" HEIGHT="20" PORT="l"></TD><TD FIXEDSIZE="TRUE" WIDTH="60" HEIGHT="30" PORT="r"></TD></TR></TABLE>>` and the graph has `a:r -> b`
- **THEN** `b` lies 20 points right of the center of `a`, under the center of the cell `r`, and the edge starts at the bottom of that cell

#### Scenario: Flat edge between ports

- **WHEN** nodes `a` and `b` have `shape=plaintext` and the same label, a table without borders, spacing or padding of three rows of three empty cells, whose middle cell is a fixed 40 by 20 points cell with `PORT="h"`, and the graph has `a:h -> b:h [minlen=0]`
- **THEN** `a` and `b` are neighbors of one rank, and the edge goes straight from the right side of the cell `h` of `a` to the left side of the cell `h` of `b`

#### Scenario: Order of the ports

- **WHEN** the graph has the node `a` of the scenario above and the edges `a:r -> b` and `a:l -> c`
- **THEN** `c` lies left of `b`

## MODIFIED Requirements

### Requirement: Table label size

A table label SHALL be laid out as the reference lays it out:

- a cell is `WIDTH` by `HEIGHT` points when its `FIXEDSIZE` is `TRUE`; otherwise it is at least that, and at least as large as its padding and its border on both sides
- a column is as wide as its widest cell, and a row as high as its highest cell; a cell that spans several columns and is wider than they are, with the spacing between them, widens them by equal parts, in fractions of a point if need be, the spanning cells taken in the order of the cells; every cell fills its row and its columns
- `CELLSPACING` separates the cells from each other and from the border of the table; `BORDER` is the width of the table's border, and `CELLBORDER`, or a cell's own `BORDER`, the width of the cells' borders; a cell that does not set its `CELLPADDING` or `BORDER` takes the table's, a cell's own `CELLSPACING` changes nothing, as in the reference, and the defaults are those of the Graphviz documentation (`BORDER` 1, `CELLBORDER` 1, `CELLSPACING` 2, `CELLPADDING` 2)
- a table with `FIXEDSIZE` set to `TRUE` is `WIDTH` by `HEIGHT` points, even when that is smaller than its content, and a side that it does not give takes the size of its content; any other table is at least `WIDTH` by `HEIGHT` points and at least the size of its content; the cells of a table larger than its content stay at its upper left

`WIDTH` and `HEIGHT` SHALL count in whole points, their fractions dropped, as the reference counts them. A node or a cluster SHALL hold a table label with the margins it gives a text label; an edge label, an `xlabel`, a `headlabel` and a `taillabel` SHALL be the size of the table.

#### Scenario: Node sized by its table

- **WHEN** a node has `shape=rect, width=.01, height=.01, label=<<TABLE BGCOLOR="#000007" FIXEDSIZE="TRUE" WIDTH="18" HEIGHT="11"><TR><TD></TD></TR></TABLE>>`
- **THEN** the node is 34 by 19 points

#### Scenario: Table smaller than its cell

- **WHEN** an edge has `label=<<TABLE WIDTH="30" HEIGHT="10"><TR><TD></TD></TR></TABLE>>`, and another one the same label with `FIXEDSIZE="TRUE"`
- **THEN** the first label is 30 by 12 points, as high as a table of one empty cell, and the second one 30 by 10 points

#### Scenario: Edge label table

- **WHEN** an edge has `label=<<TABLE BGCOLOR="#000009" FIXEDSIZE="TRUE" WIDTH="57" HEIGHT="17"><TR><TD></TD></TR></TABLE>>`
- **THEN** its label is 57 by 17 points, with its left side on the edge

#### Scenario: Default borders, spacing and padding

- **WHEN** a node has `shape=plaintext, label=<<TABLE><TR><TD></TD><TD></TD></TR><TR><TD COLSPAN="2"></TD></TR></TABLE>>`
- **THEN** the table is 20 by 20 points, its upper cells are 6 by 6 points and its lower cell 14 by 6 points

#### Scenario: Spanning cell

- **WHEN** a table without borders, spacing or padding has a row of cells 10 and 20 points wide and a row of one cell 60 points wide that spans both columns
- **THEN** the columns are 25 and 35 points wide

#### Scenario: Cell filling its row

- **WHEN** a table without borders, spacing or padding has one row of a fixed 40 by 20 points cell and a fixed 60 by 30 points cell, in a `plaintext` node
- **THEN** the first cell is drawn 40 by 30 points, and the node is 116 by 38 points

#### Scenario: Borders, spacing and padding

- **WHEN** a table has `BORDER="2" CELLBORDER="3" CELLSPACING="4" CELLPADDING="5"` and one row of an empty cell and a cell with `WIDTH="30"`
- **THEN** the table is 62 by 28 points

#### Scenario: Fractions of a point

- **WHEN** a table without borders, spacing or padding has one row of a fixed cell with `WIDTH="50.9" HEIGHT="20.5"`
- **THEN** the cell is 50 by 20 points
