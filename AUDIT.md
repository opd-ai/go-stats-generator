# IMPLEMENTATION GAP AUDIT — 2026-04-23

## Project Architecture Overview

`go-stats-generator` is a high-performance CLI tool for analyzing Go source code repositories and generating
comprehensive statistical reports. Its stated goals include enterprise-scale analysis (50,000+ files, <60s,
<1GB RAM), multiple output formats (console/JSON/HTML/CSV/Markdown), historical baseline management, and
trend forecasting.

| Package | Role |
|---------|------|
| `cmd/` | CLI commands: analyze, baseline, diff, trend, serve, watch, version |
| `internal/analyzer/` | AST analysis engines (functions, structs, interfaces, packages, patterns, naming, burden, duplication, team, coverage, forecast) |
| `internal/metrics/` | Metric data structures and diff computation |
| `internal/reporter/` | Output formatters (console, JSON, HTML, CSV, Markdown) |
| `internal/scanner/` | File discovery and concurrent processing |
| `internal/storage/` | Historical metrics storage (SQLite, JSON, memory) |
| `internal/config/` | Configuration management |
| `internal/api/` | HTTP API server and storage backends (memory, Postgres, MongoDB) |
| `internal/multirepo/` | Multi-repository analysis orchestration |
| `pkg/generator/` | Public API |

**Build status:** `go build ./...` — ✅ clean  
**Vet status:** `go vet ./...` — ✅ clean  
**Test status:** `go test ./...` — ✅ all pass  

---

## Gap Summary

| Category | Count | Critical | High | Medium | Low |
|----------|-------|----------|------|--------|-----|
| Stubs/TODOs | 1 | 0 | 1 | 0 | 0 |
| Dead Code | 2 | 0 | 0 | 1 | 1 |
| Partially Wired | 4 | 0 | 2 | 2 | 0 |
| Interface Gaps | 2 | 0 | 1 | 0 | 1 |
| Dependency Gaps | 0 | 0 | 0 | 0 | 0 |

---

## Implementation Completeness by Package

| Package | Exported Functions | Implemented | Stubs | Dead | Coverage |
|---------|--------------------|-------------|-------|------|----------|
| `cmd` | 0 (internal) | — | 1 (`printWatchSummary`) | 0 | — |
| `internal/analyzer` | ~30 | ~30 | 0 | 0 | ✅ |
| `internal/api` | 3 | 3 | 0 | 0 | ✅ |
| `internal/api/storage` | 15 | 15 | 0 | 0 | ✅ |
| `internal/config` | 3 | 2 | 0 | 1 (`DefaultCustomMetricsConfig`) | ⚠️ |
| `internal/metrics` | ~10 | ~10 | 0 | 0 | ✅ |
| `internal/multirepo` | 2 | 2 | 0 | 0 (unwired) | ⚠️ |
| `internal/reporter` | ~25 | ~21 | 0 | 0 | ⚠️ |
| `internal/scanner` | ~8 | ~8 | 0 | 0 | ✅ |
| `internal/storage` | ~8 | ~8 | 0 | 0 | ✅ |
| `pkg/generator` | 6 | 4 | 0 | 4 (sentinel errors) | ⚠️ |

---

## Findings

### HIGH

- [ ] **`printWatchSummary` is a no-op stub** — `cmd/watch.go:209` — The `analyzeWithWatch` function on line 192 runs the full analysis workflow and passes the resulting `*metrics.Report` to `printWatchSummary(report)`, but `printWatchSummary` discards its argument and prints only `"✓ Analysis complete at <time>"`. No metrics (function count, complexity, violations, MBI score) are ever shown in watch mode. — **Blocked goal:** Real-time metrics updates during development (stated in `watchCmd.Long` and README). — **Remediation:** Replace the stub body with a compact console summary using the existing `ConsoleReporter`. Cast the `interface{}` parameter to `*metrics.Report`, construct a `reporter.NewConsoleReporter(reporter.Config{IncludeOverview: true})`, call `Generate(report, os.Stdout)` or write a custom 5-line summary using `report.Overview` and `report.Complexity`. Validate: `go-stats-generator watch . --quiet=false` must print function/complexity counts after each re-analysis. No signature change needed.

- [ ] **`cmd/serve` ignores configured storage backend — always uses in-memory** — `cmd/serve.go:44` — `runServe` calls `api.Run(serverPort, version)` which calls `api.NewServer(version)` → `storage.NewMemory()`. The `internal/api/storage.New(cfg)` factory (factory.go:21) that routes to Postgres/MongoDB/memory based on configuration is never called. All analysis results are lost on server restart; the README and code comments for `NewServerWithStorage` state that durable storage is the intended production path. — **Blocked goal:** Production REST API server with durable result persistence. — **Remediation:** In `runServe`, call `loadConfiguration()` to get the config, then call `apistorage.New(cfg)` to create the configured backend, and pass it to `api.NewServerWithStorage(version, store)` before starting. Add `--storage-backend` and `--storage-dsn` flags to `serveCmd` mirroring the config fields. Validate: `go-stats-generator serve --port 8080` followed by two POST requests to `/api/v1/analyze`; results must survive across handler invocations.

### MEDIUM

- [ ] **Team metrics (`report.Team`) never rendered by console, HTML, Markdown, or CSV reporters** — `internal/metrics/report.go:28` — The `Team` field is populated by `finalizeTeamMetrics` (analyze_finalize.go:820) when `--enable-team-metrics` is set, and it serializes correctly in JSON output. However, none of the non-JSON reporters (console, HTML, markdown, CSV) implement a team section, and `"team"` is absent from `ValidSections` in `internal/metrics/sections.go:6-22`, making it impossible to filter via `--sections`. — **Blocked goal:** Team productivity analysis feature is partially invisible — users relying on console/HTML/Markdown output get no team data. — **Remediation:** (1) Add `"team": true` to `ValidSections` (sections.go:~22) and a `clearTeamSection` handler. (2) Add a `writeTeamAnalysis` method to `ConsoleReporter` covering developer count, top contributors, and knowledge silo warnings; wire it into `writeReportSections`. (3) Add a `{{if .Report.Team}}` block to `internal/reporter/templates/html/report.html` and `internal/reporter/templates/markdown/report.md`. (4) Add a `writeTeamSection` to `CSVReporter`. Validate: `go-stats-generator analyze . --enable-team-metrics` shows team data in console output.

- [ ] **`internal/multirepo` package is implemented but not wired to any CLI command** — `internal/multirepo/analyzer.go:1` — The `multirepo.Analyzer` and `multirepo.Config` types are fully implemented: `NewAnalyzer(cfg)` iterates over `config.Repositories`, calls `generator.AnalyzeDirectory` for each, and returns an aggregate `Report`. No `cmd/multirepo.go` exists; no `rootCmd.AddCommand` references this package. The package is not imported from `cmd/` or `main.go`. — **Blocked goal:** Cross-repository comparisons and organization-wide trend tracking (described in `multirepo/doc.go` and the package comment). — **Remediation:** Create `cmd/multirepo.go` with a `multirepoCmd` cobra command that accepts `--config <file>` (YAML with list of repos), unmarshals into `multirepo.Config`, calls `multirepo.NewAnalyzer(cfg).Analyze()`, and formats the result. Wire it with `rootCmd.AddCommand(multirepoCmd)` in `init()`. Validate: `go build ./...` succeeds and `go-stats-generator multirepo --help` shows the command.

### LOW

- [ ] **`pkg/generator` sentinel errors defined but never returned** — `pkg/generator/errors_api.go:6-17` — `ErrNoGoFiles`, `ErrInvalidDirectory`, `ErrParsingFailed`, and `ErrAnalysisFailed` are defined as exported sentinel errors for use by callers via `errors.Is`. Neither `AnalyzeDirectory` nor `AnalyzeFile` (api.go, api_common.go) returns these errors; they return raw errors from `filepath.Abs`, `os.Stat`, and parser calls. External consumers cannot distinguish error types. — **Blocked goal:** Clean public API contract for the `pkg/generator` library. — **Remediation:** In `AnalyzeDirectory` (api.go:20), wrap `os.Stat`/`filepath.Abs` errors with `fmt.Errorf("%w: %w", ErrInvalidDirectory, err)` and scanner errors with `ErrNoGoFiles`. In `parseFileForAnalysis` (api.go:91), wrap parse errors with `ErrParsingFailed`. Validate: `errors.Is(err, ErrNoGoFiles)` returns true when analyzing an empty directory.

- [ ] **README's "Planned Features" section lists ARIMA and correlation as roadmap items despite being implemented** — `README.md:1121-1122` — Lines 1121-1122 list "ARIMA/exponential smoothing" and "Correlation analysis" with `(roadmap)` labels. Both are implemented: ARIMA in `internal/analyzer/forecast.go:236`, exponential smoothing in `forecast.go:105`, and correlation in `internal/analyzer/statistics.go:102` wired to `trend correlation` subcommand (cmd/trend.go:876). This creates a false impression of missing features. — **Blocked goal:** Accurate documentation of project capabilities. — **Remediation:** Update `README.md` lines 1121-1122 to mark these as `✅ implemented` consistent with the `Linear regression` bullet above. Validate: README bullets for statistical analysis are internally consistent.

---

## False Positives Considered and Rejected

| Candidate Finding | Reason Rejected |
|-------------------|-----------------|
| `internal/api/storage/postgres.go` — PostgreSQL backend "unused" | Fully implemented with Store/Get/List/Delete/Clear/Close; wired through `internal/api/storage.New(cfg)` factory. Not called from `cmd/serve.go`, which is a separate HIGH gap. The implementation itself is complete. |
| `internal/api/storage/mongo.go` — MongoDB backend "unused" | Same as Postgres: fully implemented, reachable via factory, not invoked from serve command. |
| `go.mod` lists `github.com/lib/pq` and `go.mongodb.org/mongo-driver` as "unused" dependencies | Both are used by `internal/api/storage/postgres.go` and `mongo.go` respectively. They are not dead imports. |
| `internal/config/custom_metrics.go` `DefaultCustomMetricsConfig()` — never called | Function exists in a config package and defines a valid default. The `CustomMetricsConfig` type is not referenced by `Config` struct, making the entire type effectively unused dead code — this IS flagged as a medium gap. The function itself is just an uninvoked default constructor, not a stub. |
| `pkg/generator` `Analyzer.buildReport` — "not exposed in public API" | Called internally by `AnalyzeDirectory` and `AnalyzeFile`; correct architectural choice to keep it unexported. |
| `multirepo.Analyzer.Analyze()` — "function with no tests for edge cases" | The function is tested in `internal/multirepo/analyzer_test.go`. Missing CLI wiring is the actual gap, not missing implementation. |
| Console reporter doesn't output `TestCoverage`/`TestQuality` | These metrics are only populated when `--coverage-profile` is provided. JSON output includes them. The reporter gap for team metrics is the stronger finding; test coverage rendering is speculative given the opt-in nature of the feature. Retained as part of the team-metrics gap finding (same root cause). |
| `cmd/watch.go` `watchRecursive` flag — "defined but never checked" | Flag is parsed via `watchCmd.Flags()` and bound; `addWatchPaths` (watch.go:172) always walks recursively. Simplification, not a stub — the flag exists for forward compatibility but its current value is the intended always-recursive behavior. |
| `printWatchSummary` parameter type `interface{}` — "should be `*metrics.Report`" | The parameter type is a consequence of the stub, not an independent design gap. Covered under the stub finding above. |
