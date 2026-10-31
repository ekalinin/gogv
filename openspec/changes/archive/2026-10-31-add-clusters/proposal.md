# Proposal

## Why

Slice 1 lays out cluster subgraphs as plain subgraphs and ignores `rank`, with a "not supported yet" warning. gopuml draws packages, composite states and partitions as clusters and pins nodes with `rank=source` and `rank=sink` inside them: 36 of the 260 graphs of its harvested corpus use clusters. Every one of them also has an HTML label on a cluster, which slice 4 adds, so gopuml can judge them only after that slice; this slice lays out their structure. This is slice 2 of the v1 roadmap.

## What Changes

- Cluster subgraphs: a subgraph whose name starts with `cluster`, compared without regard to case, as the reference recognizes it. Clusters nest, a node belongs to one cluster, and empty clusters are not drawn. As in the reference, a node is taken out of the other clusters that contain it, and `canon` shows that.
- Layout with clusters: the nodes of a cluster and the bends of the edges between them are drawn together inside a box, with margins, nested boxes inside their parent's and sibling boxes apart. Other edges are routed outside the box. Clusters are ranked cluster by cluster, as the reference does by default.
- Cluster attributes of the v1 language: a plain-text `label` drawn at the top of the box, `labeljust` (`c`, `l`, `r`), `style=solid`, `color`, and `fontname`, `fontsize` and `fontcolor` for the label. `labeljust` also places the graph label.
- Rank constraints: `rank=same|min|max|source|sink` on subgraphs, inside and outside clusters.
- `svg` output: one group of class `cluster` per drawn cluster, with its name in the `title`, its box as a `polygon` and its label as `text`.
- Conformance: curated cases for clusters and rank constraints; the cases that use them are judged instead of being reported as not supported yet. A variant of the 36 gopuml graphs with clusters, with empty labels in place of their HTML cluster labels, checks the layout on gopuml's structures; 26 of them use no other feature of a later slice.

Still not supported yet, with a warning, as other attributes outside the v1 language: `cluster=true`, `labelloc`, `margin`, `pencolor`, `bgcolor`, `fillcolor`, `penwidth`, `peripheries`, cluster styles other than `solid`, `newrank`, `clusterrank`, and `compound` with `lhead` and `ltail`. HTML labels on clusters stay in slice 4.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `dot-language`: cluster subgraphs and node membership, also in `canon`; the subgraph's copy of its parent's graph attributes; `labeljust`, cluster `style` and `color`, and `rank` become supported attributes; clusters and `rank` leave the list of features not supported yet.
- `dot-layout`: ranking with clusters, rank constraints, cluster boxes and labels, routing outside clusters, and `labeljust` on the graph label.
- `output-formats`: cluster groups and boxes in `svg`.
- `upstream-compatibility`: the example of a feature of a later slice moves from clusters to edge labels.

## Impact

- Code: `internal/attrs` (cluster recognition and warnings), `internal/graph` (cluster membership, which `canon` also writes), `internal/layout` (ranking, ordering, positioning and routing with clusters, rank constraints), `internal/output` (`svg` clusters).
- Conformance: new curated cases and references; the dot-language cases with clusters are judged; the 36 gopuml graphs with clusters stay not supported yet until slice 4, and their variant with empty cluster labels is judged; known differences are updated.
- `README.md`: the list of what the current slice supports.
- No change to the library API, the command line or `plain`.
