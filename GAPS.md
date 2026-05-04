# Implementation Gaps — 2026-04-23

---

## Watch Mode Summary Is a Stub

- **Intended Behavior**: The `watch` command monitors Go files for changes, re-runs analysis after each debounced change event, and displays a compact summary of the updated metrics so developers get immediate feedback during coding sessions. The command's `Long` description (cmd/watch.go:17-28) and README team-workflow section both describe "live metrics updates during development."
- **Current State**: `printWatchSummary` at `cmd/watch.go:209` accepts a `*metrics.Report` (typed as `interface{}`) but ignores it entirely, printing only `"✓ Analysis complete at <time>"`. The full analysis is performed but zero metrics are surfaced to the user.
- **Blocked Goal**: Real-time code quality feedback during development — a stated differentiator of the watch command over one-shot analyze.
- **Implementation Path**: Replace `printWatchSummary` body with a compact renderer. The simplest correct approach:
  1. Change parameter type to `*metrics.Report` (or keep `interface{}` and type-assert).
  2. Print a 5-line summary: file count, total functions, avg complexity, MBI score, violations count using fields already populated on `report.Overview`, `report.Complexity`, and `report.Scores`.
  3. Alternatively, construct `reporter.NewConsoleReporter(reporter.Config{IncludeOverview: true, IncludeDetails: false})` and call `Generate(report, os.Stdout)`.
- **Dependencies**: None — all required data is already in the `*metrics.Report` returned by `runDirectoryAnalysis`.
- **Effort**: Small (< 1 hour)

---

## `cmd/serve` Hardcodes In-Memory Storage

- **Intended Behavior**: The API server documentation (`internal/api/handlers.go:27-29`) explicitly describes `NewServerWithStorage` as the production path: "Use this when analysis results must survive server restarts or be shared across multiple server instances." The `internal/api/storage` package provides a factory `New(cfg)` (factory.go:21) that selects memory, Postgres, or MongoDB based on `config.StorageConfig`. The `--config` file already supports `storage.type`, `storage.postgres_connection_string`, and `storage.mongo_connection_string` fields (config/analysis.go:191-199).
- **Current State**: `cmd/serve.go:44` calls `api.Run(serverPort, version)` which hardcodes `storage.NewMemory()` as the backend (api/server.go:21-26). All configured storage options are silently ignored. Every server restart loses all analysis results.
- **Blocked Goal**: Production deployment of the REST API with durable result storage. The feature is documented and the implementation exists; only the command-level wiring is missing.
- **Implementation Path**:
  1. In `runServe` (cmd/serve.go:34), call `loadConfiguration()` to get `*config.Config`.
  2. Call `apistorage.New(cfg)` to create the configured `ResultStore`.
  3. Replace `api.Run(serverPort, version)` with `api.RunWithStorage(serverPort, version, store)` — or inline the server construction: `server := api.NewServerWithStorage(version, store); api.RunServer(serverPort, server)`.
  4. Optionally add `--storage-backend` (memory/postgres/mongo) and `--storage-dsn` flags to `serveCmd` so storage can be configured without a config file.
- **Dependencies**: None — `NewServerWithStorage` and `storage.New` already exist.
- **Effort**: Small (1-2 hours)

---

## Team Metrics Not Rendered Outside JSON Output

- **Intended Behavior**: When `--enable-team-metrics` is passed, the analysis collects per-developer Git statistics (commit counts, lines added/removed, file ownership, active days) into `report.Team`. The README documents this feature with full detail (per-developer table, knowledge silo detection) and implies it is visible in all output formats.
- **Current State**: `finalizeTeamMetrics` (cmd/analyze_finalize.go:820) populates `report.Team` when the flag is set. In JSON output the `"team"` key is serialized via normal struct marshaling. However:
  - The console reporter's `writeReportSections` (internal/reporter/console.go:86) has no team entry.
  - The HTML template (`internal/reporter/templates/html/report.html`) has no team block.
  - The Markdown template (`internal/reporter/templates/markdown/report.md`) has no team block.
  - `CSVReporter.Generate` (internal/reporter/csv.go:27) has no team section.
  - `"team"` is absent from `ValidSections` (internal/metrics/sections.go:6), so `--sections team` silently does nothing and `--sections functions` cannot exclude team data.
- **Blocked Goal**: Team productivity analysis (fully implemented in `internal/analyzer/team.go`) cannot be presented to users who use any non-JSON output format.
- **Implementation Path**:
  1. Add `"team": true` to `ValidSections` (sections.go:~22) and add `"team": func(r *Report) { r.Team = nil }` to `sectionHandlers`.
  2. Add `(cr *ConsoleReporter) writeTeamAnalysis(output io.Writer, report *metrics.Report)` — guard with `report.Team != nil`, print developer count, list top 5 contributors by commit count, flag knowledge silos (developers with >40% exclusive file ownership).
  3. Add `(cr *ConsoleReporter) shouldWriteTeamAnalysis(report *metrics.Report) bool` returning `report.Team != nil`.
  4. Append `{cr.shouldWriteTeamAnalysis, cr.writeTeamAnalysis}` to the sections slice in `writeReportSections`.
  5. Add `{{if .Report.Team}}...{{end}}` block to both HTML and Markdown templates.
  6. Add a `writeTeamSection` CSV writer analogous to `writePackagesSection`.
- **Dependencies**: None — team data is already collected and available on the report.
- **Effort**: Medium (4-6 hours for all four renderers)

---

## `internal/multirepo` Package Not Wired to CLI

- **Intended Behavior**: The `multirepo` package doc (internal/multirepo/doc.go) describes "multi-repository analysis orchestration" enabling "cross-project comparisons, aggregated metrics, and organization-wide trend tracking." The ROADMAP.md references this capability. `multirepo.Analyzer.Analyze()` is fully implemented: it loops over configured repos, calls `generator.AnalyzeDirectory` for each, and returns a `*multirepo.Report` with per-repo results and errors.
- **Current State**: `internal/multirepo` is never imported outside its own package and tests. No `cmd/multirepo.go` exists. `main.go` and `cmd/root.go` have no reference to the package. The feature is entirely unreachable by end users.
- **Blocked Goal**: Cross-repository comparisons and organization-wide code health reporting.
- **Implementation Path**:
  1. Create `cmd/multirepo.go` with a `multirepoCmd` cobra command.
  2. Accept `--config <file>` flag pointing to a YAML file with `repositories: [{name: ..., path: ...}]` structure (matches `multirepo.Config`).
  3. Accept `--format` and `--output` flags matching analyze command conventions.
  4. In `RunE`, unmarshal config file into `multirepo.Config`, call `multirepo.NewAnalyzer(&cfg).Analyze()`, then format each `RepoResult` (skip errored repos with a warning, render successful ones).
  5. Register: `rootCmd.AddCommand(multirepoCmd)` in `init()`.
  6. Add basic test in `cmd/multirepo_test.go` verifying the command registers and the help text is correct.
- **Dependencies**: None — `multirepo.Analyzer` is complete.
- **Effort**: Medium (3-5 hours)

---

## Sentinel Errors in `pkg/generator` Never Returned

- **Intended Behavior**: `pkg/generator/errors_api.go` exports `ErrNoGoFiles`, `ErrInvalidDirectory`, `ErrParsingFailed`, and `ErrAnalysisFailed` as distinct sentinel errors to enable callers to use `errors.Is()` for typed error handling. The `errors_test.go` file tests these definitions but not their return from API methods.
- **Current State**: `AnalyzeDirectory` (api.go:18) returns raw errors from `filepath.Abs` and `scanner.NewDiscoverer`. `AnalyzeFile` (api.go:43) returns raw errors from `os.Stat` and the parser. None of the four sentinels are wrapped into any error chain in production code.
- **Blocked Goal**: Clean public API contract for `pkg/generator` library consumers who need to distinguish "no Go files found" from "directory doesn't exist" in automated tooling.
- **Implementation Path**:
  1. In `AnalyzeDirectory`, when `filepath.Abs` fails, return `fmt.Errorf("%w: %w", ErrInvalidDirectory, err)`.
  2. When the file list is empty after discovery, return `ErrNoGoFiles`.
  3. In `AnalyzeFile`, when `os.Stat` fails, return `fmt.Errorf("%w: %w", ErrInvalidDirectory, err)`.
  4. In `parseFileForAnalysis`, wrap parse errors: `fmt.Errorf("%w: %w", ErrParsingFailed, err)`.
  5. Add table-driven tests in `errors_test.go` asserting `errors.Is(err, ErrNoGoFiles)` for the appropriate scenarios.
- **Dependencies**: None.
- **Effort**: Small (1 hour)

---

## `CustomMetricsConfig` Defined but Never Referenced

- **Intended Behavior**: `internal/config/custom_metrics.go` defines `CustomMetricsConfig` and `DefaultCustomMetricsConfig()` for user-defined custom metric calculation — pattern matching, ratio metrics, and measurement aggregation. The code comment says "Custom metrics allow extending the tool with project-specific measurements."
- **Current State**: `CustomMetricsConfig` is never embedded in `config.Config` (analysis.go). `DefaultCustomMetricsConfig()` is never called by `config.DefaultConfig()` or any other function. No analyzer reads from a `CustomMetricsConfig`. The type is dead code.
- **Blocked Goal**: Per-project extensibility described in the type's documentation.
- **Implementation Path**: Either (a) embed `CustomMetrics CustomMetricsConfig` in `config.Config` and implement a `CustomMetricsAnalyzer` that evaluates the pattern/ratio/measurement definitions against the AST, or (b) if this feature is deferred, remove `custom_metrics.go` to avoid confusion. Option (a) is the intended path per the documentation. Minimum viable wiring: add the field to `Config`, call `DefaultCustomMetricsConfig()` from `DefaultConfig()`, and add a future-proof no-op check in `finalizeReport` that logs a warning when custom metrics are enabled but not yet implemented.
- **Dependencies**: None for the wiring; requires analyzer implementation for full feature.
- **Effort**: Small for wiring (30 min); Medium for full implementation (1-2 days)

---

## README "Planned Features" Incorrectly Labels Implemented Capabilities

- **Intended Behavior**: The README accurately reflects the project's implementation state so that users and contributors know what works today vs. what is planned.
- **Current State**: `README.md:1121-1122` lists "ARIMA/exponential smoothing for advanced time series forecasting (roadmap)" and "Correlation analysis between different metrics (roadmap)" under "Planned Features." Both are fully implemented:
  - ARIMA(1,1,1): `internal/analyzer/forecast.go:236` (`generateARIMAForecast`)
  - Exponential smoothing: `internal/analyzer/forecast.go:105` (`generateExponentialForecast`)
  - Correlation matrix: `internal/analyzer/statistics.go:245` (`ComputeCorrelationMatrix`), wired to `trend correlation` subcommand at `cmd/trend.go:876`
- **Blocked Goal**: Accurate documentation that enables users to discover and use implemented features.
- **Implementation Path**: Update `README.md:1121-1122` to mark both bullets as `✅` implemented, consistent with the linear regression bullet on line 1119. Update `README.md:41` which also lists ARIMA/exponential as "future enhancements."
- **Dependencies**: None.
- **Effort**: Trivial (5 minutes)
