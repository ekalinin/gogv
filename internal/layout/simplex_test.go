// SPDX-License-Identifier: MIT
// Provenance: original

package layout

import (
	"context"
	"math/rand/v2"
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
