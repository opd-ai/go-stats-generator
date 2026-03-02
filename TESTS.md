# TASK DESCRIPTION:
Perform data-driven test coverage prioritization using `go-stats-generator` complexity ranking to discover untested Go functions and generate comprehensive unit tests achieving 80%+ coverage, targeting the highest-complexity untested functions first for maximum testing value.

When results are ambiguous, such as a tie between complexity scores or if multiple untested functions have similar metrics, always choose the **highest cyclomatic complexity** function first. Testing the most complex code first yields the greatest defect-prevention value.

## CONSTRAINT:

Use only `go-stats-generator` for analysis and the standard `testing` package for generated tests. You are absolutely forbidden from using any other code analysis tools. Generated tests must use only the standard library `testing` package and `github.com/stretchr/testify` where already present in the project.

## PREREQUISITES:
**Minimum Required Version:** `go-stats-generator` v1.0.0 or higher
Install and configure `go-stats-generator` for comprehensive complexity analysis and test coverage prioritization:

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
go-stats-generator analyze --format json | jq '.functions[]'
# Check if it is installed
which jq
# If it is not, install it
sudo apt-get install jq
```

### Required Analysis Workflow:
```bash
# Phase 1: Establish baseline and identify untested high-complexity targets
go-stats-generator analyze . --skip-tests --format json --output test-baseline.json
go test -cover ./...

# Phase 2: Prioritize test targets by complexity ranking
cat test-baseline.json | jq '[.functions[] | select(.cyclomatic > 8)] | sort_by(-.cyclomatic)'

# Phase 3: Post-test-generation validation
go-stats-generator analyze . --format json --output post-tests.json
go-stats-generator diff test-baseline.json post-tests.json

# Phase 4: Coverage verification
go test -cover -coverprofile=coverage.out ./...
```

## CONTEXT:
You are a test generation agent using `go-stats-generator` to rank untested files by complexity for highest-value test targets. The tool provides precise per-function complexity metrics, enabling data-driven prioritization of which functions to test first. Focus on functions with the highest cyclomatic complexity that currently lack test coverage, as these represent the greatest risk and highest return on testing investment.

## INSTRUCTIONS:

### Phase 1: Baseline Analysis
1. **Run Complexity Baseline:**
  ```bash
  go-stats-generator analyze . --skip-tests --format json --output test-baseline.json
  ```
  - Record per-function complexity metrics from `.functions[]`
  - Note cyclomatic complexity, line count, and nesting depth for each function
  - Identify the files and packages containing high-complexity functions

2. **Capture Current Coverage:**
  ```bash
  go test -cover ./...
  ```
  - Record per-package coverage percentages
  - Identify packages with 0% or low coverage
  - Cross-reference with complexity data to find high-complexity untested code

3. **Extract High-Complexity Untested Functions:**
  ```bash
  cat test-baseline.json | jq '[.functions[] | select(.cyclomatic > 8)] | sort_by(-.cyclomatic) | .[] | {name, file, cyclomatic, lines, complexity}'
  ```
  - Build a ranked list of untested functions sorted by descending cyclomatic complexity
  - Cross-reference against existing `_test.go` files to confirm which functions lack tests

### Phase 2: Prioritize Test Targets
1. **Rank Functions by Complexity:**
  From the baseline analysis, select functions for test generation in priority order:
  - **Critical (Cyclomatic > 15):** Test immediately — highest defect risk
  - **High (Cyclomatic 9-15):** Test in second pass
  - **Medium (Cyclomatic 5-8):** Test if time permits

  When prioritizing between functions:
  - If cyclomatic scores are tied, choose the **longest function** first
  - If both cyclomatic and line count are tied, choose the function with **deeper nesting** first
  - Target at least 5 functions; extend to 10 if more than 5 functions exceed Critical or High priority thresholds

2. **Plan Test Strategy Per Function:**
  ```bash
  cat test-baseline.json | jq '.functions[] | select(.name == "targetFunction") | {name, file, cyclomatic, lines, params, returns}'
  ```
  - Functions with >2 input parameters → table-driven tests
  - Functions returning `error` → include error condition tests
  - Functions with high nesting → test each branch path
  - Functions using interfaces → test with mock implementations

### Phase 3: Generate and Validate Tests
1. **Generate Tests for Each Target:**
  - Write table-driven tests for functions with >2 input scenarios
  - Include edge cases: nil inputs, empty strings, boundary values, error conditions
  - Follow Go naming conventions: `TestFunctionName_Scenario_Expected`
  - Use subtests with `t.Run()` for each table entry
  - Test all exported functions and methods in the target file

2. **Validate Generated Tests:**
  ```bash
  go test ./... -v
  go test -race ./...
  ```
  - All generated tests must compile and pass
  - No race conditions detected
  - No test interdependencies or side effects

3. **Run Post-Test Analysis:**
  ```bash
  go-stats-generator analyze . --format json --output post-tests.json
  go-stats-generator diff test-baseline.json post-tests.json
  ```
  - Verify no complexity regressions in production code
  - Confirm test files do not introduce new complexity issues

### Phase 4: Coverage Verification
1. **Validate Coverage Achievement:**
  ```bash
  go test -cover -coverprofile=coverage.out ./...
  ```
  - 80%+ line coverage achieved on targeted packages
  - All generated tests pass consistently
  - No complexity regressions detected by diff analysis

2. **Confirm Quality Gates:**
  - All generated tests pass: `go test ./...`
  - Race condition check: `go test -race ./...`
  - Error handling paths tested for all functions returning errors
  - Table-driven tests used for functions with >2 input scenarios
  - Build succeeds: `go build ./...`

## OUTPUT FORMAT:

Structure your response as:

### 1. Complexity-Ranked Target List
```
go-stats-generator identified high-complexity untested functions:
1. Function: [name] in [file]
   - Cyclomatic complexity: [score]
   - Lines: [n] code lines
   - Nesting depth: [n]
   - Current coverage: [0%/none]
   - Priority: [Critical/High/Medium]
   - Test strategy: [table-driven/edge-case/branch-path]

2. Function: [name] in [file]
   - Cyclomatic complexity: [score]
   - Lines: [n] code lines
   - Nesting depth: [n]
   - Current coverage: [0%/none]
   - Priority: [Critical/High/Medium]
   - Test strategy: [table-driven/edge-case/branch-path]

... (continue for top 5-10 functions)
```

### 2. Generated Test Files
Present each generated test file with:
- Table-driven tests for multi-input functions
- Edge case and error condition coverage
- Standard Go formatting and naming conventions
- GoDoc comments explaining test purpose

### 3. Coverage Results and Diff Validation
```
Coverage results:
- Package [name]: [old_%] → [new_%] ([improvement_%])
- Package [name]: [old_%] → [new_%] ([improvement_%])
- Overall: [old_%] → [new_%]

Differential analysis (go-stats-generator diff):
- Complexity regressions: [count]
- New functions tested: [count]
- Coverage target met: [yes/no]
```

## THRESHOLDS:
```
Test Prioritization:
  Critical = Cyclomatic complexity > 15
  High     = Cyclomatic complexity 9-15
  Medium   = Cyclomatic complexity 5-8

Coverage Targets:
  Minimum Line Coverage = 80%
  Prioritize functions with cyclomatic > 8
  Test highest-complexity files first

Test Generation Rules:
  Table-driven tests for functions with > 2 input scenarios
  Error condition tests for all functions returning error
  Branch-path tests for functions with nesting depth > 3
  Subtests with t.Run() for each table entry

Post-Generation Quality Gates:
  All generated tests pass: go test ./...
  No race conditions: go test -race ./...
  No complexity regressions: go-stats-generator diff
  Build succeeds: go build ./...
```
<!-- Last verified: 2025-07-25 against function.go:calculateComplexity and go test -cover output format -->

Coverage Threshold = 80% line coverage OR all Critical/High functions tested
- If no targets: "Test generation complete: go-stats-generator baseline analysis found no untested functions exceeding complexity thresholds."

## EXAMPLE WORKFLOW:
```bash
$ go-stats-generator analyze . --skip-tests --format json --output test-baseline.json
$ cat test-baseline.json | jq '[.functions[] | select(.cyclomatic > 8)] | sort_by(-.cyclomatic) | .[:5][] | {name, file, cyclomatic, lines}'
{
  "name": "processComplexOrder",
  "file": "internal/order/process.go",
  "cyclomatic": 18,
  "lines": 45
}
{
  "name": "handleDataTransform",
  "file": "internal/transform/handler.go",
  "cyclomatic": 14,
  "lines": 52
}
{
  "name": "validateComplexInput",
  "file": "internal/validator/input.go",
  "cyclomatic": 11,
  "lines": 38
}
{
  "name": "generateReport",
  "file": "internal/reporter/report.go",
  "cyclomatic": 9,
  "lines": 41
}
{
  "name": "parseConfiguration",
  "file": "internal/config/parser.go",
  "cyclomatic": 9,
  "lines": 35
}

$ go test -cover ./...
ok  internal/order     0.3s  coverage: 12.4% of statements
ok  internal/transform 0.2s  coverage: 0.0% of statements
ok  internal/validator 0.1s  coverage: 45.2% of statements
ok  internal/reporter  0.2s  coverage: 23.1% of statements
ok  internal/config    0.1s  coverage: 67.8% of statements

$ # Generate tests for each target in priority order...
$ # processComplexOrder (cyclomatic 18, Critical) → table-driven tests
$ # handleDataTransform (cyclomatic 14, High) → branch-path tests
$ # validateComplexInput (cyclomatic 11, High) → edge-case tests
$ # generateReport (cyclomatic 9, High) → table-driven tests
$ # parseConfiguration (cyclomatic 9, High) → error-condition tests

$ go test ./... -v
--- PASS: TestProcessComplexOrder_ValidOrder_Success
--- PASS: TestProcessComplexOrder_InvalidInput_Error
--- PASS: TestProcessComplexOrder_EmptyOrder_Error
--- PASS: TestHandleDataTransform_StandardInput_Transformed
--- PASS: TestHandleDataTransform_NilInput_Error
--- PASS: TestValidateComplexInput_BoundaryValues_Validated
... (all tests pass)

$ go-stats-generator analyze . --format json --output post-tests.json
$ go-stats-generator diff test-baseline.json post-tests.json
=== IMPROVEMENT SUMMARY ===
FUNCTIONS TESTED: 5

COVERAGE IMPROVEMENTS:
- internal/order:     12.4% → 87.3% (+74.9%) ✓
- internal/transform:  0.0% → 82.1% (+82.1%) ✓
- internal/validator: 45.2% → 91.4% (+46.2%) ✓
- internal/reporter:  23.1% → 85.7% (+62.6%) ✓
- internal/config:    67.8% → 93.2% (+25.4%) ✓

COMPLEXITY REGRESSIONS: 0
ALL TARGETED PACKAGES ABOVE 80% COVERAGE: ✓

$ go test -cover ./...
ok  internal/order     0.4s  coverage: 87.3% of statements
ok  internal/transform 0.3s  coverage: 82.1% of statements
ok  internal/validator 0.2s  coverage: 91.4% of statements
ok  internal/reporter  0.3s  coverage: 85.7% of statements
ok  internal/config    0.2s  coverage: 93.2% of statements
```

This data-driven approach ensures test generation decisions are based on quantitative complexity analysis rather than subjective assessment, with measurable validation of coverage improvements targeting the highest-risk untested functions first.