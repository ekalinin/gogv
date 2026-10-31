GO ?= go
REFERENCE_IMAGE ?= gogv-reference

.PHONY: build test lint provenance reference-image reference harvest conformance determinism

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

# Records the reference outputs of the curated cases.
reference:
	$(GO) run ./tools/refgen -image $(REFERENCE_IMAGE) -cases testdata/conformance/cases -out testdata/conformance/reference

# Collects the DOT that gopuml generates for its conformance cases into
# .cache/gopuml-dot (GOPUML is a gopuml checkout), and records its references.
GOPUML ?= ../plantuml-in-go

harvest:
	$(GO) run ./tools/harvest -gopuml $(GOPUML) -image $(REFERENCE_IMAGE)
	$(GO) run ./tools/refgen -image $(REFERENCE_IMAGE) -cases .cache/gopuml-dot -out .cache/reference/gopuml-dot

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
