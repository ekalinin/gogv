# Proposal

## Why

Slices 1 to 4 lay out the whole v1 language and write `svg`, `plain` and `canon`. The other 11 output formats of v1 are still reported as not supported yet: `svg_inline`, `plain-ext`, `dot`, `gv`, `xdot`, `xdot1.2`, `xdot1.4`, `json`, `json0`, `dot_json` and `xdot_json`. `dot` is also the format that the command writes without `-T`, as Graphviz's does. This is slice 5 of the v1 roadmap, the last one: after it, the command can replace Graphviz's `dot` for every format of v1. gopuml writes only `svg`, so no gopuml graph depends on this slice.

## What Changes

- `dot` and `gv`: the graph written back in DOT, as `canon` writes it, with the attributes of the layout:
  - `bb` on the graph and its clusters, with `lp`, `lwidth` and `lheight` when they have a label;
  - `pos`, `width` and `height` on nodes, and `xlp` for an `xlabel`;
  - `pos` on edges, with the ends of their arrowheads, and `lp`, `xlp`, `head_lp` and `tail_lp` for their labels.
- `xdot`, `xdot1.2` and `xdot1.4`: the `dot` output with the drawing of every object as xdot drawing operations (`_draw_`, `_ldraw_`, `_hdraw_`, `_tdraw_`, `_hldraw_`, `_tldraw_`) and the `xdotversion` of the format, as the reference writes them.
- The JSON family, in the structure of the reference: graph, subgraphs, nodes and edges with their `_gvid` numbers.
  - `json`: the attributes of `xdot`, with the drawing operations as JSON objects.
  - `json0`: the attributes of `dot`.
  - `dot_json` and `xdot_json`: the graph as it is read, without layout.
- `plain-ext`: the `plain` output with the ports of the edge ends.
- `svg_inline`: the `svg` output without the XML declaration, the document type and the namespace attributes, for inclusion in another document.
- Command line: every v1 format is accepted, and the output without `-T` is `dot`. No format of v1 remains "not supported yet"; formats outside v1 keep the reference's "not recognized" error.
- Conformance:
  - `dot`, `xdot` and `json` are recorded and compared for every curated case;
  - all the formats are recorded and compared for a new set of curated cases, `output-formats`;
  - the outputs are parsed, and their structure must match the reference's exactly: objects, attribute names, the values that do not depend on the layout, and the kinds, colors, fonts and texts of the drawing operations;
  - the numbers of the layout are left to `plain` and `svg`, except the size of the drawing (`bb`), which must be within 5%;
  - `dot_json` and `xdot_json`, which have no layout, must be byte-identical.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `output-formats`: the `dot`, `gv`, xdot, JSON, `plain-ext` and `svg_inline` formats.
- `cli`: every v1 format is accepted; without `-T` the output is `dot`.
- `upstream-compatibility`: the similarity contract and the corpus cover the new formats.

## Impact

- Code:
  - `internal/output`: the new writers, built on the `canon` writer and on the drawing that `svg` already does;
  - `gogv.go`: the list of formats;
  - `internal/cli`;
  - `internal/conformance` and `tools/refgen`: the comparison of the new formats and their references.
- Conformance: references of `dot`, `xdot` and `json` for every curated case, and of every format for the `output-formats` cases. The CLI case `cli/later-format` becomes a case of `xdot` output.
- Determinism: `make determinism` hashes the outputs of every format, so it covers the new writers.
- `README.md`: the list of formats.
- No change to the layout, the library API or the gopuml corpora.
