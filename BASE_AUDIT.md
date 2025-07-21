# TASK DESCRIPTION:
Perform a comprehensive functional audit of a Go codebase to identify all discrepancies between documented functionality (README.md) and actual implementation, focusing on bugs, missing features, and functional misalignments. Place your findings into AUDIT.md at the base of the package.

## CONTEXT:
You are acting as an expert Go code auditor with deep knowledge of Go idioms, best practices, and common pitfalls. Your analysis will be used by development teams to identify and prioritize fixes before deployment. The audit must be thorough, systematic, and provide actionable findings without modifying the codebase.

## INSTRUCTIONS:
1. **Initial Documentation Review**
   - Read and thoroughly understand the README.md file
   - Extract all functional requirements, features, and behavioral specifications
   - Note any ambiguous or incomplete documentation

2. **Dependency-Based File Analysis Order**
   - Map import dependencies across all .go files
   - Categorize files by dependency level:
     * Level 0: No internal imports (utilities, constants, pure functions)
     * Level 1: Import only Level 0 files
     * Level N: Import files from Level N-1 or below
   - Audit files strictly in ascending level order (0→1→2...)

3. **Systematic Code Analysis**
   - Begin with Level 0 files to establish baseline correctness
   - For each file level, verify all functions before proceeding
   - Trace execution paths for each documented feature
   - Check for consistency between function signatures and documentation
   - Identify unreachable code or dead endpoints

4. **Issue Categorization**
   Classify each finding into one of these categories:
   - **CRITICAL BUG**: Causes incorrect behavior, data corruption, or crashes
   - **FUNCTIONAL MISMATCH**: Implementation differs from documented behavior
   - **MISSING FEATURE**: Documented functionality not implemented
   - **EDGE CASE BUG**: Fails under specific conditions not covered in normal flow
   - **PERFORMANCE ISSUE**: Significant inefficiency affecting usability

5. **Analysis Depth Requirements**
   - Test boundary conditions for all inputs
   - Verify concurrent operation safety (goroutines, channels, mutexes)
   - Check resource management (file handles, network connections, memory)
   - Validate error propagation and handling
   - Examine integration points between modules

## FORMATTING REQUIREMENTS:
Present each finding in a separate ~~~~ fenced codeblock using this structure:

## AUDIT SUMMARY
[Provide summary in first codeblock with totals for each issue category]

## DETAILED FINDINGS
[Each finding in its own codeblock with the following format:]

### [CATEGORY]: [Brief Issue Title]
**File:** [filename.go:line_numbers]
**Severity:** [High/Medium/Low]
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
1. Confirm dependency analysis completed before code examination
2. Verify audit progression followed dependency levels strictly
3. Ensure all findings include specific file references and line numbers
4. Validate that each bug explanation includes reproduction steps
5. Check that severity ratings align with actual impact on functionality
6. Confirm no code modifications were suggested (analysis only)

## EXAMPLES:
Example finding format:

### CRITICAL BUG: Nil Pointer Dereference in User Authentication
**File:** auth/handler.go:45-52
**Severity:** High
**Description:** The authenticateUser function does not check if the user object is nil before accessing its properties, causing a panic when invalid credentials are provided.
**Expected Behavior:** Function should return an authentication error for invalid credentials
**Actual Behavior:** Application crashes with nil pointer dereference panic
**Impact:** Denial of service vulnerability; any invalid login attempt crashes the server
**Reproduction:** Submit login request with non-existent username
**Code Reference:**
```go
func authenticateUser(username string) (*User, error) {
    user := findUserByUsername(username)
    // Missing nil check here
    if user.IsActive { // Panic occurs here when user is nil
        return user, nil
    }
}
```