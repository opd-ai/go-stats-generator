# TASK DESCRIPTION:
Perform a data-driven functional breakdown analysis to identify and refactor the **top 5-10 longest, most complex functions** in the codebase to below professional complexity thresholds. Use `go-stats-generator` baseline analysis (with --skip-tests), targeted refactoring guidance, and differential validation to ensure measurable complexity improvements while preserving functionality.

When results are ambiguous, such as a tie between complexity scores or if one threshold is exceeded but not another, always choose the longest function first.

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
  - Record the **top 5-10 highest complexity functions** and their metrics
  - Note specific complexity contributors (cyclomatic, nesting, signature) for each
  - Identify the files containing these complex functions

2. **Prioritize Refactoring Targets:**
  From the baseline analysis, select functions for refactoring in priority order based on overall complexity scores (as defined in COMPLEXITY REFERENCE below):
  - **Critical (Complexity > 20.0):** Refactor immediately - these are the highest priority
  - **High (Complexity 15.0-20.0):** Refactor in second pass
  - **Medium (Complexity 9.0-15.0):** Refactor if time permits
  
  When prioritizing between functions:
  - If complexity scores are tied, choose the **longest function first**
  - If one threshold is exceeded but not another, choose the function that exceeds more thresholds
  - Target at least 5 functions; extend to 10 if more than 5 functions exceed Critical or High priority thresholds

3. **Generate Refactoring Plan:**
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
  - **Carefully consider mutexes, locks, and thread safety:** When extracting functions that operate within a critical section (e.g., code protected by `sync.Mutex`, `sync.RWMutex`, or other synchronization primitives), ensure that lock boundaries are preserved correctly. Do not split a lock/unlock pair across the original and extracted functions in a way that could introduce race conditions or deadlocks. If the extracted function requires access to shared state, verify that the caller still holds the appropriate lock or that the extracted function acquires it safely. Always run `go test -race` after refactoring concurrency-sensitive code to detect data races.

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
go-stats-generator identified top refactoring targets:
1. Function: [name] in [file]
   - Current complexity: [score]
   - Key issues: [cyclomatic/nesting/lines breakdown]
   - Priority: [Critical/High/Medium]
   - Recommended extractions: [n] functions

2. Function: [name] in [file]
   - Current complexity: [score]
   - Key issues: [cyclomatic/nesting/lines breakdown]
   - Priority: [Critical/High/Medium]
   - Recommended extractions: [n] functions

... (continue for top 5-10 functions)
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

## EXAMPLE WORKFLOW:
```bash
$ go-stats-generator analyze .
=== TOP COMPLEX FUNCTIONS ===
1. processComplexOrder (order.go): 25.4 complexity [Critical]
  - Lines: 45 code lines 
  - Cyclomatic: 18
  - Nesting: 4
  - Recommendations: Extract 4 logical blocks

2. handleDataTransform (transform.go): 18.7 complexity [High]
  - Lines: 52 code lines
  - Cyclomatic: 14
  - Nesting: 3
  - Recommendations: Extract 3 logical blocks

3. validateComplexInput (validator.go): 15.2 complexity [High]
  - Lines: 38 code lines
  - Cyclomatic: 11
  - Nesting: 3
  - Recommendations: Extract 2 logical blocks

4. generateReport (reporter.go): 12.8 complexity [Medium]
  - Lines: 41 code lines
  - Cyclomatic: 9
  - Nesting: 2
  - Recommendations: Extract 2 logical blocks

5. parseConfiguration (config.go): 11.3 complexity [Medium]
  - Lines: 35 code lines
  - Cyclomatic: 8
  - Nesting: 2
  - Recommendations: Extract 2 logical blocks

$ # Refactor each function in priority order...

$ go-stats-generator diff baseline.json refactored.json 
=== IMPROVEMENT SUMMARY ===
FUNCTIONS REFACTORED: 5

MAJOR IMPROVEMENTS:
- processComplexOrder: 25.4 → 7.2 (72% reduction) ✓
- handleDataTransform: 18.7 → 6.8 (64% reduction) ✓
- validateComplexInput: 15.2 → 5.9 (61% reduction) ✓
- generateReport: 12.8 → 5.4 (58% reduction) ✓
- parseConfiguration: 11.3 → 4.8 (58% reduction) ✓

EXTRACTED FUNCTIONS:
  validateOrderData: 5.1 complexity ✓
  calculatePricing: 7.3 complexity ✓
  finalizeOrder: 6.8 complexity ✓
  transformInput: 4.2 complexity ✓
  mapOutputFields: 5.5 complexity ✓
  ... (additional extracted functions)
  
QUALITY SCORE: 95/67 (+28 improvement)
REGRESSIONS: 0
ALL FUNCTIONS NOW BELOW THRESHOLDS: ✓
```

This data-driven approach ensures refactoring decisions are based on quantitative analysis rather than subjective assessment, with measurable validation of improvements across the top 5-10 most complex functions.
