.PHONY: help build install test clean fmt lint vet deps release snapshot check-release

# Variables
BINARY_NAME=cflip
VERSION?=latest
COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)"

# Go related variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build directories
BUILD_DIR=bin
DIST_DIR=dist

# Platforms for cross-compilation
PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Install dependencies
	$(GOGET) -u github.com/spf13/cobra@latest
	$(GOGET) -u github.com/spf13/viper@latest
	$(GOMOD) tidy

build: ## Build the binary for current platform
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/$(BINARY_NAME)

build-all: ## Build binaries for all platforms
	@mkdir -p $(DIST_DIR)
	@$(foreach platform,$(PLATFORMS), \
		echo "Building for $(platform)..." && \
		GOOS=$(word 1,$(subst /, ,$(platform))) \
		GOARCH=$(word 2,$(subst /, ,$(platform))) \
		$(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-$(word 1,$(subst /, ,$(platform))-$(word 2,$(subst /, ,$(platform)))$(if $(findstring windows,$(platform)),.exe,) ./cmd/$(BINARY_NAME); \
	)

install: build ## Install the binary to $GOPATH/bin
	$(GOCMD) install $(LDFLAGS) ./cmd/$(BINARY_NAME)

test: ## Run tests
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -f coverage.out coverage.html

fmt: ## Format Go code
	$(GOCMD) fmt ./...

lint: ## Run linter
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install it from https://golangci-lint.run/"; \
	fi

vet: ## Run go vet
	$(GOCMD) vet ./...

pre-commit: fmt vet test ## Run pre-commit checks

release: clean test build-all ## Create a release with binaries
	@echo "Release artifacts created in $(DIST_DIR)/"

dev: ## Run in development mode
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/$(BINARY_NAME)
	./$(BUILD_DIR)/$(BINARY_NAME) --help

run: build ## Build and run
	./$(BUILD_DIR)/$(BINARY_NAME)

# GoReleaser commands
check-release: ## Check GoReleaser configuration
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser check; \
	else \
		echo "GoReleaser not installed. Install it with: brew install goreleaser"; \
	fi

snapshot: ## Build snapshot with GoReleaser
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser build --snapshot --clean; \
	else \
		echo "GoReleaser not installed. Install it with: brew install goreleaser"; \
	fi

release: ## Release with GoReleaser (requires tag)
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --clean; \
	else \
		echo "GoReleaser not installed. Install it with: brew install goreleaser"; \
	fi