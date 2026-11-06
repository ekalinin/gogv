## MODIFIED Requirements

### Requirement: Ranking

The system SHALL place every node on a rank so that, for every edge with `constraint=true` (the default), the rank of its head minus the rank of its tail is at least the edge's `minlen` (default 1), except where the requirement "Ranking with clusters" allows otherwise. Among such placements, it SHALL keep edges short, each edge weighed by its `weight` (default 1), as the network simplex ranking of Gansner et al. does; an edge of `weight` 0 SHALL NOT pull its ends together. Edges with `constraint=false` SHALL NOT affect ranks. A graph with cycles SHALL be laid out: edges reversed for ranking SHALL still be drawn from their tail to their head.

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

#### Scenario: Heavy edge kept short

- **WHEN** the graph is `digraph { a -> b -> c -> d; a -> e [weight=5]; e -> d }`
- **THEN** `e` is on the rank of `b`, right below `a`, while without the weight it is on the rank of `c`

### Requirement: Positioning

Within the separation constraints, the system SHALL place nodes to keep edges short and straight, each edge weighed by its `weight` (default 1) times the factor of Gansner et al. for the virtual nodes it joins, as the positioning of Gansner et al. does.

#### Scenario: Straight chain

- **WHEN** the graph is `digraph { a -> b -> c }`
- **THEN** `a`, `b` and `c` have the same horizontal position

#### Scenario: Heavy edge kept straight

- **WHEN** the graph is `digraph { a -> b; a -> c [weight=10] }`
- **THEN** `a` has the horizontal position of `c`, and `b` lies left of them

#### Scenario: Heavy edge across ranks

- **WHEN** the graph is `digraph { a -> b -> c; a -> c [weight=5] }`
- **THEN** `a` and `c` have the same horizontal position, right of `b`

### Requirement: Node size and shape

A node SHALL be at least `width` wide and `height` high (in inches, defaults 0.75 and 0.5), and SHALL grow to fit its label with margins. The shapes SHALL be drawn as documented by Graphviz:

- `rect`, `box`, `rectangle`: a rectangle
- `box3d`: a rectangle with a depth of 4 points drawn inside it, at its upper right, sized as a rectangle
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

#### Scenario: Box with depth

- **WHEN** a node has `shape=box3d, label="", width=2, height=1`
- **THEN** the node is 144 points wide and 72 points high, and its front face is 140 by 68 points, at its lower left

### Requirement: Arrowheads

`dir` SHALL decide which ends get an arrowhead: `forward` (the default in a directed graph) the head, `back` the tail, `both` both and `none` neither. `arrowhead` and `arrowtail` SHALL select the shape at each end: `normal` (the default) a filled triangle, `empty` an outlined triangle, `none` no arrowhead. `arrowsize` (default 1) SHALL scale the arrowheads. An edge SHALL end where its arrowhead begins. With a `penwidth` wider than 1 point, the arrowhead SHALL stop short of its node so that its stroked tip touches the node, and an arrowhead of a `penwidth` wider than 4 points SHALL be wider by the ratio of the `penwidth` to 4, as the reference draws them.

#### Scenario: Arrow at the tail

- **WHEN** an edge `a -> b` has `dir=back`
- **THEN** the arrowhead is at `a` and there is none at `b`

#### Scenario: Empty arrowhead

- **WHEN** an edge has `arrowhead=empty`
- **THEN** its arrowhead is an outlined triangle

#### Scenario: Wider pen

- **WHEN** an edge `a -> b` has `penwidth=3`
- **THEN** the tip of its arrowhead is 3 points further from `b` than with `penwidth=1`, and the arrowhead keeps its size
