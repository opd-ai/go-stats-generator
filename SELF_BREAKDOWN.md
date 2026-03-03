# TASK DESCRIPTION:
Perform a data-driven functional breakdown analysis on the `go-stats-generator` codebase using its own metrics to identify and refactor functions exceeding professional complexity thresholds. Use the tool's baseline analysis, targeted refactoring guidance, and differential validation to ensure measurable complexity improvements while preserving functionality and fixing any bugs discovered during the breakdown process.

## CONSTRAINT:

Use only `go-stats-generator` and existing tests for your analysis. You are absolutely forbidden from using any other code analysis tools. However, you may write code fixes for bugs discovered during the breakdown analysis process.

## PREREQUISITES:
**Current Repository:** `go-stats-generator` (self-analysis)  
Use the local build of `go-stats-generator` for comprehensive complexity analysis and improvement tracking of its own codebase:

### Build and Setup:
```bash
# Build the current version
make build

# Or use go build directly
go build -o go-stats-generator ./cmd/go-stats-generator
```

### Required Analysis Workflow:
```bash
# Phase 1: Establish baseline and identify targets in our own codebase
./go-stats-generator analyze . --max-complexity 10 --max-function-length 30 --skip-tests --format json --output baseline.json
./go-stats-generator analyze . --max-complexity 10 --max-function-length 30 --skip-tests

# Phase 2: Deep dive into specific complex modules
./go-stats-generator analyze ./internal/analyzer --max-complexity 8 --max-function-length 25
./go-stats-generator analyze ./internal/metrics --max-complexity 8 --max-function-length 25  
./go-stats-generator analyze ./cmd --max-complexity 8 --max-function-length 25

# Phase 3: Post-refactoring validation
./go-stats-generator analyze . --format json --output refactored.json --max-complexity 10 --max-function-length 30 --skip-tests

# Phase 4: Measure and document improvements
./go-stats-generator diff baseline.json refactored.json
./go-stats-generator diff baseline.json refactored.json --format html --output improvements.html
```

## CONTEXT:
You are performing self-analysis on the `go-stats-generator` codebase to improve its own code quality. This dogfooding approach will validate the tool's effectiveness while optimizing the codebase. Focus on the core analysis engine, CLI commands, and metric calculation functions that likely contain the highest complexity due to their comprehensive feature sets.

## INSTRUCTIONS:

### Phase 1: Self-Analysis Target Identification
1. **Run Comprehensive Baseline Analysis:**
  ```bash
  ./go-stats-generator analyze . --format json --output self-baseline.json
  ./go-stats-generator analyze .
  ```
  - Focus on core modules: `internal/analyzer`, `internal/metrics`, `internal/storage`
  - Record the highest complexity functions and their locations
  - Identify complexity hotspots in the analysis engine itself
  - Note any patterns of high complexity in CLI command handlers

2. **Module-Specific Deep Analysis:**
  ```bash
  ./go-stats-generator analyze ./internal/analyzer --format json
  ./go-stats-generator analyze ./internal/metrics --format json
  ./go-stats-generator analyze ./cmd/go-stats-generator --format json
  ```
  - Target functions exceeding thresholds:
    * Overall complexity > 10.0 
    * Line count > 30 (code lines only)
    * Cyclomatic complexity > 10
    * Nesting depth > 3
  - Pay special attention to AST processing, metric calculation, and CLI orchestration functions

### Phase 2: Guided Self-Refactoring Implementation
1. **Follow Tool's Own Recommendations:**
  - Use `go-stats-generator`'s extraction suggestions as the primary guide for refactoring itself
  - Target each suggested extraction point in the analysis engine
  - Maintain the tool's precision in metric calculation during refactoring

2. **Create Focused Extractions in Core Modules:**
  - Extract complex AST traversal logic into focused functions
  - Separate metric calculation algorithms from orchestration logic
  - Split CLI command handlers into validation, execution, and formatting phases
  - Name functions using verb-first camelCase (e.g., `calculateCyclomaticComplexity`, `parseASTNode`)
  - Target metrics per extracted function:
    * <20 lines of code
    * Cyclomatic complexity <8
    * Clear single responsibility
  - Add comprehensive GoDoc comments for all public and complex private functions

3. **Bug Discovery and Fixing Protocol:**
  - During function breakdown, test each extracted component independently
  - If bugs are discovered in metric calculations or AST processing:
    * Document the bug with specific test cases
    * Fix the bug while maintaining API compatibility
    * Add regression tests to prevent reoccurrence
  - Validate that refactoring preserves exact metric calculation accuracy

### Phase 3: Self-Validation and Bug Testing
1. **Measure Self-Improvements:**
  ```bash
  ./go-stats-generator diff self-baseline.json refactored.json
  ```
  - Verify target functions show significant complexity reduction (>50%)
  - Ensure no regressions in the tool's own analysis accuracy
  - Validate that the tool can still analyze itself correctly after refactoring

2. **Bug Regression Testing:**
  ```bash
  go test ./... -v
  ./go-stats-generator analyze . --format json --output validation.json
  # Compare against known good baseline to ensure accuracy preserved
  ```

### Phase 4: Repository Optimization and Quality Verification
1. **Validate Self-Analysis Metrics:**
  - Original complex functions reduced by ≥50%
  - All extracted functions meet target thresholds
  - No complexity regressions detected by self-analysis
  - Tool maintains same analysis accuracy before/after refactoring

2. **Comprehensive Quality Assurance:**
  - All existing tests pass without modification
  - Tool produces identical analysis results on test codebases
  - Performance characteristics maintained or improved
  - Memory usage patterns unchanged

3. **Documentation and Bug Report Updates:**
  - Update any discovered inaccuracies in metric calculations
  - Document any algorithmic improvements made during refactoring
  - Update complexity calculation documentation if formulas were refined

## OUTPUT FORMAT:

Structure your response as:

### 1. Self-Analysis Summary
```
go-stats-generator self-analysis identified target functions:
- Function: [name] in [file]
- Current complexity: [score]  
- Module: [internal/analyzer|internal/metrics|cmd]
- Key issues: [cyclomatic/nesting/lines breakdown]
- Bugs discovered: [list any calculation errors or logic issues found]
```

### 2. Complete Refactored Files
Present the fully refactored Go files with:
- Original complex functions reduced to coordination logic
- Extracted private functions with comprehensive GoDoc
- Any bug fixes implemented during breakdown
- Standard Go formatting maintained

### 3. Bug Fixes and Improvements
```
Bugs discovered and fixed during breakdown:
- [Bug description]: [Fix implemented]
- [Performance issue]: [Optimization applied]
- [Edge case]: [Handling added]
```

### 4. Self-Improvement Validation
```
go-stats-generator differential self-analysis results:
- Original function: [old_score] → [new_score] ([improvement_%])
- New functions: [list with complexities]
- Regressions: [count]
- Analysis accuracy: [preserved/improved]
- Overall repository quality improvement: [score]
```

## COMPLEXITY REFERENCE (go-stats-generator self-calculation):
```
Overall Complexity = cyclomatic + (nesting_depth * 0.5) + (cognitive * 0.3)
Signature Complexity = (params * 0.5) + (returns * 0.3) + (interfaces * 0.8) + (generics * 1.5) + variadic_penalty
Refactoring Threshold = Overall Complexity > 10.0 OR Lines > 30 OR Cyclomatic > 10
```

## EXAMPLE SELF-ANALYSIS WORKFLOW:
```bash
$ ./go-stats-generator analyze .
=== TOP COMPLEX FUNCTIONS IN go-stats-generator ===
1. analyzeFunction (internal/analyzer/function.go): 28.7 complexity
  - Lines: 67 code lines 
  - Cyclomatic: 22
  - Nesting: 5
  - Bug: Incorrect handling of generic type parameters in signature complexity

2. processASTNode (internal/metrics/ast.go): 24.3 complexity
  - Lines: 52 code lines
  - Cyclomatic: 19
  - Nesting: 4

$ ./go-stats-generator diff self-baseline.json refactored.json 
=== SELF-IMPROVEMENT SUMMARY ===
MAJOR IMPROVEMENTS:
- analyzeFunction: 28.7 → 8.4 (71% reduction) ✓
- processASTNode: 24.3 → 7.9 (67% reduction) ✓

EXTRACTED FUNCTIONS:
  calculateSignatureComplexity: 6.2 complexity ✓ (bug fixed: generic multiplier corrected)
  validateFunctionParameters: 4.8 complexity ✓
  computeCyclomaticScore: 7.1 complexity ✓
  
BUGS FIXED: 1 (generic type parameter calculation)
QUALITY SCORE: 94/100 (+18 improvement)
REGRESSIONS: 0
```

This dogfooding approach ensures `go-stats-generator` maintains high code quality standards while validating its own analysis capabilities and discovering/fixing any bugs in its core algorithms.
