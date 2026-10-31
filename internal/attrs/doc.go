// SPDX-License-Identifier: MIT
// Provenance: doc https://graphviz.org/doc/info/attrs.html

// Package attrs knows which attributes the Graphviz documentation describes
// and which of them, with which values, gogv supports; it reports the
// features that are not supported yet and gives the layout the values it acts
// on, with the documented defaults in place of unsupported ones.
package attrs

//go:generate go run ../../tools/attrdoc -out documented.go

// UsedBy is the set of object kinds that use an attribute.
type UsedBy uint8

// The kinds of objects of the attribute documentation.
const (
	UsedByGraphs UsedBy = 1 << iota
	UsedBySubgraphs
	UsedByClusters
	UsedByNodes
	UsedByEdges
)
