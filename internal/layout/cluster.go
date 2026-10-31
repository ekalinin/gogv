// SPDX-License-Identifier: MIT
// Provenance: paper Gansner, Koutsofios, North, Vo 1993, A Technique for Drawing Directed Graphs, section 2
// Provenance: doc https://graphviz.org/docs/attrs/newrank/
// Provenance: doc https://graphviz.org/docs/attrs/rank/
// Provenance: observed dot 16.1.0

package layout

import (
	"context"
	"math"
	"slices"

	"github.com/ekalinin/gogv/internal/graph"
	"github.com/ekalinin/gogv/internal/label"
)

// clusterMargin is the documented default margin of a cluster: the space
// between its box and the nodes and clusters inside it. Observed on the
// reference: a node outside a cluster keeps the same space from its box, the
// drawing keeps it around the boxes of the top-level clusters, and boxes
// keep it from the nodes and boxes of the neighboring ranks.
const clusterMargin = 8

// cluster is a drawn cluster, or the root graph at index 0.
type cluster struct {
	g        *graph.Graph
	parent   int // -1 for the root
	children []int
	depth    int
	// nodes are the real nodes whose innermost cluster this is, in node
	// order.
	nodes []int
	// rankSets are the subgraphs with a rank constraint whose innermost
	// cluster this is.
	rankSets []*graph.Graph
	// minRank and maxRank are the ranks the cluster spans.
	minRank, maxRank int
	// label is the cluster's label, labelW and labelH the size of its box;
	// no lines when it has none.
	label          Text
	labelW, labelH float64
	labelJust      string
	color          string
	// ln and rn are the left and right border nodes of the cluster in the
	// auxiliary graph of the x coordinates.
	ln, rn int
	// above and below are how far the box reaches above the y of its first
	// rank and below the y of its last rank.
	above, below float64
	// box is the lower left and the upper right corner of the box, in the
	// layout orientation.
	box [2]Point
}

// clusterTree holds the drawn clusters of a graph.
type clusterTree struct {
	clusters []cluster
	// of is the innermost cluster of each real node.
	of    []int
	index map[*graph.Node]int
	// rankOf is the cluster that ranks each real node: its own, or the one
	// that a rank subgraph took it out of.
	rankOf []int
}

// buildClusterTree collects the clusters that have nodes, nested as their
// subgraphs are, through subgraphs that are not clusters, and the rank
// subgraphs of each level.
func buildClusterTree(g *graph.Graph, index map[*graph.Node]int) *clusterTree {
	t := &clusterTree{of: make([]int, len(index)), index: index}
	t.clusters = append(t.clusters, cluster{g: g, parent: -1})
	var walk func(sg *graph.Graph, c int)
	walk = func(sg *graph.Graph, c int) {
		for _, s := range sg.Subgraphs {
			if !s.IsCluster() {
				if s.Rank() != "" {
					t.clusters[c].rankSets = append(t.clusters[c].rankSets, s)
				}
				walk(s, c)
				continue
			}
			if len(s.Nodes()) == 0 {
				continue
			}
			k := len(t.clusters)
			t.clusters = append(t.clusters, cluster{g: s, parent: c, depth: t.clusters[c].depth + 1})
			t.clusters[c].children = append(t.clusters[c].children, k)
			for _, n := range s.Nodes() {
				t.of[index[n]] = k
			}
			walk(s, k)
		}
	}
	walk(g, 0)
	for v, c := range t.of {
		t.clusters[c].nodes = append(t.clusters[c].nodes, v)
	}
	byGraph := map[*graph.Graph]int{}
	for i, c := range t.clusters {
		byGraph[c.g] = i
	}
	t.rankOf = slices.Clone(t.of)
	for n, v := range index {
		if c, ok := byGraph[g.RankedIn(n)]; ok && g.RankedIn(n) != nil {
			t.rankOf[v] = c
		}
	}
	return t
}

// drawn reports whether the graph has clusters to draw.
func (t *clusterTree) drawn() bool { return len(t.clusters) > 1 }

// hasRankSets reports whether a level of the graph has rank constraints.
func (t *clusterTree) hasRankSets() bool {
	for _, c := range t.clusters {
		if len(c.rankSets) > 0 {
			return true
		}
	}
	return false
}

// contains reports whether cluster c is cluster d or encloses it.
func (t *clusterTree) contains(c, d int) bool {
	for ; d >= 0; d = t.clusters[d].parent {
		if d == c {
			return true
		}
	}
	return false
}

// common returns the innermost cluster that encloses both c and d.
func (t *clusterTree) common(c, d int) int {
	for t.clusters[c].depth > t.clusters[d].depth {
		c = t.clusters[c].parent
	}
	for t.clusters[d].depth > t.clusters[c].depth {
		d = t.clusters[d].parent
	}
	for c != d {
		c, d = t.clusters[c].parent, t.clusters[d].parent
	}
	return c
}

// childToward returns the child of c that encloses cluster d, or -1 when d
// is c.
func (t *clusterTree) childToward(c, d int) int {
	if d == c {
		return -1
	}
	for t.clusters[d].parent != c {
		d = t.clusters[d].parent
	}
	return d
}

// readLabels reads the label, its justification and the color of every
// cluster. In rankdir=LR the label box lies across the ranks.
func (t *clusterTree) readLabels() {
	for i := range t.clusters[1:] {
		c := &t.clusters[i+1]
		c.color = supported(graph.GraphKind, "color", graphValue(c.g, "color"), "black")
		c.labelJust = supported(graph.GraphKind, "labeljust", graphValue(c.g, "labeljust"), "c")
		v := graphValue(c.g, "label")
		if v.S == "" {
			continue
		}
		c.label = labelOf(v, "", label.Object{Graph: c.g.Name}, graphValue(c.g, "fontname"), graphValue(c.g, "fontcolor"), graphValue(c.g, "fontsize"))
		if c.label.Present() {
			c.labelW, c.labelH = c.label.Width+labelMarginX, c.label.Height+labelMarginY
		}
	}
}

// rankItem is a node or a child cluster ranked as one block in its parent.
type rankItem struct {
	node, child int // one of them is -1
	key         int // the smallest real node index in it, for its order
}

// rank ranks the nodes cluster by cluster, as the reference does by
// default: each cluster is ranked on its own, innermost first, and is then
// one block in its parent that keeps the ranks of its nodes relative to each
// other. edges are the edges that constrain ranks, between real nodes.
func (t *clusterTree) rank(ctx context.Context, edges []rankEdge, search int) ([]int, error) {
	// Cycles are broken between the nodes first; between blocks, an edge
	// is reversed only when the ranks of the blocks cannot satisfy it.
	of := t.rankOf
	reversed := breakCycles(len(of), edges)
	edges = slices.Clone(edges)
	for i := range edges {
		if reversed[i] {
			edges[i].tail, edges[i].head = edges[i].head, edges[i].tail
		}
	}
	local := make([]int, len(of)) // ranks within the cluster being ranked
	height := make([]int, len(t.clusters))
	var rankCluster func(c int) error
	rankCluster = func(c int) error {
		for _, d := range t.clusters[c].children {
			if err := rankCluster(d); err != nil {
				return err
			}
		}
		// The items of the cluster, in the order of their first node.
		var items []rankItem
		for v, cv := range of {
			if cv == c {
				items = append(items, rankItem{node: v, child: -1, key: v})
			}
		}
		for _, d := range t.clusters[c].children {
			key := len(of)
			for v, cv := range of {
				if t.contains(d, cv) {
					key = v
					break
				}
			}
			items = append(items, rankItem{node: -1, child: d, key: key})
		}
		slices.SortStableFunc(items, func(a, b rankItem) int { return a.key - b.key })
		itemOfNode := map[int]int{}
		itemOfChild := map[int]int{}
		for i, it := range items {
			if it.node >= 0 {
				itemOfNode[it.node] = i
			} else {
				itemOfChild[it.child] = i
			}
		}
		itemOf := func(v int) int {
			if of[v] == c {
				return itemOfNode[v]
			}
			return itemOfChild[t.childToward(c, of[v])]
		}
		offset := func(v int) int {
			if of[v] == c {
				return 0
			}
			return local[v]
		}
		itemHeight := func(i int) int {
			if items[i].child >= 0 {
				return height[items[i].child]
			}
			return 1
		}

		// Rank sets merge their nodes: same sets each into one group, min
		// and source sets into the first group, max and sink sets into the
		// last one.
		parent := make([]int, len(items))
		for i := range parent {
			parent[i] = i
		}
		var find func(i int) int
		find = func(i int) int {
			for parent[i] != i {
				parent[i] = parent[parent[i]]
				i = parent[i]
			}
			return i
		}
		union := func(a, b int) {
			a, b = find(a), find(b)
			if a > b {
				a, b = b, a
			}
			parent[b] = a
		}
		minItem, maxItem := -1, -1
		var source, sink bool
		for _, s := range t.clusters[c].rankSets {
			first := -1
			for _, n := range s.Nodes() {
				v, ok := t.index[n]
				if !ok || !t.contains(c, of[v]) {
					continue
				}
				i := itemOf(v)
				switch s.Rank() {
				case "min", "source":
					if minItem < 0 {
						minItem = i
					}
					union(minItem, i)
				case "max", "sink":
					if maxItem < 0 {
						maxItem = i
					}
					union(maxItem, i)
				default:
					if first < 0 {
						first = i
					}
					union(first, i)
				}
			}
			source = source || s.Rank() == "source"
			sink = sink || s.Rank() == "sink"
		}
		group := make([]int, len(items))
		var groups []int // group -> representative item
		groupOf := map[int]int{}
		for i := range items {
			r := find(i)
			gi, ok := groupOf[r]
			if !ok {
				gi = len(groups)
				groupOf[r] = gi
				groups = append(groups, r)
			}
			group[i] = gi
		}
		minGroup, maxGroup := -1, -1
		if minItem >= 0 {
			minGroup = group[find(minItem)]
		}
		if maxItem >= 0 {
			maxGroup = group[find(maxItem)]
		}

		var es []rankEdge
		for _, e := range edges {
			if !t.contains(c, of[e.tail]) || !t.contains(c, of[e.head]) {
				continue
			}
			ti, hi := itemOf(e.tail), itemOf(e.head)
			if ti == hi && items[ti].child >= 0 {
				continue // ranked inside the child
			}
			tg, hg := group[ti], group[hi]
			if tg == hg {
				continue
			}
			ot, oh := offset(e.tail), offset(e.head)
			re := rankEdge{tail: tg, head: hg, minlen: e.minlen + ot - oh, weight: e.weight, backMinlen: e.minlen + oh - ot, hasBack: true}
			// Edges into the minimum rank and out of the maximum rank are
			// reversed, so that those ranks keep their place.
			if hg == minGroup && tg != minGroup || tg == maxGroup && hg != maxGroup {
				re.tail, re.head, re.minlen, re.backMinlen = re.head, re.tail, re.backMinlen, re.minlen
			}
			es = append(es, re)
		}
		groupHeight := make([]int, len(groups))
		for i := range items {
			groupHeight[group[i]] = max(groupHeight[group[i]], itemHeight(i))
		}
		for gi := range groups {
			if minGroup >= 0 && gi != minGroup {
				m := 0
				if source {
					m = 1
				}
				es = append(es, rankEdge{tail: minGroup, head: gi, minlen: m})
			}
			if maxGroup >= 0 && gi != maxGroup {
				m := groupHeight[gi] - 1
				if sink {
					m++
				}
				es = append(es, rankEdge{tail: gi, head: maxGroup, minlen: m})
			}
		}
		breakPositiveCycles(len(groups), es)
		occupancy := make([][]int, len(groups))
		for i, it := range items {
			gi := group[i]
			if it.child < 0 {
				if occupancy[gi] == nil {
					occupancy[gi] = []int{0}
				}
				occupancy[gi][0]++
				continue
			}
			for len(occupancy[gi]) < height[it.child] {
				occupancy[gi] = append(occupancy[gi], 0)
			}
			for v, cv := range of {
				if t.contains(it.child, cv) {
					occupancy[gi][local[v]]++
				}
			}
		}
		r, err := rankFeasible(ctx, len(groups), es, search, occupancy)
		if err != nil {
			return err
		}
		lowest := 0
		for v, cv := range of {
			if !t.contains(c, cv) {
				continue
			}
			local[v] = r[group[itemOf(v)]] + offset(v)
			lowest = min(lowest, local[v])
		}
		for v, cv := range of {
			if t.contains(c, cv) {
				local[v] -= lowest
				height[c] = max(height[c], local[v]+1)
			}
		}
		return nil
	}
	if err := rankCluster(0); err != nil {
		return nil, err
	}
	for i := range t.clusters {
		t.clusters[i].minRank, t.clusters[i].maxRank = -1, -1
	}
	for v, c := range t.of {
		for ; c >= 0; c = t.clusters[c].parent {
			cl := &t.clusters[c]
			if cl.minRank < 0 || local[v] < cl.minRank {
				cl.minRank = local[v]
			}
			cl.maxRank = max(cl.maxRank, local[v])
		}
	}
	return local, nil
}

// clusterCompactness pulls the left and right borders of a cluster
// together in the auxiliary graph of the x coordinates.
const clusterCompactness = 2

// xEdges returns the nodes and edges that clusters add to the auxiliary
// graph of the x coordinates: a left and a right border node per cluster,
// numbered from next on. Its nodes are clusterMargin inside the borders,
// its child clusters too, and its label fits between them. On each rank, a
// node next to a cluster that it is not in keeps clusterMargin from its
// border, and so do two neighboring clusters. In rankdir=LR the label lies
// at the right, beyond the nodes. It returns the edges and the next free
// node number.
func (t *clusterTree) xEdges(g *ranked, left, right []float64, next int, lr bool) ([]nsEdge, int) {
	var es []nsEdge
	sep := func(d float64) int { return int(math.Round(d)) }
	for c := 1; c < len(t.clusters); c++ {
		t.clusters[c].ln, t.clusters[c].rn = next, next+1
		next += 2
	}
	labelSide := func(c int) float64 {
		if lr {
			return t.clusters[c].labelH
		}
		return 0
	}
	for c := 1; c < len(t.clusters); c++ {
		cl := t.clusters[c]
		inner := 0.0
		if !lr {
			inner = cl.labelW
		}
		es = append(es, nsEdge{tail: cl.ln, head: cl.rn, minlen: int(math.Ceil(inner)), weight: clusterCompactness})
		if p := cl.parent; p > 0 {
			es = append(es,
				nsEdge{tail: t.clusters[p].ln, head: cl.ln, minlen: clusterMargin},
				nsEdge{tail: cl.rn, head: t.clusters[p].rn, minlen: sep(clusterMargin + labelSide(p))})
		}
	}
	for v := range g.nodes {
		c := g.clust[v]
		if c == 0 {
			continue
		}
		cl := t.clusters[c]
		es = append(es,
			nsEdge{tail: cl.ln, head: v, minlen: sep(clusterMargin + left[v])},
			nsEdge{tail: v, head: cl.rn, minlen: sep(clusterMargin + right[v] + labelSide(c))})
	}
	for _, nodes := range g.ranks {
		for i := 0; i+1 < len(nodes); i++ {
			u, v := nodes[i], nodes[i+1]
			cu, cv := g.clust[u], g.clust[v]
			common := t.common(cu, cv)
			outerA, outerB := -1, -1
			for a := cu; a != common; a = t.clusters[a].parent {
				es = append(es, nsEdge{tail: t.clusters[a].rn, head: v, minlen: sep(clusterMargin + left[v])})
				outerA = a
			}
			for b := cv; b != common; b = t.clusters[b].parent {
				es = append(es, nsEdge{tail: u, head: t.clusters[b].ln, minlen: sep(right[u] + clusterMargin)})
				outerB = b
			}
			if outerA >= 0 && outerB >= 0 {
				es = append(es, nsEdge{tail: t.clusters[outerA].rn, head: t.clusters[outerB].ln, minlen: clusterMargin})
			}
		}
	}
	return es, next
}

// yGaps returns, for each rank but the last, the least distance between
// its y and the next rank's that clusters need: what reaches below the rank
// (its nodes, and the boxes of the clusters that end on it) and what reaches
// above the next one (its nodes, and the boxes and labels of the clusters
// that start on it), with clusterMargin between them. It also sets how far
// each box reaches above its first rank and below its last one.
func (t *clusterTree) yGaps(g *ranked, height []float64, lr bool) []float64 {
	nodeHalf := make([]float64, len(g.ranks))
	for v, rn := range g.nodes {
		nodeHalf[rn.rank] = max(nodeHalf[rn.rank], half(height[v]))
	}
	for c := len(t.clusters) - 1; c > 0; c-- {
		cl := &t.clusters[c]
		for v, rn := range g.nodes {
			if g.clust[v] != c {
				continue
			}
			if rn.rank == cl.minRank {
				cl.above = max(cl.above, half(height[v]))
			}
			if rn.rank == cl.maxRank {
				cl.below = max(cl.below, half(height[v]))
			}
		}
		for _, d := range cl.children {
			if t.clusters[d].minRank == cl.minRank {
				cl.above = max(cl.above, t.clusters[d].above)
			}
			if t.clusters[d].maxRank == cl.maxRank {
				cl.below = max(cl.below, t.clusters[d].below)
			}
		}
		cl.above += clusterMargin
		cl.below += clusterMargin
		if !lr {
			cl.above += cl.labelH
		}
	}
	above := slices.Clone(nodeHalf)
	below := slices.Clone(nodeHalf)
	for _, cl := range t.clusters[1:] {
		above[cl.minRank] = max(above[cl.minRank], cl.above)
		below[cl.maxRank] = max(below[cl.maxRank], cl.below)
	}
	gaps := make([]float64, max(len(g.ranks)-1, 0))
	for r := range gaps {
		gaps[r] = below[r] + clusterMargin + above[r+1]
	}
	return gaps
}

// boxes sets the box of every cluster in the layout orientation, from the x
// of its border nodes and the y of its first and last ranks. In
// rankdir=LR a box grows along the ranks, on both sides, to be as long as
// its label is wide.
func (t *clusterTree) boxes(xs, ys []float64, lr bool) {
	for c := 1; c < len(t.clusters); c++ {
		cl := &t.clusters[c]
		cl.box = [2]Point{{xs[cl.ln], ys[cl.maxRank] - cl.below}, {xs[cl.rn], ys[cl.minRank] + cl.above}}
		if lr {
			if d := cl.labelW - (cl.box[1].Y - cl.box[0].Y); d > 0 {
				cl.box[0].Y -= half(d)
				cl.box[1].Y += half(d)
			}
		}
	}
}

// breakPositiveCycles reverses edges until no cycle is longer than 0, so
// that the ranks can satisfy every edge: on each such cycle, the edge that
// goes back to the node that comes first.
func breakPositiveCycles(n int, edges []rankEdge) {
	for {
		dist := make([]int, n)
		pred := make([]int, n)
		for v := range pred {
			pred[v] = -1
		}
		last := -1
		for range n {
			last = -1
			for i, e := range edges {
				if e.tail != e.head && dist[e.tail]+e.minlen > dist[e.head] {
					dist[e.head] = dist[e.tail] + e.minlen
					pred[e.head] = i
					last = e.head
				}
			}
			if last < 0 {
				return
			}
		}
		// last still relaxes after n rounds: walking back n edges from it
		// lands on a cycle.
		v := last
		for range n {
			v = edges[pred[v]].tail
		}
		back := -1
		for u := v; ; {
			i := pred[u]
			if back < 0 || edges[i].head < edges[back].head {
				back = i
			}
			u = edges[i].tail
			if u == v {
				break
			}
		}
		e := &edges[back]
		e.tail, e.head = e.head, e.tail
		if e.hasBack {
			e.minlen, e.backMinlen = e.backMinlen, e.minlen
		}
	}
}

// joins returns pairs of real nodes that are in the same rank subgraph.
func (t *clusterTree) joins() [][2]int {
	var out [][2]int
	for _, c := range t.clusters {
		for _, s := range c.rankSets {
			first := -1
			for _, n := range s.Nodes() {
				v, ok := t.index[n]
				if !ok {
					continue
				}
				if first < 0 {
					first = v
					continue
				}
				out = append(out, [2]int{first, v})
			}
		}
	}
	return out
}
