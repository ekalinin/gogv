# Tasks

## 1. Conformance cases

- [x] 1.1 Add curated cases under `testdata/conformance/cases/dot-language` and `testdata/conformance/cases/dot-layout` for every scenario of the delta specs (cluster names in another case, shared and nested nodes, empty clusters, rank sets across clusters, the graph label received by a cluster, cluster attributes, HTML cluster labels, rank constraints inside and outside clusters, rank on the root graph and on a cluster, margins, nested and sibling boxes, edges around clusters, colors, labels with `labeljust` in `TB` and `LR`, the graph label with `labeljust`) and record them with `make reference`; verify that a second run leaves the references unchanged and that the reference exits with status 0 on every new case
- [x] 1.2 Change the "not supported yet" fixture of the conformance tests from `subgraph cluster_x` to an edge with a `label`; verify `go test ./internal/conformance/...` passes
- [x] 1.3 Add a tool that writes the gopuml graphs with clusters into `.cache/gopuml-cluster-dot/` with `label=""` in place of each HTML cluster label, record their references in `make harvest`, and add them to the conformance run as an optional corpus (design D8); verify the cache holds the 36 graphs with no HTML cluster label and that 26 of them render without a "not supported yet" warning

## 2. Language and graph model

- [x] 2.1 Recognize clusters by the prefix `cluster` without regard to case, drop the cluster and `rank` warnings from `attrs.Check`, and add `labeljust` (`c`, `l`, `r`), cluster `style` (`solid`), cluster `color` and subgraph `rank` (`same`, `min`, `max`, `source`, `sink`) to the supported attributes; verify the "Supported attributes" and "Graphviz features that are not supported yet" scenarios, including the HTML cluster label and `style=filled` on a cluster
- [x] 2.2 Resolve cluster membership on the model after `Build`, for every format (design D2): first cluster in input order, innermost cluster inside it, removal from losing clusters and the subgraphs inside them, removal of rank-set nodes from clusters that do not contain the rank subgraph with a warning; verify the "Cluster subgraphs" scenarios and that `canon` of every dot-language case with clusters is byte-identical to the reference's
- [x] 2.3 Add a test for the graph attributes that a subgraph receives from its parent; verify the "Graph label received by a cluster" scenario

## 3. Ranking

- [x] 3.1 Implement rank sets on subgraphs that are not clusters (design D3): `same` merged into one node, `min` and `max` at the ends, `source` and `sink` alone on their rank, no effect of `rank` on the root graph and on clusters, and `rank` copied from the root into later subgraphs; verify the "Rank constraints" scenarios outside clusters and a test with `digraph { rank=same; subgraph s { a -> b } }`
- [x] 3.2 Implement the recursive ranking of clusters (design D3), with rank sets that apply within their cluster; verify the "Cluster ranked as a block" and "Source inside a cluster" scenarios, and that the conformance cases without clusters and rank sets keep their outcomes

## 4. Ordering

- [x] 4.1 Order ranks top-down on the cluster tree with skeleton nodes and their expansion (design D4), with virtual nodes in the innermost cluster that contains both ends of their edge; verify a test that the nodes of every cluster are contiguous on each of its ranks for the curated cases and the gopuml graphs with clusters

## 5. Positioning

- [x] 5.1 Add left and right border nodes of clusters to the auxiliary graph (design D5): 8 points of margin, nested borders inside their parent's, neighboring nodes and sibling clusters 8 points from a box, borders at least as far apart as the label is wide; verify the "Margin", "Nested clusters", "Sibling clusters" and "Label wider than the nodes" scenarios
- [x] 5.2 Grow the gaps between ranks where clusters start or end so that their margins and labels fit; verify the "Label at the top" scenario and that the box heights of the curated cluster cases are within 5% of the reference's

## 6. Routing

- [x] 6.1 Bound edge corridors by the boxes of the clusters that the edge does not enter (design D6); verify the "Edge around a cluster" scenario and that no route of an edge with no end in a cluster crosses its box in the curated cases

## 7. Output

- [x] 7.1 Add cluster boxes and labels to the layout result and write them in `svg` (design D7): one group of class `cluster` per drawn cluster in preorder before the nodes, with the reference's ids, a `polygon` without fill in the cluster's color and the label placed by `labeljust`; place the graph label by `labeljust`; verify the "Cluster groups", "Cluster box", "Cluster color", "Label at the left" and "Graph label at the left" scenarios

## 8. Acceptance

- [x] 8.1 Run the conformance suite with the gopuml corpus and list the remaining differences in `known-differences.yaml` with their reasons under `tune-layout-heuristics`; verify `make conformance` reports no failed case, and that the 26 graphs of the gopuml cluster variant with no feature of a later slice are judged rather than reported as not supported yet
- [x] 8.2 Check determinism and hygiene for the new code: provenance headers, FMA guards, race safety; verify `go run ./tools/provenance`, the FMA assembly tests, `go test -race ./...` and `make determinism` (byte-identical on arm64, amd64 and wasm) pass
- [x] 8.3 Update the list of what the current slice supports in `README.md`; verify it names clusters, cluster labels with `labeljust` and the `rank` values
