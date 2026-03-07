package metrics

import (
	"time"
)

// Report represents the complete analysis report for a repository
type Report struct {
	Metadata             ReportMetadata       `json:"metadata"`
	Overview             OverviewMetrics      `json:"overview"`
	Functions            []FunctionMetrics    `json:"functions"`
	Structs              []StructMetrics      `json:"structs"`
	Interfaces           []InterfaceMetrics   `json:"interfaces"`
	Packages             []PackageMetrics     `json:"packages"`
	CircularDependencies []CircularDependency `json:"circular_dependencies"`
	Patterns             PatternMetrics       `json:"patterns"`
	Complexity           ComplexityMetrics    `json:"complexity"`
	Documentation        DocumentationMetrics `json:"documentation"`
	Generics             GenericMetrics       `json:"generics"`
	Duplication          DuplicationMetrics   `json:"duplication"`
	Naming               NamingMetrics        `json:"naming"`
	Placement            PlacementMetrics     `json:"placement"`
	Organization         OrganizationMetrics  `json:"organization"`
	Burden               BurdenMetrics        `json:"burden"`
	Scores               ScoringMetrics       `json:"scores"`
	TestCoverage         TestCoverageMetrics  `json:"test_coverage,omitempty"`
	TestQuality          TestQualityMetrics   `json:"test_quality,omitempty"`
	Team                 *TeamMetrics         `json:"team,omitempty"`
	Suggestions          []SuggestionInfo     `json:"suggestions,omitempty"`
}

// ReportMetadata contains information about the analysis run
type ReportMetadata struct {
	Repository     string        `json:"repository"`
	GeneratedAt    time.Time     `json:"generated_at"`
	AnalysisTime   time.Duration `json:"analysis_time"`
	FilesProcessed int           `json:"files_processed"`
	ToolVersion    string        `json:"tool_version"`
	GoVersion      string        `json:"go_version"`
}

// OverviewMetrics provides high-level statistics for total lines, functions, and structural elements.
type OverviewMetrics struct {
	TotalLinesOfCode int `json:"total_lines_of_code"`
	TotalFunctions   int `json:"total_functions"`
	TotalMethods     int `json:"total_methods"`
	TotalStructs     int `json:"total_structs"`
	TotalInterfaces  int `json:"total_interfaces"`
	TotalPackages    int `json:"total_packages"`
	TotalFiles       int `json:"total_files"`
}

// LineMetrics represents line counting information for total, code, comments, and blank lines.
type LineMetrics struct {
	Total    int `json:"total"`
	Code     int `json:"code"`
	Comments int `json:"comments"`
	Blank    int `json:"blank"`
}

// FunctionMetrics contains detailed function analysis including complexity, signature, and documentation metrics.
type FunctionMetrics struct {
	Name          string            `json:"name"`
	Package       string            `json:"package"`
	File          string            `json:"file"`
	Line          int               `json:"line"`
	IsExported    bool              `json:"is_exported"`
	IsMethod      bool              `json:"is_method"`
	ReceiverType  string            `json:"receiver_type,omitempty"`
	Lines         LineMetrics       `json:"lines"`
	Signature     FunctionSignature `json:"signature"`
	Complexity    ComplexityScore   `json:"complexity"`
	Documentation DocumentationInfo `json:"documentation"`
}

// FunctionSignature represents function signature complexity including parameters, returns, and generic constraints.
type FunctionSignature struct {
	ParameterCount  int            `json:"parameter_count"`
	ReturnCount     int            `json:"return_count"`
	VariadicUsage   bool           `json:"has_variadic"`
	ErrorReturn     bool           `json:"returns_error"`
	InterfaceParams int            `json:"interface_parameters"`
	GenericParams   []GenericParam `json:"generic_parameters"`
	ComplexityScore float64        `json:"signature_complexity"`
}

// GenericParam represents a generic type parameter
type GenericParam struct {
	Name        string   `json:"name"`
	Constraints []string `json:"constraints"`
}

// ComplexityScore represents various complexity measurements including cyclomatic, cognitive, and nesting depth.
type ComplexityScore struct {
	Cyclomatic   int     `json:"cyclomatic"`
	Cognitive    int     `json:"cognitive"`
	NestingDepth int     `json:"nesting_depth"`
	Overall      float64 `json:"overall"`
}

// DocumentationInfo contains documentation quality metrics including comment presence, length, and quality score.
type DocumentationInfo struct {
	HasComment    bool    `json:"has_comment"`
	CommentLength int     `json:"comment_length"`
	HasExample    bool    `json:"has_example"`
	QualityScore  float64 `json:"quality_score"`
}

// StructMetrics contains detailed struct analysis including fields, embedded types, methods, and complexity.
type StructMetrics struct {
	Name          string            `json:"name"`
	Package       string            `json:"package"`
	File          string            `json:"file"`
	Line          int               `json:"line"`
	IsExported    bool              `json:"is_exported"`
	TotalFields   int               `json:"total_fields"`
	FieldsByType  map[FieldType]int `json:"fields_by_type"`
	EmbeddedTypes []EmbeddedType    `json:"embedded_types"`
	Methods       []MethodInfo      `json:"methods"`
	Tags          map[string]int    `json:"tag_usage"`
	Complexity    ComplexityScore   `json:"complexity"`
	Documentation DocumentationInfo `json:"documentation"`
}

// FieldType represents the category of a struct field
type FieldType string

const (
	FieldTypePrimitive FieldType = "primitive"
	FieldTypeSlice     FieldType = "slice"
	FieldTypeMap       FieldType = "map"
	FieldTypeChannel   FieldType = "channel"
	FieldTypeInterface FieldType = "interface"
	FieldTypeStruct    FieldType = "struct"
	FieldTypePointer   FieldType = "pointer"
	FieldTypeFunction  FieldType = "function"
	FieldTypeEmbedded  FieldType = "embedded"
)

// EmbeddedType represents an embedded type in a struct
type EmbeddedType struct {
	Name       string `json:"name"`
	Package    string `json:"package"`
	IsPointer  bool   `json:"is_pointer"`
	IsExported bool   `json:"is_exported"`
}

// MethodInfo represents method information including receiver type, signature, and complexity metrics.
type MethodInfo struct {
	Name          string            `json:"name"`
	IsExported    bool              `json:"is_exported"`
	IsPointer     bool              `json:"is_pointer_receiver"`
	Signature     FunctionSignature `json:"signature"`
	Lines         LineMetrics       `json:"lines"`
	Complexity    ComplexityScore   `json:"complexity"`
	Documentation DocumentationInfo `json:"documentation"`
}

// InterfaceMetrics contains interface analysis including methods, embedding depth, and implementation tracking.
type InterfaceMetrics struct {
	Name                string            `json:"name"`
	Package             string            `json:"package"`
	File                string            `json:"file"`
	Line                int               `json:"line"`
	IsExported          bool              `json:"is_exported"`
	MethodCount         int               `json:"method_count"`
	Methods             []InterfaceMethod `json:"methods"`
	EmbeddedInterfaces  []string          `json:"embedded_interfaces"`
	Implementations     []string          `json:"implementations"`
	ImplementationCount int               `json:"implementation_count"`
	ImplementationRatio float64           `json:"implementation_ratio"`
	EmbeddingDepth      int               `json:"embedding_depth"`
	ComplexityScore     float64           `json:"complexity_score"`
	Documentation       DocumentationInfo `json:"documentation"`
}

// InterfaceMethod represents a method in an interface
type InterfaceMethod struct {
	Name      string            `json:"name"`
	Signature FunctionSignature `json:"signature"`
}

// PackageMetrics contains package-level analysis for dependencies, cohesion, and coupling metrics.
type PackageMetrics struct {
	Name          string            `json:"name"`
	Path          string            `json:"path"`
	Files         []string          `json:"files"`
	Lines         LineMetrics       `json:"lines"`
	Functions     int               `json:"functions"`
	Structs       int               `json:"structs"`
	Interfaces    int               `json:"interfaces"`
	Dependencies  []string          `json:"dependencies"`
	Dependents    []string          `json:"dependents"`
	CohesionScore float64           `json:"cohesion_score"`
	CouplingScore float64           `json:"coupling_score"`
	Documentation DocumentationInfo `json:"documentation"`
}

// PackageReport contains comprehensive package analysis results
type PackageReport struct {
	Packages                   []PackageMetrics     `json:"packages"`
	TotalPackages              int                  `json:"total_packages"`
	CircularDependencies       []CircularDependency `json:"circular_dependencies"`
	DependencyGraph            map[string][]string  `json:"dependency_graph"`
	AverageFilesPerPackage     float64              `json:"average_files_per_package"`
	AverageFunctionsPerPackage float64              `json:"average_functions_per_package"`
	AverageTypesPerPackage     float64              `json:"average_types_per_package"`
}

// CircularDependency represents a circular dependency in the package graph
type CircularDependency struct {
	Packages []string `json:"packages"`
	Severity string   `json:"severity"` // "low", "medium", "high"
}

// PatternMetrics contains design pattern detection results
type PatternMetrics struct {
	DesignPatterns      DesignPatternMetrics      `json:"design_patterns"`
	ConcurrencyPatterns ConcurrencyPatternMetrics `json:"concurrency_patterns"`
	AntiPatterns        AntiPatternMetrics        `json:"anti_patterns"`
}

// DesignPatternMetrics tracks various design patterns including singleton, factory, builder, observer, and strategy.
type DesignPatternMetrics struct {
	Singleton []PatternInstance `json:"singleton"`
	Factory   []PatternInstance `json:"factory"`
	Builder   []PatternInstance `json:"builder"`
	Observer  []PatternInstance `json:"observer"`
	Strategy  []PatternInstance `json:"strategy"`
}

// ConcurrencyPatternMetrics tracks concurrency patterns including worker pools, pipelines, and synchronization primitives.
type ConcurrencyPatternMetrics struct {
	WorkerPools []PatternInstance `json:"worker_pools"`
	Pipelines   []PatternInstance `json:"pipelines"`
	FanOut      []PatternInstance `json:"fan_out"`
	FanIn       []PatternInstance `json:"fan_in"`
	Semaphores  []PatternInstance `json:"semaphores"`
	Goroutines  GoroutineMetrics  `json:"goroutines"`
	Channels    ChannelMetrics    `json:"channels"`
	SyncPrims   SyncPrimitives    `json:"sync_primitives"`
}

// GoroutineMetrics tracks goroutine usage patterns including total count, anonymous vs named, and leak warnings.
type GoroutineMetrics struct {
	TotalCount     int                    `json:"total_count"`
	AnonymousCount int                    `json:"anonymous_count"`
	NamedCount     int                    `json:"named_count"`
	GoroutineLeaks []GoroutineLeakWarning `json:"potential_leaks"`
	Instances      []GoroutineInstance    `json:"instances"`
}

// ChannelMetrics tracks channel usage patterns including buffered, unbuffered,
// ChannelMetrics includes directional channel counts along with detailed instance information.
type ChannelMetrics struct {
	TotalCount       int               `json:"total_count"`
	BufferedCount    int               `json:"buffered_count"`
	UnbufferedCount  int               `json:"unbuffered_count"`
	DirectionalCount int               `json:"directional_count"`
	Instances        []ChannelInstance `json:"instances"`
}

// SyncPrimitives tracks synchronization primitive usage including mutexes, wait groups, and atomic operations.
type SyncPrimitives struct {
	Mutexes    []SyncPrimitiveInstance `json:"mutexes"`
	RWMutexes  []SyncPrimitiveInstance `json:"rw_mutexes"`
	WaitGroups []SyncPrimitiveInstance `json:"wait_groups"`
	Once       []SyncPrimitiveInstance `json:"once"`
	Cond       []SyncPrimitiveInstance `json:"cond"`
	Atomic     []SyncPrimitiveInstance `json:"atomic"`
}

// GoroutineInstance represents a goroutine usage
type GoroutineInstance struct {
	File        string `json:"file"`
	Line        int    `json:"line"`
	Function    string `json:"function"`
	IsAnonymous bool   `json:"is_anonymous"`
	HasDefer    bool   `json:"has_defer"`
	Context     string `json:"context"`
}

// GoroutineLeakWarning represents a potential goroutine leak
type GoroutineLeakWarning struct {
	File           string `json:"file"`
	Line           int    `json:"line"`
	Function       string `json:"function"`
	RiskLevel      string `json:"risk_level"`
	Description    string `json:"description"`
	Recommendation string `json:"recommendation"`
}

// ChannelInstance represents a channel usage
type ChannelInstance struct {
	File          string `json:"file"`
	Line          int    `json:"line"`
	Function      string `json:"function"`
	Type          string `json:"type"`
	IsBuffered    bool   `json:"is_buffered"`
	BufferSize    int    `json:"buffer_size"`
	IsDirectional bool   `json:"is_directional"`
	Direction     string `json:"direction"`
}

// SyncPrimitiveInstance represents a synchronization primitive usage
type SyncPrimitiveInstance struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function"`
	Type     string `json:"type"`
	Variable string `json:"variable"`
	Context  string `json:"context"`
}

// AntiPatternMetrics tracks code smells and anti-patterns
type AntiPatternMetrics struct {
	GodObjects              []AntiPatternWarning     `json:"god_objects"`
	LongMethods             []AntiPatternWarning     `json:"long_methods"`
	DeepNesting             []AntiPatternWarning     `json:"deep_nesting"`
	MagicNumbers            []AntiPatternWarning     `json:"magic_numbers"`
	PerformanceAntipatterns []PerformanceAntipattern `json:"performance_antipatterns"`
}

// PatternInstance represents a detected pattern
type PatternInstance struct {
	Name            string  `json:"name"`
	File            string  `json:"file"`
	Line            int     `json:"line"`
	ConfidenceScore float64 `json:"confidence_score"`
	Description     string  `json:"description"`
	Example         string  `json:"example"`
}

// AntiPatternWarning represents a detected anti-pattern
type AntiPatternWarning struct {
	Type           string  `json:"type"`
	File           string  `json:"file"`
	Line           int     `json:"line"`
	Function       string  `json:"function"`
	Severity       string  `json:"severity"`
	Description    string  `json:"description"`
	Recommendation string  `json:"recommendation"`
	ItemName       string  `json:"item_name,omitempty"`
	Metric         string  `json:"metric,omitempty"`
	ActualValue    float64 `json:"actual_value,omitempty"`
	Threshold      float64 `json:"threshold,omitempty"`
}

// ComplexityMetrics provides overall complexity analysis
type ComplexityMetrics struct {
	AverageFunction   float64          `json:"average_function_complexity"`
	AverageStruct     float64          `json:"average_struct_complexity"`
	HighestComplexity []ComplexityItem `json:"highest_complexity"`
	Distribution      map[string]int   `json:"complexity_distribution"`
}

// ComplexityItem represents a high-complexity item
type ComplexityItem struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	File        string  `json:"file"`
	Line        int     `json:"line"`
	Complexity  float64 `json:"complexity"`
	Severity    string  `json:"severity,omitempty"`
	Suggestion  string  `json:"suggestion,omitempty"`
	ItemName    string  `json:"item_name,omitempty"`
	Metric      string  `json:"metric,omitempty"`
	ActualValue float64 `json:"actual_value,omitempty"`
	Threshold   float64 `json:"threshold,omitempty"`
}

// DocumentationMetrics contains documentation quality analysis
type DocumentationMetrics struct {
	Coverage              DocumentationCoverage `json:"coverage"`
	Quality               DocumentationQuality  `json:"quality"`
	TODOComments          []TODOComment         `json:"todo_comments"`
	FIXMEComments         []FIXMEComment        `json:"fixme_comments"`
	HACKComments          []HACKComment         `json:"hack_comments"`
	BUGComments           []BUGComment          `json:"bug_comments"`
	XXXComments           []XXXComment          `json:"xxx_comments"`
	DEPRECATEDComments    []DEPRECATEDComment   `json:"deprecated_comments"`
	NOTEComments          []NOTEComment         `json:"note_comments"`
	StaleAnnotations      int                   `json:"stale_annotations"`
	AnnotationsByCategory map[string]int        `json:"annotations_by_category"`
}

// DocumentationCoverage tracks GoDoc coverage percentages for packages,
// DocumentationCoverage includes functions, types, and methods, plus an overall weighted coverage score.
type DocumentationCoverage struct {
	Packages  float64 `json:"packages"`
	Functions float64 `json:"functions"`
	Types     float64 `json:"types"`
	Methods   float64 `json:"methods"`
	Overall   float64 `json:"overall"`
}

// DocumentationQuality tracks comment quality metrics
type DocumentationQuality struct {
	AverageLength  float64 `json:"average_length"`
	CodeExamples   int     `json:"code_examples"`
	InlineComments int     `json:"inline_comments"`
	BlockComments  int     `json:"block_comments"`
	QualityScore   float64 `json:"quality_score"`
}

// TODOComment represents a TODO comment
type TODOComment struct {
	File        string `json:"file"`
	Line        int    `json:"line"`
	Author      string `json:"author,omitempty"`
	Description string `json:"description"`
}

// FIXMEComment represents a FIXME comment
type FIXMEComment struct {
	File        string `json:"file"`
	Line        int    `json:"line"`
	Author      string `json:"author,omitempty"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
}

// HACKComment represents a HACK comment
type HACKComment struct {
	File        string `json:"file"`
	Line        int    `json:"line"`
	Author      string `json:"author,omitempty"`
	Description string `json:"description"`
	Reason      string `json:"reason"`
}

// BUGComment represents a BUG comment
type BUGComment struct {
	File        string `json:"file"`
	Line        int    `json:"line"`
	Author      string `json:"author,omitempty"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
}

// XXXComment represents a XXX comment
type XXXComment struct {
	File        string `json:"file"`
	Line        int    `json:"line"`
	Author      string `json:"author,omitempty"`
	Description string `json:"description"`
}

// DEPRECATEDComment represents a DEPRECATED comment
type DEPRECATEDComment struct {
	File        string `json:"file"`
	Line        int    `json:"line"`
	Author      string `json:"author,omitempty"`
	Description string `json:"description"`
	Alternative string `json:"alternative,omitempty"`
}

// NOTEComment represents a NOTE comment
type NOTEComment struct {
	File        string `json:"file"`
	Line        int    `json:"line"`
	Author      string `json:"author,omitempty"`
	Description string `json:"description"`
}

// GenericMetrics contains Go 1.18+ generics analysis
type GenericMetrics struct {
	TypeParameters  GenericTypeParameters `json:"type_parameters"`
	Instantiations  GenericInstantiations `json:"instantiations"`
	ConstraintUsage map[string]int        `json:"constraint_usage"`
	ComplexityScore float64               `json:"complexity_score"`
}

// GenericTypeParameters tracks type parameter usage
type GenericTypeParameters struct {
	Count       int                 `json:"count"`
	Constraints map[string]int      `json:"constraints"`
	Complexity  []GenericComplexity `json:"complexity"`
}

// GenericInstantiations tracks generic instantiations for functions, types, and methods with type parameters.
type GenericInstantiations struct {
	Functions []GenericInstantiation `json:"functions"`
	Types     []GenericInstantiation `json:"types"`
	Methods   []GenericInstantiation `json:"methods"`
}

// GenericComplexity represents generic type complexity
type GenericComplexity struct {
	Name            string  `json:"name"`
	File            string  `json:"file"`
	Line            int     `json:"line"`
	ParameterCount  int     `json:"parameter_count"`
	ConstraintCount int     `json:"constraint_count"`
	ComplexityScore float64 `json:"complexity_score"`
}

// GenericInstantiation represents a generic instantiation
type GenericInstantiation struct {
	GenericName string   `json:"generic_name"`
	TypeArgs    []string `json:"type_args"`
	File        string   `json:"file"`
	Line        int      `json:"line"`
	Usage       string   `json:"usage"`
}

// Duplication Analysis Types

// DuplicationMetrics contains code duplication analysis results for clone pair detection and size tracking.
type DuplicationMetrics struct {
	ClonePairs       int         `json:"clone_pairs"`
	DuplicatedLines  int         `json:"duplicated_lines"`
	DuplicationRatio float64     `json:"duplication_ratio"`
	LargestCloneSize int         `json:"largest_clone_size"`
	Clones           []ClonePair `json:"clones"`
}

// ClonePair represents a set of duplicated code blocks
type ClonePair struct {
	Hash       string          `json:"hash"`
	Type       CloneType       `json:"type"`
	Instances  []CloneInstance `json:"instances"`
	LineCount  int             `json:"line_count"`
	Severity   string          `json:"severity"`
	Suggestion string          `json:"suggestion"`
	ItemName   string          `json:"item_name,omitempty"`
	Metric     string          `json:"metric,omitempty"`
	ActualValue float64        `json:"actual_value,omitempty"`
	Threshold   float64        `json:"threshold,omitempty"`
}

// CloneInstance represents a single instance of duplicated code
type CloneInstance struct {
	File      string `json:"file"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	NodeCount int    `json:"node_count"`
}

// CloneType represents the category of code duplication
type CloneType string

const (
	CloneTypeExact   CloneType = "exact"   // Type 1: exact duplicates
	CloneTypeRenamed CloneType = "renamed" // Type 2: same structure, different identifiers
	CloneTypeNear    CloneType = "near"    // Type 3: similar structure above threshold
)

// NamingMetrics contains naming convention analysis results
type NamingMetrics struct {
	FileNameViolations    int                    `json:"file_name_violations"`
	IdentifierViolations  int                    `json:"identifier_violations"`
	PackageNameViolations int                    `json:"package_name_violations"`
	OverallNamingScore    float64                `json:"overall_naming_score"`
	FileNameIssues        []FileNameViolation    `json:"file_name_issues"`
	IdentifierIssues      []IdentifierViolation  `json:"identifier_issues"`
	PackageNameIssues     []PackageNameViolation `json:"package_name_issues"`
}

// PlacementMetrics contains misplaced declaration analysis results
type PlacementMetrics struct {
	MisplacedFunctions int                      `json:"misplaced_functions"`
	MisplacedMethods   int                      `json:"misplaced_methods"`
	LowCohesionFiles   int                      `json:"low_cohesion_files"`
	AvgFileCohesion    float64                  `json:"avg_file_cohesion"`
	FunctionIssues     []MisplacedFunctionIssue `json:"function_issues"`
	MethodIssues       []MisplacedMethodIssue   `json:"method_issues"`
	CohesionIssues     []FileCohesionIssue      `json:"cohesion_issues"`
}

// MisplacedFunctionIssue represents a function that may be better placed in another file
type MisplacedFunctionIssue struct {
	Name              string   `json:"name"`
	CurrentFile       string   `json:"current_file"`
	SuggestedFile     string   `json:"suggested_file"`
	CurrentAffinity   float64  `json:"current_affinity"`
	SuggestedAffinity float64  `json:"suggested_affinity"`
	ReferencedSymbols []string `json:"referenced_symbols"`
	Severity          string   `json:"severity"`
	Suggestion        string   `json:"suggestion,omitempty"`
	ItemName          string   `json:"item_name,omitempty"`
	Metric            string   `json:"metric,omitempty"`
	ActualValue       float64  `json:"actual_value,omitempty"`
	Threshold         float64  `json:"threshold,omitempty"`
}

// MisplacedMethodIssue represents a method defined away from its receiver type
type MisplacedMethodIssue struct {
	MethodName   string  `json:"method_name"`
	ReceiverType string  `json:"receiver_type"`
	CurrentFile  string  `json:"current_file"`
	ReceiverFile string  `json:"receiver_file"`
	Distance     string  `json:"distance"` // "same_package" or "different_package"
	Severity     string  `json:"severity"`
	Suggestion   string  `json:"suggestion,omitempty"`
	ItemName     string  `json:"item_name,omitempty"`
	Metric       string  `json:"metric,omitempty"`
	ActualValue  float64 `json:"actual_value,omitempty"`
	Threshold    float64 `json:"threshold,omitempty"`
}

// FileCohesionIssue represents a file with low internal cohesion
type FileCohesionIssue struct {
	File            string   `json:"file"`
	CohesionScore   float64  `json:"cohesion_score"`
	IntraFileRefs   int      `json:"intra_file_refs"`
	TotalRefs       int      `json:"total_refs"`
	SuggestedSplits []string `json:"suggested_splits"`
	Severity        string   `json:"severity"`
	Suggestion      string   `json:"suggestion,omitempty"`
	ItemName        string   `json:"item_name,omitempty"`
	Metric          string   `json:"metric,omitempty"`
	ActualValue     float64  `json:"actual_value,omitempty"`
	Threshold       float64  `json:"threshold,omitempty"`
}

// FileNameViolation represents a file naming convention violation
type FileNameViolation struct {
	File          string  `json:"file"`
	ViolationType string  `json:"violation_type"`
	Description   string  `json:"description"`
	SuggestedName string  `json:"suggested_name"`
	Severity      string  `json:"severity"`
	ItemName      string  `json:"item_name,omitempty"`
	Metric        string  `json:"metric,omitempty"`
	ActualValue   float64 `json:"actual_value,omitempty"`
	Threshold     float64 `json:"threshold,omitempty"`
}

// IdentifierViolation represents an identifier naming convention violation
type IdentifierViolation struct {
	Name          string  `json:"name"`
	File          string  `json:"file"`
	Line          int     `json:"line"`
	Type          string  `json:"type"` // function, method, type, const, var
	ViolationType string  `json:"violation_type"`
	Description   string  `json:"description"`
	SuggestedName string  `json:"suggested_name"`
	Severity      string  `json:"severity"`
	ItemName      string  `json:"item_name,omitempty"`
	Metric        string  `json:"metric,omitempty"`
	ActualValue   float64 `json:"actual_value,omitempty"`
	Threshold     float64 `json:"threshold,omitempty"`
}

// PackageNameViolation represents a package naming convention violation
type PackageNameViolation struct {
	Package       string  `json:"package"`
	Directory     string  `json:"directory"`
	ViolationType string  `json:"violation_type"`
	Description   string  `json:"description"`
	SuggestedName string  `json:"suggested_name"`
	Severity      string  `json:"severity"`
	ItemName      string  `json:"item_name,omitempty"`
	Metric        string  `json:"metric,omitempty"`
	ActualValue   float64 `json:"actual_value,omitempty"`
	Threshold     float64 `json:"threshold,omitempty"`
}

// Organization Analysis Types

// OrganizationMetrics contains organizational structure and health analysis for package stability assessment.
type OrganizationMetrics struct {
	OversizedFiles      []OversizedFile    `json:"oversized_files"`
	OversizedPackages   []OversizedPackage `json:"oversized_packages"`
	DeepDirectories     []DeepDirectory    `json:"deep_directories"`
	HighFanInPackages   []FanInPackage     `json:"high_fan_in_packages"`
	HighFanOutPackages  []FanOutPackage    `json:"high_fan_out_packages"`
	AvgPackageStability float64            `json:"avg_package_instability"`
}

// OversizedFile represents a file that exceeds recommended size thresholds
type OversizedFile struct {
	File              string      `json:"file"`
	Lines             LineMetrics `json:"lines"`
	FunctionCount     int         `json:"function_count"`
	TypeCount         int         `json:"type_count"`
	MaintenanceBurden float64     `json:"maintenance_burden"`
	Severity          string      `json:"severity"`
	Suggestions       []string    `json:"suggestions"`
}

// OversizedPackage represents a package that may be too large
type OversizedPackage struct {
	Package         string   `json:"package"`
	FileCount       int      `json:"file_count"`
	ExportedSymbols int      `json:"exported_symbols"`
	TotalFunctions  int      `json:"total_functions"`
	CohesionScore   float64  `json:"cohesion_score"`
	IsMegaPackage   bool     `json:"is_mega_package"`
	Severity        string   `json:"severity"`
	Suggestions     []string `json:"suggestions"`
}

// DeepDirectory represents a directory structure that may be too nested
type DeepDirectory struct {
	Path       string `json:"path"`
	Depth      int    `json:"depth"`
	FileCount  int    `json:"file_count"`
	Severity   string `json:"severity"`
	Suggestion string `json:"suggestion"`
}

// FanInPackage represents a package with high incoming dependencies (hub)
type FanInPackage struct {
	Package      string   `json:"package"`
	FanIn        int      `json:"fan_in"`
	Dependents   []string `json:"dependents"`
	IsBottleneck bool     `json:"is_bottleneck"`
	RiskLevel    string   `json:"risk_level"`
	Suggestion   string   `json:"suggestion"`
}

// FanOutPackage represents a package with high outgoing dependencies (authority)
type FanOutPackage struct {
	Package      string   `json:"package"`
	FanOut       int      `json:"fan_out"`
	Dependencies []string `json:"dependencies"`
	Instability  float64  `json:"instability"`
	CouplingRisk string   `json:"coupling_risk"`
	Suggestion   string   `json:"suggestion"`
}

// Diff and Historical Analysis Types

// Snapshot represents a complete snapshot of code metrics at a point in time for baseline comparison.
type Snapshot struct {
	ID       string           `json:"id"`
	Report   Report           `json:"report"`
	Metadata SnapshotMetadata `json:"metadata"`
}

// SnapshotMetadata contains versioning and context information for a metrics snapshot
type SnapshotMetadata struct {
	Timestamp   time.Time         `json:"timestamp"`
	GitCommit   string            `json:"git_commit,omitempty"`
	GitBranch   string            `json:"git_branch,omitempty"`
	GitTag      string            `json:"git_tag,omitempty"`
	Version     string            `json:"version,omitempty"`
	Author      string            `json:"author,omitempty"`
	Description string            `json:"description,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// ComplexityDiff represents comprehensive diff between two metrics snapshots
type ComplexityDiff struct {
	Baseline     Snapshot        `json:"baseline"`
	Current      Snapshot        `json:"current"`
	Summary      DiffSummary     `json:"summary"`
	Changes      []MetricChange  `json:"changes"`
	Regressions  []Regression    `json:"regressions"`
	Improvements []Improvement   `json:"improvements"`
	Timestamp    time.Time       `json:"timestamp"`
	Config       ThresholdConfig `json:"config"`
}

// DiffSummary provides high-level overview of changes between snapshots
type DiffSummary struct {
	TotalChanges       int            `json:"total_changes"`
	SignificantChanges int            `json:"significant_changes"`
	RegressionCount    int            `json:"regression_count"`
	ImprovementCount   int            `json:"improvement_count"`
	NeutralChangeCount int            `json:"neutral_change_count"`
	CriticalIssues     int            `json:"critical_issues"`
	OverallTrend       TrendDirection `json:"overall_trend"`
	QualityScore       float64        `json:"quality_score"`
	QualityDelta       float64        `json:"quality_delta"`
}

// MetricChange represents a single metric change between snapshots
type MetricChange struct {
	Category    string        `json:"category"`
	Name        string        `json:"name"`
	Path        string        `json:"path"`
	File        string        `json:"file"`
	Line        int           `json:"line,omitempty"`
	OldValue    interface{}   `json:"old_value"`
	NewValue    interface{}   `json:"new_value"`
	Delta       Delta         `json:"delta"`
	Impact      ImpactLevel   `json:"impact"`
	Severity    SeverityLevel `json:"severity"`
	Description string        `json:"description"`
	Suggestion  string        `json:"suggestion,omitempty"`
}

// Delta represents quantified change between two values
type Delta struct {
	Absolute    float64         `json:"absolute"`
	Percentage  float64         `json:"percentage"`
	Direction   ChangeDirection `json:"direction"`
	Significant bool            `json:"significant"`
	Magnitude   ChangeMagnitude `json:"magnitude"`
}

// Regression represents a negative change that exceeds thresholds
type Regression struct {
	Type        RegressionType `json:"type"`
	Location    string         `json:"location"`
	File        string         `json:"file"`
	Line        int            `json:"line,omitempty"`
	Function    string         `json:"function,omitempty"`
	Description string         `json:"description"`
	OldValue    interface{}    `json:"old_value"`
	NewValue    interface{}    `json:"new_value"`
	Delta       Delta          `json:"delta"`
	Impact      ImpactLevel    `json:"impact"`
	Severity    SeverityLevel  `json:"severity"`
	Threshold   float64        `json:"threshold"`
	Suggestion  string         `json:"suggestion"`
	Priority    int            `json:"priority"` // 1-10, higher = more urgent
}

// Improvement represents a positive change
type Improvement struct {
	Type        ImprovementType `json:"type"`
	Location    string          `json:"location"`
	File        string          `json:"file"`
	Line        int             `json:"line,omitempty"`
	Function    string          `json:"function,omitempty"`
	Description string          `json:"description"`
	OldValue    interface{}     `json:"old_value"`
	NewValue    interface{}     `json:"new_value"`
	Delta       Delta           `json:"delta"`
	Impact      ImpactLevel     `json:"impact"`
	Benefit     string          `json:"benefit"`
}

// Enum types for classification

// ChangeDirection represents the direction of a metric change between snapshots for trend analysis.
type ChangeDirection string

const (
	ChangeDirectionIncrease ChangeDirection = "increase"
	ChangeDirectionDecrease ChangeDirection = "decrease"
	ChangeDirectionNeutral  ChangeDirection = "neutral"
)

// ChangeMagnitude represents the relative size of a metric change.
type ChangeMagnitude string

const (
	ChangeMagnitudeMinor       ChangeMagnitude = "minor"
	ChangeMagnitudeModerate    ChangeMagnitude = "moderate"
	ChangeMagnitudeSignificant ChangeMagnitude = "significant"
	ChangeMagnitudeMajor       ChangeMagnitude = "major"
	ChangeMagnitudeCritical    ChangeMagnitude = "critical"
)

// ImpactLevel represents the potential impact severity of a change or issue.
type ImpactLevel string

const (
	ImpactLevelLow      ImpactLevel = "low"
	ImpactLevelMedium   ImpactLevel = "medium"
	ImpactLevelHigh     ImpactLevel = "high"
	ImpactLevelCritical ImpactLevel = "critical"
)

// SeverityLevel represents the severity classification for issues and warnings.
type SeverityLevel string

const (
	SeverityLevelInfo     SeverityLevel = "info"
	SeverityLevelWarning  SeverityLevel = "warning"
	SeverityLevelError    SeverityLevel = "error"
	SeverityLevelCritical SeverityLevel = "critical"
)

// TrendDirection represents the overall direction of a metric trend over time.
type TrendDirection string

const (
	TrendImproving TrendDirection = "improving"
	TrendStable    TrendDirection = "stable"
	TrendDegrading TrendDirection = "degrading"
	TrendVolatile  TrendDirection = "volatile"
)

// RegressionType categorizes the type of code quality regression detected.
type RegressionType string

const (
	ComplexityRegression    RegressionType = "complexity_increase"
	SizeRegression          RegressionType = "size_increase"
	CouplingRegression      RegressionType = "coupling_increase"
	CohesionRegression      RegressionType = "cohesion_decrease"
	CoverageRegression      RegressionType = "coverage_decrease"
	PatternRegression       RegressionType = "anti_pattern_introduction"
	DocumentationRegression RegressionType = "documentation_decrease"
	PerformanceRegression   RegressionType = "performance_decrease"
	BurdenRegression        RegressionType = "burden_increase"
	DuplicationRegression   RegressionType = "duplication_increase"
	NamingRegression        RegressionType = "naming_violations_increase"
)

// ImprovementType categorizes the type of code quality improvement detected.
type ImprovementType string

const (
	ComplexityImprovement    ImprovementType = "complexity_decrease"
	SizeImprovement          ImprovementType = "size_decrease"
	CouplingImprovement      ImprovementType = "coupling_decrease"
	CohesionImprovement      ImprovementType = "cohesion_increase"
	CoverageImprovement      ImprovementType = "coverage_increase"
	PatternImprovement       ImprovementType = "pattern_introduction"
	DocumentationImprovement ImprovementType = "documentation_increase"
	PerformanceImprovement   ImprovementType = "performance_increase"
)

// ThresholdConfig defines configurable thresholds for change detection
type ThresholdConfig struct {
	FunctionComplexity struct {
		Warning     int     `yaml:"warning" json:"warning"`
		Error       int     `yaml:"error" json:"error"`
		MaxIncrease float64 `yaml:"max_increase_percent" json:"max_increase_percent"`
		MinDecrease float64 `yaml:"min_decrease_percent" json:"min_decrease_percent"`
	} `yaml:"function_complexity" json:"function_complexity"`

	StructComplexity struct {
		MaxFields      int     `yaml:"max_fields" json:"max_fields"`
		FieldIncrease  float64 `yaml:"max_field_increase" json:"max_field_increase"`
		MethodIncrease float64 `yaml:"max_method_increase" json:"max_method_increase"`
	} `yaml:"struct_complexity" json:"struct_complexity"`

	PackageMetrics struct {
		MaxCoupling     float64 `yaml:"max_coupling" json:"max_coupling"`
		MinCohesion     float64 `yaml:"min_cohesion" json:"min_cohesion"`
		MaxDependencies int     `yaml:"max_dependencies" json:"max_dependencies"`
	} `yaml:"package_metrics" json:"package_metrics"`

	Documentation struct {
		MinCoverage float64 `yaml:"min_coverage" json:"min_coverage"`
		MaxDecrease float64 `yaml:"max_decrease_percent" json:"max_decrease_percent"`
	} `yaml:"documentation" json:"documentation"`

	BurdenMetrics struct {
		FileMBIThreshold    float64 `yaml:"file_mbi_threshold" json:"file_mbi_threshold"`
		PackageMBIThreshold float64 `yaml:"package_mbi_threshold" json:"package_mbi_threshold"`
		MaxDuplicationRatio float64 `yaml:"max_duplication_ratio" json:"max_duplication_ratio"`
		MaxNamingViolations int     `yaml:"max_naming_violations" json:"max_naming_violations"`
	} `yaml:"burden_metrics" json:"burden_metrics"`

	Global struct {
		MaxRegressions    int     `yaml:"max_regressions" json:"max_regressions"`
		FailOnError       bool    `yaml:"fail_on_error" json:"fail_on_error"`
		FailOnCritical    bool    `yaml:"fail_on_critical" json:"fail_on_critical"`
		SignificanceLevel float64 `yaml:"significance_level" json:"significance_level"`
	} `yaml:"global" json:"global"`
}

// DefaultThresholdConfig returns sensible default thresholds for code quality assessment based on industry best practices.
// Includes cyclomatic complexity <= 10, function length <= 30 lines, documentation coverage >= 70%, and duplication ratio < 5%.
// These thresholds represent a balanced approach between strict quality enforcement and practical development workflows.
func DefaultThresholdConfig() ThresholdConfig {
	config := ThresholdConfig{}

	config.FunctionComplexity.Warning = 10
	config.FunctionComplexity.Error = 20
	config.FunctionComplexity.MaxIncrease = 25.0
	config.FunctionComplexity.MinDecrease = 10.0

	config.StructComplexity.MaxFields = 15
	config.StructComplexity.FieldIncrease = 30.0
	config.StructComplexity.MethodIncrease = 20.0

	config.PackageMetrics.MaxCoupling = 0.75
	config.PackageMetrics.MinCohesion = 0.60
	config.PackageMetrics.MaxDependencies = 20

	config.Documentation.MinCoverage = 70.0
	config.Documentation.MaxDecrease = 15.0

	config.BurdenMetrics.FileMBIThreshold = 10.0
	config.BurdenMetrics.PackageMBIThreshold = 5.0
	config.BurdenMetrics.MaxDuplicationRatio = 0.10
	config.BurdenMetrics.MaxNamingViolations = 10

	config.Global.MaxRegressions = 5
	config.Global.FailOnError = true
	config.Global.FailOnCritical = true
	config.Global.SignificanceLevel = 5.0

	return config
}

// Trend Analysis Types

// TrendAnalysis represents trend analysis across multiple data points for forecasting and velocity tracking.
type TrendAnalysis struct {
	Metric       string           `json:"metric"`
	Direction    TrendDirection   `json:"direction"`
	Velocity     float64          `json:"velocity"`
	Acceleration float64          `json:"acceleration"`
	Confidence   float64          `json:"confidence"`
	Forecast     ForecastPoint    `json:"forecast"`
	DataPoints   []TrendPoint     `json:"data_points"`
	Regression   LinearRegression `json:"regression"`
}

// TrendPoint represents a single data point in a trend
type TrendPoint struct {
	Timestamp time.Time   `json:"timestamp"`
	Value     float64     `json:"value"`
	Metadata  interface{} `json:"metadata,omitempty"`
}

// ForecastPoint represents a forecasted future value
type ForecastPoint struct {
	Timestamp  time.Time `json:"timestamp"`
	Value      float64   `json:"value"`
	Confidence float64   `json:"confidence"`
	Upper      float64   `json:"upper_bound"`
	Lower      float64   `json:"lower_bound"`
}

// LinearRegression represents simple linear regression results
type LinearRegression struct {
	Slope       float64 `json:"slope"`
	Intercept   float64 `json:"intercept"`
	RSquared    float64 `json:"r_squared"`
	StdError    float64 `json:"std_error"`
	Significant bool    `json:"significant"`
}

// ChangeGranularity defines what aspects of code to track for changes
type ChangeGranularity struct {
	Function struct {
		LineCount     bool `json:"track_line_count"`
		Complexity    bool `json:"track_complexity"`
		Parameters    bool `json:"track_parameters"`
		Returns       bool `json:"track_returns"`
		Documentation bool `json:"track_documentation"`
	} `json:"function"`

	Struct struct {
		FieldCount    bool `json:"track_field_count"`
		FieldTypes    bool `json:"track_field_types"`
		Methods       bool `json:"track_methods"`
		Embedding     bool `json:"track_embedding"`
		Documentation bool `json:"track_documentation"`
	} `json:"struct"`

	Package struct {
		Dependencies  bool `json:"track_dependencies"`
		Cohesion      bool `json:"track_cohesion"`
		Coupling      bool `json:"track_coupling"`
		Coverage      bool `json:"track_coverage"`
		Documentation bool `json:"track_documentation"`
	} `json:"package"`
}

// DefaultChangeGranularity returns default tracking settings for differential analysis and change detection.
// Determines the sensitivity thresholds for flagging complexity increases, duplication growth, and documentation decreases.
// Lower granularity values make the diff engine more sensitive to minor changes, higher values focus on significant regressions.
func DefaultChangeGranularity() ChangeGranularity {
	granularity := ChangeGranularity{}

	granularity.Function.LineCount = true
	granularity.Function.Complexity = true
	granularity.Function.Parameters = true
	granularity.Function.Returns = true
	granularity.Function.Documentation = true

	granularity.Struct.FieldCount = true
	granularity.Struct.FieldTypes = true
	granularity.Struct.Methods = true
	granularity.Struct.Embedding = true
	granularity.Struct.Documentation = true

	granularity.Package.Dependencies = true
	granularity.Package.Cohesion = true
	granularity.Package.Coupling = true
	granularity.Package.Coverage = true
	granularity.Package.Documentation = true

	return granularity
}

// BurdenMetrics contains maintenance burden indicators
type BurdenMetrics struct {
	MagicNumbers          []MagicNumber      `json:"magic_numbers"`
	DeadCode              DeadCodeMetrics    `json:"dead_code"`
	ComplexSignatures     []SignatureIssue   `json:"complex_signatures"`
	DeeplyNestedFunctions []NestingIssue     `json:"deeply_nested_functions"`
	FeatureEnvyMethods    []FeatureEnvyIssue `json:"feature_envy_methods"`
}

// MagicNumber represents a detected magic number or string
type MagicNumber struct {
	File        string  `json:"file"`
	Line        int     `json:"line"`
	Column      int     `json:"column"`
	Value       string  `json:"value"`
	Type        string  `json:"type"`
	Context     string  `json:"context"`
	Function    string  `json:"function"`
	Severity    string  `json:"severity"`
	Suggestion  string  `json:"suggestion"`
	ItemName    string  `json:"item_name,omitempty"`
	Metric      string  `json:"metric,omitempty"`
	ActualValue float64 `json:"actual_value,omitempty"`
	Threshold   float64 `json:"threshold,omitempty"`
}

// DeadCodeMetrics contains dead code detection results
type DeadCodeMetrics struct {
	UnreferencedFunctions []UnreferencedSymbol `json:"unreferenced_functions"`
	UnreachableCode       []UnreachableBlock   `json:"unreachable_code"`
	TotalDeadLines        int                  `json:"total_dead_lines"`
	DeadCodePercent       float64              `json:"dead_code_percent"`
}

// UnreferencedSymbol represents an unreferenced unexported symbol
type UnreferencedSymbol struct {
	Name        string  `json:"name"`
	File        string  `json:"file"`
	Line        int     `json:"line"`
	Type        string  `json:"type"`
	Package     string  `json:"package"`
	Severity    string  `json:"severity,omitempty"`
	Suggestion  string  `json:"suggestion,omitempty"`
	ItemName    string  `json:"item_name,omitempty"`
	Metric      string  `json:"metric,omitempty"`
	ActualValue float64 `json:"actual_value,omitempty"`
	Threshold   float64 `json:"threshold,omitempty"`
}

// UnreachableBlock represents unreachable code after control flow statements
type UnreachableBlock struct {
	File        string  `json:"file"`
	StartLine   int     `json:"start_line"`
	EndLine     int     `json:"end_line"`
	Function    string  `json:"function"`
	Reason      string  `json:"reason"`
	Lines       int     `json:"lines"`
	Severity    string  `json:"severity,omitempty"`
	Suggestion  string  `json:"suggestion,omitempty"`
	ItemName    string  `json:"item_name,omitempty"`
	Metric      string  `json:"metric,omitempty"`
	ActualValue float64 `json:"actual_value,omitempty"`
	Threshold   float64 `json:"threshold,omitempty"`
}

// SignatureIssue represents a function with excessive parameters or returns
type SignatureIssue struct {
	Function       string   `json:"function"`
	File           string   `json:"file"`
	Line           int      `json:"line"`
	ParameterCount int      `json:"parameter_count"`
	ReturnCount    int      `json:"return_count"`
	BoolParams     []string `json:"bool_params,omitempty"`
	Severity       string   `json:"severity"`
	Suggestion     string   `json:"suggestion"`
	ItemName       string   `json:"item_name,omitempty"`
	Metric         string   `json:"metric,omitempty"`
	ActualValue    float64  `json:"actual_value,omitempty"`
	Threshold      float64  `json:"threshold,omitempty"`
}

// NestingIssue represents deep nesting in a function
type NestingIssue struct {
	Function    string  `json:"function"`
	File        string  `json:"file"`
	Line        int     `json:"line"`
	MaxDepth    int     `json:"max_depth"`
	Location    string  `json:"location"`
	Severity    string  `json:"severity"`
	Suggestion  string  `json:"suggestion"`
	ItemName    string  `json:"item_name,omitempty"`
	Metric      string  `json:"metric,omitempty"`
	ActualValue float64 `json:"actual_value,omitempty"`
	Threshold   float64 `json:"threshold,omitempty"`
}

// FeatureEnvyIssue represents a method with excessive external references
type FeatureEnvyIssue struct {
	Method         string  `json:"method"`
	File           string  `json:"file"`
	Line           int     `json:"line"`
	ReceiverType   string  `json:"receiver_type"`
	SelfReferences int     `json:"self_references"`
	ExternalType   string  `json:"external_type"`
	ExternalRefs   int     `json:"external_references"`
	Ratio          float64 `json:"ratio"`
	Severity       string  `json:"severity"`
	SuggestedMove  string  `json:"suggested_move"`
	ItemName       string  `json:"item_name,omitempty"`
	Metric         string  `json:"metric,omitempty"`
	ActualValue    float64 `json:"actual_value,omitempty"`
	Threshold      float64 `json:"threshold,omitempty"`
}

// ScoringMetrics holds maintenance burden index scores
type ScoringMetrics struct {
	FileScores    []FileScore    `json:"file_scores"`
	PackageScores []PackageScore `json:"package_scores"`
}

// FileScore represents the MBI for a single file
type FileScore struct {
	File      string         `json:"file"`
	Score     float64        `json:"score"`
	Risk      string         `json:"risk"`
	Breakdown ScoreBreakdown `json:"breakdown"`
}

// PackageScore represents the MBI for a package
type PackageScore struct {
	Package   string         `json:"package"`
	Score     float64        `json:"score"`
	Risk      string         `json:"risk"`
	Breakdown ScoreBreakdown `json:"breakdown"`
}

// ScoreBreakdown shows contribution of each category
type ScoreBreakdown struct {
	Duplication   float64 `json:"duplication"`
	Naming        float64 `json:"naming"`
	Placement     float64 `json:"placement"`
	Documentation float64 `json:"documentation"`
	Organization  float64 `json:"organization"`
	Burden        float64 `json:"burden"`
}

// SuggestionInfo represents a refactoring suggestion for inclusion in reports
type SuggestionInfo struct {
	Action        string  `json:"action"`
	Target        string  `json:"target"`
	Location      string  `json:"location"`
	Description   string  `json:"description"`
	Effort        string  `json:"effort"`
	MBIImpact     float64 `json:"mbi_impact"`
	ImpactEffort  float64 `json:"impact_effort"`
	Category      string  `json:"category"`
	AffectedLines int     `json:"affected_lines"`
}

// PerformanceAntipattern represents a detected performance anti-pattern
type PerformanceAntipattern struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	File        string `json:"file"`
	Line        int    `json:"line"`
	Suggestion  string `json:"suggestion"`
}

// TestCoverageMetrics represents test coverage correlation analysis
type TestCoverageMetrics struct {
	FunctionCoverageRate   float64            `json:"function_coverage_rate"`
	ComplexityCoverageRate float64            `json:"complexity_coverage_rate"`
	HighRiskFunctions      []HighRiskFunction `json:"high_risk_functions"`
	CoverageGaps           []CoverageGap      `json:"coverage_gaps"`
}

// HighRiskFunction represents a function with high complexity and low coverage
type HighRiskFunction struct {
	Name       string  `json:"name"`
	File       string  `json:"file"`
	Line       int     `json:"line"`
	Complexity int     `json:"complexity"`
	Coverage   float64 `json:"coverage"`
	RiskScore  float64 `json:"risk_score"`
}

// CoverageGap represents a coverage gap in the codebase
type CoverageGap struct {
	Name        string  `json:"name"`
	File        string  `json:"file"`
	Line        int     `json:"line"`
	Complexity  int     `json:"complexity"`
	Coverage    float64 `json:"coverage"`
	GapSeverity string  `json:"gap_severity"`
}

// TestQualityMetrics represents test suite quality assessment
type TestQualityMetrics struct {
	TotalTests           int            `json:"total_tests"`
	AvgAssertionsPerTest float64        `json:"avg_assertions_per_test"`
	TestFiles            []TestFileInfo `json:"test_files"`
}

// TestFileInfo represents test file statistics
type TestFileInfo struct {
	File           string  `json:"file"`
	TestCount      int     `json:"test_count"`
	SubtestCount   int     `json:"subtest_count"`
	AssertionCount int     `json:"assertion_count"`
	AssertionRatio float64 `json:"assertion_ratio"`
}

// TeamMetrics represents team productivity analysis
type TeamMetrics struct {
	Developers      map[string]*DeveloperMetrics `json:"developers"`
	TotalDevelopers int                          `json:"total_developers"`
}

// DeveloperMetrics represents individual contributor stats
type DeveloperMetrics struct {
	Name            string    `json:"name"`
	CommitCount     int       `json:"commit_count"`
	LinesAdded      int       `json:"lines_added"`
	LinesRemoved    int       `json:"lines_removed"`
	FilesModified   int       `json:"files_modified"`
	FirstCommitDate time.Time `json:"first_commit_date"`
	LastCommitDate  time.Time `json:"last_commit_date"`
	ActiveDays      int       `json:"active_days"`
}
