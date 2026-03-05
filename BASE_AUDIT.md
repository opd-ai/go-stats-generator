# TASK: Perform a functional audit of the Go codebase, comparing documented behavior (README.md) against actual implementation, and output findings to AUDIT.md.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output audit-baseline.json --sections functions,documentation,naming,packages
```

### Phase 2: Audit
1. Read README.md and extract every claimed feature and capability.
2. For each claim, verify against the `go-stats-generator` JSON output and manual code inspection:
   - Does the feature exist in the codebase?
   - Does it produce correct output when invoked?
   - Are there edge cases or partial implementations?
3. Cross-reference function metrics for risk indicators:
   - Functions with cyclomatic complexity >15 or length >50 lines are high-risk.
   - Packages with <70% doc coverage may have undocumented behavioral differences.
4. Run `go test -race ./...` to confirm existing tests pass.
5. Run `go vet ./...` to check for static analysis issues.
6. Inspect each package for internal consistency (exported symbols vs. doc coverage).

### Phase 3: Report
Generate `AUDIT.md` in the repository root:

```markdown
# AUDIT — [date]
## Summary
[One paragraph: overall health, count of findings by severity]

## Findings

### CRITICAL
- [ ] [Finding title] — [file:line] — [description with evidence]

### HIGH
- [ ] ...

### MEDIUM
- [ ] ...

### LOW
- [ ] ...

## Metrics Snapshot
[Key numbers: total functions, avg complexity, doc coverage, duplication ratio]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | Feature documented but non-functional, or data corruption risk |
| HIGH | Feature partially implemented, or complexity >15 with no tests |
| MEDIUM | Edge case failures, or documentation coverage gap >20% |
| LOW | Style issues, minor naming inconsistencies |

## Thresholds
- Cyclomatic complexity warning: >10
- Function length warning: >30 lines
- Doc coverage minimum: 70%
- High-risk function: cyclomatic >15 OR length >50 OR params >7

## Tiebreaker
Prioritize by severity (CRITICAL > HIGH > MEDIUM > LOW), then by descending cyclomatic complexity.

## Constraints
- Output ONLY `AUDIT.md` — no code changes permitted.
- Use `go-stats-generator` metrics as primary evidence source for all findings.
- Verify against the currently installed binary, not an older cached version.
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Distinguish between production code issues and test-only issues.
- If a prior AUDIT.md exists, diff findings against it and note new vs. known issues.
