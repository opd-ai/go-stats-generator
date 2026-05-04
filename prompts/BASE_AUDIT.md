# TASK: Perform a functional audit comparing the project's stated goals against its actual implementation, and produce a root-level audit with an accompanying gaps analysis.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the audit report
2. **`GAPS.md`** — gaps between stated goals and implementation

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project's Goals
1. Read the project README thoroughly — it is the primary source of behavioral claims and goals to verify.
2. Extract every stated goal, feature claim, capability, performance target, and audience promise. These form the **audit checklist**.
3. Examine `go.mod` for module path, Go version, and dependencies.
4. List packages (`go list ./...`) and understand the project's architecture — which packages serve which goals.
5. Identify any other documentation (API docs, user guides, design docs, `--help` output) that makes verifiable claims.
6. Note the project's own conventions for error handling, testing, and code organization — evaluate the code against its own standards, not external ones.

### Phase 1: Online Research
Use web search to build context that isn't available in the repository:
1. Search for the project on GitHub — read open issues, recent PRs, and community discussions to understand known pain points.
2. Research key dependencies from `go.mod` for known vulnerabilities, deprecations, or upcoming breaking changes.
3. Look up best practices in the project's domain to calibrate audit expectations against its stated goals.

Keep research brief (≤10 minutes). Record only findings that are directly relevant to the project's stated goals.

### Phase 2: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --sections functions,documentation,packages,patterns,duplication > tmp/audit-baseline.json
go-stats-generator analyze . --skip-tests
```
Delete `tmp/audit-baseline.json` when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Goal-Achievement Audit
1. For each stated goal or feature claim from Phase 0, perform a systematic audit:
   - **Does the feature exist in the codebase?** Trace the execution path from entry point to output.
   - **Does it produce correct output when invoked?** Test with representative inputs.
   - **Are there edge cases or partial implementations?** Check boundary conditions.
   - **Does the implementation match what the documentation promises?**
   - **Are there bugs?** Look for logic errors, off-by-one errors, nil pointer dereferences, resource leaks, race conditions, and incorrect error handling. Bugs on critical goal paths are CRITICAL severity.

2. Use dependency-level analysis for systematic coverage:
   - Map import dependencies across all `.go` files.
   - Categorize by dependency level: Level 0 (no internal imports) → Level 1 (imports Level 0) → Level N.
   - Audit in ascending level order to establish baseline correctness before examining higher-level code.

3. Cross-reference with `go-stats-generator` metrics for risk indicators:
   - Functions with cyclomatic complexity >15 or length >50 lines are high-risk for bugs.
   - Packages with <70% doc coverage may have undocumented behavioral differences.
   - Check `.duplication` for copy-paste that may have introduced drift.

4. Run `go test -race ./...` and `go vet ./...` to confirm baseline health. When `go vet` or linters report warnings, read the comments surrounding the flagged code. If a comment explicitly acknowledges the warning (e.g., `//nolint:`, an explanatory comment justifying the pattern, or a TODO tracking a known issue), treat it as an acknowledged false positive — do not report it as a new finding.

### Phase 4: Report
Generate **`AUDIT.md`** in the repository root:

```markdown
# AUDIT — [date]

## Project Goals
[What the project claims to do, who it serves, and what it promises]

## Goal-Achievement Summary
| Goal | Status | Evidence |
|------|--------|----------|
| [Stated goal] | ✅ Achieved / ⚠️ Partial / ❌ Missing | [file:line or metric reference] |

## Findings
### CRITICAL
- [ ] [Finding title] — [file:line] — [description with evidence] — **Remediation:** [specific, actionable fix]
### HIGH / MEDIUM / LOW
- [ ] ...

## Metrics Snapshot
[Key numbers: total functions, avg complexity, doc coverage, duplication ratio]
```

Generate **`GAPS.md`** in the repository root:

```markdown
# Implementation Gaps — [date]

## [Gap Title]
- **Stated Goal**: [what the README/docs promise]
- **Current State**: [what actually exists]
- **Impact**: [how this gap affects users]
- **Closing the Gap**: [what needs to happen to achieve the stated goal]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | Feature documented but non-functional, data corruption risk, or confirmed bug on a critical path |
| HIGH | Feature partially implemented, or high complexity with no tests |
| MEDIUM | Edge case failures, or documentation coverage gap >20% |
| LOW | Minor inconsistencies that don't affect stated goals |

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Complete solutions**: State the full fix — what to change and where. Do not recommend "consider doing X."
2. **Respect project idioms**: Recommendations must follow the existing codebase's conventions.
3. **Verifiable**: Include a validation command (e.g., `go test -race ./pkg/...`, `go-stats-generator analyze . --format json | jq '.complexity'`).

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Use `go-stats-generator` metrics as primary evidence source.
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Evaluate the code against its **own stated goals**, not arbitrary external standards.

## Tiebreaker
Prioritize by severity (CRITICAL > HIGH > MEDIUM > LOW), then by impact on the project's core stated goals.
