// SPDX-License-Identifier: MIT
// Provenance: original

package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ekalinin/gogv"
	"github.com/ekalinin/gogv/internal/conformance"
)

// referenceRunner lays out inputs with the reference image, as tools/refgen
// runs it, and caches its plain outputs, and its errors, by the SHA-256 of
// the input.
type referenceRunner struct {
	image, cache string
	// runs and cached count the containers run and the outputs read from
	// the cache.
	runs, cached int
}

func (r *referenceRunner) layout(src string) (conformance.Plain, error) {
	sum := sha256.Sum256([]byte(src))
	base := filepath.Join(r.cache, hex.EncodeToString(sum[:]))
	if out, err := os.ReadFile(base + ".plain"); err == nil {
		r.cached++
		return conformance.ParsePlain(out)
	}
	if msg, err := os.ReadFile(base + ".err"); err == nil {
		r.cached++
		return conformance.Plain{}, errors.New(string(msg))
	}
	if err := os.MkdirAll(r.cache, 0o755); err != nil {
		return conformance.Plain{}, err
	}
	r.runs++
	cmd := exec.Command("docker", "run", "--rm", "-i", "--platform", "linux/amd64", r.image, "dot", "-Tplain")
	cmd.Stdin = bytes.NewReader([]byte(src))
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		if _, ok := errors.AsType[*exec.ExitError](err); !ok {
			return conformance.Plain{}, fmt.Errorf("docker: %w", err)
		}
		msg := fmt.Sprintf("the reference fails: %v: %s", err, bytes.TrimSpace(stderr.Bytes()))
		return conformance.Plain{}, errors.Join(errors.New(msg), os.WriteFile(base+".err", []byte(msg), 0o644))
	}
	if err := os.WriteFile(base+".plain", out, 0o644); err != nil {
		return conformance.Plain{}, err
	}
	return conformance.ParsePlain(out)
}

// gogvLayout lays out an input with gogv, which must give no warning.
func gogvLayout(src string) (conformance.Plain, error) {
	out, warnings, err := gogv.Render(context.Background(), src, "plain")
	if err != nil {
		return conformance.Plain{}, err
	}
	if len(warnings) > 0 {
		return conformance.Plain{}, fmt.Errorf("warning: %s", warnings[0])
	}
	return conformance.ParsePlain(out)
}
