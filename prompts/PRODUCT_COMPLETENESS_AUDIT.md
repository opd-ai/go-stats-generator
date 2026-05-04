# TASK: Perform a product completeness audit of a Go project, systematically verifying that every documented feature, capability, and user-facing promise is fully implemented, functional, and accessible to the target audience.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the product completeness audit report
2. **`GAPS.md`** — gaps between documented product surface and actual implementation

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Extract the Full Product Surface
1. Read the project README end-to-end. Extract every:
   - Feature claim ("supports X", "provides Y", "includes Z")
   - Capability statement ("can handle N concurrent users", "processes files in <M seconds")
   - User-facing promise ("simple API", "zero configuration", "drop-in replacement for X")
   - Audience claim ("designed for DevOps teams", "ideal for large codebases")
   - Integration claim ("works with Docker", "supports CI/CD pipelines")
   - Installation and usage instructions (do they actually work?)
2. Read `--help` output for every CLI command and subcommand. Extract every flag, option, and argument with its documented behavior.
3. Read API documentation, GoDoc comments, and exported function signatures. Each exported symbol is an implicit promise to users.
4. Read CHANGELOG, release notes, and any migration guides for features claimed in past releases.
5. Examine `go.mod` for module path and Go version compatibility claims.
6. Search for configuration file schemas, environment variable documentation, and example configurations.
7. Compile a **Product Surface Checklist** — every verifiable claim becomes a checklist item.

### Phase 1: Online Research
Use web search to build context that isn't available in the repository:
1. Search for the project on GitHub — read open issues labeled "bug", "feature request", or "documentation" to understand what users expect vs what they get.
2. Read community discussions, blog posts, or tutorials mentioning the project to understand how users perceive the product surface.
3. Check if competing tools exist and whether the project claims parity or superiority.
4. Look for user complaints about missing features or broken functionality.

Keep research brief (≤10 minutes). Record only findings that are directly relevant to product completeness.

### Phase 2: Baseline
```bash
set -o pipefail
mkdir -p tmp
go-stats-generator analyze . --skip-tests --format json --sections functions,documentation,packages,patterns,interfaces > tmp/product-audit-metrics.json
go-stats-generator analyze . --skip-tests
go build ./... 2>&1 | tee tmp/product-build-results.txt
go test -race -count=1 ./... 2>&1 | tee tmp/product-test-results.txt
```
Delete temporary files when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Product Completeness Audit

#### 3a. Feature Completeness
For every feature claimed in README, documentation, or help text:

- [ ] The feature's entry point exists in the codebase (CLI command, API endpoint, exported function).
- [ ] The feature produces correct output for the documented happy-path use case.
- [ ] The feature handles documented edge cases (empty input, large input, invalid input).
- [ ] The feature's behavior matches its documentation — no undocumented deviations.
- [ ] The feature is accessible to the target audience without undocumented prerequisites.
- [ ] The feature is not a stub, TODO, or partial implementation that silently returns incomplete results.

#### 3b. CLI and User Interface Completeness
For every CLI command, flag, and option:

- [ ] The command exists and is registered in the command tree.
- [ ] Every documented flag is implemented and has the documented effect.
- [ ] Default values match documentation.
- [ ] `--help` text accurately describes current behavior.
- [ ] Error messages for invalid input are clear and actionable.
- [ ] Exit codes follow documented conventions (0 for success, non-zero for failure).
- [ ] Output formats (JSON, CSV, HTML, etc.) match documented schemas.

#### 3c. API and Library Completeness
For every exported type, function, and method:

- [ ] GoDoc comments exist and accurately describe behavior.
- [ ] Function signatures match documented parameters and return types.
- [ ] Exported error types and sentinel errors are documented.
- [ ] Interface contracts are complete — all documented methods exist with correct signatures.
- [ ] Breaking changes from previous versions are documented.

#### 3d. Installation and Onboarding Completeness
- [ ] Installation instructions work on a clean system (verify `go install`, binary downloads, package managers).
- [ ] Quick-start or getting-started instructions produce the documented output.
- [ ] Example code compiles and runs without modification.
- [ ] Configuration file examples are valid and complete.
- [ ] Environment variable documentation matches actual usage in the code.

#### 3e. Integration and Compatibility Completeness
- [ ] Claimed integrations (Docker, CI/CD, editors, etc.) are actually implemented and functional.
- [ ] Claimed Go version compatibility is correct (`go.mod` version vs claimed version).
- [ ] Claimed OS/architecture support is verified (build tags, platform-specific code).
- [ ] Dependency versions in `go.mod` are compatible with claimed functionality.

#### 3f. False-Positive Prevention (MANDATORY)
Before recording ANY finding, apply these checks:

1. **Verify the claim exists**: Do not invent product claims that the documentation does not make. Only audit against explicit, verifiable statements.
2. **Check for intentional omission**: A feature mentioned as "planned" or "coming soon" is not a completeness gap — it is a documented roadmap item.
3. **Read surrounding context**: A feature described with caveats ("experimental", "beta", "limited support") should be evaluated against those caveats, not against full production expectations.
4. **Check for alternative implementations**: A feature may be implemented differently than expected but still fulfill its documented purpose.
5. **Verify user impact**: A missing feature that no user has requested or noticed may be LOW priority. Check issue tracker activity.

**Rule**: If the documentation does not claim a feature, its absence is not a product completeness finding. Only audit against actual promises.

### Phase 4: Report
Generate **`AUDIT.md`**:
```markdown
# PRODUCT COMPLETENESS AUDIT — [date]

## Product Surface Summary
[What the project claims to be: purpose, target audience, key capabilities]

## Product Completeness Scorecard
| Category | Claimed | Implemented | Partial | Missing |
|----------|---------|-------------|---------|---------|
| CLI Commands | N | N | N | N |
| Features | N | N | N | N |
| Output Formats | N | N | N | N |
| Integrations | N | N | N | N |

## Feature-by-Feature Assessment
| Feature | Status | Evidence |
|---------|--------|----------|
| [Documented feature] | ✅ Complete / ⚠️ Partial / ❌ Missing / 🔇 Stub | [file:line or test evidence] |

## Findings
### CRITICAL
- [ ] [Finding] — [file:line] — [documented claim vs actual behavior] — [user impact] — **Remediation:** [specific fix]
### HIGH / MEDIUM / LOW
- [ ] ...

## False Positives Considered and Rejected
| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| [description] | [why it is not a real completeness gap] |
```

Generate **`GAPS.md`**:
```markdown
# Product Completeness Gaps — [date]

## [Gap Title]
- **Documented Claim**: [exact text from README/docs/help]
- **Current State**: [what actually exists in the code]
- **User Impact**: [how this affects the target audience]
- **Closing the Gap**: [what needs to be implemented or fixed]
- **Effort Estimate**: [small/medium/large based on implementation complexity]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | Core feature documented but non-functional, installation instructions broken, or output format produces incorrect data |
| HIGH | Feature partially implemented with user-visible limitations not mentioned in docs, or CLI flag documented but ignored |
| MEDIUM | Feature works but deviates from documentation in edge cases, or help text is misleading |
| LOW | Minor documentation inaccuracies, cosmetic differences from documented behavior, or missing examples |

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Concrete fix**: State exactly what needs to be implemented, fixed, or documented. Do not recommend "consider completing the feature."
2. **Respect project idioms**: Recommendations must follow the existing codebase's conventions and architecture.
3. **Verifiable**: Include a validation approach (e.g., specific test case, CLI invocation that demonstrates the fix).
4. **User-focused**: Frame remediation in terms of what the user will be able to do after the fix.

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Use `go-stats-generator` metrics as supporting evidence.
- Every finding must reference a specific documented claim and the code location where it should be fulfilled.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Evaluate the code against its **own documented product surface**, not competitor features or arbitrary expectations.
- Apply the Phase 3f false-positive prevention checks to every candidate finding before including it.

## Tiebreaker
Prioritize: broken core features → missing documented features → partial implementations → documentation inaccuracies → cosmetic issues. Within a level, prioritize by user impact (features used daily > features used occasionally > edge cases).
