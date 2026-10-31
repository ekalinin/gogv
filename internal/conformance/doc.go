// SPDX-License-Identifier: MIT
// Provenance: original

// Package conformance compares the outputs of gogv with those of the
// reference dot (Graphviz 16.1.0 built without Pango), layer by layer, as the
// upstream-compatibility spec defines: canon byte for byte, plain by nodes,
// edges and size, svg by size and the structural similarity of the
// rasterized documents.
package conformance
