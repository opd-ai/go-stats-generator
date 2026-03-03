# Implementation Plan: Phase 7 — Composite Scoring & Actionable Output (Steps 7.2-7.4)

## Phase Overview
- **Objective**: Complete Phase 7 by implementing prioritized refactoring suggestions, baseline/trend integration, and CI/CD quality gates
- **Source Document**: ROADMAP.md (Phase 7: Composite Scoring & Actionable Output)
- **Prerequisites**: Phases 1-6 complete, Step 7.1 (MBI scoring) complete
- **Estimated Scope**: Large — 12 functions above complexity threshold in target areas, 33.42% duplication ratio, 66.6% doc coverage

## Metrics Summary
- **Complexity Hotspots**: 12 functions above threshold (>15) in production code, 6 critical (>20)
- **Duplication Ratio**: 33.42% overall (critical — well above 10% threshold)
- **Documentation Coverage**: 66.6% overall, 10.0% package-level coverage (medium)
- **Package Coupling**: `cmd` package has highest coupling (3.5), `analyzer` package has 352 functions

### Critical Complexity Functions Requiring Attention:
| Function | File | Complexity |
|----------|------|------------|
| WriteDiff | internal/reporter/csv.go | 24.9 |
| Cleanup | internal/storage/json.go | 24.1 |
| extractNestedBlocks | internal/analyzer/duplication.go | 21.5 |
| buildSymbolIndex | internal/analyzer/placement.go | 20.7 |
| walkForNestingDepth | internal/analyzer/burden.go | 19.2 |
| List | internal/storage/json.go | 19.2 |
| checkStmtForUnreachable | internal/analyzer/burden.go | 18.9 |
| Cleanup | internal/storage/sqlite.go | 18.4 |

## Implementation Steps

### Step 1: Implement Prioritized Refactoring Suggestions (7.2)

1.1. **Create suggestion generator infrastructure**
   - **Deliverable**: `internal/analyzer/suggestions.go` with `RefactoringSuggestion` struct and `SuggestionGenerator` type
   - **Dependencies**: Step 7.1 MBI scoring complete
   - **Metric Justification**: Required to surface actionable insights from existing MBI scores
   - **Technical Details**:
     - Each suggestion includes: action type, target location, estimated MBI impact, effort classification
     - Action types: `extract_function`, `rename`, `move_to_file`, `add_documentation`, `reduce_complexity`, `deduplicate`
     - Effort levels: low (<1 hour), medium (1-4 hours), high (>4 hours)

1.2. **Implement impact-to-effort scoring algorithm**
   - **Deliverable**: `calculateImpactEffortRatio()` function that computes ROI for each suggestion
   - **Dependencies**: Step 1.1
   - **Metric Justification**: 135 clone pairs and 33.42% duplication need prioritized remediation
   - **Technical Details**:
     - Impact = delta in MBI score if suggestion implemented
     - Effort = estimated based on affected lines and complexity
     - Sort suggestions by impact/effort ratio descending

1.3. **Generate suggestions for each burden category**
   - **Deliverable**: Category-specific generators for duplication, naming, placement, documentation, organization, and burden metrics
   - **Dependencies**: Step 1.2
   - **Metric Justification**: All six burden categories need actionable output per ROADMAP
   - **Technical Details**:
     - Duplication: "Extract to shared function in X" with affected locations
     - Complexity: "Split function into N helpers" for functions >15 complexity
     - Documentation: "Add GoDoc to exported symbol X" for undocumented exports

1.4. **Integrate suggestions into output formats**
   - **Deliverable**: `Refactoring Suggestions` section in console, JSON, HTML, and Markdown reporters
   - **Dependencies**: Step 1.3
   - **Metric Justification**: MBI scores exist in JSON but not console/HTML/Markdown (noted in ROADMAP Step 7.1)

### Step 2: Extend Baseline & Diff Commands for Burden Metrics (7.3)

2.1. **Add burden metrics to baseline storage schema**
   - **Deliverable**: Updated `internal/storage/*.go` with burden metrics columns/fields
   - **Dependencies**: None
   - **Metric Justification**: Current baseline/diff only covers basic metrics, not MBI scores
   - **Technical Details**:
     - Add `mbi_score`, `duplication_ratio`, `doc_coverage` fields to baseline snapshots
     - Maintain backward compatibility with existing baselines

2.2. **Implement burden-specific regression detection**
   - **Deliverable**: `detectBurdenRegressions()` in `internal/metrics/diff.go`
   - **Dependencies**: Step 2.1
   - **Metric Justification**: WriteDiff in csv.go has 24.9 complexity — needs regression protection
   - **Technical Details**:
     - Alert when file MBI increases by configurable threshold (default: 10 points)
     - Alert when package MBI increases by configurable threshold (default: 5 points)
     - Add `--burden-regression-threshold` flag

2.3. **Integrate burden trends into trend command**
   - **Deliverable**: Burden metrics in `cmd/trend.go` time-series output
   - **Dependencies**: Step 2.2
   - **Metric Justification**: cmd/trend.go has 4 functions above complexity threshold — extend carefully
   - **Technical Details**:
     - Track MBI score over time per file and package
     - Show trend direction (improving/degrading/stable)
     - Add visual indicators in console output

### Step 3: Implement CI/CD Quality Gates (7.4)

3.1. **Add `--max-burden-score` flag to analyze command**
   - **Deliverable**: Updated `cmd/analyze.go` with flag and exit-code logic
   - **Dependencies**: Step 7.1 MBI scores
   - **Metric Justification**: cmd/analyze.go has 9 functions above threshold — add flag carefully
   - **Technical Details**:
     - Exit code 1 when any file or package exceeds threshold
     - Default threshold: 70 (critical level per MBI scale)
     - Output which files/packages exceeded threshold before exit

3.2. **Add per-category threshold flags**
   - **Deliverable**: `--max-duplication-ratio`, `--max-undocumented-exports`, `--max-complexity` flags
   - **Dependencies**: Step 3.1
   - **Metric Justification**: 33.42% duplication ratio needs enforceable ceiling
   - **Technical Details**:
     - `--max-duplication-ratio` (default: 0.10 = 10%)
     - `--max-undocumented-exports` (default: 10 symbols)
     - `--max-complexity` (default: 15.0)
     - Each flag triggers exit code 1 on violation

3.3. **Add CI/CD documentation and examples**
   - **Deliverable**: `docs/ci-cd-integration.md` with GitHub Actions, GitLab CI, and Jenkins examples
   - **Dependencies**: Steps 3.1, 3.2
   - **Metric Justification**: 10.0% package doc coverage indicates need for better docs
   - **Technical Details**:
     - Example workflows for common CI systems
     - Recommended thresholds for new vs legacy codebases
     - How to gradually tighten thresholds over time

### Step 4: Address High-Complexity Functions (Prerequisite Cleanup)

4.1. **Refactor WriteDiff in internal/reporter/csv.go (complexity 24.9)**
   - **Deliverable**: Refactored function with complexity ≤15
   - **Dependencies**: None — standalone cleanup
   - **Metric Justification**: Highest complexity in production code outside testdata
   - **Technical Details**:
     - Extract diff formatting logic into helper functions
     - Target: 3-4 smaller functions with single responsibilities

4.2. **Refactor Cleanup functions in storage package (complexity 24.1, 18.4)**
   - **Deliverable**: Refactored `internal/storage/json.go` and `internal/storage/sqlite.go` Cleanup methods
   - **Dependencies**: None — standalone cleanup
   - **Metric Justification**: Two functions above critical threshold in same package
   - **Technical Details**:
     - Extract error handling and resource cleanup into helpers
     - Target complexity ≤12 for each

4.3. **Refactor buildSymbolIndex in placement.go (complexity 20.7)**
   - **Deliverable**: Refactored function with complexity ≤15
   - **Dependencies**: None — standalone cleanup
   - **Metric Justification**: Highest complexity in analyzer package
   - **Technical Details**:
     - Extract AST walking logic into separate functions
     - Consider using visitor pattern for cleaner structure

## Technical Specifications

- **Suggestion Generator Architecture**: Implement as interface with per-category implementations to allow easy extension
- **Impact Estimation**: Use linear model based on affected lines × complexity × category weight
- **Effort Classification**: low = <30 LoC changed, medium = 30-100 LoC, high = >100 LoC
- **Backward Compatibility**: Baseline files from v1.0.0 must remain readable; add new fields as optional
- **Exit Codes**: 0 = success, 1 = quality gate violation, 2 = analysis error
- **Configuration**: All new thresholds configurable via `.go-stats-generator.yaml` under `maintenance.scoring` section

## Validation Criteria

- [ ] `go-stats-generator analyze` outputs refactoring suggestions section with at least 10 prioritized items
- [ ] `go-stats-generator baseline save && go-stats-generator diff` includes burden metric comparisons
- [ ] `go-stats-generator analyze --max-burden-score 50` exits with code 1 on this codebase (current MBI likely >50)
- [ ] `go-stats-generator analyze --max-duplication-ratio 0.30` exits with code 1 (current: 33.42%)
- [ ] `go-stats-generator trend` shows MBI trend line in output
- [ ] All refactored functions have complexity ≤15 per `go-stats-generator analyze`
- [ ] `go-stats-generator diff baseline.json final.json` shows no regressions in unrelated areas
- [ ] Documentation coverage for new code ≥80%
- [ ] All tests pass: `go test ./...`
- [ ] Package doc coverage improves from 10.0% to ≥30% with new docs

## Known Gaps

### Gap 1: Shotgun Surgery Detection Incomplete
- **Description**: Step 6.5 notes shotgun surgery detection was deferred (requires git history analysis)
- **Impact**: Suggestions generator cannot include shotgun surgery recommendations
- **Metrics Context**: Would help identify functions frequently changed together
- **Resolution**: Could be added in future phase; not blocking for Phase 7 completion

### Gap 2: MBI Console/HTML/Markdown Output Not Implemented
- **Description**: ROADMAP Step 7.1 notes "Console/HTML/Markdown output for MBI scores not yet implemented"
- **Impact**: Step 1.4 must include MBI score display alongside suggestions
- **Metrics Context**: MBI data exists in `.scores` JSON key but not rendered
- **Resolution**: Include in Step 1.4 deliverables

### Gap 3: High Duplication Ratio May Affect Effort Estimates
- **Description**: 33.42% duplication ratio is extremely high and may indicate test data inflation
- **Impact**: Effort estimates for deduplication suggestions may be skewed
- **Metrics Context**: 135 clone pairs, 6182 duplicated lines — verify these are production vs test files
- **Resolution**: Analyze duplication with `--skip-tests` to get production-only ratio; adjust estimates accordingly

## Appendix: Raw Metrics Data

```json
{
  "complexity": {
    "functions_above_threshold": 12,
    "critical_functions": 6,
    "average_complexity": 4.9
  },
  "duplication": {
    "clone_pairs": 135,
    "duplicated_lines": 6182,
    "duplication_ratio": 0.3342,
    "largest_clone_size": 35
  },
  "documentation": {
    "overall_coverage": 66.6,
    "package_coverage": 10.0,
    "function_coverage": 68.5,
    "type_coverage": 58.1
  },
  "organization": {
    "oversized_files": 19,
    "oversized_packages": 5,
    "high_coupling_packages": ["cmd (3.5)", "go_stats_generator (2.0)"]
  }
}
```

---

*Generated: 2026-03-03 by go-stats-generator metrics analysis*
*Phase selection based on ROADMAP.md — Phase 7 Steps 7.2-7.4 incomplete*
