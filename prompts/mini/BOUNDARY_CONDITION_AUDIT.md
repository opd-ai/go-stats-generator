# TASK: Perform a focused audit of boundary condition and off-by-one logic errors in Go code, identifying incorrect index arithmetic, fence-post errors, loop termination bugs, and edge-case mishandling while rigorously preventing false positives.

## Output
Write two files in repo root: `AUDIT.md` (boundary audit report) and `GAPS.md` (boundary correctness gaps).

## Workflow

### Phase 0: Boundary Inventory
1. Read README for data structures and iteration patterns.
2. Check `go.mod` for dependencies with pagination/indexed APIs.
3. Scan for: direct index access `s[i]`, slice arithmetic `s[i:j]`, loop bounds, off-by-one comparisons `<` vs `<=`, nil/empty collection access, string index `-1` handling, pagination arithmetic.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages > tmp/boundary-metrics.json
go test -race ./... && go vet ./...
```

### Phase 2: Audit Checklist
- [ ] **Index access**: Bounds checked before `s[i]` or `s[i:j]`
- [ ] **Loop bounds**: Correct `<` vs `<=`; no off-by-one in `for i := 0; i < len(s); i++`
- [ ] **Empty/nil guards**: Check `len(s) > 0` before accessing element 0
- [ ] **String ops**: Handle `-1` from `strings.Index`, `bytes.Index`, `strings.Cut`
- [ ] **Pagination**: Correct `offset + limit`, `page * pageSize` arithmetic
- [ ] **Fence-post**: Element count vs index arithmetic correct

### Phase 3: Report
**AUDIT.md**: Project data access profile, findings by class (CRITICAL/HIGH/MEDIUM/LOW) with file:line, counterexample input, remediation.

**GAPS.md**: Per gap — Stated Goal, Current State, Risk, Closing the Gap.

## Constraints
- Report ONLY — no code changes
- Every finding: file:line, concrete counterexample, actionable fix
- No speculative findings
- Use unchecked `- [ ]` checkboxes
- Evaluate against project's own indexing conventions