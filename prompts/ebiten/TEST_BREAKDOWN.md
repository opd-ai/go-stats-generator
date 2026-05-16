# TASK: [Ebitengine Edition] Identify and refactor the top 5–10 most complex **test** functions below test-appropriate complexity thresholds.

## Execution Mode
**Autonomous action** — refactor test functions, validate with tests and diff.

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

### Phase 0: Understand the Test Strategy
1. Read the project README and discover the testing philosophy: unit-focused, integration-heavy, or BDD?
2. Identify the test framework in use (`testing` only, `testify`, `gomock`, etc.).
3. Discover existing test conventions: how are helpers structured, do they use `t.Helper()`, table-driven tests, `t.Parallel()`?
4. Note the assertion style — refactored tests must match existing patterns.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --only-tests --format json --output baseline.json --sections functions --max-complexity 14 --max-function-length 60
go-stats-generator analyze . --only-tests --max-complexity 14 --max-function-length 60
```

### Phase 2: Refactor
1. Select the top 5–10 test functions exceeding thresholds (sorted by overall complexity descending).
2. For each target test function, apply test-appropriate refactoring matching the project's conventions:
   - Extract shared setup into test helpers using `t.Helper()`.
   - Convert repetitive assertions into table-driven subtests with `t.Run`.
   - Extract complex assertion logic into named helper functions.
   - Each extracted helper: <30 lines, cyclomatic <12 (tunable defaults).
   - Preserve all existing test coverage and pass/fail behavior.
3. Run `go test -race ./...` after each refactoring.

### Phase 3: Validate
```bash
go-stats-generator analyze . --only-tests --format json --output post.json --sections functions --max-complexity 14 --max-function-length 60
go-stats-generator diff baseline.json post.json
```
Confirm: zero regressions, all target test functions now below thresholds.

## Default Thresholds (test-appropriate — ~50% relaxed vs. production)
| Metric | Warning | Critical |
|--------|---------|----------|
| Overall complexity | >15.0 | >22.0 |
| Cyclomatic complexity | >15 | >22 |
| Function length | >45 | >90 |
| Nesting depth | >5 | >7 |
| Extracted helper length | — | >30 |
| Extracted helper cyclomatic | — | >12 |


### Ebitengine-Specific Refactoring Patterns
- **Extract Update Logic**: Move entity update logic into separate systems
- **Extract Draw Logic**: Separate rendering concerns from game state
- **State Pattern**: Complex Update() methods should use state machines
- **Component Systems**: Break monolithic entities into component-based architecture
- **Spatial Partitioning**: Extract collision detection into dedicated spatial structures

## Refactoring Rules
- **Table-driven tests**: preferred strategy — consolidate repetitive cases into `[]struct{...}` with `t.Run`.
- **Test helpers**: extract shared setup/teardown into functions marked with `t.Helper()`.
- **Assertion helpers**: extract complex assertion sequences into named helpers.
- Match the project's existing test naming and assertion patterns.
- Never change test coverage or pass/fail behavior.

## Output Format
```
[function] [file]: [old_complexity] -> [new_complexity] ([reduction_%])
  Extracted: [helper_1], [helper_2], ...
  Tests: PASS
```

## Tiebreaker
When complexity scores are tied, refactor the longest test function first.
