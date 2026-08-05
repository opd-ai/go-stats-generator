# TASK: Perform a focused audit of arithmetic and numeric logic errors in Go code, identifying integer overflow, incorrect division, sign handling bugs, precision loss, bit manipulation errors, and incorrect numeric comparisons while rigorously preventing false positives.

## Output
Write two files in repo root: `AUDIT.md` (arithmetic audit report) and `GAPS.md` (numeric correctness gaps).

## Workflow

### Phase 0: Numeric Inventory
1. Read README for computational purpose.
2. Check `go.mod` for dependencies with numeric behavior.
3. Scan for: integer math on large values, division/modulo with user-controlled divisors, mixed-type arithmetic, explicit type conversions, bit operations, percentage/ratio computations, float comparisons, float-to-int conversions.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages > tmp/arithmetic-metrics.json
go test -race ./... && go vet ./...
```

### Phase 2: Audit Checklist
- [ ] **Overflow**: No silent integer overflow in accumulators, multipliers, or size calculations
- [ ] **Division**: No divide-by-zero; correct truncation vs rounding for percentages/ratios
- [ ] **Sign**: No uint underflow used as size; no sign errors in mixed signed/unsigned ops
- [ ] **Precision**: No float equality on computed values; no precision loss in accumulation
- [ ] **Conversions**: Safe int/float conversions; no truncation of large values
- [ ] **Bit ops**: Shifts within type width; no sign-extension bugs

### Phase 3: Report
**AUDIT.md**: Project numeric domain, findings by class (CRITICAL/HIGH/MEDIUM/LOW) with file:line, triggering value, remediation.

**GAPS.md**: Per gap — Stated Goal, Current State, Risk, Closing the Gap.

## Constraints
- Report ONLY — no code changes
- Every finding: file:line, bug class, concrete trigger, actionable fix
- No speculative findings
- Use unchecked `- [ ]` checkboxes
- Check Go version for numeric behavior changes