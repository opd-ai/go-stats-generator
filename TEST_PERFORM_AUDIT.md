# TASK: Perform a comprehensive risk-prioritized functional audit of **test code** using `go-stats-generator` metrics as the primary evidence source.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 1: Baseline
```bash
go-stats-generator analyze . --only-tests --format json --output test-audit-metrics.json --sections functions,documentation,naming,patterns,duplication
go-stats-generator analyze . --only-tests
```

### Phase 2: Risk-Prioritized Audit
1. Extract test function metrics and classify by risk:
   - **HIGH RISK**: length >75 lines OR cyclomatic >22 OR nesting >7
   - **MEDIUM RISK**: length >45 lines OR cyclomatic >15 OR nesting >5
   - **LOW RISK**: all metrics within test-appropriate thresholds
2. For each HIGH RISK test function, perform detailed review:
   - Verify test setup error handling (are setup failures caught with `t.Fatal`?)
   - Check for race conditions in parallel tests
   - Verify resource cleanup (`t.Cleanup()`, temp files, goroutines)
   - Check for flaky test patterns (timing, hardcoded ports, file system)
   - Check documentation accuracy for test helpers
3. For MEDIUM RISK test functions, perform targeted review.
4. Cross-reference findings with:
   - `.duplication.clone_pairs` for duplicated test code
   - `.naming` for test naming convention violations
   - `.documentation` for undocumented test helpers

### Phase 3: Report
Generate `TEST_AUDIT.md` in the repository root with findings organized by risk level, each with:
- Severity classification (CRITICAL/HIGH/MEDIUM/LOW)
- File:line reference
- Metric evidence (complexity, length, coverage)
- Remediation recommendation (prefer table-driven tests, `t.Helper()` extraction)

## Thresholds (Test-Appropriate)
| Risk Level | Criteria |
|------------|----------|
| HIGH | length >75 OR cyclomatic >22 OR nesting >7 |
| MEDIUM | length >45 OR cyclomatic >15 OR nesting >5 |
| LOW | within all test-appropriate thresholds |

> **Note**: Risk thresholds are relaxed by ~50% for test code.

## Report Structure
```markdown
# TEST AUDIT — [date]
## Risk Summary
[HIGH: N test functions, MEDIUM: N test functions, critical findings: N]

## Findings
### CRITICAL
- [ ] [Finding] — [file:line] — [evidence] — [remediation]

### HIGH / MEDIUM / LOW
- [ ] ...
```

## Tiebreaker
Prioritize: HIGH RISK → MEDIUM RISK → LOW RISK. Within a level, highest complexity first.
## Audit Scope
- All `*_test.go` files in the repository.
- All exported test helpers and their documentation.
- All test setup/teardown paths in HIGH RISK test functions.
- Race condition risks in parallel tests.

## Constraints
- Output ONLY `TEST_AUDIT.md` — no code changes.
- Use `go-stats-generator --only-tests` metrics as primary evidence for all findings.
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes.
- Recommend table-driven tests and `t.Helper()` extraction as primary remediation strategies.

## Validation Checklist
- [ ] All HIGH RISK test functions reviewed
- [ ] All findings have file:line references
- [ ] All findings have metric evidence
- [ ] TEST_AUDIT.md follows the severity classification structure
