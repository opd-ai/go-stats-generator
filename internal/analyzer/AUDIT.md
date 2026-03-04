# Audit: github.com/opd-ai/go-stats-generator/internal/analyzer
**Date**: 2026-03-04
**Status**: Needs Work

## Summary
The `analyzer` package is the core AST analysis engine with 23 files and 5566 LOC. Documentation coverage (94.6%) and test coverage (82.5%) exceed thresholds, but cyclomatic complexity (max 15), function length (max 79 lines), and duplication ratio (6.22%) violate quality gates. Critical issues include significant code duplication between interface.go and struct.go, 5 functions exceeding complexity threshold, and naming convention violations.

## go-stats-generator Metrics
| Metric               | Value   | Threshold | Status |
|----------------------|---------|-----------|--------|
| Doc Coverage         | 94.6%   | ≥70%      | ✓      |
| Max Cyclomatic       | 15      | ≤10       | ✗      |
| Max Function Length  | 79      | ≤30 lines | ✗      |
| Test Coverage        | 82.5%   | ≥65%      | ✓      |
| Duplication Ratio    | 6.22%   | ≤5%       | ✗      |
| Naming Violations    | 10      | 0         | ✗      |

## Issues Found
- [x] **high** complexity — extractNestedBlocks exceeds threshold with cyclomatic 15 (`duplication.go:47`)
- [x] **high** complexity — walkForNestingDepth exceeds threshold with cyclomatic 14 (`burden.go:73`)
- [x] **high** complexity — checkStmtForUnreachable exceeds threshold with cyclomatic 13 (`burden.go:42`)
- [x] **high** complexity — walkForNestingDepth exceeds threshold with cyclomatic 12 (`function.go:66`)
- [x] **high** complexity — findCommentOutsideStrings exceeds threshold with cyclomatic 11 (`function.go:48`)
- [x] **high** duplication — 35-line renamed clone between interface.go:557-591 and struct.go:296-330
- [x] **high** duplication — 33-line renamed clone between interface.go:557-589 and struct.go:296-328
- [x] **high** duplication — 31-line renamed clone between interface.go:561-591 and struct.go:300-330
- [x] **high** duplication — 29-line renamed clone between interface.go:561-589 and struct.go:300-328
- [x] **high** duplication — 28-line renamed clone between interface.go:564-591 and struct.go:303-330
- [x] **high** function-length — NewNamingAnalyzer exceeds threshold with 79 lines (`naming.go:79`)
- [x] **med** function-length — walkForNestingDepth at 73 lines (`burden.go:73`)
- [x] **med** function-length — walkForNestingDepth at 66 lines (`function.go:66`)
- [x] **med** function-length — detectBuilder at 51 lines (`pattern.go:51`)
- [x] **med** function-length — findCommentOutsideStrings at 48 lines (`function.go:48`)
- [x] **med** function-length — extractNestedBlocks at 47 lines (`duplication.go:47`)
- [x] **med** function-length — AnalyzeFunctionAffinity at 60 lines (`placement.go:60`)
- [x] **med** function-length — AnalyzeCorrelation at 45 lines (`testcoverage.go:45`)
- [x] **med** function-length — detectBuilder at 43 lines (`pattern.go:43`)
- [x] **med** function-length — checkStmtForUnreachable at 42 lines (`burden.go:42`)
- [x] **med** naming — testcoverage.go should be coverage.go per Go conventions (`testcoverage.go`)
- [x] **med** naming — Package name "analyzer" does not match directory "." (`analyzer` package)
- [x] **low** naming — Acronym casing: processIdentRef should be processIDentRef (`placement.go:180`)
- [x] **low** naming — Acronym casing: AnalyzeIdentifiers should be AnalyzeIDentifiers (`naming.go:293`)
- [x] **low** naming — Acronym casing: checkIdentifier should be checkIDentifier (`naming.go:380`)
- [x] **low** naming — Acronym casing: checkIdentifierWithSingleLetter should be checkIDentifierWithSingleLetter (`naming.go:404`)
- [x] **low** naming — Acronym casing: checkIdentifierStuttering should be checkIDentifierStuttering (`naming.go:542`)
- [x] **low** naming — Acronym casing: ComputeIdentifierQualityScore should be ComputeIDentifierQualityScore (`naming.go:597`)
- [x] **low** naming — Single-letter parameter name: x (`statistics.go:18`)
- [x] **low** naming — Single-letter parameter name: y (`statistics.go:18`)
- [x] **low** stub-code — Placeholder comment for complex pattern detection (`concurrency.go`)

## Concurrency Assessment
**Goroutines**: 0 detected (no goroutines in this synchronous analysis package)
**Channels**: 0 detected
**Sync Primitives**: None detected
**Race Check**: PASS (go test -race completed successfully)
**Assessment**: Package is purely synchronous with no concurrency primitives. No race conditions or goroutine leaks possible.

## Dependencies
**External Dependencies**: 
- `github.com/opd-ai/go-stats-generator/internal/metrics` (data types)
- `github.com/opd-ai/go-stats-generator/internal/config` (configuration)

**Cohesion Score**: 4.27 (moderate - room for improvement)
**Coupling Score**: 1.0 (minimal external coupling)
**Circular Import Risk**: None detected
**Assessment**: Package has appropriate external dependencies for metrics definitions and config. Moderate cohesion suggests potential for splitting into more focused sub-packages.

## Recommendations
1. **Extract duplicated analysis logic** — 25 clone pairs detected, especially between interface.go and struct.go (lines 526-591). Extract shared visitor/walker logic to eliminate 579 duplicated lines (6.22% ratio).
2. **Refactor high-complexity functions** — Break down extractNestedBlocks (complexity 15), walkForNestingDepth (complexity 14), and checkStmtForUnreachable (complexity 13) using Extract Method pattern.
3. **Split oversized functions** — NewNamingAnalyzer (79 lines) should delegate initialization to helper functions. Apply to 12 functions exceeding 50 lines.
4. **Fix naming violations** — Rename testcoverage.go to coverage.go. Correct acronym casing in identifier-related method names (use ID not Ident).
5. **Improve package cohesion** — Consider splitting into sub-packages (e.g., analyzer/ast, analyzer/metrics, analyzer/quality) to improve cohesion score from 4.27 to <3.0.
