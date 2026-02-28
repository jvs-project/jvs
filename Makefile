.PHONY: build test lint conformance verify security sec fuzz test-race test-cover test-all integration release-gate clean

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

test-race:
	go test -race -count=1 ./internal/... ./pkg/...

test-cover:
	go test -coverprofile=coverage.out -covermode=atomic ./internal/... ./pkg/...
	@go tool cover -func=coverage.out | awk '/^total:/ { gsub(/%/, "", $$3); if ($$3+0 < 60) { printf "FAIL: coverage %.1f%% < 60%% threshold\n", $$3; exit 1 } else { printf "OK: coverage %.1f%% >= 60%% threshold\n", $$3 } }'

test-all: test conformance fuzz

integration: build conformance

release-gate: test-race test-cover lint build conformance fuzz
	@echo "RELEASE GATE PASSED"

clean:
	rm -rf bin/
	rm -f coverage.out gosec-report.json
