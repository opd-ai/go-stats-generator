# TASK: Execute the next planned task from the project's backlog in strict priority order: audit findings first, then planned steps, then roadmap items.

## Execution Mode
**Autonomous action** — implement the task fully, validate with tests and diff.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project
Before executing any task, build context:
1. Read the project README to learn what it does, its domain, and its users.
2. Examine `go.mod` for module path, Go version, and key dependencies.
3. Scan existing source files to identify the project's code style: error handling patterns, naming conventions, test strategy, and preferred idioms.
4. Note any CI configuration, linter configs, or code generation patterns to respect.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output baseline.json --sections functions,duplication,documentation
```

### Phase 2: Select and Execute Task
1. **Task selection** — strict file priority order, NO EXCEPTIONS:
   - **First**: Discover any audit findings file (e.g., `AUDIT.md`, `*AUDIT*.md`). Take the first unchecked `- [ ]` item.
   - **Second**: If no audit items remain, find the project's implementation plan (e.g., `PLAN.md`, issue tracker). Take the first incomplete step.
   - **Third**: If no plan items remain, find the project's roadmap or backlog. Take the first incomplete item.
   - Do NOT skip priority levels. Do NOT reorder.

2. **Task grouping** — if the next task has logical sub-items, execute the entire group as one unit.

3. **Implementation** — match the project's existing conventions:
   - Mirror the codebase's error handling style (wrapping pattern, sentinel errors, etc.).
   - Follow the project's naming conventions and package structure.
   - Respect established function length and complexity norms (default targets: <=30 lines, cyclomatic <=10).
   - Preserve all existing public API signatures.
   - Run `go test -race ./...` and `go vet ./...` after implementation.

4. **Mark completion**: Check off completed items (`- [x]`) in the source file.

### Phase 3: Validate
```bash
go-stats-generator analyze . --skip-tests --format json --output post.json --sections functions,duplication,documentation
go-stats-generator diff baseline.json post.json
```
Confirm: zero regressions in complexity, duplication, or doc coverage.

## Default Thresholds (calibrate to project baseline)
- Max function length: 30 lines
- Max cyclomatic complexity: 10
- Min doc coverage: 70%
- Zero regression tolerance on diff

## Priority Rules
The task priority order is absolute and non-negotiable:
1. **Audit findings** — bug fixes and critical findings always come first
2. **Planned steps** — implementation plan items come second
3. **Roadmap items** — strategic improvements come last

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
