# TASK: Generate unit tests targeting the highest-complexity untested functions to achieve >=80% coverage.

## Execution Mode
**Autonomous action** — generate tests, validate coverage improvement.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output baseline.json --sections functions
go test -cover ./... 2>&1 | tee coverage-baseline.txt
```

### Phase 2: Generate Tests
1. From baseline JSON, extract all functions sorted by cyclomatic complexity descending.
2. Cross-reference with `go test -cover` output to identify untested or low-coverage packages.
3. Prioritize test generation:
   - CRITICAL: cyclomatic >15, no test coverage → test first
   - HIGH: cyclomatic 9–15, partial coverage → add edge case tests
   - MEDIUM: cyclomatic 5–8, low coverage → add basic tests
4. For each target function, generate tests following these patterns:
   - **Table-driven tests** for functions with multiple input/output combinations.
   - **Error path tests** for functions that return errors.
   - **Boundary tests** for functions with numeric parameters.
   - **Nil input tests** for functions accepting pointers/slices/maps.
5. Test conventions:
   - Use `testing` package and `testify` (where already present in the project).
   - Name tests `Test[FunctionName]_[Scenario]`.
   - Use `t.Helper()` in test helper functions.
   - Use `t.Parallel()` where safe.
   - Use `t.Run` for subtests in table-driven tests.
6. Run `go test -race ./...` after each test file to confirm tests pass.

### Phase 3: Validate
```bash
go test -cover ./... 2>&1 | tee coverage-post.txt
go-stats-generator analyze . --skip-tests --format json --output post.json --sections functions
```
Confirm: coverage increased, all tests pass, no flaky tests.

## Coverage Targets
| Package Type | Target |
|--------------|--------|
| Core analyzer | >=80% |
| CLI commands | >=70% |
| Utility packages | >=80% |
| Reporter/output | >=60% |

## Test Generation Rules
- Only use `testing` package and `testify` (if already a dependency).
- Each test must be deterministic (no random data, no time-dependent assertions).
- Each test must clean up after itself (temp files, goroutines).
- Test the public API surface first, then critical unexported functions.
- Do not test auto-generated code or trivial getters/setters.

## Test Quality Standards
- Each test has a clear, descriptive name indicating what it validates.
- Error path tests verify both the error value and that no partial state leaked.
- Concurrent tests use `t.Parallel()` and avoid shared mutable state.
- Table-driven tests have at least 3 cases: happy path, edge case, error case.

## Output Format
```
[package]: coverage [before]% -> [after]%
  Added: [N] tests for [M] functions
  New tests: [TestName1], [TestName2], ...
Overall: [before]% -> [after]%
```

## Tiebreaker
Test the highest cyclomatic complexity function first. If tied, longest function. If still tied, deepest nesting.
## Test Anti-Patterns to Avoid
- Testing implementation details instead of behavior.
- Hardcoding file paths or environment-specific values.
- Using `time.Sleep` for synchronization (use channels or sync primitives).
- Asserting on exact error messages (assert on error type or `errors.Is`).

## Validation Checklist
- [ ] Coverage improved toward 80% target
- [ ] All new tests pass deterministically (run 3x to confirm)
- [ ] No flaky tests introduced
- [ ] All tests clean up after themselves (temp files, goroutines)
