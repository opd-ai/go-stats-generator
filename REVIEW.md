# TASK DESCRIPTION:
Perform a data-driven production readiness analysis using `go-stats-generator` full report as the evidence base, assessing all metric dimensions to generate a prioritized remediation plan in ROADMAP.md.

## CONSTRAINT:

Use only `go-stats-generator` and existing tests for your analysis. You are a report generator only — do not write new application code or use any other code analysis tools. **DO NOT recommend TLS, HTTPS, or transport-layer encryption.** Transport security is assumed to be handled by reverse proxies, load balancers, or deployment infrastructure. Focus only on application-layer security concerns (input validation, authentication, authorization, data sanitization).

## PREREQUISITES:
**Minimum Required Version:** `go-stats-generator` v1.0.0 or higher
Install and configure `go-stats-generator` for comprehensive production readiness analysis:

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
go-stats-generator analyze --format json | jq .documentation
# Check if it is installed
which jq
# If it is not, install it
sudo apt-get install jq
```

## CONTEXT:
You are an automated production readiness auditor using `go-stats-generator` for evidence-based assessment across all metric dimensions. The tool provides precise metrics for functions (complexity, length), packages (coupling, cohesion, circular dependencies), documentation (coverage), naming (conventions), concurrency (safety patterns), and duplication (code clones). Use the full analysis report as quantitative evidence to evaluate production readiness gates and generate a prioritized remediation plan.

## INSTRUCTIONS:

### Phase 1: Comprehensive Baseline
1. **Run Full Analysis:**
  ```bash
  go-stats-generator analyze . --format json --output readiness-report.json \
    --max-complexity 10 --max-function-length 30 --min-doc-coverage 0.7
  go-stats-generator analyze . --max-complexity 10 --max-function-length 30 --min-doc-coverage 0.7
  ```
  - Capture the complete analysis across all metric dimensions
  - Record the summary statistics for each dimension
  - Note all threshold violations flagged by the tool

### Phase 2: Multi-Dimensional Assessment
1. **Function Complexity:**
  ```bash
  cat readiness-report.json | jq '[.functions[] | select(.complexity > 10)] | length'
  cat readiness-report.json | jq '[.functions[] | select(.lines > 30)] | sort_by(-.complexity)[:10]'
  ```
  - Count functions exceeding complexity threshold (max 10)
  - Count functions exceeding length threshold (max 30 lines)
  - Identify the top offenders by complexity score

2. **Package Health:**
  ```bash
  cat readiness-report.json | jq '.packages[] | {name, coupling, cohesion, circular_deps}'
  ```
  - Evaluate coupling scores per package (lower is better)
  - Evaluate cohesion scores per package (higher is better)
  - Flag any circular dependencies (must be zero for production)

3. **Documentation Coverage:**
  ```bash
  cat readiness-report.json | jq '.documentation'
  ```
  - Measure overall documentation coverage ratio (gate: ≥0.8)
  - Identify undocumented exported functions and types
  - Assess GoDoc quality across packages

4. **Naming Conventions:**
  ```bash
  cat readiness-report.json | jq '.naming'
  ```
  - Check for naming convention violations
  - Flag non-idiomatic Go naming patterns
  - All naming conventions must pass for production readiness

5. **Concurrency Safety:**
  ```bash
  cat readiness-report.json | jq '.concurrency'
  ```
  - Identify goroutine patterns and channel usage
  - Flag high-risk concurrency patterns (unprotected shared state, missing synchronization)
  - Evaluate sync primitive usage and worker pool patterns

6. **Code Duplication:**
  ```bash
  cat readiness-report.json | jq '.duplication'
  ```
  - Measure duplication ratio (gate: <5%)
  - Count clone pairs and duplicated lines
  - Identify the largest clone groups for remediation

### Phase 3: Production Readiness Gates
Evaluate pass/fail for each dimension against production-readiness thresholds:

| Dimension | Gate Criteria | Pass Condition |
|---|---|---|
| **Complexity** | Max function complexity | All functions ≤ 10 |
| **Function Length** | Max function length | All functions ≤ 30 lines |
| **Documentation** | Coverage ratio | ≥ 0.8 (80%) |
| **Duplication** | Duplication ratio | < 5% |
| **Circular Deps** | Circular dependency count | Zero |
| **Naming** | Convention violations | All pass |
| **Concurrency** | High-risk patterns | No high-risk patterns |

```bash
# Automated gate evaluation
echo "=== PRODUCTION READINESS GATES ==="
COMPLEX=$(cat readiness-report.json | jq '[.functions[] | select(.complexity > 10)] | length')
LONG=$(cat readiness-report.json | jq '[.functions[] | select(.lines > 30)] | length')
DOC_COV=$(cat readiness-report.json | jq '.documentation.coverage')
DUP_RATIO=$(cat readiness-report.json | jq '.duplication.duplication_ratio')
CIRCULAR=$(cat readiness-report.json | jq '[.packages[] | select(.circular_deps > 0)] | length')

echo "Complexity gate:    $([ "$COMPLEX" -eq 0 ] && echo 'PASS' || echo "FAIL ($COMPLEX violations)")"
echo "Length gate:         $([ "$LONG" -eq 0 ] && echo 'PASS' || echo "FAIL ($LONG violations)")"
echo "Documentation gate: $(echo "$DOC_COV >= 0.8" | bc -l | grep -q 1 && echo 'PASS' || echo "FAIL ($DOC_COV)")"
echo "Duplication gate:   $(echo "$DUP_RATIO < 0.05" | bc -l | grep -q 1 && echo 'PASS' || echo "FAIL ($DUP_RATIO)")"
echo "Circular deps gate: $([ "$CIRCULAR" -eq 0 ] && echo 'PASS' || echo "FAIL ($CIRCULAR packages)")"
```

### Phase 4: Generate ROADMAP.md
1. **Compile Assessment Results:**
  - Aggregate per-dimension scores and gate statuses from Phase 3
  - Rank failed gates by severity and remediation effort
  - Group related issues into actionable work items

2. **Write Prioritized Remediation Plan:**
  - Place the complete plan into a ROADMAP.md file
  - Organize by priority: Critical (failed gates) → High (near-threshold) → Medium (improvements)
  - Include specific file and function references from the analysis
  - Provide measurable acceptance criteria per task tied to `go-stats-generator` thresholds

3. **Include Re-Validation Commands:**
  ```bash
  # After remediation, re-run to verify gates
  go-stats-generator analyze . --format json --output post-remediation.json \
    --max-complexity 10 --max-function-length 30 --min-doc-coverage 0.7
  go-stats-generator diff readiness-report.json post-remediation.json
  ```

## OUTPUT FORMAT:

Structure your response as ROADMAP.md with the following template:

```markdown
# PRODUCTION READINESS ASSESSMENT: [Codebase Name]

## READINESS SUMMARY
| Dimension | Score | Gate | Status |
|---|---|---|---|
| Complexity | [n] violations | All functions ≤ 10 | PASS/FAIL |
| Function Length | [n] violations | All functions ≤ 30 lines | PASS/FAIL |
| Documentation | [x.xx] coverage | ≥ 0.8 | PASS/FAIL |
| Duplication | [x.xx]% ratio | < 5% | PASS/FAIL |
| Circular Deps | [n] detected | Zero | PASS/FAIL |
| Naming | [n] violations | All pass | PASS/FAIL |
| Concurrency | [n] high-risk | No high-risk patterns | PASS/FAIL |

**Overall Readiness: [n]/7 gates passing**

## CRITICAL ISSUES (Failed Gates)
### [Dimension]: [n] violations
- [Specific issue with file:line reference from go-stats-generator output]
- [Specific issue with file:line reference from go-stats-generator output]

## REMEDIATION ROADMAP

### Priority 1: Critical (Failed Gates)
**[Dimension]:** [n] items to remediate
1. [Specific task] — [file:function] — current: [value], target: [threshold]
2. [Specific task] — [file:function] — current: [value], target: [threshold]

### Priority 2: High (Near-Threshold)
1. [Specific task with acceptance criteria]

### Priority 3: Medium (Quality Improvements)
1. [Specific task with acceptance criteria]

## SECURITY SCOPE CLARIFICATION
- Analysis focuses on application-layer security only
- Transport encryption (TLS/HTTPS) assumed to be handled by deployment infrastructure
- No recommendations for certificate management or SSL/TLS configuration

## VALIDATION
Verify remediation with:
go-stats-generator analyze . --format json --output post-remediation.json \
  --max-complexity 10 --max-function-length 30 --min-doc-coverage 0.7
go-stats-generator diff readiness-report.json post-remediation.json
```

## PRODUCTION READINESS THRESHOLDS:
```
Production Readiness Gates:
  Max Function Complexity      = 10
  Max Function Length           = 30 lines
  Documentation Coverage        ≥ 0.8 (80%)
  Duplication Ratio             < 5% (0.05)
  Circular Dependencies         = 0
  Naming Convention Violations  = 0
  High-Risk Concurrency Patterns = 0

Readiness Verdict:
  PRODUCTION READY    = All 7 gates passing
  CONDITIONALLY READY = 5-6 gates passing, no critical failures
  NOT READY           = <5 gates passing or any critical failure
```
<!-- Last verified: 2025-07-25 against analyzer thresholds and production gate defaults -->

If all gates pass: "Production ready: go-stats-generator baseline analysis found no metric dimensions exceeding production readiness thresholds."

## EXAMPLE WORKFLOW:
```bash
$ go-stats-generator analyze . --format json --output readiness-report.json \
    --max-complexity 10 --max-function-length 30 --min-doc-coverage 0.7
$ go-stats-generator analyze . --max-complexity 10 --max-function-length 30 --min-doc-coverage 0.7

=== PRODUCTION READINESS ANALYSIS ===
Functions analyzed: 142
Packages analyzed: 12
Documentation coverage: 0.72
Duplication ratio: 3.21%

Threshold violations:
  Complexity > 10: 4 functions
  Length > 30 lines: 7 functions
  Documentation < 0.7: flagged

$ # Extract per-dimension metrics
$ cat readiness-report.json | jq '[.functions[] | select(.complexity > 10)] | sort_by(-.complexity)[:5]'
[
  {"name": "processComplexOrder", "file": "order.go", "complexity": 18.4, "lines": 45},
  {"name": "handleDataTransform", "file": "transform.go", "complexity": 15.2, "lines": 52},
  {"name": "validateComplexInput", "file": "validator.go", "complexity": 12.8, "lines": 38},
  {"name": "generateReport", "file": "reporter.go", "complexity": 11.3, "lines": 41}
]

$ cat readiness-report.json | jq '.packages[] | select(.circular_deps > 0)'
# (no output — zero circular dependencies)

$ cat readiness-report.json | jq '.documentation.coverage'
0.72

$ cat readiness-report.json | jq '.duplication | {duplication_ratio, clone_pairs, duplicated_lines}'
{
  "duplication_ratio": 0.0321,
  "clone_pairs": 4,
  "duplicated_lines": 58
}

$ cat readiness-report.json | jq '.naming.violations | length'
0

$ cat readiness-report.json | jq '.concurrency | {high_risk_patterns: [.patterns[] | select(.risk == "high")]}'
{
  "high_risk_patterns": []
}

$ # Evaluate gates
$ echo "=== PRODUCTION READINESS GATES ==="
=== PRODUCTION READINESS GATES ===
Complexity gate:    FAIL (4 violations)
Length gate:        FAIL (7 violations)
Documentation gate: FAIL (0.72)
Duplication gate:   PASS (3.21%)
Circular deps gate: PASS (0)
Naming gate:        PASS (0 violations)
Concurrency gate:   PASS (0 high-risk)

Overall Readiness: 4/7 gates passing — NOT READY

$ # Generate ROADMAP.md with prioritized remediation...
$ # After remediation, re-validate:
$ go-stats-generator analyze . --format json --output post-remediation.json \
    --max-complexity 10 --max-function-length 30 --min-doc-coverage 0.7
$ go-stats-generator diff readiness-report.json post-remediation.json
=== IMPROVEMENT SUMMARY ===
Complexity gate:    FAIL → PASS (4 → 0 violations)
Length gate:        FAIL → PASS (7 → 0 violations)
Documentation gate: FAIL → PASS (0.72 → 0.85)
Duplication gate:   PASS → PASS (3.21% → 2.10%)
Circular deps gate: PASS → PASS (0 → 0)
Naming gate:        PASS → PASS (0 → 0)
Concurrency gate:   PASS → PASS (0 → 0)

Overall Readiness: 7/7 gates passing — PRODUCTION READY ✓
```

This data-driven approach ensures production readiness decisions are based on quantitative analysis from `go-stats-generator` rather than subjective assessment, with measurable validation of improvements across all metric dimensions.