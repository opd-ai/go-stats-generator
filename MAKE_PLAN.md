# Generate Implementation Plan from Repository Documentation

## Execution Mode
**Report generation** — Analyze repository documentation and produce structured markdown output. Do not commit files or open PRs.

## Objective
Identify the next incomplete development phase from project documentation and generate an actionable implementation plan as `PLAN.md`, with an optional `GAPS.md` for unresolved blockers.

## Input Review (in priority order)
1. `ROADMAP.md` — Identify the earliest incomplete milestone or phase
2. `README.md` — Understand project purpose, architecture, and current capabilities
3. `docs/*.md` — Gather implementation details, specs, and existing plans
4. Source code structure — Assess current state against documented goals

**If `ROADMAP.md` does not exist**, infer the next phase from any planning documents found (e.g., `FOUNDATION.md`, `TODO.md`, `CHANGELOG.md`, project board references), then note this substitution in the output.

## Phase Selection Rules
- Select the **earliest incomplete milestone** in the roadmap
- If all milestones are complete: propose the next logical enhancement based on project trajectory
- If no roadmap or planning document exists: state this clearly and derive a phase from open issues, audit findings, or `README` feature gaps

## Output Format
Produce **three separate markdown code blocks**, each labeled with its filename:

### 1. `PLAN.md`

```
# Implementation Plan: [Phase Name]

## Phase Overview
- **Objective**: [One-sentence goal]
- **Source Document**: [File used to identify this phase]
- **Prerequisites**: [Completed items required]
- **Estimated Scope**: [Small / Medium / Large]

## Implementation Steps
1. [Actionable task]
   - **Deliverable**: [Specific, verifiable output]
   - **Dependencies**: [If any]

2. [Actionable task]
   - **Deliverable**: [Specific, verifiable output]
   - **Dependencies**: [If any]

## Technical Specifications
- [Key technical decision or constraint]
- [Key technical decision or constraint]

## Validation Criteria
- [ ] [Measurable success criterion]
- [ ] [Measurable success criterion]

## Known Gaps
- [Gap description, or "None identified"]
```

### 2. `GAPS.md` (only if gaps exist; otherwise output "Not needed — no gaps identified")

```
# Implementation Gaps: [Phase Name]

## [Gap Title]
- **Description**: [What information is missing]
- **Impact**: [How it blocks implementation]
- **Resolution**: [What is needed to close this gap]
```

### 3. `ROADMAP.md` addition (exact text to append)

```
- [YYYY-MM-DD] PLAN.md created for [Phase Name]
```

## Quality Criteria
- Every step is independently actionable by a developer
- No step requires information that is undefined — flag it as a gap instead
- Dependencies between steps are explicit
- Deliverables are concrete artifacts (files, functions, passing tests), not vague outcomes
- Technical specifications answer "how", not just "what"
