# TASK: [Ebitengine Edition] Analyze how well this codebase achieves its stated goals and generate a prioritized roadmap for closing the gaps.

## Execution Mode
**Report generation only** — produce a roadmap document. Do not modify source code.

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

## Output
Write exactly one file: **`ROADMAP.md`** in the repository root (the directory containing `go.mod`).
If `ROADMAP.md` already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

### Ebitengine-Specific Audit Criteria

#### Rendering & Graphics
- [ ] No `ebiten.NewImage()` or `ebiten.NewImageFromImage()` calls inside `Draw()` or `Update()` (per-frame allocation anti-pattern)
- [ ] Image atlases and sprite sheets loaded once during initialization
- [ ] `DrawImageOptions` pooled or reused, not allocated per sprite
- [ ] `GeoM` and `ColorM` transformations applied efficiently
- [ ] Screen clearing strategy appropriate (`SetScreenClearedEveryFrame` usage)
- [ ] Offscreen images used appropriately for caching/render-to-texture

#### Input Handling
- [ ] Input state checked at beginning of `Update()`, not in `Draw()`
- [ ] `inpututil` used for frame-accurate press/release detection
- [ ] Gamepad connection state validated before reading inputs
- [ ] Touch/mouse coordinates adjusted for window DPI scaling
- [ ] Input handling accounts for both desktop and mobile platforms

#### Audio Management
- [ ] Audio players created during initialization, not per-frame
- [ ] Audio player `Close()` deferred appropriately
- [ ] Streaming audio used for music, buffered for short SFX
- [ ] Audio context sample rate matches expected platform rates

#### Game Loop & Timing
- [ ] `Update()` and `Draw()` methods have appropriate complexity (<15 cyclomatic for frame budget)
- [ ] TPS (ticks per second) set appropriately with `ebiten.SetTPS()`
- [ ] No blocking operations in `Update()` or `Draw()` (I/O, network, heavy computation)
- [ ] State transitions don't cause dropped frames
- [ ] Delta time handling if using variable time steps

#### Resource Management
- [ ] Images disposed via `Dispose()` when no longer needed
- [ ] Global image cache has bounded size
- [ ] Scene transitions clean up previous scene resources
- [ ] No resource leaks when switching between game states

#### Mobile & Cross-Platform
- [ ] `Layout()` returns consistent logical dimensions
- [ ] Touch input implemented alongside mouse input
- [ ] UI elements sized appropriately for touch targets (44×44+ logical pixels)
- [ ] Text rendering readable on high-DPI displays
- [ ] Back button handling on Android (`ebiten.AppendInputChars` for back)

## ## Workflow

### Phase 0: Discover Project Goals
Before assessing anything, understand what the project is trying to accomplish:
1. Read the project README thoroughly — extract every stated goal, feature claim, capability promise, performance target, and audience statement. These are the **acceptance criteria** for the review.
2. Examine `go.mod` for module path, Go version, and dependency footprint.
3. Scan for existing CI (`.github/workflows/`, `.gitlab-ci.yml`, `Makefile`) and note what quality checks already run.
4. List packages (`go list ./...`) and identify the architectural layers and their responsibilities.
5. Check for an existing roadmap, backlog, issue tracker, or changelog that reveals maintainer priorities and planned work.
6. Look for design documents, ADRs, or spec files that clarify intent beyond the README.

### Phase 1: Online Research
Use web search to build context that isn't available in the repository:
1. Search for the project on GitHub — read open issues, recent PRs, and community discussions to understand known pain points and user feedback.
2. Research key dependencies from `go.mod` for known vulnerabilities, deprecations, or upcoming breaking changes.
3. Look up best practices and conventions in the project's domain to calibrate expectations for its stated goals.
4. Check whether comparable tools exist — understanding the competitive landscape helps evaluate whether the project's goals are ambitious, typical, or outdated.

Keep research brief (≤10 minutes). Record only findings that are directly relevant to the project's stated goals.

### Phase 2: Metrics Collection
```bash
mkdir -p tmp
go-stats-generator analyze . --skip-tests --format json > tmp/review-metrics.json
go-stats-generator analyze . --skip-tests
```
Delete `tmp/review-metrics.json` when done — the only persistent output is `ROADMAP.md`.

### Phase 3: Goal-Achievement Assessment
For each stated goal or feature claim discovered in Phase 0, evaluate:

1. **Does the feature exist?** Trace through the codebase to confirm the claimed functionality is implemented, not just stubbed.
2. **Does it work correctly?** Cross-reference with `go-stats-generator` metrics to identify risk areas:
   - Functions with cyclomatic complexity >15 or length >50 lines on critical paths are high-risk for bugs.
   - Packages with <70% doc coverage may have undocumented behavioral differences from what the README claims.
   - Cross-reference `.duplication` to find areas where copy-paste may have introduced behavioral drift.
3. **Does it meet its own stated quality bar?** If the project claims performance targets, test coverage levels, or scalability guarantees, verify them with evidence.
4. **Are there gaps between ambition and reality?** Identify features that are documented but incomplete, partially implemented, or non-functional.

Run `go test -race ./...` and `go vet ./...` to confirm baseline health.

When `go vet` or linters report warnings, read the comments surrounding the flagged code before reporting a finding. If a comment explicitly acknowledges the warning (e.g., `//nolint:`, an explanatory comment justifying the pattern, or a TODO tracking the known issue), treat it as an acknowledged false positive and do not report it as a new finding.

Use the project's own conventions and architecture as the standard — do not impose external standards that the project does not claim to follow.

### Phase 4: Generate ROADMAP.md
```markdown
# Goal-Achievement Assessment

## Project Context
- **What it claims to do**: [summary from README]
- **Target audience**: [who the project serves]
- **Architecture**: [key packages and their roles]
- **Existing CI/quality gates**: [list]

## Goal-Achievement Summary
| Stated Goal | Status | Evidence | Gap Description |
|-------------|--------|----------|-----------------|
| [Goal from README] | ✅ Achieved / ⚠️ Partial / ❌ Missing | [metric or code reference] | [what's missing, if anything] |

**Overall: [N]/[total] goals fully achieved**

## Roadmap
### Priority 1: [Most impactful unachieved or partially achieved goal]
- [ ] [Specific step with file/function reference]
- [ ] [Validation: how to confirm this goal is now achieved]

### Priority 2: [Next most impactful gap]
- [ ] ...
```

## Review Rules
- **Goal-first**: Every finding must trace back to a stated goal or feature claim. Do not invent requirements the project does not claim.
- **Evidence-based**: Use `go-stats-generator` metrics as quantitative evidence. Cite specific files, functions, and metric values.
- **Context-sensitive**: A 50-line function in a parser may be perfectly fine. A high complexity score in a utility function is more concerning. Use the project's own baseline distribution to calibrate.
- Do NOT recommend TLS, HTTPS, or transport-layer encryption — transport security is handled by infrastructure.
- Prioritize roadmap items by how much they would advance the project's stated goals, not by arbitrary code-quality checklists.

## Tiebreaker
Prioritize gaps that affect the most users or the most critical claimed functionality first. Within a priority level, address the highest-risk items (most complexity, least test coverage) first.
