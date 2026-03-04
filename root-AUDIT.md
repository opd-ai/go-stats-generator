# Audit: github.com/opd-ai/go-stats-generator (root package)
**Date**: 2026-03-04
**Status**: Needs Work

## Summary
The root package serves as the CLI entry point, delegating to the cmd package. Analysis includes the entire codebase (104 files). Documentation coverage is good at 73.3%, but complexity and function length thresholds are exceeded in multiple packages. High duplication ratio (47%) and 47 naming violations require attention. No test coverage exists for main.go itself (expected for simple entry point).

## go-stats-generator Metrics
| Metric               | Value   | Threshold | Status |
|----------------------|---------|-----------|--------|
| Doc Coverage         | 73.3%   | ≥70%      | ✓      |
| Max Cyclomatic       | 24      | ≤10       | ✗      |
| Max Function Length  | 118     | ≤30 lines | ✗      |
| Test Coverage        | 0.0%    | ≥65%      | ✗      |
| Duplication Ratio    | 47.1%   | ≤5%       | ✗      |
| Naming Violations    | 47      | 0         | ✗      |

## Issues Found

### High Priority
- [x] high complexity — FilterReportSections exceeds threshold (`internal/metrics/sections.go:45`) complexity 23, 68 lines (COMPLETED: refactored to complexity 2, 6 lines)
- [x] high complexity — runWatch exceeds threshold (`cmd/watch.go:71`) complexity 14, 44 lines (COMPLETED: refactored to complexity 3, 21 lines)
- [x] high complexity — extractNestedBlocks exceeds threshold (`internal/analyzer/duplication.go:112`) complexity 15, 45 lines (COMPLETED: refactored to complexity 2, 20 lines)
- [x] high complexity — List exceeds threshold (`internal/storage/json.go:89`) complexity 14, 63 lines (COMPLETED: refactored to complexity 2, 8 lines)
- [x] high complexity — checkStmtForUnreachable exceeds threshold (`internal/analyzer/burden.go:156`) complexity 13, 40 lines (COMPLETED: refactored to complexity 2, 16 lines)
- [x] high function-length — init function too long (`cmd/analyze.go:68`) 118 lines, complexity 1 (COMPLETED: split into 15 helper functions, now 10 lines)
- [ ] high duplication — Overall duplication ratio 47.1% (125 clone pairs, 11,384 duplicated lines)

### Medium Priority
- [ ] med naming — 11 file name violations (generic names like types.go)
- [ ] med naming — 25 identifier violations (non-idiomatic names)
- [ ] med naming — 11 package name violations
- [ ] med documentation — 1 TODO comment (`internal/metrics/types.go:395`)
- [ ] med documentation — 1 FIXME comment (critical severity, `internal/metrics/types.go:403`)
- [ ] med documentation — 1 HACK comment (`internal/metrics/types.go:412`)
- [ ] med documentation — 1 BUG comment (critical severity, `internal/metrics/types.go:421`)
- [ ] med function-length — 69 functions exceed complexity or length thresholds

### Low Priority
- [ ] low documentation — 1 XXX comment (`internal/metrics/types.go:430`)
- [ ] low documentation — 2 DEPRECATED comments (`internal/metrics/types.go:438`, `internal/api/storage.go:3`)
- [ ] low documentation — 2 NOTE comments

## Concurrency Assessment
**Goroutines**: 33 total (30 anonymous, 3 named)
**Potential Leaks**: None detected
**Channels**: Analyzed across scanner, api packages
**Sync Primitives**: Proper worker pool patterns detected (scanner package)
**Race Detection**: PASS (go test -race .)

Notable patterns:
- Worker pool implementation in scanner package (`internal/scanner/worker.go:65`)
- Anonymous goroutine for analysis (`internal/api/server.go:71`)
- Proper context usage for cancellation

## Dependencies
**External Dependencies**: 
- github.com/spf13/cobra v1.9.1
- github.com/spf13/viper v1.20.1
- modernc.org/sqlite v1.31.1
- github.com/stretchr/testify v1.10.0

**Cohesion/Coupling**: Analyzed 22 packages with proper separation of concerns
**Circular Imports**: None detected

## Recommendations
1. **CRITICAL**: Reduce duplication ratio from 47.1% to ≤5% — consolidate clone pairs across analyzer and reporter packages
2. **HIGH**: Refactor FilterReportSections (complexity 23) — extract conditional logic into helper functions
3. **HIGH**: Split init function in cmd/analyze.go (118 lines) — move flag definitions to separate functions
4. **HIGH**: Address critical FIXME and BUG comments in internal/metrics/types.go
5. **MEDIUM**: Rename generic file names (types.go) to descriptive names
6. **MEDIUM**: Fix 25 identifier naming violations for Go idiomaticity
7. **MEDIUM**: Add test coverage for main.go entry point (currently 0%)
8. **LOW**: Resolve or remove TODO/XXX/DEPRECATED comments
