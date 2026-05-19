# TASK: [Ebitengine Edition] Identify and refactor the top 5–10 most complex functions below professional complexity thresholds.

## Execution Mode
**Autonomous action** — refactor functions, validate with tests and diff.

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
Before refactoring, understand what you're changing:
1. Read the project README to learn the project's domain and purpose.
2. Examine `go.mod` for dependencies and Go version.
3. Identify the project's coding patterns: do they use builders, functional options, table-driven dispatch, or other idioms? Refactored code must match.
4. Note the project's error handling style and test strategy — extracted helpers must be consistent.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output baseline.json --sections functions --max-complexity 9 --max-function-length 40
go-stats-generator analyze . --skip-tests --max-complexity 9 --max-function-length 40
```

### Phase 2: Refactor
1. Select the top 5–10 functions exceeding thresholds (sorted by overall complexity descending).
2. For each target function, **understand its role** in the project before refactoring:
   - What does this function do? Is it a parser, handler, orchestrator, or algorithm?
   - A complex parser or state machine may warrant higher thresholds than a simple handler.
3. Apply extract-method refactoring idiomatically:
   - Identify cohesive blocks (loop bodies, conditional branches, setup/teardown, error paths).
   - Extract into named helpers matching the project's naming conventions (default: verb-first).
   - Each extracted function: <20 lines, cyclomatic <8 (or <6 for Update/Draw methods) (tunable defaults).
   - Preserve all existing public API signatures.
4. Run `go test -race ./...` after each refactoring.

### Phase 3: Validate
```bash
go-stats-generator analyze . --skip-tests --format json --output post.json --sections functions --max-complexity 9 --max-function-length 40
go-stats-generator diff baseline.json post.json
```
Confirm: zero regressions, all target functions now below thresholds.

## Complexity Formula
```
Overall = Cyclomatic + (NestingDepth * 0.5) + (Cognitive * 0.3)
```
Where Cognitive currently equals Cyclomatic.

## Default Thresholds (calibrate to project baseline)
| Metric | Warning | Critical |
|--------|---------|----------|
| Overall complexity | >9.0 | >15.0 |
| Cyclomatic complexity | >9 | >15 |
| Function length (code lines) | >40 | >80 |
| Nesting depth | >3 | >5 |
| Extracted function length | — | >20 |
| Extracted function cyclomatic | — | >8 |

### Ebitengine-Specific Refactoring Patterns
- **Extract Update Logic**: Move entity update logic into separate systems
- **Extract Draw Logic**: Separate rendering concerns from game state
- **State Pattern**: Complex Update() methods should use state machines
- **Component Systems**: Break monolithic entities into component-based architecture
- **Spatial Partitioning**: Extract collision detection into dedicated spatial structures

## Refactoring Rules
- **Extract method**: move cohesive blocks into named helpers.
- **Decompose conditional**: replace complex boolean chains with predicate functions.
- **Replace loop body**: extract inner loop logic into a function.
- **Consolidate error handling**: merge repeated error patterns into a shared helper.
- Match the project's naming conventions (default: verb-first, e.g., `buildDependencyMap`).
- Never change exported function signatures.
- Add GoDoc to extracted functions with >3 lines of logic.

## Output Format
```
[function] [file]: [old_complexity] -> [new_complexity] ([reduction_%])
  Extracted: [helper_1], [helper_2], ...
  Tests: PASS
```

## Tiebreaker
When complexity scores are tied, refactor the longest function first.
