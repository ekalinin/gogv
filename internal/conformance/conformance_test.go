// SPDX-License-Identifier: MIT
// Provenance: original

//go:build conformance

package conformance_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ekalinin/gogv/internal/cli"
	"github.com/ekalinin/gogv/internal/conformance"
)

// root is the repository root, relative to this package.
const root = "../.."

// render runs the command line in this process, as `dot -T<format>` with
// the source on stdin.
func render(src []byte, format string) conformance.Result {
	var stdout, stderr bytes.Buffer
	status := cli.Run(context.Background(), []string{"-T" + format}, bytes.NewReader(src), &stdout, &stderr, "test")
	return conformance.Result{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), Status: status}
}

// TestConformance runs every corpus against its recorded references and
// writes conformance-report.md. It fails when a case fails: known
// differences and cases not supported yet do not fail it.
func TestConformance(t *testing.T) {
	known, err := conformance.ReadKnownDifferences(filepath.Join(root, "testdata/conformance/known-differences.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	rasterizer := conformance.Resvg{
		Path:     filepath.Join(root, ".cache/tools/resvg"),
		FontsDir: filepath.Join(root, ".cache/fonts/liberation"),
	}
	if _, err := os.Stat(rasterizer.Path); err != nil {
		t.Fatal("resvg not fetched: run tools/reference/fetch-tools.sh")
	}
	reportPath, err := filepath.Abs(filepath.Join(root, "conformance-report.md"))
	if err != nil {
		t.Fatal(err)
	}
	var all conformance.Report
	var failures []string
	for _, c := range conformance.Corpora {
		cases, err := conformance.LoadCases(filepath.Join(root, c.Cases))
		if err != nil || len(cases) == 0 {
			if c.Optional {
				t.Logf("corpus %s: not collected, skipped", c.Name)
				continue
			}
			t.Fatalf("corpus %s: %v", c.Name, err)
		}
		// A harvested corpus records where it comes from.
		source, _ := os.ReadFile(filepath.Join(root, c.Cases, "SOURCE"))
		all.Corpora = append(all.Corpora, conformance.CorpusRun{Name: c.Name, Cases: len(cases), Source: strings.TrimSpace(string(source))})
		report, err := conformance.Run(cases, filepath.Join(root, c.References), known, render, rasterizer)
		all.Results = append(all.Results, report.Results...)
		if err != nil {
			failures = append(failures, c.Name+": "+err.Error())
		}
	}
	cliCases, err := conformance.ReadCLICases(filepath.Join(root, "testdata/conformance/cli/cases.json"))
	if err != nil {
		t.Fatal(err)
	}
	cliRef, err := conformance.ReadCLIReference(filepath.Join(root, "testdata/conformance/cli/reference.json"))
	if err != nil {
		t.Fatal(err)
	}
	// Relative file names of the cases resolve in an empty directory.
	t.Chdir(t.TempDir())
	for _, c := range cliCases {
		var stdout, stderr bytes.Buffer
		status := cli.Run(context.Background(), c.Args, strings.NewReader(c.Stdin), &stdout, &stderr, "test")
		res := conformance.JudgeCLI(c, cliRef[c.ID], conformance.CLIResult{Stdout: stdout.String(), Stderr: stderr.String(), Status: status})
		all.Results = append(all.Results, res)
		if res.Outcome == conformance.Failed {
			failures = append(failures, c.ID+": "+res.Problems["cli"])
		}
	}
	all.Corpora = append(all.Corpora, conformance.CorpusRun{Name: "cli", Cases: len(cliCases)})
	if err := os.WriteFile(reportPath, []byte(all.Markdown()), 0o644); err != nil {
		t.Fatal(err)
	}
	if len(failures) > 0 {
		t.Fatal(strings.Join(failures, "\n"))
	}
}
