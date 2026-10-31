// SPDX-License-Identifier: MIT
// Provenance: gopuml internal/jmath/doc.go

// Package fdlibm provides deterministic elementary functions: Sin, Cos, Atan
// and Atan2, a port of fdlibm 5.3 taken from gopuml. Their results are
// bit-identical on every operating system and architecture, unlike those of
// package math.
//
// Go may fuse a*b+c into a single fused multiply-add on some architectures
// (arm64, amd64 v3), which changes the last bits of a result. Every product
// that feeds an addition or a subtraction is therefore written float64(a*b):
// the Go specification guarantees that an explicit conversion rounds the
// product and prevents fusion. From package math only exact functions are
// used: bit access, Abs, Floor and Copysign.
package fdlibm
