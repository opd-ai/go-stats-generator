# TASK: Discover all `*AUDIT*.md` files, extract unchecked findings, enrich with metrics, and produce a single prioritized root-level AUDIT.md.

## Execution Mode
**Report generation only** — do NOT modify source code or existing audit files.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output audit-metrics.json --sections functions,packages,documentation
```

### Phase 2: Collate
1. Find all `*AUDIT*.md` files in the repository:
   ```bash
   find . -name '*AUDIT*.md' -not -path './vendor/*'
   ```
2. From each file, extract every unchecked finding (`- [ ]` items).
3. Skip findings that are test-only or already resolved (checked `- [x]`).
4. For each finding, look up the referenced function/file in the `go-stats-generator` JSON:
   - Add cyclomatic complexity, line count, and doc coverage to the finding.
   - Escalate severity if metrics indicate higher risk (never downgrade).
5. Deduplicate findings that appear in multiple audit files (keep the highest severity version).

### Phase 3: Generate AUDIT.md
Produce a single `AUDIT.md` in the repository root:

```markdown
# AUDIT — Collated [date]
## Summary
[Total findings, breakdown by severity, source audit files]

## CRITICAL
- [ ] [Finding] — [file:line] — complexity: [N], lines: [N] — [remediation steps]

## HIGH
- [ ] ...

## MEDIUM
- [ ] ...

## LOW
- [ ] ...

## Source Audits
[List of *AUDIT*.md files discovered and their finding counts]
```

## Severity Escalation Rules
Metrics can only **escalate** severity, never downgrade:
| Original Severity | Escalate to CRITICAL if | Escalate to HIGH if |
|-------------------|------------------------|---------------------|
| HIGH | complexity >20 OR cyclomatic >15 OR lines >60 | — |
| MEDIUM | complexity >20 OR cyclomatic >15 | complexity >15 OR cyclomatic >10 OR lines >40 |
| LOW | complexity >20 | complexity >15 OR cyclomatic >10 |

## Thresholds
- Cyclomatic complexity warning: >10
- Function length warning: >30 lines
- Doc coverage minimum: 70%

## Remediation Instructions
Each finding must include step-by-step remediation:
1. What to change (specific function/file)
2. Why it needs changing (metric evidence)
3. How to validate the fix (`go test`, `go-stats-generator diff`)

## Output Rules
- Only output `AUDIT.md` — do not modify any other files.
- Exclude test-only findings unless they indicate production-code bugs.
- Order findings: CRITICAL → HIGH → MEDIUM → LOW, then by descending complexity within each group.

## Tiebreaker
Within a severity group, order by descending complexity score. If tied, order by line count descending.
## Deduplication Rules
- If the same finding appears in multiple audit files, keep the version with:
  1. Highest severity
  2. Most specific file:line reference
  3. Most detailed remediation steps
- Note the source audit files for each finding.
