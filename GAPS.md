# Performance Gaps — 2026-04-20

This document identifies gaps between the stated performance goals of `go-stats-generator` and what its current implementation measurably achieves. All measurements were taken on a 4-core CI runner (Go 1.24) against the project's own codebase (211 files). Projections to 50,000 files assume linear scaling from benchmark data.

---

## Gap 1: Throughput Target for Large Codebases

- **Stated Goal**: Process 50,000+ files within 60 seconds (README: "Enterprise Scale: Designed for large codebases with concurrent processing"; PERFORMANCE.md: "~50.7 seconds" at 50K files)
- **Current State**: `BenchmarkFullAnalysis_WithMemoryTracking-4` measures 611 files/sec on 4 cores. Linear projection: 50,000 files ÷ 611 files/sec = **81.8 seconds** — 36% over the 60-second target. Even the claimed 987 files/sec on 16 cores projects to ~51 seconds, which sits at the very edge of the target with no headroom for larger functions, deeper nesting, or complex duplication patterns.
- **Bottleneck**: Two compounding factors push throughput below target: (1) Per-function `os.ReadFile` in `function.go:219` amplifies disk I/O by a factor proportional to average functions-per-file; (2) O(S²) sliding-window block extraction in `duplication.go:109` dominates CPU time for files with large function bodies. Duplication analysis alone consumed 270ms for the 211-file codebase — 73% of the 368ms total analysis time.
- **Closing the Gap**:
  1. Cache file content in `FunctionAnalyzer` to eliminate per-function re-reads (`function.go:219`). Expected reduction: ~20–40% of total analysis time at realistic function density.
  2. Replace the O(S²) sliding-window with a rolling-hash or suffix-array based clone detector, or cap max window size. Expected reduction: duplication analysis from O(N²) to O(N log N) — dominant gain for large repos.
  3. Eliminate the second `os.ReadFile` in `discover.go:111` by reusing the bytes already read in `analyzeFile`. Expected reduction: ~5–10% of total I/O time.
- **Expected Improvement**: Combined, these changes should achieve 2–4× throughput improvement, bringing 50,000-file analysis well within the 60-second target.

---

## Gap 2: Memory Target for Large Codebases

- **Stated Goal**: "Memory usage under 1GB" (README implied for enterprise-scale use; PERFORMANCE.md: "~3.1 GB projected for 50K files — ❌ No" — the project's own documentation acknowledges this gap)
- **Current State**: `BenchmarkFullAnalysis_CurrentCodebase-4` allocates 155,264,292 bytes (155 MB) for 211 files. Extrapolating: 50,000 files × (155 MB / 211 files) ≈ **36.7 GB total allocations**. Peak RSS is lower (17.31 MB for 211 files → ~4.1 GB projected peak RSS for 50,000 files) due to GC reclaiming memory between allocations, but still exceeds the 1 GB target by ~4×.
- **Bottleneck**: `collectedMetrics.files` in `api_common.go:42` retains all parsed `*ast.File` ASTs in memory simultaneously until `finalizeReport` completes. The O(S²) block list from duplication analysis adds a second large working set. Both prevent incremental GC of processed-file data.
- **Closing the Gap**:
  1. Stream duplication block extraction: replace `collected.files map[string]*ast.File` with `collected.blocks []StatementBlock`. Extract blocks immediately after each file is parsed and discard the AST. This reduces peak memory to the size of block slices (~2 KB/file) rather than full ASTs (~20–100 KB/file).
  2. Cap the sliding-window max size: limiting windows to `minBlockLines * 3` (e.g., 18 instead of the full function body) drastically reduces block count and block-list memory without meaningfully reducing clone detection quality for the most actionable clones.
  3. Apply the channel-buffer cap (AUDIT.md MEDIUM finding): replacing `make(chan Result, fileCount)` with `make(chan Result, workerCount*2)` allows GC to collect results as they flow through the pipeline rather than accumulating all at once.
- **Expected Improvement**: Streaming AST disposal should reduce peak RSS to near-linear with the number of files simultaneously in-flight (bounded by worker count), bringing 50K-file peak RSS to approximately `workerCount × avg_ast_size` ≈ `16 × 50 KB` = 800 KB in flight + analysis result storage (~50 MB for 50K file summaries). Total under 100 MB — well within the 1 GB target.

---

## Gap 3: Memory-per-File Claim

- **Stated Goal**: README states "~62 KB peak memory per file analyzed"
- **Current State**: `BenchmarkFullAnalysis_CurrentCodebase-4` allocates 155 MB for 211 files = **735 KB allocated per file** (11.8× the claimed figure). Peak RSS is ~82 KB/file, which is closer to the claim but still 32% higher.
- **Bottleneck**: The README figure appears to reflect only peak RSS per file at the operating system level (approximately 17.31 MB / 211 files ≈ 82 KB/file), but does not account for total heap allocations (735 KB/file) that drive GC pressure. The PERFORMANCE.md document correctly notes "145 MB allocated/op" but the README headline uses the smaller RSS metric.
- **Closing the Gap**:
  1. Correct the README to distinguish between "peak RSS" and "total heap allocated". The 62 KB figure is an underrepresentation.
  2. Reduce total heap allocations by implementing the per-function file-read cache and streaming duplication analysis (see Gaps 1 and 2). Both changes will bring allocated/file significantly closer to the RSS/file figure.
- **Expected Improvement**: After streaming AST disposal, total allocations should fall to ~100–150 KB/file (dominated by analysis result structs), closing the gap between the claimed and actual figures.

---

## Gap 4: Concurrency Scalability

- **Stated Goal**: "Concurrent Processing: Worker pools for analyzing large codebases efficiently" (README); default workers = `runtime.NumCPU()`
- **Current State**: All workers share a single `token.FileSet` via the `Discoverer` struct (`scanner/discovery.go:30`). `token.FileSet.AddFile` is mutex-protected internally, so each `ParseFile` call acquires a global lock while registering the file's position range. On a 16-core machine this serialises all concurrent file additions through a single `sync.Mutex`, capping parallel speedup below the theoretical 16×.
- **Bottleneck**: The shared `token.FileSet` mutex in `go/token` is a contended serialisation point during the parse phase. The effect is proportional to parse rate: at 987 files/sec on 16 cores, each core adds a file roughly every 16ms, but the `AddFile` critical section — while short — contributes non-trivial overhead when all 16 goroutines are simultaneously active.
- **Closing the Gap**:
  1. Assign a per-worker `token.FileSet` and use a `sync.Map` or post-processing merge to combine position data only for functions that require cross-file source extraction (primarily `readFileLines`).
  2. Alternatively, since `readFileLines` only needs the file's line count and content — not the `token.FileSet` position — it can bypass the `FileSet` entirely and read the file directly, removing the dependency on a shared mutable structure.
- **Expected Improvement**: Removing the `token.FileSet` contention should allow near-linear scaling with CPU cores, recovering potentially 20–30% of throughput on 16+ core machines (based on Amdahl's law with the mutex as the serial fraction).

---

## Gap 5: Duplication Analysis Does Not Scale to Enterprise Codebases

- **Stated Goal**: "Enterprise Scale: Designed for large codebases with concurrent processing" (README); duplication detection is a listed production-ready feature
- **Current State**: `BenchmarkDuplicationAnalysis_LargeCodebase-4` takes 270ms and allocates 78 MB for this ~200-file codebase. The O(S²) sliding-window in `extractBlocksFromStmtList` means duplication analysis time grows quadratically with function body size. For a 50,000-file codebase with functions 10× larger than this repo's average, duplication analysis alone could take tens of minutes and consume tens of gigabytes.
- **Bottleneck**: The algorithm generates every sub-window of every statement list at every nesting level, resulting in a block count of O(F × S²) where F = total functions and S = average statements per function. Fingerprinting, normalization, and grouping are then applied to this enormous block set.
- **Closing the Gap**:
  1. Switch to a rolling-hash (Rabin-Karp) or suffix-array approach over a serialised token sequence, producing O(N log N) candidate pairs from N total tokens.
  2. Alternatively, adopt a fixed-window-only approach: extract blocks only at the minimum size (e.g., exactly 6 statements), not all sizes from 6 to S. This reduces block count from O(S²) to O(S) per function while detecting the most actionable (longest) clones separately with a merge step.
  3. Apply a maximum window size cap immediately (e.g., `maxWindow = minBlockLines * 5`) to bound worst-case behaviour without requiring algorithmic redesign.
- **Expected Improvement**: O(S²) → O(S) or O(N log N) would reduce duplication analysis from a super-linear blocker to a linear or near-linear step, enabling enterprise-scale usage as advertised.

---

## Summary

| Gap | Stated Goal | Observed Gap | Severity | Primary Fix |
|-----|-------------|-------------|---------|-------------|
| Throughput | ≤60s for 50K files | ~82s projected | **Critical** | Cache file reads; fix O(S²) duplication |
| Memory | ≤1 GB peak | ~4.1 GB projected peak RSS | **Critical** | Stream AST disposal; cap channels |
| Memory-per-file claim | ~62 KB/file | ~735 KB allocated/file | **High** | Fix root causes above; update README |
| Concurrency scaling | Linear with CPU cores | Serialised by shared `token.FileSet` | **High** | Per-worker `FileSet` |
| Duplication at scale | Enterprise-scale feature | O(S²) blocks; not viable at 50K files | **Critical** | Rolling-hash / fixed-window algorithm |

---

*Generated 2026-04-20 as part of go-stats-generator performance audit*
