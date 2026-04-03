# TASK: Reorganize **test file** structure using package cohesion/coupling metrics — move test code between files without modifying test logic.

## Execution Mode
**Autonomous action** — reorganize test files one sub-package at a time, validate with tests and diff.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Test Infrastructure
1. Read the project README and discover the testing philosophy.
2. Identify existing test organization patterns: how are test files named, are there `testutil_test.go` files, how are fixtures stored?
3. Note whether tests use build tags, test suites, or integration test separation.
4. Discover the project's test naming conventions (e.g., `Test_funcName`, `TestFuncName_scenario`).

### Phase 1: Baseline
```bash
go-stats-generator analyze . --only-tests --format json --output baseline.json --sections packages,functions,duplication
```

### Phase 2: Identify and Reorganize
1. Identify packages with poor test organization:
   - Test files with >500 lines or >20 test functions
   - Multiple test files testing the same function group
   - Test helpers scattered across many test files
2. Select the worst-organized package (largest test files first).
3. Analyze the test file structure and apply reorganization matching the project's conventions:
   - **Group by tested function**: move tests into files named `<feature>_test.go`.
   - **Consolidate helpers**: move shared test helpers into `testutil_test.go`.
   - **Split large test files**: break files with >500 lines into focused files.
   - **Consolidate tiny test files**: merge files with <20 lines.
4. Move rules:
   - Only move entire test functions, helpers, or fixtures.
   - Do NOT modify test logic, assertions, or expected values.
   - Ensure all moved helpers use `t.Helper()`.
   - Maintain `*_test.go` naming conventions.
5. Run `go build ./...` and `go test -race ./...` after each reorganization.

### Phase 3: Validate
```bash
go-stats-generator analyze . --only-tests --format json --output post.json --sections packages,functions,duplication
go-stats-generator diff baseline.json post.json
```

## Default Thresholds (test-appropriate)
| Metric | Critical | Warning | Target |
|--------|----------|---------|--------|
| Test file size | >500 lines | >300 lines | <300 lines |
| Test functions per file | >20 | >15 | <=15 |
| Test helper duplication | >10% | >5% | <5% |

## Move Rules
- Only move test code — never modify test logic or production code.
- Each move must improve at least one metric.
- Preserve all existing test behavior.
- Document each move: `[TestFunc/Helper] [from_file] -> [to_file] — [reason]`

## Output Format
```
Package: [name] (test files: [before] -> [after])
Moves:
  [TestFunc/Helper] [old_file] -> [new_file]
  ...
Tests: PASS
```

## Tiebreaker
Reorganize the package with the largest test files first. If tied, most test functions.
