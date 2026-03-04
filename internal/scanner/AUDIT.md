# Audit: internal/scanner
**Date**: 2026-03-03
**Status**: Needs Work

## Summary
The `internal/scanner` package provides file discovery and concurrent processing capabilities for the go-stats-generator tool. Overall health is good with 90% documentation coverage and no duplication, but has 1 function exceeding complexity/length thresholds and 3 naming violations. Critical risk: ProcessFiles function exceeds both complexity (8) and length (31 lines) thresholds, potentially impacting maintainability.

## go-stats-generator Metrics
| Metric               | Value   | Threshold | Status |
|----------------------|---------|-----------|--------|
| Doc Coverage         | 90.0%   | ≥70%      | ✓      |
| Max Cyclomatic       | 8       | ≤10       | ✓      |
| Max Function Length  | 31      | ≤30 lines | ✗      |
| Test Coverage        | 43.8%   | ≥65%      | ✗      |
| Duplication Ratio    | 0.0%    | ≤5%       | ✓      |
| Naming Violations    | 3       | 0         | ✗      |

## Issues Found
- [x] high — function length — ProcessFiles exceeds 30-line threshold (`worker.go:50`, 31 code lines)
- [x] med — test coverage — Package coverage 43.8% below 65% threshold
- [x] med — naming — Generic file name "types.go" lacks descriptive context (`types.go`)
- [x] med — naming — Generic file name "helpers.go" lacks descriptive context (`helpers.go`)
- [x] low — naming — Package name "scanner" has directory mismatch violation (`doc.go`)

## Concurrency Assessment
**Patterns Detected:**
- Worker pool pattern (worker.go:65) with 6 goroutines, channels, and WaitGroup — confidence: 1.0
- Pipeline pattern (worker.go:65) with 6 stages and 27 channels — confidence: 1.1

**Goroutines:** 7 total (5 anonymous, 2 named)
- worker.go:65 — wp.worker (named)
- worker.go:217 — bp.processBatchesAsync (named)
- worker.go:75, 87, 145, 168 — anonymous functions
- worker_wasm.go:50 — anonymous function

**Channels:** 32 total (6 buffered, 26 unbuffered, 16 directional)
- Proper use of buffered channels for job/result queuing
- Good use of directional channels for type safety

**Sync Primitives:**
- WaitGroup (worker.go:62) for worker synchronization

**Race Check:** PASS — `go test -race` completed successfully

**Assessment:** Excellent concurrency implementation with proper worker pool pattern, channel directionality, and synchronization. No potential goroutine leaks detected. Context cancellation handling present in ProcessFiles.

## Dependencies
**Package Dependencies:** 1 internal package
- github.com/opd-ai/go-stats-generator/internal/config

**Cohesion/Coupling:**
- Package cohesion: 1.6 (Low — below 2.0 threshold)
- 7 files, 42 functions across package

**Circular Import Risks:** None detected

**External Dependencies:** Standard library only (go/ast, go/parser, context, sync, path/filepath, strings, fmt, os)

**Assessment:** Clean dependency structure with no external dependencies. Low cohesion score suggests file organization could be improved (consider consolidating related functionality).

## Recommendations
1. **Refactor ProcessFiles (worker.go:50)** — Split into smaller helper functions to reduce length from 31 to ≤30 lines (extract channel setup, worker initialization, progress tracking)
2. **Improve test coverage** — Add tests to increase coverage from 43.8% to ≥65% (focus on helper functions, edge cases in discover.go)
3. **Rename generic files** — Rename `types.go` to `file_info.go` and `helpers.go` to `file_filters.go` for better discoverability
4. **Address package cohesion** — Consider consolidating or splitting files to improve cohesion score from 1.6 to ≥2.0
5. **Add package-level documentation** — Ensure doc.go has comprehensive package comment explaining scanner purpose and usage patterns
