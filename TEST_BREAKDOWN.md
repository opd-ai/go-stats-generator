# TASK: Identify and refactor the top 5–10 most complex **test** functions below test-appropriate complexity thresholds.

## Execution Mode
**Autonomous action** — refactor test functions, validate with tests and diff.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 1: Baseline
```bash
go-stats-generator analyze . --only-tests --format json --output baseline.json --sections functions --max-complexity 14 --max-function-length 60
go-stats-generator analyze . --only-tests --max-complexity 14 --max-function-length 60
```

### Phase 2: Refactor
1. From the baseline, select the top 5–10 test functions exceeding thresholds (sorted by overall complexity descending).
2. For each target test function, apply test-appropriate refactoring:
   - Extract shared setup into test helpers using `t.Helper()`.
   - Convert repetitive assertions into table-driven subtests with `t.Run`.
   - Extract complex assertion logic into named helper functions.
   - Each extracted helper must be <30 lines with cyclomatic complexity <12.
   - Preserve all existing test coverage and pass/fail behavior.
3. Run `go test -race ./...` after each refactoring to confirm no regressions.
4. Run `go vet ./...` to confirm no new issues.

### Phase 3: Validate
```bash
go-stats-generator analyze . --only-tests --format json --output post.json --sections functions --max-complexity 14 --max-function-length 60
go-stats-generator diff baseline.json post.json
```
Confirm: zero regressions, all target test functions now below thresholds.

## Complexity Formula
Overall complexity is a weighted composite:
```
Overall = (Cyclomatic * 0.3) + (Lines * 0.2) + (Nesting * 0.2) + (Cognitive * 0.15) + (Signature * 0.15)
```

**Signature complexity** = `(params * 2) + (returns * 1.5) + (2 if variadic) + (1.5 per interface{}/any param)`

## Thresholds (Test-Appropriate)
| Metric | Warning | Critical |
|--------|---------|----------|
| Overall complexity | >15.0 | >22.0 |
| Cyclomatic complexity | >15 | >22 |
| Function length (code lines) | >45 | >90 |
| Nesting depth | >5 | >7 |
| Extracted helper length | — | >30 |
| Extracted helper cyclomatic | — | >12 |

> **Note**: Test thresholds are relaxed by ~50% compared to production code. Test functions naturally have higher complexity due to multiple assertions, table-driven cases, and setup/teardown logic.

## Refactoring Rules
- **Table-driven tests**: consolidate repetitive test cases into `[]struct{...}` with `t.Run` subtests — this is the preferred strategy for test refactoring.
- **Test helpers**: extract shared setup/teardown into functions marked with `t.Helper()`.
- **Assertion helpers**: extract complex assertion sequences into named helpers.
- **Decompose conditional**: replace complex test branching with predicate functions.
- Name helpers with test-conventional Go names (e.g., `setupTestServer`, `assertMetricsEqual`).
- Never change test coverage or pass/fail behavior.
- Each extracted helper gets a GoDoc comment if it has >3 lines of logic.

## Go Test Coding Standards
- Use `t.Helper()` in all test helper functions.
- Use `t.Parallel()` where safe.
- Use `t.Run("subtest name", ...)` for table-driven tests.
- Explicit error messages in assertions: `t.Errorf("got %v, want %v", got, want)`.
- Prefer `testify/assert` if already present in the project.

## Output Format
For each refactored test function:
```
[function] [file]: [old_complexity] -> [new_complexity] ([reduction_%])
  Extracted: [helper_1], [helper_2], ...
  Tests: PASS
```

## Tiebreaker
When complexity scores are tied, refactor the longest test function first.
## Validation Checklist
- [ ] All target test functions now below overall complexity 15.0
- [ ] No new functions introduced above thresholds
- [ ] All existing tests pass with -race flag
- [ ] No test coverage reduced
