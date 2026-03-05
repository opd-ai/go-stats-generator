# TASK: Generate a data-driven implementation plan by combining `go-stats-generator` metrics with the project's own priorities and backlog.

## Execution Mode
**Report generation only** — produce a plan document. Do not modify source code.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project
Before generating any plan, build deep context:
1. Read the project README to learn: what it does, who uses it, and what "done" means for the current milestone.
2. Examine `go.mod` for module path and dependency profile.
3. List packages (`go list ./...`) and identify the architectural layers and their responsibilities.
4. Discover the project's own backlog: look for roadmap files, issue trackers, TODO comments, changelog, or milestone documents.
5. Identify the project's conventions: test strategy, CI gates, deployment model, and what quality attributes the maintainers prioritize.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output metrics.json --sections functions,duplication,documentation,packages,patterns
```

### Phase 2: Plan Generation
1. Identify the first incomplete milestone from the project's backlog (discovered in Phase 0).
2. Cross-reference with metrics to quantify scope:
   - Count functions above complexity thresholds for that milestone.
   - Measure current duplication ratio if deduplication is planned.
   - Check doc coverage if documentation work is planned.
   - Review package coupling/cohesion for structural work.
3. Break the milestone into ordered implementation steps:
   - Each step must be independently testable.
   - Each step must have a clear acceptance criterion tied to a `go-stats-generator` metric.
   - Order by dependency (prerequisites first), then by impact (highest value first).
   - Account for the project's architecture — don't plan changes that ignore package boundaries or established patterns.
4. Estimate scope using the codebase's own baseline distribution:
   - Small: <5 items above threshold
   - Medium: 5–15 items above threshold
   - Large: >15 items above threshold

### Phase 3: Write Plan
```markdown
# Implementation Plan: [Milestone Name]

## Project Context
- **What it does**: [one sentence from README]
- **Current milestone**: [from backlog]
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
```

Optionally generate a gaps document for findings outside the current plan's scope.

## Default Thresholds for Scope Assessment (calibrate to project)
| Metric | Small | Medium | Large |
|--------|-------|--------|-------|
| Functions above complexity 9.0 | <5 | 5–15 | >15 |
| Duplication ratio | <3% | 3–10% | >10% |
| Doc coverage gap | <10% | 10–25% | >25% |

## Plan Rules
- Each step must reference specific files or packages.
- Each step must have a validation command using `go-stats-generator`.
- Steps must be ordered by dependency, then by descending impact.
- Do not plan work that the project's backlog does not call for.
- Plans should reflect the project's own priorities, not just mechanically list metric violations.

## Tiebreaker
Target the earliest incomplete milestone. If all milestones are complete, propose the next logical enhancement based on metrics.
