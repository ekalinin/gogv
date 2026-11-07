# upstream-compatibility Specification

## Purpose

Defines how closely this implementation follows Graphviz 16.1.0's `dot` built without Pango, and how that closeness is measured and reported.

## Requirements

### Requirement: Pinned reference

Compatibility SHALL be measured against Graphviz 16.1.0, built from its release sources without Pango and run in a container image whose base image, Graphviz sources and SVG rasterizer (`resvg`) are pinned by digest or checksum. References SHALL be regenerated only when a pinned version changes.

#### Scenario: Reference without Pango

- **WHEN** the reference image runs `dot -V` and lays out a node labeled `Hello world`
- **THEN** the version is 16.1.0, and the font of the label is resolved without a text layout plugin

### Requirement: Layered similarity contract

For every case of the corpus, the outputs SHALL be compared with the reference's outputs for the same input:

- `canon`: byte-identical
- `plain`: the same node names and the same edges (tail and head), and graph width and height each within 5% of the reference's
- `svg`: width and height each within 5% of the reference's, and a structural similarity (SSIM, grayscale) of at least 0.90 between both documents rasterized by the pinned `resvg` at the reference's size, with the Liberation 2.1.5 fonts as the only fonts
- `svg_inline`: the comparison of `svg`, and the same lines as the reference's before the root `svg` element, but for the version of the generator
- `plain-ext`: the comparison of `plain`, and the same ports on the edge lines
- `dot`, `gv`, `xdot`, `xdot1.2`, `xdot1.4`, `json` and `json0`: both outputs parsed, with the same graphs, subgraphs, nodes and edges in the same order, the same attribute names on each, and the same values for the attributes that do not depend on the layout; the same drawing operations, in the same order, with the same colors, styles, fonts and texts; and the width and height of the graph's `bb` each within 5% of the reference's. The numbers of the layout are not compared one by one: `plain` and `svg` judge the layout
- `dot_json` and `xdot_json`: byte-identical

A case passes when the comparisons of all its recorded formats pass. A conformance run without the Liberation fonts SHALL fail before comparing any case, and name the command that fetches them.

#### Scenario: Similar layout

- **WHEN** a case has the reference's `canon`, the same nodes and edges in `plain` with a size 3% larger, and an SVG at SSIM 0.94 with a size 3% larger
- **THEN** the case passes

#### Scenario: Different order within a rank

- **WHEN** a case orders the nodes of a rank differently from the reference, and its SVG reaches SSIM 0.82
- **THEN** the case fails

#### Scenario: Layout numbers in dot

- **WHEN** a case has the reference's objects and attributes in `dot`, a `pos` of a node 2 points away from the reference's, and a `bb` 3% larger
- **THEN** its `dot` output passes

#### Scenario: Missing drawing operation

- **WHEN** the `xdot` output of a case has no `_hdraw_` for an edge whose reference has one
- **THEN** its `xdot` output fails

#### Scenario: Fonts not fetched

- **WHEN** the conformance run starts on a clone where `resvg` is fetched and the Liberation fonts are not
- **THEN** the run fails before comparing any case, names `go generate ./internal/textmetrics`, and writes no report

### Requirement: Corpus

The corpus SHALL include:

- curated cases, one DOT file per case grouped by capability, committed with their reference outputs as text: `canon`, `plain`, `svg`, `dot`, `xdot` and `json` for every case, and every format of v1 for the cases of `output-formats`
- the DOT that gopuml generates for its conformance cases, harvested into a cache and not committed, with the reference outputs of `canon`, `plain` and `svg`
- the DOT that `go tool pprof` writes for profiles of a workload of the repository, committed as the curated cases of `pprof`, with the tool that profiles the workload and writes it again

The gopuml corpus SHALL be harvested from the released version of gopuml pinned in the repository, fetched by its module path and verified by the Go checksum database, without a gopuml checkout. A local gopuml checkout SHALL be used only when it is named explicitly. The harvested corpus SHALL record its source: the pinned version, or the commit of the local checkout marked as local, and as dirty when the checkout has uncommitted changes.

#### Scenario: Harvested corpus is not committed

- **WHEN** the harvested gopuml corpus is collected
- **THEN** it is stored in the cache directory, which is ignored by version control

#### Scenario: Formats of a curated case

- **WHEN** the references of the curated cases are recorded
- **THEN** a case of `dot-layout` has references of `canon`, `plain`, `svg`, `dot`, `xdot` and `json`, and a case of `output-formats` has references of the 14 formats of v1

#### Scenario: Pinned gopuml version

- **WHEN** the gopuml corpus is harvested on a machine with no gopuml checkout, without naming one
- **THEN** gopuml v0.3.0 is fetched by its module path, the corpus holds the DOT of its curated and Vega cases, and the corpus records `gopuml v0.3.0` as its source

#### Scenario: Same corpus from the release

- **WHEN** the gopuml corpus is harvested from v0.3.0
- **THEN** its DOT files are byte-identical to those harvested from gopuml commit `becdd08`, which v0.3.0 contains with the same cases

#### Scenario: Local checkout with changes

- **WHEN** the gopuml corpus is harvested from a local gopuml checkout named explicitly, which has an uncommitted change
- **THEN** the corpus holds the DOT generated by that working tree and records the checkout's commit as its source, marked as local and dirty

#### Scenario: Corpus of pprof

- **WHEN** the tool that writes the pprof corpus runs
- **THEN** it writes, under the curated cases of `pprof`, the DOT that `go tool pprof -dot` writes for CPU and heap profiles of its workload, with and without `-nodecount`, `-tagfocus`, `-focus` and `-diff_base`

### Requirement: Case outcomes and report

Each case SHALL have one outcome:

- passed
- known difference: listed in a known-differences file with a reason and the follow-up change that resolves it
- not supported yet: the run of the case emitted a "not supported yet" warning or error
- failed: any other case that does not pass

The conformance run SHALL fail when a case is failed, and SHALL NOT fail for known differences and cases not supported yet. It SHALL write a report with the number of cases of each outcome per capability and per format. The report SHALL name the corpora of the run, with the number of cases of each and the recorded source of each harvested corpus. When the run includes a harvested gopuml corpus, the report SHALL give the number of its cases that pass the proxy of gopuml's contract, and list the others with their largest deviation.

#### Scenario: Feature of a later slice

- **WHEN** a case uses a compass point on a node ID, as in `a:n -> b`, before compass points are supported
- **THEN** the case is reported as not supported yet, and the run does not fail because of it

#### Scenario: Unlisted failure

- **WHEN** a case fails and is not listed in the known-differences file
- **THEN** the conformance run fails and the report names the case

#### Scenario: Source of the harvested corpus

- **WHEN** the conformance run includes the gopuml corpus harvested from v0.3.0 and its variant with clusters
- **THEN** the report names the curated corpus, the gopuml corpus and the variant with their numbers of cases, and `gopuml v0.3.0` as the source of both harvested corpora

#### Scenario: Harvested corpus not collected

- **WHEN** the conformance run finds no harvested gopuml corpus
- **THEN** the report names only the curated corpus and the command line cases

#### Scenario: Proxy in the report

- **WHEN** the conformance run includes the gopuml corpus and its variant with clusters, and 23 of their cases do not pass the proxy
- **THEN** the report gives the number of cases that pass the proxy, and lists the 23 others with their largest deviation, while their outcomes stay those of the layered similarity contract

### Requirement: Proxy of gopuml's contract

For every case of the harvested gopuml corpora, the conformance run SHALL judge a proxy of gopuml's contract: whether gopuml's drawing of the case keeps the reference's size within 5% and its arrangement. A case passes the proxy when, compared with the reference's outputs for the same input:

- in `plain`, the graph width and height are each within 2% of the reference's;
- in `plain`, the same nodes and the same edge labels are present, and the center of each node and of each edge label is within 2% of the reference's graph width on the x axis and within 2% of its graph height on the y axis; edges are matched by tail, head and their order between the same nodes;
- the `svg` reaches an SSIM of at least 0.90, as in the layered similarity contract.

The largest deviation of a case is its largest difference among these sizes and centers, as a fraction of the reference's graph size. The proxy SHALL NOT change the outcome of any case, and the conformance run SHALL NOT fail because of it.

#### Scenario: Close layout

- **WHEN** a case of the gopuml corpus has a `plain` output 1% wider than the reference's, every node and label center within 1% of the reference's graph size, and an SVG at SSIM 0.95
- **THEN** the case passes the proxy

#### Scenario: Label moved

- **WHEN** a case of the gopuml corpus has the reference's graph size, and the center of one edge label is 3% of the graph width away from the reference's
- **THEN** the case does not pass the proxy, with a largest deviation of 3%, and its outcome is the outcome of the layered similarity contract

### Requirement: Reduction of a layout difference

A tool SHALL reduce a DOT input whose layout by this implementation does not pass the proxy against the reference. It SHALL remove statements, attributes and subgraphs, keeping their contents when a subgraph is removed. It SHALL keep a removal while the layout of the smaller input still does not pass the `plain` part of the proxy, on a node or an edge label that does not pass it in the original input, and while the input gives no warning. It SHALL write the smallest input it finds, as DOT. It SHALL run the reference as a black box, as the other conformance tools do, and record its outputs in the cache.

#### Scenario: Unrelated statements removed

- **WHEN** the tool reduces an input whose difference lies in two nodes and an edge between them, and which has ten other nodes and edges that the difference does not depend on
- **THEN** the input it writes still does not pass the proxy on those nodes, and it has none of the ten other nodes and edges

#### Scenario: Another difference

- **WHEN** removing a statement leaves a layout that does not pass the proxy only on nodes and labels that pass it in the original input
- **THEN** the tool keeps the statement
