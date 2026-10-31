## MODIFIED Requirements

### Requirement: SVG drawing

Node outlines SHALL be `polygon` elements for polygonal shapes, `ellipse` elements for `ellipse`, `circle` and `point`, and `path` elements of cubic Bezier segments for rounded rectangles. Cluster boxes SHALL be `polygon` elements with no fill. Edges SHALL be `path` elements whose `d` attribute is a move followed by cubic Bezier segments (`M x,y C ...`), and arrowheads SHALL be `polygon` elements. Labels SHALL be `text` elements, one per line, with the font family, the font size and the color of the label. The labels of an edge (`label`, `xlabel`, `headlabel`, `taillabel`) SHALL be in the edge's group, after its arrowheads, and the `xlabel` of a node in the node's group, after its label. A table label SHALL be drawn as the reference draws it: a `polygon` over the table filled with its `BGCOLOR` and without a stroke, when it has one; a `polygon` for the border of the table; and a `polygon` for the border of its cell, of the cell's size, at the upper left of the table; the borders in the color of the outline of the object that holds the label. A color given as `#RRGGBB` SHALL be written in lower case in the `stroke` and `fill` attributes, and the default color SHALL be black.

#### Scenario: Node color

- **WHEN** a node has `color="#FF0000"`
- **THEN** its outline element has `stroke="#ff0000"`

#### Scenario: Edge path

- **WHEN** `digraph { a -> b }` is rendered as `svg`
- **THEN** the edge group has a `path` whose `d` starts with `M` and continues with `C`, and a `polygon` for the arrowhead

#### Scenario: Cluster box

- **WHEN** a cluster with `color="#FF0000"` and `label="A"` is rendered as `svg`
- **THEN** its group has a `polygon` with `fill="none"` and `stroke="#ff0000"`, and a `text` element `A`

#### Scenario: Edge label

- **WHEN** `digraph { a -> b [label="x"] }` is rendered as `svg`
- **THEN** the edge group has a `text` element `x` after the arrowhead `polygon`

#### Scenario: Table label that gopuml reads

- **WHEN** `digraph { a -> b [label=<<TABLE BGCOLOR="#000009" FIXEDSIZE="TRUE" WIDTH="57" HEIGHT="17"><TR><TD></TD></TR></TABLE>>] }` is rendered as `svg`
- **THEN** the edge group has a `polygon` with `fill="#000009"` and `stroke="none"` whose points are the corners of a 57 by 17 points box, then a `polygon` for the border of the table and one for the border of its cell

### Requirement: Plain output

`plain` output SHALL follow the format documented by Graphviz: a `graph` line with the scale and the size, one `node` line per node with its center, size, label, style, shape and colors, one `edge` line per edge with its control points, its label and the label's center when it has a `label`, its style and its color, and a final `stop` line. Sizes and coordinates SHALL be in inches, with `y` growing upward. HTML labels SHALL be written as the reference writes them.

#### Scenario: Plain lines

- **WHEN** `digraph { a -> b }` is rendered as `plain`
- **THEN** the output is a `graph` line, two `node` lines, one `edge` line from `a` to `b` and a `stop` line

#### Scenario: Edge label in plain

- **WHEN** `digraph { a -> b [label="x"] }` is rendered as `plain`
- **THEN** the `edge` line has, after the control points, `x` and the center of the label, then the style and the color
