# TASK: [Ebitengine Edition] Investigate a broken Go program to discover and fix the bug(s) blocking its basic utility.

## Execution Mode
**Autonomous action** — diagnose the failure, fix the blocking bug(s), validate with tests and diff. Do not fix anything that is not blocking basic utility.

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

## Core Directive
The program is already known to be broken. Do not start by reading all the code or running a comprehensive checklist. **Start by observing the failure.** Every subsequent step must be driven by what you observe, working backward from symptom to root cause.

## Scope Rule
Fix only bugs that block basic utility. A bug is blocking if it satisfies at least one of:
1. The program **crashes or panics** on a documented invocation.
2. The program **fails to start** on a valid configuration.
3. A documented CLI command **produces no output or garbage output** for any input.
4. A **core data path produces an incorrect result** that contradicts the tool's stated purpose.
5. A documented flag or option is **silently ignored**, making a feature unreachable.

If a candidate finding does not satisfy at least one of the above, skip it. Do not fix it.

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

### Phase 0: Observe the Failure First
Before reading any code, run the program and record what actually happens:

```bash
go build ./... 2>&1
```

If the build fails, the build error IS the blocking bug. Go to Phase 2 immediately with the build output as your symptom.

If the build succeeds, run the primary documented invocation from the README:
```bash
# Replace with the actual documented happy-path command from the README
<primary-invocation> 2>&1
```

Record **exactly** what the user sees:
- Crash/panic message and stack trace
- Wrong or empty output (capture it)
- Non-zero exit code with no message
- Program that hangs indefinitely
- Misleading error message

This observed failure is the **anchor**. Every subsequent step must trace back to it. Do not report bugs that are not connected to this failure.

### Phase 1: Understand What Should Happen
Now that you know how it fails, establish what it should do:

1. Read the project README. Extract:
   - The primary value proposition and documented happy-path invocation(s).
   - The expected output for the primary invocation on this codebase.
   - Every documented CLI command and flag.
2. Note the Go version in `go.mod` (loop variable capture is eliminated in Go 1.22+).
3. Identify the **critical path**: the call chain from the entry point to the output that the primary invocation depends on. This is the only code you need to investigate.

### Phase 2: Triage — Classify the Failure Mode
Classify the observed failure into one of these modes. The mode determines where to look:

| Failure Mode | Where to Look First |
|---|---|
| **Build failure** | Compiler error location; missing types, methods, or imports |
| **Startup crash** | `main()`, `init()` functions, config parsing, storage initialization |
| **Runtime panic** | Stack trace → panicking goroutine → the nil/bounds/type-assertion site |
| **Silent wrong output** | The function responsible for the metric or section that is wrong |
| **Hang / deadlock** | Goroutine dump (`SIGQUIT` or `kill -ABRT`); channel send/receive pairs |
| **Empty output / no-op** | CLI command registration; `Run`/`RunE` wiring; flag-to-logic connection |

For **runtime panics**, capture the full stack trace:
```bash
GOTRACEBACK=all <primary-invocation> 2>&1 | head -100
```

For **hangs**, get a goroutine dump:
```bash
# Send SIGQUIT to the running process to print all goroutine stacks
kill -QUIT <pid>
```

### Phase 3: Root Cause Investigation
Starting from the failure mode identified in Phase 2, trace backward through the call chain to the root cause. Do NOT read code that is not on the path from the entry point to the failure.

Use `go-stats-generator` metrics to identify high-risk zones on the critical path:
```bash
mkdir -p tmp
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages > tmp/breaking-audit-metrics.json
```
Functions with cyclomatic complexity >20 or nesting >5 on the critical path are high-risk and should be read carefully. Delete `tmp/breaking-audit-metrics.json` when done.

#### Patterns to look for once you have a hypothesis
Look for these only within the code you are already tracing — do not use these as an independent checklist:

**Crash patterns**
- Nil dereference: pointer used without nil check on a path that runs for valid inputs
- Nil map write: `m[k] = v` where `m` was never initialized with `make`
- Slice out-of-bounds: index derived from input without a bounds check
- Type assertion without ok-check: `x.(T)` where `x` may not hold type `T`
- Unrecovered panic in goroutine: goroutine panics and crashes the whole program
- `init()` panics on a valid installation

**Startup / wiring failures**
- Storage backend not initialized before first use (e.g., SQLite schema not created)
- Required config file not documented, causing an error on clean install
- CLI command with `Run`/`RunE` not set: registered in help but does nothing
- Flag parsed but value never read by any code path

**Silent wrong output**
- Integer division truncation: `a / b * 100` instead of `a * 100 / b`
- Accumulator reassigned instead of incremented inside the loop it aggregates
- Result struct field never written: always emits the zero value
- Swallowed parse error: error discarded, zero-value result included in aggregates
- Error from goroutine silently dropped: affected files produce no output

**Concurrency / hang**
- Goroutines blocked on channel send with no consumer (goroutine leak → hang)
- WaitGroup `Done()` called more times than `Add()` (panic)
- Channel send after close (panic)
- Shared map written from multiple goroutines without synchronization (race → panic)

### Phase 4: Fix
Apply the **minimum change** that restores basic utility:

1. Fix only the confirmed root cause. Do not refactor, improve style, or fix unrelated bugs.
2. Match the project's existing error handling convention, variable naming, and code style.
3. Preserve all existing API contracts and behavior outside the broken path.
4. After each fix, run:
   ```bash
   go build ./... && go test -race ./... && <primary-invocation>
   ```
   Confirm the failure from Phase 0 is resolved before declaring done.

If fixing the root cause reveals a second blocking bug on the same critical path, fix that too and re-run. Stop when the primary documented invocation works correctly.

### Phase 5: Validate and Report
```bash
go-stats-generator analyze . --skip-tests --format json --output post-fix.json --sections functions,patterns
go-stats-generator diff baseline.json post-fix.json
```
(Run the baseline before fixing if you have not already.)

Delete temporary files. Write one file in the repository root:

**`AUDIT.md`**:

```markdown
# BREAKING BUG AUDIT — [date]

## Observed Failure
[Exact symptom: the command run, the output/crash seen, the exit code]

## Root Cause
[File:line, function name, and a one-sentence description of the bug]

## Fix Applied
[What was changed: file, function, before/after diff summary]

## Verification
[The command run after the fix and its output confirming the failure is resolved]

## Other Blocking Bugs Found
- [ ] [If additional blocking bugs were found on the same critical path, list them here with file:line and remediation]

## Discarded Candidates
| Candidate | Reason Discarded |
|-----------|-----------------|
| [description] | [which scope rule it failed, or which guard prevented it from being reachable] |
```

## Fix Rules
- Fix only the confirmed root cause. Do not refactor, improve style, or fix unrelated bugs in the same pass.
- Match the project's existing error handling convention, variable naming, and code style.
- Preserve all existing API contracts and behavior outside the broken path.
- The minimum change that restores the documented behavior is the correct change.

## Tiebreaker
If multiple blocking bugs are found, fix in this order: build failure → startup crash → runtime crash on primary invocation → silent no-op command → incorrect core metric. Fix the highest-priority bug first, verify it, then continue.
