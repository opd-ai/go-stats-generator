# Project Overview

The `go-stats-generator` is a high-performance command-line tool that analyzes Go source code repositories to generate comprehensive statistical reports about code structure, complexity, and patterns. The project focuses on computing obscure and detailed metrics that standard linters don't typically capture, providing actionable insights for code quality assessment and refactoring decisions. The tool is designed for enterprise-scale codebases, supporting concurrent processing of 50,000+ files within 60 seconds while maintaining memory usage under 1GB.

The target audience includes software engineers, technical leads, and DevOps teams working with large Go codebases who need detailed code analysis beyond basic linting. The tool emphasizes advanced metrics like design pattern detection, struct member categorization, generic usage analysis, and documentation quality assessment. It serves as a comprehensive code health assessment platform that helps teams identify technical debt, architectural issues, and optimization opportunities.

## Technical Stack
- **Primary Language**: Go (targeting Go 1.18+ for generic analysis support)
- **CLI Framework**: github.com/spf13/cobra v1.7+ for command structure, github.com/spf13/viper for configuration management
- **AST Processing**: Go standard library (go/parser, go/ast, go/token) for zero-dependency core functionality
- **Output Formatting**: github.com/jedib0t/go-pretty/v6 for rich console tables and progress bars, github.com/olekukonko/tablewriter for ASCII output
- **Testing**: Go standard library testing package with github.com/stretchr/testify for enhanced assertions and mocking
- **Build/Deploy**: Standard Go build tools, Makefile for automation, GitHub Actions CI/CD, GoReleaser for multi-platform releases

## Code Assistance Guidelines

1. **Function Length and Complexity**: Maintain maximum function length of 30 lines excluding comments and blank lines. Each function should have a single responsibility. Extract complex logic into separate functions rather than increasing nesting depth beyond 3 levels. Use precise line counting that excludes comments, blank lines, and braces for accurate metrics.

2. **AST Processing Patterns**: Use the visitor pattern for AST traversal with specialized analyzers for different node types. Implement proper error handling and recovery for malformed Go files. Utilize token.FileSet for accurate position tracking and line number calculations. Create separate analyzer structs for functions, structs, packages, and patterns rather than monolithic processing.

3. **Concurrent Processing Architecture**: Implement worker pool patterns for file processing with configurable worker counts (default: runtime.NumCPU()). Use channels for work distribution and result aggregation. Ensure proper cleanup and cancellation support for long-running operations. Memory management is critical for large codebases - avoid loading entire file contents into memory simultaneously.

4. **Metric Data Structures**: Design comprehensive metric types with JSON serialization support for multiple output formats. Categorize struct members by type (primitives, slices, maps, channels, interfaces, custom types, embedded types, functions). Include confidence scores for pattern detection and provide detailed examples in results.

5. **Error Handling Standards**: Use wrapped errors with fmt.Errorf and %w verb for error chains. Provide meaningful context including file paths, line numbers, and operation details. Gracefully handle syntax errors in source files without stopping analysis of other files. Log warnings for non-critical issues while continuing processing.

6. **Testing Requirements**: Achieve 85%+ code coverage on business logic functions. Write table-driven tests for metric calculations. Include integration tests against known Go projects to verify accuracy. Create comprehensive test data sets including edge cases, large files, and malformed code. Performance tests should validate requirements for processing speed and memory usage.

7. **Output Format Consistency**: Support multiple output formats (console, JSON, CSV, HTML, Markdown) with consistent data representation. JSON output must include schema validation. Console output should use professional formatting with color coding and progress indication. HTML reports should be responsive and include interactive charts for metric visualization.

## Project Context

- **Domain**: Static code analysis and software quality metrics for Go programming language. Focus on advanced, non-standard metrics that provide actionable insights for large-scale software development teams. Emphasis on detecting design patterns, architectural issues, and code complexity beyond traditional linting.

- **Architecture**: Modular design using internal packages for separation of concerns: analyzer/ for AST processing and metric calculation, reporter/ for output formatting, scanner/ for file discovery and filtering, metrics/ for data structures and aggregation, config/ for configuration management. Public API exposed through pkg/gostats/ for library usage.

- **Key Directories**: 
  - `cmd/` - CLI command definitions (root, analyze, compare, version)
  - `internal/analyzer/` - Core analysis engines for functions, structs, patterns, complexity
  - `internal/metrics/` - Metric data structures, aggregation, and calculations
  - `internal/reporter/` - Output formatters for console, JSON, CSV, HTML, Markdown
  - `internal/scanner/` - File discovery, filtering, and concurrent processing
  - `pkg/gostats/` - Public API for library integration
  - `testdata/` - Test projects for validation and benchmarking

- **Configuration**: Support for configuration files (YAML/JSON), command-line flags, and environment variables. Configurable worker pool sizes, output formats, filtering patterns, and metric thresholds. Default settings optimized for typical Go project structures and enterprise-scale analysis requirements.

## Quality Standards

- **Testing Requirements**: Maintain >85% code coverage using Go's built-in testing package with testify for enhanced assertions. Write table-driven tests for all metric calculations and AST processing functions. Include integration tests that verify accuracy against known Go projects like Kubernetes, Docker, and Prometheus. Performance benchmarks must validate processing 50,000+ files within 60 seconds with <1GB memory usage.

- **Code Review Criteria**: All functions must be under 30 lines with single responsibility. Maximum of 5 parameters per function (use structs for complex inputs). Error handling must provide meaningful context with wrapped errors. External dependencies limited to well-maintained libraries (>1000 GitHub stars, recent updates). Concurrent processing must be properly tested with race condition detection.

- **Documentation Standards**: All exported functions, types, and packages must have GoDoc comments. Include code examples in documentation for complex operations. Maintain comprehensive README with installation instructions, usage examples, and metric definitions. API documentation must explain all obscure metrics with examples and calculation methods. Update documentation when adding new metrics or changing behavior.