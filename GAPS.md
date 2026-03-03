# Implementation Gaps: Phase 7 — Composite Scoring & Actionable Output

Generated: 2026-03-03T09:54:00Z
Analysis Tool: go-stats-generator v1.0.0
Source: ROADMAP.md Phase 7, go-stats-generator metrics analysis

---

## Gap 1: Shotgun Surgery Detection Incomplete

- **Description**: ROADMAP Step 6.5 explicitly deferred shotgun surgery detection because it requires git history analysis via subprocess integration. This was noted as "deferred" and remains unimplemented.

- **Impact**: The suggestions generator (Step 7.2) cannot include recommendations to consolidate functions that are frequently changed together — a key indicator of maintenance burden that would improve refactoring prioritization.

- **Metrics Context**: 
  ```
  Current analysis scope: AST-based (go/ast, go/parser, go/token)
  Required for shotgun surgery: git subprocess (git log --name-only)
  Feature envy detection: Implemented (identifies methods with poor cohesion)
  ```
  Feature envy was implemented as a partial substitute but provides static analysis only, not temporal change patterns.

- **Resolution**: 
  1. Accept limitation for Phase 7 — proceed without shotgun surgery suggestions
  2. Future phase could integrate `git log` analysis to detect files frequently modified in the same commits
  3. Document this as a known limitation in CLI help text

---

## Gap 2: MBI Score Display Missing from Console/HTML/Markdown

- **Description**: ROADMAP Step 7.1 notes that MBI (Maintenance Burden Index) scores are computed and available in JSON output under `.scores.file_scores` and `.scores.package_scores`, but console/HTML/Markdown reporters do not render them. Additionally, the Step 7.1 notes state: "Console/HTML/Markdown output for MBI scores not yet implemented (deferred to Step 7.2)."

- **Impact**: Users running `go-stats-generator analyze` (default console output) cannot see MBI scores without switching to JSON format and using `jq`. This undermines the value of the MBI scoring feature for most users.

- **Metrics Context**: 
  ```bash
  $ cat metrics.json | jq '.scores'
  null
  ```
  The `.scores` field returns `null`, suggesting scores may not be fully integrated into the analyze command despite the ROADMAP claiming Step 7.1 is complete.

- **Resolution**:
  1. Step 1.4 of PLAN.md must include MBI score table in console output
  2. HTML reporter needs MBI heatmap or similar visualization
  3. Markdown reporter needs MBI summary table
  4. **Critical**: Verify `internal/analyzer/scoring.go` is wired into `cmd/analyze.go` — current analysis suggests it may not be

---

## Gap 3: Duplication Ratio Anomaly — Production vs Test Files

- **Description**: The duplication ratio of 33.42% (6182 duplicated lines, 135 clone pairs) is extremely high for a codebase of this maturity. Analysis with `--skip-tests` was used but `testdata/` directory may still be included. Test data files contain intentionally duplicated code for testing purposes.

- **Impact**: 
  - Effort estimates for deduplication suggestions may be significantly overestimated
  - CI/CD quality gates using `--max-duplication-ratio` may be unrealistic if based on inflated numbers
  - Top clones identified include files that may be test fixtures

- **Metrics Context**:
  ```json
  {
    "clone_pairs": 135,
    "duplicated_lines": 6182,
    "duplication_ratio": 0.3342,
    "largest_clone_size": 35
  }
  ```
  Top clone locations:
  - `internal/analyzer/interface.go:555-589` (35 lines)
  - `cmd/trend.go:106-140` (35 lines)
  - `internal/analyzer/naming.go:284-316` (33 lines, 3 instances)

- **Resolution**:
  1. Verify `--skip-tests` excludes `testdata/` directory (may need explicit `--exclude testdata`)
  2. Run analysis with `--exclude testdata` to get accurate production duplication ratio
  3. Adjust PLAN.md effort estimates if production ratio is significantly lower
  4. Document expected duplication in `testdata/` as intentional for test coverage

---

## Gap 4: Package Documentation Coverage Critically Low

- **Description**: Package-level documentation coverage is only 10.0% (2 of 20 packages), meaning most packages lack `doc.go` files or package-level comments. This is well below acceptable thresholds for a public tool.

- **Impact**: 
  - New contributors cannot understand package purposes from documentation
  - Step 3.3 (CI/CD documentation) will be less effective without proper package docs
  - The tool itself has poor discoverability for new users

- **Metrics Context**:
  ```json
  {
    "coverage": {
      "packages": 10.0,
      "functions": 68.5,
      "types": 58.1,
      "methods": 79.3,
      "overall": 66.6
    }
  }
  ```
  Package coverage (10.0%) is a significant outlier compared to other coverage metrics.

- **Resolution**:
  1. Add package documentation as a prerequisite step before Step 3.3
  2. Create `doc.go` files for core packages: `analyzer`, `reporter`, `storage`, `scanner`, `metrics`, `config`
  3. Target: raise package coverage from 10% to ≥50% before Phase 7 completion

---

## Gap 5: `.scores` JSON Field Returns Null

- **Description**: Running `cat metrics.json | jq '.scores'` returns `null`, suggesting MBI scoring is not integrated into the JSON output despite ROADMAP claiming Step 7.1 is complete.

- **Impact**: Steps 7.2-7.4 depend on MBI scores being available in the metrics output. If scoring isn't functional, the entire Phase 7 implementation plan requires revision.

- **Metrics Context**: 
  ROADMAP Step 7.1 states: "Integrated into JSON output with `.scores.file_scores` and `.scores.package_scores`"
  
  Actual output:
  ```bash
  $ cat metrics.json | jq '.scores'
  null
  ```

- **Resolution**:
  1. **Critical**: Verify Step 7.1 is actually complete by checking `internal/analyzer/scoring.go` implementation
  2. Verify `NewScoringAnalyzer()` is called in `cmd/analyze.go`
  3. If scoring analyzer exists but isn't wired into analyze command, this is a blocking issue
  4. If Step 7.1 is incomplete, it must be completed before Steps 7.2-7.4 can proceed

---

## Summary

| Gap | Severity | Phase 7 Impact | Resolution |
|-----|----------|----------------|------------|
| Shotgun Surgery | Moderate | Partial suggestions | Accept limitation |
| MBI Display Missing | High | Step 1.4 blocked | Include in Step 1.4 |
| Duplication Anomaly | Medium | Estimate accuracy | Exclude testdata/ |
| Package Doc Coverage | High | Step 3.3 quality | Add doc.go files |
| **`.scores` Returns Null** | **Critical** | **Blocks 7.2-7.4** | **Verify 7.1 complete** |

**Total Gaps**: 5
- **Critical** (blocks implementation): 1
- **High** (significant impact): 2
- **Medium** (accuracy/quality): 1
- **Moderate** (partial feature): 1

---

## Previous Phase 6 Gaps (Resolved)

The following gaps from Phase 6 have been resolved or deferred:

- ✅ **Burden Analyzer Integration**: Phase 6 Step 6.6 marked complete in ROADMAP
- ⏳ **Shotgun Surgery**: Explicitly deferred to future phase (see Gap 1 above)
- ⏳ **Annotation Age Tracking**: Deferred (requires git integration)
- ⚠️ **csv.go Complexity**: WriteDiff still at 24.9 complexity — included in PLAN.md Step 4
