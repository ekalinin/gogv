# Proposal

## Why

The README and NOTICE promise that gogv contains no code from Graphviz or from its ports. So far that holds because one maintainer follows decision D1 of the foundation change. Once the repository is public, a pull request from anyone could break the promise, and nothing tells a contributor the rules: `make provenance` checks that every Go file has `Provenance:` lines, not that they are true. The rules have to be written where contributors look, and each pull request has to state that it follows them.

## What Changes

- A `CONTRIBUTING.md` that states the clean-room rules of D1: the allowed sources (papers, the Graphviz documentation, the observed behavior of the reference `dot`, gopuml's own code), the sources that are not allowed (the source code of Graphviz and of its ports and bindings), how a Go file names its sources in `Provenance:` lines, and what to do with an AI assistant. It also says when a change needs an OpenSpec change (only when it changes behavior described in `openspec/specs`), which checks to run, how commits and branches are named, what a bug report needs, and that contributions are under the MIT License.
- A pull request template, `.github/pull_request_template.md`, in which the author checks that the change follows the clean-room rules and that every new or changed Go file names its sources.
- A short Contributing section in the README that points to `CONTRIBUTING.md`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None. The change adds documentation and a pull request template; no behavior of the library, the command line or the conformance harness changes, so the change declares `skip_specs`.

## Impact

- New `CONTRIBUTING.md` and `.github/pull_request_template.md`.
- `README.md`: a Contributing section.
- No code, test or workflow changes. The attestation is not checked by CI.
