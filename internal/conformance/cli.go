// SPDX-License-Identifier: MIT
// Provenance: original

package conformance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// CLICase runs the command line with arguments and standard input.
type CLICase struct {
	ID    string   `json:"id"`
	Args  []string `json:"args"`
	Stdin string   `json:"stdin"`
}

// CLIResult is what a CLI case produced.
type CLIResult struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	Status int    `json:"status"`
}

// ReadCLICases reads the CLI cases of a JSON file.
func ReadCLICases(path string) ([]CLICase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cases []CLICase
	return cases, json.Unmarshal(data, &cases)
}

// ReadCLIReference reads the recorded results of the reference by case ID.
func ReadCLIReference(path string) (map[string]CLIResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	ref := map[string]CLIResult{}
	return ref, json.Unmarshal(data, &ref)
}

// WriteCLIReference records the results of the reference.
func WriteCLIReference(path string, ref map[string]CLIResult) error {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(ref); err != nil {
		return err
	}
	return os.WriteFile(path, b.Bytes(), 0o644)
}

// cliLine normalizes the first stderr line for the comparison: the list of
// formats, the build of the version and the program's path differ between
// the two programs by design.
func cliLine(stderr string) string {
	line := firstLine([]byte(stderr))
	if i := strings.Index(line, " Use one of:"); i >= 0 {
		line = line[:i]
	}
	if strings.HasPrefix(line, "dot - graphviz version ") {
		if i := strings.Index(line, " ("); i >= 0 {
			line = line[:i]
		}
	}
	return line
}

// JudgeCLI compares gogv's result for a CLI case with the reference's: the
// exit status, standard output, and the first line of standard error.
func JudgeCLI(c CLICase, ref, got CLIResult) CaseResult {
	res := CaseResult{Case: Case{ID: c.ID, Capability: "cli"}, Formats: map[string]Outcome{}, Problems: map[string]string{}}
	switch {
	case strings.Contains(got.Stderr, NotSupportedYet):
		res.Outcome = NotSupported
	case got.Status != ref.Status || got.Stdout != ref.Stdout || cliLine(got.Stderr) != cliLine(ref.Stderr):
		res.Outcome = Failed
		res.Problems["cli"] = fmt.Sprintf("status %d, stderr %q; reference %d, %q", got.Status, cliLine(got.Stderr), ref.Status, cliLine(ref.Stderr))
	default:
		res.Outcome = Passed
	}
	return res
}
