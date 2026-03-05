# TASK: Identify and consolidate the top 5–10 most significant code clone groups in **test files** below test-appropriate duplication thresholds.

## Execution Mode
**Autonomous action** — deduplicate test code, validate with tests and diff.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 1: Baseline
```bash
go-stats-generator analyze . --only-tests --format json --output baseline.json --sections duplication --min-block-lines 6 --similarity-threshold 0.80
go-stats-generator analyze . --only-tests
```

### Phase 2: Deduplicate
1. From the baseline JSON, extract `.duplication.clone_pairs` sorted by line count ascending (smallest first).
2. Classify clone groups by priority:
   - CRITICAL: >=30 lines AND >=3 instances
   - HIGH: >=15 lines AND >=2 instances
   - MEDIUM: >=6 lines AND >=2 instances
3. For each clone group (starting with the shortest/simplest):
   - Identify the shared test logic and create a single canonical implementation.
   - Choose the consolidation strategy:
     - **Table-driven pattern**: if clones differ only in inputs/expected outputs (preferred strategy for test deduplication).
     - **Extract test helper**: move shared setup/assertion code into a `t.Helper()` function.
     - **Extract to testutil**: if shared across packages, consider a shared `testutil_test.go`.
   - Replace all clone instances with calls to the canonical implementation.
   - Ensure all extracted helpers use `t.Helper()`.
4. Run `go test -race ./...` after each consolidation to confirm no regressions.

### Phase 3: Validate
```bash
go-stats-generator analyze . --only-tests --format json --output post.json --sections duplication --min-block-lines 6 --similarity-threshold 0.80
go-stats-generator diff baseline.json post.json
```
Confirm: duplication ratio decreased, zero regressions, all tests pass.

## Thresholds (Test-Appropriate)
| Metric | Target |
|--------|--------|
| Duplication ratio | <10% |
| Min block size | 6 lines |
| Similarity threshold | 0.80 |

> **Note**: Test duplication threshold is relaxed to 10% (vs 5% for production code) since some test repetition is acceptable for clarity.

## Deduplication Rules
- Start with the shortest clone group within each priority tier — simpler consolidations are safer and may collapse larger clones.
- **Table-driven tests** are the preferred deduplication strategy for test code.
- Preserve all existing test coverage and pass/fail behavior.
- Do not merge test cases that validate different conceptual behaviors even if textually similar.
- Each extracted helper must use `t.Helper()` and be <45 lines.
- Add comments to extracted helpers explaining what they validate.

## Clone Types
| Type | Description | Strategy |
|------|-------------|----------|
| Exact | Identical test blocks | Table-driven consolidation |
| Renamed | Same structure, different variable names | Table-driven with parameters |
| Near-duplicate | Similar structure, minor logic differences | Extract helper with config parameter |

## Output Format
```
Clone group [N]: [file1:line]-[file2:line] ([M] lines, [K] instances)
  Strategy: [table-driven | extract helper | extract testutil]
  Consolidated into: [new_helper_name]
  Tests: PASS
Duplication: [before]% -> [after]%
```

## Tiebreaker
Within each priority tier, consolidate the shortest clone group first.
## Validation Checklist
- [ ] Duplication ratio decreased
- [ ] No new test functions introduced above complexity thresholds
- [ ] All tests pass with -race flag
- [ ] No test coverage reduced
- [ ] Each extracted helper uses `t.Helper()`
- [ ] All clone instances replaced with calls to canonical implementation
- [ ] Diff report shows zero complexity regressions
