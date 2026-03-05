# TASK: Generate a data-driven implementation plan (PLAN.md) by combining `go-stats-generator` metrics with ROADMAP.md priorities.

## Execution Mode
**Report generation only** — produce PLAN.md (and optionally GAPS.md) in the repository root. Do not modify source code.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output metrics.json --sections functions,duplication,documentation,packages,patterns
```

### Phase 2: Plan Generation
1. Read ROADMAP.md and identify the first incomplete milestone/phase.
2. Cross-reference with metrics to quantify the scope:
   - Count functions above complexity thresholds for that phase.
   - Measure current duplication ratio if deduplication is planned.
   - Check doc coverage if documentation work is planned.
   - Review package coupling/cohesion for structural work.
3. Break the milestone into ordered implementation steps:
   - Each step must be independently testable.
   - Each step must have a clear acceptance criterion tied to a `go-stats-generator` metric.
   - Steps should be ordered by dependency (prerequisites first) then by impact (highest value first).
4. Estimate scope for the overall plan:
   - Small: <5 functions above threshold
   - Medium: 5–15 functions above threshold
   - Large: >15 functions above threshold

### Phase 3: Write PLAN.md
```markdown
# Implementation Plan: [Milestone Name]

## Phase Overview
- **Objective**: [one sentence]
- **Source**: [ROADMAP.md section reference]
- **Estimated Scope**: [Small | Medium | Large]

## Metrics Summary
- Complexity hotspots: [N] functions above threshold
- Duplication ratio: [N]%
- Doc coverage: [N]%
- Package coupling: [notable packages]

## Implementation Steps

### Step 1: [Title]
- **Deliverable**: [what changes]
- **Dependencies**: [prerequisites]
- **Acceptance**: [go-stats-generator metric target]
- **Validation**: `go-stats-generator analyze ... | jq '[specific query]'`

### Step 2: ...
```

Optionally generate GAPS.md for findings that don't fit the current plan but should be tracked.

## Thresholds for Scope Assessment
| Metric | Small | Medium | Large |
|--------|-------|--------|-------|
| Functions above complexity 9.0 | <5 | 5–15 | >15 |
| Duplication ratio | <3% | 3–10% | >10% |
| Doc coverage gap | <10% | 10–25% | >25% |

## Plan Rules
- Each step must reference specific files or packages.
- Each step must have a validation command using `go-stats-generator`.
- Steps must be ordered by dependency, then by descending impact.
- Do not plan work that ROADMAP.md does not call for.

## Output
- `PLAN.md` — ordered implementation steps with acceptance criteria.
- `GAPS.md` (optional) — findings outside current milestone scope.

## Tiebreaker
Target the earliest incomplete milestone in ROADMAP.md. If all milestones are complete, propose the next logical enhancement based on metrics.
