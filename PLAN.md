# Implementation Plan: Phase 2.2–2.4 Naming Convention Analysis

## Phase Overview
- **Objective**: Complete identifier name quality scoring, package name analysis, and naming metrics integration
- **Source Document**: `ROADMAP.md` (Phase 2: Naming Convention Analysis, Steps 2.2–2.4)
- **Prerequisites**: Phase 2.1 (File Name Linting) — Complete ✅
- **Estimated Scope**: Medium

## Implementation Steps

### 1. Implement Identifier Name Quality Scoring (Step 2.2) ✅ COMPLETE

**Status**: All sub-steps implemented and tested with >95% coverage

#### 1.1 Add MixedCaps Verification ✅
- **Deliverable**: `checkMixedCaps(name string, isTest bool) *metrics.IdentifierViolation` function
- **Implementation**: Verifies identifiers use MixedCaps (no underscores except test functions `Test_*`)
- **Testing**: Comprehensive unit tests covering valid/invalid cases

#### 1.2 Add Single-Letter Name Detection ✅
- **Deliverable**: `checkSingleLetterName(name string, idType string, context IdentifierContext) *metrics.IdentifierViolation` function
- **Implementation**: Flags single-letter names outside short loops (`i`, `j`, `k`) and receivers (`s`, `r`, `w`)
- **Testing**: Verified with table-driven tests

#### 1.3 Add Acronym Casing Checker ✅
- **Deliverable**: `checkAcronymCasing(name string) *metrics.IdentifierViolation` function
- **Implementation**: Detects common acronyms (URL, HTTP, ID, API, JSON, XML, SQL, HTML, CSS, EOF, IP, TCP, UDP, RPC, TLS, SSL, GRPC, UI, URI, UUID, ASCII, UTF) and flags incorrect casing
- **Testing**: Tests for "Url" → "URL", "Http" → "HTTP", "Id" → "ID", etc.

#### 1.4 Add Stuttering Detection for Identifiers ✅
- **Deliverable**: `checkIdentifierStuttering(name string, packageName string, receiverType string) *metrics.IdentifierViolation` function
- **Implementation**: Flags `User.UserName`, `user.UserService`, method name stuttering
- **Testing**: Verified stuttering detection with various patterns

#### 1.5 Add Name Quality Scoring ✅
- **Deliverable**: `ComputeIdentifierQualityScore(name string, violations []metrics.IdentifierViolation) float64` method
- **Implementation**: Base score using severity weights (low=0.1, medium=0.3, high=0.5)
- **Testing**: Score calculation verified with different violation counts

#### 1.6 Add AST Walking for Identifiers ✅
- **Deliverable**: `AnalyzeIdentifiers(file *ast.File, filePath string, fset *token.FileSet) []metrics.IdentifierViolation` method
- **Implementation**: 
  - Walks AST to collect all function, method, type, const, and var declarations
  - Applies all identifier checks to each declaration
  - Tracks context (loop scope, receiver position) during walk
  - Integration test with sample Go code validates end-to-end functionality
- **Testing**: Integration test covers multiple violation types in realistic code

**Coverage**: naming.go overall 81.5%, new methods 68-100%

### 2. Implement Package Name Analysis (Step 2.3)

TODO: Not yet implemented
- **Deliverable**: New methods in `internal/analyzer/naming.go` for package name validation
- **Dependencies**: None

#### 2.1 Add Package Name Convention Checker
- **Deliverable**: `AnalyzePackageName(pkgName string, dirName string, filePath string) []metrics.PackageNameViolation` method
- **Logic**:
  - Verify lowercase, single word preferred, no underscores or mixedCaps
  - Flag generic names: `util`, `common`, `base`, `shared`, `lib`, `core`, `misc`, `helpers`
  - Flag standard library collisions: `http`, `fmt`, `io`, `os`, `net`, `sync`, `time`, `strings`, `bytes`

#### 2.2 Add Directory Name Mismatch Detection
- **Deliverable**: Check within `AnalyzePackageName` that package name matches directory name
- **Logic**:
  - Compare `package <name>` declaration with directory basename
  - Flag mismatches with suggested rename

### 3. Integrate Naming Metrics into Reporting (Step 2.4)
- **Deliverable**: Updated `NamingMetrics` population and output format integration
- **Dependencies**: Steps 1 and 2

#### 3.1 Update Analyze Command Integration
- **Deliverable**: Modify `cmd/analyze.go` to call identifier and package name analysis
- **Logic**:
  - After file parsing, run `AnalyzeIdentifiers` on each AST
  - Run `AnalyzePackageName` for each unique package
  - Aggregate results into `report.Naming`

#### 3.2 Add Console Reporter Section
- **Deliverable**: Naming violations section in `internal/reporter/console.go`
- **Logic**:
  - Table with columns: Name, File, Line, Type, Violation, Suggested Fix
  - Summary line with violation counts and overall naming score

#### 3.3 Add JSON/HTML Reporter Fields
- **Deliverable**: Ensure `NamingMetrics` serializes correctly in JSON and renders in HTML
- **Logic**:
  - JSON: Verify `IdentifierIssues` and `PackageNameIssues` arrays serialize
  - HTML: Add naming section to template in `internal/reporter/html.go`

### 4. Add Configuration Options
- **Deliverable**: Configuration keys in `.go-stats-generator.yaml` parsing
- **Dependencies**: None

#### 4.1 Add Config Struct Fields
- **Deliverable**: Update `internal/config/config.go` with naming config options
- **Logic**:
  - `naming.flag_generic_filenames` (bool, default: true)
  - `naming.flag_stuttering` (bool, default: true)
  - `naming.min_name_length` (int, default: 2)
  - `naming.check_acronyms` (bool, default: true)

#### 4.2 Wire Config to Analyzer
- **Deliverable**: `NamingAnalyzer` accepts config and respects settings
- **Logic**:
  - Skip checks based on config flags
  - Use `min_name_length` for single-letter detection threshold

### 5. Add Comprehensive Tests
- **Deliverable**: Test coverage >95% for new naming analysis code
- **Dependencies**: Steps 1–4

#### 5.1 Unit Tests for Identifier Checks
- **Deliverable**: Table-driven tests in `internal/analyzer/naming_test.go`
- **Test Cases**:
  - MixedCaps violations: `get_user`, `Get_User`, `getUserID` (valid)
  - Single-letter names in/out of context
  - Acronym casing: `Url` → `URL`, `HttpClient` → `HTTPClient`
  - Stuttering: `user.NewUser`, `User.GetUser`

#### 5.2 Unit Tests for Package Name Checks
- **Deliverable**: Table-driven tests for package name analysis
- **Test Cases**:
  - Generic names: `util`, `common`, `helpers`
  - Standard library collisions: `http`, `fmt`
  - Directory mismatches: `package foo` in `/bar/` directory

#### 5.3 Integration Tests
- **Deliverable**: End-to-end tests with sample Go files in `testdata/naming/`
- **Test Cases**:
  - `testdata/naming/bad_identifiers.go` — file with various violations
  - `testdata/naming/good_identifiers.go` — file with no violations
  - Verify analyze command output includes naming section

## Technical Specifications

- **AST Traversal**: Use `ast.Inspect` with type switch for `*ast.FuncDecl`, `*ast.GenDecl` (with `*ast.TypeSpec`, `*ast.ValueSpec`)
- **Acronym List**: Hardcoded map of common Go acronyms with correct casing; extensible via config in future
- **Scoring Formula**: `score = 1.0 - (sum(violation_weights) / total_identifiers)`, clamped to [0, 1]
- **Severity Weights**: `low=0.1`, `medium=0.3`, `high=0.5` (consistent with existing file name scoring)
- **Context Tracking**: Use a stack-based approach during AST walk to track loop scope and receiver context

## Validation Criteria

- [ ] `go test ./internal/analyzer/... -v` passes with no failures
- [ ] Test coverage for `naming.go` ≥95% as reported by `go test -cover`
- [ ] `go-stats-generator analyze ./testdata/naming` produces expected violations
- [ ] JSON output includes populated `naming.identifier_issues` and `naming.package_name_issues` arrays
- [ ] Console output displays naming violations table when violations exist
- [ ] HTML report includes naming section with violation details
- [ ] No performance regression: analyze 1000 files in <5s on standard hardware
- [ ] All new config options documented in `.go-stats-generator.yaml` example in README

## Known Gaps

- **Git Blame Integration**: Determining annotation age (Phase 4.3) requires git integration not yet implemented; this is a Phase 4 dependency, not blocking Phase 2
- **Custom Acronym List**: The acronym list is hardcoded; future enhancement could allow user-defined acronyms via config
