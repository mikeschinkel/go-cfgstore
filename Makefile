.PHONY: test test-verbose test-coverage build clean help

# Enable JSON v2 experiment for Go 1.25+
export GOEXPERIMENT=jsonv2

help:
	@echo "Available targets:"
	@echo "  test          - Run tests"
	@echo "  test-verbose  - Run tests with verbose output"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  build         - Build the package"
	@echo "  clean         - Clean build artifacts"
	@echo "  help          - Show this help message"

test:
	go test ./...

test-verbose:
	go test -v ./...

test-coverage:
	go test -cover ./...

build:
	go build ./...

clean:
	go clean ./...
