# output-formats Specification

## Purpose

Writes laid-out graphs in the reference `dot`'s formats of v1 (`svg`, `svg_inline`, `plain`, `plain-ext`, `canon`, `dot`, `gv`, `xdot`, `xdot1.2`, `xdot1.4`, `json`, `json0`, `dot_json` and `xdot_json`), with the same structure that programs reading `dot` output (gopuml first) rely on, and identically on every platform.

## Requirements

### Requirement: SVG document

`svg` output SHALL be an SVG document with the structure that the reference writes:

- a root `svg` element whose `width` and `height` are whole numbers of points (`width="62pt"`) and whose `viewBox` covers the drawing
- one group for the graph, with the graph name in its `title` and a transform that puts the origin at the bottom left of the drawing plus the pad, so that `y` coordinates are negative
- one group per drawn cluster, of class `cluster`, with the cluster name in its `title`, written before the node groups, a cluster before the clusters nested in it
- one group per drawn node, of class `node`, with the node name in its `title`, whose `id` is the node's `id` attribute, or else `node` followed by the node's place among the drawn nodes
- one group per drawn edge, of class `edge`, with the edge name (`tail->head`) in its `title`, each end followed by its port as it is written (`tail:port->head`), whose `id` is `edge` followed by the edge's place among the edges in the order they are read, as the reference numbers them
- the node and edge groups in the reference's order: the nodes in node order, each followed by its out-edges in the order of their heads, an edge preceded by its head when the head is not written yet; an invisible edge leaves only its comment, and keeps its place in the numbering
- for a node with a `tooltip`, its outline and its label, but not its `xlabel`, within a group whose `id` is `a_` followed by the node group's `id`, around an `a` element whose `xlink:title` is the tooltip; for an edge with a `tooltip`, its curve and arrowheads likewise, and for an edge with a `labeltooltip`, its `label` within a group whose `id` is `a_` followed by the edge group's `id` and `-label`
- coordinates written with at most two decimals
- text escaped for XML as the reference escapes it

#### Scenario: Groups that gopuml reads

- **WHEN** `digraph { a -> b }` is rendered as `svg`
- **THEN** the document has a node group titled `a`, a node group titled `b` and an edge group titled `a->b` (escaped as the reference escapes it), and its `svg` element has `width` and `height` in whole points

#### Scenario: Cluster groups

- **WHEN** `digraph { subgraph cluster_a { subgraph cluster_b { a } } }` is rendered as `svg`
- **THEN** the document has a cluster group titled `cluster_a`, then a cluster group titled `cluster_b`, both before the node group titled `a`

#### Scenario: Edge title with a port

- **WHEN** `digraph { a [shape=rect]; a:P -> b }` is rendered as `svg`
- **THEN** the edge group is titled `a:P->b`, escaped as the reference escapes it

#### Scenario: Tooltips and ids that pprof uses

- **WHEN** a node `N1_0` has `id="N1_0", tooltip="10ms"` and an edge `N1 -> N1_0` has `tooltip="10ms", labeltooltip="10ms", label=" 10ms"`, rendered as `svg`
- **THEN** the node group has `id="N1_0"` and holds a group `a_N1_0` whose `a` element has `xlink:title="10ms"` around the outline and the label; the edge group holds a group `a_edgeN` around its curve and arrowhead and a group `a_edgeN-label` around its label, both with `xlink:title="10ms"`

#### Scenario: Edge groups in edge order

- **WHEN** `digraph { a -> c; b -> a; a -> b }` is rendered as `svg`
- **THEN** the edge groups of `a->c`, `b->a` and `a->b` have the ids `edge1`, `edge2` and `edge3`, and the groups come in the order `a`, `c`, `a->c`, `b`, `a->b`, `b->a`

### Requirement: SVG drawing

Node outlines SHALL be `polygon` elements for polygonal shapes, `ellipse` elements for `ellipse`, `circle` and `point`, and `path` elements of cubic Bezier segments for rounded rectangles. A node with `style=filled` SHALL fill its outline with its `fillcolor`, or else its `color`, or else `lightgrey`, and a filled `plaintext` node SHALL be a filled `polygon` without a stroke; a `box3d` node SHALL be a `polygon` and three `polyline` elements for its depth. Cluster boxes SHALL be `polygon` elements with no fill. Edges SHALL be `path` elements whose `d` attribute is a move followed by cubic Bezier segments (`M x,y C ...`), and arrowheads SHALL be `polygon` elements; an edge with `style=dotted` SHALL have `stroke-dasharray="1,5"` on its curve and not on its arrowheads, and an edge with a `penwidth` other than 1 SHALL have it as the `stroke-width` of its curve and of its arrowheads. Labels SHALL be `text` elements, one per line, with the font family, the font size and the color of the label. The labels of an edge (`label`, `xlabel`, `headlabel`, `taillabel`) SHALL be in the edge's group, after its arrowheads, and the `xlabel` of a node in the node's group, after its label. A table label SHALL be drawn as the reference draws it: a `polygon` over the table filled with its `BGCOLOR` and without a stroke, when it has one, and a `polygon` for the border of the table, when its border is not 0; then, for each cell in order, a filled `polygon` for its `BGCOLOR` and a `polygon` for its border, likewise. Borders SHALL be in the `COLOR` of the cell or the table, or else in the color of the outline of the object that holds the label; a border wider than 1 point SHALL have a `stroke-width` and be drawn half its width inside its box. A color given as `#RRGGBB` SHALL be written in lower case in the `stroke` and `fill` attributes, and the default color SHALL be black.

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

#### Scenario: Cells and their colors

- **WHEN** a table with `BGCOLOR="#000009" COLOR="#00000F"` has a cell with `BGCOLOR="#000001" COLOR="#00000E"` and a cell with `BGCOLOR="#000002"`
- **THEN** the `svg` has, in this order, the table's fill and its border in `#00000f`, the first cell's fill and its border in `#00000e`, and the second cell's fill and its border in `#00000f`

#### Scenario: Grid without borders

- **WHEN** a table with `BORDER="0" CELLBORDER="0"` has one cell with `BGCOLOR="#000006"` among empty cells
- **THEN** the only `polygon` of the table is that cell's fill, without a stroke

#### Scenario: Filled nodes

- **WHEN** nodes have `style=filled`, `style=filled, color="#FF0000"`, `fillcolor="#00FF00"` without `style=filled`, and `shape=box, style=filled, fillcolor="#0000FF"`, rendered as `svg`
- **THEN** their outlines have `fill="lightgrey"`, `fill="#ff0000"`, `fill="none"` and `fill="#0000ff"`

#### Scenario: Box with depth in svg

- **WHEN** a node has `shape=box3d` and is rendered as `svg`
- **THEN** its group has a `polygon` of 6 corners and three `polyline` elements

#### Scenario: Dotted and wide edges

- **WHEN** an edge has `style=dotted` and another one `penwidth=3`, rendered as `svg`
- **THEN** the first one's `path` has `stroke-dasharray="1,5"` and its arrowhead has none, and the second one's `path` and arrowhead have `stroke-width="3"`

### Requirement: Plain output

`plain` output SHALL follow the format documented by Graphviz: a `graph` line with the scale and the size, one `node` line per node with its center, size, label, style, shape and colors, one `edge` line per edge with its control points, its label and the label's center when it has a `label`, its style and its color, and a final `stop` line. Sizes and coordinates SHALL be in inches, with `y` growing upward. HTML labels SHALL be written as the reference writes them. A node's style SHALL be written as the reference writes it (`filled`), and its fill color SHALL be its `fillcolor` as written, or else its `color`, or else `lightgrey`; an edge's style SHALL be written likewise (`dotted`).

#### Scenario: Plain lines

- **WHEN** `digraph { a -> b }` is rendered as `plain`
- **THEN** the output is a `graph` line, two `node` lines, one `edge` line from `a` to `b` and a `stop` line

#### Scenario: Edge label in plain

- **WHEN** `digraph { a -> b [label="x"] }` is rendered as `plain`
- **THEN** the `edge` line has, after the control points, `x` and the center of the label, then the style and the color

#### Scenario: Styles in plain

- **WHEN** `digraph { a [style=filled, fillcolor="#00FF00"]; a -> b [style=dotted] }` is rendered as `plain`
- **THEN** the line of `a` ends with `filled ellipse black #00FF00`, and the edge line ends with `dotted black`

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

### Requirement: Attributed DOT output

`dot` output SHALL be the `canon` output of the graph to which the layout adds its attributes, as the reference adds them:

- the graph and each drawn cluster: `bb`, the box of the drawing or of the cluster as `"llx,lly,urx,ury"`; with a label, `lp`, the label's center, and `lwidth` and `lheight`, its size in inches
- each node: `pos`, its center as `"x,y"`, `width` and `height`, its size in inches, and `xlp`, the center of its `xlabel`
- each edge, invisible ones included: `pos`, the control points of its curve, after `e,x,y`, the tip of the arrowhead at its head, and `s,x,y`, the tip of the one at its tail, when it has them; `lp`, `xlp`, `head_lp` and `tail_lp`, the centers of its `label`, `xlabel`, `headlabel` and `taillabel`

Coordinates SHALL be in points, with `y` growing upward from the bottom left corner of the drawing, and numbers SHALL be written as the reference writes them, with at most five significant digits (`88.651`). An object that gets a layout attribute SHALL be written as `canon` writes an object with attributes; unlike in `canon`, an anonymous subgraph of the graph or of a cluster SHALL be written even when it has nothing of its own, and a subgraph SHALL have the graph attributes of the layout only when they are its own, as the reference writes them. `gv` output SHALL be the same as `dot` output.

#### Scenario: Layout attributes

- **WHEN** `digraph G { a -> b [label="x"]; a [shape=rect, color="#FF0000"] }` is rendered as `dot`
- **THEN** the output is the reference's `canon` output with `bb` on the graph, `height`, `pos` and `width` on `a` and on `b`, which gets a statement of its own, and `lp` and `pos` on the edge, whose `pos` starts with `e,`

#### Scenario: Clusters and external labels

- **WHEN** a graph with `label="Title"` has a cluster with `label="C"` and an edge with `headlabel="H"` is rendered as `dot`
- **THEN** the graph has `bb`, `lheight`, `lp` and `lwidth`, the cluster has `bb`, `lheight`, `lp` and `lwidth`, and the edge has `head_lp`

#### Scenario: Same output as gv

- **WHEN** a graph is rendered as `dot` and as `gv`
- **THEN** the two outputs are byte-identical

### Requirement: xdot output

`xdot`, `xdot1.2` and `xdot1.4` output SHALL be the `dot` output with the graph attribute `xdotversion` set to `1.7`, `1.2` and `1.4`, and with the drawing of each object as the drawing operations of the xdot format of the Graphviz documentation, in the attributes that the reference uses:

- `_draw_`: the background of the graph, the box of a cluster, the outline of a node and the curve of an edge
- `_ldraw_`: the label of the graph or of a cluster; the label and the `xlabel` of a node; the `label` and the `xlabel` of an edge
- `_hdraw_` and `_tdraw_`: the arrowheads at the head and at the tail of an edge; `_hldraw_` and `_tldraw_`: its `headlabel` and `taillabel`

The operations SHALL be those that draw what `svg` draws, in the same order: pen and fill colors (`c`, `C`), styles (`S`), polygons (`p`, `P`), ellipses (`e`, `E`), Bezier curves (`B`), fonts (`F`) and texts (`T`), with colors in lower case, numbers with at most two decimals and the backslashes of texts doubled, as the reference writes them; a border wider than 1 point SHALL set its pen width (`S 15 -setlinewidth(2)`) when it changes, and an object that sets `style=solid` SHALL start its outline with `S 5 -solid`. A filled outline SHALL set its fill color (`C`) and use the filled operation (`P`, `E`), with `lightgrey` written `#d3d3d3`; the depth of a `box3d` node SHALL be three polylines (`L`); an edge with `style=dotted` SHALL start its curve with `S 6 -dotted`, and an edge with a `penwidth` other than 1 SHALL start its curve and its arrowheads with `S 15 -setlinewidth(N)`. An invisible object SHALL have no drawing operations. For the formats of v1, the three versions SHALL differ only in `xdotversion`.

#### Scenario: Node in xdot

- **WHEN** `digraph { a }` is rendered as `xdot`
- **THEN** the output is byte-identical to the reference's: the graph has `_draw_="c 9 -#fffffe00 C 7 -#ffffff P 4 0 0 0 36 54 36 54 0 "`, `bb` and `xdotversion=1.7`, and `a` has `_draw_="c 7 -#000000 e 27 18 27 18 "` and `_ldraw_="F 14 11 -Times-Roman c 7 -#000000 T 27 13.8 0 6.21 1 -a "`

#### Scenario: Edge in xdot

- **WHEN** `digraph { a -> b [label="x"] }` is rendered as `xdot`
- **THEN** the edge has a `_draw_` with one `B` curve, a `_hdraw_` with a filled `P` triangle, and an `_ldraw_` with the text `x`

#### Scenario: Table label in xdot

- **WHEN** a node has `shape=plaintext, label=<<TABLE><TR><TD PORT="p" BGCOLOR="#000001"></TD><TD WIDTH="20"></TD></TR></TABLE>>` and is rendered as `xdot`
- **THEN** its `_ldraw_` has, in the order of the `svg` drawing, a `p` polygon for the border of the table, a `P` polygon filled with `#000001` and a `p` polygon for the first cell, and a `p` polygon for the second cell

#### Scenario: Versions

- **WHEN** a graph is rendered as `xdot1.2`
- **THEN** the output is its `xdot` output with `xdotversion=1.2`

#### Scenario: Filled node in xdot

- **WHEN** `digraph { a [style=filled] }` is rendered as `xdot`
- **THEN** `a` has `_draw_="c 7 -#000000 C 7 -#d3d3d3 E 27 18 27 18 "`

#### Scenario: Dotted and wide edges in xdot

- **WHEN** an edge has `style=dotted` and another one `penwidth=3`, rendered as `xdot`
- **THEN** the first one's `_draw_` starts with `S 6 -dotted` and its `_hdraw_` with `S 5 -solid`, and the second one's `_draw_` and `_hdraw_` start with `S 15 -setlinewidth(3)`

### Requirement: JSON output

`json`, `json0`, `dot_json` and `xdot_json` output SHALL be JSON documents with the structure of the reference's, indented as the reference indents them:

- the graph: `name`, `directed`, `strict`, its attributes, `_subgraph_cnt`, `objects` and `edges`; an anonymous graph or subgraph is named as the reference names it, `%N`, with the graphs without a name and the edges numbered 1, 2, 3 and so on in the order they are read, and `N` twice that number less one
- `objects`: the subgraphs, then the nodes, numbered together by `_gvid` from 0; a subgraph has its `name`, its attributes, its `_gvid`, and the `_gvid` numbers of the subgraphs nested in it (`subgraphs`), of its nodes (`nodes`) and of its edges (`edges`); a node has its `_gvid`, its `name` and its attributes
- `edges`: each edge, in the order they are read, with its `_gvid`, numbered from 0 by tail in node order and by head, the `_gvid` of its `tail` and `head`, and its attributes

Attribute values SHALL be JSON strings, escaped as the reference escapes them (`\"`, `\\` and `\/`), with an HTML label written without its outer angle brackets. An object SHALL list its attributes that have a value, and its `label` whenever the attribute is declared; a subgraph SHALL list the attributes of the layout and the drawing only when they are its own. The attributes SHALL be:

- `json`: those of `xdot` output, with the drawing operations as arrays of JSON objects, one per operation, with its `op` and its fields as the reference names them (`color`, `grad`, `points`, `rect`, `pt`, `align`, `width`, `text`, `size`, `face`, `style`), numbers with three decimals
- `json0`: those of `dot` output
- `dot_json` and `xdot_json`: those of the graph as it is read, without layout, so that the two are the same for the inputs of v1

For an input with several graphs, the output SHALL be the JSON document of each graph, one after another, as the reference writes them.

#### Scenario: Graph as it is read

- **WHEN** `digraph G { a -> b [label="x"]; a [shape=rect, color="#FF0000"] }` is rendered as `dot_json`
- **THEN** the output is byte-identical to the reference's: `objects` holds `a`, with `_gvid` 0, `color`, `label` `\\N` and `shape`, and `b`, with `_gvid` 1 and `label`, and `edges` holds one edge with `tail` 0, `head` 1 and `label` `x`

#### Scenario: Subgraphs in json0

- **WHEN** `digraph G { subgraph cluster_x { label="C"; a; subgraph cluster_y { b } } { rank=same; c; d } a -> b }` is rendered as `json0`
- **THEN** `_subgraph_cnt` is 3, `objects` holds `cluster_x`, `cluster_y` and the anonymous subgraph with `_gvid` 0, 1 and 2, then `a`, `b`, `c` and `d` with `_gvid` 3 to 6; `cluster_x` lists `subgraphs` `[1]`, `nodes` `[3,4]` and `edges` `[0]`; and the nodes and the edge have `pos`

#### Scenario: Drawing operations in json

- **WHEN** `digraph { a }` is rendered as `json`
- **THEN** `a` has a `_draw_` array with a `c` operation of `color` `#000000` and an `e` operation with a `rect`, and an `_ldraw_` array with an `F` operation of `size` 14 and `face` `Times-Roman`, a `c` operation and a `T` operation with the `text` `a`

#### Scenario: Anonymous subgraphs

- **WHEN** `digraph { {rank=same; a} {rank=same; b} subgraph x {c} {d} }` is rendered as `dot_json`
- **THEN** the graph is named `%1` and its subgraphs `%3`, `%5`, `x` and `%7`

#### Scenario: Anonymous subgraph after edges

- **WHEN** `digraph { a -> b -> c; d; { rank=max; d } }` is rendered as `json`
- **THEN** the graph is named `%1` and its subgraph `%7`, after the edges numbered 2 and 3

#### Scenario: Escapes

- **WHEN** a node has `label=<<TABLE><TR><TD></TD></TR></TABLE>>` and another one `label="say \"hi\""`, rendered as `dot_json`
- **THEN** their labels are `"<TABLE><TR><TD><\/TD><\/TR><\/TABLE>"` and `"say \"hi\""`

### Requirement: Plain output with ports

`plain-ext` output SHALL be the `plain` output in which the tail and the head of each `edge` line are followed by their ports as they are written (`a:p`).

#### Scenario: Port on an edge line

- **WHEN** a graph with an edge `a:p -> b` is rendered as `plain-ext`
- **THEN** the edge line starts with `edge a:p b`, and the rest of the output is its `plain` output

### Requirement: Inline SVG

`svg_inline` output SHALL be the `svg` output without the XML declaration, without the document type, and without the namespace attributes of the root `svg` element, as the reference writes it for inclusion in another document.

#### Scenario: No prolog

- **WHEN** `digraph G { a }` is rendered as `svg_inline`
- **THEN** the output starts with the comment `<!-- Generated by graphviz version`, its root element starts with `<svg width="62pt" height="44pt"` and has no `xmlns` attribute, and the rest is its `svg` output
