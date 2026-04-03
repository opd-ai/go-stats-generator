# TASK: Execute autonomous iterative refactoring cycles until no functions exceed complexity thresholds, then halt.

## Execution Mode
**Autonomous iterative loop** — self-terminating on success, regression, or max iterations.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Phase 0: Understand the Codebase
Before the first iteration:
1. Read the project README to understand its domain and purpose.
2. Identify the project's coding patterns, naming conventions, and error handling style.
3. Note which packages are core logic vs. infrastructure — core functions may justify slightly higher thresholds.

## Initialization
```bash
go-stats-generator analyze . --skip-tests --format json --output iteration-0.json --sections functions --max-complexity 10 --max-function-length 30
```

## Per-Iteration Cycle (max 5 iterations)
For each iteration N:

1. **Identify target**: Select the single highest-complexity function exceeding thresholds.
2. **Understand its role**: Read the function and its callers. Is it a parser, handler, orchestrator? Choose a decomposition strategy that fits the function's purpose and the project's idioms.
3. **Refactor**: Apply extract-method decomposition:
   - Extract cohesive blocks into named helpers (<20 lines, cyclomatic <8).
   - Preserve all public API signatures.
   - Match the project's naming and error handling conventions.
4. **Test**: Run `go test -race ./...` — halt loop if tests fail.
5. **Measure**:
   ```bash
   go-stats-generator analyze . --skip-tests --format json --output iteration-N.json --sections functions --max-complexity 10 --max-function-length 30
   go-stats-generator diff iteration-$((N-1)).json iteration-N.json
   ```
6. **Check termination conditions**:
   - **Success**: No remaining functions exceed thresholds → halt.
   - **Regression**: Diff shows any metric worsening → halt with rollback warning.
   - **Max iterations**: N >= 5 → halt with remaining violations count.

## Default Thresholds (calibrate to project)
| Metric | Maximum |
|--------|---------|
| Overall complexity | 10.0 |
| Cyclomatic complexity | 10 |
| Function length | 30 lines |
| Nesting depth | 4 |
| Extracted function length | 20 lines |
| Extracted function cyclomatic | 8 |

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
  Refactored: [old] -> [new] ([reduction]%)
  Extracted: [helper_1], [helper_2]
  Tests: PASS
  Remaining violations: [count]
```

## Final Summary
```
LOOP COMPLETE: [iterations] iterations, [functions_fixed] functions refactored
Remaining violations: [count] (or "none")
```

## Recovery
If tests fail: revert (`git checkout -- <modified files>`), log the failure, halt immediately.

## Tiebreaker
Refactor the highest-complexity function. If tied, choose the longest function.
