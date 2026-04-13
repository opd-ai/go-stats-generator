# TASK: Perform a focused audit of memory management in Go code, identifying memory leaks, excessive allocations, unsafe pointer misuse, and GC pressure while rigorously preventing false positives.

## Execution Mode
**Report generation only** — do NOT modify any source code.

## Output
Write exactly two files in the repository root (the directory containing `go.mod`):
1. **`AUDIT.md`** — the memory audit report
2. **`GAPS.md`** — gaps in memory safety relative to the project's stated goals

If either file already exists, delete it and create a fresh one.

## Prerequisite
```bash
which go-stats-generator || go install github.com/opd-ai/go-stats-generator@latest
```

## Workflow

### Phase 0: Understand the Project's Memory Model
1. Read the project README to understand its purpose, users, and any claims about performance, memory efficiency, or scalability.
2. Examine `go.mod` for module path, Go version, and dependencies that affect memory behavior (e.g., `unsafe`, protobuf, cgo, pooling libraries).
3. List packages (`go list ./...`) and identify which packages handle large data, long-lived objects, or high-throughput processing.
4. Build a **memory inventory** by scanning for:
   - `unsafe.Pointer` and `uintptr` conversions
   - `reflect.SliceHeader` and `reflect.StringHeader` (deprecated in Go 1.20+)
   - `runtime.SetFinalizer` usage
   - `sync.Pool` usage and reuse patterns
   - Large slice/map allocations (`make([]T, n)` with large or unbounded `n`)
   - String↔byte conversions (`[]byte(s)`, `string(b)`) in hot paths
   - Append patterns that may cause repeated reallocation (`append` in a loop without pre-allocation)
   - `cgo` allocations (`C.malloc`, `C.CString`) and their corresponding frees
   - Global variables that accumulate data over time
   - Closures that capture large objects or slices
5. Identify the project's memory management conventions — does it use object pools? Does it document capacity expectations? Does it use `pprof`?
6. Determine the expected memory profile: is this a long-running server (leak-sensitive), a CLI tool (short-lived), or a library (caller-dependent)?

### Phase 1: Online Research
Use web search to build context that isn't available in the repository:
1. Search for the project on GitHub — read open issues mentioning "memory", "OOM", "leak", "allocation", or "GC" to understand known memory pain points.
2. Research key dependencies from `go.mod` for known memory issues, allocation overhead, or recommended usage patterns.
3. Look up common Go memory pitfalls relevant to the project's domain (e.g., HTTP body leaks, protobuf allocation, large JSON parsing).

Keep research brief (≤10 minutes). Record only findings that are directly relevant to the project's memory behavior.

### Phase 2: Baseline
```bash
set -o pipefail
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages,structs > /tmp/memory-audit-metrics.json
go-stats-generator analyze . --skip-tests
go test -race -count=1 ./... 2>&1 | tee /tmp/memory-test-results.txt
go vet ./... 2>&1 | tee /tmp/memory-vet-results.txt
```
Delete temporary files when done — the only persistent outputs are `AUDIT.md` and `GAPS.md`.

### Phase 3: Memory Audit

#### 3a. Memory Leak Analysis
For every long-lived object, goroutine, and accumulating data structure, verify:

**Goroutine-retained memory:**
- [ ] No goroutine holds references to large objects after the parent scope completes.
- [ ] Goroutines that block on channels or timers do not retain references to request-scoped data.
- [ ] `time.After` is not used in loops — each call creates a new timer that is not GC'd until it fires; use `time.NewTimer` with `Stop()` and drain instead.
- [ ] `time.NewTicker` is always stopped with `defer ticker.Stop()`.
- [ ] Context-derived goroutines terminate when the context is cancelled — stranded goroutines retain their closure variables.

**Slice and map retention:**
- [ ] Slices returned from functions do not retain a reference to a larger backing array (use `copy` or `slices.Clone` to decouple).
- [ ] Maps that grow during operation are periodically pruned or replaced — Go maps never shrink their internal bucket array.
- [ ] Nil-ing out slice/map entries that hold pointers, interfaces, or structs with pointers, to allow GC of referenced objects.
- [ ] `append` to a sub-slice does not silently overwrite the original backing array's subsequent elements.

**Global and package-level state:**
- [ ] Global maps, slices, or caches have a bounded size or TTL eviction strategy.
- [ ] `sync.Map` entries are periodically cleaned if the map is used as a cache.
- [ ] Package-level `init()` functions do not allocate large structures that persist for the process lifetime.

**HTTP and I/O body leaks:**
- [ ] `http.Response.Body` is always closed, even on error paths: `defer resp.Body.Close()`.
- [ ] `io.ReadCloser` values from any source are closed on all code paths.
- [ ] `io.LimitReader` or equivalent is used to bound reads from untrusted sources.
- [ ] `bufio.Scanner` and `bufio.Reader` buffers are not retained beyond their useful lifetime.

#### 3b. Excessive Allocation Analysis
For hot paths and high-throughput code, verify:

- [ ] String concatenation in loops uses `strings.Builder` or `bytes.Buffer`, not `+` or `fmt.Sprintf`.
- [ ] Slice capacity is pre-allocated with `make([]T, 0, expectedCap)` when the approximate size is known.
- [ ] Interface boxing of small value types in hot paths is minimized (each boxing allocates).
- [ ] Variadic function calls in hot paths do not create unnecessary slice allocations.
- [ ] `fmt.Sprintf` is not used for simple integer-to-string conversions where `strconv.Itoa` suffices.
- [ ] Map literals in loops are allocated once outside the loop or use pre-allocated maps with `clear()`.
- [ ] `regexp.Compile` is called once (at package level or `init()`), not inside functions called repeatedly.
- [ ] Struct pointers are preferred over large struct values in function signatures where the struct exceeds ~128 bytes.

#### 3c. Unsafe Pointer and Low-Level Memory
For every use of `unsafe`, verify:

- [ ] `unsafe.Pointer` conversions follow the six valid patterns documented in the `unsafe` package (no arbitrary arithmetic).
- [ ] `uintptr` values are not stored in variables — they must be used in a single expression to prevent the GC from moving the referenced object.
- [ ] No `reflect.SliceHeader` or `reflect.StringHeader` manipulation occurs (use `unsafe.Slice` and `unsafe.String` in Go 1.17+).
- [ ] `cgo` allocations (`C.CString`, `C.CBytes`, `C.malloc`) have corresponding `C.free` on all code paths, including error returns.
- [ ] `//go:nosplit`, `//go:noescape`, and `//go:linkname` directives are justified and do not mask memory safety issues.
- [ ] Pointer alignment assumptions are correct for the target architecture.

#### 3d. GC Pressure and Performance
- [ ] `sync.Pool` is used for frequently allocated/discarded objects in hot paths (e.g., buffers, temporary structs).
- [ ] `sync.Pool.Put` resets the object before returning it to the pool to avoid retaining references.
- [ ] Large allocations (>32KB) are aware that they go directly to the heap, bypassing size-class allocation.
- [ ] `runtime.SetFinalizer` is not used as a substitute for explicit cleanup (finalizers are non-deterministic and delay GC).
- [ ] Closures in hot paths do not inadvertently capture and retain large outer-scope variables.
- [ ] `runtime.KeepAlive` is used where necessary to prevent premature GC of objects passed to cgo or unsafe code.

#### 3e. False-Positive Prevention (MANDATORY)
Before recording ANY finding, apply these checks:

1. **Verify the execution context**: A memory pattern in a CLI tool that runs for <1 second is not a "leak." Reserve leak findings for long-running processes or code paths that execute repeatedly.
2. **Check for bounded lifetimes**: An allocation inside a request handler that is released when the handler returns is not a leak — it is normal per-request allocation.
3. **Trace the allocation path**: Confirm the allocation actually occurs on the path you claim. Use `go build -gcflags='-m'` escape analysis output as evidence where possible.
4. **Check for pooling or caching**: An object that appears "leaked" may be intentionally pooled or cached. Verify there is no `sync.Pool`, LRU cache, or similar mechanism managing its lifecycle.
5. **Read surrounding comments**: If a comment explicitly acknowledges a memory decision (e.g., `// intentionally cached`, `// pool managed`, `//nolint:`, or a TODO tracking a known issue), treat it as an acknowledged pattern — do not report it as a new finding.
6. **Assess materiality**: A 64-byte allocation in a cold path is not worth reporting. Focus on allocations that are either large (>1KB), frequent (hot path), or cumulative (grow unbounded over time).

**Rule**: If you cannot demonstrate that the memory issue causes measurable impact (unbounded growth, OOM risk, or significant GC pressure), do NOT report it. Speculative findings waste remediation effort.

### Phase 4: Report
Generate **`AUDIT.md`**:
```markdown
# MEMORY AUDIT — [date]

## Project Memory Profile
[Summary: long-running vs short-lived, expected memory behavior, stated performance goals, memory-sensitive dependencies]

## Memory Inventory
| Package | unsafe | sync.Pool | Large Allocs | Closures | cgo | Global State |
|---------|--------|-----------|-------------|----------|-----|-------------|
| [pkg]   | N      | N         | N           | N        | N   | ✅/❌        |

## Findings
### CRITICAL
- [ ] [Finding] — [file:line] — [evidence: allocation path, growth pattern] — [impact] — **Remediation:** [specific fix]
### HIGH / MEDIUM / LOW
- [ ] ...

## False Positives Considered and Rejected
| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| [description] | [why it is not a real memory issue] |
```

Generate **`GAPS.md`**:
```markdown
# Memory Safety Gaps — [date]

## [Gap Title]
- **Stated Goal**: [what the project claims about performance/memory/scalability]
- **Current State**: [what memory management exists]
- **Risk**: [what could go wrong under load or extended runtime]
- **Closing the Gap**: [specific changes needed]
```

## Severity Classification
| Severity | Criteria |
|----------|----------|
| CRITICAL | Confirmed unbounded memory growth, OOM risk under normal operation, use-after-free via unsafe, or cgo memory leak on a critical path |
| HIGH | Unbounded cache/map growth without eviction, goroutine leak retaining significant memory, or missing `resp.Body.Close()` on a request-handling path |
| MEDIUM | Excessive allocation in hot paths causing GC pressure, sub-optimal pre-allocation, or `time.After` in a loop |
| LOW | Minor allocation inefficiencies in cold paths, style preferences (e.g., `strings.Builder` vs `+` for small concatenations) |

## Remediation Standards
Every finding MUST include a **Remediation** section:
1. **Concrete fix**: State exactly what to change — specific allocation strategy, pooling approach, or cleanup pattern. Do not recommend "consider reducing allocations."
2. **Respect project idioms**: If the project uses `sync.Pool`, recommend pool-based fixes. If it avoids `unsafe`, don't introduce it.
3. **Verifiable**: Include a validation approach (e.g., `go test -bench=. -benchmem ./pkg/...`, `go build -gcflags='-m' ./...`, `go tool pprof`).
4. **Minimal scope**: Fix the memory issue without restructuring unrelated code.

## Constraints
- Output ONLY the two report files — no code changes permitted.
- Use `go-stats-generator` metrics as supporting evidence for complexity and function length.
- Every finding must reference a specific file and line number.
- All findings must use unchecked `- [ ]` checkboxes for downstream processing.
- Every finding must include evidence of material impact — no speculative findings.
- Evaluate the code against its **own stated goals** and memory profile, not arbitrary external standards.
- Apply the Phase 3e false-positive prevention checks to every candidate finding before including it.

## Tiebreaker
Prioritize: confirmed unbounded growth → unsafe/cgo memory corruption → hot-path allocation pressure → cold-path inefficiency → style. Within a level, prioritize by proximity to the project's critical paths and by magnitude of memory impact.
