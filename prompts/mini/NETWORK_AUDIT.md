# TASK: Perform a focused audit of networking code in Go, identifying connection leaks, timeout misconfigurations, TLS issues, HTTP client/server bugs, and protocol handling errors while rigorously preventing false positives.

## Output
Write two files in repo root: `AUDIT.md` (networking audit report) and `GAPS.md` (networking robustness gaps).

## Workflow

### Phase 0: Networking Inventory
1. Read README for networking capabilities, performance, reliability claims.
2. Check `go.mod` for networking deps (`net/http`, `grpc`, `websocket`, DB drivers, `x/net`).
3. Scan for: `http.Client`/`Transport`, `http.Server`, `net.Dial`/`Listen`/`Conn`, `tls.Config`, `context.Context` in net calls, DNS, URL parsing, gRPC, WebSocket, DB pools, retry/backoff, circuit breakers.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages > tmp/network-metrics.json
go vet ./...
```

### Phase 2: Audit Checklist
- [ ] **HTTP Client timeouts**: `Client.Timeout` set; `Transport.DialContext`, `TLSHandshakeTimeout`, `ResponseHeaderTimeout`, `IdleConnTimeout` configured; per-request `context` with deadline
- [ ] **Response body**: Every `resp.Body.Close()` called (use `defer`); no body leaks
- [ ] **Connection pooling**: `MaxIdleConns`, `MaxConnsPerHost`, `MaxIdleConnsPerHost` tuned; no exhaustion
- [ ] **Server timeouts**: `Server.ReadTimeout`, `WriteTimeout`, `IdleTimeout`, `ReadHeaderTimeout` set; graceful shutdown with `Shutdown()`
- [ ] **TLS**: `MinVersion` >= 1.2; `InsecureSkipVerify` false; cert validation; no weak ciphers
- [ ] **Retries**: Bounded retries with exponential backoff; `Retry-After` respected; idempotency keys for non-GET
- [ ] **Context propagation**: `context` passed through all net calls; cancellation respected; no context leaks

### Phase 3: Report
**AUDIT.md**: Project networking profile, inventory, findings by class (CRITICAL/HIGH/MEDIUM/LOW) with file:line, failure scenario, remediation.

**GAPS.md**: Per gap — Stated Goal, Current State, Risk, Closing the Gap.

## Constraints
- Report ONLY — no code changes
- Every finding: file:line, concrete failure scenario, actionable fix
- No speculative findings
- Use unchecked `- [ ]` checkboxes
- Evaluate against project's deployment model and goals
- Prioritize: leaks/hangs > TLS vulns > missing error handling > suboptimal config