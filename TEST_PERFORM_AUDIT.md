# TASK: Perform a comprehensive risk-prioritized functional audit of **test code** using `go-stats-generator` metrics as the primary evidence source.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Test Strategy
1. Read the project README to understand its domain and expected behavior.
2. Discover the test framework in use and the project's assertion patterns.
3. Identify the test organization: how are tests structured, what testing conventions exist?
4. Note whether the project uses `t.Parallel()`, `t.Cleanup()`, test suites, or integration test separation.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --only-tests --format json --output test-audit-metrics.json --sections functions,documentation,naming,patterns,duplication
go-stats-generator analyze . --only-tests
```

### Phase 2: Risk-Prioritized Audit
1. Extract test function metrics and classify by risk (tunable defaults):
   - **HIGH RISK**: length >75 lines OR cyclomatic >22 OR nesting >7
   - **MEDIUM RISK**: length >45 lines OR cyclomatic >15 OR nesting >5
   - **LOW RISK**: within test-appropriate thresholds
2. For each HIGH RISK test function, review:
   - Test setup error handling (are setup failures caught with `t.Fatal`?)
   - Race conditions in parallel tests
   - Resource cleanup (`t.Cleanup()`, temp files, goroutines)
   - Flaky test patterns (timing, hardcoded ports, file system)
   - Documentation accuracy for test helpers
3. Cross-reference with `.duplication.clone_pairs`, `.naming`, and `.documentation`.

### Phase 3: Report
Generate a test audit document with findings organized by risk level:
```markdown
# TEST AUDIT — [date]
## Test Infrastructure Context
[Test framework, conventions, coverage approach]
## Risk Summary
[HIGH: N, MEDIUM: N, critical findings: N]
## Findings
### CRITICAL
- [ ] [Finding] — [file:line] — [evidence] — [remediation]
### HIGH / MEDIUM / LOW
- [ ] ...
```

## Risk Thresholds (test-appropriate — ~50% relaxed)
| Risk Level | Criteria |
|------------|----------|
| HIGH | length >75 OR cyclomatic >22 OR nesting >7 |
| MEDIUM | length >45 OR cyclomatic >15 OR nesting >5 |
| LOW | within thresholds |

## Constraints
- Output ONLY the audit report — no code changes.
- Use `go-stats-generator --only-tests` metrics as primary evidence.
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes.
- Recommend table-driven tests and `t.Helper()` extraction as primary remediation.

## Tiebreaker
Prioritize: HIGH RISK → MEDIUM RISK → LOW RISK. Within a level, highest complexity first.
