# Go Stats Generator Makefile

# Variables
BINARY_NAME=go-stats-generator
VERSION=1.0.0
BUILD_DIR=build
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Default target
.PHONY: all
all: build

# Build the application
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

# Build for multiple platforms
.PHONY: build-all
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

# Install the application
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) .

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmarks
.PHONY: bench
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Lint the code
.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		go vet ./...; \
	fi

# Format the code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Tidy dependencies
.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Run the application on test data
.PHONY: demo
demo: build
	@echo "Running demo on test data..."
	./$(BUILD_DIR)/$(BINARY_NAME) analyze testdata/simple

# Generate test data
.PHONY: generate-testdata
generate-testdata:
	@echo "Generating additional test data..."
	@mkdir -p testdata/complex
	@# This would generate more complex test scenarios

# Development setup
.PHONY: dev-setup
dev-setup:
	@echo "Setting up development environment..."
	go mod download
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi

# Check for security vulnerabilities
.PHONY: security
security:
	@echo "Checking for security vulnerabilities..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not installed. Install with: go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi

# Analyze the current project with itself
.PHONY: self-analyze
self-analyze: build
	@echo "Analyzing this project with itself..."
	./$(BUILD_DIR)/$(BINARY_NAME) analyze . --format console --verbose

# Generate documentation
.PHONY: docs
docs:
	@echo "Generating documentation..."
	@if command -v godoc >/dev/null 2>&1; then \
		echo "Starting godoc server on http://localhost:6060"; \
		godoc -http=:6060; \
	else \
		echo "godoc not installed. Install with: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

# Release preparation
.PHONY: release-prep
release-prep: clean tidy fmt lint test build-all
	@echo "Release preparation complete"

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          Build the application"
	@echo "  build-all      Build for multiple platforms"
	@echo "  install        Install the application"
	@echo "  test           Run tests"
	@echo "  test-coverage  Run tests with coverage report"
	@echo "  bench          Run benchmarks"
	@echo "  lint           Run linter"
	@echo "  fmt            Format code"
	@echo "  tidy           Tidy dependencies"
	@echo "  clean          Clean build artifacts"
	@echo "  demo           Run demo on test data"
	@echo "  self-analyze   Analyze this project with itself"
	@echo "  dev-setup      Set up development environment"
	@echo "  security       Check for security vulnerabilities"
	@echo "  docs           Generate and serve documentation"
	@echo "  release-prep   Prepare for release"
	@echo "  help           Show this help message"
