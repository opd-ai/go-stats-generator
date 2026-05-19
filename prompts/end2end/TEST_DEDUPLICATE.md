# TASK (END-TO-END): Identify and consolidate ALL code clone groups in **test files** above test-appropriate duplication thresholds — no cap per session.

## Execution Mode
**Autonomous action** — deduplicate test code, validate with tests and diff.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Test Strategy
1. Read the project README and understand its testing philosophy.
2. Discover the test framework in use and existing test helper patterns.
3. Identify whether the project already uses table-driven tests, shared helpers, or `testutil_test.go` patterns.
4. Note the assertion style — consolidated tests must match existing conventions.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --only-tests --format json --output baseline.json --sections duplication --min-block-lines 6 --similarity-threshold 0.80
go-stats-generator analyze . --only-tests
```

### Phase 2: Deduplicate
1. Extract `.duplication.clones` sorted by `line_count` ascending (smallest first).
2. Classify clone groups:
   - CRITICAL: >=30 lines AND >=3 instances
   - HIGH: >=15 lines AND >=2 instances
   - MEDIUM: >=6 lines AND >=2 instances
3. For each clone group (starting with the shortest/simplest):
   - Choose the consolidation strategy:
     - **Table-driven pattern**: preferred for test deduplication when clones differ only in inputs/outputs.
     - **Extract test helper**: move shared setup/assertion into a `t.Helper()` function.
     - **Extract to testutil**: if shared across multiple test files in a package.
   - Do NOT merge test cases that validate different conceptual behaviors.
   - Ensure all helpers use `t.Helper()`.
4. Run `go test -race ./...` after each consolidation.

### Phase 3: Validate
```bash
go-stats-generator analyze . --only-tests --format json --output post.json --sections duplication --min-block-lines 6 --similarity-threshold 0.80
go-stats-generator diff baseline.json post.json
```

## Default Thresholds (test-appropriate)
| Metric | Target |
|--------|--------|
| Duplication ratio | <10% |
| Min block size | 6 lines |
| Similarity threshold | 0.80 |

> Test duplication threshold is relaxed to 10% (vs 5% production) since some test repetition aids clarity.

## Deduplication Rules
- Start with the shortest clone group per tier — simpler consolidations are safer.
- **Table-driven tests** are the preferred deduplication strategy for test code.
- Preserve all existing test coverage and pass/fail behavior.
- Each extracted helper must use `t.Helper()` and be <45 lines.

## Output Format
```
Clone group [N]: [file1:line]-[file2:line] ([M] lines, [K] instances)
  Strategy: [table-driven | extract helper | extract testutil]
  Consolidated into: [new_helper_name]
  Tests: PASS
Duplication: [before]% -> [after]%
```


## End-to-End Policy
This is an **end-to-end variant**. The following rules override any conflicting instructions above:
- **No finding cap** — report or fix every issue that meets the threshold. Do not stop at 10, 5, or any other fixed count.
- **Complete coverage** — process every file, every function, and every package. Do not sample or skip lower-priority items.
- **Iterative until done** — if the session's context is running low, commit progress, document the remaining scope, and continue in a fresh session. Never abandon remaining work.
- **Findings are cumulative** — each pass may surface new issues; repeat until a full pass produces zero new findings above the threshold.

## Tiebreaker
Within each priority tier, consolidate the shortest clone group first.
