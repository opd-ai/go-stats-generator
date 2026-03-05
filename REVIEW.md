# TASK: Perform a multi-dimensional production readiness analysis and generate a prioritized remediation roadmap.

## Execution Mode
**Report generation only** — produce a remediation roadmap document. Do not modify source code.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Discover Project Context
Before assessing any gate, understand the target codebase:
1. Read the project README to identify: **project type** (library, CLI, service, framework), audience, and stated guarantees.
2. Examine `go.mod` for module path, Go version, and dependency footprint.
3. Scan for existing CI (`.github/workflows/`, `.gitlab-ci.yml`, `Makefile`) and note what quality checks already run.
4. List packages (`go list ./...`) and identify the architectural layers.
5. Check for an existing backlog, issue tracker, or changelog that reveals what maintainers care about.

### Phase 1: Full Analysis
```bash
go-stats-generator analyze . --skip-tests --format json --output review-metrics.json
go-stats-generator analyze . --skip-tests
```

### Phase 2: Gate Assessment
Evaluate against 7 readiness gates. Thresholds below are **tunable defaults** — calibrate against the codebase's own baseline distribution when appropriate.

| Gate | Default Threshold | Metric Source |
|------|-------------------|---------------|
| Complexity | All functions cyclomatic <=10 | `.functions` |
| Function Length | All functions <=30 lines | `.functions` |
| Documentation | >=80% overall coverage | `.documentation` |
| Duplication | <5% ratio | `.duplication` |
| Circular Dependencies | Zero detected | `.packages` |
| Naming | Zero violations | `.naming` |
| Concurrency Safety | No high-risk patterns | `.patterns.concurrency_patterns` |

For each gate: record pass/fail, current value, violation count, worst offenders (top 5).

#### Gate Weighting by Project Type
- **Library**: Documentation and API naming are critical; concurrency matters only if goroutines are exposed.
- **CLI tool**: Complexity and input validation are critical; documentation matters for `--help` and examples.
- **Service/server**: Concurrency safety and graceful shutdown are critical; circular dependencies affect deploy stability.
- **Framework**: All gates carry high weight; API stability and doc coverage are paramount.

### Phase 3: Generate Roadmap
```markdown
# PRODUCTION READINESS ASSESSMENT
## Project Context
- Type: [library | CLI | service | framework]
- Deployment model: [discovered]
- Existing CI checks: [list]

## Readiness Summary
| Gate | Score | Threshold | Status | Weight for Project Type |
|------|-------|-----------|--------|------------------------|

**Overall: [N]/7 gates passing — [READY | CONDITIONALLY READY | NOT READY]**

## Remediation Plan
### Phase 1: [Highest-weight failed gate for this project type]
- [ ] [Specific step with file/function reference]
```

## Readiness Verdicts
| Gates Passing | Verdict |
|---------------|---------|
| 7/7 | PRODUCTION READY |
| 5–6/7 | CONDITIONALLY READY — note which failed gates are low-weight for this project type |
| <5/7 | NOT READY |

## Review Rules
- Do NOT recommend TLS, HTTPS, or transport-layer encryption — transport security is handled by infrastructure.
- Focus on application-layer concerns: input validation, error handling, data sanitization.
- Base all assessments on quantitative `go-stats-generator` metrics, not subjective opinion.
- Prioritize remediation by gate weight for the discovered project type, not a fixed order.
- Context matters: a 50-line function in a parser may be acceptable; a 50-line handler is suspect. Use the project's baseline distribution to calibrate judgments.

## Tiebreaker
Within a remediation phase, fix the highest-impact items first (most violations or highest severity).
