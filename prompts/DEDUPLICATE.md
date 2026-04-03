# TASK: Identify and consolidate the top 5–10 most significant code clone groups below duplication thresholds.

## Execution Mode
**Autonomous action** — deduplicate code, validate with tests and diff.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Codebase
1. Read the project README to understand its domain and architecture.
2. Identify the project's coding patterns and idioms — consolidation must produce idiomatic code.
3. Discover whether the project uses specific patterns (functional options, builders, table dispatch) that inform the consolidation strategy.
4. Note which packages are core vs. utility — duplication in core code is higher priority.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output baseline.json --sections duplication --min-block-lines 6 --similarity-threshold 0.80
go-stats-generator analyze . --skip-tests
```

### Phase 2: Deduplicate
1. Extract `.duplication.clone_pairs` sorted by line count ascending (smallest first).
2. Classify clone groups by priority:
   - CRITICAL: >=20 lines AND >=3 instances
   - HIGH: >=10 lines AND >=2 instances
   - MEDIUM: >=6 lines AND >=2 instances
3. For each clone group (starting with the shortest/simplest):
   - Identify the shared logic and create a single canonical implementation.
   - Choose the consolidation strategy:
     - **Extract function**: move shared code into a new helper.
     - **Extract method**: if clones are in methods on the same type.
     - **Table-driven pattern**: if clones differ only in inputs/expected outputs.
   - Do NOT merge clones that serve different conceptual purposes even if textually similar.
   - Replace all instances with calls to the canonical implementation.
   - Name helpers matching the project's conventions (default: verb-first).
4. Run `go test -race ./...` after each consolidation.

### Phase 3: Validate
```bash
go-stats-generator analyze . --skip-tests --format json --output post.json --sections duplication --min-block-lines 6 --similarity-threshold 0.80
go-stats-generator diff baseline.json post.json
```
Confirm: duplication ratio decreased, zero regressions, all tests pass.

## Default Thresholds (calibrate to project)
| Metric | Target |
|--------|--------|
| Duplication ratio | <5% |
| Min block size | 6 lines |
| Similarity threshold | 0.80 |

## Clone Types
| Type | Description | Strategy |
|------|-------------|----------|
| Exact | Identical code blocks | Direct extraction |
| Renamed | Same structure, different variable names | Extract with parameters |
| Near-duplicate | Similar structure, minor logic differences | Extract with config parameter or callback |

## Deduplication Rules
- Start with the shortest clone group per tier — simpler consolidations are safer.
- Preserve all existing public API signatures.
- Each extracted helper must be <30 lines with GoDoc comments.

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
