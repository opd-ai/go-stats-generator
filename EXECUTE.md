# TASK: Execute the next planned task from the project backlog in strict priority order: AUDIT.md first, then PLAN.md, then ROADMAP.md.

## Execution Mode
**Autonomous action** — implement the task fully, validate with tests and diff.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output baseline.json --sections functions,duplication,documentation
```

### Phase 2: Select and Execute Task
1. **Task selection** — strict file priority order, NO EXCEPTIONS:
   - **First**: Check AUDIT.md for unchecked `- [ ]` findings. Take the first unchecked item.
   - **Second**: If no AUDIT.md items remain, check PLAN.md for the first incomplete step.
   - **Third**: If no PLAN.md items remain, check ROADMAP.md for the first incomplete item.
   - Do NOT skip files. Do NOT reorder. AUDIT.md items ALWAYS take priority over PLAN.md, which ALWAYS takes priority over ROADMAP.md.

2. **Task grouping** — if the next task is part of a logical group (e.g., "Step 3" in PLAN.md has sub-items a, b, c), execute the entire group as one unit.

3. **Implementation**:
   - Follow Go best practices: functions <=30 lines, explicit error handling, GoDoc comments.
   - Preserve all existing public API signatures.
   - Run `go test -race ./...` after implementation.
   - Run `go vet ./...` to confirm no new issues.

4. **Mark completion**: Check off completed items (`- [x]`) in the source file.

### Phase 3: Validate
```bash
go-stats-generator analyze . --skip-tests --format json --output post.json --sections functions,duplication,documentation
go-stats-generator diff baseline.json post.json
```
Confirm: zero regressions in complexity, duplication, or doc coverage.

## Thresholds
- Max function length: 30 lines
- Max cyclomatic complexity: 10
- Min doc coverage: 70%
- Zero regression tolerance on diff

## Priority Rules
The task priority order is absolute and non-negotiable:
1. **AUDIT.md** — bug fixes and critical findings always come first
2. **PLAN.md** — planned implementation steps come second
3. **ROADMAP.md** — strategic improvements come last

Do NOT second-guess this order. The humans put them in that order deliberately. Execute what is next, not what seems most interesting or impactful.

## Go Coding Standards
- Verb-first function names: `parseConfig`, `buildReport`.
- Explicit error handling: `fmt.Errorf("context: %w", err)`.
- GoDoc comment on every exported symbol.
- Prefer stdlib over external dependencies.
- Race-safe: pass `go test -race ./...`.

## Task Completion Rules
- If an AUDIT.md finding is already resolved (code matches expectation), check it off and move to the next.
- If a task requires information not available in the codebase, note the blocker and skip to the next task.
- After completing all items in AUDIT.md, delete the file to signal completion to loop.sh.
- After completing all items in PLAN.md, delete the file to signal completion to loop.sh.

## Output Format
```
Source: [AUDIT.md | PLAN.md | ROADMAP.md]
Task: [description]
Files modified: [list]
Tests: PASS
Diff: [summary of changes]
```

## Tiebreaker
Always take the first unchecked item in the current priority file. Never skip ahead.
## Backlog Completion
- When all items in AUDIT.md are checked off, delete AUDIT.md.
- When all items in PLAN.md are checked off, delete PLAN.md.
- This signals loop.sh that the backlog is empty and triggers the termination check.
