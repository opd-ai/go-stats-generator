# TASK: Perform a universal, comprehensive bug-hunting audit of a Go project — no domain is off-limits.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the consolidated bug-hunting audit report
2. **`GAPS.md`** — gaps between what the project claims to do and what it actually does

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## What This Prompt Covers
This prompt is a universal bug-hunter. Unlike domain-specific audits, it hunts across **all** bug classes simultaneously:

| Domain | Bug Classes |
|--------|-------------|
| Logic | Arithmetic errors, boolean logic, boundary conditions, off-by-one, unreachable branches |
| Memory | Leaks, dangling pointers, excessive allocation, goroutine retention |
| Concurrency | Race conditions, deadlocks, channel misuse, sync primitive errors |
| Error handling | Swallowed errors, incorrect wrapping, missing propagation, panic/recover misuse |
| Resources | File/connection leaks, missing defers, unclosed bodies |
| Security | Injection, path traversal, hardcoded secrets, weak crypto |
| API | Breaking changes, inconsistent behavior, undocumented invariants |
| Initialization | Dependency order, zero-value misuse, global state races |
| Data aliasing | Shallow copy bugs, unintended mutation, loop variable capture |
| Performance | Hot-path allocations, O(n²) algorithms, lock contention |
| Testing | Coverage gaps on critical paths, test-only bugs |
| Documentation | Behavioral claims that code does not satisfy |

## Workflow

### Phase 0: Understand the Project
1. Read the project README — extract every stated goal, feature claim, performance target, and audience promise. These are the **acceptance criteria** for the audit.
2. Examine `go.mod` for module path, Go version, and dependencies.
3. List packages with `go list ./...` and classify each by role: core logic, CLI, storage, output, utility.
4. Identify the project's conventions: error handling style (sentinel, `%w`, custom types), nil handling, concurrency patterns, test strategy.
5. Note which packages are on **critical paths** — packages that implement the project's primary stated goals deserve deeper scrutiny.
6. Map the project's trust boundaries: where does untrusted input enter, and how far does it travel before validation?

### Phase 1: Online Research
1. Search for the project on GitHub — read open issues, recent PRs, and security advisories.
2. Research key dependencies from `go.mod` for known CVEs and deprecations.
3. Look up best practices and known failure modes for the project's primary domain.

Keep research brief (≤10 minutes). Record only findings with direct bearing on the code.

### Phase 2: Baseline
```bash
mkdir -p tmp
go-stats-generator analyze . --skip-tests --format json \
  --sections functions,packages,documentation,duplication,patterns,interfaces,structs \
  > tmp/generic-audit-baseline.json
go-stats-generator analyze . --skip-tests
go test -race ./... 2>&1 | tee tmp/test-results.txt
go vet ./... 2>&1 | tee tmp/vet-results.txt
```
Delete `tmp/` when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Systematic Bug Hunt

Work through the following checklist. For each category, inspect every relevant code site — do not stop at the first finding.

#### 3a. Structural Risk Scan (use go-stats-generator output)
- Functions with cyclomatic complexity >15 or length >50 lines are **automatically high-risk** — inspect all of them manually.
- Packages with doc coverage <50% may have undocumented behavioral contracts — inspect exported APIs.
- Duplicate code blocks (similarity >0.80, >=10 lines) may have drifted — verify all copies are consistent.

#### 3b. Logic Bugs
- [ ] Off-by-one errors in slice indexing, loop bounds, and length/capacity comparisons.
- [ ] Arithmetic overflow: integer addition/multiplication where the result is used in `make()`, slice bounds, or compared against a limit.
- [ ] Boolean logic inversions: `!= nil` vs `== nil`, `>` vs `>=`, `&&` vs `||`.
- [ ] Unreachable branches: conditions that can never be true given preceding guards.
- [ ] Incorrect operator precedence in complex boolean or arithmetic expressions.
- [ ] State machines with missing, duplicate, or impossible transitions.
- [ ] Comparison of floating-point values with `==` where epsilon comparison is needed.

#### 3c. Nil and Boundary Safety
- [ ] Pointer dereferences without nil checks after conditional assignment or type assertion.
- [ ] Map reads/writes without prior `make()` where the map may be nil.
- [ ] Slice access `s[i]` where `i` is user-controlled or derived from external data without bounds validation.
- [ ] Type assertions `x.(T)` without the two-value form `v, ok := x.(T)`.
- [ ] `interface{}` / `any` values cast without verifying the dynamic type.

#### 3d. Error Handling
- [ ] Errors silently discarded: `val, _ := f()` where the error is meaningful.
- [ ] Errors checked but not propagated: early return missing or condition inverted.
- [ ] `fmt.Sprintf` used instead of `fmt.Errorf` for error construction (loses context).
- [ ] `panic` used outside of truly unrecoverable initialization paths.
- [ ] `recover()` used to silently swallow panics without logging or re-raising.
- [ ] Error wrapping inconsistent with the project's established convention.
- [ ] Sentinel errors compared with `==` across package boundaries (should use `errors.Is`).

#### 3e. Resource Lifecycle
- [ ] `os.Open`, `os.Create`, `http.Get`, database queries — missing `defer f.Close()` or equivalent.
- [ ] `defer` inside a loop that defers resource release until function return, not loop iteration.
- [ ] HTTP response bodies not closed after reading (goroutine and FD leak).
- [ ] Database rows not iterated to completion or explicitly closed.
- [ ] `context.WithCancel`, `context.WithTimeout` — cancel functions not called (context leak).

#### 3f. Concurrency
- [ ] Shared variables read/written from multiple goroutines without synchronization.
- [ ] Goroutines that block on a channel send/receive with no consumer/producer (goroutine leak).
- [ ] Mutexes copied by value (pass by pointer or embed in the type).
- [ ] `sync.WaitGroup` used incorrectly: `Add` called after goroutine starts, or `Done` called in wrong goroutine.
- [ ] Loop variables captured by goroutine closures without copying first (pre-Go 1.22).
- [ ] `select` with no `default` that can block indefinitely without a context cancellation path.

#### 3g. Security
- [ ] User-controlled input reaches `os/exec.Command` arguments without allowlist validation.
- [ ] File paths derived from user input not validated to stay within a trusted root directory.
- [ ] SQL queries constructed with string concatenation instead of parameterized queries.
- [ ] `html/template` vs `text/template` misuse (use `html/template` for HTML output).
- [ ] `math/rand` used for security-sensitive operations instead of `crypto/rand`.
- [ ] Hardcoded credentials, tokens, or private keys in source code.
- [ ] `InsecureSkipVerify: true` in TLS configurations.
- [ ] `//go:embed` directives that may include `.env`, key files, or other secrets.

#### 3h. Data Aliasing and Mutation
- [ ] `s2 := s1` (slice header copy) followed by `append` to `s2` — may silently modify `s1` if `len < cap`.
- [ ] Map or struct passed by value where caller expects modifications to be visible.
- [ ] Shallow copy of a struct containing pointer fields — both copies share the pointed-to data.
- [ ] In-place sort (`sort.Slice`, `slices.Sort`) applied to a slice the caller still references.
- [ ] Return of a slice backed by a shared buffer — caller modifications corrupt the buffer.

#### 3i. Initialization Order
- [ ] `init()` functions that depend on global variables set by other `init()` functions in the same package (order is undefined within a package).
- [ ] Global variables initialized to a value that depends on a function call that may fail silently.
- [ ] Types whose zero value is invalid but which can be created without the constructor.
- [ ] Lazy initialization of shared globals without a `sync.Once` or mutex.

#### 3j. API and Behavioral Contracts
- [ ] Functions that silently ignore invalid input instead of returning an error.
- [ ] Exported functions that panic on invalid input rather than returning an error.
- [ ] Behavioral difference between what the README/GoDoc promises and what the code does.
- [ ] Functions with non-obvious side effects not documented in their GoDoc.
- [ ] Interface contracts not enforced: implementations that skip required behavior.

#### 3k. Performance Red Flags (only flag confirmed hot-path issues)
- [ ] O(n²) or worse algorithms on paths that process user-controlled or large inputs.
- [ ] Allocations inside tight loops that could be hoisted or pre-allocated.
- [ ] Unbounded channel or slice growth on paths triggered by external input.
- [ ] `fmt.Sprintf` in hot paths where `strings.Builder` would avoid allocation.

#### 3l. False-Positive Prevention (MANDATORY)
Before recording ANY finding, apply these checks:
1. **Trace the data flow**: Confirm the bug is actually reachable via a real call path.
2. **Check for upstream guards**: Input may be validated earlier — trace upward through the call chain.
3. **Read surrounding comments**: A comment explicitly acknowledging a pattern (e.g., `//nolint:`, `// safe: ...`, `// intentional`) is an acknowledged pattern — do not report it.
4. **Assess exploitability and impact**: A theoretical issue with no practical consequence in the project's deployment model is at most LOW.
5. **Run the tests**: `go test -race ./...` passing is evidence (not proof) against some race conditions and panics — note this in findings where relevant.

**Rule**: Only report confirmed findings where you can state the exact file, line, code path, and concrete consequence. Speculative findings must be labeled LOW with explicit uncertainty noted.

### Phase 4: Report

Generate **`AUDIT.md`**:
```markdown
# UNIVERSAL BUG AUDIT — [date]

## Project Profile
[Purpose, target users, deployment model, critical paths]

## Audit Scope
[Packages audited, total functions inspected, go-stats-generator metrics summary]

## Goal-Achievement Summary
| Stated Goal | Status | Blocking Findings |
|-------------|--------|-------------------|
| [Goal from README] | ✅ / ⚠️ / ❌ | [finding refs] |

## Findings

### CRITICAL
- [ ] [Title] — [file:line] — [bug class] — [concrete consequence] — **Remediation:** [specific fix with validation command]

### HIGH
- [ ] ...

### MEDIUM
- [ ] ...

### LOW
- [ ] ...

## Metrics Snapshot
| Metric | Value |
|--------|-------|
| Total functions | N |
| Functions above complexity 15 | N |
| Avg cyclomatic complexity | N |
| Doc coverage | N% |
| Duplication ratio | N% |
| Test pass rate | N/N |
| go vet warnings | N |

## False Positives Considered and Rejected
| Candidate | Reason Rejected |
|-----------|----------------|
| [description] | [why not exploitable/reachable] |
```

Generate **`GAPS.md`**:
```markdown
# Implementation Gaps — [date]

## [Gap Title]
- **Stated Goal**: [what the project claims]
- **Current State**: [what actually exists]
- **Impact**: [how this gap affects users or correctness]
- **Closing the Gap**: [specific, actionable remediation]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | Confirmed bug on a critical path, data corruption, security exploit with traceable data flow, or feature fully documented but non-functional |
| HIGH | Bug in a high-complexity function, missing error propagation on a primary path, race condition with empirical evidence |
| MEDIUM | Edge-case failure, low-severity security issue, performance problem on a secondary path |
| LOW | Code smell with theoretical risk, missing documentation, minor inconsistency |

## Remediation Standards
Every finding MUST include a **Remediation** section that:
1. States exactly what to change (file, function, line range).
2. Respects the project's existing idioms and error handling conventions.
3. Includes a validation command (`go test -race ./...`, `go vet ./...`, or a specific test case).
4. Is proportionate — do not recommend architectural rewrites for LOW findings.

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Every CRITICAL or HIGH finding must include a concrete data flow or code path demonstrating the issue.
- Evaluate the code against **its own stated goals**, not arbitrary external standards.
- Apply the Phase 3l false-positive prevention checks to every candidate finding before including it.

## Tiebreaker
Within each severity group: confirmed bugs on critical goal paths first → security findings → resource safety → performance. Then descending by cyclomatic complexity. If tied, descending by function length.
