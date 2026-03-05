# TASK: Audit exactly ONE unaudited Go sub-package per invocation — generate a package-level audit report and update the root audit tracker.

## Execution Mode
**Autonomous action** — create `<package>/AUDIT.md` and update the root audit tracker.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project
1. Read the project README to understand its purpose and architecture.
2. List all Go packages: `go list ./...`
3. Identify which packages are core logic vs. utilities vs. infrastructure — core packages deserve deeper scrutiny.

### Phase 1: Select Package
1. Discover which packages already have audit files: `find . -name 'AUDIT.md'`
2. Select the first unaudited package, prioritizing:
   - Packages listed in any root-level audit tracker but unchecked
   - Packages with highest integration surface (most imports/importers)
   - Core business logic packages over utility packages

### Phase 2: Analyze
```bash
go-stats-generator analyze ./<package> --skip-tests --format json --output pkg-audit.json --sections functions,documentation,naming,patterns,duplication,interfaces,structs,packages
go test -race -count=1 ./<package>/...
go vet ./<package>/...
```

### Phase 3: Audit
Evaluate the selected package against these gates (thresholds are tunable defaults — adjust if the project's conventions warrant):

| Gate | Default Threshold | Check |
|------|-------------------|-------|
| Documentation | >=70% coverage | `.documentation` |
| Complexity | All functions cyclomatic <=10 | `.functions` |
| Function length | All functions <=30 lines | `.functions` |
| Test coverage | >=65% | `go test -cover` |
| Duplication | <5% internal ratio | `.duplication` |
| Naming | 0 violations | `.naming` |

For each gate failure, create a finding with:
- Severity (CRITICAL/HIGH/MEDIUM/LOW)
- Specific file:line reference
- Metric value vs. threshold
- Remediation suggestion that respects the project's idioms

### Phase 4: Report
Create `<package>/AUDIT.md`:
```markdown
# AUDIT: [package name] — [date]
## Package Role
[What this package does in the context of the project]
## Summary
[Gate pass/fail counts, overall assessment]
## Findings
### [SEVERITY]
- [ ] [Finding] — [file:line] — [metric]: [value] (threshold: [target])
```

Update root audit tracker (create if absent):
```markdown
- [x] [package]: [pass_count]/[total_gates] gates passing — see [package]/AUDIT.md
```

## Output Format
```
Package: [name]
Gates: [passed]/[total] passing
Findings: [count] ([critical], [high], [medium], [low])
Created: [package]/AUDIT.md
Updated: root audit tracker
```

## Tiebreaker
Audit the package with the highest integration surface (most importers) first. If tied, choose alphabetically.
