# TASK: Perform a focused audit of error handling in Go code, identifying swallowed errors, incorrect wrapping, panic/recover misuse, and error propagation failures while rigorously preventing false positives.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the error handling audit report
2. **`GAPS.md`** — gaps in error handling relative to the project's stated goals

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project's Error Handling Model
1. Read the project README to understand its purpose, users, and any claims about reliability, error reporting, or failure behavior.
2. Examine `go.mod` for module path, Go version, and error-handling-related dependencies (e.g., `github.com/pkg/errors`, `github.com/hashicorp/go-multierror`, `go.uber.org/multierr`).
3. List packages (`go list ./...`) and identify which packages are on critical paths where error handling failures would have the most impact.
4. Build an **error handling inventory** by identifying:
   - The project's error wrapping convention: `fmt.Errorf("...: %w", err)` vs `errors.Wrap` vs custom wrappers
   - Custom error types (types implementing the `error` interface)
   - Sentinel errors (`var ErrFoo = errors.New(...)`)
   - Error checking patterns: `errors.Is`, `errors.As`, `==` comparison, type assertions
   - `panic` and `recover` usage
   - Error logging patterns: where and how errors are logged vs returned
   - `_` assignments that discard errors
   - Error return conventions: `(T, error)` vs `(T, bool)` vs panic
5. Map the error propagation paths: where do errors originate (I/O, parsing, external calls), and how do they reach the user or caller?
6. Identify the project's error handling conventions — does it wrap at every level? Does it use sentinel errors? Does it log-and-return or log-and-discard?

### Phase 1: Online Research
Use web search to build context that isn't available in the repository:
1. Search for the project on GitHub — read open issues mentioning "error", "panic", "crash", "nil pointer", or "unexpected" to understand known error handling pain points.
2. Research key dependencies from `go.mod` for their error return conventions and expected error handling patterns.
3. Look up Go error handling best practices relevant to the project's domain (e.g., API error responses, CLI error formatting, library error contracts).

Keep research brief (≤10 minutes). Record only findings that are directly relevant to the project's error handling model.

### Phase 2: Baseline
```bash
set -o pipefail
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages > tmp/error-audit-metrics.json
go-stats-generator analyze . --skip-tests
go vet ./... 2>&1 | tee tmp/error-vet-results.txt
```
Delete temporary files when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Error Handling Audit

#### 3a. Swallowed Errors
For every error-returning function call, verify the error is handled:

- [ ] No `err` variable is assigned and then ignored (not checked, not returned, not logged with context).
- [ ] No `_` is used to discard an error from a function that can fail in ways the caller should handle (e.g., `Close()`, `Flush()`, `Write()` on critical paths).
- [ ] `defer f.Close()` on writable files checks the error — a failed `Close()` on a write can mean data loss. Use a named return `err` and `defer func() { cerr := f.Close(); if err == nil { err = cerr } }()`.
- [ ] Errors from `json.Unmarshal`, `json.Marshal`, `xml.Unmarshal`, and similar serialization functions are checked — these can fail with unexpected input.
- [ ] Errors from `fmt.Fprintf`, `io.WriteString`, and `bufio.Writer.Flush` to network connections or files are checked (writes to `os.Stdout`/`os.Stderr` are acceptable to ignore).
- [ ] `log.Fatal` or `os.Exit` is not called in library code — only in `main()` or CLI entry points.

#### 3b. Incorrect Error Wrapping
For every `fmt.Errorf` and error wrapping call, verify:

- [ ] `%w` is used (not `%v` or `%s`) when the caller needs to unwrap the error with `errors.Is` or `errors.As`.
- [ ] `%w` is NOT used when the error should be opaque to callers (implementation detail, not part of the API contract).
- [ ] Error messages add context about what the current function was doing: `fmt.Errorf("opening config file %s: %w", path, err)` — not `fmt.Errorf("error: %w", err)`.
- [ ] Error messages do not start with a capital letter or end with punctuation (Go convention for composable error chains).
- [ ] Double-wrapping is avoided: `fmt.Errorf("failed: %w", fmt.Errorf("also failed: %w", err))` creates redundant nesting.
- [ ] Error wrapping at package boundaries converts internal errors to the package's public error types where the package defines an error contract.

#### 3c. Error Checking Patterns
For every `errors.Is`, `errors.As`, and error comparison, verify:

- [ ] `errors.Is` is used instead of `==` for sentinel error comparison (supports wrapped errors).
- [ ] `errors.As` is used instead of type assertions for error type checking (supports wrapped errors).
- [ ] `errors.As` target is a pointer to the error type, not a value: `var pathErr *os.PathError; errors.As(err, &pathErr)`.
- [ ] Error type switches (`switch err.(type)`) are not used where `errors.As` is needed (type switches do not unwrap).
- [ ] Sentinel errors used across package boundaries are exported and documented.
- [ ] `errors.Is(err, io.EOF)` is used correctly — `io.EOF` is not an error in many contexts (e.g., `bufio.Scanner`, `io.ReadAll`), and wrapping it with `%w` can break callers.

#### 3d. Nil Pointer and Nil Error Confusion
- [ ] Functions that return `(T, error)` do not return a non-nil error with a non-zero-value `T` unless documented (callers may check only `err`).
- [ ] Functions that return `(*T, error)` do not return `(nil, nil)` when a result is expected — this forces callers to handle an ambiguous case.
- [ ] Interface-typed error returns do not return a typed nil: `var err *MyError; return err` returns a non-nil `error` interface. Use `return nil` explicitly.
- [ ] Methods on pointer receivers check for nil receiver where the type may be used as a nil pointer (e.g., `func (t *T) Error() string` when `t` may be nil).
- [ ] Error checking happens before using the success value: `resp, err := http.Get(url); if err != nil { ... }; defer resp.Body.Close()` — not the reverse.

#### 3e. Panic and Recover
For every `panic` and `recover` usage, verify:

- [ ] `panic` is used only for truly unrecoverable programmer errors (e.g., invalid state that indicates a bug), not for expected runtime errors like I/O failures or invalid user input.
- [ ] `recover` is used in goroutines that must not crash the process (e.g., HTTP handlers, worker pool goroutines).
- [ ] `recover` only catches panics in the same goroutine — it cannot catch panics in child goroutines.
- [ ] Recovered panics are logged with a stack trace (`runtime.Stack` or `debug.Stack`) — silently recovering hides bugs.
- [ ] `panic` in `init()` is justified — a panic during initialization crashes the entire program.
- [ ] API boundaries (exported functions) do not panic on invalid input — they return errors.

#### 3f. Error Propagation in Concurrent Code
- [ ] Errors from goroutines are propagated back to the caller via channels, `errgroup.Group`, or similar — not silently dropped.
- [ ] `errgroup.Group` collects the first error and cancels remaining goroutines (verify the context is connected).
- [ ] Error channels have sufficient buffer or a guaranteed reader to prevent goroutine leaks.
- [ ] Errors in deferred cleanup functions do not mask the original error from the function body.
- [ ] `select` statements that read from error channels include a `context.Done()` case to prevent blocking forever.

#### 3g. False-Positive Prevention (MANDATORY)
Before recording ANY finding, apply these checks:

1. **Verify the error matters**: A discarded error from `fmt.Println` to stdout is not a bug. A discarded error from `f.Close()` on a read-only file is generally acceptable. Evaluate the consequence of the error being non-nil.
2. **Check the project's conventions**: If the project consistently uses `log.Printf` + return for a class of errors, a single instance following that pattern is not a finding — it is consistency.
3. **Trace the error path**: Confirm the error can actually be non-nil. A `json.Marshal` on a struct with common basic data is usually safe, but it can still fail (for example, if float fields contain `NaN`, `+Inf`, or `-Inf`). A `strings.NewReader` cannot return an error from `Read`.
4. **Read surrounding comments**: If a comment explicitly acknowledges an error handling decision (e.g., `// error intentionally ignored: best-effort logging`, `//nolint:`, or a TODO tracking a known issue), treat it as an acknowledged pattern — do not report it as a new finding.
5. **Assess the impact**: A swallowed error in a debug logging path is LOW. A swallowed error in a data persistence path is CRITICAL. Classify by consequence, not by pattern.
6. **Check for alternative error handling**: An error that appears unhandled may be handled by a higher-level mechanism (e.g., HTTP middleware that catches panics, or a deferred function that checks a named return).

**Rule**: If you cannot demonstrate that the error handling issue leads to a concrete problem (data loss, silent failure, crash, incorrect behavior), do NOT report it. Speculative findings waste remediation effort.

### Phase 4: Report
Generate **`AUDIT.md`**:
```markdown
# ERROR HANDLING AUDIT — [date]

## Project Error Handling Model
[Summary: wrapping convention, custom error types, sentinel errors, panic policy, logging vs returning, error propagation approach]

## Error Handling Inventory
| Package | Custom Errors | Sentinels | Panics | Recovers | Discarded (_) | Error Wrapping |
|---------|--------------|-----------|--------|----------|---------------|----------------|
| [pkg]   | N            | N         | N      | N        | N             | %w / %v / pkg  |

## Findings
### CRITICAL
- [ ] [Finding] — [file:line] — [evidence: error path, consequence] — [impact] — **Remediation:** [specific fix]
### HIGH / MEDIUM / LOW
- [ ] ...

## False Positives Considered and Rejected
| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| [description] | [why it is not a real error handling issue] |
```

Generate **`GAPS.md`**:
```markdown
# Error Handling Gaps — [date]

## [Gap Title]
- **Stated Goal**: [what the project claims about reliability/error reporting]
- **Current State**: [what error handling exists]
- **Risk**: [what could go wrong: silent data loss, confusing errors, crashes]
- **Closing the Gap**: [specific error handling changes needed]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | Swallowed error on a data persistence/mutation path causing silent data loss, panic in library code reachable by callers, or typed nil error return causing downstream nil pointer dereference |
| HIGH | Swallowed error on a critical path (not data loss but incorrect behavior), missing error propagation from goroutines, or `%v` instead of `%w` breaking a documented error contract |
| MEDIUM | Incorrect error wrapping reducing debuggability, missing `defer Close()` error check on writable resources, or `panic` for expected runtime errors |
| LOW | Error message style inconsistencies, missing context in error wrapping, or error handling patterns that differ from project conventions without consequence |

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Concrete fix**: State exactly what error check, wrapping change, or recovery mechanism to add and where. Do not recommend "consider handling the error."
2. **Respect project idioms**: If the project uses `fmt.Errorf` with `%w`, do not recommend `github.com/pkg/errors`. If it uses sentinel errors, recommend `errors.Is` patterns.
3. **Verifiable**: Include a validation approach (e.g., `go vet ./...`, `errcheck ./...`, or a specific test case).
4. **Minimal scope**: Fix the error handling issue without restructuring unrelated code.

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Use `go-stats-generator` metrics as supporting evidence for complexity and function length.
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Every finding must include the consequence of the error handling failure — no speculative findings.
- Evaluate the code against its **own error handling conventions** and stated reliability goals, not arbitrary external standards.
- Apply the Phase 3g false-positive prevention checks to every candidate finding before including it.

## Tiebreaker
Prioritize: silent data loss → crashes/panics → incorrect behavior → lost debuggability → style inconsistency. Within a level, prioritize by proximity to the project's critical paths and by severity of the consequence.
