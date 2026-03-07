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
- **Team Productivity Analysis**: Git-based metrics for understanding team contributions and code ownership
  - Per-developer commit statistics and contribution volumes
  - Code ownership analysis using `git blame` to identify primary file maintainers
  - Activity patterns and engagement metrics (active days, commit frequency)
  - Knowledge silo detection and expertise area identification
- **Test Coverage Correlation**: Analyze relationship between code complexity and test coverage
  - Identify high-risk functions (high complexity + low coverage)
  - Coverage gap detection for exported APIs
  - Risk scoring based on complexity, coverage, and size
  - Function-level and complexity-weighted coverage rates
- **Test Quality Assessment**: Evaluate test suite effectiveness and thoroughness
  - Test file structure and organization analysis
  - Assertion density metrics (assertions per test)
  - Subtest pattern detection
  - Test quality scoring and weak test identification

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
| `--format` | Output format (console, json, html, csv, markdown) | console |
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
| `--enable-team-metrics` | Enable team productivity analysis (requires Git repository) | false |
| `--coverage-profile` | Path to Go coverage profile for test coverage correlation and quality analysis | - |
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

### Team Productivity Analysis

The `--enable-team-metrics` flag enables Git-based analysis of team contributions and code ownership patterns. This feature requires the analyzed directory to be a Git repository.

**Requirements:**
- Git must be installed and available in PATH
- Directory must be a valid Git repository with commit history
- Analysis is read-only and does not modify the repository

**Usage:**

```bash
# Analyze with team metrics
go-stats-generator analyze . --enable-team-metrics

# Export team metrics to JSON for detailed analysis
go-stats-generator analyze . --enable-team-metrics --format json --output report.json

# Combine with other analysis options
go-stats-generator analyze . --enable-team-metrics --skip-tests
```

**Metrics Produced:**

The team metrics analysis provides per-developer statistics extracted from Git history:

| Metric | Description | Use Case |
|--------|-------------|----------|
| `commit_count` | Total commits by the developer | Activity level and contribution frequency |
| `lines_added` | Total lines added across all commits | Code contribution volume |
| `lines_removed` | Total lines removed/deleted | Refactoring activity and code cleanup |
| `files_modified` | Number of files where developer is primary owner (>50% lines) | Code ownership and expertise areas |
| `first_commit_date` | Timestamp of first commit | Tenure on the project |
| `last_commit_date` | Timestamp of most recent commit | Current activity status |
| `active_days` | Number of unique calendar days with commits | Consistency and engagement patterns |

**JSON Output Structure:**

```json
{
  "team": {
    "total_developers": 3,
    "developers": {
      "Alice Smith": {
        "name": "Alice Smith",
        "commit_count": 247,
        "lines_added": 18432,
        "lines_removed": 9821,
        "files_modified": 42,
        "first_commit_date": "2025-01-15T10:23:00Z",
        "last_commit_date": "2026-03-05T14:52:00Z",
        "active_days": 89
      },
      "Bob Johnson": {
        "name": "Bob Johnson",
        "commit_count": 156,
        "lines_added": 12109,
        "lines_removed": 8342,
        "files_modified": 28,
        "first_commit_date": "2025-02-01T09:15:00Z",
        "last_commit_date": "2026-03-06T16:20:00Z",
        "active_days": 62
      }
    }
  }
}
```

**Interpreting Results:**

- **Knowledge Silos**: Developers with high `files_modified` counts may indicate concentrated code ownership. Consider cross-training or pair programming to distribute knowledge.
- **Code Churn**: High `lines_removed` relative to `lines_added` may indicate refactoring work or exploratory development patterns.
- **Engagement Patterns**: Compare `active_days` to commit count to identify burst vs. consistent contribution patterns.
- **Onboarding Effectiveness**: Track `first_commit_date` alongside early contribution metrics to evaluate new developer ramp-up.

**Limitations:**

- File ownership is calculated using `git blame`, which attributes lines to the most recent modifier. Large refactorings may shift ownership attribution.
- Merge commits and automated commits (bots, CI systems) are included in metrics. Filter by author name if needed.
- Metrics reflect Git history only. Contributions like code reviews, design discussions, and documentation outside the codebase are not captured.

**Example Analysis Workflow:**

```bash
# 1. Generate team report
go-stats-generator analyze . --enable-team-metrics --format json --output team-report.json

# 2. Extract top contributors (using jq)
jq '.team.developers | to_entries | sort_by(-.value.commit_count) | .[0:5]' team-report.json

# 3. Identify potential knowledge silos (developers owning >20 files)
jq '.team.developers | to_entries | map(select(.value.files_modified > 20))' team-report.json

# 4. Track recent activity (commits in last 30 days)
# Compare last_commit_date against current date in your analysis scripts
```

### Test Coverage and Quality Analysis

The tool provides comprehensive test coverage correlation and test quality assessment features to help identify testing gaps and evaluate test suite effectiveness.

#### Test Coverage Correlation

Analyzes the relationship between code complexity and test coverage to identify high-risk untested code. This feature requires a Go coverage profile generated by `go test -coverprofile`.

**Requirements:**
- Go coverage profile file (generate with `go test -coverprofile=coverage.out ./...`)
- Access to coverage data during analysis

**Usage:**

```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./...

# Analyze with coverage correlation
go-stats-generator analyze . --coverage-profile coverage.out --format json --output report.json

# View coverage metrics
jq '.test_coverage' report.json
```

**Metrics Provided:**

| Metric | Description | Use Case |
|--------|-------------|----------|
| `function_coverage_rate` | Percentage of functions with >0% coverage | Overall coverage health |
| `complexity_coverage_rate` | Coverage weighted by cyclomatic complexity | Focus on complex code testing |
| `high_risk_functions` | Functions with high complexity + low coverage | Prioritize test writing efforts |
| `coverage_gaps` | Exported functions below 70% coverage | API testing completeness |

**High-Risk Functions:**

Functions are flagged as high-risk when:
- Cyclomatic complexity >5 AND coverage <50%, OR
- Cyclomatic complexity >10 AND coverage <80%

Each high-risk function includes a `risk_score` calculated as: `complexity × (1 - coverage) × size_multiplier`

**Coverage Gap Severity:**

| Severity | Criteria |
|----------|----------|
| `critical` | Coverage <30% AND complexity >5 |
| `high` | Coverage <50% OR complexity >8 |
| `medium` | Coverage <70% |
| `low` | Coverage ≥70% |

**Example JSON Output:**

```json
{
  "test_coverage": {
    "function_coverage_rate": 0.72,
    "complexity_coverage_rate": 0.65,
    "high_risk_functions": [
      {
        "name": "ProcessComplexData",
        "file": "internal/processor/data.go",
        "line": 45,
        "complexity": 15,
        "coverage": 0.32,
        "risk_score": 15.3
      }
    ],
    "coverage_gaps": [
      {
        "name": "PublicAPI",
        "file": "pkg/api/handler.go",
        "line": 120,
        "complexity": 8,
        "coverage": 0.45,
        "gap_severity": "high"
      }
    ]
  }
}
```

**Analysis Workflow:**

```bash
# 1. Generate coverage for all packages
go test -coverprofile=coverage.out -covermode=atomic ./...

# 2. Run analysis with coverage correlation
go-stats-generator analyze . --coverage-profile coverage.out --format json --output full-report.json

# 3. Extract high-risk functions (complexity >10, coverage <50%)
jq '.test_coverage.high_risk_functions | sort_by(.risk_score) | reverse | .[:10]' full-report.json

# 4. Find critical coverage gaps
jq '.test_coverage.coverage_gaps | map(select(.gap_severity == "critical"))' full-report.json

# 5. Calculate untested complexity (functions with 0% coverage and complexity >5)
jq '[.functions[] | select(.complexity.cyclomatic > 5)] | length as $total |
    [.test_coverage.high_risk_functions[] | select(.coverage == 0)] | 
    length as $untested | 
    {total_complex: $total, untested: $untested, risk_percentage: ($untested / $total * 100)}' full-report.json
```

#### Test Quality Assessment

Analyzes test suite structure and quality by examining test files for assertion density, test organization, and testing patterns. This feature runs automatically when `--coverage-profile` is provided.

**Metrics Provided:**

| Metric | Description | Interpretation |
|--------|-------------|----------------|
| `total_tests` | Total number of test functions across all test files | Test suite size |
| `avg_assertions_per_test` | Average assertion count per test function | Test thoroughness indicator |
| `test_files` | Per-file test statistics and assertion ratios | File-level test quality |

**Per-File Test Metrics:**

Each test file is analyzed for:
- `test_count`: Number of `Test*` functions in the file
- `subtest_count`: Number of `t.Run()` subtest calls
- `assertion_count`: Number of assertion/verification statements
- `assertion_ratio`: Assertions per test (higher is more thorough)

**What Counts as an Assertion:**

The analyzer detects common testing patterns:
- `testify` assertions: `assert.*`, `require.*`
- Standard library checks: `t.Error`, `t.Errorf`, `t.Fatal`, `t.Fatalf`
- Custom checks: `if !condition { t.Error() }` patterns
- Comparison checks: `reflect.DeepEqual`, `bytes.Equal`, etc.

**Example JSON Output:**

```json
{
  "test_quality": {
    "total_tests": 247,
    "avg_assertions_per_test": 3.8,
    "test_files": [
      {
        "file": "internal/analyzer/function_test.go",
        "test_count": 15,
        "subtest_count": 23,
        "assertion_count": 67,
        "assertion_ratio": 4.5
      },
      {
        "file": "pkg/generator/analyzer_test.go",
        "test_count": 8,
        "subtest_count": 12,
        "assertion_count": 18,
        "assertion_ratio": 2.3
      }
    ]
  }
}
```

**Quality Indicators:**

| Assertion Ratio | Interpretation | Action |
|----------------|----------------|--------|
| <1.0 | Very weak tests (minimal verification) | Add assertions to validate behavior |
| 1.0-2.0 | Basic tests (single outcome check) | Consider edge cases and error paths |
| 2.0-4.0 | Good tests (multiple verifications) | Maintain quality standard |
| >4.0 | Thorough tests (comprehensive verification) | Excellent - may indicate complex scenarios |

**Analysis Examples:**

```bash
# 1. Find test files with low assertion ratios (weak tests)
jq '.test_quality.test_files | map(select(.assertion_ratio < 2.0)) | 
    sort_by(.assertion_ratio)' report.json

# 2. Calculate test coverage percentage (test files vs source files)
jq '{
    source_files: .metadata.files_processed,
    test_files: (.test_quality.test_files | length),
    coverage_ratio: ((.test_quality.test_files | length) / .metadata.files_processed)
}' report.json

# 3. Identify files with many tests but few assertions (possible test quality issue)
jq '.test_quality.test_files | 
    map(select(.test_count > 5 and .assertion_ratio < 2.0))' report.json

# 4. Summary of test quality across the project
jq '{
    total_tests: .test_quality.total_tests,
    total_test_files: (.test_quality.test_files | length),
    avg_assertions: .test_quality.avg_assertions_per_test,
    weak_files: [.test_quality.test_files[] | select(.assertion_ratio < 2.0)] | length,
    strong_files: [.test_quality.test_files[] | select(.assertion_ratio >= 4.0)] | length
}' report.json
```

**Limitations:**

- Test quality analysis runs only when `--coverage-profile` is provided
- Assertion counting is heuristic-based and may miss custom assertion helpers
- Subtest detection relies on `t.Run()` pattern recognition
- Does not evaluate test correctness or semantic quality
- High assertion count doesn't guarantee good test design (may indicate over-specification)

**Best Practices:**

1. **Use with Coverage Data**: Test quality is most valuable when correlated with coverage metrics
2. **Focus on Ratio Trends**: Compare assertion ratios across similar code modules
3. **Investigate Outliers**: Files with very low or very high ratios may need review
4. **Combine with Complexity**: Prioritize adding assertions to tests covering complex functions
5. **Monitor Over Time**: Track average assertion ratio trends using baseline snapshots

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

## Output Formats

`go-stats-generator` supports multiple output formats to suit different use cases, from interactive console output to machine-readable formats for integration with other tools.

### Console Output (Default)

The default human-friendly format with colored text, tables, and progress indicators. Best for interactive terminal use and quick analysis.

```bash
go-stats-generator analyze .
# or explicitly
go-stats-generator analyze . --format console
```

**Features:**
- Color-coded severity indicators
- Formatted tables for easy reading
- Progress bars during analysis
- Summary statistics with emoji indicators

**Use Cases:**
- Interactive development and debugging
- Quick code reviews
- Manual analysis and exploration

### JSON Output

Structured JSON format for programmatic access and integration with other tools. All metrics are preserved with full precision.

```bash
go-stats-generator analyze . --format json --output report.json
```

**Features:**
- Complete metrics preservation
- Nested structure for complex data
- Type safety for programmatic parsing
- Suitable for version control and diffing

**Use Cases:**
- CI/CD integration and automation
- Custom tooling and analysis scripts
- Historical tracking and trending
- API responses and data exchange

**Example Structure:**
```json
{
  "metadata": {
    "repository": "/path/to/project",
    "generated_at": "2026-03-07T03:13:07Z",
    "analysis_time": "849.601886ms",
    "tool_version": "1.0.0"
  },
  "overview": {
    "total_lines": 14362,
    "total_functions": 650,
    "total_methods": 828
  },
  "functions": [...],
  "structs": [...],
  "packages": [...]
}
```

### HTML Output

Interactive HTML report with embedded CSS and JavaScript for rich visualization in web browsers.

```bash
go-stats-generator analyze . --format html --output report.html
```

**Features:**
- Sortable and filterable tables
- Interactive charts and graphs
- Hyperlinked navigation between sections
- Embedded styling (no external dependencies)

**Use Cases:**
- Shareable reports for team reviews
- Documentation and presentations
- Archival of analysis results
- Non-technical stakeholder communication

### CSV Output

Comma-separated values format optimized for spreadsheet applications and data analysis tools. Each section (functions, structs, interfaces, packages) is exported as a separate table with headers.

```bash
go-stats-generator analyze . --format csv --output report.csv
```

**Structure:**
The CSV output contains multiple sections, each prefixed with a comment header:
- `# METADATA` - Repository info, generation time, tool version
- `# OVERVIEW` - Summary statistics
- `# FUNCTIONS` - Detailed function metrics (one row per function)
- `# STRUCTS` - Struct analysis data
- `# INTERFACES` - Interface metrics
- `# PACKAGES` - Package-level statistics

**Function Columns:**
- Name, Package, File, Line, Is Exported, Is Method
- Lines Total, Lines Code, Lines Comments, Lines Blank
- Cyclomatic Complexity, Cognitive Complexity, Nesting Depth
- Overall Complexity, Parameter Count, Return Count
- Has Variadic, Returns Error, Has Documentation, Documentation Quality

**Use Cases:**
- Import into Excel, Google Sheets, or other spreadsheet tools
- Statistical analysis with R, Python pandas, or similar
- Data warehouse integration
- Custom reporting and charting
- Filtering and sorting large datasets

**Example:**
```csv
# METADATA
Repository,/path/to/project
Generated At,2026-03-07 03:13:07
Analysis Time,849.601886ms

# FUNCTIONS
Name,Package,File,Line,Cyclomatic Complexity,Lines Code
ProcessData,analyzer,analyzer.go,45,8,23
ValidateInput,validator,validator.go,12,4,15
```

**Tips:**
- Use `grep "^# "` to identify section boundaries
- Import with "comma" delimiter and "quote" text qualifier
- Filter rows by exporting specific sections only with `--sections` flag
- Combine with `--skip-tests` for production code analysis

### Markdown Output

GitHub-flavored Markdown format with tables, emoji indicators, and formatted sections. Perfect for README files, pull request comments, and documentation.

```bash
go-stats-generator analyze . --format markdown --output report.md
```

**Features:**
- GitHub-flavored Markdown tables
- Emoji indicators for visual appeal (📊, 🔧, 🏗️, 🔌, 📦, ⚡)
- Collapsible sections (when rendered on GitHub)
- Direct copy-paste into issues and PRs
- Renders beautifully on GitHub, GitLab, and documentation sites

**Structure:**
```markdown
# Go Code Analysis Report

## 📊 Overview
| Metric | Value |
|--------|-------|
| Repository | /path/to/project |
| Total Functions | 650 |

## 🔧 Functions
| Function | File | Lines | Complexity | Exported | Documentation |
|----------|------|-------|------------|----------|---------------|
| ProcessData | analyzer.go | 23 | 8 | ✅ | 85.2% |

## 🏗️ Structs
...
```

**Use Cases:**
- Adding reports to repository README files
- Pull request and code review comments
- GitHub Issues and discussion threads
- Static site documentation (Jekyll, Hugo, MkDocs)
- Markdown-based wikis and knowledge bases

**Tips:**
- Use `--sections` to include only relevant sections for focused reports
- Pipe to clipboard: `go-stats-generator analyze . --format markdown | pbcopy` (macOS) or `xclip` (Linux)
- Combine with baseline diffs for changelog generation
- Top 50 results are shown by default for readability (full data available in JSON/CSV)

### Choosing the Right Format

| Format | Interactive | Machine-Readable | Human-Readable | Shareable | Best For |
|--------|-------------|------------------|----------------|-----------|----------|
| **Console** | ✅ | ❌ | ✅ | ❌ | Development, debugging |
| **JSON** | ❌ | ✅ | ⚠️ | ✅ | Automation, APIs, CI/CD |
| **HTML** | ✅ | ❌ | ✅ | ✅ | Reports, presentations |
| **CSV** | ❌ | ✅ | ⚠️ | ✅ | Spreadsheets, data analysis |
| **Markdown** | ❌ | ⚠️ | ✅ | ✅ | Documentation, PRs, issues |

### Filtering Output Sections

All output formats support the `--sections` flag to include only specific parts of the analysis:

```bash
# Export only function metrics to CSV for analysis
go-stats-generator analyze . --format csv --sections functions --output functions.csv

# Create a focused Markdown report for documentation
go-stats-generator analyze . --format markdown --sections overview,documentation --output docs.md

# Generate JSON with specific sections for API integration
go-stats-generator analyze . --format json --sections functions,complexity,burden --output api-report.json
```

**Available Sections:**
- `metadata` - Repository info and generation metadata (always included)
- `overview` - High-level summary statistics (always included)
- `functions` - Function and method metrics
- `structs` - Struct analysis
- `interfaces` - Interface metrics
- `packages` - Package-level statistics
- `patterns` - Design and concurrency patterns
- `complexity` - Complexity analysis
- `documentation` - Documentation coverage
- `duplication` - Code duplication detection
- `naming` - Naming convention analysis
- `placement` - Code organization analysis
- `burden` - Maintenance burden indicators
- `scores` - Quality scores (MBI, etc.)
- `suggestions` - Refactoring suggestions

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