package metrics

import (
	"time"
)

// Report represents the complete analysis report for a repository
type Report struct {
	Metadata      ReportMetadata       `json:"metadata"`
	Overview      OverviewMetrics      `json:"overview"`
	Functions     []FunctionMetrics    `json:"functions"`
	Structs       []StructMetrics      `json:"structs"`
	Interfaces    []InterfaceMetrics   `json:"interfaces"`
	Packages      []PackageMetrics     `json:"packages"`
	Patterns      PatternMetrics       `json:"patterns"`
	Complexity    ComplexityMetrics    `json:"complexity"`
	Documentation DocumentationMetrics `json:"documentation"`
	Generics      GenericMetrics       `json:"generics"`
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

// OverviewMetrics provides high-level statistics
type OverviewMetrics struct {
	TotalLinesOfCode int `json:"total_lines_of_code"`
	TotalFunctions   int `json:"total_functions"`
	TotalMethods     int `json:"total_methods"`
	TotalStructs     int `json:"total_structs"`
	TotalInterfaces  int `json:"total_interfaces"`
	TotalPackages    int `json:"total_packages"`
	TotalFiles       int `json:"total_files"`
}

// LineMetrics represents line counting information
type LineMetrics struct {
	Total    int `json:"total"`
	Code     int `json:"code"`
	Comments int `json:"comments"`
	Blank    int `json:"blank"`
}

// FunctionMetrics contains detailed function analysis
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

// FunctionSignature represents function signature complexity
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

// ComplexityScore represents various complexity measurements
type ComplexityScore struct {
	Cyclomatic   int     `json:"cyclomatic"`
	Cognitive    int     `json:"cognitive"`
	NestingDepth int     `json:"nesting_depth"`
	Overall      float64 `json:"overall"`
}

// DocumentationInfo contains documentation quality metrics
type DocumentationInfo struct {
	HasComment    bool    `json:"has_comment"`
	CommentLength int     `json:"comment_length"`
	HasExample    bool    `json:"has_example"`
	QualityScore  float64 `json:"quality_score"`
}

// StructMetrics contains detailed struct analysis
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

// MethodInfo represents method information
type MethodInfo struct {
	Name          string            `json:"name"`
	IsExported    bool              `json:"is_exported"`
	IsPointer     bool              `json:"is_pointer_receiver"`
	Signature     FunctionSignature `json:"signature"`
	Lines         LineMetrics       `json:"lines"`
	Complexity    ComplexityScore   `json:"complexity"`
	Documentation DocumentationInfo `json:"documentation"`
}

// InterfaceMetrics contains interface analysis
type InterfaceMetrics struct {
	Name               string            `json:"name"`
	Package            string            `json:"package"`
	File               string            `json:"file"`
	Line               int               `json:"line"`
	IsExported         bool              `json:"is_exported"`
	MethodCount        int               `json:"method_count"`
	Methods            []InterfaceMethod `json:"methods"`
	EmbeddedInterfaces []string          `json:"embedded_interfaces"`
	Implementations    []string          `json:"implementations"`
	Documentation      DocumentationInfo `json:"documentation"`
}

// InterfaceMethod represents a method in an interface
type InterfaceMethod struct {
	Name      string            `json:"name"`
	Signature FunctionSignature `json:"signature"`
}

// PackageMetrics contains package-level analysis
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

// PatternMetrics contains design pattern detection results
type PatternMetrics struct {
	DesignPatterns      DesignPatternMetrics      `json:"design_patterns"`
	ConcurrencyPatterns ConcurrencyPatternMetrics `json:"concurrency_patterns"`
	AntiPatterns        AntiPatternMetrics        `json:"anti_patterns"`
}

// DesignPatternMetrics tracks various design patterns
type DesignPatternMetrics struct {
	Singleton []PatternInstance `json:"singleton"`
	Factory   []PatternInstance `json:"factory"`
	Builder   []PatternInstance `json:"builder"`
	Observer  []PatternInstance `json:"observer"`
	Strategy  []PatternInstance `json:"strategy"`
}

// ConcurrencyPatternMetrics tracks concurrency patterns
type ConcurrencyPatternMetrics struct {
	WorkerPools []PatternInstance `json:"worker_pools"`
	Pipelines   []PatternInstance `json:"pipelines"`
	FanOut      []PatternInstance `json:"fan_out"`
	FanIn       []PatternInstance `json:"fan_in"`
	Semaphores  []PatternInstance `json:"semaphores"`
}

// AntiPatternMetrics tracks code smells and anti-patterns
type AntiPatternMetrics struct {
	GodObjects   []AntiPatternWarning `json:"god_objects"`
	LongMethods  []AntiPatternWarning `json:"long_methods"`
	DeepNesting  []AntiPatternWarning `json:"deep_nesting"`
	MagicNumbers []AntiPatternWarning `json:"magic_numbers"`
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
	Type           string `json:"type"`
	File           string `json:"file"`
	Line           int    `json:"line"`
	Function       string `json:"function"`
	Severity       string `json:"severity"`
	Description    string `json:"description"`
	Recommendation string `json:"recommendation"`
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
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	File       string  `json:"file"`
	Line       int     `json:"line"`
	Complexity float64 `json:"complexity"`
}

// DocumentationMetrics contains documentation quality analysis
type DocumentationMetrics struct {
	Coverage      DocumentationCoverage `json:"coverage"`
	Quality       DocumentationQuality  `json:"quality"`
	TODOComments  []TODOComment         `json:"todo_comments"`
	FIXMEComments []FIXMEComment        `json:"fixme_comments"`
	HACKComments  []HACKComment         `json:"hack_comments"`
}

// DocumentationCoverage tracks GoDoc coverage
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

// GenericInstantiations tracks generic instantiations
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
