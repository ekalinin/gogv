// SPDX-License-Identifier: MIT
// Provenance: original

// Command metricscheck measures, as a black box, how wide the reference dot
// sets each printable ASCII character in its Times, Helvetica and Courier
// metrics, and reports the differences with the tables of
// internal/textmetrics into .cache/metricscheck/report.md.
//
// Each character is measured as a plaintext node whose label repeats it 50
// times, read back from -Tplain. The measured widths only go into the report
// under .cache: they are never written into the module (design D5).
package main

import (
	"bytes"
	"flag"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ekalinin/gogv/internal/textmetrics"
)

const (
	repeat   = 50
	fontsize = 14
	// margins is the horizontal margin of a plaintext node of the reference
	// at width 0, in points (8 points on each side).
	margins = 16
	// tolerance is the difference per character, in points, that the report
	// counts as a mismatch.
	tolerance = 0.01
)

var families = []struct {
	fontname string
	family   textmetrics.Family
}{
	{"Times-Roman", textmetrics.Serif},
	{"Helvetica", textmetrics.Sans},
	{"Courier", textmetrics.Mono},
}

func main() {
	image := flag.String("image", "gogv-reference", "reference image")
	out := flag.String("out", ".cache/metricscheck/report.md", "report file")
	flag.Parse()
	if err := run(*image, *out); err != nil {
		fmt.Fprintln(os.Stderr, "metricscheck:", err)
		os.Exit(1)
	}
}

func run(image, out string) error {
	var src bytes.Buffer
	src.WriteString("digraph {\n\tnode [shape=plaintext, width=0, height=0];\n")
	for fi, f := range families {
		for c := byte(0x20); c <= 0x7e; c++ {
			fmt.Fprintf(&src, "\tn%d_%d [fontname=%q, label=\"%s\"];\n", fi, c, f.fontname, strings.Repeat(escape(c), repeat))
		}
	}
	src.WriteString("}\n")

	cmd := exec.Command("docker", "run", "--rm", "-i", "--platform", "linux/amd64", image, "dot", "-Tplain")
	cmd.Stdin = &src
	cmd.Stderr = os.Stderr
	plain, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("reference dot: %w", err)
	}
	widths := map[string]float64{} // node name -> width in points
	for line := range strings.SplitSeq(string(plain), "\n") {
		f := strings.Fields(line)
		if len(f) > 4 && f[0] == "node" {
			w, err := strconv.ParseFloat(f[4], 64)
			if err != nil {
				return fmt.Errorf("width of %s: %w", f[1], err)
			}
			widths[f[1]] = w * 72
		}
	}

	var report strings.Builder
	report.WriteString("# Text metrics check\n\n")
	fmt.Fprintf(&report, "Width per character at %d points: reference dot (measured over %d characters) against the tables of internal/textmetrics. A mismatch differs by more than %.2f points.\n\n", fontsize, repeat, tolerance)
	report.WriteString("| Family | Characters | Mismatches | Largest difference |\n|---|---|---|---|\n")
	var details strings.Builder
	for fi, f := range families {
		face := textmetrics.Face{Family: f.family}
		var count, mismatches int
		var largest float64
		for c := byte(0x20); c <= 0x7e; c++ {
			ref, ok := widths[fmt.Sprintf("n%d_%d", fi, c)]
			if !ok {
				return fmt.Errorf("no width for %q in %s", c, f.fontname)
			}
			ref = (ref - margins) / repeat
			got := face.Width(string(c), fontsize)
			diff := got - ref
			count++
			largest = math.Max(largest, math.Abs(diff))
			if math.Abs(diff) > tolerance {
				mismatches++
				fmt.Fprintf(&details, "| %s | %q | %.4f | %.4f |\n", f.fontname, c, ref, got)
			}
		}
		fmt.Fprintf(&report, "| %s | %d | %d | %.4f |\n", f.fontname, count, mismatches, largest)
	}
	if details.Len() > 0 {
		report.WriteString("\n## Mismatches\n\n| Family | Character | Reference | gogv |\n|---|---|---|---|\n")
		report.WriteString(details.String())
	}

	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(out, []byte(report.String()), 0o644); err != nil {
		return err
	}
	fmt.Print(report.String())
	return nil
}

// escape writes c inside a double-quoted DOT string.
func escape(c byte) string {
	switch c {
	case '"':
		return `\"`
	case '\\':
		return `\\`
	}
	return string(c)
}
