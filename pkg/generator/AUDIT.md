# Audit: github.com/opd-ai/go-stats-generator/pkg/go-stats-generator
**Date**: 2026-03-03
**Status**: Needs Work

## Summary
The `pkg/go-stats-generator` package provides the public programmatic API for analyzing Go source code. Overall health is good with strong test coverage (77.1%) and no complexity violations, but documentation coverage (53.8%) falls short of the 70% threshold. The package has 6 files implementing both standard filesystem-based analysis and WASM-compatible in-memory analysis. Critical risk: package naming convention violations due to underscores in the name.

## go-stats-generator Metrics
| Metric               | Value   | Threshold | Status |
|----------------------|---------|-----------|--------|
| Doc Coverage         | 53.8%   | ≥70%      | ✗      |
| Max Cyclomatic       | 5       | ≤10       | ✓      |
| Max Function Length  | 16 lines| ≤30 lines | ✓      |
| Test Coverage        | 77.1%   | ≥65%      | ✓      |
| Duplication Ratio    | 0.0%    | ≤5%       | ✓      |
| Naming Violations    | 6       | 0         | ✗      |

## Issues Found

### Documentation (Medium Priority)
- [ ] **med** documentation — Package-level documentation missing (doc.go:1) — 0% package coverage vs 70% threshold
- [ ] **med** documentation — 21 exported functions missing godoc comments — only 0% function coverage
- [ ] **low** documentation — Overall documentation coverage 53.8% below 70% threshold

### Naming Conventions (High Priority)
- [ ] **high** naming — Package name `go_stats_generator` violates Go conventions (should not contain underscores) — use `gostatsgenerator` instead
- [ ] **high** naming — Package name doesn't match directory name `go-stats-generator` (package_name_issues)
- [ ] **med** naming — Identifier stuttering: `analyzeResults` method should be renamed to `Results` (api_common.go:35)
- [ ] **med** naming — Type stuttering: `analyzerSet` should be renamed to `Set` (api_common.go:61)
- [ ] **low** naming — Generic file name: `types.go` too generic, should describe content (types.go)
- [ ] **low** naming — Generic file name: `errors.go` too generic, should describe content (errors.go)

### API Design (Low Priority)
- [ ] **low** api — Low package cohesion score (1.17) suggests functions may not be tightly related — consider splitting into focused sub-packages
- [ ] **low** api — 4 dependencies create moderate coupling (coupling score: 2.0)

### Code Organization (Low Priority)
- [ ] **low** organization — File `types.go` flagged as oversized with organization burden 0.40
- [ ] **low** organization — Suggested refactoring: Move `Report` function to `api_common.go` for better cohesion (ROI: 12.31)

## Concurrency Assessment
**Goroutine Patterns**: 1 anonymous goroutine detected (api.go:53)
**Channel Usage**: 7 total channels (2 buffered, 5 unbuffered, 3 directional)
**Sync Primitives**: None detected (no mutexes, wait groups, or atomic operations)
**Race Check**: PASS — no data races detected in tests
**Safety Assessment**: Concurrency usage is minimal and appears safe. The single goroutine and channel usage are for result streaming in the analysis pipeline. No shared state protection issues identified.

## Dependencies
**External Dependencies** (4 total):
- `github.com/opd-ai/go-stats-generator/internal/metrics` — core metrics types
- `github.com/opd-ai/go-stats-generator/internal/scanner` — file discovery and parsing
- `github.com/opd-ai/go-stats-generator/internal/analyzer` — analysis implementations
- `github.com/opd-ai/go-stats-generator/internal/config` — configuration management

**Circular Import Risk**: None detected
**Cohesion Score**: 1.17 (low — suggests loosely related functionality)
**Coupling Score**: 2.0 (moderate — reasonable dependency count)
**Assessment**: All dependencies are internal packages, which is appropriate for a public API facade. The moderate coupling is justified given this is the main entry point that orchestrates multiple internal components.

## Recommendations
1. **HIGH PRIORITY** — Rename package from `go_stats_generator` to `gostats` or match directory name (currently `go-stats-generator` uses hyphen, not underscore)
2. **HIGH PRIORITY** — Add package-level godoc comment in `doc.go` to explain the public API surface
3. **MEDIUM PRIORITY** — Document all 21 exported functions with godoc comments to reach 70% threshold
4. **MEDIUM PRIORITY** — Rename `analyzeResults` to `Results` and `analyzerSet` to `Set` to eliminate stuttering
5. **LOW PRIORITY** — Rename `types.go` to `api_types.go` and `errors.go` to `api_errors.go` for clarity
