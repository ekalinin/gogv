# Design

## Context

See proposal.md for the motivation and the scope, and the delta specs for the behavior.

Current state:

- `internal/attrs` supports `style` with one value per kind of object (`solid`, `rounded`, `invis`), and warns for `fillcolor`, `penwidth`, `weight`, `tooltip`, `labeltooltip`, `id` and `shape=box3d`.
- The ranking already carries a weight per rank edge, set to 1. The x positioning pulls the ends of each ranked edge with Ω times the number of edges that it stands for (`omegaTable`, `e.count`).
- `arrowAtEnd` stops an edge short of its node by `arrowMiter`, which is half a point divided by the sine of the half angle of the arrowhead: the miter of a 1 point pen.
- `svg.go` numbers edge groups in the order in which it draws them, and node groups in node order. It writes no `a` elements.

Facts observed that shape the approach:

- **pprof's DOT** (`internal/graph/dotgraph.go` of pprof, in the Go distribution):
  - Nodes get `style=filled` from `node [style=filled fillcolor="#f8f8f8"]`, plus `fillcolor`, `color`, `fontsize`, `shape=box`, `tooltip` and `id="nodeN"`.
  - Tag nodes are `shape=box3d` with `id="N1_0"`.
  - Edges get `weight` (1 to 100), `penwidth` (1 to 6), `color`, `tooltip`, `labeltooltip`, `style="dotted"` for residual edges and for edges to cumulative numeric tags, and `minlen=2`.
  - `bold`, `peripheries`, `URL` and `target` exist only for callers of the library.
- **pprof's Graph view** (`-http`) finds nodes by the group `id` `node<N>` and highlights the first `polygon` inside by changing its fill.
- **The reference's SVG:**
  - Fill: `fillcolor`, else `color`, else `lightgrey`. `fillcolor` without `filled` fills nothing, and a filled `plaintext` node is a `polygon` without a stroke.
  - `box3d`: a 6-corner `polygon` and three `polyline` elements, with a depth of 4 points whatever the size, sized as a box.
  - Dotted edges have `stroke-dasharray="1,5"` on the path only. Pen widths are the `stroke-width` of the path and of the arrowheads.
  - With a pen of width `p`, the tip of the arrowhead stops `(p/2)/sin(θ)` short of the node: 1.51 points more per point of pen. Above 4 points, the arrowhead widens by `p/4`.
  - Tooltips wrap the node's outline and label, or the edge's path and arrowheads, in `<g id="a_<id>"><a xlink:title="...">`. A `labeltooltip` wraps the edge's label in `<g id="a_<id>-label">`.
  - Node groups take the `id` attribute. Edge groups are `edge<N>` in the order the edges are read: `digraph { a -> c; b -> a; a -> b }` numbers `b->a` 2, while gogv numbers it 3 today.
- **The reference's other formats:**
  - xdot: a filled outline is `c ... C ... P` or `E`, and `lightgrey` is `#d3d3d3`. `box3d` adds `c ... L 2` three times. Dotted is `S 6 -dotted`, while arrowheads keep `S 5 -solid`. Pen widths are `S 15 -setlinewidth(N)` on the curve and the arrowheads.
  - plain: writes `filled` and `dotted` as the style, and the fill color as written.
- **Weight:**
  - `a -> e [weight=5]` lifts `e` to the rank right below `a`.
  - In `a -> b; a -> c [weight=10]`, `a` moves above `c`.
  - gogv does neither today.
- **The pprof corpus:** 7 DOT files from CPU and heap profiles of a small workload give 584 `tooltip` warnings, 351 `style`, 293 `labeltooltip`, 291 `id`, 291 `fillcolor`, 286 `weight`, 95 `shape` and 32 `penwidth` in gogv today. With gogv as `dot`, `go tool pprof -svg` already succeeds.

## Goals / Non-Goals

**Goals:**

- The DOT that `go tool pprof` writes laid out and drawn as the reference does, without a warning.
- The Graph view of `pprof -http` selects nodes in gogv's SVG as in the reference's.

**Non-Goals:**

- The styles, shapes and attributes that `go tool pprof` does not write, even where they are cheap: `dashed`, `bold`, `peripheries`, `URL`, `target`, `penwidth` on nodes and clusters, `id` on edges and graphs, `tooltip` on graphs and clusters.
- Raster and PDF output.

## Decisions

### D1. Sources

The clean-room rule holds. Sources added for this change:

- pprof's DOT writer, in the Go distribution: it describes the input, not Graphviz. It is read for what it writes, never for how a graph is laid out.
- the Graphviz documentation of `style`, `fillcolor`, `color`, `penwidth`, `weight`, `tooltip`, `labeltooltip`, `id`, the `box3d` shape and the color names
- Gansner, Koutsofios, North and Vo, sections 2 and 4: the weight `ω(e)` of an edge in the ranking and in the x positioning
- black-box observation of the reference

### D2. Styles as lists

`internal/attrs` checks `style` as a comma-separated list: each value must be supported for the kind of object, else the whole value is not supported and falls back to `solid`, as one unsupported value does today. The layout reads styles through one helper that returns the set of values, so `filled` and `rounded` combine. Node and edge results carry the fill color (empty when the node is not filled) and the edge style and pen width.

Alternative considered: supporting only the single values `filled` and `dotted`. Rejected: `filled,rounded` is how a filled rounded box is written, and the list costs one split.

### D3. Weight in the ranking and the positioning

Each rank edge takes its edge's `weight` instead of 1, as section 2 of the paper uses `ω(e)`. A ranked edge of the positioning carries the sum of the weights of the edges it stands for. The pull of its auxiliary edges becomes Ω times that sum, as section 4 of the paper does, instead of Ω times the number of edges. The number of edges still sizes the separations of virtual nodes. An edge of weight 0 does not pull.

### D4. Pen width at the arrowheads

`arrowAtEnd` takes the pen width: the miter becomes `(p/2)·hypot(length, halfwidth)/halfwidth`, with the half width scaled by `p/4` above 4 points, which is today's constant for `p = 1`. `Arrow` carries the pen width, so that `ArrowPolygon` widens the arrowhead in the same way for every output.

### D5. `box3d` as a geometry helper

`internal/layout` gives the `box3d` outline, its polygon and its three depth lines, from the node's size, next to `PolygonVertices` and `RoundedOutline`, so that `svg` and xdot draw the same points. The node is sized as a box.

### D6. SVG structure

`svg.go` gives node groups the `id` attribute when set, and edge groups their place in edge order. A tooltip opens `<g id="a_<id>"><a xlink:title="...">` around what the reference wraps. Tooltip texts are escaped as titles are.

Fills, dash arrays and stroke widths go on the elements that the reference puts them on. The xdot writer follows `svg.go` element by element, as it does today, with the same fill, dash and width decisions. JSON follows xdot.

### D7. The pprof corpus

`tools/pprofcorpus` builds a small workload, embedded in the tool:

- hashing, sorting and string building under `pprof.Do` labels;
- allocations of three sizes.

It runs the workload with CPU and memory profiling, and writes the DOT of `go tool pprof -dot` for 7 option sets into `testdata/conformance/cases/pprof`:

- CPU and heap profiles;
- `-nodecount`, `-tagfocus`, `-focus` and `-diff_base`.

Profiles are sampled, so each run writes different files. The committed files are the corpus, and the tool shows how to make more. The cases go through the curated contract (`canon`, `plain`, `svg`, `dot`, `xdot` and `json`), recorded with `make reference`.

Alternative considered: harvesting into the cache, as for gopuml. Rejected: the profiles change with each run, so a committed corpus is the only one that two machines share.

## Risks / Trade-offs

- [Weights change the layout of every graph that sets them] → Graphs without `weight` keep their ranks and positions, since every weight is then 1. The conformance run shows the curated and gopuml cases unchanged.
- [Edge groups get new ids in svg] → gopuml reads node titles and colors, not edge ids, and the contract compares rasterized SVG. The ids now match the reference's.
- [Calibration of the arrowhead with pen widths rests on four probes] → The model is the miter of a stroked triangle and matches all four to 0.1 point. The pprof corpus, with pen widths 1 to 6, checks it at scale.
- [Profiles are machine-dependent: function names, times and file names end up in the corpus] → They are inputs, committed once. They hold no secrets: the workload, the Go runtime and the standard library.

## Migration Plan

No migration: the change only adds behavior. The README's Status and "Not supported yet" lists change with it.
