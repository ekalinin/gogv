// SPDX-License-Identifier: MIT
// Provenance: original

package textmetrics

// faceTable is the advance-width table of one font face.
type faceTable struct {
	name       string
	unitsPerEm uint16
	ranges     []widthRange
}

// widthRange holds the advance widths of consecutive characters, starting at
// first.
type widthRange struct {
	first  rune
	widths []uint16
}
