# TASK: Execute ONE highest-priority incomplete task from the project's backlog with full implementation and baseline/diff validation.

## Execution Mode
**Autonomous action** — implement the task, validate with tests and diff.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project
1. Read the project README to understand its domain, users, and architecture.
2. Examine `go.mod` for module path and dependency profile.
3. Discover the project's coding conventions: error handling style, naming patterns, test strategy.
4. Find the project's backlog: roadmap files, issue tracker, TODO comments, or milestone documents.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output baseline.json --sections functions,duplication,documentation
```

### Phase 2: Select and Implement
1. Identify the highest-priority incomplete task from the project's backlog.
2. Understand the task requirements and acceptance criteria.
3. Implement following the project's established conventions:
   - Match the codebase's error handling style and naming patterns.
   - Respect established function length and complexity norms (default targets: <=30 lines, cyclomatic <=10).
   - Add GoDoc comments on all exported symbols.
   - Prefer stdlib over external dependencies unless the project already uses a relevant library.
4. Preserve all existing public API signatures.
5. Run `go test -race ./...` and `go vet ./...` after implementation.

### Phase 3: Validate
```bash
go-stats-generator analyze . --skip-tests --format json --output post.json --sections functions,duplication,documentation
go-stats-generator diff baseline.json post.json
```
Confirm: zero regressions in complexity, duplication, or doc coverage.

## Default Thresholds (calibrate to project)
| Metric | Target |
|--------|--------|
| Cyclomatic complexity | <=10 |
| Function length | <=30 lines |
| Doc coverage | >=70% |
| Duplication ratio | <5% |

## Implementation Rules
- Execute exactly ONE task per invocation.
- Preserve all existing public APIs.
- Match the project's naming conventions.
- Add tests for new functionality where practical.
- Do not introduce new dependencies without justification.

## Mark Completion
After successful implementation and validation:
- Check off the completed item in the backlog file (`- [x]`).
- Note the diff summary in your output.

## Output Format
```
Task: [description from backlog]
Files modified: [list]
Tests: PASS
Diff summary: [key metric changes]
Remaining: [count of incomplete items]
```

## Tiebreaker
Take the highest-priority incomplete task. If tied, choose the task with broadest file impact.
