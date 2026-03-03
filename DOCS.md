# TASK DESCRIPTION:
Perform a data-driven documentation audit to identify and fix all **undocumented exported symbols and inaccurate documentation** in the codebase above professional documentation coverage thresholds. Use `go-stats-generator` baseline analysis (with --skip-tests), targeted documentation corrections, and differential validation to ensure measurable documentation improvements while preserving accurate existing content.

When results are ambiguous, such as a tie between documentation coverage scores across packages, always choose the package with the **most exported symbols** first.

## CONSTRAINT:

Use only `go-stats-generator` and existing tests for your analysis. You are absolutely forbidden from writing new code of any kind or using any other code analysis tools. Apply all changes autonomously without prompting for approval. Do not modify auto-generated files (detected by generator signatures such as `godocdown`-style headers, CI-stamped headers, or similar tool artifacts).

## PREREQUISITES:
**Minimum Required Version:** `go-stats-generator` v1.0.0 or higher
Install and configure `go-stats-generator` for comprehensive documentation coverage analysis and improvement tracking:

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
go-stats-generator analyze --format json | jq '{documentation: .documentation, naming: .naming}'
which jq || sudo apt-get install -y jq
```
**Section filter**: Use only `.documentation` and `.naming` from the report. Exclude `.functions`, `.structs`, `.interfaces`, `.packages`, `.patterns`, `.concurrency`, `.complexity`, `.generics`, `.duplication`, `.placement`, `.organization`, `.burden`, `.scores`, `.suggestions` — they are not relevant to documentation auditing and godoc corrections.

### Required Analysis Workflow:
```bash
# Phase 1: Establish baseline and identify documentation gaps
go-stats-generator analyze . --min-doc-coverage 0.8 --skip-tests --format json --output baseline.json --sections documentation,naming
go-stats-generator analyze . --min-doc-coverage 0.8 --skip-tests

# Phase 2: Inspect documentation and naming metrics
cat baseline.json | jq '.documentation'
cat baseline.json | jq '.naming'

# Phase 3: Post-correction validation
go-stats-generator analyze . --min-doc-coverage 0.8 --skip-tests --format json --output corrected.json --sections documentation,naming

# Phase 4: Measure and document improvements
go-stats-generator diff baseline.json corrected.json
go-stats-generator diff baseline.json corrected.json --format html --output docs-report.html
```

## CONTEXT:
You are an automated Go documentation auditor using `go-stats-generator` for enterprise-grade documentation coverage analysis and correction validation. The tool provides precise doc-coverage metrics, identifies undocumented exported symbols, detects naming convention violations, and measures improvements through differential analysis. Focus on exported symbols flagged by the tool's `.documentation` and `.naming` analysis outputs, correcting inaccurate godoc comments, adding missing godoc for exported symbols, and updating non-generated Markdown files to reflect the current codebase state.

## INSTRUCTIONS:

### Phase 1: Data-Driven Target Identification
1. **Run Baseline Analysis:**
  ```bash
  go-stats-generator analyze . --min-doc-coverage 0.8 --skip-tests
  ```
  - Record the **overall documentation coverage** and per-package breakdowns
  - Note which exported symbols lack godoc comments
  - Identify naming convention violations for exported symbols

2. **Extract Documentation and Naming Metrics:**
  ```bash
  go-stats-generator analyze . --min-doc-coverage 0.8 --skip-tests --format json --output baseline.json --sections documentation,naming
  cat baseline.json | jq '.documentation'
  cat baseline.json | jq '.naming'
  ```
  - Capture the overall documentation metrics:
    * Doc Coverage (percentage of exported symbols with godoc)
    * Undocumented Symbols (list of exported symbols missing godoc)
    * Stale Comments (godoc comments that no longer match their symbol's signature or behavior)
  - Capture naming metrics:
    * Exported symbols with naming convention violations
    * Symbols whose godoc does not start with the symbol name

3. **Prioritize Documentation Targets:**
  From the baseline analysis, select symbols and files for correction in priority order:
  - **Critical (exported public API with no godoc):** Add godoc immediately — these are the highest priority
  - **High (godoc present but inaccurate or stale):** Correct in second pass
  - **Medium (godoc present but does not start with symbol name):** Fix format if time permits

  When prioritizing between symbols:
  - If coverage gaps are tied across packages, choose the package with **more exported symbols** first
  - If a symbol has both missing godoc and a naming violation, fix the godoc first
  - Target all undocumented exported symbols; extend to stale/inaccurate comments after coverage reaches ≥80%

4. **Identify Markdown Files for Update:**
  ```bash
  cat baseline.json | jq '.documentation.markdown_files'
  ```
  - Identify which Markdown files are auto-generated (by header patterns, generator comments, or tooling config) and **exclude them from edits**
  - For each non-generated Markdown file, compare its content against the current codebase
  - Flag inaccurate feature checklists, package/module descriptions, import paths, and code examples

### Phase 2: Guided Documentation Corrections
1. **Add Missing Godoc Comments:**
  - For each undocumented exported symbol, add a godoc comment starting with the symbol name
  - Use concise, technical prose consistent with Go documentation conventions
  - Do not add marketing or subjective language
  - Example:
    ```go
    // AnalyzePackage performs static analysis on a single Go package,
    // collecting function, struct, and interface metrics.
    func AnalyzePackage(path string) (*PackageResult, error) {
        // ...
    }
    ```

2. **Correct Inaccurate Godoc Comments:**
  - Compare each godoc comment against its symbol's current signature and behavior
  - Fix parameter descriptions that reference renamed or removed parameters
  - Update return value documentation to match current return types
  - Remove references to deleted functionality
  - Do not rewrite accurate documentation — minimum viable changes only

3. **Fix Naming Convention Violations in Godoc:**
  - Ensure every godoc comment starts with the exported symbol name
  - Example fix:
    ```go
    // Bad:  "This function analyzes packages."
    // Good: "AnalyzePackage performs static analysis on a single Go package."
    ```

4. **Update Non-Generated Markdown Files:**
  - Verify feature checklists match implemented functionality
  - Correct package/module descriptions and import paths
  - Update code examples to compile against the current API
  - Do not add new documentation sections that are not already present
  - Do not modify auto-generated files

### Phase 3: Differential Validation
1. **Measure Improvements:**
  ```bash
  go-stats-generator analyze . --min-doc-coverage 0.8 --skip-tests --format json --output corrected.json --sections documentation,naming
  go-stats-generator diff baseline.json corrected.json
  ```
  - Verify documentation coverage increased to ≥80%
  - Confirm undocumented symbol count decreased
  - Confirm no new naming violations were introduced
  - Check for zero regressions in unchanged code

2. **Generate Improvement Report:**
  ```bash
  go-stats-generator diff baseline.json corrected.json --format html --output docs-report.html
  ```

### Phase 4: Quality Verification
1. **Validate Metrics Achievement:**
  - Documentation coverage ≥80% across all packages
  - All exported symbols have godoc comments
  - All godoc comments start with the symbol name
  - No stale or inaccurate godoc comments remain
  - No auto-generated files modified
  - No marketing or subjective language introduced
  - Overall codebase documentation quality trend positive

2. **Confirm Functional Preservation:**
  - All tests pass: `go test ./...`
  - Build succeeds: `go build ./...`
  - No source code logic changes — only comments and Markdown

## OUTPUT FORMAT:

Structure your response as:

### 1. Baseline Documentation Summary
```
go-stats-generator identified documentation metrics:
  Doc Coverage: [x.xx]%
  Undocumented Exported Symbols: [n]
  Stale/Inaccurate Comments: [n]
  Naming Convention Violations: [n]
  Non-Generated Markdown Files: [n]

Top documentation targets:
1. Symbol: [name] in [file]
   - Issue: [missing godoc/stale comment/naming violation]
   - Priority: [Critical/High/Medium]
   - Action: [add godoc/correct comment/fix format]

2. Symbol: [name] in [file]
   - Issue: [missing godoc/stale comment/naming violation]
   - Priority: [Critical/High/Medium]
   - Action: [add godoc/correct comment/fix format]

... (continue for all flagged symbols)
```

### 2. Corrected Files
Present each corrected file with:
- Added or corrected godoc comments
- Updated Markdown content
- Standard Go formatting preserved

### 3. Improvement Validation
```
Differential analysis results:
- Doc Coverage: [old_%] → [new_%] ([improvement_%] increase)
- Undocumented Symbols: [old_count] → [new_count] ([fixed_count] fixed)
- Stale Comments: [old_count] → [new_count] ([corrected_count] corrected)
- Naming Violations: [old_count] → [new_count] ([fixed_count] fixed)
- Markdown Files Updated: [count]
- Regressions: [count]
- Overall quality improvement: [score]
```

## DOCUMENTATION THRESHOLDS:
```
Documentation Coverage:
  Minimum Doc Coverage = 0.80 (80% of exported symbols must have godoc)
  Godoc Format = comment must start with the exported symbol name

Correction Priority:
  Critical = exported public API symbol with no godoc comment
  High     = godoc present but inaccurate, stale, or references removed parameters
  Medium   = godoc present but does not start with symbol name

Post-Correction Quality Gates:
  Doc Coverage ≥ 80%
  All exported symbols have godoc
  All godoc comments start with symbol name
  No auto-generated files modified
  No marketing or subjective language introduced
  No source code logic changes
```
<!-- Last verified: 2025-07-25 against documentation.go:AnalyzeDocCoverage and naming.go:AnalyzeNaming -->

Documentation Threshold = Doc Coverage < 80% OR Undocumented Symbols > 0 OR Stale Comments > 0
- If no targets: "Documentation audit complete: go-stats-generator baseline analysis found no exported symbols below professional documentation coverage thresholds."

## EXAMPLE WORKFLOW:
```bash
$ go-stats-generator analyze . --min-doc-coverage 0.8 --skip-tests
=== DOCUMENTATION ANALYSIS ===
Doc Coverage: 62.5%
Undocumented Exported Symbols: 15
Stale/Inaccurate Comments: 3
Naming Convention Violations: 5
Non-Generated Markdown Files: 4

Undocumented Symbols:
Package             Symbol                   Type
--------------------------------------------------------------------------------
internal/analyzer   AnalyzePackage           func
internal/analyzer   PackageResult            struct
internal/reporter   FormatConsole            func
internal/reporter   HTMLRenderer             struct
internal/scanner    ScanDirectory            func
cmd                 NewAnalyzeCmd            func
...

Naming Violations:
File                             Symbol            Issue
--------------------------------------------------------------------------------
internal/reporter/console.go     FormatConsole     godoc does not start with symbol name
internal/metrics/types.go        MetricSet         godoc references removed field
...

$ # Extract baseline JSON for diff comparison
$ go-stats-generator analyze . --min-doc-coverage 0.8 --skip-tests --format json --output baseline.json --sections documentation,naming
$ cat baseline.json | jq '.documentation | {doc_coverage, undocumented_count, stale_count}'
{
  "doc_coverage": 0.625,
  "undocumented_count": 15,
  "stale_count": 3
}

$ cat baseline.json | jq '.naming | {violations_count, symbols: [.violations[] | .symbol]}'
{
  "violations_count": 5,
  "symbols": ["FormatConsole", "MetricSet", "BuildReport", "ScanWorker", "ConfigLoader"]
}

$ # Correct each documentation gap in priority order...

$ go-stats-generator analyze . --min-doc-coverage 0.8 --skip-tests --format json --output corrected.json --sections documentation,naming
$ go-stats-generator diff baseline.json corrected.json
=== IMPROVEMENT SUMMARY ===
SYMBOLS DOCUMENTED: 15
COMMENTS CORRECTED: 3
NAMING VIOLATIONS FIXED: 5

DOCUMENTATION METRICS:
- Doc Coverage: 62.5% → 95.0% (32.5% increase) ✓
- Undocumented Symbols: 15 → 0 (15 fixed) ✓
- Stale Comments: 3 → 0 (3 corrected) ✓
- Naming Violations: 5 → 0 (5 fixed) ✓

MARKDOWN FILES UPDATED:
  README.md: corrected import paths, updated feature checklist
  ROADMAP.md: updated completion status for implemented features
  CONTRIBUTING.md: fixed code example to match current API

QUALITY GATES:
  Doc coverage above 80% threshold: ✓
  All exported symbols have godoc: ✓
  All godoc comments start with symbol name: ✓
  No auto-generated files modified: ✓
  No marketing language introduced: ✓
  All tests passing: ✓

QUALITY SCORE: 95/62 (+33 improvement)
REGRESSIONS: 0
DOC COVERAGE NOW ABOVE 80% THRESHOLD: ✓
```

This data-driven approach ensures documentation corrections are based on quantitative coverage analysis rather than subjective assessment, with measurable validation of improvements across all exported symbols and non-generated Markdown files.
