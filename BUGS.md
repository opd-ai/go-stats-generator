# TASK DESCRIPTION:
Perform a data-driven bug identification and remediation pass to find and autonomously fix **bug-prone functions** in the codebase using complexity metrics as risk indicators. Use `go-stats-generator` baseline analysis (with --skip-tests), targeted bug fixing guided by complexity and concurrency metrics, and differential validation to ensure fixes introduce no regressions while reducing defect risk.

When results are ambiguous, such as a tie between complexity scores or if one threshold is exceeded but not another, always choose the function with the highest cyclomatic complexity first.

## CONSTRAINT:

Use only `go-stats-generator` and existing tests for your analysis. Only fix bugs with clear, deterministic solutions. Preserve existing functionality and API contracts. Maintain code style consistency with surrounding code. Do not modify comments, imports, or formatting unless necessary for the fix. Skip fixes that require architectural changes or unclear requirements.

## PREREQUISITES:
**Minimum Required Version:** `go-stats-generator` v1.0.0 or higher
Install and configure `go-stats-generator` for complexity-guided bug detection and regression validation:

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
# Phase 1: Establish baseline and identify high-risk functions
go-stats-generator analyze . --max-complexity 10 --max-function-length 30 --skip-tests --format json --output baseline.json
go-stats-generator analyze . --max-complexity 10 --max-function-length 30 --skip-tests

# Phase 2: Inspect bug-prone areas
# Using the results generated in phase 1, select high-complexity functions as bug candidates.

# Phase 3: Post-fix validation
go-stats-generator analyze . --format json --output fixed.json --max-complexity 10 --max-function-length 30 --skip-tests

# Phase 4: Regression check
go-stats-generator diff baseline.json fixed.json
go-stats-generator diff baseline.json fixed.json --format html --output bug-fixes-report.html
```

## CONTEXT:
You are an automated Go bug hunter using `go-stats-generator` as the primary analysis engine for enterprise-grade defect detection and regression validation. The tool provides precise complexity, concurrency, and structural metrics that identify bug-prone code regions. High cyclomatic complexity (>15) and deep nesting (>4) are strong predictors of latent defects. Focus on functions flagged by the tool's analysis engine as exceeding risk thresholds, then systematically scan them for the following bug categories:
1. Runtime errors (nil pointers, race conditions, deadlocks)
2. Resource leaks (unclosed files, goroutine leaks, memory leaks)
3. Logic errors (incorrect conditionals, off-by-one errors, improper error handling)
4. Concurrency issues (unsafe concurrent access, missing synchronization)
5. Security vulnerabilities (injection risks, unsafe input handling)

## INSTRUCTIONS:

### Phase 1: Data-Driven Risk Identification
1. **Run Baseline Analysis:**
  ```bash
  go-stats-generator analyze . --skip-tests --format json --output baseline.json
  go-stats-generator analyze . --skip-tests
  ```
  - Record all functions exceeding bug-risk thresholds (see BUG RISK THRESHOLDS below)
  - Note specific risk contributors (cyclomatic complexity, nesting depth, concurrency primitives) for each
  - Identify the files containing these high-risk functions

2. **Extract Concurrency Risk Profile:**
  ```bash
  cat baseline.json | jq '.concurrency'
  ```
  - Capture goroutine spawn sites, channel operations, sync primitive usage
  - Flag functions with concurrent access patterns lacking synchronization
  - Identify goroutine leaks (spawns without cancellation or timeout)

3. **Prioritize Bug Candidates:**
  From the baseline analysis, select functions for bug inspection in priority order:
  - **CRITICAL (Cyclomatic > 20 OR nesting > 5):** Inspect and fix immediately — highest defect probability
  - **HIGH (Cyclomatic > 15 OR nesting > 4):** Inspect and fix in second pass
  - **MEDIUM (Cyclomatic 10-15 OR nesting 3-4):** Inspect if time permits
  - **LOW (Cyclomatic < 10 AND nesting < 3):** Inspect only if flagged by concurrency analysis

  When prioritizing between functions:
  - If cyclomatic scores are tied, choose the function with **deeper nesting** first
  - If both metrics are tied, choose the function with **more concurrency primitives** first
  - Prioritize fixes: CRITICAL > HIGH > MEDIUM > LOW

4. **Generate Risk Map:**
  ```bash
  cat baseline.json | jq '[.functions[] | select(.complexity.cyclomatic > 10 or .complexity.nesting_depth > 3) | {name: .name, file: .file, cyclomatic: .complexity.cyclomatic, nesting_depth: .complexity.nesting_depth, line_count: .lines.code}]'
  ```
  - Produce a ranked list of bug-candidate functions with their risk indicators
  - Cross-reference with `.concurrency` data for race condition candidates

### Phase 2: Targeted Bug Detection and Fixing
1. **Scan High-Risk Functions:**
  For each function identified in Phase 1, inspect for:
  - **Nil pointer dereferences:** Unchecked returns from functions that may return nil
  - **Resource leaks:** Opened files/connections without deferred close; goroutines without exit paths
  - **Error swallowing:** Errors assigned to `_` or checked but not propagated
  - **Off-by-one errors:** Loop bounds, slice indexing, range calculations
  - **Race conditions:** Shared state accessed without synchronization (cross-reference `.concurrency`)
  - **Deadlocks:** Lock ordering violations, missing unlocks in error paths

2. **Apply Deterministic Fixes:**
  - Add nil checks before pointer dereferences
  - Add `defer resource.Close()` for unclosed resources
  - Propagate swallowed errors to callers
  - Correct loop bounds and slice indices
  - Add mutex protection for unsynchronized shared state
  - Fix lock ordering to prevent deadlocks
  - Add context cancellation for goroutine lifecycle management

3. **Preserve Correctness:**
  - Maintain error propagation chains from all original call sites
  - Keep defer statements in correct scope
  - Preserve variable access patterns and side effects
  - **Carefully consider mutexes, locks, and thread safety:**
    - Preserve lock boundaries for critical sections (e.g., code protected by `sync.Mutex`, `sync.RWMutex`, or other synchronization primitives)
    - Avoid introducing new lock/unlock pairs that could create deadlocks
    - Ensure fixes that add synchronization are consistent with existing locking strategy
    - After fixing concurrency-sensitive code, validate with `go test -race` to detect data races

### Phase 3: Differential Validation
1. **Measure Regression Risk:**
  ```bash
  go-stats-generator analyze . --skip-tests --format json --output fixed.json
  go-stats-generator diff baseline.json fixed.json
  ```
  - Verify no function's complexity increased after fixes
  - Confirm no new functions exceed risk thresholds
  - Check for zero regressions in unchanged code
  - Verify concurrency metrics remain stable or improve

2. **Generate Fix Report:**
  ```bash
  go-stats-generator diff baseline.json fixed.json --format html --output bug-fixes-report.html
  ```

### Phase 4: Quality Verification
1. **Validate Fix Safety:**
  - All tests pass: `go test ./...`
  - Race condition check: `go test -race ./...`
  - Build succeeds: `go build ./...`
  - No complexity regressions detected by diff analysis
  - Error handling paths preserved or strengthened

2. **Confirm Bug Elimination:**
  - Each fix addresses a specific, identifiable defect
  - No fix introduces new error paths without handling
  - Return value semantics preserved
  - Flag any bugs that require manual review with: ⚠️ MANUAL REVIEW NEEDED

## OUTPUT FORMAT:

Structure your response as:

### 1. Baseline Risk Summary
```
go-stats-generator identified high-risk functions:
  Functions exceeding risk thresholds: [n]
  Concurrency hotspots: [n]
  Total functions analyzed: [n]

Top bug candidates:
1. Function: [name] in [file]
   - Cyclomatic: [score] | Nesting: [depth] | Lines: [count]
   - Risk Level: [CRITICAL/HIGH/MEDIUM]
   - Concurrency primitives: [goroutines/channels/mutexes or none]
   - Bug categories to inspect: [list]

2. Function: [name] in [file]
   - Cyclomatic: [score] | Nesting: [depth] | Lines: [count]
   - Risk Level: [CRITICAL/HIGH/MEDIUM]
   - Concurrency primitives: [goroutines/channels/mutexes or none]
   - Bug categories to inspect: [list]

... (continue for all functions exceeding thresholds)
```

### 2. Fixes Applied
For each file modified, show:

**Fixed: `path/to/file.go`**
- Line [X]: [CRITICAL/HIGH/MEDIUM/LOW] [Brief issue description]
  - Applied: [Specific change made]
  - Risk indicator: cyclomatic [score], nesting [depth]

### 3. Regression Validation
```
Differential analysis results (go-stats-generator diff):
- Functions fixed: [count]
- Complexity regressions: [count] (must be 0)
- Concurrency metrics: [stable/improved]
- New threshold violations: [count] (must be 0)
- Overall quality trend: [positive/neutral]
```

### 4. Summary
```
SUMMARY: Fixed [X] bugs across [Y] files
  CRITICAL: [X] | HIGH: [X] | MEDIUM: [X] | LOW: [X]
  ⚠️ MANUAL REVIEW NEEDED: [X] items
  Regressions: [count]
```

## BUG RISK THRESHOLDS:
```
Bug Risk Classification (go-stats-generator metrics):
  CRITICAL = Cyclomatic > 20 OR Nesting Depth > 5
  HIGH     = Cyclomatic > 15 OR Nesting Depth > 4
  MEDIUM   = Cyclomatic 10-15 OR Nesting Depth 3-4
  LOW      = Cyclomatic < 10 AND Nesting Depth < 3

Concurrency Risk Indicators:
  Goroutine spawns without context cancellation
  Channel operations without timeout or select default
  Shared state access without sync primitives in .concurrency output
  Lock/unlock pairs spanning multiple control flow branches

Post-Fix Quality Gates:
  All tests pass (go test ./...)
  Race detector clean (go test -race ./...)
  No complexity regressions (go-stats-generator diff)
  No new threshold violations introduced
  Build succeeds (go build ./...)
```
<!-- Last verified: 2025-07-25 against function.go:calculateComplexity and concurrency.go:AnalyzeConcurrency -->

Bug Fix Threshold = Cyclomatic > 15 OR Nesting > 4 OR Concurrency Risk Detected
- If no targets: "Bug scan complete: go-stats-generator baseline analysis found no functions exceeding bug-risk thresholds."

## EXAMPLE WORKFLOW:
```bash
$ go-stats-generator analyze . --skip-tests
=== BUG RISK ANALYSIS ===
Functions exceeding risk thresholds: 6
Concurrency hotspots: 2
Total functions analyzed: 142

HIGH-RISK FUNCTIONS:
1. processComplexOrder (order.go): cyclomatic 22, nesting 5 [CRITICAL]
   - Lines: 58 code lines
   - Concurrency: 2 goroutine spawns, 1 mutex
   - Bug risk: nil pointer, resource leak, race condition

2. handleConcurrentUpdate (sync.go): cyclomatic 18, nesting 4 [HIGH]
   - Lines: 45 code lines
   - Concurrency: 3 goroutine spawns, 2 channels, 1 mutex
   - Bug risk: deadlock, race condition, goroutine leak

3. parseUntrustedInput (parser.go): cyclomatic 16, nesting 4 [HIGH]
   - Lines: 41 code lines
   - Concurrency: none
   - Bug risk: nil pointer, injection, error swallowing

4. transformDataPipeline (transform.go): cyclomatic 14, nesting 4 [MEDIUM]
   - Lines: 52 code lines
   - Concurrency: 1 goroutine spawn, 2 channels
   - Bug risk: goroutine leak, off-by-one

5. validateAndStore (storage.go): cyclomatic 12, nesting 3 [MEDIUM]
   - Lines: 38 code lines
   - Concurrency: none
   - Bug risk: resource leak, error swallowing

6. buildResponsePayload (handler.go): cyclomatic 11, nesting 3 [MEDIUM]
   - Lines: 35 code lines
   - Concurrency: none
   - Bug risk: nil pointer, logic error

$ cat baseline.json | jq '[.functions[] | select(.complexity.cyclomatic > 15) | {name: .name, file: .file, cyclomatic: .complexity.cyclomatic, nesting_depth: .complexity.nesting_depth}]'
[
  {"name": "processComplexOrder", "file": "order.go", "cyclomatic": 22, "nesting_depth": 5},
  {"name": "handleConcurrentUpdate", "file": "sync.go", "cyclomatic": 18, "nesting_depth": 4},
  {"name": "parseUntrustedInput", "file": "parser.go", "cyclomatic": 16, "nesting_depth": 4}
]

$ cat baseline.json | jq '.concurrency'
{
  "goroutine_spawns": 8,
  "channel_operations": 12,
  "mutex_usage": 5,
  "waitgroup_usage": 3,
  "hotspots": [
    {"function": "handleConcurrentUpdate", "file": "sync.go", "goroutines": 3, "channels": 2, "mutexes": 1},
    {"function": "processComplexOrder", "file": "order.go", "goroutines": 2, "channels": 0, "mutexes": 1}
  ]
}

$ # Fix each bug-prone function in priority order...

$ go-stats-generator analyze . --skip-tests --format json --output fixed.json
$ go-stats-generator diff baseline.json fixed.json
=== REGRESSION CHECK ===
FUNCTIONS MODIFIED: 6

BUG FIXES VALIDATED:
- processComplexOrder: cyclomatic 22 → 22 (stable, 3 bugs fixed) ✓
- handleConcurrentUpdate: cyclomatic 18 → 17 (improved, 2 bugs fixed) ✓
- parseUntrustedInput: cyclomatic 16 → 15 (improved, 2 bugs fixed) ✓
- transformDataPipeline: cyclomatic 14 → 14 (stable, 1 bug fixed) ✓
- validateAndStore: cyclomatic 12 → 11 (improved, 2 bugs fixed) ✓
- buildResponsePayload: cyclomatic 11 → 11 (stable, 1 bug fixed) ✓

COMPLEXITY REGRESSIONS: 0
NEW THRESHOLD VIOLATIONS: 0
CONCURRENCY METRICS: improved (2 race conditions eliminated)

SUMMARY: Fixed 11 bugs across 6 files
  CRITICAL: 3 | HIGH: 4 | MEDIUM: 3 | LOW: 1
  ⚠️ MANUAL REVIEW NEEDED: 0 items
  REGRESSIONS: 0
```

This data-driven approach uses `go-stats-generator` complexity and concurrency metrics as bug-risk indicators, ensuring inspection effort is concentrated on the highest-risk functions with measurable regression validation after every fix.
