# TASK: Execute the next planned task from the target project's backlog in strict priority order: audit findings first, then planned steps, then roadmap items.

## Execution Mode
**Autonomous action** — implement the task fully, validate with tests and diff.
This prompt operates on a **third-party Go project**, not on go-stats-generator itself. Every decision must serve the target project's own stated goals and conventions.

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
1. Search for the project on GitHub — read open issues and recent PRs related to the task you are about to execute.
2. Research any dependencies or APIs involved in the task for known issues or deprecations.
3. Look up implementation best practices relevant to the specific change you will make.

Keep research brief (≤10 minutes). Record only findings that directly affect how you implement or validate the task.

### Phase 2: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --sections functions,duplication,documentation > /tmp/baseline-exec.json
```
Delete `/tmp/baseline-exec.json` after validation is complete.

### Phase 3: Select and Execute Task
1. **Task selection** — strict file priority order, NO EXCEPTIONS:
   - **First**: `AUDIT.md` (or any `*AUDIT*.md`). Take the first unchecked `- [ ]` item.
   - **Second**: `PLAN.md`. Take the first incomplete step.
   - **Third**: `ROADMAP.md`. Take the first incomplete item.
   - Do NOT skip priority levels. Do NOT reorder.

2. **Task grouping** — if the next task has logical sub-items, execute the entire group as one unit.

3. **Implementation** — match the project's existing conventions and advance its stated goals:
   - Mirror the codebase's error handling style (wrapping pattern, sentinel errors, etc.).
   - Follow the project's naming conventions and package structure.
   - Respect established function length and complexity norms (default targets: <=30 lines, cyclomatic <=10).
   - Preserve all existing public API signatures.
   - Verify the change serves the project's stated goals — do not introduce code that contradicts or is irrelevant to what the project claims to do.
   - Run `go test -race ./...` and `go vet ./...` after implementation.

4. **Mark completion**: Check off completed items (`- [x]`) in the source file.

### Phase 4: Validate
```bash
go-stats-generator analyze . --skip-tests --format json --sections functions,duplication,documentation > /tmp/post-exec.json
go-stats-generator diff /tmp/baseline-exec.json /tmp/post-exec.json
```
Delete `/tmp/baseline-exec.json` and `/tmp/post-exec.json` when done.

Confirm ALL of the following:
1. **No metric regressions** in complexity, duplication, or doc coverage.
2. **Tests pass**: `go test -race ./...` and `go vet ./...` succeed.
3. **Compliance with the project's stated goals**: The change advances (or at minimum does not regress) the project's stated goals as documented in its README. Evaluate the change against the project's **own stated goals** first, then against general engineering best practices. Cross-reference with AUDIT.md/PLAN.md/ROADMAP.md to confirm the task's goal-achievement intent was fulfilled.

## Success Criteria
| Criterion | Check |
|-----------|-------|
| No metric regressions | `go-stats-generator diff` shows zero regressions |
| Tests pass | `go test -race ./...` exits 0 |
| Vet clean | `go vet ./...` exits 0 |
| Compliance with the project's stated goals | Change advances the project's own stated goals (README) and fulfills the intent documented in AUDIT.md / PLAN.md / ROADMAP.md |

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
- If a finding is already resolved (code matches expectation), check it off and move to the next.
- If a task requires information not available, note the blocker and skip to the next task.
- After completing all items in an audit file, delete it to signal completion.

## Output Format
```
Source: [audit file | plan file | roadmap file]
Task: [description]
Files modified: [list]
Tests: PASS
Diff: [summary of changes]
```

## Tiebreaker
Always take the first unchecked item in the current priority file. Never skip ahead.
