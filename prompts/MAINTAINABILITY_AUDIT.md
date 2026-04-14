# TASK: Perform a code maintainability and technical debt audit of a Go project, evaluating code complexity, duplication, coupling, documentation debt, and long-term sustainability while rigorously preventing false positives.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the maintainability audit report
2. **`GAPS.md`** — technical debt items with prioritized remediation roadmap

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project's Maintenance Context
1. Read the project README to understand its purpose, maturity, team size, and contribution model.
2. Examine `go.mod` for module path, Go version, and dependency age.
3. List packages (`go list ./...`) and understand the project's architecture.
4. Assess the project's maintenance context:
   - **Team size**: Single developer, small team, large org, open source community?
   - **Change velocity**: How frequently is the code modified? (Recent commit history)
   - **Contributor diversity**: How many people contribute? (Bus factor)
   - **Maturity stage**: Active development, stable maintenance, legacy, or declining?
5. Identify areas of the code that change most frequently (these benefit most from maintainability improvements).
6. Note existing quality infrastructure: linters, CI checks, code review requirements, documentation standards.

### Phase 1: Online Research
Use web search to build context:
1. Search for the project on GitHub — read issues about code quality, refactoring requests, or difficulty contributing.
2. Review recent PRs for patterns: are changes concentrated in certain files? Are PRs large and complex or small and focused?
3. Check contributor guidelines for maintainability standards the project has adopted.

Keep research brief (≤10 minutes). Record only findings relevant to maintainability assessment.

### Phase 2: Baseline
```bash
set -o pipefail
go-stats-generator analyze . --skip-tests --format json --sections functions,documentation,packages,patterns,duplication,interfaces,structs > /tmp/maintain-audit-metrics.json
go-stats-generator analyze . --skip-tests
go vet ./... 2>&1 | tee /tmp/maintain-vet-results.txt
```
Delete temporary files when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Maintainability Audit

#### 3a. Complexity Hotspots
Use `go-stats-generator` function metrics to identify complexity debt:

- [ ] Functions with cyclomatic complexity >15 are identified and assessed — can they be decomposed without loss of clarity?
- [ ] Functions exceeding 50 lines are identified — long functions are harder to understand, test, and modify.
- [ ] Functions with >5 parameters are identified — many parameters suggest the function has too many responsibilities.
- [ ] Deeply nested code (>3 levels of indentation) is identified — flatten with early returns or extraction.
- [ ] Switch statements with >10 cases are identified — consider map-based dispatch or the strategy pattern.
- [ ] Files exceeding 500 lines are identified — large files indicate unclear package responsibility boundaries.

#### 3b. Code Duplication
Use `go-stats-generator` duplication metrics:

- [ ] Clone pairs are identified and assessed — are they true duplication or coincidentally similar code?
- [ ] Exact clones (identical code) are prioritized over near-clones (similar but parameterized differently).
- [ ] Duplication across packages indicates missing shared abstractions.
- [ ] Duplication within a package indicates extract-function opportunities.
- [ ] Copy-pasted error handling patterns indicate a missing error handling helper.
- [ ] Duplicated test setup indicates missing test helpers or fixtures.

#### 3c. Coupling and Cohesion
- [ ] Package dependencies are acyclic — circular dependencies make independent testing and modification impossible.
- [ ] Package fan-in (number of importers) and fan-out (number of imports) are balanced — a package imported by everything is a risk; a package importing everything has too many responsibilities.
- [ ] God packages (doing too many unrelated things) are identified — split by responsibility.
- [ ] Shotgun surgery risk: changing one feature requires modifying many packages (high coupling).
- [ ] Feature envy: functions in one package primarily operate on types from another package (misplaced responsibility).
- [ ] Interface segregation: large interfaces force implementors to implement methods they do not need.

#### 3d. Documentation Debt
Use `go-stats-generator` documentation metrics:

- [ ] Package-level documentation coverage is assessed — packages without doc comments are difficult for new contributors.
- [ ] Exported API documentation is complete — every exported type, function, and method has a GoDoc comment.
- [ ] Comments are accurate — stale comments that describe old behavior are worse than no comments.
- [ ] README accurately describes the current project (not an outdated version).
- [ ] Architecture decisions are documented somewhere (ADRs, design docs, or inline comments explaining "why").
- [ ] Onboarding documentation exists for new contributors (how to build, test, and contribute).

#### 3e. Code Freshness and Rot
- [ ] Go version in `go.mod` is current — more than 2 minor versions behind indicates potential maintenance debt.
- [ ] Deprecated standard library functions are not used (e.g., `ioutil` package deprecated since Go 1.16).
- [ ] Dependencies with security advisories are identified and updatable.
- [ ] `TODO` and `FIXME` comments are assessed — are they recent (active development) or ancient (forgotten debt)?
- [ ] Dead code (unreachable functions, unused types) is identified — dead code confuses readers and has maintenance cost.
- [ ] Build tags and platform-specific code are still relevant to the project's supported platforms.

#### 3f. Testability
- [ ] Functions are testable in isolation — they accept interfaces, not concrete types, for dependencies.
- [ ] Side effects (file I/O, network, time) are injectable, not hardcoded.
- [ ] Global state is minimal — package-level variables make tests non-independent.
- [ ] `init()` functions do not perform side effects that make testing difficult.
- [ ] Test coverage correlates with code complexity — high-complexity functions have high test coverage.

#### 3g. Change Safety
- [ ] High-churn files (frequently modified) have good test coverage — these are the files most likely to introduce regressions.
- [ ] The project has CI that runs tests on every change.
- [ ] Linters are configured and enforced — they catch common maintainability issues automatically.
- [ ] Type safety is leveraged — `interface{}` / `any` is used sparingly where type-safe alternatives exist.
- [ ] Error handling is consistent — developers can predict how errors flow without reading every function.

#### 3h. False-Positive Prevention (MANDATORY)
Before recording ANY finding, apply these checks:

1. **Assess the project's lifecycle stage**: A prototype or proof-of-concept has different maintainability expectations than a production system. Evaluate against the appropriate standard.
2. **Check for intentional complexity**: Some functions are inherently complex (parsers, state machines, protocol handlers). High complexity is justified if the problem domain requires it and the function is well-tested.
3. **Verify the debt is material**: A function with complexity 16 (threshold: 15) is not the same debt as complexity 45. Report proportionally.
4. **Respect the team's capacity**: Recommending a large refactoring to a single-maintainer project is not actionable. Prioritize high-value, low-effort improvements.
5. **Read surrounding context**: If code includes `// TODO: refactor when X is resolved`, it is a tracked item. Note it but classify as LOW.
6. **Check duplication intent**: Some duplication is intentional for clarity (e.g., similar but meaningfully different test cases). Do not blindly flag all duplication.

**Rule**: Maintainability is contextual. A function that is complex but well-tested, well-documented, and rarely modified is not urgent debt. Prioritize debt that actively hinders the team's ability to make changes safely.

### Phase 4: Report
Generate **`AUDIT.md`**:
```markdown
# MAINTAINABILITY AUDIT — [date]

## Project Maintenance Context
[Team size, maturity stage, change velocity, quality infrastructure]

## Maintainability Scorecard
| Category | Status | Key Metric |
|----------|--------|------------|
| Complexity | ✅/⚠️/❌ | N functions above threshold |
| Duplication | ✅/⚠️/❌ | N clone pairs, N% ratio |
| Coupling | ✅/⚠️/❌ | N circular deps, max fan-out: N |
| Documentation | ✅/⚠️/❌ | N% coverage |
| Code Freshness | ✅/⚠️/❌ | Go version, N stale TODOs |
| Testability | ✅/⚠️/❌ | N untestable functions |

## Complexity Hotspots
| Function | File:Line | Complexity | Lines | Params | Risk |
|----------|-----------|-----------|-------|--------|------|
| [func] | [file:line] | N | N | N | HIGH/MEDIUM |

## Findings
### CRITICAL
- [ ] [Finding] — [file:line] — [metric evidence] — [maintainability impact] — **Remediation:** [specific refactoring]
### HIGH / MEDIUM / LOW
- [ ] ...

## Technical Debt Inventory
| Debt Item | Category | Effort | Impact | Priority |
|-----------|----------|--------|--------|----------|
| [item] | complexity/duplication/coupling/docs | S/M/L | [description] | P1/P2/P3 |

## False Positives Considered and Rejected
| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| [description] | [why it is not actionable maintainability debt] |
```

Generate **`GAPS.md`**:
```markdown
# Technical Debt Gaps — [date]

## [Gap Title]
- **Current State**: [complexity metric, duplication count, or coupling issue]
- **Impact**: [how this debt affects development velocity, bug risk, or contributor onboarding]
- **Remediation**: [specific refactoring, extraction, or documentation to add]
- **Effort**: [small/medium/large — hours, not days, where possible]
- **Dependencies**: [what must be done first, if anything]
- **Quick Win**: [yes/no — can this be fixed in a single PR?]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | Circular dependencies that prevent testing, functions with complexity >30 on critical paths with no tests, or code so coupled that any change risks breaking unrelated features |
| HIGH | Functions with complexity >20, duplication >10% in actively maintained packages, or missing documentation on exported API used by external consumers |
| MEDIUM | Functions with complexity >15, minor duplication, stale TODOs older than 1 year, or deprecated stdlib usage |
| LOW | Cosmetic complexity (complexity 11-15 in well-tested code), minor documentation gaps, or single-use internal abstractions |

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Specific refactoring**: Describe the extraction, decomposition, or restructuring needed. Do not recommend "reduce complexity."
2. **Preserve behavior**: All refactoring must be behavior-preserving — no functionality changes disguised as maintainability improvements.
3. **Verifiable**: Include validation (e.g., `go test ./...`, `go-stats-generator analyze . | grep complexity`, `go vet ./...`).
4. **Incremental**: Break large refactorings into independent, shippable steps. Each step should leave the code in a working state.

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Use `go-stats-generator` metrics as the **primary evidence source** for all complexity, duplication, and documentation findings.
- Every finding must reference a specific file, line number, and metric value.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Evaluate maintainability against the project's **own development context** (team size, maturity, velocity).
- Apply the Phase 3h false-positive prevention checks to every candidate finding before including it.

## Tiebreaker
Prioritize: circular dependencies → high-complexity untested functions → actively-maintained duplicated code → stale documentation → cosmetic debt. Within a level, prioritize by change frequency (high-churn files benefit most from maintainability improvements).
