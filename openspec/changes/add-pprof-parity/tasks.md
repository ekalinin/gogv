# Tasks

## 1. Corpus and cases

- [x] 1.1 Add `tools/pprofcorpus` (design D7):
  - an embedded workload of hashing, sorting and string building under `pprof.Do` labels, and allocations of three sizes;
  - the DOT of `go tool pprof -dot` for CPU and heap profiles, with and without `-nodecount`, `-tagfocus`, `-focus` and `-diff_base`, written into `testdata/conformance/cases/pprof`;
  - a Makefile target that runs it.

  Run it and commit the corpus. Verify that the tool writes the 7 files, that gogv parses each of them, and that `make reference` records their references with exit status 0.
- [x] 1.2 Add curated cases for every scenario of the delta specs, and record them with `make reference`:
  - the attributes that pprof writes, style lists and an unsupported style in a list;
  - weights in the ranking and the positioning;
  - `box3d`, and pen widths 1 to 6 at arrowheads;
  - fills with each choice of color and on `plaintext`;
  - dotted edges, tooltips, ids, and edge order in svg.

  Verify that a second run leaves the references unchanged and that the reference exits with status 0 on every new case.

## 2. Language

- [ ] 2.1 Support the attributes and values of the `dot-language` delta in `internal/attrs`, with `style` checked as a list (design D2). Verify:
  - the "Attributes that pprof writes give no warning", "List of styles" and "Style that is not supported in a list" scenarios;
  - that the existing attribute tests pass.

## 3. Layout

- [ ] 3.1 Weigh rank edges and the pulls of the x positioning by `weight` (design D3). Verify:
  - the "Heavy edge kept short", "Heavy edge kept straight" and "Heavy edge across ranks" scenarios;
  - that the curated cases without `weight` keep their positions, compared with `main`.
- [ ] 3.2 Size `box3d` nodes as boxes and give their outline from a geometry helper (design D5); verify the "Box with depth" scenario
- [ ] 3.3 Carry the fill color, the edge style and the pen width in `layout.Result`, and stop arrowheads short of their nodes by the miter of their pen, widening them above 4 points (design D4). Verify:
  - the "Wider pen" scenario;
  - the arrowheads of the pen width cases within 0.2 points of the reference's.

## 4. Output

- [ ] 4.1 Draw fills, `box3d` outlines, dotted edges and pen widths in `svg` (design D6); verify the "Filled nodes", "Box with depth in svg" and "Dotted and wide edges" scenarios, and that `svg` is byte-identical to `main` for the curated cases without the new attributes, edge group ids aside
- [ ] 4.2 Give node groups their `id`, number edge groups in edge order, and wrap tooltips in `a` elements as the reference does (design D6); verify the "Tooltips and ids that pprof uses" and "Edge groups in edge order" scenarios
- [ ] 4.3 Write the new styles, fills and pen widths in `plain` and xdot, and through xdot in JSON. Verify:
  - the "Styles in plain", "Filled node in xdot" and "Dotted and wide edges in xdot" scenarios;
  - the `dot`, `xdot` and `json` comparisons on the pprof cases.

## 5. Acceptance

- [ ] 5.1 Run the conformance suite with the gopuml corpora, and list the remaining differences in `known-differences.yaml` with their reasons. Verify:
  - `make conformance` reports no failed case;
  - the pprof cases are judged, none of them reported as not supported yet;
  - no other case changes its outcome.
- [ ] 5.2 Check pprof end to end with gogv installed as `dot`. Verify:
  - `go tool pprof -svg` writes the graph without a warning;
  - the Graph view of `go tool pprof -http` finds the groups `node<N>` with a `polygon` to highlight, fetched from its page.
- [ ] 5.3 Check determinism and hygiene. Verify that these pass:
  - `go run ./tools/provenance`;
  - the fused multiply-add tests;
  - `go test -race ./...`;
  - `make lint`;
  - `make determinism`;
  - `make perf`, with no benchmark significantly slower than `main`.
- [ ] 5.4 Update the Status and "Not supported yet" lists of `README.md`; verify they name the newly supported attributes and values, and still list `dashed`, `bold`, `peripheries` and `URL` as not supported yet
