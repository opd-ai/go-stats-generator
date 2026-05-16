# TASK: [Ebitengine Edition] Use `go-stats-generator` to analyze and refactor the target codebase (dogfooding) — reduce complexity of functions below thresholds while verifying the build still works.

## Execution Mode
**Autonomous action** — refactor and fix discovered bugs, validate with tests, rebuild, and diff.

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
1. Read the project README to understand purpose, domain, and build process.
2. Examine `go.mod` and discover the build/install command (e.g., `go build`, `go install`, `make build`).
3. Identify the project's coding patterns, naming conventions, and error handling style.
4. Note whether the project is a tool/binary (requires rebuild verification) or a library (build verification sufficient).

### Phase 1: Self-Analysis Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output baseline.json --sections functions --max-complexity 10 --max-function-length 30
go-stats-generator analyze . --skip-tests --max-complexity 10 --max-function-length 30
```

### Phase 2: Refactor
1. Identify functions exceeding thresholds from the baseline.
2. For each violating function (sorted by overall complexity descending):
   - **Understand its role** before refactoring — read callers and context.
   - Apply extract-method refactoring matching the project's idioms:
     - Extract cohesive blocks into named helpers (<20 lines, cyclomatic <8 (or <6 for Update/Draw methods)).
     - Preserve all public API signatures.
   - If a bug is discovered during review, fix it as part of this pass.
3. Run `go test -race ./...` after each refactoring.
4. Run `go vet ./...` to confirm no issues.
5. If the project is a buildable tool, rebuild and verify it still works:
   ```bash
   go build ./... && echo "BUILD PASS"
   ```

### Phase 3: Validate
```bash
go-stats-generator analyze . --skip-tests --format json --output post.json --sections functions --max-complexity 10 --max-function-length 30
go-stats-generator diff baseline.json post.json
```
Confirm: all target functions below thresholds, zero regressions, project still builds and runs.

## Default Thresholds (calibrate to project)
| Metric | Maximum |
|--------|---------|
| Overall complexity | 10.0 |
| Cyclomatic complexity | 10 |
| Function length | 30 lines |
| Nesting depth | 3 |
| Extracted function length | 20 |
| Extracted function cyclomatic | 8 |

## Dogfooding Rules
- After refactoring, rebuild the project and verify it still produces correct output.
- If the project's own output changes (beyond metric improvements), investigate whether a bug was introduced.
- May fix bugs discovered during analysis (unlike BREAKDOWN.md which is refactor-only).

## Output Format
```
Analysis: [N] functions above thresholds
Refactored:
  [function] [file]: [old] -> [new] ([reduction]%)
Bugs fixed: [count] (or "none")
Build verification: PASS
Tests: PASS
```

## Tiebreaker
Refactor the longest function first when complexity scores are tied.
