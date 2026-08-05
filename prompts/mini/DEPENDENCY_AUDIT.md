# TASK: Perform a dependency health and supply chain audit evaluating freshness, vulnerabilities, licenses, maintenance status, and supply chain risk.

## Output
Write two files in repo root: `AUDIT.md` (dependency audit report) and `GAPS.md` (dependency management gaps).

## Workflow

### Phase 0: Dependency Map
1. Read README for dependency claims (minimalism, zero-dep).
2. Parse `go.mod`: module, Go version, direct/indirect deps, `replace`/`retract` directives.
3. Run `go mod graph` for tree: diamond deps, deep chains, large transitive trees.
4. Classify direct deps: Core, Testing, Build, Optional.
5. Check for `vendor/` directory.

### Phase 1: Research & Baseline
```bash
go mod tidy -diff && go mod verify
govulncheck ./... 2>&1 | tee tmp/dep-vuln.txt
# For each direct dep: check last commit, CVEs, license, popularity
```

### Phase 2: Audit Checklist
- [ ] **Vulnerabilities**: `govulncheck` clean — no reachable CVEs. Verify each reported vuln's call path.
- [ ] **Maintenance**: No abandoned deps on critical paths. Last commit <2 years for core deps.
- [ ] **Licenses**: All deps compatible with project license (MIT/BSD/Apache OK; GPL/LGPL need review).
- [ ] **Supply Chain**: `go.sum` committed. No `replace` to untrusted sources. No `//go:embed` of sensitive files.
- [ ] **Freshness**: Core deps not >2 major versions behind. No stale deps with better alternatives.
- [ ] **Minimality**: No unused deps (`go mod tidy` clean). No heavy deps where stdlib suffices.

### Phase 3: Report
**AUDIT.md**: Dependency overview table (dep, version, latest, last commit, license, CVEs, status), vulnerability assessment, findings (CRITICAL/HIGH/MEDIUM/LOW) with dep:version, reachability, impact, remediation.

**GAPS.md**: Per gap — Current State, Risk, Recommendation, Priority.

## Constraints
- Report ONLY — no code changes
- Every finding: specific dep:version, quantified risk, actionable upgrade/replacement
- No "this is old" without specific risk
- Use unchecked `- [ ]` checkboxes