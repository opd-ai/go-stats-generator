# TASK DESCRIPTION:
Perform a data-driven functional breakdown analysis on a single Go file within the `go-stats-generator` codebase using the tool's own metrics to identify and refactor functions exceeding professional complexity thresholds. Use the tool's baseline analysis, targeted refactoring guidance, and differential validation to ensure measurable complexity improvements while preserving functionality.

## PREREQUISITES:
**Development Environment:** Working within the `go-stats-generator` repository  
Build and configure the local `go-stats-generator` for comprehensive complexity analysis and improvement tracking:

### Local Build:
```bash
# Build the current version for self-analysis
go build -o go-stats-generator ./cmd/go-stats-generator
```

### Required Analysis Workflow:
```bash
# Phase 1: Establish baseline and identify targets
./go-stats-generator analyze . --max-complexity 10 --max-function-length 30 --skip-tests --exclude testdata --format json --output baseline.json
./go-stats-generator analyze . --max-complexity 10 --max-function-length 30 --skip-tests --exclude testdata

# Phase 2: Generate refactoring recommendations  
Using the results generated in phase 1, select a high-complexity function suitable for refactoring.

# Phase 3: Post-refactoring validation
./go-stats-generator analyze . --format json --output refactored.json --max-complexity 13 --max-function-length 30 --skip-tests --exclude testdata

# Phase 4: Measure and document improvements
./go-stats-generator diff baseline.json refactored.json
./go-stats-generator diff baseline.json refactored.json --format html --output improvements.html
```

## CONTEXT:
You are an automated Go code auditor using `go-stats-generator` for enterprise-grade complexity analysis and refactoring validation on the tool's own codebase. The tool provides precise metrics, identifies refactoring targets, and measures improvements through differential analysis. Focus on functions with the highest complexity scores identified by the tool's analysis engine. Ignore all files in the `./testdata` directory as they contain sample code for testing purposes.

## INSTRUCTIONS:

### Phase 1: Data-Driven Target Identification
1. **Run Baseline Analysis:**
  ```bash
  ./go-stats-generator analyze . --exclude testdata
  ```
  - Record the highest complexity function and its metrics
  - Note specific complexity contributors (cyclomatic, nesting, signature)
  - Identify the file containing the most complex function

2. **Generate Refactoring Plan:**
  ```bash
  ./go-stats-generator analyze [target-file] --format json --exclude testdata
  ```
  - Use tool's suggestions for logical extraction points
  - Identify functions exceeding thresholds:
    * Overall complexity > 13.0
    * Line count > 30 (code lines only)
    * Cyclomatic complexity > 10
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
    func validateInput(input string) error {
        // ...
    }
    ```
  - Add GoDoc comments starting with function name

3. **Preserve Analysis-Verified Patterns:**
  - Maintain error propagation chains
  - Keep defer statements in correct scope
  - Preserve variable access patterns

### Phase 3: Differential Validation
1. **Measure Improvements:**
  ```bash
  ./go-stats-generator diff baseline.json refactored.json
  ```
  - Verify target function shows significant complexity reduction (>50%)
  - Confirm no new functions exceed thresholds
  - Check for zero regressions in unchanged code

2. **Generate Improvement Report:**
  ```bash
  ./go-stats-generator diff baseline.json refactored.json --format html --output report.html --include-recommendations
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

Signature Complexity = (params * 0.5) + (returns * 0.3) + (interfaces * 0.8) + generics_penalty
- generics_penalty: An additional score (typically 1.0 per type parameter) added for each generic type parameter in the function signature to reflect increased complexity.

Refactoring Threshold = Overall Complexity > 13.0 OR Lines > 30 OR Cyclomatic > 10
- If no targets: "Refactor complete: go-stats-generator baseline analysis found no functions exceeding professional complexity thresholds."

## COMPLEXITY REFERENCE (go-stats-generator calculation):
```
Overall Complexity = cyclomatic + (nesting_depth * 0.5) + (cognitive * 0.3)
Signature Complexity = (params * 0.5) + (returns * 0.3) + (interfaces * 0.8) + generics_penalty
Refactoring Threshold = Overall Complexity > 13.0 OR Lines > 30 OR Cyclomatic > 10
```

## EXAMPLE WORKFLOW:
```bash
$ ./go-stats-generator analyze . --exclude testdata
=== TOP COMPLEX FUNCTIONS ===
1. processComplexOrder (order.go): 25.4 complexity
  - Lines: 45 code lines 
  - Cyclomatic: 18
  - Nesting: 4
  - Recommendations: Extract 4 logical blocks

$ ./go-stats-generator diff baseline.json refactored.json 
=== IMPROVEMENT SUMMARY ===
MAJOR IMPROVEMENTS:
EXTRACTED FUNCTIONS:
(All steps validated by automated differential analysis to ensure measurable, data-driven improvements.)
  
EXTRACTED FUNCTIONS:
  validateOrderData: 5.1 complexity ✓
  calculatePricing: 7.3 complexity ✓
  finalizeOrder: 6.8 complexity ✓
  
QUALITY SCORE: 95/100 (+22 improvement)
REGRESSIONS: 0
```

This data-driven approach ensures refactoring decisions are based on quantitative analysis rather than subjective assessment, with measurable validation of improvements.

