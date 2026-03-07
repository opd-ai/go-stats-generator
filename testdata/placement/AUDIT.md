# Audit: testdata/placement
**Date**: 2026-03-04
**Status**: Needs Work

## Summary
The `testdata/placement` package is a test fixture designed to demonstrate code placement analysis scenarios (high cohesion, low cohesion, misplaced functions, and misplaced methods). While the code serves its intended purpose as test data, it exhibits significant documentation coverage gaps (36.8% vs ≥70% threshold) and intentional placement anti-patterns. No critical runtime issues exist, but the package demonstrates textbook examples of code organization problems that the `go-stats-generator` tool is designed to detect.

## go-stats-generator Metrics
| Metric               | Value   | Threshold | Status |
|----------------------|---------|-----------|--------|
| Doc Coverage         | 36.8%   | ≥70%      | ✗      |
| Max Cyclomatic       | 3       | ≤10       | ✓      |
| Max Function Length  | 8 lines | ≤30 lines | ✓      |
| Test Coverage        | 0.0%    | ≥65%      | ✗      |
| Duplication Ratio    | 0.0%    | ≤5%       | ✓      |
| Naming Violations    | 1       | 0         | ✗      |

## Issues Found

### High Priority
- [x] **high** documentation — Package missing doc.go file (no package-level documentation)
- [ ] **high** documentation — Overall doc coverage 36.8% below 70% threshold (missing godoc for 16 of 22 functions)
- [ ] **high** placement — ProcessUser function in handler.go should be in database.go (`misplaced_function/handler.go:19`)
- [ ] **high** placement — ValidateUser function in handler.go should be in database.go (`misplaced_function/handler.go:7`)
- [ ] **high** placement — ProcessAll function in handlers.go should be in mixed.go (`low_cohesion/handlers.go:24`)
- [ ] **high** test — Zero test coverage (0.0% vs ≥65% threshold)

### Medium Priority
- [ ] **med** documentation — User struct in misplaced_method/user.go has only 26-char comment (`misplaced_method/user.go:4`)
- [ ] **med** documentation — User struct in low_cohesion/mixed.go has only 26-char comment (`low_cohesion/mixed.go:7`)
- [ ] **med** documentation — Product struct has only 32-char comment (`low_cohesion/mixed.go:13`)
- [ ] **med** naming — Package name "placement" does not match directory name "misplaced_method" (`misplaced_method/`)
- [ ] **med** placement — User.Activate method misplaced in high_cohesion/user.go (receiver User in misplaced_method/user.go) (`high_cohesion/user.go:26`)
- [ ] **med** placement — User.Deactivate method misplaced in high_cohesion/user.go (receiver User in misplaced_method/user.go) (`high_cohesion/user.go:30`)
- [ ] **med** placement — User.Display method misplaced in high_cohesion/user.go (receiver User in misplaced_method/user.go) (`high_cohesion/user.go:35`)
- [ ] **med** placement — User.ValidateEmail method misplaced in high_cohesion/user.go (receiver User in misplaced_method/user.go) (`high_cohesion/user.go:44`)
- [ ] **med** placement — User.IsActive method misplaced in high_cohesion/user.go (receiver User in misplaced_method/user.go) (`high_cohesion/user.go:49`)
- [ ] **med** placement — User.IsAdmin method misplaced in validator.go (receiver User in misplaced_method/user.go) (`misplaced_method/validator.go:15`)
- [ ] **med** placement — User.Validate method misplaced in validator.go (receiver User in misplaced_method/user.go) (`misplaced_method/validator.go:7`)

### Low Priority
- [ ] **low** cohesion — Package has low cohesion score (1.1 vs typical ≥2.0 for healthy packages)
- [ ] **low** cohesion — mixed.go file has 0.00 cohesion (all functions reference external symbols) (`low_cohesion/mixed.go`)
- [ ] **low** cohesion — handler.go file has 0.00 cohesion (`misplaced_function/handler.go`)
- [ ] **low** cohesion — handlers.go file has 0.17 cohesion (`low_cohesion/handlers.go`)
- [ ] **low** documentation — Multiple struct comments under 50 characters (User, Product, Order, Database)
- [ ] **low** maintenance — 20 magic numbers detected (literals in code without named constants)

## Concurrency Assessment
**Goroutines**: 0 instances  
**Channels**: 0 instances  
**Sync Primitives**: None detected  
**Race Detector**: PASS (no concurrent code)  

The package contains no concurrent code patterns. All functions are synchronous and do not use goroutines, channels, or synchronization primitives.

## Dependencies
**External Dependencies**: None (pure test fixture package)  
**Internal Imports**: `fmt`, `strings` (standard library only)  
**Cohesion Score**: 1.1 (low — indicates functions don't reference each other within files)  
**Coupling Score**: 0.0 (no external package dependencies)  
**Circular Dependencies**: None detected  

The package intentionally demonstrates low cohesion through scattered struct definitions and misplaced methods across 7 files in 4 subdirectories (high_cohesion, low_cohesion, misplaced_function, misplaced_method).

## Struct Design Analysis
- **7 structs total**: User (defined 3x in different files), Product, Order, Database
- **Duplicate User definitions**: Defined in 3 separate files (high_cohesion/user.go, low_cohesion/mixed.go, misplaced_method/user.go) with different field sets
- **Methods scattered**: 9 User methods spread across 3 files instead of co-located with receiver type
- **High cohesion example**: high_cohesion/user.go properly groups User struct with all related methods
- **Low cohesion example**: misplaced_method/user.go defines User struct but methods live in validator.go

## Placement Analysis Details
**Misplaced Functions**: 6 instances  
**Misplaced Methods**: 7 instances  
**Low Cohesion Files**: 4 files  
**Average File Cohesion**: 0.31 (target: ≥0.60)  

Top misplaced items by affinity gain:
1. `ProcessUser` in handler.go → should move to database.go (+1.00 affinity)
2. `ProcessAll` in handlers.go → should move to mixed.go (+1.00 affinity)
3. `ValidateUser` in handler.go → should move to database.go (+0.67 affinity)

All misplaced methods involve User type methods defined in files other than the User struct definition file, violating Go convention of co-locating receiver methods with type definitions.

## Recommendations
1. **Add package documentation**: Create doc.go with package-level godoc explaining the test fixture purpose and scenarios demonstrated
2. **Consolidate User definitions**: The package intentionally has 3 different User struct definitions across files — document this is intentional test data, not production code
3. **Add comprehensive godoc**: Increase function/method documentation to meet 70% threshold (currently 27.3% for functions, 66.7% for methods)
4. **Add test coverage**: Create placement_test.go to validate the placement detection scenarios and bring coverage above 65%
5. **Document naming violation**: Add comment in misplaced_method/ explaining why package name differs from directory name (intentional for testing)
6. **Fix or document placement issues**: Either fix the 13 misplaced functions/methods or add comments explaining they are intentionally misplaced for test purposes
