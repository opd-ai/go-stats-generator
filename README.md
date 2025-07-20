# Go Source Code Statistics Generator

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/opd-ai/go-stats-generator)](https://goreportcard.com/report/github.com/opd-ai/go-stats-generator)

`go-stats-generator` is a high-performance command-line tool that analyzes Go source code repositories to generate comprehensive statistical reports about code structure, complexity, and patterns. The project focuses on computing obscure and detailed metrics that standard linters don't typically capture, providing actionable insights for code quality assessment and refactoring decisions.

## Features

- **Function and Method Analysis**: Precise line counting, cyclomatic complexity, signature complexity
- **Struct Complexity Metrics**: Detailed member categorization by type
- **Advanced Pattern Detection**: Design patterns, concurrency patterns, anti-patterns
- **Historical Metrics Storage**: SQLite/JSON backends for tracking metrics over time
- **Complexity Differential Analysis**: Compare metrics snapshots with multi-dimensional comparisons
- **Regression Detection**: Automated detection of complexity regressions with configurable thresholds
- **Trend Analysis**: Statistical analysis of metrics over time with forecasting capabilities
- **Baseline Management**: Create and manage reference snapshots for comparisons
- **CI/CD Integration**: Exit codes and reporting for automated quality gates
- **Concurrent Processing**: Worker pools for analyzing large codebases efficiently
- **Multiple Output Formats**: Console, JSON, CSV, HTML, Markdown with rich reporting
- **Enterprise Scale**: Process 50,000+ files within 60 seconds, <1GB memory usage
- **Configurable Analysis**: Flexible filtering, thresholds, and analysis options

## Installation

### From Source

```bash
git clone https://github.com/opd-ai/go-stats-generator.git
cd go-stats-generator
go build -o go-stats-generator .
```

### Using Go Install

```bash
go install github.com/opd-ai/go-stats-generator@latest
```

After installation, the binary will be available as `go-stats-generator` in your `$GOPATH/bin` directory (which should be in your `$PATH`). You can verify the installation by running:

```bash
go-stats-generator version
```

## Quick Start

```bash
# Analyze current directory
go-stats-generator analyze .

# Analyze with JSON output
go-stats-generator analyze ./src --format json --output report.json

# Analyze excluding test files
go-stats-generator analyze . --skip-tests

# Analyze with custom complexity thresholds
go-stats-generator analyze . --max-function-length 50 --max-complexity 15
```

## Usage

### Basic Analysis

```bash
go-stats-generator analyze [directory] [flags]
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--format` | Output format (console, json, csv, html, markdown) | console |
| `--output` | Output file (default: stdout) | - |
| `--workers` | Number of worker goroutines | CPU cores |
| `--timeout` | Analysis timeout | 10m |
| `--skip-vendor` | Skip vendor directories | true |
| `--skip-tests` | Skip test files (*_test.go) | false |
| `--skip-generated` | Skip generated files | true |
| `--include` | Include patterns (glob) | **/*.go |
| `--exclude` | Exclude patterns (glob) | - |
| `--max-function-length` | Maximum function length threshold | 30 |
| `--max-complexity` | Maximum cyclomatic complexity threshold | 10 |
| `--verbose` | Verbose output | false |

### Example Output

```
=== GO SOURCE CODE STATISTICS REPORT ===
Repository: /path/to/project
Generated: 2025-07-20 16:20:02
Analysis Time: 1.234s
Files Processed: 156

=== OVERVIEW ===
Total Lines of Code: 45,123
Total Functions: 1,234
Total Methods: 856
Total Structs: 245
Total Interfaces: 67
Total Packages: 23
Total Files: 156

=== FUNCTION ANALYSIS ===
Function Statistics:
  Average Function Length: 15.4 lines
  Longest Function: ProcessComplexData (127 lines)
  Functions > 50 lines: 23 (1.9%)
  Functions > 100 lines: 3 (0.2%)
  Average Complexity: 4.2
  High Complexity (>10): 45 functions

=== COMPLEXITY ANALYSIS ===
Top 10 Most Complex Functions:
Function                       Package                 Lines Cyclomatic    Overall
--------------------------------------------------------------------------------
ProcessComplexData             processor                  127         23       45.2
HandleUserRequest              handler                     89         18       32.1
ValidateConfiguration          config                      67         15       28.3
...
```

## Configuration

Create a `.gostats.yaml` file in your home directory or project root:

```yaml
analysis:
  include_functions: true
  include_structs: true
  include_patterns: true
  max_function_length: 30
  max_cyclomatic_complexity: 10

output:
  format: console
  use_colors: true
  show_progress: true
  include_examples: false

performance:
  worker_count: 8
  timeout: 10m
  enable_cache: true

filters:
  skip_vendor: true
  skip_test_files: false
  skip_generated: true
  include_patterns:
    - "**/*.go"
  exclude_patterns:
    - "vendor/**"
    - "*.pb.go"
```

## Metrics Explained

### Function Metrics

- **Cyclomatic Complexity**: Number of independent paths through the code
- **Cognitive Complexity**: How difficult the code is to understand
- **Nesting Depth**: Maximum level of nested blocks
- **Signature Complexity**: Based on parameter count, return values, generics

### Line Counting

Lines are counted precisely, excluding:
- Comments (single-line `//` and multi-line `/* */`)
- Blank lines
- Opening and closing braces of function bodies

### Complexity Thresholds

| Complexity | Rating | Recommendation |
|------------|--------|----------------|
| 1-5 | Low | Good |
| 6-10 | Moderate | Acceptable |
| 11-20 | High | Consider refactoring |
| 21+ | Very High | Refactor immediately |

## Architecture

```
github.com/opd-ai/go-stats-generator/
├── cmd/                    # CLI commands
├── internal/
│   ├── analyzer/          # AST analysis engines
│   ├── metrics/           # Metric data structures
│   ├── reporter/          # Output formatters
│   ├── scanner/           # File discovery and processing
│   └── config/            # Configuration management
├── pkg/gostats/          # Public API
└── testdata/             # Test data
```

## Performance

- **Large Codebases**: Tested on repositories with 50,000+ Go files
- **Memory Efficient**: Processes files in batches, <1GB memory usage
- **Concurrent**: Configurable worker pools (default: number of CPU cores)
- **Fast**: Completes analysis of most projects in seconds

### Benchmarks

| Repository | Files | LOC | Analysis Time | Memory Usage |
|------------|-------|-----|---------------|--------------|
| Standard Library | 400+ | 500K+ | <10s | <100MB |
| Kubernetes | 10K+ | 2M+ | <60s | <800MB |
| Docker | 2K+ | 300K+ | <15s | <200MB |

## Development

### Building

```bash
make build
```

### Testing

```bash
make test
make test-coverage
```

### Linting

```bash
make lint
```

## API Usage

```go
package main

import (
    "context"
    "fmt"
    "os"
    
    "github.com/opd-ai/go-stats-generator/pkg/gostats"
)

func main() {
    analyzer := gostats.NewAnalyzer()
    
    report, err := analyzer.AnalyzeDirectory(context.Background(), "./src")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Analysis failed: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Printf("Found %d functions with average complexity %.1f\n", 
        len(report.Functions), report.Complexity.AverageFunction)
}
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Functions must be under 30 lines
- Maximum cyclomatic complexity of 10
- Test coverage >85% for business logic
- All exported functions must have GoDoc comments

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Inspired by the need for advanced Go code analysis beyond standard linters
- Built with the Go standard library AST package for zero-dependency core functionality
- Uses Cobra for CLI framework and go-pretty for rich console output