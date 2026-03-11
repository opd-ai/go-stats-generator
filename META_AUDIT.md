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

### Phase 1: Select Package
1. Discover which packages already have audit files: `find . -name 'AUDIT.md'`
2. Select the first unaudited package, prioritizing:
   - Packages listed in any root-level audit tracker but unchecked
   - Packages that implement the project's core stated goals
   - Packages with highest integration surface (most imports/importers)

### Phase 2: Analyze
```bash
go-stats-generator analyze ./<package> --skip-tests --format json --sections functions,documentation,patterns,duplication,interfaces,structs,packages
go test -race -count=1 ./<package>/...
go vet ./<package>/...
```

### Phase 3: Goal-Focused Audit
Evaluate the selected package against its role in achieving the project's stated goals:

1. **Role clarity**: Does this package have a clear, well-defined responsibility? Does it serve one of the project's stated goals?
2. **Functional correctness**: Do the package's exported functions do what their documentation (and the project README) claims?
3. **Implementation completeness**: Are there stubs, TODOs, or partial implementations that prevent goal achievement?

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

### Phase 4: Report
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
