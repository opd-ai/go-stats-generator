# Audit: testdata/simple
**Date**: 2026-03-04
**Status**: Needs Work

## Summary
The testdata/simple package is a test fixture containing example code demonstrating various complexity levels, concurrency patterns, and interface designs. It contains intentionally complex code (VeryComplexFunction with cyclomatic complexity of 24) that violates multiple thresholds, along with multiple package naming violations. The package is unsuitable for production but serves its purpose as test data for the analyzer. Critical issues include extremely high complexity functions, very low documentation coverage, and multiple package naming violations.

## go-stats-generator Metrics
| Metric               | Value   | Threshold | Status |
|----------------------|---------|-----------|--------|
| Doc Coverage         | 25.7%   | ≥70%      | ✗      |
| Max Cyclomatic       | 24      | ≤10       | ✗      |
| Max Function Length  | 75      | ≤30 lines | ✗      |
| Test Coverage        | N/A     | ≥65%      | N/A    |
| Duplication Ratio    | 0.0%    | ≤5%       | ✓      |
| Naming Violations    | 2       | 0         | ✗      |

## Issues Found
- [x] high complexity — VeryComplexFunction exceeds cyclomatic threshold with 24 (calculator.go:128)
- [x] high complexity — ComplexFunction exceeds cyclomatic threshold with 15 (user.go:120)
- [x] high complexity — VeryComplexFunction exceeds function length threshold with 75 lines (calculator.go:128)
- [x] high complexity — FanOutExample exceeds cyclomatic threshold with 10 (concurrency.go:51)
- [x] high complexity — FanInExample exceeds function length threshold with 49 lines (concurrency.go:60)
- [x] high complexity — ComplexFunction exceeds function length threshold with 36 lines (user.go:120)
- [x] high complexity — FanOutExample exceeds function length threshold with 40 lines (concurrency.go:51)
- [x] high complexity — Statistics exceeds function length threshold with 32 lines (calculator.go)
- [x] high documentation — Overall documentation coverage at 25.7%, far below 70% threshold
- [x] high documentation — Package documentation missing (0% coverage)
- [x] high documentation — Function documentation at 25.0%, below 70% threshold
- [x] med documentation — Method documentation at 21.4%, below 70% threshold
- [x] med documentation — Type documentation at 33.3%, below 70% threshold
- [x] med naming — Package name "concurrency" does not match directory "." (directory_mismatch)
- [x] med naming — Package name "simple" does not match directory "." (directory_mismatch)
- [x] med complexity — VeryComplexFunction has nesting depth of 7 levels
- [x] low organization — File has low cohesion score of 0.00 (user.go)
- [x] low maintenance — 80 magic numbers detected across the package
- [x] low maintenance — 1 dead code function (unreferenced)

## Concurrency Assessment
The package demonstrates multiple concurrency patterns including worker pools (1 instance with 24 goroutines), pipelines (1 instance with 24 stages and 30 channels), and semaphores (1 buffered channel with size 3). Total of 24 goroutines launched, all anonymous. Channel usage includes 30 total channels (27 unbuffered, 3 buffered), primarily int channels. Sync primitives include 5 WaitGroups, 1 Mutex, 1 RWMutex, 1 Once, and 1 Cond. No potential goroutine leaks detected. Race detection not applicable (go vet and go test fail due to multiple package declarations in same directory).

## Dependencies
No external dependencies detected. Zero coupling score indicates no inter-package dependencies. The package structure has two logical packages in the same directory ("simple" and "concurrency"), causing build/test failures. Both packages have low cohesion scores (simple: 1.87, concurrency: 1.60). No circular dependencies detected.

## Recommendations
1. This is intentionally complex test data; issues are by design to test analyzer capabilities
2. Split the directory into separate packages if real-world usage intended (simple/ and concurrency/)
3. Add package-level documentation to improve doc coverage from 0%
4. VeryComplexFunction should be split into smaller helpers to reduce cyclomatic complexity from 24 to ≤10
5. Refactor ComplexFunction to reduce cyclomatic complexity from 15 to ≤10
6. Extract constants for the 80 magic numbers to improve maintainability
7. Add documentation to all exported functions, types, and methods to reach 70% threshold
