# Implementation Gaps: Statistical Trend Analysis Phase

Generated: 2026-03-03T19:26:00Z
Analysis Tool: go-stats-generator v1.0.0
Source: ROADMAP.md, README.md "Planned Features", go-stats-generator metrics analysis

---

## Gap 1: Trend Forecast and Regressions Commands Are Placeholders

- **Description**: The `trend forecast` and `trend regressions` subcommands are explicitly marked as "PLACEHOLDER - implementation planned" in the CLI help text and source code (cmd/trend.go lines 54-77).

- **Impact**: Users cannot perform time-series forecasting or automated regression detection. This blocks the statistical analysis features promised in README.md under "Planned Features: Statistical Trend Analysis".

- **Metrics Context**: 
  ```
  cmd/trend.go contains 13 placeholder markers
  internal/analyzer/concurrency.go contains 1 placeholder marker
  internal/analyzer/duplication.go contains 4 placeholder markers
  Total placeholders: 18 across codebase
  ```
  The `trend` subcommands are the most user-visible incomplete features.

- **Resolution**: 
  1. Implement linear regression analysis in `internal/analyzer/statistics.go`
  2. Implement forecast generation in `internal/analyzer/forecast.go`
  3. Implement regression detection in `internal/analyzer/regression_detect.go`
  4. Update cmd/trend.go to call new analyzer functions
  5. Remove "PLACEHOLDER" warnings from help text

---

## Gap 2: Insufficient Historical Data for Statistical Analysis

- **Description**: Statistical trend analysis requires multiple historical metric snapshots stored via `baseline create`. New users or projects without baseline history cannot use forecast/regression features.

- **Impact**: First-time users running `trend forecast` will receive no useful output if they haven't established a baseline history. This creates a poor onboarding experience.

- **Metrics Context**:
  ```
  Required minimum data points: 7 (for regression)
  Current recommendation: "use diff command for production regression detection"
  Storage backend: SQLite (internal/storage/sqlite.go)
  ```

- **Resolution**:
  1. Add helpful error message when insufficient data points exist
  2. Document baseline workflow requirement in README.md
  3. Consider generating synthetic "cold start" baseline from current snapshot

---

## Gap 3: ARIMA/Exponential Smoothing Not Implemented

- **Description**: README.md "Planned Features" promises "ARIMA/exponential smoothing for time series forecasting" but this is beyond the scope of linear regression.

- **Impact**: Linear regression assumes a monotonic trend, which may not fit cyclical or seasonal patterns in real codebases. Forecasts may be inaccurate for complex metric histories.

- **Metrics Context**:
  ```
  README.md line 450: "ARIMA/exponential smoothing for time series forecasting"
  Current implementation: None (placeholder)
  Complexity: ARIMA requires parameter estimation (p,d,q), differencing, stationarity tests
  ```

- **Resolution**:
  1. Implement linear regression first (covers 80% of use cases)
  2. Add ARIMA/exponential smoothing in future phase
  3. Document limitations of linear regression in help text
  4. Accept this as a known limitation for the current phase

---

## Gap 4: High Duplication Ratio Affects Metric Accuracy

- **Description**: The duplication ratio of 49.23% is extremely high, likely inflated by intentional duplicates in `testdata/` directory. This affects effort estimates and may trigger CI/CD quality gate failures unexpectedly.

- **Impact**: 
  - Effort estimates for deduplication may be significantly overestimated
  - `--max-duplication-ratio` quality gates may fail on false positives
  - Developers may spend time investigating duplicates that are test fixtures

- **Metrics Context**:
  ```json
  {
    "clone_pairs": 116,
    "duplicated_lines": 9870,
    "duplication_ratio": 0.49231843575418993,
    "largest_clone_size": 36
  }
  ```
  Running with `--exclude testdata` would provide accurate production metrics.

- **Resolution**:
  1. Run analysis with `--exclude testdata` for production metrics
  2. Document expected duplication in `testdata/` as intentional
  3. Consider adding `testdata` to default exclusion list in config

---

## Gap 5: Package Documentation Coverage Low (30%)

- **Description**: Package-level documentation coverage is only 30% (6 of 20 packages have doc.go or package comments), well below the 70% target.

- **Impact**: 
  - New contributors cannot understand package purposes from documentation
  - godoc output is incomplete for the public API
  - Undermines the tool's credibility as a documentation analyzer

- **Metrics Context**:
  ```json
  {
    "coverage": {
      "packages": 30.0,
      "functions": 70.77,
      "types": 65.38,
      "methods": 82.91,
      "overall": 71.43
    }
  }
  ```
  Package coverage (30%) is a significant outlier compared to other metrics.

- **Resolution**:
  1. Create doc.go files for core packages: analyzer, reporter, storage, scanner, metrics, config
  2. Target: raise package coverage from 30% to ≥50%
  3. Include as a validation criterion for phase completion

---

## Previous Gaps (Resolved or Deferred)

### From Phase 7 Analysis (2026-03-03T09:54:00Z):

| Gap | Status | Resolution |
|-----|--------|------------|
| Shotgun Surgery Detection | Deferred | Requires git history subprocess integration |
| MBI Display in Console/HTML | Deferred | Included in PLAN.md Step 7 |
| `.scores` Returns Null | Resolved | Scores now populated in metrics.json |
| Annotation Age Tracking | Deferred | Requires git integration |

---

## Summary

| Gap | Severity | Impact | Resolution Status |
|-----|----------|--------|-------------------|
| **Placeholder Trend Commands** | **Critical** | Blocks statistical features | PLAN.md Steps 1-5 |
| **Insufficient Historical Data** | High | Poor onboarding | PLAN.md Step 8 |
| **ARIMA Not Implemented** | Medium | Limited forecasting | Deferred to future phase |
| **High Duplication Ratio** | Medium | Inflated metrics | Use --exclude testdata |
| **Low Package Doc Coverage** | Medium | Poor discoverability | PLAN.md validation criteria |

**Total Gaps**: 5
- **Critical** (blocks implementation): 1
- **High** (significant impact): 1  
- **Medium** (accuracy/quality): 3
