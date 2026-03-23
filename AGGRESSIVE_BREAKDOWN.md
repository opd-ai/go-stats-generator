# TASK: Identify and refactor ALL functions exceeding professional complexity thresholds — no cap per session.

## Execution Mode
**Autonomous action** — refactor every function above threshold, validate with tests and diff. Do not stop at a fixed count; continue until every flagged function is resolved or the session's context is exhausted.

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

### Phase 2: Build the Full Worklist
1. Collect **every** function that exceeds any threshold (sorted by overall complexity descending).
2. Do not cap the list — include all violators.
3. Group functions by package/file for efficient batch refactoring (touching nearby code in one pass avoids redundant test cycles).

### Phase 3: Refactor (Iterate Until Exhausted)
Loop over the worklist. For each target function:
1. **Understand its role** in the project before refactoring:
   - What does this function do? Is it a parser, handler, orchestrator, or algorithm?
   - A complex parser or state machine may warrant higher thresholds than a simple handler.
2. Apply extract-method refactoring idiomatically:
   - Identify cohesive blocks (loop bodies, conditional branches, setup/teardown, error paths).
   - Extract into named helpers matching the project's naming conventions (default: verb-first).
   - Each extracted function: <20 lines, cyclomatic <8 (tunable defaults).
   - Preserve all existing public API signatures.
3. Run `go vet ./...` after each refactoring to catch mistakes early.
4. Run `go test -race ./...` after every batch of refactorings within the same package (or after each individual refactoring if the function is high-risk).
5. If tests fail, fix immediately before moving on — never accumulate broken state.

**Stopping conditions** (in priority order):
- ✅ Worklist exhausted — every flagged function is now below thresholds.
- ⚠️ Context/session boundary reached — commit progress and document remaining items.
- 🛑 Unrecoverable test regression — stop, revert the last refactoring, and report.

### Phase 4: Validate
```bash
go-stats-generator analyze . --skip-tests --format json --output post.json --sections functions --max-complexity 9 --max-function-length 40
go-stats-generator diff baseline.json post.json
```
Confirm: zero regressions, all target functions now below thresholds (or document justified exceptions).

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

After all refactorings, print a summary table:
```
Total functions refactored: N
Total extracted helpers:     M
Remaining above threshold:   R (with justifications if R > 0)
```

## Tiebreaker
When complexity scores are tied, refactor the longest function first.

## Session Strategy
- Prioritize breadth: get every function below threshold rather than perfecting a few.
- Batch refactorings within the same file/package to minimize test cycle overhead.
- Commit working progress frequently so partial sessions still deliver value.
- If the session is running low on context, commit and document the remaining worklist for the next session.
