// SPDX-License-Identifier: MIT
// Provenance: doc https://graphviz.org/doc/info/lang.html
// Provenance: observed dot 16.1.0

package graph

import "strings"

// IsCluster reports whether g is a cluster: a subgraph whose name starts
// with "cluster", compared without regard to case.
func (g *Graph) IsCluster() bool {
	return !g.IsRoot() && len(g.Name) >= len("cluster") && strings.EqualFold(g.Name[:len("cluster")], "cluster")
}

// RankValues are the values of rank that constrain the nodes of a subgraph.
var RankValues = []string{"same", "min", "max", "source", "sink"}

// Rank returns the rank constraint of a subgraph that is not a cluster, or
// "" when it has none. rank has no effect on the root graph and on clusters.
func (g *Graph) Rank() string {
	if g.IsRoot() || g.IsCluster() {
		return ""
	}
	v, _ := g.Attr("rank")
	for _, r := range RankValues {
		if v.S == r {
			return r
		}
	}
	return ""
}

// ResolveClusters makes every node a member of one cluster at each level of
// nesting, as the reference does before it lays a graph out: within the root
// graph or a cluster, the nodes of the rank subgraphs that are not inside a
// nested cluster are taken first, then each nested cluster, in input order,
// keeps the nodes that are not taken yet. A node taken by an earlier cluster
// is removed from a later one and from its subgraphs, with its edges; a node
// taken by a rank subgraph is removed in the same way, with a warning. It
// returns the warnings. Then every edge becomes a member of the innermost
// cluster that has both of its ends, and of the graphs that enclose it.
func (g *Graph) ResolveClusters() []string {
	var warnings []string
	g.resolveLevel(&warnings)
	for _, e := range g.r.edges {
		var inner *Graph
		for _, c := range g.r.graphs {
			if c.IsCluster() && c.nodeIn[e.Tail] && c.nodeIn[e.Head] && (inner == nil || c.depth() > inner.depth()) {
				inner = c
			}
		}
		for sg := inner; sg != nil; sg = sg.Parent {
			sg.addEdge(e)
		}
	}
	return warnings
}

func (g *Graph) depth() int {
	d := 0
	for sg := g; !sg.IsRoot(); sg = sg.Parent {
		d++
	}
	return d
}

func (g *Graph) resolveLevel(warnings *[]string) {
	var ranks, clusters []*Graph
	var walk func(sg *Graph)
	walk = func(sg *Graph) {
		for _, c := range sg.Subgraphs {
			if c.IsCluster() {
				clusters = append(clusters, c)
				continue
			}
			if c.Rank() != "" {
				ranks = append(ranks, c)
			}
			walk(c)
		}
	}
	walk(g)
	inRankSet := map[*Node]bool{}
	for _, r := range ranks {
		for _, n := range r.nodes {
			inRankSet[n] = true
		}
	}
	taken := map[*Node]bool{}
	for _, c := range clusters {
		for _, n := range append([]*Node(nil), c.nodes...) {
			switch {
			case inRankSet[n]:
				*warnings = append(*warnings, n.Name+" was already in a rankset, deleted from cluster "+c.Name)
				if g.r.rankedIn == nil {
					g.r.rankedIn = map[*Node]*Graph{}
				}
				g.r.rankedIn[n] = c.innermostWith(n)
				c.RemoveNode(n)
			case taken[n]:
				c.RemoveNode(n)
			default:
				taken[n] = true
			}
		}
	}
	for _, c := range clusters {
		c.resolveLevel(warnings)
	}
}

// innermostWith returns the innermost cluster in g, or g, that has n.
func (g *Graph) innermostWith(n *Node) *Graph {
	for _, sg := range g.Subgraphs {
		if !sg.nodeIn[n] {
			continue
		}
		if in := sg.innermostWith(n); in.IsCluster() {
			return in
		}
	}
	return g
}

// RankedIn returns the cluster that a rank subgraph took n out of, or nil.
// The reference ranks such a node as if it were still in that cluster.
func (g *Graph) RankedIn(n *Node) *Graph { return g.r.rankedIn[n] }
