# Proposal

## Why

There is no native Go implementation of Graphviz `dot`. Go programs either run the external `dot` binary or embed Graphviz compiled to WebAssembly (goccy/go-graphviz and its forks). gopuml (`github.com/ekalinin/gopuml`) needs `dot` for class, object, description, state and legacy activity diagrams, and its WebAssembly build cannot lay them out at all. A standalone Go clone of `dot` removes the external dependency for any Go program, and gopuml is its first consumer.

## What Changes

This change is the first of the vertical slices that make up v1 of the project. Each slice gives a working `dot`, from DOT input to SVG, on a growing part of the v1 language, and closes with a passing conformance run.

Project-wide decisions:

- A new Go module that implements the Graphviz `dot` program in Go: no cgo, no embedded WebAssembly, no system Graphviz. MIT license.
- Clean-room implementation: from published papers, the Graphviz documentation and the observed behavior of `dot`, without reading the source code of Graphviz or of its ports (Smetana, graphviz-pure, go-graphviz).
- Reference behavior: Graphviz 16.1.0 built without Pango, run in a pinned container and used as a black box.
- Parity target: layouts similar to the reference's within a tolerance in the spirit of gopuml's (output size, structural similarity of the rasterized output). `canon` output matches the reference exactly.
- Text width comes from font metrics bundled with the module, never from system fonts.
- Output is deterministic: byte-identical across runs, machines and operating systems, as gopuml requires of its own output.

The v1 language is the DOT attribute subset that gopuml emits, plus plain-text labels:

- graphs: `digraph`, subgraphs, nested clusters, `rank=same|min|max|source|sink`; `nodesep`, `ranksep`, `rankdir=LR`, `splines=spline|polyline`, `remincross`, `searchsize`
- nodes: shapes `rect`, `plaintext`, `point`, `ellipse`, `circle`, `diamond`, `hexagon`, `octagon`; `width`, `height`, `color`; `style=rounded` on `rect`
- edges: node ports (`a:port`), `minlen`, `color`, `style=invis`, `constraint=false`, `sametail`, `arrowhead` and `arrowtail` (`none`, `empty`, `normal`), `dir`, `arrowsize`
- clusters: `label`, `labeljust`, `style=solid`, `color`
- labels: plain-text `label` on nodes, edges, graphs and clusters, with the escapes `\N`, `\G`, `\E`, `\T`, `\H`, `\L` and the line breaks `\n`, `\l`, `\r`; `fontname`, `fontsize`, `fontcolor`; `xlabel`, `taillabel`, `headlabel`
- HTML-like labels: `TABLE`, `TR`, `TD` with `BORDER`, `CELLBORDER`, `CELLSPACING`, `CELLPADDING`, `BGCOLOR`, `COLOR`, `FIXEDSIZE`, `WIDTH`, `HEIGHT`, `COLSPAN`, `PORT`, and empty cells

The v1 output formats are `svg`, `svg_inline`, `plain`, `plain-ext`, `canon`, `dot`, `gv`, `xdot`, `xdot1.2`, `xdot1.4`, `json`, `json0`, `dot_json`, `xdot_json`.

This change (slice 1, foundation):

- the Go module, a Go library API safe for concurrent use, and a `dot` command line with `-T`, `-o` and `-V`
- a DOT parser for the whole v1 syntax, and `canon` output
- the pinned reference container, the conformance harness and its corpus. Outputs are compared with the reference layer by layer: `canon` exactly, `plain` and `svg` within the tolerance
- font metric tables for text width
- layout of graphs without clusters, edge labels and HTML-like labels: ranking, crossing minimization, positioning and spline routing (`splines=spline` and `splines=polyline`), with the node, edge and graph attributes of the v1 language that do not belong to a later slice, and plain-text labels on nodes and graphs
- `svg` and `plain` output

Roadmap of v1, one follow-up change per slice:

1. foundation (this change)
2. clusters: nested clusters, cluster labels, `rank=same|min|max|source|sink`
3. edge labels: `label` on edges, `xlabel`, `taillabel` and `headlabel`, `sametail`
4. HTML tables: `TABLE`, `TR`, `TD` and node ports
5. formats: `svg_inline`, `plain-ext`, `dot`, `gv`, the `xdot` and the `json` families

After slice 5, gopuml can start its switch.

Out of scope of v1, planned as later changes:

- the rest of the Graphviz attribute language: other node shapes including `record` and `Mrecord`, HTML text markup (`FONT`, `B`, `I`, ...), other styles and arrow types, color names and schemes, `concentrate`, compound edges (`lhead`, `ltail`), images (`IMG`, `image`, `shapefile`)
- text metrics of the fonts that the reference knows besides the Times, Helvetica and Courier families (`DejaVu Sans`, `Verdana`, `Georgia`, `Consolas`, `Calibri`, `Trebuchet MS`, `Open Sans`)
- `cmapx` and `cmapx_np`, together with the `URL`, `href`, `tooltip` and `target` attributes
- `splines=ortho`
- the other layout engines (`-K neato`, `fdp`, `sfdp`, `circo`, `twopi`, `osage`, `patchwork`)
- raster output (PNG, JPG) and PDF

## Capabilities

### New Capabilities

- `dot-language`: parsing graphs in the DOT language: graphs, subgraphs, clusters, attributes, quoted and HTML strings, node ports, plain-text labels and their escapes
- `dot-layout`: the `dot` layout engine for graphs without clusters, edge labels and HTML-like labels: ranking, crossing minimization, positioning, spline routing, node shapes, text width from bundled font metrics
- `output-formats`: the `svg`, `plain` and `canon` outputs, deterministic across platforms
- `cli`: a `dot` command line compatible with Graphviz for `-T`, `-o` and `-V`
- `library-api`: a Go package API that lays out and renders a DOT graph, safe for concurrent use
- `upstream-compatibility`: the similarity contract with Graphviz 16.1.0 built without Pango and how it is measured

### Modified Capabilities

None. The project has no specs yet.

## Impact

- New repository `gogv` with the Go module `github.com/ekalinin/gogv` (Go 1.26), under the MIT license.
- A pinned reference container with Graphviz 16.1.0 built without Pango, used to record reference outputs. Recording references needs Docker.
- gopuml (`github.com/ekalinin/gopuml`) is the first consumer. Its switch is done in its own changes, not in this one:
  - move its reference container to Graphviz 16.1.0 built without Pango and regenerate its references
  - lay out with this module by default once gopuml's conformance passes with it; for `splines=ortho`, run the external `dot` when one is found and report "not supported yet" otherwise (always in the WebAssembly build)
