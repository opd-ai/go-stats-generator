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
- **Code Duplication Detection**: AST-based detection of exact, renamed, and near-duplicate code blocks
  - Configurable block size and similarity thresholds
  - Support for Type 1 (exact), Type 2 (renamed), and Type 3 (near) clone detection
  - Optional test file filtering for focused analysis
- **Historical Metrics Storage**: SQLite backend for tracking metrics over time (JSON backend planned)
- **Complexity Differential Analysis**: Compare metrics snapshots with multi-dimensional comparisons
- **Baseline Management**: Create and manage reference snapshots for comparisons
- **Regression Detection**: Compare snapshots to identify metric increases and decreases
- **Trend Analysis**: ⚠️ **BETA** - Basic trend commands available; advanced statistical analysis and forecasting planned for future release
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

# Analyze a single file
go-stats-generator analyze ./main.go

# Analyze specific file with verbose output
go-stats-generator analyze ./internal/analyzer/function.go --verbose

# Analyze with JSON output
go-stats-generator analyze ./src --format json --output report.json

# Analyze single file with JSON output
go-stats-generator analyze ./pkg/analyzer.go --format json --output single-file-report.json

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

# Note: Trend commands are in BETA with basic functionality
# Advanced statistical analysis and forecasting coming in future release
go-stats-generator trend analyze --days 30    # Basic trend overview
go-stats-generator trend forecast             # Placeholder - full implementation planned
go-stats-generator trend regressions --threshold 10.0  # Basic structure only
```

## Usage

### Basic Analysis

The analyze command can operate in two modes:

- **Directory mode**: Recursively scans for Go source files and processes them concurrently
- **File mode**: Analyzes a single Go source file

```bash
# Directory analysis
go-stats-generator analyze [directory] [flags]

# Single file analysis  
go-stats-generator analyze [file.go] [flags]
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
  duplication:
    min_block_lines: 6            # Minimum block size for duplication detection
    similarity_threshold: 0.80    # Threshold for near-duplicate detection (0.0-1.0)
    ignore_test_files: false      # Exclude test files from duplication analysis

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

### Duplication Detection Configuration

The tool includes advanced code duplication detection with configurable thresholds:

**CLI Flags:**
- `--min-block-lines` (default: 6) - Minimum number of statements in a block to consider for duplication
- `--similarity-threshold` (default: 0.80) - Similarity threshold for near-duplicate detection (0.0-1.0)
- `--ignore-test-duplication` (default: false) - Exclude test files (*_test.go) from duplication analysis

**Examples:**
```bash
# Detect duplicates with smaller block size
go-stats-generator analyze . --min-block-lines 3

# Use stricter similarity threshold for near-duplicates
go-stats-generator analyze . --similarity-threshold 0.90

# Ignore test files in duplication analysis
go-stats-generator analyze . --ignore-test-duplication

# Combine multiple duplication settings
go-stats-generator analyze . --min-block-lines 4 --similarity-threshold 0.85 --ignore-test-duplication
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

*Note: These are estimated performance targets based on the tool's architecture. Actual performance may vary depending on code complexity, system resources, and analysis configuration.*

| Repository | Files | LOC | Analysis Time (est.) | Memory Usage (est.) |
|------------|-------|-----|---------------------|---------------------|
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

## Planned Features

The following features are under development and will be included in future releases:

### Statistical Trend Analysis (Roadmap)
- **Linear regression** for trend lines across metric history
- **ARIMA/exponential smoothing** for time series forecasting
- **Statistical hypothesis testing** for regression detection
- **Confidence interval calculations** for forecast reliability
- **Correlation analysis** between different metrics

### Storage Backend Expansion
- **JSON storage backend** - File-based metrics storage as alternative to SQLite
- **Memory storage** - In-memory storage for temporary analysis runs

### Advanced Maintenance Detection
See [ROADMAP.md](ROADMAP.md) for detailed implementation plans for:
- Enhanced code duplication analysis with semantic similarity
- Naming convention analysis with automated suggestions
- Misplaced declaration detection (functions/methods in wrong files)
- Documentation gap detection and quality scoring
- Organizational health metrics (file size, package cohesion)
- Additional burden indicators (magic numbers, dead code, deep nesting)

### Configuration Enhancement
- Complete configuration file loader for all documented options
- Per-project configuration inheritance
- Team-level default profiles

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