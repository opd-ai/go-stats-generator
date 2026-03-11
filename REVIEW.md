# TASK: Analyze how well this codebase achieves its stated goals and generate a prioritized roadmap for closing the gaps.

## Execution Mode
**Report generation only** — produce a roadmap document. Do not modify source code.

## Output
Write exactly one file: **`ROADMAP.md`** in the repository root (the directory containing `go.mod`).
If `ROADMAP.md` already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Discover Project Goals
Before assessing anything, understand what the project is trying to accomplish:
1. Read the project README thoroughly — extract every stated goal, feature claim, capability promise, performance target, and audience statement. These are the **acceptance criteria** for the review.
2. Examine `go.mod` for module path, Go version, and dependency footprint.
3. Scan for existing CI (`.github/workflows/`, `.gitlab-ci.yml`, `Makefile`) and note what quality checks already run.
4. List packages (`go list ./...`) and identify the architectural layers and their responsibilities.
5. Check for an existing roadmap, backlog, issue tracker, or changelog that reveals maintainer priorities and planned work.
6. Look for design documents, ADRs, or spec files that clarify intent beyond the README.

### Phase 1: Metrics Collection
```bash
go-stats-generator analyze . --skip-tests --format json --output review-metrics.json
go-stats-generator analyze . --skip-tests
```

### Phase 2: Goal-Achievement Assessment
For each stated goal or feature claim discovered in Phase 0, evaluate:

1. **Does the feature exist?** Trace through the codebase to confirm the claimed functionality is implemented, not just stubbed.
2. **Does it work correctly?** Cross-reference with `go-stats-generator` metrics to identify risk areas:
   - Functions with cyclomatic complexity >15 or length >50 lines on critical paths are high-risk for bugs.
   - Packages with <70% doc coverage may have undocumented behavioral differences from what the README claims.
   - Cross-reference `.duplication` to find areas where copy-paste may have introduced behavioral drift.
3. **Does it meet its own stated quality bar?** If the project claims performance targets, test coverage levels, or scalability guarantees, verify them with evidence.
4. **Are there gaps between ambition and reality?** Identify features that are documented but incomplete, partially implemented, or non-functional.

Run `go test -race ./...` and `go vet ./...` to confirm baseline health.

Use the project's own conventions and architecture as the standard — do not impose external standards that the project does not claim to follow.

### Phase 3: Generate ROADMAP.md
```markdown
# Goal-Achievement Assessment

## Project Context
- **What it claims to do**: [summary from README]
- **Target audience**: [who the project serves]
- **Architecture**: [key packages and their roles]
- **Existing CI/quality gates**: [list]

## Goal-Achievement Summary
| Stated Goal | Status | Evidence | Gap Description |
|-------------|--------|----------|-----------------|
| [Goal from README] | ✅ Achieved / ⚠️ Partial / ❌ Missing | [metric or code reference] | [what's missing, if anything] |

**Overall: [N]/[total] goals fully achieved**

## Roadmap
### Priority 1: [Most impactful unachieved or partially achieved goal]
- [ ] [Specific step with file/function reference]
- [ ] [Validation: how to confirm this goal is now achieved]

### Priority 2: [Next most impactful gap]
- [ ] ...
```

## Review Rules
- **Goal-first**: Every finding must trace back to a stated goal or feature claim. Do not invent requirements the project does not claim.
- **Evidence-based**: Use `go-stats-generator` metrics as quantitative evidence. Cite specific files, functions, and metric values.
- **Context-sensitive**: A 50-line function in a parser may be perfectly fine. A high complexity score in a utility function is more concerning. Use the project's own baseline distribution to calibrate.
- Do NOT recommend TLS, HTTPS, or transport-layer encryption — transport security is handled by infrastructure.
- Prioritize roadmap items by how much they would advance the project's stated goals, not by arbitrary code-quality checklists.

## Tiebreaker
Prioritize gaps that affect the most users or the most critical claimed functionality first. Within a priority level, address the highest-risk items (most complexity, least test coverage) first.
