# Tasks

## 1. Conformance cases

- [x] 1.1 Add curated cases under `testdata/conformance/cases/dot-language` and `testdata/conformance/cases/dot-layout` for every scenario of the delta specs and record them with `make reference`. Cover:
  - tables with spans, fills, fixed sizes, fractions of a point, borders, spacing, padding, `COLOR`, and `TD` attributes over the table's;
  - tables that are not table labels: cell content and `ROWSPAN`;
  - ports in `TB` and `LR`, two and three ports of one node, flat edges between ports;
  - unknown ports, a port named like a compass point, a port on a loop, compass points;
  - `tailport` and `headport` set as attributes, and ports inside clusters.

  Verify that a second run leaves the references unchanged and that the reference exits with status 0 on every new case.
- [x] 1.2 Change the "not supported yet" fixture of the conformance tests from a node port to a compass point (`a:n -> b`); verify `go test ./internal/conformance/...` passes

## 2. Language

- [x] 2.1 Read table labels of several cells (design D2): the whole form of the dot-language spec, with each value checked and the fractions of `WIDTH` and `HEIGHT` dropped. Verify table tests for:
  - gopuml's grids and single-cell tables;
  - each `TABLE` and `TD` attribute, and `COLSPAN`;
  - other cases of names, and white space between the elements;
  - the forms that are not table labels: cell content, `ROWSPAN`, other attributes, attributes on `TR`, bad values.
- [x] 2.2 Accept `tailport` and `headport`, and warn that a compass point or a port on a loop is not supported yet, using `center` (design D4); verify the "Port attributes give no warning", "Compass point", "Port on a loop", "Cell with a port", "Grid that gopuml writes", "Row span" and "Cell with content" scenarios
- [x] 2.3 Return the reference's `node a, port P unrecognized` warnings in its order, and append them in `gogv.go` (design D4); verify the "Unknown port" and "Port named like a compass point" scenarios, and a test of the order over several nodes and edges

## 3. Table layout

- [x] 3.1 Lay out tables when they are read (design D3). Cover column widths and row heights, spanning cells, borders, spacing and padding with the `TD` values over the table's, `FIXEDSIZE` and the content size of tables without it, and the cell boxes. Verify:
  - the "Default borders, spacing and padding", "Spanning cell", "Cell filling its row", "Borders, spacing and padding" and "Fractions of a point" scenarios;
  - slice 3's "Table label size" scenarios;
  - the division of a spanning cell's extra width, in fractions of a point and in the order of the cells, against `table-span-uneven`.

## 4. Ports in the layout

- [x] 4.1 Resolve each edge end's port into its offsets along and across the rank and its box, in `TB` and `LR`, and give edges with ports chains of their own (design D4); verify unit tests on the offsets of the "Node in line with a port" node in `TB` and `LR`
- [x] 4.2 Read port positions in the median and the crossing counts (design D5); verify the "Order of the ports" scenario, and that the curated cases without ports keep their outcomes
- [x] 4.3 Add port offsets to the auxiliary graph of the x positioning as δ (design D6); verify the "Node in line with a port" scenario, and the positions of the two-port case within the tolerance of the reference's `plain`
- [x] 4.4 Route edges from and to the port's box, for chains, flat edges between neighbors and flat arcs (design D7); verify the "Node in line with a port" and "Flat edge between ports" scenarios, and that in the `LR` case the edges leave the right side of the cell at its vertical center

## 5. Output

- [x] 5.1 Draw tables from their cell boxes in `svg` (design D8); verify the "Cells and their colors" and "Grid without borders" scenarios, the `stroke-width` and the inset of the "Borders, spacing and padding" case, and that slice 3's curated table cases keep the polygons of their references
- [x] 5.2 Write the ports in the `title` of edge groups, and keep the comment before the group as `tail->head`; verify the "Edge title with a port" scenario

## 6. Acceptance

- [x] 6.1 Run the conformance suite with the gopuml corpora. List the remaining differences in `known-differences.yaml` with their reasons, under `tune-layout-heuristics`, and re-check slice 3's entries. Verify:
  - `make conformance` reports no failed case;
  - the 17 gopuml graphs are judged rather than reported as not supported yet.
- [x] 6.2 Check determinism and hygiene for the new code: provenance headers, fused multiply-add guards and race safety. Verify that these pass:
  - `go run ./tools/provenance`;
  - the fused multiply-add tests;
  - `go test -race ./...`;
  - `make lint`;
  - `make determinism`, byte-identical on arm64, amd64 and wasm.
- [x] 6.3 Update the list of what the current slice supports in `README.md`; verify it names table labels of several empty cells and node ports, and that the next slices of version 1 are the formats
