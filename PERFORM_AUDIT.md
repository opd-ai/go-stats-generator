# TASK: Perform a comprehensive risk-prioritized functional audit of the Go codebase using `go-stats-generator` metrics as the primary evidence source.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Codebase
Before auditing, learn what you're auditing:
1. Read the project README to understand its purpose, users, and claimed capabilities.
2. Examine `go.mod` for module path, Go version, and dependency profile.
3. List packages (`go list ./...`) and identify the architecture: core logic, API surface, infrastructure, and utilities.
4. Discover the project's error handling conventions, test strategy, and existing quality gates.
5. Note which packages are on critical paths (handle user input, manage state, perform I/O).

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output audit-metrics.json --sections functions,documentation,naming,patterns,duplication,interfaces,structs,packages
go-stats-generator analyze . --skip-tests
```

### Phase 2: Risk-Prioritized Audit
1. Extract function metrics and classify by risk (thresholds are tunable defaults):
   - **HIGH RISK**: length >50 lines OR cyclomatic >15 OR params >7
   - **MEDIUM RISK**: length >30 lines OR cyclomatic >10 OR params >5
   - **LOW RISK**: all metrics within thresholds
2. For each HIGH RISK function, perform detailed code review, focusing on:
   - Error handling completeness (does it match the project's conventions?)
   - Nil pointer risks and bounds checking
   - Concurrency safety (cross-reference `.patterns.concurrency_patterns`)
   - Whether the function's complexity is justified by its role (parsers and state machines may warrant higher thresholds)
3. For MEDIUM RISK functions on critical paths (discovered in Phase 0), perform targeted review.
4. Cross-reference with `.duplication.clone_pairs`, `.naming`, and `.documentation` for additional findings.
5. Run `go test -race ./...` and `go vet ./...` for baseline health.

### Phase 3: Report
Generate an audit document with findings organized by risk level, each with:
- Severity classification (CRITICAL/HIGH/MEDIUM/LOW)
- File:line reference
- Metric evidence (complexity, length, coverage)
- Remediation recommendation that respects the project's patterns

## Risk Thresholds (tunable defaults)
| Risk Level | Criteria |
|------------|----------|
| HIGH | length >50 OR cyclomatic >15 OR params >7 |
| MEDIUM | length >30 OR cyclomatic >10 OR params >5 |
| LOW | within all thresholds |

## Report Structure
```markdown
# AUDIT — [date]
## Project Context
[Project type, critical paths, conventions discovered]
## Risk Summary
[HIGH: N functions, MEDIUM: N functions, critical findings: N]
## Findings
### CRITICAL
- [ ] [Finding] — [file:line] — [evidence] — [remediation]
### HIGH / MEDIUM / LOW
- [ ] ...
```

## Constraints
- Output ONLY the audit report — no code changes.
- Use `go-stats-generator` metrics as primary evidence for all findings.
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes.
- Distinguish production code issues from test-only issues.

## Tiebreaker
Prioritize: HIGH RISK → critical-path functions → MEDIUM RISK → LOW RISK. Within a level, highest complexity first.
