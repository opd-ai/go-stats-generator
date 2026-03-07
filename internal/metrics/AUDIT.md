# Audit: internal/metrics
**Date**: 2026-03-04
**Status**: Needs Work

## Summary
The `internal/metrics` package defines core data structures and diff computation logic for the go-stats-generator analysis engine. While the package has good test coverage (34.0%) and passes all vet/race checks, it fails multiple quality gates: documentation coverage (66.7% vs ≥70%), excessive cyclomatic complexity (max 23 vs ≤10), excessive function length (max 68 lines vs ≤30), and extremely high code duplication (22.91% vs ≤5%). The package contains critical technical debt markers (FIXME, BUG annotations) and 105 struct definitions that create significant maintenance burden.

## go-stats-generator Metrics
| Metric               | Value   | Threshold | Status |
|----------------------|---------|-----------|--------|
| Doc Coverage         | 66.7%   | ≥70%      | ✗      |
| Max Cyclomatic       | 23      | ≤10       | ✗      |
| Max Function Length  | 68      | ≤30 lines | ✗      |
| Test Coverage        | 34.0%   | ≥65%      | ✗      |
| Duplication Ratio    | 22.91%  | ≤5%       | ✗      |
| Naming Violations    | 4       | 0         | ✗      |

## Issues Found

### High Severity (11)
- [x] **high** complexity — FilterReportSections exceeds cyclomatic threshold 23 vs ≤10 (`sections.go:31`) — **RESOLVED**: Function refactored from 68 lines/complexity 23 to 5 lines/complexity 2 by extracting buildSectionKeepSet and clearUnrequestedSections helper functions with table-driven clearing via sectionHandlers map
- [x] **high** function-length — FilterReportSections exceeds length threshold 68 lines vs ≤30 (`sections.go:31`) — **RESOLVED**: Same refactoring reduced function to 5 lines
- [x] **high** duplication — 22.91% code duplication ratio (541 lines across 13 clone pairs, largest 18 lines) exceeds ≤5% threshold — **RESOLVED**: Duplication reduced to 0.50% (2 clone pairs, 15 lines) via types.go refactoring
- [x] **high** duplication — Overlapping struct definitions in types.go:1018-1036 (18-line clone with 2+ instances) — **RESOLVED**: types.go split into focused files, struct definitions consolidated
- [x] **high** duplication — Repeated patterns in types.go:915-934 (11-line clone pairs) — **RESOLVED**: Same refactoring eliminated these patterns
- [x] **high** function-length — compareFunctionMetrics exceeds length threshold 56 lines vs ≤30 (`diff.go:84`) — **RESOLVED**: Refactored to 3 lines by extracting buildFunctionMaps, collectAllFunctionKeys, and compareFunctionsByKey
- [x] **high** function-length — compareFunctionComplexity exceeds length threshold 49 lines vs ≤30 (`diff.go:156`) — **RESOLVED**: Refactored to 8 lines via extraction pattern
- [x] **high** function-length — categorizeChanges exceeds length threshold 39 lines vs ≤30 (`diff.go:503`) — **RESOLVED**: Refactored to 13 lines by extracting buildRegression and buildImprovement helper functions
- [x] **high** function-length — generateDiffSummary exceeds length threshold 31 lines vs ≤30 (`diff.go:550`) — **RESOLVED**: Refactored to 11 lines
- [ ] **high** test-coverage — Test coverage 34.0% is below ≥65% threshold (31% gap)
- [ ] **high** documentation — Critical FIXME annotation at `types.go:403`

### Medium Severity (8)
- [ ] **med** documentation — Coverage 66.7% is below ≥70% threshold (3.3% gap)
- [ ] **med** documentation — Critical BUG annotation at `types.go:421`
- [ ] **med** documentation — HACK annotation at `types.go:412`
- [ ] **med** documentation — TODO annotation at `types.go:395`
- [ ] **med** documentation — XXX annotation at `types.go:430`
- [ ] **med** naming — Package name "metrics" causes stuttering in exported type MetricsSnapshot (`types.go:683`)
- [ ] **med** naming — Package directory mismatch: package "metrics" in directory "." (severity: medium)
- [ ] **med** organization — types.go is oversized with 108 types, creating 7.23 maintenance burden index

### Low Severity (6)
- [ ] **low** naming — File name "types.go" is too generic (violation: generic_name)
- [ ] **low** naming — IdentifierViolation has incorrect acronym casing, should be IDentifierViolation (`types.go:595`)
- [ ] **low** documentation — DEPRECATED annotation at `types.go:438`
- [ ] **low** documentation — NOTE annotation at `types.go:447`
- [ ] **low** organization — diff.go is oversized with 39 functions, creating 0.72 maintenance burden index
- [ ] **low** cohesion — types.go has 0.00 file cohesion score, suggesting split into types_related.go

## Concurrency Assessment
**Goroutines**: 0 detected  
**Channels**: 0 detected  
**Sync Primitives**: None (no mutexes, wait groups, or atomic operations)  
**Race Detection**: PASS (go test -race completed successfully)  
**Assessment**: Package is purely data-structure focused with no concurrency patterns. No safety concerns.

## Dependencies
**External Dependencies**: 1 (time package)  
**Circular Dependencies**: None detected  
**Average Dependencies per Package**: 0.0  
**Cohesion/Coupling**: Package instability: 0.00 (stable)  
**Assessment**: Minimal external dependencies. Package serves as central type definition hub for the analyzer.

## Recommendations
1. **Critical**: Refactor FilterReportSections (sections.go:31) to reduce cyclomatic complexity from 23 to ≤10 by extracting section-clearing logic into separate functions
2. **Critical**: Eliminate 22.91% code duplication by consolidating overlapping struct definitions in types.go:1018-1036 (consider table-driven generation or embedding)
3. **Critical**: Split oversized types.go (108 types, 7.23 MBI) into focused files: diff_types.go, documentation_types.go, metrics_types.go, trend_types.go
4. **High**: Break down long functions in diff.go (compareFunctionMetrics:56 lines, compareFunctionComplexity:49 lines) using Extract Method refactoring
5. **High**: Increase test coverage from 34.0% to ≥65% by adding tests for diff computation and section filtering logic
6. **High**: Resolve critical FIXME (types.go:403) and BUG (types.go:421) annotations before production deployment
7. **Medium**: Add package-level documentation to increase coverage from 66.7% to ≥70%
8. **Medium**: Rename MetricsSnapshot to Snapshot to eliminate package stuttering
9. **Low**: Rename types.go to a more descriptive name (e.g., core_types.go or report_types.go)
10. **Low**: Fix IdentifierViolation acronym casing to IDentifierViolation per Go conventions
