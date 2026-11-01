## MODIFIED Requirements

### Requirement: Output format

`-T<format>` SHALL select the output format among `svg`, `svg_inline`, `plain`, `plain-ext`, `canon`, `dot`, `gv`, `xdot`, `xdot1.2`, `xdot1.4`, `json`, `json0`, `dot_json` and `xdot_json`, the formats of v1. Without `-T`, the format SHALL be `dot`, as in Graphviz. For any other format, the command SHALL report it in the reference's wording (`Format: "NAME" not recognized. Use one of: ...`, listing the supported formats) and SHALL exit with status 1.

#### Scenario: Supported format

- **WHEN** `dot -Tplain` runs on `digraph { a }`
- **THEN** the `plain` output is written to standard output

#### Scenario: Format of a later slice

- **WHEN** `dot -Txdot` runs on `digraph { a }`, a format that the last slice of v1 adds
- **THEN** the `xdot` output, byte-identical to the reference's, is written to standard output and the exit status is 0

#### Scenario: Default format

- **WHEN** `dot` runs without `-T` on `digraph G { a }`
- **THEN** the `dot` output is written to standard output, starting with `digraph G {` and `graph [bb="0,0,54,36"];`, and the exit status is 0

#### Scenario: Unknown format

- **WHEN** `dot -Tfoo` runs
- **THEN** the command reports `Format: "foo" not recognized. Use one of: ...` and the exit status is 1
