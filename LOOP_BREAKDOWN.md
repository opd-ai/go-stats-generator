# TASK: Execute iterative breakdown cycles using BREAKDOWN.md logic until no functions exceed complexity thresholds, then halt.

## Execution Mode
**Autonomous iterative loop** — wraps BREAKDOWN.md for repeated execution.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Initialization
```bash
go-stats-generator analyze . --skip-tests --format json --output iteration-0.json --sections functions --max-complexity 10 --max-function-length 30
go-stats-generator analyze . --skip-tests --max-complexity 10 --max-function-length 30
```

### Per-Iteration Cycle (max 10 iterations)
For each iteration N:

1. **Analyze**: Identify ALL functions exceeding thresholds from the current analysis.
2. **Select target**: Choose the single worst-offending function (longest first if tied).
3. **Refactor**: Apply the BREAKDOWN.md refactoring rules:
   - Extract cohesive blocks into named helpers (<20 lines, cyclomatic <8).
   - Preserve all public API signatures.
   - Each extracted function gets a GoDoc comment.
4. **Test**: Run `go test -race ./...`
   - If tests FAIL → halt immediately, report failure.
5. **Measure**:
   ```bash
   go-stats-generator analyze . --skip-tests --format json --output iteration-N.json --sections functions --max-complexity 10 --max-function-length 30
   go-stats-generator diff iteration-$((N-1)).json iteration-N.json
   ```
6. **Validate**: Target function must show >=50% complexity reduction.
   - If <50% improvement → halt (diminishing returns).
7. **Check continuation**:
   - Any function still exceeds thresholds AND iteration < 10 → continue.
   - Otherwise → halt.

## Thresholds
| Metric | Maximum |
|--------|---------|
| Overall complexity | 10.0 |
| Cyclomatic complexity | 10 |
| Function length | 30 lines |

## Continuation Criteria (ALL must be true)
- At least one function exceeds thresholds
- Maximum iterations (10) not reached
- Previous iteration achieved >=50% complexity reduction on target
- All tests still pass

## Output Format (per iteration)
```
ITERATION [N]:
  Function: [name] in [file]
  Complexity: [old] -> [new] ([reduction]%)
  Extracted: [count] helpers
  Tests: PASS | FAIL
  CONTINUE? [YES | NO] — [reason]
```

## Final Summary
```
REFACTORING COMPLETE
Total iterations: [N]
Functions refactored: [count]
Remaining violations: [count] (or "none")
```

## Tiebreaker
Choose the longest function first when complexity scores are tied.
## Halt Conditions (any triggers immediate stop)
- Tests fail after refactoring
- Target function shows <50% complexity reduction
- Max iterations (10) reached
- No remaining functions exceed thresholds (success)

## Validation Checklist
- [ ] Every iteration achieved >=50% reduction on its target
- [ ] All tests pass after every iteration
- [ ] No exported API signatures changed
