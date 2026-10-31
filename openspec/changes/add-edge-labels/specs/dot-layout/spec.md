## ADDED Requirements

### Requirement: Edge labels

The `label` of an edge between two ranks SHALL be drawn beside the edge, halfway between its ends, with its left side on the edge, and SHALL take room in the layout, as the reference's off-center virtual nodes do (Gansner et al., section 5.3), so that it overlaps no node, edge or other label. As the reference does, a graph with at least one edge label SHALL have its ranks laid out with a rank between every two of its ranks, half `ranksep` apart from them, and each edge label SHALL lie on the rank halfway along its edge. The label SHALL be set in the edge's `fontname`, `fontsize` and `fontcolor`. In `rankdir=LR`, the label SHALL lie above the edge. Two edges with labels between the same nodes SHALL be drawn as two curves, each with its own label.

#### Scenario: Label beside the edge

- **WHEN** the graph is `digraph { a -> b [label="x"] }`
- **THEN** `x` is drawn between `a` and `b`, its left side on the edge, and `a` is farther above `b` than in `digraph { a -> b }`

#### Scenario: Ranks of the other edges

- **WHEN** the graph is `digraph { a -> b [label="x"]; c -> d }`
- **THEN** `c` is on the rank of `a` and `d` on the rank of `b`

#### Scenario: Font of an edge label

- **WHEN** an edge has `label="x"`, `fontname="Helvetica"`, `fontsize=20` and `fontcolor="#FF0000"`
- **THEN** `x` is set in Helvetica, 20 points, in red

#### Scenario: Left to right

- **WHEN** the graph is `digraph { rankdir=LR; a -> b [label="x"] }`
- **THEN** `x` is drawn above the edge, between `a` and `b`

#### Scenario: Two labeled edges between the same nodes

- **WHEN** the graph is `digraph { a -> b [label="x"]; a -> b [label="y"] }`
- **THEN** the two edges are two different curves, `x` beside one and `y` beside the other

### Requirement: Labels of flat edges and loops

The label of an edge between two nodes of the same rank SHALL be drawn above the edge, between its ends, which SHALL be far enough apart for the label to fit. The label of an edge from a node to itself SHALL be drawn beside its loop, beyond the loop's far side.

#### Scenario: Flat edge

- **WHEN** the graph is `digraph { a -> b [minlen=0, label="flat"] }`
- **THEN** `flat` is drawn above the edge, between `a` and `b`, and `a` and `b` are farther apart than in `digraph { a -> b [minlen=0] }`

#### Scenario: Loop

- **WHEN** the graph is `digraph { a -> a [label="loop"] }`
- **THEN** `loop` is drawn to the right of the loop

### Requirement: External labels

The `xlabel` of nodes and edges, and the `headlabel` and `taillabel` of edges, SHALL be placed after the layout, without moving any node or edge, next to their object, as the reference places them when `labelangle` and `labeldistance` are not set: the `xlabel` of a node outside the node, at its upper left; the `xlabel` of an edge beside the middle of the edge; a `headlabel` where the edge reaches its head and a `taillabel` where it leaves its tail, on the side of the edge. They SHALL be set in the object's `fontname`, `fontsize` and `fontcolor`, and the drawing SHALL grow to hold them.

#### Scenario: Node xlabel

- **WHEN** the graph is `digraph { a [xlabel="X"] }`
- **THEN** `X` is drawn outside `a`, left of and above it, and the drawing is wider than without it

#### Scenario: Head and tail labels

- **WHEN** the graph is `digraph { a -> b [headlabel="H", taillabel="T"] }`
- **THEN** `H` is drawn next to the end of the edge at `b` and `T` next to its start at `a`, and `a` and `b` are where they are in `digraph { a -> b }`

#### Scenario: Edge xlabel

- **WHEN** the graph is `digraph { a -> b [xlabel="X"] }`
- **THEN** `X` is drawn beside the middle of the edge, and `a` and `b` are where they are in `digraph { a -> b }`

### Requirement: Same tail

Edges that have the same tail and the same `sametail` value SHALL leave the tail from one point of its outline, as documented by Graphviz.

#### Scenario: Shared tail point

- **WHEN** the graph is `digraph { a -> b [sametail=t]; a -> c [sametail=t]; a -> d }`
- **THEN** the edges to `b` and to `c` start at the same point, and the edge to `d` starts elsewhere

### Requirement: Table label size

A table label SHALL be `WIDTH` by `HEIGHT` points, with or without `FIXEDSIZE`, even when that is smaller than its cell, as the reference does; a side that the table does not give SHALL be 12 points, the size of a table of one empty cell with the default border, cell spacing, cell border and cell padding of the Graphviz documentation. A node or a cluster SHALL hold a table label with the margins it gives a text label; an edge label, an `xlabel`, a `headlabel` and a `taillabel` SHALL be the size of the table.

#### Scenario: Node sized by its table

- **WHEN** a node has `shape=rect, width=.01, height=.01, label=<<TABLE BGCOLOR="#000007" FIXEDSIZE="TRUE" WIDTH="18" HEIGHT="11"><TR><TD></TD></TR></TABLE>>`
- **THEN** the node is 34 by 19 points

#### Scenario: Table smaller than its cell

- **WHEN** an edge has `label=<<TABLE WIDTH="30" HEIGHT="10"><TR><TD></TD></TR></TABLE>>`
- **THEN** its label is 30 by 10 points

#### Scenario: Edge label table

- **WHEN** an edge has `label=<<TABLE BGCOLOR="#000009" FIXEDSIZE="TRUE" WIDTH="57" HEIGHT="17"><TR><TD></TD></TR></TABLE>>`
- **THEN** its label is 57 by 17 points, with its left side on the edge

## MODIFIED Requirements

### Requirement: Cluster labels

A plain-text or table `label` of a cluster SHALL be drawn inside its box, at the top, a plain-text label set in the cluster's `fontname`, `fontsize` and `fontcolor`. With `labeljust=c` (the default) the label SHALL be centered, with `labeljust=l` at the left and with `labeljust=r` at the right of the box. The box SHALL grow to hold its label: it SHALL be at least as wide as the label, and its nodes SHALL be below the label. In `rankdir=LR`, the label SHALL also be at the top of the box.

#### Scenario: Label at the top

- **WHEN** the graph is `digraph { subgraph cluster_a { label="Group"; a } }`
- **THEN** `Group` is drawn inside the box of `cluster_a`, centered, above `a`

#### Scenario: Label at the left

- **WHEN** a cluster has `label="T"` and `labeljust=l`
- **THEN** `T` is drawn at the left of the top of the box

#### Scenario: Label wider than the nodes

- **WHEN** the graph is `digraph { subgraph cluster_a { label="Long cluster title"; a } }`
- **THEN** the box of `cluster_a` is wider than the label

#### Scenario: Table label of a cluster

- **WHEN** a cluster holding a 72 by 36 points node has `label=<<TABLE BGCOLOR="#000007" FIXEDSIZE="TRUE" WIDTH="51" HEIGHT="11"><TR><TD></TD></TR></TABLE>>`
- **THEN** the table is drawn centered at the top of the box, and the box is 88 by 71 points
