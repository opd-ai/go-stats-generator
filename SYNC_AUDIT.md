# TASK: Perform a focused audit of concurrency and synchronization problems in Go code, identifying deadlocks, race conditions, and resource conflicts while rigorously preventing false positives.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the concurrency/synchronization audit report
2. **`GAPS.md`** — gaps in concurrency safety relative to the project's stated goals

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project's Concurrency Model
1. Read the project README to understand its purpose, users, and any claims about concurrency, parallelism, or thread-safety.
2. Examine `go.mod` for module path, Go version, and concurrency-related dependencies (e.g., `golang.org/x/sync`, `github.com/sourcegraph/conc`).
3. List packages (`go list ./...`) and identify which packages use concurrency primitives.
4. Build a **concurrency inventory** by scanning for:
   - `go ` statements (goroutine launches)
   - `chan` declarations and operations (`<-`, `select`)
   - `sync.Mutex`, `sync.RWMutex`, `sync.WaitGroup`, `sync.Once`, `sync.Map`, `sync.Cond`, `sync.Pool`
   - `atomic.*` operations
   - `context.Context` propagation and cancellation
   - `errgroup.Group` and similar coordination primitives
5. Map which goroutines communicate with which, and through what mechanism (channels, shared memory, or both).
6. Identify the project's concurrency conventions — does it prefer channels over mutexes? Does it use context consistently? Does it document goroutine ownership?

### Phase 1: Online Research
Use web search to build context that isn't available in the repository:
1. Search for the project on GitHub — read open issues mentioning "race", "deadlock", "hang", "goroutine leak", or "concurrent" to understand known concurrency pain points.
2. Research key concurrency-related dependencies from `go.mod` for known issues, thread-safety caveats, or correct usage patterns.
3. Look up common Go concurrency anti-patterns relevant to the project's domain (e.g., HTTP handler concurrency, database connection pool safety, file I/O under concurrency).

Keep research brief (≤10 minutes). Record only findings that are directly relevant to the project's concurrency model.

### Phase 2: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages > /tmp/sync-audit-metrics.json
go-stats-generator analyze . --skip-tests
go test -race ./... 2>&1 | tee /tmp/sync-race-results.txt
go vet ./... 2>&1 | tee /tmp/sync-vet-results.txt
```
Delete temporary files when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

**Critical**: The `-race` detector results are primary evidence. Any race detected by `-race` is a confirmed finding, not a candidate — escalate immediately to CRITICAL.

### Phase 3: Concurrency & Synchronization Audit

#### 3a. Deadlock Analysis
For every mutex, channel, and wait group usage, verify:

**Mutex deadlocks:**
- [ ] No function holds two or more mutexes simultaneously without a documented, consistent lock ordering.
- [ ] No recursive locking (a goroutine attempting to lock a mutex it already holds).
- [ ] Every `Lock()` has a corresponding `Unlock()` on all code paths, including error returns and panics (prefer `defer mu.Unlock()` immediately after `Lock()`).
- [ ] `RWMutex` usage is correct: no `RLock()` followed by `Lock()` in the same goroutine.
- [ ] No lock held across a blocking operation (channel send/receive, I/O, sleep) unless explicitly justified.

**Channel deadlocks:**
- [ ] Every unbuffered channel send has a guaranteed concurrent receiver (and vice versa).
- [ ] Buffered channels are sized appropriately — a full buffer with no reader causes a deadlock.
- [ ] All channels are eventually closed by the sender (or documented as intentionally long-lived).
- [ ] No goroutine blocks forever on a channel operation without a `select` with `context.Done()` or timeout.
- [ ] `select` statements with `default` cases do not spin-loop (burning CPU).

**WaitGroup deadlocks:**
- [ ] `wg.Add()` is called before launching the goroutine, not inside it.
- [ ] Every `wg.Add(n)` has exactly `n` corresponding `wg.Done()` calls on all paths.
- [ ] `wg.Wait()` is not called from a goroutine that also needs to call `wg.Done()`.

#### 3b. Race Condition Analysis
For every piece of shared mutable state, verify:

- [ ] All reads and writes to shared variables are protected by the same synchronization mechanism.
- [ ] No struct fields are accessed concurrently without synchronization (especially in types that launch goroutines or are used by handlers).
- [ ] Map access is protected — Go maps are not safe for concurrent read/write.
- [ ] Slice append operations on shared slices are synchronized.
- [ ] Interface values and pointers assigned from multiple goroutines are protected.
- [ ] `sync.Once` is used (not ad-hoc `if !initialized` checks) for lazy initialization.
- [ ] `atomic` operations use the correct memory ordering and are not mixed with non-atomic access to the same variable.

#### 3c. Goroutine Lifecycle Analysis
For every goroutine launch (`go func()` or `go namedFunc()`), verify:

- [ ] The goroutine has a clear termination condition — no goroutine runs forever without a shutdown signal.
- [ ] The goroutine respects `context.Context` cancellation where applicable.
- [ ] Errors from goroutines are propagated back to the caller (via channels, `errgroup`, or similar).
- [ ] Goroutine leaks: no goroutine blocks indefinitely on a channel or lock after the parent function returns.
- [ ] Panic recovery: goroutines that can panic have `recover()` or the caller handles the crash.
- [ ] Closure captures: loop variables captured by goroutine closures are either passed as parameters or explicitly copied.

#### 3d. Resource Contention & Starvation
- [ ] `RWMutex` read-heavy workloads do not starve writers.
- [ ] Worker pools have bounded concurrency — unbounded `go` in a loop is flagged.
- [ ] Rate limiters and semaphores are used where external resource access is concurrent.
- [ ] Database connections, file handles, and network connections used from multiple goroutines are either pooled or individually protected.

#### 3e. False-Positive Prevention (MANDATORY)
Before recording ANY finding, apply these checks:

1. **Trace the full execution path**: Confirm the problematic code is actually reachable from concurrent call sites. Code that only runs in `init()`, `main()`, or sequential test functions is not concurrent.
2. **Check for higher-level synchronization**: A seemingly unprotected access may be safe because a channel, `sync.Once`, or a higher-level lock serializes all callers. Trace upward.
3. **Verify the goroutine model**: If a struct is documented (or provably used) as single-goroutine-only, concurrent access findings on it are false positives.
4. **Read surrounding comments**: If a comment explicitly acknowledges a concurrency decision (e.g., `// safe: only accessed from the main goroutine`, `//nolint:`, or a TODO tracking a known issue), treat it as an acknowledged pattern — do not report it as a new finding.
5. **Check `-race` output**: If `go test -race` passes for the package in question, downgrade any speculative race condition finding to MEDIUM at most — the dynamic race detector has higher authority than static reasoning for confirmed absence.
6. **Confirm channel direction**: Before flagging a channel deadlock, verify the channel's direction constraints (`chan<-` vs `<-chan`) and confirm no goroutine in the call graph provides the missing send/receive.

**Rule**: If you cannot demonstrate a concrete execution path that leads to the concurrency bug, do NOT report it. Speculative findings waste remediation effort.

### Phase 4: Report
Generate **`AUDIT.md`**:
```markdown
# SYNC AUDIT — [date]

## Project Concurrency Model
[Summary of the project's concurrency approach: goroutine patterns, synchronization primitives used, channel vs mutex preference, context propagation]

## Concurrency Inventory
| Package | Goroutines | Channels | Mutexes | Atomics | WaitGroups | Context |
|---------|-----------|----------|---------|---------|------------|---------|
| [pkg]   | N         | N        | N       | N       | N          | ✅/❌    |

## Race Detector Results
[Summary of `go test -race` output — confirmed races are CRITICAL]

## Findings
### CRITICAL
- [ ] [Finding] — [file:line] — [evidence: concrete execution path] — [impact] — **Remediation:** [specific fix]
### HIGH / MEDIUM / LOW
- [ ] ...

## False Positives Considered and Rejected
| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| [description] | [why it is not a real concurrency bug] |
```

Generate **`GAPS.md`**:
```markdown
# Concurrency Safety Gaps — [date]

## [Gap Title]
- **Stated Goal**: [what the project claims about concurrency/performance/safety]
- **Current State**: [what synchronization exists]
- **Risk**: [what could go wrong under concurrent use]
- **Closing the Gap**: [specific synchronization changes needed]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | Confirmed race from `-race` detector, demonstrated deadlock path, or goroutine leak on a critical path |
| HIGH | Unprotected shared mutable state on a concurrent path with no higher-level serialization, or missing context cancellation in long-running goroutines |
| MEDIUM | Potential contention under high load, suboptimal lock granularity, or missing `defer Unlock()` pattern |
| LOW | Style issues (e.g., using mutex where channel is idiomatic), missing documentation of concurrency invariants |

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Concrete fix**: State exactly what synchronization primitive to add/change and where. Do not recommend "consider adding a mutex."
2. **Respect project idioms**: If the project prefers channels, don't recommend mutexes (and vice versa) unless there is a clear correctness reason.
3. **Verifiable**: Include a validation command (e.g., `go test -race ./pkg/...`, `go vet -copylocks ./...`).
4. **Minimal scope**: Fix the concurrency bug without restructuring unrelated code.

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Use `go test -race` results as the highest-authority evidence source.
- Use `go-stats-generator` `.patterns.concurrency_patterns` metrics as supporting evidence.
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Every finding must include a concrete execution path demonstrating the bug — no speculative findings.
- Evaluate the code against its **own stated goals** and concurrency model, not arbitrary external standards.
- Apply the Phase 3e false-positive prevention checks to every candidate finding before including it.

## Tiebreaker
Prioritize: confirmed `-race` detections → demonstrated deadlocks → unprotected shared state → goroutine leaks → contention issues → style. Within a level, prioritize by proximity to the project's critical paths.
