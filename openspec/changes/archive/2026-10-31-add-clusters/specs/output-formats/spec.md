## MODIFIED Requirements

### Requirement: SVG document

`svg` output SHALL be an SVG document with the structure that the reference writes:

- a root `svg` element whose `width` and `height` are whole numbers of points (`width="62pt"`) and whose `viewBox` covers the drawing
- one group for the graph, with the graph name in its `title` and a transform that puts the origin at the bottom left of the drawing plus the pad, so that `y` coordinates are negative
- one group per drawn cluster, of class `cluster`, with the cluster name in its `title`, written before the node groups, a cluster before the clusters nested in it
- one group per drawn node, of class `node`, with the node name in its `title`
- one group per drawn edge, of class `edge`, with the edge name (`tail->head`) in its `title`
- coordinates written with at most two decimals
- text escaped for XML as the reference escapes it

#### Scenario: Groups that gopuml reads

- **WHEN** `digraph { a -> b }` is rendered as `svg`
- **THEN** the document has a node group titled `a`, a node group titled `b` and an edge group titled `a->b` (escaped as the reference escapes it), and its `svg` element has `width` and `height` in whole points

#### Scenario: Cluster groups

- **WHEN** `digraph { subgraph cluster_a { subgraph cluster_b { a } } }` is rendered as `svg`
- **THEN** the document has a cluster group titled `cluster_a`, then a cluster group titled `cluster_b`, both before the node group titled `a`

### Requirement: SVG drawing

Node outlines SHALL be `polygon` elements for polygonal shapes, `ellipse` elements for `ellipse`, `circle` and `point`, and `path` elements of cubic Bezier segments for rounded rectangles. Cluster boxes SHALL be `polygon` elements with no fill. Edges SHALL be `path` elements whose `d` attribute is a move followed by cubic Bezier segments (`M x,y C ...`), and arrowheads SHALL be `polygon` elements. Labels SHALL be `text` elements, one per line, with the font family, the font size and the color of the label. A color given as `#RRGGBB` SHALL be written in lower case in the `stroke` and `fill` attributes, and the default color SHALL be black.

#### Scenario: Node color

- **WHEN** a node has `color="#FF0000"`
- **THEN** its outline element has `stroke="#ff0000"`

#### Scenario: Edge path

- **WHEN** `digraph { a -> b }` is rendered as `svg`
- **THEN** the edge group has a `path` whose `d` starts with `M` and continues with `C`, and a `polygon` for the arrowhead

#### Scenario: Cluster box

- **WHEN** a cluster with `color="#FF0000"` and `label="A"` is rendered as `svg`
- **THEN** its group has a `polygon` with `fill="none"` and `stroke="#ff0000"`, and a `text` element `A`
