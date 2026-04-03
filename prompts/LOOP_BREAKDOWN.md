# TASK: Execute iterative breakdown cycles using BREAKDOWN.md logic until no functions exceed complexity thresholds, then halt.

## Execution Mode
**Autonomous iterative loop** — wraps BREAKDOWN.md for repeated execution.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Phase 0: Understand the Codebase
Before the first iteration:
1. Read the project README to understand its domain, purpose, and architecture.
2. Identify the project's coding idioms: naming conventions, error handling patterns, and any established decomposition patterns (builders, functional options, etc.).
3. Note which functions serve critical roles — their refactoring deserves extra care.

## Initialization
```bash
go-stats-generator analyze . --skip-tests --format json --output iteration-0.json --sections functions --max-complexity 10 --max-function-length 30
go-stats-generator analyze . --skip-tests --max-complexity 10 --max-function-length 30
```

## Per-Iteration Cycle (max 10 iterations)
For each iteration N:

1. **Analyze**: Identify ALL functions exceeding thresholds from the current analysis.
2. **Select target**: Choose the single worst-offending function (longest first if tied).
3. **Understand its role**: Read the function and its context. Determine whether it's a parser, handler, orchestrator, etc. — this informs the decomposition strategy.
4. **Refactor**: Apply the BREAKDOWN.md refactoring rules idiomatically:
   - Extract cohesive blocks into named helpers (<20 lines, cyclomatic <8).
   - Preserve all public API signatures.
   - Match the project's naming and error handling conventions.
5. **Test**: Run `go test -race ./...`
   - If tests FAIL → halt immediately, report failure.
6. **Measure**:
   ```bash
   go-stats-generator analyze . --skip-tests --format json --output iteration-N.json --sections functions --max-complexity 10 --max-function-length 30
   go-stats-generator diff iteration-$((N-1)).json iteration-N.json
   ```
7. **Validate**: Target function must show >=50% complexity reduction.
   - If <50% improvement → halt (diminishing returns).

## Default Thresholds (calibrate to project)
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

## Halt Conditions (any triggers immediate stop)
- Tests fail after refactoring
- Target function shows <50% complexity reduction
- Max iterations (10) reached
- No remaining functions exceed thresholds (success)

## Tiebreaker
Choose the longest function first when complexity scores are tied.
