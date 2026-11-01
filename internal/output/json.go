// SPDX-License-Identifier: MIT
// Provenance: doc https://graphviz.org/docs/outputs/json/
// Provenance: observed dot 16.1.0

package output

import (
	"bytes"
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/ekalinin/gogv/internal/graph"
	"github.com/ekalinin/gogv/internal/layout"
)

// JSON writes the laid-out graph as the reference's -Tjson does, with the
// drawing operations of xdot, or as -Tjson0 does without them.
func JSON(res *layout.Result, draw bool) []byte {
	setLayoutAttrs(res)
	if draw {
		setDrawAttrs(res, "1.7")
	}
	return writeJSON(res.Graph, true)
}

// GraphJSON writes a graph as it is read, without layout, as the reference's
// -Tdot_json and -Txdot_json do.
func GraphJSON(g *graph.Graph) []byte { return writeJSON(g, false) }

// jsonWriter writes the lines of a JSON document, two spaces per level.
type jsonWriter struct {
	b bytes.Buffer
}

func (w *jsonWriter) line(level int, s string) {
	w.b.WriteString(strings.Repeat("  ", level))
	w.b.WriteString(s)
	w.b.WriteByte('\n')
}

// jsonString writes a JSON string, escaped as the reference escapes it.
func jsonString(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == '"', c == '\\', c == '/':
			b.WriteByte('\\')
			b.WriteByte(c)
		case c == '\n':
			b.WriteString(`\n`)
		case c == '\r':
			b.WriteString(`\r`)
		case c == '\t':
			b.WriteString(`\t`)
		case c < 0x20:
			fmt.Fprintf(&b, `\u%04x`, c)
		default:
			b.WriteByte(c)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// jsonNumber writes a number of a drawing operation with three decimals.
func jsonNumber(v float64) string { return strconv.FormatFloat(v, 'f', 3, 64) }

// member is a key and its value, written on one or more lines.
type member struct {
	key string
	// value is written after the key on the same line; ops, when value is
	// empty, are drawing operations written as an array below the key.
	value string
	ops   []map[string]string
}

// writeObject writes an object at level with its members.
func (w *jsonWriter) writeObject(level int, members []member, last bool) {
	w.line(level, "{")
	for i, m := range members {
		comma := ","
		if i == len(members)-1 {
			comma = ""
		}
		if m.ops != nil {
			w.writeArrayMember(level+1, m, comma)
			continue
		}
		w.line(level+1, jsonString(m.key)+": "+m.value+comma)
	}
	if last {
		w.line(level, "}")
	} else {
		w.line(level, "},")
	}
}

// opFieldOrder lists the fields of each drawing operation, in the
// reference's order.
var opFieldOrder = map[string][]string{
	"c": {"op", "grad", "color"}, "C": {"op", "grad", "color"},
	"S": {"op", "style"},
	"F": {"op", "size", "face"},
	"T": {"op", "pt", "align", "width", "text"},
	"t": {"op", "fontchar"},
	"P": {"op", "points"}, "p": {"op", "points"}, "L": {"op", "points"}, "b": {"op", "points"},
	"E": {"op", "rect"}, "e": {"op", "rect"},
}

func opFields(op map[string]string) []string {
	name, _ := strconv.Unquote(op["op"])
	return opFieldOrder[name]
}

// jsonOps turns xdot drawing operations into JSON objects: field names
// with their values written in JSON.
func jsonOps(s string) []map[string]string {
	var ops []map[string]string
	r := strings.NewReader(s)
	word := func() string {
		var b strings.Builder
		for {
			c, err := r.ReadByte()
			if err != nil {
				return b.String()
			}
			if c == ' ' {
				if b.Len() > 0 {
					return b.String()
				}
				continue
			}
			b.WriteByte(c)
		}
	}
	num := func() float64 {
		v, _ := strconv.ParseFloat(word(), 64)
		return v
	}
	// text reads a string after its length in bytes and a dash. A doubled
	// backslash counts as one byte, and stays doubled, as the reference
	// keeps it.
	text := func() string {
		n, _ := strconv.Atoi(word())
		r.ReadByte() // the dash
		var b []byte
		for count := 0; count < n; count++ {
			c, err := r.ReadByte()
			if err != nil {
				break
			}
			b = append(b, c)
			if c == '\\' {
				if next, err := r.ReadByte(); err == nil && next == '\\' {
					b = append(b, next)
				} else if err == nil {
					r.UnreadByte()
				}
			}
		}
		return string(b)
	}
	point := func() string { return "[" + jsonNumber(num()) + "," + jsonNumber(num()) + "]" }
	for {
		op := word()
		if op == "" {
			return ops
		}
		o := map[string]string{}
		switch op {
		case "c", "C":
			o["grad"] = jsonString("none")
			o["color"] = jsonString(text())
		case "S":
			o["style"] = jsonString(text())
		case "F":
			o["size"] = jsonNumber(num())
			o["face"] = jsonString(text())
		case "T":
			pt := point()
			align := map[string]string{"-1": "l", "0": "c", "1": "r"}[word()]
			o["pt"], o["align"] = pt, jsonString(align)
			o["width"] = jsonNumber(num())
			o["text"] = jsonString(text())
		case "t":
			o["fontchar"] = word()
		case "P", "p", "L", "B", "b":
			n, _ := strconv.Atoi(word())
			pts := make([]string, n)
			for i := range pts {
				pts[i] = point()
			}
			o["points"] = "[" + strings.Join(pts, ",") + "]"
			if op == "B" {
				// The reference names a Bezier curve b in JSON.
				op = "b"
			}
		case "E", "e":
			o["rect"] = "[" + jsonNumber(num()) + "," + jsonNumber(num()) + "," + jsonNumber(num()) + "," + jsonNumber(num()) + "]"
		}
		o["op"] = jsonString(op)
		ops = append(ops, o)
	}
}

// ids writes a list of object numbers on one line.
func ids(list []int) string {
	s := make([]string, len(list))
	for i, v := range list {
		s[i] = strconv.Itoa(v)
	}
	return strings.Join(s, ",")
}

// attrMembers returns the attributes of an object that the reference lists
// in JSON, in name order: those with a value, drawing operations as arrays.
func attrMembers(symbols []*graph.Symbol, value func(*graph.Symbol) graph.Value, keep func(*graph.Symbol, graph.Value) bool) []member {
	sorted := slices.Clone(symbols)
	slices.SortStableFunc(sorted, func(a, b *graph.Symbol) int { return cmp.Compare(a.Name, b.Name) })
	var out []member
	for _, s := range sorted {
		v := value(s)
		if !keep(s, v) {
			continue
		}
		if drawAttr[s.Name] {
			out = append(out, member{key: s.Name, ops: jsonOps(v.S)})
			continue
		}
		out = append(out, member{key: s.Name, value: jsonString(v.S)})
	}
	return out
}

// drawAttr are the attributes that hold xdot drawing operations.
var drawAttr = map[string]bool{"_draw_": true, "_ldraw_": true, "_hdraw_": true, "_tdraw_": true, "_hldraw_": true, "_tldraw_": true}

// writeJSON writes the root graph of g as a JSON document. laidOut is set
// when the graph went through the layout: the graph attributes of the
// layout and the drawing are then listed only where they are set.
func writeJSON(g *graph.Graph, laidOut bool) []byte {
	root := g.Root()
	subs := root.Graphs()[1:]
	nodes := root.Nodes()
	gvid := map[*graph.Graph]int{}
	for i, sg := range subs {
		gvid[sg] = i
	}
	nodeID := map[*graph.Node]int{}
	for i, n := range nodes {
		nodeID[n] = len(subs) + i
	}
	// Edges are listed in creation order, and numbered by tail, in node
	// order, and by head, as the reference numbers them.
	edges := root.Edges()
	edgeID := map[*graph.Edge]int{}
	for _, n := range nodes {
		var out []*graph.Edge
		for _, e := range edges {
			if e.Tail == n {
				out = append(out, e)
			}
		}
		slices.SortStableFunc(out, func(a, b *graph.Edge) int { return cmp.Compare(a.Head.Seq, b.Head.Seq) })
		for _, e := range out {
			edgeID[e] = len(edgeID)
		}
	}
	name := func(sg *graph.Graph) string {
		if sg.Name != "" {
			return sg.Name
		}
		return "%" + strconv.Itoa(2*sg.Unnamed-1)
	}
	present := func(v graph.Value) bool { return v.S != "" || v.HTML }
	graphKeep := func(sg *graph.Graph) func(*graph.Symbol, graph.Value) bool {
		return func(s *graph.Symbol, v graph.Value) bool {
			if laidOut && laidOutGraphAttr[s.Name] {
				return sg.AttrSetHere(s)
			}
			// A declared label is listed even when it is empty.
			return s.Name == "label" || present(v)
		}
	}
	objectKeep := func(s *graph.Symbol, v graph.Value) bool { return s.Name == "label" || present(v) }

	var w jsonWriter
	w.line(0, "{")
	top := []member{
		{key: "name", value: jsonString(name(root))},
		{key: "directed", value: strconv.FormatBool(root.Directed())},
		{key: "strict", value: strconv.FormatBool(root.Strict())},
	}
	top = append(top, attrMembers(root.Symbols(graph.GraphKind), root.AttrAt, graphKeep(root))...)
	top = append(top, member{key: "_subgraph_cnt", value: strconv.Itoa(len(subs))})
	hasObjects, hasEdges := len(subs)+len(nodes) > 0, len(edges) > 0
	// The members of the graph before its objects, then its objects and
	// edges as arrays.
	for i, m := range top {
		comma := ","
		if i == len(top)-1 && !hasObjects && !hasEdges {
			comma = ""
		}
		if m.ops != nil {
			w.writeArrayMember(1, m, comma)
			continue
		}
		w.line(1, jsonString(m.key)+": "+m.value+comma)
	}
	if hasObjects {
		w.line(1, `"objects": [`)
		count := 0
		total := len(subs) + len(nodes)
		for _, sg := range subs {
			count++
			members := []member{{key: "name", value: jsonString(name(sg))}}
			members = append(members, attrMembers(root.Symbols(graph.GraphKind), sg.AttrAt, graphKeep(sg))...)
			members = append(members, member{key: "_gvid", value: strconv.Itoa(gvid[sg])})
			var children, ns, es []int
			for _, c := range sg.Subgraphs {
				children = append(children, gvid[c])
			}
			for _, n := range sg.Nodes() {
				ns = append(ns, nodeID[n])
			}
			for _, e := range sg.Edges() {
				es = append(es, edgeID[e])
			}
			slices.Sort(ns)
			slices.Sort(es)
			for _, l := range []struct {
				key  string
				list []int
			}{{"subgraphs", children}, {"nodes", ns}, {"edges", es}} {
				if len(l.list) > 0 {
					members = append(members, member{key: l.key, value: "[\n" + strings.Repeat("  ", 4) + ids(l.list) + "\n" + strings.Repeat("  ", 3) + "]"})
				}
			}
			w.writeObject(2, members, count == total)
		}
		for _, n := range nodes {
			count++
			members := []member{{key: "_gvid", value: strconv.Itoa(nodeID[n])}, {key: "name", value: jsonString(n.Name)}}
			members = append(members, attrMembers(root.Symbols(graph.NodeKind), n.AttrAt, objectKeep)...)
			w.writeObject(2, members, count == total)
		}
		end := "],"
		if !hasEdges {
			end = "]"
		}
		w.line(1, end)
	}
	if hasEdges {
		w.line(1, `"edges": [`)
		for i, e := range edges {
			members := []member{
				{key: "_gvid", value: strconv.Itoa(edgeID[e])},
				{key: "tail", value: strconv.Itoa(nodeID[e.Tail])},
				{key: "head", value: strconv.Itoa(nodeID[e.Head])},
			}
			members = append(members, attrMembers(root.Symbols(graph.EdgeKind), e.AttrAt, objectKeep)...)
			w.writeObject(2, members, i == len(edges)-1)
		}
		w.line(1, "]")
	}
	w.line(0, "}")
	return w.b.Bytes()
}

// writeArrayMember writes drawing operations as a member of the graph.
func (w *jsonWriter) writeArrayMember(level int, m member, comma string) {
	w.line(level, jsonString(m.key)+": ")
	w.line(level, "[")
	for j, op := range m.ops {
		w.line(level+1, "{")
		fields := opFields(op)
		for k, f := range fields {
			sep := ","
			if k == len(fields)-1 {
				sep = ""
			}
			w.line(level+2, jsonString(f)+": "+op[f]+sep)
		}
		end := "},"
		if j == len(m.ops)-1 {
			end = "}"
		}
		w.line(level+1, end)
	}
	w.line(level, "]"+comma)
}
