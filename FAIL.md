# TASK: Classify and resolve Go test failures using complexity metrics for root cause correlation.

## Execution Mode
**Autonomous action** — analyze failures, fix root causes, validate with tests.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 1: Identify Failures
```bash
go test -race -count=1 ./... 2>&1 | tee test-output.txt
go-stats-generator analyze . --skip-tests --format json --output baseline.json --sections functions,patterns
```

### Phase 2: Classify and Fix
1. Parse test output to extract every failing test: name, package, error message, file:line.
2. For each failing test, look up the function-under-test in the baseline JSON:
   - Get cyclomatic complexity, line count, nesting depth.
   - Check `.patterns.concurrency_patterns` for related concurrency issues.
3. Classify each failure into one of three categories:

| Category | Description | Fix Strategy |
|----------|-------------|-------------|
| Cat 1: Implementation Bug | Test is correct, code is wrong | Fix the production code |
| Cat 2: Test Spec Error | Code is correct, test expectation is wrong | Fix the test |
| Cat 3: Negative Test Gap | Test expects success but should test error path | Convert to proper error test |

4. For each failure (in order of function complexity, highest first):
   - Read the failing test and the function under test.
   - Determine the root cause and category.
   - Apply the minimal fix according to the category.
   - Run `go test -race -run TestName ./package` to confirm the specific fix.
5. After all individual fixes, run the full suite: `go test -race ./...`

### Phase 3: Validate
```bash
go-stats-generator analyze . --skip-tests --format json --output post.json --sections functions,patterns
go-stats-generator diff baseline.json post.json
```
Confirm: all tests pass, zero complexity regressions.

## Risk Indicators
- Cyclomatic complexity >12: high-risk for implementation bugs
- Nesting depth >3: high-risk for logic errors
- Function length >30: high-risk for untested code paths
- Concurrency primitives present: check for race conditions

## Fix Rules
- Cat 1 fixes must not change the public API.
- Cat 2 fixes must update test expectations to match documented behavior.
- Cat 3 conversions must use `t.Errorf` or `assert.Error` patterns.
- Never delete a failing test — fix it or convert it.
- Each fix must be minimal and targeted to the specific failure.

## Output Format
```
[Cat N] [TestName] [package] — [root cause]
  Function: [name] (complexity: [N], lines: [N])
  Fix: [description of change]
  Status: PASS
```

## Tiebreaker
Fix the failure in the highest-complexity function first — complex code is most likely to harbor defects.
## Concurrency Failure Patterns
- Race condition: test passes alone but fails with `-race` → add proper synchronization.
- Goroutine leak: test hangs or times out → check channel/context lifecycle.
- Flaky test: test passes intermittently → investigate shared state or timing dependencies.

## Resolution Order
1. Fix all Cat 1 (implementation bugs) first — they affect production code.
2. Fix Cat 2 (test spec errors) second — they mask real issues.
3. Convert Cat 3 (negative test gaps) last — they improve coverage.

## Validation Checklist
- [ ] All previously failing tests now pass
- [ ] No new test failures introduced
- [ ] `go test -race ./...` passes (full suite)
- [ ] Diff report shows zero complexity regressions
