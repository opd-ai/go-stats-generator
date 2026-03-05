# TASK: Reorganize **test file** structure using package cohesion/coupling metrics — move test code between files without modifying test logic.

## Execution Mode
**Autonomous action** — reorganize test files one sub-package at a time, validate with tests and diff.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 1: Baseline
```bash
go-stats-generator analyze . --only-tests --format json --output baseline.json --sections packages,functions,duplication
```

### Phase 2: Identify and Reorganize
1. From baseline JSON, identify packages with poor test organization:
   - Test files with >500 lines or >20 test functions
   - Multiple test files testing the same function group
   - Test helpers scattered across many test files
2. Select the worst-organized package (largest test files first).
3. Analyze the test file structure:
   - What test functions are defined in each file?
   - Which test helpers are shared vs. file-local?
   - Are there test files with mixed concerns (testing unrelated functions)?
4. Apply reorganization moves:
   - **Group by tested function**: move tests into files named `<feature>_test.go`.
   - **Consolidate helpers**: move shared test helpers into `testutil_test.go`.
   - **Split large test files**: break test files with >500 lines into focused files.
   - **Consolidate tiny test files**: merge test files with <20 lines.
5. Rules for moving test code:
   - Only move entire test functions, test helpers, or test fixtures.
   - Do NOT modify test logic, assertions, or expected values.
   - Ensure all moved helpers use `t.Helper()`.
   - Maintain test file naming conventions (`*_test.go`).
6. Run `go build ./...` and `go test -race ./...` after each package reorganization.

### Phase 3: Validate
```bash
go-stats-generator analyze . --only-tests --format json --output post.json --sections packages,functions,duplication
go-stats-generator diff baseline.json post.json
```
Confirm: test organization improved, zero test regressions, no logic changes.

## Thresholds (Test-Appropriate)
| Metric | Critical | Warning | Target |
|--------|----------|---------|--------|
| Test file size | >500 lines | >300 lines | <300 lines |
| Test functions per file | >20 | >15 | <=15 |
| Test helper duplication | >10% | >5% | <5% |

## Move Rules
- Only move test code — never modify test logic, assertions, or production code.
- Each move must improve at least one metric (file size, organization, duplication).
- Preserve all existing test behavior.
- After reorganization, all tests must pass.
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
Reorganize the package with the largest test files first. If tied, most test functions. If still tied, alphabetical.
## Test File Organization Patterns
| Pattern | When to Apply |
|---------|---------------|
| testutil_test.go | When helpers are duplicated across test files |
| <feature>_test.go | When a test file tests functions from multiple features |
| Split by test type | When unit tests and integration tests are mixed |

## Validation Checklist
- [ ] Test file organization improved
- [ ] All tests pass with -race flag
- [ ] No test logic changes — only code movement
- [ ] All imports updated correctly
- [ ] `go build ./...` succeeds
