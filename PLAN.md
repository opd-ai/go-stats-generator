# Implementation Plan: Phase 5 — Organizational & Structural Problems

## Phase Overview
- **Objective**: Detect structural issues that make a codebase hard to navigate and maintain (file size, package size, import graph health)
- **Source Document**: ROADMAP.md (Phase 5: Organizational & Structural Problems)
- **Prerequisites**: Phases 1-4 complete (Duplication ✅, Naming ✅, Placement ✅, Documentation ✅)
- **Estimated Scope**: Large — 14 functions above complexity threshold (15.0), 26.35% duplication ratio, 62.9% doc coverage

## Metrics Summary
- **Complexity Hotspots**: 14 functions above threshold (9.0) in target areas; 14 critical (>15.0)
- **Duplication Ratio**: 26.35% overall (critical — exceeds 10% threshold)
- **Documentation Coverage**: 62.9% overall (medium priority)
- **Package Coupling**: `cmd` (3.5), `go_stats_generator` (2.0), `reporter` (1.0), `storage` (1.0) — cmd has high coupling

### Critical Files by Complexity
| File | Function Count >9.0 | Max Complexity |
|------|---------------------|----------------|
| cmd/analyze.go | 7 | 21.3 |
| internal/analyzer/naming.go | 5 | 32.9 |
| internal/analyzer/concurrency.go | 5 | 11.9 |
| internal/analyzer/interface.go | 5 | 13.4 |
| internal/analyzer/placement.go | 4 | 20.7 |
| cmd/trend.go | 4 | 13.2 |
| internal/metrics/diff.go | 4 | 13.7 |
| internal/storage/json.go | 3 | 24.1 |
| internal/storage/sqlite.go | 3 | 18.4 |
| internal/reporter/json.go | 2 | 65.7 (CRITICAL) |

## Implementation Steps

### ✅ Step 1: Create OrganizationMetrics Type Definition (COMPLETED 2026-03-03)
- **Deliverable**: Add `OrganizationMetrics` struct to `internal/metrics/types.go` ✅
  - `OversizedFiles []OversizedFile` ✅
  - `OversizedPackages []OversizedPackage` ✅
  - `DeepDirectories []DeepDirectory` ✅
  - `HighFanInPackages []FanInPackage` ✅
  - `HighFanOutPackages []FanOutPackage` ✅
  - `AvgPackageStability float64` ✅
- **Dependencies**: None
- **Metric Justification**: Required foundation for Phase 5 reporting integration
- **Verification**: All structs added, Report struct updated, builds successfully, all tests pass

### ✅ Step 2: Implement File Size Analysis (COMPLETED 2026-03-03)
- **Deliverable**: `internal/analyzer/organization.go` with `AnalyzeFileSizes()` function ✅
  - Report total lines, code lines, comment lines, blank lines per file (reuse `LineMetrics`) ✅
  - Flag files exceeding `max_file_lines` (default: 500) ✅
  - Flag files exceeding `max_file_functions` (default: 20) ✅
  - Flag files exceeding `max_file_types` (default: 5) ✅
  - Compute maintenance burden score (composite of size, complexity, declaration count) ✅
- **Dependencies**: Step 1 (OrganizationMetrics type)
- **Metric Justification**: Addresses high complexity in `cmd/analyze.go` (7 functions >9.0) and `internal/reporter/json.go` (65.7 max complexity) — both likely oversized
- **Verification**: All functions <30 lines, all functions complexity ≤10, test coverage 100%, all tests pass, zero regressions ✅

### Step 3: Implement Package Size & Depth Analysis (Step 5.2)
- **Deliverable**: Extend `internal/analyzer/organization.go` with `AnalyzePackageSizes()` function
  - Flag packages with >20 files (`max_package_files`)
  - Flag packages with >50 exported symbols (`max_exported_symbols`)
  - Flag directories deeper than 5 levels (`max_directory_depth`)
  - Detect "mega-packages" (low cohesion + high symbol count)
- **Dependencies**: Step 2
- **Metric Justification**: `analyzer` package has 232 functions — potential mega-package candidate

### Step 4: Implement Import Graph Health Analysis (Step 5.3)
- **Deliverable**: Extend `internal/analyzer/organization.go` with `AnalyzeImportGraph()` function
  - Flag files with >15 imports (`max_file_imports`)
  - Identify "hub" packages (high fan-in — change bottleneck)
  - Identify "authority" packages (high fan-out — coupling indicator)
  - Compute instability metric: `fan-out / (fan-in + fan-out)`
- **Dependencies**: Step 3, existing `PackageAnalyzer`
- **Metric Justification**: `cmd` package has coupling score 3.5 with 7 dependencies — highest coupling in codebase

### Step 5: Add Configuration Options
- **Deliverable**: Update `.go-stats-generator.yaml` schema and `cmd/analyze.go` with:
  ```yaml
  maintenance:
    organization:
      max_file_lines: 500
      max_file_functions: 20
      max_file_types: 5
      max_package_files: 20
      max_exported_symbols: 50
      max_directory_depth: 5
      max_file_imports: 15
  ```
- **Dependencies**: Steps 2-4
- **Metric Justification**: Configuration enables project-specific thresholds for enforcement

### Step 6: Integrate into Report Generation
- **Deliverable**: Update all reporters (`console.go`, `json.go`, `html.go`, `markdown.go`) with "Organization Health" section
- **Dependencies**: Steps 1-5
- **Metric Justification**: `internal/reporter/json.go` has Generate function with 65.7 complexity — will need careful integration

### Step 7: Add Comprehensive Unit Tests
- **Deliverable**: `internal/analyzer/organization_test.go` with:
  - Table-driven tests for each threshold
  - Integration tests against `testdata/` samples
  - Edge case tests for empty packages, single-file packages
  - Benchmark tests for performance validation
- **Dependencies**: Step 6
- **Metric Justification**: Follows existing test patterns (>95% coverage target per ROADMAP.md)

### Step 8: Refactor Generate Function (Prerequisite Cleanup)
- **Deliverable**: Split `internal/reporter/json.go:Generate` (complexity 65.7) into smaller helper functions
  - Target: <15 complexity per extracted function
  - Extract section generators: `generateOverviewSection()`, `generateFunctionSection()`, etc.
- **Dependencies**: None (can be done in parallel with Steps 1-5)
- **Metric Justification**: Generate has 65.7 complexity (CRITICAL) — must be addressed before adding more output sections

## Technical Specifications
- **File Size Analysis**: Reuse existing `LineMetrics` from function analysis; extend AST visitor to count declarations per file
- **Package Metrics**: Extend `PackageAnalyzer` to track file counts, symbol counts, and import relationships
- **Instability Calculation**: Per Robert C. Martin's dependency metrics: I = Ce / (Ca + Ce) where Ca = afferent coupling (fan-in), Ce = efferent coupling (fan-out)
- **Threshold Storage**: Add `OrganizationConfig` struct to `internal/config/config.go` for YAML parsing
- **Reporter Integration**: Follow existing patterns in `console.go` sections (table-based output with ranking)

## Validation Criteria
- [ ] `go-stats-generator analyze .` includes "Organization Health" section in output
- [ ] All new functions have cyclomatic complexity <15 (validated via `go-stats-generator analyze`)
- [ ] `internal/reporter/json.go:Generate` complexity reduced from 65.7 to <20
- [ ] Test coverage ≥95% for `internal/analyzer/organization.go`
- [ ] `go test ./internal/analyzer/...` passes with no failures
- [ ] Documentation coverage for new exported symbols ≥80%
- [ ] No regressions: `go-stats-generator diff baseline.json final.json` shows improvements only
- [ ] Performance: analysis time increase <10% on 50,000+ file codebases

## Known Gaps
- **Git History Integration**: Step 5.3 mentions fan-in/fan-out based on imports, but shotgun surgery detection (Phase 6) requires git log parsing — defer git integration to Phase 6
- **Mega-Package Heuristics**: Definition of "low cohesion + high symbol count" needs specific thresholds — propose: cohesion <0.5 AND symbols >30
- **Reporter Complexity**: The Generate function refactoring (Step 8) may surface additional complexity in HTML/Markdown reporters that also need refactoring — assess after Step 8 completion

## Estimated Timeline
- Step 1: 0.5 day (type definitions)
- Steps 2-4: 2 days each (6 days total for core analysis)
- Step 5: 0.5 day (configuration)
- Step 6: 1 day (reporter integration)
- Step 7: 1.5 days (testing)
- Step 8: 1 day (refactoring — parallel track)

**Total: ~10-11 developer days**

## Risk Assessment
| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Generate refactoring breaks existing output | Medium | High | Snapshot-based regression tests before refactoring |
| Performance degradation on large codebases | Low | Medium | Benchmark tests; lazy evaluation where possible |
| Configuration schema conflicts with existing tools | Low | Low | Document in `.go-stats-generator.yaml.example` |

---

*Generated: 2026-03-03 based on go-stats-generator v1.0.0 metrics analysis*
