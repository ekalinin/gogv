## ADDED Requirements

### Requirement: Output formats

The package SHALL list the output formats that it renders: the 14 formats of v1, always in the same order. Every call SHALL return a new list, and changing a returned list SHALL NOT change which formats are rendered, how an unknown format is reported, or the lists of later calls.

#### Scenario: List of formats

- **WHEN** the list of formats is read
- **THEN** it holds the 14 formats of v1, and every format in it renders `digraph { a }` without an error

#### Scenario: Changed list

- **WHEN** a caller appends `png` to a returned list and replaces its `svg` with `pdf`
- **THEN** rendering as `png` returns the error for an unknown format, rendering as `svg` writes the SVG, the error for an unknown format lists `svg` and not `pdf` or `png`, and the next list holds the 14 formats of v1
