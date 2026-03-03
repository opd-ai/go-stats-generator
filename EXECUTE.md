# TASK DESCRIPTION:
Execute the next planned task from the project backlog (AUDIT.md → PLAN.md → ROADMAP.md, in strict priority order) with `go-stats-generator` baseline analysis before changes, implementation following Go best practices, and differential validation to ensure zero regressions while delivering measurable progress toward a finished product.

When results are ambiguous, such as multiple tasks appearing equally important, always choose the task listed first in its source file.

## CONSTRAINT:

Use `go-stats-generator` to establish a metrics baseline before any code changes and to validate after changes that no complexity, duplication, or quality regressions were introduced. You must execute tasks in strict file priority order: AUDIT.md first, then PLAN.md, then ROADMAP.md. Tasks completed out-of-order will be summarily rejected without consideration.

Do not second guess the order of the files. The humans put them in that order deliberately. If you think these files are misplaced or mislabeled, you're wrong. Go fuck yourself, then do the thing in the order the humans told you to.

## PREREQUISITES:
**Minimum Required Version:** `go-stats-generator` v1.0.0 or higher
Install and configure `go-stats-generator` for pre/post implementation validation:

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
go-stats-generator analyze --format json | jq '{functions: .functions, duplication: .duplication, documentation: .documentation}'
which jq || sudo apt-get install -y jq
```
**Section filter**: Use only `.functions`, `.duplication`, and `.documentation` from the report. Exclude `.structs`, `.interfaces`, `.packages`, `.patterns`, `.concurrency`, `.complexity`, `.generics`, `.naming`, `.placement`, `.organization`, `.burden`, `.scores`, `.suggestions` — they are not relevant to backlog task execution.

### Required Analysis Workflow:
```bash
# Phase 1: Establish baseline before any code changes
go-stats-generator analyze . --max-function-length 30 --max-complexity 10 --min-doc-coverage 0.7 --skip-tests --format json --output baseline.json --sections functions,duplication,documentation
go-stats-generator analyze . --max-function-length 30 --max-complexity 10 --min-doc-coverage 0.7 --skip-tests

# Phase 2: Identify task, implement changes, run tests
# (see INSTRUCTIONS below)

# Phase 3: Post-implementation validation
go-stats-generator analyze . --max-function-length 30 --max-complexity 10 --min-doc-coverage 0.7 --skip-tests --format json --output post-change.json --sections functions,duplication,documentation

# Phase 4: Measure and confirm zero regressions
go-stats-generator diff baseline.json post-change.json
go-stats-generator diff baseline.json post-change.json --format json --output diff-report.json
```

## CONTEXT:
You are an autonomous Go developer using `go-stats-generator` as the primary quality gate for every code change. The tool provides precise function-level complexity, duplication, naming, and documentation metrics. You establish a quantitative baseline before changes, implement the highest-priority backlog task, and use differential analysis to prove zero regressions. You execute tasks from AUDIT.md first (immediate action), then PLAN.md (medium-term), then ROADMAP.md (long-term) — never out of order. ALWAYS choose items from AUDIT.md, if it exists, no matter what. Optional tasks from AUDIT.md take priority over any tasks from any other file. No skipping anything in AUDIT.md. Perform optional tasks. If AUDIT.md or PLAN.md appears complete, verify it is actually complete. If it is actually complete, delete the file.

## INSTRUCTIONS:

### Phase 1: Baseline Capture
1. **Establish Pre-Change Metrics:**
  ```bash
  go-stats-generator analyze . --max-function-length 30 --max-complexity 10 --skip-tests --format json --output baseline.json
  go-stats-generator analyze . --max-function-length 30 --max-complexity 10 --skip-tests
  ```
  - Record current violation counts (functions over length, over complexity, under doc coverage)
  - Note existing duplication ratio and clone pair count
  - This baseline is the regression boundary — no metric may worsen after changes

### Phase 2: Task Selection
1. **Identify the Next Task (Strict Priority Order):**
  - **First:** Read AUDIT.md. If any incomplete tasks exist, select from here. Do not proceed to other files.
  - **Second:** Only if AUDIT.md does not exist or is fully complete, read PLAN.md. Select from here.
  - **Third:** Only if both AUDIT.md and PLAN.md are absent or complete, read ROADMAP.md. Select from here.
  - Never execute a PLAN.md task before an AUDIT.md task. Never execute a ROADMAP.md task before a PLAN.md task. Prioritize important tasks before resorting to trivial ones.

2. **Task Grouping (Optional):**
  You may execute 1–3 tasks together if they meet ALL of these criteria:
  - **Same Component**: Tasks affect the same file, function, or module
  - **Shared Context**: Changes require understanding the same code area
  - **Dependent Changes**: Completing one task naturally leads to or enables the next
  - **Similar Scope**: Each task is small and together they stay under 500 lines of changes
  - **Common Testing**: The tasks can be validated with a shared test suite

  **Examples of related tasks (group together):**
  - Adding 2–3 similar struct fields and their getter/setter methods
  - Implementing multiple related interface methods for the same type
  - Fixing multiple validation issues in the same function

  **Examples of unrelated tasks (execute separately):**
  - Changes to different packages or unrelated modules
  - Tasks requiring different testing strategies
  - Large refactoring combined with new feature development

### Phase 3: Implementation
1. **Design:**
  - Document your approach and library choices in comments before coding
  - For grouped tasks, ensure they share a coherent design

2. **Code Standards:**
  - Use standard library first, then well-maintained libraries (>1000 GitHub stars, updated within 6 months)
  - Keep functions under 30 lines with single responsibility
  - Handle all errors explicitly — no ignored error returns
  - Write self-documenting code with descriptive names over abbreviations
  - Name functions using verb-first camelCase (e.g., `validateInput`, `calculateResult`)
  - Add GoDoc comments for all exported functions

3. **Implementation:**
  - Write the minimal viable solution using existing libraries where possible
  - For grouped tasks, implement them in logical order
  - **SIMPLICITY RULE**: If your solution requires more than 3 levels of abstraction or clever patterns, redesign it for clarity. Choose boring, maintainable solutions over elegant complexity.

4. **Testing:**
  ```bash
  go test ./... -race
  ```
  - Create unit tests with >80% coverage for business logic
  - Include error case testing
  - For grouped tasks, ensure comprehensive test coverage across all changes

### Phase 4: Differential Validation
1. **Capture Post-Change Metrics:**
  ```bash
  go-stats-generator analyze . --max-function-length 30 --max-complexity 10 --skip-tests --format json --output post-change.json
  ```

2. **Measure Regressions:**
  ```bash
  go-stats-generator diff baseline.json post-change.json
  ```
  - Verify zero complexity regressions in unchanged code
  - Confirm all new functions are under 30 lines
  - Confirm all new functions have complexity ≤ 10
  - Confirm no new duplication introduced
  - Verify documentation coverage did not decrease

3. **Generate Validation Report:**
  ```bash
  go-stats-generator diff baseline.json post-change.json --format json --output diff-report.json
  cat diff-report.json | jq '{regressions: .regressions, new_violations: .new_violations, quality_delta: .quality_delta}'
  ```

### Phase 5: Documentation and Reporting
1. **Update Backlog:**
  - Mark completed tasks in AUDIT.md, PLAN.md, or ROADMAP.md
  - If AUDIT.md or PLAN.md is fully complete, delete the file
  - Add GoDoc comments for exported functions and update README if needed

2. **Confirm Functional Preservation:**
  - All tests pass: `go test ./...`
  - Race condition check: `go test -race ./...`
  - Build succeeds: `go build ./...`

## OUTPUT FORMAT:

Structure your response as:

### 1. Task Selection
```
Source file: [AUDIT.md | PLAN.md | ROADMAP.md]
Task(s) selected:
  1. [Task description from source file]
  2. [Optional grouped task, if criteria met]
Grouping justification: [why these tasks are related, or "single task"]
```

### 2. Baseline Summary
```
go-stats-generator pre-change baseline:
  Functions over 30 lines: [n]
  Functions over complexity 10: [n]
  Duplication ratio: [x.xx]%
  Documentation coverage: [x.xx]%
  Total violations: [n]
```

### 3. Implementation Summary
Present what was implemented:
- Files changed and why
- Design decisions and rationale
- Test coverage achieved

### 4. Validation Results
```
Differential analysis (baseline → post-change):
  Complexity regressions: [count] (must be 0)
  New functions over 30 lines: [count] (must be 0)
  New functions over complexity 10: [count] (must be 0)
  Duplication ratio: [old_%] → [new_%]
  Documentation coverage: [old_%] → [new_%]
  Tests passing: [yes/no]
  Race conditions: [none detected / details]
  Overall quality delta: [+n / -n / unchanged]
```

### 5. Backlog Update
```
Updated: [file name]
Completed: [task description(s)]
Remaining tasks in file: [n]
File deleted: [yes/no]
```

## THRESHOLDS:
```
Pre/Post Validation (go-stats-generator enforce):
  Max Function Length  = 30 lines (--max-function-length 30)
  Max Complexity       = 10 (--max-complexity 10)
  Min Doc Coverage     = 0.7 (--min-doc-coverage 0.7)
  Similarity Threshold = 0.80 (--similarity-threshold 0.80)
  Min Block Lines      = 6 (--min-block-lines 6)

Regression Gate (zero tolerance):
  Complexity regressions in unchanged code = 0
  New functions exceeding length threshold = 0
  New functions exceeding complexity threshold = 0
  Duplication ratio increase = 0%

Implementation Quality:
  Test coverage for business logic ≥ 80%
  All errors explicitly handled
  All exported functions have GoDoc comments
  Functions have single responsibility

Task Grouping Limits:
  Max grouped tasks = 3
  Max combined change size = 500 lines
```
<!-- Last verified: 2025-07-25 against go-stats-generator CLI flags and default thresholds -->

Execution Threshold = Any incomplete task in AUDIT.md, then PLAN.md, then ROADMAP.md
- If no targets: "Execution complete: all backlog files (AUDIT.md, PLAN.md, ROADMAP.md) are fully resolved. No tasks remain."

## EXAMPLE WORKFLOW:
```bash
$ # Phase 1: Baseline before any changes
$ go-stats-generator analyze . --max-function-length 30 --max-complexity 10 --skip-tests
=== ANALYSIS SUMMARY ===
Functions analyzed: 142
Functions over 30 lines: 3
Functions over complexity 10: 2
Duplication ratio: 3.21%
Documentation coverage: 78.4%

$ go-stats-generator analyze . --max-function-length 30 --max-complexity 10 --skip-tests --format json --output baseline.json

$ # Phase 2: Select task from AUDIT.md (highest priority)
$ head -20 AUDIT.md
# Audit Findings
- [ ] Fix error handling in internal/scanner/worker.go:processFile
- [ ] Add missing GoDoc for exported functions in pkg/metrics/
- [x] Resolve circular import in internal/analyzer/

$ # Task selected: "Fix error handling in internal/scanner/worker.go:processFile"

$ # Phase 3: Implement the fix
$ # ... edit internal/scanner/worker.go ...
$ go test ./internal/scanner/... -race
ok  	github.com/example/project/internal/scanner	0.342s

$ # Phase 4: Validate zero regressions
$ go-stats-generator analyze . --max-function-length 30 --max-complexity 10 --skip-tests --format json --output post-change.json
$ go-stats-generator diff baseline.json post-change.json
=== DIFFERENTIAL ANALYSIS ===
CHANGED FUNCTIONS: 1

IMPROVEMENTS:
  processFile (worker.go):
    - Complexity: 12.4 → 8.1 (35% reduction) ✓
    - Lines: 34 → 22 ✓
    - Error paths: 2 → 5 (all handled) ✓

REGRESSIONS: 0
NEW VIOLATIONS: 0
QUALITY DELTA: +4.3

$ # Phase 5: Update backlog
$ # Mark task complete in AUDIT.md
$ go build ./...
$ go test ./... -race
ok  	github.com/example/project/...	1.247s
```

This data-driven approach ensures every backlog task is validated against a quantitative baseline, with `go-stats-generator` differential analysis proving zero regressions before any change is accepted.
