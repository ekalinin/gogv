// SPDX-License-Identifier: MIT
// Provenance: gopuml internal/conformance/compare.go
// Provenance: doc https://graphviz.org/docs/outputs/plain/

package conformance

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"maps"
	"math"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// Thresholds of the upstream-compatibility spec. They may only be tightened
// over time.
const (
	SizeTolerance = 0.05
	MinSSIM       = 0.90
)

// CompareCanon checks that the canon output is byte-identical to the
// reference's.
func CompareCanon(ref, got []byte) error {
	if bytes.Equal(ref, got) {
		return nil
	}
	r := strings.Split(string(ref), "\n")
	g := strings.Split(string(got), "\n")
	for i := range min(len(r), len(g)) {
		if r[i] != g[i] {
			return fmt.Errorf("canon line %d differs: got %q, reference %q", i+1, g[i], r[i])
		}
	}
	return fmt.Errorf("canon has %d lines, reference has %d", len(g), len(r))
}

// Plain is what the comparison reads from plain output: the graph size in
// inches, the node names, the edges as tail and head names, and the centers
// of the nodes and of the edge labels.
type Plain struct {
	Width, Height float64
	Nodes         []string
	Edges         [][2]string
	// NodeCenters holds the center of each node, by name, and LabelCenters
	// the center of each edge label, by the key of its edge: its tail, its
	// head and its place among the edges between them, as "a->b#0".
	NodeCenters, LabelCenters map[string]Point
}

// Point is a position of plain output, in inches.
type Point struct{ X, Y float64 }

// parsePoint reads the coordinates of a plain statement.
func parsePoint(x, y string) (Point, bool) {
	px, err1 := strconv.ParseFloat(x, 64)
	py, err2 := strconv.ParseFloat(y, 64)
	return Point{px, py}, err1 == nil && err2 == nil
}

// ParsePlain reads the graph, node and edge statements of plain or
// plain-ext output, up to the stop statement. A quoted string may hold
// newlines, and a backslash before a newline continues it, so statements are
// split at the newlines outside quotes. Several graphs are not supported.
func ParsePlain(data []byte) (Plain, error) {
	p := Plain{NodeCenters: map[string]Point{}, LabelCenters: map[string]Point{}}
	stmts, err := plainStatements(string(data))
	if err != nil {
		return p, err
	}
	var sawGraph, sawStop bool
	between := map[[2]string]int{}
	for n, f := range stmts {
		if sawStop {
			return p, fmt.Errorf("plain statement %d: text after stop", n+1)
		}
		switch f[0] {
		case "graph":
			if len(f) < 4 {
				return p, fmt.Errorf("plain statement %d: short graph line", n+1)
			}
			w, err1 := strconv.ParseFloat(f[2], 64)
			h, err2 := strconv.ParseFloat(f[3], 64)
			if err1 != nil || err2 != nil {
				return p, fmt.Errorf("plain statement %d: bad graph size", n+1)
			}
			p.Width, p.Height, sawGraph = w, h, true
		case "node":
			if len(f) < 4 {
				return p, fmt.Errorf("plain statement %d: short node line", n+1)
			}
			center, ok := parsePoint(f[2], f[3])
			if !ok {
				return p, fmt.Errorf("plain statement %d: bad node center", n+1)
			}
			p.Nodes = append(p.Nodes, f[1])
			p.NodeCenters[f[1]] = center
		case "edge":
			if len(f) < 4 {
				return p, fmt.Errorf("plain statement %d: short edge line", n+1)
			}
			e := [2]string{f[1], f[2]}
			p.Edges = append(p.Edges, e)
			// The control points, then the label and its center when the
			// edge has one, then the style and the color.
			points, err := strconv.Atoi(f[3])
			if err != nil || points < 0 || len(f) < 4+2*points+2 {
				return p, fmt.Errorf("plain statement %d: bad edge points", n+1)
			}
			if rest := f[4+2*points:]; len(rest) >= 5 {
				center, ok := parsePoint(rest[1], rest[2])
				if !ok {
					return p, fmt.Errorf("plain statement %d: bad edge label center", n+1)
				}
				p.LabelCenters[fmt.Sprintf("%s->%s#%d", e[0], e[1], between[e])] = center
			}
			between[e]++
		case "stop":
			sawStop = true
		default:
			return p, fmt.Errorf("plain statement %d: unknown statement %q", n+1, f[0])
		}
	}
	if !sawGraph || !sawStop {
		return p, errors.New("plain output lacks its graph or stop line")
	}
	return p, nil
}

// plainStatements splits plain output into statements and their fields. A
// field is either a run of non-space characters or a double-quoted string,
// in which a backslash escapes the next character.
func plainStatements(text string) ([][]string, error) {
	var stmts [][]string
	var fields []string
	for i := 0; i < len(text); {
		switch text[i] {
		case ' ', '\t', '\r':
			i++
		case '\n':
			if len(fields) > 0 {
				stmts = append(stmts, fields)
				fields = nil
			}
			i++
		case '"':
			var b strings.Builder
			i++
			for ; i < len(text) && text[i] != '"'; i++ {
				if text[i] == '\\' && i+1 < len(text) {
					i++
					if text[i] == '\n' {
						continue // a continued line
					}
				}
				b.WriteByte(text[i])
			}
			if i >= len(text) {
				return nil, errors.New("unterminated string")
			}
			i++
			fields = append(fields, b.String())
		default:
			j := i
			for j < len(text) && text[j] != ' ' && text[j] != '\n' {
				j++
			}
			fields = append(fields, text[i:j])
			i = j
		}
	}
	if len(fields) > 0 {
		stmts = append(stmts, fields)
	}
	return stmts, nil
}

// ComparePlain checks that got has the reference's node names and edges, as
// multisets, and a graph width and height each within tolerance (a fraction)
// of the reference's.
func ComparePlain(ref, got Plain, tolerance float64) error {
	var errs []error
	if !sameMultiset(ref.Nodes, got.Nodes) {
		errs = append(errs, fmt.Errorf("nodes %v differ from the reference's %v", sorted(got.Nodes), sorted(ref.Nodes)))
	}
	edgeKeys := func(es [][2]string) []string {
		keys := make([]string, len(es))
		for i, e := range es {
			keys[i] = e[0] + "->" + e[1]
		}
		return keys
	}
	if !sameMultiset(edgeKeys(ref.Edges), edgeKeys(got.Edges)) {
		errs = append(errs, fmt.Errorf("edges %v differ from the reference's %v", sorted(edgeKeys(got.Edges)), sorted(edgeKeys(ref.Edges))))
	}
	errs = append(errs, within("plain width", ref.Width, got.Width, tolerance), within("plain height", ref.Height, got.Height, tolerance))
	return errors.Join(errs...)
}

// ProxyTolerance is the largest deviation of the proxy of gopuml's contract,
// as a fraction of the reference's graph size.
const ProxyTolerance = 0.02

// ProxyDeviation returns the largest deviation of got from the reference
// for the proxy of gopuml's contract, and the object that has it: the
// relative difference of the graph width and height, and the distance of
// each node and edge label center from the reference's on each axis, as a
// fraction of the reference's graph width or height. A node or label that
// only one side has deviates by 1.
func ProxyDeviation(ref, got Plain) (float64, string) {
	worst, object := 0.0, ""
	note := func(d float64, what string) {
		if d > worst {
			worst, object = d, what
		}
	}
	relative := func(d, size float64) float64 {
		if size == 0 {
			return 0
		}
		return math.Abs(d) / size
	}
	note(relative(got.Width-ref.Width, ref.Width), "graph width")
	note(relative(got.Height-ref.Height, ref.Height), "graph height")
	centers := func(kind string, refCenters, gotCenters map[string]Point) {
		for _, name := range slices.Sorted(maps.Keys(refCenters)) {
			r := refCenters[name]
			g, ok := gotCenters[name]
			if !ok {
				note(1, kind+" "+name+" missing")
				continue
			}
			note(relative(g.X-r.X, ref.Width), kind+" "+name)
			note(relative(g.Y-r.Y, ref.Height), kind+" "+name)
		}
		for _, name := range slices.Sorted(maps.Keys(gotCenters)) {
			if _, ok := refCenters[name]; !ok {
				note(1, kind+" "+name+" not in the reference")
			}
		}
	}
	centers("node", ref.NodeCenters, got.NodeCenters)
	centers("label of", ref.LabelCenters, got.LabelCenters)
	return worst, object
}

// PassesProxy reports whether a case passes the proxy of gopuml's contract,
// from its deviation and the SSIM of its svg.
func PassesProxy(deviation, ssim float64) bool {
	return deviation <= ProxyTolerance && ssim >= MinSSIM
}

func sameMultiset(a, b []string) bool {
	return slices.Equal(sorted(a), sorted(b))
}

func sorted(s []string) []string {
	c := slices.Clone(s)
	slices.Sort(c)
	return c
}

func within(name string, ref, got, tolerance float64) error {
	if math.Abs(got-ref) > tolerance*ref {
		return fmt.Errorf("%s %g is not within %.0f%% of the reference %g", name, got, tolerance*100, ref)
	}
	return nil
}

var svgRootSize = regexp.MustCompile(`<svg\b[^>]*?\swidth="([0-9.]+)(?:pt|px)?"[^>]*?\sheight="([0-9.]+)(?:pt|px)?"`)

// SVGSize returns the width and height declared on the SVG root element, in
// points rounded to whole numbers.
func SVGSize(svg []byte) (width, height int, ok bool) {
	m := svgRootSize.FindSubmatch(svg)
	if m == nil {
		return 0, 0, false
	}
	w, err1 := strconv.ParseFloat(string(m[1]), 64)
	h, err2 := strconv.ParseFloat(string(m[2]), 64)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return int(math.Round(w)), int(math.Round(h)), true
}

// JudgeSVG applies the svg layer of the contract to measured values: the
// width and height each within SizeTolerance of the reference's, and an SSIM
// of at least MinSSIM.
func JudgeSVG(refW, refH, gotW, gotH int, ssim float64) error {
	errs := []error{
		within("svg width", float64(refW), float64(gotW), SizeTolerance),
		within("svg height", float64(refH), float64(gotH), SizeTolerance),
	}
	if ssim < MinSSIM {
		errs = append(errs, fmt.Errorf("SSIM %.4f is below %.2f", ssim, MinSSIM))
	}
	return errors.Join(errs...)
}

// CompareSVG rasterizes both documents at the reference's size with r and
// applies JudgeSVG. It returns the SSIM it measured.
func CompareSVG(r Rasterizer, ref, got []byte) (float64, error) {
	refW, refH, ok1 := SVGSize(ref)
	gotW, gotH, ok2 := SVGSize(got)
	if !ok1 || !ok2 {
		return 0, errors.New("svg size not found")
	}
	refImg, err := r.Rasterize(ref, refW, refH)
	if err != nil {
		return 0, fmt.Errorf("reference: %w", err)
	}
	gotImg, err := r.Rasterize(got, refW, refH)
	if err != nil {
		return 0, err
	}
	score := ImageSSIM(refImg, gotImg)
	return score, JudgeSVG(refW, refH, gotW, gotH, score)
}

// ImageSSIM returns the structural similarity of got against ref after
// resampling got to the reference size.
func ImageSSIM(ref, got image.Image) float64 {
	b := ref.Bounds()
	return SSIM(gray(ref, b.Dx(), b.Dy()), gray(got, b.Dx(), b.Dy()), b.Dx(), b.Dy())
}

// gray converts img to grayscale luminance in [0,255], composited onto a
// white background and resampled (bilinear) to w x h.
func gray(img image.Image, w, h int) []float64 {
	b := img.Bounds()
	sw, sh := b.Dx(), b.Dy()
	src := make([]float64, sw*sh)
	for y := range sh {
		for x := range sw {
			r, g, bl, a := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			// Premultiplied colors over white: c + (1 - alpha) * white.
			white := float64(0xffff - a)
			lum := 0.299*(float64(r)+white) + 0.587*(float64(g)+white) + 0.114*(float64(bl)+white)
			src[y*sw+x] = lum / 0xffff * 255
		}
	}
	if sw == w && sh == h {
		return src
	}
	dst := make([]float64, w*h)
	for y := range h {
		fy := (float64(y)+0.5)*float64(sh)/float64(h) - 0.5
		y0 := clamp(int(math.Floor(fy)), sh)
		y1 := clamp(y0+1, sh)
		ty := fy - math.Floor(fy)
		for x := range w {
			fx := (float64(x)+0.5)*float64(sw)/float64(w) - 0.5
			x0 := clamp(int(math.Floor(fx)), sw)
			x1 := clamp(x0+1, sw)
			tx := fx - math.Floor(fx)
			top := src[y0*sw+x0]*(1-tx) + src[y0*sw+x1]*tx
			bottom := src[y1*sw+x0]*(1-tx) + src[y1*sw+x1]*tx
			dst[y*w+x] = top*(1-ty) + bottom*ty
		}
	}
	return dst
}

func clamp(v, n int) int { return min(max(v, 0), n-1) }

// SSIM returns the mean structural similarity of two grayscale images of
// size w x h, using an 11x11 Gaussian window with sigma 1.5 (Wang et al.).
func SSIM(a, b []float64, w, h int) float64 {
	const (
		c1 = (0.01 * 255) * (0.01 * 255)
		c2 = (0.03 * 255) * (0.03 * 255)
	)
	blur := gaussianBlur(w, h)
	mul := func(x, y []float64) []float64 {
		out := make([]float64, len(x))
		for i := range x {
			out[i] = x[i] * y[i]
		}
		return out
	}
	muA, muB := blur(a), blur(b)
	sAA, sBB, sAB := blur(mul(a, a)), blur(mul(b, b)), blur(mul(a, b))
	var sum float64
	for i := range muA {
		ma, mb := muA[i], muB[i]
		varA, varB, cov := sAA[i]-ma*ma, sBB[i]-mb*mb, sAB[i]-ma*mb
		sum += ((2*ma*mb + c1) * (2*cov + c2)) / ((ma*ma + mb*mb + c1) * (varA + varB + c2))
	}
	return sum / float64(len(muA))
}

// gaussianBlur returns a separable 11x11 Gaussian filter (sigma 1.5) for
// w x h images. Edges are handled by renormalizing the truncated kernel.
func gaussianBlur(w, h int) func([]float64) []float64 {
	const radius = 5
	var kernel [2*radius + 1]float64
	for i := range kernel {
		d := float64(i - radius)
		kernel[i] = math.Exp(-d * d / (2 * 1.5 * 1.5))
	}
	pass := func(in []float64, horizontal bool) []float64 {
		out := make([]float64, len(in))
		for y := range h {
			for x := range w {
				var sum, weight float64
				for k := -radius; k <= radius; k++ {
					sx, sy := x, y
					if horizontal {
						sx += k
					} else {
						sy += k
					}
					if sx < 0 || sx >= w || sy < 0 || sy >= h {
						continue
					}
					kw := kernel[k+radius]
					sum += kw * in[sy*w+sx]
					weight += kw
				}
				out[y*w+x] = sum / weight
			}
		}
		return out
	}
	return func(in []float64) []float64 { return pass(pass(in, true), false) }
}

// Rasterizer converts an SVG document to an image of the given size.
type Rasterizer interface {
	Rasterize(svg []byte, width, height int) (image.Image, error)
}

// Resvg rasterizes with the resvg command, mapping the generic families to
// the Liberation fonts, so that both sides of a comparison draw text with the
// same fonts on every host.
type Resvg struct {
	// Path is the resvg executable; "resvg" when empty.
	Path string
	// FontsDir, when set, is the only font source.
	FontsDir string
}

// Rasterize implements Rasterizer.
func (r Resvg) Rasterize(svg []byte, width, height int) (image.Image, error) {
	exe := r.Path
	if exe == "" {
		exe = "resvg"
	}
	args := []string{
		"--sans-serif-family", "Liberation Sans",
		"--serif-family", "Liberation Serif",
		"--monospace-family", "Liberation Mono",
		"-w", strconv.Itoa(width), "-h", strconv.Itoa(height),
	}
	if r.FontsDir != "" {
		args = append(args, "--skip-system-fonts", "--use-fonts-dir", r.FontsDir)
	}
	cmd := exec.Command(exe, append(args, "-", "-c")...)
	cmd.Stdin = bytes.NewReader(svg)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("resvg: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return png.Decode(bytes.NewReader(out))
}
