# TASK: [Ebitengine Edition] Execute ONE highest-priority incomplete task from the project's backlog with full implementation and baseline/diff validation.

## Execution Mode
**Autonomous action** — implement the task, validate with tests and diff.

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

### Phase 0: Understand the Project
1. Read the project README to understand its domain, users, and architecture.
2. Examine `go.mod` for module path and dependency profile.
3. Discover the project's coding conventions: error handling style, naming patterns, test strategy.
4. Find the project's backlog: roadmap files, issue tracker, TODO comments, or milestone documents.

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --output baseline.json --sections functions,duplication,documentation
```

### Phase 2: Select and Implement
1. Identify the highest-priority incomplete task from the project's backlog.
2. Understand the task requirements and acceptance criteria.
3. Implement following the project's established conventions:
   - Match the codebase's error handling style and naming patterns.
   - Respect established function length and complexity norms (default targets: <=30 lines, cyclomatic <=10).
   - Add GoDoc comments on all exported symbols.
   - Prefer stdlib over external dependencies unless the project already uses a relevant library.
4. Preserve all existing public API signatures.
5. Run `go test -race ./...` and `go vet ./...` after implementation.

### Phase 3: Validate
```bash
go-stats-generator analyze . --skip-tests --format json --output post.json --sections functions,duplication,documentation
go-stats-generator diff baseline.json post.json
```
Confirm: zero regressions in complexity, duplication, or doc coverage.

## Default Thresholds (calibrate to project)
| Metric | Target |
|--------|--------|
| Cyclomatic complexity | <=10 |
| Function length | <=30 lines |
| Doc coverage | >=70% |
| Duplication ratio | <5% |

## Implementation Rules
- Execute exactly ONE task per invocation.
- Preserve all existing public APIs.
- Match the project's naming conventions.
- Add tests for new functionality where practical.
- Do not introduce new dependencies without justification.

## Mark Completion
After successful implementation and validation:
- Check off the completed item in the backlog file (`- [x]`).
- Note the diff summary in your output.

## Output Format
```
Task: [description from backlog]
Files modified: [list]
Tests: PASS
Diff summary: [key metric changes]
Remaining: [count of incomplete items]
```

## Tiebreaker
Take the highest-priority incomplete task. If tied, choose the task with broadest file impact.
