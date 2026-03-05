# TASK: Discover all audit files, extract unchecked findings, enrich with metrics, and produce a single prioritized consolidated audit.

## Execution Mode
**Report generation only** — do NOT modify source code or existing audit files.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Codebase
1. Read the project README to understand its purpose and architecture.
2. Examine `go.mod` to understand the module structure.
3. Note the project's conventions for error handling, testing, and code organization — findings should be evaluated in this context.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output audit-metrics.json --sections functions,packages,documentation
```

### Phase 2: Collate
1. Find all audit-related files in the repository:
   ```bash
   find . -name '*AUDIT*.md' -not -path './vendor/*'
   ```
2. From each file, extract every unchecked finding (`- [ ]` items).
3. Skip findings that are test-only or already resolved (checked `- [x]`).
4. For each finding, look up the referenced function/file in the `go-stats-generator` JSON:
   - Add cyclomatic complexity, line count, and doc coverage to the finding.
   - Escalate severity if metrics indicate higher risk (never downgrade).
5. Deduplicate findings that appear in multiple audit files (keep the highest severity version).

### Phase 3: Generate Consolidated Audit
```markdown
# AUDIT — Collated [date]
## Project Context
[Project type and key architectural observations]
## Summary
[Total findings, breakdown by severity, source audit files]
## CRITICAL
- [ ] [Finding] — [file:line] — complexity: [N], lines: [N] — [remediation steps]
## HIGH / MEDIUM / LOW
- [ ] ...
## Source Audits
[List of audit files discovered and their finding counts]
```

## Severity Escalation Rules
Metrics can only **escalate** severity, never downgrade:
| Original Severity | Escalate to CRITICAL if | Escalate to HIGH if |
|-------------------|------------------------|---------------------|
| HIGH | complexity >20 OR lines >60 | — |
| MEDIUM | complexity >20 | cyclomatic >10 OR lines >40 |
| LOW | complexity >20 | complexity >15 OR cyclomatic >10 |

## Remediation Instructions
Each finding must include:
1. What to change (specific function/file)
2. Why (metric evidence)
3. How to validate (`go test`, `go-stats-generator diff`)

## Deduplication Rules
- Keep the version with: highest severity, most specific file:line reference, most detailed remediation.
- Note the source audit files for each finding.

## Output Rules
- Only output the collated audit — do not modify any other files.
- Order: CRITICAL → HIGH → MEDIUM → LOW, then descending complexity within group.

## Tiebreaker
Within a severity group, order by descending complexity score. If tied, line count descending.
