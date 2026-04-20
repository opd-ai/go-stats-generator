# PERFORMANCE AUDIT — 2026-04-20

## Project Performance Profile

| Dimension | Value |
|-----------|-------|
| **Workload type** | CPU-bound (AST parsing, complexity analysis) + I/O-bound (file discovery, line counting) |
| **Primary use-case** | Batch analysis of Go source trees; CLI invoked once per CI run or developer session |
| **Stated throughput target** | 50,000+ files in ≤60 seconds |
| **Stated memory target** | Peak memory usage ≤ 1 GB |
| **Stated throughput (README)** | 987 files/sec on AMD Ryzen 7 7735HS (16 cores) |
| **Stated memory per file (README)** | ~62 KB peak memory per file |
| **Concurrency model** | Fixed-size worker pool (default: `runtime.NumCPU()`) with buffered job/result channels |
| **Hot paths** | `DiscoverFiles` → `ProcessFiles` (parallel) → per-file: `AnalyzeFunctionsWithPath`, `ExtractBlocks`, `AnalyzeDuplication` → `finalizeReport` |
| **Cold paths** | Configuration loading, storage backend initialisation, reporter construction |
| **Caching strategy** | None at analysis layer; `enable_cache: true` in config has no runtime effect on the core AST/analysis pipeline |

## Benchmark Results (This Run — 4-Core CI Runner)

```
BenchmarkFullAnalysis_CurrentCodebase-4          4   368 153 469 ns/op   155 264 292 B/op   1 040 927 allocs/op
BenchmarkFullAnalysis_WithMemoryTracking-4       3   381 801 273 ns/op    17.31 MB_peak       611 files/sec
BenchmarkDuplicationAnalysis_LargeCodebase-4     4   269 798 853 ns/op    78 001 254 B/op   1 491 182 allocs/op
BenchmarkDuplicationAnalysis_BlockNormalization-4  412   2 779 103 ns/op      796 474 B/op      17 147 allocs/op
BenchmarkDuplicationAnalysis_MultipleFiles-4    3858     280 714 ns/op      102 045 B/op       2 106 allocs/op
BenchmarkPlacementAnalysis_LargeCodebase-4       214   5 468 812 ns/op      193 072 B/op       1 585 allocs/op
```

**Observed**: 611 files/sec on 4 cores (vs 987 files/sec on 16 cores in PERFORMANCE.md).  
**Allocated per run**: 155 MB for 211 files ≈ **735 KB allocated/file** (vs 62 KB claimed).  
**Peak RSS per run**: 17.31 MB for 211 files ≈ 82 KB RSS/file.  
Projecting linearly to 50,000 files: **~82 seconds** throughput, **~19.5 GB allocated** — both exceed stated targets.

## Performance Inventory

| Package | Hot-Path Functions | Allocations in Hot Path | I/O Operations | Concurrency Primitives | Caching |
|---------|-------------------|------------------------|----------------|----------------------|---------|
| `internal/scanner` | `DiscoverFiles`, `analyzeFile`, `ParseFile`, `ProcessFiles` | Medium | **High** — 2 `os.ReadFile` per file in discovery | Bounded worker pool, buffered channels | ❌ |
| `internal/analyzer` (function) | `AnalyzeFunctionsWithPath`, `readFileLines`, `countLinesInRange` | **High** — `os.ReadFile` + `strings.Split` per function | **High** — 1 disk read per function analysed | None | ❌ |
| `internal/analyzer` (duplication) | `ExtractBlocks`, `extractBlocksFromStmtList`, `NormalizeBlock`, `FingerprintBlocks`, `DetectClonePairs` | **High** — 17K allocs per normalization batch | None (in-memory AST) | None | ❌ |
| `internal/analyzer` (naming) | `toSnakeCase`, `AnalyzeFileNames` | Medium — regex compiled per call | None | None | ❌ |
| `internal/analyzer` (package) | `analyzePackageImports`, `mergeUniqueStrings`, `isKnownInternalPrefix`, `isKnownExternalPackage` | Medium | None | None | ❌ |
| `pkg/generator` | `buildReport`, `processFile`, `finalizeReport`, `AnalyzeDuplication` | **High** — accumulates all ASTs in `collected.files` | None (post-parse) | None | ❌ |
| `internal/reporter` | `Generate` (JSON), `Generate` (console) | Low | Low (buffered `json.NewEncoder`) | None | ❌ |

---

## Findings

### CRITICAL

- [ ] **Per-function file re-read — I/O amplification on every function analysed** — `internal/analyzer/function.go:219-231` — `readFileLines` calls `os.ReadFile(fileName)` for every function declaration in every file. For a file with F functions, the file is read F times during analysis. Discovery already reads the file twice (`discover.go:66`, `discover.go:111`). A file with 50 functions incurs 52 disk reads. At 50,000 files averaging 20 functions each, this produces ~1,050,000 file-read syscalls. With duplication analysis holding all ASTs, the OS page cache does absorb some of this, but hot-path allocation (`strings.Split(string(src), "\n")`) still occurs per function, allocating a full string copy and line slice for every function on every file. Confirmed by benchmark: 1,040,927 allocs/op for 211 files = **~4,900 allocs/file**, much of which originates here.  
  **Remediation**: Cache the file's line slice (or the raw bytes) in `FunctionAnalyzer` keyed by `token.File.Name()`. Populate once per unique file encountered in `AnalyzeFunctionsWithPath` and reuse across all functions in that file. Complexity: O(F reads) → O(1 read) per file. Validate with `go test -bench=BenchmarkFunctionAnalyzer -benchmem -count=5 ./internal/analyzer/` and `go build -gcflags='-m' ./internal/analyzer/`.

- [ ] **O(S²) sliding-window block extraction in duplication analysis** — `internal/analyzer/duplication.go:109-143` — `extractBlocksFromStmtList` applies a sliding window of all sizes from `minBlockLines` to `len(stmts)`. For a statement list of length S and minimum block size M, the number of generated blocks is `Σ(w=M to S) (S-w+1) = (S-M+1)*(S-M+2)/2`, which is O(S²). For a function body with 80 statements and M=6, that is `(75*76)/2 = 2,850` blocks — before recursive nested-block extraction doubles or triples the count. Benchmark confirms: `BenchmarkDuplicationAnalysis_LargeCodebase-4: 270ms, 78MB, 1.49M allocs` for this codebase alone. At 50,000 files with typical larger code, block count would be in the hundreds of millions, making duplication analysis a hard throughput blocker.  
  **Remediation**: Switch to a suffix-array or rolling-hash (Rabin-Karp) approach that produces O(N log N) candidate pairs from N total statements across all files, or cap the maximum window size (e.g. `maxWindowSize = min(len(stmts), minBlockLines*3)`) to bound growth. Additionally, early-exit when `CountNodes(blockStmts) > MaxDeepCopyNodes` to skip blocks that will be bucketed as `LARGE_BLOCK_*` anyway. Validate with `go test -bench=BenchmarkDuplicationAnalysis -benchmem -count=5 ./internal/analyzer/` before and after.

### HIGH

- [ ] **All parsed ASTs retained in memory for duplication analysis** — `pkg/generator/api_common.go:42-43,188` — `collectedMetrics.files` is a `map[string]*ast.File` populated in `processFile` (line 114) for every result. At `finalizeReport` (line 188), the entire map is passed to `AnalyzeDuplication`. This prevents the garbage collector from reclaiming any AST until the analysis is fully complete. For 50,000 files, each `*ast.File` carries its full node tree; at even 20 KB average per AST, that is 1 GB just for the AST map before any analysis results are accumulated. Projecting from the benchmark's 155 MB allocated for 211 files gives **~35 GB total allocations** for 50,000 files.  
  **Remediation**: Stream duplication analysis per-batch: extract statement blocks from each file's AST immediately after parsing (`ExtractBlocks`), accumulate only the `[]StatementBlock` slices (far smaller than full ASTs), then discard the AST. Fingerprinting and clone detection can then run on the accumulated blocks after all files are processed, without keeping any `*ast.File` alive. This decouples AST lifetime from analysis completion. Validate: `go test -bench=BenchmarkDuplicationAnalysis_LargeCodebase -memprofile=before.prof ./internal/analyzer/`; apply change; re-profile.

- [ ] **Two `os.ReadFile` calls per file during discovery** — `internal/scanner/discover.go:66` and `discover.go:111` — `analyzeFile` reads the full file contents to detect generation markers and parse the package clause (`parser.PackageClauseOnly`). `ParseFile` then reads the same file a second time for full AST parsing. Both reads are sequential on the same path. For 50,000 files, this is 100,000 redundant syscalls in the discovery/parse phase alone.  
  **Remediation**: Read the file once in `analyzeFile`, store `src []byte` in `FileInfo`, and pass it to `ParseFile` via `parser.ParseFile(fset, path, src, parser.ParseComments)` (the `src` parameter accepts `[]byte` or `string`). This eliminates the second disk read entirely. Verify with `go test -bench=BenchmarkFullAnalysis -benchmem ./cmd/`.

- [ ] **Shared `token.FileSet` serialises concurrent workers** — `internal/scanner/discovery.go:30`, `internal/scanner/worker.go:160` — A single `token.FileSet` is created in `NewDiscoverer` and shared across all worker goroutines that call `discoverer.ParseFile`. `token.FileSet.AddFile` acquires an internal mutex (`sync.Mutex`) on every new file added. With the default worker count set to `runtime.NumCPU()`, all goroutines contend on this single mutex during the parse phase, partially serialising what should be a fully parallel operation. The effect is most pronounced when I/O latency is low (e.g. NVMe + warm page cache) and the CPU-to-parse ratio is high.  
  **Remediation**: Give each worker its own `token.FileSet` (or a per-worker `FileSet` shard), then merge position information centrally after all workers complete. Alternatively use a `sync.Map`-based approach where each file's positions are stored independently. Benchmark the lock contention with `go test -bench=BenchmarkFullAnalysis -cpuprofile=cpu.prof ./cmd/` and `go tool pprof cpu.prof` to confirm the mutex shows up in the flame graph before optimising.

### MEDIUM

- [ ] **`regexp.MustCompile` compiled inside hot-path function** — `internal/analyzer/naming.go:259` — `toSnakeCase` calls `regexp.MustCompile(`_+`)` on every invocation. `toSnakeCase` is called from `checkSnakeCase` (naming.go:143) once per file with a naming violation during `AnalyzeFileNames`. For repositories with many non-snake-case files, this compiles the same regex repeatedly. `regexp.MustCompile` is documented as safe to call at package level.  
  **Remediation**: Extract to a package-level variable:
  ```go
  var multipleUnderscoreRE = regexp.MustCompile(`_+`)
  ```
  Replace the in-function call with `multipleUnderscoreRE.ReplaceAllString(resultStr, "_")`. Validate: `go build ./internal/analyzer/` (no functional change; zero-allocation improvement).

- [ ] **Linear searches through static slices in per-import hot path** — `internal/analyzer/package.go:358-388` — `isKnownInternalPrefix` iterates over a `[]string{...}` and `isKnownExternalPackage` iterates over another, both using `strings.HasPrefix`. Both are called for every import in every file. For a project with 500 files each importing 20 packages, that is 10,000 calls, each doing a 2- or 4-element linear search. Maps or a sorted-slice binary search would be O(1) amortised.  
  **Remediation**: Replace both slices with `map[string]struct{}` for prefix matching, or compile them as a single `strings.Replacer` / prefix tree. Alternatively, since the prefixes are short and fixed, the current linear search is O(1) at current scale but should be converted to avoid regression as the lists grow. Validate with `go test -bench=BenchmarkAnalyzeImportGraph -benchmem ./internal/analyzer/`.

- [ ] **`mergeUniqueStrings` allocates a new map on every call** — `internal/analyzer/package.go:392-411` — Called from `analyzePackageImports` per file, allocating a `map[string]bool` seen-set even when merging zero or one new import. For 50,000 files, 50,000 map allocations are created and immediately discarded.  
  **Remediation**: Replace with an append-then-deduplicate approach using a pre-sorted slice and `sort.SearchStrings`, which avoids heap allocation for the common small-list case, or accumulate imports in a `map[string][]string` keyed by package name from the start, eliminating the repeated merge entirely. Validate: `go test -bench=BenchmarkAnalyzeImportGraph -benchmem ./internal/analyzer/`.

- [ ] **Oversized channel pre-allocation** — `internal/scanner/worker.go:72-73` — `createChannels` allocates both `jobChan` and `resultChan` with capacity `fileCount`. For 50,000 files, this eagerly allocates ~800 KB of channel descriptor memory (two slices of 50,000 pointers) before any processing begins. The result channel in particular holds parsed `*ast.File` pointers, causing all parsed ASTs to remain live simultaneously.  
  **Remediation**: Cap channel buffer at `min(fileCount, workerCount*2)`. The workers can keep the pipeline full with a small buffer; the large pre-allocation provides no throughput benefit over a modest buffer size. Validate throughput parity with `go test -bench=BenchmarkFullAnalysis -benchmem ./cmd/`.

- [ ] **`BatchProcessor` result channel defeats its own batching purpose** — `internal/scanner/worker.go:257` — `ProcessInBatches` allocates `resultChan := make(chan Result, totalFiles)` buffered to the full file count, then the outer goroutine (`processBatchesAsync`) fully drains each batch before starting the next (line 282: `processSingleBatch` blocks until `collectBatchResults` finishes). This means all results can be buffered simultaneously, eliminating the memory-reduction benefit of batch processing. The design offers no advantage over a single `ProcessFiles` call for the typical codebase.  
  **Remediation**: Either (a) remove `BatchProcessor` and rely on a single bounded `ProcessFiles` with the channel-cap fix above, or (b) restructure batch processing to forward results immediately as they arrive rather than collecting them fully, and use a small output channel buffer. Verify with heap profiling before and after.

- [ ] **`filepath.Walk` instead of `filepath.WalkDir`** — `internal/scanner/discover.go:16`, `internal/analyzer/coverage.go:269`, `cmd/watch.go:170` — All three use the pre-Go-1.16 `filepath.Walk` API which calls `os.Lstat` then `os.Stat` for each file system entry, issuing two syscalls per entry. `filepath.WalkDir` (available since Go 1.16, and this project requires Go 1.24) provides `fs.DirEntry` with the `Lstat` result already available, halving filesystem syscalls for directory traversal.  
  **Remediation**: Replace `filepath.Walk(root, func(path string, info os.FileInfo, err error) error {...})` with `filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {...})` in all three locations. Adjust the function body to use `d.IsDir()` and `d.Type()` instead of `info.IsDir()` and `info.Mode()`. The `os.Stat` call in `analyzeFile` to retrieve `info.Size()` must also be updated to use `d.Info()` (which returns the cached stat). Validate: `go test ./internal/scanner/...`.

### LOW

- [ ] **`sortClaimedRanges` allocates and sorts on every containment check** — `internal/analyzer/duplication.go:742-754` — `isRangeSubsumed` calls `sortClaimedRanges` which creates a sorted copy of the claimed-ranges slice on every call. During `filterSubsumedClonePairs`, this is invoked once per instance per clone pair. For large codebases with many detected clones, this causes O(K log K) sorting work per containment check where K grows with clone count.  
  **Remediation**: Maintain `claimed[file]` as a sorted slice and perform insertion in sorted order when calling `claimPairRanges`. The containment check then uses `sort.Search` on an already-sorted slice with no copy. Validate correctness with existing `TestFilterSubsumedClonePairs` tests.

- [ ] **Missing benchmark for `readFileLines`** — `internal/analyzer/function.go` — Despite `readFileLines` being the single most I/O-intensive function in the hot path (one `os.ReadFile` + `strings.Split` per function), no benchmark exists to measure its cost or regression-detect future changes.  
  **Remediation**: Add `BenchmarkFunctionAnalyzer_ReadFileLines` and `BenchmarkFunctionAnalyzer_CountLinesInRange` to `function_test.go` using a synthetic or testdata file with a realistic number of functions. These benchmarks will immediately quantify the I/O amplification described in the CRITICAL finding above.

- [ ] **`OrganizationAnalyzer.countFileLines` re-reads files already parsed** — `internal/analyzer/organization.go:76` — Called during organisation analysis, this function reads the file from disk again with `os.ReadFile` to count lines, even though the same file was already read during discovery and full AST parsing. This adds a third or fourth `os.ReadFile` per file.  
  **Remediation**: Pass `LineMetrics` computed during function analysis (which already counts lines) into the organisation analyser, or compute line counts from the existing `*ast.File` using `fset.File(file.Pos()).LineCount()`. No re-read should be needed.

---

## False Positives Considered and Rejected

| Candidate Finding | Reason Rejected |
|-------------------|----------------|
| `sort.Slice` in `sortPackagesByName` (package.go:169) | Called once per report generation (cold path), not per file. O(P log P) for P packages. Not a hot-path concern. |
| `json.NewEncoder` recreated per section in JSON reporter (json.go:46, 104) | `json.NewEncoder` is lightweight; streaming to an `io.Writer` is correct practice. Allocation is not on the analysis hot path. |
| `ast.Inspect` used in multiple analyzers | Correct idiomatic Go AST traversal. The overhead is proportional to AST size, which is unavoidable. No O(n²) concern at the function level. |
| Linear search in `matchesIncludePatterns`/`matchesExcludePatterns` (file_filters.go) | Pattern lists in practice contain 1–5 entries. Linear search over ≤5 items is faster than a map due to cache locality. Not a measurable bottleneck. |
| `make([]metrics.FunctionMetrics, 0, len(...))` not used in `AnalyzeFunctionsWithPath` | Slice appends without pre-allocation add minor reallocation overhead but are not a hot-path bottleneck at per-file scale. LOW severity at best; rejected as a significant finding. |
| `detectCircularDependencies` uses DFS with maps (package.go:231-249) | Runs once at report generation time, not per file. Input size is number of packages (typically ≤100), making this O(P + E) — well within acceptable bounds. |
| `isGeneratedFile` scans up to 10 lines of each file (file_filters.go:127) | Correctly bounded. 10-line scan is O(1) relative to file size and runs only once per file during discovery. Not a hot path issue. |
| `BenchmarkDuplicationAnalysis_SimilarityComputation-4: 8µs/op` | Per-pair similarity (Jaccard on token sets) is fast and called only for detected duplicate groups, not all block pairs. Not a bottleneck. |
| `strings.Split` in `readFileLines` | Flagged in CRITICAL finding only for the per-function repetition. A single `strings.Split` per file is appropriate; the issue is the repeat, not the operation itself. |
| `sync.Pool` absence in `NormalizeBlock` | Block normalization allocates `bytes.Buffer` and AST copies per block, but these are short-lived and GC efficiently. `sync.Pool` would help only if normalization were on the absolute critical path; duplication analysis itself needs the O(S²) fix first. |
