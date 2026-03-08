# TASK: Perform a functional audit comparing documented behavior (README) against actual implementation, and output findings to an audit report.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project
1. Read the project README thoroughly — it is the primary source of behavioral claims to verify.
2. Examine `go.mod` for module path, Go version, and dependencies.
3. List packages (`go list ./...`) and understand the project's architecture.
4. Identify any other documentation (API docs, user guides, `--help` output) that makes verifiable claims.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output audit-baseline.json --sections functions,documentation,naming,packages
```

### Phase 2: Audit
1. Extract every claimed feature and capability from the README and other docs.
2. For each claim, verify against `go-stats-generator` JSON output and manual code inspection:
   - Does the feature exist in the codebase?
   - Does it produce correct output when invoked?
   - Are there edge cases or partial implementations?
3. Cross-reference function metrics for risk indicators (tunable defaults):
   - Functions with cyclomatic complexity >15 or length >50 lines are high-risk.
   - Packages with <70% doc coverage may have undocumented behavioral differences.
4. Run `go test -race ./...` and `go vet ./...` to confirm baseline health.
5. Inspect each package for internal consistency (exported symbols vs. doc coverage).

### Phase 3: Report
Generate an audit document in the repository root:

```markdown
# AUDIT — [date]
## Project Context
[What the project claims to do, its type, and its audience]
## Summary
[Overall health, count of findings by severity]
## Findings
### CRITICAL
- [ ] [Finding title] — [file:line] — [description with evidence] — **Remediation:** [complete, production-ready fix with specific steps]
### HIGH / MEDIUM / LOW
- [ ] ...
## Metrics Snapshot
[Key numbers: total functions, avg complexity, doc coverage, duplication ratio]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | Feature documented but non-functional, or data corruption risk |
| HIGH | Feature partially implemented, or high complexity with no tests |
| MEDIUM | Edge case failures, or documentation coverage gap >20% |
| LOW | Style issues, minor naming inconsistencies |

## Default Thresholds (calibrate to project)
- Cyclomatic complexity warning: >10
- Function length warning: >30 lines
- Doc coverage minimum: 70%
- High-risk function: cyclomatic >15 OR length >50 OR params >7

## Remediation Standards
Every finding MUST include a **Remediation** section that meets these requirements:

1. **Production-ready**: The recommended fix must be a complete, deployable solution — not a sketch, placeholder, or partial workaround. If the fix requires multiple steps, list all of them.
2. **No half measures**: Do not recommend "consider doing X" or "investigate Y." State exactly what to change, in which file, and how. If a function needs refactoring, specify the extraction points and resulting function signatures. If error handling is missing, show the exact error propagation path.
3. **No missing features**: If the remediation requires new code (helper functions, types, tests), describe all of it. Do not leave gaps for the implementer to fill in. Include validation steps (`go test`, `go vet`, `go-stats-generator diff`) that prove the fix is complete.
4. **Respect project idioms**: Recommendations must follow the existing codebase's conventions for error handling, naming, package structure, and testing patterns. Study the codebase before prescribing solutions.
5. **Verifiable**: Every remediation must include a concrete validation command or check that confirms the fix works. Example: "Run `go test -race ./pkg/...` and verify zero failures" or "Run `go-stats-generator analyze . --format json | jq '.complexity'` and confirm average < 10."

### Remediation Template
Each finding's remediation should follow this structure:
- **What**: Specific file(s) and function(s) to change
- **How**: Exact changes required (refactoring steps, new code, configuration)
- **Why**: Metric evidence justifying the change
- **Validate**: Command(s) to confirm the fix is correct and complete

## Constraints
- Output ONLY the audit report — no code changes permitted.
- Use `go-stats-generator` metrics as primary evidence source.
- Verify against the currently installed binary, not an older cached version.
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Every finding must include a production-ready remediation — findings without actionable fixes are incomplete.
- If a prior audit exists, diff findings against it and note new vs. known issues.

## Tiebreaker
Prioritize by severity (CRITICAL > HIGH > MEDIUM > LOW), then by descending cyclomatic complexity.
