# TASK DESCRIPTION:
Perform a comprehensive functional audit of a Go codebase using `go-stats-generator` complexity analysis capabilities to identify all discrepancies between documented functionality (README.md) and actual implementation, focusing on bugs, missing features, and functional misalignments.

## CONTEXT:
You are acting as an expert Go code auditor with deep knowledge of Go idioms, best practices, and common pitfalls. Your analysis will be used by development teams to identify and prioritize fixes before deployment. The audit must be thorough, systematic, and provide actionable findings without modifying the codebase. You will leverage the `go-stats-generator` tool to assess code complexity and structure as part of your analysis methodology.

## INSTRUCTIONS:
1. **Initial Documentation Review**
   - Read and thoroughly understand the README.md file
   - Extract all functional requirements, features, and behavioral specifications
   - Note any ambiguous or incomplete documentation

2. **Complexity Analysis with go-stats-generator**
   - Run `go-stats-generator analyze . --format json --output complexity-report.json` to generate baseline complexity metrics
   - Analyze function length violations and cyclomatic complexity patterns
   - Identify high-complexity functions that may harbor bugs
   - Review struct complexity metrics and member categorization
   - Examine pattern detection results for architectural consistency
   - Use complexity data to prioritize which code areas need deeper manual review

3. **Dependency-Based File Analysis Order**
   - Map import dependencies across all .go files using both the complexity report and manual inspection
   - Categorize files by dependency level:
     * Level 0: No internal imports (utilities, constants, pure functions)
     * Level 1: Import only Level 0 files
     * Level N: Import files from Level N-1 or below
   - Audit files strictly in ascending level order (0→1→2...)
   - Cross-reference complexity metrics with dependency depth to identify architectural issues

4. **Systematic Code Analysis Enhanced by Metrics**
   - Begin with Level 0 files to establish baseline correctness
   - For each file level, verify all functions before proceeding
   - Use complexity metrics from go-stats-generator to prioritize review of high-complexity functions
   - Trace execution paths for each documented feature
   - Check for consistency between function signatures and documentation
   - Identify unreachable code or dead endpoints
   - Pay special attention to functions flagged as high complexity or violating common standards

5. **Issue Categorization with Complexity Context**
   Classify each finding into one of these categories:
   - **CRITICAL BUG**: Causes incorrect behavior, data corruption, or crashes
   - **FUNCTIONAL MISMATCH**: Implementation differs from documented behavior
   - **MISSING FEATURE**: Documented functionality not implemented
   - **EDGE CASE BUG**: Fails under specific conditions not covered in normal flow
   - **PERFORMANCE ISSUE**: Significant inefficiency affecting usability
   - **COMPLEXITY CONCERN**: High complexity functions that may need refactoring for maintainability
   - **ARCHITECTURAL INCONSISTENCY**: Pattern detection reveals design violations

6. **Analysis Depth Requirements**
   - Test boundary conditions for all inputs, especially in high-complexity functions identified by go-stats-generator
   - Verify concurrent operation safety (goroutines, channels, mutexes)
   - Check resource management (file handles, network connections, memory)
   - Validate error propagation and handling
   - Examine integration points between modules
   - Cross-reference manual findings with automated complexity analysis results

## FORMATTING REQUIREMENTS:
Present each finding in a separate ~~~~ fenced codeblock using this structure:

## AUDIT SUMMARY
[Provide summary in first codeblock with totals for each issue category, including relevant complexity statistics from go-stats-generator analysis]

## COMPLEXITY ANALYSIS OVERVIEW
[Second codeblock showing key insights from go-stats-generator output that informed the audit]

## DETAILED FINDINGS
[Each finding in its own codeblock with the following format:]

### [CATEGORY]: [Brief Issue Title]
**File:** [filename.go:line_numbers]
**Severity:** [High/Medium/Low]
**Complexity Context:** [Relevant metrics from go-stats-generator if applicable]
**Description:** [Detailed explanation of the issue]
**Expected Behavior:** [What the documentation specifies]
**Actual Behavior:** [What the code actually does]
**Impact:** [Consequences of this issue]
**Reproduction:** [Steps or conditions to trigger the issue]
**Code Reference:**
```go
[Relevant code snippet]
```

## QUALITY CHECKS:
1. Confirm go-stats-generator analysis completed and results reviewed before manual examination
2. Verify complexity metrics guided prioritization of code review areas
3. Confirm dependency analysis completed before code examination
4. Verify audit progression followed dependency levels strictly
5. Ensure all findings include specific file references and line numbers
6. Validate that each bug explanation includes reproduction steps
7. Check that severity ratings align with actual impact on functionality
8. Confirm complexity insights enhanced the quality of manual review
9. Confirm no code modifications were suggested (analysis only)

## EXAMPLES:
Example finding format incorporating complexity analysis:

### COMPLEXITY CONCERN: High Complexity Function May Hide Logic Errors
**File:** internal/processor/handler.go:125-178
**Severity:** Medium
**Complexity Context:** Function length: 54 lines, Cyclomatic complexity: 15 (flagged by go-stats-generator as high complexity)
**Description:** The processUserRequest function has high cyclomatic complexity and contains multiple nested conditions that make error handling paths difficult to verify against documentation
**Expected Behavior:** User requests should be processed with clear error responses for invalid inputs
**Actual Behavior:** Complex branching logic may allow some invalid requests to be processed incorrectly
**Impact:** Potential for incorrect processing of edge cases due to complex control flow
**Reproduction:** Submit requests with boundary value combinations to test all code paths
**Code Reference:**
```go
func processUserRequest(req *UserRequest) (*Response, error) {
    // 54 lines with multiple nested if-else chains
    if req != nil {
        if req.Type == "premium" {
            if req.User.IsPaid {
                // Multiple levels of nesting continue...
            }
        }
    }
    // Complex logic continues with potential edge case issues
}
```

### CRITICAL BUG: Nil Pointer Dereference in High-Traffic Function
**File:** api/auth/validator.go:45-52
**Severity:** High
**Complexity Context:** Function complexity: 8 (moderate), but handles critical authentication logic
**Description:** The validateUser function does not check if the user object is nil before accessing its properties, despite being flagged as a critical authentication path
**Expected Behavior:** Function should return authentication error for invalid/missing user data
**Actual Behavior:** Application crashes with nil pointer dereference panic
**Impact:** Denial of service vulnerability; any invalid authentication attempt crashes the server
**Reproduction:** Submit authentication request with malformed or empty user data
**Code Reference:**
```go
func validateUser(user *User) error {
    // Missing nil check here - identified during complexity-guided review
    if user.IsActive && user.HasPermission("read") { // Panic occurs here when user is nil
        return nil
    }
    return errors.New("user validation failed")
}