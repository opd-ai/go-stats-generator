# TASK: Identify and refactor the top 5–10 most complex functions below professional complexity thresholds.

## Execution Mode
**Autonomous action** — refactor functions, validate with tests and diff.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output baseline.json --sections functions --max-complexity 9 --max-function-length 40
go-stats-generator analyze . --skip-tests --max-complexity 9 --max-function-length 40
```

### Phase 2: Refactor
1. From the baseline, select the top 5–10 functions exceeding thresholds (sorted by overall complexity descending).
2. For each target function, apply extract-method refactoring:
   - Identify cohesive blocks (loop bodies, conditional branches, setup/teardown, error-handling paths).
   - Extract each block into a named helper with a clear, verb-first name.
   - Each extracted function must be <20 lines with cyclomatic complexity <8.
   - Preserve all existing public API signatures.
3. Run `go test -race ./...` after each refactoring to confirm no regressions.
4. Run `go vet ./...` to confirm no new issues.

### Phase 3: Validate
```bash
go-stats-generator analyze . --skip-tests --format json --output post.json --sections functions --max-complexity 9 --max-function-length 40
go-stats-generator diff baseline.json post.json
```
Confirm: zero regressions, all target functions now below thresholds.

## Complexity Formula
Overall complexity is a weighted composite:
```
Overall = (Cyclomatic * 0.3) + (Lines * 0.2) + (Nesting * 0.2) + (Cognitive * 0.15) + (Signature * 0.15)
```

**Signature complexity** = `(params * 2) + (returns * 1.5) + (2 if variadic) + (1.5 per interface{}/any param)`

## Thresholds
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
- Name helpers with verb-first Go conventions (e.g., `buildDependencyMap`, `validateThreshold`).
- Never change exported function signatures.
- Each extracted function gets a GoDoc comment if it has >3 lines of logic.

## Go Coding Standards
- Verb-first function names: `parseConfig`, `buildReport`, `validateInput`.
- Explicit error handling with `fmt.Errorf("context: %w", err)`.
- GoDoc comment on every exported function and any unexported function >3 lines.
- Prefer stdlib over external dependencies.

## Output Format
For each refactored function:
```
[function] [file]: [old_complexity] -> [new_complexity] ([reduction_%])
  Extracted: [helper_1], [helper_2], ...
  Tests: PASS
```

## Tiebreaker
When complexity scores are tied, refactor the longest function first.
## Validation Checklist
- [ ] All target functions now below overall complexity 9.0
- [ ] No new functions introduced above thresholds
- [ ] All existing tests pass with -race flag
- [ ] No exported API signatures changed
