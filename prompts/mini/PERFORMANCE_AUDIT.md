# TASK: Perform a performance audit identifying hot-path inefficiencies, algorithmic issues, unnecessary allocations, I/O bottlenecks, and scalability limiters.

## Output
Write two files in repo root: `AUDIT.md` (performance audit report) and `GAPS.md` (performance gaps).

## Workflow

### Phase 0: Performance Profile
1. Read README for performance claims (throughput, latency, scale).
2. Check `go.mod` for perf deps (sync, pooling, serialization).
3. Identify hot paths (per-request, per-item, tight loops) and cold paths (init, config).

### Phase 1: Baseline
```bash
go-stats-generator analyze . --skip-tests --format json --sections functions,patterns,packages,structs > tmp/perf-metrics.json
go test -bench=. -benchmem ./... 2>&1 | tee tmp/perf-bench.txt
```

### Phase 2: Audit Checklist
- [ ] **Algorithmic**: No O(n²+) on hot paths where O(n log n) or O(n) exists. No repeated sorting. Map lookups over linear search for repeated access.
- [ ] **Allocations**: Pre-allocate slices/maps with known size. `strings.Builder` for concat in loops. No `[]byte(s)`/`string(b)` in hot paths. Pool frequent objects via `sync.Pool`.
- [ ] **I/O**: Buffered I/O (`bufio`). Connection pooling (HTTP, DB). No unbounded `io.Copy` — enforce size limits.
- [ ] **Concurrency**: Worker pools over unbounded goroutines. Right-sized lock granularity. No lock in hot path where lock-free works.
- [ ] **Caching**: Cache expensive computations with invalidation strategy.

### Phase 3: Report
**AUDIT.md**: Performance profile, inventory table, benchmark summary, findings (CRITICAL/HIGH/MEDIUM/LOW) with file:line, evidence (complexity/alloc/IO), impact on targets, remediation.

**GAPS.md**: Per gap — Stated Goal, Current State, Bottleneck, Closing the Gap, Expected Improvement.

## Constraints
- Report ONLY — no code changes
- Every finding: file:line, measured evidence, actionable optimization
- No speculative optimizations as CRITICAL
- Use unchecked `- [ ]` checkboxes