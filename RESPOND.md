# TASK DESCRIPTION:
Perform a data-driven roadmap item execution cycle: select ONE highest-priority incomplete task from ROADMAP.md, implement it completely, and validate measurable quality using `go-stats-generator` baseline/diff analysis to ensure no regressions and confirmed metric improvements.

## CONSTRAINT:

Use only `go-stats-generator` and existing tests for analysis and validation. Preserve all existing public APIs. Follow Go best practices: prefer stdlib, functions under 30 lines, explicit error handling with context, comprehensive GoDoc comments for any logic over 3 steps. If external libraries are needed, choose well-maintained options with >1000 GitHub stars and recent activity.

## PREREQUISITES:
**Minimum Required Version:** `go-stats-generator` v1.0.0 or higher
Install and configure `go-stats-generator` for comprehensive analysis and improvement tracking:

### Installation:
```bash
# First, check if go-stats-generator is already installed
which go-stats-generator
# If not, install it with `go install`
go install github.com/opd-ai/go-stats-generator@latest
```

## Recommendations:
```bash
# When long json outputs are encountered, use `jq`
go-stats-generator analyze --format json | jq .functions
# Check if it is installed
which jq
# If it is not, install it
sudo apt-get install jq
```

### Required Analysis Workflow:
```bash
# Phase 1: Establish baseline before any changes
go-stats-generator analyze . --format json --output baseline.json --skip-tests

# Phase 2: Select and implement roadmap task
# Using ROADMAP.md, select the highest-priority incomplete task and implement it.

# Phase 3: Post-implementation validation
go-stats-generator analyze . --format json --output post-impl.json --skip-tests
go-stats-generator diff baseline.json post-impl.json

# Phase 4: Quality verification
# Confirm no complexity regressions, all tests pass, metrics improved or stable.
```

## CONTEXT:
You are an automated Go development agent using `go-stats-generator` for measurable completion validation of roadmap tasks. The tool provides precise metrics across functions, structs, packages, interfaces, concurrency, duplication, naming, and documentation quality. Every implementation must be bookended by baseline capture and differential analysis to prove the change improved or maintained codebase health. Focus on selecting the single highest-impact incomplete task and delivering a production-ready implementation verified by quantitative analysis.

## INSTRUCTIONS:

### Phase 1: Baseline Capture
1. **Capture Pre-Implementation Metrics:**
  ```bash
  go-stats-generator analyze . --format json --output baseline.json --skip-tests
  go-stats-generator analyze . --skip-tests
  ```
  - Record current codebase metrics: function complexities, documentation coverage, duplication ratio
  - Note any existing threshold violations for comparison after implementation
  - Save baseline.json for differential analysis in Phase 3

2. **Review Current Quality State:**
  ```bash
  cat baseline.json | jq '{functions: (.functions | length), packages: (.packages | length), doc_coverage: .documentation.coverage.overall, duplication_ratio: .duplication.duplication_ratio}'
  ```
  - Establish the quantitative starting point for measuring task impact

### Phase 2: Roadmap Analysis and Task Implementation
1. **Select Highest-Priority Task:**
  - Analyze ROADMAP.md to identify incomplete tasks (marked "TODO", "In Progress", or similar)
  - Select the task that appears most critical or has explicit priority markers
  - If priorities are tied, prefer the task that affects the most files or has the broadest impact

2. **Implement the Selected Task:**
  - Prefer Go standard library over third-party packages when possible
  - Write functions under 30 lines with descriptive verb-first camelCase names
  - Handle all errors explicitly with context wrapping
  - Add GoDoc comments starting with function name for any exported or complex function
  - Include unit tests with table-driven test patterns where appropriate
  - Ensure all code follows `gofmt` standards and includes proper package documentation

3. **Update ROADMAP.md:**
  - Mark the selected task as completed with today's date
  - Note any sub-tasks or follow-up items discovered during implementation

### Phase 3: Post-Implementation Validation
1. **Capture Post-Implementation Metrics:**
  ```bash
  go-stats-generator analyze . --format json --output post-impl.json --skip-tests
  ```

2. **Run Differential Analysis:**
  ```bash
  go-stats-generator diff baseline.json post-impl.json
  ```
  - Verify no complexity regressions in existing functions
  - Confirm all new functions are below threshold limits
  - Check documentation coverage is maintained or improved

3. **Inspect New and Changed Functions:**
  ```bash
  cat post-impl.json | jq '.functions[] | select(.complexity.overall > 10) | {name, file, complexity: .complexity.overall, lines: .lines.code}'
  ```
  - Ensure no new function exceeds maximum complexity of 10
  - Ensure no new function exceeds 30 lines of code

### Phase 4: Quality Verification
1. **Validate Metrics Achievement:**
  - All new functions: complexity ≤ 10, line count ≤ 30
  - No complexity regressions detected by diff analysis
  - Documentation coverage ≥ 0.7
  - No new duplication introduced
  - Overall codebase quality trend stable or positive

2. **Confirm Functional Preservation:**
  - All tests pass: `go test ./...`
  - Race condition check: `go test -race ./...`
  - Build succeeds: `go build ./...`
  - Error handling paths unchanged in existing code
  - Return value semantics preserved

## OUTPUT FORMAT:

Structure your response as:

### 1. Baseline Metrics Summary
```
go-stats-generator baseline analysis:
  Total Functions: [n]
  Total Packages: [n]
  Documentation Coverage: [x.xx]
  Duplication Ratio: [x.xx]%
  Functions Exceeding Thresholds: [n]
```

### 2. Task Selected
```
Task: [task name from ROADMAP.md]
Priority: [priority level]
Rationale: [why this task was selected]
Files Affected: [list of files created or modified]
Libraries Used: [list with justification, or "stdlib only"]
```

### 3. Implementation Summary
```
Changes Made:
- [file1]: [description of change]
- [file2]: [description of change]
- ...

Tests Added/Modified:
- [test file]: [description]
```

### 4. Diff Results
```
Differential analysis (baseline → post-implementation):
  Functions Added: [n]
  Functions Modified: [n]
  Complexity Regressions: [count]
  Documentation Coverage: [old] → [new]
  Duplication Ratio: [old]% → [new]%
  New functions above thresholds: [count]
  Overall quality trend: [improved/stable/regressed]
```

## THRESHOLDS:
```
Implementation Quality Gates:
  Max Function Complexity  = 10
  Max Function Length       = 30 lines
  Min Documentation Coverage = 0.70 (70%)
  Max Duplication Ratio     = 5%
  Complexity Regressions    = 0 (no regressions allowed in diff)

Post-Implementation Checks:
  All tests pass:           go test ./...
  Race condition free:      go test -race ./...
  Build succeeds:           go build ./...
  gofmt compliant:          gofmt -l .
```
<!-- Last verified: 2025-07-25 against go-stats-generator analyze and diff commands -->

Task Complete Threshold = All tests pass AND complexity regressions = 0 AND doc coverage ≥ 0.7
- If no suitable task: "Roadmap execution complete: no incomplete tasks found in ROADMAP.md matching priority criteria."

## EXAMPLE WORKFLOW:
```bash
$ # Phase 1: Capture baseline
$ go-stats-generator analyze . --format json --output baseline.json --skip-tests
$ go-stats-generator analyze . --skip-tests
=== ANALYSIS SUMMARY ===
Total Functions: 142
Total Packages: 12
Documentation Coverage: 0.73
Duplication Ratio: 3.21%
Functions Exceeding Thresholds: 2

$ cat baseline.json | jq '{functions: (.functions | length), doc_coverage: .documentation.coverage.overall}'
{"functions": 142, "doc_coverage": 0.73}

$ # Phase 2: Select and implement roadmap task
$ # (Read ROADMAP.md, select highest-priority TODO item)
$ # Task selected: "Add CSV export support for analysis results"
$ # Implementation: created internal/reporter/csv.go, csv_test.go
$ # Updated ROADMAP.md to mark task complete

$ # Phase 3: Post-implementation validation
$ go-stats-generator analyze . --format json --output post-impl.json --skip-tests
$ go-stats-generator diff baseline.json post-impl.json
=== IMPROVEMENT SUMMARY ===
FUNCTIONS ADDED: 4
FUNCTIONS MODIFIED: 1

NEW FUNCTIONS:
  writeCSVReport: 6.2 complexity, 22 lines ✓
  formatCSVRow: 3.1 complexity, 15 lines ✓
  escapeCSVField: 2.0 complexity, 12 lines ✓
  newCSVReporter: 4.5 complexity, 18 lines ✓

MODIFIED FUNCTIONS:
  createReporter: 5.8 → 6.1 complexity (added CSV case) ✓

METRICS COMPARISON:
  Documentation Coverage: 0.73 → 0.75 (+0.02) ✓
  Duplication Ratio: 3.21% → 3.18% (-0.03%) ✓
  Complexity Regressions: 0 ✓
  All new functions below thresholds: ✓

QUALITY SCORE: 88/85 (+3 improvement)
REGRESSIONS: 0
ALL QUALITY GATES PASSED: ✓

$ # Phase 4: Final verification
$ go test ./...
ok  	github.com/opd-ai/go-stats-generator/internal/reporter	0.234s
ok  	github.com/opd-ai/go-stats-generator/...	1.456s
$ go test -race ./...
ok  	(all packages pass race detection)
$ go build ./...
$ echo "Task complete: CSV export support implemented and validated"
```

This data-driven approach ensures roadmap task implementation is validated by quantitative analysis rather than subjective assessment, with measurable confirmation that codebase quality is maintained or improved through every change.