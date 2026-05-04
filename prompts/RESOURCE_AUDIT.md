# TASK: Perform a focused audit of resource lifecycle management in Go code, identifying file descriptor leaks, database connection leaks, unclosed handles, defer ordering issues, and cleanup failures while rigorously preventing false positives.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the resource management audit report
2. **`GAPS.md`** — gaps in resource lifecycle management relative to the project's stated goals

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project's Resource Usage
1. Read the project README to understand its purpose, users, and any claims about reliability, resource efficiency, or long-running operation.
2. Examine `go.mod` for module path, Go version, and dependencies that manage external resources (database drivers, file libraries, network clients, cgo bindings).
3. List packages (`go list ./...`) and identify which packages acquire and release external resources.
4. Build a **resource inventory** by scanning for:
   - `os.Open`, `os.Create`, `os.OpenFile` — file handles
   - `os.MkdirTemp`, `os.CreateTemp` — temporary files and directories
   - `sql.Open`, `sql.DB`, `sql.Tx`, `sql.Rows`, `sql.Stmt` — database resources
   - `net.Dial`, `net.Listen`, `net.Conn` — network connections
   - `net/http.Client`, `net/http.Response` — HTTP resources
   - `exec.Command`, `exec.Cmd.Start` — child processes
   - `os.Pipe`, `io.Pipe` — pipe handles
   - `plugin.Open` — plugin handles
   - `compress/gzip.NewReader`, `compress/zlib.NewReader` — compression readers
   - `crypto/tls.Conn` — TLS connections
   - `bufio.Writer` — buffered writers that need `Flush()`
   - `io.Closer` implementations — any custom closeable resource
   - CGo resources: `C.malloc`, `C.fopen`, file descriptors passed to/from C
   - `runtime.SetFinalizer` — finalizer-based cleanup (fragile)
5. Map the resource lifecycle: for each resource type, where is it acquired, how long does it live, and where is it released?
6. Identify the project's resource management conventions — does it use `defer` consistently? Does it use helper functions that wrap acquire/release? Does it document resource ownership?

### Phase 1: Online Research
Use web search to build context that isn't available in the repository:
1. Search for the project on GitHub — read open issues mentioning "leak", "close", "file descriptor", "too many open files", "connection", or "cleanup" to understand known resource management pain points.
2. Research key dependencies from `go.mod` for resource management requirements, cleanup APIs, and known leak patterns.
3. Look up common Go resource management pitfalls relevant to the project's domain (e.g., `sql.Rows` must be closed, `gzip.Reader` must be closed, temporary files must be removed).

Keep research brief (≤10 minutes). Record only findings that are directly relevant to the project's resource lifecycle.

### Phase 2: Baseline
```bash
set -o pipefail
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages > /tmp/resource-audit-metrics.json
go-stats-generator analyze . --skip-tests
go vet ./... 2>&1 | tee /tmp/resource-vet-results.txt
```
Delete temporary files when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Resource Lifecycle Audit

#### 3a. File Handle Leaks
For every `os.Open`, `os.Create`, and `os.OpenFile` call, verify:

- [ ] The file is closed on ALL code paths: success, error, and panic. Prefer `defer f.Close()` immediately after the nil error check.
- [ ] When writing, the `Close()` error is checked — a failed close on a writable file means data may not be flushed to disk.
- [ ] Files opened in a loop are closed within the loop iteration, not deferred to after the loop (deferred closes accumulate and all execute at function exit).
- [ ] `os.CreateTemp` and `os.MkdirTemp` have corresponding cleanup: `defer os.Remove(f.Name())` and `defer os.RemoveAll(dir)`.
- [ ] File handles are not stored in long-lived structs without a `Close()` method and documented ownership.
- [ ] `filepath.Walk` or `filepath.WalkDir` callbacks that open files close them before returning.

#### 3b. Database Resource Leaks
For every database operation, verify:

- [ ] `sql.Rows` is always closed: `defer rows.Close()` immediately after the error check on `Query`.
- [ ] `sql.Rows.Err()` is checked after the iteration loop — iteration may have stopped due to an error, not EOF.
- [ ] `sql.Tx` is always committed or rolled back on all paths. Prefer: `defer tx.Rollback()` at the start, then `tx.Commit()` at the end (rollback on already-committed tx is a no-op).
- [ ] `sql.Stmt` prepared statements are closed when no longer needed: `defer stmt.Close()`.
- [ ] `sql.DB` connections are not opened per-request — a single `sql.DB` instance manages a connection pool.
- [ ] `sql.DB.SetMaxOpenConns` is configured to prevent file descriptor exhaustion.
- [ ] Database operations use `context.Context` with timeouts to prevent connections from being held indefinitely.

#### 3c. Network Resource Leaks
For every network resource, verify:

- [ ] `net.Conn` from `net.Dial` or `net.Listener.Accept` is closed on all paths: `defer conn.Close()`.
- [ ] `http.Response.Body` is fully read and closed: `defer resp.Body.Close()` after the error check, and `io.Copy(io.Discard, resp.Body)` to drain before close for connection reuse.
- [ ] `net.Listener` is closed on shutdown: `defer listener.Close()`.
- [ ] WebSocket connections are closed with a proper close frame, not just dropped.
- [ ] gRPC client connections (`grpc.ClientConn`) are closed when no longer needed.

#### 3d. Child Process and Pipe Leaks
- [ ] `exec.Cmd.Start` is followed by `cmd.Wait()` on all paths — un-waited processes become zombies and their pipes leak.
- [ ] When using `exec.Cmd.StdoutPipe` and `StderrPipe`, stdout/stderr are drained concurrently (for example, via goroutines), then `cmd.Wait()` is called, and the drain goroutines are allowed to finish before returning — reading either pipe synchronously to completion before `Wait()` can deadlock if the child blocks on a full pipe buffer, while calling `Wait()` too early can truncate reads because it closes the pipes.
- [ ] `os.Pipe` and `io.Pipe` writers are closed to signal EOF to readers.
- [ ] Child process stdout/stderr are captured or redirected — uncaptured output can fill OS pipe buffers and cause deadlocks.
- [ ] Long-running child processes have a kill mechanism (context-based or signal-based) for shutdown.

#### 3e. Defer Ordering and Correctness
For every `defer` statement, verify:

- [ ] `defer` is placed immediately after the resource is acquired and the error is checked — not after additional operations that might return early.
- [ ] `defer` in a loop body is intentional — deferred calls accumulate until the enclosing function returns, not until the loop iteration ends. Use an anonymous function or explicit close instead.
- [ ] `defer` order is correct when multiple resources depend on each other — defers execute LIFO (last-in, first-out). E.g., `defer db.Close()` then `defer rows.Close()` means rows close first (correct).
- [ ] `defer` with method calls on interfaces captures the correct receiver — `defer r.Close()` where `r` is reassigned later still closes the original value (correct but confusing).
- [ ] Named return values modified in deferred functions have the intended effect — `defer func() { err = f.Close() }()` works, but `defer f.Close()` does not modify the named return.
- [ ] `defer` is not used for performance-critical tight loops where the overhead matters (rare, but document if so).

#### 3f. Custom Resource Types
For every type that manages an external resource (file, connection, handle), verify:

- [ ] The type implements `io.Closer` or has a documented `Close`/`Shutdown`/`Release` method.
- [ ] The `Close` method is idempotent — calling it twice does not panic or return a confusing error.
- [ ] The `Close` method releases ALL held resources, not just some.
- [ ] Constructors (`New*` functions) that can fail clean up partially-acquired resources before returning an error.
- [ ] Documentation states who owns the resource and who is responsible for closing it.
- [ ] `runtime.SetFinalizer` is NOT used as the primary cleanup mechanism (finalizers are non-deterministic).

#### 3g. False-Positive Prevention (MANDATORY)
Before recording ANY finding, apply these checks:

1. **Verify the resource is actually leaked**: Trace the full lifecycle. A file opened and stored in a struct may be closed by the struct's `Close()` method, called by a higher-level owner.
2. **Check for ownership transfer**: A resource returned from a function may be intentionally handed off to the caller for lifecycle management. This is not a leak — it is an ownership transfer.
3. **Verify the execution context**: A resource "leak" in a short-lived CLI process that exits immediately is not the same severity as in a long-running server. The OS reclaims resources on exit.
4. **Read surrounding comments**: If a comment explicitly acknowledges a resource management decision (e.g., `// closed by caller`, `// cleaned up in Shutdown()`, `//nolint:`, or a TODO tracking a known issue), treat it as an acknowledged pattern — do not report it as a new finding.
5. **Check for cleanup at a higher level**: A resource that appears unclosed in one function may be closed in a `defer` in the calling function, in a `Shutdown()` method, or in a `t.Cleanup()` in tests.
6. **Assess materiality**: A leaked `strings.Reader` (which holds no OS resources) is not a finding. Focus on resources that hold OS file descriptors, network sockets, database connections, or external process handles.

**Rule**: If you cannot demonstrate that the resource leak causes a concrete problem (file descriptor exhaustion, connection pool depletion, zombie processes, data loss), do NOT report it. Speculative findings waste remediation effort.

### Phase 4: Report
Generate **`AUDIT.md`**:
```markdown
# RESOURCE AUDIT — [date]

## Project Resource Profile
[Summary: resource types used, lifecycle patterns, long-running vs short-lived, stated reliability goals]

## Resource Inventory
| Package | File Handles | DB Resources | Net Connections | Child Processes | Custom Closers | Temp Files |
|---------|-------------|-------------|-----------------|----------------|---------------|------------|
| [pkg]   | N           | N           | N               | N              | N             | N          |

## Findings
### CRITICAL
- [ ] [Finding] — [file:line] — [evidence: resource lifecycle, leak path] — [impact: FD exhaustion, data loss] — **Remediation:** [specific fix]
### HIGH / MEDIUM / LOW
- [ ] ...

## False Positives Considered and Rejected
| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| [description] | [why it is not a real resource leak in this context] |
```

Generate **`GAPS.md`**:
```markdown
# Resource Management Gaps — [date]

## [Gap Title]
- **Stated Goal**: [what the project claims about reliability/resource management]
- **Current State**: [what resource lifecycle management exists]
- **Risk**: [what could go wrong: FD exhaustion, data loss, zombie processes]
- **Closing the Gap**: [specific resource management changes needed]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | File descriptor or connection leak on a critical path in a long-running process, `sql.Rows` not closed in a request handler (connection pool exhaustion), or `Close()` error ignored on writable file causing data loss |
| HIGH | Resource not closed on error paths, `defer` in a loop accumulating unclosed resources, or child process not waited (zombie leak) |
| MEDIUM | Missing `defer Close()` where the resource is closed later but not on all paths, `Close()` error unchecked on non-critical resource, or temp files not cleaned up |
| LOW | `Close()` error unchecked on read-only file, resource ownership not documented, or minor defer ordering concerns |

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Concrete fix**: State exactly where to add `defer Close()`, how to restructure the acquisition/release, or what cleanup to add. Do not recommend "consider closing the resource."
2. **Respect project idioms**: If the project uses helper functions for resource management, recommend fixes using those patterns.
3. **Verifiable**: Include a validation approach (e.g., `go vet ./...`, `lsof` monitoring, or a specific test case that checks resource cleanup).
4. **Minimal scope**: Fix the resource management issue without restructuring unrelated code.

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Use `go-stats-generator` metrics as supporting evidence for complexity and function length.
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Every finding must include a concrete resource lifecycle demonstrating the leak or mismanagement — no speculative findings.
- Evaluate the code against its **own stated goals** and resource management conventions, not arbitrary external standards.
- Apply the Phase 3g false-positive prevention checks to every candidate finding before including it.

## Tiebreaker
Prioritize: OS resource exhaustion (FDs, connections) → data loss from unchecked close → zombie processes → temp file accumulation → documentation gaps. Within a level, prioritize by resource scarcity (database connections > file handles > memory) and proximity to the project's critical paths.
