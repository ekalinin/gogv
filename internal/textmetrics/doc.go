// SPDX-License-Identifier: MIT
// Provenance: original

// Package textmetrics measures text with advance-width tables of the
// Liberation fonts, which are metrically compatible with Times New Roman,
// Arial and Courier New. The tables are generated from the fonts by
// tools/fontmetrics; no font file is part of the module, and text size never
// depends on the fonts installed on the machine.
package textmetrics

//go:generate go run ../../tools/fontmetrics -out tables.go
