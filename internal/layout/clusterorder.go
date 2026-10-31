// SPDX-License-Identifier: MIT
// Provenance: paper Gansner, Koutsofios, North, Vo 1993, A Technique for Drawing Directed Graphs, section 3
// Provenance: observed dot 16.1.0

package layout

import (
	"context"
	"slices"
)

// skeletonCount weighs the edges between the skeleton nodes of a cluster in
// the crossing count, so that edges do not cross a cluster.
const skeletonCount = 64

// orderClusters orders the ranks so that the nodes of every cluster are next
// to each other on each of its ranks. The root graph is ordered first with
// each child cluster standing as one skeleton node per rank, joined by heavy
// edges; then each cluster, from the outside in, replaces its skeleton nodes
// by its own nodes and the skeletons of its children, ordered by the same
// heuristics with everything outside the cluster fixed. Virtual nodes belong
// to the innermost cluster that has both ends of their edge.
func (g *ranked) orderClusters(ctx context.Context, t *clusterTree, maxIter int, mirror bool) error {
	of := make([]int, len(g.nodes))
	for v, rn := range g.nodes {
		if rn.real >= 0 {
			of[v] = t.of[rn.real]
		}
	}
	// Virtual nodes: the innermost cluster of the two real ends of their
	// chain.
	for v, rn := range g.nodes {
		if rn.real >= 0 {
			continue
		}
		up, down := v, v
		for g.nodes[up].real < 0 {
			up = g.edges[g.nodes[up].up[0]].upper
		}
		for g.nodes[down].real < 0 {
			down = g.edges[g.nodes[down].down[0]].lower
		}
		of[v] = t.common(of[up], of[down])
	}
	g.clust = of

	// Tokens: the ranked nodes, then one skeleton node per cluster and rank.
	n := len(g.nodes)
	skel := make([][]int, len(t.clusters))
	var skelCluster, skelRank []int
	for c := 1; c < len(t.clusters); c++ {
		cl := t.clusters[c]
		for r := cl.minRank; r <= cl.maxRank; r++ {
			skel[c] = append(skel[c], n+len(skelCluster))
			skelCluster = append(skelCluster, c)
			skelRank = append(skelRank, r)
		}
	}
	tokenRank := func(tok int) int {
		if tok < n {
			return g.nodes[tok].rank
		}
		return skelRank[tok-n]
	}
	// The smallest node index in each cluster, to order its skeleton.
	first := make([]int, len(t.clusters))
	for c := range first {
		first[c] = n
	}
	for v := n - 1; v >= 0; v-- {
		for c := of[v]; c >= 0; c = t.clusters[c].parent {
			first[c] = v
		}
	}

	expanded := make([]bool, len(t.clusters))
	expanded[0] = true
	// rep returns the token that stands for ranked node v while the
	// clusters are expanded down to level c: v itself, or the skeleton of
	// the outermost cluster around v that is not expanded yet.
	rep := func(v int) int {
		top := -1
		for c := of[v]; c > 0; c = t.clusters[c].parent {
			if !expanded[c] {
				top = c
			}
		}
		if top < 0 {
			return v
		}
		return skel[top][g.nodes[v].rank-t.clusters[top].minRank]
	}

	var cur [][]int
	// Clusters are numbered parents first.
	for c := 0; c < len(t.clusters); c++ {
		expanded[c] = true
		// The units of level c: its own ranked nodes and the skeletons of
		// its children.
		var units []int
		for v := range g.nodes {
			if of[v] == c {
				units = append(units, v)
			}
		}
		for _, d := range t.clusters[c].children {
			units = append(units, skel[d]...)
		}
		key := func(tok int) int {
			if tok < n {
				return tok
			}
			return first[skelCluster[tok-n]]
		}
		slices.SortStableFunc(units, func(a, b int) int { return key(a) - key(b) })
		inside := func(v int) bool { return t.contains(c, of[v]) }

		lg := &ranked{ranks: make([][]int, len(g.ranks)), startSkeletons: c == 0}
		local := map[int]int{}
		var tokens []int
		addNode := func(tok int, fixed bool) int {
			if i, ok := local[tok]; ok {
				return i
			}
			i := len(lg.nodes)
			lg.nodes = append(lg.nodes, rnode{real: -1, rank: tokenRank(tok), fixed: fixed, skeleton: tok >= n})
			local[tok] = i
			tokens = append(tokens, tok)
			return i
		}
		for _, u := range units {
			addNode(u, false)
		}
		// unit maps a ranked node inside level c to its unit there.
		unit := func(v int) int {
			if of[v] == c {
				return v
			}
			d := t.childToward(c, of[v])
			return skel[d][g.nodes[v].rank-t.clusters[d].minRank]
		}
		type pair struct{ a, b int }
		counts := map[pair]int{}
		var pairs []pair
		addEdge := func(a, b, count int) {
			p := pair{a, b}
			if _, ok := counts[p]; !ok {
				pairs = append(pairs, p)
			}
			counts[p] += count
		}
		// The skeleton edges come first, so that a search reaches the rest
		// of a cluster before what lies beside it.
		for _, d := range t.clusters[c].children {
			for i := 0; i+1 < len(skel[d]); i++ {
				addEdge(local[skel[d][i]], local[skel[d][i+1]], skeletonCount)
			}
		}
		for _, e := range g.edges {
			ui, li := inside(e.upper), inside(e.lower)
			var a, b int
			switch {
			case ui && li:
				a, b = unit(e.upper), unit(e.lower)
				if a >= n && b >= n && skelCluster[a-n] == skelCluster[b-n] {
					continue // inside one child
				}
				a, b = local[a], local[b]
			case ui && c > 0:
				a, b = local[unit(e.upper)], addNode(rep(e.lower), true)
			case li && c > 0:
				a, b = addNode(rep(e.upper), true), local[unit(e.lower)]
			default:
				continue
			}
			addEdge(a, b, e.count)
		}
		for _, p := range pairs {
			lg.addEdge(p.a, p.b, counts[p])
		}
		for _, j := range g.joins {
			if inside(j[0]) && inside(j[1]) && unit(j[0]) != unit(j[1]) {
				lg.joins = append(lg.joins, [2]int{local[unit(j[0])], local[unit(j[1])]})
			}
		}
		for _, f := range g.flats {
			if !inside(f.left) || !inside(f.right) {
				continue
			}
			a, b := unit(f.left), unit(f.right)
			if a == b {
				continue
			}
			lg.flats = append(lg.flats, flatEdge{left: local[a], right: local[b]})
			fi := len(lg.flats) - 1
			lg.nodes[local[a]].flat = append(lg.nodes[local[a]].flat, fi)
			lg.nodes[local[b]].flat = append(lg.nodes[local[b]].flat, fi)
		}

		if c == 0 {
			if err := lg.order(ctx, maxIter, mirror); err != nil {
				return err
			}
			cur = make([][]int, len(g.ranks))
			for r, nodes := range lg.ranks {
				for _, i := range nodes {
					cur[r] = append(cur[r], tokens[i])
				}
			}
			continue
		}

		// Expand the cluster in place of its skeleton: the units in the
		// order of a breadth-first search, between the fixed nodes that
		// were left and right of the skeleton.
		// Inside a cluster, the search starts from its child clusters, as
		// observed on the reference.
		var movable []int
		for _, u := range units {
			if u >= n {
				movable = append(movable, local[u])
			}
		}
		for _, u := range units {
			if u < n {
				movable = append(movable, local[u])
			}
		}
		start := lg.initialOrder(movable)
		ranks := make([][]int, len(g.ranks))
		for r := range cur {
			if mirror {
				slices.Reverse(start[r])
			}
			pos := -1
			if r >= t.clusters[c].minRank && r <= t.clusters[c].maxRank {
				pos = slices.Index(cur[r], skel[c][r-t.clusters[c].minRank])
			}
			for i, tok := range cur[r] {
				if i == pos {
					ranks[r] = append(ranks[r], start[r]...)
					continue
				}
				if j, ok := local[tok]; ok && lg.nodes[j].fixed {
					ranks[r] = append(ranks[r], j)
				}
			}
		}
		best, err := lg.mincross(ctx, ranks, maxIter)
		if err != nil {
			return err
		}
		for r := t.clusters[c].minRank; r <= t.clusters[c].maxRank; r++ {
			var placed []int
			for _, i := range best[r] {
				if !lg.nodes[i].fixed {
					placed = append(placed, tokens[i])
				}
			}
			pos := slices.Index(cur[r], skel[c][r-t.clusters[c].minRank])
			cur[r] = slices.Concat(cur[r][:pos], placed, cur[r][pos+1:])
		}
	}

	for r, toks := range cur {
		g.ranks[r] = toks
		for i, v := range toks {
			g.nodes[v].order = i
		}
	}
	return nil
}
