// SPDX-License-Identifier: MIT
// Provenance: doc https://graphviz.org/docs/outputs/canon/
// Provenance: doc https://graphviz.org/docs/outputs/json/
// Provenance: doc https://graphviz.org/docs/outputs/svg/

package conformance

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/ekalinin/gogv/internal/dotlang"
	"github.com/ekalinin/gogv/internal/graph"
)

// layoutAttrs are the attributes whose values the layout computes: the
// contract compares that an object has them, not their values, which plain
// and svg judge.
var layoutAttrs = map[string]bool{
	"bb": true, "pos": true, "lp": true, "xlp": true, "head_lp": true, "tail_lp": true,
	"width": true, "height": true, "lwidth": true, "lheight": true,
}

// drawAttrs are the attributes that hold xdot drawing operations.
var drawAttrs = map[string]bool{
	"_draw_": true, "_ldraw_": true, "_hdraw_": true, "_tdraw_": true, "_hldraw_": true, "_tldraw_": true,
}

// problems collects the differences of a comparison, and reports the first
// few of them.
type problems []string

func (p *problems) add(format string, args ...any) { *p = append(*p, fmt.Sprintf(format, args...)) }

func (p problems) err() error {
	switch {
	case len(p) == 0:
		return nil
	case len(p) > 5:
		return fmt.Errorf("%s; and %d more", strings.Join(p[:5], "; "), len(p)-5)
	}
	return errors.New(strings.Join(p, "; "))
}

// CompareDOT compares outputs in DOT with the attributes of the layout
// (dot, gv and the xdot formats): the same graphs, subgraphs, nodes and
// edges, in the same order, the same attributes on each, the same values
// for those that do not come from the layout, the same drawing operations,
// and a size of the drawing within tolerance (a fraction) of the
// reference's.
func CompareDOT(ref, got []byte, tolerance float64) error {
	rgs, err := parseDOT(ref)
	if err != nil {
		return fmt.Errorf("reference: %w", err)
	}
	ggs, err := parseDOT(got)
	if err != nil {
		return err
	}
	var p problems
	if len(rgs) != len(ggs) {
		p.add("%d graphs, the reference has %d", len(ggs), len(rgs))
		return p.err()
	}
	for i := range rgs {
		compareGraph(&p, rgs[i], ggs[i], tolerance)
	}
	return p.err()
}

func parseDOT(data []byte) ([]*graph.Graph, error) {
	file, _, err := dotlang.Parse(string(data))
	if err != nil {
		return nil, err
	}
	var gs []*graph.Graph
	for _, d := range file.Graphs {
		gs = append(gs, dotlang.Build(d))
	}
	return gs, nil
}

// attrs returns the attributes of an object that have a value.
func attrs(symbols []*graph.Symbol, value func(*graph.Symbol) graph.Value) map[string]graph.Value {
	out := map[string]graph.Value{}
	for _, s := range symbols {
		if v := value(s); v.S != "" || v.HTML {
			out[s.Name] = v
		}
	}
	return out
}

func compareGraph(p *problems, ref, got *graph.Graph, tolerance float64) {
	rsub, gsub := ref.Graphs(), got.Graphs()
	if len(rsub) != len(gsub) {
		p.add("%d graphs and subgraphs, the reference has %d", len(gsub), len(rsub))
		return
	}
	for i := range rsub {
		r, g := rsub[i], gsub[i]
		name := "graph " + r.Name
		if r.Name != g.Name && !(r.Anonymous && g.Anonymous) {
			p.add("subgraph %d is %q, the reference's %q", i, g.Name, r.Name)
			continue
		}
		compareAttrs(p, name, attrs(ref.Symbols(graph.GraphKind), r.AttrAt), attrs(got.Symbols(graph.GraphKind), g.AttrAt))
		if names(r.Nodes()) != names(g.Nodes()) {
			p.add("%s has the nodes %s, the reference %s", name, names(g.Nodes()), names(r.Nodes()))
		}
		if len(r.Edges()) != len(g.Edges()) {
			p.add("%s has %d edges, the reference %d", name, len(g.Edges()), len(r.Edges()))
		}
	}
	rn, gn := ref.Nodes(), got.Nodes()
	if names(rn) != names(gn) {
		p.add("nodes %s, the reference's %s", names(gn), names(rn))
		return
	}
	for i := range rn {
		compareAttrs(p, "node "+rn[i].Name, attrs(ref.Symbols(graph.NodeKind), rn[i].AttrAt), attrs(got.Symbols(graph.NodeKind), gn[i].AttrAt))
	}
	re, ge := ref.Edges(), got.Edges()
	if len(re) != len(ge) {
		p.add("%d edges, the reference has %d", len(ge), len(re))
		return
	}
	for i := range re {
		name := "edge " + re[i].Tail.Name + "->" + re[i].Head.Name
		if re[i].Tail.Name != ge[i].Tail.Name || re[i].Head.Name != ge[i].Head.Name {
			p.add("edge %d is %s->%s, the reference's %s->%s", i, ge[i].Tail.Name, ge[i].Head.Name, re[i].Tail.Name, re[i].Head.Name)
			continue
		}
		compareAttrs(p, name, attrs(ref.Symbols(graph.EdgeKind), re[i].AttrAt), attrs(got.Symbols(graph.EdgeKind), ge[i].AttrAt))
	}
	rb, _ := ref.Attr("bb")
	gb, _ := got.Attr("bb")
	compareBB(p, rb.S, gb.S, tolerance)
}

func names(ns []*graph.Node) string {
	s := make([]string, len(ns))
	for i, n := range ns {
		s[i] = n.Name
	}
	return "[" + strings.Join(s, " ") + "]"
}

// compareAttrs compares the attributes of one object.
func compareAttrs(p *problems, object string, ref, got map[string]graph.Value) {
	for name, rv := range ref {
		gv, ok := got[name]
		switch {
		case !ok:
			p.add("%s lacks %s", object, name)
		case layoutAttrs[name]:
		case drawAttrs[name]:
			compareOps(p, object+" "+name, rv.S, gv.S)
		case rv != gv:
			p.add("%s has %s=%q, the reference %q", object, name, gv.S, rv.S)
		}
	}
	for name := range got {
		if _, ok := ref[name]; !ok {
			p.add("%s has %s, which the reference has not", object, name)
		}
	}
}

// compareBB checks the size of the drawing from the bounding boxes.
func compareBB(p *problems, ref, got string, tolerance float64) {
	box := func(s string) (w, h float64, ok bool) {
		parts := strings.Split(s, ",")
		if len(parts) != 4 {
			return 0, 0, false
		}
		var v [4]float64
		for i, part := range parts {
			f, err := strconv.ParseFloat(part, 64)
			if err != nil {
				return 0, 0, false
			}
			v[i] = f
		}
		return v[2] - v[0], v[3] - v[1], true
	}
	rw, rh, rok := box(ref)
	gw, gh, gok := box(got)
	switch {
	case !rok:
		return
	case !gok:
		p.add("bb %q is not a box", got)
		return
	}
	for _, err := range []error{within("bb width", rw, gw, tolerance), within("bb height", rh, gh, tolerance)} {
		if err != nil {
			p.add("%v", err)
		}
	}
}

// xop is an xdot drawing operation: its letter, and the fields that do not
// come from the layout. Coordinates are left out; so is the number of
// points of a Bezier curve, which depends on the routing.
type xop struct {
	op     string
	fields string
}

// parseOps reads the drawing operations of an xdot attribute.
func parseOps(s string) ([]xop, error) {
	var ops []xop
	r := strings.NewReader(s)
	word := func() (string, error) {
		var b strings.Builder
		for {
			c, err := r.ReadByte()
			if err == io.EOF && b.Len() > 0 {
				return b.String(), nil
			}
			if err != nil {
				return "", err
			}
			if c == ' ' {
				if b.Len() > 0 {
					return b.String(), nil
				}
				continue
			}
			b.WriteByte(c)
		}
	}
	nums := func(n int) ([]string, error) {
		out := make([]string, n)
		for i := range out {
			w, err := word()
			if err != nil {
				return nil, err
			}
			out[i] = w
		}
		return out, nil
	}
	// text reads a string of n bytes after its count and a dash; a doubled
	// backslash counts as one byte.
	text := func() (string, error) {
		w, err := word()
		if err != nil {
			return "", err
		}
		n, err := strconv.Atoi(w)
		if err != nil {
			return "", err
		}
		if c, err := r.ReadByte(); err != nil || c != '-' {
			return "", fmt.Errorf("no string after %d", n)
		}
		buf := make([]byte, 0, n)
		for len(buf) < n {
			c, err := r.ReadByte()
			if err != nil {
				return "", err
			}
			if c == '\\' {
				if next, err := r.ReadByte(); err == nil && next != '\\' {
					r.UnreadByte()
				}
			}
			buf = append(buf, c)
		}
		return string(buf), nil
	}
	for {
		op, err := word()
		if err == io.EOF {
			return ops, nil
		}
		if err != nil {
			return nil, err
		}
		var fields string
		switch op {
		case "E", "e":
			_, err = nums(4)
		case "P", "p", "L", "B", "b":
			var n []string
			if n, err = nums(1); err == nil {
				var count int
				if count, err = strconv.Atoi(n[0]); err == nil {
					_, err = nums(2 * count)
					if op != "B" && op != "b" {
						fields = n[0]
					}
				}
			}
		case "T":
			var n []string
			if n, err = nums(4); err == nil {
				var t string
				t, err = text()
				fields = n[2] + " " + t
			}
		case "t":
			var n []string
			n, err = nums(1)
			if err == nil {
				fields = n[0]
			}
		case "C", "c", "S":
			fields, err = text()
		case "F":
			var n []string
			if n, err = nums(1); err == nil {
				var t string
				t, err = text()
				fields = n[0] + " " + t
			}
		case "I":
			if _, err = nums(4); err == nil {
				fields, err = text()
			}
		default:
			return nil, fmt.Errorf("unknown operation %q", op)
		}
		if err != nil {
			return nil, fmt.Errorf("operation %s: %w", op, err)
		}
		ops = append(ops, xop{op, fields})
	}
}

// compareOps compares two lists of drawing operations by kind, colors,
// styles, fonts and texts.
func compareOps(p *problems, object, ref, got string) {
	rops, err := parseOps(ref)
	if err != nil {
		p.add("%s: reference: %v", object, err)
		return
	}
	gops, err := parseOps(got)
	if err != nil {
		p.add("%s: %v", object, err)
		return
	}
	if opString(rops) != opString(gops) {
		p.add("%s is %s, the reference %s", object, opString(gops), opString(rops))
	}
}

func opString(ops []xop) string {
	var b strings.Builder
	for i, o := range ops {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(o.op)
		if o.fields != "" {
			b.WriteByte('(')
			b.WriteString(o.fields)
			b.WriteByte(')')
		}
	}
	return b.String()
}

// CompareJSON compares the json and json0 outputs as CompareDOT compares
// DOT: the same keys and structure, the same values but for those of the
// layout, the same drawing operations, and the size of the drawing within
// tolerance.
func CompareJSON(ref, got []byte, tolerance float64) error {
	rdocs, err := decodeJSON(ref)
	if err != nil {
		return fmt.Errorf("reference: %w", err)
	}
	gdocs, err := decodeJSON(got)
	if err != nil {
		return err
	}
	var p problems
	if len(rdocs) != len(gdocs) {
		p.add("%d documents, the reference has %d", len(gdocs), len(rdocs))
		return p.err()
	}
	for i := range rdocs {
		compareJSONValue(&p, "", rdocs[i], gdocs[i])
		rm, _ := rdocs[i].(map[string]any)
		gm, _ := gdocs[i].(map[string]any)
		rb, _ := rm["bb"].(string)
		gb, _ := gm["bb"].(string)
		compareBB(&p, rb, gb, tolerance)
	}
	return p.err()
}

func decodeJSON(data []byte) ([]any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var docs []any
	for {
		var v any
		err := dec.Decode(&v)
		if err == io.EOF {
			return docs, nil
		}
		if err != nil {
			return nil, err
		}
		docs = append(docs, v)
	}
}

func compareJSONValue(p *problems, path string, ref, got any) {
	switch r := ref.(type) {
	case map[string]any:
		g, ok := got.(map[string]any)
		if !ok {
			p.add("%s is not an object", path)
			return
		}
		for k, rv := range r {
			gv, ok := g[k]
			switch {
			case !ok:
				p.add("%s lacks %s", path, k)
			case layoutAttrs[k]:
			case drawAttrs[k]:
				compareJSONOps(p, path+"."+k, rv, gv)
			default:
				compareJSONValue(p, path+"."+k, rv, gv)
			}
		}
		for k := range g {
			if _, ok := r[k]; !ok {
				p.add("%s has %s, which the reference has not", path, k)
			}
		}
	case []any:
		g, ok := got.([]any)
		if !ok || len(g) != len(r) {
			p.add("%s has %d items, the reference %d", path, len(g), len(r))
			return
		}
		for i := range r {
			compareJSONValue(p, fmt.Sprintf("%s[%d]", path, i), r[i], g[i])
		}
	default:
		if ref != got {
			p.add("%s is %v, the reference %v", path, got, ref)
		}
	}
}

// compareJSONOps compares drawing operations written as JSON objects.
func compareJSONOps(p *problems, path string, ref, got any) {
	ops := func(v any) string {
		list, _ := v.([]any)
		var b strings.Builder
		for _, item := range list {
			o, _ := item.(map[string]any)
			op := fmt.Sprint(o["op"])
			b.WriteString(op)
			b.WriteByte('(')
			for _, k := range []string{"color", "grad", "style", "face", "size", "align", "text"} {
				if v, ok := o[k]; ok {
					fmt.Fprintf(&b, "%s=%v ", k, v)
				}
			}
			if pts, ok := o["points"].([]any); ok && op != "B" && op != "b" {
				fmt.Fprintf(&b, "points=%d", len(pts))
			}
			b.WriteString(") ")
		}
		return b.String()
	}
	if ops(ref) != ops(got) {
		p.add("%s is %s, the reference %s", path, ops(got), ops(ref))
	}
}

// CompareInlineSVG compares svg_inline outputs: the lines before the root
// element, but for the version of the generator, then the drawing as
// CompareSVG compares it, once the namespace that the rasterizer needs is
// put back.
func CompareInlineSVG(r Rasterizer, ref, got []byte) error {
	prolog := func(svg []byte) string {
		head, _, _ := bytes.Cut(svg, []byte("<svg"))
		var lines []string
		for line := range strings.Lines(string(head)) {
			if strings.HasPrefix(line, "<!-- Generated by graphviz version") {
				line = "<!-- Generated by graphviz version"
			}
			lines = append(lines, strings.TrimRight(line, "\n"))
		}
		return strings.Join(lines, "\n")
	}
	if prolog(ref) != prolog(got) {
		return fmt.Errorf("prolog %q, the reference's %q", prolog(got), prolog(ref))
	}
	ns := func(svg []byte) []byte {
		return bytes.Replace(svg, []byte("<svg "), []byte(`<svg xmlns="http://www.w3.org/2000/svg" `), 1)
	}
	_, err := CompareSVG(r, ns(ref), ns(got))
	return err
}
