# TASK: Perform a multi-dimensional production readiness analysis and generate a prioritized remediation plan in ROADMAP.md.

## Execution Mode
**Report generation only** — produce ROADMAP.md with remediation plan. Do not modify source code.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 1: Full Analysis
```bash
go-stats-generator analyze . --skip-tests --format json --output review-metrics.json
go-stats-generator analyze . --skip-tests
```

### Phase 2: Gate Assessment
Evaluate the codebase against 7 production readiness gates:

| Gate | Threshold | Metric Source |
|------|-----------|---------------|
| Complexity | All functions cyclomatic <=10 | `.functions` |
| Function Length | All functions <=30 lines | `.functions` |
| Documentation | >=80% overall coverage | `.documentation` |
| Duplication | <5% ratio | `.duplication` |
| Circular Dependencies | Zero detected | `.packages` |
| Naming | Zero violations | `.naming` |
| Concurrency Safety | No high-risk patterns | `.patterns.concurrency_patterns` |

For each gate, record: pass/fail, current value, violation count, worst offenders (top 5).

### Phase 3: Generate ROADMAP.md
```markdown
# PRODUCTION READINESS ASSESSMENT: go-stats-generator

## Readiness Summary
| Dimension | Score | Gate | Status |
|-----------|-------|------|--------|
| [gate] | [value] | [threshold] | PASS/FAIL |

**Overall: [N]/7 gates passing — [READY | CONDITIONALLY READY | NOT READY]**

## Critical Issues (Failed Gates)
### [Gate]: [current value] vs [threshold]
Top offenders:
- [function/file]: [metric value]
...

## Remediation Plan
### Phase 1: [Highest priority failed gate]
- [ ] [Specific remediation step]
...

### Phase 2: [Next priority]
...
```

## Readiness Verdicts
| Gates Passing | Verdict |
|---------------|---------|
| 7/7 | PRODUCTION READY |
| 5–6/7 | CONDITIONALLY READY |
| <5/7 | NOT READY |

## Remediation Priority Order
1. Concurrency safety (data corruption risk)
2. Circular dependencies (build stability)
3. Complexity (bug density correlation)
4. Function length (maintainability)
5. Documentation (onboarding / API clarity)
6. Duplication (maintenance burden)
7. Naming (consistency)

## Review Rules
- Do NOT recommend TLS, HTTPS, or transport-layer encryption — transport security is handled by infrastructure.
- Focus on application-layer concerns: input validation, error handling, data sanitization.
- Base all assessments on quantitative `go-stats-generator` metrics, not subjective opinion.

## Output
- `ROADMAP.md` with gate assessment and prioritized remediation plan.

## Tiebreaker
Within a remediation phase, fix the highest-impact items first (most violations or highest severity).
