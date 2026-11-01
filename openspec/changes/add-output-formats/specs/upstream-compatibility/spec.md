## MODIFIED Requirements

### Requirement: Layered similarity contract

For every case of the corpus, the outputs SHALL be compared with the reference's outputs for the same input:

- `canon`: byte-identical
- `plain`: the same node names and the same edges (tail and head), and graph width and height each within 5% of the reference's
- `svg`: width and height each within 5% of the reference's, and a structural similarity (SSIM, grayscale) of at least 0.90 between both documents rasterized by the pinned `resvg` at the reference's size
- `svg_inline`: the comparison of `svg`, and the same lines as the reference's before the root `svg` element, but for the version of the generator
- `plain-ext`: the comparison of `plain`, and the same ports on the edge lines
- `dot`, `gv`, `xdot`, `xdot1.2`, `xdot1.4`, `json` and `json0`: both outputs parsed, with the same graphs, subgraphs, nodes and edges in the same order, the same attribute names on each, and the same values for the attributes that do not depend on the layout; the same drawing operations, in the same order, with the same colors, styles, fonts and texts; and the width and height of the graph's `bb` each within 5% of the reference's. The numbers of the layout are not compared one by one: `plain` and `svg` judge the layout
- `dot_json` and `xdot_json`: byte-identical

A case passes when the comparisons of all its recorded formats pass.

#### Scenario: Similar layout

- **WHEN** a case has the reference's `canon`, the same nodes and edges in `plain` with a size 3% larger, and an SVG at SSIM 0.94 with a size 3% larger
- **THEN** the case passes

#### Scenario: Different order within a rank

- **WHEN** a case orders the nodes of a rank differently from the reference, and its SVG reaches SSIM 0.82
- **THEN** the case fails

#### Scenario: Layout numbers in dot

- **WHEN** a case has the reference's objects and attributes in `dot`, a `pos` of a node 2 points away from the reference's, and a `bb` 3% larger
- **THEN** its `dot` output passes

#### Scenario: Missing drawing operation

- **WHEN** the `xdot` output of a case has no `_hdraw_` for an edge whose reference has one
- **THEN** its `xdot` output fails

### Requirement: Corpus

The corpus SHALL include:

- curated cases, one DOT file per case grouped by capability, committed with their reference outputs as text: `canon`, `plain`, `svg`, `dot`, `xdot` and `json` for every case, and every format of v1 for the cases of `output-formats`
- the DOT that gopuml generates for its conformance cases, harvested into a cache and not committed, with the reference outputs of `canon`, `plain` and `svg`

#### Scenario: Harvested corpus is not committed

- **WHEN** the harvested gopuml corpus is collected
- **THEN** it is stored in the cache directory, which is ignored by version control

#### Scenario: Formats of a curated case

- **WHEN** the references of the curated cases are recorded
- **THEN** a case of `dot-layout` has references of `canon`, `plain`, `svg`, `dot`, `xdot` and `json`, and a case of `output-formats` has references of the 14 formats of v1
