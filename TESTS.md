# PROJECT: Go Test Coverage Discovery and Implementation System

## OBJECTIVE:
Create a systematic approach for AI coding agents to discover untested Go files and generate comprehensive, maintainable unit tests that achieve 80%+ line coverage while prioritizing high-value testing targets.

## TECHNICAL SPECIFICATIONS:
- Language: Go
- Type: Testing automation system
- Key Features: File discovery, complexity analysis, test generation, coverage validation
- Performance Requirements: Complete analysis in <30 seconds for projects with <500 files

## ARCHITECTURE GUIDELINES:

### Preferred Libraries:
| Library | Use Case | Justification |
|---------|----------|---------------|
| `go/parser` | AST analysis for function discovery | Standard library, comprehensive Go code parsing |
| `go/ast` | Abstract syntax tree walking | Standard library, robust code structure analysis |
| `go/token` | Source code position tracking | Standard library, precise location mapping |
| `path/filepath` | File system traversal | Standard library, cross-platform path handling |
| `testing` | Test execution framework | Standard library, Go's built-in testing infrastructure |

### Project Structure:
```
discovery-system/
├── analyzer/
│   ├── file_scanner.go      # File discovery and filtering
│   ├── complexity_calc.go   # Code complexity assessment  
│   └── metrics_collector.go # Line counts, imports, functions
├── selector/
│   ├── priority_ranker.go   # Selection algorithm implementation
│   └── criteria_validator.go # Validation of selection rules
├── generator/
│   ├── test_builder.go      # Test file generation logic
│   └── coverage_validator.go # Coverage verification
└── main.go                  # CLI interface
```

### Design Patterns:
- **Strategy Pattern**: For different selection criteria implementations
- **Builder Pattern**: For constructing test files incrementally
- **Command Pattern**: For executing analysis phases independently

## IMPLEMENTATION PHASES:

### Phase 1: Enhanced File Discovery System
**Tasks:**
- Implement recursive directory scanning with `.gitignore` respect
- Create AST-based analysis for accurate function counting
- Build import dependency graph for depth calculation
- Generate file metadata cache for performance

**Acceptance Criteria:**
- [ ] Scan 1000+ files in <5 seconds
- [ ] Exclude vendor/, internal test files, generated code
- [ ] Accurate function count including methods and closures
- [ ] Import depth calculation including transitive dependencies

### Phase 2: Intelligent Selection Algorithm
**Tasks:**
- Implement weighted scoring system for file prioritization
- Create exclusion filters for problematic file types
- Build testability assessment for complex dependencies
- Generate selection justification reports

**Acceptance Criteria:**
- [ ] Score files using 6+ weighted criteria
- [ ] Exclude files with database connections, HTTP clients, file I/O
- [ ] Identify mock-able interfaces vs concrete dependencies
- [ ] Provide detailed reasoning for selection decisions

### Phase 3: Test Generation Engine
**Tasks:**
- Generate comprehensive test suites using AST analysis
- Implement table-driven test patterns automatically
- Create mock generation for interface dependencies
- Build coverage validation and gap identification

**Acceptance Criteria:**
- [ ] Achieve 80%+ line coverage on selected files
- [ ] Generate tests for all exported functions and methods
- [ ] Include edge cases and error conditions
- [ ] Validate test compilation and execution

## CODE STANDARDS:

### File Discovery Rules:
```go
// ✅ GOOD: Comprehensive exclusion criteria
func ShouldAnalyzeFile(path string, info os.FileInfo) bool {
    // Exclude test files, generated code, vendor directories
    if strings.HasSuffix(path, "_test.go") ||
       strings.Contains(path, "/vendor/") ||
       strings.Contains(path, "/.git/") ||
       strings.Contains(path, "/testdata/") ||
       isGeneratedFile(path) {
        return false
    }
    return strings.HasSuffix(path, ".go")
}

// ❌ AVOID: Simple pattern matching
func ShouldAnalyzeFile(path string) bool {
    return strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go")
}
```

### Selection Criteria Implementation:
```go
// ✅ GOOD: Weighted scoring system
type FileScore struct {
    DependencyScore  float64 // 0.3 weight - fewer deps = higher score
    ComplexityScore  float64 // 0.25 weight - moderate complexity preferred
    SizeScore        float64 // 0.2 weight - optimal size range
    TestabilityScore float64 // 0.15 weight - interface usage, mockability
    UtilityScore     float64 // 0.1 weight - reusability assessment
    TotalScore       float64
}

// ❌ AVOID: Simple boolean filtering
func IsGoodCandidate(file GoFile) bool {
    return file.ImportCount <= 3 && file.FunctionCount <= 5
}
```

## VALIDATION CHECKLIST:

### Discovery Phase:
- [ ] All non-test `.go` files identified correctly
- [ ] Generated files and vendor code excluded
- [ ] AST parsing succeeds for all target files
- [ ] Import dependency graph complete and accurate
- [ ] File metrics calculated within 5% accuracy

### Selection Phase:
- [ ] Scoring algorithm prioritizes testable files
- [ ] Files with external dependencies ranked lower
- [ ] Utility/helper files prioritized over application logic
- [ ] Selection reasoning clearly documented
- [ ] Alternative candidates identified and ranked

### Test Generation Phase:
- [ ] All exported functions have corresponding tests
- [ ] Table-driven tests used for functions with >2 input scenarios
- [ ] Error conditions tested for all functions returning errors
- [ ] Tests compile without errors
- [ ] 80%+ line coverage achieved
- [ ] Tests pass consistently (run 10 times)
- [ ] No test interdependencies or side effects

### Code Quality:
- [ ] Test names follow Go conventions: `TestFunctionName_Scenario_Expected`
- [ ] Subtests used appropriately with `t.Run()`
- [ ] Test setup/teardown isolated to individual tests
- [ ] Mock interfaces generated for external dependencies
- [ ] Test documentation explains complex scenarios

## ENHANCED SELECTION ALGORITHM:

```markdown
SYSTEMATIC FILE DISCOVERY PROCESS:

1. **Repository Scanning**:
   - Walk directory tree respecting .gitignore
   - Filter by file extension and exclusion rules
   - Parse each file's AST for structural analysis
   - Cache results for performance optimization

2. **Dependency Analysis**:
   - Count direct imports (standard library weighted differently)
   - Analyze import types: interfaces vs concrete types
   - Calculate testability score based on dependency injection patterns
   - Flag files requiring complex mocking infrastructure

3. **Complexity Assessment**:
   - Function count (exported vs unexported)
   - Cyclomatic complexity per function
   - Nesting depth and branching factors
   - Interface usage and abstraction levels

4. **Testability Scoring**:
   Weight Factor: Description
   0.30: Low external dependencies (≤3 non-stdlib imports)
   0.25: Moderate complexity (3-7 functions, <10 cyclomatic complexity)
   0.20: Optimal size (50-200 lines excluding comments)
   0.15: High testability (interface usage, dependency injection)
   0.10: Utility classification (pure functions, data structures)

5. **Final Selection**:
   - Rank by composite score
   - Apply exclusion filters (network I/O, file operations, database access)
   - Validate selection produces meaningful tests
   - Document selection rationale with metrics
```

## COVERAGE VALIDATION REQUIREMENTS:

```go
// Mandatory coverage validation after test generation
func ValidateCoverage(packagePath string) error {
    cmd := exec.Command("go", "test", "-cover", "-coverprofile=coverage.out", packagePath)
    output, err := cmd.Output()
    if err != nil {
        return fmt.Errorf("test execution failed: %w", err)
    }
    
    coverage := extractCoveragePercentage(output)
    if coverage < 80.0 {
        return fmt.Errorf("insufficient coverage: %.1f%% (required: 80%%)", coverage)
    }
    
    return validateTestQuality(packagePath)
}

// Additional quality checks beyond coverage percentage
func validateTestQuality(packagePath string) error {
    // Check for test independence
    // Verify error case coverage
    // Validate table-driven test usage
    // Confirm no skipped tests
    return nil
}
```