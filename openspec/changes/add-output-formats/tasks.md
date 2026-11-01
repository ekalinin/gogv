# Tasks

## 1. Conformance harness and cases

- [x] 1.1 Compare the new formats in the conformance runner (design D7), for the formats that a case has references for. The comparisons:
  - parsed DOT for `dot`, `gv` and the xdot family, with the drawing attributes tokenized into operations;
  - decoded JSON for `json` and `json0`;
  - bytes for `dot_json` and `xdot_json`;
  - `plain` with ports for `plain-ext`, and `svg` with its prolog for `svg_inline`.

  Verify unit tests in `internal/conformance` on fixture outputs:
  - the "Layout numbers in dot" and "Missing drawing operation" scenarios;
  - a missing attribute, object or operation, and a different color or text, fail;
  - a `bb` beyond 5% fails.
- [x] 1.2 Record the references of each corpus in its formats (design D7):
  - `canon`, `plain`, `svg`, `dot`, `xdot` and `json` for the curated cases;
  - all 14 formats for `output-formats`;
  - `canon`, `plain` and `svg` for the gopuml corpora.

  Run `make reference` and verify that the curated cases get their new references, that the existing `canon`, `plain` and `svg` references are unchanged, and that `make harvest` keeps three formats.
- [x] 1.3 Add curated cases under `testdata/conformance/cases/output-formats` for every scenario of the `output-formats` delta, and record them with `make reference`. Cover:
  - layout attributes, clusters and external labels;
  - xdot nodes, edges, tables and versions;
  - JSON structure, anonymous subgraphs and escapes, and several graphs in one input;
  - `plain-ext` ports and `svg_inline`;
  - every shape, rounded rectangles, colors, `rankdir=LR`, graph and cluster labels, invisible edges, and arrowheads at both ends.

  Verify that a second run leaves the references unchanged and that the reference exits with status 0 on every new case.
- [x] 1.4 Rename the CLI case `cli/later-format` to `cli/xdot`, add `cli/default-format` (no `-T`), and record them; verify `go test ./internal/conformance/...` passes

## 2. DOT output

- [x] 2.1 Set the layout attributes on the laid-out graph and write `dot` and `gv` with `Canon` (design D2):
  - `bb`, `lp`, `lwidth` and `lheight` on the graph and the clusters;
  - `pos`, `width`, `height` and `xlp` on nodes;
  - `pos`, with `e,` and `s,`, and `lp`, `xlp`, `head_lp` and `tail_lp` on edges;
  - numbers with five significant digits.

  Verify the "Layout attributes", "Clusters and external labels" and "Same output as gv" scenarios, and the `dot` comparison on every curated case but the known differences.

## 3. xdot output

- [x] 3.1 Write the xdot drawing operations of the background, clusters, nodes, labels, tables, edges and arrowheads, following `svg.go` (design D3), and the `xdotversion` of each version. Verify:
  - the "Node in xdot" scenario byte for byte;
  - the "Edge in xdot", "Table label in xdot" and "Versions" scenarios;
  - the `xdot` comparison on every curated case but the known differences.

## 4. JSON output

- [x] 4.1 Write `json`, `json0`, `dot_json` and `xdot_json` (design D4): `_gvid` numbering, subgraph lists, anonymous names, the attributes of each object as the reference lists them, the drawing operations as objects, escapes and indentation. Verify:
  - the "Graph as it is read" scenario byte for byte;
  - the "Subgraphs in json0", "Drawing operations in json", "Anonymous subgraphs" and "Escapes" scenarios, and several graphs in one input;
  - the `json` comparison on every curated case, and `dot_json` and `xdot_json` byte for byte on the `output-formats` cases.

## 5. Variants of plain and svg

- [x] 5.1 Add the `plain-ext` and `svg_inline` options of `Plain` and `SVG` (design D5); verify the "Port on an edge line" and "No prolog" scenarios, and that `plain` and `svg` output is unchanged

## 6. Library and command

- [x] 6.1 List the 14 formats of v1 in `gogv.Formats`, empty `laterFormats`, and skip the layout for `dot_json` and `xdot_json` (design D6). Verify:
  - the "Supported format", "Format of a later slice", "Default format" and "Unknown format" scenarios through the CLI cases;
  - the library tests of `Render` for every format.

## 7. Acceptance

- [x] 7.1 Run the conformance suite with the gopuml corpora, and list the remaining differences in `known-differences.yaml` with their reasons. Verify:
  - `make conformance` reports no failed case;
  - no case changes its outcome in `canon`, `plain` or `svg`;
  - every curated case is judged in `dot`, `xdot` and `json`.
- [x] 7.2 Check determinism and hygiene for the new code: provenance headers, fused multiply-add guards and race safety. Verify that these pass:
  - `go run ./tools/provenance`;
  - the fused multiply-add test of `internal/output`;
  - `go test -race ./...`;
  - `make lint`;
  - `make determinism`, byte-identical on arm64, amd64 and wasm for all 14 formats.
- [x] 7.3 Update `README.md`; verify it lists the 14 formats of v1 and no longer names a next slice of v1
