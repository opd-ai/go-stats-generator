# TASK: Perform a focused security audit of Go code, identifying injection vulnerabilities, authentication/authorization flaws, cryptographic misuse, input validation gaps, and secrets exposure while rigorously preventing false positives.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the security audit report
2. **`GAPS.md`** — gaps in security posture relative to the project's stated goals

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project's Security Surface
1. Read the project README to understand its purpose, users, and any claims about security, authentication, data handling, or compliance.
2. Examine `go.mod` for module path, Go version, and security-relevant dependencies (e.g., `crypto/*`, `golang.org/x/crypto`, `net/http`, database drivers, template engines, JWT libraries).
3. List packages (`go list ./...`) and identify which packages handle:
   - User input (HTTP handlers, CLI arguments, file uploads, API endpoints)
   - Authentication and authorization (tokens, sessions, RBAC, middleware)
   - Data storage (databases, files, caches)
   - External communication (HTTP clients, gRPC, WebSockets, email)
   - Cryptographic operations (hashing, encryption, signing, TLS)
   - Process execution (os/exec, syscall)
4. Build a **security surface inventory** by scanning for:
   - `os/exec.Command` and `syscall.Exec` — command injection vectors
   - `database/sql` query construction — SQL injection vectors
   - `html/template` vs `text/template` — XSS vectors
   - `net/http` handler registration — endpoint enumeration
   - `filepath.Join` with user input — path traversal vectors
   - `io.Copy` without size limits — denial of service vectors
   - `crypto/*` and `golang.org/x/crypto` usage — cryptographic implementation
   - `os.Getenv` for secrets — environment-based secret management
   - Hardcoded strings that look like tokens, passwords, API keys, or connection strings
   - `//go:embed` directives that may include sensitive files
5. Map the trust boundaries: where does untrusted input enter the system, and how far does it propagate before validation?
6. Identify the project's security conventions — does it use middleware for auth? Does it validate input at boundaries? Does it use parameterized queries?

### Phase 1: Online Research
Use web search to build context that isn't available in the repository:
1. Search for the project on GitHub — read open issues and security advisories mentioning "vulnerability", "CVE", "injection", "XSS", "auth", or "security" to understand known security concerns.
2. Research key dependencies from `go.mod` for known CVEs, security advisories, or deprecated crypto algorithms.
3. Run `govulncheck ./...` if available, or check https://pkg.go.dev/vuln/ for the project's dependencies.
4. Look up OWASP guidelines relevant to the project's domain (web application, API, CLI tool, library).

Keep research brief (≤10 minutes). Record only findings that are directly relevant to the project's security posture.

### Phase 2: Baseline
```bash
set -o pipefail
mkdir -p tmp
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages > tmp/security-audit-metrics.json
go-stats-generator analyze . --skip-tests
go vet ./... 2>&1 | tee tmp/security-vet-results.txt
```
Delete temporary files when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Security Audit

#### 3a. Injection Vulnerabilities
For every path from untrusted input to a sensitive sink, verify:

**Command injection:**
- [ ] `os/exec.Command` arguments are never constructed from user input without validation.
- [ ] Shell metacharacters are not passed through — prefer argument lists over shell invocation (`Command("cmd", args...)` not `Command("sh", "-c", userInput)`).
- [ ] `syscall.Exec` and `os.StartProcess` inputs are validated against an allowlist.

**SQL injection:**
- [ ] All SQL queries use parameterized queries with the database driver's parameter placeholders and pass arguments separately, never string concatenation or `fmt.Sprintf`.
- [ ] ORM query builders do not interpolate user input into raw SQL fragments.
- [ ] Table and column names derived from user input are validated against an allowlist (parameterization cannot protect identifiers).

**Template injection:**
- [ ] `html/template` is used for HTML output, not `text/template` (which does not escape).
- [ ] Template names or template content are never derived from user input.
- [ ] `template.HTML()`, `template.JS()`, and `template.CSS()` type conversions are not applied to user-controlled data.

**Path traversal:**
- [ ] File paths derived from user input are constrained to a trusted base directory; do not treat `filepath.Clean` or `filepath.Rel` alone as sufficient validation.
- [ ] Code that calls `os.Open`, `os.Create`, `os.ReadFile`, or `os.WriteFile` with user-derived paths first joins them to a trusted root and verifies the resolved path remains within that root.
- [ ] Where the file system is untrusted or symlinks may be present, `filepath.EvalSymlinks` is used as part of the trusted-root containment check, not as a standalone safeguard.
- [ ] Archive extraction (zip, tar) validates each entry path against the intended extraction directory to prevent writes outside the target directory (Zip Slip).

**SSRF (Server-Side Request Forgery):**
- [ ] URLs constructed from user input are validated against an allowlist of permitted hosts/schemes.
- [ ] HTTP redirects from user-controlled URLs are limited or disabled to prevent redirect-based SSRF.
- [ ] DNS rebinding is considered for long-lived connections to user-specified hosts.

#### 3b. Authentication and Authorization
For every access-controlled operation, verify:

- [ ] Authentication middleware is applied to all routes that require it — no route is accidentally unprotected.
- [ ] Authorization checks verify the authenticated user has permission for the specific resource, not just that they are authenticated.
- [ ] JWT validation checks signature, expiration (`exp`), issuer (`iss`), and audience (`aud`) — not just signature.
- [ ] Session tokens have sufficient entropy (≥128 bits from `crypto/rand`).
- [ ] Password comparison uses `subtle.ConstantTimeCompare` or `bcrypt.CompareHashAndPassword` — not `==`.
- [ ] Failed authentication does not reveal whether the username or password was wrong (timing-safe).
- [ ] Rate limiting or account lockout is applied to authentication endpoints.
- [ ] CORS configuration does not use `Access-Control-Allow-Origin: *` with credentials.

#### 3c. Cryptographic Misuse
For every cryptographic operation, verify:

- [ ] Random number generation uses `crypto/rand`, not `math/rand` (which is deterministic and predictable).
- [ ] Password hashing uses `bcrypt`, `scrypt`, or `argon2` — not MD5, SHA-1, SHA-256, or custom schemes.
- [ ] Encryption uses authenticated encryption (AES-GCM, XChaCha20-Poly1305) — not AES-CBC without HMAC.
- [ ] Nonces/IVs are generated from `crypto/rand` and never reused with the same key.
- [ ] RSA key sizes are ≥2048 bits; ECDSA uses P-256 or stronger.
- [ ] TLS configuration specifies `MinVersion: tls.VersionTLS12` and does not include known-weak cipher suites.
- [ ] `InsecureSkipVerify: true` is not used in production TLS configurations.
- [ ] HMAC comparison uses `hmac.Equal`, not `bytes.Equal` or `==` (timing side-channel).

#### 3d. Input Validation and Data Handling
- [ ] All external input (HTTP parameters, headers, body, CLI arguments, file content, environment variables) is validated before use.
- [ ] Input size limits are enforced: `http.MaxBytesReader` for request bodies, `io.LimitReader` for streams.
- [ ] Integer overflow is considered for user-provided sizes or counts used in `make()` or loop bounds.
- [ ] Deserialization of untrusted data (JSON, XML, YAML, gob, protobuf) uses strict decoders that reject unknown fields where appropriate.
- [ ] Error messages do not leak internal paths, stack traces, SQL queries, or configuration details to users.
- [ ] Sensitive data (passwords, tokens, PII) is not logged, even at debug level.
- [ ] Sensitive fields in structs use `json:"-"` or custom marshalers to prevent accidental serialization.

#### 3e. Secrets and Configuration
- [ ] No hardcoded secrets (API keys, passwords, tokens, private keys) exist in source code.
- [ ] Connection strings and credentials are loaded from environment variables or secret management systems.
- [ ] `.gitignore` excludes configuration files that may contain secrets (`.env`, `*.pem`, `*.key`).
- [ ] `//go:embed` directives do not include files containing secrets.
- [ ] Default passwords and tokens in example configurations are clearly marked as non-production values.

#### 3f. False-Positive Prevention (MANDATORY)
Before recording ANY finding, apply these checks:

1. **Trace the data flow**: Confirm that untrusted input actually reaches the vulnerable sink. If all callers pass compile-time constants or validated data, it is not exploitable.
2. **Check for upstream validation**: A seemingly unvalidated input may be validated by middleware, a wrapper function, or a type system constraint earlier in the call chain. Trace upward.
3. **Verify the threat model**: A "SQL injection" in a CLI tool that only the operator runs is a different severity than one in a web-facing API. Evaluate against the project's actual deployment model.
4. **Read surrounding comments**: If a comment explicitly acknowledges a security decision (e.g., `// nosec`, `// safe: input is from trusted config`, `//nolint:`, or a TODO tracking a known issue), treat it as an acknowledged pattern — do not report it as a new finding.
5. **Check dependency context**: A vulnerable function in a dependency may not be called by this project, or may be called with safe inputs. Verify the actual call site.
6. **Assess exploitability**: A theoretical vulnerability with no practical exploit path (e.g., requires physical access to the server) should be downgraded, not reported as CRITICAL.

**Rule**: If you cannot demonstrate a concrete data flow from untrusted input to a vulnerable sink, do NOT report it as an injection vulnerability. Speculative findings waste remediation effort and erode trust in the audit.

### Phase 4: Report
Generate **`AUDIT.md`**:
```markdown
# SECURITY AUDIT — [date]

## Project Security Profile
[Summary: deployment model (web app, CLI, library), trust boundaries, authentication model, data sensitivity, stated security goals]

## Security Surface Inventory
| Package | HTTP Handlers | DB Queries | Exec Calls | File I/O | Crypto | Auth |
|---------|--------------|------------|------------|----------|--------|------|
| [pkg]   | N            | N          | N          | N        | N      | ✅/❌ |

## Dependency Vulnerability Check
[Summary of govulncheck or manual CVE research results]

## Findings
### CRITICAL
- [ ] [Finding] — [file:line] — [evidence: data flow from untrusted input to sink] — [impact: what an attacker can do] — **Remediation:** [specific fix]
### HIGH / MEDIUM / LOW
- [ ] ...

## False Positives Considered and Rejected
| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| [description] | [why it is not exploitable in this context] |
```

Generate **`GAPS.md`**:
```markdown
# Security Gaps — [date]

## [Gap Title]
- **Stated Goal**: [what the project claims about security/authentication/data protection]
- **Current State**: [what security controls exist]
- **Risk**: [what an attacker could exploit]
- **Closing the Gap**: [specific security controls to add]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | Exploitable injection (SQL, command, template) with confirmed data flow, authentication bypass, or secret exposure in source code |
| HIGH | Missing input validation on an external-facing endpoint, weak cryptographic algorithm on a security-critical path, or missing authorization checks |
| MEDIUM | Insecure defaults (e.g., `InsecureSkipVerify`), overly permissive CORS, or sensitive data in logs |
| LOW | Missing security headers, informational exposure in error messages, or defense-in-depth recommendations |

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Concrete fix**: State exactly what to change — the specific function, parameter, or configuration. Do not recommend "consider validating input."
2. **Respect project idioms**: If the project uses a specific auth middleware or validation library, recommend fixes using those tools.
3. **Verifiable**: Include a validation approach (e.g., `govulncheck ./...`, `go vet ./...`, or a specific test case demonstrating the fix).
4. **Minimal scope**: Fix the security issue without restructuring unrelated code.

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Use `go-stats-generator` metrics as supporting evidence for complexity and function length.
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Every finding must include a concrete data flow demonstrating exploitability — no speculative findings.
- Evaluate the code against its **own stated goals** and threat model, not arbitrary external standards.
- Apply the Phase 3f false-positive prevention checks to every candidate finding before including it.

## Tiebreaker
Prioritize: exploitable injection → authentication/authorization bypass → secret exposure → cryptographic weakness → input validation gaps → defense-in-depth. Within a level, prioritize by attack surface exposure (internet-facing > internal > CLI-only).
