# TASK: Perform a comprehensive risk-prioritized functional audit of the Go codebase using `go-stats-generator` metrics as the primary evidence source.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output audit-metrics.json --sections functions,documentation,naming,patterns,duplication,interfaces,structs,packages
go-stats-generator analyze . --skip-tests
```

### Phase 2: Risk-Prioritized Audit
1. Extract function metrics and classify by risk:
   - **HIGH RISK**: length >50 lines OR cyclomatic >15 OR params >7
   - **MEDIUM RISK**: length >30 lines OR cyclomatic >10 OR params >5
   - **LOW RISK**: all metrics within thresholds
2. For each HIGH RISK function, perform detailed code review:
   - Verify error handling completeness
   - Check for nil pointer risks
   - Verify concurrency safety (cross-reference `.patterns.concurrency_patterns`)
   - Check documentation accuracy
3. For MEDIUM RISK functions on critical paths, perform targeted review.
4. Cross-reference findings with:
   - `.duplication.clone_pairs` for duplicated code risks
   - `.naming` for convention violations
   - `.documentation` for undocumented public APIs
5. Run `go test -race ./...` and `go vet ./...` for baseline health.

### Phase 3: Report
Generate `AUDIT.md` in the repository root with findings organized by risk level, each with:
- Severity classification (CRITICAL/HIGH/MEDIUM/LOW)
- File:line reference
- Metric evidence (complexity, length, coverage)
- Remediation recommendation

## Thresholds
| Risk Level | Criteria |
|------------|----------|
| HIGH | length >50 OR cyclomatic >15 OR params >7 |
| MEDIUM | length >30 OR cyclomatic >10 OR params >5 |
| LOW | within all thresholds |

## Report Structure
```markdown
# AUDIT — [date]
## Risk Summary
[HIGH: N functions, MEDIUM: N functions, critical findings: N]

## Findings
### CRITICAL
- [ ] [Finding] — [file:line] — [evidence] — [remediation]

### HIGH / MEDIUM / LOW
- [ ] ...
```

## Tiebreaker
Prioritize: HIGH RISK → critical-path functions → MEDIUM RISK → LOW RISK. Within a level, highest complexity first.
## Audit Scope
- All packages in the repository (excluding vendor/ and testdata/).
- All exported symbols and their documentation.
- All error handling paths in HIGH RISK functions.
- Concurrency patterns and their safety.

## Constraints
- Output ONLY `AUDIT.md` — no code changes.
- Use `go-stats-generator` metrics as primary evidence for all findings.
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes.
- Distinguish production code issues from test-only issues.

## Validation Checklist
- [ ] All HIGH RISK functions reviewed
- [ ] All findings have file:line references
- [ ] All findings have metric evidence
- [ ] AUDIT.md follows the severity classification structure
