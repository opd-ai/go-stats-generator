# TASK: [Ebitengine Edition] Perform a focused code-organization and architecture audit of Ebitengine game code for a Go project, evaluating library-forward design, entrypoint thinness, interface-driven boundaries, and separation of concerns while rigorously preventing false positives.

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
1. **`AUDIT.md`** — the code organization audit report
2. **`GAPS.md`** — gaps between intended architecture and current organization

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

### Phase 0: Map the Project's Organization Model
1. Read the project README to understand intended architecture, target users, and any design claims.
2. Examine `go.mod` for module path, Go version, and dependency profile.
3. List packages (`go list ./...`) and classify:
   - **Entrypoints**: `main` packages under project root or `cmd/`
   - **Library packages**: reusable packages intended to hold feature logic
   - **Internal packages**: implementation details under `internal/`
4. Build a package responsibility map:
   - Which package owns orchestration?
   - Which package owns business/domain logic?
   - Which package owns integration concerns (I/O, DB, network, filesystem)?
5. Catalog exported structs, exported interfaces, and public function signatures across packages.
6. Identify current directory and file-level conventions (naming, grouping by feature/layer, data structure placement).

### Phase 1: Online Research
Use web search to build context:
1. Search for the project on GitHub and review issues/PRs discussing architecture, organization, refactoring, or extensibility pain.
2. Check whether contributors report difficulties extending functionality due to package boundaries or concrete coupling.
3. Review Go project organization guidance to calibrate expectations for CLI + library split.

Keep research brief (≤10 minutes). Record only findings relevant to code organization and extensibility.

### Phase 2: Baseline
```bash
set -o pipefail
mkdir -p tmp
go-stats-generator analyze . --skip-tests --format json --sections packages,functions,interfaces,structs,patterns,duplication > tmp/organize-audit-metrics.json
go-stats-generator analyze . --skip-tests
go build ./... 2>&1 | tee tmp/organize-build-results.txt
go test -race ./... 2>&1 | tee tmp/organize-test-results.txt
```
Delete temporary files when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Code Organization Audit

#### 3a. Library-Forward Architecture
- [ ] Core feature logic is implemented in library packages, not in `main`.
- [ ] `main` functions are thin: parse flags/config, construct dependencies, call library APIs, and format output.
- [ ] `main` does not contain business rules, algorithmic logic, or direct feature implementation.
- [ ] CLI/framework code is separated from reusable domain logic.
- [ ] Library code can be reused without invoking CLI-specific components.

#### 3b. Entrypoint Quality and Orchestration Boundaries
- [ ] Entrypoints orchestrate, they do not implement.
- [ ] Dependency construction is localized and does not leak throughout the codebase.
- [ ] Error handling/reporting in `main` is consistent and delegated where appropriate.
- [ ] Side-effect-heavy setup code is isolated from business workflows.
- [ ] Multiple commands/subcommands share library services instead of duplicating logic.

#### 3c. Struct/Interface Architecture
- [ ] Exported structs that represent extension points have corresponding exported interfaces where abstraction is intended.
- [ ] Public functions accept interfaces for dependencies instead of concrete types.
- [ ] Interface contracts are cohesive and minimal (avoid large, multipurpose interfaces).
- [ ] Concrete implementations remain behind package boundaries where possible.
- [ ] Struct and interface placement follows ownership boundaries (consumer-facing contracts near consumers, implementations near providers).

#### 3d. Separation of Concerns (Directories, Files, Data Structures)
- [ ] Directory structure expresses clear boundaries (domain, infrastructure, transport/CLI, utilities).
- [ ] Files group related concerns and avoid unrelated mixed responsibilities.
- [ ] Data structures are colocated with the package that owns their behavior.
- [ ] Cross-layer imports do not violate intended architectural direction.
- [ ] Circular dependencies are absent; package dependency flow is understandable.
- [ ] Naming and placement conventions are consistent across similar components.

#### 3e. Extensibility and Changeability
- [ ] Adding a new feature can be done by extending library packages, not by bloating entrypoints.
- [ ] New implementations can be introduced via interfaces without widespread edits.
- [ ] Shared abstractions reduce duplication across commands/packages.
- [ ] Package boundaries support independent testing and replacement of components.

#### 3f. False-Positive Prevention (MANDATORY)
Before recording ANY finding, apply these checks:

1. **Confirm intended scope**: Some projects are app-first (not library-first). If README/docs explicitly choose this model, treat as context, not immediate failure.
2. **Verify real coupling pain**: Concrete parameters are findings only when they meaningfully block testing, substitution, or extension.
3. **Avoid abstraction cargo-culting**: Not every struct requires an interface. Record findings only for extension seams, dependency boundaries, or public API contracts where abstraction adds value.
4. **Respect package context**: Internal concrete usage within a package can be appropriate; prioritize package-boundary and public API issues.
5. **Check for intentional deviations**: If comments/docs justify a boundary choice, classify proportionally and avoid over-reporting.

**Rule**: Evaluate organization against the project's stated architecture and extensibility goals, not arbitrary layering ideals.

### Phase 4: Report
Generate **`AUDIT.md`**:
```markdown
# ORGANIZATION AUDIT — [date]

## Architecture Summary
[Entrypoints, library packages, internal boundaries, dependency flow]

## Organization Scorecard
| Category | Rating | Evidence |
|----------|--------|----------|
| Library-Forward Design | ✅/⚠️/❌ | [summary] |
| Entrypoint Thinness | ✅/⚠️/❌ | [summary] |
| Struct/Interface Boundaries | ✅/⚠️/❌ | [summary] |
| Separation of Concerns | ✅/⚠️/❌ | [summary] |
| Extensibility | ✅/⚠️/❌ | [summary] |

## Findings
### CRITICAL
- [ ] [Finding] — [file:line] — [organization issue] — [impact on extensibility/maintainability] — **Remediation:** [specific structural fix]
### HIGH / MEDIUM / LOW
- [ ] ...

## False Positives Considered and Rejected
| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| [description] | [why it is not a true organization issue in this context] |
```

Generate **`GAPS.md`**:
```markdown
# Organization Gaps — [date]

## [Gap Title]
- **Desired Organization**: [what structure/boundary should exist]
- **Current State**: [what exists today]
- **Impact**: [how this harms extensibility/testability/maintainability]
- **Closing the Gap**: [specific package/file/interface restructuring needed]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | Core business logic implemented in `main`/entrypoints, or architecture that fundamentally blocks extension without invasive rewrites |
| HIGH | Public/package-boundary APIs tightly coupled to concrete types where interfaces are required for substitution and testing |
| MEDIUM | Mixed responsibilities across directories/files/data structures causing recurring maintenance friction |
| LOW | Inconsistent naming/placement or minor boundary drift without immediate architectural risk |

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Specific restructuring**: State exactly which package/file/interface boundary should change. Do not recommend "improve organization."
2. **Library-first orientation**: Prefer moving feature logic to reusable library packages and keeping entrypoints orchestration-only.
3. **Verifiable**: Include validation steps (e.g., `go build ./...`, `go test -race ./...`, `go-stats-generator analyze . --sections packages,interfaces,structs`).
4. **Incremental and safe**: Recommend shippable refactoring steps that preserve behavior.

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Use `go-stats-generator` package/interface/struct/function metrics as primary evidence.
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Evaluate organization against the project's **own architecture goals and conventions** before applying generic standards.
- Apply the Phase 3f false-positive prevention checks to every candidate finding before including it.

## Tiebreaker
Prioritize: business logic in entrypoints → concrete coupling on public boundaries → broken separation of concerns → poor extensibility seams → naming/placement inconsistencies. Within a level, prioritize by impact on core project workflows.
