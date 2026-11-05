GO ?= go
REFERENCE_IMAGE ?= gogv-reference

.PHONY: build test lint provenance reference-image reference harvest conformance determinism perf release

build:
	$(GO) build ./...

test:
	$(GO) test ./...

lint:
	$(GO) vet ./...
	@test -z "$$(gofmt -l .)" || (echo "gofmt needed:"; gofmt -l .; exit 1)

provenance:
	$(GO) run ./tools/provenance

# Builds the reference image: Graphviz 16.1.0 without Pango, and resvg.
reference-image:
	docker build --platform linux/amd64 -t $(REFERENCE_IMAGE) tools/reference

# Records the reference outputs of the curated cases: canon, plain, svg, dot,
# xdot and json, and every format for the cases of output-formats.
reference:
	$(GO) run ./tools/refgen -image $(REFERENCE_IMAGE) -cases testdata/conformance/cases -out testdata/conformance/reference -formats "canon plain svg dot xdot json" -all output-formats

# Collects the DOT that gopuml generates for its conformance cases into
# .cache/gopuml-dot, writes the variant of its graphs with clusters into
# .cache/gopuml-cluster-dot, and records the references of both. gopuml is
# fetched at GOPUML_VERSION through the Go module proxy, or taken from the
# checkout GOPUML when it is set.
GOPUML ?=
GOPUML_VERSION ?= v0.3.0

harvest:
	$(GO) run ./tools/harvest -gopuml-version $(GOPUML_VERSION) $(if $(GOPUML),-gopuml $(GOPUML)) -image $(REFERENCE_IMAGE)
	$(GO) run ./tools/refgen -image $(REFERENCE_IMAGE) -cases .cache/gopuml-dot -out .cache/reference/gopuml-dot
	$(GO) run ./tools/clustervariant -in .cache/gopuml-dot -out .cache/gopuml-cluster-dot
	$(GO) run ./tools/refgen -image $(REFERENCE_IMAGE) -cases .cache/gopuml-cluster-dot -out .cache/reference/gopuml-cluster-dot

# Runs the conformance suite and writes conformance-report.md.
conformance:
	$(GO) test -tags conformance -count=1 ./internal/conformance

# Checks that the curated corpus renders byte for byte the same natively,
# for amd64 and in the WebAssembly build (run on an arm64 host with Rosetta
# and Node.js).
determinism:
	$(GO) build -o .cache/tools/hash-native ./internal/corpushash
	GOARCH=amd64 $(GO) build -o .cache/tools/hash-amd64 ./internal/corpushash
	GOOS=js GOARCH=wasm $(GO) build -o .cache/tools/hash.wasm ./internal/corpushash
	.cache/tools/hash-native testdata/conformance/cases > .cache/tools/hash-native.txt
	.cache/tools/hash-amd64 testdata/conformance/cases > .cache/tools/hash-amd64.txt
	env -i PATH="$$PATH" "$$($(GO) env GOROOT)/lib/wasm/go_js_wasm_exec" .cache/tools/hash.wasm $$PWD/testdata/conformance/cases > .cache/tools/hash-wasm.txt
	cmp .cache/tools/hash-native.txt .cache/tools/hash-amd64.txt
	cmp .cache/tools/hash-native.txt .cache/tools/hash-wasm.txt
	@tail -1 .cache/tools/hash-native.txt

# Compares the speed of the working tree with BASE (make perf BASE=main), as
# the performance job of CI does for a pull request: it fails when a benchmark
# is significantly slower than BASE by more than THRESHOLD percent (default
# 20), over COUNT rounds (default 10).
BASE ?= main

perf:
	$(if $(COUNT),COUNT=$(COUNT)) $(if $(THRESHOLD),THRESHOLD=$(THRESHOLD)) tools/perf/compare.sh $(BASE)

# Releases VERSION (make release VERSION=v0.1.0): tags the commit of main that
# CI passed with an annotated tag, pushes the tag, and creates the GitHub
# release with the notes of docs/releases/VERSION.md. A pushed tag is never
# moved: once the module proxy has fetched it, its content is fixed.
release:
	@echo "$(VERSION)" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$$' || { echo "release: VERSION must look like v0.1.0"; exit 1; }
	@test -f docs/releases/$(VERSION).md || { echo "release: docs/releases/$(VERSION).md is missing"; exit 1; }
	@test "$$(git rev-parse --abbrev-ref HEAD)" = main || { echo "release: check out main"; exit 1; }
	@test -z "$$(git status --porcelain --untracked-files=no)" || { echo "release: tracked files have uncommitted changes"; exit 1; }
	git fetch -q origin main
	@test "$$(git rev-parse HEAD)" = "$$(git rev-parse origin/main)" || { echo "release: main differs from origin/main"; exit 1; }
	@test -z "$$(git tag -l $(VERSION))$$(git ls-remote --tags origin refs/tags/$(VERSION))" || { echo "release: $(VERSION) is already tagged"; exit 1; }
	@test "$$(gh run list --commit $$(git rev-parse HEAD) --workflow ci.yml --json conclusion -q '.[0].conclusion')" = success || { echo "release: CI has not passed on $$(git rev-parse --short HEAD)"; exit 1; }
	git tag -a $(VERSION) -m "gogv $(VERSION)"
	git push origin $(VERSION)
	gh release create $(VERSION) --title $(VERSION) --notes-file docs/releases/$(VERSION).md --verify-tag
