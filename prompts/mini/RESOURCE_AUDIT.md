# TASK: Perform a resource lifecycle audit identifying FD leaks, connection leaks, unclosed handles, defer ordering issues, and cleanup failures.

## Output
Write two files in repo root: `AUDIT.md` (resource audit report) and `GAPS.md` (resource management gaps).

## Workflow

### Phase 0: Resource Inventory
1. Read README for reliability/long-running claims.
2. Check `go.mod` for resource deps (DB drivers, network, file libs).
3. Scan for: `os.Open/Create`, `os.MkdirTemp/CreateTemp`, `sql.Open/DB/Tx/Rows/Stmt`, `net.Dial/Listen/Conn`, `http.Client/Response`, `exec.Command`, `io.Pipe`, `gzip/zlib` readers, `tls.Conn`, `bufio.Writer`, custom `io.Closer`.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages > tmp/resource-metrics.json
go vet ./...
```

### Phase 2: Audit Checklist
- [ ] **Files**: Every `os.Open/Create` has `defer Close()` on all paths. Writable files check Close error.
- [ ] **Temp Files**: `os.CreateTemp`/`MkdirTemp` cleaned up via `defer os.Remove`/`RemoveAll`.
- [ ] **Database**: `sql.Rows`, `sql.Stmt`, `sql.Tx` closed on all paths (including errors). `sql.DB` closed on shutdown.
- [ ] **Network**: `net.Conn`, `http.Response.Body`, `tls.Conn` closed on all paths. HTTP client reused (not created per-request).
- [ ] **Defer Ordering**: No `defer` in loops accumulating resources. Acquire → defer Close immediately.
- [ ] **Processes**: `exec.Cmd.Wait()` called to avoid zombies. Stdout/stderr pipes drained.
- [ ] **Compression**: `gzip.Reader`/`zlib.Reader` closed (they wrap `io.Closer`).
- [ ] **Custom Closers**: All `io.Closer` implementations closed on all paths.

### Phase 3: Report
**AUDIT.md**: Resource profile, inventory table, findings (CRITICAL/HIGH/MEDIUM/LOW) with file:line, leak path, impact (FD exhaustion, data loss), remediation.

**GAPS.md**: Per gap — Stated Goal, Current State, Risk, Closing the Gap.

## Constraints
- Report ONLY — no code changes
- Every finding: file:line, concrete lifecycle leak, actionable fix
- Use unchecked `- [ ]` checkboxes