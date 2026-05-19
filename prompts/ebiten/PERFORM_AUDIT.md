# TASK: [Ebitengine Edition] Perform a comprehensive goal-focused functional audit of Ebitengine game code for the Go codebase using `go-stats-generator` metrics as the primary evidence source.

## Execution Mode
**Report generation only** — do NOT modify any source code.

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
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the audit report
2. **`GAPS.md`** — gaps between stated goals and implementation

If either file already exists, delete it and create a fresh one.

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

### Phase 0: Understand the Codebase's Goals
Before auditing, learn what the code is supposed to achieve:
1. Read the project README to understand its purpose, users, and claimed capabilities. Extract every verifiable claim — these are your audit targets.
2. Examine `go.mod` for module path, Go version, and dependency profile.
3. List packages (`go list ./...`) and identify the architecture: which packages serve which stated goals.
4. Discover the project's error handling conventions, test strategy, and existing quality gates.
5. Note which packages are on critical paths for the project's core goals (handle user input, implement key features, manage state, perform I/O).
6. Look for design documents, ADRs, or spec files that clarify intent beyond the README.

### Phase 1: Online Research
Use web search to build context that isn't available in the repository:
1. Search for the project on GitHub — read open issues, recent PRs, and community discussions to understand known pain points.
2. Research key dependencies from `go.mod` for known vulnerabilities, deprecations, or upcoming breaking changes.
3. Look up best practices in the project's domain to calibrate audit expectations against its stated goals.

Keep research brief (≤10 minutes). Record only findings that are directly relevant to the project's stated goals.

### Phase 2: Baseline
```bash
mkdir -p tmp
go-stats-generator analyze . --skip-tests --format json --sections functions,documentation,patterns,duplication,interfaces,structs,packages > tmp/audit-metrics.json
go-stats-generator analyze . --skip-tests
```
Delete `tmp/audit-metrics.json` when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Goal-Focused Audit
1. For each stated goal or feature claim, trace through the codebase to verify implementation:
   - **Is it implemented?** Find the entry point and trace execution to output.
   - **Does it work correctly?** Check boundary conditions and error paths.
   - **Does it match the documentation?** Compare promised behavior to actual behavior.
   - **Are there bugs?** Look for logic errors, nil dereferences, resource leaks, race conditions, and incorrect error handling on critical paths.

2. Use `go-stats-generator` metrics to identify risk areas that threaten goal achievement:
   - **HIGH RISK**: Functions on critical goal paths with length >50 lines OR cyclomatic >15 — most likely to contain bugs that prevent goals from being met.
   - **MEDIUM RISK**: Functions on critical goal paths with length >30 lines OR cyclomatic >10.
   - **LOW RISK**: All metrics within thresholds.

3. For each HIGH RISK function on a goal-critical path, perform detailed review:
   - Error handling completeness (does it match the project's conventions?)
   - Nil pointer risks and bounds checking
   - Concurrency safety (cross-reference `.patterns.concurrency_patterns`)
   - Whether the function's complexity is justified by its role (parsers and state machines may warrant higher thresholds)

4. Use dependency-level analysis for systematic coverage:
   - Map import dependencies across `.go` files.
   - Categorize by level: Level 0 (no internal imports) → Level N.
   - Verify correctness in ascending level order.

5. Cross-reference with `.duplication.clones[]` and `.documentation` for additional findings.
6. Run `go test -race ./...` and `go vet ./...` for baseline health. When `go vet` or linters report warnings, read the comments surrounding the flagged code. If a comment explicitly acknowledges the warning (e.g., `//nolint:`, an explanatory comment justifying the pattern, or a TODO tracking a known issue), treat it as an acknowledged false positive — do not report it as a new finding.

### Phase 4: Report
Generate **`AUDIT.md`**:
```markdown
# AUDIT — [date]

## Project Goals
[What the project claims to do, who it serves, and what it promises]

## Goal-Achievement Summary
| Goal | Status | Evidence |
|------|--------|----------|
| [Stated goal] | ✅ Achieved / ⚠️ Partial / ❌ Missing | [file:line or metric reference] |

## Risk Summary
[HIGH: N functions on critical paths, MEDIUM: N functions, critical findings: N]

## Findings
### CRITICAL
- [ ] [Finding] — [file:line] — [evidence] — [how it prevents a stated goal] — [remediation]
### HIGH / MEDIUM / LOW
- [ ] ...
```

Generate **`GAPS.md`**:
```markdown
# Implementation Gaps — [date]

## [Gap Title]
- **Stated Goal**: [what the README/docs promise]
- **Current State**: [what actually exists]
- **Impact**: [how this gap affects users or the project's mission]
- **Closing the Gap**: [what needs to happen]
```

## Risk Thresholds (tunable defaults — calibrate to project)
| Risk Level | Criteria |
|------------|----------|
| HIGH | length >50 OR cyclomatic >15 OR params >7 |
| MEDIUM | length >30 OR cyclomatic >10 OR params >5 |
| LOW | within all thresholds |

## Constraints
- Output ONLY the two report files — no code changes.
- Use `go-stats-generator` metrics as primary evidence for all findings.
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes.
- Evaluate the code against its **own stated goals**, not arbitrary external standards.
- Distinguish findings that block goal achievement from cosmetic issues.

## Tiebreaker
Prioritize: findings that block stated goals → HIGH RISK on critical paths → MEDIUM RISK → LOW RISK. Within a level, highest complexity first.
