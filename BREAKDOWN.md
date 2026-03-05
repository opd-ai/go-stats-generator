# TASK: Identify and refactor the top 5–10 most complex functions below professional complexity thresholds.

## Execution Mode
**Autonomous action** — refactor functions, validate with tests and diff.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Codebase
Before refactoring, understand what you're changing:
1. Read the project README to learn the project's domain and purpose.
2. Examine `go.mod` for dependencies and Go version.
3. Identify the project's coding patterns: do they use builders, functional options, table-driven dispatch, or other idioms? Refactored code must match.
4. Note the project's error handling style and test strategy — extracted helpers must be consistent.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output baseline.json --sections functions --max-complexity 9 --max-function-length 40
go-stats-generator analyze . --skip-tests --max-complexity 9 --max-function-length 40
```

### Phase 2: Refactor
1. Select the top 5–10 functions exceeding thresholds (sorted by overall complexity descending).
2. For each target function, **understand its role** in the project before refactoring:
   - What does this function do? Is it a parser, handler, orchestrator, or algorithm?
   - A complex parser or state machine may warrant higher thresholds than a simple handler.
3. Apply extract-method refactoring idiomatically:
   - Identify cohesive blocks (loop bodies, conditional branches, setup/teardown, error paths).
   - Extract into named helpers matching the project's naming conventions (default: verb-first).
   - Each extracted function: <20 lines, cyclomatic <8 (tunable defaults).
   - Preserve all existing public API signatures.
4. Run `go test -race ./...` after each refactoring.

### Phase 3: Validate
```bash
go-stats-generator analyze . --skip-tests --format json --output post.json --sections functions --max-complexity 9 --max-function-length 40
go-stats-generator diff baseline.json post.json
```
Confirm: zero regressions, all target functions now below thresholds.

## Complexity Formula
```
Overall = (Cyclomatic * 0.3) + (Lines * 0.2) + (Nesting * 0.2) + (Cognitive * 0.15) + (Signature * 0.15)
```

## Default Thresholds (calibrate to project baseline)
| Metric | Warning | Critical |
|--------|---------|----------|
| Overall complexity | >9.0 | >15.0 |
| Cyclomatic complexity | >9 | >15 |
| Function length (code lines) | >40 | >80 |
| Nesting depth | >3 | >5 |
| Extracted function length | — | >20 |
| Extracted function cyclomatic | — | >8 |

## Refactoring Rules
- **Extract method**: move cohesive blocks into named helpers.
- **Decompose conditional**: replace complex boolean chains with predicate functions.
- **Replace loop body**: extract inner loop logic into a function.
- **Consolidate error handling**: merge repeated error patterns into a shared helper.
- Match the project's naming conventions (default: verb-first, e.g., `buildDependencyMap`).
- Never change exported function signatures.
- Add GoDoc to extracted functions with >3 lines of logic.

## Output Format
```
[function] [file]: [old_complexity] -> [new_complexity] ([reduction_%])
  Extracted: [helper_1], [helper_2], ...
  Tests: PASS
```

## Tiebreaker
When complexity scores are tied, refactor the longest function first.
