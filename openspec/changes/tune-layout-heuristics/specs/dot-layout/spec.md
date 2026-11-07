## MODIFIED Requirements

### Requirement: Order within ranks

The system SHALL order the nodes of each rank to reduce edge crossings, with the median and transposition heuristics of Gansner et al. When no edge decides between two nodes, the system SHALL keep their input order. Between orders with the same number of crossings, the system SHALL choose the reference's order, and SHALL NOT lay out its mirror. `remincross` and `searchsize` SHALL be accepted.

#### Scenario: Crossing removed

- **WHEN** the graph is `digraph { a; b; c; d; a -> d; b -> c }`
- **THEN** the drawing has no edge crossing

#### Scenario: Input order of unconnected nodes

- **WHEN** the graph is `digraph { a; b; c }`
- **THEN** `a`, `b` and `c` are placed from left to right in that order

#### Scenario: Orders with the same crossings

- **WHEN** the graph is `digraph { a -> b [label="LLLLLL"]; a -> c; d -> b }`
- **THEN** `a` lies left of `d`, and `c` left of `b`, as the reference orders them

### Requirement: Positioning

Within the separation constraints, the system SHALL place nodes to keep edges short and straight, each edge weighed by its `weight` (default 1) times the factor of Gansner et al. for the virtual nodes it joins, as the positioning of Gansner et al. does. Among positions of equal cost, a node, or a part of the drawing, that the edges pull equally to the left and to the right SHALL lie halfway across the range in which its cost does not change, within a point, as the reference places it.

#### Scenario: Straight chain

- **WHEN** the graph is `digraph { a -> b -> c }`
- **THEN** `a`, `b` and `c` have the same horizontal position

#### Scenario: Heavy edge kept straight

- **WHEN** the graph is `digraph { a -> b; a -> c [weight=10] }`
- **THEN** `a` has the horizontal position of `c`, and `b` lies left of them

#### Scenario: Heavy edge across ranks

- **WHEN** the graph is `digraph { a -> b -> c; a -> c [weight=5] }`
- **THEN** `a` and `c` have the same horizontal position, right of `b`

#### Scenario: Head between two tails

- **WHEN** the graph is `digraph { a -> x; c -> y; b -> x; d -> y }`
- **THEN** `x` lies halfway between `a` and `b`, and `y` halfway between `c` and `d`

#### Scenario: Tail between a node and a labeled edge

- **WHEN** the graph is `digraph { a -> b; c -> d; c -> e [minlen=2, label="L"] }`
- **THEN** `c` lies halfway between `d` and `e`, within a point

### Requirement: Edge routing

With `splines=spline` (the default), edges SHALL be smooth curves made of cubic Bezier segments. With `splines=polyline`, edges SHALL be made of straight segments. Edges SHALL start and end on the outlines of their nodes, SHALL be routed around nodes other than their own, and edges between the same two nodes SHALL be drawn as distinct curves. An edge from a node to itself SHALL be drawn as a loop beside the node. An edge between two nodes of the same rank SHALL be drawn between them; when they are not neighbors, it SHALL go over the nodes between them, as high as the reference's curve goes, and the drawing's box SHALL include it.

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

#### Scenario: Flat edge over a node

- **WHEN** the graph is `digraph { a -> c [minlen=0]; b -> c [minlen=0] }`
- **THEN** the curve of `a -> c` goes over `b`, and the drawing is 60 points high, within 2%, where its nodes are 36 points high

### Requirement: Edge labels

The `label` of an edge between two ranks SHALL be drawn beside the edge, halfway between its ends, with its left side on the edge, and SHALL take room in the layout, as the reference's off-center virtual nodes do (Gansner et al., section 5.3), so that it overlaps no node, edge or other label. As the reference does, a graph with at least one edge label SHALL have its ranks laid out with a rank between every two of its ranks, half `ranksep` apart from them, and each edge label SHALL lie on the rank halfway along its edge. The label SHALL be set in the edge's `fontname`, `fontsize` and `fontcolor`. In `rankdir=LR`, the label SHALL lie above the edge. Two edges with labels between the same nodes SHALL be drawn as two curves, each with its own label. An edge label SHALL take the room that the reference gives it also when its edge shares its nodes with another edge, lies in a cycle, or is in one of several components.

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

#### Scenario: Labeled and unlabeled edges between the same nodes

- **WHEN** the graph is `digraph { a -> b [label="LLLLLLLLLLLL"]; a -> b }`
- **THEN** the drawing is 108 points wide, within 2%, as the reference's

#### Scenario: Labeled edge in a cycle

- **WHEN** the graph is `digraph { a -> b [label="LLLLLLLLLLLL"]; b -> c; c -> a }`
- **THEN** the drawing is 112 points wide, and the centers of the nodes and of the label are within 2% of its size of the reference's

### Requirement: Labels of flat edges and loops

The label of an edge between two nodes of the same rank SHALL be drawn above the edge, between its ends, which SHALL be far enough apart for the label to fit. When the edge has a port at an end, its label SHALL lie where the reference puts it, which can be higher above the edge or below it. The label of an edge from a node to itself SHALL be drawn beside its loop, beyond the loop's far side.

#### Scenario: Flat edge

- **WHEN** the graph is `digraph { a -> b [minlen=0, label="flat"] }`
- **THEN** `flat` is drawn above the edge, between `a` and `b`, and `a` and `b` are farther apart than in `digraph { a -> b [minlen=0] }`

#### Scenario: Loop

- **WHEN** the graph is `digraph { a -> a [label="loop"] }`
- **THEN** `loop` is drawn to the right of the loop

#### Scenario: Flat edge from a port

- **WHEN** the graph is `digraph { a [shape=plaintext, label=<<TABLE><TR><TD PORT="p" WIDTH="40" HEIGHT="30"></TD></TR></TABLE>>]; a:p -> b [minlen=0, label="L"] }`
- **THEN** the label is 8.5 points higher than the label of the same edge without the port, within a point, and the drawing is higher than without the port
