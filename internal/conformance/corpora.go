// SPDX-License-Identifier: MIT
// Provenance: original

package conformance

// Corpus is a set of cases with the directory of their recorded references.
type Corpus struct {
	Name string
	// Cases and References are directories relative to the repository root.
	Cases, References string
	// Optional corpora are skipped when their cases have not been collected.
	Optional bool
	// Proxy marks the corpora of gopuml, whose cases the report judges with
	// the proxy of gopuml's contract.
	Proxy bool
}

// Corpora are the corpora of the conformance run, as the
// upstream-compatibility spec lists them: the curated cases, committed with
// their references, and the DOT that gopuml generates, harvested into the
// cache by tools/harvest and recorded there by tools/refgen; and the variant
// of gopuml's graphs with clusters that tools/clustervariant writes.
var Corpora = []Corpus{
	{Name: "curated", Cases: "testdata/conformance/cases", References: "testdata/conformance/reference"},
	{Name: "gopuml", Cases: ".cache/gopuml-dot", References: ".cache/reference/gopuml-dot", Optional: true, Proxy: true},
	{Name: "gopuml-clusters", Cases: ".cache/gopuml-cluster-dot", References: ".cache/reference/gopuml-cluster-dot", Optional: true, Proxy: true},
}
