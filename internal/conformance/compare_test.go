// SPDX-License-Identifier: MIT
// Provenance: original

package conformance

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"strings"
	"testing"
)

func TestCompareCanon(t *testing.T) {
	ref := []byte("digraph {\n\ta -> b;\n}\n")
	if err := CompareCanon(ref, []byte("digraph {\n\ta -> b;\n}\n")); err != nil {
		t.Fatalf("identical canon: %v", err)
	}
	err := CompareCanon(ref, []byte("digraph {\n\tb -> a;\n}\n"))
	if err == nil || !strings.Contains(err.Error(), "line 2") {
		t.Fatalf("different canon: got %v, want an error on line 2", err)
	}
	if err := CompareCanon(ref, []byte("digraph {\n\ta -> b;\n}")); err == nil {
		t.Fatal("missing final newline: got no error")
	}
}

const refPlain = `graph 1 0.75 1.5
node a 0.375 1.25 0.75 0.5 a solid ellipse black lightgrey
node "my node" 0.375 0.25 0.75 0.5 "my node" solid ellipse black lightgrey
edge a "my node" 4 0.375 0.99 0.375 0.88 0.375 0.75 0.375 0.62 solid black
stop
`

func TestParsePlain(t *testing.T) {
	p, err := ParsePlain([]byte(refPlain))
	if err != nil {
		t.Fatal(err)
	}
	if p.Width != 0.75 || p.Height != 1.5 {
		t.Errorf("size = %g x %g, want 0.75 x 1.5", p.Width, p.Height)
	}
	if strings.Join(p.Nodes, ",") != "a,my node" {
		t.Errorf("nodes = %q", p.Nodes)
	}
	if len(p.Edges) != 1 || p.Edges[0] != [2]string{"a", "my node"} {
		t.Errorf("edges = %q", p.Edges)
	}
	if _, err := ParsePlain([]byte("graph 1 1 1\n")); err == nil {
		t.Error("output without stop: got no error")
	}
}

func TestComparePlain(t *testing.T) {
	ref, err := ParsePlain([]byte(refPlain))
	if err != nil {
		t.Fatal(err)
	}
	larger := ref
	larger.Width, larger.Height = ref.Width*1.03, ref.Height*1.03
	if err := ComparePlain(ref, larger, SizeTolerance); err != nil {
		t.Errorf("size +3%%: %v", err)
	}

	tooLarge := ref
	tooLarge.Height = ref.Height * 1.06
	if err := ComparePlain(ref, tooLarge, SizeTolerance); err == nil || !strings.Contains(err.Error(), "plain height") {
		t.Errorf("height +6%%: got %v", err)
	}

	otherNodes := ref
	otherNodes.Nodes = []string{"a", "b"}
	if err := ComparePlain(ref, otherNodes, SizeTolerance); err == nil || !strings.Contains(err.Error(), "nodes") {
		t.Errorf("other nodes: got %v", err)
	}

	reversed := ref
	reversed.Edges = [][2]string{{"my node", "a"}}
	if err := ComparePlain(ref, reversed, SizeTolerance); err == nil || !strings.Contains(err.Error(), "edges") {
		t.Errorf("reversed edge: got %v", err)
	}
}

func TestJudgeSVG(t *testing.T) {
	if err := JudgeSVG(100, 200, 103, 206, 0.94); err != nil {
		t.Errorf("SSIM 0.94, size +3%%: %v", err)
	}
	if err := JudgeSVG(100, 200, 100, 200, 0.82); err == nil || !strings.Contains(err.Error(), "SSIM") {
		t.Errorf("SSIM 0.82: got %v", err)
	}
	if err := JudgeSVG(100, 200, 106, 200, 0.99); err == nil || !strings.Contains(err.Error(), "svg width") {
		t.Errorf("width +6%%: got %v", err)
	}
}

func TestSVGSize(t *testing.T) {
	w, h, ok := SVGSize([]byte(`<svg width="62pt" height="116pt" viewBox="0.00 0.00 62.00 116.00">`))
	if !ok || w != 62 || h != 116 {
		t.Fatalf("SVGSize = %d, %d, %v; want 62, 116, true", w, h, ok)
	}
}

// fakeRasterizer draws a filled rectangle, moved to the right when the
// document contains "shifted", so that tests control the similarity.
type fakeRasterizer struct{}

func (fakeRasterizer) Rasterize(svg []byte, width, height int) (image.Image, error) {
	img := image.NewGray(image.Rect(0, 0, width, height))
	for i := range img.Pix {
		img.Pix[i] = 0xff
	}
	x0 := 10
	if strings.Contains(string(svg), "shifted") {
		x0 = 60
	}
	for y := 10; y < 40; y++ {
		for x := x0; x < x0+30; x++ {
			img.SetGray(x, y, color.Gray{})
		}
	}
	return img, nil
}

func TestCompareSVG(t *testing.T) {
	ref := []byte(`<svg width="100pt" height="50pt">`)
	same := []byte(`<svg width="102pt" height="51pt">`)
	if score, err := CompareSVG(fakeRasterizer{}, ref, same); err != nil || score < 0.999 {
		t.Errorf("same drawing: score %.4f, err %v", score, err)
	}
	shifted := []byte(`<svg width="100pt" height="50pt"> shifted`)
	if score, err := CompareSVG(fakeRasterizer{}, ref, shifted); err == nil {
		t.Errorf("shifted drawing: score %.4f, want an SSIM error", score)
	}
}

// TestResvg rasterizes with the pinned resvg of tools/reference/fetch-tools.sh
// when it is present.
func TestResvg(t *testing.T) {
	const exe = "../../.cache/tools/resvg"
	if _, err := os.Stat(exe); err != nil {
		t.Skip("resvg not fetched; run tools/reference/fetch-tools.sh")
	}
	doc := func(x int) []byte {
		return []byte(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="100pt" height="50pt" viewBox="0 0 100 50">`+
			`<rect x="%d" y="10" width="30" height="30" fill="black"/></svg>`, x))
	}
	r := Resvg{Path: exe}
	if score, err := CompareSVG(r, doc(10), doc(10)); err != nil || score < 0.999 {
		t.Errorf("same document: score %.4f, err %v", score, err)
	}
	if score, err := CompareSVG(r, doc(10), doc(60)); err == nil {
		t.Errorf("moved rectangle: score %.4f, want an SSIM error", score)
	}
}
