# TASK DESCRIPTION:
Perform a functional breakdown analysis on a single Go file using advanced complexity metrics from `go-stats-generator`, refactoring functions that exceed professional quality thresholds into smaller, well-named private functions that improve code readability and maintainability. Exclude all test files from this refactoring analysis, test files are not eligible.

## PREREQUISITES:
Before starting the analysis, install and use `go-stats-generator` to identify complex functions:

### Installation:
```bash
go install github.com/opd-ai/go-stats-generator@latest
```

### Usage for Complexity Analysis:
```bash
# Analyze the target codebase to identify complex functions
go-stats-generator analyze . --format json --output complexity-report.json

# Or use console output to quickly identify functions exceeding thresholds
go-stats-generator analyze . --max-function-length 30 --max-complexity 10
```

The tool will provide detailed metrics including:
- Function line counts (excluding comments, blank lines, braces)
- Cyclomatic complexity scores
- Nesting depth measurements
- Signature complexity analysis
- Overall complexity scores with weighted calculations

Use this output to prioritize which functions require refactoring based on actual measured complexity rather than manual assessment.

## CONTEXT:
You are acting as an automated Go code auditor specializing in functional decomposition using enterprise-grade complexity analysis powered by `go-stats-generator`. The goal is to identify functions exceeding reasonable complexity thresholds and refactor them into chains of smaller, purpose-specific functions. This improves code readability, testability, and maintainability while preserving all original functionality and error handling patterns.

Target metrics (following go-stats-generator analysis patterns):
- Functions exceeding 30 lines of code (excluding comments, blank lines, and braces)
- Functions with cyclomatic complexity > 10
- Functions with nesting depth > 3
- Functions with overall complexity score > 15.0 (calculated as: cyclomatic + nesting_depth*0.5 + cognitive*0.3)
- Functions with signature complexity > 8.0 (based on parameter count, return values, interface usage, generics)
- Functions performing multiple distinct logical operations

## INSTRUCTIONS:
1. **Initial Analysis Phase:**
   - Run `go-stats-generator analyze .` on the target codebase to identify functions exceeding complexity thresholds
   - Review the console output or JSON report to locate functions with:
     * Line counts > 30 (excluding comments, blank lines, braces)
     * Cyclomatic complexity > 10
     * Overall complexity score > 15.0
   - Use the tool's detailed metrics to guide refactoring decisions

2. **Complexity Analysis Phase:**
   - Calculate precise line counts excluding comments, blank lines, and function braces
   - Compute cyclomatic complexity by counting decision points (if, for, range, switch, select, case clauses)
   - Measure nesting depth of control structures
   - Evaluate signature complexity considering:
     * Parameter count (each param adds 0.5 to complexity)
     * Return count (each return adds 0.3 to complexity)
     * Interface parameters (each adds 0.8 to complexity)
     * Variadic parameters (adds 1.0 to complexity)
     * Generic parameters (each adds 1.5 to complexity)
   - Calculate overall complexity score using weighted formula

3. **File Selection Phase:**
   - Based on `go-stats-generator` output, identify the file containing the most complex function
   - Prioritize functions with highest overall complexity scores first
   - Select exactly one file for refactoring
   - If no functions exceed thresholds according to the tool, skip to step 8

3. **Function Analysis Phase:**
   - Identify the highest complexity function from the `go-stats-generator` report
   - Map distinct logical tasks within this function by identifying:
     * Initialization/setup blocks
     * Input validation/preprocessing steps
     * Core business logic segments with single responsibilities
     * Error handling patterns and cleanup operations
     * Loop bodies performing discrete operations
     * Conditional blocks with substantial logic
     * Resource management (defer statements, file operations)
     * Data transformation steps

4. **Refactoring Design Phase:**
   - Plan extraction of each identified task into a private function
   - Design function signatures that:
     * Accept only necessary parameters (target <5 parameters per function)
     * Return appropriate values including error types
     * Maintain the same error handling patterns as the original
     * Minimize signature complexity score
   - Ensure extracted functions will be attached to the correct receiver (if methods)
   - Verify that variable scoping remains correct and no unintended closures are created

5. **Implementation Phase:**
   - Extract each identified task into a private function following these conventions:
     * Use camelCase starting with lowercase letter
     * Begin with a verb describing the action
     * Be specific about the function's purpose
     * Target <30 lines per extracted function
     * Target cyclomatic complexity <10 per extracted function
     * Examples: `validateUserInput()`, `calculateTaxRate()`, `buildResponseHeader()`
   - Add precise GoDoc comments above each new function:
     * Start with the function name
     * Describe what the function does and returns
     * Include error conditions if applicable
     * Example: `// validateUserInput checks that all required fields are present and valid.`
   - Preserve all error handling:
     * Propagate errors up the call chain with proper wrapping
     * Maintain original error context and annotations
     * Keep defer statements in appropriate scope
   - Update the original function to call the new private functions in sequence

6. **Verification Phase:**
   - Run `go-stats-generator analyze .` again to confirm complexity reduction
   - Verify that the refactored function now meets all threshold requirements
   - Confirm functional equivalence:
     * All original logic is preserved
     * Error handling paths remain identical
     * Return values are unchanged
   - Validate Go best practices:
     * Functions follow single responsibility principle
     * Error handling follows Go idioms
     * Variable scoping is correct
     * No unnecessary global state access

7. **Completion Phase:**
   - If refactoring was performed: Output message "Refactor complete: [filename] - Reduced complexity from [original_score] to [new_score] across [n] extracted functions."
   - If no refactoring needed: Output message "Refactor complete: No functions in the codebase exceed professional complexity thresholds according to go-stats-generator analysis."

## COMPLEXITY CALCULATION REFERENCE:
```
Cyclomatic Complexity = 1 + count(if, for, range, switch, select, case, comm clauses)
Nesting Depth = maximum depth of nested block statements
Signature Complexity = (parameters * 0.5) + (returns * 0.3) + (interfaces * 0.8) + (variadic ? 1.0 : 0) + (generics * 1.5)
Overall Complexity = cyclomatic + (nesting_depth * 0.5) + (cognitive * 0.3)
```

## FORMATTING REQUIREMENTS:
Present the refactored code using:
- Standard Go formatting (as produced by `go fmt`)
- Clear separation between the refactored main function and extracted helper functions
- Consistent indentation and spacing
- Proper placement of GoDoc comments according to Go conventions
- Extracted functions placed immediately after the main function

Structure your response as:
1. Complexity analysis summary showing before/after metrics
2. The complete refactored file with all changes
3. Completion message with quantified improvements

## QUALITY CHECKS:
Before presenting the refactored code, verify:
- Original function's overall complexity score is reduced below 15.0
- Each extracted function has complexity score <10.0
- Line counts are accurate (excluding comments/blanks/braces)
- The refactored code compiles without errors
- All tests that passed before refactoring still pass
- No business logic has been altered
- Error handling is preserved exactly as in the original
- Function names accurately describe their single responsibility
- GoDoc comments follow Go documentation standards
- No code duplication has been introduced
- Variable scoping is correct with no unintended memory leaks

## EXAMPLES:
Example of a function requiring breakdown (complexity analysis):
```go
func processComplexOrder(order Order, config Config, validators []Validator) (*Result, error) {
    // 67 lines total (45 code lines excluding comments/blanks/braces)
    // Cyclomatic complexity: 18 (multiple if/for/switch statements)
    // Nesting depth: 4 (deeply nested conditionals)
    // Signature complexity: 6.3 (3 params * 0.5 + 2 returns * 0.3 + 1 interface * 0.8)
    // Overall complexity: 18 + (4 * 0.5) + (18 * 0.3) = 25.4 (exceeds 15.0 threshold)
}
```

After refactoring (complexity reduced):
```go
func processComplexOrder(order Order, config Config, validators []Validator) (*Result, error) {
    // Overall complexity: 4.9 (within threshold)
    if err := validateOrderData(order, validators); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    pricing, err := calculateOrderPricing(order, config)
    if err != nil {
        return nil, fmt.Errorf("pricing calculation failed: %w", err)
    }
    
    result, err := finalizeOrderProcessing(order, pricing, config)
    if err != nil {
        return nil, fmt.Errorf("order finalization failed: %w", err)
    }
    
    return result, nil
}

// validateOrderData ensures all required order fields are present and pass validation rules.
// It returns an error if any validation rule is violated.
func validateOrderData(order Order, validators []Validator) error {
    // Complexity: 6.2 (within threshold)
    // validation logic extracted here
}

// calculateOrderPricing computes the final price including tax, discounts, and fees.
// It returns the pricing details or an error if calculation fails.
func calculateOrderPricing(order Order, config Config) (*PricingDetails, error) {
    // Complexity: 8.1 (within threshold)
    // pricing calculation logic extracted here
}

// finalizeOrderProcessing completes the order processing workflow.
// It returns the final result or an error if any step fails.
func finalizeOrderProcessing(order Order, pricing *PricingDetails, config Config) (*Result, error) {
    // Complexity: 7.3 (within threshold)
    // finalization logic extracted here
}