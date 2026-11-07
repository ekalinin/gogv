// SPDX-License-Identifier: MIT
// Provenance: original

package conformance

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const okSVG = `<svg width="100pt" height="50pt">`

func okReference() Reference {
	return Reference{
		"canon": {Stdout: []byte("digraph {\n\ta;\n}\n")},
		"plain": {Stdout: []byte(refPlain)},
		"svg":   {Stdout: []byte(okSVG)},
	}
}

func clone(ref Reference) map[string]Result {
	got := map[string]Result{}
	for f, r := range ref {
		got[f] = r
	}
	return got
}

func TestJudgeOutcomes(t *testing.T) {
	known := map[string]KnownDifference{
		"dot-layout/listed": {ID: "dot-layout/listed", Reason: "order within a rank", Change: "add-x"},
	}
	ref := okReference()

	shifted := clone(ref)
	shifted["svg"] = Result{Stdout: []byte(okSVG + " shifted")}

	compassWarning := clone(ref)
	for _, f := range Formats {
		r := compassWarning[f]
		r.Stderr = []byte("Warning: edge a->b: tailport=n is not supported yet, using center\n")
		compassWarning[f] = r
	}

	otherNodes := clone(ref)
	otherNodes["plain"] = Result{Stdout: []byte(strings.Replace(refPlain, `"my node"`, "b", -1))}

	tests := []struct {
		name    string
		id      string
		got     map[string]Result
		want    Outcome
		formats map[string]Outcome
	}{
		{"same outputs", "dot-layout/same", clone(ref), Passed,
			map[string]Outcome{"canon": Passed, "plain": Passed, "svg": Passed}},
		{"listed failure", "dot-layout/listed", shifted, KnownDiff,
			map[string]Outcome{"canon": Passed, "plain": Passed, "svg": KnownDiff}},
		{"compass point before its slice", "dot-language/compass", compassWarning, NotSupported,
			map[string]Outcome{"canon": NotSupported, "plain": NotSupported, "svg": NotSupported}},
		{"unlisted svg failure", "dot-layout/shifted", shifted, Failed,
			map[string]Outcome{"canon": Passed, "plain": Passed, "svg": Failed}},
		{"unlisted plain failure", "dot-layout/nodes", otherNodes, Failed,
			map[string]Outcome{"canon": Passed, "plain": Failed, "svg": Passed}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Case{ID: tt.id, Capability: strings.Split(tt.id, "/")[0]}
			res := Judge(c, ref, tt.got, known, fakeRasterizer{})
			if res.Outcome != tt.want {
				t.Errorf("outcome %q, want %q (problems %v)", res.Outcome, tt.want, res.Problems)
			}
			for f, want := range tt.formats {
				if res.Formats[f] != want {
					t.Errorf("%s: outcome %q, want %q", f, res.Formats[f], want)
				}
			}
		})
	}
}

func TestJudgeErrors(t *testing.T) {
	errRef := Reference{}
	for _, f := range Formats {
		errRef[f] = Result{Stderr: []byte("Error: <stdin>: syntax error in line 1 near '}'\n"), Status: 1}
	}
	same := map[string]Result{}
	other := map[string]Result{}
	for _, f := range Formats {
		same[f] = Result{Stderr: []byte("Error: in.gv: syntax error in line 1 near '}'\n"), Status: 1}
		other[f] = Result{Stderr: []byte("Error: <stdin>: syntax error in line 2 near 'a'\n"), Status: 1}
	}
	c := Case{ID: "dot-language/error", Capability: "dot-language"}
	if res := Judge(c, errRef, same, nil, fakeRasterizer{}); res.Outcome != Passed {
		t.Errorf("same error: outcome %q, problems %v", res.Outcome, res.Problems)
	}
	if res := Judge(c, errRef, other, nil, fakeRasterizer{}); res.Outcome != Failed {
		t.Errorf("other error: outcome %q, want failed", res.Outcome)
	}
}

func TestReportFailsOnUnlistedFailuresOnly(t *testing.T) {
	ref := okReference()
	shifted := clone(ref)
	shifted["svg"] = Result{Stdout: []byte(okSVG + " shifted")}
	known := map[string]KnownDifference{"dot-layout/listed": {ID: "dot-layout/listed", Reason: "r", Change: "c"}}
	var report Report
	for _, id := range []string{"dot-layout/listed", "dot-layout/shifted", "dot-layout/same"} {
		got := shifted
		if id == "dot-layout/same" {
			got = clone(ref)
		}
		report.Results = append(report.Results, Judge(Case{ID: id, Capability: "dot-layout"}, ref, got, known, fakeRasterizer{}))
	}
	failed := report.Failed()
	if len(failed) != 1 || failed[0].Case.ID != "dot-layout/shifted" {
		t.Fatalf("failed = %v, want only dot-layout/shifted", failed)
	}
	md := report.Markdown()
	for _, want := range []string{
		"| dot-layout | 3 | 1 | 1 | 0 | 1 |",
		"| svg | 3 | 1 | 1 | 0 | 1 |",
		"| dot-layout/shifted | failed |",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("report lacks %q:\n%s", want, md)
		}
	}
}

func TestReportCorpora(t *testing.T) {
	for _, tc := range []struct {
		name    string
		corpora []CorpusRun
		want    string
	}{
		{
			"harvested corpus from v0.3.0",
			[]CorpusRun{{"curated", 255, ""}, {"gopuml", 260, "gopuml v0.3.0"}, {"gopuml-clusters", 36, "gopuml v0.3.0"}, {"cli", 14, ""}},
			"- Cases: 0\n- Corpora: curated 255, gopuml 260 (gopuml v0.3.0), gopuml-clusters 36 (gopuml v0.3.0), cli 14\n\n",
		},
		{
			"harvested corpus not collected",
			[]CorpusRun{{"curated", 255, ""}, {"cli", 14, ""}},
			"- Cases: 0\n- Corpora: curated 255, cli 14\n\n",
		},
	} {
		md := Report{Corpora: tc.corpora}.Markdown()
		if !strings.Contains(md, tc.want) {
			t.Errorf("%s: report lacks %q:\n%s", tc.name, tc.want, md)
		}
	}
}

func TestReadKnownDifferences(t *testing.T) {
	path := filepath.Join(t.TempDir(), "known-differences.yaml")
	content := `# Cases that differ from the reference, with the change that resolves them.
- id: dot-layout/crossing
  reason: "order within a rank: \"b\" before \"a\""
  change: add-graphviz-dot-foundation
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	known, err := ReadKnownDifferences(path)
	if err != nil {
		t.Fatal(err)
	}
	k, ok := known["dot-layout/crossing"]
	if !ok || k.Reason != `order within a rank: "b" before "a"` || k.Change != "add-graphviz-dot-foundation" {
		t.Fatalf("known = %+v", known)
	}

	if err := os.WriteFile(path, []byte("- id: x\n  reason: r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadKnownDifferences(path); err == nil {
		t.Error("entry without a change: got no error")
	}
	missing, err := ReadKnownDifferences(filepath.Join(t.TempDir(), "none.yaml"))
	if err != nil || len(missing) != 0 {
		t.Errorf("missing file: %v, %v", missing, err)
	}
}

func TestReferenceRoundTripAndCases(t *testing.T) {
	root := t.TempDir()
	cases := filepath.Join(root, "cases")
	if err := os.MkdirAll(filepath.Join(cases, "dot-layout"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cases, "dot-layout", "chain.gv"), []byte("digraph { a -> b }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadCases(cases)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 1 || loaded[0].ID != "dot-layout/chain" || loaded[0].Capability != "dot-layout" {
		t.Fatalf("cases = %+v", loaded)
	}

	refDir := filepath.Join(root, "reference")
	ref := okReference()
	ref["svg"] = Result{Stdout: []byte(okSVG), Stderr: []byte("Warning: x\n"), Status: 0}
	if err := WriteReference(refDir, "dot-layout/chain", ref); err != nil {
		t.Fatal(err)
	}
	back, err := ReadReference(refDir, "dot-layout/chain")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range Formats {
		if string(back[f].Stdout) != string(ref[f].Stdout) || string(back[f].Stderr) != string(ref[f].Stderr) || back[f].Status != ref[f].Status {
			t.Errorf("%s: round trip gave %+v, want %+v", f, back[f], ref[f])
		}
	}
}

func TestRun(t *testing.T) {
	refDir := t.TempDir()
	ref := okReference()
	cases := []Case{
		{ID: "dot-layout/same", Capability: "dot-layout", Source: []byte("same")},
		{ID: "dot-language/compass", Capability: "dot-language", Source: []byte("digraph { a:n -> b }")},
	}
	for _, c := range cases {
		if err := WriteReference(refDir, c.ID, ref); err != nil {
			t.Fatal(err)
		}
	}
	render := func(src []byte, format string) Result {
		r := ref[format]
		if strings.Contains(string(src), ":n") {
			r.Stderr = []byte("Warning: edge a->b: tailport=n is not supported yet, using center\n")
		}
		return r
	}
	report, err := Run(cases, refDir, nil, render, fakeRasterizer{})
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if report.Results[1].Outcome != NotSupported {
		t.Errorf("compass case: outcome %q, want not supported yet", report.Results[1].Outcome)
	}

	failing := func(src []byte, format string) Result {
		r := ref[format]
		if format == "svg" {
			r.Stdout = []byte(okSVG + " shifted")
		}
		return r
	}
	if _, err := Run(cases[:1], refDir, nil, failing, fakeRasterizer{}); err == nil || !strings.Contains(err.Error(), "dot-layout/same") {
		t.Errorf("unlisted failure: err %v, want the run to fail naming the case", err)
	}
}

func TestReportProxyScenarios(t *testing.T) {
	// Spec scenario "Proxy in the report" of the requirement "Case outcomes
	// and report" of upstream-compatibility, on fake results: the section
	// counts the cases that pass the proxy and lists the others, and the
	// outcomes stay those of the contract.
	results := []CaseResult{
		{Case: Case{ID: "vega-state/close", Capability: "vega-state"}, Outcome: Passed, SSIM: 0.99, ProxyDeviation: 0.01, ProxyObject: "node a"},
		{Case: Case{ID: "vega-state/moved", Capability: "vega-state"}, Outcome: Passed, SSIM: 0.95, ProxyDeviation: 0.03, ProxyObject: "label of a->b#0"},
		{Case: Case{ID: "vega-state/far", Capability: "vega-state"}, Outcome: KnownDiff, SSIM: 0.80, ProxyDeviation: 0.25, ProxyObject: "node c"},
		{Case: Case{ID: "vega-state/blurred", Capability: "vega-state"}, Outcome: KnownDiff, SSIM: 0.85, ProxyDeviation: 0.0, ProxyObject: ""},
	}
	md := Report{Results: results, Proxy: results}.Markdown()
	for _, want := range []string{
		"| vega-state | 4 | 2 | 2 | 0 | 0 |",
		"## gopuml contract proxy\n\n1 of 4 cases of the gopuml corpora pass the proxy",
		"| Case | Largest deviation | Object | SSIM |\n|---|---|---|---|\n" +
			"| vega-state/far | 25.0% | node c | 0.8000 |\n" +
			"| vega-state/moved | 3.0% | label of a->b#0 | 0.9500 |\n" +
			"| vega-state/blurred | 0.0% |  | 0.8500 |\n",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("report lacks %q:\n%s", want, md)
		}
	}
	if strings.Contains(md, "| vega-state/close |") {
		t.Errorf("a case that passes the proxy is listed:\n%s", md)
	}

	// A run without a harvested corpus has no such section.
	if md := (Report{Results: results}).Markdown(); strings.Contains(md, "gopuml contract proxy") {
		t.Errorf("report without a harvested corpus has the proxy section:\n%s", md)
	}
}

func TestRunMeasuresProxy(t *testing.T) {
	// Run records the deviation of each case from its plain outputs, without
	// changing its outcome.
	refDir := t.TempDir()
	ref := okReference()
	ref["plain"] = Result{Stdout: []byte("graph 1 2 2\nnode a 1 1.8 0.75 0.5 a solid ellipse black lightgrey\nstop\n")}
	c := Case{ID: "dot-layout/moved", Capability: "dot-layout", Source: []byte("moved")}
	if err := WriteReference(refDir, c.ID, ref); err != nil {
		t.Fatal(err)
	}
	render := func(src []byte, format string) Result {
		r := ref[format]
		if format == "plain" {
			r.Stdout = []byte("graph 1 2 2\nnode a 1.06 1.8 0.75 0.5 a solid ellipse black lightgrey\nstop\n")
		}
		return r
	}
	report, err := Run([]Case{c}, refDir, nil, render, fakeRasterizer{})
	if err != nil {
		t.Fatal(err)
	}
	res := report.Results[0]
	if res.Outcome != Passed || math.Abs(res.ProxyDeviation-0.03) > 1e-9 || res.ProxyObject != "node a" {
		t.Errorf("outcome %q, deviation %g of %q; want passed, and 3%% of node a", res.Outcome, res.ProxyDeviation, res.ProxyObject)
	}
}
