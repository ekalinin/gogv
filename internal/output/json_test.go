// SPDX-License-Identifier: MIT
// Provenance: original

package output

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ekalinin/gogv/internal/dotlang"
	"github.com/ekalinin/gogv/internal/layout"
)

// renderResolved lays out each graph of src with its clusters resolved, as
// Render does, and writes it.
func renderResolved(t *testing.T, src string, write func(*layout.Result) []byte) string {
	t.Helper()
	f, _, err := dotlang.Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	for _, d := range f.Graphs {
		g := dotlang.Build(d)
		g.ResolveClusters()
		res, err := layout.Layout(context.Background(), g)
		if err != nil {
			t.Fatal(err)
		}
		out.Write(write(res))
	}
	return out.String()
}

func graphJSON(res *layout.Result) []byte { return GraphJSON(res.Graph) }
func json0(res *layout.Result) []byte     { return JSON(res, false) }
func jsonDraw(res *layout.Result) []byte  { return JSON(res, true) }

func decode(t *testing.T, s string) map[string]any {
	t.Helper()
	var v map[string]any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatalf("%v:\n%s", err, s)
	}
	return v
}

func TestJSONScenarios(t *testing.T) {
	// Graph as it is read: the reference's output, byte for byte.
	want := `{
  "name": "G",
  "directed": true,
  "strict": false,
  "_subgraph_cnt": 0,
  "objects": [
    {
      "_gvid": 0,
      "name": "a",
      "color": "#FF0000",
      "label": "\\N",
      "shape": "rect"
    },
    {
      "_gvid": 1,
      "name": "b",
      "label": "\\N"
    }
  ],
  "edges": [
    {
      "_gvid": 0,
      "tail": 0,
      "head": 1,
      "label": "x"
    }
  ]
}
`
	if got := renderResolved(t, `digraph G { a -> b [label="x"]; a [shape=rect, color="#FF0000"] }`, graphJSON); got != want {
		t.Errorf("graph as it is read:\n%s\nwant\n%s", got, want)
	}

	// Subgraphs in json0.
	v := decode(t, renderResolved(t, `digraph G { subgraph cluster_x { label="C"; a; subgraph cluster_y { b } } { rank=same; c; d } a -> b }`, json0))
	objects := v["objects"].([]any)
	if v["_subgraph_cnt"] != 3.0 || len(objects) != 7 {
		t.Fatalf("subgraphs in json0: %v", v)
	}
	for i, name := range []string{"cluster_x", "cluster_y", "%1", "a", "b", "c", "d"} {
		o := objects[i].(map[string]any)
		if o["name"] != name || o["_gvid"] != float64(i) {
			t.Errorf("object %d: %v", i, o)
		}
		if i >= 3 && o["pos"] == nil {
			t.Errorf("node %s has no pos", name)
		}
	}
	x := objects[0].(map[string]any)
	if s, _ := json.Marshal([]any{x["subgraphs"], x["nodes"], x["edges"]}); string(s) != "[[1],[3,4],[0]]" {
		t.Errorf("cluster_x lists %s", s)
	}
	if e := v["edges"].([]any)[0].(map[string]any); e["pos"] == nil {
		t.Errorf("edge has no pos: %v", e)
	}

	// Drawing operations in json.
	v = decode(t, renderResolved(t, `digraph { a }`, jsonDraw))
	a := v["objects"].([]any)[0].(map[string]any)
	draw, ldraw := a["_draw_"].([]any), a["_ldraw_"].([]any)
	if len(draw) != 2 || draw[0].(map[string]any)["color"] != "#000000" || draw[1].(map[string]any)["op"] != "e" || draw[1].(map[string]any)["rect"] == nil {
		t.Errorf("_draw_: %v", draw)
	}
	if len(ldraw) != 3 || ldraw[0].(map[string]any)["size"] != 14.0 || ldraw[0].(map[string]any)["face"] != "Times-Roman" || ldraw[2].(map[string]any)["text"] != "a" {
		t.Errorf("_ldraw_: %v", ldraw)
	}
	if v["xdotversion"] != "1.7" {
		t.Errorf("xdotversion %v", v["xdotversion"])
	}

	// Anonymous subgraphs, numbered with the edges.
	for _, tt := range []struct {
		src   string
		names []string
	}{
		{`digraph { {rank=same; a} {rank=same; b} subgraph x {c} {d} }`, []string{"%1", "%3", "%5", "x", "%7"}},
		{`digraph { a -> b -> c; d; { rank=max; d } }`, []string{"%1", "%7"}},
	} {
		v := decode(t, renderResolved(t, tt.src, graphJSON))
		got := []string{v["name"].(string)}
		for _, o := range v["objects"].([]any) {
			if n := o.(map[string]any)["name"].(string); n == "x" || strings.HasPrefix(n, "%") {
				got = append(got, n)
			}
		}
		if strings.Join(got, " ") != strings.Join(tt.names, " ") {
			t.Errorf("%s: names %v, want %v", tt.src, got, tt.names)
		}
	}

	// Escapes.
	out := renderResolved(t, `digraph { a [label=<<TABLE><TR><TD></TD></TR></TABLE>>]; b [label="say \"hi\""] }`, graphJSON)
	for _, want := range []string{`"label": "<TABLE><TR><TD><\/TD><\/TR><\/TABLE>"`, `"label": "say \"hi\""`} {
		if !strings.Contains(out, want) {
			t.Errorf("escapes: no %s in\n%s", want, out)
		}
	}

	// Several graphs: one document after another.
	out = renderResolved(t, "digraph A { a }\ndigraph B { b }", graphJSON)
	dec := json.NewDecoder(strings.NewReader(out))
	var names []string
	for dec.More() {
		var v map[string]any
		if err := dec.Decode(&v); err != nil {
			t.Fatal(err)
		}
		names = append(names, v["name"].(string))
	}
	if strings.Join(names, " ") != "A B" {
		t.Errorf("several graphs: %v", names)
	}
}
