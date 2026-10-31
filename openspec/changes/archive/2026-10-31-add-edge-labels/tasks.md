# Tasks

## 1. Conformance cases

- [x] 1.1 Add curated cases under `testdata/conformance/cases/dot-language` and `testdata/conformance/cases/dot-layout` for every scenario of the delta specs (edge labels in `TB` and `LR`, long labels, labels with `minlen=2`, fonts, two labeled edges between the same nodes, flat edges and loops with labels, node and edge `xlabel`, `headlabel` and `taillabel`, `sametail`, single-cell tables on nodes, edges and clusters with and without `FIXEDSIZE` and `BGCOLOR`, tables that are not single-cell, escapes in edge labels, labeled edges inside and between clusters) and record them with `make reference`; verify that a second run leaves the references unchanged and that the reference exits with status 0 on every new case
- [x] 1.2 Change the "not supported yet" fixture of the conformance tests from an edge label to a node port; verify `go test ./internal/conformance/...` passes

## 2. Language

- [x] 2.1 Read single-cell table labels (design D2): one `TABLE` with one `TR` with one empty `TD`, names without regard to case, white space between elements, only `BGCOLOR`, `FIXEDSIZE`, `WIDTH` and `HEIGHT` on the table; verify table tests for the gopuml form, each attribute, other cases of names, and the forms that are not single-cell (cell content, `PORT`, two cells, other attributes)
- [x] 2.2 Accept `xlabel` on nodes and `label`, `xlabel`, `headlabel`, `taillabel`, `fontname`, `fontsize`, `fontcolor` and `sametail` on edges, and table labels where the spec allows them; keep the warning for other HTML labels and for HTML labels of the root graph; expand escapes in edge, head and tail labels; verify the "Supported attributes", "Graphviz features that are not supported yet", "Single-cell table labels" and "Plain-text labels" scenarios

## 3. Sizes

- [x] 3.1 Give table labels their size (`WIDTH` by `HEIGHT`, 12 points for a side not given) and size nodes and clusters around them with the margins of text labels; verify the "Node sized by its table", "Edge label table" and "Table label of a cluster" scenarios

## 4. Layout of edge labels

- [x] 4.1 Double the ranks of a graph with an edge label and halve `ranksep` (design D3), with chains of their own for labeled edges and the label on the middle rank; verify the "Label beside the edge" and "Ranks of the other edges" scenarios, the rank distances of the curated cases against the reference's `plain`, and that the cases without edge labels keep their outcomes
- [x] 4.2 Give label nodes their off-center extents in the positioning (design D4) and route the edge on the left of its label, above it in `rankdir=LR`; verify the "Label beside the edge", "Left to right", "Two labeled edges between the same nodes" and "Edge label table" scenarios, and label centers within 2 points of the reference's on the curated cases but those listed as known differences
- [x] 4.3 Lay out the labels of flat edges and loops (design D5); verify the "Flat edge" and "Loop" scenarios against the curated cases

## 5. External labels and sametail

- [x] 5.1 Place node and edge `xlabel`, `headlabel` and `taillabel` after the layout (design D6), growing the drawing; verify the "Node xlabel", "Head and tail labels" and "Edge xlabel" scenarios and the label centers of the curated cases against the reference's
- [x] 5.2 Make the edges of a `sametail` group share their tail point (design D7); verify the "Shared tail point" scenario

## 6. Output

- [x] 6.1 Write edge labels, external labels and table labels in `svg` (design D8), and edge labels with their centers in `plain`, HTML labels as the reference writes them; verify the "Edge label", "Table label that gopuml reads", "Edge label in plain" and "Font of an edge label" scenarios

## 7. Acceptance

- [x] 7.1 Run the conformance suite with the gopuml corpora and list the remaining differences in `known-differences.yaml` with their reasons under `tune-layout-heuristics`; verify `make conformance` reports no failed case and that the 77 gopuml graphs with edge labels or single-cell tables and no feature of a later slice are judged rather than reported as not supported yet
- [x] 7.2 Check determinism and hygiene for the new code: provenance headers, fused multiply-add guards, race safety; verify `go run ./tools/provenance`, the fused multiply-add tests, `go test -race ./...`, `make lint` and `make determinism` (byte-identical on arm64, amd64 and wasm) pass
- [x] 7.3 Update the list of what the current slice supports in `README.md`; verify it names edge labels, external labels, `sametail` and single-cell table labels
