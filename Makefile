.PHONY: build test lint conformance verify

build:
	go build -o bin/jvs ./cmd/jvs

test:
	go test ./internal/... ./pkg/...

conformance:
	go test -tags conformance -count=1 -v ./test/conformance/...

lint:
	golangci-lint run ./...

verify: test lint
