# TASK: Perform a focused test quality and coverage audit of a Go project, evaluating test completeness, test design quality, coverage gaps on critical paths, flaky test risk, and test infrastructure health while rigorously preventing false positives.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the test quality audit report
2. **`GAPS.md`** — gaps in test coverage and test infrastructure

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project's Testing Strategy
1. Read the project README for any claims about test coverage, testing approach, or quality standards.
2. Examine the project for testing infrastructure:
   - Test files (`*_test.go`) — which packages have them, which do not?
   - Test helpers, fixtures, and `testdata/` directories
   - Testing libraries in `go.mod` (testify, gomock, go-cmp, httptest)
   - CI configuration for test execution (GitHub Actions, Makefile targets)
   - Coverage reporting tools or thresholds
3. List packages (`go list ./...`) and identify which packages have test files.
4. Understand the project's testing conventions:
   - Table-driven tests vs individual test functions
   - White-box (`package foo`) vs black-box (`package foo_test`) testing
   - Mock/stub patterns (interfaces, generated mocks, test doubles)
   - Integration test separation (build tags, `_integration_test.go`)
5. Identify critical paths that MUST be tested: core algorithms, data transformations, error handling, user-facing commands.

### Phase 1: Online Research
Use web search to build context:
1. Search for the project on GitHub — read issues about test failures, flaky tests, or requests for better coverage.
2. Check CI history for patterns of test instability.
3. Review Go testing best practices relevant to the project's domain.

Keep research brief (≤10 minutes). Record only findings relevant to test quality assessment.

### Phase 2: Baseline
```bash
set -o pipefail
go-stats-generator analyze . --skip-tests --format json --sections functions,documentation,packages,patterns > /tmp/test-audit-metrics.json
go-stats-generator analyze . --skip-tests
go test -race -count=1 -coverprofile=/tmp/test-coverage.out ./... 2>&1 | tee /tmp/test-results.txt
go tool cover -func=/tmp/test-coverage.out | tee /tmp/test-coverage-func.txt
```
Delete temporary files when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Test Quality Audit

#### 3a. Coverage Analysis
- [ ] Overall test coverage is assessed — note the percentage and whether it meets any stated targets.
- [ ] Critical path functions (core algorithms, data processing, error handling) have test coverage.
- [ ] Exported API functions are tested — untested exported functions are implicit promises without verification.
- [ ] Error paths are tested — not just happy paths. Functions returning errors should have tests that trigger error returns.
- [ ] Edge cases are covered: empty input, nil pointers, boundary values, maximum sizes, unicode/special characters.
- [ ] Packages with 0% coverage are identified and assessed — is the absence of tests justified or an oversight?

#### 3b. Test Design Quality
For each test file, evaluate the quality of the tests themselves:

- [ ] Tests are independent — no test depends on the execution order or side effects of another test.
- [ ] Table-driven tests are used for functions with multiple input/output cases (instead of copy-pasted test bodies).
- [ ] `t.Run` subtests provide clear names for each test case — failure messages identify which case failed.
- [ ] Test helpers call `t.Helper()` so failure locations point to the test case, not the helper.
- [ ] Assertions are specific — `assert.Equal(expected, actual)` not just `assert.NotNil(result)`.
- [ ] Test names describe behavior: `TestParseConfig_InvalidJSON_ReturnsError` not `TestParseConfig3`.
- [ ] Tests verify behavior, not implementation — they test what a function does, not how it does it (no testing of unexported internals unless justified).

#### 3c. Flaky Test Risk
Identify tests that are likely to fail intermittently:

- [ ] No tests use `time.Sleep` for synchronization — use channels, `sync.WaitGroup`, or test helpers instead.
- [ ] No tests depend on wall-clock time (`time.Now()`) without using a mockable clock interface or `time.After` with generous margins.
- [ ] No tests depend on specific port availability — use `net.Listen(":0")` for random port assignment.
- [ ] No tests depend on file system ordering — directory listings are sorted before comparison.
- [ ] No tests depend on map iteration order — Go maps iterate in random order by design.
- [ ] No tests depend on goroutine scheduling order without explicit synchronization.
- [ ] No tests read environment variables set by other tests without cleanup (`t.Setenv` or `defer os.Unsetenv`).
- [ ] Parallel tests (`t.Parallel()`) do not share mutable state without synchronization.

#### 3d. Test Infrastructure
- [ ] `go test ./...` passes consistently with no manual setup required.
- [ ] `go test -race ./...` passes — no race conditions in tests or code under test.
- [ ] Test fixtures in `testdata/` are committed and not generated at test time from external sources.
- [ ] CI runs tests on every PR and blocks merge on failure.
- [ ] Test timeouts are set (via `-timeout` flag or `context.WithTimeout` in tests) to prevent hanging tests from blocking CI.
- [ ] Integration tests are separated from unit tests (build tags or naming convention) so unit tests can run fast.
- [ ] Mock/stub generation is reproducible and committed (not generated on-the-fly during tests).

#### 3e. Test Completeness for Stated Goals
For each project goal from the README:

- [ ] There exists at least one test that verifies the goal is achievable end-to-end.
- [ ] Error scenarios for goal-critical paths are tested (what happens when the goal cannot be achieved?).
- [ ] Performance-critical goals have benchmarks (`func Benchmark*`).
- [ ] Concurrency-related goals are tested with `-race` and concurrent test scenarios.

#### 3f. Missing Test Categories
Identify categories of tests that are absent:

- [ ] **Benchmark tests**: Do `Benchmark*` functions exist for performance-critical code?
- [ ] **Example tests**: Do `Example*` functions exist for documenting API usage?
- [ ] **Fuzz tests**: Do `Fuzz*` functions exist for input parsing and validation code (Go 1.18+)?
- [ ] **Regression tests**: Are previously reported bugs covered by tests to prevent recurrence?
- [ ] **Negative tests**: Are invalid inputs, error conditions, and boundary violations tested?
- [ ] **Concurrent tests**: Are types that claim thread-safety tested under concurrent access?

#### 3g. False-Positive Prevention (MANDATORY)
Before recording ANY finding, apply these checks:

1. **Assess the project's testing strategy**: A project that uses integration tests heavily may have low unit test coverage by design. Evaluate the overall test strategy, not just unit test metrics.
2. **Check for alternative verification**: Code may be verified through `go vet`, static analysis, or type system guarantees instead of runtime tests. These are valid quality assurance approaches.
3. **Verify the function needs testing**: A one-line function that delegates to a well-tested standard library function may not need its own test. Focus on logic, not delegation.
4. **Respect project maturity**: A v0 project may intentionally prioritize implementation speed over test coverage. Note the gap but classify appropriately.
5. **Check for generated code**: Generated code (protobuf, mock implementations) typically does not need tests — the generator is tested upstream.

**Rule**: Coverage numbers are a guide, not a verdict. A project with 60% coverage on all critical paths may be better tested than one with 90% coverage that skips error handling. Evaluate test quality, not just quantity.

### Phase 4: Report
Generate **`AUDIT.md`**:
```markdown
# TEST QUALITY AUDIT — [date]

## Testing Strategy Summary
[Testing approach, libraries used, CI configuration, coverage targets]

## Coverage Summary
| Package | Coverage | Critical Path | Test Files | Benchmark | Fuzz |
|---------|----------|---------------|-----------|-----------|------|
| [pkg] | N% | ✅/❌ | N | ✅/❌ | ✅/❌ |

## Test Results
[Summary of `go test -race ./...` results — any failures are CRITICAL]

## Findings
### CRITICAL
- [ ] [Finding] — [file:line or package] — [test quality issue] — [risk: what could break undetected] — **Remediation:** [specific test to add or fix]
### HIGH / MEDIUM / LOW
- [ ] ...

## False Positives Considered and Rejected
| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| [description] | [why it is not a real test quality issue] |
```

Generate **`GAPS.md`**:
```markdown
# Test Coverage Gaps — [date]

## [Gap Title]
- **Untested Code**: [function/package/path that lacks tests]
- **Risk**: [what could go wrong without test coverage]
- **Recommended Tests**: [specific test cases to add]
- **Priority**: [based on criticality of the untested code]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | Tests actively fail or have race conditions, critical path (data processing, user-facing output) has 0% coverage, or test infrastructure is broken |
| HIGH | Core functionality lacks error path testing, flaky tests in CI, or no tests for functions with cyclomatic complexity >10 |
| MEDIUM | Below-target coverage on non-critical packages, missing benchmarks for performance-claimed features, or test design issues (no subtests, weak assertions) |
| LOW | Missing example tests, cosmetic test naming issues, or coverage gaps in utility/helper code |

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Specific test**: Describe the test function to add — what it tests, what inputs it uses, what it asserts. Do not recommend "add more tests."
2. **Respect project patterns**: Use the project's existing testing conventions (testify vs stdlib, table-driven vs individual).
3. **Verifiable**: Include the command to run the new test and verify coverage improvement.
4. **Focused**: Each recommended test should cover a specific risk, not be a generic "test everything" recommendation.

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Use `go-stats-generator` function metrics to identify high-risk untested functions (high complexity, many parameters).
- Every finding must reference a specific package or function.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Evaluate tests against the project's **own stated quality goals and testing conventions**.
- Apply the Phase 3g false-positive prevention checks to every candidate finding before including it.

## Tiebreaker
Prioritize: broken tests → untested critical paths → flaky test risk → missing error path tests → coverage gaps → missing benchmarks → missing examples. Within a level, prioritize by the criticality of the untested code to the project's stated goals.
