# TASK: Perform a combined breaking-bug audit of a Go project, identifying only bugs that block or break the basic utility of the program — crashes, silent wrong output on documented core paths, non-functional CLI commands, and startup failures.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Scope Principle
This audit is NOT about style, coverage, performance, or cosmetic correctness. Every finding must answer "yes" to at least one of these questions:
1. Does this bug cause a **crash or panic** reachable through normal usage?
2. Does this bug cause **wrong output** on a documented core use case?
3. Does this bug cause a documented CLI command or flag to **silently do nothing or produce garbage**?
4. Does this bug cause the program to **fail to start** on a valid configuration?
5. Does this bug cause a **core data path to produce incorrect metrics** that contradict the tool's stated purpose?

If a candidate finding does not satisfy at least one of the above, discard it. Do not report it.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the breaking-bug audit report
2. **`GAPS.md`** — gaps between the program's stated basic utility and its actual operational state

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Map the Program's Basic Utility
Before searching for bugs, establish what "basic utility" means for this project:

1. Read the project README end-to-end. Extract:
   - The primary value proposition: what is the one thing this tool does that justifies its existence?
   - The documented happy-path invocation(s): what command does a new user run first?
   - The documented output formats and what each produces.
   - Every documented CLI command, flag, and argument.
   - Any stated performance or correctness guarantees ("analyzes 50,000 files in 60 seconds", "zero false positives").
2. Run `--help` on the binary and every subcommand. Record every flag and its documented default.
3. Identify the **critical paths**: the code that must execute correctly for the primary value proposition to hold. These are the only paths this audit cares about.
4. Record the project's error handling convention (sentinel errors, `%w` wrapping, custom types) and its concurrency patterns. Bugs on critical paths that violate these conventions are high-risk.
5. Note the Go version in `go.mod`. Loop variable capture bugs are eliminated in Go 1.22+; do not report them for newer versions.

### Phase 1: Baseline
```bash
set -o pipefail
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages > /tmp/breaking-audit-metrics.json
go-stats-generator analyze . --skip-tests
go build ./... 2>&1 | tee /tmp/breaking-build-results.txt
go test -race -count=1 ./... 2>&1 | tee /tmp/breaking-test-results.txt
go vet ./... 2>&1 | tee /tmp/breaking-vet-results.txt
```

Record:
- Build failures (CRITICAL — program cannot run at all)
- Test failures with race conditions (CRITICAL if on a critical path, HIGH otherwise)
- `go vet` errors (not warnings — errors only)
- Functions with cyclomatic complexity >20 or nesting >5 on critical paths (high-risk zones for breaking bugs)

Delete all `/tmp/breaking-*.txt` and `/tmp/breaking-audit-metrics.json` when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

---

### Phase 2: Breaking Bug Categories

Audit each category below, but only report findings that satisfy the Scope Principle above.

#### 2a. Crash and Panic Paths
Trace every code path reachable from the documented happy-path invocations:

- [ ] **Nil dereference on normal input**: a pointer is used without a nil check on a path that executes for valid, documented inputs. Verify by tracing: can this pointer be nil when the caller follows documented usage?
- [ ] **Slice/array out-of-bounds on valid input**: `slice[i]` or `array[i]` where `i` is derived from input data and is not guarded by a bounds check. Do not report this for internal indices that are always computed from the slice's own length.
- [ ] **Nil map write**: `m[k] = v` where `m` could be nil at the call site (not initialized by `make` and not guaranteed non-nil by the constructor). Confirm by tracing the initialization path.
- [ ] **Division by zero**: `a / b` or `a % b` on a critical-path calculation where `b` is derived from file contents, user input, or collection size that can legitimately be zero (e.g., average over an empty set).
- [ ] **Type assertion without ok-check**: `x.(T)` (not `x, ok := x.(T)`) where `x` is an interface value that may not hold type `T` at runtime.
- [ ] **Unrecorvered panic in goroutine**: a goroutine on a critical path can panic and crash the entire program. Verify the panic path is reachable for normal inputs.
- [ ] **`init()` panic**: an `init()` function panics for a valid installation (e.g., missing file assumed to exist, invalid hardcoded regex).

#### 2b. Silent Wrong Output on Core Use Cases
For each documented core metric or output section, verify the calculation is correct:

- [ ] **Integer division truncation producing wrong metric**: a ratio or percentage is computed as `a / b * 100` (truncates to zero when `a < b`) instead of `a * 100 / b` or a floating-point form. Only report if the truncated value contradicts the metric's documented meaning.
- [ ] **Off-by-one in line/position counting**: a line count, token position, or range includes or excludes one element incorrectly relative to the documented definition. Do not report as off-by-one if the documented definition is ambiguous.
- [ ] **Accumulator reset inside loop**: a running total or counter is reset (reassigned rather than incremented) inside the loop it is supposed to accumulate, producing a final value equal to the last element rather than the true aggregate.
- [ ] **Wrong aggregation level**: a per-function metric is emitted at the per-file or per-package level (or vice versa), causing the output to attribute a value to the wrong unit.
- [ ] **Stale or uninitialized result**: a result struct field used in output is never written after construction, producing the zero value instead of a computed value for all inputs.
- [ ] **Incorrect filter applied to wrong collection**: a filter (e.g., "skip test files") is applied to the wrong slice, causing either no filtering or filtering of the wrong items.

#### 2c. Non-Functional CLI Commands and Flags
For every documented CLI command and flag:

- [ ] **Command registered but `Run`/`RunE` not set**: the command exists in `--help` but produces no output and exits 0 for all inputs.
- [ ] **Flag parsed but never read**: a flag appears in `--help` and is accepted without error, but its value is never used by any code path. Verify by tracing from flag registration to usage site.
- [ ] **Flag silently ignored when combined with another flag**: two flags are documented as independent but one cancels or overrides the other without documentation or a user-visible warning.
- [ ] **Default value mismatch**: the default value documented in `--help` differs from the default value in code, causing the program to behave differently than a user following the docs expects on first run.
- [ ] **Output file flag that silently writes nothing**: a `--output` or `--file` flag is accepted, but the output is still written to stdout (or vice versa), contradicting the documented behavior.
- [ ] **Format flag that produces unparseable output**: a `--format json` (or other machine-readable format) flag produces output that fails to parse with the format's standard parser due to structural errors (not just style).

#### 2d. Startup and Initialization Failures
- [ ] **Configuration file required but not documented**: the program refuses to start without a configuration file that is not mentioned in the installation or quick-start documentation.
- [ ] **Database or storage initialization failure on clean install**: the program crashes or returns a non-zero exit code on first run because a storage backend (e.g., SQLite database) is not initialized before first use.
- [ ] **Package-level `var` initialization ordering bug**: a package-level variable depends on another package-level variable in the same or another package, but Go's initialization order does not guarantee the dependency is ready. This produces wrong zero-value behavior, not a panic.
- [ ] **`init()` that silently sets wrong state**: an `init()` function sets a default that is always wrong for production use (e.g., sets a log level to debug, disables a required feature flag).
- [ ] **Missing required dependency check**: the program uses an external binary or resource (beyond the Go standard library) on a critical path but does not check for its presence before use, causing a confusing error or panic instead of a clear diagnostic.

#### 2e. Error Handling Failures That Produce Broken Output
Only report these if the swallowed or mishandled error causes the program to produce wrong output or exit successfully when it should fail:

- [ ] **Swallowed parse error produces zero-value metric**: a file or AST parse error is discarded (assigned to `_` or logged-and-continued) and the function returns a zero-value struct that is included in aggregate metrics, producing incorrect totals.
- [ ] **Error from concurrent worker silently dropped**: an error from a goroutine processing a file on a critical path is not propagated back to the coordinator, causing the final report to silently omit or miscount results for that file.
- [ ] **Typed nil error causing downstream nil dereference**: a function returns a `(*T, error)` where the error is a typed nil (`var err *MyError; return nil, err`), the caller checks `err != nil` (which is true for a typed nil), and then dereferences the non-nil error interface value, causing a panic.
- [ ] **Context cancellation not respected on critical path**: a long-running analysis ignores `ctx.Done()`, causing the program to continue processing after the user sends SIGINT, potentially writing a partial result to an output file.

#### 2f. Concurrency Bugs That Break Core Results
Only report these if the race condition causes incorrect output or a crash, not merely a theoretical data race:

- [ ] **Shared result map written from multiple goroutines without synchronization**: concurrent goroutines write to the same map, causing a race condition that panics at runtime (Go's map write is not concurrent-safe).
- [ ] **Goroutine leak that exhausts resources on large input**: goroutines are started but never terminated (blocked channel send/receive with no consumer/producer), causing the program to hang or OOM on large codebases.
- [ ] **WaitGroup counter goes negative**: `wg.Done()` is called more times than `wg.Add()`, causing a panic. Trace the Add/Done pairs.
- [ ] **Channel send after close**: a goroutine sends to a channel after it has been closed, causing a panic. Trace the close and send sites.

---

### Phase 3: False-Positive Prevention (MANDATORY)
Before recording ANY finding, apply ALL of these checks:

1. **Reproduce the path**: Confirm the bug is reachable through a documented or reasonably expected invocation. A bug in dead code is not a breaking bug.
2. **Verify the scope**: Does this finding satisfy at least one of the five Scope Principle questions? If not, discard it.
3. **Check for guards elsewhere**: The crash or wrong-output path may be guarded at a higher level (e.g., the function is only called after a nil check in its only caller). Trace the full call chain before reporting.
4. **Read surrounding comments**: If a comment explicitly acknowledges the issue (e.g., `// intentionally not checked`, `// safe because X`, `//nolint:`), classify as an acknowledged pattern and do not report as a new finding.
5. **Check Go version for eliminated bugs**: Loop variable capture (`for _, v := range`; goroutine closes over `v`) is eliminated in Go 1.22+. Integer overflow is undefined behavior in C but well-defined (wrapping) in Go — report it only if wrapping produces a demonstrably wrong result for valid inputs.
6. **Confirm the metric definition**: Before reporting an off-by-one or wrong-calculation bug, read the project's own definition of the metric (README, GoDoc, test assertions). What looks wrong may be intentional.

**Rule**: If you cannot construct a concrete input or invocation that triggers the broken behavior using the documented interface, do NOT report it. Hypothetical bugs are not breaking bugs.

---

### Phase 4: Report
Generate **`AUDIT.md`**:

```markdown
# BREAKING BUG AUDIT — [date]

## Program's Basic Utility
[One paragraph: what this program does, its primary invocation, its core output, and its target users]

## Critical Paths Examined
| Path | Entry Point | Output | Verdict |
|------|------------|--------|---------|
| [documented use case] | [command/function] | [metric/file] | ✅ Functional / ❌ Broken / ⚠️ Degraded |

## Baseline Health
- Build: [PASS / FAIL — details]
- Tests: [N passed, N failed, N races]
- `go vet`: [clean / N errors]

## Findings
### CRITICAL — Program broken or crashes on documented usage
- [ ] [Finding title] — [file:line] — [concrete invocation or input that triggers it] — [what the user sees] — **Remediation:** [exact fix: what to change, where, and how to verify]

### HIGH — Core output is wrong for valid inputs
- [ ] [Finding title] — [file:line] — [concrete invocation or input that triggers it] — [what is wrong vs what is expected] — **Remediation:** [exact fix]

## False Positives Considered and Rejected
| Candidate | Reason Discarded |
|-----------|-----------------|
| [description] | [which scope check failed or which guard was found] |
```

Generate **`GAPS.md`**:

```markdown
# Breaking Utility Gaps — [date]

## [Gap Title]
- **Documented Capability**: [exact claim from README, --help, or GoDoc]
- **Broken Behavior**: [what actually happens — crash, wrong value, silent no-op]
- **Triggering Condition**: [the input or invocation that exposes the break]
- **User Impact**: [what the user loses — wrong data, wasted time, corrupted output file]
- **Fix**: [exact remediation: function, file, and what to change]
```

## Severity Classification

| Severity | Criteria |
|----------|----------|
| CRITICAL | Program crashes or panics on a documented invocation; program writes corrupted output and exits 0; program fails to start on a clean install |
| HIGH | A documented metric is computed incorrectly for valid inputs; a documented CLI flag is silently ignored; a documented output format produces unparseable content |
| DISCARD | Everything else — style issues, test coverage gaps, performance problems, cosmetic wrong values in edge cases not covered by documentation |

There are only two reportable severity levels: CRITICAL and HIGH. If a finding does not meet CRITICAL or HIGH criteria, it is discarded. Do not create MEDIUM or LOW sections.

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Name the exact location**: file path and line number, function name.
2. **State the exact fix**: what to add, change, or remove. Do not say "consider adding a nil check" — say "add `if x == nil { return nil, fmt.Errorf(...) }` before line N in function F".
3. **Respect project idioms**: use the project's existing error wrapping convention, variable naming style, and test patterns.
4. **Include a verification command**: `go test -race -run TestFunctionName ./pkg/...` or a specific CLI invocation that demonstrates the fix.

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Every finding must satisfy at least one of the five Scope Principle questions (listed at the top of this prompt).
- Every finding must reference a specific file and line number.
- Every finding must include a concrete triggering condition — not "could be nil" but "is nil when the input file is empty".
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Apply all Phase 3 false-positive prevention checks to every candidate finding before including it.
- Evaluate against the project's **own documented utility**, not hypothetical ideal behavior.

## Tiebreaker
Prioritize: crashes on default invocation → corrupted output file → wrong core metric → non-functional documented command → wrong default behavior. Within CRITICAL, order by how likely the triggering condition is on a real codebase. Within HIGH, order by how visible the wrong output is to the end user.
