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

// AnalysisConfig controls what gets analyzed
type AnalysisConfig struct {
	IncludeFunctions     bool `mapstructure:"include_functions" json:"include_functions"`
	IncludeStructs       bool `mapstructure:"include_structs" json:"include_structs"`
	IncludeInterfaces    bool `mapstructure:"include_interfaces" json:"include_interfaces"`
	IncludePatterns      bool `mapstructure:"include_patterns" json:"include_patterns"`
	IncludeComplexity    bool `mapstructure:"include_complexity" json:"include_complexity"`
	IncludeDocumentation bool `mapstructure:"include_documentation" json:"include_documentation"`
	IncludeGenerics      bool `mapstructure:"include_generics" json:"include_generics"`

	// Thresholds for warnings
	MaxFunctionLength        int     `mapstructure:"max_function_length" json:"max_function_length"`
	MaxCyclomaticComplexity  int     `mapstructure:"max_cyclomatic_complexity" json:"max_cyclomatic_complexity"`
	MaxStructFields          int     `mapstructure:"max_struct_fields" json:"max_struct_fields"`
	MinDocumentationCoverage float64 `mapstructure:"min_documentation_coverage" json:"min_documentation_coverage"`
}

// OutputConfig controls output formatting
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
	WorkerCount     int           `mapstructure:"worker_count" json:"worker_count"`
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
	Type        string `mapstructure:"type" json:"type"`               // "sqlite", "json", "memory"
	Path        string `mapstructure:"path" json:"path"`               // File path for sqlite/json
	Compression bool   `mapstructure:"compression" json:"compression"` // Enable compression for stored data

	// Retention policy
	MaxSnapshots int           `mapstructure:"max_snapshots" json:"max_snapshots"`
	MaxAge       time.Duration `mapstructure:"max_age" json:"max_age"`
}

// DefaultConfig returns the default configuration
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
			CacheDirectory:  ".gostats-cache",
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
