# TASK: Execute iterative breakdown cycles on **test functions** using TEST_BREAKDOWN.md logic until no test functions exceed complexity thresholds, then halt.

## Execution Mode
**Autonomous iterative loop** — wraps TEST_BREAKDOWN.md for repeated execution.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Initialization
```bash
go-stats-generator analyze . --only-tests --format json --output iteration-0.json --sections functions --max-complexity 15 --max-function-length 45
go-stats-generator analyze . --only-tests --max-complexity 15 --max-function-length 45
```

### Per-Iteration Cycle (max 10 iterations)
For each iteration N:

1. **Analyze**: Identify ALL test functions exceeding thresholds from the current analysis.
2. **Select target**: Choose the single worst-offending test function (longest first if tied).
3. **Refactor**: Apply the TEST_BREAKDOWN.md refactoring rules:
   - Convert to table-driven tests where applicable.
   - Extract setup/assertion helpers with `t.Helper()` (<30 lines, cyclomatic <12).
   - Preserve all test coverage and pass/fail behavior.
4. **Test**: Run `go test -race ./...`
   - If tests FAIL → halt immediately, report failure.
5. **Measure**:
   ```bash
   go-stats-generator analyze . --only-tests --format json --output iteration-N.json --sections functions --max-complexity 15 --max-function-length 45
   go-stats-generator diff iteration-$((N-1)).json iteration-N.json
   ```
6. **Validate**: Target test function must show >=50% complexity reduction.
   - If <50% improvement → halt (diminishing returns).
7. **Check continuation**:
   - Any test function still exceeds thresholds AND iteration < 10 → continue.
   - Otherwise → halt.

## Thresholds (Test-Appropriate)
| Metric | Maximum |
|--------|---------|
| Overall complexity | 15.0 |
| Cyclomatic complexity | 15 |
| Function length | 45 lines |

> **Note**: Thresholds are relaxed by ~50% for test code.

## Continuation Criteria (ALL must be true)
- At least one test function exceeds thresholds
- Maximum iterations (10) not reached
- Previous iteration achieved >=50% complexity reduction on target
- All tests still pass

## Output Format (per iteration)
```
ITERATION [N]:
  Function: [name] in [file]
  Complexity: [old] -> [new] ([reduction]%)
  Strategy: [table-driven | extract helper | decompose]
  Extracted: [count] helpers
  Tests: PASS | FAIL
  CONTINUE? [YES | NO] — [reason]
```

## Final Summary
```
TEST REFACTORING COMPLETE
Total iterations: [N]
Test functions refactored: [count]
Remaining violations: [count] (or "none")
```

## Tiebreaker
Choose the longest test function first when complexity scores are tied.
## Halt Conditions (any triggers immediate stop)
- Tests fail after refactoring
- Target test function shows <50% complexity reduction
- Max iterations (10) reached
- No remaining test functions exceed thresholds (success)

## Validation Checklist
- [ ] Every iteration achieved >=50% reduction on its target
- [ ] All tests pass after every iteration
- [ ] No test coverage reduced
