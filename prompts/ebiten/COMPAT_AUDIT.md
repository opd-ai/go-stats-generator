# TASK: [Ebitengine Edition] Perform a focused cross-platform compatibility audit of an Ebitengine Go game, verifying compatibility with Windows, macOS, Linux, FreeBSD, OpenBSD, NetBSD, Android, and iOS while enforcing pure-Go (no CGO), Go standard library idioms, and platform-agnostic code practices.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Ebitengine-Specific Context

This prompt variant is optimized for Go codebases using the Ebitengine (github.com/hajimehoshi/ebiten/v2) game framework. Cross-platform compatibility is a first-class Ebitengine feature, but requires careful attention to:

### Ebitengine Platform Support Matrix
- **Desktop**: Windows, macOS, Linux — full feature support via `ebiten.RunGame`
- **BSD**: FreeBSD — supported via the standard Linux-like build path; OpenBSD/NetBSD — community supported, may require `golang.org/x/sys` version alignment
- **Mobile**: Android (arm64, arm) and iOS (arm64) — via `golang.org/x/mobile` integration; `ebiten.RunGame` replaced by `ebitenmobile.SetGame` or `ebitenmobile.Run`
- **Web (WASM)**: `GOOS=js GOARCH=wasm` — not listed in the task scope but note Ebitengine supports it

### Ebitengine CGO Policy
Ebitengine itself is **pure Go** on most platforms when `CGO_ENABLED=0`. However:
- On Linux, Ebitengine may require X11 or Wayland headers (CGO) for the native windowing backend — use the `ebitenmobile` build path or headless mode to avoid this
- On Android/iOS, `golang.org/x/mobile` uses CGO internally for the runtime bridge — this is acceptable only in the mobile entry-point package (`main` or `app` package), not in game logic packages
- All game logic, asset management, and non-platform-entry-point packages must be pure Go with no CGO

### Ebitengine Mobile Entry Points
- **Android**: `android.go` must call `ebitenmobile.SetGame()` in `func init()` with `//go:build android` constraint
- **iOS**: `ios.go` must call `ebitenmobile.SetGame()` in `func init()` with `//go:build ios` constraint
- **Desktop**: `main.go` calls `ebiten.RunGame()` with no platform constraint (or `//go:build !android,!ios`)

### Platform-Specific Ebitengine Concerns
- **Audio**: `audio.NewContext` sample rate should match platform expectations; iOS may restrict background audio
- **Input**: Touch input must be implemented for Android/iOS; keyboard-only games are unplayable on mobile
- **Screen**: `ebiten.SetWindowSize` and `ebiten.SetFullscreen` are no-ops on mobile — use `Layout()` for logical size
- **Assets**: Must be bundled with `go:embed` — mobile apps cannot read from arbitrary file system paths
- **Fonts**: Must be embedded — system fonts are not reliably accessible across platforms
- **Permissions**: Android requires manifest entries for storage, network, etc.; iOS requires Info.plist entries

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the cross-platform compatibility audit report
2. **`GAPS.md`** — gaps in cross-platform support relative to the project's stated goals

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Target Platforms
Every finding must be tagged with the affected platform(s):

| Tag | Platform |
|-----|----------|
| `[WIN]` | Windows (amd64, arm64) |
| `[MAC]` | macOS (amd64, arm64) |
| `[LIN]` | Linux (amd64, arm64, 386) |
| `[BSD]` | FreeBSD, OpenBSD, NetBSD |
| `[AND]` | Android (arm64, arm) |
| `[IOS]` | iOS (arm64) |
| `[MOB]` | Android + iOS (mobile-specific) |
| `[ALL]` | All platforms |

## Workflow

### Phase 0: Understand the Project's Cross-Platform Posture
1. Read the project README — extract any stated target platforms, mobile support claims, or cross-platform goals.
2. Examine `go.mod` for module path, Go version, and dependencies:
   - Confirm `github.com/hajimehoshi/ebiten/v2` is present
   - Check for `golang.org/x/mobile` (required for Android/iOS entry points)
   - Look for any dependency that pulls in CGO outside the allowed mobile entry-point exception
3. List packages (`go list ./...`) and classify:
   - Entry point packages (`main`, mobile `app` packages with `ebitenmobile.SetGame`)
   - Game logic packages (must be pure Go, CGO-free)
   - Asset packages (`go:embed`, must have no file-system reads at runtime)
4. Build a **platform surface inventory** by scanning for:
   - `import "C"` outside the mobile entry-point package
   - `//go:build` constraints — check for missing platform coverage
   - `ebiten.SetWindowSize`, `ebiten.SetWindowTitle`, `ebiten.SetFullscreen` — no-ops on mobile
   - `ebiten.RunGame` — must be guarded with `//go:build !android,!ios`
   - `ebitenmobile.SetGame` / `ebitenmobile.Run` — must be guarded with `//go:build android` or `//go:build ios`
   - `os/exec` — forbidden on iOS, severely restricted on Android
   - `os.Open`, `os.ReadFile` — forbidden on mobile for arbitrary paths; only sandbox paths allowed
   - `filepath.Join` vs hardcoded `/` separators
   - `go:embed` for all asset loading
   - Touch input via `ebiten.TouchIDs` or `inpututil.AppendJustPressedTouchIDs`
   - Keyboard-only input with no touch fallback
5. Map any `_android.go`, `_ios.go`, `_windows.go`, `_linux.go`, `_darwin.go`, `_freebsd.go` files.
6. Check that `Layout()` implementation returns consistent logical dimensions independent of actual screen size.

### Phase 1: Online Research
Use web search to build context:
1. Search for the project on GitHub — read issues mentioning "android", "ios", "mobile", "windows", "linux", "mac", "build", "cross-platform", or "CGO".
2. Check the Ebitengine compatibility page (https://ebitengine.org/en/documents/platforms.html) for platform-specific limitations.
3. Research `golang.org/x/mobile` for Android/iOS build requirements and CGO policy.
4. Check if any Ebitengine version-specific platform bugs apply to the version in `go.mod`.

Keep research brief (≤10 minutes). Record only findings directly relevant to cross-platform compatibility.

### Phase 2: Baseline
```bash
set -o pipefail
mkdir -p tmp

go-stats-generator analyze . --skip-tests --format json --sections functions,packages,patterns > tmp/compat-audit-metrics.json
go-stats-generator analyze . --skip-tests

# Verify CGO-free build of game logic (exclude mobile entry-point packages)
CGO_ENABLED=0 go list ./... | grep -Ev '/(android|ios|mobile)(/|$)' | xargs -r go build 2>&1 | tee tmp/cgo-free-build.txt

# Cross-compile spot checks
GOOS=windows GOARCH=amd64  CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-windows.txt
GOOS=darwin  GOARCH=arm64  CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-darwin.txt
GOOS=linux   GOARCH=amd64  CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-linux.txt
GOOS=freebsd GOARCH=amd64  CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-freebsd.txt
GOOS=openbsd GOARCH=amd64  CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-openbsd.txt
GOOS=netbsd  GOARCH=amd64  CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-netbsd.txt
GOOS=android GOARCH=arm64  CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-android.txt
GOOS=ios     GOARCH=arm64  CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-ios.txt

go vet ./... 2>&1 | tee tmp/compat-vet.txt
```
Delete temporary files when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Cross-Platform Compatibility Audit

#### 3a. CGO and Native Dependencies
For every package in the module, verify:

- [ ] No `import "C"` appears outside the designated mobile entry-point package (if the project targets Android/iOS).
- [ ] No `#cgo` directives appear in any Go source file outside the mobile entry-point package.
- [ ] `CGO_ENABLED=0 go build ./...` succeeds for all non-mobile-entry-point packages.
- [ ] No dependency in `go.mod` transitively requires CGO in game logic packages.
- [ ] The mobile entry-point package (if present) is clearly isolated and documented as the only CGO-permitted zone.

#### 3b. Ebitengine Entry Point Structure
For the game's platform entry points, verify:

- [ ] `ebiten.RunGame()` in `main.go` (or equivalent) is guarded with `//go:build !android,!ios` or equivalent to prevent compilation on mobile.
- [ ] `ebitenmobile.SetGame()` in `android.go` is guarded with `//go:build android`.
- [ ] `ebitenmobile.SetGame()` in `ios.go` is guarded with `//go:build ios`.
- [ ] No single file attempts to call both `ebiten.RunGame` and `ebitenmobile.SetGame` without build constraints.
- [ ] Mobile entry points are in their own `app` or dedicated package, not mixed with game logic.
- [ ] `ebiten.SetWindowSize`, `ebiten.SetWindowTitle`, `ebiten.SetFullscreen`, and `ebiten.SetWindowResizingMode` calls are guarded with `//go:build !android,!ios` (these are no-ops on mobile but their presence in shared code adds confusion).

#### 3c. Asset Loading and Embedding
For every asset (images, audio, fonts, data files), verify:

- [ ] All assets are loaded via `go:embed` directives — not `os.Open`, `os.ReadFile`, or `ioutil.ReadFile` with relative or absolute paths (which fail on mobile).
- [ ] `embed.FS` is passed to `ebitenutil.NewImageFromReader` or equivalent, not from a raw `os.File`.
- [ ] Font files are embedded with `//go:embed` — system fonts (`/usr/share/fonts`, `C:\Windows\Fonts`, etc.) are not used.
- [ ] Audio files are embedded with `//go:embed` — no runtime file-system reads for audio data.
- [ ] The `go:embed` glob patterns do not accidentally embed sensitive files (`.env`, secrets, large test fixtures).
- [ ] Embedded assets use forward-slash paths in `//go:embed` directives — the `embed` package always uses forward slashes regardless of OS.

#### 3d. Input Handling for All Platforms
For every input handling path, verify:

- [ ] Touch input is implemented alongside mouse/keyboard input for mobile compatibility:
  - `ebiten.TouchIDs()` or `inpututil.AppendJustPressedTouchIDs` checked in `Update()`
  - Touch position translated to game coordinates via `ebiten.TouchPosition`
- [ ] Keyboard-only game actions have touch/tap equivalents for mobile — there is no physical keyboard on most Android/iOS devices.
- [ ] Gamepad input checks `ebiten.AppendGamepadIDs` before reading gamepad state — not all platforms have gamepads, and Android/iOS gamepad support is optional.
- [ ] Virtual on-screen controls or touch UI are present if the game requires directional or action input on mobile.
- [ ] `inpututil.IsKeyJustPressed` and `inpututil.IsKeyJustReleased` are not used as the sole input mechanism for a mobile-targeted game.
- [ ] Back button / escape key handling on Android uses `ebiten.IsKeyPressed(ebiten.KeyEscape)` or the platform-appropriate mechanism.

#### 3e. Screen and Layout
For every screen size and layout concern, verify:

- [ ] `Layout(outsideWidth, outsideHeight int)` returns fixed logical dimensions — the game logic uses logical, not physical, pixels.
- [ ] The game does not hardcode pixel positions assuming a specific screen resolution — all layout uses the logical dimensions returned by `Layout()`.
- [ ] DPI scaling is handled: `ebiten.DeviceScaleFactor()` is used if physical pixel precision is needed (e.g., for pixel-art sharpness).
- [ ] Safe area insets for iOS notch and Android cutout are considered for UI element placement if the game uses full-screen mode.
- [ ] Text and UI elements are legible at mobile screen densities (high DPI) — font sizes in logical pixels that translate to readable physical sizes.
- [ ] `ebiten.SetWindowSize` is called only on desktop — this call is a no-op on mobile but is a code smell in shared code.

#### 3f. Audio Cross-Platform
For every audio operation, verify:

- [ ] `audio.NewContext` sample rate (typically 44100 or 48000) is compatible with all target platforms — iOS prefers 44100, Android supports both.
- [ ] Audio playback does not depend on system audio APIs not available through Ebitengine's abstraction.
- [ ] Audio files are in a format Ebitengine can decode cross-platform (MP3, OGG, WAV, FLAC via the appropriate `ebiten/audio/*` sub-package).
- [ ] Background audio behavior on iOS is considered — iOS pauses audio when the app is backgrounded unless the appropriate capability is configured.
- [ ] No direct `syscall` or platform-specific audio API calls bypass Ebitengine's audio abstraction.

#### 3g. File System and Path Handling
For every file path operation outside of embedded assets, verify:

- [ ] `filepath.Join` is used for all OS file paths — never hardcoded `/` or `\`.
- [ ] `os.TempDir()` is used for temporary files — not `/tmp` (which does not exist on Windows or mobile sandboxes).
- [ ] `os.UserHomeDir()`, `os.UserCacheDir()`, `os.UserConfigDir()` are used for user-data storage on desktop — not `~/` expansion.
- [ ] Save game data on mobile uses platform-appropriate directories (obtained via `os.UserCacheDir()` or injected by the mobile framework), not hardcoded paths.
- [ ] File I/O on Android/iOS is restricted to the app sandbox — no reads from `/sdcard`, `/etc`, or other system paths.

#### 3h. Build Constraints and Platform-Specific Code
For every `//go:build` constraint and platform-specific file, verify:

- [ ] All build tags use the modern `//go:build` syntax — not the deprecated `// +build` syntax alone.
- [ ] Every platform-specific file name suffix (`_android.go`, `_ios.go`, `_windows.go`) has a matching `//go:build` constraint.
- [ ] No platform is silently excluded — every supported platform has either a shared implementation or a platform-specific stub.
- [ ] `runtime.GOOS` and `runtime.GOARCH` switches are not used for logic that could be expressed with build constraints (prefer compile-time over runtime platform detection for significant divergence).

#### 3i. BSD-Specific Concerns (FreeBSD, OpenBSD, NetBSD)
For any code that may run on BSD variants:

- [ ] `/proc` filesystem access is guarded with Linux-only build constraints — `/proc` does not exist on BSD.
- [ ] `inotify`-based file watching (Linux-specific) is not used without a BSD fallback via `kqueue`.
- [ ] Ebitengine's BSD build path is tested — `GOOS=freebsd CGO_ENABLED=0 go build ./...` must succeed.
- [ ] No tool names or shell commands assume Linux-specific behavior in `os/exec` calls with a BSD fallback.

#### 3j. Mobile-Specific Constraints
For any code path that reaches Android or iOS:

- [ ] No `os/exec` calls anywhere in the package graph reachable from mobile entry points.
- [ ] No `net.Listen` for serving — mobile apps cannot run as network servers without special entitlements.
- [ ] `context.Context` with timeout wraps all network operations — mobile OSes may suspend or kill background network operations.
- [ ] Memory footprint is considered — avoid loading all assets into memory at once on mobile; use lazy loading or streaming where possible.
- [ ] `time/tzdata` is blank-imported in the mobile entry-point package — mobile platforms may not have a system timezone database.
- [ ] The Android manifest (if present) declares required permissions (network, storage) and the minimum SDK version supports the targeted Ebitengine features.
- [ ] The iOS Info.plist (if present) declares required permissions and the deployment target supports the targeted Ebitengine features.

#### 3k. False-Positive Prevention (MANDATORY)
Before recording ANY finding, apply these checks:

1. **Verify the build constraint scope**: Code in `main_desktop.go` with `//go:build !android,!ios` that calls `ebiten.RunGame` is correct — it is not a mobile portability issue.
2. **Check Ebitengine's own abstraction**: Ebitengine handles many platform differences internally. Verify whether the flagged API goes through Ebitengine's abstraction layer before reporting it as a portability issue.
3. **Assess actual target platforms**: If the project README states "desktop only", mobile findings are informational. Evaluate against stated goals.
4. **CGO mobile exception**: `import "C"` in the mobile entry-point package (`android.go`, `ios.go`) with the correct build constraint is expected and acceptable — do NOT report it as a CGO violation.
5. **Verify asset path handling**: `go:embed` paths always use forward slashes — a forward slash in a `//go:embed` directive is correct, not a Windows path separator bug.
6. **Check for Ebitengine version support**: Some APIs or platforms were added in specific Ebitengine versions — verify the `go.mod` version before flagging a missing API as a portability gap.

**Rule**: If a platform-specific pattern is already guarded by a correct build constraint or routes through Ebitengine's cross-platform abstraction, it is NOT a finding. Only report code that is reachable on a platform it is incompatible with.

### Phase 4: Report
Generate **`AUDIT.md`**:
```markdown
# CROSS-PLATFORM COMPATIBILITY AUDIT (Ebitengine Edition) — [date]

## Ebitengine Version
[Version from go.mod, platform support matrix for that version]

## Target Platform Matrix
| Platform | Build Status | Ebitengine Support | Known Issues |
|----------|--------------|--------------------|--------------|
| Windows (amd64) | ✅/⚠️/❌ | ✅ | [summary] |
| macOS (arm64)   | ✅/⚠️/❌ | ✅ | [summary] |
| Linux (amd64)   | ✅/⚠️/❌ | ✅ | [summary] |
| FreeBSD (amd64) | ✅/⚠️/❌ | ✅ | [summary] |
| OpenBSD (amd64) | ✅/⚠️/❌ | ⚠️ community | [summary] |
| NetBSD (amd64)  | ✅/⚠️/❌ | ⚠️ community | [summary] |
| Android (arm64) | ✅/⚠️/❌ | ✅ via gomobile | [summary] |
| iOS (arm64)     | ✅/⚠️/❌ | ✅ via gomobile | [summary] |

## CGO Status
[Result of CGO_ENABLED=0 go build ./... for game logic packages — PASS or FAIL with details]
[Mobile entry-point CGO usage — documented exception or violation]

## Platform Surface Inventory
| Package | CGO | Entry Point | Assets | Touch Input | Desktop-Only APIs | Build Tags |
|---------|-----|-------------|--------|-------------|-------------------|------------|
| [pkg]   | ✅/❌ | desktop/mobile/both | embed/fs | ✅/❌ | N | [tags] |

## Findings
### CRITICAL
- [ ] [Finding] `[PLATFORM]` — [file:line] — [portability issue] — [impact: build failure, panic, wrong behavior] — **Remediation:** [specific fix]
### HIGH / MEDIUM / LOW
- [ ] ...

## False Positives Considered and Rejected
| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| [description] | [why it is acceptable — build constraint, Ebitengine abstraction, or stated scope] |
```

Generate **`GAPS.md`**:
```markdown
# Cross-Platform Compatibility Gaps (Ebitengine Edition) — [date]

## [Gap Title]
- **Stated Goal**: [what the project claims about platform/mobile support]
- **Current State**: [what platforms are actually supported and how]
- **Affected Platforms**: [WIN / MAC / LIN / BSD / AND / IOS]
- **Impact**: [build failure, runtime panic, unplayable on platform, incorrect behavior]
- **Closing the Gap**: [specific changes — build constraints, touch input, go:embed, ebitenmobile setup]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | CGO in game logic packages (outside mobile entry point); `ebiten.RunGame` without mobile build guard causing mobile build failure; file-system reads that fail on mobile sandbox; missing touch input making the game unplayable on mobile |
| HIGH | Missing `//go:build` guard for desktop-only APIs; asset loaded via `os.Open` instead of `go:embed`; hardcoded screen size breaking mobile layout; missing timezone data for mobile |
| MEDIUM | Audio sample rate mismatch; missing back-button handling on Android; UI elements too small for touch targets; platform-specific behavior difference that causes visual artifacts |
| LOW | Informational portability note; desktop-only API call guarded correctly but could be refactored; suboptimal but functional cross-platform pattern |

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Concrete fix**: State the exact file, function, build constraint, or `go:embed` directive to add or change.
2. **Preserve Ebitengine idioms**: Fixes must route through Ebitengine's cross-platform abstractions — do not introduce new platform-native calls.
3. **Verifiable**: Include the cross-compile command that must pass after the fix (e.g., `GOOS=android CGO_ENABLED=0 go build ./...`).
4. **Minimal scope**: Fix the portability issue without restructuring game logic.

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Use `go-stats-generator` metrics as supporting evidence for complexity and package structure.
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Every finding must be tagged with the affected platform(s) using the tags defined above.
- Every finding must include the concrete failure mode — no speculative findings.
- Evaluate the code against its **own stated platform goals**.
- Apply the Phase 3k false-positive prevention checks to every candidate finding before including it.
- CGO in game logic packages is non-negotiable CRITICAL; CGO in mobile entry-point packages with correct build constraints is acceptable.

## Tiebreaker
Prioritize: CGO violations in game logic → mobile build failures → unplayable on mobile (no touch input, broken layout) → asset loading failures → runtime panics on specific platforms → incorrect behavior → informational portability notes. Within a level, prioritize by number of affected platforms.
