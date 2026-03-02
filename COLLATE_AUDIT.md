# TASK DESCRIPTION:
Perform a data-driven audit collation and remediation prioritization to discover all `*AUDIT*.md` files, extract every **unfinished** finding (unchecked `- [ ]` items), enrich each finding with `go-stats-generator` complexity metrics, and produce a single root-level `AUDIT.md` with metric-backed severity rankings and step-by-step remediation instructions for 100% resolution of all outstanding items.

## CONSTRAINT:

**Report generation only** — do NOT modify source code or existing audit files. Use only `go-stats-generator` and standard shell tools (`find`, `grep`, `jq`) for your analysis. You are absolutely forbidden from modifying any `.go` files, test files, or existing `*AUDIT*.md` files. The sole output artifact is a new root-level `AUDIT.md`.

## PREREQUISITES:
**Minimum Required Version:** `go-stats-generator` v1.0.0 or higher
Install and configure `go-stats-generator` for complexity-backed audit prioritization:

### Installation:
```bash
# First, check if go-stats-generator is already installed
which go-stats-generator
# If not, install it with `go install`
go install github.com/opd-ai/go-stats-generator@latest
```

## Recommendations:
```bash
# When long json outputs are encountered, use `jq`
go-stats-generator analyze --format json | jq .functions
# Check if it is installed
which jq
# If it is not, install it
sudo apt-get install jq
```

### Required Analysis Workflow:
```bash
# Phase 1: Generate complexity baseline for metric-backed prioritization
go-stats-generator analyze . --skip-tests --format json --output audit-metrics.json
go-stats-generator analyze . --skip-tests

# Phase 2: Discover and extract audit findings
find . -type f -iname '*audit*.md' | sort
# For each file, extract unchecked items
grep -rn '\- \[ \]' <file>

# Phase 3: Correlate findings with complexity metrics
cat audit-metrics.json | jq '.functions[] | select(.complexity > 9.0) | {name, file, complexity, lines, cyclomatic}'
cat audit-metrics.json | jq '.packages[] | {name, coupling, cohesion}'

# Phase 4: Produce metric-enriched AUDIT.md
# Collate, deduplicate, rank by metric-backed severity, generate remediation plan
```

## EXCLUSIONS:
The following categories of findings MUST be excluded from the collated remediation plan:

- **Test coverage percentage findings** — any finding reporting that test coverage is below a target threshold (e.g., "coverage below 65%")
- **Missing test findings** — findings that flag missing unit tests, table-driven tests, benchmarks, or test files
- **Test infrastructure findings** — findings about missing test helpers, test fixtures, or test utilities
- **Coverage tooling findings** — findings related to `go test -cover`, coverage reports, or coverage configuration

Exclusion applies at all stages: extraction, deduplication, and remediation generation. Any `- [ ]` item whose category is `testing` or `test-coverage`, or whose description relates to test coverage, missing tests, or coverage percentages MUST be silently omitted and not assigned a `REM-###` tracking ID.

## CONTEXT:
You are an automated Go code auditor using `go-stats-generator` for enterprise-grade complexity analysis and audit remediation prioritization. The tool provides precise function-level, package-level, and structural metrics that transform subjective audit findings into quantitatively ranked remediation items. By correlating audit findings with actual complexity data, you ensure that the highest-risk code receives attention first and that remediation effort is proportional to measured impact.

## INSTRUCTIONS:

### Phase 1: Complexity Baseline Generation
1. **Run Full Analysis:**
  ```bash
  go-stats-generator analyze . --skip-tests --format json --output audit-metrics.json
  ```
  - Capture function complexity, line counts, cyclomatic scores, and nesting depths
  - Capture package coupling, cohesion, and dependency metrics
  - Capture documentation coverage and naming convention data

2. **Extract Key Metric Summaries:**
  ```bash
  # Functions exceeding complexity thresholds
  cat audit-metrics.json | jq '[.functions[] | select(.complexity > 9.0)] | sort_by(-.complexity) | .[:20] | .[] | {name, file, complexity, lines, cyclomatic}'

  # Packages with high coupling or low cohesion
  cat audit-metrics.json | jq '.packages[] | select(.coupling > 5 or .cohesion < 0.5) | {name, coupling, cohesion}'

  # Documentation coverage gaps
  cat audit-metrics.json | jq '.documentation'
  ```

### Phase 2: Audit Finding Discovery and Extraction
1. **Discover All Audit Files:**
  ```bash
  find . -type f -iname '*audit*.md' | sort
  ```

2. **Extract Unfinished Items:**
  For each discovered file, extract lines matching unchecked findings:
  ```bash
  grep -n '\- \[ \]' <file>
  ```
  Collect each finding with:
  - **Source file path** (e.g., `internal/AUDIT.md`)
  - **Severity** (critical / high / medium / low) — as originally marked in the audit
  - **Category** (error-handling, documentation, security, api-design, concurrency, naming, etc.)
  - **Description and code location** (e.g., `main.go:71`)

3. **Apply Exclusion Filters:**
  After extraction, discard any finding that matches the test-coverage exclusions defined in the EXCLUSIONS section above. Specifically, drop findings where:
  - The category is `testing` or `test-coverage`
  - The description mentions test coverage percentages, missing tests, missing benchmarks, or coverage targets
  - The finding originates from a `## Test Coverage` section of a sub-package audit

### Phase 3: Metric-Backed Severity Correlation
1. **Correlate Findings with Complexity Data:**
  For each remaining finding, look up the referenced file and function in `audit-metrics.json`:
  ```bash
  # Look up the function referenced by a finding
  cat audit-metrics.json | jq --arg file "path/to/file.go" --arg func "functionName" \
    '.functions[] | select(.file == $file and .name == $func) | {name, file, complexity, lines, cyclomatic, nesting_depth}'
  ```

2. **Adjust Severity Using Metrics:**
  Use `go-stats-generator` metrics to validate or escalate the original audit severity:
  - **Escalate to CRITICAL** if the finding's code location has:
    * Overall complexity > 20.0, OR
    * Cyclomatic complexity > 15, OR
    * Function length > 60 lines
  - **Escalate to HIGH** if the finding's code location has:
    * Overall complexity > 15.0, OR
    * Cyclomatic complexity > 10, OR
    * Function length > 40 lines
  - **Confirm as MEDIUM** if the finding's code location has:
    * Overall complexity > 9.0, OR
    * Cyclomatic complexity > 7, OR
    * Function length > 30 lines
  - **Confirm or downgrade to LOW** if metrics are below all thresholds

  Severity can only be **escalated** by metrics, never downgraded below the original audit severity. If `go-stats-generator` metrics indicate a higher risk than the audit originally assigned, use the higher severity.

3. **Enrich Finding Metadata:**
  Attach metric data to each finding for the output:
  - Complexity score of the affected function
  - Line count of the affected function
  - Package coupling/cohesion if the finding is structural

### Phase 4: Collation and Deduplication
1. **Group by Metric-Backed Severity:**
  - CRITICAL → HIGH → MEDIUM → LOW
  - Within each severity, group by category

2. **Deduplicate Across Sub-Audits:**
  - Merge findings that reference the same root cause across multiple `*AUDIT*.md` files
  - When merging, keep the **highest severity** (metric-backed) and combine all affected locations
  - Prefer the most specific description

3. **Assign Tracking IDs:**
  - Assign sequential `REM-###` IDs starting from `REM-001`
  - Order: CRITICAL first (by descending complexity), then HIGH, MEDIUM, LOW

### Phase 5: Remediation Plan Generation
For each finding, write:
1. **Tracking ID** and original audit source(s)
2. **One-sentence problem statement**
3. **Affected file(s) and line(s)**
4. **Complexity metrics** from `go-stats-generator` (complexity score, cyclomatic, lines)
5. **Step-by-step fix instructions** — concrete, minimal, copy-paste-ready where possible
6. **Verification command** — including a `go-stats-generator` command to confirm the fix reduced complexity:
  ```bash
  go-stats-generator analyze <file> --format json | jq '.functions[] | select(.name == "<function>") | {complexity, lines, cyclomatic}'
  ```

Fixes must be:
- **Simple** — prefer standard library, smallest diff
- **Maintainable** — no clever tricks, follow existing code style
- **Minimally invasive** — change only what the finding requires
- **Metric-validated** — include a `go-stats-generator` verification step

## OUTPUT FORMAT:

The final `AUDIT.md` must use this structure:

```markdown
# Codebase Audit Remediation Plan
**Generated**: YYYY-MM-DD
**Scope**: All *AUDIT*.md files in repository
**Analysis Engine**: go-stats-generator v1.x.x
**Total Unresolved Findings**: <N>

## Complexity Baseline
| Metric                     | Value |
|----------------------------|-------|
| Functions Above Threshold  | X     |
| Avg Complexity             | X.XX  |
| Max Complexity             | X.XX  |
| Packages with High Coupling| X     |

## Summary by Severity
| Severity | Count | Avg Complexity |
|----------|-------|----------------|
| Critical | X     | X.XX           |
| High     | X     | X.XX           |
| Medium   | X     | X.XX           |
| Low      | X     | X.XX           |

## Findings

### CRITICAL

#### REM-001: <title>
- **Source**: `path/to/AUDIT.md`
- **Location**: `file.go:LINE`
- **Complexity**: X.XX (cyclomatic: X, lines: X)
- **Problem**: <one sentence>
- **Fix**:
  1. <step>
  2. <step>
- **Verify**:
  ```bash
  go-stats-generator analyze <file> --format json | jq '.functions[] | select(.name == "<func>") | {complexity, lines, cyclomatic}'
  ```

### HIGH
...

### MEDIUM
...

### LOW
...

## Completion Criteria
- [ ] All REM-### items implemented
- [ ] All verification commands pass
- [ ] `go-stats-generator analyze . --skip-tests` shows no regressions
- [ ] No `- [ ]` items remain in any *AUDIT*.md
```

## SEVERITY THRESHOLDS (go-stats-generator correlation):
```
Escalation to CRITICAL:
  Overall Complexity > 20.0 OR Cyclomatic > 15 OR Lines > 60

Escalation to HIGH:
  Overall Complexity > 15.0 OR Cyclomatic > 10 OR Lines > 40

Confirmation as MEDIUM:
  Overall Complexity > 9.0 OR Cyclomatic > 7 OR Lines > 30

Confirmation as LOW:
  All metrics below MEDIUM thresholds

Rule: Metrics can only ESCALATE severity, never downgrade below original audit severity.
```

```
Overall Complexity = cyclomatic + (nesting_depth * 0.5) + (cognitive * 0.3)
Signature Complexity = (params * 0.5) + (returns * 0.3) + (interfaces * 0.8) + (generics * 1.5) + variadic_penalty
Refactoring Threshold = Overall Complexity > 9.0 OR Lines > 40 OR Cyclomatic > 9
```
<!-- Last verified: 2025-07-25 against function.go:calculateComplexity and calculateSignatureComplexity -->

## EXAMPLE WORKFLOW:
```bash
$ # Phase 1: Generate complexity baseline
$ go-stats-generator analyze . --skip-tests --format json --output audit-metrics.json
$ go-stats-generator analyze . --skip-tests
=== FUNCTION ANALYSIS ===
Functions analyzed: 142
Functions above threshold: 8
Average complexity: 6.4
Max complexity: 25.4 (processComplexOrder in order.go)

=== PACKAGE ANALYSIS ===
Packages analyzed: 12
High coupling (>5): 2
Low cohesion (<0.5): 1

$ # Phase 2: Discover audit files and extract findings
$ find . -type f -iname '*audit*.md' | sort
./AUDIT.md
./cmd/AUDIT.md
./internal/analyzer/AUDIT.md
./internal/reporter/AUDIT.md
./pkg/metrics/AUDIT.md

$ grep -c '\- \[ \]' ./internal/analyzer/AUDIT.md
14

$ # Phase 3: Correlate findings with metrics
$ cat audit-metrics.json | jq --arg file "internal/analyzer/function.go" \
    '[.functions[] | select(.file == $file and .complexity > 9.0)] | sort_by(-.complexity) | .[] | {name, complexity, lines, cyclomatic}'
{
  "name": "analyzeFunction",
  "complexity": 22.1,
  "lines": 58,
  "cyclomatic": 16
}
{
  "name": "calculateMetrics",
  "complexity": 15.7,
  "lines": 43,
  "cyclomatic": 11
}

$ # Finding: "error handling missing in analyzeFunction" (originally HIGH)
$ # Metrics show complexity 22.1 > 20.0 → ESCALATE to CRITICAL

$ # Phase 4: Collate, deduplicate, rank
$ # 3 sub-audits flagged analyzeFunction error handling → merge into REM-001

$ # Phase 5: Generate AUDIT.md
$ cat AUDIT.md
# Codebase Audit Remediation Plan
**Generated**: 2025-07-25
**Scope**: All *AUDIT*.md files in repository
**Analysis Engine**: go-stats-generator v1.0.0
**Total Unresolved Findings**: 23

## Complexity Baseline
| Metric                     | Value |
|----------------------------|-------|
| Functions Above Threshold  | 8     |
| Avg Complexity             | 6.40  |
| Max Complexity             | 25.40 |
| Packages with High Coupling| 2     |

## Summary by Severity
| Severity | Count | Avg Complexity |
|----------|-------|----------------|
| Critical | 4     | 21.30          |
| High     | 7     | 14.20          |
| Medium   | 8     | 8.50           |
| Low      | 4     | 4.10           |

## Findings

### CRITICAL

#### REM-001: Missing error handling in analyzeFunction
- **Source**: `internal/analyzer/AUDIT.md`, `pkg/metrics/AUDIT.md`
- **Location**: `internal/analyzer/function.go:45`
- **Complexity**: 22.10 (cyclomatic: 16, lines: 58)
- **Problem**: Error return from parseExpression is silently discarded, risking nil pointer dereference in high-complexity function.
- **Fix**:
  1. Add error check after parseExpression call at line 45
  2. Return wrapped error: `return fmt.Errorf("analyzeFunction: %w", err)`
  3. Add nil guard for the parsed result before use at line 48
- **Verify**:
  ```bash
  go vet ./internal/analyzer/...
  go-stats-generator analyze internal/analyzer/function.go --format json | jq '.functions[] | select(.name == "analyzeFunction") | {complexity, lines, cyclomatic}'
  ```

### HIGH

#### REM-005: Unchecked type assertion in calculateMetrics
- **Source**: `internal/analyzer/AUDIT.md`
- **Location**: `internal/analyzer/metrics.go:112`
- **Complexity**: 15.70 (cyclomatic: 11, lines: 43)
- **Problem**: Bare type assertion without comma-ok pattern will panic on unexpected input types.
- **Fix**:
  1. Replace `val := node.(ast.Expr)` with `val, ok := node.(ast.Expr)`
  2. Add guard: `if !ok { return fmt.Errorf("unexpected node type: %T", node) }`
- **Verify**:
  ```bash
  go vet ./internal/analyzer/...
  go-stats-generator analyze internal/analyzer/metrics.go --format json | jq '.functions[] | select(.name == "calculateMetrics") | {complexity, lines, cyclomatic}'
  ```

### MEDIUM
...

### LOW
...

## Completion Criteria
- [ ] All REM-### items implemented
- [ ] All verification commands pass
- [ ] `go-stats-generator analyze . --skip-tests` shows no regressions
- [ ] No `- [ ]` items remain in any *AUDIT*.md
```

## SUCCESS CRITERIA:
- Every `- [ ]` finding from every `*AUDIT*.md` file — **except** test-coverage related findings (see EXCLUSIONS) — has a corresponding `REM-###` entry
- Test-coverage findings are explicitly excluded and do not appear in the output
- Zero non-excluded findings omitted, skipped, or deferred
- Each finding includes `go-stats-generator` complexity metrics for the affected code location
- Severity is validated or escalated using metric-backed thresholds — never downgraded below original audit severity
- Each remediation includes a `go-stats-generator` verification command
- Each remediation is actionable without additional research
- The output is a single valid Markdown file at `./AUDIT.md`

This data-driven approach ensures audit remediation priorities are based on quantitative complexity analysis rather than subjective assessment, with `go-stats-generator` metrics providing objective evidence for severity ranking and remediation validation.
