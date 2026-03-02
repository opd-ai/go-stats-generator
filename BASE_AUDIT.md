# TASK DESCRIPTION:
Perform a data-driven comprehensive functional audit of a Go codebase to identify all discrepancies between documented functionality (README.md) and actual implementation, using `go-stats-generator` as the primary analysis engine for evidence gathering. Place your findings into AUDIT.md at the base of the package. This package has been audited many times already, carefully ensure that your concerns apply to the most recent version of the code.

When results are ambiguous, such as multiple functions exceeding thresholds or overlapping issues, always prioritize by severity (CRITICAL > FUNCTIONAL MISMATCH > MISSING FEATURE > EDGE CASE > PERFORMANCE), then by complexity score descending.

## CONSTRAINT:

Use only `go-stats-generator` and manual code review for your analysis. You are absolutely forbidden from modifying any code — this is a report-generation audit only. No other code analysis tools may be used.

## PREREQUISITES:
**Minimum Required Version:** `go-stats-generator` v1.0.0 or higher
Install and configure `go-stats-generator` for comprehensive audit evidence gathering:

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
go-stats-generator analyze --format json | jq .functions
# Check if it is installed
which jq
# If it is not, install it
sudo apt-get install jq
```

### Required Analysis Workflow:
```bash
# Phase 1: Gather quantitative evidence with go-stats-generator
go-stats-generator analyze . --max-complexity 10 --max-function-length 30 --min-doc-coverage 0.7 --skip-tests --format json --output audit-baseline.json
go-stats-generator analyze . --max-complexity 10 --max-function-length 30 --min-doc-coverage 0.7 --skip-tests

# Phase 2: Extract high-risk areas for focused manual review
cat audit-baseline.json | jq '.functions[] | select(.lines > 50 or .cyclomatic > 15)'
cat audit-baseline.json | jq '.documentation'
cat audit-baseline.json | jq '.packages[]'

# Phase 3: Cross-reference README claims against metric evidence
# Use the quantitative data to validate or refute each documented feature
```

## CONTEXT:
You are an expert Go code auditor using `go-stats-generator` as your primary analysis engine for enterprise-grade codebase auditing. The tool provides precise metrics — complexity hotspots, documentation coverage, naming analysis, and package dependency analysis — that serve as quantitative evidence for your audit findings. Use `go-stats-generator` output to identify high-risk areas for focused manual review, then cross-reference README.md claims against the metric evidence. The audit must be thorough, systematic, and provide actionable findings without modifying the codebase.

## INSTRUCTIONS:

### Phase 1: Evidence Gathering with go-stats-generator
1. **Run Baseline Analysis:**
  ```bash
  go-stats-generator analyze . --skip-tests --format json --output audit-baseline.json
  go-stats-generator analyze . --skip-tests
  ```
  - Capture complexity metrics for all functions, structs, packages, and interfaces
  - Identify functions exceeding HIGH RISK thresholds (>50 lines OR cyclomatic >15)
  - Record documentation coverage metrics across the codebase
  - Note package dependency structure and potential circular dependencies

2. **Extract High-Risk Targets:**
  ```bash
  # Identify complexity hotspots most likely to contain bugs
  cat audit-baseline.json | jq '[.functions[] | select(.lines > 50 or .cyclomatic > 15)] | sort_by(-.cyclomatic)'
  # Check documentation coverage gaps
  cat audit-baseline.json | jq '.documentation'
  # Review naming consistency issues
  cat audit-baseline.json | jq '.naming'
  # Examine package dependencies for architectural issues
  cat audit-baseline.json | jq '.packages[] | {name, dependencies, cohesion, coupling}'
  ```

3. **Prioritize Audit Focus:**
  From the baseline analysis, prioritize manual review in order:
  - **Critical Risk (complexity >20 OR lines >80):** Review first — highest bug probability
  - **High Risk (complexity >15 OR lines >50):** Review second
  - **Medium Risk (complexity >10 OR lines >30):** Review if time permits
  - Functions with low documentation coverage in high-risk areas get elevated priority

### Phase 2: Documentation Cross-Reference
1. **Initial Documentation Review:**
   - Read and thoroughly understand the README.md file
   - Extract all functional requirements, features, and behavioral specifications
   - Note any ambiguous or incomplete documentation

2. **Metric-Backed Validation:**
   - For each documented feature, locate the implementing functions in `go-stats-generator` output
   - Cross-reference claimed behaviors against actual function signatures and complexity
   - Use documentation coverage metrics to identify undocumented public APIs
   - Use package dependency analysis to verify claimed architectural patterns

### Phase 3: Dependency-Based Deep Review
1. **Map Import Dependencies:**
   - Use `go-stats-generator` package analysis to establish dependency levels:
     * Level 0: No internal imports (utilities, constants, pure functions)
     * Level 1: Import only Level 0 files
     * Level N: Import files from Level N-1 or below
   - Audit files strictly in ascending level order (0→1→2...)

2. **Systematic Code Analysis (guided by metrics):**
   - Begin with Level 0 files to establish baseline correctness
   - Focus manual review time on functions flagged as HIGH RISK by `go-stats-generator`
   - For each file level, verify all functions before proceeding
   - Trace execution paths for each documented feature
   - Check for consistency between function signatures and documentation
   - Identify unreachable code or dead endpoints

3. **Analysis Depth Requirements:**
   - Test boundary conditions for all inputs
   - Verify concurrent operation safety (use `go-stats-generator` concurrency analysis: goroutines, channels, sync primitives)
   - Check resource management (file handles, network connections, memory)
   - Validate error propagation and handling
   - Examine integration points between modules

### Phase 4: Issue Classification and Reporting
1. **Issue Categorization:**
   Classify each finding into one of these categories:
   - **CRITICAL BUG**: Causes incorrect behavior, data corruption, or crashes
   - **FUNCTIONAL MISMATCH**: Implementation differs from documented behavior
   - **MISSING FEATURE**: Documented functionality not implemented
   - **EDGE CASE BUG**: Fails under specific conditions not covered in normal flow
   - **PERFORMANCE ISSUE**: Significant inefficiency affecting usability

2. **Attach Metric Evidence:**
   - Include `go-stats-generator` metrics that flagged the area for review
   - Reference specific complexity scores, doc coverage gaps, or dependency issues
   - Correlate finding severity with quantitative risk indicators

## OUTPUT FORMAT:

Structure your response as:

### 1. Audit Evidence Summary
```
go-stats-generator baseline analysis:
  Total Functions Analyzed: [n]
  HIGH RISK Functions (>50 lines OR cyclomatic >15): [n]
  Documentation Coverage: [x.xx]%
  Functions Below --min-doc-coverage 0.7: [n]
  Package Dependency Issues: [n]
  Naming Violations: [n]
  Concurrency Patterns Detected: [n]

High-risk audit targets:
1. Function: [name] in [file]
   - Lines: [n], Cyclomatic: [n], Complexity: [score]
   - Doc Coverage: [present/missing]
   - Risk Level: [Critical/High/Medium]

2. Function: [name] in [file]
   - Lines: [n], Cyclomatic: [n], Complexity: [score]
   - Doc Coverage: [present/missing]
   - Risk Level: [Critical/High/Medium]

... (continue for all HIGH RISK functions)
```

### 2. Audit Summary
```
AUDIT RESULTS:
  CRITICAL BUG:        [n] findings
  FUNCTIONAL MISMATCH: [n] findings
  MISSING FEATURE:     [n] findings
  EDGE CASE BUG:       [n] findings
  PERFORMANCE ISSUE:   [n] findings
  TOTAL:               [n] findings
```

### 3. Detailed Findings
Each finding in its own section with the following format:

#### [CATEGORY]: [Brief Issue Title]
**File:** [filename.go:line_numbers]
**Severity:** [High/Medium/Low]
**Metric Evidence:** [go-stats-generator metrics that flagged this area]
**Description:** [Detailed explanation of the issue]
**Expected Behavior:** [What the documentation specifies]
**Actual Behavior:** [What the code actually does]
**Impact:** [Consequences of this issue]
**Reproduction:** [Steps or conditions to trigger the issue]
**Code Reference:**
```go
[Relevant code snippet]
```

## AUDIT THRESHOLDS:
```
Risk Classification (go-stats-generator metrics):
  Critical Risk = Function Lines > 80 OR Cyclomatic > 20
  High Risk     = Function Lines > 50 OR Cyclomatic > 15
  Medium Risk   = Function Lines > 30 OR Cyclomatic > 10

Documentation Quality:
  Minimum Doc Coverage  = 0.7 (70% of exported symbols documented)
  Undocumented exports in HIGH RISK functions = automatic severity elevation

Complexity Reference (go-stats-generator calculation):
  Overall Complexity = cyclomatic + (nesting_depth * 0.5) + (cognitive * 0.3)
  Signature Complexity = (params * 0.5) + (returns * 0.3) + (interfaces * 0.8) + (generics * 1.5) + variadic_penalty

Audit Trigger = HIGH RISK functions > 0 OR Doc Coverage < 0.7 OR Package Dependency Issues > 0
  If no findings: "Audit complete: go-stats-generator baseline analysis found no functions exceeding risk thresholds and documentation coverage meets minimum standards."
```
<!-- Last verified: 2025-07-25 against function.go:calculateComplexity and calculateSignatureComplexity -->

## QUALITY CHECKS:
1. Confirm `go-stats-generator` baseline analysis completed before manual code examination
2. Verify audit focus prioritized HIGH RISK functions identified by metrics
3. Verify audit progression followed dependency levels strictly
4. Ensure all findings include specific file references and line numbers
5. Ensure all findings include `go-stats-generator` metric evidence where applicable
6. Validate that each bug explanation includes reproduction steps
7. Check that severity ratings align with both quantitative metrics and actual impact
8. Confirm no code modifications were suggested (analysis only)

## EXAMPLE WORKFLOW:
```bash
$ go-stats-generator analyze . --skip-tests --max-complexity 10 --max-function-length 30 --min-doc-coverage 0.7
=== FUNCTION ANALYSIS ===
HIGH RISK FUNCTIONS:
1. authenticateUser (auth/handler.go): 22.4 complexity
   - Lines: 65 code lines
   - Cyclomatic: 18
   - Nesting: 4
   - Doc Coverage: missing

2. processTransaction (payment/processor.go): 17.8 complexity
   - Lines: 52 code lines
   - Cyclomatic: 14
   - Nesting: 3
   - Doc Coverage: present

3. parseConfiguration (config/loader.go): 12.3 complexity
   - Lines: 38 code lines
   - Cyclomatic: 10
   - Nesting: 2
   - Doc Coverage: missing

=== DOCUMENTATION COVERAGE ===
Overall: 62.5% (below 70% threshold)
Undocumented exports: 15
Packages below threshold: 3

=== PACKAGE DEPENDENCIES ===
Circular Dependencies: 0
High Coupling Packages: 2

$ go-stats-generator analyze . --skip-tests --format json --output audit-baseline.json

$ # Extract high-risk functions for focused review
$ cat audit-baseline.json | jq '[.functions[] | select(.lines > 50 or .cyclomatic > 15)] | length'
2

$ # Check documentation gaps
$ cat audit-baseline.json | jq '.documentation'
{
  "coverage": 0.625,
  "undocumented_exports": 15,
  "packages_below_threshold": 3
}

$ # Review package dependencies
$ cat audit-baseline.json | jq '.packages[] | {name, dependencies, cohesion, coupling}'
{"name": "auth", "dependencies": ["config", "models"], "cohesion": 0.85, "coupling": 0.3}
{"name": "payment", "dependencies": ["auth", "models", "config"], "cohesion": 0.72, "coupling": 0.45}

$ # Now perform manual audit guided by these metrics...
$ # Focus on authenticateUser (Critical Risk) and processTransaction (High Risk)
$ # Cross-reference README.md claims against actual implementation
$ # Document all findings in AUDIT.md with metric evidence attached
```

This data-driven approach ensures audit findings are guided by quantitative `go-stats-generator` metrics rather than random code browsing, with high-risk areas identified systematically for focused manual review and README cross-referencing.
