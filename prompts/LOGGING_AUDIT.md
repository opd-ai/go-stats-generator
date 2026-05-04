# TASK: Perform a focused logging and observability audit of a Go project, evaluating log quality, structured logging consistency, metrics instrumentation, tracing coverage, and operational readiness while rigorously preventing false positives.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the logging and observability audit report
2. **`GAPS.md`** — gaps in observability relative to the project's operational needs

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project's Observability Model
1. Read the project README for claims about logging, monitoring, debugging, or operational readiness.
2. Examine `go.mod` for observability-related dependencies:
   - Logging: `log/slog`, `go.uber.org/zap`, `github.com/sirupsen/logrus`, `github.com/rs/zerolog`
   - Metrics: `github.com/prometheus/client_golang`, `go.opentelemetry.io/otel/metric`
   - Tracing: `go.opentelemetry.io/otel/trace`, `github.com/opentracing/opentracing-go`
   - Error tracking: `github.com/getsentry/sentry-go`
3. List packages (`go list ./...`) and identify which packages produce operational output (logs, metrics, traces).
4. Build a **logging inventory** by scanning for:
   - `log.*`, `slog.*`, `zap.*`, `logrus.*`, `zerolog.*` calls
   - `fmt.Print*`, `fmt.Fprint*` to stdout/stderr (unstructured output mixed with logs)
   - `os.Stderr.Write`, `os.Stdout.Write` direct writes
   - Metric registration and instrumentation calls
   - Trace span creation and attribute setting
   - `pprof` registration and debug endpoint setup
5. Identify the project's logging conventions:
   - Structured vs unstructured logging
   - Log levels used and their meaning in this project
   - Log field naming conventions
   - Error logging vs error returning (which does the project prefer?)
6. Determine the deployment model: CLI (logs to stderr), server (logs to collector), library (should not log).

### Phase 1: Online Research
Use web search to build context:
1. Search for the project on GitHub — read issues mentioning "log", "debug", "verbose", "quiet", or "monitoring" to understand operational pain points.
2. Research the project's logging library for best practices and common misuse patterns.
3. Check if the project's domain has specific observability requirements (e.g., SLO monitoring for web services, audit logging for security tools).

Keep research brief (≤10 minutes). Record only findings relevant to operational observability.

### Phase 2: Baseline
```bash
set -o pipefail
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages > tmp/logging-audit-metrics.json
go-stats-generator analyze . --skip-tests
go vet ./... 2>&1 | tee tmp/logging-vet-results.txt
```
Delete temporary files when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Logging and Observability Audit

#### 3a. Logging Consistency
- [ ] A single logging library/approach is used throughout the project — not a mix of `log`, `fmt.Println`, `slog`, and a third-party logger.
- [ ] Log levels are used consistently: DEBUG for development detail, INFO for operational events, WARN for recoverable issues, ERROR for failures.
- [ ] ERROR log entries correspond to actual errors — not informational messages logged at ERROR level.
- [ ] Log messages are actionable — they describe what happened, what was expected, and what the operator should do.
- [ ] Log messages do not duplicate error returns — if a function logs an error AND returns it, the error will be logged multiple times as it propagates.
- [ ] Log field names are consistent across the codebase (e.g., always `"error"` not sometimes `"err"`, `"error"`, `"failure"`).

#### 3b. Structured Logging Quality
If the project uses structured logging:

- [ ] All log calls use structured fields, not string interpolation: `slog.Error("failed", "err", err)` not `slog.Error(fmt.Sprintf("failed: %v", err))`.
- [ ] Field names follow a consistent naming convention (snake_case, camelCase, or as defined by the project).
- [ ] Common fields (request ID, user ID, operation name) are included consistently via context or logger enrichment.
- [ ] Numeric values are logged as numbers, not strings (for aggregation and alerting).
- [ ] Durations are logged in a consistent unit (milliseconds, seconds) with the unit in the field name.
- [ ] Error values are logged with their type information preserved, not just `.Error()` string.

#### 3c. Sensitive Data Protection
- [ ] Passwords, tokens, API keys, and secrets are NEVER logged, even at DEBUG level.
- [ ] PII (personal identifiable information) is redacted or excluded from logs.
- [ ] Request/response bodies are not logged verbatim if they may contain sensitive data.
- [ ] Database connection strings in logs are sanitized (password removed).
- [ ] HTTP headers like `Authorization`, `Cookie`, and `X-API-Key` are not logged.
- [ ] Error messages from failed authentication attempts do not reveal whether the username or password was wrong.

#### 3d. Operational Readiness
- [ ] The application logs a startup message with version, configuration summary, and listening addresses.
- [ ] The application logs a clean shutdown message when terminated gracefully.
- [ ] Error conditions that require operator intervention are logged at ERROR or WARN with sufficient context.
- [ ] Recoverable errors (retries, fallbacks) are logged at WARN, not ERROR — ERROR should indicate a need for action.
- [ ] Log verbosity is configurable (via flag, environment variable, or config file) without code changes.
- [ ] Log output format is configurable (JSON for production/machine parsing, text for development/human reading).

#### 3e. Metrics and Instrumentation
If the project is a server or long-running process:

- [ ] Key business metrics are instrumented (requests handled, items processed, errors encountered).
- [ ] Latency is measured for critical operations (request handling, database queries, external API calls).
- [ ] Resource usage is observable (goroutine count, connection pool size, queue depth).
- [ ] Metrics have appropriate labels/dimensions without high cardinality (no per-user-ID labels).
- [ ] Health check endpoints exist and return meaningful status (not just 200 OK).
- [ ] Metrics are exposed in a standard format (Prometheus, OpenTelemetry) if applicable.

#### 3f. Debug and Troubleshooting Support
- [ ] `pprof` is available (or can be enabled) for production debugging of CPU, memory, goroutine, and mutex profiles.
- [ ] Error messages include sufficient context to diagnose the issue without access to the source code.
- [ ] Request IDs or correlation IDs are propagated through the call chain for distributed tracing.
- [ ] `debug` or `verbose` mode provides additional output for troubleshooting without overwhelming normal operation.
- [ ] Stack traces are available for unexpected panics (via recovery middleware or similar).

#### 3g. Library Logging Behavior
If the project is a library (imported by other applications):

- [ ] The library does NOT log directly — it returns errors and lets the caller decide how to log.
- [ ] If logging is necessary, the library accepts a logger interface from the caller (dependency injection).
- [ ] The library does not use `log.Fatal` or `os.Exit` — these kill the caller's process.
- [ ] The library does not write to stdout/stderr directly — these may conflict with the caller's output.

#### 3h. False-Positive Prevention (MANDATORY)
Before recording ANY finding, apply these checks:

1. **Respect the deployment model**: A CLI tool does not need structured JSON logging, metrics, or tracing. Evaluate against the project's actual operational context.
2. **Check the project's conventions**: If the project consistently uses `log.Printf`, switching to `slog` is a recommendation, not a finding. Report it as LOW if the current approach causes real problems.
3. **Verify the log matters**: A missing log in a cold path that never fails is not a finding. Focus on operational paths where logging is critical for troubleshooting.
4. **Read surrounding context**: If a `// TODO: add logging` comment exists, it is a tracked item, not an undiscovered gap.
5. **Assess the audience**: Logs for a developer tool are different from logs for a production web service. Calibrate expectations.
6. **Check for alternative observability**: Missing logs may be compensated by good error messages, metrics, or tracing. Evaluate the overall observability posture.

**Rule**: Not every function needs a log statement. Focus on operational value: would an operator need this information to diagnose a production issue? If not, its absence is not a finding.

### Phase 4: Report
Generate **`AUDIT.md`**:
```markdown
# LOGGING & OBSERVABILITY AUDIT — [date]

## Observability Profile
[Deployment model, logging library, structured/unstructured, log levels, metrics, tracing]

## Logging Inventory
| Package | Log Calls | Structured | fmt.Print* | Log Level Usage | Sensitive Data Risk |
|---------|----------|------------|------------|-----------------|-------------------|
| [pkg] | N | ✅/❌ | N | DEBUG/INFO/WARN/ERROR | ✅ Safe / ⚠️ Review |

## Findings
### CRITICAL
- [ ] [Finding] — [file:line] — [observability issue] — [operational impact] — **Remediation:** [specific fix]
### HIGH / MEDIUM / LOW
- [ ] ...

## False Positives Considered and Rejected
| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| [description] | [why it is not a real observability issue in this context] |
```

Generate **`GAPS.md`**:
```markdown
# Observability Gaps — [date]

## [Gap Title]
- **Operational Need**: [what an operator needs to observe or diagnose]
- **Current State**: [what observability exists]
- **Gap**: [what is missing]
- **Recommendation**: [specific logging, metric, or tracing to add]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | Sensitive data (passwords, tokens, PII) logged in plaintext, or a library calling `log.Fatal`/`os.Exit` that kills the host process |
| HIGH | No error logging on a critical failure path (operator cannot diagnose production issues), inconsistent logging making correlation impossible, or ERROR level used for non-errors causing alert fatigue |
| MEDIUM | Mixed logging approaches (structured and unstructured), missing request/correlation IDs, or non-configurable log verbosity |
| LOW | Minor log message formatting inconsistencies, missing DEBUG-level logging for development, or cosmetic field naming differences |

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Specific change**: State exactly what log call to add, modify, or remove, and what fields to include. Do not recommend "improve logging."
2. **Respect project idioms**: Use the project's existing logging library and conventions.
3. **Verifiable**: Include a way to verify the fix (e.g., grep for the new log field, check log output format).
4. **Operationally motivated**: Explain what operational problem the change solves.

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Use `go-stats-generator` metrics as supporting evidence for identifying critical paths that need observability.
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Evaluate observability against the project's **actual deployment model and operational needs**, not arbitrary logging standards.
- Apply the Phase 3h false-positive prevention checks to every candidate finding before including it.

## Tiebreaker
Prioritize: sensitive data exposure in logs → missing error logging on critical paths → logging consistency → metrics gaps → tracing gaps → cosmetic issues. Within a level, prioritize by operational impact (how much harder is troubleshooting without this?).
