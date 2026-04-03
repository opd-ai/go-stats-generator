# TASK: Remove binary artifacts, eliminate redundant reports, consolidate duplicate tests, and update .gitignore.

## Execution Mode
**Autonomous action** — execute all cleanup steps directly, no user approval between steps.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project
1. Read the project README to understand its build process and artifact patterns.
2. Examine `go.mod`, `Makefile`, or CI config to identify expected build outputs and their locations.
3. Discover the project's `.gitignore` patterns and note what's already excluded.
4. Identify which files are development artifacts vs. tracked project inputs.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output pre-cleanup.json --sections duplication
```

### Phase 2: Cleanup
Execute these steps in order:

**Step 1 — Remove binary artifacts:**
- Find and remove compiled binaries, `.exe`, `.so`, `.dylib` files not in expected build output directories.
- Remove any stale binaries in the repo root.

**Step 2 — Remove redundant report files:**
- Identify JSON baseline/diff files that are development artifacts (e.g., `*-baseline.json`, `*-post.json`, `diff-report*.json`).
- Remove files that are not tracked inputs to active workflows.
- Preserve README, tracked config files, and any active backlog/audit files.

**Step 3 — Consolidate duplicate tests:**
- Use the duplication report (`.duplication.clone_pairs`) to find test files with >20 duplicated lines.
- Extract shared setup/assertion code into test helpers.
- Consolidate table-driven test cases that differ only in inputs.

**Step 4 — Update .gitignore:**
- Add patterns for: compiled binaries, analysis JSON outputs, loop artifacts, and any other discovered artifact patterns.
- Do not ignore checked-in config files or test fixtures.

Run `go test -race ./...` after each step to confirm no regressions.

### Phase 3: Validate
```bash
go-stats-generator analyze . --skip-tests --format json --output post-cleanup.json --sections duplication
go-stats-generator diff pre-cleanup.json post-cleanup.json
```
Confirm: duplication ratio did not increase, all tests pass.

## Cleanup Rules
- Only remove files that are clearly artifacts (not source code, configs, or test fixtures).
- When consolidating tests, prefer table-driven patterns over identical test functions.
- If uncertain whether a file should be removed, leave it and note it in output.
- Never remove test fixture directories.

## Output Format
```
Step 1: Removed [N] binary artifacts ([list])
Step 2: Removed [N] redundant reports ([list])
Step 3: Consolidated [N] duplicate test blocks across [M] files
Step 4: Updated .gitignore with [N] new patterns
Tests: PASS | Duplication: [before]% -> [after]%
```

## Tiebreaker
When cleanup actions have equal impact, prefer the action that removes the most files first.
