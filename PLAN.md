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

2. **[COMPLETED] Refactor `extractNestedBlocks` — complexity 21.5 → 3.1 ✓**
   - **File**: `internal/analyzer/duplication.go`
   - **Deliverable**: Extract per-AST-node-type handlers into separate functions; implement visitor pattern for cleaner traversal
   - **Dependencies**: None
   - **Metric Justification**: Second highest complexity (21.5); 45 lines; core duplication detection logic
   - **Result**: Reduced complexity from 21.5 to 3.1 (85.6% improvement); extracted 5 helper functions (extractFromIfStmt, extractFromLoopBody, extractFromSwitchStmt, extractFromTypeSwitchStmt, extractFromSelectStmt); all functions under 18 lines and complexity ≤6.2

3. **[COMPLETED] Refactor `runWatch` — complexity 20.2 → 4.4 ✓**
   - **File**: `cmd/watch.go`
   - **Deliverable**: Extract event handling, error recovery, and file system notification logic into separate functions
   - **Dependencies**: None
   - **Metric Justification**: Third highest complexity (20.2); 44 lines; user-facing watch command
   - **Result**: Reduced complexity from 20.2 to 4.4 (78.2% improvement); lines reduced from 44 to 16 (63.6% reduction); extracted 7 helper functions (getWatchPath, createWatcher, printWatchStartMessage, watchEventLoop, shouldStopOnChannel, handleFileEvent, handleWatchError); all new functions ≤16 lines; primary function runWatch at 4.4 complexity (well under 10.0 threshold)

### Priority 2: High Complexity Violations (15.0-20.0)

4. **[COMPLETED] Refactor `walkForNestingDepth` (burden.go) — complexity 19.2 → 4.4 ✓**
   - **File**: `internal/analyzer/burden.go`
   - **Deliverable**: Extract per-statement-type depth calculators; create depth calculation helper functions
   - **Dependencies**: None
   - **Metric Justification**: Complexity 19.2; 64 lines; burden analysis core function
   - **Result**: Reduced complexity from 19.2 to 4.4 (77.1% improvement); lines reduced from 64 to 23 (64.1% reduction); extracted 10 helper functions (walkIfStmtNesting, walkForStmtNesting, walkRangeStmtNesting, walkSwitchStmtNesting, walkTypeSwitchStmtNesting, walkSelectStmtNesting, walkCaseClauseNesting, walkCommClauseNesting, walkBlockStmtNesting, updateMaxDepth); all new functions ≤6 lines; primary function walkForNestingDepth at 4.4 complexity (well under 10.0 threshold)

5. **[COMPLETED] Refactor `List` — complexity 19.2 → 3.1 ✓**
   - **File**: `internal/storage/json.go`
   - **Deliverable**: Separate filter, sort, and pagination logic into distinct helper functions
   - **Dependencies**: None
   - **Metric Justification**: Complexity 19.2; 63 lines; JSON storage list operation
   - **Result**: Reduced complexity from 19.2 to 3.1 (83.9% improvement); lines reduced from 63 to 7 (88.9% reduction); extracted 8 helper functions (readStorageDirectory, collectSnapshots, processSnapshotFile, isSnapshotFile, loadSnapshotMetadata, buildSnapshotInfo, sortSnapshotsByTimestamp, applyPagination); all functions ≤18 lines and complexity ≤7.0

6. **[COMPLETED] Refactor `checkStmtForUnreachable` — complexity 18.9 → 3.1 ✓**
   - **File**: `internal/analyzer/burden.go`
   - **Deliverable**: Decompose statement type checks into individual checker functions
   - **Dependencies**: Step 4 (shared burden analyzer context)
   - **Metric Justification**: Complexity 18.9; 40 lines; unreachable code detection
   - **Result**: Reduced complexity from 18.9 to 3.1 (83.6% improvement); cyclomatic reduced from 13 to 2 (84.6% improvement); extracted 4 helper functions (checkIfStmtUnreachable, checkElseClauseUnreachable, checkLoopBodyUnreachable, checkSwitchCasesUnreachable); all new functions ≤10 lines and complexity ≤6.2; primary function checkStmtForUnreachable at 3.1 complexity (well under 10.0 threshold)

7. **[COMPLETED] Refactor `Generate` (console.go) — complexity 17.4 → 1.3 ✓**
   - **File**: `internal/reporter/console.go`
   - **Deliverable**: Extract per-section generation into separate methods; use strategy pattern for section rendering
   - **Dependencies**: None
   - **Metric Justification**: Complexity 17.4; 44 lines; console report generation entry point
   - **Result**: Reduced complexity from 17.4 to 1.3 (92.5% improvement); cyclomatic from 13 to 1 (92.3% improvement); extracted 13 helper functions (writeReportSections at 4.9, 12 shouldWrite* functions at 1.0 each); primary function Generate at 1.3 complexity (well under 10.0 threshold)

8. **[COMPLETED] Refactor `walkForNestingDepth` (function.go) — complexity 16.6 → 4.4 ✓**
   - **File**: `internal/analyzer/function.go`
   - **Deliverable**: Extract per-node-type handlers; consolidate with burden.go walker if possible
   - **Dependencies**: Step 4 (potential code sharing)
   - **Metric Justification**: Complexity 16.6; 56 lines; function nesting analysis
   - **Result**: Reduced complexity from 16.6 to 4.4 (73.5% improvement); lines reduced from 56 to 21 (62.5% reduction); extracted 9 helper functions (walkIfStmtNesting, walkForStmtNesting, walkRangeStmtNesting, walkSwitchStmtNesting, walkTypeSwitchStmtNesting, walkSelectStmtNesting, walkBlockStmtNesting, walkDefaultNodeNesting, updateMaxNestingDepth); all new functions ≤7 lines and complexity ≤3.1; primary function walkForNestingDepth at 4.4 complexity (well under 10.0 threshold)

9. **[COMPLETED] Refactor `detectSingleton` — complexity 16.0 → 6.2 ✓**
   - **File**: `internal/analyzer/pattern.go`
   - **Deliverable**: Extract singleton detection heuristics into named helper functions
   - **Dependencies**: None
   - **Metric Justification**: Complexity 16.0; 37 lines; design pattern detection
   - **Result**: Reduced complexity from 16.0 to 6.2 (61.3% improvement); lines reduced from 41 to 16 (61.0% reduction); extracted 5 helper functions (inspectVarDeclForSyncOnce, checkValueSpecForSyncOnce, hasSyncOnceType, hasSyncOnceValue, addSingletonPattern); all functions ≤10 lines and complexity ≤6.2

10. **[COMPLETED] Refactor `findCommentOutsideStrings` — complexity 15.8 → 8.8 ✓**
    - **File**: `internal/analyzer/function.go`
    - **Deliverable**: Extract string literal state machine into separate function; simplify comment detection logic
    - **Dependencies**: None
    - **Metric Justification**: Complexity 15.8; 38 lines; comment parsing utility
    - **Result**: Reduced complexity from 15.8 to 8.8 (44.3% improvement); cyclomatic from 11 to 6 (45.5% improvement); extracted 4 helper functions (processDoubleQuoteChar, processBacktickChar, checkStringStart, matchesCommentMarker); all new functions ≤12 lines and complexity ≤5.7; primary function findCommentOutsideStrings at 8.8 complexity (well under 10.0 threshold)

11. **[COMPLETED] Refactor `finalizeTestCoverageMetrics` — complexity 15.3 → 3.1 ✓**
    - **File**: `cmd/analyze_finalize.go`
    - **Deliverable**: Extract correlation check and metric aggregation into helper functions
    - **Dependencies**: None
    - **Metric Justification**: Complexity 15.3; 39 lines; test coverage finalization
    - **Result**: Reduced complexity from 15.3 to 3.1 (79.7% improvement); cyclomatic from 11 to 2 (81.8% improvement); extracted 5 helper functions (loadAndAnalyzeCoverage, resolveCoveragePath, logCoverageResults, analyzeTestQualityMetrics, logVerbose); all functions under 15 lines and complexity ≤3.1

12. **[COMPLETED] Refactor `loadOutputConfiguration` — complexity 15.0 → 7.0 ✓**
    - **File**: `cmd/analyze_config.go`
    - **Deliverable**: Extract configuration validation and default assignment into separate functions
    - **Dependencies**: None
    - **Metric Justification**: Complexity 15.0; 28 lines; CLI output configuration loading
    - **Result**: Reduced complexity from 15.0 to 7.0 (53.3% improvement); cyclomatic from 10 to 5 (50.0% improvement); extracted 2 helper functions (applyVerboseDefaults at 3.1 complexity, mergeSectionFlags at 8.5 complexity); all functions under 10.0 threshold

### Priority 3: Validation and Documentation

13. **[IN PROGRESS ⏳] Reduce remaining production functions to complexity ≤10.0**
    - **Status**: 44 production functions remain above complexity 10.0 (25 completed, 44 remaining out of 69 total)
    - **Progress**: 
      - ✅ Refactored `finalizeNamingMetrics` (14.5 → 3.1, 78.6% improvement)
      - ✅ Refactored `compareFunctionMetrics` (13.7 → 1.3, 90.5% improvement)
      - ✅ Refactored `calculateDelta` (13.7 → 1.3, 90.5% improvement)
      - ✅ Refactored `AnalyzeInterfacesWithPath` (13.4 → 1.3, 90.3% improvement)
      - ✅ Refactored `CalculateDocQualityScore` (13.2 → 3.1, 76.5% improvement; cyclomatic 9 → 2, 77.8% improvement)
      - ✅ Refactored `Store` (13.2 → 7.0, 47.0% improvement)
      - ✅ Refactored `runTrendRegressions` (13.2 → 5.7, 56.8% improvement)
      - ✅ Refactored `worker` (13.2 → 4.9, 62.9% improvement; cyclomatic 9 → 3, 66.7% improvement)
      - ✅ Refactored `matchesFilter` (13.2 → 1.3, 90.2% improvement; cyclomatic 9 → 1, 88.9% improvement) - 2 instances in memory.go & json.go
      - ✅ Refactored `getAuthorStats` (13.2 → 3.1, 76.5% improvement; cyclomatic 9 → 2, 77.8% improvement)
      - ✅ Refactored `checkNodeContext` (13.2 → 3.1, 76.5% improvement; cyclomatic 9 → 2, 77.8% improvement)
      - ✅ Refactored `AnalyzeFunctionAffinity` (13.2 → 4.9, 62.9% improvement; cyclomatic 9 → 3, 66.7% improvement)
      - ✅ Refactored `isTerminating` (12.9 → 5.7, 55.8% improvement; cyclomatic 8 → 4, 50.0% improvement)
      - ✅ Refactored `getTerminationReason` (12.9 → 5.7, 55.8% improvement; cyclomatic 8 → 4, 50.0% improvement)
      - ✅ Refactored `analyzeQuality` (12.9 → 1.3, 89.9% improvement; cyclomatic 8 → 1, 87.5% improvement)
      - ✅ Refactored `dfsCircular` (12.9 → 4.9, 62.0% improvement; cyclomatic 8 → 3, 62.5% improvement)
      - ✅ Refactored `runFileAnalysis` (12.7 → 4.4, 65.4% improvement; cyclomatic 9 → 3, 66.7% improvement)
      - ✅ Refactored `detectBuilder` (12.7 → 1.3, 89.8% improvement; cyclomatic 9 → 1, 88.9% improvement)
      - ✅ Refactored `writeNamingSection` (12.7 → 5.7, 55.1% improvement; cyclomatic 9 → 4, 55.6% improvement)
      - ✅ Refactored `Retrieve` (12.7 → 7.0, 44.9% improvement; cyclomatic 9 → 5, 44.4% improvement)
      - ✅ Refactored `checkIdentifierStuttering` (12.4 → 3.1, 75.0% improvement; cyclomatic 8 → 2, 75.0% improvement; extracted 4 helper functions: checkMethodStuttering, isAllowedMethodPrefix, checkPackageStuttering, isAllowedFunctionPrefix)
      - ✅ Refactored `parseCoverageLine` (12.2 → 4.4, 63.9% improvement; cyclomatic 9 → 3, 66.7% improvement; extracted 3 helper functions: extractCoverageFields, parseFileAndRange, recordCoverage)
      - ✅ Refactored `AnalyzeStructsWithPath` (12.1 → 3.1, 74.4% improvement; cyclomatic 7 → 2, 71.4% improvement; lines 20 → 7, 65% reduction; extracted 2 helper functions: processDeclaration at 4.4 complexity, processTypeSpec at 5.7 complexity)
      - ✅ Refactored `watchEventLoop` (11.9 → 4.9, 58.8% improvement; cyclomatic 8 → 3, 62.5% improvement; lines 18 → 5, 72.2% reduction; extracted 3 helper functions: processWatchEvent at 7.0 complexity, processFileSystemEvent at 3.1 complexity, processWatcherError at 3.1 complexity)
      - ⏳ 44 functions remaining above 10.0 threshold (25 completed, 44 remaining out of 69 total)
    - **Next Targets** (highest complexity first):
      1. ⏳ finalizeDuplicationMetrics (11.9) - cmd/analyze_finalize.go [NEXT]
      2. ⏳ detectStrategy (11.9) - internal/analyzer/pattern.go
      3. ⏳ AnalyzePackage (11.9) - internal/analyzer/package.go
    - **Deliverable**: `go-stats-generator analyze . --skip-tests` shows 0 functions above complexity 10.0
    - **Dependencies**: Steps 1-12
    - **Metric Justification**: ROADMAP.md Gate: "All functions ≤ 10 cyclomatic"
    - **Quality Score**: 50.0/100 (improving trend)

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
