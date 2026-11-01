## ADDED Requirements

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

The operations SHALL be those that draw what `svg` draws, in the same order: pen and fill colors (`c`, `C`), styles (`S`), polygons (`p`, `P`), ellipses (`e`, `E`), Bezier curves (`B`), fonts (`F`) and texts (`T`), with colors in lower case, numbers with at most two decimals and the backslashes of texts doubled, as the reference writes them; a border wider than 1 point SHALL set its pen width (`S 15 -setlinewidth(2)`) when it changes, and an object that sets `style=solid` SHALL start its outline with `S 5 -solid`. An invisible object SHALL have no drawing operations. For the formats of v1, the three versions SHALL differ only in `xdotversion`.

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
