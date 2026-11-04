// SPDX-License-Identifier: MIT
// Provenance: original

package layout

import (
	"context"
	"math/rand/v2"
	"slices"
	"testing"
)

// bruteCut computes a cut value from its definition.
func bruteCut(s *simplex, f int) int {
	fe := s.edges[f]
	sub := fe.head
	if s.parent[fe.tail] == f {
		sub = fe.tail
	}
	tailSide := sub == fe.tail
	var cut int
	for _, e := range s.edges {
		ti, hi := s.inSubtree(e.tail, sub), s.inSubtree(e.head, sub)
		if ti == hi {
			continue
		}
		if ti == tailSide {
			cut += e.weight
		} else {
			cut -= e.weight
		}
	}
	return cut
}

func objective(edges []nsEdge, rank []int) int {
	var sum int
	for _, e := range edges {
		sum += e.weight * (rank[e.head] - rank[e.tail])
	}
	return sum
}

// randomDAG returns a connected acyclic graph: every node after the first
// gets an edge from an earlier node, plus extra edges forward.
func randomDAG(r *rand.Rand, n, extra int) []nsEdge {
	var edges []nsEdge
	for v := 1; v < n; v++ {
		edges = append(edges, nsEdge{tail: r.IntN(v), head: v, minlen: 1 + r.IntN(2), weight: 1 + r.IntN(3)})
	}
	for range extra {
		a, b := r.IntN(n), r.IntN(n)
		if a == b {
			continue
		}
		if a > b {
			a, b = b, a
		}
		edges = append(edges, nsEdge{tail: a, head: b, minlen: r.IntN(3), weight: 1 + r.IntN(3)})
	}
	return edges
}

func TestSimplexCutValuesAndOptimality(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	for trial := range 200 {
		n := 2 + r.IntN(30)
		edges := randomDAG(r, n, r.IntN(2*n))
		s := newSimplex(context.Background(), n, edges, 30)
		s.initRank()
		s.feasibleTree()
		s.initCutValues()
		for f := range edges {
			if s.inTree[f] && s.cut[f] != bruteCut(s, f) {
				t.Fatalf("trial %d: cut value of edge %d is %d, by definition %d", trial, f, s.cut[f], bruteCut(s, f))
			}
		}
		rank, err := s.solve(balanceNone)
		if err != nil {
			t.Fatal(err)
		}
		for i, e := range edges {
			if rank[e.head]-rank[e.tail] < e.minlen {
				t.Fatalf("trial %d: edge %d violates its minimum length", trial, i)
			}
		}
		for f := range edges {
			if s.inTree[f] {
				if s.cut[f] < 0 {
					t.Fatalf("trial %d: negative cut value at the end", trial)
				}
				if s.cut[f] != bruteCut(s, f) {
					t.Fatalf("trial %d: stale cut value", trial)
				}
			}
		}
		// No single node move improves an optimal solution.
		obj := objective(edges, rank)
		for v := range n {
			for _, d := range []int{-1, 1} {
				moved := append([]int(nil), rank...)
				moved[v] += d
				feasible := true
				for _, e := range edges {
					if moved[e.head]-moved[e.tail] < e.minlen {
						feasible = false
					}
				}
				if feasible && objective(edges, moved) < obj {
					t.Fatalf("trial %d: moving node %d by %d improves the objective", trial, v, d)
				}
			}
		}
	}
}

func TestSimplexShortestRanking(t *testing.T) {
	// a -> b -> c, a -> c: three consecutive ranks.
	edges := []nsEdge{{0, 1, 1, 1}, {1, 2, 1, 1}, {0, 2, 1, 1}}
	rank, _ := newSimplex(context.Background(), 3, edges, 30).solve(balanceNone)
	if rank[0] != 0 || rank[1] != 1 || rank[2] != 2 {
		t.Errorf("ranks %v", rank)
	}
}

func TestSimplexCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := rand.New(rand.NewPCG(3, 4))
	edges := randomDAG(r, 200, 400)
	if _, err := newSimplex(ctx, 200, edges, 30).solve(balanceNone); err == nil {
		t.Skip("solved before the first context check")
	}
}

// rebuildExchange exchanges f for e as exchange did before it updated the
// tree in place: it moves the ranks, swaps the edges, and rebuilds the tree
// and every cut value.
func rebuildExchange(s *simplex, f, e int) {
	fe := s.edges[f]
	sub := fe.head
	if s.parent[fe.tail] == f {
		sub = fe.tail
	}
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

// balance spreads the ranks as solve does after its last pivot.
func balance(s *simplex, mode balanceMode) {
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
}

// TestSimplexExchangeMatchesRebuild steps two solvers through each problem,
// one with exchange and one with rebuildExchange, and checks that they make
// the same pivots and end with the same ranks.
func TestSimplexExchangeMatchesRebuild(t *testing.T) {
	ctx := context.Background()
	r := rand.New(rand.NewPCG(5, 6))
	pivots := 0
	for trial := range 300 {
		n := 2 + r.IntN(40)
		edges := randomDAG(r, n, r.IntN(2*n))
		if trial%2 == 1 {
			// Edges of weight 0, as the edges that keep the neighbors in a
			// rank apart have in the problem of x coordinates.
			for i := range edges {
				if r.IntN(3) == 0 {
					edges[i].weight = 0
				}
			}
		}
		search := []int{30, 500, 0}[trial%3]
		a := newSimplex(ctx, n, edges, search)
		b := newSimplex(ctx, n, edges, search)
		for _, s := range []*simplex{a, b} {
			s.initRank()
			s.feasibleTree()
			s.initCutValues()
		}
		for pivot := 0; ; pivot++ {
			f := a.leaveEdge()
			if g := b.leaveEdge(); g != f {
				t.Fatalf("trial %d, pivot %d: leaving edge %d, after a rebuild %d", trial, pivot, f, g)
			}
			if f < 0 {
				break
			}
			e := a.enterEdge(f)
			if g := b.enterEdge(f); g != e {
				t.Fatalf("trial %d, pivot %d: entering edge %d, after a rebuild %d", trial, pivot, e, g)
			}
			if e < 0 {
				break
			}
			a.exchange(f, e)
			rebuildExchange(b, f, e)
			pivots++
			if !slices.Equal(a.rank, b.rank) || !slices.Equal(a.inTree, b.inTree) || !slices.Equal(a.parent, b.parent) {
				t.Fatalf("trial %d, pivot %d: ranks, tree or parent edges differ from a rebuild", trial, pivot)
			}
			for ei := range edges {
				if a.inTree[ei] && (a.cut[ei] != b.cut[ei] || a.cut[ei] != bruteCut(a, ei)) {
					t.Fatalf("trial %d, pivot %d: cut value of edge %d is %d, after a rebuild %d, by definition %d", trial, pivot, ei, a.cut[ei], b.cut[ei], bruteCut(a, ei))
				}
			}
			for u := range n {
				for v := range n {
					if a.inSubtree(u, v) != b.inSubtree(u, v) {
						t.Fatalf("trial %d, pivot %d: whether node %d is in the subtree of node %d differs from a rebuild", trial, pivot, u, v)
					}
				}
			}
		}
		final := slices.Clone(a.rank)
		for _, mode := range []balanceMode{balanceNone, balanceTopBottom, balanceLeftRight} {
			copy(a.rank, final)
			copy(b.rank, final)
			balance(a, mode)
			balance(b, mode)
			solved, err := newSimplex(ctx, n, edges, search).solve(mode)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(a.rank, b.rank) || !slices.Equal(solved, b.rank) {
				t.Fatalf("trial %d, mode %d: ranks differ from a rebuild", trial, mode)
			}
		}
	}
	if pivots == 0 {
		t.Fatal("no problem made a pivot")
	}
}
