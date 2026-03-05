# TASK: Use `go-stats-generator` to analyze and refactor its own codebase (dogfooding) — reduce complexity of its own functions below thresholds.

## Execution Mode
**Autonomous action** — self-refactor and fix discovered bugs, validate with tests and diff.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 1: Self-Analysis Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output baseline.json --sections functions --max-complexity 10 --max-function-length 30
go-stats-generator analyze . --skip-tests --max-complexity 10 --max-function-length 30
```

### Phase 2: Self-Refactor
1. From baseline, identify functions in the `go-stats-generator` codebase exceeding thresholds.
2. For each violating function (sorted by overall complexity descending):
   - Apply extract-method refactoring (same rules as BREAKDOWN.md):
     - Extract cohesive blocks into named helpers (<20 lines, cyclomatic <8).
     - Preserve all public API signatures.
     - Use verb-first Go naming conventions.
   - If a bug is discovered during review, fix it as part of this pass.
3. Run `go test -race ./...` after each refactoring.
4. Run `go vet ./...` to confirm no issues.
5. Rebuild and verify the tool still produces correct output:
   ```bash
   go install . && go-stats-generator analyze . --skip-tests --format json | jq '.functions | length'
   ```

### Phase 3: Validate
```bash
go-stats-generator analyze . --skip-tests --format json --output post.json --sections functions --max-complexity 10 --max-function-length 30
go-stats-generator diff baseline.json post.json
```
Confirm: all target functions below thresholds, zero regressions, tool still functional.

## Thresholds
| Metric | Maximum |
|--------|---------|
| Overall complexity | 10.0 |
| Cyclomatic complexity | 10 |
| Function length | 30 lines |
| Nesting depth | 3 |
| Extracted function length | 20 |
| Extracted function cyclomatic | 8 |

## Self-Analysis Rules
- This is dogfooding — the tool analyzes itself.
- After refactoring, rebuild (`go install .`) and verify the tool still works.
- If the tool's own output changes after refactoring (beyond metric improvements), investigate whether a bug was introduced.
- May fix bugs discovered during self-analysis (unlike BREAKDOWN.md which is refactor-only).

## Output Format
```
Self-analysis: [N] functions above thresholds
Refactored:
  [function] [file]: [old] -> [new] ([reduction]%)
Bugs fixed: [count] (or "none")
Tool verification: PASS
Tests: PASS
```

## Tiebreaker
Refactor the longest function first when complexity scores are tied.
## Dogfooding Verification
After refactoring, the tool must still produce correct output:
1. Rebuild: `go install .`
2. Verify: `go-stats-generator analyze . --skip-tests --format json | jq '.overview'`
3. Compare key metrics (function count, avg complexity) against baseline — they should be stable or improved.
4. If output differs unexpectedly, investigate whether a bug was introduced during refactoring.

## Validation Checklist
- [ ] All target functions below thresholds
- [ ] Tool still produces correct output after rebuild
- [ ] All tests pass with -race flag
- [ ] Diff shows zero regressions except intentional improvements
- [ ] Any bugs discovered during review have been fixed
