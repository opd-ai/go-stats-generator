# TASK: Perform a focused audit of networking code in Go, identifying connection leaks, timeout misconfigurations, TLS issues, HTTP client/server bugs, and protocol handling errors while rigorously preventing false positives.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the networking audit report
2. **`GAPS.md`** — gaps in networking robustness relative to the project's stated goals

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project's Networking Model
1. Read the project README to understand its purpose, users, and any claims about networking capabilities, API interactions, performance under load, or distributed operation.
2. Examine `go.mod` for module path, Go version, and networking-related dependencies (e.g., `net/http`, `google.golang.org/grpc`, `nhooyr.io/websocket`, database drivers, `golang.org/x/net`).
3. List packages (`go list ./...`) and identify which packages perform network I/O.
4. Build a **networking inventory** by scanning for:
   - `net/http.Client` and `net/http.Transport` usage and configuration
   - `net/http.Server` and handler registration
   - `net.Dial`, `net.Listen`, `net.Conn` usage
   - `tls.Config` and TLS-related setup
   - `context.Context` propagation in network calls
   - DNS resolution and hostname handling
   - `net/url.Parse` and URL construction
   - gRPC client/server setup
   - WebSocket connections
   - Database connection pooling (`sql.DB` configuration)
   - Redis, message queue, or other network service clients
   - Retry and backoff logic
   - Circuit breaker patterns
5. Map the network topology: which services does this project connect to, which protocols does it use, and what are the failure modes?
6. Identify the project's networking conventions — does it use a shared HTTP client? Does it configure timeouts consistently? Does it handle retries?

### Phase 1: Online Research
Use web search to build context that isn't available in the repository:
1. Search for the project on GitHub — read open issues mentioning "timeout", "connection", "EOF", "refused", "DNS", "TLS", or "retry" to understand known networking pain points.
2. Research key networking dependencies from `go.mod` for known connection handling issues, timeout defaults, or recommended configuration.
3. Look up common Go networking pitfalls relevant to the project's domain (e.g., default HTTP client has no timeout, connection pool exhaustion, DNS caching).

Keep research brief (≤10 minutes). Record only findings that are directly relevant to the project's networking behavior.

### Phase 2: Baseline
```bash
set -o pipefail
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages > tmp/network-audit-metrics.json
go-stats-generator analyze . --skip-tests
go vet ./... 2>&1 | tee tmp/network-vet-results.txt
```
Delete temporary files when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Networking Audit

#### 3a. HTTP Client Issues
For every HTTP client usage, verify:

**Timeout configuration:**
- [ ] `http.Client.Timeout` is set — the default `http.DefaultClient` has no timeout, causing goroutines to hang indefinitely on unresponsive servers.
- [ ] `http.Transport.DialContext` timeout, `TLSHandshakeTimeout`, `ResponseHeaderTimeout`, and `IdleConnTimeout` are configured appropriately for the use case.
- [ ] `context.Context` with deadline/timeout is passed to `http.NewRequestWithContext` for per-request timeout control.
- [ ] Long-polling or streaming responses use appropriate timeouts that do not kill legitimate long-lived connections.

**Connection management:**
- [ ] `http.Response.Body` is always read to completion (`io.Copy(io.Discard, resp.Body)`) and closed — partial reads prevent connection reuse.
- [ ] `defer resp.Body.Close()` is called immediately after the error check on the HTTP call, on all code paths.
- [ ] `http.Transport.MaxIdleConns`, `MaxIdleConnsPerHost`, and `MaxConnsPerHost` are configured for the expected workload (defaults may be too low for high-throughput or too high for resource-constrained environments).
- [ ] Custom `http.Transport` instances are shared across requests (creating a new `Transport` per request disables connection pooling).
- [ ] `http.Client` instances are reused, not created per-request.

**Retry and resilience:**
- [ ] Retries use exponential backoff with jitter — not fixed-interval retries that cause thundering herd.
- [ ] Retries are only performed on idempotent requests or requests known to be safe to retry.
- [ ] A maximum retry count or total timeout bounds the retry loop.
- [ ] Non-retryable HTTP status codes (4xx) are not retried.
- [ ] `context.Context` cancellation is respected during retry waits.

#### 3b. HTTP Server Issues
For every HTTP server, verify:

**Timeout configuration:**
- [ ] `http.Server.ReadTimeout`, `WriteTimeout`, and `IdleTimeout` are set — defaults are zero (no timeout).
- [ ] `ReadHeaderTimeout` is set to prevent slowloris attacks.
- [ ] Handler functions respect `request.Context()` cancellation and do not continue processing after the client disconnects.
- [ ] `http.TimeoutHandler` wraps handlers that may take a long time.

**Request handling:**
- [ ] Request body size is limited with `http.MaxBytesReader` to prevent denial of service.
- [ ] Concurrent handler safety: shared state accessed by handlers is synchronized (handlers run in separate goroutines).
- [ ] `http.Error` is followed by `return` — code after `http.Error` continues executing otherwise.
- [ ] Middleware ordering is correct: authentication before authorization before handler logic.

**Graceful shutdown:**
- [ ] `http.Server.Shutdown(ctx)` is used for graceful shutdown, not `Close()` which drops active connections.
- [ ] The shutdown context has a timeout to prevent hanging during shutdown.
- [ ] Long-lived connections (WebSockets, SSE) are properly drained during shutdown.

#### 3c. Raw TCP/UDP and Connection Management
For every `net.Dial`, `net.Listen`, or `net.Conn` usage, verify:

- [ ] `net.Dialer.Timeout` is configured — default is OS-dependent and can be very long.
- [ ] `net.Conn.SetDeadline`, `SetReadDeadline`, or `SetWriteDeadline` is set before blocking reads/writes.
- [ ] Connection close is deferred immediately after successful dial/accept.
- [ ] Listeners are properly closed on shutdown.
- [ ] TCP keepalive is configured for long-lived connections (`net.Dialer.KeepAlive`).
- [ ] Half-open connection detection is handled (peer disconnects without TCP FIN).

#### 3d. TLS Configuration
For every TLS configuration, verify:

- [ ] `tls.Config.MinVersion` is set to `tls.VersionTLS12` or higher.
- [ ] `InsecureSkipVerify` is not `true` in production code (acceptable in tests with a comment explaining why).
- [ ] Custom `tls.Config.CipherSuites` (if specified) does not include known-weak ciphers.
- [ ] Certificate rotation is handled for long-running servers (certificates are reloaded, not cached at startup forever).
- [ ] Client certificate validation (`ClientAuth`, `ClientCAs`) is correct if mutual TLS is used.
- [ ] `tls.Config` is not shared mutably across goroutines after the server starts.

#### 3e. DNS and Name Resolution
- [ ] DNS resolution failures are handled gracefully with retry or fallback logic.
- [ ] `net.Resolver` with a custom dialer is used if DNS timeout control is needed (the default resolver has no configurable timeout).
- [ ] Hardcoded IP addresses are avoided unless there is a documented reason (DNS allows failover).
- [ ] IPv6 compatibility is considered if the project may run in dual-stack environments.

#### 3f. Database and Service Client Connections
- [ ] `sql.DB.SetMaxOpenConns`, `SetMaxIdleConns`, and `SetConnMaxLifetime` are configured.
- [ ] Connection pool exhaustion is handled — queries with `context.Context` timeout rather than blocking forever waiting for a connection.
- [ ] Database connections are not leaked: every `sql.Rows` is closed (`defer rows.Close()`) and every `sql.Tx` is committed or rolled back.
- [ ] Redis, message queue, and other client libraries have appropriate timeout and pool configuration.
- [ ] Health checks or ping operations verify connectivity before assuming a connection is usable.

#### 3g. False-Positive Prevention (MANDATORY)
Before recording ANY finding, apply these checks:

1. **Verify the networking context**: A `http.DefaultClient` usage in a CLI that makes one request and exits is not the same severity as in a long-running server. Evaluate against actual deployment.
2. **Check for wrapper functions**: A seemingly unconfigured client may be wrapped by a function that sets timeouts, retries, or other configuration. Trace the full call chain.
3. **Verify the failure mode matters**: A missing timeout on a connection to `localhost` in a controlled environment is different from a missing timeout on an external API call. Assess the blast radius.
4. **Read surrounding comments**: If a comment explicitly acknowledges a networking decision (e.g., `// timeout handled by context`, `// connection pooled by library`, `//nolint:`, or a TODO tracking a known issue), treat it as an acknowledged pattern — do not report it as a new finding.
5. **Check library defaults**: Some libraries (e.g., gRPC, database drivers) configure timeouts and pooling internally. Verify the library's default behavior before flagging missing configuration.
6. **Assess production relevance**: Test helpers, example code, and development-only configurations should not be reported as production security findings.

**Rule**: If you cannot demonstrate that the networking issue causes a concrete failure mode (connection leak, hang, timeout, data loss) under realistic conditions, do NOT report it. Speculative findings waste remediation effort.

### Phase 4: Report
Generate **`AUDIT.md`**:
```markdown
# NETWORK AUDIT — [date]

## Project Networking Profile
[Summary: network services consumed/provided, protocols, connection patterns, stated performance/reliability goals]

## Networking Inventory
| Package | HTTP Client | HTTP Server | Raw TCP/UDP | TLS | DB Connections | External Services |
|---------|------------|-------------|-------------|-----|---------------|-------------------|
| [pkg]   | N          | N           | N           | N   | N             | [list]            |

## Findings
### CRITICAL
- [ ] [Finding] — [file:line] — [evidence: connection flow, timeout path] — [impact: hang, leak, data loss] — **Remediation:** [specific fix]
### HIGH / MEDIUM / LOW
- [ ] ...

## False Positives Considered and Rejected
| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| [description] | [why it is not a real networking issue in this context] |
```

Generate **`GAPS.md`**:
```markdown
# Networking Gaps — [date]

## [Gap Title]
- **Stated Goal**: [what the project claims about networking/performance/reliability]
- **Current State**: [what networking configuration exists]
- **Risk**: [what could go wrong under load, network instability, or malicious input]
- **Closing the Gap**: [specific networking changes needed]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | Connection leak causing resource exhaustion, missing timeout causing goroutine hang indefinitely, TLS misconfiguration allowing MITM, or data loss due to unhandled network errors |
| HIGH | Default HTTP client without timeout in a server context, missing `resp.Body.Close()`, no graceful shutdown, or unbounded retry loops |
| MEDIUM | Suboptimal connection pool configuration, missing per-request context timeout, or retry without backoff |
| LOW | Missing TCP keepalive, suboptimal idle connection settings, or informational TLS configuration improvements |

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Concrete fix**: State exactly what timeout, pool size, or configuration to set and where. Do not recommend "consider adding a timeout."
2. **Respect project idioms**: If the project uses a specific HTTP client wrapper or connection library, recommend fixes using those tools.
3. **Verifiable**: Include a validation approach (e.g., `go vet ./...`, `curl` test commands, or a specific test case demonstrating the fix).
4. **Minimal scope**: Fix the networking issue without restructuring unrelated code.

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Use `go-stats-generator` metrics as supporting evidence for complexity and function length.
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Every finding must include a concrete failure scenario — no speculative findings.
- Evaluate the code against its **own stated goals** and deployment model, not arbitrary external standards.
- Apply the Phase 3g false-positive prevention checks to every candidate finding before including it.

## Tiebreaker
Prioritize: connection leaks/resource exhaustion → missing timeouts causing hangs → TLS vulnerabilities → missing error handling → suboptimal configuration. Within a level, prioritize by blast radius (internet-facing > internal service > local-only).
