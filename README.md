# Go Source Code Statistics Generator

[![License: MIT](https://img.shields.io/badge/Li### Basic Analysis

```bash
go-stats-generator analyze [directory] [flags]
```-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/opd-ai/go-stats-generator)](https://goreportcard.com/report/github.com/opd-ai/go-stats-generator)

`go-stats-generator` is a high-performance command-line tool that analyzes Go source code repositories to generate comprehensive statistical reports about code structure, complexity, and patterns. The project focuses on computing obscure and detailed metrics that standard linters don't typically capture, providing actionable insights for code quality assessment and refactoring decisions.

## Features

- **Precise Line Counting**: Advanced function/method line analysis that accurately categorizes code, comments, and blank lines
  - Excludes braces, comments, and blank lines from function length calculations
  - Handles complex scenarios: inline comments, multi-line block comments, mixed lines
  - Provides detailed breakdown: total, code, comment, and blank line counts
- **Function and Method Analysis**: Cyclomatic complexity, signature complexity, parameter analysis
- **Struct Complexity Metrics**: Detailed member categorization by type with method analysis
- **Package Dependency Analysis**: Architectural insights with dependency tracking and circular detection
  - Dependency graph analysis with internal/external package filtering
  - Circular dependency detection with severity classification (low/medium/high)
  - Package cohesion metrics for design quality assessment
  - Package coupling metrics for architectural complexity measurement
- **Advanced Pattern Detection**: Design patterns, concurrency patterns, anti-patterns
- **Historical Metrics Storage**: SQLite/JSON backends for tracking metrics over time
- **Complexity Differential Analysis**: Compare metrics snapshots with multi-dimensional comparisons
- **Regression Detection**: Automated detection of complexity regressions with configurable thresholds
- **Trend Analysis**: Statistical analysis of metrics over time with forecasting capabilities
- **Baseline Management**: Create and manage reference snapshots for comparisons
- **CI/CD Integration**: Exit codes and reporting for automated quality gates
- **Concurrent Processing**: Worker pools for analyzing large codebases efficiently
- **Multiple Output Formats**: Console, JSON, HTML with rich reporting
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

# Create a baseline snapshot
go-stats-generator baseline create . --id "v1.0.0" --message "Initial baseline"

# Compare with baseline
go-stats-generator diff baseline-report.json current-report.json

# List all baselines
go-stats-generator baseline list

# Analyze trends over time
go-stats-generator trend analyze --days 30

# Detect regressions
go-stats-generator trend regressions --threshold 10.0
```

## Usage

### Basic Analysis

```bash
go-stats-generator analyze [directory] [flags]
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--format` | Output format (console, json, html) | console |
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

<!-- AUDIT_FLAG: NEEDS_REVIEW
Issue: CSV and Markdown output formats mentioned in table but not implemented
Found in code: CSV and Markdown reporters exist but return "not yet implemented" errors
Current working formats: console, json, html
Action needed: Either remove from docs or implement the formats
Last verified: 2025-01-25 against current codebase
-->

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

Create a `.go-stats-generator.yaml` file in your home directory or project root:

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

### Line Counting Methodology

The tool implements precise line counting that provides detailed breakdowns for function analysis:

#### Line Categories
- **Code Lines**: Lines containing executable code, variable declarations, control flow statements
- **Comment Lines**: Single-line (`//`) and multi-line (`/* */`) comments  
- **Blank Lines**: Empty lines or lines containing only whitespace
- **Total Lines**: Sum of all categories (excluding function braces)

#### Advanced Handling
- **Mixed Lines**: Lines with both code and comments are classified as code lines
- **Multi-line Comments**: Each line of a block comment is counted separately
- **Inline Comments**: Code with trailing comments counts as code
- **Function Boundaries**: Opening and closing braces are excluded from counts

#### Example Analysis
```go
func example() {
    // This is a comment line
    var x int = 42 // This is a code line (mixed)
    
    /*
     * Multi-line comment
     * spans multiple lines  
     */
    
    if x > 0 { // Another code line
        return x
    }
}
```

**Result**: 4 code lines, 5 comment lines, 2 blank lines

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
├── pkg/go-stats-generator/          # Public API
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
    
    "github.com/opd-ai/go-stats-generator/pkg/go-stats-generator"
)

func main() {
    analyzer := go_stats_generator.NewAnalyzer()
    
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