# TASK: Discover all audit files, extract unchecked findings, enrich with metrics, and produce a single prioritized consolidated audit with an accompanying gaps analysis.

## Execution Mode
**Report generation only** — do NOT modify source code or existing audit files.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the consolidated audit report
2. **`GAPS.md`** — consolidated gaps between stated goals and implementation

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project's Goals
1. Read the project README to understand its purpose, stated goals, and architecture.
2. Examine `go.mod` to understand the module structure.
3. Note the project's conventions — findings should be evaluated in the context of the project's own stated goals, not arbitrary external standards.

### Phase 1: Online Research
Use web search to build context that isn't available in the repository:
1. Search for the project on GitHub — read open issues and discussions to understand known pain points and user feedback.
2. Research key dependencies for known vulnerabilities or deprecations that may affect existing audit findings.

Keep research brief (≤10 minutes). Record only findings relevant to the project's stated goals.

### Phase 2: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --sections functions,packages,documentation,duplication > /tmp/audit-metrics.json
```
Delete `/tmp/audit-metrics.json` when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Collate
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
6. Tag each finding with which stated project goal it affects (if any).
7. Flag any finding that describes a confirmed or likely bug (logic error, nil dereference, resource leak, race condition). Bugs on critical paths should be escalated to at least HIGH severity.
8. When `go vet` or linter warnings appear in findings, check whether the surrounding code has comments that explicitly acknowledge the warning (e.g., `//nolint:`, an explanatory comment justifying the pattern, or a TODO tracking the issue). If so, mark the finding as an acknowledged false positive and do not escalate it.

### Phase 4: Generate Consolidated Audit and Gaps

Generate **`AUDIT.md`**:
```markdown
# AUDIT — Collated [date]

## Project Goals
[What the project claims to do, extracted from README]

## Goal-Achievement Summary
| Goal | Status | Blocking Findings |
|------|--------|-------------------|
| [Stated goal] | ✅ / ⚠️ / ❌ | [finding IDs] |

## Summary
[Total findings, breakdown by severity, source audit files]

## CRITICAL
- [ ] [Finding] — [file:line] — complexity: [N], lines: [N] — [which goal this blocks] — [remediation steps]

## HIGH / MEDIUM / LOW
- [ ] ...

## Source Audits
[List of audit files discovered and their finding counts]
```

Generate **`GAPS.md`**:
```markdown
# Implementation Gaps — Collated [date]

## [Gap Title]
- **Stated Goal**: [what the README/docs promise]
- **Current State**: [synthesized from multiple audit findings]
- **Impact**: [how this gap affects users or the project's mission]
- **Closing the Gap**: [what needs to happen, referencing specific findings]
- **Source Audits**: [which audit files identified this gap]
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
2. Why (which stated goal this supports)
3. How to validate (`go test`, `go-stats-generator diff`)

## Deduplication Rules
- Keep the version with: highest severity, most specific file:line reference, most detailed remediation.
- Note the source audit files for each finding.

## Output Rules
- Only output the two consolidated files — do not modify any other files.
- Order: CRITICAL → HIGH → MEDIUM → LOW, then descending complexity within group.

## Tiebreaker
Within a severity group, prioritize findings that block stated project goals over cosmetic issues. Then order by descending complexity score. If tied, line count descending.
