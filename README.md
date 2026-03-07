# Go Source Code Statistics Generator

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/opd-ai/go-stats-generator)](https://goreportcard.com/report/github.com/opd-ai/go-stats-generator)

`go-stats-generator` is a high-performance command-line tool that analyzes Go source code repositories to generate comprehensive statistical reports about code structure, complexity, and patterns. The project focuses on computing obscure and detailed metrics that standard linters don't typically capture, providing actionable insights for code quality assessment and refactoring decisions.

## Features

### Production-Ready Features

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
- **Historical Metrics Storage**: SQLite, JSON, and in-memory backends for tracking metrics over time
- **Complexity Differential Analysis**: Compare metrics snapshots with multi-dimensional comparisons
- **Baseline Management**: Create and manage reference snapshots for comparisons
- **Regression Detection**: Compare snapshots to identify metric increases and decreases
- **CI/CD Integration**: Exit codes and reporting for automated quality gates
- **Concurrent Processing**: Worker pools for analyzing large codebases efficiently
- **Multiple Output Formats**: Console, JSON, HTML, CSV, and Markdown with rich reporting
- **Enterprise Scale**: Designed for large codebases with concurrent processing
- **Configurable Analysis**: Flexible filtering, thresholds, and analysis options
- **Trend Analysis**: Statistical analysis of code metrics over time
  - Linear regression trend lines with R² coefficients for measuring trend strength
  - Statistical forecasting (7/14/30-day predictions) with 95% confidence intervals
  - Regression detection with hypothesis testing and p-values for significance assessment
  - Future enhancements: ARIMA forecasting, exponential smoothing (see [Planned Features](#planned-features))

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

# Trend analysis with statistical forecasting
go-stats-generator trend analyze --days 30            # Analyze trends over 30 days
go-stats-generator trend forecast --days 30           # Forecast using linear regression
go-stats-generator trend regressions --threshold 10.0 # Detect statistical regressions
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
| `--max-burden-score` | Maximum Maintenance Burden Index (MBI) score (0-100) | 70.0 |
| `--min-doc-coverage` | Minimum documentation coverage (fraction) | 0.7 |
| `--enforce-thresholds` | Exit with code 1 if thresholds exceeded | false |
| `--verbose` | Verbose output | false |

### CI/CD Integration

Use the `--enforce-thresholds` flag with threshold flags to fail builds when quality standards are not met:

```bash
# Fail build if MBI score exceeds 50 (medium risk)
go-stats-generator analyze . --max-burden-score 50 --enforce-thresholds

# Fail build if documentation coverage below 80%
go-stats-generator analyze . --min-doc-coverage 0.8 --enforce-thresholds

# Combined quality gates
go-stats-generator analyze . \
  --max-burden-score 50 \
  --min-doc-coverage 0.8 \
  --max-complexity 15 \
  --enforce-thresholds
```

When `--enforce-thresholds` is enabled, the tool exits with code 1 if any threshold is violated, making it suitable for CI/CD pipelines. Violations are printed to stderr with details about which files/packages failed.

**GitHub Actions Example:**
```yaml
- name: Code Quality Check
  run: |
    go install github.com/opd-ai/go-stats-generator@latest
    go-stats-generator analyze . \
      --max-burden-score 70 \
      --min-doc-coverage 0.7 \
      --enforce-thresholds
```

### Trend Analysis

The `trend` command provides statistical analysis of code metrics over time using historical baseline snapshots.

#### Prerequisites

Trend analysis requires historical data. Create baseline snapshots periodically:

```bash
# Create initial baseline
go-stats-generator baseline create --name "sprint-1" --tags "release-1.0.0"

# Update baselines regularly (e.g., after each sprint or release)
go-stats-generator baseline create --name "sprint-2" --tags "release-1.1.0"
```

#### Analyze Trends

View how metrics have changed over time:

```bash
# Analyze trends over the last 30 days
go-stats-generator trend analyze --days 30

# Analyze specific metric
go-stats-generator trend analyze --days 30 --metric mbi_score

# JSON output for programmatic analysis
go-stats-generator trend analyze --days 30 --format json --output trend-report.json
```

Example output:
```
=== TREND ANALYSIS ===
Period: Last 30 days (15 snapshots)

MBI Score: 45.2 → 38.7 (▼ 14.4% improvement)
Duplication: 8.5% → 6.2% (▼ 27.1% improvement)
Documentation: 68.3% → 72.1% (▲ 5.6% improvement)
Complexity Violations: 23 → 18 (▼ 21.7% improvement)
```

#### Forecast Future Metrics

Generate statistical forecasts using linear regression:

```bash
# Forecast metrics for 7, 14, and 30 days ahead
go-stats-generator trend forecast --days 30

# Focus on specific metric
go-stats-generator trend forecast --days 30 --metric duplication_ratio
```

Example output:
```
=== METRIC FORECASTS ===

Metric: mbi_score
Method: linear_regression
Data Points: 15

Trend Line:
  y = -0.2134·x + 45.6721
  R² = 0.8456 (excellent fit) ✓

Forecasts:
   7 days (2026-03-10): 44.18  [42.34 - 46.02]
  14 days (2026-03-17): 42.69  [40.21 - 45.17]
  30 days (2026-04-02): 39.26  [35.18 - 43.34]
```

The R² value indicates forecast reliability:
- **≥0.8**: Excellent fit - high confidence in predictions
- **0.5-0.8**: Moderate fit - reasonable predictions with wider confidence intervals
- **<0.5**: Poor fit - unreliable forecast (warning displayed)

#### Detect Regressions

Identify significant deviations from historical trends:

```bash
# Detect regressions with 10% threshold
go-stats-generator trend regressions --threshold 10.0 --days 30

# Stricter detection (5% threshold)
go-stats-generator trend regressions --threshold 5.0 --days 30
```

Example output:
```
=== REGRESSION DETECTION ===

Threshold: 10.0%
Historical snapshots: 12
Recent snapshots: 3
Overall Severity: medium

Detected 2 regression(s):

1. [medium] ▼ duplication_ratio
   Current: 9.45  |  Expected: 7.82  |  Deviation: 20.8%
   P-value: 0.0234

2. [low] ▼ doc_coverage
   Current: 69.2  |  Expected: 72.1  |  Deviation: 4.0%
   P-value: 0.1456

```

Indicators:
- **▼** Regression (metric worsened)
- **▲** Improvement (metric improved)
- **→** Stable (no significant change)

Severity levels based on deviation:
- **critical**: >25% deviation from expected
- **high**: 15-25% deviation
- **medium**: 10-15% deviation
- **low**: <10% deviation

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

storage:
  type: sqlite                      # Storage backend: "sqlite" or "json"
  path: .go-stats-generator/metrics.db
  compression: true

maintenance:
  burden:
    max_params: 5                   # Maximum parameters before flagging function signature
    max_returns: 3                  # Maximum return values before flagging function signature
    max_nesting: 4                  # Maximum nesting depth before flagging deep nesting
    feature_envy_ratio: 2.0         # External reference threshold for feature envy detection
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

### Maintenance Burden Configuration

The tool detects maintenance burden indicators including magic numbers, dead code, complex signatures, deep nesting, and feature envy patterns:

**CLI Flags:**
- `--max-params` (default: 5) - Maximum function parameters before flagging high signature complexity
- `--max-returns` (default: 3) - Maximum return values before flagging high signature complexity
- `--max-nesting` (default: 4) - Maximum nesting depth before flagging deeply nested code
- `--feature-envy-ratio` (default: 2.0) - Threshold ratio for detecting feature envy (external references / self references)

**What is detected:**
- **Magic Numbers**: Numeric and string literals that should be named constants (excludes 0, 1, -1, "")
- **Dead Code**: Unreferenced unexported functions and unreachable code after return/panic/os.Exit
- **Signature Complexity**: Functions with too many parameters, return values, or boolean flag parameters
- **Deep Nesting**: Functions with excessive control structure nesting that should use guard clauses
- **Feature Envy**: Methods that reference external objects more than their own receiver (misplaced methods)

**Examples:**
```bash
# Stricter signature complexity thresholds
go-stats-generator analyze . --max-params 3 --max-returns 2

# Allow deeper nesting for complex algorithms
go-stats-generator analyze . --max-nesting 6

# More sensitive feature envy detection
go-stats-generator analyze . --feature-envy-ratio 1.5

# Combine maintenance burden settings
go-stats-generator analyze . --max-params 4 --max-nesting 3 --feature-envy-ratio 2.5
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
├── pkg/generator/          # Public API
└── testdata/             # Test data
```

## Performance

- **Fast Analysis**: 987 files/second on modern hardware (AMD Ryzen 7 7735HS)
- **Memory Efficient**: ~62 KB peak memory per file analyzed
- **Concurrent Processing**: Configurable worker pools (default: number of CPU cores)
- **Scalable**: Sub-second analysis for typical projects (<1,000 files)
- **Benchmarked**: Comprehensive performance tests validate throughput and memory usage

For detailed benchmark results and scaling projections, see [docs/PERFORMANCE.md](docs/PERFORMANCE.md).

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
    
    "github.com/opd-ai/go-stats-generator/pkg/generator"
)

func main() {
    analyzer := generator.NewAnalyzer()
    
    report, err := analyzer.AnalyzeDirectory(context.Background(), "./src")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Analysis failed: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Printf("Found %d functions with average complexity %.1f\n", 
        len(report.Functions), report.Complexity.AverageFunction)
}
```

## WebAssembly Browser Version

In addition to the CLI tool, `go-stats-generator` is available as a browser-based WebAssembly application hosted on GitHub Pages. This allows you to analyze public GitHub repositories entirely in your browser without installing anything.

### Live Demo

Visit the hosted application at: **https://opd-ai.github.io/go-stats-generator/**

### Features

- **Zero Installation**: No CLI installation required - runs entirely in your browser
- **Client-Side Analysis**: All processing happens locally in your browser for privacy and speed
- **GitHub Repository Support**: Analyze any public GitHub repository by URL
- **Multiple Output Formats**: View results as interactive HTML reports or downloadable JSON
- **Branch/Tag/Commit Support**: Analyze any git ref (branch, tag, or specific commit SHA)
- **Rate Limit Management**: Optional GitHub Personal Access Token support for increased API limits

### How to Use

1. Visit the GitHub Pages site
2. Enter a GitHub repository URL (e.g., `https://github.com/golang/example`)
3. (Optional) Specify a branch, tag, or commit SHA
4. (Optional) Add a GitHub Personal Access Token to increase rate limits
5. Click "Analyze" and view the results

### GitHub Pages Deployment

The WebAssembly application is automatically deployed to GitHub Pages via GitHub Actions on every push to the `main` branch.

#### Enabling GitHub Pages (One-Time Setup)

To enable GitHub Pages for this repository:

1. Navigate to **Settings** → **Pages** in the GitHub repository
2. Under **Source**, select **GitHub Actions** (not "Deploy from a branch")
3. The workflow will automatically deploy on the next push to `main`

The deployment workflow (`.github/workflows/deploy-pages.yml`) handles:
- Compiling the Go code to WebAssembly (`GOOS=js GOARCH=wasm`)
- Content-hashing the WASM binary for aggressive browser caching
- Copying static assets (HTML, CSS, JavaScript)
- Deploying to GitHub Pages

#### Local Development

To test the WebAssembly version locally:

```bash
# Build the WASM binary and assemble the site
make build-dist

# Serve locally (requires Python 3)
cd dist && python3 -m http.server 8080

# Visit http://localhost:8080 in your browser
```

The `build-dist` target creates a production-ready `dist/` directory with:
- Content-hashed WASM binary (`go-stats-generator.<hash>.wasm`)
- Static assets (HTML, CSS, JavaScript)
- Manifest file for dynamic WASM loading
- Cache headers configuration

#### WASM Build Details

The WebAssembly build uses platform-specific implementations:
- `cmd/wasm/main.go` - WASM entry point with JavaScript API
- `internal/scanner/discover_wasm.go` - In-memory file processing
- `internal/scanner/worker_wasm.go` - Sequential processing (no OS threads)
- `internal/storage/storage_wasm.go` - Storage stubs (analysis is one-shot)

All core analyzers (functions, structs, interfaces, packages, patterns, concurrency, duplication, naming, documentation) work identically in both CLI and WASM builds.

## Planned Features

The following features are under development and will be included in future releases:

### Statistical Trend Analysis
- ✅ **Linear regression** for trend lines across metric history (implemented)
- ✅ **Statistical hypothesis testing** for regression detection (implemented)
- ✅ **Confidence interval calculations** for forecast reliability (implemented)
- **ARIMA/exponential smoothing** for advanced time series forecasting (roadmap)
- **Correlation analysis** between different metrics (roadmap)

### Advanced Maintenance Detection
See [ROADMAP.md](ROADMAP.md) for detailed implementation plans for:
- Enhanced code duplication analysis with semantic similarity
- Naming convention analysis with automated suggestions
- Misplaced declaration detection (functions/methods in wrong files)
- Documentation gap detection and quality scoring
- Organizational health metrics (file size, package cohesion)
- Additional burden indicators (magic numbers, dead code, deep nesting)

### Configuration Enhancement
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