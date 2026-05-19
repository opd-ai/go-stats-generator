# TASK: [Ebitengine Edition] Identify and refactor ALL functions exceeding complexity thresholds AND consolidate ALL code clone groups — no cap per session.

## Execution Mode
**Autonomous action** — refactor every function above threshold and deduplicate every clone group above threshold, validate with tests and diff. Do not stop at a fixed count; continue until every flagged function and clone group is resolved or the session's context is exhausted.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Ebitengine-Specific Context

This prompt variant is optimized for Go codebases using the Ebitengine (github.com/hajimehoshi/ebiten/v2) game framework. When analyzing code, prioritize game-specific patterns and concerns:

### Ebitengine Architecture Patterns
- **Game Interface**: Implementations of `ebiten.Game` interface with `Update()` and `Draw(ebiten.Image)` methods
- **Resource Lifecycle**: Image, audio, and asset management across game state transitions
- **Frame Timing**: 60 TPS (ticks per second) default update cycle, vsync-based drawing
- **Coordinate Systems**: Screen coordinates vs logical coordinates, `Layout()` method behavior

### Performance-Critical Areas
- **Draw Method**: Must complete within frame budget (~16.67ms at 60fps); avoid allocations
- **Update Method**: Game logic execution; profile for O(n²) entity interactions
- **Image Creation**: `ebiten.NewImage()` calls should be cached, not per-frame
- **DrawImageOptions**: Reuse `DrawImageOptions` instances via sync.Pool to reduce GC pressure

### Common Ebitengine Patterns
- **Input Handling**: `inpututil` for press/release detection, `ebiten.IsKeyPressed()` for state
- **Sprite Rendering**: `Image.DrawImage()` with `GeoM` for transforms, `ColorM` for tinting
- **Audio**: `audio.Player` lifecycle management, streaming vs. buffered playback
- **Text Rendering**: `text.Draw()` or `ebitenutil.DebugPrint()` considerations
- **Collision Detection**: AABB, circle, and pixel-perfect collision in Update()

## Workflow

### Phase 0: Understand the Codebase
Before refactoring or deduplicating, understand what you're changing:
1. Read the project README to learn the project's domain, purpose, and architecture.
2. Examine `go.mod` for dependencies and Go version.
3. Identify the project's coding patterns: do they use builders, functional options, table-driven dispatch, or other idioms? Refactored and consolidated code must match.
4. Note the project's error handling style and test strategy — extracted helpers must be consistent.
5. Identify which packages are core vs. utility — duplication in core code is higher priority.

### Phase 1: Baseline
```bash
# Analyze functions for complexity
go-stats-generator analyze . --skip-tests --format json --output baseline-functions.json --sections functions --max-complexity 9 --max-function-length 40

# Analyze duplication
go-stats-generator analyze . --skip-tests --format json --output baseline-duplication.json --sections duplication --min-block-lines 4 --similarity-threshold 0.75

# Display console report
go-stats-generator analyze . --skip-tests --max-complexity 9 --max-function-length 40 --min-block-lines 4 --similarity-threshold 0.75
```

### Phase 2: Build the Full Worklist
1. **Function Complexity Worklist**:
   - Collect **every** function that exceeds any threshold (sorted by overall complexity descending).
   - Do not cap the list — include all violators.
   - Group functions by package/file for efficient batch refactoring.

2. **Duplication Worklist**:
   - From `baseline-duplication.json`, extract `duplication.clones[]` and sort by `line_count` descending.
   - Do not cap the list — include all clone groups above thresholds.
   - Classify clone groups by priority:
     - CRITICAL: >=15 lines AND >=3 instances
     - HIGH: >=8 lines AND >=2 instances
     - MEDIUM: >=4 lines AND >=2 instances
   - Group clone pairs by package/file for efficient batch consolidation.

3. **Prioritization Strategy**:
   - Process CRITICAL duplication clones first (highest impact).
   - Then process HIGH complexity functions (most tangled logic).
   - Then process HIGH duplication clones.
   - Then process remaining complexity violations and MEDIUM duplication.
   - This ordering maximizes both readability gains and code reduction.

### Phase 3: Refactor and Deduplicate (Iterate Until Exhausted)

#### For Function Complexity Refactoring:
Loop over the complexity worklist. For each target function:
1. **Understand its role** in the project before refactoring:
   - What does this function do? Is it a parser, handler, orchestrator, or algorithm?
   - A complex parser or state machine may warrant higher thresholds than a simple handler.
2. Apply extract-method refactoring idiomatically:
   - Identify cohesive blocks (loop bodies, conditional branches, setup/teardown, error paths).
   - Extract into named helpers matching the project's naming conventions (default: verb-first).
   - Each extracted function: <20 lines, cyclomatic <8 (or <6 for Update/Draw methods) (tunable defaults).
   - Preserve all existing public API signatures.
3. Run `go vet ./...` after each refactoring to catch mistakes early.
4. Run `go test -race ./...` after every batch of refactorings within the same package (or after each individual refactoring if the function is high-risk).
5. If tests fail, fix immediately before moving on — never accumulate broken state.

#### For Duplication Consolidation:
Loop over the duplication worklist. For each clone group:
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
- ✅ Both worklists exhausted — every flagged function and clone group is now below thresholds.
- ⚠️ Context/session boundary reached — commit progress and document remaining items.
- 🛑 Unrecoverable test regression — stop, revert the last change, and report.

### Phase 4: Validate
```bash
# Re-analyze functions
go-stats-generator analyze . --skip-tests --format json --output post-functions.json --sections functions --max-complexity 9 --max-function-length 40

# Re-analyze duplication
go-stats-generator analyze . --skip-tests --format json --output post-duplication.json --sections duplication --min-block-lines 4 --similarity-threshold 0.75

# Show improvements
go-stats-generator diff baseline-functions.json post-functions.json
go-stats-generator diff baseline-duplication.json post-duplication.json
```
Confirm: zero regressions, all target functions now below complexity thresholds, duplication ratio decreased, all tests pass.

## Complexity Formula (as computed by go-stats-generator)
```
Overall = Cyclomatic + (NestingDepth * 0.5) + (Cognitive * 0.3)
```
Where `Cognitive` is currently set equal to `Cyclomatic`, so this simplifies to:
```
Overall = Cyclomatic * 1.3 + (NestingDepth * 0.5)
```

**Note**: The tool does NOT include function length (lines) or signature complexity in the overall score calculation. These are tracked as separate metrics and can be used as independent thresholds for refactoring decisions.

## Default Thresholds (calibrate to project baseline)

### Function Complexity
| Metric | Warning | Critical | Notes |
|--------|---------|----------|-------|
| Overall complexity | >9.0 | >15.0 | Computed from cyclomatic + nesting (see formula above) |
| Cyclomatic complexity | >9 | >15 | Used in overall score calculation |
| Nesting depth | >3 | >5 | Used in overall score calculation |
| Function length (code lines) | >40 | >80 | Independent threshold, NOT in overall score |
| Extracted function length | — | >20 | Target for refactored helpers |
| Extracted function cyclomatic | — | >8 | Target for refactored helpers |

### Duplication
| Metric | Target |
|--------|--------|
| Duplication ratio | <3% |
| Min block size | 4 lines |
| Similarity threshold | 0.75 |

> Duplication thresholds are more aggressive than the standard DEDUPLICATE prompt (3% vs 5%, 4-line blocks vs 6-line, 0.75 similarity vs 0.80). Calibrate to the project's actual baseline on first run.

## Clone Categories (manual / heuristic)
These are refactoring-oriented categories you apply manually based on `go-stats-generator` duplication output. The tool itself only classifies clones as `exact`, `renamed`, or `near`; "Type-variant" is an additional pattern you identify by inspection.

| Type | Description | Strategy |
|------|-------------|----------|
| Exact | Identical code blocks | Direct extraction |
| Renamed | Same structure, different variable names | Extract with parameters |
| Near-duplicate | Similar structure, minor logic differences | Extract with config parameter or callback |
| Type-variant (manual) | Same logic, different concrete types (not a separate clone type emitted by `go-stats-generator`) | Generic function (Go 1.18+) |

## Refactoring and Deduplication Rules
- **Extract method**: move cohesive blocks into named helpers.
- **Decompose conditional**: replace complex boolean chains with predicate functions.
- **Replace loop body**: extract inner loop logic into a function.
- **Consolidate error handling**: merge repeated error patterns into a shared helper.
- **Process CRITICAL clone groups first**, then HIGH complexity functions, then HIGH clone groups, then remaining violations.
- **Within each tier**, consolidate the largest clone group first (maximum line reduction per action), or refactor the longest/most complex function first.
- Match the project's naming conventions (default: verb-first, e.g., `buildDependencyMap`).
- Never change exported function signatures.
- Add GoDoc to extracted functions with >3 lines of logic.
- Preserve all existing public API signatures.
- Each extracted helper for duplication must be <30 lines with GoDoc comments.
- Each extracted helper for complexity must be <20 lines, cyclomatic <8 (or <6 for Update/Draw methods).
- When `go vet` or linters report warnings, read the comments surrounding the flagged code. If a comment explicitly acknowledges the warning (e.g., `//nolint:`, an explanatory comment justifying the pattern, or a TODO tracking the known issue), treat it as an acknowledged false positive — do not report it as a new finding.

## Output Format

### For Function Refactoring:
```
[function] [file]: [old_complexity] -> [new_complexity] ([reduction_%])
  Extracted: [helper_1], [helper_2], ...
  Tests: PASS
```

### For Duplication Consolidation:
```
Clone group [N]: [file1:line]-[file2:line] ([M] lines, [K] instances)
  Strategy: [extract function | table-driven | extract method | parameterize | generic]
  Consolidated into: [new_function_name]
  Tests: PASS
```

### Summary Table (after all changes):
```
Total functions refactored:        N
Total clone groups consolidated:   M
Total extracted helpers:            X
Remaining functions above threshold: R₁ (with justifications if R₁ > 0)
Remaining clone groups above threshold: R₂ (with justifications if R₂ > 0)
Duplication: [before]% -> [after]%
```

## Tiebreaker
- **For complexity**: When complexity scores are tied, refactor the longest function first.
- **For duplication**: Within each priority tier, consolidate the largest clone group first (maximum line savings per action).

## Session Strategy
- Prioritize breadth: get every function below threshold and eliminate every clone group rather than perfecting a few.
- Batch refactorings and consolidations within the same file/package to minimize test cycle overhead.
- Commit working progress frequently so partial sessions still deliver value.
- If the session is running low on context, commit and document the remaining worklist for the next session.
- Address both complexity and duplication opportunistically: if refactoring a complex function also eliminates a clone group (or vice versa), that's a bonus — but don't force the combination.
