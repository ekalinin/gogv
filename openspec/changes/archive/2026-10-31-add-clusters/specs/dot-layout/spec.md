## ADDED Requirements

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

A plain-text `label` of a cluster SHALL be drawn inside its box, at the top, set in the cluster's `fontname`, `fontsize` and `fontcolor`. With `labeljust=c` (the default) the label SHALL be centered, with `labeljust=l` at the left and with `labeljust=r` at the right of the box. The box SHALL grow to hold its label: it SHALL be at least as wide as the label, and its nodes SHALL be below the label. In `rankdir=LR`, the label SHALL also be at the top of the box.

#### Scenario: Label at the top

- **WHEN** the graph is `digraph { subgraph cluster_a { label="Group"; a } }`
- **THEN** `Group` is drawn inside the box of `cluster_a`, centered, above `a`

#### Scenario: Label at the left

- **WHEN** a cluster has `label="T"` and `labeljust=l`
- **THEN** `T` is drawn at the left of the top of the box

#### Scenario: Label wider than the nodes

- **WHEN** the graph is `digraph { subgraph cluster_a { label="Long cluster title"; a } }`
- **THEN** the box of `cluster_a` is wider than the label

## MODIFIED Requirements

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

### Requirement: Graph label

A plain-text `label` on the graph SHALL be drawn below the drawing, which SHALL grow to make room for it. With `labeljust=c` (the default) the label SHALL be centered, with `labeljust=l` at the left and with `labeljust=r` at the right of the drawing.

#### Scenario: Graph label

- **WHEN** the graph is `digraph { label="Title"; a }`
- **THEN** `Title` is drawn below `a`, centered horizontally

#### Scenario: Graph label at the left

- **WHEN** the graph is `digraph { label="Title"; labeljust=l; a -> b; a -> c }`
- **THEN** `Title` is drawn below the drawing, at its left
