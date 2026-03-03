# TASK DESCRIPTION:
Perform a data-driven codebase reorganization to achieve maximally navigable file structure by moving code segments without modification, using `go-stats-generator` package cohesion/coupling metrics to guide every reorganization decision, interface/struct analysis to identify consolidation targets, and differential validation to ensure measurable structural improvements while preserving functionality.

When results are ambiguous, such as a tie between cohesion scores or if multiple packages have equal file counts, always choose the package with the **lowest cohesion score** first (most in need of reorganization).

## CONSTRAINT:

Use only `go-stats-generator`, `go test`, `go build`, and existing tests for your analysis. You are absolutely forbidden from modifying code logic of any kind or using any other code analysis tools. You may only move code between files.

## PREREQUISITES:
**Minimum Required Version:** `go-stats-generator` v1.0.0 or higher
Install and configure `go-stats-generator` for comprehensive package structure analysis and reorganization tracking:

### Installation:
```bash
# First, check if go-stats-generator is already installed
which go-stats-generator
# If not, install it with `go install`
go install github.com/opd-ai/go-stats-generator@latest
```

## Recommendations:
```bash
# Extract only task-relevant sections from JSON; discard everything else
go-stats-generator analyze --format json | jq '{packages: .packages, structs: .structs, interfaces: .interfaces, functions: .functions}'
which jq || sudo apt-get install -y jq
```
**Section filter**: Use only `.packages`, `.structs`, `.interfaces`, and `.functions` from the report. Exclude `.patterns`, `.concurrency`, `.complexity`, `.documentation`, `.generics`, `.duplication`, `.naming`, `.placement`, `.organization`, `.burden`, `.scores`, `.suggestions` — they are not relevant to codebase reorganization.

### Required Analysis Workflow:
```bash
# Phase 0: Identify packages and select reorganization target
go-stats-generator analyze . --skip-tests --format json --output full-baseline.json --sections packages,structs,interfaces,functions
cat full-baseline.json | jq '.packages[] | {name, path, cohesion_score, coupling_score, files, structs, interfaces}'

# Phase 1: Establish per-package baseline before reorganization
go-stats-generator analyze ./[selected_package] --skip-tests --format json --output pkg-baseline.json --sections packages,structs,interfaces,functions

# Phases 2-4: Reorganize using struct/interface/package metrics as guide

# Phase 5: Post-reorganization validation
go-stats-generator analyze ./[selected_package] --skip-tests --format json --output pkg-reorganized.json --sections packages,structs,interfaces,functions

# Phase 6: Measure and document improvements
go-stats-generator diff pkg-baseline.json pkg-reorganized.json
go-stats-generator diff pkg-baseline.json pkg-reorganized.json --format html --output reorganization-report.html
```

## CONTEXT:
You are an automated Go codebase reorganizer using `go-stats-generator` for enterprise-grade structural analysis and reorganization validation. The tool provides precise package cohesion/coupling metrics, struct/interface inventories, and documentation coverage data that drive every reorganization decision. You move code between files — never modifying logic — to maximize on-disk navigability. Focus on packages with the lowest cohesion scores and highest coupling scores identified by the tool's package analysis engine, working one package at a time until each is complete with an AUDIT.md.

## INSTRUCTIONS:

### Phase 0: Data-Driven Sub-Package Selection
This phase uses `go-stats-generator` package metrics to select the most impactful reorganization target.

1. **Run Full Codebase Analysis:**
   ```bash
   go-stats-generator analyze . --skip-tests --format json --output full-baseline.json
   go-stats-generator analyze . --skip-tests
   ```

2. **Extract Package Metrics:**
   ```bash
   cat full-baseline.json | jq '[.packages[] | {name, path, cohesion_score, coupling_score, functions, structs, interfaces, files: (.files | length)}] | sort_by(.cohesion_score)'
   ```

3. **Identify Packages Lacking AUDIT.md:**
   - For each sub-package in `.packages[]`, check if an `AUDIT.md` file exists in that package directory
   - Create a list of packages that do NOT have an `AUDIT.md` file

4. **Select a Single Sub-Package** for reorganization:
   - From the packages lacking `AUDIT.md`, select ONE package to work on
   - Selection priority (choose the first applicable):
     a. Package with the **lowest cohesion score** (most disorganized, most benefit from reorganization)
     b. Package with the **highest coupling score** (most entangled, needs structural clarity)
     c. Package with the most `.go files` (highest complexity)
     d. Alphabetically first package if no other criteria apply
   - **IMPORTANT**: Only work on this single selected package until its reorganization and audit are complete

5. **Announce the Selection:**
   ```
   SELECTED_PACKAGE: [package_path]
   REASON: [selection reason with metric values]
   COHESION: [score] | COUPLING: [score]
   FILES: [count] .go files | STRUCTS: [count] | INTERFACES: [count]
   UNAUDITED_REMAINING: [count] packages without AUDIT.md
   ```

6. **Proceed to Phase 1** with the selected package as the scope for all subsequent phases

### Phase 1: Baseline Assessment and Preparation
*Scope: The single sub-package selected in Phase 0*

1. **Establish Per-Package Baseline:**
   ```bash
   go-stats-generator analyze ./[selected_package] --skip-tests --format json --output pkg-baseline.json
   go test ./[selected_package]/... > test-baseline.txt 2>&1
   ```

2. **Inventory Structs for Reorganization Targets:**
   ```bash
   cat pkg-baseline.json | jq '[.structs[] | {name, file, is_exported, total_fields, methods: (.methods | length)}]'
   ```
   - Identify files containing multiple structs (candidates for splitting)
   - Identify structs with many methods (candidates for dedicated files)

3. **Inventory Interfaces for Consolidation Targets:**
   ```bash
   cat pkg-baseline.json | jq '[.interfaces[] | {name, file, is_exported, method_count, implementation_count, embedding_depth}]'
   ```
   - Identify interfaces scattered across multiple files (candidates for consolidation)
   - Note implementation counts to understand interface importance

4. **Assess Documentation Coverage:**
   ```bash
   cat pkg-baseline.json | jq '.documentation'
   ```
   - Record baseline documentation coverage percentage
   - Identify exported symbols missing documentation

5. **Document the Current Structure:**
   - Create a file inventory listing all `.go` files and their primary contents
   - Record which structs, interfaces, constants, and utility functions live in which files

### Phase 2: Interface Consolidation
*Scope: The single sub-package selected in Phase 0*

Use `.interfaces[]` data from the baseline to identify consolidation targets.

1. **Identify Scattered Interfaces:**
   ```bash
   cat pkg-baseline.json | jq '[.interfaces[] | .file] | group_by(.) | map({file: .[0], count: length})'
   ```
   - If interfaces are spread across 3+ files, consolidation is warranted

2. **Consolidate Interfaces:**
   - Create either `interfaces.go` or `types.go` in the selected package (choose based on existing conventions)
   - For each interface found in the selected package:
     1. Copy the interface definition to the consolidated file
     2. Add comment: `// Originally from: [source_file.go]`
     3. Run `go build ./[selected_package]/...` to check for compilation errors
     4. If successful, remove the interface from its original location
     5. Run `go test ./[selected_package]/...` to verify no regressions
     6. If tests fail, revert the change and investigate dependencies

### Phase 3: Struct-Driven Structural Reorganization
*Scope: The single sub-package selected in Phase 0*

Use `.structs[]` data from the baseline to identify files needing reorganization.

1. **Identify Multi-Struct Files:**
   ```bash
   cat pkg-baseline.json | jq '[.structs[] | .file] | group_by(.) | map(select(length > 1) | {file: .[0], struct_count: length})'
   ```

2. **For Files Containing Multiple Structs:**
   1. Create new file named `[structname].go` for each struct (use lowercase of struct name)
   2. Move to the new file in this order:
      - Package declaration and imports
      - Documentation comment: `// [StructName] handles [brief description]`
      - Struct definition
      - Constructor function (typically `New[StructName]`)
      - All methods with that struct as receiver
      - Related helper functions used only by this struct
   3. Add comment: `// Code relocated from: [original_file.go]`
   4. Run `go build ./[selected_package]/...` after each struct migration
   5. Remove successfully migrated code from original file
   6. Run `go test ./[selected_package]/...` to confirm no regressions

3. **For Shared Constants:**
   1. Create `constants.go` in the selected package if it doesn't exist
   2. Group constants by logical category with section comments
   3. Move constants maintaining their original comments
   4. Add `// Originally defined in: [source_file.go]`
   5. Update imports in affected files
   6. Validate with `go build ./[selected_package]/...` and `go test ./[selected_package]/...`

4. **For Utility Functions:**
   1. Identify functions used across multiple structs/files
   2. Create `utils.go` or `helpers.go` for shared utilities
   3. Group related utilities with section comments
   4. Document each function's purpose and origin
   5. Maintain original function signatures exactly
   6. Test after each function group migration

5. **For Type Definitions (non-interface):**
   1. Simple types used by single struct: keep with that struct
   2. Shared types: move to `types.go` with interfaces
   3. Domain-specific types: create `[domain]_types.go`
   4. Always preserve type documentation

### Phase 4: Package Organization
*Scope: The single sub-package selected in Phase 0*

Use `.packages[]` metrics to guide subdomain splitting decisions.

1. **For Packages with 20+ Files:**
   ```bash
   cat pkg-baseline.json | jq '.packages[] | select((.files | length) >= 20) | {name, file_count: (.files | length), cohesion_score}'
   ```
   1. Identify logical subdomains within the selected package
   2. Create subdirectories for each subdomain
   3. Move related files maintaining import paths
   4. Update internal imports
   5. Run full test suite for the selected package after each subdirectory

2. **For Mixed Responsibility Files:**
   1. Separate by primary responsibility
   2. HTTP handlers → `handlers.go` or `[resource]_handler.go`
   3. Database operations → `db.go` or `[model]_db.go`
   4. External API clients → `[service]_client.go`
   5. Business logic → `[domain]_service.go`

### Phase 5: Implementation Gap Audit with go-stats-generator Data
*Scope: The single sub-package selected in Phase 0*

Use `go-stats-generator` analysis results to populate AUDIT.md with precise, data-driven findings.

1. **Run Post-Reorganization Analysis:**
   ```bash
   go-stats-generator analyze ./[selected_package] --skip-tests --format json --output pkg-reorganized.json
   ```

2. **Extract Documentation Gaps:**
   ```bash
   cat pkg-reorganized.json | jq '.documentation'
   ```

3. **Extract Interface Implementation Data:**
   ```bash
   cat pkg-reorganized.json | jq '[.interfaces[] | select(.implementation_count == 0) | {name, file, method_count}]'
   ```

4. **Extract Package Dependency Issues:**
   ```bash
   cat pkg-reorganized.json | jq '[.packages[] | {name, dependencies, dependents, cohesion_score, coupling_score}]'
   ```

5. **Create AUDIT.md in the Selected Package** with the following format:
   ```
   # Package Audit: [package_name]
   Generated during reorganization on: [date]
   Analysis engine: go-stats-generator

   ## Metrics Summary
   - Cohesion Score: [before] → [after]
   - Coupling Score: [before] → [after]
   - Documentation Coverage: [percentage]
   - Functions: [count] | Structs: [count] | Interfaces: [count]

   ## Implementation Gaps
   - Missing Implementations: [count]
   - Incomplete Features: [count]
   - Interface Violations: [count]
   - Untested Code: [count]
   - Dead Code: [count]
   - Error Handling Gaps: [count]
   - Documentation Gaps: [count]
   - Dependency Issues: [count]

   ## Detailed Findings

   ### Missing Implementations
   [List each missing implementation with file and line number]

   ### Incomplete Features
   [List each incomplete feature with TODO/FIXME text and location]

   ### Interface Violations
   [List each interface violation with struct, interface, and missing methods]

   ### Untested Code
   [List functions without corresponding tests]

   ### Dead Code
   [List unreachable or unused code discovered]

   ### Error Handling Gaps
   [List error handling issues]

   ### Documentation Gaps
   [List exported symbols missing documentation from go-stats-generator .documentation output]

   ### Dependency Issues
   [List dependency problems from go-stats-generator .packages[] dependency/coupling data]

   ## Recommendations
   [Prioritized list of fixes for the identified gaps]
   ```

6. **Track AUDIT.md creation** using the standardized AUDIT entry format defined in the **OUTPUT FORMAT** section below.

### Phase 6: Differential Validation and Documentation Enhancement
*Scope: The single sub-package selected in Phase 0*

1. **Measure Structural Improvements:**
   ```bash
   go-stats-generator diff pkg-baseline.json pkg-reorganized.json
   ```
   - Verify cohesion score improved (higher is better)
   - Verify coupling score improved (lower is better)
   - Confirm documentation coverage maintained or improved
   - Check for zero regressions in unchanged code

2. **Generate Improvement Report:**
   ```bash
   go-stats-generator diff pkg-baseline.json pkg-reorganized.json --format html --output reorganization-report.html
   ```

3. **Enhance Documentation:**
   For each file in the selected package after reorganization:
   1. Add file-level comment explaining the file's purpose
   2. Ensure every exported function has a comment starting with its name
   3. Document any non-obvious organizational decisions
   4. Create or update README.md in the selected package explaining the structure

### Phase 7: Completion and Next Package
After completing all phases for the selected package:

1. **Verify Completion Criteria:**
   ```bash
   go test ./[selected_package]/...
   go build ./[selected_package]/...
   go-stats-generator analyze ./[selected_package] --skip-tests --min-doc-coverage 0.7
   ```
   - AUDIT.md exists in the selected package directory
   - All tests pass (zero regressions from test-baseline.txt)
   - Build succeeds
   - Documentation coverage ≥ 70%
   - All reorganization steps documented in output

2. **Output Completion Status:**
   ```
   PACKAGE_COMPLETE: [package_path]
   AUDIT_FILE: [package_path]/AUDIT.md
   COHESION: [before] → [after] ([improvement])
   COUPLING: [before] → [after] ([improvement])
   DOC_COVERAGE: [percentage]
   TESTS: PASS - [number] tests, [number] passed
   BUILD: SUCCESS
   ```

3. **Check for Remaining Packages:**
   - List packages still lacking AUDIT.md
   - If packages remain, return to **Phase 0** and select the next package
   - If all packages have AUDIT.md, output final summary

## OUTPUT FORMAT:

Structure your output as:

### At the Start of Reorganization (Phase 0):
```
go-stats-generator identified package reorganization targets:
1. Package: [name] at [path]
   - Cohesion: [score] | Coupling: [score]
   - Files: [count] | Structs: [count] | Interfaces: [count]
   - Priority: [Critical/High/Medium]

2. Package: [name] at [path]
   - Cohesion: [score] | Coupling: [score]
   - Files: [count] | Structs: [count] | Interfaces: [count]
   - Priority: [Critical/High/Medium]

... (continue for all unaudited packages)

SELECTED_PACKAGE: [package_path]
REASON: [selection reason with metric values]
COHESION: [score] | COUPLING: [score]
FILES: [count] .go files | STRUCTS: [count] | INTERFACES: [count]
UNAUDITED_REMAINING: [count] packages without AUDIT.md
```

### After Each Reorganization Step:
```
MOVED: [description of what was moved]
FROM: [source_file.go]
TO: [destination_file.go]
TESTS: [PASS/FAIL] - [number] tests, [number] passed
BUILD: [SUCCESS/FAIL]
```

### After Completing Implementation Gap Audit:
```
AUDIT: [package_name]
GAPS_FOUND: [total_count]
  - Missing Implementations: [count]
  - Incomplete Features: [count]
  - Interface Violations: [count]
  - Untested Code: [count]
  - Dead Code: [count]
  - Error Handling Gaps: [count]
  - Documentation Gaps: [count]
  - Dependency Issues: [count]
FILE: [package_path]/AUDIT.md
```

### After Completing One Package (Phase 7):
```
PACKAGE_COMPLETE: [package_path]
AUDIT_FILE: [package_path]/AUDIT.md
COHESION: [before] → [after] ([improvement])
COUPLING: [before] → [after] ([improvement])
DOC_COVERAGE: [percentage]
TESTS: PASS - [number] tests, [number] passed
BUILD: SUCCESS
NEXT_PACKAGE: [next_package_path or "NONE - all packages audited"]
```

### Improvement Validation (per package):
```
Differential analysis results:
- Cohesion: [old_score] → [new_score] ([improvement_%])
- Coupling: [old_score] → [new_score] ([improvement_%])
- Documentation: [old_%] → [new_%]
- Files created: [count] | Files modified: [count] | Files deleted: [count]
- Regressions: [count]
```

### Final Summary (when all packages are complete):
```
REORGANIZATION COMPLETE
Packages reorganized: [number]
Files created: [number]
Files modified: [number]
Files deleted: [number]
AUDIT.md files created: [number]
Total implementation gaps found: [number]
Average cohesion improvement: [percentage]
Average coupling improvement: [percentage]
Test status: [PASS/FAIL]
Build status: [SUCCESS/FAIL]
```

## REORGANIZATION THRESHOLDS:
```
Package Selection Priority:
  Critical = Cohesion < 0.3 (severely disorganized)
  High     = Cohesion < 0.5 (needs significant restructuring)
  Medium   = Cohesion < 0.7 (could benefit from reorganization)

Structural Triggers:
  Multi-Struct Files  = File contains > 1 exported struct definition
  Scattered Interfaces = Interfaces spread across ≥ 3 files in same package
  Large Package        = Package contains ≥ 20 .go files

Post-Reorganization Quality Gates:
  Cohesion Score     ≥ 0.5 (improved from baseline)
  Documentation Coverage ≥ 70% (--min-doc-coverage 0.7)
  Zero test regressions
  Zero build errors
  AUDIT.md created for each package
```
<!-- Last verified: 2025-07-25 against package.go:calculateCohesion, calculateCoupling and config defaults -->

Reorganization Threshold = Cohesion < 0.7 OR Multi-Struct Files > 0 OR Scattered Interfaces OR Files ≥ 20
- If no targets: "Reorganization complete: go-stats-generator baseline analysis found no packages exceeding professional structural thresholds."

## QUALITY CRITERIA:
- One sub-package completed at a time with AUDIT.md created before moving to next
- Zero test regressions: test-baseline.txt matches final test output for each package
- Zero build errors throughout the process
- Every exported symbol has documentation (validated by `--min-doc-coverage 0.7`)
- File names clearly indicate contents
- Related code is co-located (validated by cohesion score improvement)
- No code logic modifications — only moves between files
- All moves are traced with comments
- AUDIT.md created for each reorganized package with go-stats-generator metric data
- All implementation gaps documented with specific file and line references
- Differential validation via `go-stats-generator diff` confirms measurable improvement

## EXAMPLE WORKFLOW:
```bash
$ go-stats-generator analyze . --skip-tests
=== PACKAGE ANALYSIS ===
Package             Cohesion  Coupling  Files  Structs  Interfaces
--------------------------------------------------------------------------------
internal/analyzer      0.28      0.72     8       6          4
internal/reporter      0.41      0.55     5       3          2
internal/config        0.65      0.30     3       2          1
internal/metrics       0.72      0.25     4       3          1

$ go-stats-generator analyze . --skip-tests --format json --output full-baseline.json
$ cat full-baseline.json | jq '[.packages[] | {name, path, cohesion_score, coupling_score, files: (.files | length), structs, interfaces}] | sort_by(.cohesion_score)'
[
  {"name": "analyzer", "path": "internal/analyzer", "cohesion_score": 0.28, "coupling_score": 0.72, "files": 8, "structs": 6, "interfaces": 4},
  {"name": "reporter", "path": "internal/reporter", "cohesion_score": 0.41, "coupling_score": 0.55, "files": 5, "structs": 3, "interfaces": 2},
  {"name": "config",   "path": "internal/config",   "cohesion_score": 0.65, "coupling_score": 0.30, "files": 3, "structs": 2, "interfaces": 1},
  {"name": "metrics",  "path": "internal/metrics",  "cohesion_score": 0.72, "coupling_score": 0.25, "files": 4, "structs": 3, "interfaces": 1}
]

SELECTED_PACKAGE: internal/analyzer
REASON: Lowest cohesion score (0.28) among unaudited packages — most disorganized
COHESION: 0.28 | COUPLING: 0.72
FILES: 8 .go files | STRUCTS: 6 | INTERFACES: 4
UNAUDITED_REMAINING: 3 packages without AUDIT.md

$ # Phase 1: Per-package baseline
$ go-stats-generator analyze ./internal/analyzer --skip-tests --format json --output pkg-baseline.json
$ cat pkg-baseline.json | jq '[.structs[] | {name, file}]'
[
  {"name": "FunctionAnalyzer", "file": "analyzer.go"},
  {"name": "StructAnalyzer",   "file": "analyzer.go"},
  {"name": "PackageAnalyzer",  "file": "analyzer.go"},
  {"name": "InterfaceTracker", "file": "tracker.go"},
  {"name": "MetricsCollector", "file": "tracker.go"},
  {"name": "ReportBuilder",    "file": "output.go"}
]

$ cat pkg-baseline.json | jq '[.interfaces[] | {name, file}]'
[
  {"name": "Analyzer",   "file": "analyzer.go"},
  {"name": "Tracker",    "file": "tracker.go"},
  {"name": "Collector",  "file": "types.go"},
  {"name": "Formatter",  "file": "output.go"}
]

$ go test ./internal/analyzer/... > test-baseline.txt 2>&1

$ # Phase 2: Consolidate interfaces (scattered across 4 files → interfaces.go)

MOVED: Analyzer interface definition
FROM: analyzer.go
TO: interfaces.go
TESTS: PASS - 47 tests, 47 passed
BUILD: SUCCESS

MOVED: Tracker interface definition
FROM: tracker.go
TO: interfaces.go
TESTS: PASS - 47 tests, 47 passed
BUILD: SUCCESS

MOVED: Collector interface definition
FROM: types.go
TO: interfaces.go
TESTS: PASS - 47 tests, 47 passed
BUILD: SUCCESS

MOVED: Formatter interface definition
FROM: output.go
TO: interfaces.go
TESTS: PASS - 47 tests, 47 passed
BUILD: SUCCESS

$ # Phase 3: Split multi-struct files (analyzer.go has 3 structs)

MOVED: FunctionAnalyzer struct and methods
FROM: analyzer.go
TO: function_analyzer.go
TESTS: PASS - 47 tests, 47 passed
BUILD: SUCCESS

MOVED: StructAnalyzer struct and methods
FROM: analyzer.go
TO: struct_analyzer.go
TESTS: PASS - 47 tests, 47 passed
BUILD: SUCCESS

MOVED: PackageAnalyzer struct and methods
FROM: analyzer.go
TO: package_analyzer.go
TESTS: PASS - 47 tests, 47 passed
BUILD: SUCCESS

MOVED: InterfaceTracker struct and methods
FROM: tracker.go
TO: interface_tracker.go
TESTS: PASS - 47 tests, 47 passed
BUILD: SUCCESS

MOVED: MetricsCollector struct and methods
FROM: tracker.go
TO: metrics_collector.go
TESTS: PASS - 47 tests, 47 passed
BUILD: SUCCESS

$ # Phase 5: Post-reorganization analysis and audit
$ go-stats-generator analyze ./internal/analyzer --skip-tests --format json --output pkg-reorganized.json

AUDIT: analyzer
GAPS_FOUND: 5
  - Missing Implementations: 1
  - Incomplete Features: 2
  - Interface Violations: 0
  - Untested Code: 1
  - Dead Code: 0
  - Error Handling Gaps: 1
  - Documentation Gaps: 0
  - Dependency Issues: 0
FILE: internal/analyzer/AUDIT.md

$ # Phase 6: Differential validation
$ go-stats-generator diff pkg-baseline.json pkg-reorganized.json
=== IMPROVEMENT SUMMARY ===
PACKAGE: internal/analyzer

STRUCTURAL METRICS:
- Cohesion: 0.28 → 0.71 (154% improvement) ✓
- Coupling: 0.72 → 0.45 (38% reduction) ✓
- Documentation: 62% → 85% ✓

FILE CHANGES:
- Files created: 6
- Files modified: 3
- Files deleted: 0

QUALITY GATES:
  Cohesion ≥ 0.5: ✓ (0.71)
  Documentation ≥ 70%: ✓ (85%)
  Zero test regressions: ✓
  Zero build errors: ✓

QUALITY SCORE: 92/65 (+27 improvement)
REGRESSIONS: 0

PACKAGE_COMPLETE: internal/analyzer
AUDIT_FILE: internal/analyzer/AUDIT.md
COHESION: 0.28 → 0.71 (154% improvement)
COUPLING: 0.72 → 0.45 (38% reduction)
DOC_COVERAGE: 85%
TESTS: PASS - 47 tests, 47 passed
BUILD: SUCCESS
NEXT_PACKAGE: internal/reporter
```

This data-driven approach ensures reorganization decisions are based on quantitative package cohesion/coupling analysis rather than subjective assessment, with measurable validation of structural improvements for each package via `go-stats-generator diff`.
