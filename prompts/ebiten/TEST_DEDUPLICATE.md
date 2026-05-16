# TASK: [Ebitengine Edition] Identify and consolidate the top 5–10 most significant code clone groups in **test files** below test-appropriate duplication thresholds.

## Execution Mode
**Autonomous action** — deduplicate test code, validate with tests and diff.

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
1. Extract `.duplication.clone_pairs` sorted by line count ascending (smallest first).
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

## Tiebreaker
Within each priority tier, consolidate the shortest clone group first.
