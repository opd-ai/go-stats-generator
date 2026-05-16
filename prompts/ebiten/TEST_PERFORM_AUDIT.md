# TASK: [Ebitengine Edition] Perform a comprehensive goal-focused functional audit of Ebitengine game code for **test code** using `go-stats-generator` metrics as the primary evidence source.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## 
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
1. **`AUDIT.md`** — the test audit report
2. **`GAPS.md`** — gaps in test coverage relative to the project's stated goals

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

### Phase 0: Understand the Project's Goals and Test Strategy
1. Read the project README to understand its domain, stated goals, and expected behavior.
2. Discover the test framework in use and the project's assertion patterns.
3. Identify the test organization: how are tests structured, what testing conventions exist?
4. Note whether the project uses `t.Parallel()`, `t.Cleanup()`, test suites, or integration test separation.
5. Map which stated goals have corresponding test coverage and which do not.

### Phase 1: Online Research
Use web search to build context that isn't available in the repository:
1. Search for the project on GitHub — read open issues and discussions to understand known test gaps and flaky test reports.
2. Research the project's test dependencies for known issues, deprecations, or better alternatives.
3. Look up testing best practices in the project's domain (e.g., table-driven tests, test helpers, integration test patterns).

Keep research brief (≤10 minutes). Record only findings that are directly relevant to the project's test strategy and stated goals.

### Phase 2: Baseline
```bash
mkdir -p tmp
go-stats-generator analyze . --only-tests --format json --sections functions,documentation,patterns,duplication > tmp/test-audit-metrics.json
go-stats-generator analyze . --only-tests
```
Delete `tmp/test-audit-metrics.json` when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Goal-Focused Test Audit
1. For each stated project goal, verify that adequate test coverage exists:
   - Are the critical paths for this goal tested?
   - Do tests cover the happy path, error paths, and edge cases?
   - Are integration points between packages tested?

2. Extract test function metrics and classify by risk (tunable defaults):
   - **HIGH RISK**: length >75 lines OR cyclomatic >22 OR nesting >7
   - **MEDIUM RISK**: length >45 lines OR cyclomatic >15 OR nesting >5
   - **LOW RISK**: within test-appropriate thresholds

3. For each HIGH RISK test function, review:
   - Test setup error handling (are setup failures caught with `t.Fatal`?)
   - Race conditions in parallel tests
   - Resource cleanup (`t.Cleanup()`, temp files, goroutines)
   - Flaky test patterns (timing, hardcoded ports, file system)

4. Cross-reference with `.duplication` and `.documentation` for additional findings.

5. **Bug-masking risk**: Identify tests that may hide bugs:
   - Tests with no assertions or only `t.Log` calls.
   - Tests that swallow errors (e.g., `err` assigned but never checked).
   - Tests that assert on outdated or hardcoded values instead of computed expectations.
   - Missing negative tests for error paths on goal-critical functions.

### Phase 4: Report

Generate **`AUDIT.md`**:
```markdown
# TEST AUDIT — [date]

## Project Goals and Test Coverage
| Stated Goal | Test Coverage | Assessment |
|-------------|--------------|------------|
| [Goal] | [which test files/functions cover this] | ✅ Well-tested / ⚠️ Partial / ❌ Untested |

## Test Infrastructure Context
[Test framework, conventions, coverage approach]

## Risk Summary
[HIGH: N, MEDIUM: N, critical findings: N]

## Findings
### CRITICAL
- [ ] [Finding] — [file:line] — [evidence] — [which goal's test coverage is affected] — [remediation]
### HIGH / MEDIUM / LOW
- [ ] ...
```

Generate **`GAPS.md`**:
```markdown
# Test Coverage Gaps — [date]

## [Gap Title]
- **Stated Goal**: [what the project claims to do]
- **Current Test State**: [what tests exist, if any]
- **Missing Coverage**: [what scenarios are untested]
- **Impact**: [risk of this goal regressing without tests]
- **Closing the Gap**: [specific tests to add]
```

## Risk Thresholds (test-appropriate — ~50% relaxed)
| Risk Level | Criteria |
|------------|----------|
| HIGH | length >75 OR cyclomatic >22 OR nesting >7 |
| MEDIUM | length >45 OR cyclomatic >15 OR nesting >5 |
| LOW | within thresholds |

## Constraints
- Output ONLY the two report files — no code changes.
- Use `go-stats-generator --only-tests` metrics as primary evidence.
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes.
- Evaluate test quality in terms of how well tests verify the project's **stated goals**, not arbitrary coverage targets.
- When `go vet` or linters report warnings in test code, read the comments surrounding the flagged code. If a comment explicitly acknowledges the warning (e.g., `//nolint:`, an explanatory comment justifying the pattern, or a TODO tracking a known issue), treat it as an acknowledged false positive — do not report it as a new finding.
- Recommend table-driven tests and `t.Helper()` extraction as primary remediation patterns.

## Tiebreaker
Prioritize: untested stated goals → HIGH RISK → MEDIUM RISK → LOW RISK. Within a level, highest complexity first.
