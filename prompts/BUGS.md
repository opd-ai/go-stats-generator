# TASK: Identify and fix bug-prone functions using complexity metrics as risk indicators.

## Execution Mode
**Autonomous action** — analyze, fix bugs, validate with tests and diff.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Codebase
Before hunting bugs, learn the project's conventions:
1. Read the project README to understand its domain, users, and critical paths.
2. Discover the project's **error handling conventions**: Does it use sentinel errors, `%w` wrapping, custom error types, or multi-return? Bug fixes must match.
3. Identify how the project handles nil values, resource lifecycle, and concurrency.
4. Note the test strategy: what testing framework, assertion style, and coverage level exist.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output baseline.json --sections functions,patterns
```

### Phase 2: Identify and Fix
1. Extract high-risk functions (sorted by cyclomatic complexity descending). Thresholds are tunable defaults:
   - CRITICAL: cyclomatic >20 OR nesting >5
   - HIGH: cyclomatic >15 OR nesting >4
   - MEDIUM: cyclomatic 10–15 OR nesting 3–4
2. For each high-risk function, perform targeted code review using the project's own conventions as the standard:
   - Error handling: are errors silently ignored? Missing wrapping per project convention?
   - Nil pointer risks: pointers dereferenced without nil checks?
   - Slice/map access: possible index-out-of-range or nil-map writes?
   - Goroutine safety: shared variables accessed without synchronization?
   - Resource leaks: missing deferred closes for files/connections?
   - Concurrency patterns: missing context cancellation, unbuffered channel deadlocks?
3. For each confirmed bug, apply the minimal fix:
   - Preserve existing API contracts and behavior.
   - Match the project's error handling style.
   - Add nil/bounds checks before access.
   - Add missing defers for resource cleanup.
4. Run `go test -race ./...` after each fix.

### Phase 3: Validate
```bash
go-stats-generator analyze . --skip-tests --format json --output post.json --sections functions,patterns
go-stats-generator diff baseline.json post.json
```
Confirm: zero new regressions, all fixes preserve existing test coverage.

## Bug Risk Classification (tunable defaults)
| Risk Level | Criteria | Action |
|------------|----------|--------|
| CRITICAL | cyclomatic >20, nesting >5, concurrency without sync | Fix immediately |
| HIGH | cyclomatic >15, nesting >4, error returns ignored | Fix in this pass |
| MEDIUM | cyclomatic 10–15, nesting 3–4 | Fix if clear solution exists |
| LOW | cyclomatic <10 | Skip unless obvious bug found |

## Fix Rules
- Only fix bugs with clear, deterministic solutions.
- Preserve existing functionality and API contracts.
- Maintain code style consistency with surrounding code and project conventions.
- Skip fixes that require architectural changes or unclear requirements.

## Common Bug Patterns
1. **Ignored errors**: `val, _ := someFunc()` where error matters
2. **Nil dereference**: pointer used without nil check after conditional assignment
3. **Slice panic**: `slice[i]` without bounds check on user-controlled `i`
4. **Map write to nil**: `m[k] = v` without prior `make(map...)`
5. **Goroutine leak**: goroutine blocks on channel with no consumer
6. **Resource leak**: `os.Open` without corresponding `defer f.Close()`
7. **Race condition**: shared variable modified in goroutine without mutex

## Output Format
```
[SEVERITY] [function] [file:line] — [bug description] -> [fix applied]
Tests: PASS
```

## Tiebreaker
Fix the function with the highest cyclomatic complexity first. If tied, prefer deeper nesting.
