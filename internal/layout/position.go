// SPDX-License-Identifier: MIT
// Provenance: paper Gansner, Koutsofios, North, Vo 1993, A Technique for Drawing Directed Graphs, section 4
// Provenance: observed dot 16.1.0

package layout

import (
	"cmp"
	"context"
	"math"
	"slices"
)

// omegaTable weighs the edges of the auxiliary graph by how many of their
// ends are virtual nodes. Observed on the reference: an edge between a real
// and a virtual node pulls as much as one between real nodes.
var omegaTable = [3]int{1, 1, 8}

// placeX assigns integer x coordinates by the network simplex on the
// auxiliary graph of section 4.2: an extra node for each edge pulls its ends
// together with weight omega times the weights of the edges it stands for, its
// ports rather than its nodes when it has ports (the offsets δ of section
// 4), and an
// edge between neighbors in a rank keeps them apart by how far they reach
// towards each other (half their widths, for nodes centered on their x) plus
// nodesep, rounded. extra, when not nil, adds nodes from next on and
// their edges; the result holds their x coordinates after those of the
// ranked nodes.
func (g *ranked) placeX(ctx context.Context, left, right []float64, nodesep float64, search int, extra func(next int) ([]nsEdge, int)) ([]float64, error) {
	n := len(g.nodes)
	var edges []nsEdge
	next := n
	omega := func(e redge) int {
		virt := 0
		if g.nodes[e.upper].real < 0 {
			virt++
		}
		if g.nodes[e.lower].real < 0 {
			virt++
		}
		return omegaTable[virt]
	}
	// With the ports pu and pv, the edges to the ends get the lengths m-pu
	// and m-pv, m the larger of the two: the cost becomes
	// |(xu+pu) - (xv+pv)| plus a constant.
	delta := func(pu, pv float64) (int, int) {
		u, v := int(math.Round(pu)), int(math.Round(pv))
		m := max(u, v)
		return m - u, m - v
	}
	// As observed, the reference lists the edges of the auxiliary graph by
	// the rank and the order of their upper ends: the network simplex breaks
	// its ties between them in this order.
	byUpper := make([]int, len(g.edges))
	for k := range byUpper {
		byUpper[k] = k
	}
	slices.SortStableFunc(byUpper, func(a, b int) int {
		u, v := g.nodes[g.edges[a].upper], g.nodes[g.edges[b].upper]
		return cmp.Or(cmp.Compare(u.rank, v.rank), cmp.Compare(u.order, v.order))
	})
	for _, k := range byUpper {
		e := g.edges[k]
		w := omega(e) * e.weight
		du, dv := delta(e.upperPort, e.lowerPort)
		edges = append(edges,
			nsEdge{tail: next, head: e.upper, minlen: du, weight: w},
			nsEdge{tail: next, head: e.lower, minlen: dv, weight: w})
		next++
	}
	for _, f := range g.flats {
		dl, dr := delta(f.leftPort, f.rightPort)
		edges = append(edges,
			nsEdge{tail: next, head: f.left, minlen: dl, weight: f.weight},
			nsEdge{tail: next, head: f.right, minlen: dr, weight: f.weight})
		next++
	}
	for _, nodes := range g.ranks {
		for i := 0; i+1 < len(nodes); i++ {
			v, w := nodes[i], nodes[i+1]
			sep := int(math.Round(right[v] + left[w] + nodesep + g.labelSep[[2]int{v, w}]))
			edges = append(edges, nsEdge{tail: v, head: w, minlen: sep, weight: 0})
		}
	}
	if extra != nil {
		es, end := extra(next)
		edges = append(edges, es...)
		next = end
	}
	total := next
	if total == 0 {
		return nil, nil
	}
	var rank, list []int
	ok := false
	if extra == nil {
		rank, list, ok = g.initialXTree(edges, byUpper, total)
	}
	var sol []int
	var err error
	if ok {
		sol, err = newSimplex(ctx, total, edges, search).solveFrom(rank, list, balanceLeftRight)
	} else {
		sol, err = solveConnected(ctx, total, edges, search, balanceLeftRight)
	}
	if err != nil {
		return nil, err
	}
	x := make([]float64, total)
	for v := range total {
		x[v] = float64(sol[v])
	}
	return x, nil
}

// initialXTree builds the initial feasible tree that section 4.3 of Gansner
// et al. gives for the auxiliary graph of placeX, whose edges are the two
// edges of each ranked edge, in the order byUpper, then those of each flat
// edge, then the separations rank by rank:
//   - each rank is packed from the left, with its separations in the tree;
//   - for each pair of adjacent ranks, both edges of one edge f between them
//     are in the tree, which places the lower rank under the upper one. As
//     observed, f goes from the leftmost node of the upper rank that has
//     edges down to the leftmost of their lower ends;
//   - every other edge has its edge to its leftmost end in the tree, to its
//     lower or right end when both ends are at the same x.
//
// It returns the ranking, and the tree edges in the order in which, as
// observed, the reference balances them: the edges of the fs, then the
// separations, then the others. It reports false when two adjacent ranks
// have no edge between them, so that the tree would not span the graph.
func (g *ranked) initialXTree(edges []nsEdge, byUpper []int, total int) ([]int, []int, bool) {
	if len(g.ranks) == 0 {
		return nil, nil, false
	}
	pos := make([]int, len(g.edges))
	for i, k := range byUpper {
		pos[k] = i
	}
	rank := make([]int, total)
	var fs, seps, others []int
	si := 2 * (len(g.edges) + len(g.flats))
	for _, nodes := range g.ranks {
		for i := 0; i+1 < len(nodes); i++ {
			rank[nodes[i+1]] = rank[nodes[i]] + edges[si].minlen
			seps = append(seps, si)
			si++
		}
	}
	isF := make([]bool, len(g.edges))
	for r := 0; r+1 < len(g.ranks); r++ {
		f := -1
		for _, v := range g.ranks[r] {
			for _, k := range g.nodes[v].down {
				if f < 0 || g.nodes[g.edges[k].lower].order < g.nodes[g.edges[f].lower].order {
					f = k
				}
			}
			if f >= 0 {
				break
			}
		}
		if f < 0 {
			return nil, nil, false
		}
		isF[f] = true
		eu, el := edges[2*pos[f]], edges[2*pos[f]+1]
		shift := (rank[eu.head] - eu.minlen) - (rank[el.head] - el.minlen)
		for _, v := range g.ranks[r+1] {
			rank[v] += shift
		}
	}
	// Each edge node lies at the leftmost of its ends, less their offsets.
	for i := range len(g.edges) + len(g.flats) {
		a, b := 2*i, 2*i+1
		xa, xb := rank[edges[a].head]-edges[a].minlen, rank[edges[b].head]-edges[b].minlen
		rank[edges[a].tail] = min(xa, xb)
		switch {
		case i < len(g.edges) && isF[byUpper[i]]:
			fs = append(fs, a, b)
		case xa < xb:
			others = append(others, a)
		default:
			others = append(others, b)
		}
	}
	return rank, slices.Concat(fs, seps, others), true
}

// solveConnected solves a network simplex problem whose graph may not be
// connected: each component is solved on its own.
func solveConnected(ctx context.Context, n int, edges []nsEdge, search int, mode balanceMode) ([]int, error) {
	var re []rankEdge
	for _, e := range edges {
		re = append(re, rankEdge{tail: e.tail, head: e.head})
	}
	sol := make([]int, n)
	for _, nodes := range components(n, re) {
		local := make(map[int]int, len(nodes))
		for i, v := range nodes {
			local[v] = i
		}
		var sub []nsEdge
		for _, e := range edges {
			if _, ok := local[e.tail]; ok {
				sub = append(sub, nsEdge{tail: local[e.tail], head: local[e.head], minlen: e.minlen, weight: e.weight})
			}
		}
		r, err := newSimplex(ctx, len(nodes), sub, search).solve(mode)
		if err != nil {
			return nil, err
		}
		for i, v := range nodes {
			sol[v] = r[i]
		}
	}
	return sol, nil
}

// placeY returns the y coordinate of each rank, y growing upward, with rank
// 0 at the top: the bottom rank's center is half its height above 0, and
// neighboring ranks are ranksep apart between the bottom of the upper one
// and the top of the lower one, and at least minGap[r] apart when minGap is
// not nil.
func (g *ranked) placeY(height []float64, ranksep float64, minGap []float64) []float64 {
	n := len(g.ranks)
	ht := make([]float64, n)
	for r, nodes := range g.ranks {
		for _, v := range nodes {
			ht[r] = max(ht[r], height[v])
		}
	}
	y := make([]float64, n)
	if n == 0 {
		return y
	}
	y[n-1] = half(ht[n-1])
	for r := n - 2; r >= 0; r-- {
		d := half(ht[r+1]) + ranksep + half(ht[r])
		if minGap != nil {
			d = max(d, minGap[r])
		}
		y[r] = y[r+1] + d
	}
	return y
}
