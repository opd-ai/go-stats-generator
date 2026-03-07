# Audit: internal/storage
**Date**: 2026-03-04
**Status**: Needs Work

## Summary
The `internal/storage` package provides persistence for historical metrics through SQLite, JSON file, and in-memory backends. Overall implementation is solid with good error handling and proper context usage. However, the package exceeds complexity and function length thresholds in 4 critical functions and has below-threshold test coverage at 49.2%. WASM stub implementations are properly documented.

## go-stats-generator Metrics
| Metric               | Value   | Threshold | Status |
|----------------------|---------|-----------|--------|
| Doc Coverage         | 87.5%   | ≥70%      | ✓      |
| Max Cyclomatic       | 14      | ≤10       | ✗      |
| Max Function Length  | 63      | ≤30 lines | ✗      |
| Test Coverage        | 49.2%   | ≥65%      | ✗      |
| Duplication Ratio    | 0.35%   | ≤5%       | ✓      |
| Naming Violations    | 2       | 0         | ✗      |

## Issues Found

### High Priority
- [x] **high** complexity — List function exceeds cyclomatic complexity threshold (`json.go:List`, cyclomatic: 14) — **FIXED**: Refactored to 7 code lines, complexity 2
- [x] **high** function-length — List function exceeds 30 line threshold (`json.go:List`, 63 lines) — **FIXED**: Refactored to 7 code lines
- [x] **high** function-length — Retrieve function exceeds 30 line threshold (`sqlite.go:Retrieve`, 61 lines) — **FIXED**: Refactored to 17 code lines
- [x] **high** function-length — Store function exceeds 30 line threshold (`sqlite.go:Store`, 56 lines) — **FIXED**: Refactored to 16 code lines
- [x] **high** function-length — initSchema function exceeds 30 line threshold (`sqlite.go:initSchema`, 50 lines) — **FIXED**: Refactored to 11 code lines by extracting SQL constants and helper methods
- [x] **high** test-coverage — Package coverage at 49.2% is below 65% threshold — **FIXED**: Created comprehensive sqlite_test.go with 11 test cases covering Store, Retrieve, List, Delete, Cleanup, GetLatest, GetByTag, and filtering. Coverage increased from 49.0% to 81.3%

### Medium Priority
- [ ] **med** naming — StorageConfig type stutters package name (`interface.go:88`) — should be `Config`
- [ ] **med** naming — Package name mismatch with directory (`interface.go:1`) — directory/package name inconsistency
- [ ] **med** duplication — Renamed clone pair detected (`sqlite.go:325-330` and `sqlite.go:507-512`, 6 lines)
- [ ] **med** complexity — matchesFilter functions have cyclomatic complexity of 9 (`json.go:matchesFilter` and `memory.go:matchesFilter`)
- [ ] **med** api-design — MetricsStorage interface has 8 methods, consider breaking into smaller focused interfaces
- [ ] **med** documentation — Package-level doc.go exists but coverage reported as 0.0%

### Low Priority
- [ ] **low** complexity — Store function cyclomatic complexity 9, approaching threshold (`sqlite.go:Store`)
- [ ] **low** complexity — Retrieve function cyclomatic complexity 9, approaching threshold (`sqlite.go:Retrieve`)
- [ ] **low** organization — 3 oversized files detected (sqlite.go, interface.go, json.go)
- [ ] **low** organization — Package has 80 functions across 6 files, consider splitting by backend type

## Concurrency Assessment
**Goroutines**: 0 detected — no goroutine usage in storage layer (synchronous operations)
**Channels**: 0 detected — no channel usage
**Sync Primitives**: 0 mutexes/RWMutexes/WaitGroups detected
**Race Check**: PASS — `go test -race` completed successfully with no data races
**Context Usage**: ✓ All storage operations accept `context.Context` for cancellation support
**Concurrency Safety**: Storage implementations rely on database/filesystem locking; no shared mutable state detected

## Dependencies
**External Dependencies**:
- `modernc.org/sqlite` — Pure-Go SQLite driver for database backend
- `github.com/opd-ai/go-stats-generator/internal/metrics` — Internal metrics types

**Cohesion/Coupling Metrics**:
- Cohesion Score: 3.17 (moderate — functions grouped by backend implementation)
- Coupling Score: 1.0 (low — minimal external dependencies)
- Circular Dependencies: 0 detected

**Import Analysis**:
- Well-isolated package with minimal external dependencies
- Proper use of build tags (`//go:build !js || !wasm`) to exclude platform-specific code
- WASM stubs properly isolated in `storage_wasm.go`

## Recommendations
1. **Refactor List function** (`json.go`) — Extract filtering logic to reduce cyclomatic complexity from 14 to ≤10 and split into smaller helper functions to reduce from 63 to ≤30 lines
2. **Add comprehensive tests** — Increase test coverage from 49.2% to ≥65% with focus on:
   - SQLite retention policy cleanup edge cases
   - JSON file storage error handling paths
   - Memory storage concurrent access scenarios
3. **Rename StorageConfig** to `Config` (`interface.go:88`) — Eliminate package stuttering per Go conventions
4. **Split oversized functions** in `sqlite.go`:
   - Extract `initSchema` schema creation SQL to constants/table definitions (50 lines → ≤30 lines)
   - Break `Store` transaction logic into smaller helper methods (56 lines → ≤30 lines)
   - Refactor `Retrieve` query and decompression logic (61 lines → ≤30 lines)
5. **Consider interface segregation** — Split `MetricsStorage` (8 methods) into focused interfaces like `SnapshotWriter`, `SnapshotReader`, `SnapshotManager` per Interface Segregation Principle
6. **Extract matchesFilter** — Consolidate duplicate filtering logic in `json.go` and `memory.go` to shared helper (DRY principle)
7. **Add integration tests** — Current 49.2% coverage suggests gaps in SQLite migration, compression, and cleanup logic testing
