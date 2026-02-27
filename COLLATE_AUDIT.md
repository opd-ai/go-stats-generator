# Codebase-Wide Audit Remediation Plan

## Objective
Discover all `*AUDIT*.md` files across the codebase, extract every **unfinished** finding (unchecked `- [ ]` items), and produce a single root-level `AUDIT.md` that provides step-by-step remediation instructions for 100% resolution of all outstanding items.

## Execution Mode
**Report generation only** — do NOT modify source code or existing audit files.

## Exclusions
The following categories of findings MUST be excluded from the collated remediation plan:

- **Test coverage percentage findings** — any finding reporting that test coverage is below a target threshold (e.g., "coverage below 65%")
- **Missing test findings** — findings that flag missing unit tests, table-driven tests, benchmarks, or test files
- **Test infrastructure findings** — findings about missing test helpers, test fixtures, or test utilities
- **Coverage tooling findings** — findings related to `go test -cover`, coverage reports, or coverage configuration

Exclusion applies at all stages: extraction, deduplication, and remediation generation. Any `- [ ]` item whose category is `testing` or `test-coverage`, or whose description relates to test coverage, missing tests, or coverage percentages MUST be silently omitted and not assigned a `REM-###` tracking ID.

## Workflow

### Step 1: Discovery
Run file discovery to locate all audit-related markdown files:
```bash
find . -type f -iname '*audit*.md' | sort
```

### Step 2: Extract Unfinished Items
For each discovered file, extract lines matching unchecked findings:
```bash
grep -n '\- \[ \]' <file>
```
Collect each finding with:
- **Source file path** (e.g., `net/AUDIT.md`)
- **Severity** (critical / high / med / low)
- **Category** (error-handling, testing, documentation, security, api-design, etc.)
- **Description and code location** (e.g., `main.go:71`)

**Filtering**: After extraction, discard any finding that matches the test-coverage exclusions defined in the Exclusions section above. Specifically, drop findings where:
- The category is `testing` or `test-coverage`
- The description mentions test coverage percentages, missing tests, missing benchmarks, or coverage targets
- The finding originates from a `## Test Coverage` section of a sub-package audit

### Step 3: Collate and Deduplicate
- Group findings by severity (CRITICAL → HIGH → MEDIUM → LOW)
- Within each severity, group by category
- Merge duplicates that reference the same root cause across sub-audits
- Assign a unique tracking ID to each (e.g., `REM-001`)

### Step 4: Generate Remediation Instructions
For each finding, write:
1. **Tracking ID** and original audit source
2. **One-sentence problem statement**
3. **Affected file(s) and line(s)**
4. **Step-by-step fix instructions** — concrete, minimal, copy-paste-ready where possible
5. **Verification command** (e.g., `go test -race ./pkg/...`, `grep -r "fmt.Fprintf"`)

Fixes must be:
- **Simple** — prefer standard library, smallest diff
- **Maintainable** — no clever tricks, follow existing code style
- **Minimally invasive** — change only what the finding requires

### Step 5: Produce Root `AUDIT.md`

## Output Format

The final `AUDIT.md` must use this structure:

```markdown
# Codebase Audit Remediation Plan
**Generated**: YYYY-MM-DD
**Scope**: All *AUDIT*.md files in repository
**Total Unresolved Findings**: <N>

## Summary by Severity
| Severity | Count |
|----------|-------|
| Critical | X     |
| High     | X     |
| Medium   | X     |
| Low      | X     |

## Findings

### CRITICAL

#### REM-001: <title>
- **Source**: `path/to/AUDIT.md`
- **Location**: `file.go:LINE`
- **Problem**: <one sentence>
- **Fix**:
  1. <step>
  2. <step>
- **Verify**: `<command>`

### HIGH
...

### MEDIUM
...

### LOW
...

## Completion Criteria
- [ ] All REM-### items implemented
- [ ] All verification commands pass
- [ ] No `- [ ]` items remain in any *AUDIT*.md
```

## Success Criteria
- Every `- [ ]` finding from every `*AUDIT*.md` file — **except** test-coverage related findings (see Exclusions) — has a corresponding `REM-###` entry
- Test-coverage findings are explicitly excluded and do not appear in the output
- Zero non-excluded findings omitted, skipped, or deferred
- Each remediation is actionable without additional research
- The output is a single valid Markdown file at `./AUDIT.md`
