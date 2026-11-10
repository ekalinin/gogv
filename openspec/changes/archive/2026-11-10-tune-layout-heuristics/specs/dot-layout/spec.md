## MODIFIED Requirements

### Requirement: Order within ranks

The system SHALL order the nodes of each rank to reduce edge crossings, with the median and transposition heuristics of Gansner et al. When no edge decides between two nodes, the system SHALL keep their input order. As the reference does, the initial order SHALL take the sources of each component in the order of a depth-first search of it from its first node, along the edges down before those up, and the sort by medians SHALL swap the nodes with equal medians on the passes down, so that between orders with the same number of crossings the system chooses the reference's order in the scenarios below. `remincross` and `searchsize` SHALL be accepted.

#### Scenario: Crossing removed

- **WHEN** the graph is `digraph { a; b; c; d; a -> d; b -> c }`
- **THEN** the drawing has no edge crossing

#### Scenario: Input order of unconnected nodes

- **WHEN** the graph is `digraph { a; b; c }`
- **THEN** `a`, `b` and `c` are placed from left to right in that order

#### Scenario: Orders with the same crossings

- **WHEN** the graph is `digraph { a -> b [label="LLLLLL"]; a -> c; d -> b }`
- **THEN** `a` lies left of `d`, and `c` left of `b`, as the reference orders them

#### Scenario: Sources in the order of a search

- **WHEN** the graph is `digraph { f -> g; c -> f; b -> g }`
- **THEN** `b` lies left of `f`, as the reference orders them

### Requirement: Positioning

Within the separation constraints, the system SHALL place nodes to keep edges short and straight, each edge weighed by its `weight` (default 1) times the factor of Gansner et al. for the virtual nodes it joins, as the positioning of Gansner et al. does. Among positions of equal cost, the system SHALL choose as the reference does in the scenarios below: it SHALL start from the initial feasible tree of section 4.3 of Gansner et al., and balance the parts of the tree whose cut value is zero in the order of the tree's edges.

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

### Requirement: Edge routing

With `splines=spline` (the default), edges SHALL be smooth curves made of cubic Bezier segments. With `splines=polyline`, edges SHALL be made of straight segments. Edges SHALL start and end on the outlines of their nodes, SHALL be routed around nodes other than their own, and edges between the same two nodes SHALL be drawn as distinct curves. An edge from a node to itself SHALL be drawn as a loop beside the node. An edge between two nodes of the same rank SHALL be drawn between them; when they are not neighbors, it SHALL go over the nodes between them, as high as the reference's curve goes, and the drawing's box SHALL include it. As section 5.1.1 of Gansner et al. describes, the shorter edges SHALL be routed first, and an edge routed SHALL keep those routed after it clear of its curve rather than of its virtual nodes.

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

#### Scenario: Label clear of an edge routed before

- **WHEN** the graph is `digraph { l -> m; x -> r; x -> l [label="LLLLLL"] }`
- **THEN** the center of the label lies 28 points right of the center of `l`, within 2 points, as the reference puts it

#### Scenario: Flat edge over a node

- **WHEN** the graph is `digraph { a -> c [minlen=0]; b -> c [minlen=0] }`
- **THEN** the curve of `a -> c` goes over `b`, and the drawing is 60 points high, within 2%, where its nodes are 36 points high

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
