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

// Diff and Historical Analysis Types

// MetricsSnapshot represents a complete snapshot of code metrics at a point in time
type MetricsSnapshot struct {
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
	Baseline     MetricsSnapshot `json:"baseline"`
	Current      MetricsSnapshot `json:"current"`
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

type ChangeDirection string

const (
	ChangeDirectionIncrease ChangeDirection = "increase"
	ChangeDirectionDecrease ChangeDirection = "decrease"
	ChangeDirectionNeutral  ChangeDirection = "neutral"
)

type ChangeMagnitude string

const (
	ChangeMagnitudeMinor       ChangeMagnitude = "minor"
	ChangeMagnitudeModerate    ChangeMagnitude = "moderate"
	ChangeMagnitudeSignificant ChangeMagnitude = "significant"
	ChangeMagnitudeMajor       ChangeMagnitude = "major"
	ChangeMagnitudeCritical    ChangeMagnitude = "critical"
)

type ImpactLevel string

const (
	ImpactLevelLow      ImpactLevel = "low"
	ImpactLevelMedium   ImpactLevel = "medium"
	ImpactLevelHigh     ImpactLevel = "high"
	ImpactLevelCritical ImpactLevel = "critical"
)

type SeverityLevel string

const (
	SeverityLevelInfo     SeverityLevel = "info"
	SeverityLevelWarning  SeverityLevel = "warning"
	SeverityLevelError    SeverityLevel = "error"
	SeverityLevelCritical SeverityLevel = "critical"
)

type TrendDirection string

const (
	TrendImproving TrendDirection = "improving"
	TrendStable    TrendDirection = "stable"
	TrendDegrading TrendDirection = "degrading"
	TrendVolatile  TrendDirection = "volatile"
)

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
)

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

	Global struct {
		MaxRegressions    int     `yaml:"max_regressions" json:"max_regressions"`
		FailOnError       bool    `yaml:"fail_on_error" json:"fail_on_error"`
		FailOnCritical    bool    `yaml:"fail_on_critical" json:"fail_on_critical"`
		SignificanceLevel float64 `yaml:"significance_level" json:"significance_level"`
	} `yaml:"global" json:"global"`
}

// DefaultThresholdConfig returns sensible default thresholds
func DefaultThresholdConfig() ThresholdConfig {
	config := ThresholdConfig{}

	// Function complexity thresholds
	config.FunctionComplexity.Warning = 10
	config.FunctionComplexity.Error = 20
	config.FunctionComplexity.MaxIncrease = 25.0
	config.FunctionComplexity.MinDecrease = 10.0

	// Struct complexity thresholds
	config.StructComplexity.MaxFields = 15
	config.StructComplexity.FieldIncrease = 30.0
	config.StructComplexity.MethodIncrease = 20.0

	// Package metrics thresholds
	config.PackageMetrics.MaxCoupling = 0.75
	config.PackageMetrics.MinCohesion = 0.60
	config.PackageMetrics.MaxDependencies = 20

	// Documentation thresholds
	config.Documentation.MinCoverage = 70.0
	config.Documentation.MaxDecrease = 15.0

	// Global settings
	config.Global.MaxRegressions = 5
	config.Global.FailOnError = true
	config.Global.FailOnCritical = true
	config.Global.SignificanceLevel = 5.0

	return config
}

// Trend Analysis Types

// TrendAnalysis represents trend analysis across multiple data points
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

// DefaultChangeGranularity returns default tracking settings
func DefaultChangeGranularity() ChangeGranularity {
	granularity := ChangeGranularity{}

	// Function tracking (enable all)
	granularity.Function.LineCount = true
	granularity.Function.Complexity = true
	granularity.Function.Parameters = true
	granularity.Function.Returns = true
	granularity.Function.Documentation = true

	// Struct tracking (enable all)
	granularity.Struct.FieldCount = true
	granularity.Struct.FieldTypes = true
	granularity.Struct.Methods = true
	granularity.Struct.Embedding = true
	granularity.Struct.Documentation = true

	// Package tracking (enable all)
	granularity.Package.Dependencies = true
	granularity.Package.Cohesion = true
	granularity.Package.Coupling = true
	granularity.Package.Coverage = true
	granularity.Package.Documentation = true

	return granularity
}
