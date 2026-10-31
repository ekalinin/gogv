# Spec Delta

## Purpose

Computes the positions of nodes and the routes of edges as the Graphviz `dot` layout engine does, for graphs without clusters, edge labels and HTML-like labels, with text measured from bundled font metrics.

## ADDED Requirements

### Requirement: Ranking

The system SHALL place every node on a rank so that, for every edge with `constraint=true` (the default), the rank of its head minus the rank of its tail is at least the edge's `minlen` (default 1). Among such placements, it SHALL keep edges short, as the network simplex ranking of Gansner et al. does. Edges with `constraint=false` SHALL NOT affect ranks. A graph with cycles SHALL be laid out: edges reversed for ranking SHALL still be drawn from their tail to their head.

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

#### Scenario: Label wider than the node

- **WHEN** a node has `label="a fairly long label"` and the default `width`
- **THEN** the node is wider than 0.75 inches and its label fits inside its outline

#### Scenario: Node without a label

- **WHEN** a node has `shape=rect, label="", width=2, height=1`
- **THEN** the node is a rectangle 144 points wide and 72 points high

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

A plain-text `label` on the graph SHALL be drawn centered below the drawing, which SHALL grow to make room for it.

#### Scenario: Graph label

- **WHEN** the graph is `digraph { label="Title"; a }`
- **THEN** `Title` is drawn below `a`, centered horizontally
