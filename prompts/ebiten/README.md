# Ebitengine-Optimized Prompts

This directory contains specialized variants of all prompts from the parent `prompts/` directory, optimized specifically for Go codebases using the [Ebitengine (Ebiten)](https://github.com/hajimehoshi/ebiten) game framework.

## Purpose

These prompts extend the standard prompts with game development-specific context, patterns, and best practices for Ebitengine applications. They are designed to help analyze, audit, refactor, and optimize Go game code built with Ebiten.

## What's Different?

Each Ebitengine prompt variant includes:

### 1. **Game-Specific Context**
- **Game Interface Patterns**: Focus on `ebiten.Game` interface implementations (`Update()`, `Draw()`, `Layout()`)
- **Resource Lifecycle**: Image, audio, and asset management across game state transitions
- **Frame Timing**: 60 TPS (ticks per second) default, vsync-based drawing, frame budget considerations
- **Coordinate Systems**: Screen coordinates vs logical coordinates

### 2. **Performance-Critical Areas**
- **Draw Method**: Must complete within frame budget (~16.67ms at 60fps); avoid allocations
- **Update Method**: Game logic execution; profile for O(n²) entity interactions
- **Image Creation**: `ebiten.NewImage()` calls should be cached, not per-frame
- **DrawImageOptions**: Reuse instances via `sync.Pool` to reduce GC pressure

### 3. **Ebitengine-Specific Audit Criteria**

#### Rendering & Graphics
- No `ebiten.NewImage()` or `ebiten.NewImageFromImage()` calls inside `Draw()` or `Update()`
- Image atlases and sprite sheets loaded once during initialization
- `DrawImageOptions` pooled or reused, not allocated per sprite
- `GeoM` and `ColorM` transformations applied efficiently
- Screen clearing strategy appropriate (`SetScreenClearedEveryFrame` usage)
- Offscreen images used appropriately for caching/render-to-texture

#### Input Handling
- Input state checked at beginning of `Update()`, not in `Draw()`
- `inpututil` used for frame-accurate press/release detection
- Gamepad connection state validated before reading inputs
- Touch/mouse coordinates adjusted for window DPI scaling
- Input handling accounts for both desktop and mobile platforms

#### Audio Management
- Audio players created during initialization, not per-frame
- Audio player `Close()` deferred appropriately
- Streaming audio used for music, buffered for short SFX
- Audio context sample rate matches expected platform rates

#### Game Loop & Timing
- `Update()` and `Draw()` methods have appropriate complexity (<15 cyclomatic for frame budget)
- TPS (ticks per second) set appropriately with `ebiten.SetTPS()`
- No blocking operations in `Update()` or `Draw()` (I/O, network, heavy computation)
- State transitions don't cause dropped frames
- Delta time handling if using variable time steps

#### Resource Management
- Images disposed via `Dispose()` when no longer needed
- Global image cache has bounded size
- Scene transitions clean up previous scene resources
- No resource leaks when switching between game states

#### Mobile & Cross-Platform
- `Layout()` returns consistent logical dimensions
- Touch input implemented alongside mouse input
- UI elements sized appropriately for touch targets (44×44+ logical pixels)
- Text rendering readable on high-DPI displays
- Back button handling on Android

## When to Use

Use these Ebitengine-optimized prompts when working with:

- **Game Projects**: Any Go codebase that imports `github.com/hajimehoshi/ebiten/v2`
- **Performance-Critical Code**: Game loops, rendering pipelines, entity systems
- **Cross-Platform Games**: Desktop, mobile, and web (WASM) builds
- **Real-Time Applications**: Anything requiring 60fps performance

## How to Use

### With CLI Tools

These prompts are designed for AI assistants to use when analyzing `go-stats-generator` output. The tool itself provides metrics, not prompt-based analysis:

```bash
# Generate metrics that can be analyzed using these prompts
go-stats-generator analyze . --format json > metrics.json

# Then use AI assistants with the Ebitengine-specific prompts
# Example: "Analyze metrics.json using prompts/ebiten/PERFORMANCE_AUDIT.md"
```

### With AI Assistants

When working with AI coding assistants (GitHub Copilot, Claude, ChatGPT, etc.), reference these prompts:

```
Use the prompt at prompts/ebiten/MEMORY_AUDIT.md to audit memory usage in my Ebitengine game code.
```

### Direct Usage

Simply copy the content of any prompt and use it as instructions for code review, refactoring, or audit tasks specific to Ebitengine game development.

## Prompt Categories

### Audits
- **Performance**: `PERFORMANCE_AUDIT.md` - Hot-path optimization, frame budget adherence
- **Memory**: `MEMORY_AUDIT.md` - GC pressure, image lifecycle, resource leaks
- **Concurrency**: `SYNC_AUDIT.md` - Goroutine safety in game loops, asset loading
- **Security**: `SECURITY_AUDIT.md` - Input validation, resource access
- **Game-Specific**: `GAME_DEBUG.md` - UI/UX defects, collision detection, input handling

### Refactoring
- **Complexity**: `BREAKDOWN.md` - Reduce complexity in `Update()` and `Draw()` methods
- **Deduplication**: `DEDUPLICATE.md` - Eliminate code duplication
- **Organization**: `ORGANIZE.md` - Improve code structure

### Testing
- **Test Coverage**: `TESTS.md` - Add tests for game logic
- **Test Quality**: `TEST_BUGS.md` - Fix test issues
- **Test Refactoring**: `TEST_BREAKDOWN.md` - Simplify complex tests

### Documentation
- **API Docs**: `DOCS.md` - Document game interfaces and systems
- **Code Comments**: `DOC.md` - Improve inline documentation

## Relationship to Standard Prompts

Each Ebitengine prompt corresponds to a standard prompt in the parent directory:

| Standard Prompt | Ebitengine Variant | Key Additions |
|-----------------|-------------------|---------------|
| `PERFORMANCE_AUDIT.md` | `ebiten/PERFORMANCE_AUDIT.md` | Frame budget analysis, per-frame allocation detection |
| `MEMORY_AUDIT.md` | `ebiten/MEMORY_AUDIT.md` | Image lifecycle, DrawImageOptions pooling |
| `BREAKDOWN.md` | `ebiten/BREAKDOWN.md` | Update/Draw complexity thresholds, ECS patterns |
| `GAME_DEBUG.md` | `ebiten/GAME_DEBUG.md` | Already Ebitengine-focused |

## Contributing

When adding new prompts to the parent `prompts/` directory:

1. Create a corresponding Ebitengine variant in this directory
2. Add game-specific context and audit criteria
3. Adjust complexity thresholds for frame-budget constraints
4. Include examples from Ebitengine APIs where relevant

## Resources

- [Ebitengine Official Documentation](https://ebitengine.org/en/documents/)
- [Ebitengine GitHub Repository](https://github.com/hajimehoshi/ebiten)
- [Ebitengine Examples](https://github.com/hajimehoshi/ebiten/tree/main/examples)
- [Ebitengine Wiki](https://github.com/hajimehoshi/ebiten/wiki)

## License

These prompts follow the same license as the parent repository.
