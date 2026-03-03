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

### 2. Implement Exported Symbol Documentation Coverage (Step 4.1)
- **Deliverable**: `AnalyzeExportedSymbols()` function detecting missing/inadequate GoDoc
- **Dependencies**: Step 1
- **Metric Justification**: 589 functions, 143 structs, 7 interfaces need coverage tracking; currently 0% measured
- **Specification**:
  - Check for GoDoc comment immediately preceding each exported declaration
  - Verify comment starts with symbol name (Go convention)
  - Flag empty or trivially short comments (<5 words after symbol name)
  - Compute per-package and per-file documentation coverage percentages
  - Update `DocumentationCoverage` struct with computed values

### 3. Implement Package-Level Documentation Detection (Step 4.2)
- **Deliverable**: `AnalyzePackageDocs()` function checking for `doc.go` and package comments
- **Dependencies**: Step 1
- **Metric Justification**: 20 packages need package-level documentation tracking
- **Specification**:
  - Scan each package for `doc.go` file presence
  - Check for package-level comment in at least one file per package
  - Flag packages without any package-level documentation
  - Score package documentation quality: presence (40%), length (30%), examples (20%), synopsis (10%)
  - Add `PackagesWithoutDocGo` field to `DocumentationMetrics`

### 4. Implement Stale Annotation Tracking (Step 4.3)
- **Deliverable**: `AnalyzeAnnotations()` function scanning for TODO/FIXME/HACK/BUG/XXX/DEPRECATED/NOTE
- **Dependencies**: Step 1
- **Metric Justification**: Existing `TODOComment`, `FIXMEComment`, `HACKComment` types in metrics require population
- **Specification**:
  - Scan all comments for annotation patterns with regex: `(?i)(TODO|FIXME|HACK|BUG|XXX|DEPRECATED|NOTE)[\s:](.*)`
  - Extract annotation text, file, line, and category
  - Optional git blame integration for author and age (when `.git` directory exists)
  - Flag annotations older than configurable threshold (default: 180 days)
  - Categorize by severity: `FIXME` and `BUG` (critical) > `HACK` (high) > `TODO` (medium) > `NOTE` (low)
  - Add `StaleAnnotations` and `AnnotationsByCategory` fields to `DocumentationMetrics`

### 5. Integrate with Analyze Command
- **Deliverable**: Updated `cmd/analyze.go` to invoke documentation analyzer and populate `DocumentationMetrics`
- **Dependencies**: Steps 2, 3, 4
- **Metric Justification**: `cmd/analyze.go` has 7 functions above complexity threshold (max 21.3); integration should minimize additional complexity
- **Specification**:
  - Add `--require-exported-doc` flag (default: true)
  - Add `--require-package-doc` flag (default: true)
  - Add `--stale-annotation-days` flag (default: 180)
  - Call `DocumentationAnalyzer.Analyze()` in main analysis pipeline
  - Populate `report.Documentation` with results

### 6. Implement Reporter Output for Documentation Metrics
- **Deliverable**: Updated `internal/reporter/console.go`, `json.go`, `html.go`, `markdown.go` with documentation sections
- **Dependencies**: Step 5
- **Metric Justification**: `internal/reporter/json.go` has highest complexity (65.7); extend existing `Generate()` function pattern
- **Specification**:
  - Console: Add "=== DOCUMENTATION ANALYSIS ===" section with coverage percentages and annotation summary
  - JSON: Populate existing `documentation` object fields
  - HTML: Add documentation coverage tables and annotation lists
  - Markdown: Add documentation section with coverage breakdown

### 7. Add Configuration Support
- **Deliverable**: Updated `.go-stats-generator.yaml` schema and `internal/config/` loader
- **Dependencies**: Step 5
- **Metric Justification**: Aligns with existing `maintenance.naming` and `maintenance.placement` config patterns
- **Specification**:
  ```yaml
  maintenance:
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
