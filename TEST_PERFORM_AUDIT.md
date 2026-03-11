# TASK: Perform a comprehensive goal-focused functional audit of **test code** using `go-stats-generator` metrics as the primary evidence source.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the test audit report
2. **`GAPS.md`** — gaps in test coverage relative to the project's stated goals

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project's Goals and Test Strategy
1. Read the project README to understand its domain, stated goals, and expected behavior.
2. Discover the test framework in use and the project's assertion patterns.
3. Identify the test organization: how are tests structured, what testing conventions exist?
4. Note whether the project uses `t.Parallel()`, `t.Cleanup()`, test suites, or integration test separation.
5. Map which stated goals have corresponding test coverage and which do not.

### Phase 1: Online Research
Use web search to build context that isn't available in the repository:
1. Search for the project on GitHub — read open issues and discussions to understand known test gaps and flaky test reports.
2. Research the project's test dependencies for known issues, deprecations, or better alternatives.
3. Look up testing best practices in the project's domain (e.g., table-driven tests, test helpers, integration test patterns).

Keep research brief (≤10 minutes). Record only findings that are directly relevant to the project's test strategy and stated goals.

### Phase 2: Baseline
```bash
go-stats-generator analyze . --only-tests --format json --sections functions,documentation,patterns,duplication > /tmp/test-audit-metrics.json
go-stats-generator analyze . --only-tests
```
Delete `/tmp/test-audit-metrics.json` when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Goal-Focused Test Audit
1. For each stated project goal, verify that adequate test coverage exists:
   - Are the critical paths for this goal tested?
   - Do tests cover the happy path, error paths, and edge cases?
   - Are integration points between packages tested?

2. Extract test function metrics and classify by risk (tunable defaults):
   - **HIGH RISK**: length >75 lines OR cyclomatic >22 OR nesting >7
   - **MEDIUM RISK**: length >45 lines OR cyclomatic >15 OR nesting >5
   - **LOW RISK**: within test-appropriate thresholds

3. For each HIGH RISK test function, review:
   - Test setup error handling (are setup failures caught with `t.Fatal`?)
   - Race conditions in parallel tests
   - Resource cleanup (`t.Cleanup()`, temp files, goroutines)
   - Flaky test patterns (timing, hardcoded ports, file system)

4. Cross-reference with `.duplication` and `.documentation` for additional findings.

### Phase 4: Report

Generate **`AUDIT.md`**:
```markdown
# TEST AUDIT — [date]

## Project Goals and Test Coverage
| Stated Goal | Test Coverage | Assessment |
|-------------|--------------|------------|
| [Goal] | [which test files/functions cover this] | ✅ Well-tested / ⚠️ Partial / ❌ Untested |

## Test Infrastructure Context
[Test framework, conventions, coverage approach]

## Risk Summary
[HIGH: N, MEDIUM: N, critical findings: N]

## Findings
### CRITICAL
- [ ] [Finding] — [file:line] — [evidence] — [which goal's test coverage is affected] — [remediation]
### HIGH / MEDIUM / LOW
- [ ] ...
```

Generate **`GAPS.md`**:
```markdown
# Test Coverage Gaps — [date]

## [Gap Title]
- **Stated Goal**: [what the project claims to do]
- **Current Test State**: [what tests exist, if any]
- **Missing Coverage**: [what scenarios are untested]
- **Impact**: [risk of this goal regressing without tests]
- **Closing the Gap**: [specific tests to add]
```

## Risk Thresholds (test-appropriate — ~50% relaxed)
| Risk Level | Criteria |
|------------|----------|
| HIGH | length >75 OR cyclomatic >22 OR nesting >7 |
| MEDIUM | length >45 OR cyclomatic >15 OR nesting >5 |
| LOW | within thresholds |

## Constraints
- Output ONLY the two report files — no code changes.
- Use `go-stats-generator --only-tests` metrics as primary evidence.
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes.
- Evaluate test quality in terms of how well tests verify the project's **stated goals**, not arbitrary coverage targets.
- Recommend table-driven tests and `t.Helper()` extraction as primary remediation patterns.

## Tiebreaker
Prioritize: untested stated goals → HIGH RISK → MEDIUM RISK → LOW RISK. Within a level, highest complexity first.
