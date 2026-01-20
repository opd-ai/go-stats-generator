# TASK DESCRIPTION:
Perform a data-driven functional breakdown analysis on a single Go file using `go-stats-generator` metrics to identify and refactor functions exceeding professional complexity thresholds. Use the tool's baseline analysis(with --skip-tests), targeted refactoring guidance, and differential validation to ensure measurable complexity improvements while preserving functionality. When results are ambiguous, such as a tie between complexity scores or if one threshold is exceeded but not another, always choose the longest function first.

## CONSTRAINT:

Use only `go-stats-generator` and existing tests for your analysis. You are absolutely forbidden from writing new code of any kind or using any other code analysis tools.

## PREREQUISITES:
**Minimum Required Version:** `go-stats-generator` v1.0.0 or higher  
Install and configure `go-stats-generator` for comprehensive complexity analysis and improvement tracking:

### Installation:
```bash
# First, check if go-stats-generator is already installed
which go-stats-generator
# If not, install it with `go install`
go install github.com/opd-ai/go-stats-generator@latest
```

## Recommendations:
```bash
# When long json outputs are encountered, use `jq`
go-stats-generator analyze --output json | jq .example
# Check if it is installed
which jq
# If it is not, install it
sudo apt-get install jq
```

### Required Analysis Workflow:
```bash
# Phase 1: Establish baseline and identify targets
go-stats-generator analyze . --max-complexity 9 --max-function-length 40 --skip-tests --format json --output baseline.json
go-stats-generator analyze . --max-complexity 9 --max-function-length 40 --skip-tests

# Phase 2: Generate refactoring recommendations  
Using the results generated in phase 1, select a high-complexity function suitable for refactoring.

# Phase 3: Post-refactoring validation
go-stats-generator analyze . --format json --output refactored.json --max-complexity 9 --max-function-length 40 --skip-tests

# Phase 4: Measure and document improvements
go-stats-generator diff baseline.json refactored.json
go-stats-generator diff baseline.json refactored.json --format html --output improvements.html
```

## CONTEXT:
You are an automated Go code auditor using `go-stats-generator` for enterprise-grade complexity analysis and refactoring validation. The tool provides precise metrics, identifies refactoring targets, and measures improvements through differential analysis. Focus on functions with the highest complexity scores identified by the tool's analysis engine.

## INSTRUCTIONS:

### Phase 1: Data-Driven Target Identification
1. **Run Baseline Analysis:**
  ```bash
  go-stats-generator analyze .
  ```
  - Record the highest complexity function and its metrics
  - Note specific complexity contributors (cyclomatic, nesting, signature)
  - Identify the file containing the most complex function

2. **Generate Refactoring Plan:**
  ```bash
  go-stats-generator analyze [target-file] --format json
  ```
  - Use tool's suggestions for logical extraction points
  - Identify functions exceeding thresholds:
    * Overall complexity > 9.0 (default threshold)
    * Line count > 40 (code lines only)
    * Cyclomatic complexity > 9
    * Nesting depth > 3

### Phase 2: Guided Refactoring Implementation
1. **Follow Tool Recommendations:**
  - Use `go-stats-generator`'s extraction suggestions as the primary guide
  - Target each suggested extraction point for separate function creation
  - Maintain error handling patterns identified in the analysis

2. **Create Focused Extractions:**
  - Extract each logical block identified by the tool
  - Name functions using verb-first camelCase (e.g., `validateInput`, `calculateResult`)
    - ❌ Avoid noun-first or snake_case names (e.g., `inputValidator`, `calculate_result`)
  - Target metrics per extracted function:
    * <20 lines of code
    * Cyclomatic complexity <8
  - Add GoDoc comments starting with function name  
    *Example:*  
    ```go
    // validateInput checks if the provided input meets all required criteria.
    // etc...
    func validateInput(input string) error {
        // ...
    }
    ```
  - Add GoDoc comments starting with function name and containing a description of the function's purpose and operation

3. **Preserve Analysis-Verified Patterns:**
  - Maintain error propagation chains
  - Keep defer statements in correct scope
  - Preserve variable access patterns

### Phase 3: Differential Validation
1. **Measure Improvements:**
  ```bash
  go-stats-generator diff baseline.json refactored.json
  ```
  - Verify target function shows significant complexity reduction (>50%)
  - Confirm no new functions exceed thresholds
  - Check for zero regressions in unchanged code

2. **Generate Improvement Report:**
  ```bash
  go-stats-generator diff baseline.json refactored.json --format html --output report.html
  ```

### Phase 4: Quality Verification
1. **Validate Metrics Achievement:**
  - Original function complexity reduced by ≥50%
  - All extracted functions meet target thresholds
  - No complexity regressions detected by diff analysis
  - Overall codebase complexity trend positive

2. **Confirm Functional Preservation:**
  - All tests pass (if present)
  - Error handling paths unchanged
  - Return value semantics preserved

## OUTPUT FORMAT:

Structure your response as:

### 1. Baseline Analysis Summary
```
go-stats-generator identified target function:
- Function: [name] in [file]
- Current complexity: [score]
- Key issues: [cyclomatic/nesting/lines breakdown]
- Recommended extractions: [n] functions
```

### 2. Complete Refactored File
Present the fully refactored Go file with:
- Original function reduced to coordination logic
- Extracted private functions with GoDoc
- Standard Go formatting

### 3. Improvement Validation
```
Differential analysis results:
- Original function: [old_score] → [new_score] ([improvement_%])
- New functions: [list with complexities]
- Regressions: [count]
- Overall quality improvement: [score]
```

Signature Complexity = (params * 0.5) + (returns * 0.3) + (interfaces * 0.8) + (generics * 1.5) + variadic_penalty
- variadic_penalty: An additional score (1.0) added for variadic parameters (...args) to reflect increased complexity.
- generics: The actual multiplier is 1.5 per generic type parameter, not 1.0 as previously documented.

Refactoring Threshold = Overall Complexity > 9.0 OR Lines > 40 OR Cyclomatic > 9
- If no targets: "Refactor complete: go-stats-generator baseline analysis found no functions exceeding professional complexity thresholds."

## COMPLEXITY REFERENCE (go-stats-generator calculation):
```
Overall Complexity = cyclomatic + (nesting_depth * 0.5) + (cognitive * 0.3)
Signature Complexity = (params * 0.5) + (returns * 0.3) + (interfaces * 0.8) + (generics * 1.5) + variadic_penalty
Refactoring Threshold = Overall Complexity > 9.0 OR Lines > 40 OR Cyclomatic > 9
```
<!-- Last verified: 2025-07-25 against function.go:calculateComplexity and calculateSignatureComplexity -->

## TOP 10 REFACTORING TARGETS

The following functions have been identified as priority refactoring targets based on `go-stats-generator` analysis. Each exceeds professional complexity thresholds (Overall Complexity > 9.0 OR Lines > 40 OR Cyclomatic > 9).

### 1. Generate (internal/reporter/console.go)
| Metric | Value | Threshold | Status |
|--------|-------|-----------|--------|
| Lines | 183 | 40 | ❌ EXCEEDS |
| Cyclomatic | 26 | 9 | ❌ EXCEEDS |
| Overall Complexity | 35.3 | 9.0 | ❌ EXCEEDS |

**Recommended Extractions:**
- `writeReportHeader` - Extract header generation logic
- `writeReportSections` - Extract section iteration and conditional rendering
- `writeReportBody` - Extract main content generation
- `writeReportFooter` - Extract footer and summary logic

**Target Complexity After Refactoring:** < 8.0 per function

---

### 2. WriteDiff (internal/reporter/console.go)
| Metric | Value | Threshold | Status |
|--------|-------|-----------|--------|
| Lines | 91 | 40 | ❌ EXCEEDS |
| Cyclomatic | 18 | 9 | ❌ EXCEEDS |
| Overall Complexity | 24.9 | 9.0 | ❌ EXCEEDS |

**Recommended Extractions:**
- `formatDiffHeader` - Extract diff header formatting
- `formatRegressions` - Extract regression display logic
- `formatImprovements` - Extract improvement display logic
- `formatChangeSummary` - Extract summary generation

**Target Complexity After Refactoring:** < 8.0 per function

---

### 3. Cleanup (internal/storage/sqlite.go)
| Metric | Value | Threshold | Status |
|--------|-------|-----------|--------|
| Lines | 67 | 40 | ❌ EXCEEDS |
| Cyclomatic | 13 | 9 | ❌ EXCEEDS |
| Overall Complexity | 18.4 | 9.0 | ❌ EXCEEDS |

**Recommended Extractions:**
- `cleanupByAge` - Extract age-based deletion logic with policy checks
- `cleanupByCount` - Extract count-based deletion logic
- `buildCleanupQuery` - Extract dynamic query construction
- `logCleanupResults` - Extract result logging

**Target Complexity After Refactoring:** < 7.0 per function

---

### 4. walkForNestingDepth (internal/analyzer/function.go)
| Metric | Value | Threshold | Status |
|--------|-------|-----------|--------|
| Lines | 66 | 40 | ❌ EXCEEDS |
| Cyclomatic | 12 | 9 | ❌ EXCEEDS |
| Overall Complexity | 16.6 | 9.0 | ❌ EXCEEDS |

**Recommended Extractions:**
- `handleConditionalNesting` - Extract if/else depth tracking
- `handleLoopNesting` - Extract for/range loop depth tracking
- `handleSwitchNesting` - Extract switch/select statement handling
- `updateMaxDepth` - Extract depth comparison and update logic

**Target Complexity After Refactoring:** < 6.0 per function

---

### 5. compareFunctionMetrics (internal/metrics/diff.go)
| Metric | Value | Threshold | Status |
|--------|-------|-----------|--------|
| Lines | 68 | 40 | ❌ EXCEEDS |
| Cyclomatic | 9 | 9 | ⚠️ AT LIMIT |
| Overall Complexity | 13.7 | 9.0 | ❌ EXCEEDS |

**Recommended Extractions:**
- `buildMetricsMap` - Extract baseline metrics indexing
- `detectNewFunctions` - Extract new function detection
- `detectRemovedFunctions` - Extract removed function detection
- `compareExistingFunctions` - Extract metric comparison for matching functions

**Target Complexity After Refactoring:** < 6.0 per function

---

### 6. calculateDelta (internal/metrics/diff.go)
| Metric | Value | Threshold | Status |
|--------|-------|-----------|--------|
| Lines | 34 | 40 | ✅ OK |
| Cyclomatic | 9 | 9 | ⚠️ AT LIMIT |
| Overall Complexity | 13.7 | 9.0 | ❌ EXCEEDS |

**Recommended Extractions:**
- `computeDeltaDirection` - Extract direction calculation (increase/decrease/unchanged)
- `computeDeltaPercentage` - Extract percentage change calculation
- `assessDeltaSeverity` - Extract threshold-based severity assessment

**Target Complexity After Refactoring:** < 5.0 per function

---

### 7. AnalyzeInterfacesWithPath (internal/analyzer/interface.go)
| Metric | Value | Threshold | Status |
|--------|-------|-----------|--------|
| Lines | 32 | 40 | ✅ OK |
| Cyclomatic | 8 | 9 | ✅ OK |
| Overall Complexity | 13.4 | 9.0 | ❌ EXCEEDS |

**Recommended Extractions:**
- `extractInterfaceMethods` - Extract method signature analysis
- `calculateInterfaceComplexity` - Extract complexity scoring
- `buildInterfaceMetrics` - Extract metrics struct construction

**Target Complexity After Refactoring:** < 6.0 per function

---

### 8. runTrendRegressions (cmd/trend.go)
| Metric | Value | Threshold | Status |
|--------|-------|-----------|--------|
| Lines | 69 | 40 | ❌ EXCEEDS |
| Cyclomatic | 9 | 9 | ⚠️ AT LIMIT |
| Overall Complexity | 13.2 | 9.0 | ❌ EXCEEDS |

**Recommended Extractions:**
- `loadTrendData` - Extract data loading and validation
- `analyzeTrendPatterns` - Extract pattern analysis logic
- `formatTrendOutput` - Extract output formatting by format type
- `writeTrendReport` - Extract report writing logic

**Target Complexity After Refactoring:** < 6.0 per function

---

### 9. Store (internal/storage/sqlite.go)
| Metric | Value | Threshold | Status |
|--------|-------|-----------|--------|
| Lines | 59 | 40 | ❌ EXCEEDS |
| Cyclomatic | 9 | 9 | ⚠️ AT LIMIT |
| Overall Complexity | 13.2 | 9.0 | ❌ EXCEEDS |

**Recommended Extractions:**
- `prepareSnapshotData` - Extract data preparation and serialization
- `executeSnapshotInsert` - Extract database insertion logic
- `storeSnapshotMetadata` - Extract metadata handling
- `handleStoreError` - Extract error handling and rollback logic

**Target Complexity After Refactoring:** < 6.0 per function

---

### 10. runFileAnalysis (cmd/analyze.go)
| Metric | Value | Threshold | Status |
|--------|-------|-----------|--------|
| Lines | 78 | 40 | ❌ EXCEEDS |
| Cyclomatic | 9 | 9 | ⚠️ AT LIMIT |
| Overall Complexity | 12.7 | 9.0 | ❌ EXCEEDS |

**Recommended Extractions:**
- `parseSourceFile` - Extract file parsing logic
- `analyzeFileStructures` - Extract struct/interface analysis
- `analyzeFileFunctions` - Extract function analysis
- `buildFileReport` - Extract report construction

**Target Complexity After Refactoring:** < 6.0 per function

---

## REFACTORING PRIORITY ORDER

Process functions in this order based on complexity impact:

| Priority | Function | File | Complexity | Impact |
|----------|----------|------|------------|--------|
| 1 | Generate | internal/reporter/console.go | 35.3 | Critical |
| 2 | WriteDiff | internal/reporter/console.go | 24.9 | High |
| 3 | Cleanup | internal/storage/sqlite.go | 18.4 | High |
| 4 | walkForNestingDepth | internal/analyzer/function.go | 16.6 | Medium |
| 5 | compareFunctionMetrics | internal/metrics/diff.go | 13.7 | Medium |
| 6 | calculateDelta | internal/metrics/diff.go | 13.7 | Medium |
| 7 | AnalyzeInterfacesWithPath | internal/analyzer/interface.go | 13.4 | Medium |
| 8 | runTrendRegressions | cmd/trend.go | 13.2 | Medium |
| 9 | Store | internal/storage/sqlite.go | 13.2 | Medium |
| 10 | runFileAnalysis | cmd/analyze.go | 12.7 | Medium |

## EXAMPLE WORKFLOW:
```bash
$ go-stats-generator analyze .
=== TOP COMPLEX FUNCTIONS ===
1. Generate (internal/reporter/console.go): 35.3 complexity
  - Lines: 183 code lines 
  - Cyclomatic: 26
  - Nesting: 3
  - Recommendations: Extract 4 logical blocks

2. WriteDiff (internal/reporter/console.go): 24.9 complexity
  - Lines: 91 code lines 
  - Cyclomatic: 18
  - Nesting: 2
  - Recommendations: Extract 4 logical blocks

$ go-stats-generator diff baseline.json refactored.json 
=== IMPROVEMENT SUMMARY ===
MAJOR IMPROVEMENTS:
- Generate: 35.3 → 7.2 (79.6% reduction) ✓
- WriteDiff: 24.9 → 6.8 (72.7% reduction) ✓

EXTRACTED FUNCTIONS:
  writeReportHeader: 4.1 complexity ✓
  writeReportSections: 6.3 complexity ✓
  writeReportBody: 5.8 complexity ✓
  writeReportFooter: 3.2 complexity ✓
  formatDiffHeader: 3.5 complexity ✓
  formatRegressions: 5.9 complexity ✓
  formatImprovements: 5.7 complexity ✓
  formatChangeSummary: 4.8 complexity ✓
  
QUALITY SCORE: 95/67 (+28 improvement)
REGRESSIONS: 0
```

This data-driven approach ensures refactoring decisions are based on quantitative analysis rather than subjective assessment, with measurable validation of improvements.
