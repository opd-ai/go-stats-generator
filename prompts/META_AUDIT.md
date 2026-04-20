# TASK: Audit ALL unaudited Go sub-packages in a single invocation — evaluate how well each fulfills its role in achieving the project's stated goals, and update the root audit tracker.

## Execution Mode
**Autonomous action** — create package-level audit and gaps files for every unaudited package, and update the root audit tracker.

## Output
For each unaudited package, write exactly two files in that package's directory:
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

### Phase 2: Select Packages
1. Discover which packages already have audit files: `find . -name 'AUDIT.md'`
2. Build the full list of unaudited packages by comparing `go list ./...` output against the packages that already have `AUDIT.md` files.
3. Order the unaudited packages for processing, prioritizing:
   - Packages listed in any root-level audit tracker but unchecked
   - Packages that implement the project's core stated goals
   - Packages with highest integration surface (most imports/importers)
   - Alphabetically as a final tiebreaker
4. Audit **every** package in this ordered list — do not stop after the first one.

### Phase 3: Analyze
Repeat the following for each package in the ordered list from Phase 2:
```bash
go-stats-generator analyze ./<package> --skip-tests --format json --sections functions,documentation,patterns,duplication,interfaces,structs,packages
go test -race -count=1 ./<package>/...
go vet ./<package>/...
```

When `go vet` or linters report warnings, read the comments surrounding the flagged code. If a comment explicitly acknowledges the warning (e.g., `//nolint:`, an explanatory comment justifying the pattern, or a TODO tracking a known issue), treat it as an acknowledged false positive — do not report it as a new finding.

### Phase 4: Goal-Focused Audit
For **each** package in the ordered list, evaluate it against its role in achieving the project's stated goals:

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
For **each** audited package, create **`<package>/AUDIT.md`** and **`<package>/GAPS.md`** using the templates below, then append a line to **`AUDIT_TRACKER.md`** for that package.

After all packages have been processed, write the final state of **`AUDIT_TRACKER.md`** to disk once (create if absent).

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
Print one summary block per audited package:
```
Package: [name]
Gates: [passed]/[total] passing
Findings: [count] ([critical], [high], [medium], [low])
Created: [package]/AUDIT.md, [package]/GAPS.md
Updated: AUDIT_TRACKER.md
```
After all packages are processed, print a final totals line:
```
Total packages audited: [N]
Total findings: [count] ([critical], [high], [medium], [low])
```

## Package Processing Order
Process packages in the following priority order (descending):
1. Packages listed in any root-level audit tracker but unchecked.
2. Packages that contribute most to the project's core stated goals.
3. Packages with the highest integration surface (most importers).
4. Alphabetically.
