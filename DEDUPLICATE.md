# TASK DESCRIPTION:
Perform a data-driven duplication elimination analysis to identify and consolidate the **top 5-10 most significant code clone groups** in the codebase below professional duplication thresholds. Use `go-stats-generator` baseline analysis (with --skip-tests), targeted deduplication guidance, and differential validation to ensure measurable duplication reduction while preserving functionality.

When results are ambiguous, such as a tie between clone sizes or if multiple clone types exist for the same block, always choose the **shortest clone group** (by line count) first. Working from shortest to longest ensures simpler, lower-risk consolidations happen first and can collapse larger clones that contain them.

## CONSTRAINT:

Use only `go-stats-generator` and existing tests for your analysis. You are absolutely forbidden from writing new code of any kind or using any other code analysis tools.

## PREREQUISITES:
**Minimum Required Version:** `go-stats-generator` v1.0.0 or higher  
Install and configure `go-stats-generator` for comprehensive duplication analysis and improvement tracking:

### Installation:
```bash
# First, check if go-stats-generator is already installed
which go-stats-generator
# If not, install it with `go install`
go install github.com/opd-ai/go-stats-generator@latest
```

## Recommendations:
```bash
# Extract only task-relevant sections from JSON; discard everything else
go-stats-generator analyze --format json | jq '{duplication: .duplication}'
which jq || sudo apt-get install -y jq
```
**Section filter**: Use only `.duplication` from the report. Exclude `.functions`, `.structs`, `.interfaces`, `.packages`, `.patterns`, `.concurrency`, `.complexity`, `.documentation`, `.generics`, `.naming`, `.placement`, `.organization`, `.burden`, `.scores`, `.suggestions` — they are not relevant to code clone elimination.

### Required Analysis Workflow:
```bash
# Phase 1: Establish baseline and identify duplication targets
go-stats-generator analyze . --min-block-lines 6 --similarity-threshold 0.80 --skip-tests --format json --output baseline.json --sections duplication
go-stats-generator analyze . --min-block-lines 6 --similarity-threshold 0.80 --skip-tests

# Phase 2: Generate deduplication plan
# Using the results generated in phase 1, select significant clone groups for consolidation.

# Phase 3: Post-deduplication validation
go-stats-generator analyze . --format json --output deduplicated.json --min-block-lines 6 --similarity-threshold 0.80 --skip-tests --sections duplication

# Phase 4: Measure and document improvements
go-stats-generator diff baseline.json deduplicated.json
go-stats-generator diff baseline.json deduplicated.json --format html --output deduplication-report.html
```

## CONTEXT:
You are an automated Go code auditor using `go-stats-generator` for enterprise-grade duplication detection and elimination validation. The tool provides precise clone metrics, identifies duplication targets by type (exact, renamed, near), and measures improvements through differential analysis. Focus on clone groups identified by the tool's duplication analysis engine, working from shortest to longest within each priority tier for incremental, low-risk consolidation.

## INSTRUCTIONS:

### Phase 1: Data-Driven Target Identification
1. **Run Baseline Analysis:**
  ```bash
  go-stats-generator analyze . --skip-tests
  ```
  - Record the **top 5-10 largest clone groups** and their metrics
  - Note specific clone attributes (type, line count, instance count, file locations) for each
  - Identify the files containing these duplicated blocks

2. **Extract Duplication Summary:**
  ```bash
  go-stats-generator analyze . --skip-tests --format json --output baseline.json
  cat baseline.json | jq '.duplication'
  ```
  - Capture the overall duplication metrics:
    * Clone Pairs Detected (total number of clone groups)
    * Duplicated Lines (total duplicated lines across all clones)
    * Duplication Ratio (percentage of codebase that is duplicated)
    * Largest Clone Size (lines in the biggest clone)

3. **Prioritize Deduplication Targets:**
  From the baseline analysis, select clone groups for consolidation in priority order:
  - **Critical (≥20 duplicated lines, ≥3 instances):** Consolidate immediately - highest impact
  - **High (≥10 duplicated lines, ≥2 instances):** Consolidate in second pass
  - **Medium (≥6 duplicated lines, ≥2 instances):** Consolidate if time permits
  
  When prioritizing between clone groups:
  - Within each priority tier, work from **shortest to longest** clone groups
  - If line counts are tied, choose the clone group with **more instances** first
  - If both line count and instance count are tied, choose the clone group appearing in **more distinct files** first
  - Target at least 5 clone groups; extend to 10 if more than 5 clone groups exceed Critical or High priority thresholds
  
  **Rationale:** Shorter clones are safer to consolidate, and eliminating them first may reduce or collapse larger overlapping clones, simplifying subsequent passes.

4. **Classify Clone Types for Strategy Selection:**
  ```bash
  cat baseline.json | jq '.duplication.clones[] | {type, line_count, instances: (.instances | length), locations: [.instances[] | "\(.file):\(.start_line)-\(.end_line)"]}'
  ```
  - **Type 1 (exact):** Identical code after whitespace normalization → extract shared function directly
  - **Type 2 (renamed):** Same structure, different identifiers → extract parameterized function
  - **Type 3 (near):** Similar structure above similarity threshold → extract with configuration parameter or strategy pattern

### Phase 2: Guided Deduplication Implementation
1. **Follow Clone-Type-Specific Strategies:**

  **For Exact Clones (Type 1):**
  - Extract the duplicated block into a single shared function
  - Replace all instances with calls to the shared function
  - No parameterization needed — code is identical
  - **Placement rules:**
    - If all instances reside in the **same package**, keep the shared function private in that package
    - If instances span **multiple sub-packages**, move the shared function to a `common` sub-package and export it (uppercase name) so all consumer packages can import it

  **For Renamed Clones (Type 2):**
  - Identify the differing identifiers between instances
  - Extract a parameterized function where differing identifiers become parameters
  - Replace all instances with calls passing the appropriate arguments
  - **Placement rules:**
    - If all instances reside in the **same package**, keep the shared function private in that package
    - If instances span **multiple sub-packages**, move the shared function to `common` and export it
  - Example: If two blocks differ only in variable names `userCount` vs `orderCount`, extract a function with a generic parameter name

  **For Near Clones (Type 3):**
  - Identify the structural differences between instances
  - Determine if differences can be abstracted via:
    * **Parameters:** If differences are values or variable names
    * **Function arguments:** If differences are behavioral (pass a function/closure)
    * **Interfaces:** If differences represent distinct strategies
  - Extract the common structure and inject the varying parts
  - Accept slightly higher complexity in the shared function if it eliminates significant duplication
  - **Placement rules:**
    - If all instances reside in the **same package**, keep the shared function private in that package
    - If instances span **multiple sub-packages**, move the shared function to `common` and export it; define any required interfaces in `common` as well to avoid import cycles

2. **Create Consolidated Functions:**
  - Name functions using verb-first camelCase (e.g., `processBlock`, `validateEntry`)
    - ❌ Avoid noun-first or snake_case names (e.g., `blockProcessor`, `validate_entry`)
  - Target metrics per consolidated function:
    * Overall complexity < 9.0
    * Cyclomatic complexity < 9
    * Line count < 40
  - Add GoDoc comments starting with function name  
    *Example:*  
    ```go
    // processBlock handles the common processing logic shared across
    // multiple analysis stages, reducing code duplication.
    func processBlock(block StatementBlock, threshold float64) error {
        // ...
    }
    ```
  - Add GoDoc comments starting with function name and containing a description of the function's purpose and the duplication it eliminates

3. **Preserve Code Correctness:**
  - Maintain error propagation chains from all original call sites
  - Keep defer statements in correct scope
  - Preserve variable access patterns and side effects
  - **Carefully consider mutexes, locks, and thread safety:**
    - Preserve lock boundaries for critical sections (e.g., code protected by `sync.Mutex`, `sync.RWMutex`, or other synchronization primitives)
    - Avoid splitting a lock/unlock pair between the original and extracted functions in ways that could introduce race conditions or deadlocks
    - Ensure extracted functions that access shared state are used only when the caller holds the appropriate lock, or that the extracted function acquires and releases it safely
    - After refactoring concurrency-sensitive code, validate with `go test -race` to detect data races
  - **Maintain import consistency:**
    - If a shared function is moved to a different package, update all import paths
    - Avoid circular imports — if consolidation would create a cycle, keep the function in `common`
  - **Cross-package deduplication via `common`:**
    - When clones are detected across different sub-packages, create a `common` package to house the shared implementation
    - Export the consolidated function with an uppercase name (e.g., `common.ProcessBlock`) so all consumer packages can import it
    - Keep shared types, helper functions, and interfaces that serve multiple packages in `common`
    - The `common` package must remain a **leaf dependency** — it must not import other project packages to prevent circular imports
    - If the shared function requires types defined in another package, either:
      * Move those types to `common` as well, or
      * Define a minimal interface in `common` that the other package's types satisfy
    - Update all call sites to import `common` and call the exported function

### Phase 3: Differential Validation
1. **Measure Improvements:**
  ```bash
  go-stats-generator analyze . --skip-tests --format json --output deduplicated.json
  go-stats-generator diff baseline.json deduplicated.json
  ```
  - Verify overall duplication ratio decreased significantly
  - Confirm clone pair count is reduced
  - Confirm no new clone groups were introduced
  - Check that consolidated functions do not exceed complexity thresholds:
    * Overall complexity ≤ 9.0
    * Line count ≤ 40
    * Cyclomatic complexity ≤ 9
  - Check for zero regressions in unchanged code

2. **Generate Improvement Report:**
  ```bash
  go-stats-generator diff baseline.json deduplicated.json --format html --output deduplication-report.html
  ```

### Phase 4: Quality Verification
1. **Validate Metrics Achievement:**
  - Duplication ratio reduced by ≥50%
  - Clone pair count reduced proportionally
  - All consolidated functions meet complexity thresholds
  - No new duplication introduced by refactoring
  - No complexity regressions detected by diff analysis
  - Overall codebase quality trend positive

2. **Confirm Functional Preservation:**
  - All tests pass: `go test ./...`
  - Race condition check: `go test -race ./...`
  - Error handling paths unchanged
  - Return value semantics preserved
  - Build succeeds: `go build ./...`

## OUTPUT FORMAT:

Structure your response as:

### 1. Baseline Duplication Summary
```
go-stats-generator identified duplication metrics:
  Clone Pairs Detected: [n]
  Duplicated Lines: [n]
  Duplication Ratio: [x.xx]%
  Largest Clone Size: [n] lines

Top deduplication targets:
1. Clone Group: [hash/identifier]
   - Type: [exact/renamed/near]
   - Line Count: [n] lines
   - Instances: [n] locations
   - Files: [file1:L1-L2, file2:L3-L4, ...]
   - Priority: [Critical/High/Medium]
   - Strategy: [direct extraction/parameterization/interface abstraction]

2. Clone Group: [hash/identifier]
   - Type: [exact/renamed/near]
   - Line Count: [n] lines
   - Instances: [n] locations
   - Files: [file1:L1-L2, file2:L3-L4, ...]
   - Priority: [Critical/High/Medium]
   - Strategy: [direct extraction/parameterization/interface abstraction]

... (continue for top 5-10 clone groups)
```

### 2. Complete Deduplicated Files
Present each fully deduplicated Go file with:
- Original duplicated blocks replaced with calls to shared functions
- New shared functions with GoDoc comments
- Standard Go formatting

### 3. Improvement Validation
```
Differential analysis results:
- Duplication Ratio: [old_%] → [new_%] ([improvement_%] reduction)
- Clone Pairs: [old_count] → [new_count] ([eliminated_count] eliminated)
- Duplicated Lines: [old_count] → [new_count] ([lines_saved] lines saved)
- New shared functions: [list with complexities]
- Complexity regressions: [count]
- Overall quality improvement: [score]
```

## DEDUPLICATION THRESHOLDS:
```
Duplication Detection:
  Minimum Block Lines = 6 (default; blocks smaller than this are ignored)
  Similarity Threshold = 0.80 (80% structural similarity for near-clone detection)

Clone Classification:
  Type 1 (exact):   Identical after whitespace normalization
  Type 2 (renamed): Same structure, different identifiers (similarity ≥ 0.95)
  Type 3 (near):    Structural similarity ≥ similarity_threshold (default 0.80)

Consolidation Priority:
  Critical = ≥20 duplicated lines AND ≥3 instances
  High     = ≥10 duplicated lines AND ≥2 instances
  Medium   = ≥6 duplicated lines AND ≥2 instances

Post-Deduplication Quality Gates:
  Overall Complexity ≤ 9.0
  Cyclomatic Complexity ≤ 9
  Function Length ≤ 40 lines
  No new clone pairs introduced
```
<!-- Last verified: 2025-07-25 against duplication.go:AnalyzeDuplication, ClassifyCloneType, and config defaults -->

Deduplication Threshold = Duplication Ratio > 5% OR Clone Pairs > 10 OR Duplicated Lines > 100
- If no targets: "Deduplication complete: go-stats-generator baseline analysis found no significant code clones exceeding professional duplication thresholds."

## EXAMPLE WORKFLOW:
```bash
$ go-stats-generator analyze . --skip-tests
=== DUPLICATION ANALYSIS ===
Clone Pairs Detected: 12
Duplicated Lines: 187
Duplication Ratio: 8.42%
Largest Clone Size: 24 lines

Top 10 Clone Pairs (by size):
Type            Lines Instances Locations
--------------------------------------------------------------------------------
exact              24        3 internal/analyzer/function.go:45-68 (+2 more)
renamed            18        2 internal/reporter/console.go:120-137 (+1 more)
near               15        4 cmd/analyze.go:200-214 (+3 more)
exact              12        2 internal/scanner/worker.go:30-41 (+1 more)
near               10        3 internal/metrics/types.go:85-94 (+2 more)

$ # Extract baseline JSON for diff comparison
$ go-stats-generator analyze . --skip-tests --format json --output baseline.json
$ cat baseline.json | jq '.duplication | {clone_pairs, duplicated_lines, duplication_ratio, largest_clone_size}'
{
  "clone_pairs": 12,
  "duplicated_lines": 187,
  "duplication_ratio": 0.0842,
  "largest_clone_size": 24
}

$ # Examine specific clone groups
$ cat baseline.json | jq '.duplication.clones[:3][] | {type, line_count, instances: (.instances | length)}'
{"type": "exact", "line_count": 24, "instances": 3}
{"type": "renamed", "line_count": 18, "instances": 2}
{"type": "near", "line_count": 15, "instances": 4}

$ # Consolidate each clone group in priority order...

$ go-stats-generator analyze . --skip-tests --format json --output deduplicated.json
$ go-stats-generator diff baseline.json deduplicated.json
=== IMPROVEMENT SUMMARY ===
CLONE GROUPS CONSOLIDATED: 5

DUPLICATION METRICS:
- Duplication Ratio: 8.42% → 2.15% (74% reduction) ✓
- Clone Pairs: 12 → 4 (8 eliminated) ✓
- Duplicated Lines: 187 → 38 (149 lines saved) ✓

SHARED FUNCTIONS CREATED:
  processAnalysisBlock: 7.2 complexity ✓
  formatReportSection: 5.8 complexity ✓
  validateInputRange: 4.1 complexity ✓
  buildMetricsMap: 6.3 complexity ✓
  applyThresholdCheck: 3.9 complexity ✓

QUALITY GATES:
  All shared functions below complexity threshold: ✓
  No new clone pairs introduced: ✓
  All tests passing: ✓
  Race condition check: ✓
  
QUALITY SCORE: 95/67 (+28 improvement)
REGRESSIONS: 0
DUPLICATION RATIO NOW BELOW 5% THRESHOLD: ✓
```

This data-driven approach ensures deduplication decisions are based on quantitative clone analysis rather than subjective assessment, with measurable validation of improvements across the top 5-10 most significant code clone groups.
