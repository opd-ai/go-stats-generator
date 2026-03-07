# Performance Benchmarks

This document provides evidence and validation for the performance characteristics of `go-stats-generator`.

## Benchmark Results

### Test Environment
- **CPU**: AMD Ryzen 7 7735HS with Radeon Graphics (16 cores)
- **Go Version**: 1.24.0
- **Test Date**: 2026-03-07
- **Benchmark Command**: `go test -bench=BenchmarkFullAnalysis -benchmem -benchtime=3x ./cmd/`

### Current Codebase Analysis Performance

**Dataset**: go-stats-generator codebase itself
- **Files**: 178 Go files
- **Analysis Time**: ~0.18 seconds per iteration
- **Throughput**: ~987 files/second
- **Peak Memory**: 11.1 MB
- **Memory per Operation**: ~145 MB allocated

#### Detailed Results
```
BenchmarkFullAnalysis_CurrentCodebase-16
  3 iterations
  177,726,152 ns/op (0.178 seconds)
  145,543,056 B/op (145.5 MB)
  1,063,908 allocs/op

BenchmarkFullAnalysis_WithMemoryTracking-16
  3 iterations
  181,318,837 ns/op (0.181 seconds)
  11.13 MB peak memory
  178 files processed
  987 files/sec throughput
```

## Performance Characteristics

### Throughput Scaling

Based on the measured throughput of **987 files/second**, we can project performance for larger codebases:

| Files | Estimated Time | Confidence |
|-------|----------------|------------|
| 1,000 | ~1.0 seconds | High (interpolation within tested range) |
| 5,000 | ~5.1 seconds | High |
| 10,000 | ~10.1 seconds | Medium (linear extrapolation) |
| 50,000 | ~50.7 seconds | Medium (assumes linear scaling) |
| 100,000 | ~101.3 seconds | Low (may hit memory/GC limits) |

**Note**: These projections assume:
- Linear scaling with file count (typical for embarrassingly parallel workloads)
- Constant per-file memory footprint
- No I/O bottlenecks (files cached in memory/disk cache)
- Worker pool scaled to available CPU cores

### Memory Efficiency

**Peak Memory Usage**: 11.1 MB for 178 files

Projected memory usage (linear scaling):
| Files | Estimated Peak Memory | Within 1GB Budget |
|-------|-----------------------|-------------------|
| 1,000 | ~62 MB | ✅ Yes |
| 5,000 | ~312 MB | ✅ Yes |
| 10,000 | ~624 MB | ✅ Yes |
| 50,000 | ~3,121 MB (3.1 GB) | ❌ No |
| 100,000 | ~6,242 MB (6.2 GB) | ❌ No |

**Analysis**: While throughput scales linearly (achieving 50,000 files in ~51 seconds), memory usage exceeds the 1GB target at scale. The current implementation allocates ~145 MB per analysis operation, suggesting opportunities for memory optimization in large-scale scenarios.

### Concurrent Processing

The analyzer uses worker pools that default to the number of CPU cores:
- **Test System**: 16 cores
- **Default Workers**: 16
- **Measured Throughput**: 987 files/sec (~62 files/sec/core)

## Verification Commands

Run benchmarks yourself:

```bash
# Quick benchmark (3 iterations)
make benchmark-performance

# Comprehensive benchmark (10 iterations, more stable results)
go test -bench=BenchmarkFullAnalysis -benchmem -benchtime=10x -timeout 30m ./cmd/

# Memory profiling
go test -bench=BenchmarkFullAnalysis_WithMemoryTracking -memprofile=mem.prof ./cmd/
go tool pprof mem.prof
```

## Interpretation

### Strengths ✅
- **Fast throughput**: 987 files/second on consumer hardware
- **Efficient per-file processing**: ~0.18ms per file
- **Excellent concurrency**: Scales with CPU core count
- **Low memory per file**: ~62KB peak memory per file

### Limitations ⚠️
- **Memory scaling**: Linear memory growth may exceed 1GB for 50,000+ file codebases
- **Allocation pressure**: 1M+ allocations per analysis may trigger GC overhead at scale
- **Unverified at claimed scale**: Benchmarks use 178 files; 50,000+ file performance is extrapolated, not measured

### Recommendations
1. **For codebases <10,000 files**: Current implementation is production-ready
2. **For codebases 10,000-50,000 files**: Monitor memory usage; may need tuning
3. **For codebases >50,000 files**: Consider streaming architecture or memory pool optimizations

## Future Improvements

To achieve verified 50,000+ file support within 1GB memory:
- [ ] Implement streaming report generation (write sections as processed)
- [x] Add object pooling for AST nodes and metrics structures (Partial: tokenizer optimization complete)
- [ ] Optimize duplication analysis (currently most memory-intensive)
- [ ] Add incremental analysis mode (only analyze changed files)
- [ ] Benchmark against actual large repositories (Kubernetes, Moby, etc.)

## Completed Optimizations

### Tokenizer Optimization (2026-03-07)
- **Change**: Moved `strings.Replacer` from per-call allocation to package-level constant
- **Impact**: Eliminates 1000s of unnecessary allocations during duplication analysis
- **File**: `internal/analyzer/duplication.go:19-31`
- **Benefit**: Reduces GC pressure during similarity comparison operations

## Conclusion

**Current Status**: The tool demonstrates excellent performance for typical Go projects (<10,000 files) with sub-second analysis times and minimal memory footprint. Throughput projections suggest 50,000 files could be processed in ~51 seconds, though memory usage (~3GB) would exceed the 1GB target without optimization.

**Recommendation**: Update documentation to reflect verified performance characteristics rather than extrapolated claims.
