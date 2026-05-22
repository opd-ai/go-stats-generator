# Implementation Gaps — 2026-05-22

## Gap 1: REST API Lacks Authentication and Authorization

- **Stated Goal**: README documents a "REST API Server" mode (`serve` command) for remote analysis.
- **Current State**: The API endpoints have no authentication mechanism (no API keys, no tokens, no OAuth). Anyone with network access can trigger analysis, read results, and access the filesystem.
- **Impact**: In any non-localhost deployment, the API is an open filesystem scanner. Combined with finding #3 (path traversal), this allows unauthenticated directory enumeration and source code reading of any Go file on the server.
- **Closing the Gap**: Add at minimum an API key middleware (configurable via environment variable). Document that the server should only be bound to localhost unless authentication is configured.

## Gap 2: REST API Has No Rate Limiting or Resource Bounding

- **Stated Goal**: The API is intended to service analysis requests.
- **Current State**: No rate limiting, no request size limits, no concurrent analysis cap. `HandleAnalyze` spawns unbounded goroutines (finding #9).
- **Impact**: A single client can exhaust server memory/CPU by submitting many analysis requests for large directories simultaneously. Trivial denial of service.
- **Closing the Gap**: Add a semaphore (bounded channel) limiting concurrent analyses; add request body size limits; add per-IP rate limiting middleware.

## Gap 3: Trend Analysis Display Functions Panic on Type Mismatch

- **Stated Goal**: README advertises "Trend Analysis" with regression detection and visual display.
- **Current State**: `displayFloatTrend`, `displayIntTrend`, and other display functions in `cmd/trend.go` use bare type assertions (`data["start"].(float64)`) without comma-ok form (finding #10). If trend data contains unexpected types (e.g., `int` instead of `float64` from JSON decode edge cases), these functions panic.
- **Impact**: Users running `go-stats-generator trend` may get unexpected crashes instead of helpful error messages. The feature is unreliable for production use.
- **Closing the Gap**: Convert all bare type assertions to comma-ok form with graceful fallbacks.

## Gap 4: Storage Backend Inconsistency — Foreign Key Enforcement

- **Stated Goal**: README describes "SQLite Database" storage with "built-in" reliability. The schema defines foreign key relationships.
- **Current State**: `PRAGMA foreign_keys=ON` is per-connection in SQLite, but `MaxOpenConns` allows multiple connections. Only ~1/N connections have FK enforcement (finding #4). Orphaned metrics records can accumulate.
- **Impact**: Data integrity guarantees are illusory. Deleting a snapshot may leave orphaned records; cascading deletes don't fire for most connections.
- **Closing the Gap**: Set `MaxOpenConns(1)` (standard for SQLite) or use the `_pragma=foreign_keys(1)` DSN parameter that applies to every connection.

## Gap 5: Gzip Compressed Storage Silently Corrupts Data

- **Stated Goal**: JSON storage supports "configurable compression" for efficient storage.
- **Current State**: `writeCompressed` (finding #1) defers `gzWriter.Close()`, discarding the error. Gzip's `Close()` performs the final flush and writes the footer. If it fails, the file is truncated but `Store()` returns nil.
- **Impact**: Users who enable compression may silently lose data. `Retrieve()` will fail with an opaque "failed to decompress" error later, after the original data is lost.
- **Closing the Gap**: Handle `gzWriter.Close()` error explicitly and propagate it.

## Gap 6: Missing `rows.Err()` Checks in All SQL Backends

- **Stated Goal**: Reliable metric storage and retrieval across backends.
- **Current State**: After every `for rows.Next()` loop in SQLite and PostgreSQL storage, `rows.Err()` is never checked (finding #13). If the database encounters an error mid-iteration (disk full, network drop, corruption), the code silently returns partial results.
- **Impact**: Users may see truncated metric histories with no error indication. Particularly dangerous for trend analysis which makes decisions based on historical data completeness.
- **Closing the Gap**: Add `if err := rows.Err(); err != nil` after each iteration loop.

## Gap 7: Nesting Depth Metric Is Inflated

- **Stated Goal**: "Struct Complexity Metrics" including "nesting depth analysis" are advertised.
- **Current State**: `calculateNestingDepth` in `struct.go` (finding #19) never decrements `currentDepth` when exiting a nesting scope. Sequential (non-nested) control structures are reported as nested, inflating the metric.
- **Impact**: Users see exaggerated nesting depth values, potentially causing false alarms in CI quality gates.
- **Closing the Gap**: Track enter/exit of nested scopes properly (decrement on `node == nil` callback from `ast.Inspect`).

## Gap 8: Builder Pattern Detection Produces Garbage Output

- **Stated Goal**: "Advanced Pattern Detection" including builder patterns.
- **Current State**: `pattern.go:289` uses `string(rune(candidate.setterCount))` which converts an integer to a Unicode codepoint character, not a decimal string (finding #14). A setter count of 65 becomes "A", count of 5 becomes an invisible control character.
- **Impact**: Pattern detection output for builder patterns contains garbage characters instead of meaningful setter counts. The feature is effectively broken for reporting.
- **Closing the Gap**: Use `strconv.Itoa(candidate.setterCount)` instead.

## Gap 9: Multi-Repository Analysis Has No Parallel Execution

- **Stated Goal**: README mentions "Multi-Repository Analysis" capabilities.
- **Current State**: `internal/multirepo/` processes repositories sequentially in a loop. Despite the project's concurrent processing capabilities for single repos, multi-repo mode does not parallelize across repositories.
- **Impact**: Multi-repo analysis of N repos takes N× single-repo time instead of benefiting from parallelism.
- **Closing the Gap**: Add a worker pool (project already uses this pattern in `internal/scanner/worker.go`) to analyze multiple repos concurrently.

## Gap 10: Documentation Coverage Claims vs. Actual Metric Accuracy

- **Stated Goal**: "Documentation Coverage Analysis" to accurately assess code documentation quality.
- **Current State**: The organization analyzer (finding #18) uses `strings.Contains(line, "//")` to detect comments, which matches URLs in string literals (e.g., `"https://example.com"`). This inflates comment counts and doc coverage percentages.
- **Impact**: Documentation coverage metrics are unreliable for code that contains URLs or paths with `//` in string literals. Reports may significantly overstate documentation quality.
- **Closing the Gap**: Use proper Go AST comment extraction rather than line-based string matching.
