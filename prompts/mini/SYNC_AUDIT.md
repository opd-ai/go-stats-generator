# TASK: Perform a focused audit of concurrency and synchronization problems in Go code, identifying deadlocks, race conditions, and resource conflicts while rigorously preventing false positives.

## Output
Write two files in repo root: `AUDIT.md` (concurrency audit report) and `GAPS.md` (concurrency safety gaps).

## Workflow

### Phase 0: Concurrency Inventory
1. Read README for concurrency claims and parallelism model.
2. Check `go.mod` for concurrency dependencies (`golang.org/x/sync`, `sourcegraph/conc`).
3. Scan for: `go` statements, `chan` operations, `sync.Mutex`/`RWMutex`/`WaitGroup`/`Once`/`Map`/`Cond`/`Pool`, `atomic.*`, `context.Context`, `errgroup.Group`.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages > tmp/sync-metrics.json
go test -race ./... && go vet ./...
```
**Critical**: `-race` detector results are primary evidence — confirmed races are CRITICAL.

### Phase 2: Audit Checklist
- [ ] **Mutex deadlocks**: No multi-lock without consistent ordering; no recursive locking; `Lock()`/`Unlock()` paired on all paths (use `defer`)
- [ ] **RWMutex**: No `RLock()` followed by `Lock()` in same goroutine
- [ ] **Channels**: Every unbuffered send has receiver; buffered channels sized; all channels closed by sender; no indefinite blocking without `context.Done()`/`timeout`
- [ ] **Goroutine leaks**: All goroutines exit on context cancel; no orphaned goroutines
- [ ] **Shared state**: No unprotected mutable state on concurrent paths; `atomic` or mutex for shared vars
- [ ] **Context**: Propagated through call chains; cancellation respected

### Phase 3: Report
**AUDIT.md**: Project concurrency model, `-race` results, findings by class (CRITICAL/HIGH/MEDIUM/LOW) with file:line, execution path, remediation.

**GAPS.md**: Per gap — Stated Goal, Current State, Risk, Closing the Gap.

## Constraints
- Report ONLY — no code changes
- Every finding: file:line, concrete execution path, actionable fix
- `-race` detections = confirmed CRITICAL findings
- No speculative findings
- Use unchecked `- [ ]` checkboxes
- Respect project's channel vs mutex idioms