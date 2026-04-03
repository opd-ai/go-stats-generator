# TASK: Identify and fix bug-prone **test** functions using complexity metrics as risk indicators.

## Execution Mode
**Autonomous action** — analyze, fix test bugs, validate with tests and diff.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Test Strategy
1. Read the project README and understand the testing philosophy.
2. Discover the test framework in use and the project's assertion patterns.
3. Identify how the project handles test setup/teardown, resource cleanup, and parallel tests.
4. Note whether the project uses `t.Cleanup()`, `t.TempDir()`, or manual cleanup patterns.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --only-tests --format json --output baseline.json --sections functions,patterns
```

### Phase 2: Identify and Fix
1. Extract high-risk test functions (sorted by cyclomatic complexity descending). Tunable defaults:
   - CRITICAL: cyclomatic >30 OR nesting >7
   - HIGH: cyclomatic >22 OR nesting >6
   - MEDIUM: cyclomatic 15–22 OR nesting 5–6
2. For each high-risk test function, perform targeted review:
   - Error handling: are setup errors silently ignored? Missing `t.Fatal` on setup failure?
   - Resource leaks: missing cleanup for temp files, goroutines, or test servers?
   - Race conditions: shared test state accessed without synchronization?
   - Flaky patterns: time-dependent assertions, hardcoded ports, file system assumptions?
   - Goroutine leaks: test goroutines that outlive the test function?
3. For each confirmed bug, apply the minimal fix matching the project's test conventions:
   - Preserve existing test coverage and pass/fail semantics.
   - Add `t.Helper()` to extracted helpers.
   - Add `t.Cleanup()` for resource management.
4. Run `go test -race ./...` after each fix.

### Phase 3: Validate
```bash
go-stats-generator analyze . --only-tests --format json --output post.json --sections functions,patterns
go-stats-generator diff baseline.json post.json
```

## Bug Risk Classification (test-appropriate — ~50% relaxed)
| Risk Level | Criteria | Action |
|------------|----------|--------|
| CRITICAL | cyclomatic >30, nesting >7, race condition | Fix immediately |
| HIGH | cyclomatic >22, nesting >6, resource leak | Fix in this pass |
| MEDIUM | cyclomatic 15–22, nesting 5–6 | Fix if clear solution exists |
| LOW | cyclomatic <15 | Skip unless obvious bug found |

## Common Test Bug Patterns
1. **Ignored setup errors**: `f, _ := os.CreateTemp(...)` where error matters
2. **Missing cleanup**: temp files/servers not cleaned up with `t.Cleanup()`
3. **Race in parallel tests**: shared slice/map modified in `t.Parallel()` subtests
4. **Flaky time assertions**: `time.Sleep` for synchronization instead of channels
5. **Goroutine leak**: test goroutine blocks on channel after test returns
6. **Missing t.Helper()**: helpers that don't call `t.Helper()`, hiding failure locations

## Output Format
```
[SEVERITY] [function] [file:line] — [bug description] -> [fix applied]
Tests: PASS
```

## Tiebreaker
Fix the test function with the highest cyclomatic complexity first. If tied, prefer deeper nesting.
