.PHONY: build install test clean run help

# Binary name
BINARY_NAME=unifidoc
BINARY_PATH=./bin/$(BINARY_NAME)

# Go related variables
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard *.go)

# Build flags
BUILD_FLAGS=-ldflags="-s -w"

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(GOBIN)
	@go build $(BUILD_FLAGS) -o $(BINARY_PATH) ./cmd/unifidoc
	@echo "✓ Built $(BINARY_PATH)"

install: build ## Install the binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@go install ./cmd/unifidoc
	@echo "✓ Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

test-coverage: test ## Run tests and show coverage in browser
	@go tool cover -html=coverage.out

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(GOBIN)
	@rm -f coverage.out
	@rm -rf docs
	@echo "✓ Cleaned"

run: build ## Build and run
	@$(BINARY_PATH)

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Formatted"

lint: ## Run linter
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies downloaded"

example: build ## Generate example documentation
	@echo "Generating example documentation..."
	@$(BINARY_PATH) init --auto-detect
	@$(BINARY_PATH) generate
	@echo "✓ Example documentation generated in ./docs"

serve-example: example ## Generate and serve example documentation
	@$(BINARY_PATH) serve

# Development targets
dev-setup: deps ## Setup development environment
	@echo "Setting up development environment..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✓ Development environment ready"

all: clean deps build test ## Clean, download deps, build, and test

.DEFAULT_GOAL := help
