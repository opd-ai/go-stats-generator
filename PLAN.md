# Implementation Plan: Phase 6 — Complete Maintenance Burden Indicators Integration

## Phase Overview
- **Objective**: Complete integration of maintenance burden detection features (Steps 6.2-6.6) into the analyze command and JSON output
- **Source Document**: ROADMAP.md (Phase 6: Additional Maintenance Burden Indicators)
- **Prerequisites**: Phases 1-5 complete; Step 6.1 (Magic Numbers) complete
- **Estimated Scope**: Medium — 6 functions above complexity threshold in target area, 34.82% overall duplication, 65.75% doc coverage

## Metrics Summary
- **Complexity Hotspots**: 6 functions above threshold in `internal/analyzer/burden.go`
  - `walkForNestingDepth` (19.2), `checkStmtForUnreachable` (18.9), `checkNodeContext` (13.2), `isTerminating` (12.9), `getTerminationReason` (12.9), `countParameters` (9.3)
- **Duplication Ratio**: 34.82% overall (Critical — exceeds 10% threshold)
- **Documentation Coverage**: 65.75% overall (Medium priority)
- **Package Coupling**: `cmd` has coupling score 3.5 (7 dependencies), `internal/analyzer` has 0.5 (acceptable)

## Current Gap Analysis

The `BurdenAnalyzer` in `internal/analyzer/burden.go` (703 lines) implements the following detection methods:
- ✅ `DetectMagicNumbers()` — Integrated in `cmd/analyze.go`
- ⚠️ `DetectDeadCode()` — Implemented but NOT integrated
- ⚠️ `AnalyzeSignatureComplexity()` — Implemented but NOT integrated
- ⚠️ `DetectDeepNesting()` — Implemented but NOT integrated
- ⚠️ `DetectFeatureEnvy()` — Implemented but NOT integrated

**JSON output currently shows `"burden": null`** — only `MagicNumbers` is being populated.

## Implementation Steps

### 1. Integrate Dead Code Detection (Step 6.2)
- **Deliverable**: `analyzeBurdenInFile()` in `cmd/analyze.go` calls `DetectDeadCode()` and populates `report.Burden.DeadCode`
- **Dependencies**: None (analyzer already implemented)
- **Metric Justification**: ROADMAP.md Step 6.2 incomplete; dead code contributes to 34.82% duplication burden
- **Implementation**:
  ```go
  // In analyzeBurdenInFile(), add after magicNumbers detection:
  deadCode := burdenAnalyzer.DetectDeadCode([]*ast.File{result.File}, result.FileInfo.Package)
  if deadCode != nil {
      report.Burden.DeadCode.UnreferencedFunctions = append(report.Burden.DeadCode.UnreferencedFunctions, deadCode.UnreferencedFunctions...)
      report.Burden.DeadCode.UnreachableCode = append(report.Burden.DeadCode.UnreachableCode, deadCode.UnreachableCode...)
      report.Burden.DeadCode.TotalDeadLines += deadCode.TotalDeadLines
  }
  ```

### 2. Integrate Signature Complexity Analysis (Step 6.3)
- **Deliverable**: `analyzeBurdenInFile()` calls `AnalyzeSignatureComplexity()` for each function and populates `report.Burden.ComplexSignatures`
- **Dependencies**: None (analyzer already implemented)
- **Metric Justification**: ROADMAP.md Step 6.3 incomplete; parameter/return complexity not visible in output
- **Implementation**:
  ```go
  // After parsing functions, for each FuncDecl:
  sigIssue := burdenAnalyzer.AnalyzeSignatureComplexity(fn, cfg.Analysis.Burden.MaxParams, cfg.Analysis.Burden.MaxReturns)
  if sigIssue != nil {
      report.Burden.ComplexSignatures = append(report.Burden.ComplexSignatures, *sigIssue)
  }
  ```

### 3. Integrate Deep Nesting Detection (Step 6.4)
- **Deliverable**: `analyzeBurdenInFile()` calls `DetectDeepNesting()` for each function and populates `report.Burden.DeeplyNestedFunctions`
- **Dependencies**: None (analyzer already implemented)
- **Metric Justification**: ROADMAP.md Step 6.4 incomplete; `walkForNestingDepth` has complexity 19.2 — already implemented but unused
- **Implementation**:
  ```go
  // After parsing functions, for each FuncDecl:
  nestingIssue := burdenAnalyzer.DetectDeepNesting(fn, cfg.Analysis.Burden.MaxNesting)
  if nestingIssue != nil {
      report.Burden.DeeplyNestedFunctions = append(report.Burden.DeeplyNestedFunctions, *nestingIssue)
  }
  ```

### 4. Integrate Feature Envy Detection (Step 6.5)
- **Deliverable**: `analyzeBurdenInFile()` calls `DetectFeatureEnvy()` for each method and populates `report.Burden.FeatureEnvyMethods`
- **Dependencies**: None (analyzer already implemented)
- **Metric Justification**: ROADMAP.md Step 6.5 incomplete; feature envy detection implemented but not called
- **Implementation**:
  ```go
  // For methods (FuncDecl with Recv):
  envyIssue := burdenAnalyzer.DetectFeatureEnvy(fn, file, cfg.Analysis.Burden.FeatureEnvyRatio)
  if envyIssue != nil {
      report.Burden.FeatureEnvyMethods = append(report.Burden.FeatureEnvyMethods, *envyIssue)
  }
  ```

### 5. Refactor `analyzeBurdenInFile()` for Function-Level Analysis
- **Deliverable**: Modified `analyzeBurdenInFile()` that iterates over functions in the file to apply signature, nesting, and feature envy analysis
- **Dependencies**: Steps 2, 3, 4
- **Metric Justification**: Current implementation only processes file-level magic numbers; function-level analyzers require AST traversal
- **Implementation**:
  ```go
  func analyzeBurdenInFile(burdenAnalyzer *analyzer.BurdenAnalyzer, result scanner.Result, report *metrics.Report, cfg *config.Config) error {
      // Existing magic number detection
      magicNumbers := burdenAnalyzer.DetectMagicNumbers(result.File, result.FileInfo.Package)
      report.Burden.MagicNumbers = append(report.Burden.MagicNumbers, magicNumbers...)

      // Dead code detection (file-level)
      deadCode := burdenAnalyzer.DetectDeadCode([]*ast.File{result.File}, result.FileInfo.Package)
      if deadCode != nil {
          // merge results...
      }

      // Function-level analysis
      ast.Inspect(result.File, func(n ast.Node) bool {
          fn, ok := n.(*ast.FuncDecl)
          if !ok || fn.Body == nil {
              return true
          }

          // Signature complexity
          if sigIssue := burdenAnalyzer.AnalyzeSignatureComplexity(fn, cfg.Analysis.Burden.MaxParams, cfg.Analysis.Burden.MaxReturns); sigIssue != nil {
              report.Burden.ComplexSignatures = append(report.Burden.ComplexSignatures, *sigIssue)
          }

          // Deep nesting
          if nestIssue := burdenAnalyzer.DetectDeepNesting(fn, cfg.Analysis.Burden.MaxNesting); nestIssue != nil {
              report.Burden.DeeplyNestedFunctions = append(report.Burden.DeeplyNestedFunctions, *nestIssue)
          }

          // Feature envy (methods only)
          if fn.Recv != nil {
              if envyIssue := burdenAnalyzer.DetectFeatureEnvy(fn, result.File, cfg.Analysis.Burden.FeatureEnvyRatio); envyIssue != nil {
                  report.Burden.FeatureEnvyMethods = append(report.Burden.FeatureEnvyMethods, *envyIssue)
              }
          }

          return true
      })

      return nil
  }
  ```

### 6. Add Burden Metrics Summary to Console Reporter
- **Deliverable**: `internal/reporter/console.go` displays burden metrics in the analysis summary
- **Dependencies**: Steps 1-5 (burden data must be populated first)
- **Metric Justification**: Step 6.6 requires "Maintenance Burden Summary" section in all output formats
- **Implementation**:
  - Add `writeBurdenSummary()` function to console reporter
  - Display counts: magic numbers, dead code lines, complex signatures, deeply nested functions, feature envy methods
  - Integrate into `Generate()` output flow

### 7. Calculate DeadCodePercent in Finalization
- **Deliverable**: `finalizeReport()` calculates `report.Burden.DeadCode.DeadCodePercent` based on total lines
- **Dependencies**: Step 1
- **Metric Justification**: `DeadCodePercent` field exists but is always 0.0 (hardcoded in analyzer)
- **Implementation**:
  ```go
  // In finalizeReport():
  totalLines := report.Overview.TotalLines
  if totalLines > 0 {
      report.Burden.DeadCode.DeadCodePercent = float64(report.Burden.DeadCode.TotalDeadLines) / float64(totalLines) * 100
  }
  ```

### 8. Add Burden Section to HTML and Markdown Reporters
- **Deliverable**: `internal/reporter/html.go` and `internal/reporter/markdown.go` include burden metrics section
- **Dependencies**: Steps 1-5
- **Metric Justification**: ROADMAP.md Step 6.6 requires burden data in "all output formats"
- **Implementation**:
  - Add `renderBurdenSection()` to HTML reporter
  - Add `writeBurdenSection()` to Markdown reporter
  - Include tables for each burden category with issue details

### 9. Write Unit Tests for Integration
- **Deliverable**: `cmd/analyze_burden_test.go` with tests for burden integration
- **Dependencies**: Steps 1-5
- **Metric Justification**: Testing strategy requires integration tests for each phase
- **Implementation**:
  - Test that JSON output includes populated `.burden` object
  - Test that all burden categories appear in output when issues exist
  - Test threshold configurations affect detection counts

## Technical Specifications

- **AST Traversal**: Function-level burden analysis requires `ast.Inspect()` to iterate over `*ast.FuncDecl` nodes in each file
- **Configuration Integration**: All thresholds already defined in `config.Config.Analysis.Burden` struct:
  - `MaxParams` (default: 5)
  - `MaxReturns` (default: 3)
  - `MaxNesting` (default: 4)
  - `FeatureEnvyRatio` (default: 2.0)
  - `IgnoreBenignMagic` (default: true)
- **Thread Safety**: Burden analysis runs sequentially per file; no mutex required for `report.Burden` access
- **Memory**: Burden issues are appended to slices; no significant memory impact expected

## Validation Criteria

- [ ] `go-stats-generator analyze . --format json | jq '.burden'` returns populated object (not `null`)
- [ ] `go-stats-generator analyze . --format json | jq '.burden.dead_code.unreferenced_functions | length'` returns count > 0 when dead code exists
- [ ] `go-stats-generator analyze . --format json | jq '.burden.complex_signatures | length'` returns count > 0 when functions exceed thresholds
- [ ] `go-stats-generator analyze . --format json | jq '.burden.deeply_nested_functions | length'` returns count > 0 when nesting exceeds threshold
- [ ] `go-stats-generator analyze . --format json | jq '.burden.feature_envy_methods | length'` returns count > 0 when feature envy detected
- [ ] Console output includes "Maintenance Burden Summary" section
- [ ] All existing tests pass: `go test ./... -race`
- [ ] No regressions in complexity: `go-stats-generator analyze . --skip-tests` shows no new functions above complexity 20

## Known Gaps

1. **Shotgun Surgery Detection (Step 6.5 partial)**: Requires git history analysis (`git log --name-only`). The ROADMAP describes this feature but `BurdenAnalyzer` does not implement it. This is deferred — would require subprocess integration similar to the deferred annotation age tracking in Phase 4.

2. **Cross-Package Dead Code**: Current `DetectDeadCode()` only analyzes within a single package (files passed to it). True dead code detection requires cross-package reference analysis, which is not implemented.

3. **BurdenMetrics Aggregation**: The `BurdenMetrics` struct lacks summary fields like `TotalBurdenScore` or counts. Step 6.6 mentions a "composite burden score per file and per package" but this is not implemented in the struct definition.

## Files to Modify

| File | Changes |
|------|---------|
| `cmd/analyze.go` | Expand `analyzeBurdenInFile()` to call all burden analyzers |
| `internal/reporter/console.go` | Add burden summary section |
| `internal/reporter/html.go` | Add burden section to HTML output |
| `internal/reporter/markdown.go` | Add burden section to Markdown output |
| `cmd/analyze_burden_test.go` (new) | Integration tests for burden analysis |

## Estimated Effort

- **Total**: 2-3 days for a developer familiar with the codebase
- **Step 5 (refactor analyzeBurdenInFile)**: ~2 hours — core integration work
- **Step 6 (console reporter)**: ~1 hour
- **Steps 7-8 (other reporters)**: ~2 hours
- **Step 9 (tests)**: ~2 hours
- **Testing and refinement**: ~4 hours
