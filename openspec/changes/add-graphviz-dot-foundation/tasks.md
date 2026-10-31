# Tasks

## 1. Module and ground rules

- [x] 1.1 Initialize the git repository and the module `github.com/ekalinin/gogv` with `go 1.26`, the MIT `LICENSE`, a `NOTICE` file and a `.gitignore` that ignores `.cache/`; verify `go build ./...` and `go vet ./...` succeed and `git check-ignore .cache/x` matches
- [x] 1.2 Adapt gopuml's `tools/provenance` into a check that every non-test Go file starts with `SPDX-License-Identifier: MIT` and names its sources (paper and section, Graphviz documentation page, observed behavior of the reference, `fdlibm`, gopuml, or `original`) as in design D1; verify its tests reject a file without a header and `go run ./tools/provenance` passes on the repository
- [x] 1.3 Copy gopuml's `internal/jmath` fdlibm functions needed by layout and rendering into `internal/fdlibm` with their `Provenance: fdlibm` headers and Sun notices, and list fdlibm in `NOTICE`; verify their tests pass natively on darwin/arm64 and with `GOARCH=amd64` (Rosetta), with identical results

## 2. Reference container and conformance harness

- [x] 2.1 Write `tools/reference/Dockerfile`: base image pinned by digest, Graphviz 16.1.0 built from its release tarball (SHA-256 checked) without Pango, `resvg` pinned by checksum, for linux/amd64; verify, on the arm64 host under Rosetta emulation, that `dot -V` reports 16.1.0 and that `dot -v` resolves the font of a `Hello world` label without a text layout plugin
- [x] 2.2 Add `make reference`, which runs the reference `dot` with `-Tcanon`, `-Tplain` and `-Tsvg` on every case in `testdata/conformance/cases/<capability>/*.gv` and records stdout, stderr and exit status under `testdata/conformance/reference/`; verify that a second run leaves the recorded files unchanged
- [x] 2.3 Implement the comparison in `internal/conformance`, reusing gopuml's SSIM and `resvg` driver: `canon` byte-identical, `plain` same node names and edges with graph size within 5%, `svg` size within 5% and grayscale SSIM of at least 0.90 at the reference's size; verify with fixture cases that pass, fail on each layer, and pass at SSIM 0.94 and size +3%
- [x] 2.4 Implement case outcomes (passed, known difference from `testdata/conformance/known-differences.yaml`, not supported yet, failed), the failing exit of the run on failed cases only, and `conformance-report.md` with counts per capability and format; verify with a fixture for each outcome, including an unlisted failure that fails the run and a `subgraph cluster_x` case reported as not supported yet
- [x] 2.5 Add `tools/harvest`, which runs gopuml's conformance cases with `GRAPHVIZ_DOT` pointing to a wrapper that saves each DOT input into `.cache/gopuml-dot/` before running the reference `dot`, and add that corpus to the conformance run; verify the cache holds the DOT of the class, description, state and legacy activity cases and is not tracked by git

## 3. Font metrics

- [x] 3.1 Add `tools/fontmetrics`, which downloads Liberation 2.x Serif, Sans and Mono (regular, bold, italic, bold italic) by checksum into `.cache/`, and generates advance-width tables into `internal/textmetrics` with `go generate`; add the OFL notice to `NOTICE`; verify that regenerating gives byte-identical tables and that no font file is committed
- [x] 3.2 Implement text measurement: line width from the tables scaled by `fontsize`, line height, and multi-line labels; verify tests that `iiii` and `WWWW` have the same width in `Courier` and that `fontsize=28` doubles the width of `fontsize=14`
- [x] 3.3 Map `fontname` to a metrics family as the reference does: whole family names without regard to case for Times, Helvetica and Courier, Times for any other name including names with a weight or a slant, and "not supported yet" for `DejaVu Sans`, `Verdana`, `Georgia`, `Consolas`, `Calibri`, `Trebuchet MS` and `Open Sans`; verify a table test and curated cases for those names (`Helvetica-Bold`, `Sans`, `Serif`, `Monospace`, `Verdana`, an unknown name)
- [x] 3.4 Add a check tool that measures node widths of one-character strings per family and style on the reference (through `-Tplain`) and reports the differences with the tables into `.cache/`; verify the report is produced and that no measured value is written into the module

## 4. DOT language

- [x] 4.1 Implement the lexer: identifiers, numerals, double-quoted strings with escaped quotes and backslash-newline continuation, HTML strings with nested angle brackets, `//`, `/* */` and `#` comments, case-insensitive keywords; verify lexer tests for each token kind
- [x] 4.2 Implement the parser and the graph model: `graph`, `digraph`, `strict`, statements, named and anonymous subgraphs, edge chains with subgraph operands, `+` concatenation, ports and compass points, attribute defaults and their scope, strict edge merging, several graphs per input, attributes kept in input order; verify tests for every scenario of the dot-language spec except errors and labels
- [x] 4.3 Report syntax errors as `syntax error in line N near 'TOKEN'` with no output for the graph; verify the scenario and curated error cases against the reference's stderr
- [x] 4.4 Implement plain-text labels: `\N`, `\G`, `\E`, `\T`, `\H`, `\L`, the line breaks `\n`, `\l`, `\r` with their justification, and the default node label `\N`; verify the label scenarios
- [x] 4.5 Build the table of attributes documented by Graphviz (attribute documentation page) and the table of supported attributes and values of the dot-language spec, including `#rrggbb` colors in either case; verify that an unknown attribute gives no warning and every supported one is accepted
- [x] 4.6 Implement "not supported yet" warnings with their defaults: unsupported attributes and values, clusters laid out as plain subgraphs, ignored `rank`, ignored ports and compass points, default labels instead of HTML labels, `dir=none` in undirected graphs; verify the scenarios of the spec and that each warning names the object, the feature and the default

## 5. Canonical output

- [x] 5.1 Implement the `canon` writer, keeping every attribute of the input; add curated cases for the whole DOT grammar; verify all of them are byte-identical to the reference's `canon` in the conformance run

## 6. Layout

- [x] 6.1 Create the per-run layout state (no package-level mutable state) with context checks in long loops; verify `go test -race` with parallel layouts and a test that a canceled context stops a large layout
- [x] 6.2 Implement cycle breaking and network simplex ranking (Gansner et al. 1993, section 2) with `minlen`, `constraint=false` and edges of length zero; verify the ranking scenarios (shortest ranking, minimum length, length zero, no constraint, cycle)
- [x] 6.3 Add virtual nodes for edges that span several ranks; verify that every rank between the ends of a long edge holds one virtual node of it
- [x] 6.4 Implement ordering within ranks: initial order from input order, weighted median and transposition (Gansner et al. 1993, section 3), input order kept on ties, `remincross` and `searchsize` accepted; verify the crossing and input-order scenarios
- [x] 6.5 Implement node sizes and shapes: `rect`, `box`, `rectangle`, `ellipse`, `oval`, `circle`, `diamond`, `hexagon`, `octagon`, `point`, `plaintext`, minimum `width` and `height`, growth to the label with the documented margins, and the intersection of a line with each outline; verify the size and shape scenarios
- [x] 6.6 Implement positioning: x coordinates by network simplex on the auxiliary graph (Gansner et al. 1993, section 4) with `nodesep` and node widths, y coordinates from rank heights and `ranksep`, `rankdir=LR`; verify the separation, straight chain and left-to-right scenarios
- [x] 6.7 Implement edge routing with `splines=spline` (Dobkin et al. 1997) and `splines=polyline`: ends clipped at the outlines, routes around other nodes, distinct curves for edges between the same nodes, self loops and edges within a rank; verify the routing scenarios
- [x] 6.8 Implement arrowheads: `dir`, `arrowhead`, `arrowtail` with `normal`, `empty` and `none`, `arrowsize`, and edges ending where their arrowhead begins; verify the arrowhead scenarios
- [x] 6.9 Implement invisible edges and the graph label below the drawing; verify their scenarios
- [x] 6.10 Add curated layout cases for every dot-layout scenario and run them against the reference; verify each passes or is listed in known differences with a reason
- [x] 6.11 Implement `style=rounded` on `rect`, `box` and `rectangle` (corner radius a third of the smaller side, at most 12 points; drawn as an SVG path; `rounded` in plain) and the "not supported yet" warning for it on `diamond`, `hexagon` and `octagon`; verify the rounded rectangle scenario and curated cases against the reference

## 7. SVG and plain output

- [x] 7.1 Implement number formatting per format (at most two decimals in SVG, the reference's format in `plain`) with fdlibm and `float64(a*b)` guards in layout and rendering arithmetic; verify unit tests of the formatting
- [x] 7.2 Implement the SVG writer: `width` and `height` in whole points, `viewBox`, the graph group with its transform, node and edge groups with their titles escaped as the reference escapes them, `polygon` and `ellipse` outlines, `M ... C ...` paths, arrowhead polygons, `text` elements with font attributes, lower-case colors and black by default; verify the SVG scenarios and a test that reads the output with the patterns of gopuml's `SvgResult` (`<svg width="Npt"`, `<title>`, `points="`, `d="`, `stroke="#..."`)
- [x] 7.3 Implement the `plain` writer as documented by Graphviz; verify the plain scenario and the `plain` layer of the curated cases
- [x] 7.4 Write the outputs of several graphs one after another; verify the two-graph scenario for `svg`, `plain` and `canon`

## 8. Library API

- [x] 8.1 Add the root-package function that takes a context, DOT text and a format name and returns the output, the warnings (object, feature, default) and an error carrying the line and token of a syntax error; verify the library-api scenarios, including byte-identical output with the command line
- [x] 8.2 Verify concurrent use and cancellation: a test with 100 goroutines under `-race` that compares their outputs, and a canceled-context test
- [x] 8.3 Keep the library self-contained: a test that the library packages import neither `os/exec`, `net` nor file or environment access, and a `GOOS=js GOARCH=wasm` build run under Node.js; verify the wasm output of `digraph { a -> b }` is byte-identical to the native one

## 9. Command line

- [x] 9.1 Implement `cmd/dot`: input from files or stdin, `-T` with `svg`, `plain` and `canon`, `dot` as default format, "not supported yet" for the other v1 formats, the reference's message for unknown formats, `-o` joined or separate as observed on the reference, `-Kdot`, other `-K` values and other flags rejected with exit status 1, `Warning:` and `Error:` prefixes; verify the cli scenarios as tests
- [x] 9.2 Implement `-V` with `dot - graphviz version 16.1.0 (gogv VERSION)` on stderr, `VERSION` taken from the build information; verify the version scenario and that gopuml's `NewGraphvizVersion` reads 16.1 from the line
- [x] 9.3 Add CLI conformance cases that run the reference `dot` and `cmd/dot` with the same arguments and compare exit status, stdout and the first stderr line; verify them in the conformance run

## 10. Determinism and wrap-up

- [x] 10.1 Render the curated corpus on darwin/arm64, with `GOARCH=amd64` and in the wasm build, and compare the outputs; verify they are byte-identical
- [x] 10.2 Run the whole conformance suite with the curated and the harvested corpora and commit `conformance-report.md`; verify that no curated case of this slice is failed and that harvested cases using later slices are reported as not supported yet
- [x] 10.3 Write the README: what gogv is, the clean-room sources, build and install, the supported subset and formats, the roadmap of v1, the conformance status; verify that `go run ./tools/provenance` and the full test suite pass
