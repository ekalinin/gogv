# Proposal

## Why

gopuml lays out its graph-based diagrams with gogv, and on dense graphs gogv takes 2 to 4 times as long as Graphviz `dot` (issue #29): 6.18 s against 1.69 s for a class diagram of 150 classes, and 125.78 s against 55.07 s for 300 classes, where both exceed gopuml's default layout timeout of 30 seconds. Nearly all of that time goes to the network simplex that places the nodes in x. On the same auxiliary graph, gogv makes about as many pivots as the reference does, but after each pivot it rebuilds the whole spanning tree and recomputes every cut value, where section 2.4 of Gansner et al. updates only the part of the tree that the pivot changes.

## What Changes

- After each pivot, the network simplex updates its spanning tree as section 2.4 of Gansner et al. describes: it recomputes the cut values on the tree path between the ends of the entering edge, and the postorder numbers and parent edges of the nodes below the least common ancestor of those ends. It no longer rebuilds the tree or recomputes every cut value. The simplex is shared by ranking and by x coordinates, so both get the update.
- The pivots stay the same and come in the same order, so the ranks and coordinates do not change, and neither does any output of gogv, byte for byte.
- A test checks the updated tree against a full rebuild after every pivot.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None. The change makes the layout faster and keeps its results. No behavior that `openspec/specs` describes changes, so the change declares `skip_specs`.

## Impact

- `internal/layout/simplex.go`: `exchange` and the tree bookkeeping it relies on.
- `internal/layout/simplex_test.go`: the new test.
- No change to any output, `conformance-report.md`, `testdata/conformance/known-differences.yaml`, the library API or the command line.
