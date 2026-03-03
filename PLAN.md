# Implementation Plan: Phase 3 — Misplaced Declarations (Functions/Methods in Wrong Files)

## Phase Overview
- **Objective**: Identify declarations (functions, methods, types) that would be easier to maintain if relocated to more appropriate files or packages
- **Source Document**: ROADMAP.md (Phase 3: Misplaced Declarations)
- **Prerequisites**: Phase 1 (Code Duplication Detection) ✅, Phase 2 (Naming Convention Analysis) ✅
- **Estimated Scope**: Large

## Implementation Steps

### 1. Create PlacementMetrics Types ✅ COMPLETE
- **Deliverable**: Add `PlacementMetrics` struct and supporting types to `internal/metrics/types.go`
- **Dependencies**: None
- **Completed**: 2026-03-02

```go
// PlacementMetrics contains misplaced declaration analysis results
type PlacementMetrics struct {
    MisplacedFunctions int                       `json:"misplaced_functions"`
    MisplacedMethods   int                       `json:"misplaced_methods"`
    LowCohesionFiles   int                       `json:"low_cohesion_files"`
    AvgFileCohesion    float64                   `json:"avg_file_cohesion"`
    FunctionIssues     []MisplacedFunctionIssue  `json:"function_issues"`
    MethodIssues       []MisplacedMethodIssue    `json:"method_issues"`
    CohesionIssues     []FileCohesionIssue       `json:"cohesion_issues"`
}

type MisplacedFunctionIssue struct {
    Name              string   `json:"name"`
    CurrentFile       string   `json:"current_file"`
    SuggestedFile     string   `json:"suggested_file"`
    CurrentAffinity   float64  `json:"current_affinity"`
    SuggestedAffinity float64  `json:"suggested_affinity"`
    ReferencedSymbols []string `json:"referenced_symbols"`
    Severity          string   `json:"severity"`
}

type MisplacedMethodIssue struct {
    MethodName       string  `json:"method_name"`
    ReceiverType     string  `json:"receiver_type"`
    CurrentFile      string  `json:"current_file"`
    ReceiverFile     string  `json:"receiver_file"`
    Distance         string  `json:"distance"` // "same_package" or "different_package"
    Severity         string  `json:"severity"`
}

type FileCohesionIssue struct {
    File            string   `json:"file"`
    CohesionScore   float64  `json:"cohesion_score"`
    IntraFileRefs   int      `json:"intra_file_refs"`
    TotalRefs       int      `json:"total_refs"`
    SuggestedSplits []string `json:"suggested_splits"`
    Severity        string   `json:"severity"`
}
```

### 2. Add Placement Field to Report Struct ✅ COMPLETE
- **Deliverable**: Update `Report` struct in `internal/metrics/types.go` to include `Placement PlacementMetrics`
- **Dependencies**: Step 1
- **Completed**: 2026-03-02

### 3. Build Symbol Reference Index ✅ COMPLETE
- **Deliverable**: Create `internal/analyzer/placement.go` with `PlacementAnalyzer` struct and `buildSymbolIndex()` method
- **Dependencies**: Steps 1, 2
- **Completed**: 2026-03-02

Implementation details:
- Parse all files in package to build complete symbol table
- For each file, record: defined symbols (functions, types, vars, consts), referenced symbols
- Store `map[symbolName]definedInFile` and `map[file][]referencedSymbols`
- Track cross-file symbol references for affinity calculation

### 4. Implement Symbol-to-File Affinity Analysis (Step 3.1) ✅ COMPLETE
- **Deliverable**: Add `AnalyzeFunctionAffinity()` method to `PlacementAnalyzer`
- **Dependencies**: Step 3
- **Completed**: 2026-03-02

Algorithm:
```
For each function F in file X:
  sameFileRefs = count(symbols in F that are defined in X)
  otherFileRefs = map[file]count(symbols in F defined in that file)
  
  currentAffinity = sameFileRefs / totalRefs
  
  For each file Y with refs > margin * totalRefs:
    if otherFileRefs[Y] > currentAffinity + margin:
      flag F as misplaced, suggest Y
```

Configurable threshold: `placement.affinity_margin` (default: 0.25)

### 5. Implement Method Receiver Placement Check (Step 3.2) ✅ COMPLETE
- **Deliverable**: Add `AnalyzeMethodPlacement()` method to `PlacementAnalyzer`
- **Dependencies**: Step 3
- **Completed**: 2026-03-02

Implementation:
- For each method, extract receiver type name
- Find file where receiver type is defined (from symbol index)
- If method file ≠ receiver type file, flag as misplaced
- Classify distance: "same_package" or "different_package"
- Sort results by distance (different_package is higher severity)

### 6. Implement File Cohesion Scoring (Step 3.3) ✅ COMPLETE
- **Deliverable**: Add `AnalyzeFileCohesion()` method to `PlacementAnalyzer`
- **Dependencies**: Step 3
- **Completed**: 2026-03-02

Algorithm:
```
For each file F:
  intraRefs = count(references from F to symbols defined in F)
  totalRefs = count(all symbol references from F)
  cohesion = intraRefs / totalRefs
  
  if cohesion < minCohesion:
    flag F as low cohesion
    identify clusters of related declarations
    suggest logical splits
```

Configurable threshold: `placement.min_cohesion` (default: 0.3)

Cluster detection:
- Group declarations by shared references
- Declarations referencing same set of symbols form a cluster
- Suggest split file names based on dominant cluster theme

### 7. Create Unified Analysis Entry Point ✅ COMPLETE
- **Deliverable**: Add `Analyze()` method that orchestrates all placement checks
- **Dependencies**: Steps 4, 5, 6
- **Completed**: 2026-03-02

```go
func (pa *PlacementAnalyzer) Analyze(files []*ast.File, fset *token.FileSet) PlacementMetrics {
    pa.buildSymbolIndex(files, fset)
    
    functionIssues := pa.AnalyzeFunctionAffinity()
    methodIssues := pa.AnalyzeMethodPlacement()
    cohesionIssues := pa.AnalyzeFileCohesion()
    
    return PlacementMetrics{
        MisplacedFunctions: len(functionIssues),
        MisplacedMethods:   len(methodIssues),
        LowCohesionFiles:   len(cohesionIssues),
        AvgFileCohesion:    calculateAvgCohesion(cohesionIssues),
        FunctionIssues:     functionIssues,
        MethodIssues:       methodIssues,
        CohesionIssues:     cohesionIssues,
    }
}
```

### 8. Add Configuration Options ✅ COMPLETE
- **Deliverable**: Update `internal/config/config.go` with placement config section
- **Dependencies**: None (can be done in parallel)
- **Completed**: 2026-03-02

```go
type PlacementConfig struct {
    AffinityMargin float64 `yaml:"affinity_margin" mapstructure:"affinity_margin"`
    MinCohesion    float64 `yaml:"min_cohesion" mapstructure:"min_cohesion"`
}
```

Update `.go-stats-generator.yaml`:
```yaml
analysis:
  placement:
    affinity_margin: 0.25
    min_cohesion: 0.3
```

### 9. Integrate into Analyze Command ✅ COMPLETE
- **Deliverable**: Update `cmd/analyze.go` to call `PlacementAnalyzer` and populate report
- **Dependencies**: Steps 7, 8
- **Completed**: 2026-03-02

Implementation:
- Added `Placement *analyzer.PlacementAnalyzer` to `AnalyzerSet` struct
- Updated `createAnalyzers()` to accept config and initialize PlacementAnalyzer with config values
- Created `finalizePlacementMetrics()` function similar to `finalizeNamingMetrics()`
- Called finalizePlacementMetrics() in both `runFileAnalysis()` and `runAnalysisWorkflow()`
- Added placement config loading to `loadAnalysisConfiguration()`

Verified:
- All existing tests pass
- JSON output includes placement metrics
- Analysis runs successfully on real codebase

### 10. Add Console Reporter Output ✅ COMPLETE
- **Deliverable**: Update `internal/reporter/console.go` with `writePlacementAnalysis()` method
- **Dependencies**: Steps 1, 9
- **Completed**: 2026-03-02

Implementation:
- Added placement section check in `Generate()` method to call `writePlacementAnalysis()` when violations exist
- Created `writePlacementAnalysis()` method to display summary statistics (counts and average cohesion)
- Created `writeMisplacedFunctions()` to display top misplaced function issues in table format with affinity gains
- Created `writeMisplacedMethods()` to display top misplaced method issues with receiver information
- Created `writeFileCohesionIssues()` to display low cohesion files with suggested splits
- All methods use severity-based sorting (high > medium > low) and respect the configured display limit
- Added comprehensive unit tests for both basic output and sorting behavior
- Verified end-to-end: `./go-stats-generator analyze .` produces placement analysis in console output

Output sections:
- Placement Summary (counts, avg cohesion)
- Misplaced Functions table (name, current file, suggested file, affinity scores)
- Misplaced Methods table (method, receiver, current file, receiver file, distance)
- Low Cohesion Files table (file, cohesion score, suggested splits)

### 11. Add JSON Reporter Output ✅ COMPLETE
- **Deliverable**: Update `internal/reporter/json.go` to include placement metrics
- **Dependencies**: Steps 1, 9
- **Completed**: 2026-03-02

JSON output automatically includes full report struct serialization; placement field is serialized correctly.
Verified with test output showing complete placement metrics in JSON format.

### 12. Add HTML Reporter Output ✅ COMPLETE
- **Deliverable**: Update `internal/reporter/templates/html/report.html` with Placement tab
- **Dependencies**: Steps 1, 9
- **Completed**: 2026-03-02

Implementation:
- Added conditional placement tab button in navigation (displays when violations exist)
- Created placement content section with summary cards showing:
  - Misplaced Functions count
  - Misplaced Methods count
  - Low Cohesion Files count
  - Average File Cohesion score
- Added three detailed tables:
  - Misplaced Functions: shows function name, current/suggested files, affinity scores, affinity gain, and severity
  - Misplaced Methods: shows method name, receiver type, current/receiver files, distance, and severity
  - Low Cohesion Files: shows file, cohesion score, intra-file refs, total refs, suggested splits, and severity
- Added `subtract` template function to html.go for calculating affinity gain
- Created comprehensive unit test in html_test.go (TestHTMLReporter_WithPlacement)
- All existing tests continue to pass
- Verified end-to-end: `./go-stats-generator analyze . --format=html` produces placement section when violations exist

Output sections:
- Placement tab button (conditional on violations)
- Placement Summary (counts, avg cohesion)
- Misplaced Functions table (name, current file, suggested file, affinity scores with gain calculation)
- Misplaced Methods table (method, receiver, current file, receiver file, distance)
- Low Cohesion Files table (file, cohesion score, intra/total refs, suggested splits)

### 13. Add Markdown Reporter Output ✅ COMPLETE
- **Deliverable**: Update `internal/reporter/templates/markdown/report.md` with Placement section
- **Dependencies**: Steps 1, 9
- **Completed**: 2026-03-02

Implementation:
- Added conditional placement section that displays when violations exist (MisplacedFunctions + MisplacedMethods + LowCohesionFiles > 0)
- Created placement summary table showing:
  - Misplaced Functions count
  - Misplaced Methods count
  - Low Cohesion Files count
  - Average File Cohesion score
- Added three detailed markdown tables:
  - Misplaced Functions: shows function name, current/suggested files, affinity scores (current, suggested, gain), and severity
  - Misplaced Methods: shows method name, receiver type, current/receiver files, distance, and severity
  - Low Cohesion Files: shows file, cohesion score, intra-file refs, total refs, suggested splits (comma-separated), and severity
- Added `subtract` helper function to markdown.go template function map for calculating affinity gain
- Created comprehensive unit test in markdown_test.go (TestMarkdownReporter_WithPlacement)
- All existing tests continue to pass
- Verified end-to-end: `./go-stats-generator analyze . --format=markdown` produces placement section when violations exist

Output sections:
- Placement summary (counts, avg cohesion)
- Misplaced Functions table (name, current file, suggested file, affinity scores with gain calculation)
- Misplaced Methods table (method, receiver, current file, receiver file, distance)
- Low Cohesion Files table (file, cohesion score, intra/total refs, suggested splits)

### 14. Create Unit Tests ✅ COMPLETE
- **Deliverable**: Create `internal/analyzer/placement_test.go` with comprehensive test coverage
- **Dependencies**: Steps 3-6
- **Completed**: 2026-03-02
- **Coverage**: >80% on all placement analyzer methods

Test cases:
- Function with all references to same file (100% affinity)
- Function with references split across files (identify best fit)
- Method defined in same file as receiver (no issue)
- Method defined in different file from receiver (flag)
- Method with receiver in different package (high severity)
- High cohesion file (>0.3 score)
- Low cohesion file (<0.3 score)
- File with distinct clusters of declarations

### 15. Create Integration Tests
- **Deliverable**: Add test fixtures in `testdata/placement/` and integration test
- **Dependencies**: Steps 9-13

Test fixtures:
- `testdata/placement/misplaced_function/` - function referencing wrong file
- `testdata/placement/misplaced_method/` - method away from receiver
- `testdata/placement/low_cohesion/` - file with unrelated declarations
- `testdata/placement/high_cohesion/` - well-organized file

### 16. Add Benchmark Tests
- **Deliverable**: Create `internal/analyzer/placement_bench_test.go`
- **Dependencies**: Step 7

Ensure placement analysis doesn't degrade performance below 50,000-file-in-60-seconds target.

## Technical Specifications

- **AST Walking Strategy**: Use `go/ast.Inspect()` for symbol extraction; track `*ast.Ident` references within function bodies
- **Symbol Resolution**: Build per-package symbol table mapping names to defining files; handle imports and qualified names
- **Affinity Calculation**: Use ratio-based scoring: `affinity = sameFileRefs / (sameFileRefs + otherFileRefs)`
- **Cohesion Algorithm**: LCOM-style metric based on shared method/field access patterns adapted for file scope
- **Cluster Detection**: Use simple graph partitioning: declarations sharing >50% of references form a cluster
- **Performance**: Process symbol index once per package, cache for all placement checks; O(n) where n = total references

## Validation Criteria

- [x] `PlacementMetrics` struct added to `internal/metrics/types.go`
- [x] `Report.Placement` field added and serializes correctly in JSON output
- [x] `PlacementAnalyzer` correctly identifies functions with low affinity to current file
- [x] `PlacementAnalyzer` flags methods defined in different files from their receiver type
- [x] `PlacementAnalyzer` computes file cohesion scores and flags files below threshold
- [x] Configuration options `placement.affinity_margin` and `placement.min_cohesion` work correctly
- [x] PlacementAnalyzer integrated into analyze command workflow
- [x] finalizePlacementMetrics() function created and called correctly
- [x] JSON reporter includes placement metrics
- [x] Console reporter displays placement analysis when violations exist
- [x] HTML reporter includes Placement tab when violations exist
- [x] Markdown reporter includes Placement section when violations exist
- [x] Unit tests achieve >85% code coverage for `internal/analyzer/placement.go`
- [ ] Integration tests pass for all test fixtures
- [ ] Benchmark tests confirm no performance regression (50K files in 60s target maintained)
- [x] Running `go-stats-generator analyze .` on this repository produces placement metrics

## Known Gaps

- **Cross-package affinity**: Current design focuses on intra-package placement; cross-package suggestions may require import analysis and could suggest inappropriate moves
  - **Impact**: May miss optimization opportunities or suggest moves that would create circular dependencies
  - **Resolution**: Phase 1 implementation focuses on same-package placement; cross-package analysis deferred to Phase 3.1 enhancement

- **Generic type handling**: Methods on generic types may have complex receiver type resolution
  - **Impact**: May produce false positives for methods on parameterized types
  - **Resolution**: Add special handling for `*ast.IndexExpr` receiver types (generic instantiations)

- **Interface implementation scattering**: No detection for interface implementations spread across many files
  - **Impact**: Won't flag when a type's interface methods are scattered
  - **Resolution**: Could be added as enhancement after core placement analysis is stable
