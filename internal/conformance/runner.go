// SPDX-License-Identifier: MIT
// Provenance: original

package conformance

import (
	"bufio"
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// Formats are the output formats that the conformance run can compare, in
// the order of the layers of the contract. A case is compared in the formats
// that its reference records.
var Formats = []string{"canon", "plain", "plain-ext", "svg", "svg_inline", "dot", "gv", "xdot", "xdot1.2", "xdot1.4", "json", "json0", "dot_json", "xdot_json"}

// formatsOf returns the formats of ref, in the order of Formats.
func formatsOf(ref Reference) []string {
	var fs []string
	for _, f := range Formats {
		if _, ok := ref[f]; ok {
			fs = append(fs, f)
		}
	}
	return fs
}

// NotSupportedYet is the phrase of the warnings and errors that gogv emits
// for a feature it does not support yet.
const NotSupportedYet = "not supported yet"

// Case is one DOT input of a corpus.
type Case struct {
	// ID is "<capability>/<name>", from the path under the corpus root.
	ID         string
	Capability string
	Source     []byte
}

// LoadCases reads every *.gv file in the capability directories under root.
func LoadCases(root string) ([]Case, error) {
	var cases []Case
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".gv" {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(strings.TrimSuffix(rel, ".gv"))
		capability, _, ok := strings.Cut(rel, "/")
		if !ok {
			return fmt.Errorf("%s: a case must be in a capability directory", path)
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		cases = append(cases, Case{ID: rel, Capability: capability, Source: src})
		return nil
	})
	slices.SortFunc(cases, func(a, b Case) int { return cmp.Compare(a.ID, b.ID) })
	return cases, err
}

// Result is what a run of dot for one format produced.
type Result struct {
	Stdout []byte
	Stderr []byte
	Status int
}

// Reference is the recorded run of the reference dot for each format of a
// case.
type Reference map[string]Result

// meta is the JSON file that records the exit status and the stderr of the
// reference for each format; stdout is recorded in a file per format.
type meta map[string]struct {
	Status int    `json:"status"`
	Stderr string `json:"stderr"`
}

// WriteReference records ref for the case id under dir, as
// <id>.<format> files with stdout for each of its formats and
// <id>.meta.json.
func WriteReference(dir, id string, ref Reference) error {
	base := filepath.Join(dir, filepath.FromSlash(id))
	if err := os.MkdirAll(filepath.Dir(base), 0o755); err != nil {
		return err
	}
	if len(formatsOf(ref)) == 0 {
		return fmt.Errorf("%s: no result", id)
	}
	m := meta{}
	for _, f := range formatsOf(ref) {
		r := ref[f]
		if err := os.WriteFile(base+"."+f, r.Stdout, 0o644); err != nil {
			return err
		}
		m[f] = struct {
			Status int    `json:"status"`
			Stderr string `json:"stderr"`
		}{r.Status, string(r.Stderr)}
	}
	var data bytes.Buffer
	enc := json.NewEncoder(&data)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(m); err != nil {
		return err
	}
	return os.WriteFile(base+".meta.json", data.Bytes(), 0o644)
}

// ReadReference reads what WriteReference recorded.
func ReadReference(dir, id string) (Reference, error) {
	base := filepath.Join(dir, filepath.FromSlash(id))
	data, err := os.ReadFile(base + ".meta.json")
	if err != nil {
		return nil, err
	}
	var m meta
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("%s.meta.json: %w", base, err)
	}
	ref := Reference{}
	for _, f := range Formats {
		if _, ok := m[f]; !ok {
			continue
		}
		out, err := os.ReadFile(base + "." + f)
		if err != nil {
			return nil, err
		}
		ref[f] = Result{Stdout: out, Stderr: []byte(m[f].Stderr), Status: m[f].Status}
	}
	return ref, nil
}

// KnownDifference is an entry of known-differences.yaml.
type KnownDifference struct {
	ID     string
	Reason string
	Change string
}

// ReadKnownDifferences reads a known-differences file: a YAML list of
// entries with the keys id, reason and change, each a plain or double-quoted
// scalar on one line. A missing file is an empty list.
func ReadKnownDifferences(path string) (map[string]KnownDifference, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return map[string]KnownDifference{}, nil
	}
	if err != nil {
		return nil, err
	}
	known := map[string]KnownDifference{}
	var cur *KnownDifference
	flush := func() error {
		if cur == nil {
			return nil
		}
		if cur.ID == "" || cur.Reason == "" || cur.Change == "" {
			return fmt.Errorf("%s: entry %q needs an id, a reason and a change", path, cur.ID)
		}
		known[cur.ID] = *cur
		return nil
	}
	sc := bufio.NewScanner(bytes.NewReader(data))
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimRight(sc.Text(), " \t")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || trimmed == "[]" {
			continue
		}
		if rest, ok := strings.CutPrefix(line, "- "); ok {
			if err := flush(); err != nil {
				return nil, err
			}
			cur = &KnownDifference{}
			line = "  " + rest
		}
		if cur == nil || !strings.HasPrefix(line, "  ") {
			return nil, fmt.Errorf("%s:%d: expected a list entry", path, n)
		}
		key, value, ok := strings.Cut(strings.TrimSpace(line), ":")
		if !ok {
			return nil, fmt.Errorf("%s:%d: expected key: value", path, n)
		}
		value = strings.TrimSpace(value)
		if strings.HasPrefix(value, `"`) {
			if value, err = strconv.Unquote(value); err != nil {
				return nil, fmt.Errorf("%s:%d: %w", path, n, err)
			}
		}
		switch key {
		case "id":
			cur.ID = value
		case "reason":
			cur.Reason = value
		case "change":
			cur.Change = value
		default:
			return nil, fmt.Errorf("%s:%d: unknown key %q", path, n, key)
		}
	}
	if err := flush(); err != nil {
		return nil, err
	}
	return known, sc.Err()
}

// Outcome is the result of a case.
type Outcome string

// The outcomes of the upstream-compatibility spec.
const (
	Passed       Outcome = "passed"
	KnownDiff    Outcome = "known difference"
	NotSupported Outcome = "not supported yet"
	Failed       Outcome = "failed"
)

// Outcomes lists the outcomes in report order.
var Outcomes = []Outcome{Passed, KnownDiff, NotSupported, Failed}

// CaseResult is the outcome of a case and of each of its formats.
type CaseResult struct {
	Case    Case
	Outcome Outcome
	// Formats holds the outcome of each format; Problems the reasons of the
	// formats that did not pass.
	Formats  map[string]Outcome
	Problems map[string]string
	SSIM     float64
}

// Judge compares gogv's results for a case with the reference's.
func Judge(c Case, ref Reference, got map[string]Result, known map[string]KnownDifference, r Rasterizer) CaseResult {
	res := CaseResult{Case: c, Formats: map[string]Outcome{}, Problems: map[string]string{}}
	notSupported := false
	for _, f := range formatsOf(ref) {
		g := got[f]
		if bytes.Contains(g.Stderr, []byte(NotSupportedYet)) {
			notSupported = true
			res.Formats[f] = NotSupported
			continue
		}
		if err := judgeFormat(f, ref[f], g, r, &res); err != nil {
			res.Formats[f] = Failed
			res.Problems[f] = err.Error()
		} else {
			res.Formats[f] = Passed
		}
	}
	_, isKnown := known[c.ID]
	switch {
	case notSupported:
		res.Outcome = NotSupported
	case isKnown:
		res.Outcome = KnownDiff
		for f, o := range res.Formats {
			if o == Failed {
				res.Formats[f] = KnownDiff
			}
		}
	case len(res.Problems) == 0:
		res.Outcome = Passed
	default:
		res.Outcome = Failed
	}
	return res
}

func judgeFormat(format string, ref, got Result, r Rasterizer, res *CaseResult) error {
	if ref.Status != 0 || got.Status != 0 {
		if ref.Status == got.Status && firstLine(ref.Stderr) == firstLine(got.Stderr) {
			return nil
		}
		return fmt.Errorf("exit status %d and stderr %q, reference %d and %q",
			got.Status, firstLine(got.Stderr), ref.Status, firstLine(ref.Stderr))
	}
	switch format {
	case "canon":
		return CompareCanon(ref.Stdout, got.Stdout)
	case "plain", "plain-ext":
		return eachDocument(ref.Stdout, got.Stdout, "stop\n", func(ref, got []byte) error {
			rp, err := ParsePlain(ref)
			if err != nil {
				return fmt.Errorf("reference: %w", err)
			}
			gp, err := ParsePlain(got)
			if err != nil {
				return err
			}
			return ComparePlain(rp, gp, SizeTolerance)
		})
	case "svg":
		return eachDocument(ref.Stdout, got.Stdout, "</svg>\n", func(ref, got []byte) error {
			score, err := CompareSVG(r, ref, got)
			res.SSIM = score
			return err
		})
	case "svg_inline":
		return eachDocument(ref.Stdout, got.Stdout, "</svg>\n", func(ref, got []byte) error {
			return CompareInlineSVG(r, ref, got)
		})
	case "dot", "gv", "xdot", "xdot1.2", "xdot1.4":
		return CompareDOT(ref.Stdout, got.Stdout, SizeTolerance)
	case "json", "json0":
		return CompareJSON(ref.Stdout, got.Stdout, SizeTolerance)
	case "dot_json", "xdot_json":
		if !bytes.Equal(ref.Stdout, got.Stdout) {
			return errors.New("differs from the reference")
		}
		return nil
	}
	return fmt.Errorf("unknown format %q", format)
}

// eachDocument compares the outputs of an input with several graphs graph by
// graph: each document of the output ends with end.
func eachDocument(ref, got []byte, end string, compare func(ref, got []byte) error) error {
	split := func(data []byte) [][]byte {
		var docs [][]byte
		for len(data) > 0 {
			i := bytes.Index(data, []byte(end))
			if i < 0 {
				return append(docs, data)
			}
			docs = append(docs, data[:i+len(end)])
			data = data[i+len(end):]
		}
		return docs
	}
	rd, gd := split(ref), split(got)
	if len(rd) != len(gd) {
		return fmt.Errorf("%d documents, the reference has %d", len(gd), len(rd))
	}
	for i := range rd {
		if err := compare(rd[i], gd[i]); err != nil {
			if len(rd) > 1 {
				return fmt.Errorf("graph %d: %w", i+1, err)
			}
			return err
		}
	}
	return nil
}

// firstLine returns the first line of a stderr output, with the input name
// that dot puts in its messages removed, so that reading from a file and from
// stdin compare equal.
func firstLine(stderr []byte) string {
	line, _, _ := strings.Cut(string(stderr), "\n")
	line = strings.TrimSpace(line)
	if rest, ok := strings.CutPrefix(line, "Error: "); ok {
		if _, msg, ok := strings.Cut(rest, ": "); ok {
			return "Error: " + msg
		}
	}
	return line
}

// Report collects the results of a conformance run.
type Report struct {
	Results []CaseResult
	// Corpora are the corpora of the run, in the order they ran.
	Corpora []CorpusRun
}

// CorpusRun is a corpus of a conformance run: its name, its number of cases
// and, for a harvested corpus, the source recorded with it.
type CorpusRun struct {
	Name   string
	Cases  int
	Source string
}

// Failed returns the results whose outcome is Failed. The run fails when it
// is not empty.
func (r Report) Failed() []CaseResult {
	var failed []CaseResult
	for _, res := range r.Results {
		if res.Outcome == Failed {
			failed = append(failed, res)
		}
	}
	return failed
}

// Markdown writes the report: the corpora of the run, counts of each outcome
// per capability and per format, then every case that is not passed.
func (r Report) Markdown() string {
	var b strings.Builder
	b.WriteString("# Conformance report\n\n")
	b.WriteString("- Reference: Graphviz 16.1.0 built without Pango\n")
	fmt.Fprintf(&b, "- Cases: %d\n", len(r.Results))
	if len(r.Corpora) > 0 {
		var corpora []string
		for _, c := range r.Corpora {
			s := fmt.Sprintf("%s %d", c.Name, c.Cases)
			if c.Source != "" {
				s += " (" + c.Source + ")"
			}
			corpora = append(corpora, s)
		}
		fmt.Fprintf(&b, "- Corpora: %s\n", strings.Join(corpora, ", "))
	}
	b.WriteString("\n")

	header := "| Name | Cases | Passed | Known differences | Not supported yet | Failed |\n|---|---|---|---|---|---|\n"
	row := func(name string, counts map[Outcome]int) {
		total := 0
		for _, n := range counts {
			total += n
		}
		fmt.Fprintf(&b, "| %s | %d | %d | %d | %d | %d |\n", name, total,
			counts[Passed], counts[KnownDiff], counts[NotSupported], counts[Failed])
	}

	b.WriteString("## By capability\n\n" + header)
	byCap := map[string]map[Outcome]int{}
	for _, res := range r.Results {
		if byCap[res.Case.Capability] == nil {
			byCap[res.Case.Capability] = map[Outcome]int{}
		}
		byCap[res.Case.Capability][res.Outcome]++
	}
	for _, name := range slices.Sorted(maps.Keys(byCap)) {
		row(name, byCap[name])
	}

	b.WriteString("\n## By format\n\n" + header)
	for _, f := range Formats {
		counts := map[Outcome]int{}
		for _, res := range r.Results {
			if o, ok := res.Formats[f]; ok {
				counts[o]++
			}
		}
		if len(counts) > 0 {
			row(f, counts)
		}
	}

	var other []CaseResult
	for _, res := range r.Results {
		if res.Outcome != Passed {
			other = append(other, res)
		}
	}
	if len(other) > 0 {
		b.WriteString("\n## Cases not passed\n\n| Case | Outcome | Problems |\n|---|---|---|\n")
		for _, res := range other {
			var problems []string
			for _, f := range Formats {
				if p, ok := res.Problems[f]; ok {
					problems = append(problems, f+": "+p)
				}
			}
			fmt.Fprintf(&b, "| %s | %s | %s |\n", res.Case.ID, res.Outcome, mdCell(strings.Join(problems, "; ")))
		}
	}
	return b.String()
}

func mdCell(s string) string {
	s = strings.ReplaceAll(s, "|", `\|`)
	return strings.ReplaceAll(s, "\n", " ")
}

// Renderer runs gogv on a DOT source for one format, as the command line
// `dot -T<format>` with the source on stdin would.
type Renderer func(src []byte, format string) Result

// Run judges every case against its reference recorded under refDir and
// returns the report. The error is not nil when a case failed: that is what
// makes the conformance run fail. Known differences and cases not supported
// yet do not.
func Run(cases []Case, refDir string, known map[string]KnownDifference, render Renderer, r Rasterizer) (Report, error) {
	var report Report
	for _, c := range cases {
		ref, err := ReadReference(refDir, c.ID)
		if err != nil {
			return report, fmt.Errorf("%s: no reference: %w", c.ID, err)
		}
		got := map[string]Result{}
		for _, f := range formatsOf(ref) {
			got[f] = render(c.Source, f)
		}
		report.Results = append(report.Results, Judge(c, ref, got, known, r))
	}
	if failed := report.Failed(); len(failed) > 0 {
		ids := make([]string, len(failed))
		for i, res := range failed {
			ids[i] = res.Case.ID
		}
		return report, fmt.Errorf("%d case(s) failed: %s", len(failed), strings.Join(ids, ", "))
	}
	return report, nil
}
