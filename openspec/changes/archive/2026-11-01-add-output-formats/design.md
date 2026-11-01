# Design

## Context

See proposal.md for the motivation and the scope of this slice, and the delta specs for the behavior.

Current state:

- `output.Canon` writes the graph back in DOT, byte-identical to the reference's `canon` on every curated case. It handles statement placement, defaults, inherited attributes, sorted attribute lists and the line breaks of long strings.
- `output.SVG` draws `layout.Result`: node outlines, cluster boxes, edge curves and arrowheads, text and table labels. `output.Plain` writes the `plain` lines. Both read `layout.Result` only.
- `gogv.Formats` lists `canon`, `plain` and `svg`. `laterFormats` lists the 11 formats of this slice, which `Render` and the command report as not supported yet. The command already defaults to `dot` without `-T`.
- The conformance runner compares `canon`, `plain` and `svg` for every case, and `tools/refgen` records those three. `make determinism` hashes every format of `gogv.Formats` for every curated case.

Facts observed on the reference that shape the approach:

- **`dot` and `gv`:**
  - `dot` is `canon` with layout attributes added to the objects: `bb`, `lp`, `lwidth` and `lheight` on the graph and the clusters; `pos`, `width`, `height` and `xlp` on nodes; `pos`, `lp`, `xlp`, `head_lp` and `tail_lp` on edges.
  - A node that `canon` leaves out, because it only appears in edges, gets a statement of its own once it has `pos`.
  - Unlike `canon`, `dot` writes every anonymous subgraph of the root or of a cluster, even an empty one, and gives a subgraph `bb`, `lp`, `lwidth` and `lheight` only when the layout sets them on it: they are not inherited.
  - Numbers have at most five significant digits (`88.651`, `36.41`, `124.8`). An edge's `pos` starts with `e,x,y` for a head arrowhead, then lists the clipped curve.
  - `gv` is byte-identical to `dot` on all 236 curated cases.
- **xdot:**
  - xdot adds `_draw_`, `_ldraw_`, `_hdraw_`, `_tdraw_`, `_hldraw_` and `_tldraw_`, and `xdotversion`.
  - The graph's `_draw_` paints the background: `c 9 -#fffffe00 C 7 -#ffffff P 4 ...`.
  - Text is `F size n -face`, a color, then `T x y j width n -text`, with `j` -1, 0 or 1 for left, center and right. Table labels are `S 5 -solid`, colors and `p` or `P` polygons, in the order of the `svg` drawing.
  - An edge's `xlabel` goes to its `_ldraw_`, and a node's `xlabel` to the node's `_ldraw_`, after its label.
  - Numbers have at most two decimals. Texts double their backslashes. Long strings break after 128 characters, as in `canon`.
  - An object that sets `style=solid` starts its outline with `S 5 -solid`. A table border wider than 1 point sets `S 15 -setlinewidth(N)` when the width changes.
  - On the curated cases, `xdot1.2` and `xdot1.4` differ from `xdot` only in `xdotversion`, except for HTML text markup, which v1 does not support.
- **JSON:**
  - The structure: `objects` lists the subgraphs, in preorder of declaration, then the nodes, all numbered by `_gvid`. A subgraph lists the `_gvid` numbers of its nested subgraphs, nodes and edges.
  - Anonymous graphs and subgraphs are named `%N`: the graphs without a name and the edges share one count in the order they are read, and `N` is twice their number less one. An unnamed root takes `%1`, and in `a -> b -> c; { rank=max; d }` the subgraph takes `%7`.
  - Edges are listed in the order they are read, but numbered by tail in node order and by head. A declared `label` is listed for every object, even when it is empty.
  - Values are strings. Strings escape `"`, `\` and `/`, and HTML labels lose their outer angle brackets. Operation numbers have three decimals.
  - A subgraph lists the attribute values it inherits (`cluster_y` shows its parent's `label`), and the root lists `label` `""` when a cluster has a label.
  - `dot_json` and `xdot_json` are the graph as read, without layout, byte-identical to each other on all curated cases.
- **`plain-ext` and `svg_inline`:**
  - `plain-ext` differs from `plain` only in the ports on edge lines, on all curated cases.
  - `svg_inline` drops the XML declaration, the document type and the `xmlns` attributes.
- **Text widths:** our Times metrics give the reference's widths in xdot `T` operations (62.14 for ten letters `a` at 14 points).

## Goals / Non-Goals

**Goals:**

- Every v1 format written with the reference's structure, so that programs that read `dot`, xdot or JSON output find the same objects, attributes and drawing operations.
- `svg`, `plain` and `canon` output unchanged, byte for byte.

**Non-Goals:**

- HTML text markup and the `t` operation of xdot 1.7, gradients, and the other features of xdot beyond what v1 draws.
- Formats outside v1 (`cmapx`, `ps`, `png`, ...): they keep the reference's "not recognized" error.
- New formats for the gopuml corpora, which keep `canon`, `plain` and `svg`.

## Decisions

### D1. Sources

The foundation's clean-room rule holds. Sources added for this slice:

- the Graphviz documentation: the output formats (`canon`, `dot`, `gv`, `xdot`, `json`, `plain`, `plain-ext`, `svg`, `svg_inline`), the xdot drawing operations, and the attributes `bb`, `pos`, `lp`, `xlp`, `head_lp`, `tail_lp`, `lwidth`, `lheight`, `xdotversion` and `_draw_` and its kin
- black-box observation of the reference

### D2. `dot` is `canon` over the laid-out graph

After the layout, the attributes of the layout are set on the graph itself: symbols declared at the root without a default, values set on the graph, the clusters, the nodes and the edges. The `canon` writer then writes it, in a mode for laid-out graphs that keeps the anonymous subgraphs of the root and of clusters and writes the graph attributes of the layout only where they are set. The graph belongs to one `Render` call, so setting attributes on it keeps `Render` safe for concurrent use. A number formatter writes the reference's five significant digits; `gv` is the same writer.

Alternative considered: a writer of its own for `dot`. Rejected: `canon` already places statements, defaults and inherited values as the reference does, and the reference's `dot` follows the same rules once the layout attributes are set.

### D3. xdot operations follow the `svg` drawing

A writer in `internal/output` turns `layout.Result` into xdot operations, object by object and element by element as `svg.go` draws them, with the same geometry helpers. It writes:

- the background;
- cluster boxes and labels;
- node outlines, ellipses, rounded rectangles and points;
- text, with its font and justification;
- table labels;
- edge curves and arrowheads.

Each object's operation strings are set as its `_draw_`, `_ldraw_`, `_hdraw_`, `_tdraw_`, `_hldraw_` and `_tldraw_` attributes before `Canon` writes the graph (D2), which also breaks the long strings. Coordinates keep `y` growing upward, as in `plain`, not negated as in `svg`. `xdot1.2` and `xdot1.4` change only `xdotversion`.

Alternative considered: one drawing model shared by `svg` and xdot. Rejected for this slice: it would rewrite `svg.go`, whose output must stay byte-identical. Two writers that follow the same order are checked against the reference on every curated case.

### D4. JSON over the same attributed graph

A JSON writer walks the graph after D2 and D3:

1. It numbers the subgraphs in preorder of declaration and then the nodes, numbers the edges in edge order, and names anonymous graphs and subgraphs `%1`, `%3`, `%5` and so on.
2. It writes each object's attributes as the reference lists them.
3. In `json`, it turns the xdot strings into arrays of operation objects.

`json0` skips the xdot attributes, and `dot_json` and `xdot_json` run on the graph as read, without layout, as `canon` does. Escaping and indentation follow the reference. Which attributes an object lists, inherited values and root declarations included, is calibrated on the `output-formats` cases.

### D5. `plain-ext` and `svg_inline` are options of the existing writers

`plain-ext` is `plain` with the ports appended to the edge ends. `svg_inline` is `svg` without the prolog, the document type and the namespace attributes. Both are small options of `Plain` and `SVG`, so the shared output cannot drift.

### D6. Formats in the library and the command

`gogv.Formats` lists the 14 formats of v1, and `laterFormats` becomes empty. `FormatError` keeps its `NotSupportedYet` field for the API, now never set for a format of v1. The command already uses `dot` without `-T` and reports it as not supported yet today, so it writes `dot` output with no change of its own. `dot_json` and `xdot_json` skip the layout, like `canon`.

### D7. Conformance compares parsed outputs

The runner compares each format that a case has a reference for:

- **DOT formats** (`dot`, `gv`, the xdot family):
  1. Both outputs are parsed with `internal/dotlang` and walked side by side: subgraphs, nodes and edges in order, then the attribute names of each.
  2. Values must be equal, except for the layout attributes (`bb`, `pos`, `lp`, `xlp`, `head_lp`, `tail_lp`, `width`, `height`, `lwidth`, `lheight`) and the drawing attributes.
  3. Drawing attributes are tokenized into operations and compared by kind, color, style, font and text.
  4. The root's `bb` is checked within 5%.
- **JSON formats** (`json`, `json0`): decoded and compared the same way.
- **`dot_json` and `xdot_json`:** compared byte for byte.
- **`plain-ext`:** compared as `plain`, plus its edge ends with ports.
- **`svg_inline`:** compared as `svg`, plus its lines before the root element.

`tools/refgen` records, per case, the formats its corpus asks for:

- `canon`, `plain`, `svg`, `dot`, `xdot` and `json` for the curated cases;
- all 14 formats for `testdata/conformance/cases/output-formats`;
- `canon`, `plain` and `svg` for the gopuml corpora.

The CLI case `cli/later-format` becomes `cli/xdot`, and `cli/default-format` is added. Their standard output is compared byte for byte, as for every CLI case.

## Risks / Trade-offs

- [The reference's number formats have edge cases that the documentation does not give (trailing zeros, negative zero, large coordinates)] → The formatter is checked on the curated cases, and `cli/xdot` and `dot_json` are compared byte for byte.
- [Which attributes the reference lists for each JSON object, inherited values and root declarations included, is not documented] → `output-formats` cases with nested clusters, anonymous subgraphs, defaults and inherited labels calibrate the writer, and the structural comparison reports every difference.
- [Two writers, `svg` and xdot, draw the same things and can drift apart] → The xdot writer follows `svg.go` element by element. The structural comparison of xdot runs on all curated cases, and the `svg` references are unchanged.
- [About 700 new reference files] → They are small text files, recorded in the same `make reference` run.
- [The `plain` and `svg` comparisons took one document; an input with several graphs gives several] → The runner compares them document by document.

## Migration Plan

No migration: the change only adds behavior. When it is archived, the Purpose of the `output-formats` main spec names every format of v1 instead of `svg`, `plain` and `canon`.
