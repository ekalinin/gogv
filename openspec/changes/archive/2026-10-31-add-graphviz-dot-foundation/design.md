# Design

## Context

See proposal.md for the motivation and the scope of this slice.

Facts that shape the approach:

- gopuml talks to a layout engine through `dot.Engine` (`gopuml/internal/dot/graphviz.go`): `Layout(ctx, dot string) ([]byte, error)` returns the SVG document that `dot -Tsvg` writes, and implementations must be safe for concurrent use. svek reads node positions, cluster boxes and edge paths back from that SVG by `<title>`, `points="..."`, `d="M...C..."` and the unique `stroke` and `fill` colors it gave each object.
- svek sends no text to `dot`: nodes have `label=""` or HTML tables of empty, fixed-size cells. Text width matters only for standalone use.
- gopuml's reference `dot` (Graphviz 14.0.1 in `plantuml-go-reference`) has no text layout plugin: `fontname: "Times-Roman" resolved to: [internal hard-coded]`. With Pango (Homebrew `dot` 15.0.0), the same "Hello world" node is 55.45 points wide in radius instead of 57.44.
- gopuml's parity tolerance is: width and height within 5% of the reference, and SSIM of at least 0.90 between the rasterized outputs (`gopuml/openspec/specs/upstream-compatibility/spec.md`).
- gopuml requires byte-identical output across runs, machines and operating systems (`gopuml/openspec/specs/output-formats/spec.md`). Go may fuse `a*b+c` into one multiply-add on arm64, and the results of Go's `math` package are not identical across architectures (gopuml design D17).
- No conformance case of gopuml uses `linetype ortho` or `polyline`.

## Goals / Non-Goals

**Goals:**

- A `dot` whose layouts look like those of Graphviz 16.1.0, measured automatically against that reference.
- An implementation whose provenance is papers, documentation and observed behavior only.
- The same output bytes on every platform.
- An architecture where each stage can be checked against the reference on its own.

**Non-Goals:**

- Byte-identical layouts with Graphviz. The tie-breaking of its heuristics is not documented, so a clean-room implementation cannot reproduce it exactly.
- Text layout with Pango, HarfBuzz or system fonts.
- Anything listed as out of scope in proposal.md.

## Decisions

### D1. Clean-room from papers, documentation and observed behavior

Allowed sources:

- papers, first of all Gansner, Koutsofios, North and Vo, "A Technique for Drawing Directed Graphs" (IEEE TSE, 1993), for ranking with network simplex, crossing minimization and positioning; and Dobkin, Gansner, Koutsofios and North, "Implementing a General-Purpose Edge Router" (Graph Drawing 1997), for spline routing
- the Graphviz documentation: the DOT language, attributes, node shapes, arrow shapes, output formats, "Drawing graphs with dot"
- the observed behavior of the reference `dot`: outputs for chosen inputs, error messages, exit codes

Not allowed: the source code of Graphviz and of its ports and bindings (Smetana in PlantUML, graphviz-pure, goccy/go-graphviz and its forks). gopuml's own code is allowed: it is a port of PlantUML, not of Graphviz.

Each Go file names the sources it is based on (paper and section, documentation page, observed behavior) in a header, as gopuml names its upstream files.

Alternatives considered:

- A line-by-line port of the C sources of Graphviz 16.1.0: exact parity is reachable (graphviz-pure reached it in Rust), but the module would be EPL-2.0. Rejected by the user.
- A port of Smetana: Graphviz 2.38.0 from 2014, with no DOT parser, no SVG output and no `ortho`, and a derived work of Graphviz.
- Wrapping Graphviz compiled to WebAssembly: already exists (go-graphviz) and is not a Go implementation.

### D2. Reference: Graphviz 16.1.0 without Pango, as a black box

A pinned container image (base image by digest, Graphviz 16.1.0 built from its release sources without Pango, a pinned `resvg` for rasterizing) records the reference outputs. 16.1.0 is the latest stable release, and gopuml moves its own reference to it. Without Pango, text width does not depend on installed fonts, so references are reproducible, and the comparison measures layout rather than font differences.

### D3. Layered comparison and the similarity contract

Each stage of the pipeline has an output format that exposes it, and the harness compares each one with the reference:

```
 DOT --> parse  --> canon   exact text equality
           |
           v
        layout  --> plain   same nodes and edges; graph width and
           |                height within 5%
           v
        render  --> svg     width and height within 5%;
                            SSIM >= 0.90 after rasterizing both with resvg
```

The thresholds are gopuml's, so that a graph that passes here is likely to pass in gopuml. A difference first visible in `plain` is a layout problem, one visible only in `svg` is a rendering problem. Cases that do not meet the contract are listed in a known-differences file with a reason, as in gopuml.

Alternatives considered:

- Byte-identical output: not reachable with a clean-room implementation (D1).
- Quality properties only ("not worse than `dot`": rank constraints kept, no overlaps, comparable crossings and area): rejected, because gopuml needs layouts similar to `dot`'s to switch by default.

### D4. Corpus

- Curated cases, one DOT file per case, grouped by capability, committed with their reference outputs as text.
- The DOT that gopuml's svek generates for its conformance cases, harvested into a cache and not committed (gopuml does not commit the upstream inputs either). It is the acceptance corpus for gopuml's switch and grows as slices add clusters, edge labels and tables.

### D5. Text width from free metric-compatible fonts

A generator tool extracts the advance widths of Liberation 2.x (Serif, Sans and Mono, metrically compatible with Times New Roman, Arial and Courier New; regular, bold, italic, bold italic) into Go tables. Liberation 2.x is under the SIL Open Font License 1.1 and is a fork of the ChromeOS core fonts (Tinos, Arimo, Cousine), whose license is reported differently by different sources. The font files are downloaded by checksum into the cache; only the tables are committed, with the OFL notice in NOTICE. The `fontname` of a label selects a family, and `fontsize` scales the widths.

Observing the reference showed how its built-in metrics work, and gogv follows it:

- the Times metrics of the reference are those of Liberation Serif, to the hundredth of a point
- it recognizes only whole family names, without regard to case (the list is in the dot-layout spec); a name with a weight or a slant (`Helvetica-Bold`) falls back to Times, so in v1 only the regular tables are used, and the bold and italic tables wait for HTML text markup
- every byte of a non-ASCII character is as wide as a space, so only Basic Latin is tabulated
- it also has metrics for `DejaVu Sans`, `Verdana`, `Georgia`, `Consolas`, `Calibri`, `Trebuchet MS` and `Open Sans`, which Liberation does not cover; gogv reports them as not supported yet and uses Times, and a later change can add metrics from free fonts (DejaVu Sans and Open Sans are free, Carlito matches Calibri, Gelasio matches Georgia)

The reference is measured as a black box (strings of one repeated character per family and style) only to check how close the tables are. The measured data is not shipped.

Alternatives considered:

- Tables measured from the reference: closest to it, but derived from Graphviz's behavior. Rejected in favor of an independent source.
- An average character width times the font size: too far from the reference for the similarity contract.
- System fonts: output would depend on the machine (D6).

### D6. Deterministic arithmetic and ordering

- Products that feed an addition are written `float64(a*b)` in layout and rendering code: an explicit conversion forbids fused multiply-add.
- Elementary functions (`sin`, `cos`, `atan2`, `pow`, `exp`, `log`, ...) come from an internal port of fdlibm, taken from gopuml's `internal/jmath`. fdlibm is under Sun's permissive license and is not part of Graphviz, so D1 allows it. Its files keep the notice.
- Sorting that affects output is stable, and no output depends on the iteration order of a Go map. Attributes keep their input order.
- Numbers are written with one formatting routine per output format.

Alternative considered: Go's `math` package, which gives different bits on different architectures.

### D7. No global state

All state of a run lives in a value created for that run. The library holds no mutable package-level state, so it is safe for concurrent use, as gopuml's `dot.Engine` requires. Long layouts check the context and stop when it is canceled.

### D8. Pipeline and packages

```
 DOT text --> parse --> graph model --> layout --> laid-out graph --> render --> bytes
                                        rank
                                        order
                                        position
                                        route
```

The public API lives in the root package: DOT text and an output format in, bytes out, with a context. A gopuml `dot.Engine` is then a few lines. The stages live in internal packages, one per stage, so each can be tested against the layer of D3 that exposes it. A graph-building API (nodes and edges created from Go code) can be added later without changing the stages.

### D9. Module, command and reused tooling

- The module is `github.com/ekalinin/gogv`, named like gopuml: Go plus the Graphviz file extension `.gv`. It targets Go 1.26, as gopuml does, so that gopuml can import it without raising its own Go version.
- The command is built from `cmd/dot`, so that its binary can replace Graphviz's `dot`.
- From gopuml, which is MIT and not derived from Graphviz, the project takes the fdlibm port (D6), the SSIM comparison and the `resvg` driver of its conformance harness, and the source header check of `tools/provenance`, adapted to the headers of D1.

### D10. Delivery in vertical slices

v1 is delivered as five changes (proposal.md, "Roadmap of v1"). The similarity contract compares rasterized SVG, so every slice needs the whole path from DOT to SVG to be accepted. Cutting by layers instead (parser, then layout, then rendering) would leave the first changes with nothing to accept but `canon`.

## Risks / Trade-offs

- [The order of nodes within ranks differs from `dot`'s on some graphs, and SSIM falls below 0.90] → Study `dot`'s choices on small graphs as a black box, list the remaining cases in known differences, and let gopuml switch only once its own conformance passes.
- [Clean-room with an AI assistant is relative: the models were trained on public code, including Graphviz] → Every implementation session gets the list of allowed sources (D1). No code of Graphviz or its ports is fetched or quoted, and every file names its sources.
- [This exploration read fragments of Graphviz's `lib/common/splines.c` (`place_portlabel`) and of gopuml's svek] → The one fact taken from them, that head and tail labels are placed as external labels unless `labelangle` or `labeldistance` is set, is documented behavior that can also be observed. No code from that reading is used.
- [Building Graphviz 16.1.0 from source under amd64 emulation on arm64 hosts crashed at random in gopuml] → The reference image is built for linux/amd64 on the arm64 host, under the Rosetta emulation of colima, which built it without a crash; an amd64 runner is not needed. `canon` does not depend on the architecture, and small floating-point differences of the reference stay well within the tolerance.
- [Widths from the free fonts differ from the reference's built-in widths] → The black-box check of D5 reports the differences per family and style. Differences that keep cases within the contract are accepted.
- [gopuml emits DOT outside the v1 subset in cases that the exploration did not see] → The harvested corpus (D4) covers every gopuml conformance case, so gaps show up as failures in this project before gopuml switches.

## Migration Plan

This is a new module, so nothing migrates. The path for gopuml, in its own changes:

1. Move gopuml's reference container to Graphviz 16.1.0 built without Pango and regenerate its references. This does not depend on this project and can come first.
2. After slice 5, add an engine based on this module and make it the default once gopuml's conformance passes with it. For `splines=ortho`, use the external `dot` when one is found and report "not supported yet" otherwise, always in the WebAssembly build.

Rollback: gopuml keeps its external `dot` engine, so going back is a change of the default engine.

