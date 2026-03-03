# TASK DESCRIPTION:
Perform a data-driven repository cleanup to remove binary artifacts, eliminate redundant reports, consolidate duplicate tests, and update `.gitignore`, using `go-stats-generator` baseline metrics and differential validation to ensure measurable codebase quality improvement with zero regressions.

## Execution Mode
**Autonomous Action** — Execute all steps directly. No user approval required between steps.

## CONSTRAINT:

Use only `go-stats-generator` and existing tests for your analysis. You are absolutely forbidden from using any other code analysis tools. All cleanup decisions must be validated against `go-stats-generator` metrics before and after changes.

## PREREQUISITES:
**Minimum Required Version:** `go-stats-generator` v1.0.0 or higher
Install and configure `go-stats-generator` for comprehensive cleanup analysis and improvement tracking:

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
go-stats-generator analyze --format json | jq '{duplication: .duplication}'
which jq || sudo apt-get install -y jq
```
**Section filter**: Use only `.duplication` from the report. Exclude `.functions`, `.structs`, `.interfaces`, `.packages`, `.patterns`, `.concurrency`, `.complexity`, `.documentation`, `.generics`, `.naming`, `.placement`, `.organization`, `.burden`, `.scores`, `.suggestions` — they are not relevant to repository cleanup.

### Required Analysis Workflow:
```bash
# Phase 1: Establish pre-cleanup baseline
go-stats-generator analyze . --skip-tests --format json --output pre-cleanup.json --sections duplication
go-stats-generator analyze . --skip-tests

# Phase 2: Execute cleanup steps (remove binaries, reports; consolidate tests; update .gitignore)
# Perform all cleanup actions described in the INSTRUCTIONS section below.

# Phase 3: Post-cleanup validation
go-stats-generator analyze . --skip-tests --format json --output post-cleanup.json --sections duplication

# Phase 4: Measure and document improvements
go-stats-generator diff pre-cleanup.json post-cleanup.json
go-stats-generator diff pre-cleanup.json post-cleanup.json --format html --output cleanup-report.html
```

## CONTEXT:
You are an automated Go repository maintainer using `go-stats-generator` for enterprise-grade codebase hygiene validation. The tool provides precise duplication, complexity, and documentation metrics that quantify the impact of cleanup operations. Use baseline analysis before cleanup to establish measurable targets, then differential analysis after cleanup to verify improvements. Focus on duplication metrics to identify redundant reports and overlapping test files, and overall codebase quality metrics to confirm no regressions are introduced.

## INSTRUCTIONS:

### Phase 1: Pre-Cleanup Baseline
1. **Run Baseline Analysis:**
  ```bash
  go-stats-generator analyze . --skip-tests --format json --output pre-cleanup.json
  go-stats-generator analyze . --skip-tests
  ```
  - Record the current codebase metrics as the baseline:
    * Total files and packages
    * Duplication ratio and clone pair count
    * Overall complexity distribution
    * Documentation coverage
  - These metrics serve as the reference point for validating that cleanup improves (or at minimum does not degrade) codebase quality

2. **Extract Duplication Summary for Redundancy Detection:**
  ```bash
  cat pre-cleanup.json | jq '.duplication'
  ```
  - Capture duplication metrics to help identify redundant reports and overlapping test files:
    * Clone Pairs Detected (total number of clone groups)
    * Duplicated Lines (total duplicated lines across all clones)
    * Duplication Ratio (percentage of codebase that is duplicated)
  - Use duplication data to inform which reports are truly redundant and which test files overlap

### Phase 2: Execute Cleanup Actions

1. **Remove Binary Files:**
  - Scan the repository for committed binary artifacts (e.g., `.o`, `.so`, `.dll`, `.exe`, `.bin`, `.pyc`, `.class`, `.wasm`, compiled executables, and similar non-source files)
  - Delete all identified binary files from the repository
  - **Do not remove** binaries that serve as intentional test fixtures or are documented as required assets

2. **Remove Redundant Reports:**
  - Cross-reference duplication analysis from Phase 1 to identify duplicate or auto-generated report files (e.g., repeated coverage reports, stale build logs, duplicate analysis outputs)
  - Delete files that are exact duplicates or superseded by newer versions
  - **Do not remove** the most recent version of any report that has no replacement

3. **Consolidate Tests:**
  - Use duplication metrics from Phase 1 to identify test files with overlapping or duplicate test cases:
    ```bash
    cat pre-cleanup.json | jq '.duplication.clones[] | select(.instances[].file | test("_test\\.go$")) | {type, line_count, instances: (.instances | length), locations: [.instances[] | "\(.file):\(.start_line)-\(.end_line)"]}'
    ```
  - Merge overlapping test files into single cohesive test files, grouped by module or feature
  - Remove duplicate test cases while preserving full test coverage (no test logic lost)
  - Ensure all consolidated test files pass after merging:
    ```bash
    go test ./...
    ```

4. **Update `.gitignore`:**
  - Append entries to `.gitignore` (or create it if absent) to prevent future commits of:
    - Binary artifact types removed in Step 1 (e.g., `*.o`, `*.so`, `*.dll`, `*.exe`, `*.bin`, `*.pyc`, `*.class`, `*.wasm`)
    - Auto-generated report directories/patterns removed in Step 2
  - Do not duplicate entries already present in `.gitignore`
  - Group new entries under a `# Repository cleanup` comment header

### Phase 3: Differential Validation
1. **Measure Improvements:**
  ```bash
  go-stats-generator analyze . --skip-tests --format json --output post-cleanup.json
  go-stats-generator diff pre-cleanup.json post-cleanup.json
  ```
  - Verify duplication ratio decreased (or remained stable if no code duplication existed)
  - Confirm no complexity regressions in any remaining functions
  - Confirm all tests still pass:
    ```bash
    go test ./...
    go test -race ./...
    ```
  - Confirm build succeeds:
    ```bash
    go build ./...
    ```

2. **Generate Improvement Report:**
  ```bash
  go-stats-generator diff pre-cleanup.json post-cleanup.json --format html --output cleanup-report.html
  ```

### Phase 4: Quality Verification
1. **Validate Metrics Achievement:**
  - Duplication ratio decreased or held steady (no increase)
  - No complexity regressions detected by diff analysis
  - All committed binary artifacts are removed (except intentional fixtures)
  - No actively used or necessary files are deleted
  - Full test coverage is preserved — no test cases lost
  - `.gitignore` prevents reintroduction of all removed file types
  - Overall codebase quality trend positive or neutral

2. **Confirm Functional Preservation:**
  - All tests pass: `go test ./...`
  - Race condition check: `go test -race ./...`
  - Build succeeds: `go build ./...`
  - Error handling paths unchanged
  - Return value semantics preserved

## OUTPUT FORMAT:

Structure your response as:

### 1. Pre-Cleanup Baseline
```
go-stats-generator pre-cleanup baseline:
  Total Files: [n]
  Total Packages: [n]
  Duplication Ratio: [x.xx]%
  Clone Pairs: [n]
  Duplicated Lines: [n]
  Overall Complexity (avg): [x.x]
  Documentation Coverage: [x.xx]%
```

### 2. Cleanup Actions Performed
```
Cleanup Summary:
- Binaries removed: [count] files ([total size])
- Reports removed: [count] files
- Tests consolidated: [count] files merged into [count] files
- .gitignore entries added: [count] new patterns
```

### 3. Improvement Validation
```
Differential analysis results (go-stats-generator diff):
- Duplication Ratio: [old_%] → [new_%] ([change_%])
- Clone Pairs: [old_count] → [new_count] ([eliminated_count] eliminated)
- Duplicated Lines: [old_count] → [new_count] ([lines_saved] lines saved)
- Complexity regressions: [count]
- Tests passing: Yes/No
- Race check passing: Yes/No
- Build passing: Yes/No
- Overall quality improvement: [score]
```

## CLEANUP THRESHOLDS:
```
Pre/Post Validation Gates:
  Duplication Ratio    — must not increase after cleanup
  Complexity (avg)     — must not increase after cleanup (no regressions)
  Test Pass Rate       — must remain 100%
  Build Status         — must remain passing

Cleanup Targets:
  Binary Artifacts     — 0 committed binaries (except intentional test fixtures)
  Redundant Reports    — 0 exact-duplicate or superseded report files
  Test Duplication     — overlapping test cases consolidated; duplication ratio should decrease
  .gitignore Coverage  — all removed artifact patterns added

Post-Cleanup Quality Gates:
  Overall Complexity ≤ 9.0 per function (no new violations)
  Cyclomatic Complexity ≤ 9 per function (no new violations)
  Function Length ≤ 40 lines (no new violations)
  No new clone pairs introduced
```
<!-- Last verified: 2025-07-25 against go-stats-generator analyze and diff commands -->

Cleanup Threshold = Duplication Ratio increased OR Complexity regressed OR Tests failing
- If no cleanup needed: "Cleanup complete: go-stats-generator baseline analysis found no binary artifacts, redundant reports, or duplicate tests requiring action."

## EXAMPLE WORKFLOW:
```bash
$ go-stats-generator analyze . --skip-tests
=== CODEBASE SUMMARY ===
Total Files: 42
Total Packages: 8
Duplication Ratio: 6.31%
Clone Pairs: 9
Duplicated Lines: 112
Avg Complexity: 5.8
Documentation Coverage: 72.4%

=== DUPLICATION ANALYSIS ===
Top Clone Pairs (by size):
Type            Lines Instances Locations
--------------------------------------------------------------------------------
exact              14        2 internal/reporter/html_test.go:30-43 (+1 more)
renamed            10        3 internal/analyzer/function_test.go:88-97 (+2 more)
near                8        2 internal/scanner/worker_test.go:15-22 (+1 more)

$ # Capture pre-cleanup baseline
$ go-stats-generator analyze . --skip-tests --format json --output pre-cleanup.json
$ cat pre-cleanup.json | jq '{duplication_ratio: .duplication.duplication_ratio, clone_pairs: .duplication.clone_pairs, duplicated_lines: .duplication.duplicated_lines}'
{
  "duplication_ratio": 0.0631,
  "clone_pairs": 9,
  "duplicated_lines": 112
}

$ # Phase 2: Execute cleanup actions
$ find . -name "*.exe" -o -name "*.bin" -o -name "*.o" -o -name "*.so" | head
./build/go-stats-generator.exe
./tmp/test-output.bin

$ rm ./build/go-stats-generator.exe ./tmp/test-output.bin
$ # Remove superseded reports...
$ rm ./reports/coverage-2024-01-15.html ./reports/coverage-2024-02-20.html
$ # Consolidate overlapping test cases...
$ # Update .gitignore with removed artifact patterns...

$ # Phase 3: Validate with go-stats-generator diff
$ go-stats-generator analyze . --skip-tests --format json --output post-cleanup.json
$ go-stats-generator diff pre-cleanup.json post-cleanup.json
=== IMPROVEMENT SUMMARY ===
CLEANUP VALIDATION:

DUPLICATION METRICS:
- Duplication Ratio: 6.31% → 3.87% (39% reduction) ✓
- Clone Pairs: 9 → 5 (4 eliminated) ✓
- Duplicated Lines: 112 → 58 (54 lines saved) ✓

COMPLEXITY CHECK:
- Avg Complexity: 5.8 → 5.8 (no change) ✓
- Regressions: 0 ✓

QUALITY GATES:
  All functions below complexity threshold: ✓
  No new clone pairs introduced: ✓
  All tests passing: ✓
  Race condition check: ✓
  Build passing: ✓

QUALITY SCORE: 88/74 (+14 improvement)
REGRESSIONS: 0
CLEANUP VALIDATION PASSED: ✓

$ go-stats-generator diff pre-cleanup.json post-cleanup.json --format html --output cleanup-report.html

### Cleanup Summary
- Binaries removed: 2 files (14.2 MB)
- Reports removed: 2 files
- Tests consolidated: 4 files merged into 2 files
- .gitignore entries added: 6 new patterns
- Tests passing: Yes
```

This data-driven approach ensures cleanup decisions are validated against quantitative `go-stats-generator` metrics rather than subjective assessment, with measurable before/after comparison confirming no regressions are introduced across any codebase quality dimension.
