// SPDX-License-Identifier: MIT
// Provenance: original

// Command refgen records the outputs of the reference dot for every case of a
// corpus: it runs `dot -Tcanon`, `-Tplain` and `-Tsvg` in the reference image
// (tools/reference/Dockerfile) with each case on stdin, and writes stdout,
// stderr and the exit status with conformance.WriteReference.
//
// All cases run in one container, with the corpus mounted read-only and a
// work directory under .cache, which colima and Docker Desktop share with the
// host.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/ekalinin/gogv/internal/conformance"
)

// script runs inside the reference image. For each case of /work/cases.txt,
// a line with its path and its formats, it writes <case>.<format>.out, .err
// and .status under /work.
const script = `set -u
cd /cases
while read -r f fmts; do
	mkdir -p "/work/$(dirname "$f")"
	for fmt in $fmts; do
		dot -T"$fmt" < "$f" > "/work/$f.$fmt.out" 2> "/work/$f.$fmt.err"
		echo $? > "/work/$f.$fmt.status"
	done
done < /work/cases.txt
`

func main() {
	image := flag.String("image", "gogv-reference", "reference image")
	cases := flag.String("cases", "testdata/conformance/cases", "corpus directory")
	out := flag.String("out", "testdata/conformance/reference", "directory of the recorded references")
	cli := flag.String("cli", "", "CLI cases to record instead (JSON); written next to them as reference.json")
	formats := flag.String("formats", "canon plain svg", "formats to record for every case")
	all := flag.String("all", "", "capabilities whose cases are recorded in every format the conformance run compares")
	flag.Parse()

	var err error
	if *cli != "" {
		err = recordCLI(*image, *cli)
	} else {
		err = run(*image, *cases, *out, strings.Fields(*formats), strings.Fields(*all))
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "refgen:", err)
		os.Exit(1)
	}
}

// recordCLI runs each CLI case with the reference dot, with the case's
// standard input, in an empty working directory.
func recordCLI(image, path string) error {
	cases, err := conformance.ReadCLICases(path)
	if err != nil {
		return err
	}
	ref := map[string]conformance.CLIResult{}
	for _, c := range cases {
		args := append([]string{"run", "--rm", "-i", "--platform", "linux/amd64", "-w", "/tmp", "--entrypoint", "dot", image}, c.Args...)
		cmd := exec.Command("docker", args...)
		cmd.Stdin = strings.NewReader(c.Stdin)
		var stdout, stderr strings.Builder
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		status := 0
		if err := cmd.Run(); err != nil {
			var ee *exec.ExitError
			if !errors.As(err, &ee) {
				return err
			}
			status = ee.ExitCode()
		}
		ref[c.ID] = conformance.CLIResult{Stdout: stdout.String(), Stderr: stderr.String(), Status: status}
	}
	if err := conformance.WriteCLIReference(filepath.Join(filepath.Dir(path), "reference.json"), ref); err != nil {
		return err
	}
	fmt.Printf("refgen: %d CLI case(s) recorded\n", len(cases))
	return nil
}

// run records the references of the cases under casesDir: in formats, or in
// every format of conformance.Formats for the capabilities of all.
func run(image, casesDir, out string, formats, all []string) error {
	cases, err := conformance.LoadCases(casesDir)
	if err != nil {
		return err
	}
	formatsOf := func(c conformance.Case) []string {
		if slices.Contains(all, c.Capability) {
			return conformance.Formats
		}
		return formats
	}
	absCases, err := filepath.Abs(casesDir)
	if err != nil {
		return err
	}
	// A fresh directory per run: colima's shared file system can keep serving
	// a removed directory of the same name to the container.
	if err := os.MkdirAll(".cache", 0o755); err != nil {
		return err
	}
	work, err := os.MkdirTemp(".cache", "refgen-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(work)
	if work, err = filepath.Abs(work); err != nil {
		return err
	}

	var list strings.Builder
	for _, c := range cases {
		fmt.Fprintf(&list, "./%s.gv %s\n", c.ID, strings.Join(formatsOf(c), " "))
	}
	if err := os.WriteFile(filepath.Join(work, "cases.txt"), []byte(list.String()), 0o644); err != nil {
		return err
	}

	cmd := exec.Command("docker", "run", "--rm", "--platform", "linux/amd64",
		"-v", absCases+":/cases:ro", "-v", work+":/work",
		image, "sh", "-c", script)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker: %w", err)
	}

	// The references are generated files: start from an empty directory so
	// that removed cases leave nothing behind.
	if err := os.RemoveAll(out); err != nil {
		return err
	}
	for _, c := range cases {
		ref := conformance.Reference{}
		for _, f := range formatsOf(c) {
			base := filepath.Join(work, filepath.FromSlash(c.ID)+".gv."+f)
			stdout, err := os.ReadFile(base + ".out")
			if err != nil {
				return err
			}
			stderr, err := os.ReadFile(base + ".err")
			if err != nil {
				return err
			}
			statusText, err := os.ReadFile(base + ".status")
			if err != nil {
				return err
			}
			status, err := strconv.Atoi(strings.TrimSpace(string(statusText)))
			if err != nil {
				return fmt.Errorf("%s: %w", base, err)
			}
			ref[f] = conformance.Result{Stdout: stdout, Stderr: stderr, Status: status}
		}
		if err := conformance.WriteReference(out, c.ID, ref); err != nil {
			return err
		}
	}
	fmt.Printf("refgen: %d case(s) recorded in %s\n", len(cases), out)
	return nil
}
