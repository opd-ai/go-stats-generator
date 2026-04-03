# TASK: Use `go-stats-generator` to analyze and refactor the target project's **test code** (dogfooding) — reduce complexity of test functions below test-appropriate thresholds.

## Execution Mode
**Autonomous action** — refactor test code and fix discovered test bugs, validate with tests and diff.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Test Strategy
1. Read the project README to understand its domain and build process.
2. Discover the test framework, assertion style, and existing helper patterns.
3. Identify whether the project is a tool/binary (requires rebuild verification) or a library.
4. Note conventions: `t.Helper()`, `t.Parallel()`, table-driven tests, cleanup patterns.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --only-tests --format json --output baseline.json --sections functions --max-complexity 15 --max-function-length 45
go-stats-generator analyze . --only-tests --max-complexity 15 --max-function-length 45
```

### Phase 2: Refactor Tests
1. Identify test functions exceeding thresholds from the baseline.
2. For each violating test function (sorted by overall complexity descending):
   - Apply test-appropriate refactoring matching the project's conventions:
     - Convert to table-driven tests (preferred).
     - Extract setup/assertion helpers with `t.Helper()` (<30 lines, cyclomatic <12).
     - Preserve all test coverage and pass/fail behavior.
   - If a test bug is discovered, fix it as part of this pass.
3. Run `go test -race ./...` after each refactoring.
4. If the project is a buildable tool, rebuild and verify:
   ```bash
   go build ./... && echo "BUILD PASS"
   ```

### Phase 3: Validate
```bash
go-stats-generator analyze . --only-tests --format json --output post.json --sections functions --max-complexity 15 --max-function-length 45
go-stats-generator diff baseline.json post.json
```
Confirm: all target test functions below thresholds, zero regressions.

## Default Thresholds (test-appropriate — ~50% relaxed)
| Metric | Maximum |
|--------|---------|
| Overall complexity | 15.0 |
| Cyclomatic complexity | 15 |
| Function length | 45 lines |
| Nesting depth | 5 |
| Extracted helper length | 30 |
| Extracted helper cyclomatic | 12 |

## Dogfooding Rules
- After refactoring, rebuild the project and verify it still produces correct output.
- If test behavior changes, investigate whether a test bug was introduced during refactoring.
- May fix test bugs discovered during analysis.
- Prefer table-driven tests and `t.Helper()` extraction.

## Output Format
```
Analysis: [N] test functions above thresholds
Refactored:
  [function] [file]: [old] -> [new] ([reduction]%)
  Strategy: [table-driven | extract helper | decompose]
Test bugs fixed: [count] (or "none")
Build verification: PASS
Tests: PASS
```

## Tiebreaker
Refactor the longest test function first when complexity scores are tied.
