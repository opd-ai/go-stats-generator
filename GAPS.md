# TASK DESCRIPTION:
Perform a data-driven implementation gap analysis to identify **all discrepancies between documented features in README.md and actual implementation** in the codebase. Use `go-stats-generator` baseline analysis (with --skip-tests) to extract comprehensive metrics, cross-reference every documented capability against the JSON output, and produce a structured gap report. This is a report-generation task — no code modifications are made.

When results are ambiguous, such as a feature partially present in metrics or a documented capability that maps to multiple code paths, always flag the gap and include evidence from both the README and the `go-stats-generator` output.

## CONSTRAINT:

Use only `go-stats-generator` and standard command-line tools (jq, grep, diff) for your analysis. You are absolutely forbidden from modifying source code or using any other code analysis tools. This is a read-only audit producing a gap report.

## PREREQUISITES:
**Minimum Required Version:** `go-stats-generator` v1.0.0 or higher
Install and configure `go-stats-generator` for comprehensive feature coverage analysis:

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

### Required Analysis Workflow:
```bash
# Phase 1: Generate comprehensive metrics baseline
go-stats-generator analyze . --skip-tests --format json --output gap-baseline.json
go-stats-generator analyze . --skip-tests --min-doc-coverage 0.7 --verbose

# Phase 2: Extract feature inventory from metrics
cat gap-baseline.json | jq '{functions: .functions | length, structs: .structs | length, packages: .packages | length, interfaces: .interfaces | length, concurrency: .concurrency, duplication: .duplication, naming: .naming, documentation: .documentation}'

# Phase 3: Cross-reference README claims against metrics
# Compare each documented feature against corresponding JSON fields

# Phase 4: Produce structured AUDIT.md gap report
```

## CONTEXT:
You are an automated Go code auditor using `go-stats-generator` for enterprise-grade documentation-vs-implementation gap detection. The application has undergone multiple previous audits and is approaching production readiness. Most obvious issues have been resolved. Your analysis must focus on precise, subtle implementation gaps that previous audits may have missed, using the tool's comprehensive JSON output as the ground truth for what the codebase actually contains. Focus on discrepancies between what README.md promises and what `go-stats-generator` metrics confirm.

## INSTRUCTIONS:

### Phase 1: Documentation Inventory
1. **Parse README.md for Documented Features:**
   - Extract every behavioral specification, feature claim, and capability promise
   - Note specific promises about edge cases, error handling, and performance
   - Identify implicit guarantees in API descriptions
   - Document any version-specific features mentioned
   - Catalog each claim with its README line number for traceability

2. **Categorize Documented Claims:**
   Classify each README claim by the `go-stats-generator` output section it maps to:
   - **Function-level claims** → `.functions[]` (line counts, complexity, signatures)
   - **Struct-level claims** → `.structs[]` (member categorization, methods, tags)
   - **Package-level claims** → `.packages[]` (dependencies, cohesion, coupling)
   - **Interface-level claims** → `.interfaces[]` (implementations, embedding depth)
   - **Concurrency claims** → `.concurrency` (goroutines, channels, sync primitives)
   - **Duplication claims** → `.duplication` (clone pairs, duplication ratio)
   - **Naming claims** → `.naming` (convention adherence)
   - **Documentation claims** → `.documentation` (doc coverage, quality)

### Phase 2: Metrics-Driven Verification
1. **Run Comprehensive Analysis:**
   ```bash
   go-stats-generator analyze . --skip-tests --format json --output gap-baseline.json
   ```
   - Capture the full metrics for every analysis dimension
   - This is the ground truth for what the codebase actually implements

2. **Cross-Reference Each Documented Feature:**
   ```bash
   # Verify function-level features
   cat gap-baseline.json | jq '.functions[] | {name, file, lines, complexity, documented}'

   # Verify struct-level features
   cat gap-baseline.json | jq '.structs[] | {name, file, members, methods, embedded_types}'

   # Verify package-level features
   cat gap-baseline.json | jq '.packages[] | {name, dependencies, cohesion, coupling}'

   # Verify interface-level features
   cat gap-baseline.json | jq '.interfaces[] | {name, implementations, embedding_depth}'

   # Verify concurrency features
   cat gap-baseline.json | jq '.concurrency'

   # Verify documentation coverage
   cat gap-baseline.json | jq '.documentation'
   ```
   - For each documented claim, confirm a matching pattern exists in the metrics output
   - Flag any claim where the corresponding metric is absent, zero, or contradictory

3. **Detect Gap Categories:**
   - **Partial Implementations**: Features present in metrics but incomplete (e.g., function exists but lacks documented edge case handling)
   - **Phantom Features**: Documented in README but no matching code pattern in metrics output
   - **Behavioral Nuances**: Metric values contradict documented behavior (e.g., README claims complexity < 10 but metrics show > 15)
   - **Silent Failures**: Operations referenced in README but not reflected in error-handling or concurrency metrics
   - **Documentation Drift**: Code metrics show evolved behavior not reflected in README
   - **Integration Gaps**: Individual components present in metrics but documented cross-component behavior has no matching pattern

### Phase 3: Evidence Collection
1. **For Each Identified Gap, Collect:**
   ```bash
   # Extract specific function metrics as evidence
   cat gap-baseline.json | jq '.functions[] | select(.name == "targetFunction")'

   # Extract documentation coverage for specific packages
   cat gap-baseline.json | jq '.documentation | {coverage, undocumented_exports}'
   ```
   - Exact README.md quote with line number
   - Corresponding `go-stats-generator` JSON output (or absence thereof)
   - Specific file and line location from metrics
   - Clear explanation of the discrepancy

2. **Prioritize Findings:**
   From the cross-reference analysis, classify each gap:
   - **Critical**: Documented feature entirely absent from metrics (phantom feature)
   - **Moderate**: Feature present but metrics contradict documented behavior
   - **Minor**: Feature present but edge cases or nuances missing from implementation

### Phase 4: Gap Report Generation
1. **Produce AUDIT.md** with structured findings (see OUTPUT FORMAT below)
2. **Validate Report Quality:**
   - Every finding references specific README.md text
   - Every finding includes `go-stats-generator` JSON evidence
   - All gaps are reproducible by re-running the analysis commands
   - No false positives from outdated analysis
   - Findings are actionable, not theoretical

## OUTPUT FORMAT:

Create AUDIT.md with the following structure:

### 1. Executive Summary
```
go-stats-generator gap analysis results:
  README Claims Audited: [n]
  Total Gaps Found: [n]
  - Critical (phantom features): [count]
  - Moderate (behavioral contradictions): [count]
  - Minor (missing edge cases): [count]

Analysis Command:
  go-stats-generator analyze . --skip-tests --format json --output gap-baseline.json
```

### 2. Detailed Findings

```markdown
### Gap #[N]: [Precise Description]
**Severity:** [Critical/Moderate/Minor]

**Documentation Reference:**
> "[Exact quote from README.md]" (README.md:L[line])

**Expected Behavior:** [What README specifies]

**go-stats-generator Evidence:**
```json
// Relevant JSON output from gap-baseline.json showing absence or contradiction
```

**Actual Implementation:** [What metrics confirm]

**Gap Details:** [Precise explanation of discrepancy]

**File Location:** `[file.go:lines]` (from .functions[]/.structs[] output)

**Production Impact:** [Specific consequences]

**Recommended Action:** [What should change — in docs or code]
```

### 3. Coverage Matrix
```
Feature Category        README Claims  Verified  Gaps  Coverage
--------------------------------------------------------------
Functions               [n]            [n]       [n]   [x]%
Structs                 [n]            [n]       [n]   [x]%
Packages                [n]            [n]       [n]   [x]%
Interfaces              [n]            [n]       [n]   [x]%
Concurrency             [n]            [n]       [n]   [x]%
Duplication             [n]            [n]       [n]   [x]%
Naming                  [n]            [n]       [n]   [x]%
Documentation           [n]            [n]       [n]   [x]%
--------------------------------------------------------------
TOTAL                   [n]            [n]       [n]   [x]%
```

## GAP DETECTION THRESHOLDS:
```
Feature Presence:
  Any documented feature with zero matching code pattern = Critical gap
  Documented default value differs from actual default   = Moderate gap
  Documented edge case with no test or handler           = Minor gap

Documentation Coverage:
  Min Doc Coverage < 0.7 (--min-doc-coverage default)    = Flag for review
  Exported function without GoDoc                        = Documentation gap
  README feature list entry with no corresponding export = Phantom feature

Metric Contradictions:
  README claims complexity < X but metrics show > X      = Behavioral gap
  README claims N analysis dimensions but JSON lacks one = Missing implementation
  Documented flag/option not present in CLI help         = Interface gap

Report Quality:
  Each finding must reference README text                = Required
  Each finding must include go-stats-generator evidence  = Required
  False positive rate must be 0%                         = Required
```
<!-- Last verified: 2025-07-25 against analyzer output fields and CLI flag defaults -->

Gap Threshold = Any documented feature with no matching metric OR doc coverage < 70% OR metric contradicts README claim
- If no gaps: "Gap analysis complete: go-stats-generator baseline analysis confirms all documented features match implementation. No discrepancies detected."

## EXAMPLE WORKFLOW:
```bash
$ go-stats-generator analyze . --skip-tests --min-doc-coverage 0.7
=== ANALYSIS SUMMARY ===
Functions Analyzed: 142
Structs Analyzed: 38
Packages Analyzed: 12
Interfaces Analyzed: 15
Documentation Coverage: 72.3%
Concurrency Patterns: 8 goroutines, 5 channels, 3 sync primitives

$ go-stats-generator analyze . --skip-tests --format json --output gap-baseline.json
$ cat gap-baseline.json | jq '{functions: .functions | length, structs: .structs | length, interfaces: .interfaces | length, concurrency: .concurrency | keys}'
{
  "functions": 142,
  "structs": 38,
  "interfaces": 15,
  "concurrency": ["goroutines", "channels", "sync_primitives", "worker_pools"]
}

$ # Cross-reference: README claims "pipeline pattern detection"
$ cat gap-baseline.json | jq '.concurrency.pipelines'
null

$ # Gap found: README documents pipeline detection but metrics show no pipeline data

$ # Cross-reference: README claims "design pattern detection"
$ cat gap-baseline.json | jq '.functions[] | select(.name | test("pattern|detect"; "i"))'
# No matching functions found — potential phantom feature

$ # Cross-reference: README documents --min-doc-coverage default of 0.7
$ go-stats-generator analyze . --skip-tests --format json | jq '.documentation.coverage'
0.723

$ # Verified: documentation coverage metric exists and reports 72.3%

$ # Compile findings into AUDIT.md
$ cat AUDIT.md
# Implementation Gap Analysis
Generated: 2025-07-25T14:30:00Z
Analysis Tool: go-stats-generator v1.0.0

## Executive Summary
README Claims Audited: 47
Total Gaps Found: 5
- Critical (phantom features): 2
- Moderate (behavioral contradictions): 2
- Minor (missing edge cases): 1

## Detailed Findings

### Gap #1: Pipeline Pattern Detection Not Implemented
**Severity:** Critical

**Documentation Reference:**
> "Concurrency Analysis: Goroutine patterns, channel analysis, sync primitives,
>  worker pools, pipelines" (README.md:L28)

**Expected Behavior:** Analysis output includes pipeline pattern detection

**go-stats-generator Evidence:**
  $ cat gap-baseline.json | jq '.concurrency.pipelines'
  null

**Actual Implementation:** Concurrency analysis covers goroutines, channels,
sync primitives, and worker pools but omits pipeline detection entirely.

**File Location:** internal/analyzer/concurrency.go

**Production Impact:** Moderate — users relying on pipeline analysis get no data

**Recommended Action:** Implement pipeline detection or remove from README

### Gap #2: Design Pattern Detection Absent
**Severity:** Critical

**Documentation Reference:**
> "advanced metrics like design pattern detection" (README.md:L14)

**Expected Behavior:** Analysis detects and reports design patterns

**go-stats-generator Evidence:**
  $ cat gap-baseline.json | jq '.functions[] | select(.name | test("pattern"; "i"))'
  # No results

**Actual Implementation:** No function, struct, or output field implements
design pattern detection.

**File Location:** N/A — no matching implementation found

**Production Impact:** High — advertised feature entirely missing

**Recommended Action:** Implement design pattern detection or remove from README

## Coverage Matrix
Feature Category        README Claims  Verified  Gaps  Coverage
--------------------------------------------------------------
Functions               12             12        0     100%
Structs                 8              7         1     87%
Packages                6              6         0     100%
Interfaces              5              5         0     100%
Concurrency             7              5         2     71%
Duplication             4              4         0     100%
Naming                  3              3         0     100%
Documentation           2              2         0     100%
--------------------------------------------------------------
TOTAL                   47             44        3     93%
```

This data-driven approach ensures gap detection is based on quantitative `go-stats-generator` metrics rather than subjective code reading, with every finding traceable to both a README claim and a concrete presence or absence in the analysis output.
