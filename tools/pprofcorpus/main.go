// SPDX-License-Identifier: MIT
// Provenance: original

// Command pprofcorpus writes the corpus of DOT that go tool pprof writes: it
// builds the program of ./workload, runs it twice, and writes the DOT of
// go tool pprof -dot for its CPU and heap profiles with several option sets
// into the curated cases of pprof. Profiles are sampled, so each run writes
// other files; the committed files are the corpus.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// pprofCase is a case of the corpus: the options of go tool pprof and the
// profile it reads.
type pprofCase struct {
	name    string
	options []string
	profile string
}

// cases are the cases of the corpus. base is the directory of the first run
// and other that of the second one, the base of the diff.
func cases(base, other string) []pprofCase {
	cpu, heap := filepath.Join(base, "cpu.prof"), filepath.Join(base, "heap.prof")
	return []pprofCase{
		{"cpu", nil, cpu},
		{"cpu-nodecount", []string{"-nodecount=12"}, cpu},
		{"cpu-tags", []string{"-tagfocus=work=hash"}, cpu},
		{"cpu-focus", []string{"-focus=sortLoop"}, cpu},
		{"cpu-diff", []string{"-diff_base=" + filepath.Join(other, "cpu.prof")}, cpu},
		{"heap", []string{"-sample_index=alloc_space"}, heap},
		{"heap-nodecount", []string{"-sample_index=alloc_space", "-nodecount=6"}, heap},
	}
}

func main() {
	out := flag.String("out", "testdata/conformance/cases/pprof", "directory of the corpus")
	flag.Parse()
	if err := run(*out); err != nil {
		fmt.Fprintln(os.Stderr, "pprofcorpus:", err)
		os.Exit(1)
	}
}

func run(out string) error {
	work, err := os.MkdirTemp("", "pprofcorpus-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(work)
	bin := filepath.Join(work, "workload")
	if err := command("go", "build", "-o", bin, "./tools/pprofcorpus/workload").Run(); err != nil {
		return fmt.Errorf("build the workload: %w", err)
	}
	base, other := filepath.Join(work, "1"), filepath.Join(work, "2")
	for _, dir := range []string{base, other} {
		if err := os.Mkdir(dir, 0o755); err != nil {
			return err
		}
		if err := command(bin, dir).Run(); err != nil {
			return fmt.Errorf("run the workload: %w", err)
		}
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	old, err := filepath.Glob(filepath.Join(out, "*.gv"))
	if err != nil {
		return err
	}
	for _, f := range old {
		if err := os.Remove(f); err != nil {
			return err
		}
	}
	cs := cases(base, other)
	for _, c := range cs {
		var dot bytes.Buffer
		args := append([]string{"tool", "pprof", "-dot"}, c.options...)
		cmd := command("go", append(args, bin, c.profile)...)
		cmd.Stdout = &dot
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s: go tool pprof: %w", c.name, err)
		}
		if err := os.WriteFile(filepath.Join(out, c.name+".gv"), dot.Bytes(), 0o644); err != nil {
			return err
		}
	}
	fmt.Printf("pprofcorpus: %d case(s) written in %s\n", len(cs), out)
	return nil
}

// command returns a command that writes its errors to stderr.
func command(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	return cmd
}
