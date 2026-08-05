# TASK: Perform a focused audit of initialization order and use-before-ready logic errors in Go code, identifying use-before-initialization, incorrect operation ordering, nil map/slice write panics, zero-value assumption violations, circular initialization, and time-of-check to time-of-use (TOCTOU) logic bugs while rigorously preventing false positives.

## Output
Write two files in repo root: `AUDIT.md` (initialization audit report) and `GAPS.md` (initialization correctness gaps).

## Workflow

### Phase 0: Initialization Inventory
1. Read README for startup sequence, config loading, data structure lifecycle.
2. Check `go.mod` for dependencies requiring init order (DB drivers, plugins).
3. Scan for: package-level `var` with cross-refs, `init()` functions, `sync.Once`, constructors (`New*`), struct zero values, nil map/slice writes, pointer fields, `flag.Parse()`/`viper.ReadInConfig()`, DB migrations.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages,structs > tmp/init-metrics.json
go test -race ./... && go vet ./...
```

### Phase 2: Audit Checklist
- [ ] **Nil map writes**: No `m[key] = val` on uninitialized `var m map[K]V`
- [ ] **Nil pointer deref**: Methods on nil receivers guarded; pointer fields checked
- [ ] **Zero value assumptions**: No division by zero counter; no boolean default disabling feature
- [ ] **Init order**: No circular init deps; `init()` order respects import graph; constructors validate inputs
- [ ] **Config before use**: `flag.Parse()`/`viper.ReadInConfig()` called before accessing flags/config
- [ ] **DB before query**: Migrations/migrations run before queries
- [ ] **Lazy init**: `sync.Once` used correctly; error handling in lazy init

### Phase 3: Report
**AUDIT.md**: Project initialization model, findings by class (CRITICAL/HIGH/MEDIUM/LOW) with file:line, triggering sequence, remediation.

**GAPS.md**: Per gap — Stated Goal, Current State, Risk, Closing the Gap.

## Constraints
- Report ONLY — no code changes
- Every finding: file:line, concrete triggering sequence, actionable fix
- No speculative findings
- Use unchecked `- [ ]` checkboxes
- Respect project's constructor/init patterns