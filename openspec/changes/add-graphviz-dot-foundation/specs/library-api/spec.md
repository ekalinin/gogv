# Spec Delta

## Purpose

Lets Go programs, gopuml first, lay out and render DOT graphs in their own process, without an external program, including in WebAssembly builds.

## ADDED Requirements

### Requirement: Render DOT text

The package SHALL provide a function that takes a context, DOT source text and an output format name, and returns the output bytes, the warnings, and an error. For the same input and format, the bytes SHALL be those that the command line writes to standard output. A syntax error, or a format that the command line rejects, SHALL be returned as an error.

#### Scenario: Same bytes as the command line

- **WHEN** `digraph { a -> b }` is rendered as `svg` through the package and through `dot -Tsvg`
- **THEN** the two outputs are byte-identical

#### Scenario: Syntax error

- **WHEN** `digraph { a -> }` is rendered through the package
- **THEN** an error is returned that carries the line and the token of the syntax error

### Requirement: Warnings

The warnings of the "not supported yet" kind SHALL be returned separately from the output, as values from which a caller can read the object, the feature and the default used, and SHALL NOT be written to standard error by the package.

#### Scenario: Unsupported shape

- **WHEN** `digraph { a [shape=record] }` is rendered through the package
- **THEN** the output is returned with one warning that names node `a` and `shape=record`, and nothing is written to standard error

### Requirement: Concurrent use

The package SHALL be safe for concurrent use: calls running at the same time SHALL return the same results as the same calls made one after another.

#### Scenario: Parallel renders

- **WHEN** 100 goroutines render the same input at the same time
- **THEN** all of them return byte-identical outputs, equal to the output of a single call

### Requirement: Cancellation

When the context is canceled during a call, the call SHALL stop and return the context's error.

#### Scenario: Canceled context

- **WHEN** the context of a call is canceled before the layout finishes
- **THEN** the call returns the context's error and no output

### Requirement: Self-contained

The package SHALL NOT start processes, read files, read environment variables or use the network. It SHALL build and run with `GOOS=js GOARCH=wasm`.

#### Scenario: WebAssembly build

- **WHEN** the package is built with `GOOS=js GOARCH=wasm` and renders `digraph { a -> b }` as `svg`
- **THEN** the output is byte-identical to the output of a native build
