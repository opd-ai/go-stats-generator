# TASK DESCRIPTION:
Perform a data-driven documentation accuracy audit to identify and correct the **top documentation gaps and inaccuracies** across all markdown files in the repository below professional documentation coverage thresholds. Use `go-stats-generator` baseline analysis (with --skip-tests), documentation and naming metrics, and autonomous correction to ensure measurable documentation quality improvements while preserving existing accurate content.

When results are ambiguous, such as a tie between documentation coverage scores or if multiple issues exist in the same file, always choose the file with the **lowest doc coverage** first.

## CONSTRAINT:

Use only `go-stats-generator` and existing tests for your analysis. You are absolutely forbidden from writing new code of any kind or using any other code analysis tools.

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
**Section filter**: Use only `.documentation` and `.naming` from the report. Exclude `.functions`, `.structs`, `.interfaces`, `.packages`, `.patterns`, `.concurrency`, `.complexity`, `.generics`, `.duplication`, `.placement`, `.organization`, `.burden`, `.scores`, `.suggestions` — they are not relevant to documentation accuracy auditing.

### Required Analysis Workflow:
```bash
# Phase 1: Establish baseline and identify documentation gaps
go-stats-generator analyze . --min-doc-coverage 0.8 --skip-tests --format json --output baseline.json --sections documentation,naming
go-stats-generator analyze . --min-doc-coverage 0.8 --skip-tests

# Phase 2: Extract documentation and naming metrics for audit targets
cat baseline.json | jq '.documentation'
cat baseline.json | jq '.naming'

# Phase 3: Post-correction validation
go-stats-generator analyze . --format json --output corrected.json --min-doc-coverage 0.8 --skip-tests --sections documentation,naming

# Phase 4: Measure and document improvements
go-stats-generator diff baseline.json corrected.json
go-stats-generator diff baseline.json corrected.json --format html --output doc-audit-report.html
```

## CONTEXT:
You are an autonomous documentation auditor using `go-stats-generator` for enterprise-grade documentation coverage analysis and accuracy verification. The tool provides precise doc coverage metrics, identifies exported symbols missing godoc comments, and measures improvements through differential analysis. Focus on documentation gaps identified by the tool's documentation and naming analysis engines, cross-referencing markdown content against actual code artifacts to ensure zero documentation drift.

Accuracy is the paramount concern. You must never introduce false information or make assumptions. When uncertainty exists, you must flag it rather than guess. Your corrections should be traceable to specific code artifacts.

## INSTRUCTIONS:

### Phase 1: Data-Driven Target Identification
1. **Run Baseline Analysis:**
  ```bash
  go-stats-generator analyze . --min-doc-coverage 0.8 --skip-tests
  ```
  - Record the overall **documentation coverage** percentage
  - Identify all **exported symbols missing godoc** from `.naming` output
  - Note packages with the lowest doc coverage ratios

2. **Extract Documentation Metrics:**
  ```bash
  go-stats-generator analyze . --min-doc-coverage 0.8 --skip-tests --format json --output baseline.json --sections documentation,naming
  cat baseline.json | jq '.documentation'
  cat baseline.json | jq '.naming | {exported_without_doc: [.symbols[] | select(.exported == true and .has_doc == false) | .name]}'
  ```
  - Capture the overall documentation metrics:
    * Documentation Coverage (percentage of exported symbols with godoc)
    * Exported Symbols Missing Docs (count and list)
    * Packages Below Threshold (packages with coverage < 0.8)

3. **Scan Repository for Markdown Files:**
  ```bash
  find . -name '*.md' -not -path './.git/*' | sort
  ```
  - Create a processing queue ordered by:
    * README.md files (highest priority)
    * API and configuration documentation
    * Setup/installation guides
    * Other documentation

4. **Prioritize Audit Targets:**
  From the baseline analysis and markdown scan, prioritize corrections:
  - **Critical:** Exported symbols with zero godoc AND referenced in markdown with incorrect signatures
  - **High:** Markdown references to file paths, functions, or configs that no longer exist
  - **Medium:** Documentation coverage below 0.8 threshold with minor inaccuracies

  When prioritizing between files:
  - If coverage scores are tied, choose the file with **more exported symbols missing docs** first
  - If both are tied, choose the file appearing in **more markdown references** first
  - Target at least 5 files; extend to 10 if more than 5 exceed Critical or High priority thresholds

### Phase 2: Systematic Verification and Autonomous Correction
1. **Cross-Reference Markdown Against Code:**
  For each markdown file, extract and verify:
  - All code blocks with language identifiers
  - Inline code references (text within backticks)
  - File paths (patterns like `internal/...`, `cmd/...`, `pkg/...`)
  - Function/method names (patterns like `functionName()`)
  - Struct and interface names (PascalCase references)
  - Configuration flags and default values
  - Command-line examples using `go-stats-generator`

2. **Verify Each Reference Against Codebase:**
  ```bash
  # Verify file paths exist
  find . -name '*.go' -not -path './.git/*' -not -name '*_test.go' | sort

  # Verify function signatures match documentation
  grep -rn '^func ' --include='*.go' --exclude='*_test.go' .

  # Verify struct definitions
  grep -rn '^type .* struct' --include='*.go' --exclude='*_test.go' .

  # Verify interface definitions
  grep -rn '^type .* interface' --include='*.go' --exclude='*_test.go' .
  ```

3. **Determine Correction Type and Apply:**
  ```
  - SAFE_AUTO_FIX: Unambiguous correction (e.g., parameter name typo, updated file path)
  - NEEDS_REVIEW: Multiple valid options or behavioral changes
  - INFO_MISSING: Required information not found in code
  - DEPRECATED: Referenced code marked as deprecated
  ```

  Apply safe corrections autonomously:
  - Function signature updates matching `go-stats-generator` output
  - File path corrections verified against `.packages[].files[]`
  - Flag/default value updates verified against CLI flag definitions
  - Add missing godoc comments for exported symbols identified by naming analysis

4. **Flag Uncertain Items for Review:**
  ```markdown
  <!-- AUDIT_FLAG: NEEDS_REVIEW
  Issue: Function 'processData' not found. Similar functions found:
  - processDataAsync() at internal/analyzer/processor.go:34
  - processUserData() at internal/analyzer/user.go:12
  Original text: "Call processData() to handle the input"
  -->
  ```

5. **Add Verification Timestamps:**
  ```markdown
  <!-- Last verified: YYYY-MM-DD against commit: [hash] -->
  ```

### Phase 3: Differential Validation
1. **Measure Improvements:**
  ```bash
  go-stats-generator analyze . --skip-tests --format json --output corrected.json --sections documentation,naming
  go-stats-generator diff baseline.json corrected.json
  ```
  - Verify documentation coverage increased toward 0.8 threshold
  - Confirm exported symbols missing docs count decreased
  - Confirm no new documentation gaps were introduced
  - Check for zero regressions in unchanged code

2. **Generate Improvement Report:**
  ```bash
  go-stats-generator diff baseline.json corrected.json --format html --output doc-audit-report.html
  ```

### Phase 4: Quality Verification
1. **Validate Metrics Achievement:**
  - Documentation coverage increased above --min-doc-coverage 0.8 threshold
  - All critical exported symbols now have godoc comments
  - All markdown references verified against codebase
  - No inaccurate references remain uncorrected or unflagged
  - Overall documentation quality trend positive

2. **Confirm Functional Preservation:**
  - All tests pass: `go test ./...`
  - Build succeeds: `go build ./...`
  - Only documentation files and godoc comments modified
  - No behavioral changes introduced

## OUTPUT FORMAT:

Structure your response as:

### 1. Baseline Documentation Summary
```
go-stats-generator identified documentation metrics:
  Documentation Coverage: [x.xx]%
  Exported Symbols Missing Docs: [n]
  Packages Below Threshold: [n]
  Markdown Files Audited: [n]
  Total References Checked: [n]

Top audit targets:
1. File: [path]
   - Doc Coverage: [x.xx]%
   - Missing Godoc: [n] exported symbols
   - Markdown References: [n] verified, [n] incorrect
   - Priority: [Critical/High/Medium]
   - Corrections: [n] SAFE_AUTO_FIX, [n] NEEDS_REVIEW

2. File: [path]
   - Doc Coverage: [x.xx]%
   - Missing Godoc: [n] exported symbols
   - Markdown References: [n] verified, [n] incorrect
   - Priority: [Critical/High/Medium]
   - Corrections: [n] SAFE_AUTO_FIX, [n] NEEDS_REVIEW

... (continue for top 5-10 files)
```

### 2. Corrections Applied
Present each correction with:
- File path and line number
- Correction type (SAFE_AUTO_FIX / NEEDS_REVIEW / INFO_MISSING / DEPRECATED)
- Original text and corrected text
- Traceability reference to source code artifact

### 3. Improvement Validation
```
Differential analysis results:
- Documentation Coverage: [old_%] → [new_%] ([improvement_%] increase)
- Exported Symbols Missing Docs: [old_count] → [new_count] ([fixed_count] fixed)
- Markdown References Corrected: [count]
- Items Flagged for Review: [count]
- Regressions: [count]
- Overall quality improvement: [score]
```

## DOCUMENTATION THRESHOLDS:
```
Documentation Coverage:
  Minimum Doc Coverage = 0.8 (80% of exported symbols must have godoc)
  Flag all exported symbols with missing godoc comments

Correction Classification:
  SAFE_AUTO_FIX:  Unambiguous correction traceable to specific code artifact
  NEEDS_REVIEW:   Multiple valid options or behavioral change implied
  INFO_MISSING:   Referenced information not found in codebase
  DEPRECATED:     Referenced code marked as deprecated

Audit Priority:
  Critical = Exported symbols with zero godoc AND incorrect markdown references
  High     = Markdown references to nonexistent file paths, functions, or configs
  Medium   = Documentation coverage below 0.8 with minor inaccuracies

Post-Audit Quality Gates:
  Documentation Coverage ≥ 0.8
  Zero uncorrected inaccurate references
  All corrections traceable to code artifacts
  No behavioral changes introduced
```
<!-- Last verified: 2025-07-25 against documentation.go:AnalyzeDocumentation, naming.go:AnalyzeNaming, and config defaults -->

Documentation Threshold = Doc Coverage < 0.8 OR Exported Symbols Missing Docs > 0 OR Incorrect Markdown References > 0
- If no targets: "Documentation audit complete: go-stats-generator baseline analysis found no documentation gaps exceeding professional coverage thresholds."

## EXAMPLE WORKFLOW:
```bash
$ go-stats-generator analyze . --min-doc-coverage 0.8 --skip-tests
=== DOCUMENTATION ANALYSIS ===
Documentation Coverage: 62.5%
Exported Symbols Missing Docs: 15
Packages Below Threshold (0.8): 4

Packages by Coverage:
Package                          Coverage  Missing Docs
--------------------------------------------------------------------------------
internal/analyzer                   0.55          6
internal/reporter                   0.60          4
cmd                                 0.72          3
pkg/metrics                         0.78          2

Exported Symbols Missing Godoc:
  AnalyzeFile           internal/analyzer/file.go:23
  ProcessResults        internal/analyzer/results.go:45
  FormatConsole         internal/reporter/console.go:18
  WriteHTML             internal/reporter/html.go:31
  ... (11 more)

$ # Extract baseline JSON for diff comparison
$ go-stats-generator analyze . --min-doc-coverage 0.8 --skip-tests --format json --output baseline.json --sections documentation,naming
$ cat baseline.json | jq '.documentation | {coverage, missing_docs, packages_below_threshold}'
{
  "coverage": 0.625,
  "missing_docs": 15,
  "packages_below_threshold": 4
}

$ # Cross-reference markdown against code
$ cat baseline.json | jq '.naming | {exported_without_doc: [.symbols[] | select(.exported == true and .has_doc == false) | .name]}'
{
  "exported_without_doc": [
    "AnalyzeFile",
    "ProcessResults",
    "FormatConsole",
    "WriteHTML"
  ]
}

$ # Audit and correct markdown files, add missing godoc...

$ go-stats-generator analyze . --min-doc-coverage 0.8 --skip-tests --format json --output corrected.json --sections documentation,naming
$ go-stats-generator diff baseline.json corrected.json
=== IMPROVEMENT SUMMARY ===
DOCUMENTATION IMPROVEMENTS:

METRICS:
- Documentation Coverage: 62.5% → 88.3% (25.8% increase) ✓
- Exported Symbols Missing Docs: 15 → 3 (12 fixed) ✓
- Packages Below Threshold: 4 → 1 (3 brought above 0.8) ✓

CORRECTIONS APPLIED:
  Markdown references corrected: 8
  Godoc comments added: 12
  File path references updated: 3
  Flag/config references updated: 2

ITEMS FLAGGED FOR REVIEW:
  NEEDS_REVIEW: 2 (ambiguous function references)
  INFO_MISSING: 1 (external dependency reference)

QUALITY GATES:
  Documentation coverage above 0.8 threshold: ✓
  All corrections traceable to code: ✓
  No behavioral changes introduced: ✓
  All tests passing: ✓

QUALITY SCORE: 92/67 (+25 improvement)
REGRESSIONS: 0
DOCUMENTATION COVERAGE NOW ABOVE 0.8 THRESHOLD: ✓
```

This data-driven approach ensures documentation audit decisions are based on quantitative coverage analysis rather than subjective assessment, with measurable validation of improvements and full traceability of corrections to code artifacts.