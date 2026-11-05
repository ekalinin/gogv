// SPDX-License-Identifier: MIT
// Provenance: original

package gogv_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ekalinin/gogv"
)

// BenchmarkRenderClasses renders as svg the graphs of issue #29: the DOT
// that gopuml writes for class diagrams of 60, 100 and 150 classes, each
// linked to the next one and to class (i*7+3) mod n, saved by the script of
// the issue. tools/perf/compare.sh runs it on the base and the head of a pull
// request. It uses only Render, so that it builds on any base.
func BenchmarkRenderClasses(b *testing.B) {
	for _, n := range []int{60, 100, 150} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			data, err := os.ReadFile(filepath.Join("testdata", "bench", fmt.Sprintf("classes-%d.gv", n)))
			if err != nil {
				b.Fatal(err)
			}
			src := string(data)
			b.ReportAllocs()
			for b.Loop() {
				if _, _, err := gogv.Render(b.Context(), src, "svg"); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
