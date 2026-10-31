// SPDX-License-Identifier: MIT
// Provenance: paper Gansner, Koutsofios, North, Vo 1993, A Technique for Drawing Directed Graphs, sections 2.3, 2.4, 4.2

// Package layout computes the dot layout: ranks, the order of nodes within
// ranks, coordinates, and the routes of edges.
package layout

import (
	"context"
	"slices"
)

// nsEdge is an edge of a network simplex problem: rank(head) - rank(tail)
// must be at least minlen, and the solution minimizes the sum of
// weight * (rank(head) - rank(tail)).
type nsEdge struct {
	tail, head int
	minlen     int
	weight     int
}

// balanceMode chooses how equally good solutions are spread.
type balanceMode int

const (
	balanceNone balanceMode = iota
	// balanceTopBottom moves nodes with equal in and out weights to the
	// feasible rank with the fewest nodes (section 2.3).
	balanceTopBottom
	// balanceLeftRight slides a part of the tree attached by an edge with a
	// zero cut value to the middle of its free range.
	balanceLeftRight
)

// simplex solves a network simplex problem on a connected graph of n nodes.
// It returns the ranks, normalized so that the smallest is 0.
type simplex struct {
	n     int
	edges []nsEdge
	out   [][]int // edge indices by tail
	in    [][]int // edge indices by head

	rank    []int
	inTree  []bool // by edge
	treeOut [][]int
	treeIn  [][]int
	parent  []int // tree edge to the parent, -1 at the root
	low     []int
	lim     []int
	cut     []int // cut value by edge
	order   []int // nodes in postorder: leaves before their parents
	ctx     context.Context
	// search is how many tree edges with a negative cut value leaveEdge
	// examines before it takes the most negative one found.
	search int
	next   int
}

func newSimplex(ctx context.Context, n int, edges []nsEdge, search int) *simplex {
	s := &simplex{n: n, edges: edges, ctx: ctx, search: search}
	s.out = make([][]int, n)
	s.in = make([][]int, n)
	for i, e := range edges {
		s.out[e.tail] = append(s.out[e.tail], i)
		s.in[e.head] = append(s.in[e.head], i)
	}
	return s
}

// solve runs the network simplex and returns the ranks.
func (s *simplex) solve(mode balanceMode) ([]int, error) {
	if s.n == 0 {
		return nil, nil
	}
	s.initRank()
	s.feasibleTree()
	s.initCutValues()
	for iter := 0; ; iter++ {
		if iter%64 == 0 {
			if err := s.ctx.Err(); err != nil {
				return nil, err
			}
		}
		f := s.leaveEdge()
		if f < 0 {
			break
		}
		e := s.enterEdge(f)
		if e < 0 {
			break
		}
		s.exchange(f, e)
	}
	switch mode {
	case balanceTopBottom:
		s.normalize()
		s.balanceTB()
	case balanceLeftRight:
		s.balanceLR()
		s.normalize()
	default:
		s.normalize()
	}
	return s.rank, nil
}

func (s *simplex) slack(e int) int {
	ed := s.edges[e]
	return s.rank[ed.head] - s.rank[ed.tail] - ed.minlen
}

// initRank gives an initial feasible ranking: nodes in topological order,
// each as low as its in-edges allow.
func (s *simplex) initRank() {
	s.rank = make([]int, s.n)
	indeg := make([]int, s.n)
	for _, e := range s.edges {
		indeg[e.head]++
	}
	var queue []int
	for v := range s.n {
		if indeg[v] == 0 {
			queue = append(queue, v)
		}
	}
	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]
		for _, ei := range s.out[v] {
			e := s.edges[ei]
			s.rank[e.head] = max(s.rank[e.head], s.rank[v]+e.minlen)
			indeg[e.head]--
			if indeg[e.head] == 0 {
				queue = append(queue, e.head)
			}
		}
	}
}

// feasibleTree grows a spanning tree of tight edges, shifting the tree's
// ranks to make more edges tight when it gets stuck (section 2.3).
func (s *simplex) feasibleTree() {
	s.inTree = make([]bool, len(s.edges))
	inTree := make([]bool, s.n)
	if s.n == 1 {
		s.buildTree()
		return
	}
	inTree[0] = true
	size := 1
	var grow func(v int)
	grow = func(v int) {
		for _, ei := range s.out[v] {
			h := s.edges[ei].head
			if !inTree[h] && s.slack(ei) == 0 {
				inTree[h] = true
				s.inTree[ei] = true
				size++
				grow(h)
			}
		}
		for _, ei := range s.in[v] {
			t := s.edges[ei].tail
			if !inTree[t] && s.slack(ei) == 0 {
				inTree[t] = true
				s.inTree[ei] = true
				size++
				grow(t)
			}
		}
	}
	for v := range s.n {
		if inTree[v] {
			grow(v)
		}
	}
	for size < s.n {
		// The non-tree edge incident on the tree with the smallest slack.
		best, bestSlack := -1, 0
		for ei, e := range s.edges {
			if inTree[e.tail] == inTree[e.head] {
				continue
			}
			if sl := s.slack(ei); best < 0 || sl < bestSlack {
				best, bestSlack = ei, sl
			}
		}
		if best < 0 {
			break // not connected: cannot happen for the graphs given
		}
		delta := bestSlack
		if inTree[s.edges[best].head] {
			delta = -delta
		}
		for v := range s.n {
			if inTree[v] {
				s.rank[v] += delta
			}
		}
		for v := range s.n {
			if inTree[v] {
				grow(v)
			}
		}
	}
	s.buildTree()
}

// buildTree indexes the tree edges and computes the postorder numbers.
func (s *simplex) buildTree() {
	s.treeOut = make([][]int, s.n)
	s.treeIn = make([][]int, s.n)
	for ei, e := range s.edges {
		if s.inTree[ei] {
			s.treeOut[e.tail] = append(s.treeOut[e.tail], ei)
			s.treeIn[e.head] = append(s.treeIn[e.head], ei)
		}
	}
	s.parent = make([]int, s.n)
	s.low = make([]int, s.n)
	s.lim = make([]int, s.n)
	s.order = s.order[:0]
	s.postorder(0, -1, 1)
}

// postorder numbers the subtree of v: lim is v's number, low the smallest
// number in its subtree.
func (s *simplex) postorder(v, parentEdge, next int) int {
	s.parent[v] = parentEdge
	s.low[v] = next
	for _, ei := range s.treeOut[v] {
		if ei != parentEdge {
			next = s.postorder(s.edges[ei].head, ei, next)
		}
	}
	for _, ei := range s.treeIn[v] {
		if ei != parentEdge {
			next = s.postorder(s.edges[ei].tail, ei, next)
		}
	}
	s.lim[v] = next
	s.order = append(s.order, v)
	return next + 1
}

// inSubtree reports whether u is in the subtree rooted at v.
func (s *simplex) inSubtree(u, v int) bool {
	return s.low[v] <= s.lim[u] && s.lim[u] <= s.lim[v]
}

// initCutValues computes the cut value of every tree edge: the sum of the
// weights of the edges from the tail component to the head component minus
// those in the other direction, when the tree edge is removed. As section 2.4
// describes, the nodes are taken from the leaves inward, so that the cut
// value of a node's parent edge follows from the edges incident to the node
// and the cut values of its child edges.
func (s *simplex) initCutValues() {
	s.cut = make([]int, len(s.edges))
	for _, v := range s.order {
		if f := s.parent[v]; f >= 0 {
			s.cut[f] = s.localCut(v, f)
		}
	}
}

// localCut returns the cut value of f, the tree edge between v and its
// parent, from the edges incident to v.
func (s *simplex) localCut(v, f int) int {
	vIsTail := s.edges[f].tail == v
	cut := s.edges[f].weight
	add := func(ei int, out bool) {
		if ei == f {
			return
		}
		e := s.edges[ei]
		other := e.head
		if !out {
			other = e.tail
		}
		// Whether e crosses in the same direction as f would if other were
		// outside v's subtree.
		same := out == vIsTail
		if same {
			cut += e.weight
		} else {
			cut -= e.weight
		}
		if s.inTree[ei] && s.parent[other] == ei {
			// other is a child of v: its subtree is inside v's, so its own
			// cut value corrects the edges counted above.
			if same {
				cut -= s.cut[ei]
			} else {
				cut += s.cut[ei]
			}
		}
	}
	for _, ei := range s.out[v] {
		add(ei, true)
	}
	for _, ei := range s.in[v] {
		add(ei, false)
	}
	return cut
}

// leaveEdge returns a tree edge with a negative cut value, or -1. It
// searches cyclically from where the previous search stopped and, after
// finding s.search candidates, takes the most negative one.
func (s *simplex) leaveEdge() int {
	best, found := -1, 0
	m := len(s.edges)
	for k := 0; k < m; k++ {
		ei := (s.next + k) % m
		if !s.inTree[ei] || s.cut[ei] >= 0 {
			continue
		}
		if best < 0 || s.cut[ei] < s.cut[best] {
			best = ei
		}
		found++
		if s.search > 0 && found >= s.search {
			s.next = (ei + 1) % m
			return best
		}
	}
	if best >= 0 {
		s.next = (best + 1) % m
	}
	return best
}

// enterEdge returns the non-tree edge with the smallest slack that crosses
// the cut of f in the direction opposite to f.
func (s *simplex) enterEdge(f int) int {
	fe := s.edges[f]
	sub := fe.head
	if s.parent[fe.tail] == f {
		sub = fe.tail
	}
	// The replacing edge goes from the head component to the tail
	// component of f.
	subIsTail := sub == fe.tail
	best, bestSlack := -1, 0
	for ei, e := range s.edges {
		if s.inTree[ei] {
			continue
		}
		ti, hi := s.inSubtree(e.tail, sub), s.inSubtree(e.head, sub)
		if ti == hi {
			continue
		}
		// e must go from f's head component to f's tail component.
		if subIsTail {
			if ti {
				continue
			}
		} else if hi {
			continue
		}
		if sl := s.slack(ei); best < 0 || sl < bestSlack {
			best, bestSlack = ei, sl
		}
	}
	return best
}

// exchange replaces the tree edge f by e, updates the ranks so that e is
// tight, and recomputes the postorder numbers and the cut values.
func (s *simplex) exchange(f, e int) {
	fe := s.edges[f]
	sub := fe.head
	if s.parent[fe.tail] == f {
		sub = fe.tail
	}
	// Move the component of sub so that e becomes tight.
	delta := s.slack(e)
	if !s.inSubtree(s.edges[e].tail, sub) {
		delta = -delta
	}
	for v := range s.n {
		if s.inSubtree(v, sub) {
			s.rank[v] += delta
		}
	}
	s.inTree[f] = false
	s.inTree[e] = true
	s.buildTree()
	s.initCutValues()
}

func (s *simplex) normalize() {
	m := slices.Min(s.rank)
	for v := range s.rank {
		s.rank[v] -= m
	}
}

// balanceTB moves each node whose in and out weights are equal to the
// feasible rank with the fewest nodes, which spreads ranks without changing
// the total edge length.
func (s *simplex) balanceTB() {
	maxRank := slices.Max(s.rank)
	count := make([]int, maxRank+1)
	for _, r := range s.rank {
		count[r]++
	}
	for v := range s.n {
		var inW, outW int
		low, high := 0, maxRank
		for _, ei := range s.in[v] {
			e := s.edges[ei]
			inW += e.weight
			low = max(low, s.rank[e.tail]+e.minlen)
		}
		for _, ei := range s.out[v] {
			e := s.edges[ei]
			outW += e.weight
			high = min(high, s.rank[e.head]-e.minlen)
		}
		if inW != outW || len(s.in[v])+len(s.out[v]) == 0 {
			continue
		}
		best := s.rank[v]
		for r := low; r <= high; r++ {
			if count[r] < count[best] {
				best = r
			}
		}
		count[s.rank[v]]--
		count[best]++
		s.rank[v] = best
	}
}

// balanceLR slides, for each tree edge whose cut value is 0, the part of the
// tree below it to the middle of the range where all its edges stay
// feasible: such moves do not change the objective, and they center nodes
// between their neighbors.
func (s *simplex) balanceLR() {
	for f := range s.edges {
		if !s.inTree[f] || s.cut[f] != 0 {
			continue
		}
		fe := s.edges[f]
		sub := fe.head
		if s.parent[fe.tail] == f {
			sub = fe.tail
		}
		// How far the subtree can move down (negative) or up (positive)
		// with every edge between it and the rest staying feasible.
		down, up := -1<<30, 1<<30
		for ei, e := range s.edges {
			ti, hi := s.inSubtree(e.tail, sub), s.inSubtree(e.head, sub)
			if ti == hi {
				continue
			}
			sl := s.slack(ei)
			if hi {
				// Moving the subtree down shrinks this edge.
				down = max(down, -sl)
			} else {
				up = min(up, sl)
			}
		}
		if down == -1<<30 || up == 1<<30 {
			continue
		}
		shift := (down + up) / 2
		if shift == 0 {
			continue
		}
		for v := range s.n {
			if s.inSubtree(v, sub) {
				s.rank[v] += shift
			}
		}
	}
}
