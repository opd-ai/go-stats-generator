# TASK: Perform a focused audit of data aliasing and mutation logic errors in Go code, identifying unintended shared state from slice/map aliasing, shallow copy bugs, unexpected mutation through pointers, iteration variable capture errors, and incorrect equality semantics while rigorously preventing false positives.

## Output
Write two files in repo root: `AUDIT.md` (aliasing audit report) and `GAPS.md` (data isolation gaps).

## Workflow

### Phase 0: Data Ownership Inventory
1. Read README for data types collected, transformed, stored.
2. Check `go.mod` for dependencies with shared buffers (AST, scanner, JSON).
3. Scan for: struct assignments, slice header copies `s2 := s1`, map assignments `m2 := m1`, pointer aliasing, pointer params mutated, return values sharing backing storage, struct fields that are pointers/slices/maps, `append` on shared slices, loop var capture in closures, shallow `copy()`.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages,structs > tmp/aliasing-metrics.json
go test -race ./... && go vet ./...
```

### Phase 2: Audit Checklist
- [ ] **Slice aliasing**: `copy()` or `append(nil, src...)` before storing/returning shared slices
- [ ] **Map aliasing**: `maps.Clone()` or explicit copy before storing/returning shared maps
- [ ] **Struct pointer fields**: Deep copy when struct copied; no shared mutable pointers
- [ ] **Loop capture**: No closure capturing `i`, `v` from `for range` (fixed in Go 1.22+)
- [ ] **Shared backing arrays**: `append` on shared slice doesn't corrupt other references
- [ ] **Scanner/buffer reuse**: No mutation of reused `bufio.Scanner` tokens or JSON decoder buffers

### Phase 3: Report
**AUDIT.md**: Project data ownership model, findings by class (CRITICAL/HIGH/MEDIUM/LOW) with file:line, mutation path, remediation.

**GAPS.md**: Per gap — Stated Goal, Current State, Risk, Closing the Gap.

## Constraints
- Report ONLY — no code changes
- Every finding: file:line, concrete mutation path, actionable fix
- No speculative findings
- Use unchecked `- [ ]` checkboxes
- Check Go version for loop var capture (fixed in 1.22+)
- Respect project's copy idioms