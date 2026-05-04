# TASK: Execute as many planned tasks as possible from the target project's backlog in strict priority order: audit findings first, then planned steps, then roadmap items.

## Execution Mode
**Autonomous action** — implement tasks fully, validate each with tests and diff, then continue to the next task until a stopping condition is met.
This prompt operates on a **third-party Go project**, not on go-stats-generator itself. Every decision must serve the target project's own stated goals and conventions.

**IMPORTANT** some tests may require a display. If a display is not available, use `xvfb-run`.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Target Project's Goals
Before executing any task, build deep context on the project you are working in:
1. Read the project README thoroughly — extract every stated goal, feature claim, capability promise, performance target, and audience statement. These are the **acceptance criteria** for your work.
2. Examine `go.mod` for module path, Go version, and key dependencies.
3. Scan existing source files to identify the project's code style: error handling patterns, naming conventions, test strategy, and preferred idioms.
4. Note any CI configuration, linter configs, or code generation patterns to respect.
5. Identify which packages and functions are on critical paths for the project's stated goals — changes to these require extra care.

### Phase 1: Online Research
Use web search to build context before executing:
1. Search for the project on GitHub — read open issues and recent PRs related to planned tasks.
2. Research any dependencies or APIs involved in the upcoming tasks for known issues or deprecations.
3. Look up implementation best practices relevant to the planned changes.

Keep research brief (≤10 minutes). Record only findings that directly affect how you implement or validate the tasks.

### Phase 2: Session Baseline
Capture a single baseline at the start of the session. This baseline persists across all tasks and is used for the final session-level diff.
```bash
go-stats-generator analyze . --skip-tests --format json --sections functions,duplication,documentation > tmp/baseline-exec.json
```
Delete `tmp/baseline-exec.json` only after the final session validation is complete.

### Phase 3: Task Execution Loop
Repeat the following cycle for each task until a **stopping condition** is met:

#### 3a. Select Task
Strict file priority order, NO EXCEPTIONS:
- **First**: `AUDIT.md` (or any `*AUDIT*.md`). Take the first unchecked `- [ ]` item.
- **Second**: `PLAN.md`. Take the first incomplete step.
- **Third**: `ROADMAP.md`. Take the first incomplete item.
- Do NOT skip priority levels. Do NOT reorder.

If the next task has logical sub-items, execute the entire group as one unit.

#### 3b. Implement
Match the project's existing conventions and advance its stated goals:
- Mirror the codebase's error handling style (wrapping pattern, sentinel errors, etc.).
- Follow the project's naming conventions and package structure.
- Respect established function length and complexity norms (default targets: <=30 lines, cyclomatic <=10).
- Preserve all existing public API signatures.
- Verify the change serves the project's stated goals — do not introduce code that contradicts or is irrelevant to what the project claims to do.

#### 3c. Per-Task Validation
After each task, confirm:
1. `go test -race ./...` exits 0.
2. `go vet ./...` exits 0.
3. The change advances (or does not regress) the project's stated goals.

If per-task validation fails, fix the issue before continuing. If the fix would require modifying files outside the task's scope or adding more than 20 lines of unplanned code, revert the task changes, note the blocker, and proceed to the next task.

#### 3d. Mark Completion
Check off completed items (`- [x]`) in the source file. Record the task in the session log for the final output.

#### Stopping Conditions
Stop the loop when **any** of the following is true:
- **Backlog exhausted**: All items in `AUDIT.md` and `PLAN.md` have been completed and both files have been deleted (see Task Completion Rules).
- **Unrecoverable regression**: A task causes test or validation failures, or clearly regresses the project's stated goals, and the issue cannot be resolved quickly — revert it, log it, and stop.
- **Context boundary**: The next task requires modifying files in a top-level package not yet touched in this session and involves a subsystem with different domain concerns (e.g., switching from data processing to HTTP handlers, or from core logic to CI configuration).
- **High-risk threshold**: The next task involves changes to public API signatures, database schemas, or other high-blast-radius modifications that warrant isolated review.

When no stopping condition is met and you are unsure whether to continue, execute one more task rather than stopping early. Prefer completing more work per session.

### Phase 4: Session Validation
After the loop ends, perform a final session-level validation against the original baseline:
```bash
go-stats-generator analyze . --skip-tests --format json --sections functions,duplication,documentation > tmp/post-exec.json
go-stats-generator diff tmp/baseline-exec.json tmp/post-exec.json
```
Delete `tmp/baseline-exec.json` and `tmp/post-exec.json` when done.

Confirm ALL of the following:
1. **No metric regressions** in complexity, duplication, or doc coverage relative to the session baseline.
2. **Tests pass**: `go test -race ./...` and `go vet ./...` succeed.
3. **Compliance with the project's stated goals**: The cumulative changes advance (or at minimum do not regress) the project's stated goals as documented in its README. Evaluate changes against the project's **own stated goals** first, then against general engineering best practices. Cross-reference with AUDIT.md/PLAN.md/ROADMAP.md to confirm each task's goal-achievement intent was fulfilled.

## Success Criteria
| Criterion | Check |
|-----------|-------|
| No metric regressions | `go-stats-generator diff` shows zero regressions across the entire session |
| Tests pass | `go test -race ./...` exits 0 |
| Vet clean | `go vet ./...` exits 0 |
| At least one task completed | Session completed ≥1 task successfully |
| Compliance with the project's stated goals | Changes advance the project's own stated goals (README) and fulfill the intent documented in AUDIT.md / PLAN.md / ROADMAP.md |

## Default Thresholds (calibrate to project baseline)
- Max function length: 30 lines
- Max cyclomatic complexity: 10
- Min doc coverage: 70%
- Zero regression tolerance on diff

## Priority Rules
The task priority order is absolute and non-negotiable:
1. **`AUDIT.md`** — bug fixes and critical findings always come first
2. **`PLAN.md`** — implementation plan items come second
3. **`ROADMAP.md`** — strategic improvements come last

Execute what is next, not what seems most interesting or impactful.

## Task Completion Rules
- If a finding is already resolved (code matches expectation), check it off and move to the next task in the loop.
- If a task requires information not available, note the blocker and skip to the next task in the loop.
- After completing all items in `AUDIT.md`, delete it to signal completion.
- After completing all items in `PLAN.md`, delete it to signal completion.

**Critical**: The development loop (`loop.sh`) halts only when both `AUDIT.md` and `PLAN.md` are deleted. Failure to delete these files upon completion will cause the loop to run indefinitely.

## Output Format
```
## Session Summary
Tasks completed: [N]
Tasks skipped: [N] (with reasons)
Tasks reverted: [N] (with reasons)

## Task Log

### Task 1
Source: [audit file | plan file | roadmap file]
Task: [description]
Files modified: [list]
Tests: PASS
Result: COMPLETED

### Task 2
Source: [audit file | plan file | roadmap file]
Task: [description]
Files modified: [list]
Tests: PASS
Result: COMPLETED

...

## Session Diff
[go-stats-generator diff summary comparing session baseline to final state]

## Stop Reason
[backlog exhausted | unrecoverable regression | context boundary | high-risk threshold]
```

## Tiebreaker
Always take the first unchecked item in the current priority file. Never skip ahead.
