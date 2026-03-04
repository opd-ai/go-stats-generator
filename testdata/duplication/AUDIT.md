# Audit: testdata/duplication
**Date**: 2026-03-04
**Status**: Needs Work

## Summary
The `testdata/duplication` directory contains test fixtures for duplication detection feature validation, with 6 separate package definitions intentionally demonstrating various code clone patterns. The code exhibits extremely high duplication (16.52%) by design to test the analyzer's clone detection capabilities. Primary issues include multiple package declarations in the same directory (build error), lack of package-level documentation, and naming convention violations.

## go-stats-generator Metrics
| Metric               | Value   | Threshold | Status |
|----------------------|---------|-----------|--------|
| Doc Coverage         | 86.5%   | ≥70%      | ✓      |
| Max Cyclomatic       | 6       | ≤10       | ✓      |
| Max Function Length  | 21      | ≤30 lines | ✓      |
| Test Coverage        | N/A     | ≥65%      | N/A    |
| Duplication Ratio    | 16.52%  | ≤5%       | ✗      |
| Naming Violations    | 6       | 0         | ✗      |

## Issues Found
- [x] **high** build — Multiple package declarations in single directory causes `go vet` and `go test` build failures (below_threshold.go, duplicate_blocks.go, exact_clone.go, near_clone.go, renamed_clone.go, small_blocks.go)
- [x] **high** duplication — Intentional exact clone pair (20 lines) for testing duplication detection (`duplicate_blocks.go:13-32`, `duplicate_blocks.go:38-57`)
- [x] **high** duplication — 6 total clone pairs detected, 96 duplicated lines, 16.52% duplication ratio (threshold: ≤5%)
- [x] **med** naming — Package name `belowthreshold` does not match directory name `duplication` (`below_threshold.go`)
- [x] **med** naming — Package name `exactclone` does not match directory name `duplication` (`exact_clone.go`)
- [x] **med** naming — Package name `nearclone` does not match directory name `duplication` (`near_clone.go`)
- [x] **med** naming — Package name `renamedclone` does not match directory name `duplication` (`renamed_clone.go`)
- [x] **med** naming — Package name `smallblocks` does not match directory name `duplication` (`small_blocks.go`)
- [x] **med** naming — Package name `duplication` technically matches directory, but 5 other packages conflict (`duplicate_blocks.go`)
- [x] **med** documentation — All 6 packages missing package-level documentation (0% package doc coverage)
- [x] **low** cohesion — Package `duplication` has low cohesion score of 1.0 (threshold: ≥2.0)
- [x] **low** cohesion — Package `exactclone` has low cohesion score of 1.4 (threshold: ≥2.0)

## Concurrency Assessment
**No concurrency patterns detected:**
- 0 goroutines (anonymous or named)
- 0 channels (buffered or unbuffered)
- 0 sync primitives (mutexes, wait groups, etc.)
- 0 worker pools, pipelines, or fan-out/fan-in patterns
- **Race check result**: FAIL (build error prevents test execution)

This is expected for test fixture data that focuses on demonstrating duplication patterns rather than concurrent behavior.

## Dependencies
**External dependencies**: None
- All 6 packages have 0 external dependencies
- Coupling score: 0.0 (fully isolated)
- **Circular import risks**: None detected

**Package cohesion scores:**
- `belowthreshold`: 2.4 (acceptable)
- `nearclone`: 2.2 (acceptable)
- `renamedclone`: 2.0 (threshold)
- `smallblocks`: 2.0 (threshold)
- `exactclone`: 1.4 (low)
- `duplication`: 1.0 (low)

## Root Cause Analysis
This directory is **intentionally designed as test fixtures** to validate the duplication detection feature of `go-stats-generator`. The high duplication ratio, multiple package declarations, and naming violations are **deliberate test cases**, not production code defects.

However, the build errors from multiple package declarations in a single directory prevent standard Go tooling (`go vet`, `go test`) from functioning, which limits the usability of these test fixtures.

## Recommendations
1. **[HIGH PRIORITY]** Reorganize test fixtures: Move each package into its own subdirectory (`testdata/duplication/belowthreshold/`, `testdata/duplication/exactclone/`, etc.) to eliminate build errors and enable `go vet`/`go test` execution
2. **[MEDIUM PRIORITY]** Add package-level documentation to each test fixture explaining its purpose (e.g., "Package belowthreshold contains test fixtures with duplication below detection thresholds")
3. **[LOW PRIORITY]** Consider renaming packages to match their subdirectories after reorganization, or add a README explaining why naming conventions are intentionally violated for testing purposes
4. **[LOW PRIORITY]** Document the expected duplication metrics for each test fixture as part of the test validation suite

## Test Fixture Validation
As test fixtures, these files successfully demonstrate:
- ✓ Exact clone detection (20-line clone pair in `duplicate_blocks.go`)
- ✓ Multiple exact clone variations (6 clone pairs total)
- ✓ Near-clone patterns (`near_clone.go`)
- ✓ Renamed clone patterns (`renamed_clone.go`)
- ✓ Below-threshold duplication scenarios (`below_threshold.go`)
- ✓ Small block scenarios (`small_blocks.go`)

**Baseline snapshot created**: `testdata_duplication` (2026-03-04 00:56:24)
