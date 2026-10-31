// SPDX-License-Identifier: MIT
// Provenance: original

package main

import "testing"

func TestVariant(t *testing.T) {
	src := `digraph unix {
subgraph cluster6p0 {label="";subgraph cluster6 {style=solid;color="#000006";labeljust="c";label=<<TABLE BGCOLOR="#000007" FIXEDSIZE="TRUE" WIDTH="51" HEIGHT="11"><TR><TD></TD></TR></TABLE>>;
a [shape=rect,label=<<TABLE><TR><TD></TD></TR></TABLE>>];
}}
}`
	want := `digraph unix {
subgraph cluster6p0 {label="";subgraph cluster6 {style=solid;color="#000006";labeljust="c";label="";
a [shape=rect,label=<<TABLE><TR><TD></TD></TR></TABLE>>];
}}
}`
	if got := variant(src); got != want {
		t.Errorf("variant:\n%s\nwant:\n%s", got, want)
	}
}
