# Ebitengine UI/UX Debug Audit

## Objective
Autonomously audit all Ebitengine-dependent Go code in this repository for UI/UX defects. Produce a structured diagnostic report covering every identified issue with severity, location, and a concrete fix.

## Execution Mode
**Autonomous report generation** — analyze code and produce findings. Do not modify any files.

## Scope
Audit only code that imports or interacts with `github.com/hajimehoshi/ebiten/v2` or related Ebitengine packages (`ebitenutil`, `inpututil`, `text`, `audio`, `mobile`, `vector`). Ignore unrelated packages, server-side logic, and non-rendering tests.

## Audit Checklist

### 1. Performance
- `ebiten.NewImage()` or `ebiten.NewImageFromImage()` called inside `Draw()` or `Update()` (per-frame allocation)
- Missing dirty flags causing unnecessary full-screen redraws
- Sprite/asset caches without size limits or eviction
- `sync.Pool` Get/Put asymmetry for `DrawImageOptions`, vertex slices, or index slices
- Logging or string formatting on the hot path (`Draw`/`Update`) without level guards
- Inappropriate `SetTPS`, `SetVsyncEnabled`, or `SetScreenClearedEveryFrame` configuration

### 2. Collision & Physics Math
- Strict `<` vs `<=` in distance-based collision checks (off-by-epsilon)
- `Normalize()` on zero-length vectors without guard (NaN/Inf propagation)
- Sequential single-axis AABB resolution causing corner-case tunneling
- Floating-point quantization or rounding that accumulates drift
- Negative penetration depth or zero-normal edge cases left unhandled
- Hardcoded screen boundaries (magic numbers instead of dynamic bounds)
- Slide/reflection vectors with incorrect dot-product sign

### 3. Layout, Overlap & Responsiveness
- Hardcoded pixel positions that break at non-target resolutions
- Elements rendered at overlapping coordinates without z-order management
- `Layout()` returning inconsistent or non-matching logical dimensions
- Touch/click targets below 44×44 logical pixels on mobile builds
- Scroll offset clamping against stale `maxScroll` values
- Window resize events not propagated to child widgets or UI panels

### 4. Text Readability
- Text-on-background color pairs failing WCAG AA contrast (4.5:1)
- Font scale factors producing unreadable sizes at non-default DPI
- Long strings overflowing container bounds without truncation or wrapping
- Production UI relying on `ebitenutil.DebugPrint` instead of proper text rendering

### 5. Input Handling
- Keyboard shortcuts active during text-input mode (missing mode guard)
- Key bindings that shadow each other in the same game state
- Mouse/touch coordinates not adjusted for scroll offset or camera transform
- Drag/hover state not reset on focus loss, window minimize, or touch cancel
- `ebiten.AppendInputChars` called without consuming all buffered characters

### 6. Resource & Memory Management
- `ebiten.Image` created but never reused or allowed to be GC'd
- Goroutines (transport, audio, asset loading) not terminated on scene change or shutdown
- Notification/event channels with silent drops on backpressure

## Output Format
For each finding:

```
### [SEVERITY] Short description
- **File**: `path/to/file.go#L<start>-L<end>`
- **Category**: Performance | Collision | Layout | Text | Input | Resource
- **Problem**: One-sentence description of the defect.
- **Evidence**: The specific code pattern or value causing the issue.
- **Fix**: Concrete code change or approach.
```

Severity levels:
- `CRITICAL` — crash, hang, data loss, or infinite loop
- `HIGH` — visible user-facing bug under normal use
- `MEDIUM` — performance degradation or edge-case incorrect behavior
- `LOW` — code quality, minor UX polish, or defensive hardening

## Success Criteria
- Every file importing Ebitengine is evaluated or explicitly noted as skipped with reason.
- Zero findings reference code outside the Ebitengine interaction boundary.
- All CRITICAL and HIGH findings include a testable fix description.
- Report is ordered by severity descending, then by file path alphabetically.
- If no issues are found in a category, emit `**No issues found.**` for that section.
