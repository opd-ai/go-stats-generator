# TASK: Perform a dependency health and supply chain audit of a Go project, evaluating dependency freshness, vulnerability exposure, license compliance, maintenance status, and supply chain risk while rigorously preventing false positives.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the dependency audit report
2. **`GAPS.md`** — gaps in dependency management and supply chain security

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Map the Dependency Landscape
1. Read the project README for any claims about dependency minimalism, zero-dependency operation, or specific dependency choices.
2. Parse `go.mod` completely:
   - Module path and Go version
   - Direct dependencies (count and categorize)
   - Indirect dependencies (count and note large transitive trees)
   - `replace` directives (temporary development aids or permanent forks?)
   - `retract` directives (any retracted versions of this module?)
3. Parse `go.sum` to understand the full dependency tree depth.
4. Run `go mod graph` to visualize the dependency tree and identify:
   - Diamond dependencies (same module required at different versions)
   - Deep dependency chains (dependencies of dependencies of dependencies)
   - Large transitive dependency trees (a single import pulling dozens of packages)
5. Classify each direct dependency by role:
   - **Core**: Required for the project's primary functionality
   - **Testing**: Used only in tests (should be in `_test.go` imports)
   - **Build**: Used for code generation, linting, or CI
   - **Optional**: Used for non-essential features
6. Identify any vendored dependencies (`vendor/` directory) and whether `go mod vendor` is the project's strategy.

### Phase 1: Online Research
Use web search to assess dependency health:
1. For each direct dependency, check:
   - **Maintenance status**: Last commit date, open issue count, maintainer activity, archived/deprecated status
   - **Vulnerability history**: Known CVEs, security advisories, `govulncheck` results
   - **License**: License type and compatibility with the project's license
   - **Popularity and trust**: Star count, fork count, organizational backing, security audit history
2. Run `govulncheck ./...` if available to check for known vulnerabilities in used code paths.
3. Check https://pkg.go.dev/vuln/ for advisories affecting dependencies.
4. Search for dependency-related issues in the project's GitHub issues.

Keep research brief (≤15 minutes for dependency research). Record findings systematically.

### Phase 2: Baseline
```bash
set -o pipefail
go-stats-generator analyze . --skip-tests --format json --sections packages > tmp/dep-audit-metrics.json
go-stats-generator analyze . --skip-tests
go mod tidy -diff 2>&1 | tee tmp/dep-tidy-results.txt || true
go mod verify 2>&1 | tee tmp/dep-verify-results.txt
govulncheck ./... 2>&1 | tee tmp/dep-vulncheck-results.txt || true
```
Delete temporary files when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Dependency Audit

#### 3a. Vulnerability Assessment
- [ ] Run `govulncheck ./...` and assess each reported vulnerability for actual reachability (govulncheck already checks this, but verify the call path).
- [ ] Check each direct dependency against https://pkg.go.dev/vuln/ and the GitHub Advisory Database.
- [ ] Verify that the Go version itself is not affected by known vulnerabilities (check Go release notes).
- [ ] Identify any dependencies pinned to versions with known security patches available.
- [ ] Check for dependencies that use `unsafe`, `cgo`, or `reflect` extensively (higher risk surface).

#### 3b. Maintenance Health
For each direct dependency, assess:

- [ ] **Active maintenance**: Last commit within 12 months, issues being triaged, PRs being reviewed.
- [ ] **Abandoned risk**: No commits in >2 years, archived repository, maintainer inactive.
- [ ] **Bus factor**: Single maintainer vs team/organization. Single-maintainer dependencies are higher risk.
- [ ] **Go version compatibility**: Dependency's `go.mod` specifies a Go version compatible with this project's requirements.
- [ ] **Deprecation**: Check for deprecation notices in the dependency's README, GoDoc, or GitHub description.
- [ ] **Fork availability**: For high-risk dependencies, note whether maintained forks exist.

#### 3c. License Compliance
- [ ] All direct dependencies have licenses compatible with the project's license.
- [ ] No dependency uses a copyleft license (GPL, AGPL) that would impose obligations on the project unless the project itself uses the same license.
- [ ] All dependencies have a license file — unlicensed code is legally risky.
- [ ] License changes between versions are noted (some projects change licenses between major versions).
- [ ] Transitive dependencies do not introduce incompatible licenses.

#### 3d. Dependency Minimalism and Necessity
- [ ] Each direct dependency is actually used — `go mod tidy` does not remove any dependencies.
- [ ] No dependency is imported for a single trivial function that could be implemented in a few lines (e.g., importing a library just for `StringContains`).
- [ ] Standard library alternatives are not available for what the dependency provides.
- [ ] Dependencies are not duplicating functionality — two different logging libraries, two different HTTP routers, etc.
- [ ] Test-only dependencies are imported only in `_test.go` files, not in production code.

#### 3e. Supply Chain Security
- [ ] `go.sum` is committed and not in `.gitignore` — it provides integrity verification.
- [ ] `go mod verify` passes — all downloaded modules match their expected hashes.
- [ ] No `replace` directives point to local paths (development leftover) or suspicious remote repositories.
- [ ] Dependencies are imported from canonical sources (not forks, mirrors, or vanity URLs that could be hijacked).
- [ ] No `go:generate` commands download or execute code from the internet during build.
- [ ] CI/CD pipelines use `GONOSUMCHECK` or `GONOSUMDB` only with documented justification.
- [ ] The project does not use `go get` in CI for fetching tools (use `go install tool@version` with pinned versions).

#### 3f. Dependency Version Management
- [ ] Dependencies are updated regularly — no dependency is more than 2 major versions behind.
- [ ] Security patches are applied promptly — no dependency has a newer patch version with security fixes.
- [ ] Major version upgrades are evaluated for breaking changes before adoption.
- [ ] `go.mod` does not use commit hashes where tagged releases are available (hashes are harder to audit).
- [ ] Diamond dependency conflicts are resolved (different parts of the tree requiring incompatible versions).

#### 3g. False-Positive Prevention (MANDATORY)
Before recording ANY finding, apply these checks:

1. **Verify reachability**: A CVE in a dependency function that this project never calls is informational, not CRITICAL. Check `govulncheck` output for actual call paths.
2. **Assess the project's context**: A CLI tool with no network exposure has different dependency risk than a web server. Evaluate against the actual deployment model.
3. **Check for intentional choices**: A pinned old version may be intentional due to a breaking change in newer versions. Check for comments in `go.mod` or issues tracking the upgrade.
4. **Verify maintenance claims**: A dependency with no recent commits may be feature-complete and stable (e.g., `errors` package). Not every stable package needs constant updates.
5. **License compatibility depends on usage**: A GPL dependency used only in tests may not trigger copyleft obligations. Check the specific legal requirements.

**Rule**: Dependency findings must be actionable. "This dependency is old" is not a finding unless there is a specific risk (vulnerability, incompatibility, abandonment). Quantify the risk.

### Phase 4: Report
Generate **`AUDIT.md`**:
```markdown
# DEPENDENCY AUDIT — [date]

## Dependency Overview
[Total direct: N, indirect: N, Go version: X, vendored: yes/no]

## Dependency Health Summary
| Dependency | Version | Latest | Last Commit | License | CVEs | Status |
|-----------|---------|--------|-------------|---------|------|--------|
| [module] | v1.2.3 | v1.4.0 | 2024-01 | MIT | 0 | ✅ Healthy / ⚠️ Stale / ❌ Risk |

## Vulnerability Assessment
[govulncheck results, CVE analysis, reachability assessment]

## Findings
### CRITICAL
- [ ] [Finding] — [dependency] — [CVE/issue] — [reachability evidence] — [impact] — **Remediation:** [specific version upgrade or replacement]
### HIGH / MEDIUM / LOW
- [ ] ...

## False Positives Considered and Rejected
| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| [description] | [why it is not an actionable dependency risk] |
```

Generate **`GAPS.md`**:
```markdown
# Dependency Management Gaps — [date]

## [Gap Title]
- **Current State**: [dependency version, maintenance status, or configuration issue]
- **Risk**: [what could go wrong: vulnerability exploitation, build breakage, license violation]
- **Recommendation**: [specific upgrade, replacement, or configuration change]
- **Priority**: [based on risk and effort]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | Known CVE in a reachable code path, dependency with confirmed malicious code, or license violation that creates legal liability |
| HIGH | Abandoned dependency with no maintained fork on a critical path, `go.sum` not committed, or `replace` directive pointing to untrusted source |
| MEDIUM | Dependency >2 major versions behind with security patches available, stale dependency with known alternatives, or minor license concern |
| LOW | Dependency cosmetics (using commit hash instead of tag), minor version lag with no security implications, or dependency that could be replaced by stdlib |

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Specific action**: State exactly which dependency to upgrade, replace, or remove, and to what version. Do not recommend "consider updating dependencies."
2. **Migration guidance**: If replacing a dependency, note the API differences between old and new.
3. **Verifiable**: Include validation commands (`go mod tidy`, `go mod verify`, `govulncheck ./...`, `go test ./...`).
4. **Risk assessment**: Note the risk of the upgrade itself (breaking changes, behavioral differences).

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Use `go-stats-generator` package metrics as supporting evidence for dependency usage patterns.
- Every finding must reference a specific dependency and version.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Evaluate dependencies against the project's **actual risk profile and deployment model**, not theoretical worst cases.
- Apply the Phase 3g false-positive prevention checks to every candidate finding before including it.

## Tiebreaker
Prioritize: exploitable CVEs → abandoned critical dependencies → license violations → supply chain risks → version staleness → cosmetic issues. Within a level, prioritize by the dependency's proximity to the project's critical paths.
