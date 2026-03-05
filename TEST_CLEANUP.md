# TASK: Remove redundant test artifacts, consolidate duplicate test helpers, and clean up test infrastructure.

## Execution Mode
**Autonomous action** — execute all cleanup steps directly, no user approval between steps.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 1: Baseline
```bash
go-stats-generator analyze . --only-tests --format json --output pre-cleanup.json --sections duplication
```

### Phase 2: Cleanup
Execute these steps in order:

**Step 1 — Identify redundant test helpers:**
- Find duplicate test helper functions across `*_test.go` files.
- Identify test setup code that is copy-pasted between test files.
- Flag test helpers that are defined but never called.

**Step 2 — Consolidate duplicate test helpers:**
- Use the duplication report (`.duplication.clone_pairs`) to find test files with >30 duplicated lines.
- Extract shared setup/assertion code into package-level `testutil_test.go` files.
- Consolidate table-driven test cases that differ only in inputs.
- Ensure all extracted helpers use `t.Helper()`.

**Step 3 — Standardize test patterns:**
- Convert repetitive assertion blocks into table-driven subtests.
- Replace manual setup/teardown with `t.Cleanup()` patterns.
- Standardize error assertion patterns across test files.

**Step 4 — Remove stale test fixtures:**
- Identify `testdata/` files not referenced by any test.
- Remove unused test fixtures (preserve all referenced ones).

Run `go test -race ./...` after each step to confirm no regressions.

### Phase 3: Validate
```bash
go-stats-generator analyze . --only-tests --format json --output post-cleanup.json --sections duplication
go-stats-generator diff pre-cleanup.json post-cleanup.json
```
Confirm: duplication ratio did not increase, all tests pass.

## Thresholds (Test-Appropriate)
- Duplication ratio: must not increase after cleanup (target: <10%)
- All tests: must remain passing
- No complexity regressions in diff report

> **Note**: Test duplication threshold is relaxed to 10% (vs 5% for production code) since some test repetition is acceptable for clarity.

## Cleanup Rules
- When consolidating test helpers, prefer table-driven patterns over identical test functions.
- Always mark extracted helpers with `t.Helper()`.
- If uncertain whether a test fixture is used, leave it and note it in output.
- Never remove `testdata/` fixtures that are referenced by tests.
- Prefer `t.TempDir()` over manual temp directory management.

## Output Format
```
Step 1: Found [N] redundant test helpers across [M] files
Step 2: Consolidated [N] duplicate test blocks across [M] files
Step 3: Converted [N] test functions to table-driven pattern
Step 4: Removed [N] unused test fixtures ([list])
Tests: PASS | Duplication: [before]% -> [after]%
```

## Tiebreaker
When cleanup actions have equal impact, prefer the action that consolidates the most duplicate code first.
## Test File Classification Guide
| Pattern | Action | Rationale |
|---------|--------|-----------|
| `*_test.go` with duplicate helpers | Consolidate | Reduce maintenance burden |
| `testutil_test.go` | Keep/extend | Central test helper location |
| `testdata/**` (referenced) | Keep | Test fixtures |
| `testdata/**` (unreferenced) | Remove | Stale fixture |
| Duplicate table-driven cases | Merge | Redundant coverage |

## Validation Checklist
- [ ] All duplicate test helpers consolidated
- [ ] All extracted helpers use `t.Helper()`
- [ ] All tests pass after cleanup
- [ ] Duplication ratio did not increase
