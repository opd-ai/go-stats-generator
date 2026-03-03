# TASK DESCRIPTION:
Execute autonomous iterative refactoring cycles on the codebase until no remaining functions exceed professional quality thresholds, then halt. Use `go-stats-generator` as the analysis engine: each iteration runs a baseline analysis, applies targeted improvements, and validates with differential analysis. The loop terminates when diff shows no remaining violations, max iterations are reached, or diff detects a regression.

## CONSTRAINT:

Use only `go-stats-generator` and existing tests for your analysis. You are absolutely forbidden from writing new code analysis tools or using any other code analysis tools. Each iteration must produce measurable improvement validated by `go-stats-generator diff`.

## PREREQUISITES:
**Minimum Required Version:** `go-stats-generator` v1.0.0 or higher
Install and configure `go-stats-generator` for iterative complexity analysis and improvement tracking:

### Installation:
```bash
# First, check if go-stats-generator is already installed
which go-stats-generator
# If not, install it with `go install`
go install github.com/opd-ai/go-stats-generator@latest
```

## Recommendations:
```bash
# Extract only task-relevant sections from JSON; discard everything else
go-stats-generator analyze --format json | jq '{functions: .functions}'
which jq || sudo apt-get install -y jq
```
**Section filter**: Use only `.functions` from the report. Exclude `.structs`, `.interfaces`, `.packages`, `.concurrency`, `.duplication`, `.naming`, `.documentation`, `.placement`, `.organization`, `.burden`, `.scores`, `.generics`, `.patterns`, `.suggestions` — they are not relevant to iterative complexity refactoring.

### Required Analysis Workflow:
```bash
# Phase 1: Establish initial baseline
go-stats-generator analyze . --max-complexity 10 --max-function-length 30 --skip-tests --format json --output iteration-0.json --sections functions
go-stats-generator analyze . --max-complexity 10 --max-function-length 30 --skip-tests

# Phase 2: Per-iteration cycle (repeat for each iteration N)
# 2a. Identify targets from current analysis
cat iteration-0.json | jq '[.functions[] | select(.complexity.overall > 10.0 or .lines.code > 30 or .complexity.cyclomatic > 10)]'

# 2b. Apply refactoring changes to highest-priority target(s)

# 2c. Post-refactoring validation
go-stats-generator analyze . --max-complexity 10 --max-function-length 30 --skip-tests --format json --output iteration-1.json --sections functions

# 2d. Measure iteration improvement
go-stats-generator diff iteration-0.json iteration-1.json

# Phase 3: Continue or terminate based on diff results
```

## CONTEXT:
You are an autonomous Go code auditor using `go-stats-generator` for iterative, data-driven refactoring. Each iteration identifies the highest-impact remaining target, applies a focused refactoring, and validates the improvement with differential analysis. The loop runs until all functions are below thresholds, a regression is detected, or the maximum iteration count is reached. Focus on measurable per-iteration progress rather than attempting all improvements at once.

## INSTRUCTIONS:

### Phase 1: Initial Baseline
1. **Establish Iteration-0 Baseline:**
  ```bash
  go-stats-generator analyze . --max-complexity 10 --max-function-length 30 --skip-tests --format json --output iteration-0.json
  go-stats-generator analyze . --max-complexity 10 --max-function-length 30 --skip-tests
  ```
  - Record all functions exceeding any threshold (see THRESHOLDS below)
  - Count total violations — this is the starting violation count

2. **Check Termination Before First Iteration:**
  ```bash
  cat iteration-0.json | jq '[.functions[] | select(.complexity.overall > 10.0 or .lines.code > 30 or .complexity.cyclomatic > 10)] | length'
  ```
  - If **0 violations**, output the no-targets message and halt immediately
  - Otherwise, proceed to the iteration loop

### Phase 2: Iteration Loop (max 5 iterations)

For each iteration N (1 through 5):

1. **Select Iteration Target:**
  ```bash
  cat iteration-$((N-1)).json | jq '[.functions[] | select(.complexity.overall > 10.0 or .lines.code > 30 or .complexity.cyclomatic > 10)] | sort_by(-.complexity.overall) | .[0]'
  ```
  - Choose the **single highest-complexity function** that still exceeds thresholds
  - If complexity scores are tied, choose the **longest function** first
  - Record its current metrics as the iteration's target baseline

2. **Apply Focused Refactoring:**
  - Refactor only the selected target function in this iteration
  - Use `go-stats-generator`'s analysis to identify logical extraction points
  - Extract helper functions following these rules:
    * Name functions using verb-first camelCase (e.g., `validateInput`, `calculateResult`)
    * Target < 20 lines and cyclomatic complexity < 8 per extracted function
    * Add GoDoc comments starting with function name
  - Preserve error handling, defer scopes, variable access patterns, and lock boundaries

3. **Validate Iteration Results:**
  ```bash
  go-stats-generator analyze . --max-complexity 10 --max-function-length 30 --skip-tests --format json --output iteration-$N.json
  go-stats-generator diff iteration-$((N-1)).json iteration-$N.json
  ```

4. **Evaluate Continuation Criteria:**
  - **CONTINUE** if ALL of the following are true:
    * `go-stats-generator diff` shows the target function's complexity decreased
    * No regressions detected (no previously-passing function now exceeds thresholds)
    * Remaining violation count > 0
    * Current iteration N < 5
  - **TERMINATE** if ANY of the following are true:
    * No functions exceed thresholds (all violations resolved) → **Success**
    * `go-stats-generator diff` shows a regression in any metric → **Regression halt**
    * Target function's complexity did not decrease → **No progress halt**
    * Iteration N = 5 → **Max iterations reached**

5. **Run Tests:**
  ```bash
  go test ./...
  ```
  - If tests fail, revert the iteration's changes and terminate with failure report

### Phase 3: Final Summary
After the loop terminates:
```bash
go-stats-generator diff iteration-0.json iteration-$N.json
go-stats-generator diff iteration-0.json iteration-$N.json --format html --output loop-report.html
```

## OUTPUT FORMAT:

Structure your response as:

### Per-Iteration Output (repeat for each iteration):
```
ITERATION [N]:

TARGET: [function_name] in [file]
  Before: complexity=[score] lines=[count] cyclomatic=[count]

CHANGES MADE:
  - [Specific refactoring action] → extracted [function_name]
  - [Specific refactoring action] → extracted [function_name]

VALIDATION (go-stats-generator diff iteration-[N-1].json iteration-[N].json):
  Target: [old_score] → [new_score] ([improvement_%] reduction)
  Extracted functions: [list with complexities]
  Regressions: [count]
  Remaining violations: [count]

CONTINUE? [YES/NO]
  Reason: [Brief explanation based on continuation criteria]
```

### Final Summary Output:
```
LOOP COMPLETE

Termination reason: [all violations resolved / max iterations / regression / no progress]
Total iterations: [N]

Overall improvement (go-stats-generator diff iteration-0.json iteration-[N].json):
  Functions refactored: [count]
  Violations resolved: [start_count] → [end_count]
  Regressions: [count]
  Quality score change: [before] → [after] ([delta])

Per-iteration summary:
  Iteration 1: [function] [old] → [new] ([%] reduction) ✓/✗
  Iteration 2: [function] [old] → [new] ([%] reduction) ✓/✗
  ...
```

## THRESHOLDS:
```
Iteration Trigger (any function exceeding ANY of these):
  Overall Complexity > 10.0
  Function Length > 30 lines (code lines only)
  Cyclomatic Complexity > 10

Per-Extracted-Function Targets:
  Lines < 20
  Cyclomatic Complexity < 8
  GoDoc comment present

Continuation Criteria:
  Target complexity decreased (validated by go-stats-generator diff)
  Zero regressions in unchanged functions
  Remaining violations > 0
  Iteration count < 5

Termination Criteria (any one triggers halt):
  Remaining violations = 0           → Success
  Regression detected by diff        → Regression halt
  Target complexity unchanged        → No progress halt
  Iteration count = 5                → Max iterations reached
  Tests fail after refactoring       → Failure halt
```

Refactoring Threshold = Overall Complexity > 10.0 OR Lines > 30 OR Cyclomatic > 10
- If no targets: "Loop complete: go-stats-generator baseline analysis found no functions exceeding professional quality thresholds."

## COMPLEXITY REFERENCE (go-stats-generator calculation):
```
Overall Complexity = cyclomatic + (nesting_depth * 0.5) + (cognitive * 0.3)
Signature Complexity = (params * 0.5) + (returns * 0.3) + (interfaces * 0.8) + (generics * 1.5) + variadic_penalty
Refactoring Threshold = Overall Complexity > 10.0 OR Lines > 30 OR Cyclomatic > 10
```
<!-- Last verified: 2025-07-25 against function.go:calculateComplexity and calculateSignatureComplexity -->

## EXAMPLE WORKFLOW:
```bash
$ go-stats-generator analyze . --max-complexity 10 --max-function-length 30 --skip-tests
=== TOP COMPLEX FUNCTIONS ===
1. processComplexOrder (order.go): 25.4 complexity [Critical]
   - Lines: 45 code lines
   - Cyclomatic: 18
   - Nesting: 4
   - Recommendations: Extract 4 logical blocks

2. handleDataTransform (transform.go): 18.7 complexity [High]
   - Lines: 52 code lines
   - Cyclomatic: 14
   - Nesting: 3
   - Recommendations: Extract 3 logical blocks

3. validateComplexInput (validator.go): 15.2 complexity [High]
   - Lines: 38 code lines
   - Cyclomatic: 11
   - Nesting: 3
   - Recommendations: Extract 2 logical blocks

$ go-stats-generator analyze . --max-complexity 10 --max-function-length 30 --skip-tests --format json --output iteration-0.json
$ cat iteration-0.json | jq '[.functions[] | select(.complexity.overall > 10.0 or .lines.code > 30 or .complexity.cyclomatic > 10)] | length'
3

$ # ITERATION 1: Target processComplexOrder (highest complexity)
$ # ... refactor processComplexOrder → extract 4 helper functions ...

$ go-stats-generator analyze . --max-complexity 10 --max-function-length 30 --skip-tests --format json --output iteration-1.json
$ go-stats-generator diff iteration-0.json iteration-1.json
=== IMPROVEMENT SUMMARY ===
FUNCTIONS REFACTORED: 1

MAJOR IMPROVEMENTS:
- processComplexOrder: 25.4 → 7.2 (72% reduction) ✓

EXTRACTED FUNCTIONS:
  validateOrderData: 5.1 complexity ✓
  calculatePricing: 7.3 complexity ✓
  applyDiscounts: 4.8 complexity ✓
  finalizeOrder: 6.8 complexity ✓

REGRESSIONS: 0
REMAINING VIOLATIONS: 2

$ go test ./...
ok  	./... (all tests pass)

$ # CONTINUE: target improved, 0 regressions, 2 violations remain, iteration 1 < 5

$ # ITERATION 2: Target handleDataTransform (next highest)
$ # ... refactor handleDataTransform → extract 3 helper functions ...

$ go-stats-generator analyze . --max-complexity 10 --max-function-length 30 --skip-tests --format json --output iteration-2.json
$ go-stats-generator diff iteration-1.json iteration-2.json
=== IMPROVEMENT SUMMARY ===
FUNCTIONS REFACTORED: 1

MAJOR IMPROVEMENTS:
- handleDataTransform: 18.7 → 6.8 (64% reduction) ✓

EXTRACTED FUNCTIONS:
  transformInput: 4.2 complexity ✓
  mapOutputFields: 5.5 complexity ✓
  applyTransformRules: 6.1 complexity ✓

REGRESSIONS: 0
REMAINING VIOLATIONS: 1

$ go test ./...
ok  	./... (all tests pass)

$ # CONTINUE: target improved, 0 regressions, 1 violation remains, iteration 2 < 5

$ # ITERATION 3: Target validateComplexInput (last violation)
$ # ... refactor validateComplexInput → extract 2 helper functions ...

$ go-stats-generator analyze . --max-complexity 10 --max-function-length 30 --skip-tests --format json --output iteration-3.json
$ go-stats-generator diff iteration-2.json iteration-3.json
=== IMPROVEMENT SUMMARY ===
FUNCTIONS REFACTORED: 1

MAJOR IMPROVEMENTS:
- validateComplexInput: 15.2 → 5.9 (61% reduction) ✓

EXTRACTED FUNCTIONS:
  checkRequiredFields: 3.8 complexity ✓
  validateFieldRanges: 4.5 complexity ✓

REGRESSIONS: 0
REMAINING VIOLATIONS: 0

$ go test ./...
ok  	./... (all tests pass)

$ # TERMINATE: 0 violations remain → Success

$ go-stats-generator diff iteration-0.json iteration-3.json
=== OVERALL IMPROVEMENT SUMMARY ===
TOTAL ITERATIONS: 3
FUNCTIONS REFACTORED: 3

CUMULATIVE IMPROVEMENTS:
- processComplexOrder: 25.4 → 7.2 (72% reduction) ✓
- handleDataTransform: 18.7 → 6.8 (64% reduction) ✓
- validateComplexInput: 15.2 → 5.9 (61% reduction) ✓

ALL EXTRACTED FUNCTIONS BELOW THRESHOLDS: ✓
QUALITY SCORE: 95/67 (+28 improvement)
REGRESSIONS: 0
ALL FUNCTIONS NOW BELOW THRESHOLDS: ✓

$ go-stats-generator diff iteration-0.json iteration-3.json --format html --output loop-report.html
```

This data-driven iterative approach ensures each refactoring cycle produces measurable, validated improvement. The loop self-terminates on success, regression, or iteration limits, with every decision driven by `go-stats-generator` differential analysis rather than subjective assessment.
