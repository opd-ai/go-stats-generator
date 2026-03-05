# TASK: Execute ONE highest-priority incomplete task from ROADMAP.md with full implementation and baseline/diff validation.

## Execution Mode
**Autonomous action** — implement the task, validate with tests and diff.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output baseline.json --sections functions,duplication,documentation
```

### Phase 2: Select and Implement
1. Read ROADMAP.md and identify the highest-priority incomplete task.
2. Understand the task requirements and acceptance criteria.
3. Implement the task following Go best practices:
   - Functions <=30 lines, cyclomatic complexity <=10.
   - Explicit error handling with `fmt.Errorf("context: %w", err)`.
   - GoDoc comments on all exported symbols.
   - Prefer stdlib over external dependencies.
   - If external libraries are needed, choose well-maintained options (>1000 GitHub stars, recent activity).
4. Preserve all existing public API signatures.
5. Run `go test -race ./...` after implementation.
6. Run `go vet ./...` to confirm no new issues.

### Phase 3: Validate
```bash
go-stats-generator analyze . --skip-tests --format json --output post.json --sections functions,duplication,documentation
go-stats-generator diff baseline.json post.json
```
Confirm: zero regressions in complexity, duplication, or doc coverage.

## Thresholds
| Metric | Maximum/Minimum |
|--------|----------------|
| Cyclomatic complexity | <=10 |
| Function length | <=30 lines |
| Doc coverage | >=70% |
| Duplication ratio | <5% |

## Implementation Rules
- Execute exactly ONE task per invocation.
- Preserve all existing public APIs.
- Follow Go naming conventions (verb-first functions, CamelCase types).
- Add tests for new functionality where practical.
- Do not introduce new dependencies without justification.

## Mark Completion
After successful implementation and validation:
- Check off the completed item in ROADMAP.md (`- [x]`).
- Note the diff summary in your output.

## Output Format
```
Task: [description from ROADMAP.md]
Files modified: [list]
Tests: PASS
Diff summary: [key metric changes]
Remaining: [count of incomplete ROADMAP items]
```

## Tiebreaker
Take the highest-priority incomplete task. If tied, choose the task with broadest file impact.
## Implementation Quality Standards
- Verb-first function names: `parseConfig`, `buildReport`.
- Explicit error handling: `fmt.Errorf("context: %w", err)`.
- GoDoc comment on every exported symbol.
- Prefer stdlib. If external deps needed, >1000 GitHub stars and recent activity.
- All new code covered by existing or new tests.

## Constraints
- Execute exactly ONE task per invocation.
- Preserve all existing public APIs.
- Do not introduce new dependencies without justification.
- If the task is blocked, note the blocker and skip to the next task.

## Validation Checklist
- [ ] Task fully implemented
- [ ] All tests pass with -race flag
- [ ] Diff shows zero regressions
- [ ] ROADMAP.md updated with completion mark
