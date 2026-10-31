// SPDX-License-Identifier: MIT
// Provenance: paper Gansner, Koutsofios, North, Vo 1993, A Technique for Drawing Directed Graphs, section 4
// Provenance: observed dot 16.1.0

package layout

import (
	"context"
	"math"
)

// omegaTable weighs the edges of the auxiliary graph by how many of their
// ends are virtual nodes. Observed on the reference: an edge between a real
// and a virtual node pulls as much as one between real nodes.
var omegaTable = [3]int{1, 1, 8}

// placeX assigns integer x coordinates by the network simplex on the
// auxiliary graph of section 4.2: an extra node for each edge pulls its ends
// together with weight omega times the number of edges it stands for, and an
// edge between neighbors in a rank keeps them apart by their half widths
// plus nodesep, rounded. extra, when not nil, adds nodes from next on and
// their edges; the result holds their x coordinates after those of the
// ranked nodes.
func (g *ranked) placeX(ctx context.Context, width []float64, nodesep float64, search int, extra func(next int) ([]nsEdge, int)) ([]float64, error) {
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
	for _, e := range g.edges {
		w := omega(e) * e.count
		edges = append(edges,
			nsEdge{tail: next, head: e.upper, minlen: 0, weight: w},
			nsEdge{tail: next, head: e.lower, minlen: 0, weight: w})
		next++
	}
	for _, f := range g.flats {
		edges = append(edges,
			nsEdge{tail: next, head: f.left, minlen: 0, weight: 1},
			nsEdge{tail: next, head: f.right, minlen: 0, weight: 1})
		next++
	}
	for _, nodes := range g.ranks {
		for i := 0; i+1 < len(nodes); i++ {
			v, w := nodes[i], nodes[i+1]
			sep := int(math.Round(half(width[v]) + half(width[w]) + nodesep))
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
	sol, err := solveConnected(ctx, total, edges, search, balanceLeftRight)
	if err != nil {
		return nil, err
	}
	x := make([]float64, total)
	for v := range total {
		x[v] = float64(sol[v])
	}
	return x, nil
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
