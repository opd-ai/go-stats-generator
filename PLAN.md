# PROJECT: Go Source Code Stats Generator

## OBJECTIVE:
Create a command-line tool that analyzes Go source code repositories and generates comprehensive statistical reports about code structure, complexity, and patterns. The tool should provide actionable insights for code quality assessment and refactoring decisions, with a focus on computing obscure and detailed metrics that standard linters don't typically capture.

## TECHNICAL SPECIFICATIONS:
- Language: Go
- Type: CLI application  
- Module: `github.com/opd-ai/go-stats-generator`
- Key Features:
  - Function and method length analysis (precise line counting)
  - Object complexity metrics (detailed member categorization)
  - Advanced AST pattern detection
  - Concurrent file processing with worker pools
  - Multiple output formats (console, JSON, CSV, HTML)
  - Filtering and aggregation capabilities
  - Historical comparison support
- Performance Requirements: 
  - Process 50,000+ files within 60 seconds
  - Memory usage under 1GB for enterprise codebases
  - Support for repositories with 100MB+ of Go source
  - Concurrent processing with configurable worker count

## ARCHITECTURE GUIDELINES:
### Preferred Libraries:
| Library | Use Case | Justification |
|---------|----------|---------------|
| `go/parser` | AST parsing | Standard library, zero dependencies |
| `go/ast` | AST traversal | Standard library, complete node access |
| `go/token` | Position/line tracking | Standard library, accurate positioning |
| `github.com/spf13/cobra` | CLI framework | 35k+ stars, excellent docs, industry standard |
| `github.com/spf13/viper` | Configuration | Pairs with cobra, mature config handling |
| `github.com/olekukonko/tablewriter` | Console tables | 4k+ stars, professional ASCII output |
| `github.com/jedib0t/go-pretty/v6` | Advanced formatting | 4k+ stars, rich table/progress features |

### Project Structure:
```
github.com/opd-ai/go-stats-generator/
├── cmd/
│   ├── root.go              # Root command setup
│   ├── analyze.go           # Main analyze command
│   ├── compare.go           # Historical comparison
│   └── version.go           # Version information
├── internal/
│   ├── analyzer/
│   │   ├── function.go      # Function/method analysis
│   │   ├── struct.go        # Struct/interface complexity
│   │   ├── package.go       # Package-level metrics
│   │   ├── complexity.go    # Cyclomatic complexity
│   │   ├── patterns.go      # Design pattern detection
│   │   └── walker.go        # Concurrent AST processing
│   ├── metrics/
│   │   ├── types.go         # Core metric data structures
│   │   ├── aggregator.go    # Statistical aggregation
│   │   ├── calculator.go    # Advanced calculations
│   │   └── history.go       # Historical tracking
│   ├── reporter/
│   │   ├── console.go       # Rich console output
│   │   ├── json.go          # JSON export
│   │   ├── csv.go           # CSV export
│   │   ├── html.go          # HTML dashboard
│   │   └── markdown.go      # Markdown reports
│   ├── scanner/
│   │   ├── discover.go      # File discovery
│   │   ├── filter.go        # File filtering logic
│   │   └── worker.go        # Worker pool management
│   └── config/
│       ├── config.go        # Configuration structures
│       └── defaults.go      # Default settings
├── pkg/
│   └── gostats/
│       ├── api.go           # Public API
│       ├── types.go         # Public types
│       └── errors.go        # Error definitions
├── testdata/
│   ├── simple/              # Simple test projects
│   ├── complex/             # Complex test scenarios
│   └── benchmarks/          # Performance test data
├── docs/
│   ├── metrics.md           # Metric definitions
│   ├── examples.md          # Usage examples
│   └── api.md               # API documentation
├── scripts/
│   ├── build.sh             # Build automation
│   └── test.sh              # Test automation
├── .github/
│   └── workflows/
│       └── ci.yml           # GitHub Actions CI
├── go.mod
├── go.sum
├── README.md
├── LICENSE
├── Makefile
└── .goreleaser.yml
```

### Design Patterns:
- **Visitor Pattern**: For AST traversal with specialized analyzers
- **Strategy Pattern**: For different output formats and analysis strategies
- **Builder Pattern**: For constructing complex metric reports
- **Factory Pattern**: For creating analyzers based on file types
- **Observer Pattern**: For progress reporting during analysis
- **Worker Pool Pattern**: For concurrent file processing

## IMPLEMENTATION PHASES:

### Phase 1: Foundation & Core Parsing
**Tasks:**
- [ ] Initialize Go module `github.com/opd-ai/go-stats-generator`
- [ ] Implement CLI structure with cobra (analyze, compare, version commands)
- [ ] Create core metric data structures with JSON serialization
- [ ] Implement concurrent file discovery and filtering
- [ ] Basic AST parsing with proper error handling and recovery
- [ ] Worker pool for concurrent file processing

**Acceptance Criteria:**
- CLI accepts directory paths, glob patterns, and exclusion filters
- Successfully parses Go files and handles syntax errors gracefully
- Concurrent processing with configurable worker count (default: runtime.NumCPU())
- Progress indication for large repositories
- Proper memory management for large file sets

### Phase 2: Core Analysis Engine
**Tasks:**
- [ ] Function/method length analyzer with precise line counting
- [ ] Struct complexity analyzer with detailed member categorization
- [ ] Cyclomatic complexity calculator using standard algorithms
- [ ] Package dependency analysis and circular detection
- [ ] Interface analysis (implementation ratios, method counts)
- [ ] Concurrency pattern detection (goroutines, channels, mutexes)

**Acceptance Criteria:**
- Accurate line counting excluding comments, blank lines, and braces
- Struct members categorized by: primitives, slices, maps, channels, interfaces, custom types, embedded types, functions
- Cyclomatic complexity matches established tools (gocyclo)
- Package metrics include cohesion and coupling scores
- Detection of common Go concurrency patterns

### Phase 3: Advanced Metrics & Pattern Detection
**Tasks:**
- [ ] Design pattern detection (Singleton, Factory, Builder, Observer)
- [ ] Comment quality analysis (GoDoc coverage, TODO/FIXME tracking)
- [ ] Code smell detection (long parameter lists, deep nesting)
- [ ] Generic usage analysis (type parameters, constraints)
- [ ] Performance anti-pattern detection
- [ ] Test coverage correlation analysis

**Acceptance Criteria:**
- Pattern detection with confidence scores and examples
- Comment density and quality metrics per package
- Identification of functions violating best practices
- Generic type usage statistics and complexity metrics
- Correlation between code metrics and test coverage

### Phase 4: Reporting & Output Formats
**Tasks:**
- [ ] Rich console output with tables, charts, and progress bars
- [ ] JSON export with schema validation
- [ ] CSV export for spreadsheet analysis
- [ ] HTML dashboard with interactive charts
- [ ] Markdown reports for documentation
- [ ] Historical comparison and trend analysis

**Acceptance Criteria:**
- Professional console output with color coding and formatting
- Valid JSON schema for programmatic consumption
- CSV format compatible with Excel and Google Sheets
- HTML dashboard with responsive design and JavaScript charts
- Markdown reports suitable for Git repositories

## LIBRARY SELECTION PROCESS:
1. **AST Processing**: Go standard library (`go/parser`, `go/ast`, `go/token`)
   - Zero external dependencies for core functionality
   - Most reliable and complete AST access
   - Official Go toolchain components

2. **CLI Framework**: `github.com/spf13/cobra` + `github.com/spf13/viper`
   - Cobra: 35k+ stars, last updated this month, excellent documentation
   - Viper: 25k+ stars, seamless integration with cobra
   - Industry standard, avoid custom CLI parsing

3. **Output Formatting**: `github.com/jedib0t/go-pretty/v6`
   - 4k+ stars, actively maintained, rich formatting capabilities
   - Better than basic tablewriter for complex output
   - Progress bars, colors, and advanced table features

4. **Testing**: Standard library `testing` + `github.com/stretchr/testify`
   - Testify: 22k+ stars, assertion helpers for cleaner tests
   - Mock generation for interface testing

## CODE STANDARDS:

### Function Length and Complexity:
```go
// ✅ GOOD: Single responsibility, under 30 lines
func countFunctionLines(fn *ast.FuncDecl, fset *token.FileSet) (LineMetrics, error) {
    if fn.Body == nil {
        return LineMetrics{}, nil
    }
    
    start := fset.Position(fn.Body.Lbrace)
    end := fset.Position(fn.Body.Rbrace)
    
    // Count only non-blank, non-comment lines
    totalLines := end.Line - start.Line - 1
    
    return LineMetrics{
        Total:    totalLines,
        Code:     calculateCodeLines(fn.Body, fset),
        Comments: calculateCommentLines(fn.Body, fset),
    }, nil
}

// ❌ AVOID: Multiple responsibilities, unclear purpose
func analyzeFunction(fn *ast.FuncDecl, fset *token.FileSet) map[string]interface{} {
    // Don't combine line counting, complexity, and pattern detection
}
```

### Struct Complexity Analysis:
```go
// ✅ GOOD: Detailed categorization
type StructComplexity struct {
    Name           string                 `json:"name"`
    TotalFields    int                   `json:"total_fields"`
    FieldsByType   map[FieldType]int     `json:"fields_by_type"`
    EmbeddedTypes  []EmbeddedType        `json:"embedded_types"`
    Methods        []MethodInfo          `json:"methods"`
    Tags           map[string]int        `json:"tag_usage"`
    IsExported     bool                  `json:"is_exported"`
    Complexity     ComplexityScore       `json:"complexity"`
}

type FieldType string

const (
    FieldTypePrimitive   FieldType = "primitive"
    FieldTypeSlice       FieldType = "slice"
    FieldTypeMap         FieldType = "map"
    FieldTypeChannel     FieldType = "channel"
    FieldTypeInterface   FieldType = "interface"
    FieldTypeStruct      FieldType = "struct"
    FieldTypePointer     FieldType = "pointer"
    FieldTypeFunction    FieldType = "function"
    FieldTypeEmbedded    FieldType = "embedded"
)
```

### Error Handling Pattern:
```go
// ✅ GOOD: Comprehensive error context
func analyzePackage(pkgPath string) (*PackageMetrics, error) {
    fset := token.NewFileSet()
    pkgs, err := parser.ParseDir(fset, pkgPath, nil, parser.ParseComments)
    if err != nil {
        return nil, fmt.Errorf("failed to parse package %s: %w", pkgPath, err)
    }
    
    if len(pkgs) == 0 {
        return nil, fmt.Errorf("no Go packages found in %s", pkgPath)
    }
    
    // Process packages...
    metrics, err := processPackages(pkgs, fset)
    if err != nil {
        return nil, fmt.Errorf("failed to analyze package %s: %w", pkgPath, err)
    }
    
    return metrics, nil
}
```

## SIMPLICITY RULES:
- **Maximum function length**: 30 lines (excluding comments and blank lines)
- **Maximum parameter count**: 5 parameters (use structs for complex inputs)
- **Maximum nesting depth**: 3 levels (extract functions for deeper logic)
- **Prefer explicit over implicit**: Use type switches instead of reflection
- **Single concern per file**: Don't mix analysis types in one file
- **Use standard library first**: Only external deps for CLI, formatting, and testing

## OBSCURE STATS REQUIREMENTS:
The tool must calculate these advanced, non-standard metrics:

### 1. Function Signature Complexity:
```go
type FunctionSignature struct {
    ParameterCount     int                    `json:"parameter_count"`
    ReturnCount        int                    `json:"return_count"`
    VariadicUsage      bool                   `json:"has_variadic"`
    ErrorReturn        bool                   `json:"returns_error"`
    InterfaceParams    int                    `json:"interface_parameters"`
    GenericParams      []GenericParam         `json:"generic_parameters"`
    ComplexityScore    float64               `json:"signature_complexity"`
}
```

### 2. Struct Member Analysis by Category:
```go
type StructMemberAnalysis struct {
    Variables struct {
        Primitives    []FieldInfo `json:"primitives"`
        Collections   []FieldInfo `json:"collections"`
        CustomTypes   []FieldInfo `json:"custom_types"`
        Pointers      []FieldInfo `json:"pointers"`
        Interfaces    []FieldInfo `json:"interfaces"`
        Channels      []FieldInfo `json:"channels"`
        Functions     []FieldInfo `json:"function_fields"`
    } `json:"variables"`
    
    Methods struct {
        Exported      []MethodInfo `json:"exported"`
        Unexported    []MethodInfo `json:"unexported"`
        Receivers     []MethodInfo `json:"pointer_receivers"`
        Generics      []MethodInfo `json:"generic_methods"`
    } `json:"methods"`
    
    EmbeddedTypes struct {
        Interfaces    []string `json:"embedded_interfaces"`
        Structs       []string `json:"embedded_structs"`
        Aliases       []string `json:"embedded_aliases"`
    } `json:"embedded"`
}
```

### 3. Advanced Code Pattern Detection:
```go
type CodePatternMetrics struct {
    DesignPatterns struct {
        Singleton      []SingletonPattern    `json:"singleton_usage"`
        Factory        []FactoryPattern      `json:"factory_patterns"`
        Builder        []BuilderPattern      `json:"builder_patterns"`
        Observer       []ObserverPattern     `json:"observer_patterns"`
        Strategy       []StrategyPattern     `json:"strategy_patterns"`
    } `json:"design_patterns"`
    
    ConcurrencyPatterns struct {
        WorkerPools    []WorkerPoolPattern   `json:"worker_pools"`
        Pipelines      []PipelinePattern     `json:"pipelines"`
        FanOut         []FanOutPattern       `json:"fan_out"`
        FanIn          []FanInPattern        `json:"fan_in"`
        Semaphores     []SemaphorePattern    `json:"semaphores"`
    } `json:"concurrency_patterns"`
    
    AntiPatterns struct {
        GodObjects     []GodObjectWarning    `json:"god_objects"`
        LongMethods    []LongMethodWarning   `json:"long_methods"`
        DeepNesting    []DeepNestingWarning  `json:"deep_nesting"`
        MagicNumbers   []MagicNumberWarning  `json:"magic_numbers"`
    } `json:"anti_patterns"`
}
```

### 4. Comment Quality and Documentation Metrics:
```go
type DocumentationMetrics struct {
    GoDocCoverage struct {
        Packages      float64 `json:"package_coverage"`
        Functions     float64 `json:"function_coverage"`
        Types         float64 `json:"type_coverage"`
        Methods       float64 `json:"method_coverage"`
    } `json:"godoc_coverage"`
    
    CommentQuality struct {
        AverageLength     float64            `json:"average_comment_length"`
        CodeExamples      int                `json:"code_examples_count"`
        TODOs             []TODOComment      `json:"todo_comments"`
        FIXMEs            []FIXMEComment     `json:"fixme_comments"`
        HACKs             []HACKComment      `json:"hack_comments"`
        InlineComments    int                `json:"inline_comments"`
        BlockComments     int                `json:"block_comments"`
    } `json:"comment_quality"`
}
```

### 5. Generic Usage Analysis (Go 1.18+):
```go
type GenericUsageMetrics struct {
    TypeParameters struct {
        Count         int                    `json:"total_count"`
        Constraints   map[string]int         `json:"constraint_usage"`
        Complexity    []GenericComplexity    `json:"complexity_analysis"`
    } `json:"type_parameters"`
    
    Instantiations struct {
        Functions     []GenericInstantiation `json:"function_instantiations"`
        Types         []GenericInstantiation `json:"type_instantiations"`
        Methods       []GenericInstantiation `json:"method_instantiations"`
    } `json:"instantiations"`
}
```

## VALIDATION CHECKLIST:
- [ ] All functions under 30 lines with single responsibility
- [ ] Concurrent processing tested with 10,000+ files
- [ ] Memory usage profiled and optimized for large codebases
- [ ] Line counting accuracy verified against manual counts
- [ ] Cyclomatic complexity matches established tools (gocyclo, gometalinter)
- [ ] All external dependencies are well-maintained (>1000 stars, recent updates)
- [ ] Error handling covers all file I/O and parsing operations
- [ ] Output formats validated (JSON schema, CSV headers, HTML rendering)
- [ ] Performance benchmarks for different repository sizes
- [ ] Cross-platform compatibility (Windows, macOS, Linux)
- [ ] Unit tests achieve 85%+ coverage on business logic
- [ ] Integration tests verify accuracy against known Go projects
- [ ] CLI help text includes examples and metric explanations
- [ ] README includes installation, usage, and metric definitions
- [ ] Documentation explains all obscure metrics with examples

## PERFORMANCE REQUIREMENTS:
- Process standard library (400+ files): <10 seconds
- Process Kubernetes codebase (10,000+ files): <60 seconds  
- Memory usage: <1GB for enterprise codebases
- Concurrent processing: Configurable worker pools (default: NumCPU)
- Progress reporting: Real-time progress for operations >5 seconds
- Incremental analysis: Skip unchanged files when possible

## OUTPUT EXAMPLE:
```
Go Source Code Statistics Report
Repository: github.com/kubernetes/kubernetes
Generated: 2024-01-15 14:30:22
Analysis Time: 45.2 seconds
Files Processed: 12,847

=== OVERVIEW ===
Total Lines of Code: 1,247,832
Total Functions: 28,156
Total Methods: 15,234
Total Structs: 4,445
Total Interfaces: 1,223

=== FUNCTION COMPLEXITY ===
Average Function Length: 15.4 lines
Longest Function: processNodeUpdate (127 lines) in pkg/controller/node/controller.go
Functions >50 lines: 234 (0.8%)
Functions >100 lines: 23 (0.08%)
Cyclomatic Complexity >10: 156 functions

=== STRUCT ANALYSIS ===
Average Fields per Struct: 6.8
Most Complex Struct: PodSpec (67 fields) in core/v1/types.go
  - Primitives: 23 fields
  - Slices: 18 fields  
  - Maps: 8 fields
  - Custom Types: 12 fields
  - Pointers: 6 fields

=== DESIGN PATTERNS DETECTED ===
Singleton Patterns: 45 instances
Factory Patterns: 123 instances
Builder Patterns: 78 instances
Observer Patterns: 34 instances

=== CONCURRENCY ANALYSIS ===
Goroutine Creation Sites: 1,247
Channel Usage: 2,456 (buffered: 1,234, unbuffered: 1,222)
Mutex Usage: 567 instances
Worker Pool Patterns: 89 instances

=== DOCUMENTATION QUALITY ===
GoDoc Coverage: 78.4%
Functions with Documentation: 21,984 (78.1%)
TODO Comments: 234
FIXME Comments: 67
Code Examples in Comments: 145
```