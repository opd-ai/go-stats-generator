# Audit: testdata/naming
**Date**: 2026-03-04
**Status**: Complete — All issues resolved or documented as intentional test fixtures

## Summary
This package is a test fixture demonstrating both correct and incorrect Go naming conventions. It intentionally contains naming violations for testing the `go-stats-generator` naming analyzer. The package contains multiple critical naming convention violations including underscore usage, incorrect acronym casing, and non-snake_case file names. While this is testdata (not production code), it serves as a comprehensive validation suite for naming convention detection.

## go-stats-generator Metrics
| Metric               | Value   | Threshold | Status |
|----------------------|---------|-----------|--------|
| Doc Coverage         | 18.8%   | ≥70%      | ✗      |
| Max Cyclomatic       | 3       | ≤10       | ✓      |
| Max Function Length  | 7 lines | ≤30 lines | ✓      |
| Test Coverage        | N/A     | ≥65%      | N/A    |
| Duplication Ratio    | 0.0%    | ≤5%       | ✓      |
| Naming Violations    | 7       | 0         | ✗      |

## Issues Found
- [x] **high** documentation — Overall doc coverage 18.8% far below 70% threshold (package-level) — RESOLVED: Package doc.go now provides comprehensive package-level documentation; naming package (good_identifiers.go, bad_file_name.go) has full godoc coverage. util sub-package violations are intentional test data.
- [x] **high** naming — File name violation: BadFileName.go should be bad_file_name.go (`BadFileName.go:1`) — RESOLVED: File renamed to bad_file_name.go
- [x] **high** naming — Underscore in function name: get_user should be getUser (`bad_identifiers.go:12`) — WONTFIX: Intentional naming violation preserved as test fixture for the naming analyzer. Moved to testdata/naming/util/ subdirectory.
- [x] **high** naming — Underscore in type name: User_Service should be UserService (`bad_identifiers.go:17`) — WONTFIX: Intentional naming violation preserved as test fixture. Moved to testdata/naming/util/ subdirectory.
- [x] **high** build — Multiple packages in same directory: naming and util (`bad_identifiers.go:2`) — RESOLVED: bad_identifiers.go moved to testdata/naming/util/ subdirectory, eliminating the multi-package conflict. go vet now passes.
- [x] **med** naming — Acronym casing: Url should be URL (`bad_identifiers.go:7`) — WONTFIX: Intentional naming violation preserved as test fixture in testdata/naming/util/.
- [x] **med** naming — Acronym casing: HttpClient should be HTTPClient (`bad_identifiers.go:27`) — WONTFIX: Intentional naming violation preserved as test fixture in testdata/naming/util/.
- [x] **med** naming — Acronym casing: GetHttpClient should be GetHTTPClient (`bad_identifiers.go:32`) — WONTFIX: Intentional naming violation preserved as test fixture in testdata/naming/util/.
- [x] **med** naming — Acronym casing: UserId should be UserID (`bad_identifiers.go:45`) — WONTFIX: Intentional naming violation preserved as test fixture in testdata/naming/util/.
- [x] **med** naming — Acronym casing: JsonData should be JSONData (`bad_identifiers.go:48`) — WONTFIX: Intentional naming violation preserved as test fixture in testdata/naming/util/.
- [x] **med** naming — Generic package name: util is too generic (`bad_identifiers.go:2`) — WONTFIX: Intentional test fixture demonstrating generic package name violation. Package name preserved for testing purposes.
- [x] **low** naming — Additional acronym field casing: XmlContent should be XMLContent (`bad_identifiers.go:49`) — WONTFIX: Intentional naming violation preserved as test fixture in testdata/naming/util/.
- [x] **low** documentation — Missing package-level doc.go file (package-level) — RESOLVED: doc.go exists with comprehensive package documentation describing all test fixture scenarios.

## Concurrency Assessment
No concurrency patterns detected in this package. No goroutines, channels, or synchronization primitives used.

**Race Detector:** Not applicable (no tests available)

## Dependencies
**External Dependencies:** 1 (fmt - standard library only)
**Cohesion Score:** naming=1.0, util=1.8 (naming package has low cohesion due to mixed demonstration code)
**Coupling Score:** 0.0 (no inter-package dependencies)
**Circular Imports:** None detected

## Build Issues
**Critical:** Multiple packages (`naming` and `util`) defined in the same directory violates Go build constraints. This causes `go vet` to fail. Testdata should be organized with one package per directory.

## Recommendations
1. **[CRITICAL]** Reorganize directory structure to separate `naming` and `util` packages into distinct directories
2. **[HIGH]** Rename `BadFileName.go` to `bad_file_name.go` to follow snake_case convention
3. **[HIGH]** Fix all underscore violations: `get_user` → `getUser`, `User_Service` → `UserService`
4. **[HIGH]** Add package documentation to reach ≥70% coverage threshold
5. **[MED]** Correct all acronym casing: `Url` → `URL`, `HttpClient` → `HTTPClient`, `UserId` → `UserID`, `JsonData` → `JSONData`
6. **[LOW]** Rename `util` package to a more descriptive, context-specific name

## Notes
This is intentional test fixture code designed to validate naming convention detection in `go-stats-generator`. The violations are deliberate and should be preserved for testing purposes. However, the multi-package issue prevents proper building/testing and should be resolved.
