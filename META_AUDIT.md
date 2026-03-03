# TASK DESCRIPTION:
Perform a data-driven implementation audit of exactly one unaudited Go sub-package using `go-stats-generator` as the primary analysis engine. Generate a package-level AUDIT.md with quantitative findings, update the root audit tracker, and produce a structured chat report — all in a single autonomous pass.

## CONSTRAINT:

Use only `go-stats-generator`, `go vet`, `go test`, and existing tests for your analysis. You are absolutely forbidden from using any other code analysis tools. Audit exactly one package per invocation.

## PREREQUISITES:
**Minimum Required Version:** `go-stats-generator` v1.0.0 or higher
Install and configure `go-stats-generator` for comprehensive per-package audit analysis:

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
go-stats-generator analyze --format json | jq '{functions: .functions, documentation: .documentation, naming: .naming, concurrency: .patterns.concurrency_patterns, duplication: .duplication, interfaces: .interfaces, structs: .structs, packages: .packages}'
which jq || sudo apt-get install -y jq
```
**Section filter**: Use only `.functions`, `.documentation`, `.naming`, `.patterns.concurrency_patterns`, `.duplication`, `.interfaces`, `.structs`, and `.packages` from the report (`--sections concurrency` includes the `patterns` section). Exclude `.complexity`, `.generics`, `.placement`, `.organization`, `.burden`, `.scores`, `.suggestions` — they are not relevant to per-package implementation auditing.

## CONTEXT:
You are an automated Go package auditor using `go-stats-generator` for enterprise-grade implementation analysis and historical audit tracking. The tool provides precise per-package metrics across complexity, documentation coverage, naming conventions, concurrency patterns, and duplication — replacing manual code review with quantitative assessment. Focus on a single unaudited sub-package per invocation, generating actionable findings backed by file-and-line citations from the tool's analysis output.

## INSTRUCTIONS:

### Phase 1: Package Selection
1. **Discover Unaudited Packages:**
   ```bash
   # List all Go sub-packages that lack an AUDIT.md
   find ./cmd ./internal ./pkg -type f -name '*.go' -exec dirname {} \; | sort -u | while read dir; do [ ! -f "$dir/AUDIT.md" ] && echo "$dir"; done

   # List all sub-packages that already have an AUDIT.md (for reference)
   find ./cmd ./internal ./pkg -type f -name 'AUDIT.md' -exec dirname {} \; | sort -u
   ```
   - If all packages audited, report completion and exit

2. **Select ONE Unaudited Package** from `pkg/`, `internal/`, or `cmd/`, prioritizing:
   - Packages listed in root `AUDIT.md` but unchecked
   - High integration surface (many imports/importers)
   - Core business logic packages

3. **State Selection:**
   - Chosen package path and 1-sentence rationale

### Phase 2: Data-Driven Package Audit
1. **Run `go-stats-generator` Analysis:**
   ```bash
   go-stats-generator analyze ./path/to/pkg \
     --max-complexity 10 --max-function-length 30 --min-doc-coverage 0.7 \
     --skip-tests --format json --output pkg-audit.json --sections functions,documentation,naming,concurrency,duplication,interfaces,structs,packages

   go-stats-generator analyze ./path/to/pkg \
     --max-complexity 10 --max-function-length 30 --min-doc-coverage 0.7 \
     --skip-tests
   ```

2. **Extract Key Metrics with `jq`:**
   ```bash
   # Documentation coverage
   cat pkg-audit.json | jq '.documentation'

   # Functions exceeding thresholds
   cat pkg-audit.json | jq '[.functions[] | select(.complexity.cyclomatic > 10 or .lines.code > 30)] | sort_by(-.complexity.cyclomatic)'

   # Naming convention violations
   cat pkg-audit.json | jq '.naming'

   # Concurrency patterns and safety
   cat pkg-audit.json | jq '.patterns.concurrency_patterns'

   # Duplication within the package
   cat pkg-audit.json | jq '.duplication'

   # Interface implementations
   cat pkg-audit.json | jq '.interfaces'

   # Struct analysis
   cat pkg-audit.json | jq '.structs'

   # Package-level dependency and cohesion metrics
   cat pkg-audit.json | jq '.packages'
   ```

3. **Audit Categories** — cite `file.go:LINE` for every finding using analysis output:

   **Stub/Incomplete Code**
   - Functions returning only `nil`/zero values (check `.functions[]` with zero complexity and minimal lines)
   - `TODO`/`FIXME`/`placeholder` comments
   - Empty method bodies or unimplemented interfaces

   **Complexity & Function Length**
   - Functions exceeding `--max-complexity 10` threshold (from `.functions[]`)
   - Functions exceeding `--max-function-length 30` threshold (from `.functions[]`)
   - High nesting depth contributors

   **Documentation Quality**
   - Overall doc coverage vs `--min-doc-coverage 0.7` threshold (from `.documentation`)
   - Exported types/functions missing godoc comments
   - Package missing `doc.go` file

   **Naming Conventions**
   - Violations detected in `.naming` output
   - Exported identifiers not following Go conventions

   **API Design**
   - Interfaces not minimal or focused (from `.interfaces`)
   - Unnecessary concrete type exposure
   - Struct design issues (from `.structs`)

   **Concurrency Safety**
   - Goroutine patterns, channel usage, sync primitives (from `.patterns.concurrency_patterns`)
   - Shared state protection gaps
   - Context cancellation handling

   **Error Handling**
   - Unchecked returned errors
   - Swallowed errors (`_ = err`)
   - Missing error wrapping (`fmt.Errorf` with `%w`)

   **Duplication**
   - Clone pairs detected within the package (from `.duplication`)
   - Candidates for extraction or consolidation

   **Dependencies**
   - Circular import risks (from `.packages`)
   - External dependency justification
   - Cohesion and coupling metrics

4. **Run Supplementary Checks:**
   ```bash
   go vet ./path/to/pkg/...
   go test -cover -race ./path/to/pkg/...
   ```

5. **Save Baseline for Historical Tracking:**
   ```bash
   go-stats-generator baseline ./path/to/pkg --format json --output pkg-baseline.json
   ```

### Phase 3: File Operations
Execute in this order:

1. **Create `<package-dir>/AUDIT.md`:**
   ```markdown
   # Audit: <package-import-path>
   **Date**: YYYY-MM-DD
   **Status**: Complete | Incomplete | Needs Work

   ## Summary
   <2-3 sentences: scope, overall health, critical risks>

   ## go-stats-generator Metrics
   | Metric               | Value   | Threshold | Status |
   |----------------------|---------|-----------|--------|
   | Doc Coverage         | <N>%    | ≥70%      | ✓/✗    |
   | Max Cyclomatic       | <N>     | ≤10       | ✓/✗    |
   | Max Function Length  | <N>     | ≤30 lines | ✓/✗    |
   | Test Coverage        | <N>%    | ≥65%      | ✓/✗    |
   | Duplication Ratio    | <N>%    | ≤5%       | ✓/✗    |
   | Naming Violations    | <N>     | 0         | ✓/✗    |

   ## Issues Found
   - [ ] <high|med|low> <category> — <description> (`file.go:LINE`)

   ## Concurrency Assessment
   <Goroutine patterns, channel usage, sync primitives, race check result>

   ## Dependencies
   <External dependencies, cohesion/coupling metrics, circular import risks>

   ## Recommendations
   1. <highest-priority fix>
   2. <next priority fix>
   ```

2. **Update Root `AUDIT.md`:**
   - If package listed: change `[ ]` to `[x]`, append: `— <Status> — <N> issues (<H> high, <M> med, <L> low) — doc:<N>% complexity:<N> test:<N>%`
   - If package not listed: append new checked entry with status and metrics

### Phase 4: Chat Report (max 500 words)
   - Created file path: `<package-dir>/AUDIT.md`
   - `go-stats-generator` metrics summary table (doc coverage, max complexity, max function length, duplication ratio)
   - Test coverage: `<N>%`
   - Top 3-5 critical findings with `file.go:LINE` citations
   - `go vet` result: PASS/FAIL
   - `go test -race` result: PASS/FAIL
   - Updated root `AUDIT.md`: YES
   - Baseline saved: YES/NO

## OUTPUT FORMAT:

Structure your response as:

### 1. Package Selection
```
Selected: ./path/to/pkg
Rationale: <1 sentence>
```

### 2. go-stats-generator Analysis Summary
```
go-stats-generator audit metrics for <package>:
  Doc Coverage:        <N>% (threshold: ≥70%)     [PASS/FAIL]
  Max Cyclomatic:      <N>  (threshold: ≤10)      [PASS/FAIL]
  Max Function Length: <N>  (threshold: ≤30 lines) [PASS/FAIL]
  Test Coverage:       <N>% (threshold: ≥65%)      [PASS/FAIL]
  Duplication Ratio:   <N>% (threshold: ≤5%)       [PASS/FAIL]
  Naming Violations:   <N>  (threshold: 0)         [PASS/FAIL]

Functions exceeding thresholds:
  1. <name> in <file.go>: complexity <N>, <N> lines
  2. <name> in <file.go>: complexity <N>, <N> lines
  ...
```

### 3. Issues Found
```
- [high] <category> — <description> (file.go:LINE)
- [med]  <category> — <description> (file.go:LINE)
- [low]  <category> — <description> (file.go:LINE)
```

### 4. Supplementary Checks
```
go vet:       PASS/FAIL
go test -race: PASS/FAIL
Test Coverage: <N>%
```

### 5. Files Updated
```
Created: <package-dir>/AUDIT.md
Updated: AUDIT.md (root tracker)
Baseline: pkg-baseline.json
```

## THRESHOLDS:
```
Per-Package Audit Quality Gates:
  Documentation Coverage  ≥ 70%  (--min-doc-coverage 0.7)
  Cyclomatic Complexity   ≤ 10   (--max-complexity 10)
  Function Length          ≤ 30   (--max-function-length 30, code lines only)
  Test Coverage            ≥ 65%  (go test -cover)
  Duplication Ratio        ≤ 5%   (--min-block-lines 6, --similarity-threshold 0.80)
  Naming Violations        = 0    (from .naming output)

Audit Status Classification:
  Complete   = All thresholds met, 0 high-severity issues
  Needs Work = Any threshold failed OR ≥1 high-severity issue
  Incomplete = Analysis could not fully complete (e.g., build errors)
```
<!-- Last verified: 2025-07-25 against go-stats-generator CLI flags and JSON output schema -->

## EXAMPLE WORKFLOW:
```bash
$ # Phase 1: Discover unaudited packages
$ find ./cmd ./internal ./pkg -type f -name '*.go' -exec dirname {} \; | sort -u | while read dir; do [ ! -f "$dir/AUDIT.md" ] && echo "$dir"; done
./internal/analyzer
./internal/reporter
./pkg/metrics

$ # Phase 2: Analyze the selected package
$ go-stats-generator analyze ./internal/analyzer \
    --max-complexity 10 --max-function-length 30 --min-doc-coverage 0.7 \
    --skip-tests
=== PACKAGE AUDIT: internal/analyzer ===
Doc Coverage:        62.5% [FAIL — below 70%]
Max Cyclomatic:      14    [FAIL — exceeds 10]
Max Function Length: 42    [FAIL — exceeds 30 lines]
Naming Violations:   2     [FAIL]
Duplication Ratio:   3.1%  [PASS]

Functions Exceeding Thresholds:
  analyzeFunction (function.go:45):  complexity 14, 42 lines
  processStruct (struct.go:112):     complexity 11, 28 lines

Documentation Gaps:
  Missing godoc: ParseFile (parser.go:20), WalkAST (walker.go:55)

Naming Issues:
  non-idiomatic: get_metrics (metrics.go:30) → should be getMetrics
  stuttering: analyzer.AnalyzerConfig (types.go:12) → should be analyzer.Config

$ go-stats-generator analyze ./internal/analyzer \
    --max-complexity 10 --max-function-length 30 --min-doc-coverage 0.7 \
    --skip-tests --format json --output pkg-audit.json --sections functions,documentation,naming,concurrency,duplication,interfaces,structs,packages

$ cat pkg-audit.json | jq '{doc: .documentation, funcs_over: [.functions[] | select(.complexity.cyclomatic > 10) | {name, file, complexity: .complexity.overall, cyclomatic: .complexity.cyclomatic, lines: .lines.code}], naming: .naming}'
{
  "doc": {"coverage": {"overall": 0.625, "functions": 0.60, "types": 0.70}, "missing": ["ParseFile", "WalkAST"]},
  "funcs_over": [
    {"name": "analyzeFunction", "file": "function.go", "complexity": 14, "cyclomatic": 12, "lines": 42},
    {"name": "processStruct", "file": "struct.go", "complexity": 11, "cyclomatic": 9, "lines": 28}
  ],
  "naming": {"violations": 2, "issues": ["get_metrics", "analyzer.AnalyzerConfig"]}
}

$ # Supplementary checks
$ go vet ./internal/analyzer/...
$ go test -cover -race ./internal/analyzer/...
ok   internal/analyzer  0.8s  coverage: 71.2% of statements
PASS

$ # Phase 3: Save baseline for historical tracking
$ go-stats-generator baseline ./internal/analyzer --format json --output pkg-baseline.json

$ # Phase 3: Create package AUDIT.md and update root tracker
$ # (create ./internal/analyzer/AUDIT.md with template)
$ # (update root AUDIT.md: [x] internal/analyzer — Needs Work — 6 issues (2 high, 3 med, 1 low) — doc:62% complexity:14 test:71%)

$ # Phase 4: Chat report
- Created: ./internal/analyzer/AUDIT.md
- go-stats-generator metrics:
  | Metric             | Value | Threshold | Status |
  |--------------------|-------|-----------|--------|
  | Doc Coverage       | 62.5% | ≥70%      | ✗      |
  | Max Cyclomatic     | 14    | ≤10       | ✗      |
  | Max Function Length| 42    | ≤30 lines | ✗      |
  | Test Coverage      | 71.2% | ≥65%      | ✓      |
  | Duplication Ratio  | 3.1%  | ≤5%       | ✓      |
  | Naming Violations  | 2     | 0         | ✗      |
- Top findings:
  1. [high] complexity — analyzeFunction exceeds threshold (function.go:45)
  2. [high] complexity — processStruct exceeds threshold (struct.go:112)
  3. [med]  documentation — ParseFile missing godoc (parser.go:20)
  4. [med]  documentation — WalkAST missing godoc (walker.go:55)
  5. [med]  naming — get_metrics non-idiomatic (metrics.go:30)
- go vet: PASS
- go test -race: PASS
- Updated root AUDIT.md: YES
- Baseline saved: YES
```

This data-driven approach ensures package audits are based on quantitative metrics from `go-stats-generator` rather than subjective review, with measurable thresholds for documentation coverage, complexity, naming conventions, and duplication.
