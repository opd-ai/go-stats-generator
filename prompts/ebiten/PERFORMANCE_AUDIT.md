# TASK: [Ebitengine Edition] Perform a focused performance audit of Ebitengine game code for Go code, identifying hot-path inefficiencies, algorithmic complexity issues, unnecessary allocations, I/O bottlenecks, and scalability limiters while rigorously preventing false positives.

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
1. **`AUDIT.md`** — the performance audit report
2. **`GAPS.md`** — gaps between stated performance goals and actual implementation

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

### Phase 0: Understand the Project's Performance Profile
1. Read the project README to understand its purpose, users, and any claims about performance, throughput, latency, scalability, or resource usage.
2. Examine `go.mod` for module path, Go version, and performance-relevant dependencies (e.g., `sync`, `golang.org/x/sync`, caching libraries, connection pools, serialization libraries).
3. List packages (`go list ./...`) and identify which packages are on critical performance paths (request handling, data processing, I/O, computation).
4. Build a **performance profile** by identifying:
   - The project's primary workload type: I/O-bound, CPU-bound, or memory-bound
   - Hot paths: code executed per-request, per-item, or in tight loops
   - Cold paths: initialization, configuration, shutdown
   - Stated performance targets: throughput, latency, file count, processing time
   - Concurrency model: single-threaded, goroutine-per-request, worker pool, pipeline
5. Identify the project's performance conventions — does it use benchmarks? Does it pool allocations? Does it cache results? Does it use profiling?
6. Run existing benchmarks if available: `go test -bench=. -benchmem ./...`

### Phase 1: Online Research
Use web search to build context that isn't available in the repository:
1. Search for the project on GitHub — read open issues mentioning "slow", "performance", "memory", "timeout", "scale", or "bottleneck" to understand known performance pain points.
2. Research key dependencies from `go.mod` for known performance characteristics, scaling limits, or recommended configuration for high-throughput use.
3. Look up performance benchmarks for similar tools in the project's domain to calibrate expectations.

Keep research brief (≤10 minutes). Record only findings that are directly relevant to the project's performance goals.

### Phase 2: Baseline
```bash
set -o pipefail
mkdir -p tmp
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages,structs > tmp/perf-audit-metrics.json
go-stats-generator analyze . --skip-tests
go test -race -count=1 ./... 2>&1 | tee tmp/perf-test-results.txt
go test -bench=. -benchmem ./... 2>&1 | tee tmp/perf-bench-results.txt 2>/dev/null || true
```
Delete temporary files when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Performance Audit

#### 3a. Algorithmic Complexity
For every function on a hot path, verify the algorithm is appropriate for the expected input size:

- [ ] No O(n²) or worse algorithms where O(n log n) or O(n) alternatives exist and the input can grow large.
- [ ] No linear search (`for range` over a slice) where a map lookup would be appropriate for repeated lookups.
- [ ] No repeated sorting of the same data — sort once and reuse.
- [ ] String matching uses efficient algorithms (not nested loops or repeated `strings.Contains` in a loop).
- [ ] Regular expressions are compiled once (`regexp.MustCompile` at package level), not inside functions called repeatedly.
- [ ] Map keys and hash computations are efficient — complex struct keys may cause expensive hashing.
- [ ] Recursive algorithms have bounded depth or use iterative alternatives for large inputs.

#### 3b. Allocation and GC Pressure
For every hot path, verify allocation efficiency:

- [ ] Slices are pre-allocated with `make([]T, 0, expectedCap)` when the approximate size is known.
- [ ] String concatenation in loops uses `strings.Builder` or `bytes.Buffer`, not `+` or `fmt.Sprintf`.
- [ ] `sync.Pool` is used for frequently allocated/discarded objects (buffers, temporary structs).
- [ ] Interface boxing of small values in hot paths is minimized (each boxing may allocate).
- [ ] `fmt.Sprintf` is not used for simple conversions where `strconv.Itoa`, `strconv.AppendInt`, etc. suffice.
- [ ] Map literals in loops are pre-allocated or reused with `clear()` instead of re-created.
- [ ] Large struct values are passed by pointer, not by value, in hot paths.
- [ ] Closure captures do not inadvertently retain large objects beyond their useful lifetime.

#### 3c. I/O Efficiency
For every I/O operation on a hot path, verify:

- [ ] File reads use buffered I/O (`bufio.Reader`, `bufio.Scanner`) instead of unbuffered byte-at-a-time reads.
- [ ] File writes are buffered and flushed appropriately — not flushing after every write in a loop.
- [ ] Network I/O uses connection pooling (HTTP client reuse, database connection pools) instead of connect-per-request.
- [ ] Disk I/O is minimized — data is processed in memory where possible, not written to temp files unnecessarily.
- [ ] `io.Copy` is used for stream-to-stream transfers instead of reading into memory and writing back out.
- [ ] JSON/XML encoding uses `json.NewEncoder` to stream to an `io.Writer` instead of `json.Marshal` + `Write` for large payloads.
- [ ] Directory traversal uses `filepath.WalkDir` (not `filepath.Walk`) to avoid unnecessary `os.Stat` calls.

#### 3d. Concurrency Efficiency
For every concurrent operation, verify the concurrency model is appropriate:

- [ ] Worker pools are bounded — unbounded `go func()` in a loop can exhaust memory and cause scheduler thrashing.
- [ ] Channels are appropriately buffered — unbuffered channels on hot paths cause unnecessary goroutine context switches.
- [ ] Lock granularity is appropriate — a single global mutex on a hot path serializes all goroutines.
- [ ] `sync.RWMutex` is used instead of `sync.Mutex` for read-heavy workloads.
- [ ] Context cancellation is checked early in long-running operations to avoid wasted work.
- [ ] Pipeline stages are balanced — a fast producer with a slow consumer creates backpressure without flow control.
- [ ] `runtime.GOMAXPROCS` is not artificially limited (unless there is a documented reason).

#### 3e. Caching and Memoization
- [ ] Repeated expensive computations on the same input use caching or memoization.
- [ ] Cache invalidation is correct — stale cache entries do not cause incorrect results.
- [ ] Cache size is bounded to prevent memory exhaustion.
- [ ] Cache hit rates are observable — if a cache never hits, it is wasted memory.
- [ ] Immutable data is computed once and shared, not recomputed per-request.

#### 3f. Serialization and Parsing Efficiency
- [ ] JSON parsing of large payloads uses streaming decoders (`json.NewDecoder`) instead of `json.Unmarshal` on the full byte slice.
- [ ] Custom `MarshalJSON`/`UnmarshalJSON` methods avoid unnecessary allocations.
- [ ] Protocol buffer or other binary formats are used instead of JSON/XML for high-throughput internal communication.
- [ ] AST parsing and file reading operations minimize repeated parsing of the same file.

#### 3g. False-Positive Prevention (MANDATORY)
Before recording ANY finding, apply these checks:

1. **Verify it is a hot path**: An inefficiency in code that runs once during initialization is not a performance finding. Focus on code that executes per-request, per-file, per-item, or in tight loops.
2. **Assess the input scale**: An O(n²) algorithm on a list that never exceeds 10 items is not a performance issue. Evaluate against the project's stated or realistic input sizes.
3. **Check for existing optimization**: The code may already use caching, pooling, or batching that you haven't traced. Verify the full call chain before flagging.
4. **Read surrounding comments**: If a comment explains a performance decision (e.g., `// pre-allocated for expected size`, `// pooled`, `//nolint:`, or a TODO tracking a known optimization opportunity), treat it as acknowledged.
5. **Measure, don't guess**: If benchmarks exist, use their results. If they don't, state your complexity analysis clearly but acknowledge that profiling would confirm the impact.
6. **Respect the project's performance tier**: A CLI tool processing a few files does not need the same optimization as a high-throughput API server. Calibrate expectations.

**Rule**: If you cannot demonstrate that the inefficiency causes measurable impact at the project's stated scale, do NOT report it as HIGH or CRITICAL. Label it as a potential optimization and classify as MEDIUM or LOW.

### Phase 4: Report
Generate **`AUDIT.md`**:
```markdown
# PERFORMANCE AUDIT — [date]

## Project Performance Profile
[Workload type, hot paths, cold paths, stated performance targets, concurrency model]

## Performance Inventory
| Package | Hot Path Functions | Allocations in Hot Path | I/O Operations | Concurrency Primitives | Caching |
|---------|-------------------|------------------------|----------------|----------------------|---------|
| [pkg] | N | [high/medium/low] | N | [type] | ✅/❌ |

## Benchmark Results
[Summary of existing benchmark results, if available]

## Findings
### CRITICAL
- [ ] [Finding] — [file:line] — [complexity/allocation/IO evidence] — [impact on stated performance target] — **Remediation:** [specific optimization]
### HIGH / MEDIUM / LOW
- [ ] ...

## False Positives Considered and Rejected
| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| [description] | [why it is not a real performance issue at this project's scale] |
```

Generate **`GAPS.md`**:
```markdown
# Performance Gaps — [date]

## [Gap Title]
- **Stated Goal**: [what the project claims about performance/throughput/latency]
- **Current State**: [what the implementation achieves]
- **Bottleneck**: [what limits performance]
- **Closing the Gap**: [specific optimization needed]
- **Expected Improvement**: [estimated impact, if measurable]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | Algorithmic complexity that prevents the project from meeting its stated performance targets at expected scale, or a hot-path bottleneck causing timeouts/OOM under normal load |
| HIGH | O(n²) on a hot path with growing input, unbounded goroutine creation, or missing connection pooling causing resource exhaustion under load |
| MEDIUM | Unnecessary allocations in hot paths causing GC pressure, missing buffered I/O, or suboptimal lock granularity |
| LOW | Cold-path inefficiencies, potential optimizations that would not measurably improve user experience, or missing benchmarks |

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Concrete optimization**: State exactly what to change — algorithm, data structure, pooling strategy, buffering approach. Do not recommend "consider optimizing."
2. **Respect project idioms**: If the project avoids external dependencies, do not recommend importing a caching library. Use the standard library.
3. **Verifiable**: Include a validation approach (e.g., `go test -bench=BenchmarkX -benchmem`, `go tool pprof`, `go build -gcflags='-m'`).
4. **Measured**: Whenever possible, include the expected complexity improvement (e.g., "O(n²) → O(n) for N items").

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Use `go-stats-generator` metrics as supporting evidence (especially function length and complexity for hot-path identification).
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Every finding must include evidence of the performance impact — no speculative optimizations reported as CRITICAL.
- Evaluate the code against its **own stated performance targets**, not arbitrary benchmarks from different projects.
- Apply the Phase 3g false-positive prevention checks to every candidate finding before including it.

## Tiebreaker
Prioritize: algorithmic complexity on hot paths → allocation pressure in tight loops → I/O bottlenecks → concurrency inefficiency → cold-path optimizations. Within a level, prioritize by frequency of execution (per-request > per-file > per-session > one-time).
