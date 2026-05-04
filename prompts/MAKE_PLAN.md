# TASK: Generate a data-driven implementation plan by combining `go-stats-generator` metrics with the project's own stated goals and priorities.

## Execution Mode
**Report generation only** — produce a plan document. Do not modify source code.

## Output
Write exactly one file: **`PLAN.md`** in the repository root (the directory containing `go.mod`).
If `PLAN.md` already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project's Goals
Before generating any plan, build deep context about what the project is trying to achieve:
1. Read the project README to learn: what it does, who uses it, what it promises, and what "done" means.
2. Examine `go.mod` for module path and dependency profile.
3. List packages (`go list ./...`) and identify the architectural layers and their responsibilities.
4. Discover the project's own backlog: look for roadmap files, issue trackers, TODO comments, changelog, or milestone documents.
5. Identify the project's conventions: test strategy, CI gates, deployment model, and what quality attributes the maintainers prioritize.
6. Look for design documents, ADRs, or spec files that clarify the project's direction.

### Phase 1: Online Research
Use web search to build context that isn't available in the repository:
1. Search for the project on GitHub — read open issues, recent PRs, and community discussions to understand known pain points and user priorities.
2. Research key dependencies from `go.mod` for known vulnerabilities, deprecations, or upcoming breaking changes that should be planned for.
3. Look up best practices and conventions in the project's domain to ensure planned work aligns with community standards.
4. Check whether comparable tools exist — understanding the competitive landscape helps prioritize which goals matter most.

Keep research brief (≤10 minutes). Record only findings that should influence the implementation plan.

### Phase 2: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --sections functions,duplication,documentation,packages,patterns > tmp/metrics.json
```
Delete `tmp/metrics.json` when done — the only persistent output is `PLAN.md`.

### Phase 3: Plan Generation
1. Identify the most important unachieved goals from the project's own documentation (discovered in Phase 0).
2. Cross-reference with metrics to quantify scope and identify blockers:
   - Which functions on goal-critical paths have high complexity?
   - What is the current duplication ratio if consolidation work is planned?
   - What is doc coverage if documentation work is planned?
   - What package coupling/cohesion issues affect the project's architecture goals?
3. Break the work into ordered implementation steps:
   - Each step must be independently testable.
   - Each step must have a clear acceptance criterion tied to a `go-stats-generator` metric or a verifiable behavior.
   - Order by dependency (prerequisites first), then by impact on stated goals (highest value first).
   - Account for the project's architecture — don't plan changes that ignore package boundaries or established patterns.
4. Estimate scope using the codebase's own baseline distribution:
   - Small: <5 items above threshold
   - Medium: 5–15 items above threshold
   - Large: >15 items above threshold

### Phase 4: Write PLAN.md
Write the completed plan to **`PLAN.md` in the repository root** (the directory that contains `go.mod`). Do not write it to the copilot session working directory or any other location.

```markdown
# Implementation Plan: [Goal or Milestone Name]

## Project Context
- **What it does**: [one sentence from README]
- **Current goal**: [the most important unachieved goal]
- **Estimated Scope**: [Small | Medium | Large]

## Goal-Achievement Status
| Stated Goal | Current Status | This Plan Addresses |
|-------------|---------------|---------------------|
| [Goal] | ✅ / ⚠️ / ❌ | Yes / No |

## Metrics Summary
- Complexity hotspots on goal-critical paths: [N] functions above threshold
- Duplication ratio: [N]%
- Doc coverage: [N]%
- Package coupling: [notable packages]

## Implementation Steps

### Step 1: [Title]
- **Deliverable**: [what changes — specific files, functions, or behaviors]
- **Dependencies**: [prerequisites]
- **Goal Impact**: [which stated goal this advances]
- **Acceptance**: [go-stats-generator metric target or verifiable behavior]
- **Validation**: `go-stats-generator analyze ... | jq '[specific query]'`
```

## Default Thresholds for Scope Assessment (calibrate to project)
| Metric | Small | Medium | Large |
|--------|-------|--------|-------|
| Functions above complexity 9.0 | <5 | 5–15 | >15 |
| Duplication ratio | <3% | 3–10% | >10% |
| Doc coverage gap | <10% | 10–25% | >25% |

## Plan Rules
- Each step must reference specific files or packages.
- Each step must have a validation command using `go-stats-generator` or `go test`.
- Steps must be ordered by dependency, then by descending impact on stated goals.
- **Plan what the project needs to achieve its own goals**, not what an arbitrary checklist says.
- When `go vet` or linters report warnings, read the comments surrounding the flagged code. If a comment explicitly acknowledges the warning (e.g., `//nolint:`, an explanatory comment justifying the pattern, or a TODO tracking a known issue), treat it as an acknowledged false positive — do not plan work to "fix" it.
- Plans should reflect the project's own priorities and conventions.
- Every step is independently actionable by a developer — no step requires information that is undefined.
- Dependencies between steps are explicit.
- Deliverables are concrete artifacts (files, functions, passing tests), not vague outcomes.

## Tiebreaker
Target the most impactful unachieved stated goal. If all goals are achieved, propose the next logical enhancement based on the project's trajectory and metrics.
