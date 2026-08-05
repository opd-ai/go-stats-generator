# TASK: Execute Go project backlog tasks in strict priority: `AUDIT.md` ➔ `PLAN.md` ➔ `ROADMAP.md`.

## 🚨 TO PREVENT 429 API RATE LIMITS (NEMOTRON/QWEN)
- DO NOT dump large files or repetitive codebase readouts into the chat.
- Keep tool inputs/outputs tightly scoped.
- If a 429 error occurs, pause for 5 seconds and retry.

## Execution Rules
- **Autonomous Action**: Implement fully, validate with tests, and advance sequentially.
- **Display**: Use `xvfb-run` for tests if a display is missing.
- **Prerequisite**: Run `which go-stats-generator || go install ://github.com`

---

## WORKFLOW STEPS

### Phase 0: Extraction & Session Baseline
1. Read `README.md` (goals) and `go.mod` (dependencies). Scan ≤3 core files for style context.
2. Generate baseline metrics:
   ```bash
   mkdir -p tmp && go-stats-generator analyze . --skip-tests --format json --sections functions,duplication,documentation > tmp/baseline-exec.json
   ```

### Phase 1: Task Execution Loop
Find the first unchecked `- [ ]` item in priority order. Repeat loop until stopped.

*   **Standards**: Match error wrapping. Maintain API signatures. Max 30 lines/function. Max 10 cyclomatic complexity. Min 70% doc coverage.
*   **Validation**: After *every* task, run `go test -race ./...` and `go vet ./...`. If a fix takes >20 lines or modifies out-of-scope files, revert changes and skip.
*   **Completion**: Check off completed items (`- [x]`) in the markdown file.
*   **Stopping Criteria**: Stop if backlog is exhausted, an unrecoverable test failure occurs, a new domain boundary is crossed, or high-blast-radius breaking changes are met.

---

### Phase 2: Session Validation
After loop exits, run final validation:
```bash
mkdir -p tmp && go-stats-generator analyze . --skip-tests --format json --sections functions,duplication,documentation > tmp/post-exec.json
go-stats-generator diff tmp/baseline-exec.json tmp/post-exec.json && rm -rf tmp
```
**Acceptance**: Zero metric regressions via `diff`, tests pass, and `go vet` is clean.

---

## OUTPUT SCHEMA (Strictly Match)
```markdown
## Session Summary
Tasks completed: [N] | Tasks skipped: [N] (reasons) | Tasks reverted: [N] (reasons)

## Task Log
### Task [N]
Source: [AUDIT.md | PLAN.md | ROADMAP.md]
Task: [Description]
Files modified: [List]
Tests: PASS
Result: [COMPLETED | SKIPPED | REVERTED]
```
