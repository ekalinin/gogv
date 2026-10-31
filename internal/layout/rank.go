// SPDX-License-Identifier: MIT
// Provenance: paper Gansner, Koutsofios, North, Vo 1993, A Technique for Drawing Directed Graphs, section 2

package layout

import (
	"context"
	"slices"
)

// rankEdge is an edge that constrains ranks: from tail to head, after the
// reversal that breaks cycles.
type rankEdge struct {
	tail, head int
	minlen     int
	weight     int
}

// assignRanks ranks n nodes: each edge's head is at least minlen ranks below
// its tail, and the total weighted edge length is minimal. Cycles are broken
// first by reversing the edges that a depth-first search, in node order and
// each node's out-edges by head, finds going back to a node on the current
// path. Each connected component is ranked on its own, from rank 0.
func assignRanks(ctx context.Context, n int, edges []rankEdge, search int) ([]int, error) {
	// A node's out-edges are visited by head, in node order, then in edge
	// order, as the reference visits them.
	out := make([][]int, n)
	for i, e := range edges {
		if e.tail != e.head {
			out[e.tail] = append(out[e.tail], i)
		}
	}
	for _, es := range out {
		slices.SortStableFunc(es, func(a, b int) int { return edges[a].head - edges[b].head })
	}
	reversed := make([]bool, len(edges))
	state := make([]byte, n) // 0 unvisited, 1 on the path, 2 done
	var dfs func(v int)
	dfs = func(v int) {
		state[v] = 1
		for _, i := range out[v] {
			h := edges[i].head
			switch state[h] {
			case 0:
				dfs(h)
			case 1:
				reversed[i] = true
			}
		}
		state[v] = 2
	}
	for v := range n {
		if state[v] == 0 {
			dfs(v)
		}
	}

	// Merge the edges between the same nodes: the longest minimum length
	// and the sum of the weights.
	type key struct{ t, h int }
	merged := map[key]int{}
	var dag []rankEdge
	for i, e := range edges {
		if e.tail == e.head {
			continue
		}
		if reversed[i] {
			e.tail, e.head = e.head, e.tail
		}
		k := key{e.tail, e.head}
		if j, ok := merged[k]; ok {
			dag[j].minlen = max(dag[j].minlen, e.minlen)
			dag[j].weight += e.weight
			continue
		}
		merged[k] = len(dag)
		dag = append(dag, e)
	}

	// Connected components, in the order of their first node.
	comp := components(n, dag)
	rank := make([]int, n)
	for _, nodes := range comp {
		local := make(map[int]int, len(nodes))
		for i, v := range nodes {
			local[v] = i
		}
		var nsEdges []nsEdge
		for _, e := range dag {
			if _, ok := local[e.tail]; ok {
				nsEdges = append(nsEdges, nsEdge{tail: local[e.tail], head: local[e.head], minlen: e.minlen, weight: e.weight})
			}
		}
		r, err := newSimplex(ctx, len(nodes), nsEdges, search).solve(balanceTopBottom)
		if err != nil {
			return nil, err
		}
		for i, v := range nodes {
			rank[v] = r[i]
		}
	}
	return rank, nil
}

// components returns the connected components of a graph, each as its nodes
// in increasing order, ordered by their first node.
func components(n int, edges []rankEdge) [][]int {
	parent := make([]int, n)
	for v := range parent {
		parent[v] = v
	}
	var find func(v int) int
	find = func(v int) int {
		for parent[v] != v {
			parent[v] = parent[parent[v]]
			v = parent[v]
		}
		return v
	}
	for _, e := range edges {
		a, b := find(e.tail), find(e.head)
		if a != b {
			if a < b {
				parent[b] = a
			} else {
				parent[a] = b
			}
		}
	}
	index := map[int]int{}
	var comps [][]int
	for v := range n {
		r := find(v)
		i, ok := index[r]
		if !ok {
			i = len(comps)
			index[r] = i
			comps = append(comps, nil)
		}
		comps[i] = append(comps[i], v)
	}
	return comps
}
