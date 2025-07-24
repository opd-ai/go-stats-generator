# PROJECT: Go Test Failure Analysis and Categorized Resolution System

## OBJECTIVE:
Execute systematic analysis of Go test failures to determine the correct solution category through evidence-based classification, then implement precisely targeted fixes that address root causes without introducing regressions.

## TECHNICAL SPECIFICATIONS:
- Language: Go
- Type: Test Analysis and Debugging System
- Key Features: Three-tier solution categorization, root cause analysis, targeted remediation
- Performance Requirements: Complete analysis within 3 minutes, 100% accuracy in solution categorization

## ARCHITECTURE GUIDELINES:
### Required Go Tools:
| Tool | Command | Purpose |
|------|---------|---------|
| Test Runner | `go test -v ./...` | Execute all tests with verbose output |
| Coverage | `go test -cover -coverprofile=coverage.out` | Analyze test coverage |
| Race Detection | `go test -race ./...` | Detect concurrency issues |
| Benchmarking | `go test -bench=. -benchmem` | Performance validation |

### Critical Decision Matrix:
```go
// SOLUTION CATEGORY CLASSIFICATION RULES:
// Apply these rules in order - stop at first match

// CATEGORY 1: IMPLEMENTATION CODE FIX
// Indicators:
// - Test logic follows Go testing best practices
// - Test assertions match documented requirements
// - Implementation fails to satisfy business rules
// - Code doesn't handle expected input/output correctly

// CATEGORY 2: TEST SPECIFICATION FIX  
// Indicators:
// - Implementation satisfies business requirements
// - Test uses incorrect expected values
// - Test setup creates invalid conditions
// - Test assertions contradict actual requirements

// CATEGORY 3: NEGATIVE TEST CONVERSION
// Indicators:
// - Test expects success for invalid input
// - Missing validation for edge cases
// - Test should verify error handling but doesn't
// - Security boundaries not properly tested
```

## IMPLEMENTATION PHASES:

### Phase 1: Comprehensive Test Execution
**Mandatory Steps:**
1. Execute: `go test -v -race -cover ./...` 
2. Capture complete output including panic stack traces
3. Document each failure with exact error messages and file locations
4. Rank failures by severity: panics > assertion failures > timeouts

**Acceptance Criteria:**
- [ ] All test failures documented with specific `testing.T` error messages
- [ ] File paths and line numbers captured for each failure
- [ ] Highest-priority failure selected with clear justification

### Phase 2: Solution Category Determination (CRITICAL PHASE)
**MANDATORY CLASSIFICATION PROCESS:**

#### Step 2A: Evidence Collection
```go
// For each failing test, document:
// 1. Test function name and file location
// 2. Expected vs actual values from assertion failures  
// 3. Business requirement being tested
// 4. Implementation code being validated
```

#### Step 2B: Apply Decision Tree
```
START → Is the test logic sound and following Go conventions?
├─ NO → CATEGORY 2: Test Specification Fix
└─ YES → Does implementation satisfy business requirements?
    ├─ YES → Should this test verify error conditions instead?
    │   ├─ YES → CATEGORY 3: Negative Test Conversion
    │   └─ NO → CATEGORY 2: Test Specification Fix  
    └─ NO → CATEGORY 1: Implementation Code Fix
```

#### Step 2C: Validation Questions
**Before proceeding, answer ALL questions:**
1. **For Implementation Fix**: "Does the failing code violate documented behavior or business logic?"
2. **For Test Fix**: "Are the test assertions incorrect given the actual requirements?"
3. **For Negative Test**: "Should this test validate failure scenarios instead of success?"

**Acceptance Criteria:**
- [ ] Solution category determined using decision tree
- [ ] Evidence supports categorization choice
- [ ] All validation questions answered with specific examples

### Phase 3: Targeted Resolution
**Category-Specific Implementation:**

#### CATEGORY 1: Implementation Code Fix
```go
// Rules:
// - Modify only the implementation code
// - Preserve existing function signatures
// - Add error handling if missing
// - Fix business logic to match requirements
// - Do NOT change test assertions
```

#### CATEGORY 2: Test Specification Fix
```go
// Rules:
// - Modify only the test code
// - Update expected values to match correct behavior
// - Fix test setup or teardown if incorrect
// - Correct assertion logic
// - Do NOT change implementation code
```

#### CATEGORY 3: Negative Test Conversion
```go
// Rules:
// - Convert test to validate error conditions
// - Add input validation tests
// - Test boundary conditions and edge cases
// - Verify proper error types returned
// - Ensure implementation handles invalid input gracefully
```

## CODE STANDARDS:
### Go-Specific Requirements:
```go
// Simplicity Rules for Go:
// - Use standard library testing package
// - Follow table-driven test patterns
// - Prefer explicit error checking over assertions
// - Use testify/assert only for complex comparisons
// - Keep test functions under 50 lines
```

### Error Handling Standards:
```go
// Required patterns:
if err != nil {
    t.Fatalf("unexpected error: %v", err)
}

// For negative tests:
if err == nil {
    t.Fatal("expected error but got none")
}
```

## FORMATTING REQUIREMENTS:
Structure analysis using exactly these sections:

```markdown
## TEST EXECUTION RESULTS
- Command: `go test -v -race ./...`
- Failed Tests: [count] 
- Selected Failure: [test_name] in [file:line]
- Error Message: [exact output]
- Stack Trace: [if panic occurred]

## SOLUTION CATEGORY DETERMINATION
- Evidence Analysis: [specific details supporting category choice]
- Decision Tree Result: [Category 1/2/3]
- Validation Questions: [answers to all three validation questions]
- Justification: [why this category was selected over others]

## TARGETED CODE FIX
- Category: [1: Implementation / 2: Test Spec / 3: Negative Test]
- Files Modified: [exact file paths]
- Changes Made: [before/after code blocks with line numbers]
- Verification: `go test -v [specific_test]` output
```

## QUALITY CHECKS:
### Pre-Fix Validation:
1. **Category Accuracy**: Does evidence clearly support the chosen solution category?
2. **Scope Verification**: Will the fix address exactly one test failure?
3. **Impact Assessment**: Are there any dependency risks with other tests?

### Post-Fix Validation:
1. **Resolution Confirmation**: Does `go test -v [target_test]` pass?
2. **Regression Testing**: Does `go test ./...` show no new failures?
3. **Coverage Maintenance**: Has test coverage been preserved or improved?

## SUCCESS CRITERIA:
- [ ] Correct solution category identified through systematic analysis
- [ ] Evidence-based justification provided for category selection
- [ ] Exactly one test failure resolved
- [ ] No regressions introduced in test suite
- [ ] Code changes follow Go conventions and project patterns
- [ ] Fix is minimal and targeted to root cause

## ANTI-PATTERNS TO AVOID:
```go
// ❌ Don't do this:
// - Changing both implementation AND test code
// - Skipping category determination phase
// - Applying fixes without understanding root cause
// - Converting working positive tests to negative tests

// ✅ Do this instead:
// - Follow the decision tree systematically
// - Document evidence for category choice
// - Apply single-category fixes only
// - Verify the fix resolves the specific failure
```