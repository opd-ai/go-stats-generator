# TASK: Perform a memory audit identifying leaks, excessive allocations, unsafe pointer misuse, and GC pressure.

## Output
Write two files in repo root: `AUDIT.md` (memory audit report) and `GAPS.md` (memory safety gaps).

## Workflow

### Phase 0: Memory Inventory
1. Read README for memory/performance claims.
2. Check `go.mod` for `unsafe`, `cgo`, pooling libs.
3. Scan for: `unsafe.Pointer`, `reflect.SliceHeader`, `runtime.SetFinalizer`, `sync.Pool`, large `make()` calls, string/byte conversions in loops, append without pre-allocation, global accumulators, closures capturing large objects.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages,structs > tmp/memory-metrics.json
go test -bench=. -benchmem ./... 2>&1 | tee tmp/mem-bench.txt
go build -gcflags='-m' ./... 2>&1 | tee tmp/escape.txt
```

### Phase 2: Audit Checklist
- [ ] **Leaks**: No unbounded map/slice/cache growth. No goroutines retaining large objects. No `resp.Body.Close()` missing. No `time.After` in loops.
- [ ] **Unsafe**: No `unsafe.Pointer`/`uintptr` conversions violating pointer rules. No `reflect.SliceHeader`/`StringHeader` (use `unsafe.Slice`/`String`).
- [ ] **CGO**: All `C.malloc`/`C.CString` freed. No cgo pointer passing violations.
- [ ] **Finalizers**: No `runtime.SetFinalizer` for critical cleanup (use explicit Close).
- [ ] **Allocations**: Pre-allocate known-size slices. Pool hot-path objects. Avoid string/byte conversion in loops.

### Phase 3: Report
**AUDIT.md**: Memory profile, inventory table, findings (CRITICAL/HIGH/MEDIUM/LOW) with file:line, allocation path/growth pattern, impact, remediation.

**GAPS.md**: Per gap — Stated Goal, Current State, Risk, Closing the Gap.

## Constraints
- Report ONLY — no code changes
- Every finding: file:line, measurable impact evidence, actionable fix
- No speculative findings — must demonstrate unbounded growth/OOM risk/GC pressure
- Use unchecked `- [ ]` checkboxes