# TASK: Identify and fix bug-prone **test** functions using complexity metrics as risk indicators.

## Execution Mode
**Autonomous action** — analyze, fix test bugs, validate with tests and diff.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 1: Baseline
```bash
go-stats-generator analyze . --only-tests --format json --output baseline.json --sections functions,patterns
```

### Phase 2: Identify and Fix
1. Extract high-risk test functions from baseline JSON (sorted by cyclomatic complexity descending):
   - CRITICAL: cyclomatic >30 OR nesting >7
   - HIGH: cyclomatic >22 OR nesting >6
   - MEDIUM: cyclomatic 15–22 OR nesting 5–6
2. For each high-risk test function, perform targeted code review:
   - Error handling: are test setup errors silently ignored? Missing `t.Fatal` on setup failure?
   - Nil pointer risks: pointers dereferenced without nil checks in test assertions?
   - Resource leaks: missing cleanup for temp files, goroutines, or test servers?
   - Race conditions: shared test state accessed without synchronization? Missing `t.Parallel()` guards?
   - Flaky patterns: time-dependent assertions, hardcoded ports, file system assumptions?
   - Goroutine leaks: test goroutines that outlive the test function?
3. For each confirmed bug, apply the minimal fix:
   - Preserve existing test coverage and pass/fail semantics.
   - Add `t.Helper()` to extracted helpers.
   - Add `t.Cleanup()` for resource management.
   - Add proper error checks on test setup with `t.Fatal`.
4. Run `go test -race ./...` after each fix to confirm no regressions.

### Phase 3: Validate
```bash
go-stats-generator analyze . --only-tests --format json --output post.json --sections functions,patterns
go-stats-generator diff baseline.json post.json
```
Confirm: zero new regressions, all fixes preserve existing test behavior.

## Bug Risk Classification (Test-Appropriate)
| Risk Level | Criteria | Action |
|------------|----------|--------|
| CRITICAL | cyclomatic >30, nesting >7, race condition | Fix immediately |
| HIGH | cyclomatic >22, nesting >6, resource leak | Fix in this pass |
| MEDIUM | cyclomatic 15–22, nesting 5–6 | Fix if clear solution exists |
| LOW | cyclomatic <15 | Skip unless obvious bug found |

> **Note**: Test thresholds are relaxed by ~50% compared to production code.

## Fix Rules
- Only fix bugs with clear, deterministic solutions.
- Preserve existing test coverage and pass/fail behavior.
- Prefer table-driven test refactoring and `t.Helper()` helpers.
- Maintain code style consistency with surrounding test code.
- Skip fixes that require changing production code.

## Common Test Bug Patterns
1. **Ignored setup errors**: `f, _ := os.CreateTemp(...)` where error matters for test validity
2. **Missing cleanup**: temp files or test servers not cleaned up with `t.Cleanup()`
3. **Race in parallel tests**: shared slice/map modified in `t.Parallel()` subtests
4. **Flaky time assertions**: `time.Sleep` used for synchronization instead of channels
5. **Goroutine leak**: test goroutine blocks on channel after test returns
6. **Missing t.Helper()**: test helpers that don't call `t.Helper()`, hiding failure locations
7. **Hardcoded paths**: test assumes specific working directory or absolute paths

## Output Format
```
[SEVERITY] [function] [file:line] — [bug description] -> [fix applied]
Tests: PASS
```

## Tiebreaker
Fix the test function with the highest cyclomatic complexity first. If tied, prefer deeper nesting. If still tied, prefer test functions with more goroutines.
## Test Cleanup Checklist
- Check all test goroutine launches for proper lifecycle management.
- Verify all temp files/dirs use `t.TempDir()` or `t.Cleanup()`.
- Confirm parallel tests don't share mutable state.
- Check that test servers are properly shut down.
- Verify test helpers call `t.Helper()`.

## Validation Checklist
- [ ] All confirmed test bugs have fixes applied
- [ ] All fixes pass `go test -race ./...`
- [ ] No test coverage reduced
- [ ] Diff report shows zero complexity regressions
