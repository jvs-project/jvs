.PHONY: build test lint conformance verify security sec fuzz

build:
	go build -o bin/jvs ./cmd/jvs

test:
	go test ./internal/... ./pkg/...

conformance:
	go test -tags conformance -count=1 -v ./test/conformance/...

lint:
	golangci-lint run ./...

verify: test lint

security: sec

sec:
	@echo "Running security scans..."
	go install github.com/securecodewarrior/gosec/v2@latest || true
	gosec -verbose=text -fmt=json -out gosec-report.json ./... || true
	go install honnef.co/go/tools/cmd/staticcheck@latest || true
	staticcheck ./... || true
	@echo "Security scan complete. See gosec-report.json for details."

fuzz:
	@echo "Running fuzzing tests (10 seconds each)..."
	@for target in FuzzValidateName FuzzValidateTag FuzzParseSnapshotID FuzzCanonicalMarshal FuzzDescriptorJSON FuzzSnapshotIDString; do \
		echo "Fuzzing $$target..."; \
		go test -fuzz="$$target" -fuzztime=10s ./test/fuzz/... || exit 1; \
	done
	@echo "All fuzzing tests passed."
