# TASK: Perform a red team security audit of a Go project, adopting an adversarial mindset to discover exploitable vulnerabilities by simulating real attack scenarios against the project's actual deployment model and trust boundaries.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the red team audit report
2. **`GAPS.md`** — exploitable attack surface gaps and defense recommendations

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Threat Modeling
1. Read the project README to understand its purpose, deployment model (CLI tool, web service, library, daemon), and target users.
2. Determine the **trust boundary map**:
   - What is the outermost trust boundary? (internet, local network, same machine, same process)
   - Where does untrusted input enter the system? (HTTP requests, CLI args, file content, environment variables, IPC, stdin)
   - What sensitive assets does the system handle? (credentials, PII, financial data, configuration, other users' data)
   - What privilege level does the system run at? (root, user, container, sandboxed)
3. Examine `go.mod` for the attack surface introduced by dependencies.
4. List packages (`go list ./...`) and classify each by trust level:
   - **Exposed**: Directly processes untrusted input
   - **Internal**: Processes data from Exposed packages
   - **Trusted**: Only handles internal/validated data
5. Identify the project's security assumptions — what does it assume about its environment, callers, and input?
6. Build an **attack tree** for each entry point: what can an attacker achieve if they control the input?

### Phase 1: Online Research
Use web search to build context for attack planning:
1. Search for the project on GitHub — read security advisories, CVE reports, and issues mentioning "security", "vulnerability", "exploit", or "bypass".
2. Research dependencies for known CVEs that are reachable from this project's code paths.
3. Study attack techniques relevant to the project's domain (e.g., deserialization attacks for APIs, path traversal for file-serving tools, timing attacks for auth systems).
4. Check for public exploit code or security research targeting similar projects.

Keep research brief (≤10 minutes). Record only findings that inform concrete attack scenarios.

### Phase 2: Baseline
```bash
set -o pipefail
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages > /tmp/redteam-audit-metrics.json
go-stats-generator analyze . --skip-tests
go vet ./... 2>&1 | tee /tmp/redteam-vet-results.txt
```
Delete temporary files when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Red Team Audit

#### 3a. Input Weaponization
For every untrusted input entry point, attempt to construct a malicious payload:

**Command injection chains:**
- [ ] Trace every user-controlled string to `os/exec.Command`, `syscall.Exec`, or shell invocations. Construct a payload that escapes intended argument boundaries (e.g., `; rm -rf /`, backtick injection, `$()` substitution).
- [ ] Test for argument injection — even without shell metacharacters, can an attacker inject flags that change behavior? (e.g., `--admin=true` via a filename argument).

**Path traversal exploitation:**
- [ ] Trace every user-controlled path component to file system operations. Construct a traversal payload (`../../etc/passwd`, `..%2f..%2f`, null byte injection, symlink chains).
- [ ] Test for Zip Slip — can a malicious archive extract files outside the intended directory?
- [ ] Test for path truncation — can overly long paths bypass validation?

**Injection into structured formats:**
- [ ] SQL: Construct payloads that break out of parameterization (if string concatenation is used) or exploit identifier injection (table/column names from user input).
- [ ] Template: Construct payloads that execute arbitrary code via template injection.
- [ ] LDAP, XML, JSON, YAML: Construct payloads that exploit parser-specific behaviors (e.g., YAML anchors causing exponential expansion, XML entity expansion).

**Deserialization attacks:**
- [ ] Identify every deserialization boundary (`json.Unmarshal`, `gob.Decode`, `xml.Unmarshal`, protobuf, custom decoders). Can malicious payloads cause excessive memory allocation, type confusion, or logic bypass?

#### 3b. Authentication and Authorization Bypass
For every access-controlled operation:

- [ ] **Direct object reference**: Can an authenticated user access resources belonging to other users by manipulating IDs, paths, or query parameters?
- [ ] **Privilege escalation**: Can a low-privilege user invoke admin-only operations by bypassing middleware, manipulating tokens, or exploiting race conditions?
- [ ] **Token manipulation**: Can JWT claims be tampered with? Is the `alg: none` attack possible? Can token expiration be bypassed?
- [ ] **Session fixation/hijacking**: Can an attacker set or predict session tokens?
- [ ] **Route bypass**: Are there unprotected routes that should require authentication? Check for middleware ordering issues, wildcard route conflicts, or debug endpoints left enabled.
- [ ] **CORS exploitation**: Can a malicious website make authenticated cross-origin requests?

#### 3c. Cryptographic Attacks
- [ ] **Key/nonce reuse**: Can an attacker trigger nonce reuse by controlling input timing or exploiting deterministic nonce generation?
- [ ] **Downgrade attacks**: Can an attacker force the use of weaker algorithms by manipulating negotiation (TLS version, cipher suite, hash algorithm)?
- [ ] **Timing side-channels**: Are secret comparisons (passwords, tokens, HMAC) done in constant time? Can response timing reveal information?
- [ ] **Entropy attacks**: Is randomness sourced from `crypto/rand`? Can an attacker predict random values used for security-critical operations?

#### 3d. Denial of Service
- [ ] **Algorithmic complexity attacks**: Can user input trigger O(n²) or worse behavior in parsing, sorting, or matching? (e.g., regex backtracking, hash collision attacks on map keys).
- [ ] **Resource exhaustion**: Can an attacker exhaust file descriptors, goroutines, memory, or database connections by sending crafted requests?
- [ ] **Amplification attacks**: Can a small request trigger a disproportionately large response or internal operation?
- [ ] **Infinite loops and hangs**: Can malicious input cause the program to loop forever or block indefinitely?
- [ ] **Panic-based DoS**: Can crafted input trigger a panic that crashes the entire process (especially in HTTP handlers without recovery middleware)?

#### 3e. Supply Chain and Dependency Exploitation
- [ ] **Vulnerable dependencies**: Are there known CVEs in dependencies that are reachable from this project's code?
- [ ] **Dependency confusion**: Could an attacker publish a package with the same name in a public registry to hijack imports?
- [ ] **Build-time attacks**: Are there `//go:generate` commands, Makefile targets, or build scripts that execute external code?
- [ ] **Embedded secrets**: Does `//go:embed` include files that should not be distributed (private keys, credentials, internal configs)?

#### 3f. Data Exfiltration and Information Disclosure
- [ ] **Error message leakage**: Do error responses reveal internal paths, stack traces, SQL queries, or configuration details?
- [ ] **Log injection**: Can an attacker inject newlines or escape sequences into log files to forge log entries or exploit log analysis tools?
- [ ] **Debug endpoints**: Are pprof, debug, or metrics endpoints accessible without authentication?
- [ ] **Timing oracles**: Can response timing differences reveal whether a user exists, a password prefix is correct, or a resource is accessible?

#### 3g. False-Positive Prevention (MANDATORY)
Before recording ANY finding, apply these checks:

1. **Prove exploitability**: Construct a concrete attack scenario with specific input, expected behavior, and attacker outcome. "This could theoretically be exploited" is not sufficient — describe HOW.
2. **Respect the deployment model**: A "SQL injection" in an offline CLI tool the operator runs on their own machine is a different severity than one in a public-facing API. Evaluate against the actual threat model.
3. **Verify reachability**: Confirm the vulnerable code is actually reachable from an attacker-controlled input with no intervening validation.
4. **Check for defense-in-depth**: Even if one control is weak, other controls may prevent exploitation (e.g., WAF, container sandboxing, network segmentation). Note mitigating factors but still report if the code-level vulnerability exists.
5. **Read security comments**: If `// nosec`, `//nolint:`, or security-justification comments exist, treat them as acknowledged decisions — note them but do not report as new findings.
6. **Assess real-world impact**: What does a successful attack actually achieve? "Read arbitrary files" on a system with no sensitive files is LOW. "Execute arbitrary code" on a production server is CRITICAL.

**Rule**: Every finding must include a concrete attack scenario with specific malicious input. If you cannot construct one, it is not a red team finding.

### Phase 4: Report
Generate **`AUDIT.md`**:
```markdown
# RED TEAM AUDIT — [date]

## Threat Model
[Deployment model, trust boundaries, attacker profile, sensitive assets]

## Attack Surface Map
| Entry Point | Trust Level | Input Type | Downstream Sinks | Risk |
|-------------|-------------|------------|-------------------|------|
| [endpoint/function] | Untrusted/Internal/Trusted | [type] | [exec/SQL/file/template] | CRITICAL/HIGH/MEDIUM/LOW |

## Attack Scenarios
### CRITICAL
- [ ] **[Attack Name]** — [file:line] — **Vector:** [how the attacker delivers the payload] — **Payload:** [specific malicious input] — **Impact:** [what the attacker achieves] — **Remediation:** [specific defense]
### HIGH / MEDIUM / LOW
- [ ] ...

## Exploitation Chains
[Multi-step attack scenarios that combine individual findings]

## False Positives Considered and Rejected
| Candidate Attack | Reason Rejected |
|-----------------|----------------|
| [description] | [why it is not exploitable in this context] |
```

Generate **`GAPS.md`**:
```markdown
# Security Defense Gaps — [date]

## [Gap Title]
- **Attack Scenario**: [how an attacker would exploit this gap]
- **Current Defenses**: [what protection exists (if any)]
- **Missing Defense**: [what security control is absent]
- **Remediation**: [specific defense to implement]
- **Priority**: [based on exploitability and impact]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | Demonstrated remote code execution, authentication bypass, or data exfiltration with a concrete exploit path |
| HIGH | Demonstrated injection with data access, privilege escalation to admin, or DoS that crashes the process |
| MEDIUM | Information disclosure, CSRF, or DoS that degrades but does not crash the service |
| LOW | Defense-in-depth improvements, minor information leakage, or attacks requiring unlikely preconditions |

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Specific defense**: State exactly what control to add — input validation, parameterized query, rate limit, authentication check. Do not recommend "consider improving security."
2. **Defense-in-depth**: Recommend layered defenses where appropriate (input validation AND parameterized queries, not just one).
3. **Verifiable**: Include a test case that demonstrates the attack is blocked after remediation.
4. **Minimal scope**: Fix the vulnerability without restructuring unrelated code.

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Use `go-stats-generator` metrics to identify high-complexity functions on attack paths (complexity correlates with vulnerability density).
- Every finding must include a concrete attack scenario with specific malicious input.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Evaluate the code against its **actual deployment model and threat boundaries**, not hypothetical worst-case deployments.
- Apply the Phase 3g false-positive prevention checks to every candidate finding before including it.

## Tiebreaker
Prioritize: remote exploitation → authentication bypass → data exfiltration → privilege escalation → denial of service → information disclosure. Within a level, prioritize by attack complexity (easiest to exploit first).
