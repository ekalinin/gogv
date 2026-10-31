# upstream-compatibility Specification

## Purpose

Defines how closely this implementation follows Graphviz 16.1.0's `dot` built without Pango, and how that closeness is measured and reported.

## Requirements

### Requirement: Pinned reference

Compatibility SHALL be measured against Graphviz 16.1.0, built from its release sources without Pango and run in a container image whose base image, Graphviz sources and SVG rasterizer (`resvg`) are pinned by digest or checksum. References SHALL be regenerated only when a pinned version changes.

#### Scenario: Reference without Pango

- **WHEN** the reference image runs `dot -V` and lays out a node labeled `Hello world`
- **THEN** the version is 16.1.0, and the font of the label is resolved without a text layout plugin

### Requirement: Layered similarity contract

For every case of the corpus, the outputs SHALL be compared with the reference's outputs for the same input:

- `canon`: byte-identical
- `plain`: the same node names and the same edges (tail and head), and graph width and height each within 5% of the reference's
- `svg`: width and height each within 5% of the reference's, and a structural similarity (SSIM, grayscale) of at least 0.90 between both documents rasterized by the pinned `resvg` at the reference's size

A case passes when all three comparisons pass.

#### Scenario: Similar layout

- **WHEN** a case has the reference's `canon`, the same nodes and edges in `plain` with a size 3% larger, and an SVG at SSIM 0.94 with a size 3% larger
- **THEN** the case passes

#### Scenario: Different order within a rank

- **WHEN** a case orders the nodes of a rank differently from the reference, and its SVG reaches SSIM 0.82
- **THEN** the case fails

### Requirement: Corpus

The corpus SHALL include:

- curated cases, one DOT file per case grouped by capability, committed with their reference outputs as text
- the DOT that gopuml generates for its conformance cases, harvested into a cache and not committed

#### Scenario: Harvested corpus is not committed

- **WHEN** the harvested gopuml corpus is collected
- **THEN** it is stored in the cache directory, which is ignored by version control

### Requirement: Case outcomes and report

Each case SHALL have one outcome:

- passed
- known difference: listed in a known-differences file with a reason and the follow-up change that resolves it
- not supported yet: the run of the case emitted a "not supported yet" warning or error
- failed: any other case that does not pass

The conformance run SHALL fail when a case is failed, and SHALL NOT fail for known differences and cases not supported yet. It SHALL write a report with the number of cases of each outcome per capability and per format.

#### Scenario: Feature of a later slice

- **WHEN** a case uses a `label` on an edge before edge labels are supported
- **THEN** the case is reported as not supported yet, and the run does not fail because of it

#### Scenario: Unlisted failure

- **WHEN** a case fails and is not listed in the known-differences file
- **THEN** the conformance run fails and the report names the case
