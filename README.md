# gogv

A Go implementation of the Graphviz `dot` program: it lays out graphs written
in the DOT language and writes them as SVG and in the other output formats of
`dot`. It needs no cgo, no embedded WebAssembly and no installed Graphviz, and
it works as:

- a Go library (`gogv.Render`), safe for concurrent use
- a `dot` command line
- a WebAssembly module (`GOOS=js GOARCH=wasm`)

![The stages of gogv, laid out by gogv](docs/example.svg)

The picture is gogv's SVG output for [`docs/example.gv`](docs/example.gv).

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

gogv supports:

- the whole DOT grammar, read as the reference reads it
- the layout of graphs without compass points and HTML-like labels other
  than table labels:
  shapes `rect`, `box`, `rectangle`, `plaintext`, `point`, `ellipse`, `oval`,
  `circle`, `diamond`, `hexagon`, `octagon`, and `style=rounded` on `rect`;
  plain-text labels; `width`, `height`, `fontname`, `fontsize`, and `color`
  and `fontcolor` written `#rrggbb`; `minlen`,
  `constraint`, `style=invis`, `arrowhead` and `arrowtail` (`normal`, `empty`,
  `none`), `dir`, `arrowsize`; `nodesep`, `ranksep`, `rankdir=LR`,
  `splines=spline|polyline`, a graph label placed by `labeljust`
- clusters, nested, with a plain-text label placed by `labeljust` (`c`, `l`,
  `r`), `style=solid` and `color`; `rank=same|min|max|source|sink` on
  subgraphs
- edge labels, on edges between ranks, flat edges and loops, in the edge's
  `fontname`, `fontsize` and `fontcolor`; external labels: `xlabel` on nodes
  and edges, `headlabel` and `taillabel`; `sametail`
- table labels (`TABLE` of rows of empty cells) on nodes, edges and
  clusters, with `BORDER`, `CELLBORDER`, `CELLSPACING`, `CELLPADDING`,
  `BGCOLOR`, `COLOR`, `FIXEDSIZE`, `WIDTH`, `HEIGHT` and `PORT` on the table,
  and `BORDER`, `CELLPADDING`, `BGCOLOR`, `COLOR`, `COLSPAN`, `FIXEDSIZE`,
  `WIDTH`, `HEIGHT` and `PORT` on its cells
- node ports: edge ends written `node:port`, or set by `tailport` and
  `headport`, attach to the cell of a table label with that `PORT`, in the
  positioning, the order and the routes
- the output formats: `svg`, `svg_inline`, `plain`, `plain-ext`, `canon`,
  `dot`, `gv`, `xdot`, `xdot1.2`, `xdot1.4`, `json`, `json0`, `dot_json` and
  `xdot_json`

Not supported yet, among others:

- color names such as `red` (colors are written `#rrggbb`), `fillcolor`,
  `bgcolor`, `penwidth`
- styles other than `solid`, `invis` on edges and `rounded` on `rect`, such
  as `filled` and `dashed`
- other shapes, such as `record`, `Mrecord` and `box3d`, and other arrow
  types, such as `vee`
- `rankdir=BT` and `rankdir=RL`; `splines` other than `spline` and
  `polyline`, such as `ortho`
- compass points (`a:n -> b`), `compound` with `lhead` and `ltail`,
  `concentrate`
- HTML-like labels other than tables of empty cells: text in cells, and text
  markup such as `<B>` and `<FONT>`
- fonts other than the Times, Helvetica and Courier families (including
  Times New Roman, Arial and Courier New), which are measured as Times
- `URL`, `href`, `tooltip` and images
- the other layout engines (`-Kneato`, `-Kfdp`, ...) and the other output
  formats, such as `png`, `pdf` and `cmapx`

A Graphviz feature that is not supported yet gives a warning that says so and
names the default used instead; attributes that Graphviz does not document
are ignored, as Graphviz ignores them.

## Installation

Go 1.26 or later is needed. The command line:

```sh
go install github.com/ekalinin/gogv/cmd/dot@latest
```

The command is named `dot`, as Graphviz's is: when Graphviz is installed
too, the `dot` found first in `PATH` runs. `dot -V` tells them apart (see
[Command line](#command-line)).

The library:

```sh
go get github.com/ekalinin/gogv
```

## Building

```sh
go build -o dot ./cmd/dot
echo 'digraph { a -> b -> c; a -> c }' | ./dot -Tsvg > out.svg
./dot -V
```

## Library

`gogv.Render` lays out every graph of DOT text and writes it in one of
`gogv.Formats()`, with the bytes that `dot -T<format>` writes:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ekalinin/gogv"
)

func main() {
	out, warnings, err := gogv.Render(context.Background(), "digraph { a -> b }", "svg")
	if err != nil {
		log.Fatal(err)
	}
	for _, w := range warnings {
		fmt.Fprintln(os.Stderr, "warning:", w)
	}
	os.Stdout.Write(out)
}
```

The warnings are returned, not written: a `gogv.Warning` names the object,
the feature that is not supported yet and the default used instead. A syntax
error is a `*gogv.SyntaxError` with its line, and a format that `Render` does
not write is a `*gogv.FormatError`. `Render` is safe for concurrent use and
stops with the context's error when the context is canceled.

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
of at least 0.90). `dot`, xdot and JSON output are parsed and must have the
reference's objects, attributes and drawing operations, with values equal but
for the numbers of the layout, and a size within 5%; `dot_json` and
`xdot_json` are compared byte for byte. The corpus is the curated cases of
`testdata/conformance/cases` and, when collected with `make harvest`, the DOT
that gopuml generates for its own conformance cases.

`make harvest` fetches gopuml at the version `GOPUML_VERSION` of the
`Makefile` through the Go module proxy, and the Vega cases of its corpus with
gopuml's own script, so it needs network access as well as Docker and the
reference image; no gopuml checkout is needed. `make harvest GOPUML=<dir>`
harvests from a local gopuml checkout instead. The report names the corpora
of the run and the gopuml that each harvested corpus comes from.

```sh
make reference-image      # build the reference image (Docker)
make reference            # record the references of the curated cases
make harvest              # collect the gopuml corpus and record its references
tools/reference/fetch-tools.sh
go generate ./internal/textmetrics  # fetch the Liberation fonts
make conformance          # compare and write conformance-report.md
make determinism          # compare native, amd64 and WebAssembly outputs
```

`conformance-report.md` has the current results. Cases that use a feature
that is not supported yet are reported as such; the remaining differences
are listed in `testdata/conformance/known-differences.yaml`.

## Contributing

Contributions follow the clean-room rules, the checks and the conventions of
[CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT. See `LICENSE` and `NOTICE`.
