# Proposal

## Why

`go tool pprof` draws its call graphs with the external `dot`: `-svg`, `-web` and the Graph view of `-http` run `dot -Tsvg`, and fail with "Is Graphviz installed?" without it. With gogv installed as `dot`, they already work, but the drawing is poorer than the reference's, and the Graph view of `-http` cannot select nodes:

- nodes are not filled with their colors;
- hot edges are not thicker;
- heavy edges are not kept straight;
- residual edges are not dotted;
- tag nodes are plain boxes;
- there are no tooltips;
- node groups do not take the `id` that pprof gives them, and the Graph view finds nodes by it.

This is the first slice of the work that makes gogv a standalone replacement for `dot`, with pprof as its first consumer, as gopuml was for v1.

## What Changes

- Attributes that pprof writes become supported:
  - on nodes: `style=filled`, `fillcolor`, `shape=box3d`, `tooltip`, `id`;
  - on edges: `style=dotted`, `penwidth`, `weight`, `tooltip`, `labeltooltip`.

  `style` takes a comma-separated list of supported styles (`filled,rounded`). `style=filled` is supported on every supported node shape.
- Layout:
  - `weight` weighs edges in the ranking and in the x positioning, as in Gansner et al.;
  - `box3d` nodes are sized as boxes.
- `svg` output:
  - fills, with the reference's choice of color: `fillcolor`, else `color`, else `lightgrey`;
  - the outline of `box3d`;
  - dotted edges and wider pens, with their arrowheads;
  - tooltips as the reference's `a` elements;
  - node groups with their `id`;
  - edge groups numbered in edge order, as the reference numbers them.
- `plain`, xdot and JSON output: the styles, fills, pen widths and outlines of the drawing, as the reference writes them.
- Conformance: a corpus of DOT written by `go tool pprof` for profiles of a small workload, committed as curated cases with a tool that produces it again.

Out of scope:
- what pprof's library allows but `go tool pprof` does not write: `bold`, `peripheries`, `URL` and `target`;
- these attributes on other objects, such as `penwidth` on nodes and clusters, `tooltip` on graphs and clusters, and `id` on edges;
- the other styles, such as `dashed` and `bold`;
- the `png`, `gif`, `pdf` and `ps` formats, which pprof also asks for.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `dot-language`: the attributes and values that pprof writes become supported, and `style` takes lists.
- `dot-layout`: `weight` in the ranking and the positioning; the size of `box3d` nodes.
- `output-formats`: fills, `box3d`, dotted edges, pen widths, tooltips and ids in `svg`, and the same drawing in `plain`, xdot and JSON.
- `upstream-compatibility`: the corpus includes the DOT that `go tool pprof` writes.

## Impact

- Code:
  - `internal/attrs`: the supported attributes and values;
  - `internal/layout`: weights, `box3d`, and the fill, style and pen width of nodes and edges in `layout.Result`;
  - `internal/output`: `svg`, `plain`, xdot and JSON.
- Tools: `tools/pprofcorpus`, which builds a workload, profiles it and writes the DOT of `go tool pprof`.
- Conformance: curated cases under `testdata/conformance/cases/pprof`, with their references.
- `README.md`: the Status and "Not supported yet" lists.
- No change to the library API or the command line.
