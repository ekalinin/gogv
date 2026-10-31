# Design

## Context

See proposal.md for the motivation and the scope of this slice, and the delta specs for the behavior.

Current state:

- `gogv.Render` builds the graph model, collects warnings with `attrs.Check`, then writes `canon` from the model, or runs `layout.Layout` and writes `plain` or `svg`. `attrs.IsCluster` recognizes clusters by the prefix `cluster` with case, and `Check` warns for every cluster and every `rank`.
- The graph model already copies a parent's graph attributes into a subgraph when it is created, which is why a cluster shows the label of its parent unless it sets its own.
- The layout ranks with network simplex, adds virtual nodes for long edges, orders ranks with median and transposition, places x by network simplex on the auxiliary graph, and routes splines through corridors between neighbors.

Facts observed on the reference that shape the approach:

- `canon` shows the resolved membership: a node that loses a cluster (shared by sibling clusters, or in a rank subgraph outside its cluster) is missing from that cluster in `canon` too, and the reference warns `b was already in a rankset, deleted from cluster %1` for the second case.
- Ranking is recursive by default: in `digraph { subgraph cluster_a { a; b } a -> c -> d -> b }`, `a` and `b` share a rank and `d -> b` points upward.
- `rank` on a cluster has no effect; `rank` on the root graph has no effect on the root but is copied into later subgraphs, where it applies.
- A cluster box is 8 points from its nodes; the label adds its height plus a margin at the top, and ranks move apart to make room for it.
- Cluster groups come before node groups in `svg`, in preorder, with ids `clust1`, `clust2`, ...; `plain` has no cluster lines.
- gopuml writes `style=solid`, `color`, `labeljust="c"` and either `label=""` or an HTML table label on every cluster, nests clusters three deep, and uses `rank=source` and `rank=sink` only inside clusters.

## Goals / Non-Goals

**Goals:**

- Cluster boxes, positions and routes close enough to the reference that the gopuml graphs with clusters, with empty cluster labels, pass the similarity contract.
- One membership resolution shared by `canon` and the layout, so both agree with the reference.

**Non-Goals:**

- `newrank=true` and `clusterrank=global|none`: the global ranking is a different algorithm, and gopuml does not use it.
- Byte-identical cluster layouts, for the reason of the foundation's design.
- The reference's internal error on some rank subgraphs across clusters (`install_in_rank`): this system lays such graphs out.

## Decisions

### D1. Sources

The foundation's clean-room rule holds. Sources added for this slice:

- North, "Drawing ranked digraphs with recursive clusters" (ALCOM Workshop on Graph Drawing, 1993), which describes the method of `dot` for clusters, when a copy can be obtained
- Sander, "Layout of compound directed graphs" (Technical Report A/03/96, Universität des Saarlandes, 1996), for ordering and positioning with border nodes
- the Graphviz documentation: "The DOT Language" (cluster subgraphs, attribute inheritance), the attributes `rank`, `newrank`, `clusterrank`, `labeljust`, `labelloc`, `margin`, and "Drawing graphs with dot" (clusters, rank constraints)
- black-box observation of the reference, as for the foundation

The source code of Graphviz and of its ports stays forbidden.

### D2. Membership is resolved on the model, before every format

A new step after `Build` and before `Check` resolves cluster membership on the graph model: it recognizes clusters (prefix `cluster` without regard to case), keeps each node in the first cluster that has it in input order and in the innermost cluster inside it, takes nodes out of clusters that lose them and of the subgraphs inside those clusters, and takes nodes of rank subgraphs out of clusters that do not contain the rank subgraph, with a warning. `canon` and the layout then see the same model.

Alternative considered: resolving membership in the layout only. Rejected, because the reference's `canon` shows the resolved membership.

### D3. Ranking: recursive on clusters, rank sets merged

Rank sets are handled by merging the nodes of a `same` set into one node for network simplex, and by pinning `min`, `source`, `max` and `sink` sets to the ends with edges from or to the merged set, with `source` and `sink` kept alone on their rank. A cluster is ranked on its own first, innermost first, with the rank sets inside it, and is then one node in its parent that stands for the whole block and keeps its relative ranks. Edges between the inside and the outside of a cluster become edges to that block, with their `minlen` shifted by the rank of their end inside the cluster.

Cycles are broken once, between the nodes, as without clusters. Between blocks, an edge is reversed only on a cycle of positive length, where the ranks of the blocks cannot satisfy every edge; the reversed edge is the one that goes back to the block that comes first. Two edges in opposite directions between two blocks are often both satisfiable, and the reference keeps both, so a search that reverses one of them gives other ranks. Balancing counts the nodes that a block has on each rank. As observed on the reference, a node that a rank subgraph takes out of a cluster is ranked as if it were still in it.

Alternative considered: one global network simplex with order constraints between cluster members (the `newrank=true` behavior). Rejected, because it is not the default of the reference and gives other ranks.

### D4. Ordering: clusters stay contiguous

Ordering works top-down on the cluster tree. In a graph, each child cluster is represented on each of its ranks by one skeleton node, with heavy edges between its skeleton nodes; median and transposition run as today. Each cluster is then expanded in place: its skeleton nodes are replaced by its own nodes and nested skeletons, which are ordered by the same heuristics with everything outside the cluster fixed. Virtual nodes of an edge belong to the innermost cluster that contains both of its ends, so the edges of other nodes cannot enter the cluster.

The initial orders follow the reference's on small graphs: at the root, the search starts from the nodes without an upper neighbor and from the clusters, in node order; inside a cluster, it starts from the child clusters, then from the cluster's own nodes; a cluster is not reached through a flat edge; and the nodes of a rank subgraph belong to one component.

Alternative considered: ordering the whole graph at once and then sorting each rank by cluster. Rejected, because it adds crossings that the reference does not have.

### D5. Positioning: border nodes in the auxiliary graph

The auxiliary graph of the x positioning gets a left and a right border node per cluster. Separation edges keep each member at least its half-width plus 8 points from the borders, nested borders 8 points inside their parent's, a node next to a cluster that it is not in 8 points from its border, two neighboring clusters 8 points apart, and the borders at least as far apart as the label is wide. Observed on the reference, these 8 points are the margin of the box on the side of its members and on the side of other nodes, and the margin of the parent between two boxes. A weight between the borders of a cluster keeps it narrow. y positions grow the gap between two ranks when a cluster starts or ends between them, so that the margins and the label fit, with 8 points between a box and the nodes or boxes of the next rank.

### D6. Routing: corridors bounded by cluster borders

The corridor of an edge is bounded by the boxes of the clusters that it does not enter, in addition to its neighbors, so an edge with no end in a cluster stays outside its box. Next to such a box, the gate of a virtual node widens from the band of its rank to the height of the box, up to the bands of the neighboring ranks. An edge that enters a cluster is bounded by the box only where it is inside.

### D7. SVG: cluster groups before nodes

`Result` gets the cluster boxes and labels in preorder. `svg` writes one group per cluster before the nodes, with the reference's ids, a `polygon` without fill and the label `text` positioned by `labeljust`. `plain` is unchanged, since its format has no clusters.

### D8. Conformance

Curated cases are added for every scenario of the delta specs, recorded with `make reference`. The dot-language cases that use clusters are now judged by `canon`.

Every gopuml graph with clusters has an HTML label on a cluster, so it stays not supported yet until slice 4. To check the layout on gopuml's structures now, a tool writes a variant of these graphs into the cache, with `label=""` in place of each HTML label of a cluster, and `make harvest` records its references. The conformance run judges the variant as an optional corpus; 26 of its 36 graphs use no other feature of a later slice.

Remaining differences are listed in `known-differences.yaml` under `tune-layout-heuristics`, as in the foundation.

Alternative considered: judging gopuml's graphs with the HTML cluster labels left out of the comparison. Rejected, because the reference's boxes make room for those labels, so sizes and positions would differ for a reason that is not a layout error.

## Risks / Trade-offs

- [The full text of North's 1993 paper may not be obtainable] → Sander's report, the documentation and black-box observation cover the method; the paper is optional.
- [Ordering and positioning inside clusters may differ from the reference's tie-breaking, and SSIM falls below 0.90] → Probe small graphs on the reference as in the foundation, and list remaining cases as known differences.
- [The rank merging and the recursive ranking change the ranks of graphs without clusters] → Graphs without clusters and rank sets keep today's path; the existing conformance cases guard it.
- [Rank subgraphs that cross clusters make the reference fail with an internal error in some cases] → Such cases are not added to the corpus; this system lays them out with the documented warning.

## Migration Plan

No migration: the change only adds behavior. When it is archived, the Purpose of the `dot-layout` main spec drops the words "clusters" from the list of graphs that it does not cover.
