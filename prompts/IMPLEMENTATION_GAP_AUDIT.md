# TASK: Perform an implementation gap discovery audit of a Go project, systematically identifying stubs, TODOs, incomplete code paths, dead code, unreachable features, and partially wired components that indicate unfinished work.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the implementation gap audit report
2. **`GAPS.md`** — detailed gap analysis with implementation roadmap

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Map the Intended Architecture
1. Read the project README to understand its purpose, stated goals, and architectural vision.
2. Read design documents, ADRs, ROADMAP.md, PLAN.md, or any planning files that describe intended but possibly unimplemented features.
3. Examine `go.mod` for the module path, Go version, and dependencies. Note any dependency that is imported but potentially unused.
4. List all packages (`go list ./...`) and build a package dependency graph. Identify packages that exist structurally but may lack substantive implementation.
5. Read exported interfaces and types across all packages to understand the intended API surface.
6. Identify the project's stated architecture — which packages own which responsibilities.

### Phase 1: Online Research
Use web search to build context that isn't available in the repository:
1. Search for the project on GitHub — read open issues, milestones, and project boards for planned features and known incomplete areas.
2. Read recent PRs for patterns of incremental implementation that may have left gaps.
3. Check if a roadmap or feature tracker exists outside the repository.

Keep research brief (≤10 minutes). Record only findings that identify expected but unimplemented functionality.

### Phase 2: Baseline
```bash
set -o pipefail
go-stats-generator analyze . --skip-tests --format json --sections functions,documentation,packages,patterns,interfaces,structs,duplication > tmp/gap-audit-metrics.json
go-stats-generator analyze . --skip-tests
go build ./... 2>&1 | tee tmp/gap-build-results.txt
go vet ./... 2>&1 | tee tmp/gap-vet-results.txt
```
Delete temporary files when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Implementation Gap Discovery

#### 3a. Stub and TODO Detection
Scan the entire codebase for indicators of incomplete implementation:

- [ ] `TODO`, `FIXME`, `HACK`, `XXX`, `TEMP`, `STUB` comments — classify each by what remains to be done and its severity.
- [ ] Functions that contain only `panic("not implemented")`, `panic("TODO")`, or `return nil` / `return nil, nil` as placeholders.
- [ ] Functions whose body is a single `return` with zero-value results and no logic.
- [ ] Exported functions with no tests — these may be stubs awaiting real implementation.
- [ ] Empty or near-empty source files that were scaffolded but never filled in.
- [ ] Interfaces defined but never implemented (or only implemented by mock types).
- [ ] Struct types defined but never instantiated or used outside of tests.

#### 3b. Dead and Unreachable Code
Identify code that exists but cannot be executed:

- [ ] Exported functions that are never called from within the project (and are not part of an intended public API).
- [ ] Unexported functions that are never called from any code path.
- [ ] Switch/select cases that can never match due to type constraints or value ranges.
- [ ] Code paths guarded by conditions that are always true or always false.
- [ ] Feature flags or configuration options that are defined but never read or acted upon.
- [ ] Build-tagged files for platforms or configurations that are not part of the project's stated support matrix.

#### 3c. Partially Wired Components
Identify components that are structurally present but not connected to the main execution path:

- [ ] Commands registered in the CLI framework but missing their `RunE`/`Run` implementation or always returning immediately.
- [ ] Configuration fields parsed from config files or flags but never used in business logic.
- [ ] Middleware, interceptors, or hooks defined but not registered in the processing pipeline.
- [ ] Database tables, schemas, or migrations defined but not used by any query.
- [ ] Event handlers or callbacks registered but for events that are never emitted.
- [ ] Metrics, counters, or gauges defined but never incremented or observed.

#### 3d. Interface and Contract Gaps
Identify mismatches between defined contracts and their implementations:

- [ ] Interfaces with methods that some implementations leave as no-ops (empty method body or immediate return).
- [ ] Interfaces that have only one implementation — is the abstraction premature or is a second implementation planned?
- [ ] Exported error types or sentinel errors that are defined but never returned.
- [ ] Custom types with methods defined in documentation but not in code.
- [ ] Package-level documentation describing features that do not exist in the package.

#### 3e. Dependency and Import Gaps
- [ ] Packages imported in `go.mod` but not used by any `.go` file (candidates for removal).
- [ ] Internal packages imported by only one other package (may indicate planned but unfinished modularization).
- [ ] Vendor or third-party wrappers that wrap a library but expose only a subset of its functionality with TODOs for the rest.

#### 3f. False-Positive Prevention (MANDATORY)
Before recording ANY finding, apply these checks:

1. **Verify it is actually incomplete**: A minimalist implementation is not a gap. If the function does what its documentation says, it is complete even if you think it should do more.
2. **Check for intentional minimalism**: Some functions are deliberately simple (e.g., a `Close()` that only sets a flag). Read the documentation and tests before flagging as a stub.
3. **Check for external callers**: An exported function with no internal callers may be part of a public API used by external consumers.
4. **Read TODO context**: A TODO with a linked issue or milestone is a tracked item, not an undiscovered gap. Note it but classify as LOW unless it blocks a stated goal.
5. **Verify dead code**: Run `go vet` and check for `deadcode` or `unused` warnings. Do not rely solely on text search — a function may be called via reflection, interface dispatch, or code generation.

**Rule**: If the code fulfills its documented purpose, it is not an implementation gap regardless of what additional features you think it should have. Audit against stated intent, not aspirational completeness.

### Phase 4: Report
Generate **`AUDIT.md`**:
```markdown
# IMPLEMENTATION GAP AUDIT — [date]

## Project Architecture Overview
[Intended architecture, package responsibilities, stated goals]

## Gap Summary
| Category | Count | Critical | High | Medium | Low |
|----------|-------|----------|------|--------|-----|
| Stubs/TODOs | N | N | N | N | N |
| Dead Code | N | N | N | N | N |
| Partially Wired | N | N | N | N | N |
| Interface Gaps | N | N | N | N | N |
| Dependency Gaps | N | N | N | N | N |

## Implementation Completeness by Package
| Package | Exported Functions | Implemented | Stubs | Dead | Coverage |
|---------|-------------------|-------------|-------|------|----------|
| [pkg] | N | N | N | N | N% |

## Findings
### CRITICAL
- [ ] [Finding] — [file:line] — [what is missing] — [which stated goal is blocked] — **Remediation:** [specific implementation needed]
### HIGH / MEDIUM / LOW
- [ ] ...

## False Positives Considered and Rejected
| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| [description] | [why it is not actually an implementation gap] |
```

Generate **`GAPS.md`**:
```markdown
# Implementation Gaps — [date]

## [Gap Title]
- **Intended Behavior**: [what the architecture/docs/interfaces say should exist]
- **Current State**: [what actually exists — stub, partial, missing, dead]
- **Blocked Goal**: [which stated project goal this prevents]
- **Implementation Path**: [what specific code needs to be written]
- **Dependencies**: [what must be completed first]
- **Effort**: [small/medium/large]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | Core feature is a stub or placeholder on a critical execution path — the project cannot fulfill its stated purpose without it |
| HIGH | Feature partially implemented with missing branches, error paths, or edge cases that cause incorrect behavior |
| MEDIUM | Component structurally present but not wired into execution, or dead code that causes confusion and maintenance burden |
| LOW | TODOs with linked issues, minor dead code in non-critical paths, or premature abstractions that add complexity without current value |

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Specific implementation**: Describe what code needs to be written — which function, what logic, what tests. Do not recommend "consider implementing this feature."
2. **Respect project architecture**: Recommendations must fit the existing package structure and conventions.
3. **Verifiable**: Include a validation approach (e.g., `go build ./...`, specific test case, `go vet ./...`).
4. **Dependency-aware**: Note if a gap depends on another gap being closed first.

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Use `go-stats-generator` metrics as primary evidence (especially documentation coverage and function analysis).
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Evaluate gaps against the project's **own stated goals and architecture**, not hypothetical features.
- Apply the Phase 3f false-positive prevention checks to every candidate finding before including it.

## Tiebreaker
Prioritize: stubs on critical paths → partially wired features → dead code on maintained paths → interface gaps → dependency gaps → tracked TODOs. Within a level, prioritize by proximity to the project's core stated goals.
