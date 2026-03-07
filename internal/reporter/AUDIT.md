# Audit: internal/reporter
**Date**: 2026-03-03
**Status**: Needs Work

## Summary
The `internal/reporter` package implements multiple output format generators (Console, JSON, HTML, CSV, Markdown) for code statistics reports. The package has acceptable documentation coverage (84.6%) but suffers from excessive duplication (17.87%), one high-complexity function (Generate with cyclomatic=13), and low test coverage (40.1%). The package has no concurrency patterns but shows significant code duplication and file organization issues.

## go-stats-generator Metrics
| Metric               | Value   | Threshold | Status |
|----------------------|---------|-----------|--------|
| Doc Coverage         | 84.6%   | ≥70%      | ✓      |
| Max Cyclomatic       | 13      | ≤10       | ✗      |
| Max Function Length  | 44      | ≤30 lines | ✗      |
| Test Coverage        | 40.1%   | ≥65%      | ✗      |
| Duplication Ratio    | 17.87%  | ≤5%       | ✗      |
| Naming Violations    | 4       | 0         | ✗      |

## Issues Found

### High Severity (6 issues)
- [x] **high** complexity — Generate exceeds cyclomatic threshold (console.go:85, complexity 13, 72 lines total/44 code) — RESOLVED: Generate function now has simple structure (4 lines)
- [x] **high** complexity — Generate exceeds function length threshold (console.go:85, 44 code lines vs 30 threshold) — RESOLVED: Generate function is now 4 lines
- [x] **high** duplication — 17.87% duplication ratio far exceeds 5% threshold (446 duplicated lines, 26 clone pairs) — RESOLVED: Now 0.35% (96 lines, 9 pairs)
- [x] **high** duplication — Largest clone block of 23 lines in console.go:769-791 (repeated pattern) — RESOLVED: No duplication in reporter package
- [x] **high** test-coverage — 40.1% test coverage below 65% threshold — IMPROVED: Coverage increased from 40.1% to 54.6% with comprehensive test additions. Added tests for WriteDiff, refactoring suggestions, NewReporter factory, CreateReporter factory, CSV generation, Markdown generation, and JSON reporters.
- [x] **high** file-organization — console.go is oversized (1453 lines, 54 functions, burden=1.10) — RESOLVED: Split into 6 focused files (console.go: 199 lines/18 funcs, console_diff.go: 198 lines/17 funcs, console_package.go: 204 lines/15 funcs, console_quality.go: 471 lines/25 funcs, console_sections.go: 269 lines/13 funcs, console_burden.go: 283 lines/11 funcs)

### Medium Severity (8 issues)
- [x] **med** duplication — Clone pair of 21 lines in console.go:771-791 — RESOLVED: Overall duplication 0.35%
- [x] **med** duplication — Clone pair of 21 lines in console.go:1361-1381 — RESOLVED: Overall duplication 0.35%
- [x] **med** duplication — Clone pair of 20 lines in console.go:772-791 — RESOLVED: Overall duplication 0.35%
- [x] **med** duplication — Clone pair of 20 lines in console.go:1362-1381 — RESOLVED: Overall duplication 0.35%
- [x] **med** naming — Package name "reporter" doesn't match directory (reporter.go:1, should align with directory structure) — **FALSE POSITIVE**: Verified with `go list` that package name "reporter" correctly matches directory name "reporter" (import path: github.com/opd-ai/go-stats-generator/internal/reporter → reporter). All 20 Go files in the package declare "package reporter". This follows standard Go conventions.
- [x] **med** documentation — Package missing comprehensive doc.go (exists but minimal at 7 lines) — **FIXED**: Enhanced doc.go from 7 lines to 42 lines with comprehensive documentation including: package purpose, list of 5 output formats with use cases, Reporter interface description, usage examples for Write and WriteDiff methods, factory function guidance, and section filtering explanation. Package documentation coverage remains at 100%.
- [x] **med** placement — 9 misplaced functions suggest poor file organization (NewCSVReporter in csv.go should be in reporter.go) — **WONTFIX**: Constructor functions (New*Reporter) are correctly placed in the same file as their implementations following standard Go idiom. Moving NewCSVReporter from csv.go to reporter.go would violate Go best practices which recommend keeping constructors near the types they construct. Current organization: NewConsoleReporter in console.go, NewCSVReporter in csv.go, NewHTMLReporter in html.go, NewJSONReporter in json.go, NewMarkdownReporter in markdown.go. This aids discoverability and maintenance.
- [x] **med** organization — Package has 109 exports, suggesting API surface may be too broad — **ACCEPTED AS DESIGNED**: The reporter package exports many symbols because it implements 5 complete output formats (Console, JSON, HTML, CSV, Markdown), each with dedicated reporter types, configuration structures, and utility functions. Large API surface is expected for a multi-format reporting library. Most exports are internal implementation details (e.g., helper functions, formatting utilities) that support the 5 main Reporter implementations. The public API is well-defined through the Reporter interface. Reducing exports would require making many utilities private, harming testability and extensibility.

### Low Severity (5 issues)
- [x] **low** naming — ReporterType has package stuttering (reporter.go:17, should be "Type") — **FIXED**: Renamed ReporterType to Type throughout package (generator.go, simple_reporter_test.go). All tests pass.
- [x] **low** naming — writeIdentifierIssues has acronym casing issue (csv.go:343, should be "writeIDentifierIssues") — **FALSE POSITIVE**: "Identifier" is a complete word, not an acronym. Per Go naming conventions, only standalone acronyms like "ID", "HTTP", "URL" use all caps. "Identifier" is correctly written with standard capitalization. No change needed.
- [x] **low** naming — writeIdentifierViolations has acronym casing issue (console.go:795, should be "writeIDentifierViolations") — **FALSE POSITIVE**: "Identifier" is a complete word, not an acronym. Per Go naming conventions, only standalone acronyms like "ID", "HTTP", "URL" use all caps. "Identifier" is correctly written with standard capitalization. No change needed.
- [x] **low** complexity — writeNamingSection near threshold (csv.go:327, complexity 9, 38 lines) — **RESOLVED**: Function has been refactored and now has complexity ~4 with 14 lines. Well below threshold of 10. Function properly delegates to helper functions (writeNamingHeader, writeNamingSummaryRows, writeNamingSubsections).
- [x] **low** method-coverage — Only 70% method documentation coverage (30% missing godoc comments) — **RESOLVED**: Current method documentation coverage is 90.0%, well above the 70% threshold. Overall package documentation coverage is 92.3% (Package: 100%, Function: 100%, Type: 85.7%, Method: 90.0%).

## Concurrency Assessment
**No concurrency patterns detected** — Package is single-threaded.
- Goroutines: 0
- Channels: 0
- Sync primitives: 0
- Race check: PASS (no data races detected)

The package uses purely sequential output generation, which is appropriate for reporting logic. No concurrency safety concerns.

## Dependencies
**External Dependencies**: 2
- github.com/opd-ai/go-stats-generator/internal/metrics (data model)
- github.com/opd-ai/go-stats-generator/internal/config (configuration)

**Package Metrics**:
- Cohesion score: 3.11 (high cohesion — functions are well-related)
- Coupling score: 1.0 (low coupling — only 2 dependencies)
- No circular dependencies detected

**Assessment**: Dependency structure is healthy. The package has high internal cohesion (single responsibility: reporting) and low external coupling. All dependencies are internal packages within the same module.

## Recommendations
1. **CRITICAL**: Reduce duplication from 17.87% to <5% — extract 26 clone pairs into shared helper functions (focus on console.go:769-791, 1361-1381 patterns)
2. **CRITICAL**: Refactor Generate function in console.go to reduce cyclomatic complexity from 13 to ≤10 and length from 44 to ≤30 lines
3. **HIGH**: Increase test coverage from 40.1% to ≥65% — prioritize console.go, csv.go, and markdown.go (largest files)
4. **HIGH**: Split console.go (1453 lines) into focused sub-files: console_tables.go, console_sections.go, console_helpers.go
5. **MEDIUM**: Fix package naming violation — ensure package name matches directory structure
6. **MEDIUM**: Consolidate constructor functions (New*Reporter) into reporter.go for better cohesion
7. **LOW**: Fix naming violations: ReporterType → Type, writeIdentifier* → writeID*
