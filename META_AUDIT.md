# TASK: Audit exactly ONE unaudited Go sub-package per invocation — generate a package-level AUDIT.md and update the root audit tracker.

## Execution Mode
**Autonomous action** — create `<package>/AUDIT.md` and update root `AUDIT.md`.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 1: Select Package
1. List all Go packages: `go list ./...`
2. Check which packages already have an AUDIT.md: `find . -name AUDIT.md`
3. Select the first unaudited package, prioritizing:
   - Packages listed in root AUDIT.md but unchecked
   - Packages with highest integration surface (most imports/importers)
   - Core business logic packages over utility packages

### Phase 2: Analyze
```bash
go-stats-generator analyze ./<package> --skip-tests --format json --output pkg-audit.json --sections functions,documentation,naming,patterns,duplication,interfaces,structs,packages
go test -race -count=1 ./<package>/...
go vet ./<package>/...
```

### Phase 3: Audit
For the selected package, evaluate against these gates:

| Gate | Threshold | Check |
|------|-----------|-------|
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
- Remediation suggestion

### Phase 4: Report
Create `<package>/AUDIT.md`:
```markdown
# AUDIT: [package name] — [date]
## Summary
[Gate pass/fail counts, overall assessment]

## Findings
### [SEVERITY]
- [ ] [Finding] — [file:line] — [metric]: [value] (threshold: [target])
```

Update root `AUDIT.md` (create if absent):
```markdown
- [x] [package]: [pass_count]/[total_gates] gates passing — see [package]/AUDIT.md
```

## Thresholds
- Doc coverage: >=70%
- Cyclomatic complexity: <=10
- Function length: <=30 lines
- Test coverage: >=65%
- Duplication: <5%
- Naming violations: 0

## Output Format
```
Package: [name]
Gates: [passed]/[total] passing
Findings: [count] ([critical], [high], [medium], [low])
Created: [package]/AUDIT.md
Updated: AUDIT.md
```

## Tiebreaker
Audit the package with the highest integration surface (most importers) first. If tied, choose alphabetically.
