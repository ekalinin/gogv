// SPDX-License-Identifier: MIT
// Provenance: original

package main

import "testing"

func TestSource(t *testing.T) {
	for _, tc := range []struct {
		name, version, commit, status, want string
	}{
		{"version", "v0.3.0", "", "", "gopuml v0.3.0"},
		{"clean checkout", "", "6a2bb64", "", "gopuml 6a2bb64 local"},
		{"dirty checkout", "", "6a2bb64", "?? .TODO.md", "gopuml 6a2bb64 local dirty"},
		{"git failed", "", "", "", "gopuml unknown local"},
	} {
		if got := source(tc.version, tc.commit, tc.status); got != tc.want {
			t.Errorf("%s: source = %q, want %q", tc.name, got, tc.want)
		}
	}
}
