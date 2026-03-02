# Maintenance Burden Detection — Roadmap

This roadmap describes a step-by-step plan for adding maintenance burden detection to `go-stats-generator`. Each phase builds on the project's existing AST-based analysis infrastructure (function, struct, interface, package, and concurrency analyzers) and introduces new analyzers that surface concrete, actionable sources of maintenance cost in Go codebases.

---

## Goals

1. **Detect code duplication** — find near-identical blocks that increase change cost.
2. **Flag poorly named identifiers** — files, functions, methods, variables, and constants that violate Go naming conventions or are ambiguous.
3. **Identify misplaced declarations** — functions, methods, and types that live in the wrong file or package relative to their usage.
4. **Surface missing or inadequate comments** — exported symbols without GoDoc, packages without doc.go, and stale TODOs.
5. **Measure organizational problems** — too-large files, too-large packages, deep directory nesting, and incoherent package groupings.
6. **Report additional burden indicators** — magic numbers, dead code, excessive parameter lists, deeply nested logic, shotgun surgery patterns, and feature envy.

---

## Phase 1: Code Duplication Detection

Detect duplicated or near-duplicate code blocks that force developers to make the same change in multiple places.

### Step 1.1 — AST-Based Block Fingerprinting

- Walk every function and method body to extract statement-level sub-trees.
- Normalize each sub-tree (strip identifiers, literals, comments) and compute a structural hash.
- Store `(hash, file, startLine, endLine, nodeCount)` tuples for later comparison.

### Step 1.2 — Clone Pair Detection

- Group fingerprints by hash; groups with two or more entries are clone pairs.
- Classify clones:
  - **Type 1** — exact duplicates (identical after whitespace normalization).
  - **Type 2** — renamed duplicates (identical structure, different identifier names).
  - **Type 3** — near duplicates (structural similarity above a configurable threshold, e.g. ≥80%).
- Report clone set size, total duplicated lines, and affected files.

### Step 1.3 — Duplication Metrics & Reporting

- Add a `DuplicationMetrics` struct to `internal/metrics/types.go`:
  - `ClonePairs`, `DuplicatedLines`, `DuplicationRatio`, `LargestCloneSize`.
- Add a `duplication` section to every output format (console, JSON, HTML, Markdown).
- Include a per-file duplication score so teams can prioritize extraction.

### Step 1.4 — Configuration & Thresholds

- Add config keys in `.go-stats-generator.yaml`:
  - `maintenance.duplication.min_block_lines` (default: 6) — minimum block size to consider.
  - `maintenance.duplication.similarity_threshold` (default: 0.80) — threshold for Type 3 clones.
  - `maintenance.duplication.ignore_test_files` (default: false).
- Wire thresholds into the `analyze` command flags.

---

## Phase 2: Naming Convention Analysis

Flag identifiers that violate Go conventions or harm readability.

### Step 2.1 — File Name Linting

- Check every `.go` file name against Go conventions:
  - Must be `snake_case` (lowercase, underscores).
  - No stuttering with the parent directory name (e.g. `http/http_client.go`).
  - `_test.go` suffix only for test files.
  - Flag single-character or overly generic names (`utils.go`, `helpers.go`, `misc.go`, `common.go`).
- Report a list of offending files with suggested renames.

### Step 2.2 — Identifier Name Quality Scoring

- For every exported function, method, type, const, and var:
  - Verify `MixedCaps` (no underscores in Go identifiers except for test functions).
  - Flag single-letter names outside of short loops and receivers.
  - Flag acronym casing violations (e.g. `Url` should be `URL`, `Id` should be `ID`).
  - Flag stuttering: a method `User.GetUser()` or a package-qualified name `user.UserService`.
  - Compute a name quality score based on length, specificity, and convention adherence.
- For unexported identifiers, apply the same rules with relaxed length requirements.

### Step 2.3 — Package Name Analysis

- Verify package names follow Go conventions:
  - Lowercase, single word preferred, no underscores or mixedCaps.
  - Not overly generic (`util`, `common`, `base`, `shared`, `lib`, `core`).
  - Not the same as a well-known standard library package when the intent differs.
- Flag packages whose name does not match their directory name.

### Step 2.4 — Naming Metrics & Reporting

- Add `NamingMetrics` to `internal/metrics/types.go`:
  - `FileNameViolations`, `IdentifierViolations`, `PackageNameViolations`, `OverallNamingScore`.
- Integrate into all output formats with violation details and suggested fixes.

---

## Phase 3: Misplaced Declarations (Functions/Methods in Wrong Files)

Identify declarations that would be easier to maintain if they lived elsewhere.

### Step 3.1 — Symbol-to-File Affinity Analysis

- For each function and method, compute an affinity score to its current file:
  - Count references to other symbols defined in the same file vs. other files.
  - Count references from other files to this symbol.
  - A function that mostly references symbols from another file has low affinity to its current file.
- Flag functions whose affinity to another file exceeds their affinity to their current file by a configurable margin.

### Step 3.2 — Method Receiver Placement Check

- For every method, verify that it is defined in the same file (or at least the same package) as its receiver type.
- Flag methods whose receiver type is defined in a different file, sorted by distance (same package but different file vs. different package).

### Step 3.3 — File Cohesion Scoring

- For each file, compute a cohesion score:
  - Ratio of intra-file references to total references.
  - Files that declare unrelated types, functions, or constants score low.
- Flag files below a configurable cohesion threshold (default: 0.3).
- Suggest logical splits when a file contains multiple unrelated clusters of declarations.

### Step 3.4 — Placement Metrics & Reporting

- Add `PlacementMetrics` to `internal/metrics/types.go`:
  - `MisplacedFunctions`, `MisplacedMethods`, `LowCohesionFiles`, `AvgFileCohesion`.
- Report each misplaced symbol with its current location, suggested location, and affinity scores.

---

## Phase 4: Missing & Inadequate Documentation

Surface documentation gaps that slow down onboarding and increase maintenance cost.

### Step 4.1 — Exported Symbol Documentation Coverage

- For every exported type, function, method, const, and var:
  - Check for the presence of a GoDoc comment (comment immediately preceding the declaration).
  - Verify the comment starts with the symbol name (Go convention).
  - Flag empty or trivially short comments (fewer than 5 words after the symbol name).
- Compute per-package and per-file documentation coverage percentages.

### Step 4.2 — Package-Level Documentation

- Check for the presence of `doc.go` or a package-level comment in at least one file per package.
- Flag packages without any package-level documentation.
- Score package documentation quality: presence, length, examples, and synopsis.

### Step 4.3 — Stale Annotation Tracking

- Scan all comments for `TODO`, `FIXME`, `HACK`, `BUG`, `XXX`, `DEPRECATED`, and `NOTE` annotations.
- Extract annotation text, file, line, and author (if available from git blame integration).
- Track annotation age via git history when available; flag annotations older than a configurable threshold (default: 180 days).
- Categorize by severity: `FIXME` and `BUG` > `HACK` > `TODO` > `NOTE`.

### Step 4.4 — Documentation Metrics & Reporting

- Add `DocumentationMetrics` to `internal/metrics/types.go`:
  - `ExportedWithoutDoc`, `DocCoveragePercent`, `PackagesWithoutDocGo`, `StaleAnnotations`, `AnnotationsByCategory`.
- Integrate into all output formats with a per-symbol documentation status table.

---

## Phase 5: Organizational & Structural Problems

Detect structural issues that make a codebase hard to navigate and maintain.

### Step 5.1 — File Size Analysis

- For each `.go` file, report total lines, code lines, comment lines, and blank lines (reuse existing `LineMetrics`).
- Flag files exceeding configurable thresholds:
  - `maintenance.organization.max_file_lines` (default: 500) — total lines.
  - `maintenance.organization.max_file_functions` (default: 20) — functions/methods per file.
  - `maintenance.organization.max_file_types` (default: 5) — type declarations per file.
- Rank files by maintenance burden (composite score of size, complexity, and declaration count).

### Step 5.2 — Package Size & Depth Analysis

- Flag packages with too many files (`maintenance.organization.max_package_files`, default: 20).
- Flag packages with too many exported symbols (`maintenance.organization.max_exported_symbols`, default: 50).
- Flag deeply nested directory structures (`maintenance.organization.max_directory_depth`, default: 5).
- Detect "mega-packages" that combine unrelated concerns (low cohesion + high symbol count).

### Step 5.3 — Import Graph Health

- Extend the existing `PackageAnalyzer` to report:
  - Files with excessive imports (`maintenance.organization.max_file_imports`, default: 15).
  - Packages that are imported by many other packages ("hub" packages) — high fan-in indicates a change bottleneck.
  - Packages that import many other packages ("authority" packages) — high fan-out indicates potential coupling.
- Compute an instability metric per package: `fan-out / (fan-in + fan-out)`.

### Step 5.4 — Organization Metrics & Reporting

- Add `OrganizationMetrics` to `internal/metrics/types.go`:
  - `OversizedFiles`, `OversizedPackages`, `DeepDirectories`, `HighFanInPackages`, `HighFanOutPackages`, `AvgPackageInstability`.
- Add a dedicated "Organization Health" section to all output formats.

---

## Phase 6: Additional Maintenance Burden Indicators

Catch common patterns that increase the cost of understanding and changing code.

### Step 6.1 — Magic Number & String Detection

- Walk function bodies to find numeric and string literals used directly in expressions (not in const declarations or struct initialization).
- Ignore common benign values: `0`, `1`, `-1`, `""`, `true`, `false`, `nil`.
- Flag magic values with their location, value, and usage context.
- Suggest extraction to named constants.

### Step 6.2 — Dead Code Detection

- Identify unexported functions, methods, types, constants, and variables that have zero references within their package.
- Cross-reference with `_test.go` files to avoid flagging test helpers.
- Identify unreachable code after unconditional `return`, `panic`, or `os.Exit` statements.
- Report dead code volume as a percentage of total code.

### Step 6.3 — Parameter List & Return Value Complexity

- Extend the existing `FunctionAnalyzer`:
  - Flag functions with more than a configurable number of parameters (`maintenance.burden.max_params`, default: 5).
  - Flag functions with more than a configurable number of return values (`maintenance.burden.max_returns`, default: 3).
  - Flag bool parameters ("flag arguments") that indicate the function does two things.
- Suggest introducing option structs or splitting functions.

### Step 6.4 — Deep Nesting Detection

- Extend the existing nesting depth analysis:
  - Flag functions with nesting depth exceeding a configurable threshold (`maintenance.burden.max_nesting`, default: 4).
  - Report the deepest nesting point with file and line for each flagged function.
- Suggest early returns and guard clauses as refactoring strategies.

### Step 6.5 — Shotgun Surgery & Feature Envy Indicators

- **Shotgun surgery**: Identify groups of functions that are always changed together (requires git history analysis).
  - Parse `git log --name-only` to find files that frequently co-change.
  - Flag function clusters across multiple files that change in >60% of the same commits.
- **Feature envy**: Identify methods that reference another type's fields or methods more than their own receiver's.
  - Flag methods where external references exceed self-references by a configurable ratio (default: 2:1).

### Step 6.6 — Burden Metrics & Reporting

- Add `BurdenMetrics` to `internal/metrics/types.go`:
  - `MagicNumbers`, `DeadCodeLines`, `DeadCodePercent`, `LongParamFunctions`, `DeeplyNestedFunctions`, `ShotgunSurgeryClusters`, `FeatureEnvyMethods`.
- Add a "Maintenance Burden Summary" section to all output formats with a composite burden score per file and per package.

---

## Phase 7: Composite Scoring & Actionable Output

Combine all maintenance burden signals into a unified, prioritized report.

### Step 7.1 — Maintenance Burden Index (MBI)

- Define a per-file and per-package composite score that weights:
  - Code duplication (Phase 1)
  - Naming violations (Phase 2)
  - Misplaced declarations (Phase 3)
  - Documentation gaps (Phase 4)
  - Organizational problems (Phase 5)
  - Additional burden indicators (Phase 6)
- Make weights configurable in `.go-stats-generator.yaml` under a `maintenance.scoring.weights` section.
- Normalize the score to a 0–100 scale where 0 = no burden and 100 = critical maintenance risk.

### Step 7.2 — Prioritized Refactoring Suggestions

- For every flagged issue, generate a concrete suggestion:
  - What to do (extract function, rename, move to another file, add documentation, etc.).
  - Estimated impact (how much the MBI would improve).
  - Effort classification (low / medium / high).
- Sort suggestions by impact-to-effort ratio so teams can tackle the highest-value changes first.

### Step 7.3 — Baseline Integration & Trend Tracking

- Extend the existing `baseline` and `diff` commands to include all maintenance burden metrics.
- Add burden-specific regression detection: alert when the MBI for any file or package increases beyond a configurable threshold.
- Integrate burden trends into the `trend` command for time-series visualization.

### Step 7.4 — CI/CD Quality Gates

- Add a `--max-burden-score` flag to the `analyze` command.
- Exit with a non-zero code when any file or package exceeds the threshold, enabling CI/CD enforcement.
- Support per-category thresholds (e.g. `--max-duplication-ratio 0.10`, `--max-undocumented-exports 5`).

---

## Implementation Order & Dependencies

| Phase | Depends On | New Analyzer File | Estimated Effort |
|-------|-----------|-------------------|-----------------|
| 1 — Duplication | — | `internal/analyzer/duplication.go` | Large |
| 2 — Naming | — | `internal/analyzer/naming.go` | Medium |
| 3 — Placement | Phase 2 | `internal/analyzer/placement.go` | Large |
| 4 — Documentation | — | `internal/analyzer/documentation.go` | Medium |
| 5 — Organization | Existing `PackageAnalyzer` | `internal/analyzer/organization.go` | Medium |
| 6 — Burden Indicators | Phases 1–5 | `internal/analyzer/burden.go` | Large |
| 7 — Composite Scoring | Phases 1–6 | `internal/analyzer/scoring.go` | Medium |

Phases 1, 2, 4, and 5 have no cross-dependencies and can be developed in parallel. Phase 3 benefits from the naming data produced by Phase 2. Phase 6 extends analyzers from earlier phases. Phase 7 integrates everything.

---

## Configuration Summary

All new thresholds are added under a `maintenance` section in `.go-stats-generator.yaml`:

```yaml
maintenance:
  duplication:
    min_block_lines: 6
    similarity_threshold: 0.80
    ignore_test_files: false
  naming:
    flag_generic_filenames: true
    flag_stuttering: true
    min_name_length: 2
  placement:
    affinity_margin: 0.25
    min_cohesion: 0.3
  documentation:
    require_exported_doc: true
    require_package_doc: true
    stale_annotation_days: 180
  organization:
    max_file_lines: 500
    max_file_functions: 20
    max_file_types: 5
    max_package_files: 20
    max_exported_symbols: 50
    max_directory_depth: 5
    max_file_imports: 15
  burden:
    max_params: 5
    max_returns: 3
    max_nesting: 4
    feature_envy_ratio: 2.0
  scoring:
    weights:
      duplication: 0.20
      naming: 0.10
      placement: 0.15
      documentation: 0.15
      organization: 0.15
      burden: 0.25
    max_burden_score: 70
```

---

## Testing Strategy

Each phase includes its own test file (`*_test.go`) alongside the analyzer. Tests follow the existing patterns in the repository:

- **Unit tests** using `testify/assert` and `testify/require` for each detection rule.
- **Table-driven tests** with Go source snippets in `testdata/` for every category of issue.
- **Integration tests** that run the full `analyze` command against known-bad sample projects and verify the output includes the expected findings.
- **Regression tests** for false-positive cases discovered during development.
- **Benchmark tests** to ensure new analyzers do not degrade performance below the 50,000-file-in-60-seconds target.
