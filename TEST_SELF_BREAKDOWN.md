# TASK: Use `go-stats-generator` to analyze and refactor its own **test code** (dogfooding) — reduce complexity of its own test functions below test-appropriate thresholds.

## Execution Mode
**Autonomous action** — self-refactor test code and fix discovered test bugs, validate with tests and diff.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 1: Self-Analysis Baseline
```bash
go-stats-generator analyze . --only-tests --format json --output baseline.json --sections functions --max-complexity 15 --max-function-length 45
go-stats-generator analyze . --only-tests --max-complexity 15 --max-function-length 45
```

### Phase 2: Self-Refactor Tests
1. From baseline, identify test functions in the `go-stats-generator` codebase exceeding test-appropriate thresholds.
2. For each violating test function (sorted by overall complexity descending):
   - Apply test-appropriate refactoring:
     - Convert to table-driven tests where applicable (preferred strategy).
     - Extract setup/assertion helpers with `t.Helper()` (<30 lines, cyclomatic <12).
     - Preserve all test coverage and pass/fail behavior.
   - If a test bug is discovered during review, fix it as part of this pass.
3. Run `go test -race ./...` after each refactoring.
4. Run `go vet ./...` to confirm no issues.
5. Rebuild and verify the tool still produces correct output:
   ```bash
   go install . && go-stats-generator analyze . --only-tests --format json | jq '.functions | length'
   ```

### Phase 3: Validate
```bash
go-stats-generator analyze . --only-tests --format json --output post.json --sections functions --max-complexity 15 --max-function-length 45
go-stats-generator diff baseline.json post.json
```
Confirm: all target test functions below thresholds, zero regressions, tool still functional.

## Thresholds (Test-Appropriate)
| Metric | Maximum |
|--------|---------|
| Overall complexity | 15.0 |
| Cyclomatic complexity | 15 |
| Function length | 45 lines |
| Nesting depth | 5 |
| Extracted helper length | 30 |
| Extracted helper cyclomatic | 12 |

> **Note**: Thresholds are relaxed by ~50% for test code.

## Self-Analysis Rules
- This is dogfooding — the tool analyzes its own tests.
- After refactoring, rebuild (`go install .`) and verify the tool still works.
- If test output changes after refactoring, investigate whether a test bug was introduced.
- May fix test bugs discovered during self-analysis.
- Prefer table-driven tests and `t.Helper()` extraction as primary strategies.

## Output Format
```
Self-analysis: [N] test functions above thresholds
Refactored:
  [function] [file]: [old] -> [new] ([reduction]%)
  Strategy: [table-driven | extract helper | decompose]
Test bugs fixed: [count] (or "none")
Tool verification: PASS
Tests: PASS
```

## Tiebreaker
Refactor the longest test function first when complexity scores are tied.
## Dogfooding Verification
After refactoring, the tool must still produce correct output:
1. Rebuild: `go install .`
2. Verify: `go-stats-generator analyze . --only-tests --format json | jq '.overview'`
3. Compare key metrics (test function count, avg complexity) against baseline — they should be stable or improved.
4. If output differs unexpectedly, investigate whether a bug was introduced during refactoring.

## Validation Checklist
- [ ] All target test functions below thresholds
- [ ] Tool still produces correct output after rebuild
- [ ] All tests pass with -race flag
- [ ] Diff shows zero regressions except intentional improvements
- [ ] Any test bugs discovered during review have been fixed
