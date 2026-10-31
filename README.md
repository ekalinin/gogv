# gogv

A Go implementation of the Graphviz `dot` program: it lays out graphs written
in the DOT language and writes them as SVG and in the other output formats of
`dot`. It needs no cgo, no embedded WebAssembly and no installed Graphviz, and
it works as:

- a Go library (`gogv.Render`), safe for concurrent use
- a `dot` command line
- a WebAssembly module (`GOOS=js GOARCH=wasm`)

gogv is not affiliated with or endorsed by the Graphviz project.

## Clean-room implementation

gogv is written from published papers, the Graphviz documentation and the
observed behavior of Graphviz 16.1.0, used as a black box. It contains no code
from Graphviz or from its ports and bindings. Every Go file names its sources
in `Provenance:` lines (`make provenance` checks them):

- Gansner, Koutsofios, North and Vo, "A Technique for Drawing Directed Graphs"
  (IEEE TSE, 1993): ranking, crossing minimization, positioning
- Dobkin, Gansner, Koutsofios and North, "Implementing a General-Purpose Edge
  Router" (Graph Drawing 1997): edge routing
- the Graphviz documentation (https://graphviz.org/documentation/)
- outputs of Graphviz 16.1.0 for chosen inputs

Text is measured with tables generated from the Liberation 2.1.5 fonts, which
are metrically compatible with Times New Roman, Arial and Courier New; no font
is part of the module. Elementary functions come from a port of fdlibm, so
that the output is the same, byte for byte, on every platform.

## Status

This is the third slice of version 1. It supports:

- the whole DOT grammar, read as the reference reads it
- the layout of graphs without node ports and HTML-like labels other than
  single-cell tables:
  shapes `rect`, `box`, `rectangle`, `plaintext`, `point`, `ellipse`, `oval`,
  `circle`, `diamond`, `hexagon`, `octagon`; plain-text labels; `width`,
  `height`, `color`, `fontname`, `fontsize`, `fontcolor`; `minlen`,
  `constraint`, `style=invis`, `arrowhead` and `arrowtail` (`normal`, `empty`,
  `none`), `dir`, `arrowsize`; `nodesep`, `ranksep`, `rankdir=LR`,
  `splines=spline|polyline`, a graph label placed by `labeljust`
- clusters, nested, with a plain-text label placed by `labeljust` (`c`, `l`,
  `r`), `style=solid` and `color`; `rank=same|min|max|source|sink` on
  subgraphs
- edge labels, on edges between ranks, flat edges and loops, in the edge's
  `fontname`, `fontsize` and `fontcolor`; external labels: `xlabel` on nodes
  and edges, `headlabel` and `taillabel`; `sametail`
- single-cell table labels (`TABLE` with one row of one empty cell, with
  `BGCOLOR`, `FIXEDSIZE`, `WIDTH` and `HEIGHT`) on nodes, edges and clusters
- the output formats `svg`, `plain` and `canon`

A Graphviz feature that is not supported yet gives a warning that says so and
names the default used instead; attributes that Graphviz does not document
are ignored, as Graphviz ignores them.

The next slices of version 1 add the other HTML tables and node ports, and
the formats `svg_inline`, `plain-ext`, `dot`, `gv`, `xdot*` and `json*`.

## Building

Go 1.26 or later is needed.

```sh
go build -o dot ./cmd/dot
echo 'digraph { a -> b -> c; a -> c }' | ./dot -Tsvg > out.svg
./dot -V
```

As a library:

```go
out, warnings, err := gogv.Render(ctx, "digraph { a -> b }", "svg")
```

## Command line

`dot` reads the files named on its command line, or standard input, and
writes to standard output or to the file of `-o`. It accepts `-T<format>`,
`-o<file>`, `-V` and `-Kdot`; it rejects the other flags of Graphviz's `dot`
with exit status 1 instead of ignoring them. `-V` writes
`dot - graphviz version 16.1.0 (gogv VERSION)`, so that programs that read the
Graphviz version from it find the version gogv follows.

## Conformance

The outputs are compared with those of Graphviz 16.1.0 built without Pango, in
a pinned container (`tools/reference/Dockerfile`), layer by layer: `canon`
byte for byte, `plain` by nodes, edges and size (within 5%), and `svg` by size
(within 5%) and the structural similarity of the rasterized documents (SSIM
of at least 0.90). The corpus is the curated cases of
`testdata/conformance/cases` and, when collected with `make harvest`, the DOT
that gopuml generates for its own conformance cases.

```sh
make reference-image      # build the reference image (Docker)
make reference            # record the references of the curated cases
tools/reference/fetch-tools.sh
make conformance          # compare and write conformance-report.md
make determinism          # compare native, amd64 and WebAssembly outputs
```

`conformance-report.md` has the current results. Cases that use a feature of
a later slice are reported as not supported yet; the remaining differences
are listed in `testdata/conformance/known-differences.yaml`.

## License

MIT. See `LICENSE` and `NOTICE`.
