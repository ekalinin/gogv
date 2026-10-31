// SPDX-License-Identifier: MIT
// Provenance: paper Gansner, Koutsofios, North, Vo 1993, A Technique for Drawing Directed Graphs, section 3

package layout

import (
	"context"
	"slices"
)

// rnode is a node of the ranked graph: a real node or a virtual node that
// carries an edge across a rank.
type rnode struct {
	real  int // index of the real node, -1 for a virtual node
	count int // for a virtual node, the number of edges it carries
	rank  int
	order int
	comp  int
	up    []int // redges to the rank above
	down  []int // redges to the rank below
	flat  []int // flat edges, by index into ranked.flats
}

// redge joins nodes of adjacent ranks, from the upper to the lower one. One
// redge may stand for several parallel edges.
type redge struct {
	upper, lower int
	count        int
}

// flatEdge joins two nodes of the same rank; the order keeps left before
// right.
type flatEdge struct {
	left, right int
}

// ranked is the graph that the order and the positions work on.
type ranked struct {
	nodes  []rnode
	edges  []redge
	flats  []flatEdge
	ranks  [][]int // node ids by rank, in order
	nComps int
}

func (g *ranked) addEdge(upper, lower, count int) {
	g.edges = append(g.edges, redge{upper: upper, lower: lower, count: count})
	i := len(g.edges) - 1
	g.nodes[upper].down = append(g.nodes[upper].down, i)
	g.nodes[lower].up = append(g.nodes[lower].up, i)
}

// chainKey identifies the parallel edges that share a chain of virtual
// nodes: the same upper and lower real nodes.
type chainKey struct{ upper, lower int }

// buildRanked creates the ranked graph from the real nodes and their ranks,
// and the chains of the edges: chains[i] lists the ranked nodes of edge i
// from its upper to its lower end (nil for loops and flat edges).
func buildRanked(rank []int, edges [][2]int) (*ranked, [][]int) {
	g := &ranked{}
	for v, r := range rank {
		g.nodes = append(g.nodes, rnode{real: v, rank: r})
	}
	chains := make([][]int, len(edges))
	shared := map[chainKey][]int{}
	var sharedOrder []chainKey
	count := map[chainKey]int{}
	// Edges are taken by tail, then by head, in node order, then in edge
	// order, as the reference builds them: a reversed edge gets its chain
	// with its tail.
	byTail := make([]int, len(edges))
	for i := range byTail {
		byTail[i] = i
	}
	slices.SortStableFunc(byTail, func(a, b int) int {
		if edges[a][0] != edges[b][0] {
			return edges[a][0] - edges[b][0]
		}
		return edges[a][1] - edges[b][1]
	})
	for _, i := range byTail {
		e := edges[i]
		t, h := e[0], e[1]
		if t == h {
			continue
		}
		if rank[t] == rank[h] {
			g.flats = append(g.flats, flatEdge{left: t, right: h})
			fi := len(g.flats) - 1
			g.nodes[t].flat = append(g.nodes[t].flat, fi)
			g.nodes[h].flat = append(g.nodes[h].flat, fi)
			continue
		}
		upper, lower := t, h
		if rank[t] > rank[h] {
			upper, lower = h, t
		}
		k := chainKey{upper, lower}
		if c, ok := shared[k]; ok {
			chains[i] = c
			count[k]++
			continue
		}
		chain := []int{upper}
		for r := rank[upper] + 1; r < rank[lower]; r++ {
			g.nodes = append(g.nodes, rnode{real: -1, rank: r})
			chain = append(chain, len(g.nodes)-1)
		}
		chain = append(chain, lower)
		shared[k] = chain
		sharedOrder = append(sharedOrder, k)
		count[k] = 1
		chains[i] = chain
	}
	for _, k := range sharedOrder {
		chain := shared[k]
		for j := 0; j+1 < len(chain); j++ {
			g.addEdge(chain[j], chain[j+1], count[k])
		}
		for _, v := range chain[1 : len(chain)-1] {
			g.nodes[v].count = count[k]
		}
	}
	maxRank := 0
	for _, n := range g.nodes {
		maxRank = max(maxRank, n.rank)
	}
	g.ranks = make([][]int, maxRank+1)
	return g, chains
}

// order assigns the order of the nodes within their ranks: each connected
// component gets an initial order by a breadth-first search, then iterations
// of the weighted median and transposition heuristics; components are placed
// left to right in the order of their first real node. mirror reverses the
// order within each component.
func (g *ranked) order(ctx context.Context, maxIter int, mirror bool) error {
	comps := g.components()
	g.nComps = len(comps)
	for _, comp := range comps {
		ranks := g.initialOrder(comp)
		best, err := g.mincross(ctx, ranks, maxIter)
		if err != nil {
			return err
		}
		for r, nodes := range best {
			if mirror {
				// With rankdir=LR the reference mirrors each component, so
				// that the first node of a rank ends up at the top.
				slices.Reverse(nodes)
			}
			g.ranks[r] = append(g.ranks[r], nodes...)
		}
	}
	for _, nodes := range g.ranks {
		for i, v := range nodes {
			g.nodes[v].order = i
		}
	}
	return nil
}

// components returns the connected components of the ranked graph, each as
// its nodes in increasing order, ordered by their first node.
func (g *ranked) components() [][]int {
	var edges []rankEdge
	for _, e := range g.edges {
		edges = append(edges, rankEdge{tail: e.upper, head: e.lower})
	}
	for _, f := range g.flats {
		edges = append(edges, rankEdge{tail: f.left, head: f.right})
	}
	comps := components(len(g.nodes), edges)
	for i, c := range comps {
		for _, v := range c {
			g.nodes[v].comp = i
		}
	}
	return comps
}

// neighbors returns the nodes adjacent to v in the order a search visits
// them: below, then above, then on its rank.
func (g *ranked) neighbors(v int) []int {
	var ns []int
	for _, e := range g.nodes[v].down {
		ns = append(ns, g.edges[e].lower)
	}
	for _, e := range g.nodes[v].up {
		ns = append(ns, g.edges[e].upper)
	}
	for _, f := range g.nodes[v].flat {
		fe := g.flats[f]
		if fe.left == v {
			ns = append(ns, fe.right)
		} else {
			ns = append(ns, fe.left)
		}
	}
	return ns
}

// initialOrder places the nodes of a component in their ranks in the order
// a breadth-first search visits them, starting from the component's nodes
// without an upper neighbor, in node order. A tree gets an order without
// crossings.
func (g *ranked) initialOrder(comp []int) [][]int {
	ranks := make([][]int, len(g.ranks))
	seen := map[int]bool{}
	var queue []int
	visit := func(start int) {
		if seen[start] {
			return
		}
		seen[start] = true
		queue = append(queue[:0], start)
		for len(queue) > 0 {
			v := queue[0]
			queue = queue[1:]
			ranks[g.nodes[v].rank] = append(ranks[g.nodes[v].rank], v)
			for _, w := range g.neighbors(v) {
				if !seen[w] {
					seen[w] = true
					queue = append(queue, w)
				}
			}
		}
	}
	for _, v := range comp {
		if len(g.nodes[v].up) == 0 {
			visit(v)
		}
	}
	for _, v := range comp {
		visit(v)
	}
	return ranks
}

// mincrossOptions are the details of the heuristics that the paper leaves
// open; they are chosen to follow the reference.
type mincrossOptions struct {
	// tieReverse reverses the order of nodes with equal medians on the
	// iterations where iter%2 == tieReverse-1 (0 never).
	tieReverse int
	// equalSwap lets transpose swap nodes with equal crossings on the
	// iterations where iter%2 == equalSwap-1 (0 never, 3 always).
	equalSwap int
	// keepEqual replaces the best order by a later one with as many
	// crossings.
	keepEqual bool
	// initialTranspose runs transpose on the initial order.
	initialTranspose bool
	// minQuit stops after that many iterations without improvement.
	minQuit int
}

var mincrossOpts = mincrossOptions{tieReverse: 0, equalSwap: 3, keepEqual: true, initialTranspose: true, minQuit: 1 << 30}

// mincross improves an order by the weighted median and transposition
// heuristics and returns the order with the fewest crossings found.
func (g *ranked) mincross(ctx context.Context, ranks [][]int, maxIter int) ([][]int, error) {
	o := mincrossOpts
	cur := cloneRanks(ranks)
	g.enforceFlat(cur)
	g.setOrder(cur)
	if o.initialTranspose {
		g.transpose(cur, false)
	}
	best := cloneRanks(cur)
	bestCross := g.crossings(cur)
	trying := 0
	for iter := 0; iter < maxIter && bestCross > 0; iter++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if trying >= o.minQuit {
			break
		}
		trying++
		g.wmedian(cur, iter, o.tieReverse > 0 && iter%2 == o.tieReverse-1)
		g.transpose(cur, o.equalSwap == 3 || o.equalSwap > 0 && iter%2 == o.equalSwap-1)
		c := g.crossings(cur)
		if c < bestCross {
			trying = 0
		}
		if c < bestCross || o.keepEqual && c == bestCross {
			best, bestCross = cloneRanks(cur), c
		}
	}
	return best, nil
}

func cloneRanks(ranks [][]int) [][]int {
	out := make([][]int, len(ranks))
	for i, r := range ranks {
		out[i] = slices.Clone(r)
	}
	return out
}

func (g *ranked) setOrder(ranks [][]int) {
	for _, nodes := range ranks {
		for i, v := range nodes {
			g.nodes[v].order = i
		}
	}
}

// wmedian sorts each rank by the weighted median of the positions of its
// neighbors in the previous rank: going down on even iterations, up on odd
// ones. Nodes without neighbors there keep their positions; with
// reverseTies, nodes with equal medians swap their relative order.
func (g *ranked) wmedian(ranks [][]int, iter int, reverseTies bool) {
	down := iter%2 == 0
	n := len(ranks)
	for k := 1; k < n; k++ {
		r := k
		if !down {
			r = n - 1 - k
		}
		nodes := ranks[r]
		med := make(map[int]float64, len(nodes))
		for _, v := range nodes {
			med[v] = g.median(v, down)
		}
		// Sort the nodes with a median among the positions they occupy,
		// leaving the others in place.
		var movable []int
		var slots []int
		for i, v := range nodes {
			if med[v] >= 0 {
				movable = append(movable, v)
				slots = append(slots, i)
			}
		}
		if reverseTies {
			slices.Reverse(movable)
		}
		slices.SortStableFunc(movable, func(a, b int) int {
			switch {
			case med[a] < med[b]:
				return -1
			case med[a] > med[b]:
				return 1
			}
			return 0
		})
		for i, v := range movable {
			nodes[slots[i]] = v
		}
		g.enforceFlatRank(nodes)
		for i, v := range nodes {
			g.nodes[v].order = i
		}
	}
}

// median returns the weighted median of the positions of v's neighbors in
// the rank above (down) or below, or -1 when it has none (section 3).
func (g *ranked) median(v int, down bool) float64 {
	var pos []int
	if down {
		for _, e := range g.nodes[v].up {
			pos = append(pos, g.nodes[g.edges[e].upper].order)
		}
	} else {
		for _, e := range g.nodes[v].down {
			pos = append(pos, g.nodes[g.edges[e].lower].order)
		}
	}
	slices.Sort(pos)
	m := len(pos) / 2
	switch {
	case len(pos) == 0:
		return -1
	case len(pos)%2 == 1:
		return float64(pos[m])
	case len(pos) == 2:
		return float64(pos[0]+pos[1]) / 2
	}
	left := float64(pos[m-1] - pos[0])
	right := float64(pos[len(pos)-1] - pos[m])
	if left+right == 0 {
		return float64(pos[m-1]+pos[m]) / 2
	}
	return (float64(float64(pos[m-1])*right) + float64(float64(pos[m])*left)) / (left + right)
}

// transpose swaps adjacent nodes while that reduces crossings; with equal
// set, it also swaps nodes whose crossings are equal and not zero, which
// helps leave local minima.
func (g *ranked) transpose(ranks [][]int, equal bool) {
	for improved := true; improved; {
		improved = false
		for r := range ranks {
			nodes := ranks[r]
			for i := 0; i+1 < len(nodes); i++ {
				v, w := nodes[i], nodes[i+1]
				if g.flatBetween(v, w) {
					continue
				}
				c0 := g.pairCrossings(v, w)
				c1 := g.pairCrossings(w, v)
				if c1 < c0 || equal && c0 > 0 && c1 == c0 {
					nodes[i], nodes[i+1] = w, v
					g.nodes[v].order, g.nodes[w].order = i+1, i
					if c1 < c0 {
						improved = true
					}
				}
			}
		}
	}
}

// flatBetween reports whether a flat edge joins v and w.
func (g *ranked) flatBetween(v, w int) bool {
	for _, f := range g.nodes[v].flat {
		fe := g.flats[f]
		if fe.left == w || fe.right == w {
			return true
		}
	}
	return false
}

// pairCrossings counts the crossings between the edges of v and those of w,
// with v left of w, towards both neighboring ranks.
func (g *ranked) pairCrossings(v, w int) int {
	count := func(ev, ew []int, other func(redge) int) int {
		var c int
		for _, a := range ev {
			pa := g.nodes[other(g.edges[a])].order
			for _, b := range ew {
				if pb := g.nodes[other(g.edges[b])].order; pb < pa {
					c += g.edges[a].count * g.edges[b].count
				}
			}
		}
		return c
	}
	upper := func(e redge) int { return e.upper }
	lower := func(e redge) int { return e.lower }
	return count(g.nodes[v].up, g.nodes[w].up, upper) + count(g.nodes[v].down, g.nodes[w].down, lower)
}

// crossings counts the weighted crossings of an order.
func (g *ranked) crossings(ranks [][]int) int {
	g.setOrder(ranks)
	var total int
	for r := 0; r+1 < len(ranks); r++ {
		var es []redge
		for _, v := range ranks[r] {
			for _, e := range g.nodes[v].down {
				es = append(es, g.edges[e])
			}
		}
		for i := range es {
			for j := i + 1; j < len(es); j++ {
				a, b := es[i], es[j]
				ua, ub := g.nodes[a.upper].order, g.nodes[b.upper].order
				la, lb := g.nodes[a.lower].order, g.nodes[b.lower].order
				if (ua < ub && la > lb) || (ua > ub && la < lb) {
					total += a.count * b.count
				}
			}
		}
	}
	return total
}

// enforceFlat keeps the left end of every flat edge before its right end.
func (g *ranked) enforceFlat(ranks [][]int) {
	for _, nodes := range ranks {
		g.enforceFlatRank(nodes)
	}
}

// enforceFlatRank reorders a rank so that every flat edge goes left to
// right, moving as little as possible: a stable topological order.
func (g *ranked) enforceFlatRank(nodes []int) {
	if len(g.flats) == 0 {
		return
	}
	index := map[int]int{}
	for i, v := range nodes {
		index[v] = i
	}
	indeg := map[int]int{}
	for _, v := range nodes {
		for _, f := range g.nodes[v].flat {
			fe := g.flats[f]
			if fe.left == v {
				if _, ok := index[fe.right]; ok {
					indeg[fe.right]++
				}
			}
		}
	}
	placed := map[int]bool{}
	out := make([]int, 0, len(nodes))
	for len(out) < len(nodes) {
		progress := false
		for _, v := range nodes {
			if placed[v] || indeg[v] > 0 {
				continue
			}
			placed[v] = true
			out = append(out, v)
			progress = true
			for _, f := range g.nodes[v].flat {
				fe := g.flats[f]
				if fe.left == v {
					indeg[fe.right]--
				}
			}
			break
		}
		if !progress { // a cycle of flat edges: keep the rest as it is
			for _, v := range nodes {
				if !placed[v] {
					placed[v] = true
					out = append(out, v)
				}
			}
		}
	}
	copy(nodes, out)
}
