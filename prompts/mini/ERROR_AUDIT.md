# TASK: Perform an error handling audit identifying swallowed errors, incorrect wrapping, panic/recover misuse, and propagation failures.

## Output
Write two files in repo root: `AUDIT.md` (error audit report) and `GAPS.md` (error handling gaps).

## Workflow

### Phase 0: Error Handling Inventory
1. Read README for reliability claims.
2. Check `go.mod` for error deps (pkg/errors, multierror).
3. Scan for: wrapping convention (`%w` vs `%v`), custom errors, sentinels, panic/recover, `_` discards, error return patterns.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages > tmp/error-metrics.json
go vet ./... && errcheck ./... 2>/dev/null || true
```

### Phase 2: Audit Checklist
- [ ] **Swallowed Errors**: No ignored `err` returns (not checked, returned, or logged). No `_` on critical paths (Close, Flush, Write).
- [ ] **Defer Close Errors**: `defer f.Close()` on writable files checks error.
- [ ] **Wrapping**: All error returns use `%w` (not `%v`) to preserve chain for `errors.Is/As`.
- [ ] **Panic/Recover**: No `panic` in library code for expected errors. `recover` only at top-level boundaries.
- [ ] **Goroutine Errors**: Errors from goroutines propagated via channel or sync primitive.
- [ ] **Nil Error Returns**: No `return nil, nil` or typed `nil` error that breaks `errors.Is`.

### Phase 3: Report
**AUDIT.md**: Error model summary, inventory table, findings (CRITICAL/HIGH/MEDIUM/LOW) with file:line, evidence, impact, remediation.

**GAPS.md**: Per gap — Stated Goal, Current State, Risk, Closing the Gap.

## Constraints
- Report ONLY — no code changes
- Every finding: file:line, consequence, actionable fix respecting project idioms
- Use unchecked `- [ ]` checkboxes