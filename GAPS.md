# Implementation Gaps: Phase 6 — Additional Maintenance Burden Indicators

Generated: 2026-03-03T05:37:00Z
Analysis Tool: go-stats-generator v1.0.0
Source: ROADMAP.md Phase 6

---

## Gap 1: Shotgun Surgery Detection Requires Git Integration

- **Description**: ROADMAP.md Step 6.5 specifies shotgun surgery detection via `git log --name-only` parsing to identify files that frequently co-change (>60% of same commits). This requires subprocess execution and git history analysis that is outside the current AST-based analysis scope.

- **Impact**: Phase 6 cannot be fully completed without this feature. Step 6.5 will be partially implemented (feature envy only). The `BurdenMetrics.ShotgunSurgeryClusters` field will remain empty until git integration is added.

- **Metrics Context**: 
  ```
  Current analysis scope: AST-based (go/ast, go/parser, go/token)
  Required for shotgun surgery: git subprocess (git log --name-only)
  Complexity: Requires parsing commit metadata and co-change frequency analysis
  ```

- **Resolution**: Defer shotgun surgery detection to Phase 7 (Composite Scoring & Actionable Output), which already includes "Baseline Integration & Trend Tracking" (Step 7.3) that requires git history access. Implementing git integration once for both features reduces duplication.

---

## Gap 2: Annotation Age Tracking Incomplete (from Phase 4)

- **Description**: ROADMAP.md Phase 4 Step 4.3 notes "Track annotation age via git history (deferred - subprocess integration complexity)". The `documentation.stale_annotations` field currently returns 0 regardless of actual annotation age.

- **Impact**: The `go-stats-generator` output shows annotations (TODO, FIXME, HACK, BUG, XXX, DEPRECATED, NOTE) but cannot determine which are stale based on commit timestamps. Users must manually review annotation age.

- **Metrics Context**:
  ```json
  {
    "stale_annotations": 0,
    "annotations_by_category": {
      "BUG": 1, "DEPRECATED": 1, "FIXME": 1, 
      "HACK": 1, "NOTE": 1, "TODO": 1, "XXX": 1
    }
  }
  ```
  The `stale_annotation_days` configuration (default: 180) has no effect without git integration.

- **Resolution**: Address alongside Gap 1 in Phase 7. Git integration for shotgun surgery can also enable annotation age tracking via `git blame` or `git log -p`.

---

## Gap 3: testdata/ Complexity Inflates Codebase Metrics

- **Description**: Test data files in `testdata/` directory contain intentionally complex code for testing purposes. These inflate overall codebase complexity metrics and may cause confusion when validating Phase 6 implementation.

- **Impact**: 
  - `testdata/simple/calculator.go:VeryComplexFunction` has complexity 34.7 (2nd highest in codebase)
  - `testdata/simple/user.go:ComplexFunction` has complexity 21.5
  - 2 of 8 functions above complexity 20.0 are in testdata/
  - Validation criteria comparing "functions above complexity threshold" will include testdata

- **Metrics Context**:
  ```json
  {"name": "VeryComplexFunction", "file": "testdata/simple/calculator.go", "complexity": 34.7}
  {"name": "ComplexFunction", "file": "testdata/simple/user.go", "complexity": 21.5}
  ```

- **Resolution**: 
  1. Use `--exclude testdata/` when running validation analysis
  2. Update PLAN.md validation criteria to explicitly exclude testdata
  3. Consider adding `--exclude-testdata` convenience flag in future release

---

## Gap 4: High Duplication Ratio May Complicate New Development

- **Description**: Current duplication ratio is 37.04% (135 clone pairs, 6182 duplicated lines), significantly exceeding the "Large" scope threshold of 8%. This indicates substantial shared patterns that should be extracted before adding new analyzer code.

- **Impact**: Adding `internal/analyzer/burden.go` without first extracting shared patterns will likely increase duplication further. The top clone pairs are in `internal/analyzer/` and `cmd/` packages:
  ```
  - internal/analyzer/interface.go:555-589 (35 lines, renamed clone)
  - cmd/trend.go:102-136 (35 lines, renamed clone)
  - internal/analyzer/naming.go:284-316 (33 lines, 3 instances)
  ```

- **Metrics Context**:
  ```json
  {
    "clone_pairs": 135,
    "duplicated_lines": 6182,
    "duplication_ratio": 0.3704,
    "largest_clone_size": 35
  }
  ```

- **Resolution**: PLAN.md includes "Prerequisite 3: Address Duplication" to extract common AST traversal patterns into `internal/analyzer/astutil.go` before implementing Phase 6 features. This should reduce duplication ratio to <15% before new development begins.

---

## Gap 5: `internal/reporter/csv.go:Generate` Blocks Safe Extension

- **Description**: The `Generate` function in csv.go has complexity 65.7 (highest in codebase), with 296 lines and 49 cyclomatic paths. Adding burden metrics reporting to this function would further increase complexity and make the reporter unmaintainable.

- **Impact**: Phase 6 Step 9 requires integrating BurdenMetrics into all output formats including CSV. The current `Generate` function cannot be safely extended.

- **Metrics Context**:
  ```json
  {"name": "Generate", "file": "internal/reporter/csv.go", "complexity": 65.7, "lines": 296}
  ```
  This single function accounts for the highest complexity hotspot and must be refactored as a prerequisite.

- **Resolution**: PLAN.md includes "Prerequisite 1: Refactor csv.go:Generate" to extract reporting sections into dedicated helpers (target complexity ≤15.0). Refactoring must complete before adding BurdenMetrics output.

---

## Summary

| Gap | Severity | Phase 6 Impact | Resolution |
|-----|----------|----------------|------------|
| Shotgun Surgery | Moderate | Partial Step 6.5 | Defer to Phase 7 |
| Annotation Age | Minor | Documentation only | Defer to Phase 7 |
| testdata/ Inflation | Minor | Validation accuracy | Use --exclude testdata/ |
| High Duplication | High | New code quality | Extract astutil.go first |
| csv.go Complexity | Critical | Reporter integration | Refactor first |

**Total Gaps**: 5
- **Critical** (blocks implementation): 1
- **Moderate** (partial feature): 1
- **Minor** (workaround available): 3
