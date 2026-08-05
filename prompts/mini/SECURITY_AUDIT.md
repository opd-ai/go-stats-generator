# TASK: Perform a focused security audit of Go code, identifying injection vulnerabilities, auth flaws, crypto misuse, input validation gaps, and secrets exposure.

## Output
Write two files in repo root: `AUDIT.md` (security audit report) and `GAPS.md` (security posture gaps).

## Workflow

### Phase 0: Security Surface Inventory
1. Read README for deployment model and security claims.
2. Check `go.mod` for security-relevant deps (crypto, net/http, DB drivers).
3. Scan for: `os/exec.Command` (cmd injection), `database/sql` (SQL injection), `filepath.Join` with user input (path traversal), `io.Copy` without limits (DoS), hardcoded secrets, `html/template` vs `text/template` (XSS).

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages > tmp/security-metrics.json
go vet ./...
```

### Phase 2: Audit Checklist
- [ ] **Command Injection**: All `os/exec` calls validate/sanitize input
- [ ] **SQL Injection**: All DB queries use parameterized statements
- [ ] **Path Traversal**: User input in file paths validated (no `..`, absolute paths)
- [ ] **XSS**: Web handlers use `html/template`, not `text/template`
- [ ] **Secrets**: No hardcoded tokens/passwords/API keys in source
- [ ] **Crypto**: Strong algorithms (AES-GCM, bcrypt, TLS 1.2+), no custom crypto
- [ ] **Auth**: All endpoints authenticated; authorization checks enforced
- [ ] **Dependencies**: Run `govulncheck ./...` — no known CVEs

### Phase 3: Report
**AUDIT.md**: Project profile, surface inventory table, dependency check, findings (CRITICAL/HIGH/MEDIUM/LOW) with file:line, evidence, impact, remediation.

**GAPS.md**: Per gap — Stated Goal, Current State, Risk, Closing the Gap.

## Constraints
- Report ONLY — no code changes
- Every finding: file:line, concrete data flow, actionable remediation
- No speculative findings — must demonstrate exploitability
- Use unchecked `- [ ]` checkboxes