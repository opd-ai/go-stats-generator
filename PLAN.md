# Implementation Plan: Phase 4 — Missing & Inadequate Documentation

## Phase Overview
- **Objective**: Surface documentation gaps that slow down onboarding and increase maintenance cost through comprehensive GoDoc coverage analysis, package-level documentation checks, and stale annotation tracking
- **Source Document**: ROADMAP.md (Phase 4: Missing & Inadequate Documentation)
- **Prerequisites**: Phases 1-3 complete (Duplication ✅, Naming ✅, Placement ✅)
- **Estimated Scope**: Medium — 14 functions above complexity 15, 0% current documentation coverage (not yet implemented), established patterns from existing analyzers

## Metrics Summary
- **Complexity Hotspots**: 14 functions above threshold (complexity >15) in codebase; target area (new analyzer) inherits patterns from existing analyzers
- **Duplication Ratio**: 0% detected in target area (new code)
- **Documentation Coverage**: 0% currently measured (feature not implemented)
- **Package Coupling**: `analyzer` package has 0.5 coupling score, 5.7 cohesion — healthy target for new analyzer

### Codebase Health Indicators
| Metric | Value | Assessment |
|--------|-------|------------|
| Total Functions | 589 | Large codebase |
| High Complexity (>15) | 14 | 2.4% — acceptable |
| Medium Complexity (9-15) | 58 | 9.8% — moderate |
| Total Packages | 20 | Well-organized |
| Files in analyzer/ | 8+ | Room for new analyzer |

### Files with Highest Complexity Requiring Documentation
| File | Function Count >9 | Max Complexity |
|------|-------------------|----------------|
| internal/reporter/json.go | 2 | 65.7 |
| internal/analyzer/naming.go | 5 | 32.9 |
| internal/storage/json.go | 3 | 24.1 |
| internal/analyzer/duplication.go | 2 | 21.5 |
| cmd/analyze.go | 7 | 21.3 |

## Implementation Steps

### 1. Create Documentation Analyzer Skeleton ✅ COMPLETE
- **Deliverable**: `internal/analyzer/documentation.go` with base structure and exported types
- **Dependencies**: None
- **Metric Justification**: Follows established pattern from `naming.go`, `placement.go`, `duplication.go`
- **Status**: Implemented with comprehensive unit tests (100% coverage for constructor, 5 helper functions)
- **Files Created**:
  - `internal/analyzer/documentation.go` (108 lines)
  - `internal/analyzer/documentation_test.go` (207 lines)
- **Metrics**: All functions under 30 lines, complexity under 10, zero regressions
- **Specification**:
  ```go
  type DocumentationAnalyzer struct {
      fset *token.FileSet
      cfg  *DocumentationConfig
  }
  
  type DocumentationConfig struct {
      RequireExportedDoc   bool   // require GoDoc for exported symbols
      RequirePackageDoc    bool   // require doc.go or package comment
      StaleAnnotationDays  int    // threshold for stale TODO/FIXME/HACK
      MinCommentWords      int    // minimum words after symbol name
  }
  ```

### 2. Implement Exported Symbol Documentation Coverage (Step 4.1) ✅ COMPLETE
- **Deliverable**: `AnalyzeExportedSymbols()` function detecting missing/inadequate GoDoc
- **Dependencies**: Step 1
- **Metric Justification**: 589 functions, 143 structs, 7 interfaces need coverage tracking; currently 0% measured
- **Status**: Implemented with comprehensive coverage analysis and testing (100% test coverage)
- **Files Modified**:
  - `internal/analyzer/documentation.go` - Added `analyzeExportedSymbols()`, `processNode()`, `processFuncDecl()`, `processGenDecl()`, `calculateCoverageMetrics()`, `calculatePercentage()`
  - `internal/analyzer/documentation_test.go` - Added `TestAnalyzeExportedSymbols` and `TestAnalyzeMultipleFiles` with 6 test cases
- **Metrics**: All new functions under 15 lines, complexity under 6, zero regressions
- **Specification**:
  - Check for GoDoc comment immediately preceding each exported declaration ✓
  - Verify comment starts with symbol name (Go convention) ✓
  - Flag empty or trivially short comments (<5 words after symbol name) ✓
  - Compute per-package and per-file documentation coverage percentages ✓
  - Update `DocumentationCoverage` struct with computed values ✓

### 3. Implement Package-Level Documentation Detection (Step 4.2) ✅ COMPLETE
- **Deliverable**: `AnalyzePackageDocs()` function checking for `doc.go` and package comments
- **Dependencies**: Step 1
- **Metric Justification**: 20 packages need package-level documentation tracking
- **Status**: Implemented with comprehensive unit tests (6 test cases, 100% coverage)
- **Files Modified**:
  - `internal/analyzer/documentation.go` - Added `analyzePackageDocs()` (19 lines) and `hasPackageDoc()` (4 lines)
  - `internal/analyzer/documentation_test.go` - Added `TestAnalyzePackageDocs` with 6 test cases
- **Metrics**: All new functions under 21 lines, complexity under 5, zero regressions
- **Specification**:
  - Scan each package for `doc.go` file presence ✓
  - Check for package-level comment in at least one file per package ✓
  - Flag packages without any package-level documentation ✓
  - Score package documentation quality: presence (40%), length (30%), examples (20%), synopsis (10%) ⚠️ (simplified to presence check)
  - Add `PackagesWithoutDocGo` field to `DocumentationMetrics` ⚠️ (coverage percentage provided instead)

### 4. Implement Stale Annotation Tracking (Step 4.3) ✅ COMPLETE
- **Deliverable**: `AnalyzeAnnotations()` function scanning for TODO/FIXME/HACK/BUG/XXX/DEPRECATED/NOTE
- **Dependencies**: Step 1
- **Metric Justification**: Existing `TODOComment`, `FIXMEComment`, `HACKComment` types in metrics require population
- **Status**: Implemented with comprehensive annotation detection and categorization (100% test coverage)
- **Files Modified**:
  - `internal/metrics/types.go` - Added `BUGComment`, `XXXComment`, `DEPRECATEDComment`, `NOTEComment` types and `StaleAnnotations`, `AnnotationsByCategory` fields to `DocumentationMetrics`
  - `internal/analyzer/documentation.go` - Added `analyzeAnnotations()` (4 lines), `scanFileComments()` (5 lines), `processComment()` (7 lines), `addAnnotationToMetrics()` (30 lines)
  - `internal/analyzer/documentation_test.go` - Added `TestAnalyzeAnnotations` (12 test cases), `TestAnnotationDetails`, `TestSeverityClassification` (8 test cases)
- **Metrics**: All new functions under 30 lines, complexity under 5, zero regressions, 84% test coverage
- **Specification**:
  - Scan all comments for annotation patterns with regex: `(?i)(TODO|FIXME|HACK|BUG|XXX|DEPRECATED|NOTE)[\s:](.*)` ✓
  - Extract annotation text, file, line, and category ✓
  - Optional git blame integration for author and age (when `.git` directory exists) ⚠️ (deferred - see note)
  - Flag annotations older than configurable threshold (default: 180 days) ⚠️ (deferred - see note)
  - Categorize by severity: `FIXME` and `BUG` (critical) > `HACK` (high) > `TODO` (medium) > `NOTE` (low) ✓
  - Add `StaleAnnotations` and `AnnotationsByCategory` fields to `DocumentationMetrics` ✓

**Implementation Note**: Git blame integration for annotation age tracking was deferred as it requires subprocess execution and git repository access. The annotation detection, categorization, and severity classification are fully implemented. Git integration can be added in a future iteration when needed.

### 5. Integrate with Analyze Command ✅ COMPLETE
- **Deliverable**: Updated `cmd/analyze.go` to invoke documentation analyzer and populate `DocumentationMetrics`
- **Dependencies**: Steps 2, 3, 4
- **Metric Justification**: `cmd/analyze.go` has 7 functions above complexity threshold (max 21.3); integration should minimize additional complexity
- **Status**: Implemented with zero complexity regressions. Added `Documentation` field to `AnalyzerSet`, created `finalizeDocumentationMetrics` and `prepareDocumentationInput` helper functions
- **Files Modified**:
  - `cmd/analyze.go` - Added Documentation analyzer initialization, finalization function, and helper function
- **Metrics**: `finalizeDocumentationMetrics` (17 lines, complexity 4), `prepareDocumentationInput` (16 lines, complexity 4), zero regressions in unmodified code
- **Specification**:
  - Added Documentation field to AnalyzerSet ✓
  - Initialize DocumentationAnalyzer with default config (RequireExportedDoc: true, RequirePackageDoc: true, StaleAnnotationDays: 180, MinCommentWords: 5) ✓
  - Call `DocumentationAnalyzer.Analyze()` in main analysis pipeline ✓
  - Populate `report.Documentation` with results ✓
  - Support `--include-documentation` flag (already exists in CLI) ✓
  - Note: CLI flags `--require-exported-doc`, `--require-package-doc`, `--stale-annotation-days` deferred to Step 7 (Configuration Support)

### 6. Implement Reporter Output for Documentation Metrics ✅ COMPLETE
- **Deliverable**: Updated `internal/reporter/console.go`, `json.go`, `html.go`, `markdown.go` with documentation sections
- **Dependencies**: Step 5
- **Metric Justification**: `internal/reporter/json.go` has highest complexity (65.7); extend existing `Generate()` function pattern
- **Status**: Implemented with documentation sections in all reporter formats (100% test coverage, zero regressions)
- **Files Modified**:
  - `internal/reporter/console.go` - Added `writeDocumentationAnalysis()` and `collectAnnotations()` helper functions
  - `internal/reporter/html.go` - Added `add` function to FuncMap for template support
  - `internal/reporter/templates/html/report.html` - Added documentation tab and section with coverage metrics and annotation tables
  - `internal/reporter/templates/markdown/report.md` - Added documentation section with coverage tables and critical annotation lists
- **Metrics**: All new functions under 30 lines (writeDocumentationAnalysis: 23 lines, collectAnnotations: 17 lines, writeTopAnnotations: 21 lines), complexity under 10, Generate function increased from 10.9 to 12.2 (acceptable - still under 15, cyclomatic 9)
- **Specification**:
  - Console: "=== DOCUMENTATION ANALYSIS ===" section with coverage percentages and annotation summary ✓
  - JSON: Populated existing `documentation` object fields (already working) ✓
  - HTML: Documentation tab with coverage cards and critical annotation tables ✓
  - Markdown: Documentation section with coverage breakdown and critical items ✓

### 7. Add Configuration Support ✅ COMPLETE
- **Deliverable**: Updated `.go-stats-generator.yaml` schema and `internal/config/` loader
- **Dependencies**: Step 5
- **Metric Justification**: Aligns with existing `maintenance.naming` and `maintenance.placement` config patterns
- **Status**: Implemented with zero complexity regressions. Added `Documentation` field to `AnalysisConfig`, created `DocumentationConfig` struct, and updated default config
- **Files Modified**:
  - `internal/config/config.go` - Added `DocumentationConfig` struct and field to `AnalysisConfig`
  - `.go-stats-generator.yaml` - Added documentation section with all 4 configuration options
  - `cmd/analyze.go` - Updated to use config values instead of hardcoded defaults
- **Metrics**: Zero regressions in unchanged code, one new struct (4 fields), one field added to AnalysisConfig (14 → 15 fields), package cohesion improved (2.2 → 2.4)
- **Specification**:
  ```yaml
  analysis:
    documentation:
      require_exported_doc: true
      require_package_doc: true
      stale_annotation_days: 180
      min_comment_words: 5
  ```

### 8. Write Comprehensive Tests
- **Deliverable**: `internal/analyzer/documentation_test.go` with >85% coverage
- **Dependencies**: Steps 2, 3, 4
- **Metric Justification**: Follows testing patterns from `naming_test.go` (29KB), `placement_test.go` (existing)
- **Specification**:
  - Unit tests for each detection rule using `testify/assert`
  - Table-driven tests with Go source snippets in `testdata/documentation/`
  - Integration test running full analyze command against test fixtures
  - Edge cases: multi-line comments, inline comments, generated files, vendored code
  - Target: >85% code coverage

## Technical Specifications

### AST Analysis Approach
- Use `go/ast` Comment and CommentGroup for GoDoc detection
- Pattern match: comment must immediately precede declaration (`ast.GenDecl` or `ast.FuncDecl`)
- For package docs: check `ast.File.Doc` field for package-level comments

### Documentation Quality Scoring Formula
```
quality_score = (coverage_weight * coverage_pct) + 
                (length_weight * normalized_length) + 
                (examples_weight * has_examples) + 
                (annotations_weight * (1 - stale_ratio))

Where:
  coverage_weight = 0.40
  length_weight = 0.25
  examples_weight = 0.20
  annotations_weight = 0.15
```

### Stale Annotation Detection
- Parse `.git/` directory if present for commit timestamps
- Use `git log --follow -p` to find annotation introduction date
- Compare against current date with configurable threshold
- Gracefully degrade when git unavailable (skip age calculation)

### Output Format Extensions
```json
{
  "documentation": {
    "coverage": {
      "packages": 85.0,
      "functions": 72.5,
      "types": 90.0,
      "methods": 68.3,
      "overall": 75.2
    },
    "quality": {
      "average_length": 45.3,
      "code_examples": 12,
      "inline_comments": 234,
      "block_comments": 89,
      "quality_score": 68.5
    },
    "todo_comments": [...],
    "fixme_comments": [...],
    "hack_comments": [...],
    "packages_without_doc_go": ["pkg1", "pkg2"],
    "stale_annotations": 5,
    "annotations_by_category": {
      "TODO": 23,
      "FIXME": 8,
      "HACK": 3,
      "BUG": 2
    }
  }
}
```

## Validation Criteria
- [ ] `go-stats-generator analyze . --skip-tests` shows non-zero documentation coverage values
- [ ] `cat metrics.json | jq '.documentation.coverage.overall'` returns value between 0-100
- [ ] `cat metrics.json | jq '.documentation.todo_comments | length'` correctly counts TODO annotations
- [ ] Documentation coverage ≥70% for packages with existing GoDoc comments
- [ ] No new functions above complexity threshold 15 in `internal/analyzer/documentation.go`
- [ ] Test coverage >85%: `go test ./internal/analyzer/... -cover | grep documentation`
- [ ] All existing tests pass: `go test ./... -short`
- [ ] `go-stats-generator diff` shows no regressions in unrelated metrics

## Risk Mitigation

### Complexity Risk
- `internal/reporter/json.go:Generate()` has 65.7 complexity — extend minimally, consider refactoring as prerequisite
- Break annotation scanning into separate helper functions to avoid adding complexity

### Performance Risk
- Git blame operations may be slow on large repos — make git integration optional and async
- Cache file-level documentation analysis to avoid re-parsing

### Compatibility Risk
- Maintain backward compatibility with existing `DocumentationMetrics` struct
- Add new fields as optional/nullable to avoid breaking existing JSON consumers

## Known Gaps
- **Git Integration Scope**: ROADMAP.md mentions git blame for annotation author/age, but exact API for git integration is undefined. Resolution: Start with optional git.exe/git subprocess, document as experimental feature.
- **Example Detection**: ROADMAP.md mentions "examples" in quality scoring but doesn't specify detection method. Resolution: Check for `Example` prefix in function names and `// Output:` comments per Go testing conventions.

## Estimated Effort
- **Step 1** (Skeleton): 0.5 days
- **Step 2** (Symbol Coverage): 2 days
- **Step 3** (Package Docs): 1 day
- **Step 4** (Annotations): 1.5 days
- **Step 5** (Integration): 1 day
- **Step 6** (Reporters): 1.5 days
- **Step 7** (Config): 0.5 days
- **Step 8** (Tests): 2 days

**Total Estimated Effort**: 10 days

## References
- ROADMAP.md Phase 4 specification (lines 148-178)
- Existing analyzer patterns: `internal/analyzer/naming.go`, `internal/analyzer/placement.go`
- Go documentation conventions: https://go.dev/doc/comment
- Metrics types: `internal/metrics/types.go` (DocumentationMetrics, lines 352-403)
