# TASK: Identify and refactor the top 5–10 most complex **test** functions below test-appropriate complexity thresholds.

## Execution Mode
**Autonomous action** — refactor test functions, validate with tests and diff.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Test Strategy
1. Read the project README and discover the testing philosophy: unit-focused, integration-heavy, or BDD?
2. Identify the test framework in use (`testing` only, `testify`, `gomock`, etc.).
3. Discover existing test conventions: how are helpers structured, do they use `t.Helper()`, table-driven tests, `t.Parallel()`?
4. Note the assertion style — refactored tests must match existing patterns.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --only-tests --format json --output baseline.json --sections functions --max-complexity 14 --max-function-length 60
go-stats-generator analyze . --only-tests --max-complexity 14 --max-function-length 60
```

### Phase 2: Refactor
1. Select the top 5–10 test functions exceeding thresholds (sorted by overall complexity descending).
2. For each target test function, apply test-appropriate refactoring matching the project's conventions:
   - Extract shared setup into test helpers using `t.Helper()`.
   - Convert repetitive assertions into table-driven subtests with `t.Run`.
   - Extract complex assertion logic into named helper functions.
   - Each extracted helper: <30 lines, cyclomatic <12 (tunable defaults).
   - Preserve all existing test coverage and pass/fail behavior.
3. Run `go test -race ./...` after each refactoring.

### Phase 3: Validate
```bash
go-stats-generator analyze . --only-tests --format json --output post.json --sections functions --max-complexity 14 --max-function-length 60
go-stats-generator diff baseline.json post.json
```
Confirm: zero regressions, all target test functions now below thresholds.

## Default Thresholds (test-appropriate — ~50% relaxed vs. production)
| Metric | Warning | Critical |
|--------|---------|----------|
| Overall complexity | >15.0 | >22.0 |
| Cyclomatic complexity | >15 | >22 |
| Function length | >45 | >90 |
| Nesting depth | >5 | >7 |
| Extracted helper length | — | >30 |
| Extracted helper cyclomatic | — | >12 |

## Refactoring Rules
- **Table-driven tests**: preferred strategy — consolidate repetitive cases into `[]struct{...}` with `t.Run`.
- **Test helpers**: extract shared setup/teardown into functions marked with `t.Helper()`.
- **Assertion helpers**: extract complex assertion sequences into named helpers.
- Match the project's existing test naming and assertion patterns.
- Never change test coverage or pass/fail behavior.

## Output Format
```
[function] [file]: [old_complexity] -> [new_complexity] ([reduction_%])
  Extracted: [helper_1], [helper_2], ...
  Tests: PASS
```

## Tiebreaker
When complexity scores are tied, refactor the longest test function first.
