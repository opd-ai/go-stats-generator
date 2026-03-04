# Audit: internal/api
**Date**: 2026-03-04
**Status**: Needs Work

## Summary
The `internal/api` package provides REST API handlers and storage abstractions for go-stats-generator analysis results. Overall package health is good with strong documentation coverage (81.0%) and moderate test coverage (65.5% for api, 22.6% for storage sub-package), but contains several high-priority issues including swallowed JSON marshaling errors and a goroutine leak risk. The storage sub-package requires significant test coverage improvements.

## go-stats-generator Metrics
| Metric               | Value   | Threshold | Status |
|----------------------|---------|-----------|--------|
| Doc Coverage         | 81.0%   | ≥70%      | ✓      |
| Max Cyclomatic       | 8       | ≤10       | ✓      |
| Max Function Length  | 28      | ≤30 lines | ✓      |
| Test Coverage (api)  | 65.5%   | ≥65%      | ✓      |
| Test Coverage (storage) | 22.6% | ≥65%    | ✗      |
| Duplication Ratio    | 0.0%    | ≤5%       | ✓      |
| Naming Violations    | 2       | 0         | ✗      |

## Issues Found
- [x] **high** error-handling — JSON marshaling errors silently ignored in storage implementations (`storage/mongo.go:87`, `storage/postgres.go:63`)
- [x] **high** concurrency — Goroutine spawned in HandleAnalyze lacks context cancellation or timeout control (`handlers.go:71`)
- [x] **high** test-coverage — Storage sub-package critically under-tested at 22.6% vs 65% threshold (`internal/api/storage`)
- [x] **med** naming — Generic file name "types.go" violates naming conventions; should describe contents (`types.go`)
- [x] **med** naming — Package name "api" does not match directory structure pattern (`api`)
- [x] **med** error-handling — Factory function silently falls back to memory storage on errors without logging (`storage/factory.go:32-34`, `storage/factory.go:44-46`)
- [x] **med** documentation — Package "api" lacks doc.go with package-level documentation (overall package docs missing)
- [x] **med** documentation — 35% of methods lack documentation (7 of 23 methods undocumented)
- [x] **low** api-design — Deprecated storage.go file adds maintenance burden (`storage.go:3`)
- [x] **low** api-design — ResultStore interface returns bool for Delete but storage implementations don't validate existence consistently
- [x] **low** error-handling — writeJSON ignores json.Encoder error (`handlers.go:162`)

## Concurrency Assessment
**Goroutine Patterns**: 1 goroutine spawned in `HandleAnalyze` (line 71) to run async analysis. No context cancellation mechanism, potential leak if server shutdown occurs during long-running analysis.

**Channel Usage**: None detected.

**Sync Primitives**: Mutexes used in Postgres and Mongo storage implementations for thread safety (appropriate).

**Race Check Result**: PASS — no data races detected in test execution.

## Dependencies
**External Dependencies**:
- `github.com/google/uuid` — UUID generation for analysis IDs (justified)
- `github.com/lib/pq` — PostgreSQL driver (justified for storage backend)
- `go.mongodb.org/mongo-driver/*` — MongoDB driver (justified for storage backend)

**Internal Dependencies** (5):
- `internal/api/storage` (sub-package)
- `internal/config`
- `internal/metrics`
- `pkg/go-stats-generator`

**Cohesion/Coupling Metrics**:
- api package: cohesion 0.8, coupling 2.5 (acceptable)
- storage package: cohesion 1.24, coupling 3.0 (acceptable)

**Circular Import Risks**: None detected.

## Recommendations
1. **[HIGH PRIORITY]** Fix swallowed JSON marshaling errors in `storage/mongo.go:87` and `storage/postgres.go:63` — check and handle errors properly
2. **[HIGH PRIORITY]** Add context cancellation to async analysis goroutine in `handlers.go:71` — accept context from HTTP request, implement graceful shutdown
3. **[HIGH PRIORITY]** Improve storage sub-package test coverage from 22.6% to ≥65% — add comprehensive tests for Postgres, Mongo, and Memory implementations
4. **[MEDIUM PRIORITY]** Add package-level doc.go file explaining API design and usage patterns
5. **[MEDIUM PRIORITY]** Log errors in storage factory fallback paths (`storage/factory.go`) for operational visibility
6. **[LOW PRIORITY]** Rename `types.go` to descriptive name like `request_types.go` or `api_types.go`
7. **[LOW PRIORITY]** Remove deprecated `storage.go` file after ensuring no external usage
8. **[LOW PRIORITY]** Handle json.Encoder error in `writeJSON` (`handlers.go:162`)
