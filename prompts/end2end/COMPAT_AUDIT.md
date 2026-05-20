# TASK (END-TO-END): Perform an exhaustive cross-platform compatibility audit of a Go project, verifying compatibility with Windows, macOS, Linux, FreeBSD, OpenBSD, NetBSD, Android, and iOS while enforcing pure-Go (no CGO), Go standard library idioms, and platform-agnostic code practices across every file and package.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## End-to-End Policy
This is the **end-to-end variant** of COMPAT_AUDIT. The following rules are absolute:
- **No finding cap** — report every confirmed portability issue regardless of total count.
- **Complete coverage** — audit every file, every function, and every package. Do not sample or triage by priority before completing a full pass.
- **Iterative until done** — if context is running low, commit progress, document remaining scope explicitly, and continue in a fresh session. Never stop mid-audit.
- **Findings are cumulative** — each pass may reveal new issues; repeat until a complete pass produces zero new confirmed findings above LOW.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the consolidated cross-platform compatibility audit report
2. **`GAPS.md`** — gaps between what the project claims to support and what is actually cross-platform compatible

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
1. Read the project README — extract every stated platform target, cross-platform claim, mobile support statement, and deployment constraint. These are the **acceptance criteria** for the audit.
2. Examine `go.mod` for module path, Go version, and dependencies — identify any dependency that pulls in CGO or platform-specific native libraries.
3. List packages with `go list ./...` and classify each:
   - Entry point packages (desktop `main`, mobile entry points)
   - Core logic packages (must be pure Go, CGO-free)
   - Platform abstraction packages (expected to contain build-constrained files)
   - Asset/data packages (must use `go:embed`)
   - Test packages (platform constraints may be relaxed in tests)
4. Build a **complete platform surface inventory** across every file:
   - `import "C"` — CGO usage
   - `//go:build` constraints — check every constrained file for coverage gaps
   - `runtime.GOOS` / `runtime.GOARCH` switches
   - `os/exec` — forbidden on mobile
   - `syscall` and `golang.org/x/sys` — direct syscall usage
   - All file path operations: `filepath.Join`, `os.Open`, `os.ReadFile`, hardcoded path strings
   - `os.TempDir()`, `os.UserHomeDir()`, `os.UserCacheDir()`, `os.UserConfigDir()`
   - `go:embed` directives — verify coverage of all assets
   - Signal handling — `os/signal`, `syscall.SIGTERM`, POSIX-only signals
   - `time/tzdata` — check for blank import in mobile-targeting packages
   - `encoding/binary` vs unsafe pointer casts for serialization
   - Network I/O with context and timeout handling
   - Integer width assumptions (`int` vs `int32`/`int64`)
   - ANSI escape codes and terminal detection
5. Map all `_windows.go`, `_linux.go`, `_darwin.go`, `_freebsd.go`, `_openbsd.go`, `_netbsd.go`, `_android.go`, `_ios.go` files against the full package list — identify any package that has some platform-specific files but is missing coverage for a supported platform.
6. Identify the build strategy: Makefile, goreleaser, GitHub Actions matrix — does it actually build for every stated target?

### Phase 1: Online Research
1. Search for the project on GitHub — read open issues and PRs mentioning "windows", "darwin", "linux", "freebsd", "android", "ios", "CGO", "build constraint", or "cross-platform".
2. Research every dependency in `go.mod` for CGO requirements, platform restrictions, and known mobile incompatibilities.
3. Look up `GOOS`/`GOARCH` support for every third-party package used.
4. Verify whether the Go version in `go.mod` supports all stated target platforms (e.g., iOS support was added/improved in specific Go versions).

Keep research brief (≤10 minutes). Record only findings with direct bearing on cross-platform compatibility.

### Phase 2: Baseline
```bash
set -o pipefail
mkdir -p tmp

go-stats-generator analyze . --format json \
  --sections functions,packages,patterns,structs \
  > tmp/compat-e2e-baseline.json
go-stats-generator analyze .

# CGO-free build — must pass for all packages
CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/cgo-free-build.txt

# Cross-compile matrix — record pass/fail for each
GOOS=windows GOARCH=amd64  CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-windows-amd64.txt
GOOS=windows GOARCH=arm64  CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-windows-arm64.txt
GOOS=darwin  GOARCH=amd64  CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-darwin-amd64.txt
GOOS=darwin  GOARCH=arm64  CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-darwin-arm64.txt
GOOS=linux   GOARCH=amd64  CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-linux-amd64.txt
GOOS=linux   GOARCH=arm64  CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-linux-arm64.txt
GOOS=linux   GOARCH=386    CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-linux-386.txt
GOOS=freebsd GOARCH=amd64  CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-freebsd.txt
GOOS=openbsd GOARCH=amd64  CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-openbsd.txt
GOOS=netbsd  GOARCH=amd64  CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-netbsd.txt
GOOS=android GOARCH=arm64  CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-android-arm64.txt
GOOS=android GOARCH=arm    CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-android-arm.txt
GOOS=ios     GOARCH=arm64  CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-ios-arm64.txt

# Race detector on the native platform
go test -race ./... 2>&1 | tee tmp/test-race.txt
go vet ./... 2>&1 | tee tmp/compat-vet.txt
```
Delete `tmp/` when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Exhaustive Cross-Platform Audit — Full Coverage Pass

**End-to-end rule**: Inspect every code site for each category below. Do not skip files, packages, or test files. Do not stop when finding count looks high. Process everything.

#### 3a. CGO and Native Dependencies (CRITICAL — zero tolerance in pure-Go scope)
Inspect every `.go` file without exception:

- [ ] Every occurrence of `import "C"` — record file and line; classify as a CRITICAL violation of the pure-Go requirement.
- [ ] Every `#cgo` directive — classify each as a CRITICAL violation of the pure-Go requirement.
- [ ] Cross-compile failures logged in Phase 2 baseline — each failure is a finding; trace to the root cause.
- [ ] Every dependency in `go.mod` — for each, verify whether `CGO_ENABLED=0 go build` of that dependency succeeds. A transitive CGO dependency is a CRITICAL finding.
- [ ] `unsafe` package usage — not a CGO issue, but document every `unsafe` call site; flag any that encode platform-specific memory layout assumptions (struct size, alignment, pointer width).

#### 3b. File System and Path Handling — Full File Scan
Inspect every file I/O operation in every file:

- [ ] Every string literal containing `/` or `\` that is used as a file path component — flag if not inside a `go:embed` directive (where forward slashes are correct and portable) or a URL string.
- [ ] Every call to `os.Open`, `os.Create`, `os.ReadFile`, `os.WriteFile`, `os.Mkdir`, `os.MkdirAll`, `os.Remove`, `os.Rename`, `os.Stat`, `os.Lstat` — verify the path argument uses `filepath.Join` or is derived from a platform-aware directory function.
- [ ] Every `filepath.Join` call — verify it is not mixing slash-separated path segments from a `go:embed` or URL context with OS paths.
- [ ] Every occurrence of `/tmp` hardcoded as a directory — replace with `os.TempDir()`.
- [ ] Every occurrence of `~/`, `$HOME`, `%USERPROFILE%`, `%APPDATA%` — replace with `os.UserHomeDir()` or `os.UserConfigDir()`.
- [ ] Every use of `os.Getenv("HOME")`, `os.Getenv("USERPROFILE")`, `os.Getenv("APPDATA")`, `os.Getenv("TEMP")`, `os.Getenv("TMP")` — replace with the appropriate `os.User*Dir()` function.
- [ ] File path case sensitivity assumptions — any logic that depends on uppercase/lowercase file names to distinguish files.

#### 3c. Build Constraints — Complete Coverage Check
For every package:

- [ ] Enumerate all `_GOOS.go` and `_GOOS_GOARCH.go` files in the package.
- [ ] For each platform-specific file, verify the `//go:build` tag matches the file name convention.
- [ ] For each exported or package-level symbol defined only in platform-specific files, verify that every supported platform has either a definition or a documented stub. Missing symbol = compile error on that platform.
- [ ] Verify no file uses only the deprecated `// +build` syntax without the modern `//go:build` syntax.
- [ ] For every `runtime.GOOS` switch, evaluate whether it covers all supported platforms including BSDs and mobile. An unhandled `default` path should return an error or no-op, not panic or produce undefined behavior.
- [ ] Verify that the CI/CD matrix (`.github/workflows/*.yml` or equivalent) actually builds for every stated target platform — a missing build job means the portability claim is untested.

#### 3d. Process Execution and Signal Handling — Full Scan
Inspect every `os/exec` and signal usage:

- [ ] Every `exec.Command`, `exec.CommandContext`, `exec.LookPath` call — must be in a package with a `//go:build !android,!ios` constraint or equivalent.
- [ ] Every package that imports `os/exec` — verify it has a build constraint or a mobile stub.
- [ ] Every `syscall.SIGHUP`, `syscall.SIGUSR1`, `syscall.SIGUSR2`, `syscall.SIGPIPE` — POSIX-only; must have `//go:build` constraint excluding Windows and mobile.
- [ ] Every `os/signal.Notify` call — verify the signal set is portable or guarded.
- [ ] Every `os.StartProcess` and `syscall.ForkExec` call — forbidden on mobile; must be build-constrained.

#### 3e. Networking on Constrained Platforms — Full Scan
Inspect every network I/O operation:

- [ ] Every `net.Dial`, `net.DialContext`, `net.Listen` call — `net.Listen` is forbidden on mobile (no server sockets without entitlements).
- [ ] Every `http.Get`, `http.Post`, `http.DefaultClient` call — verify a `context.Context` with timeout is used; mobile OSes kill background operations.
- [ ] Every network call without a context deadline or timeout — flag as HIGH on mobile targets.
- [ ] Every DNS resolution that does not handle failure gracefully — mobile networks are unreliable.
- [ ] Every hardcoded IP address — should use DNS for failover.
- [ ] IPv4-only assumptions (`0.0.0.0`, `127.0.0.1`, IPv4-formatted addresses hardcoded in listeners) — flag for IPv6 consideration on mobile.

#### 3f. Time Zones and Locale — Full Scan
Inspect every time-related operation:

- [ ] Every `time.LoadLocation` call — verify `time/tzdata` is imported (blank import) in the same binary's main or entry package if the binary targets mobile or Alpine Linux.
- [ ] Every `time.Local` usage where a specific timezone is assumed — flag if deterministic behavior is required.
- [ ] Every time format string — verify it does not use locale-sensitive formats.
- [ ] Search for the blank import `_ "time/tzdata"` — if absent and the project targets mobile, this is a HIGH finding for every `time.LoadLocation` call.

#### 3g. Integer Sizes and Endianness — Full Scan
Inspect every integer type usage and binary serialization:

- [ ] Every `int` or `uint` used in serialization, binary protocol, or on-disk format — must use explicitly sized types (`int32`, `int64`, `uint32`, `uint64`).
- [ ] Every use of `unsafe.Sizeof`, `unsafe.Alignof`, `unsafe.Offsetof` in logic that infers platform word size — flag for portability review.
- [ ] Every byte slice cast to or from a struct via `unsafe.Pointer` — endianness assumption; use `encoding/binary` with explicit byte order.
- [ ] Every shift by 32 or more bits on an `int` or `uint` value — may differ between 32-bit and 64-bit platforms.

#### 3h. Standard I/O and Terminal Interaction — Full Scan
Inspect every terminal or console interaction:

- [ ] Every ANSI escape code emission — verify it is gated on terminal capability detection.
- [ ] Every `os.Stdout.Fd()` or `os.Stderr.Fd()` cast to verify terminal — use `golang.org/x/term.IsTerminal` rather than POSIX-only `isatty` via CGO.
- [ ] Every `\r\n` or `\n` assumption in text file writing — document the choice and ensure it is correct for the target platform's line ending convention.
- [ ] Every assumption that `os.Stdin` is a terminal — broken for services, piped input, and mobile.

#### 3i. Mobile-Specific Constraints — Full Scan
For the complete package graph reachable from any mobile entry point:

- [ ] Every `os/exec` import in the reachable graph — forbidden on iOS, severely restricted on Android.
- [ ] Every `net.Listen` in the reachable graph — forbidden on mobile without entitlements.
- [ ] Every filesystem path that is not derived from `os.User*Dir()` or `go:embed` — sandbox violation on mobile.
- [ ] Every goroutine that blocks indefinitely — mobile OSes may suspend and resume the process, causing goroutines to stall.
- [ ] Every large in-memory data structure — document memory budget considerations for mobile (typically 512MB–2GB limit).
- [ ] Every `time.Sleep` > 1 second in a main goroutine path — iOS will watchdog-kill an app that is unresponsive.

#### 3j. BSD-Specific Constraints — Full Scan
For every file reachable when `GOOS=freebsd`, `GOOS=openbsd`, or `GOOS=netbsd`:

- [ ] Every `/proc/` path access — Linux-only; no `/proc` on BSD.
- [ ] Every `inotify` reference — Linux-only file watching; use `kqueue` on BSD or abstract via a cross-platform library.
- [ ] Every Linux-specific `syscall` constant (e.g., `syscall.O_DIRECT`, `syscall.SOCK_CLOEXEC` not available on all BSD variants) — must be guarded.
- [ ] Every `GOOS=freebsd` cross-compile failure — trace to root cause.
- [ ] Every external tool invocation assuming Linux-specific flags (e.g., `ps aux`, `free -m`, `ldd`) without a BSD variant or build constraint.

#### 3k. False-Positive Prevention (MANDATORY — applies even in end-to-end mode)
Before recording ANY finding, apply these checks:
1. **Verify the build constraint scope**: Code with a correct `//go:build` constraint is already scoped — verify the constraint actually matches the platform and the file name convention before reporting.
2. **Check for existing platform abstractions**: Trace the full call chain; an abstraction layer may already handle the portability concern.
3. **Assess actual target platforms**: Cross-validate every finding against the project's README-stated platform goals. A finding for an unstated platform is informational, not blocking.
4. **Read surrounding comments**: An explicit acknowledgment (e.g., `// Windows not supported`, `// Linux only`, `//nolint:`) is an acknowledged pattern.
5. **No CGO exceptions**: Any `import "C"` occurrence is a CRITICAL finding regardless of package, platform, or build constraint.
6. **Verify transitive CGO**: Run `CGO_ENABLED=0 go build [package]` before reporting a transitive CGO dependency — confirm the build actually fails for that specific package path.

**Rule**: Only report confirmed findings where you can state the exact file, line, code path, platform, and concrete failure mode.

### Phase 4: Report

Generate **`AUDIT.md`**:
```markdown
# CROSS-PLATFORM COMPATIBILITY AUDIT (END-TO-END) — [date]

## Project Platform Profile
[Purpose, stated target platforms, deployment model, pure-Go requirement summary]

## Cross-Compile Build Matrix
| Platform + Arch | CGO_ENABLED=0 Build | Notes |
|-----------------|---------------------|-------|
| windows/amd64   | ✅/❌ | [error summary if failed] |
| windows/arm64   | ✅/❌ | |
| darwin/amd64    | ✅/❌ | |
| darwin/arm64    | ✅/❌ | |
| linux/amd64     | ✅/❌ | |
| linux/arm64     | ✅/❌ | |
| linux/386       | ✅/❌ | |
| freebsd/amd64   | ✅/❌ | |
| openbsd/amd64   | ✅/❌ | |
| netbsd/amd64    | ✅/❌ | |
| android/arm64   | ✅/❌ | |
| android/arm     | ✅/❌ | |
| ios/arm64       | ✅/❌ | |

## Audit Coverage Log
[Track per-package audit completion across all checklist categories]
| Package | 3a CGO | 3b Paths | 3c Constraints | 3d Exec/Signals | 3e Network | 3f Time | 3g Integers | 3h Terminal | 3i Mobile | 3j BSD |
|---------|--------|----------|----------------|-----------------|------------|---------|-------------|-------------|-----------|--------|
| [pkg]   | ✅     | ✅       | ✅             | ✅              | ✅         | ✅      | ✅          | ✅          | ✅        | ✅     |

## CGO Status
[Explicit statement: PURE GO / VIOLATION — list every `import "C"` occurrence as a CRITICAL violation]

## Findings

### CRITICAL
- [ ] [Title] `[PLATFORM]` — [file:line] — [portability issue] — [failure mode: build error/runtime panic/wrong behavior] — **Remediation:** [specific fix with verification command]

### HIGH
- [ ] ...

### MEDIUM
- [ ] ...

### LOW
- [ ] ...

## Goal-Achievement Summary
| Stated Goal | Status | Blocking Findings |
|-------------|--------|-------------------|
| [Platform claim from README] | ✅ / ⚠️ / ❌ | [finding refs] |

## False Positives Considered and Rejected
| Candidate | Reason Rejected |
|-----------|----------------|
| [description] | [why not a real portability issue] |

## Remaining Scope (if session ended before completion)
| Package | Status | Notes |
|---------|--------|-------|
| [pkg] | Not yet audited | Resume here next session |
```

Generate **`GAPS.md`**:
```markdown
# Cross-Platform Compatibility Gaps (End-to-End) — [date]

## [Gap Title]
- **Stated Goal**: [what the project claims about platform support]
- **Current State**: [what platforms are actually supported and how]
- **Affected Platforms**: [WIN / MAC / LIN / BSD / AND / IOS]
- **Impact**: [build failure, runtime panic, incorrect behavior, missing feature, unplayable/unusable]
- **Closing the Gap**: [specific, actionable changes — build constraints, stdlib replacements, mobile stubs, go:embed, time/tzdata]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | CGO in non-entry-point packages violating pure-Go requirement; cross-compile failure for a stated target platform; runtime panic on a specific platform due to unavailable API |
| HIGH | Hardcoded path separator or directory causing incorrect behavior on at least one target; missing timezone data causing time failures on mobile; signal handling panic on Windows; `os/exec` reachable from mobile without build constraint |
| MEDIUM | Missing build constraint for a platform-specific API that degrades gracefully; mobile networking without context timeout; suboptimal but functional cross-platform behavior |
| LOW | Informational portability note; best-practice recommendation for a platform not yet targeted; minor style deviation from cross-platform idioms |

## Remediation Standards
Every finding MUST include a **Remediation** section that:
1. States exactly what to change (file, function, line range, import to add/remove, build constraint to add).
2. Uses only Go standard library or `golang.org/x` packages — no CGO, no platform-native SDKs.
3. Includes the cross-compile command that must pass after the fix.
4. Is proportionate — do not recommend architectural rewrites for LOW findings.

## Constraints
- Do not cap findings by count — report **every** confirmed portability issue regardless of total count.
- Output ONLY the two report files — no code changes permitted.
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Every finding must be tagged with the affected platform(s).
- Every CRITICAL or HIGH finding must include the concrete failure mode (build error, runtime panic, wrong behavior).
- Evaluate the code against **its own stated platform goals**, not every possible platform.
- Apply the Phase 3k false-positive prevention checks to every candidate finding before including it.
- The pure-Go / no-CGO constraint is non-negotiable: any CGO usage is automatically CRITICAL.
- Document remaining scope explicitly if the session ends before completion.

## Session Strategy
- Audit one package at a time; complete all checklist categories (3a–3j) for a package before moving on.
- Run cross-compile checks for each supported platform after auditing a package and before moving to the next.
- Update the Coverage Log in `AUDIT.md` after each package.
- If context is running low: write the in-progress `AUDIT.md` with the Coverage Log showing which packages remain, then stop. The next session picks up from the first unaudited package.
- Do NOT skip packages to meet a time or count goal.

## Tiebreaker
Within each severity group: CGO violations → cross-compile failures → runtime panics → incorrect behavior on more platforms first → missing platform stubs → informational notes. Then descending by number of affected platforms. If tied, descending by function cyclomatic complexity.
