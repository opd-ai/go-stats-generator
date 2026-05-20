# TASK: Perform a focused cross-platform compatibility audit of a Go project, verifying compatibility with Windows, macOS, Linux, FreeBSD, OpenBSD, NetBSD, Android, and iOS while enforcing pure-Go (no CGO), Go standard library idioms, and platform-agnostic code practices.

## Execution Mode
**Report generation only** — do NOT modify any source code.

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
1. Read the project README to understand its stated target platforms, deployment model, and any cross-platform claims.
2. Examine `go.mod` for module path, Go version, and dependencies — identify any dependency that pulls in CGO or platform-specific native libraries.
3. List packages (`go list ./...`) and classify each by cross-platform risk: filesystem I/O, process execution, networking, UI, signal handling, or pure computation.
4. Build a **platform surface inventory** by scanning for:
   - `import "C"` — CGO usage (incompatible with pure-Go requirement)
   - `//go:build` constraints — platform-specific code paths
   - `runtime.GOOS` and `runtime.GOARCH` switches — runtime platform detection
   - `os/exec` — process execution (behavior differs per platform and is unavailable on some mobile targets)
   - `syscall` and `golang.org/x/sys` — direct syscall usage
   - `filepath` vs `path` package usage — path separator handling
   - `os.PathSeparator`, `os.PathListSeparator` — hardcoded separators
   - `\` or `/` hardcoded in file path strings
   - `os.TempDir()`, `os.UserHomeDir()`, `os.UserCacheDir()` — platform-aware directory resolution
   - `time/tzdata` import — embedded timezone data for platforms without a system timezone database
   - Signal handling (`os/signal`, `syscall.SIGTERM`, `syscall.SIGINT`) — signal availability differs per platform
   - `net` package usage — networking restrictions on mobile platforms
   - `go:embed` directives — verify embedded assets are portable
5. Map any `_windows.go`, `_linux.go`, `_darwin.go`, `_freebsd.go`, `_openbsd.go`, `_netbsd.go`, `_android.go`, `_ios.go` files to understand what platform divergence already exists.
6. Identify the project's build strategy: does it use a `Makefile`, `goreleaser`, or GitHub Actions matrix to build for multiple platforms?

### Phase 1: Online Research
Use web search to build context:
1. Search for the project on GitHub — read open issues mentioning "windows", "darwin", "linux", "freebsd", "android", "ios", "cross-platform", "CGO", or "build constraint" to understand known compatibility pain points.
2. Research key dependencies from `go.mod` for CGO requirements, platform restrictions, or known mobile incompatibilities.
3. Look up Go mobile documentation (`golang.org/x/mobile`) if the project targets Android or iOS.
4. Check `GOOS` and `GOARCH` support matrix for any third-party packages used.

Keep research brief (≤10 minutes). Record only findings directly relevant to cross-platform compatibility.

### Phase 2: Baseline
```bash
set -o pipefail
mkdir -p tmp

go-stats-generator analyze . --skip-tests --format json --sections functions,packages,patterns > tmp/compat-audit-metrics.json
go-stats-generator analyze . --skip-tests

# Verify CGO-free build
CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/cgo-free-build.txt

# Cross-compile spot checks (requires appropriate toolchains)
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-windows.txt
GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-darwin.txt
GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-linux.txt
GOOS=freebsd GOARCH=amd64 CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-freebsd.txt
GOOS=openbsd GOARCH=amd64 CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-openbsd.txt
GOOS=netbsd  GOARCH=amd64 CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-netbsd.txt
GOOS=android GOARCH=arm64 CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-android.txt
GOOS=ios     GOARCH=arm64 CGO_ENABLED=0 go build ./... 2>&1 | tee tmp/build-ios.txt

go vet ./... 2>&1 | tee tmp/compat-vet.txt
```
Delete temporary files when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Cross-Platform Compatibility Audit

#### 3a. CGO and Native Dependencies (CRITICAL for pure-Go requirement)
For every package in the module, verify:

- [ ] No `import "C"` appears in any `.go` file — CGO is prohibited.
- [ ] No `#cgo` directives appear in any Go source file.
- [ ] No dependency in `go.mod` transitively requires CGO (check with `CGO_ENABLED=0 go build ./...`; a build failure indicates a CGO dependency).
- [ ] No `//go:cgo_*` build directives are present.
- [ ] No `syscall.RawSyscall` or `syscall.Syscall` calls bypass the CGO-free constraint on platforms where they are unavailable.
- [ ] `golang.org/x/sys` is used instead of the deprecated `syscall` package where platform abstraction is needed (and only for platforms that support it).

#### 3b. File System and Path Handling
For every file path operation, verify:

**Path construction:**
- [ ] `filepath.Join` is used for all OS file path construction — never string concatenation with `/` or `\`.
- [ ] `path` (URL/POSIX paths) and `filepath` (OS paths) are not confused: `path` is for forward-slash paths (URLs, embedded assets), `filepath` is for OS-native paths.
- [ ] No hardcoded path separators (`/` or `\`) appear in OS file path strings outside of test fixtures.
- [ ] `os.PathSeparator` and `os.PathListSeparator` are used where path separators must be referenced programmatically.

**Directory resolution:**
- [ ] `os.TempDir()` is used for temporary files — not `/tmp` (which does not exist on Windows or mobile).
- [ ] `os.UserHomeDir()` is used for user-home-relative paths — not `~/` or `$HOME` expansion.
- [ ] `os.UserCacheDir()` is used for cache storage — not hardcoded `~/.cache` or `%APPDATA%`.
- [ ] `os.UserConfigDir()` is used for configuration storage — not hardcoded `~/.config` or `%APPDATA%`.
- [ ] `os.Executable()` is used to locate the binary's directory — not `os.Args[0]` which may be relative or a symlink.

**Case sensitivity:**
- [ ] File names are treated as case-sensitive in logic (Windows is case-insensitive, Linux/BSD are case-sensitive, macOS is case-insensitive by default).
- [ ] No logic depends on file name case to distinguish files (e.g., `Config.go` vs `config.go` in the same package).

**Mobile file system restrictions:**
- [ ] On Android/iOS targets, only the app's sandboxed directories (`os.UserCacheDir`, `os.UserHomeDir`, or platform-provided paths) are used for file I/O — no absolute system paths.
- [ ] `go:embed` is used for bundling read-only assets rather than reading from the file system at runtime on mobile.

#### 3c. Build Constraints and Platform-Specific Code
For every `//go:build` constraint and platform-specific file, verify:

- [ ] Build tags use the modern `//go:build` syntax (Go ≥ 1.17), not the deprecated `// +build` syntax alone.
- [ ] Every platform-specific file (e.g., `_windows.go`) has a matching `//go:build` constraint that matches its file name suffix.
- [ ] Each platform-specific implementation has a corresponding implementation for every other supported platform — no platform is left without a fallback.
- [ ] `runtime.GOOS` and `runtime.GOARCH` switches inside functions are used only for behavior that cannot be separated at build time; prefer build constraints over runtime switches for significant code divergence.
- [ ] Stub or no-op implementations are provided for platforms where a feature is not available, rather than compile errors or silent missing functionality.
- [ ] The `//go:build ignore` tag is not accidentally applied to production code files.

#### 3d. Process Execution and OS Interaction
For every process execution or OS interaction, verify:

**Process execution:**
- [ ] `os/exec` usage includes a build constraint excluding Android and iOS — process execution is not available on mobile platforms.
- [ ] External command names do not assume a specific shell or shell extensions (e.g., `.exe` suffix on Windows, shell builtins on POSIX).
- [ ] Commands that differ by platform (e.g., `cmd.exe /C` on Windows vs `/bin/sh -c` on POSIX) use platform-specific files or `runtime.GOOS` guards.
- [ ] `PATH` lookup via `exec.LookPath` accounts for `.exe` extensions on Windows.

**Signals:**
- [ ] Signal handling uses `syscall.SIGTERM` and `os.Interrupt` — not `syscall.SIGKILL` (which cannot be caught) or POSIX-only signals not available on Windows or mobile.
- [ ] Signal handling code is guarded with a build constraint for platforms that do not support the `os/signal` package (e.g., some mobile runtime configurations).
- [ ] `SIGHUP` (reload), `SIGUSR1`/`SIGUSR2` (custom) usage is gated on POSIX-only build constraints.

**Environment variables:**
- [ ] `os.Getenv` is used for environment variable access — not platform-specific registry or plist reads without a build constraint.
- [ ] Windows-style environment variable names (all-caps, no underscores for some) are handled where the project supports both POSIX and Windows conventions.
- [ ] `os.ExpandEnv` or `os.Expand` is used for variable expansion — not manual string manipulation of `$VAR` or `%VAR%`.

#### 3e. Networking on Constrained Platforms
For any network I/O, verify:

- [ ] Network operations use `context.Context` with timeouts — mobile platforms aggressively kill background network operations.
- [ ] The code handles `net.Error` with `Timeout() == true` and `Temporary() == true` correctly for transient mobile network changes (airplane mode, WiFi handoff).
- [ ] No assumption is made that a network connection persists indefinitely — mobile OSes may suspend the process.
- [ ] DNS resolution uses `net.DefaultResolver` — hardcoded DNS servers or resolver customization may fail on mobile.
- [ ] `net/http` usage on Android/iOS does not rely on raw socket access (`net.Dial`, `net.Listen`) which may require special permissions on mobile.
- [ ] IPv6 compatibility is considered — some mobile networks are IPv6-only.

#### 3f. Time Zones and Locale
- [ ] `time/tzdata` is imported (blank import `_ "time/tzdata"`) if the binary must work on platforms without a system timezone database (Alpine Linux, Android, iOS, scratch containers).
- [ ] Time formatting uses `time.RFC3339` or explicit format constants — not locale-sensitive `time.Format` with platform-dependent strings.
- [ ] `time.LoadLocation` errors are handled — the timezone database may not exist on the target platform without `time/tzdata`.
- [ ] `time.Local` is not assumed to be UTC or any specific zone — use explicit zones where determinism is required.

#### 3g. Integer Sizes and Endianness
- [ ] `int` and `uint` are not used where a specific bit width is required — use `int32`, `int64`, `uint32`, `uint64` (e.g., for serialization, binary protocols, or values stored on disk).
- [ ] Binary serialization uses `encoding/binary` with explicit byte order — not `unsafe.Pointer` casts or type punning that assume endianness.
- [ ] `uintptr` is not used to store pointer values that outlive the GC — it is not a pointer type and the GC does not trace it.
- [ ] `unsafe.Sizeof` and `unsafe.Alignof` are not used to infer platform word size for logic that should use `bits.UintSize` or build constraints.

#### 3h. Standard I/O and Terminal Interaction
- [ ] Terminal detection uses `os.Stdout.Fd()` with a portable method (e.g., `golang.org/x/term`) — not POSIX-only `isatty` via CGO.
- [ ] ANSI escape codes for color output are gated on terminal capability detection — not always emitted (they display as garbage on Windows cmd.exe without VT mode enabled, and in non-terminal environments).
- [ ] Line endings are handled correctly: files written with `\n` are acceptable cross-platform for text, but binary files must not normalize line endings.
- [ ] Standard input reading does not assume a terminal is always available (piped input, services, mobile).
- [ ] Console output on Windows handles the case where `CONOUT$` is not available (e.g., Windows service context).

#### 3i. Mobile-Specific Constraints (Android and iOS)
For any code intended to run on Android or iOS:

- [ ] No `os/exec` calls — process forking is not permitted on iOS and is highly restricted on Android.
- [ ] No direct filesystem access outside the app sandbox — use platform-provided directory functions.
- [ ] `gomobile bind` or `gomobile build` compatibility: no `main` package functions that take unsupported types across the FFI boundary if using `gomobile`.
- [ ] Background processing is designed for short execution windows — iOS suspends apps aggressively.
- [ ] No `cgo` import, even transitively — the app store submission process for iOS requires pure-Go or specific entitlements.
- [ ] Memory limits are respected — mobile devices have significantly less RAM than desktop; avoid large in-memory data structures.
- [ ] No `net.Listen` for arbitrary ports — mobile apps cannot act as network servers without special entitlements.

#### 3j. BSD-Specific Concerns (FreeBSD, OpenBSD, NetBSD)
For any code that may run on BSD variants:

- [ ] `kqueue`-based or epoll-based assumptions are not hardcoded — use `net/http` or `net.Conn` abstractions.
- [ ] `/proc` filesystem usage is guarded with a Linux-only build constraint — `/proc` does not exist on BSD systems.
- [ ] `inotify` usage (Linux file watching) is guarded with a Linux-only build constraint — use `golang.org/x/sys/unix` with kqueue on BSD, or abstract via a cross-platform library.
- [ ] `SO_REUSEPORT` socket option usage is gated on platforms that support it — BSD support varies.
- [ ] Tool or command names that differ on BSD (e.g., `ps`, `df`, `sed` flags) are handled with platform-specific code if the project shells out.

#### 3k. False-Positive Prevention (MANDATORY)
Before recording ANY finding, apply these checks:

1. **Verify the build constraint scope**: A function in `network_linux.go` that uses Linux-only APIs is not a cross-platform bug — it has a build constraint. Verify the constraint matches the usage.
2. **Check for existing platform abstractions**: The project may already have a `platform.go` or `os_*.go` that provides the correct abstraction. Trace the full call chain before flagging a usage as non-portable.
3. **Assess actual target platforms**: If the project README explicitly states it only targets Linux and macOS, BSD and mobile findings are informational, not blocking. Evaluate against stated goals.
4. **Read surrounding comments**: A comment acknowledging a platform limitation (e.g., `// Windows not supported`, `// Linux only`, `//nolint:`) is an acknowledged pattern — do not report it as a new finding.
5. **Verify the CGO analysis**: A build error with `CGO_ENABLED=0` from a test-only dependency or a build-tagged file may not affect the production binary. Distinguish test and production scope.
6. **Check for stdlib equivalents**: Before flagging a platform-specific API, verify there is no standard library equivalent. Many Unix-isms have portable equivalents in the Go stdlib.

**Rule**: If a platform-specific pattern is already guarded by a correct build constraint or runtime check, it is NOT a finding. Only report code that is reachable on a platform it is incompatible with.

### Phase 4: Report
Generate **`AUDIT.md`**:
```markdown
# CROSS-PLATFORM COMPATIBILITY AUDIT — [date]

## Target Platform Matrix
| Platform | Build Status | Known Issues |
|----------|-------------|--------------|
| Windows (amd64) | ✅/⚠️/❌ | [summary] |
| macOS (arm64) | ✅/⚠️/❌ | [summary] |
| Linux (amd64) | ✅/⚠️/❌ | [summary] |
| FreeBSD (amd64) | ✅/⚠️/❌ | [summary] |
| OpenBSD (amd64) | ✅/⚠️/❌ | [summary] |
| NetBSD (amd64) | ✅/⚠️/❌ | [summary] |
| Android (arm64) | ✅/⚠️/❌ | [summary] |
| iOS (arm64) | ✅/⚠️/❌ | [summary] |

## CGO Status
[Result of CGO_ENABLED=0 go build ./... — PASS or FAIL with details]

## Platform Surface Inventory
| Package | CGO | Exec | Signals | File I/O | Network | Build Tags |
|---------|-----|------|---------|----------|---------|------------|
| [pkg]   | ✅/❌ | N   | N       | N        | N       | [tags]     |

## Findings
### CRITICAL
- [ ] [Finding] `[PLATFORM]` — [file:line] — [portability issue] — [impact: build failure, panic, wrong behavior] — **Remediation:** [specific fix]
### HIGH / MEDIUM / LOW
- [ ] ...

## False Positives Considered and Rejected
| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| [description] | [why it is acceptable in this context] |
```

Generate **`GAPS.md`**:
```markdown
# Cross-Platform Compatibility Gaps — [date]

## [Gap Title]
- **Stated Goal**: [what the project claims about platform support]
- **Current State**: [what platforms are actually supported and how]
- **Affected Platforms**: [WIN / MAC / LIN / BSD / AND / IOS]
- **Impact**: [build failure, runtime panic, incorrect behavior, missing feature]
- **Closing the Gap**: [specific changes needed — build constraints, stdlib replacements, platform stubs]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | CGO dependency breaking the pure-Go requirement; cross-compile failure for a stated target platform; runtime panic on a specific platform due to unavailable API |
| HIGH | Path separator or hardcoded directory causing incorrect behavior on at least one target platform; missing timezone data causing time parsing failures on mobile; signal handling code that panics on Windows |
| MEDIUM | Missing build constraint for a platform-specific API that degrades gracefully; suboptimal but functional cross-platform behavior; missing mobile-specific fallback |
| LOW | Informational portability note; best-practice recommendation for a platform not yet targeted; minor style deviation from cross-platform idioms |

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Concrete fix**: State exactly what to change — the specific import, function, path operation, or build constraint. Do not recommend "use portable APIs."
2. **Preserve Go idioms**: Fixes must use Go standard library packages only — no CGO, no platform-native SDKs.
3. **Verifiable**: Include the cross-compile command that should pass after the fix (e.g., `GOOS=windows CGO_ENABLED=0 go build ./...`).
4. **Minimal scope**: Fix the portability issue without restructuring unrelated code.

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Use `go-stats-generator` metrics as supporting evidence for complexity and package structure.
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Every finding must be tagged with the affected platform(s) using the tags defined above.
- Every finding must include the concrete failure mode (build error, runtime panic, wrong behavior) — no speculative findings.
- Evaluate the code against its **own stated platform goals**, not every possible platform.
- Apply the Phase 3k false-positive prevention checks to every candidate finding before including it.
- The pure-Go / no-CGO constraint is non-negotiable: any CGO usage is automatically CRITICAL regardless of platform.

## Tiebreaker
Prioritize: CGO violations → cross-compile failures → runtime panics on target platforms → incorrect behavior → missing platform stubs → informational portability notes. Within a level, prioritize by number of affected platforms (more platforms = higher priority).
