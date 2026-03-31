# TASK: Identify and consolidate ALL code clone groups above duplication thresholds — no cap per session.

## Execution Mode
**Autonomous action** — deduplicate every clone group above threshold, validate with tests and diff. Do not stop at a fixed count; continue until every flagged clone group is resolved or the session's context is exhausted.

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
5. Examine `go.mod` for dependencies and Go version.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output baseline.json --sections duplication --min-block-lines 4 --similarity-threshold 0.75
go-stats-generator analyze . --skip-tests --min-block-lines 4 --similarity-threshold 0.75
```

### Phase 2: Build the Full Worklist
1. From `baseline.json`, extract `duplication.clones[]` and sort the clone groups by `line_count` descending (largest first for maximum impact).
2. Do not cap the list — include all clone groups above thresholds.
3. Classify clone groups by priority:
   - CRITICAL: >=15 lines AND >=3 instances
   - HIGH: >=8 lines AND >=2 instances
   - MEDIUM: >=4 lines AND >=2 instances
4. Group clone pairs by package/file for efficient batch consolidation (touching nearby code in one pass avoids redundant test cycles).

### Phase 3: Deduplicate (Iterate Until Exhausted)
Loop over the worklist. For each clone group:
1. **Understand the clone's role** before consolidating:
   - What does this duplicated code do? Is it setup, validation, transformation, or error handling?
   - Do NOT merge clones that serve different conceptual purposes even if textually similar.
2. Choose the consolidation strategy:
   - **Extract function**: move shared code into a new helper.
   - **Extract method**: if clones are in methods on the same type.
   - **Table-driven pattern**: if clones differ only in inputs/expected outputs.
   - **Parameterize**: if clones differ by a small number of values, extract with parameters.
   - **Generic function**: if clones are identical except for types (Go 1.18+).
3. Replace all instances with calls to the canonical implementation.
4. Name helpers matching the project's conventions (default: verb-first).
5. Each extracted helper must be <30 lines with GoDoc comments.
6. Run `go vet ./...` after each consolidation to catch mistakes early.
7. Run `go test -race ./...` after every batch of consolidations within the same package (or after each individual consolidation if the clone touches critical paths).
8. If tests fail, fix immediately before moving on — never accumulate broken state.

**Stopping conditions** (in priority order):
- ✅ Worklist exhausted — every flagged clone group is now consolidated.
- ⚠️ Context/session boundary reached — commit progress and document remaining items.
- 🛑 Unrecoverable test regression — stop, revert the last consolidation, and report.

### Phase 4: Validate
```bash
go-stats-generator analyze . --skip-tests --format json --output post.json --sections duplication --min-block-lines 4 --similarity-threshold 0.75
go-stats-generator diff baseline.json post.json
```
Confirm: duplication ratio decreased, zero regressions, all tests pass.

## Default Thresholds (calibrate to project baseline)
| Metric | Target |
|--------|--------|
| Duplication ratio | <3% |
| Min block size | 4 lines |
| Similarity threshold | 0.75 |

> Thresholds are more aggressive than the standard DEDUPLICATE prompt (3% vs 5%, 4-line blocks vs 6-line, 0.75 similarity vs 0.80). Calibrate to the project's actual baseline on first run.

## Clone Categories (manual / heuristic)
These are refactoring-oriented categories you apply manually based on `go-stats-generator` duplication output. The tool itself only classifies clones as `exact`, `renamed`, or `near`; "Type-variant" is an additional pattern you identify by inspection.
| Type | Description | Strategy |
|------|-------------|----------|
| Exact | Identical code blocks | Direct extraction |
| Renamed | Same structure, different variable names | Extract with parameters |
| Near-duplicate | Similar structure, minor logic differences | Extract with config parameter or callback |
| Type-variant (manual) | Same logic, different concrete types (not a separate clone type emitted by `go-stats-generator`) | Generic function (Go 1.18+) |

## Deduplication Rules
- Process CRITICAL clone groups first, then HIGH, then MEDIUM.
- Within each tier, consolidate the largest clone group first (maximum line reduction per action).
- Preserve all existing public API signatures.
- Each extracted helper must be <30 lines with GoDoc comments.
- When `go vet` or linters report warnings, read the comments surrounding the flagged code. If a comment explicitly acknowledges the warning (e.g., `//nolint:`, an explanatory comment justifying the pattern, or a TODO tracking the known issue), treat it as an acknowledged false positive — do not report it as a new finding.

## Output Format
```
Clone group [N]: [file1:line]-[file2:line] ([M] lines, [K] instances)
  Strategy: [extract function | table-driven | extract method | parameterize | generic]
  Consolidated into: [new_function_name]
  Tests: PASS
```

After all consolidations, print a summary table:
```
Total clone groups consolidated: N
Total extracted helpers:          M
Remaining above threshold:        R (with justifications if R > 0)
Duplication: [before]% -> [after]%
```

## Tiebreaker
Within each priority tier, consolidate the largest clone group first (maximum line savings per action).

## Session Strategy
- Prioritize breadth: eliminate every clone group rather than perfecting a few.
- Batch consolidations within the same file/package to minimize test cycle overhead.
- Commit working progress frequently so partial sessions still deliver value.
- If the session is running low on context, commit and document the remaining worklist for the next session.
