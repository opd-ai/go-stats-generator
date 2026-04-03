# TASK: Remove redundant test artifacts, consolidate duplicate test helpers, and clean up test infrastructure.

## Execution Mode
**Autonomous action** — execute all cleanup steps directly, no user approval between steps.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Test Infrastructure
1. Read the project README and understand its test strategy.
2. Discover the test framework and assertion library in use.
3. Identify existing test helper patterns: are there `testutil_test.go` files? Shared setup functions?
4. Note which packages have test files and how test fixtures are organized.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --only-tests --format json --output pre-cleanup.json --sections duplication
```

### Phase 2: Cleanup
Execute these steps in order:

**Step 1 — Identify redundant test helpers:**
- Find duplicate test helper functions across `*_test.go` files.
- Identify test setup code that is copy-pasted between files.
- Flag test helpers that are defined but never called.

**Step 2 — Consolidate duplicate test helpers:**
- Use `.duplication.clone_pairs` to find test files with >30 duplicated lines.
- Extract shared code into package-level `testutil_test.go` files.
- Consolidate table-driven test cases that differ only in inputs.
- Ensure all extracted helpers use `t.Helper()`.

**Step 3 — Standardize test patterns** (match existing project conventions):
- Convert repetitive assertion blocks into table-driven subtests.
- Replace manual setup/teardown with `t.Cleanup()` patterns.
- Standardize error assertion patterns across test files.

**Step 4 — Remove stale test fixtures:**
- Identify test fixture files not referenced by any test.
- Remove unused fixtures (preserve all referenced ones).

Run `go test -race ./...` after each step.

### Phase 3: Validate
```bash
go-stats-generator analyze . --only-tests --format json --output post-cleanup.json --sections duplication
go-stats-generator diff pre-cleanup.json post-cleanup.json
```

## Default Thresholds (test-appropriate)
- Duplication ratio: must not increase (target: <10%)
- All tests must remain passing

## Cleanup Rules
- Match the project's existing test helper and assertion patterns.
- Always mark extracted helpers with `t.Helper()`.
- If uncertain whether a fixture is used, leave it.
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
Prefer the action that consolidates the most duplicate code first.
