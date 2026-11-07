## ADDED Requirements

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

## MODIFIED Requirements

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

- **WHEN** the conformance run includes the gopuml corpus and its variant with clusters, and 22 of their cases do not pass the proxy
- **THEN** the report gives the number of cases that pass the proxy, and lists the 22 others with their largest deviation, while their outcomes stay those of the layered similarity contract
