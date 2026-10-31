// SPDX-License-Identifier: MIT
// Provenance: gopuml tools/provenance/main_test.go

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckFile(t *testing.T) {
	const spdx = spdxLine + "\n"
	tests := []struct {
		name    string
		content string
		wantErr string // substring of the first problem, empty for no problems
	}{
		{
			name:    "original",
			content: spdx + "// Provenance: original\n\npackage x\n",
		},
		{
			name: "paper, doc and observed behavior",
			content: spdx +
				"// Provenance: paper Gansner, Koutsofios, North, Vo 1993, section 2\n" +
				"// Provenance: doc https://graphviz.org/doc/info/attrs.html\n" +
				"// Provenance: observed dot 16.1.0\n\npackage x\n",
		},
		{
			name:    "gopuml",
			content: spdx + "// Provenance: gopuml internal/conformance/compare.go\n\npackage x\n",
		},
		{
			name:    "Liberation font tables",
			content: spdx + "// Provenance: liberation-fonts 2.1.5\n\npackage x\n",
		},
		{
			name:    "fdlibm with notice",
			content: spdx + "// Provenance: fdlibm 5.3 e_atan2.c\n" + fdlibmNotice + "\n\npackage x\n",
		},
		{
			name:    "missing header",
			content: spdx + "\npackage x\n",
			wantErr: "missing provenance header",
		},
		{
			name:    "missing SPDX line",
			content: "// Provenance: original\n\npackage x\n",
			wantErr: "first line is not",
		},
		{
			name:    "original and another provenance",
			content: spdx + "// Provenance: original\n// Provenance: observed dot 16.1.0\n\npackage x\n",
			wantErr: "declares both original and another provenance",
		},
		{
			name:    "Graphviz source code",
			content: spdx + "// Provenance: graphviz 16.1.0 lib/dotgen/rank.c\n\npackage x\n",
			wantErr: "unrecognized provenance",
		},
		{
			name:    "documentation outside graphviz.org",
			content: spdx + "// Provenance: doc https://example.com/dot.html\n\npackage x\n",
			wantErr: "unrecognized provenance",
		},
		{
			name:    "other reference version",
			content: spdx + "// Provenance: observed dot 14.0.1\n\npackage x\n",
			wantErr: "unrecognized provenance",
		},
		{
			name:    "fdlibm without notice",
			content: spdx + "// Provenance: fdlibm 5.3 e_atan2.c\n\npackage x\n",
			wantErr: "lacks Sun's notice",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems := checkFile("x.go", tt.content)
			if tt.wantErr == "" {
				if len(problems) != 0 {
					t.Fatalf("unexpected problems: %v", problems)
				}
				return
			}
			if len(problems) == 0 {
				t.Fatalf("expected a problem containing %q, got none", tt.wantErr)
			}
			if !strings.Contains(problems[0].Message, tt.wantErr) {
				t.Fatalf("first problem %q does not contain %q", problems[0].Message, tt.wantErr)
			}
		})
	}
}

func TestCheckTreeSkipsTestsAndTestdata(t *testing.T) {
	root := t.TempDir()
	write := func(rel, content string) {
		t.Helper()
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("ok.go", spdxLine+"\n// Provenance: original\n\npackage x\n")
	write("ok_test.go", "package x\n")
	write("testdata/bad.go", "package x\n")
	write("sub/bad.go", "package x\n")

	problems, count, err := checkTree([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("checked %d files, want 2", count)
	}
	if len(problems) == 0 || !strings.HasSuffix(problems[0].File, "sub/bad.go") {
		t.Fatalf("problems = %v, want one for sub/bad.go", problems)
	}
}
