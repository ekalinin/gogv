# Spec Delta

## Purpose

Provides a `dot` command that can replace Graphviz's `dot` for the flags and formats this project supports, and rejects the others explicitly instead of ignoring them.

## ADDED Requirements

### Requirement: Input

The command SHALL read the files named as arguments, in order, or standard input when no file is named. When a file cannot be read, the command SHALL report an error that names it and SHALL exit with a non-zero status.

#### Scenario: Standard input

- **WHEN** `echo 'digraph { a -> b }' | dot -Tsvg` runs
- **THEN** the SVG of the graph is written to standard output and the exit status is 0

#### Scenario: Missing file

- **WHEN** `dot -Tsvg missing.gv` runs and `missing.gv` does not exist
- **THEN** an error naming `missing.gv` is written to standard error and the exit status is not 0

### Requirement: Output format

`-T<format>` SHALL select the output format among `svg`, `plain` and `canon`. Without `-T`, the format SHALL be `dot`, as in Graphviz. For a format of a later slice of v1 (`svg_inline`, `plain-ext`, `dot`, `gv`, `xdot`, `xdot1.2`, `xdot1.4`, `json`, `json0`, `dot_json`, `xdot_json`), the command SHALL report that the format is not supported yet and SHALL exit with status 1. For any other format, the command SHALL report it in the reference's wording (`Format: "NAME" not recognized. Use one of: ...`, listing the supported formats) and SHALL exit with status 1.

#### Scenario: Supported format

- **WHEN** `dot -Tplain` runs on `digraph { a }`
- **THEN** the `plain` output is written to standard output

#### Scenario: Format of a later slice

- **WHEN** `dot -Txdot` runs
- **THEN** the command reports that `xdot` is not supported yet and the exit status is 1

#### Scenario: Unknown format

- **WHEN** `dot -Tfoo` runs
- **THEN** the command reports `Format: "foo" not recognized. Use one of: ...` and the exit status is 1

### Requirement: Output file

`-o<file>` SHALL write the output to that file instead of standard output. The file name SHALL be accepted joined to the flag or as the next argument, as the reference accepts it.

#### Scenario: Output to a file

- **WHEN** `dot -Tsvg -oout.svg in.gv` runs
- **THEN** `out.svg` holds the SVG of the graph and nothing is written to standard output

### Requirement: Version

`-V` SHALL write `dot - graphviz version 16.1.0 (gogv VERSION)` to standard error, where `VERSION` is the version of this project, SHALL exit with status 0 and SHALL NOT read any input. The text before the parentheses SHALL keep the reference's form, so that programs that read the Graphviz version from it (PlantUML, gopuml) find 16.1.0.

#### Scenario: Version line

- **WHEN** `dot -V` runs
- **THEN** standard error holds one line that starts with `dot - graphviz version 16.1.0 (` and the exit status is 0

### Requirement: Layout engine flag

`-Kdot` SHALL be accepted. Any other `-K` value SHALL be reported as not supported yet, with exit status 1.

#### Scenario: Other engine

- **WHEN** `dot -Kneato -Tsvg` runs
- **THEN** the command reports that `neato` is not supported yet and the exit status is 1

### Requirement: Other flags

A flag of Graphviz's `dot` other than `-T`, `-o`, `-V` and `-K` (for example `-G`, `-N`, `-E`, `-O`, `-n`, `-s`, `-y`, `-v`) SHALL be reported as not supported yet, with exit status 1, instead of being ignored.

#### Scenario: Default attribute flag

- **WHEN** `dot -Gnodesep=1 -Tsvg` runs
- **THEN** the command reports that `-G` is not supported yet and the exit status is 1

### Requirement: Messages and exit status

Warnings SHALL be written to standard error with the prefix `Warning:` and errors with the prefix `Error:`. When there are only warnings, the output SHALL be written and the exit status SHALL be 0. A syntax error SHALL give exit status 1.

#### Scenario: Warning only

- **WHEN** `dot -Tsvg` runs on `digraph { a [shape=record] }`
- **THEN** a `Warning:` line saying that `shape=record` is not supported yet is written to standard error, the SVG is written to standard output, and the exit status is 0

#### Scenario: Syntax error

- **WHEN** `dot -Tsvg` runs on `digraph { a -> }`
- **THEN** an `Error:` line with the syntax error is written to standard error and the exit status is 1
