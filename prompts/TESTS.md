# TASK: Generate unit tests targeting the highest-complexity untested functions to achieve >=80% coverage.

## Execution Mode
**Autonomous action** — generate tests, validate coverage improvement.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Test Strategy
1. Read the project README to understand what the project does and its expected behavior.
2. Discover the test framework in use: `testing` only, `testify`, `gomock`, etc.
3. Identify the project's existing test conventions: naming, assertion style, table-driven patterns, setup/teardown.
4. Note whether the project uses `t.Parallel()`, `t.Helper()`, `t.Cleanup()`.
5. Check for existing test fixtures in any `testdata/` directories.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output baseline.json --sections functions
go test -cover ./... 2>&1 | tee coverage-baseline.txt
```

### Phase 2: Generate Tests
1. From baseline JSON, extract all functions sorted by cyclomatic complexity descending.
2. Cross-reference with `go test -cover` output to identify untested or low-coverage packages.
3. Prioritize test generation (tunable defaults):
   - CRITICAL: cyclomatic >15, no test coverage → test first
   - HIGH: cyclomatic 9–15, partial coverage → add edge case tests
   - MEDIUM: cyclomatic 5–8, low coverage → add basic tests
4. For each target function, generate tests matching the project's conventions:
   - **Table-driven tests** for functions with multiple input/output combinations.
   - **Error path tests** for functions that return errors.
   - **Boundary tests** for functions with numeric parameters.
   - **Nil input tests** for functions accepting pointers/slices/maps.
5. Test conventions (match project's existing patterns):
   - Use the same assertion library the project already uses.
   - Name tests `Test[FunctionName]_[Scenario]`.
   - Use `t.Helper()` in test helpers, `t.Parallel()` where safe.
   - Use `t.Run` for subtests in table-driven tests.
6. Run `go test -race ./...` after each test file.

### Phase 3: Validate
```bash
go test -cover ./... 2>&1 | tee coverage-post.txt
go-stats-generator analyze . --skip-tests --format json --output post.json --sections functions
```
Confirm: coverage increased, all tests pass, no flaky tests.

## Default Coverage Targets (calibrate to project)
| Package Type | Target |
|--------------|--------|
| Core logic | >=80% |
| CLI commands | >=70% |
| Utility packages | >=80% |
| Reporter/output | >=60% |

## Test Generation Rules
- Only use testing framework(s) the project already depends on.
- Each test must be deterministic (no random data, no time-dependent assertions).
- Each test must clean up after itself (temp files, goroutines).
- Test the public API surface first, then critical unexported functions.
- Do not test auto-generated code or trivial getters/setters.

## Test Quality Standards
- Each test has a clear, descriptive name.
- Error path tests verify both the error value and that no partial state leaked.
- Table-driven tests have at least 3 cases: happy path, edge case, error case.

## Output Format
```
[package]: coverage [before]% -> [after]%
  Added: [N] tests for [M] functions
  New tests: [TestName1], [TestName2], ...
Overall: [before]% -> [after]%
```

## Tiebreaker
Test the highest cyclomatic complexity function first. If tied, longest function.
