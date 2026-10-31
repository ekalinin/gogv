# output-formats Specification

## Purpose

Writes laid-out graphs in the reference `dot`'s `svg`, `plain` and `canon` formats, with the same structure that programs reading `dot` output (gopuml first) rely on, and identically on every platform.

## Requirements

### Requirement: SVG document

`svg` output SHALL be an SVG document with the structure that the reference writes:

- a root `svg` element whose `width` and `height` are whole numbers of points (`width="62pt"`) and whose `viewBox` covers the drawing
- one group for the graph, with the graph name in its `title` and a transform that puts the origin at the bottom left of the drawing plus the pad, so that `y` coordinates are negative
- one group per drawn node, of class `node`, with the node name in its `title`
- one group per drawn edge, of class `edge`, with the edge name (`tail->head`) in its `title`
- coordinates written with at most two decimals
- text escaped for XML as the reference escapes it

#### Scenario: Groups that gopuml reads

- **WHEN** `digraph { a -> b }` is rendered as `svg`
- **THEN** the document has a node group titled `a`, a node group titled `b` and an edge group titled `a->b` (escaped as the reference escapes it), and its `svg` element has `width` and `height` in whole points

### Requirement: SVG drawing

Node outlines SHALL be `polygon` elements for polygonal shapes, `ellipse` elements for `ellipse`, `circle` and `point`, and `path` elements of cubic Bezier segments for rounded rectangles. Edges SHALL be `path` elements whose `d` attribute is a move followed by cubic Bezier segments (`M x,y C ...`), and arrowheads SHALL be `polygon` elements. Labels SHALL be `text` elements, one per line, with the font family, the font size and the color of the label. A color given as `#RRGGBB` SHALL be written in lower case in the `stroke` and `fill` attributes, and the default color SHALL be black.

#### Scenario: Node color

- **WHEN** a node has `color="#FF0000"`
- **THEN** its outline element has `stroke="#ff0000"`

#### Scenario: Edge path

- **WHEN** `digraph { a -> b }` is rendered as `svg`
- **THEN** the edge group has a `path` whose `d` starts with `M` and continues with `C`, and a `polygon` for the arrowhead

### Requirement: Plain output

`plain` output SHALL follow the format documented by Graphviz: a `graph` line with the scale and the size, one `node` line per node with its center, size, label, style, shape and colors, one `edge` line per edge with its control points, style and color, and a final `stop` line. Sizes and coordinates SHALL be in inches, with `y` growing upward.

#### Scenario: Plain lines

- **WHEN** `digraph { a -> b }` is rendered as `plain`
- **THEN** the output is a `graph` line, two `node` lines, one `edge` line from `a` to `b` and a `stop` line

### Requirement: Canonical output

`canon` output SHALL be the graph written back in DOT without layout, identical byte for byte to the reference's `canon` output for the same input. It SHALL keep every attribute of the input, including those that the layout does not support.

#### Scenario: Unknown attribute in canon

- **WHEN** `digraph { a -> b [foo=bar] }` is rendered as `canon`
- **THEN** the output is identical to the reference's and contains `foo=bar`

### Requirement: Several graphs

For an input with several graphs, the output SHALL be the output of each graph, in input order, written one after another.

#### Scenario: Two graphs as SVG

- **WHEN** `digraph A { a }` on one line and `digraph B { b }` on the next are rendered as `svg`
- **THEN** the output is the SVG document of `A` followed by the SVG document of `B`

### Requirement: Deterministic output

For the same input and format, the output SHALL be byte-identical across runs, machines, operating systems and processor architectures. It SHALL NOT depend on installed fonts, the time, the locale or the environment.

#### Scenario: Two platforms

- **WHEN** the same DOT input is rendered as `svg` on linux/amd64 and on darwin/arm64
- **THEN** the two outputs are byte-identical
