// SPDX-License-Identifier: MIT
// Provenance: paper Zeller, Hildebrandt 2002, Simplifying and Isolating Failure-Inducing Input, section 3

package main

import (
	"errors"
	"fmt"
	"slices"

	"github.com/ekalinin/gogv/internal/conformance"
	"github.com/ekalinin/gogv/internal/dotlang"
)

// layout renders a DOT input as plain output: gogv's, which must give no
// warning, or the reference's.
type layout func(src string) (conformance.Plain, error)

// predicate builds the test of a candidate input from the original one: a
// candidate is interesting when both sides lay it out, and it fails the
// plain part of the proxy on an object that fails it in the original. It
// returns the test and the failing objects of the original.
func predicate(src string, ref, got layout) (func(string) bool, []string, error) {
	failing, err := failingObjects(src, ref, got)
	if err != nil {
		return nil, nil, err
	}
	if len(failing) == 0 {
		return nil, nil, errors.New("the input passes the proxy: nothing to reduce")
	}
	test := func(candidate string) bool {
		now, err := failingObjects(candidate, ref, got)
		if err != nil {
			return false
		}
		return slices.ContainsFunc(now, func(o string) bool { return slices.Contains(failing, o) })
	}
	return test, failing, nil
}

// failingObjects returns the objects of an input whose deviation exceeds
// the tolerance of the proxy, sorted.
func failingObjects(src string, ref, got layout) ([]string, error) {
	r, err := ref(src)
	if err != nil {
		return nil, fmt.Errorf("reference: %w", err)
	}
	g, err := got(src)
	if err != nil {
		return nil, fmt.Errorf("gogv: %w", err)
	}
	var failing []string
	for name, d := range conformance.ProxyDeviations(r, g) {
		if d > conformance.ProxyTolerance {
			failing = append(failing, name)
		}
	}
	slices.Sort(failing)
	return failing, nil
}

// reduce returns the smallest input it finds that test still finds
// interesting: it removes statements, then attributes, then replaces
// subgraphs by their statements, and repeats until no step removes
// anything. The original input must be interesting.
func reduce(src string, test func(string) bool) (string, error) {
	f, _, err := dotlang.Parse(src)
	if err != nil {
		return "", err
	}
	if !test(printFile(f)) {
		return "", errors.New("the printed input is not interesting")
	}
	for {
		before := printFile(f)
		f = reduceStmts(f, test)
		f = reduceAttrs(f, test)
		f = unwrapSubgraphs(f, test)
		if printFile(f) == before {
			return before, nil
		}
	}
}

// ddmin returns a 1-minimal subset of items that test accepts: no single
// item can be removed from it. test must accept items.
func ddmin(items []int, test func([]int) bool) []int {
	if len(items) > 0 && test(nil) {
		return nil
	}
	n := 2
	for len(items) >= 2 {
		chunks := split(items, n)
		reduced := false
		for _, c := range chunks {
			if test(c) {
				items, n, reduced = c, 2, true
				break
			}
		}
		if !reduced {
			for i := range chunks {
				rest := slices.Concat(slices.Concat(chunks[:i]...), slices.Concat(chunks[i+1:]...))
				if test(rest) {
					items, n, reduced = rest, max(n-1, 2), true
					break
				}
			}
		}
		if !reduced {
			if n >= len(items) {
				break
			}
			n = min(2*n, len(items))
		}
	}
	return items
}

// split cuts items into n chunks of nearly equal length.
func split(items []int, n int) [][]int {
	var chunks [][]int
	for i := range n {
		lo, hi := i*len(items)/n, (i+1)*len(items)/n
		if hi > lo {
			chunks = append(chunks, items[lo:hi])
		}
	}
	return chunks
}

// stmtRefs lists the statements of a syntax tree, depth first, each as its
// list and its index in it, so that a selection of them can be kept.
func stmtRefs(f *dotlang.File) []*dotlang.Stmt {
	var refs []*dotlang.Stmt
	var walk func(stmts []dotlang.Stmt)
	walk = func(stmts []dotlang.Stmt) {
		for i := range stmts {
			refs = append(refs, &stmts[i])
			if sg, ok := stmts[i].(*dotlang.Subgraph); ok {
				walk(sg.Stmts)
			}
		}
	}
	for _, g := range f.Graphs {
		walk(g.Stmts)
	}
	return refs
}

// keepStmts returns a copy of the syntax tree with only the statements of
// keep, by their place in stmtRefs; a statement inside a removed subgraph
// is removed with it.
func keepStmts(f *dotlang.File, keep []int) *dotlang.File {
	kept := map[int]bool{}
	for _, i := range keep {
		kept[i] = true
	}
	n := 0
	var filter func(stmts []dotlang.Stmt) []dotlang.Stmt
	filter = func(stmts []dotlang.Stmt) []dotlang.Stmt {
		var out []dotlang.Stmt
		for _, s := range stmts {
			k := kept[n]
			n++
			sg, isSubgraph := s.(*dotlang.Subgraph)
			if isSubgraph {
				inner := filter(sg.Stmts)
				if k {
					out = append(out, &dotlang.Subgraph{ID: sg.ID, Stmts: inner})
				}
				continue
			}
			if k {
				out = append(out, s)
			}
		}
		return out
	}
	c := &dotlang.File{}
	for _, g := range f.Graphs {
		c.Graphs = append(c.Graphs, &dotlang.GraphDecl{Strict: g.Strict, Directed: g.Directed, ID: g.ID, Stmts: filter(g.Stmts)})
	}
	return c
}

func indexes(n int) []int {
	all := make([]int, n)
	for i := range all {
		all[i] = i
	}
	return all
}

func reduceStmts(f *dotlang.File, test func(string) bool) *dotlang.File {
	keep := ddmin(indexes(len(stmtRefs(f))), func(keep []int) bool { return test(printFile(keepStmts(f, keep))) })
	return keepStmts(f, keep)
}

// attrLists returns the attribute lists of the statements of a syntax tree,
// depth first.
func attrLists(f *dotlang.File) []*[]dotlang.Attr {
	var lists []*[]dotlang.Attr
	for _, ref := range stmtRefs(f) {
		switch s := (*ref).(type) {
		case *dotlang.NodeStmt:
			lists = append(lists, &s.Attrs)
		case *dotlang.EdgeStmt:
			lists = append(lists, &s.Attrs)
		case *dotlang.AttrStmt:
			lists = append(lists, &s.Attrs)
		}
	}
	return lists
}

// keepAttrs returns a copy of the syntax tree whose attribute lists keep
// only the attributes of keep, numbered across all lists.
func keepAttrs(f *dotlang.File, keep []int) *dotlang.File {
	c := clone(f)
	kept := map[int]bool{}
	for _, i := range keep {
		kept[i] = true
	}
	n := 0
	for _, list := range attrLists(c) {
		var out []dotlang.Attr
		for _, a := range *list {
			if kept[n] {
				out = append(out, a)
			}
			n++
		}
		*list = out
	}
	return c
}

func reduceAttrs(f *dotlang.File, test func(string) bool) *dotlang.File {
	total := 0
	for _, list := range attrLists(f) {
		total += len(*list)
	}
	keep := ddmin(indexes(total), func(keep []int) bool { return test(printFile(keepAttrs(f, keep))) })
	return keepAttrs(f, keep)
}

// unwrapSubgraphs replaces each subgraph by its statements, outermost
// first, when the input stays interesting.
func unwrapSubgraphs(f *dotlang.File, test func(string) bool) *dotlang.File {
	for i := 0; i < len(stmtRefs(f)); {
		c := clone(f)
		ref := stmtRefs(c)[i]
		sg, ok := (*ref).(*dotlang.Subgraph)
		if !ok {
			i++
			continue
		}
		if c = unwrapAt(c, sg); test(printFile(c)) {
			f = c
			continue
		}
		i++
	}
	return f
}

// unwrapAt replaces the subgraph sg of a syntax tree by its statements.
func unwrapAt(f *dotlang.File, sg *dotlang.Subgraph) *dotlang.File {
	var replace func(stmts []dotlang.Stmt) []dotlang.Stmt
	replace = func(stmts []dotlang.Stmt) []dotlang.Stmt {
		var out []dotlang.Stmt
		for _, s := range stmts {
			if s == dotlang.Stmt(sg) {
				out = append(out, sg.Stmts...)
				continue
			}
			if inner, ok := s.(*dotlang.Subgraph); ok {
				inner.Stmts = replace(inner.Stmts)
			}
			out = append(out, s)
		}
		return out
	}
	for _, g := range f.Graphs {
		g.Stmts = replace(g.Stmts)
	}
	return f
}

// clone returns a deep copy of the statements and attribute lists of a
// syntax tree; IDs are shared, as nothing changes them.
func clone(f *dotlang.File) *dotlang.File {
	var stmts func([]dotlang.Stmt) []dotlang.Stmt
	stmts = func(in []dotlang.Stmt) []dotlang.Stmt {
		out := make([]dotlang.Stmt, len(in))
		for i, s := range in {
			switch s := s.(type) {
			case *dotlang.NodeStmt:
				out[i] = &dotlang.NodeStmt{Nodes: s.Nodes, Attrs: slices.Clone(s.Attrs)}
			case *dotlang.EdgeStmt:
				ops := slices.Clone(s.Operands)
				for j, o := range ops {
					if o.Subgraph != nil {
						ops[j].Subgraph = &dotlang.Subgraph{ID: o.Subgraph.ID, Stmts: stmts(o.Subgraph.Stmts)}
					}
				}
				out[i] = &dotlang.EdgeStmt{Operands: ops, Attrs: slices.Clone(s.Attrs)}
			case *dotlang.AttrStmt:
				out[i] = &dotlang.AttrStmt{Target: s.Target, Attrs: slices.Clone(s.Attrs)}
			case *dotlang.AssignStmt:
				out[i] = &dotlang.AssignStmt{Attr: s.Attr}
			case *dotlang.Subgraph:
				out[i] = &dotlang.Subgraph{ID: s.ID, Stmts: stmts(s.Stmts)}
			}
		}
		return out
	}
	c := &dotlang.File{}
	for _, g := range f.Graphs {
		c.Graphs = append(c.Graphs, &dotlang.GraphDecl{Strict: g.Strict, Directed: g.Directed, ID: g.ID, Stmts: stmts(g.Stmts)})
	}
	return c
}
