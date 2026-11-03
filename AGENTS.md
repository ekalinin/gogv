# AGENTS.md

gogv is a clean-room Go implementation of the Graphviz `dot` program, with
Graphviz 16.1.0 built without Pango as its reference.

## Clean-room

Work only from the published papers, the Graphviz documentation, the output
of the `gogv-reference` image and the code of gopuml. The source code of
Graphviz and of its ports and bindings (Smetana, graphviz-pure,
goccy/go-graphviz) is the limit: never open, fetch or quote it, not even to
compare.

## When to read CONTRIBUTING.md

- before observing Graphviz's behavior: [Allowed sources](CONTRIBUTING.md#allowed-sources)
- when creating a Go file: [Provenance lines](CONTRIBUTING.md#provenance-lines)
- before writing layout or output code: [Determinism](CONTRIBUTING.md#determinism)
- when a change touches behavior that `openspec/specs` describes: [Changes and the specs](CONTRIBUTING.md#changes-and-the-specs)
- when writing documentation, comments or code: [Style](CONTRIBUTING.md#style)
- before calling a change done: [Checks](CONTRIBUTING.md#checks)
- before committing or creating a branch: [Commits and branches](CONTRIBUTING.md#commits-and-branches)
