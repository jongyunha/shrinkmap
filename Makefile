# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Build info
BINARY_NAME=shrinkmap
VERSION ?= $(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Test parameters
TEST_TIMEOUT=5m
COVERAGE_OUT=coverage.out

.PHONY: all build clean test coverage test-race bench lint fmt vet mod-tidy mod-verify help

all: fmt lint test build

build:
	$(GOBUILD) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(COVERAGE_OUT)

test:
	$(GOTEST) -v -timeout $(TEST_TIMEOUT) ./...

coverage:
	$(GOTEST) -v -coverprofile=$(COVERAGE_OUT) -covermode=atomic ./...
	$(GOCMD) tool cover -html=$(COVERAGE_OUT) -o coverage.html

test-race:
	$(GOTEST) -v -race -timeout $(TEST_TIMEOUT) ./...

bench:
	$(GOTEST) -bench=. -benchmem -count=3 ./...

lint:
	$(GOLINT) run --config=.golangci.yml

fmt:
	$(GOFMT) -s -w .

vet:
	$(GOCMD) vet ./...

mod-tidy:
	$(GOMOD) tidy

mod-verify:
	$(GOMOD) verify

# CI targets
ci-test: test-race coverage lint

ci-bench: bench

# Development targets
dev-setup:
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint

# Release targets
release-check:
	@echo "Checking if ready for release..."
	@$(GOTEST) -v -race ./...
	@$(GOLINT) run --config=.golangci.yml
	@$(GOMOD) verify
	@echo "Ready for release!"

# Help
help:
	@echo "Available targets:"
	@echo "  all          - Format, lint, test, and build"
	@echo "  build        - Build the project"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  coverage     - Run tests with coverage"
	@echo "  test-race    - Run tests with race detection"
	@echo "  bench        - Run benchmarks"
	@echo "  lint         - Run linters"
	@echo "  fmt          - Format code"
	@echo "  vet          - Run go vet"
	@echo "  mod-tidy     - Tidy go modules"
	@echo "  mod-verify   - Verify go modules"
	@echo "  ci-test      - Run CI test suite"
	@echo "  ci-bench     - Run CI benchmarks"
	@echo "  dev-setup    - Install development dependencies"
	@echo "  release-check - Check if ready for release"
	@echo "  help         - Show this help message"
