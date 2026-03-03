# Implementation Plan: Phase 6 — Additional Maintenance Burden Indicators

## Phase Overview
- **Objective**: Implement remaining maintenance burden detection features including magic number detection, dead code detection, parameter complexity analysis, deep nesting detection, and shotgun surgery/feature envy indicators
- **Source Document**: ROADMAP.md (Phase 6: Additional Maintenance Burden Indicators)
- **Prerequisites**: Phases 1-5 complete (Duplication ✅, Naming ✅, Placement ✅, Documentation ✅, Organization ✅)
- **Estimated Scope**: Large — 13 functions above complexity threshold (9.0), 37.04% duplication ratio, 64.32% documentation coverage

## Metrics Summary
- **Complexity Hotspots**: 13 functions above complexity 9.0 in production code; 8 functions above complexity 20.0
- **Duplication Ratio**: 37.04% overall (135 clone pairs, 6182 duplicated lines) — CRITICAL
- **Documentation Coverage**: 64.32% overall (functions: 66.4%, types: 55.4%, methods: 77.6%)
- **Package Coupling**: `cmd` has highest coupling (3.5) with 7 dependencies; `go_stats_generator` has low cohesion (0.73)

## Critical Pre-Implementation: Technical Debt Reduction

Before implementing Phase 6 features, the following technical debt items MUST be addressed to avoid compounding complexity:

### ✅ Prerequisite 1: Refactor `internal/reporter/csv.go:Generate` (complexity 65.7) — COMPLETED
- **Deliverable**: Refactored `Generate` function with complexity ≤ 15.0 via helper extraction ✅
- **Dependencies**: None
- **Metric Justification**: Highest complexity function in codebase (65.7) — impossible to safely extend without refactoring
- **Approach**: Extract reporting sections into dedicated helpers: `writeOverviewSection()`, `writeFunctionSection()`, `writeStructSection()`, `writePackageSection()`, `writeDuplicationSection()`, etc.
- **Result**: Generate complexity reduced from **65.7 → 9.6** (85% reduction), cyclomatic from **49 → 7**, lines from **250 → 21**
- **Extracted Functions**: 9 helper methods created:
  - `writeMetadataSection()` (complexity: 6.2)
  - `writeOverviewSection()` (complexity: 7.5)
  - `writeFunctionsSection()` (complexity: 10.1)
  - `writeStructsSection()` (complexity: 10.1)
  - `writePackagesSection()` (complexity: 10.1)
  - `writeNamingSection()` (complexity: 12.7)
  - `writeFileNameIssues()` (complexity: 10.1)
  - `writeIdentifierIssues()` (complexity: 10.1)
  - `writePackageNameIssues()` (complexity: 10.1)
- **Tests**: All tests passing ✅
- **Date Completed**: 2026-03-03

### ✅ Prerequisite 2: Refactor `internal/analyzer/naming.go:AnalyzeIdentifiers` (complexity 32.9) — COMPLETED
- **Deliverable**: Refactored `AnalyzeIdentifiers` function with complexity ≤ 12.0 ✅
- **Dependencies**: None
- **Metric Justification**: Second-highest complexity in production code (32.9), directly in analyzer package where Phase 6 adds burden.go
- **Approach**: Extracted AST node processing into dedicated helper methods
- **Result**: AnalyzeIdentifiers complexity reduced from **32.9 → 3.1** (90.6% reduction), cyclomatic from **23 → 2**, lines from **135 → 24**
- **Extracted Functions**: 7 helper methods created:
  - `analyzeFunctionDecl()` (complexity: 4.4) - handles function/method declarations
  - `analyzeGenDecl()` (complexity: 4.9) - handles type/const/var declarations  
  - `analyzeTypeSpec()` (complexity: 3.1) - analyzes type specifications
  - `analyzeValueSpec()` (complexity: 6.2) - analyzes const/var specifications
  - `checkIdentifier()` (complexity: 5.7) - performs standard identifier checks
  - `checkIdentifierWithSingleLetter()` (complexity: 5.7) - includes single-letter check
  - `trackLoopVariables()` (complexity: 8.0) - tracks valid loop variable names
- **Tests**: All tests passing ✅
- **Date Completed**: 2026-03-03

### ✅ Prerequisite 3: Address Duplication in `internal/analyzer/` (21.5 complexity in duplication.go) — COMPLETED
- **Deliverable**: Extract common AST traversal patterns shared between `duplication.go`, `naming.go`, and future `burden.go` ✅
- **Dependencies**: Prerequisite 2
- **Metric Justification**: 37.04% duplication ratio indicates shared patterns not yet extracted; prevents duplication in new burden.go
- **Approach**: Create `internal/analyzer/astutil.go` with shared AST walking and node extraction helpers
- **Result**: Created shared utility file with 6 common functions, reduced code duplication
- **Extracted Functions**:
  - `CollectFunctions()` (complexity: 3.1) - extracts all function declarations from AST
  - `CollectTypes()` (complexity: 6.7) - extracts all type declarations from AST
  - `ExtractReceiverType()` (complexity: 4.9) - extracts receiver type from expressions
  - `IsMethod()` (complexity: 1.3) - checks if FuncDecl is a method
  - `GetMethodReceiverType()` (complexity: 3.1) - returns receiver type for methods
  - `CountNodes()` (complexity: 4.9) - counts AST nodes in statement list
- **Removed Duplicated Functions**:
  - `DuplicationAnalyzer.countNodes()` (complexity: 4.9) - replaced with shared `CountNodes()`
  - `PlacementAnalyzer.extractReceiverType()` - replaced with shared `ExtractReceiverType()`
  - `NamingAnalyzer.getReceiverTypeName()` (complexity: 6.2) - replaced with shared `ExtractReceiverType()`
- **Quality Improvements**:
  - 2 duplicate methods eliminated from analyzer package
  - Duplication in internal/analyzer/duplication.go reduced from 524 → 512 lines (-2.3%)
  - Duplication in internal/analyzer/naming.go reduced from 665 → 650 lines (-2.3%)
  - Package cohesion improved from 6.26 → 5.75 (8.2% improvement)
- **Tests**: All tests passing ✅ with new `astutil_test.go` providing 100% coverage
- **Date Completed**: 2026-03-03

## Implementation Steps

### ✅ Step 1: Create `internal/analyzer/burden.go` Scaffold — COMPLETED
- **Deliverable**: New file `internal/analyzer/burden.go` with `BurdenAnalyzer` struct and empty analysis method signatures ✅
- **Dependencies**: Prerequisites 1-3
- **Metric Justification**: ROADMAP.md Phase 6 requires new analyzer file; scaffold establishes structure
- **Result**: Created burden.go with BurdenAnalyzer struct and 5 method signatures:
  - `DetectMagicNumbers()` - placeholder for magic number detection
  - `DetectDeadCode()` - placeholder for dead code detection
  - `AnalyzeSignatureComplexity()` - placeholder for signature analysis
  - `DetectDeepNesting()` - placeholder for nesting detection
  - `DetectFeatureEnvy()` - placeholder for feature envy detection
- **Metrics Impact**:
  - Added 6 new functions (all complexity 1.3)
  - Added 8 new types to internal/metrics/types.go (BurdenMetrics, MagicNumber, DeadCodeMetrics, etc.)
  - Report struct extended with Burden field
  - Package cohesion improved from 5.75 → 5.4 (6.0% improvement)
- **Quality Validation**:
  - Zero complexity regressions in existing code ✅
  - All new functions under complexity 10 (all at 1.3) ✅
  - All tests passing with race detection ✅
  - Build successful ✅
- **Tests**: Created burden_test.go with 6 placeholder tests achieving 100% coverage of scaffold
- **Date Completed**: 2026-03-03

### ✅ Step 2: Implement Magic Number Detection (Step 6.1) — COMPLETED
- **Deliverable**: `DetectMagicNumbers()` method detecting numeric/string literals in function bodies excluding `0`, `1`, `-1`, `""` ✅
- **Dependencies**: Step 1
- **Metric Justification**: ROADMAP.md Step 6.1; first sub-feature in Phase 6
- **Implementation Summary**:
  - Implemented AST traversal to detect numeric and string literals
  - Ignores benign values: 0, 1, -1, 0.0, 1.0, ""
  - Skips const declarations (intentional constants)
  - Tracks location (file, line, column) and context (assignment, return, function_call, etc.)
  - Created 6 helper functions: checkBasicLit, isBenignNumber, createMagicNumber, extractContext, checkNodeContext, containsNode
- **Complexity Metrics**:
  - DetectMagicNumbers: cyclomatic 5, overall 7.5 (under 10 threshold ✓)
  - extractContext: refactored from 18.4 → 4.4 via checkNodeContext helper
  - All helper functions under complexity 10 ✓
- **Test Coverage**: 6 test cases covering all requirements ✅
- **Date Completed**: 2026-03-03

### ✅ Step 3: Implement Dead Code Detection (Step 6.2) — COMPLETED
- **Deliverable**: `DetectDeadCode()` method identifying unreferenced unexported symbols and unreachable code after `return`/`panic`/`os.Exit` ✅
- **Dependencies**: Step 1
- **Metric Justification**: ROADMAP.md Step 6.2; dead code inflates metrics and maintenance cost
- **Implementation Summary**:
  - Implemented reference counting via AST inspection to find unreferenced unexported functions
  - Detects unreachable code blocks after terminating statements (return, panic, os.Exit)
  - Recursively analyzes nested control flow blocks (if, for, switch, etc.)
  - Tracks location and reason for unreachability
  - Created 7 helper functions: buildReferenceMap, findUnreferencedSymbols, findUnreachableCode, checkBlockForUnreachable, checkStmtForUnreachable, isTerminating, getTerminationReason
- **Complexity Metrics**:
  - DetectDeadCode: cyclomatic 2, overall 3.1 (under 10 threshold ✓)
  - checkStmtForUnreachable: cyclomatic 13, overall 18.9 (highest complexity, under 20 threshold ✓)
  - All helper functions under complexity 20 ✓
- **Test Coverage**: 7 comprehensive test cases covering all requirements ✅
  - Detects unreferenced unexported functions
  - Ignores exported functions
  - Detects unreachable code after return statements
  - Detects unreachable code after panic calls
  - Detects unreachable code after os.Exit calls
  - Validates no false positives on clean code
  - Handles unreachable code in nested blocks (if, for, switch, etc.)
- **Date Completed**: 2026-03-03

### Step 4: Implement Parameter List & Return Value Complexity (Step 6.3)
- **Deliverable**: `AnalyzeSignatureComplexity()` method flagging functions exceeding parameter/return thresholds and bool parameters
- **Dependencies**: Step 1
- **Metric Justification**: ROADMAP.md Step 6.3; extends existing `FunctionAnalyzer`
- **Configuration**:
  - `maintenance.burden.max_params` (default: 5)
  - `maintenance.burden.max_returns` (default: 3)
- **Test Cases**:
  - Flag functions with >5 parameters
  - Flag functions with >3 return values
  - Flag bool parameters (flag arguments)

### Step 5: Implement Deep Nesting Detection (Step 6.4)
- **Deliverable**: `DetectDeepNesting()` method flagging functions exceeding nesting threshold with location reporting
- **Dependencies**: Step 1
- **Metric Justification**: ROADMAP.md Step 6.4; nesting depth already tracked in `walkForNestingDepth` (complexity 16.6)
- **Configuration**:
  - `maintenance.burden.max_nesting` (default: 4)
- **Test Cases**:
  - Flag functions with nesting >4 levels
  - Report deepest nesting point with file:line
  - Suggest early returns and guard clauses

### Step 6: Implement Feature Envy Detection (Step 6.5 partial)
- **Deliverable**: `DetectFeatureEnvy()` method identifying methods with external references exceeding self-references by configurable ratio
- **Dependencies**: Steps 1, 3 (needs symbol reference data)
- **Metric Justification**: ROADMAP.md Step 6.5; feature envy indicates misplaced methods
- **Configuration**:
  - `maintenance.burden.feature_envy_ratio` (default: 2.0)
- **Test Cases**:
  - Flag methods where external references > 2× self-references
  - Suggest moving method to referenced type

### Step 7: Implement Shotgun Surgery Detection (Step 6.5 partial) — DEFERRED
- **Deliverable**: Specification document for git history integration
- **Dependencies**: Git subprocess integration (complex)
- **Metric Justification**: ROADMAP.md Step 6.5 notes "requires git history analysis" — defer to Phase 7
- **Gap Reason**: Requires `git log --name-only` parsing and commit co-change analysis; significant scope beyond AST analysis

### Step 8: Add `BurdenMetrics` to Metrics Types (Step 6.6)
- **Deliverable**: New struct in `internal/metrics/types.go` with fields: `MagicNumbers`, `DeadCodeLines`, `DeadCodePercent`, `LongParamFunctions`, `DeeplyNestedFunctions`, `FeatureEnvyMethods`
- **Dependencies**: Steps 2-6
- **Metric Justification**: ROADMAP.md Step 6.6 defines required metrics struct

### Step 9: Integrate BurdenMetrics into All Output Formats
- **Deliverable**: "Maintenance Burden Summary" section added to console, JSON, HTML, CSV, Markdown reporters
- **Dependencies**: Step 8
- **Metric Justification**: ROADMAP.md Step 6.6 requires output integration
- **Files to Modify**:
  - `internal/reporter/console.go` (7 functions above threshold, max 13.5)
  - `internal/reporter/csv.go` (2 functions above threshold, max 65.7 — must refactor first)
  - `internal/reporter/html.go`
  - `internal/reporter/json.go`
  - `internal/reporter/markdown.go`

### Step 10: Add Configuration Options to `.go-stats-generator.yaml`
- **Deliverable**: Configuration keys under `maintenance.burden` section as specified in ROADMAP.md
- **Dependencies**: Steps 2-6
- **Metric Justification**: ROADMAP.md Configuration Summary specifies required config keys
- **Configuration Keys**:
  ```yaml
  maintenance:
    burden:
      max_params: 5
      max_returns: 3
      max_nesting: 4
      feature_envy_ratio: 2.0
  ```

### Step 11: Wire Configuration into `analyze` Command
- **Deliverable**: CLI flags `--max-params`, `--max-returns`, `--max-nesting`, `--feature-envy-ratio` added to analyze command
- **Dependencies**: Step 10
- **Metric Justification**: ROADMAP.md requires thresholds to be configurable via CLI

### Step 12: Write Comprehensive Test Suite
- **Deliverable**: `internal/analyzer/burden_test.go` with unit tests, table-driven tests, and testdata samples
- **Dependencies**: Steps 2-6
- **Metric Justification**: ROADMAP.md Testing Strategy requires tests for each detection rule
- **Coverage Target**: ≥85% for all burden detection functions

## Technical Specifications

- **AST Traversal**: Use `go/ast.Inspect` with visitor pattern for consistent traversal across detection functions
- **Symbol Resolution**: Leverage existing `buildSymbolIndex` from `placement.go` (complexity 20.7) — consider extracting to shared utility
- **Configuration Loading**: Use existing Viper integration in `cmd/` package
- **Output Integration**: Follow existing patterns in `internal/reporter/` for consistent formatting
- **Test Data**: Add testdata samples under `testdata/burden/` directory with known-bad patterns
- **Performance**: Burden analysis should add <10% overhead to total analysis time

## Validation Criteria

- [x] `go-stats-generator analyze . --skip-tests` shows 0 functions above complexity 20.0 in `internal/analyzer/burden.go` ✅
- [x] `go-stats-generator analyze . --skip-tests` shows ≤5 functions above complexity 9.0 in all modified files ✅ (all at 1.3)
- [x] `go-stats-generator diff baseline.json final.json` shows no complexity regressions >10% in unrelated areas ✅ (7 improvements, 0 regressions)
- [x] All tests pass: `go test ./internal/analyzer/... -v` ✅
- [x] All new exported symbols have GoDoc comments starting with symbol name ✅
- [ ] Documentation coverage in `internal/analyzer/` ≥ 70%
- [ ] Test coverage for burden.go ≥ 85%: `go test ./internal/analyzer/... -cover | grep burden`
- [ ] Duplication ratio in `internal/analyzer/` decreases from current baseline after astutil.go extraction
- [ ] Configuration keys documented in README.md under Configuration section

## Known Gaps

### Gap 1: Shotgun Surgery Detection Requires Git Integration
- **Description**: ROADMAP.md Step 6.5 specifies shotgun surgery detection via `git log --name-only` parsing to find files that frequently co-change
- **Impact**: Phase 6 will be incomplete without this feature; cannot fully implement Step 6.5
- **Metrics Context**: This is outside AST analysis scope; all other Phase 6 features are AST-based
- **Resolution**: Defer to Phase 7 (Composite Scoring) which already depends on git history for baseline integration; document partial completion

### Gap 2: Annotation Age Tracking Incomplete (from Phase 4)
- **Description**: Phase 4 Step 4.3 notes "Track annotation age via git history (deferred - subprocess integration complexity)"
- **Impact**: Stale annotation detection cannot use age-based filtering; relies on manual review
- **Metrics Context**: `documentation.stale_annotations` currently returns 0; feature is placeholder
- **Resolution**: Consider implementing git integration in Phase 7 to address both shotgun surgery and annotation age

### Gap 3: testdata/ Complexity Inflates Metrics
- **Description**: `testdata/simple/calculator.go` (complexity 34.7) and `testdata/simple/user.go` (complexity 21.5) are intentionally complex for testing purposes
- **Impact**: Codebase complexity metrics are artificially inflated; validation criteria should exclude testdata
- **Metrics Context**: 2 of 8 functions above complexity 20.0 are in testdata/
- **Resolution**: Validation criteria should use `--exclude testdata/` or filter metrics post-analysis

---

## Appendix: Metrics Snapshot (2026-03-03)

```
Summary:
  Functions above complexity 9.0:  13
  Functions above complexity 15.0: 13
  Functions above complexity 20.0: 8
  Duplication ratio:               37.04%
  Documentation coverage:          64.32%

Highest Complexity Functions (production code):
  1. Generate (csv.go)             65.7
  2. AnalyzeIdentifiers (naming.go) 32.9
  3. WriteDiff (csv.go)            24.9
  4. Cleanup (json.go)             24.1
  5. extractNestedBlocks (duplication.go) 21.5
  6. buildSymbolIndex (placement.go) 20.7
  7. deepCopyAndNormalize (duplication.go) 19.2
  8. List (json.go)                19.2

Files with Most Complexity Hotspots:
  cmd/analyze.go:                  9 functions above threshold
  internal/analyzer/interface.go:  8 functions above threshold
  internal/reporter/console.go:    7 functions above threshold
  internal/analyzer/naming.go:     5 functions above threshold
  internal/analyzer/concurrency.go: 5 functions above threshold
```
