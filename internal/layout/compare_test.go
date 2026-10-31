// SPDX-License-Identifier: MIT
// Provenance: original

package layout

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/ekalinin/gogv/internal/attrs"
	"github.com/ekalinin/gogv/internal/conformance"
	"github.com/ekalinin/gogv/internal/dotlang"
)

// refNode is a node position read from the reference's plain output.
type refNode struct {
	x, y, w, h float64
}

func readPlainNodes(plain []byte) map[string]refNode {
	out := map[string]refNode{}
	for _, line := range strings.Split(string(plain), "\n") {
		f, err := plainFieldsForTest(line)
		if err != nil || len(f) < 6 || f[0] != "node" {
			continue
		}
		num := func(s string) float64 { v, _ := strconv.ParseFloat(s, 64); return v * 72 }
		out[f[1]] = refNode{num(f[2]), num(f[3]), num(f[4]), num(f[5])}
	}
	return out
}

func plainFieldsForTest(line string) ([]string, error) {
	var fields []string
	for i := 0; i < len(line); {
		switch line[i] {
		case ' ':
			i++
		case '"':
			var b strings.Builder
			i++
			for ; i < len(line) && line[i] != '"'; i++ {
				if line[i] == '\\' && i+1 < len(line) {
					i++
				}
				b.WriteByte(line[i])
			}
			i++
			fields = append(fields, b.String())
		default:
			j := i
			for j < len(line) && line[j] != ' ' {
				j++
			}
			fields = append(fields, line[i:j])
			i = j
		}
	}
	return fields, nil
}

// rankOrders groups node names by y and orders them by x.
func rankOrders(pos map[string][2]float64) string {
	byY := map[int64][]string{}
	for name, p := range pos {
		k := int64(math.Round(p[1] * 10))
		byY[k] = append(byY[k], name)
	}
	var ys []int64
	for y := range byY {
		ys = append(ys, y)
	}
	sort.Slice(ys, func(i, j int) bool { return ys[i] > ys[j] })
	var parts []string
	for _, y := range ys {
		names := byY[y]
		sort.Slice(names, func(i, j int) bool {
			if pos[names[i]][0] != pos[names[j]][0] {
				return pos[names[i]][0] < pos[names[j]][0]
			}
			return names[i] < names[j]
		})
		parts = append(parts, strings.Join(names, ","))
	}
	return strings.Join(parts, " | ")
}

// TestCompareNodePositions reports how close node sizes, rank orders and
// positions are to the reference. It is a development aid: it fails only
// when a case cannot be laid out.
func TestCompareNodePositions(t *testing.T) {
	type corpus struct{ cases, refs string }
	var sizeOK, orderOK, posOK, total int
	var report strings.Builder
	for _, c := range []corpus{
		{"../../testdata/conformance/cases", "../../testdata/conformance/reference"},
		{"../../.cache/gopuml-dot", "../../.cache/reference/gopuml-dot"},
	} {
		cases, err := conformance.LoadCases(c.cases)
		if err != nil {
			continue
		}
		for _, cs := range cases {
			ref, err := conformance.ReadReference(c.refs, cs.ID)
			if err != nil || ref["plain"].Status != 0 || bytes.Contains(cs.Source, []byte("cluster")) {
				continue
			}
			file, _, err := dotlang.Parse(string(cs.Source))
			if err != nil || len(file.Graphs) != 1 {
				continue
			}
			g := dotlang.Build(file.Graphs[0])
			if len(attrs.Check(g)) > 0 && os.Getenv("ALLCASES") == "" {
				continue // not supported yet: a later slice
			}
			lr := readParams(g).lr
			res, err := Layout(context.Background(), g)
			if err != nil {
				t.Fatalf("%s: %v", cs.ID, err)
			}
			refNodes := readPlainNodes(ref["plain"].Stdout)
			if len(refNodes) != len(res.Nodes) || len(res.Nodes) == 0 {
				continue
			}
			total++
			ours := map[string][2]float64{}
			theirs := map[string][2]float64{}
			sizes := true
			for _, n := range res.Nodes {
				r := refNodes[n.Node.Name]
				ours[n.Node.Name] = [2]float64{n.X, n.Y}
				theirs[n.Node.Name] = [2]float64{r.x, r.y}
				if lr {
					// Ranks are columns: group by x, order by y.
					ours[n.Node.Name] = [2]float64{n.Y, -n.X}
					theirs[n.Node.Name] = [2]float64{r.y, -r.x}
				}
				if math.Abs(n.W-r.w) > 0.02 || math.Abs(n.H-r.h) > 0.02 {
					sizes = false
					fmt.Fprintf(&report, "%s: size of %s %.2fx%.2f, reference %.2fx%.2f\n", cs.ID, n.Node.Name, n.W, n.H, r.w, r.h)
				}
			}
			if sizes {
				sizeOK++
			}
			if rankOrders(ours) == rankOrders(theirs) {
				orderOK++
				// Positions up to a translation.
				first := res.Nodes[0].Node.Name
				dx := theirs[first][0] - ours[first][0]
				dy := theirs[first][1] - ours[first][1]
				same := true
				for name, p := range ours {
					if math.Abs(p[0]+dx-theirs[name][0]) > 0.6 || math.Abs(p[1]+dy-theirs[name][1]) > 0.6 {
						same = false
					}
				}
				if same {
					posOK++
				} else if os.Getenv("SHOWPOS") != "" {
					fmt.Fprintf(&report, "%s: positions differ\n  ours   %v\n  theirs %v\n", cs.ID, ours, theirs)
				}
			} else if os.Getenv("SHOWORDER") != "" {
				fmt.Fprintf(&report, "%s: order\n  ours   %s\n  theirs %s\n", cs.ID, rankOrders(ours), rankOrders(theirs))
			}
		}
	}
	t.Logf("%d cases: sizes %d, rank orders %d, positions %d\n%s", total, sizeOK, orderOK, posOK, report.String())
}

// TestCompareRandom compares rank orders and positions with the reference on
// the random graphs of .cache/random (r*.gv, with the reference's r*.plain).
func TestCompareRandom(t *testing.T) {
	files, _ := filepathGlob("../../.cache/random/r*.gv")
	if len(files) == 0 {
		t.Skip("no random graphs")
	}
	var orderOK, posOK, total int
	var report strings.Builder
	for _, f := range files {
		src, _ := os.ReadFile(f)
		plain, err := os.ReadFile(strings.TrimSuffix(f, ".gv") + ".plain")
		if err != nil {
			continue
		}
		file, _, err := dotlang.Parse(string(src))
		if err != nil {
			t.Fatal(err)
		}
		res, err := Layout(context.Background(), dotlang.Build(file.Graphs[0]))
		if err != nil {
			t.Fatal(err)
		}
		ref := readPlainNodes(plain)
		total++
		ours := map[string][2]float64{}
		theirs := map[string][2]float64{}
		for _, n := range res.Nodes {
			ours[n.Node.Name] = [2]float64{n.X, n.Y}
			theirs[n.Node.Name] = [2]float64{ref[n.Node.Name].x, ref[n.Node.Name].y}
		}
		if rankOrders(ours) != rankOrders(theirs) {
			if os.Getenv("SHOWORDER") != "" {
				fmt.Fprintf(&report, "%s: order\n  ours   %s\n  theirs %s\n", f, rankOrders(ours), rankOrders(theirs))
			}
			continue
		}
		orderOK++
		first := res.Nodes[0].Node.Name
		dx, dy := theirs[first][0]-ours[first][0], theirs[first][1]-ours[first][1]
		same := true
		for name, p := range ours {
			if math.Abs(p[0]+dx-theirs[name][0]) > 0.6 || math.Abs(p[1]+dy-theirs[name][1]) > 0.6 {
				same = false
			}
		}
		if same {
			posOK++
		}
	}
	t.Logf("%d random graphs: rank orders %d, positions %d\n%s", total, orderOK, posOK, report.String())
}
func filepathGlob(p string) ([]string, error) { return filepath.Glob(p) }

// TestTuneMincross scores combinations of mincross options against the
// reference on all corpora. Run with TUNE=1.
func TestTuneMincross(t *testing.T) {
	if os.Getenv("TUNE") == "" {
		t.Skip("set TUNE=1")
	}
	saved := mincrossOpts
	defer func() { mincrossOpts = saved }()
	type score struct {
		o             mincrossOptions
		curated, rand int
	}
	var scores []score
	for _, tie := range []int{0, 1, 2} {
		for _, eq := range []int{0, 1, 2, 3} {
			for _, keep := range []bool{false, true} {
				for _, it := range []bool{false, true} {
					for _, mq := range []int{8, 1 << 30} {
						mincrossOpts = mincrossOptions{tieReverse: tie, equalSwap: eq, keepEqual: keep, initialTranspose: it, minQuit: mq}
						c := countOrders(t, false)
						r := countOrders(t, true)
						scores = append(scores, score{mincrossOpts, c, r})
					}
				}
			}
		}
	}
	sort.Slice(scores, func(i, j int) bool { return scores[i].curated+scores[i].rand > scores[j].curated+scores[j].rand })
	for _, s := range scores[:12] {
		t.Logf("%+v curated %d random %d", s.o, s.curated, s.rand)
	}
}

// countOrders counts the cases whose rank orders match the reference.
func countOrders(t *testing.T, random bool) int {
	var ok int
	check := func(src, plain []byte) {
		file, _, err := dotlang.Parse(string(src))
		if err != nil || len(file.Graphs) != 1 {
			return
		}
		g := dotlang.Build(file.Graphs[0])
		if len(attrs.Check(g)) > 0 {
			return
		}
		res, err := Layout(context.Background(), g)
		if err != nil || len(res.Nodes) == 0 {
			return
		}
		ref := readPlainNodes(plain)
		lr := readParams(g).lr
		ours := map[string][2]float64{}
		theirs := map[string][2]float64{}
		for _, n := range res.Nodes {
			r := ref[n.Node.Name]
			ours[n.Node.Name] = [2]float64{n.X, n.Y}
			theirs[n.Node.Name] = [2]float64{r.x, r.y}
			if lr {
				ours[n.Node.Name] = [2]float64{n.Y, -n.X}
				theirs[n.Node.Name] = [2]float64{r.y, -r.x}
			}
		}
		if rankOrders(ours) == rankOrders(theirs) {
			ok++
		}
	}
	if random {
		files, _ := filepath.Glob("../../.cache/random/r*.gv")
		for _, f := range files {
			src, _ := os.ReadFile(f)
			plain, _ := os.ReadFile(strings.TrimSuffix(f, ".gv") + ".plain")
			check(src, plain)
		}
		return ok
	}
	for _, c := range [][2]string{
		{"../../testdata/conformance/cases", "../../testdata/conformance/reference"},
		{"../../.cache/gopuml-dot", "../../.cache/reference/gopuml-dot"},
	} {
		cases, _ := conformance.LoadCases(c[0])
		for _, cs := range cases {
			ref, err := conformance.ReadReference(c[1], cs.ID)
			if err != nil || ref["plain"].Status != 0 {
				continue
			}
			check(cs.Source, ref["plain"].Stdout)
		}
	}
	return ok
}
