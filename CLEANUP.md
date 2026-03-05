# TASK: Remove binary artifacts, eliminate redundant reports, consolidate duplicate tests, and update .gitignore.

## Execution Mode
**Autonomous action** — execute all cleanup steps directly, no user approval between steps.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output pre-cleanup.json --sections duplication
```

### Phase 2: Cleanup
Execute these steps in order:

**Step 1 — Remove binary artifacts:**
- Find and remove compiled binaries, `.exe`, `.so`, `.dylib` files not in `build/`.
- Remove any stale `go-stats-generator` binary in the repo root (the installed one is in `$GOPATH/bin`).

**Step 2 — Remove redundant report files:**
- Identify JSON baseline/diff files that are development artifacts (e.g., `*-baseline.json`, `*-post.json`, `diff-report*.json`).
- Remove files that are not tracked inputs to active workflows.
- Preserve `ROADMAP.md`, `PLAN.md`, `README.md`, and any `AUDIT.md`.

**Step 3 — Consolidate duplicate tests:**
- Use the duplication report (`.duplication.clone_pairs`) to find test files with >20 duplicated lines.
- Extract shared setup/assertion code into test helpers (e.g., `testutil_test.go`).
- Consolidate table-driven test cases that differ only in inputs.

**Step 4 — Update .gitignore:**
- Add patterns for: compiled binaries, analysis JSON outputs, loop artifacts (`.loop-*.json`, `loop.log`, `test-output.txt`).
- Do not ignore checked-in config files or test fixtures.

Run `go test -race ./...` after each step to confirm no regressions.

### Phase 3: Validate
```bash
go-stats-generator analyze . --skip-tests --format json --output post-cleanup.json --sections duplication
go-stats-generator diff pre-cleanup.json post-cleanup.json
```
Confirm: duplication ratio did not increase, all tests pass.

## Thresholds
- Duplication ratio: must not increase after cleanup
- All tests: must remain passing
- No complexity regressions in diff report

## Cleanup Rules
- Only remove files that are clearly artifacts (not source code, configs, or test fixtures).
- When consolidating tests, prefer table-driven patterns over identical test functions.
- If uncertain whether a file should be removed, leave it and note it in output.
- Never remove `testdata/` fixtures.

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
## File Classification Guide
| Pattern | Action | Rationale |
|---------|--------|-----------|
| `*.json` (root, prefixed with baseline/diff/post) | Remove | Development artifact |
| `go-stats-generator` (root binary) | Remove | Should be in $GOPATH/bin |
| `*.exe`, `*.so`, `*.dylib` | Remove | Platform binaries |
| `testdata/**` | Keep | Test fixtures |
| `*.md` (ALL_CAPS) | Keep | Prompt files |
| `loop.log`, `test-output.txt` | Remove | Runtime artifacts |

## Validation Checklist
- [ ] No binary artifacts remain in repository root
- [ ] .gitignore updated with all artifact patterns
- [ ] All tests pass after cleanup
- [ ] Duplication ratio did not increase
