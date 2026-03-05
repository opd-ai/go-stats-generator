# TASK: Execute autonomous iterative refactoring cycles on **test functions** until no test functions exceed complexity thresholds, then halt.

## Execution Mode
**Autonomous iterative loop** — self-terminating on success, regression, or max iterations.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Initialization
```bash
go-stats-generator analyze . --only-tests --format json --output iteration-0.json --sections functions --max-complexity 15 --max-function-length 45
```

### Per-Iteration Cycle (max 5 iterations)
For each iteration N:

1. **Identify target**: Select the single highest-complexity test function exceeding thresholds.
2. **Refactor**: Apply test-appropriate decomposition:
   - Convert to table-driven subtests where applicable (preferred strategy).
   - Extract setup/assertion helpers with `t.Helper()` (<30 lines, cyclomatic <12).
   - Preserve all test coverage and pass/fail behavior.
   - Use descriptive subtest names in `t.Run`.
3. **Test**: Run `go test -race ./...` — halt loop if tests fail.
4. **Measure**:
   ```bash
   go-stats-generator analyze . --only-tests --format json --output iteration-N.json --sections functions --max-complexity 15 --max-function-length 45
   go-stats-generator diff iteration-$((N-1)).json iteration-N.json
   ```
5. **Check termination conditions**:
   - **Success**: No remaining test functions exceed thresholds → halt with success.
   - **Regression**: Diff shows any metric worsening → halt with rollback warning.
   - **Max iterations**: N >= 5 → halt with remaining violations count.

## Complexity Formula
```
Overall = (Cyclomatic * 0.3) + (Lines * 0.2) + (Nesting * 0.2) + (Cognitive * 0.15) + (Signature * 0.15)
```

## Thresholds (Test-Appropriate)
| Metric | Maximum |
|--------|---------|
| Overall complexity | 15.0 |
| Cyclomatic complexity | 15 |
| Function length | 45 lines |
| Nesting depth | 5 |
| Extracted helper length | 30 lines |
| Extracted helper cyclomatic | 12 |

> **Note**: Thresholds are relaxed by ~50% for test code. Table-driven tests and `t.Helper()` extraction are preferred strategies.

## Termination Conditions
| Condition | Action |
|-----------|--------|
| No violations remain | Halt — SUCCESS |
| Tests fail after refactoring | Halt — revert last change |
| Diff shows regression | Halt — revert last change |
| 5 iterations reached | Halt — report remaining violations |
| No improvement in iteration | Halt — report remaining violations |

## Output Format (per iteration)
```
ITERATION [N]:
  Target: [function] in [file] (complexity: [score])
  Strategy: [table-driven | extract helper | decompose]
  Refactored: [old] -> [new] ([reduction]%)
  Extracted: [helper_1], [helper_2]
  Tests: PASS
  Remaining violations: [count]
```

## Final Summary
```
LOOP COMPLETE: [iterations] iterations, [test_functions_fixed] test functions refactored
Remaining violations: [count] (or "none")
```

## Tiebreaker
Refactor the highest-complexity test function. If tied, choose the longest function.
## Recovery from Failed Iteration
If tests fail after a refactoring:
1. Revert the last change (`git checkout -- <modified files>`).
2. Log the failure and the test function that caused it.
3. Halt the loop — do not attempt further iterations.

## Validation Checklist
- [ ] Each iteration reduced at least one violation
- [ ] All tests pass after every iteration
- [ ] No test coverage reduced
