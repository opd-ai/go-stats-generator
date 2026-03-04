# Implementation Plan: Production Readiness — Complexity Remediation Phase

## Phase Overview
- **Objective**: Reduce cyclomatic complexity violations in production code from 13 critical functions (>15.0) to zero, achieving the ≤10.0 production readiness gate
- **Source Document**: ROADMAP.md (Critical Issues - Priority 1C: Complexity)
- **Prerequisites**: None — this is the highest-priority remediation gate
- **Estimated Scope**: Large — 13 functions above complexity 15.0 threshold in production code, 47% duplication ratio, 73% documentation coverage

## Metrics Summary
- **Complexity Hotspots**: 13 production functions above 15.0 threshold; 103 total above 9.0
- **Duplication Ratio**: 47.06% overall (125 clone pairs, 11,384 duplicated lines)
- **Documentation Coverage**: 73.32% overall (packages: 31.8%, functions: 72.3%)
- **Package Coupling**: `cmd` highest at 4.5 (9 deps); `storage` at 3.5 (7 deps); `api` low cohesion 0.8
- **Concurrency Safety**: 0 potential leaks, 2 worker pools, 2 pipelines — PASS ✓

## Implementation Steps

### Priority 1: Critical Complexity Violations (>20.0)

1. **[COMPLETED] Refactor `FilterReportSections` — complexity 30.4 → 3.1 ✓**
   - **File**: `internal/metrics/sections.go`
   - **Deliverable**: Replace switch/case statement with map-based section handler dispatch; extract 5-6 helper functions for section-specific filtering
   - **Dependencies**: None
   - **Metric Justification**: Highest complexity in production code (30.4); 68 lines of code; single point of failure for report filtering
   - **Result**: Reduced complexity from 30.4 to 3.1 (89.8% improvement); extracted 3 helper functions (clearPackageSection, buildSectionKeepSet, clearUnrequestedSections); all functions under 10 lines and complexity ≤5

2. **Refactor `extractNestedBlocks` — complexity 21.5 → target ≤10.0**
   - **File**: `internal/analyzer/duplication.go`
   - **Deliverable**: Extract per-AST-node-type handlers into separate functions; implement visitor pattern for cleaner traversal
   - **Dependencies**: None
   - **Metric Justification**: Second highest complexity (21.5); 45 lines; core duplication detection logic

3. **Refactor `runWatch` — complexity 20.2 → target ≤10.0**
   - **File**: `cmd/watch.go`
   - **Deliverable**: Extract event handling, error recovery, and file system notification logic into separate functions
   - **Dependencies**: None
   - **Metric Justification**: Third highest complexity (20.2); 44 lines; user-facing watch command

### Priority 2: High Complexity Violations (15.0-20.0)

4. **Refactor `walkForNestingDepth` (burden.go) — complexity 19.2 → target ≤10.0**
   - **File**: `internal/analyzer/burden.go`
   - **Deliverable**: Extract per-statement-type depth calculators; create depth calculation helper functions
   - **Dependencies**: None
   - **Metric Justification**: Complexity 19.2; 64 lines; burden analysis core function

5. **Refactor `List` — complexity 19.2 → target ≤10.0**
   - **File**: `internal/storage/json.go`
   - **Deliverable**: Separate filter, sort, and pagination logic into distinct helper functions
   - **Dependencies**: None
   - **Metric Justification**: Complexity 19.2; 63 lines; JSON storage list operation

6. **Refactor `checkStmtForUnreachable` — complexity 18.9 → target ≤10.0**
   - **File**: `internal/analyzer/burden.go`
   - **Deliverable**: Decompose statement type checks into individual checker functions
   - **Dependencies**: Step 4 (shared burden analyzer context)
   - **Metric Justification**: Complexity 18.9; 40 lines; unreachable code detection

7. **Refactor `Generate` (console.go) — complexity 17.4 → target ≤10.0**
   - **File**: `internal/reporter/console.go`
   - **Deliverable**: Extract per-section generation into separate methods; use strategy pattern for section rendering
   - **Dependencies**: None
   - **Metric Justification**: Complexity 17.4; 44 lines; console report generation entry point

8. **Refactor `walkForNestingDepth` (function.go) — complexity 16.6 → target ≤10.0**
   - **File**: `internal/analyzer/function.go`
   - **Deliverable**: Extract per-node-type handlers; consolidate with burden.go walker if possible
   - **Dependencies**: Step 4 (potential code sharing)
   - **Metric Justification**: Complexity 16.6; 56 lines; function nesting analysis

9. **Refactor `detectSingleton` — complexity 16.0 → target ≤10.0**
   - **File**: `internal/analyzer/pattern.go`
   - **Deliverable**: Extract singleton detection heuristics into named helper functions
   - **Dependencies**: None
   - **Metric Justification**: Complexity 16.0; 37 lines; design pattern detection

10. **Refactor `findCommentOutsideStrings` — complexity 15.8 → target ≤10.0**
    - **File**: `internal/analyzer/function.go`
    - **Deliverable**: Extract string literal state machine into separate function; simplify comment detection logic
    - **Dependencies**: None
    - **Metric Justification**: Complexity 15.8; 38 lines; comment parsing utility

11. **Refactor `finalizeTestCoverageMetrics` — complexity 15.3 → target ≤10.0**
    - **File**: `cmd/analyze_finalize.go`
    - **Deliverable**: Extract correlation check and metric aggregation into helper functions
    - **Dependencies**: None
    - **Metric Justification**: Complexity 15.3; 39 lines; test coverage finalization

12. **Refactor `loadOutputConfiguration` — complexity 15.0 → target ≤10.0**
    - **File**: `cmd/analyze_config.go`
    - **Deliverable**: Extract configuration validation and default assignment into separate functions
    - **Dependencies**: None
    - **Metric Justification**: Complexity 15.0; 28 lines; CLI output configuration loading

### Priority 3: Validation and Documentation

13. **Run full re-analysis and validate all complexity gates pass**
    - **Deliverable**: `go-stats-generator analyze . --skip-tests` shows 0 functions above complexity 10.0
    - **Dependencies**: Steps 1-12
    - **Metric Justification**: ROADMAP.md Gate: "All functions ≤ 10 cyclomatic"

14. **Update inline documentation for refactored functions**
    - **Deliverable**: All extracted helper functions have GoDoc comments; package documentation coverage ≥50%
    - **Dependencies**: Steps 1-12
    - **Metric Justification**: Documentation coverage currently 31.8% for packages, target ≥50%

## Technical Specifications

- **Refactoring Pattern**: Extract method — break down complex functions into smaller, focused helpers with single responsibilities
- **Map-based Dispatch**: For switch/case statements with >5 cases (e.g., FilterReportSections), use `map[string]func()` pattern to reduce cyclomatic complexity
- **Visitor Pattern**: For AST traversal functions (walkForNestingDepth, extractNestedBlocks), implement typed visitor interfaces
- **Error Handling**: Preserve existing error propagation; do not change function signatures during refactoring
- **Test Coverage**: All refactoring must pass existing tests; no new test failures allowed
- **Naming Convention**: Helper functions use descriptive verb-noun names (e.g., `handleIfStatement`, `filterBySection`)

## Validation Criteria

- [ ] `go-stats-generator analyze . --skip-tests --sections functions | grep -c "complexity.*>[0-9]\{2\}"` returns 0 for production code
- [ ] Zero functions in `internal/` and `cmd/` with cyclomatic complexity >10.0
- [ ] `go test ./...` passes with no new failures
- [ ] `go-stats-generator diff baseline.json post-refactor.json` shows complexity reductions, no regressions in other metrics
- [ ] Package documentation coverage ≥50% (up from 31.8%)
- [ ] All refactored functions have GoDoc comments explaining their purpose

## Known Gaps

- **ARIMA Forecasting**: README.md promises ARIMA/exponential smoothing but this is deferred — linear regression covers 80% of use cases
- **Trend Placeholder Commands**: `trend forecast` and `trend regressions` remain placeholders — separate phase (see GAPS.md Gap 1)
- **High Duplication Ratio**: 47% duplication inflated by testdata — recommend `--exclude testdata` for accurate production metrics
- **Annotation Resolution**: 1 TODO, 1 FIXME, 1 HACK, 1 BUG in `internal/metrics/types.go` — not blocking this phase

## Phase Completion Checklist

```bash
# Validate complexity gate
go-stats-generator analyze . --skip-tests --format json --output post-refactor.json --sections functions
cat post-refactor.json | jq '[.functions[] | select(.complexity.overall > 10) | select(.file | startswith("testdata/") | not)] | length'
# Expected: 0

# Validate no test regressions
go test ./... -count=1

# Validate documentation improvement
cat post-refactor.json | jq '.documentation.coverage.packages'
# Expected: ≥50

# Generate human-readable summary
go-stats-generator analyze . --skip-tests --max-complexity 10 --max-function-length 30
```

---

**Generated**: 2026-03-04T06:50:00Z  
**Analysis Tool**: go-stats-generator v1.0.0  
**Source Metrics**: metrics.json (104 files, --skip-tests)
