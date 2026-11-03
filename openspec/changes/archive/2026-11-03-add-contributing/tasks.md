# Tasks

## 1. CONTRIBUTING

- [x] 1.1 Write the clean-room part of `CONTRIBUTING.md` (design D1, D3): the allowed sources and the sources that are not allowed, as in D1 of the foundation design and with its examples; the `Provenance:` header, with one example line for each kind that `tools/provenance` accepts and the rule that `original` stands alone; the rule for AI assistants; and that a new `Provenance:` kind changes `CONTRIBUTING.md`, `NOTICE` and `tools/provenance` together. Verify that every example line, put in the header of a scratch Go file outside the repository's checked directories and checked with `go run ./tools/provenance` pointed at it, is accepted, and that the kinds listed match the cases of `tools/provenance`.
- [x] 1.2 Write the rest of `CONTRIBUTING.md` (D4, D5): when a change needs an OpenSpec change, the checks to run, commits and branches, what a bug report needs, and the MIT License of contributions. Verify that every command it names exists in the `Makefile` or the repository and runs as written on a clean clone (the Docker targets excepted).

## 2. Pull request template and README

- [x] 2.1 Write `.github/pull_request_template.md` (D2, D3): a Summary section, a Clean-room section with two checkboxes (the change was written without reading or copying the source code of Graphviz or of its ports and bindings, by the author or an AI assistant; every new or changed Go file names its sources in `Provenance:` lines), and a Checks section with `make lint test provenance` and the OpenSpec change when the specs change. Verify that the template's wording matches `CONTRIBUTING.md`.
- [x] 2.2 Add a Contributing section to `README.md` before License that points to `CONTRIBUTING.md` (D6). Verify that the link resolves to the file.

## 3. Verification

- [x] 3.1 Verify that no em dash appears in the new or changed files, that `make lint test provenance` passes, and that `openspec validate add-contributing --strict` passes.
- [x] 3.2 With the maintainer's go-ahead, push the branch and open the pull request. Verify that CI passes, and after the merge that `gh api repos/ekalinin/gogv/community/profile` reports both the contributing file and the pull request template (GitHub uses templates from the default branch only).
