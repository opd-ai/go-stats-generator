# TASK: Identify and consolidate the top 5–10 most significant code clone groups below duplication thresholds.

## Execution Mode
**Autonomous action** — deduplicate code, validate with tests and diff.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output baseline.json --sections duplication --min-block-lines 6 --similarity-threshold 0.80
go-stats-generator analyze . --skip-tests
```

### Phase 2: Deduplicate
1. From the baseline JSON, extract `.duplication.clone_pairs` sorted by line count ascending (smallest first).
2. Classify clone groups by priority:
   - CRITICAL: >=20 lines AND >=3 instances
   - HIGH: >=10 lines AND >=2 instances
   - MEDIUM: >=6 lines AND >=2 instances
3. For each clone group (starting with the shortest/simplest):
   - Identify the shared logic and create a single canonical implementation.
   - Choose the consolidation strategy:
     - **Extract function**: move shared code into a new helper function.
     - **Extract method**: if clones are in methods on the same type.
     - **Table-driven pattern**: if clones differ only in inputs/expected outputs (common in tests).
   - Replace all clone instances with calls to the canonical implementation.
   - Name helpers with verb-first Go conventions.
4. Run `go test -race ./...` after each consolidation to confirm no regressions.

### Phase 3: Validate
```bash
go-stats-generator analyze . --skip-tests --format json --output post.json --sections duplication --min-block-lines 6 --similarity-threshold 0.80
go-stats-generator diff baseline.json post.json
```
Confirm: duplication ratio decreased, zero regressions, all tests pass.

## Thresholds
| Metric | Target |
|--------|--------|
| Duplication ratio | <5% |
| Min block size | 6 lines |
| Similarity threshold | 0.80 |

## Deduplication Rules
- Start with the shortest clone group within each priority tier — simpler consolidations are safer and may collapse larger clones.
- Preserve all existing public API signatures.
- Do not merge clones that serve different conceptual purposes even if textually similar.
- Each extracted helper must be <30 lines.
- Add GoDoc comments to extracted helpers.

## Clone Types
| Type | Description | Strategy |
|------|-------------|----------|
| Exact | Identical code blocks | Direct extraction |
| Renamed | Same structure, different variable names | Extract with parameters |
| Near-duplicate | Similar structure, minor logic differences | Extract with config parameter or callback |

## Output Format
```
Clone group [N]: [file1:line]-[file2:line] ([M] lines, [K] instances)
  Strategy: [extract function | table-driven | extract method]
  Consolidated into: [new_function_name]
  Tests: PASS
Duplication: [before]% -> [after]%
```

## Tiebreaker
Within each priority tier, consolidate the shortest clone group first.
## Validation Checklist
- [ ] Duplication ratio decreased
- [ ] No new functions introduced above complexity thresholds
- [ ] All tests pass with -race flag
- [ ] No exported API signatures changed
- [ ] Each extracted helper has a GoDoc comment
- [ ] All clone instances replaced with calls to canonical implementation
- [ ] Diff report shows zero complexity regressions
