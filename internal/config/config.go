package config

import (
	"runtime"
	"time"
)

// Config represents the application configuration
type Config struct {
	Analysis    AnalysisConfig    `mapstructure:"analysis" json:"analysis"`
	Output      OutputConfig      `mapstructure:"output" json:"output"`
	Performance PerformanceConfig `mapstructure:"performance" json:"performance"`
	Filters     FilterConfig      `mapstructure:"filters" json:"filters"`
	Storage     StorageConfig     `mapstructure:"storage" json:"storage"`
}

// AnalysisConfig controls what gets analyzed and defines threshold limits
// AnalysisConfig includes settings for functions, structs, interfaces,
// patterns, complexity, documentation, and generics analysis.
type AnalysisConfig struct {
	IncludeFunctions     bool `mapstructure:"include_functions" json:"include_functions"`
	IncludeStructs       bool `mapstructure:"include_structs" json:"include_structs"`
	IncludeInterfaces    bool `mapstructure:"include_interfaces" json:"include_interfaces"`
	IncludePatterns      bool `mapstructure:"include_patterns" json:"include_patterns"`
	IncludeComplexity    bool `mapstructure:"include_complexity" json:"include_complexity"`
	IncludeDocumentation bool `mapstructure:"include_documentation" json:"include_documentation"`
	IncludeGenerics      bool `mapstructure:"include_generics" json:"include_generics"`
	EnableTeamMetrics    bool `mapstructure:"enable_team_metrics" json:"enable_team_metrics"`

	// Test coverage integration
	CoverageProfile string `mapstructure:"coverage_profile" json:"coverage_profile"`

	// Thresholds for warnings
	MaxFunctionLength        int     `mapstructure:"max_function_length" json:"max_function_length"`
	MaxCyclomaticComplexity  int     `mapstructure:"max_cyclomatic_complexity" json:"max_cyclomatic_complexity"`
	MaxStructFields          int     `mapstructure:"max_struct_fields" json:"max_struct_fields"`
	MinDocumentationCoverage float64 `mapstructure:"min_documentation_coverage" json:"min_documentation_coverage"`
	MaxDuplicationRatio      float64 `mapstructure:"max_duplication_ratio" json:"max_duplication_ratio"`
	MaxUndocumentedExports   int     `mapstructure:"max_undocumented_exports" json:"max_undocumented_exports"`
	EnforceThresholds        bool    `mapstructure:"enforce_thresholds" json:"enforce_thresholds"`

	// Duplication detection settings
	Duplication DuplicationConfig `mapstructure:"duplication" json:"duplication"`

	// Naming convention settings
	Naming NamingConfig `mapstructure:"naming" json:"naming"`

	// Placement analysis settings
	Placement PlacementConfig `mapstructure:"placement" json:"placement"`

	// Documentation analysis settings
	Documentation DocumentationConfig `mapstructure:"documentation" json:"documentation"`

	// Organization analysis settings
	Organization OrganizationConfig `mapstructure:"organization" json:"organization"`

	// Burden analysis settings
	Burden BurdenConfig `mapstructure:"burden" json:"burden"`

	// Scoring weights for MBI calculation
	Scoring ScoringConfig `mapstructure:"scoring" json:"scoring"`
}

// ScoringConfig controls maintenance burden index calculation
type ScoringConfig struct {
	Weights        ScoringWeights `mapstructure:"weights" json:"weights"`
	MaxBurdenScore float64        `mapstructure:"max_burden_score" json:"max_burden_score"`
}

// ScoringWeights defines weights for each maintenance category
type ScoringWeights struct {
	Duplication   float64 `mapstructure:"duplication" json:"duplication"`
	Naming        float64 `mapstructure:"naming" json:"naming"`
	Placement     float64 `mapstructure:"placement" json:"placement"`
	Documentation float64 `mapstructure:"documentation" json:"documentation"`
	Organization  float64 `mapstructure:"organization" json:"organization"`
	Burden        float64 `mapstructure:"burden" json:"burden"`
}

// DuplicationConfig controls code duplication detection
type DuplicationConfig struct {
	MinBlockLines       int     `mapstructure:"min_block_lines" json:"min_block_lines"`
	SimilarityThreshold float64 `mapstructure:"similarity_threshold" json:"similarity_threshold"`
	IgnoreTestFiles     bool    `mapstructure:"ignore_test_files" json:"ignore_test_files"`
}

// NamingConfig controls naming convention analysis
type NamingConfig struct {
	FlagGenericFilenames bool `mapstructure:"flag_generic_filenames" json:"flag_generic_filenames"`
	FlagStuttering       bool `mapstructure:"flag_stuttering" json:"flag_stuttering"`
	MinNameLength        int  `mapstructure:"min_name_length" json:"min_name_length"`
}

// PlacementConfig controls placement and cohesion analysis
type PlacementConfig struct {
	AffinityMargin float64 `mapstructure:"affinity_margin" json:"affinity_margin"`
	MinCohesion    float64 `mapstructure:"min_cohesion" json:"min_cohesion"`
}

// DocumentationConfig controls documentation analysis
type DocumentationConfig struct {
	RequireExportedDoc  bool `mapstructure:"require_exported_doc" json:"require_exported_doc"`
	RequirePackageDoc   bool `mapstructure:"require_package_doc" json:"require_package_doc"`
	StaleAnnotationDays int  `mapstructure:"stale_annotation_days" json:"stale_annotation_days"`
	MinCommentWords     int  `mapstructure:"min_comment_words" json:"min_comment_words"`
}

// OrganizationConfig controls organization and structural analysis
type OrganizationConfig struct {
	MaxFileLines       int `mapstructure:"max_file_lines" json:"max_file_lines"`
	MaxFileFunctions   int `mapstructure:"max_file_functions" json:"max_file_functions"`
	MaxFileTypes       int `mapstructure:"max_file_types" json:"max_file_types"`
	MaxPackageFiles    int `mapstructure:"max_package_files" json:"max_package_files"`
	MaxExportedSymbols int `mapstructure:"max_exported_symbols" json:"max_exported_symbols"`
	MaxDirectoryDepth  int `mapstructure:"max_directory_depth" json:"max_directory_depth"`
	MaxFileImports     int `mapstructure:"max_file_imports" json:"max_file_imports"`
}

// BurdenConfig controls maintenance burden detection
type BurdenConfig struct {
	MaxParams         int     `mapstructure:"max_params" json:"max_params"`
	MaxReturns        int     `mapstructure:"max_returns" json:"max_returns"`
	MaxNesting        int     `mapstructure:"max_nesting" json:"max_nesting"`
	FeatureEnvyRatio  float64 `mapstructure:"feature_envy_ratio" json:"feature_envy_ratio"`
	IgnoreBenignMagic bool    `mapstructure:"ignore_benign_magic" json:"ignore_benign_magic"`
}

// OutputConfig controls output formatting options including format type,
// OutputConfig sets destination file, console display settings, and report content filtering.
type OutputConfig struct {
	Format      OutputFormat `mapstructure:"format" json:"format"`
	Destination string       `mapstructure:"destination" json:"destination"`

	// Console output settings
	UseColors    bool `mapstructure:"use_colors" json:"use_colors"`
	ShowProgress bool `mapstructure:"show_progress" json:"show_progress"`
	Verbose      bool `mapstructure:"verbose" json:"verbose"`

	// Report settings
	IncludeOverview bool   `mapstructure:"include_overview" json:"include_overview"`
	IncludeDetails  bool   `mapstructure:"include_details" json:"include_details"`
	IncludeExamples bool   `mapstructure:"include_examples" json:"include_examples"`
	SortBy          string `mapstructure:"sort_by" json:"sort_by"`
	Limit           int    `mapstructure:"limit" json:"limit"`

	// Section filtering — when non-empty, only listed sections appear in output
	Sections []string `mapstructure:"sections" json:"sections,omitempty"`
}

// OutputFormat represents supported output formats
type OutputFormat string

const (
	FormatConsole  OutputFormat = "console"
	FormatJSON     OutputFormat = "json"
	FormatCSV      OutputFormat = "csv"
	FormatHTML     OutputFormat = "html"
	FormatMarkdown OutputFormat = "markdown"
)

// PerformanceConfig controls performance-related settings
type PerformanceConfig struct {
	WorkerCount int `mapstructure:"worker_count" json:"worker_count"`
	// MaxMemoryMB is reserved for future memory enforcement features (currently not enforced)
	MaxMemoryMB     int           `mapstructure:"max_memory_mb" json:"max_memory_mb"`
	Timeout         time.Duration `mapstructure:"timeout" json:"timeout"`
	EnableProfiling bool          `mapstructure:"enable_profiling" json:"enable_profiling"`

	// Caching
	EnableCache    bool   `mapstructure:"enable_cache" json:"enable_cache"`
	CacheDirectory string `mapstructure:"cache_directory" json:"cache_directory"`
}

// FilterConfig controls what files and packages to analyze
type FilterConfig struct {
	IncludePatterns []string `mapstructure:"include_patterns" json:"include_patterns"`
	ExcludePatterns []string `mapstructure:"exclude_patterns" json:"exclude_patterns"`
	IncludePackages []string `mapstructure:"include_packages" json:"include_packages"`
	ExcludePackages []string `mapstructure:"exclude_packages" json:"exclude_packages"`

	// File size limits
	MaxFileSizeKB int  `mapstructure:"max_file_size_kb" json:"max_file_size_kb"`
	SkipVendor    bool `mapstructure:"skip_vendor" json:"skip_vendor"`
	SkipTestFiles bool `mapstructure:"skip_test_files" json:"skip_test_files"`
	SkipGenerated bool `mapstructure:"skip_generated" json:"skip_generated"`
}

// StorageConfig controls historical metrics storage
type StorageConfig struct {
	Type        string `mapstructure:"type" json:"type"`               // "sqlite", "json", "memory", "postgres"
	Path        string `mapstructure:"path" json:"path"`               // File path for sqlite/json
	Compression bool   `mapstructure:"compression" json:"compression"` // Enable compression for stored data

	// PostgreSQL connection settings
	PostgresConnectionString string `mapstructure:"postgres_connection_string" json:"postgres_connection_string"`

	// Retention policy
	MaxSnapshots int           `mapstructure:"max_snapshots" json:"max_snapshots"`
	MaxAge       time.Duration `mapstructure:"max_age" json:"max_age"`
}

// DefaultConfig returns the default configuration with sensible production values
// DefaultConfig sets analysis, output, performance, filtering, and storage settings.
func DefaultConfig() *Config {
	return &Config{
		Analysis: AnalysisConfig{
			IncludeFunctions:         true,
			IncludeStructs:           true,
			IncludeInterfaces:        true,
			IncludePatterns:          true,
			IncludeComplexity:        true,
			IncludeDocumentation:     true,
			IncludeGenerics:          true,
			MaxFunctionLength:        30,
			MaxCyclomaticComplexity:  10,
			MaxStructFields:          20,
			MinDocumentationCoverage: 0.7,
			Duplication: DuplicationConfig{
				MinBlockLines:       6,
				SimilarityThreshold: 0.80,
				IgnoreTestFiles:     false,
			},
			Naming: NamingConfig{
				FlagGenericFilenames: true,
				FlagStuttering:       true,
				MinNameLength:        2,
			},
			Placement: PlacementConfig{
				AffinityMargin: 0.25,
				MinCohesion:    0.3,
			},
			Documentation: DocumentationConfig{
				RequireExportedDoc:  true,
				RequirePackageDoc:   true,
				StaleAnnotationDays: 180,
				MinCommentWords:     5,
			},
			Organization: OrganizationConfig{
				MaxFileLines:       500,
				MaxFileFunctions:   20,
				MaxFileTypes:       5,
				MaxPackageFiles:    20,
				MaxExportedSymbols: 50,
				MaxDirectoryDepth:  5,
				MaxFileImports:     15,
			},
			Burden: BurdenConfig{
				MaxParams:         5,
				MaxReturns:        3,
				MaxNesting:        4,
				FeatureEnvyRatio:  2.0,
				IgnoreBenignMagic: true,
			},
			Scoring: ScoringConfig{
				Weights: ScoringWeights{
					Duplication:   0.20,
					Naming:        0.10,
					Placement:     0.15,
					Documentation: 0.15,
					Organization:  0.15,
					Burden:        0.25,
				},
				MaxBurdenScore: 70.0,
			},
		},
		Output: OutputConfig{
			Format:          FormatConsole,
			Destination:     "stdout",
			UseColors:       true,
			ShowProgress:    true,
			Verbose:         false,
			IncludeOverview: true,
			IncludeDetails:  true,
			IncludeExamples: false,
			SortBy:          "complexity",
			Limit:           100,
		},
		Performance: PerformanceConfig{
			WorkerCount:     runtime.NumCPU(),
			MaxMemoryMB:     1024,
			Timeout:         time.Minute * 10,
			EnableProfiling: false,
			EnableCache:     true,
			CacheDirectory:  ".go-stats-generator-cache",
		},
		Filters: FilterConfig{
			IncludePatterns: []string{"**/*.go"},
			ExcludePatterns: []string{},
			IncludePackages: []string{},
			ExcludePackages: []string{},
			MaxFileSizeKB:   1024,
			SkipVendor:      true,
			SkipTestFiles:   false,
			SkipGenerated:   true,
		},
		Storage: StorageConfig{
			Type:         "sqlite",
			Path:         "metrics.db",
			Compression:  true,
			MaxSnapshots: 50,
			MaxAge:       30 * 24 * time.Hour, // 30 days
		},
	}
}
