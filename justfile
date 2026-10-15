linter := "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6.2"

# Go environment (requires jsonv2 experiment)
export GOEXPERIMENT := env("GOEXPERIMENT", "jsonv2")

# List recipes
default:
    @just --list

# Tidy, test, lint, vet and build the examples
ensure-valid: tidy test lint vet examples

# Run unit tests
test: test-unit

# Run unit tests for the package and test module
test-unit:
    go test -v -race -coverprofile=test/coverage.txt -covermode=atomic ./...
    cd test && go test -v -race ./...

# Run fuzz corpus regression tests
test-corpus:
    cd test && go test -v -run=TestFuzzCorpus

# Run all tests (unit + corpus)
test-all: test-unit test-corpus

# Run golangci-lint
lint:
    go run {{linter}} run ./... --timeout=5m

# Format code with gofmt
fmt:
    gofmt -s -w .

# Run go vet
vet:
    go vet ./...

# Run go mod tidy (main + test)
tidy:
    go mod tidy
    cd test && go mod tidy

# Build the package
build:
    go build ./...

# Build examples to ./bin/
examples:
    #!/usr/bin/env bash
    set -euo pipefail
    mkdir -p bin
    for example in examples/*/; do
        name=$(basename "$example")
        echo "Building $name..."
        (
            cd "$example"
            go mod init example 2>/dev/null || true
            go mod edit -replace github.com/mikeschinkel/go-cfgstore=../..
            go mod tidy
            go build -o "../../bin/$name" .
        )
    done
    echo "Examples built to ./bin/"

# Clean build artifacts
clean:
    go clean
    rm -f coverage.txt test/coverage.txt
    rm -rf bin
    cd test && go clean

# Run all CI checks (fmt, vet, lint, test-all)
ci: fmt vet lint test-all
    @echo "All CI checks passed!"
