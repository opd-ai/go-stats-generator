# TASK: Audit exactly ONE unaudited Go sub-package per invocation — evaluate how well it fulfills its role in achieving the project's stated goals, and update the root audit tracker.

## Execution Mode
**Autonomous action** — create package-level audit and gaps files, and update the root audit tracker.

## Output
Write exactly two files in the audited package directory:
1. **`<package>/AUDIT.md`** — the package audit report
2. **`<package>/GAPS.md`** — gaps between the package's role and its implementation

If either file already exists, delete it and create a fresh one.

Also update the root audit tracker **`AUDIT_TRACKER.md`** in the repository root (create if absent).

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project's Goals
1. Read the project README to understand its purpose, goals, and architecture.
2. List all Go packages: `go list ./...`
3. Identify which packages serve which stated goals — core packages that implement key features deserve deeper scrutiny than utility packages.

### Phase 1: Online Research
Use web search to build context that isn't available in the repository:
1. Search for the project on GitHub — read open issues and discussions for known pain points in the package area you will audit.
2. Research key dependencies used by the package for known vulnerabilities or deprecations.
3. Look up best practices relevant to the package's domain (e.g., concurrency patterns, API design, parsing).

Keep research brief (≤10 minutes). Record only findings relevant to the package's role in the project's stated goals.

### Phase 2: Select Package
1. Discover which packages already have audit files: `find . -name 'AUDIT.md'`
2. Select the first unaudited package, prioritizing:
   - Packages listed in any root-level audit tracker but unchecked
   - Packages that implement the project's core stated goals
   - Packages with highest integration surface (most imports/importers)

### Phase 3: Analyze
```bash
go-stats-generator analyze ./<package> --skip-tests --format json --sections functions,documentation,patterns,duplication,interfaces,structs,packages
go test -race -count=1 ./<package>/...
go vet ./<package>/...
# When go vet or linters report warnings, read the comments surrounding the flagged code.
# If a comment explicitly acknowledges the warning (e.g., //nolint:, an explanatory comment
# justifying the pattern, or a TODO tracking a known issue), treat it as an acknowledged
# false positive — do not report it as a new finding.
```

### Phase 4: Goal-Focused Audit
Evaluate the selected package against its role in achieving the project's stated goals:

1. **Role clarity**: Does this package have a clear, well-defined responsibility? Does it serve one of the project's stated goals?
2. **Functional correctness**: Do the package's exported functions do what their documentation (and the project README) claims?
3. **Implementation completeness**: Are there stubs, TODOs, or partial implementations that prevent goal achievement?
4. **Bug detection**: Look for logic errors, nil dereferences, resource leaks, race conditions, and incorrect error handling. Run `go vet` and inspect high-complexity functions manually.

Also evaluate these quality gates (thresholds are tunable defaults — adjust if the project's conventions warrant):

| Gate | Default Threshold | Check |
|------|-------------------|-------|
| Documentation | >=70% coverage | `.documentation` |
| Complexity | All functions cyclomatic <=10 | `.functions` |
| Function length | All functions <=30 lines | `.functions` |
| Test coverage | >=65% | `go test -cover` |
| Duplication | <5% internal ratio | `.duplication` |

For each finding, create an entry with:
- Severity (CRITICAL/HIGH/MEDIUM/LOW)
- Specific file:line reference
- Metric value vs. threshold
- How this finding impacts the project's stated goals
- Remediation that respects the project's idioms

### Phase 5: Report
Create **`<package>/AUDIT.md`**:
```markdown
# AUDIT: [package name] — [date]

## Package Role
[What this package does in the context of the project's stated goals]

## Goal-Achievement Summary
| Project Goal | Package Contribution | Status |
|-------------|---------------------|--------|
| [Goal] | [How this package helps] | ✅ / ⚠️ / ❌ |

## Findings
### [SEVERITY]
- [ ] [Finding] — [file:line] — [metric]: [value] (threshold: [target]) — [impact on goals]
```

Create **`<package>/GAPS.md`**:
```markdown
# Implementation Gaps: [package name] — [date]

## [Gap Title]
- **Package Role**: [what this package should contribute to the project's goals]
- **Current State**: [what's actually implemented]
- **Impact**: [how this gap affects the project's stated goals]
- **Closing the Gap**: [what needs to happen]
```

Update root audit tracker **`AUDIT_TRACKER.md`** (create if absent):
```markdown
- [x] [package]: [pass_count]/[total_gates] gates passing — see [package]/AUDIT.md
```

## Output Format
```
Package: [name]
Gates: [passed]/[total] passing
Findings: [count] ([critical], [high], [medium], [low])
Created: [package]/AUDIT.md, [package]/GAPS.md
Updated: AUDIT_TRACKER.md
```

## Tiebreaker
Audit the package that contributes most to the project's core stated goals first. If tied, choose the one with the highest integration surface (most importers). If still tied, choose alphabetically.
