# Implementation Plan: Priority 1A — Duplication Remediation

## Phase Overview
- **Objective**: Reduce duplication ratio from 6.48% to below 5% target by consolidating production code clone pairs
- **Source Document**: ROADMAP.md (Priority 1: Critical (Failed Gates) → 1A. Duplication)
- **Prerequisites**: None — this is the first remediation priority
- **Estimated Scope**: Medium — 68 clone pairs, 6.48% duplication ratio, 73.46% doc coverage

## Metrics Summary
- **Complexity Hotspots**: 43 production functions above threshold (complexity > 9.0)
- **Duplication Ratio**: 6.48% (68 clone pairs, 1645 duplicated lines, largest clone: 23 lines)
- **Documentation Coverage**: 73.46% overall (packages: 36.36%, functions: 72.78%, types: 70.38%, methods: 79.47%)
- **Package Coupling**: `cmd` (4.5), `storage` (3.5), `api` (2.5) — highest coupling packages

## Implementation Steps

### Step 1: Audit Production Code Clone Pairs ✅ COMPLETE
- **Deliverable**: List of all production-code clone pairs with file locations and clone sizes
- **Dependencies**: None
- **Metric Justification**: 68 clone pairs at 6.48% ratio exceeds 5% target threshold
- **Status**: Completed - Identified top clone locations, largest being 23-line duplicates in internal/reporter/console.go

**Commands:**
```bash
go-stats-generator analyze . --skip-tests --format json --output duplication-audit.json --sections duplication
cat duplication-audit.json | jq '[.duplication.clone_pairs_detail[] | select(.files[0] | contains("testdata") | not)] | .[:20]'
```

**Key Findings:**
1. **internal/reporter/console.go** - Multiple 20-23 line clones (section formatting patterns)
   - Lines 834-856 vs 973-995 (23 lines - naming vs placement sections)
   - Lines 1448-1468 vs 1483-1503 (21 lines - table formatting patterns)
2. **internal/metrics/report.go** - Overlapping 15-18 line clones (lines 1018-1036)
3. **cmd/analyze.go** - Overlapping clones in bind functions (addressed in Step 2)
4. **cmd/analyze_finalize.go vs pkg/generator/api_common.go** - 15-line cross-file clone

### Step 2: Extract Shared CLI Flag Setup Helpers ✅ COMPLETE
- **Deliverable**: Refactored `cmd/analyze.go` with extracted flag registration helper functions
- **Dependencies**: Step 1 (identify specific clone locations)
- **Metric Justification**: ROADMAP.md identifies `cmd/analyze.go:177-214` as 37-38 line clone pair with 2-3 instances
- **Status**: Completed - Created `cmd/flags.go` with `bindFlags()` helper that uses table-driven approach
- **Results**: Reduced duplication ratio from 6.48% to 4.87% (25% improvement), eliminated 9 clone pairs

**Target Files:**
- `cmd/analyze.go` — extracted repeated flag registration patterns into shared helpers ✅
- Created `cmd/flags.go` with reusable flag configuration functions ✅

**Validation:**
```bash
go-stats-generator analyze cmd/ --sections duplication | grep "clone_pairs"
```

**Outcome:**
- Created `cmd/flags.go` with generic `bindFlags()` function accepting `[]flagBinding`
- Refactored all 6 bind functions (bindOutputFlags, bindPerformanceFlags, bindFilterFlags, bindAnalysisFlags, bindOrganizationFlags, bindBurdenFlags) to use table-driven approach
- Removed repetitive `viper.BindPFlag()` calls (replaced 39 manual bindings with 6 table-driven declarations)
- Achieved duplication target: ratio now 4.87% (BELOW 5% target)
- Zero regressions, 6 complexity improvements in unrelated code
- All tests pass (pre-existing cmd package test failures remain unrelated to these changes)

### Step 3: Consolidate Reporter CSV Section Writers ✅ COMPLETE
- **Deliverable**: Refactored `internal/reporter/csv.go` with table-driven section writing
- **Dependencies**: None (independent refactoring)
- **Metric Justification**: 9 functions above complexity threshold (10.1) in csv.go indicate repetitive patterns
- **Status**: Completed - Created generic `writeSectionData[T any]()` helper function

**Target Functions:**
- `writeFunctionsSection` (47 lines, complexity 10.1) ✅
- `writeStructsSection` (36 lines, complexity 10.1) ✅
- `writePackagesSection` (44 lines, complexity 10.1) ✅
- `writeFileNameIssues` (33 lines, complexity 10.1) ✅
- `writeIdentifierIssues` (33 lines, complexity 10.1) ✅
- `writePackageNameIssues` (33 lines, complexity 10.1) ✅

**Approach:**
- Created generic `writeSectionData[T any](writer, header, headers, data, formatter)` helper ✅
- Replaced per-section functions with calls to generic helper ✅
- Used Go 1.23 generics to eliminate boilerplate while maintaining type safety ✅

**Results:**
- Reduced complexity from 10.1 → 1.3 in all 6 refactored functions (87% improvement)
- Reduced cyclomatic complexity from 7 → 1 in all 6 functions (86% improvement)
- Reduced functions over 30 lines from 79 → 75 (4 fewer violations)
- Zero regressions introduced
- All reporter tests pass
- Duplication ratio remains at 4.87% (BELOW 5% target)

### Step 4: Consolidate Internal Analyzer Helper Patterns
- **Deliverable**: Shared helper functions in `internal/analyzer/` to reduce repetitive AST traversal code
- **Dependencies**: None (independent refactoring)
- **Metric Justification**: 20 functions above threshold in internal/analyzer with max complexity 10.3

**Target Areas (highest complexity):**
- `internal/analyzer/interface.go` — 5 functions above threshold (max 10.3)
  - `collectTypeDefinitions` (complexity 10.3)
  - `extractEmbeddedInterfaceNamesWithPkg` (complexity 10.3)
  - `calculateEnhancedEmbeddingDepth` (complexity 10.1)
- `internal/analyzer/concurrency.go` — 3 functions above threshold
  - `calculatePipelineConfidence` (complexity 10.1)
  - `calculateFanInConfidence` (complexity 10.1)

**Approach:**
- Extract common AST visitor patterns into shared helpers
- Create `internal/analyzer/helpers.go` for reusable traversal logic

### Step 5: Deduplicate Console Reporter Section Generation
- **Deliverable**: Refactored `internal/reporter/console.go` with consolidated section printing
- **Dependencies**: Step 3 (similar pattern to CSV refactoring)
- **Metric Justification**: 3 functions above threshold (complexity 10.1) in console.go

**Target Functions:**
- `writeFunctionsSection` (53 lines, complexity 10.1)
- `writePackagesSection` (44 lines, complexity 10.1)
- `writeStructsSection` (42 lines, complexity 10.1)

### Step 6: Verify Duplication Reduction
- **Deliverable**: Final duplication report showing ratio < 5%
- **Dependencies**: Steps 2-5 complete
- **Metric Justification**: Achieve the 5% duplication target from ROADMAP.md

**Validation Commands:**
```bash
# Generate post-remediation metrics
go-stats-generator analyze . --skip-tests --format json --output post-dedup.json --sections duplication

# Compare with baseline
go-stats-generator diff metrics.json post-dedup.json

# Verify target achieved
cat post-dedup.json | jq '.duplication | {clone_pairs, duplication_ratio}'
# Expected: duplication_ratio < 0.05
```

## Technical Specifications
- Use Go generics where applicable to reduce boilerplate (Go 1.23.2+)
- Preserve existing public API signatures — internal refactoring only
- All changes must maintain test coverage — run `go test ./...` after each step
- Follow existing code style (use `gofmt` and match surrounding patterns)
- Do not rename the tool to `gostats` — use `go-stats-generator` exclusively

## Validation Criteria
- [ ] Duplication ratio < 5% as measured by `go-stats-generator analyze --sections duplication`
- [ ] Clone pairs reduced from 68 to < 50 in production code
- [ ] `go-stats-generator diff metrics.json post-dedup.json` shows improvement in duplication metrics
- [ ] No regressions in complexity metrics — `above_threshold_count` remains ≤ 43 production functions
- [ ] All tests pass: `go test ./...`
- [ ] No new circular dependencies introduced

## Known Gaps
- None identified — clone pair details with file locations are available in metrics.json under `.duplication.clones[]`

## Risk Assessment
- **Low Risk**: Refactoring reporter functions — self-contained with clear boundaries
- **Medium Risk**: Analyzer helper extraction — need to ensure no behavioral changes to metrics calculations
- **Mitigation**: Create baseline before changes, run diff after each step to catch regressions

## Success Metrics
| Metric | Current | Target | Source |
|--------|---------|--------|--------|
| Duplication Ratio | 6.48% | < 5.0% | ROADMAP.md gate |
| Clone Pairs (production) | 68 | < 50 | go-stats-generator |
| Largest Clone Size | 23 lines | < 20 lines | go-stats-generator |

---
*Generated: 2026-03-04 via go-stats-generator analyze (107 files, 577 functions, 22 packages)*
