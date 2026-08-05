# TASK: Perform a logic bug audit of Go code covering arithmetic, boolean, boundary, aliasing, and initialization errors.

## Output
Write two files in repo root: `AUDIT.md` (logic audit report) and `GAPS.md` (logic correctness gaps).

## Workflow

### Phase 0: Logic Inventory
1. Read README for computational purpose.
2. Check `go.mod` for Go version (loop var capture fixed in 1.22+).
3. Scan for: integer math, division/modulo, type conversions, float ops, compound booleans, index access, slice arithmetic, pointer sharing, nil map/slice writes, init() ordering.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages,structs > tmp/logic-metrics.json
go test -race ./... && go vet ./...
```

### Phase 2: Audit Checklist
- [ ] **Arithmetic**: No div-by-zero, no uint underflow, correct int/float conversions, no precision loss
- [ ] **Boolean**: No De Morgan violations, no double-negation confusion, correct short-circuit eval
- [ ] **Boundary**: No off-by-one in loops/indexes, empty/nil checks before access, correct pagination bounds
- [ ] **Aliasing**: Slice/map copies before mutation, no loop var capture (pre-1.22), no shallow copy bugs
- [ ] **Initialization**: No nil map writes, no zero-value field misuse, init() order safe, constructors used

### Phase 3: Report
**AUDIT.md**: Project logic model, findings by class (CRITICAL/HIGH/MEDIUM/LOW) with file:line, triggering value/path, remediation.

**GAPS.md**: Per gap — Bug Class, Stated Goal, Current State, Risk, Closing the Gap.

## Constraints
- Report ONLY — no code changes
- Every finding: file:line, bug class, concrete trigger, actionable fix
- No speculative findings
- Use unchecked `- [ ]` checkboxes
- Check Go version before reporting range loop capture bugs