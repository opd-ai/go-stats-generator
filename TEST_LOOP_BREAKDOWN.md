# TASK: Execute iterative breakdown cycles on **test functions** using TEST_BREAKDOWN.md logic until no test functions exceed complexity thresholds, then halt.

## Execution Mode
**Autonomous iterative loop** — wraps TEST_BREAKDOWN.md for repeated execution.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Phase 0: Understand the Test Strategy
Before the first iteration:
1. Read the project README and discover the testing philosophy.
2. Identify the test framework, assertion style, and existing helper patterns.
3. Note whether tests use `t.Parallel()`, `t.Cleanup()`, and table-driven patterns — refactoring must match.

## Initialization
```bash
go-stats-generator analyze . --only-tests --format json --output iteration-0.json --sections functions --max-complexity 15 --max-function-length 45
go-stats-generator analyze . --only-tests --max-complexity 15 --max-function-length 45
```

## Per-Iteration Cycle (max 10 iterations)
For each iteration N:

1. **Analyze**: Identify ALL test functions exceeding thresholds.
2. **Select target**: Choose the single worst-offending test function (longest first if tied).
3. **Refactor**: Apply TEST_BREAKDOWN.md rules matching project conventions:
   - Convert to table-driven tests where applicable.
   - Extract setup/assertion helpers with `t.Helper()` (<30 lines, cyclomatic <12).
   - Preserve all test coverage and pass/fail behavior.
4. **Test**: Run `go test -race ./...`
   - If tests FAIL → halt immediately.
5. **Measure**:
   ```bash
   go-stats-generator analyze . --only-tests --format json --output iteration-N.json --sections functions --max-complexity 15 --max-function-length 45
   go-stats-generator diff iteration-$((N-1)).json iteration-N.json
   ```
6. **Validate**: Target must show >=50% complexity reduction. If <50% → halt.

## Default Thresholds (test-appropriate)
| Metric | Maximum |
|--------|---------|
| Overall complexity | 15.0 |
| Cyclomatic complexity | 15 |
| Function length | 45 lines |

## Halt Conditions (any triggers immediate stop)
- Tests fail after refactoring
- Target test function shows <50% complexity reduction
- Max iterations (10) reached
- No remaining test functions exceed thresholds (success)

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
