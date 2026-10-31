# dot-layout Specification

## Purpose

Computes the positions of nodes and the routes of edges as the Graphviz `dot` layout engine does, for graphs without node ports and HTML-like labels other than single-cell tables, with text measured from bundled font metrics.

## Requirements

### Requirement: Ranking

The system SHALL place every node on a rank so that, for every edge with `constraint=true` (the default), the rank of its head minus the rank of its tail is at least the edge's `minlen` (default 1), except where the requirement "Ranking with clusters" allows otherwise. Among such placements, it SHALL keep edges short, as the network simplex ranking of Gansner et al. does. Edges with `constraint=false` SHALL NOT affect ranks. A graph with cycles SHALL be laid out: edges reversed for ranking SHALL still be drawn from their tail to their head.

#### Scenario: Shortest ranking

- **WHEN** the graph is `digraph { a -> b -> c; a -> c }`
- **THEN** `a`, `b` and `c` are on three consecutive ranks

#### Scenario: Minimum length

- **WHEN** the graph is `digraph { a -> b [minlen=2] }`
- **THEN** `b` is two ranks below `a`

#### Scenario: Edge of length zero

- **WHEN** the graph is `digraph { a -> b [minlen=0] }`
- **THEN** `a` and `b` are on the same rank

#### Scenario: Edge that does not constrain

- **WHEN** the graph is `digraph { a -> b [constraint=false] }`
- **THEN** `a` and `b` are on the same rank, and the edge is drawn

#### Scenario: Cycle

- **WHEN** the graph is `digraph { a -> b -> a }`
- **THEN** the layout has two ranks, and each edge has its arrowhead at its own head

### Requirement: Rank direction

With `rankdir=TB` (the default), ranks SHALL go from top to bottom. With `rankdir=LR`, ranks SHALL go from left to right.

#### Scenario: Left to right

- **WHEN** the graph is `digraph { rankdir=LR; a -> b }`
- **THEN** `b` is to the right of `a`, at the same height

### Requirement: Order within ranks

The system SHALL order the nodes of each rank to reduce edge crossings, with the median and transposition heuristics of Gansner et al. When no edge decides between two nodes, the system SHALL keep their input order. `remincross` and `searchsize` SHALL be accepted.

#### Scenario: Crossing removed

- **WHEN** the graph is `digraph { a; b; c; d; a -> d; b -> c }`
- **THEN** the drawing has no edge crossing

#### Scenario: Input order of unconnected nodes

- **WHEN** the graph is `digraph { a; b; c }`
- **THEN** `a`, `b` and `c` are placed from left to right in that order

### Requirement: Separation

Within a rank, adjacent nodes SHALL be at least `nodesep` apart (in inches, default 0.25, minimum 0.02). Adjacent ranks SHALL be at least `ranksep` apart (in inches, default 0.5, minimum 0.02), measured between the bottom of the nodes of one rank and the top of the nodes of the next. Nodes SHALL NOT overlap.

#### Scenario: Node separation

- **WHEN** the graph is `digraph { nodesep=1; a; b }`
- **THEN** the gap between the boxes of `a` and `b` is at least 72 points

### Requirement: Positioning

Within the separation constraints, the system SHALL place nodes to keep edges short and straight, as the positioning of Gansner et al. does.

#### Scenario: Straight chain

- **WHEN** the graph is `digraph { a -> b -> c }`
- **THEN** `a`, `b` and `c` have the same horizontal position

### Requirement: Node size and shape

A node SHALL be at least `width` wide and `height` high (in inches, defaults 0.75 and 0.5), and SHALL grow to fit its label with margins. The shapes SHALL be drawn as documented by Graphviz:

- `rect`, `box`, `rectangle`: a rectangle
- `ellipse`, `oval`: an ellipse
- `circle`: a circle, as wide as it is high
- `diamond`, `hexagon`, `octagon`: the regular polygons of the documentation
- `point`: a small filled circle, sized only by `width` and `height`, without its label
- `plaintext`: the label without an outline

With `style=rounded`, a `rect`, `box` or `rectangle` node SHALL have rounded corners whose radius is a third of its smaller side, at most 12 points.

#### Scenario: Label wider than the node

- **WHEN** a node has `label="a fairly long label"` and the default `width`
- **THEN** the node is wider than 0.75 inches and its label fits inside its outline

#### Scenario: Node without a label

- **WHEN** a node has `shape=rect, label="", width=2, height=1`
- **THEN** the node is a rectangle 144 points wide and 72 points high

#### Scenario: Rounded rectangle

- **WHEN** a node has `shape=rect, style=rounded, label="", width=1, height=0.5`
- **THEN** its outline is a 72 by 36 points rectangle whose corners are rounded with a radius of 12 points

#### Scenario: Point

- **WHEN** a node has `shape=point, label="x"`
- **THEN** the node is drawn as a filled circle and the text `x` is not drawn

### Requirement: Text size

The width of text SHALL come from font metric tables bundled with the system, scaled by `fontsize` (default 14), never from fonts installed on the machine. A line SHALL be 1.2 times the font size high. `fontname` (default `Times-Roman`) SHALL select the metrics as the reference does, comparing names without regard to case:

- `Times`, `Times-Roman`, `Times New Roman` and `Liberation Serif` select the Times metrics
- `Helvetica`, `Arial`, `Liberation Sans`, `Arimo`, `Nimbus Sans` and `FreeSans` select the Helvetica metrics
- `Courier`, `Courier New`, `Liberation Mono`, `Cousine`, `Nimbus Mono` and `FreeMono` select the Courier metrics
- any other name, including a name with a weight or a slant such as `Helvetica-Bold`, selects the Times metrics

Every byte of a non-ASCII character SHALL be as wide as a space, as in the reference.

#### Scenario: Monospaced font

- **WHEN** two nodes have `fontname="Courier"` and the labels `iiii` and `WWWW`
- **THEN** both labels have the same width

#### Scenario: Weight in the font name

- **WHEN** a label is set in `fontname="Helvetica-Bold"`
- **THEN** it has the width it has in `fontname="Times-Roman"`

#### Scenario: Non-ASCII character

- **WHEN** a label is `é`, two bytes in UTF-8
- **THEN** it is as wide as two spaces

#### Scenario: Font size

- **WHEN** a label is set in `fontsize=28` instead of `fontsize=14`
- **THEN** the text is twice as wide

### Requirement: Edge routing

With `splines=spline` (the default), edges SHALL be smooth curves made of cubic Bezier segments. With `splines=polyline`, edges SHALL be made of straight segments. Edges SHALL start and end on the outlines of their nodes, SHALL be routed around nodes other than their own, and edges between the same two nodes SHALL be drawn as distinct curves. An edge from a node to itself SHALL be drawn as a loop beside the node. An edge between two nodes of the same rank SHALL be drawn between them.

#### Scenario: Edge avoids a node

- **WHEN** the graph is `digraph { a -> b -> c; a -> c }`
- **THEN** the edge `a -> c` does not pass through `b`

#### Scenario: Two edges between the same nodes

- **WHEN** the graph is `digraph { a -> b; a -> b }`
- **THEN** the two edges are drawn as two different curves

#### Scenario: Self loop

- **WHEN** the graph is `digraph { a -> a }`
- **THEN** the edge is a loop that starts and ends on the outline of `a`

#### Scenario: Polyline

- **WHEN** the graph has `splines=polyline`
- **THEN** every edge is made of straight segments

### Requirement: Arrowheads

`dir` SHALL decide which ends get an arrowhead: `forward` (the default in a directed graph) the head, `back` the tail, `both` both and `none` neither. `arrowhead` and `arrowtail` SHALL select the shape at each end: `normal` (the default) a filled triangle, `empty` an outlined triangle, `none` no arrowhead. `arrowsize` (default 1) SHALL scale the arrowheads. An edge SHALL end where its arrowhead begins.

#### Scenario: Arrow at the tail

- **WHEN** an edge `a -> b` has `dir=back`
- **THEN** the arrowhead is at `a` and there is none at `b`

#### Scenario: Empty arrowhead

- **WHEN** an edge has `arrowhead=empty`
- **THEN** its arrowhead is an outlined triangle

### Requirement: Invisible edges

An edge with `style=invis` SHALL take part in the layout and SHALL NOT be drawn.

#### Scenario: Invisible edge

- **WHEN** the graph is `digraph { a -> b [style=invis] }`
- **THEN** `b` is one rank below `a`, and no edge is drawn

### Requirement: Graph label

A plain-text `label` on the graph SHALL be drawn below the drawing, which SHALL grow to make room for it. With `labeljust=c` (the default) the label SHALL be centered, with `labeljust=l` at the left and with `labeljust=r` at the right of the drawing.

#### Scenario: Graph label

- **WHEN** the graph is `digraph { label="Title"; a }`
- **THEN** `Title` is drawn below `a`, centered horizontally

#### Scenario: Graph label at the left

- **WHEN** the graph is `digraph { label="Title"; labeljust=l; a -> b; a -> c }`
- **THEN** `Title` is drawn below the drawing, at its left

### Requirement: Ranking with clusters

The system SHALL rank a graph with clusters cluster by cluster, as the reference does by default (Graphviz documentation of `newrank`): each cluster is ranked on its own, innermost clusters first, and is then ranked in its parent as one block that keeps the relative ranks of its nodes. Edges between a node inside a cluster and a node outside it SHALL constrain the rank of the whole block, so the head of such an edge can be on a rank above its tail.

#### Scenario: Cluster ranked as a block

- **WHEN** the graph is `digraph { subgraph cluster_a { a; b } a -> c -> d -> b }`
- **THEN** `a` and `b` are on the same rank, `c` is one rank below them and `d` two ranks below them

### Requirement: Rank constraints

`rank` on a subgraph that is not a cluster SHALL constrain the nodes of the subgraph as documented by Graphviz:

- `same`: the nodes are on one rank
- `min`: the nodes are on the minimum rank
- `source`: the nodes are on the minimum rank, and no other node is on it unless it is in a subgraph with `rank=source` or `rank=min`
- `max`: the nodes are on the maximum rank
- `sink`: the nodes are on the maximum rank, and no other node is on it unless it is in a subgraph with `rank=sink` or `rank=max`

Inside a cluster, the minimum and maximum ranks SHALL be those of the cluster, as documented for `newrank`. `rank` on the root graph or on a cluster SHALL NOT constrain their nodes.

#### Scenario: Same rank

- **WHEN** the graph is `digraph { a -> b -> c; { rank=same; a; c } }`
- **THEN** `a` and `c` are on the same rank, and `b` is one rank below them

#### Scenario: Minimum rank

- **WHEN** the graph is `digraph { a -> b -> c; d -> e; { rank=min; d } }`
- **THEN** `d` is on the rank of `a`

#### Scenario: Source

- **WHEN** the graph is `digraph { a -> b -> c; d -> e; { rank=source; d } }`
- **THEN** `d` is alone on the first rank, and `a` and `e` are on the second rank

#### Scenario: Sink

- **WHEN** the graph is `digraph { a -> b -> c; d; { rank=sink; d } }`
- **THEN** `d` is alone on the last rank, one rank below `c`

#### Scenario: Source inside a cluster

- **WHEN** the graph is `digraph { x -> a; subgraph cluster_c { a -> b -> c; d; { rank=source; d } } }`
- **THEN** `d` is alone on the first rank of `cluster_c`, and `a` is one rank below `d`

#### Scenario: Rank on a cluster

- **WHEN** the graph is `digraph { subgraph cluster_x { rank=same; a -> b } }`
- **THEN** `b` is one rank below `a`

### Requirement: Cluster boxes

Each drawn cluster SHALL be drawn as a rectangle, its box, that contains the nodes of the cluster and the routes of the edges between them, with a margin of 8 points around them, the documented default. The box of a nested cluster SHALL be inside the box of its parent, the boxes of sibling clusters SHALL NOT overlap, and nodes that do not belong to a cluster SHALL be outside its box. An edge with no end inside a cluster SHALL be routed outside its box. The box SHALL be drawn in the cluster's `color`, black by default.

#### Scenario: Margin

- **WHEN** the graph is `digraph { subgraph cluster_a { a } }`
- **THEN** the box of `cluster_a` is 70 by 52 points, 8 points larger than `a` on each side

#### Scenario: Nested clusters

- **WHEN** the graph is `digraph { subgraph cluster_a { subgraph cluster_b { a } c } }`
- **THEN** the box of `cluster_b` is inside the box of `cluster_a`, and `c` is outside the box of `cluster_b`

#### Scenario: Sibling clusters

- **WHEN** the graph is `digraph { subgraph cluster_a { a } subgraph cluster_b { b } }`
- **THEN** the boxes of `cluster_a` and `cluster_b` do not overlap

#### Scenario: Edge around a cluster

- **WHEN** the graph is `digraph { a -> d; a -> b; b -> d; subgraph cluster_x { b } }`
- **THEN** the edge `a -> d` does not cross the box of `cluster_x`

#### Scenario: Cluster color

- **WHEN** a cluster has `color="#FF0000"`
- **THEN** its box is drawn in red

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
