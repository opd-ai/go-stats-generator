package cmd

import (
	"testing"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadAnalysisConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		validate func(*testing.T, *config.Config)
	}{
		{
			name: "loads include_functions from config",
			setup: func() {
				viper.Reset()
				viper.Set("analysis.include_functions", false)
			},
			validate: func(t *testing.T, cfg *config.Config) {
				assert.False(t, cfg.Analysis.IncludeFunctions)
			},
		},
		{
			name: "loads include_structs from config",
			setup: func() {
				viper.Reset()
				viper.Set("analysis.include_structs", false)
			},
			validate: func(t *testing.T, cfg *config.Config) {
				assert.False(t, cfg.Analysis.IncludeStructs)
			},
		},
		{
			name: "loads include_interfaces from config",
			setup: func() {
				viper.Reset()
				viper.Set("analysis.include_interfaces", false)
			},
			validate: func(t *testing.T, cfg *config.Config) {
				assert.False(t, cfg.Analysis.IncludeInterfaces)
			},
		},
		{
			name: "loads duplication min_block_lines from config",
			setup: func() {
				viper.Reset()
				viper.Set("analysis.duplication.min_block_lines", 10)
			},
			validate: func(t *testing.T, cfg *config.Config) {
				assert.Equal(t, 10, cfg.Analysis.Duplication.MinBlockLines)
			},
		},
		{
			name: "loads duplication similarity_threshold from config",
			setup: func() {
				viper.Reset()
				viper.Set("analysis.duplication.similarity_threshold", 0.95)
			},
			validate: func(t *testing.T, cfg *config.Config) {
				assert.Equal(t, 0.95, cfg.Analysis.Duplication.SimilarityThreshold)
			},
		},
		{
			name: "loads duplication ignore_test_files from config",
			setup: func() {
				viper.Reset()
				viper.Set("analysis.duplication.ignore_test_files", true)
			},
			validate: func(t *testing.T, cfg *config.Config) {
				assert.True(t, cfg.Analysis.Duplication.IgnoreTestFiles)
			},
		},
		{
			name: "loads all analysis config values",
			setup: func() {
				viper.Reset()
				viper.Set("analysis.include_functions", false)
				viper.Set("analysis.include_structs", false)
				viper.Set("analysis.include_interfaces", false)
				viper.Set("analysis.include_patterns", false)
				viper.Set("analysis.include_complexity", false)
				viper.Set("analysis.include_documentation", false)
				viper.Set("analysis.include_generics", false)
				viper.Set("analysis.max_function_length", 50)
				viper.Set("analysis.max_cyclomatic_complexity", 20)
				viper.Set("analysis.min_documentation_coverage", 0.9)
				viper.Set("analysis.duplication.min_block_lines", 8)
				viper.Set("analysis.duplication.similarity_threshold", 0.85)
				viper.Set("analysis.duplication.ignore_test_files", true)
			},
			validate: func(t *testing.T, cfg *config.Config) {
				assert.False(t, cfg.Analysis.IncludeFunctions)
				assert.False(t, cfg.Analysis.IncludeStructs)
				assert.False(t, cfg.Analysis.IncludeInterfaces)
				assert.False(t, cfg.Analysis.IncludePatterns)
				assert.False(t, cfg.Analysis.IncludeComplexity)
				assert.False(t, cfg.Analysis.IncludeDocumentation)
				assert.False(t, cfg.Analysis.IncludeGenerics)
				assert.Equal(t, 50, cfg.Analysis.MaxFunctionLength)
				assert.Equal(t, 20, cfg.Analysis.MaxCyclomaticComplexity)
				assert.Equal(t, 0.9, cfg.Analysis.MinDocumentationCoverage)
				assert.Equal(t, 8, cfg.Analysis.Duplication.MinBlockLines)
				assert.Equal(t, 0.85, cfg.Analysis.Duplication.SimilarityThreshold)
				assert.True(t, cfg.Analysis.Duplication.IgnoreTestFiles)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			cfg := config.DefaultConfig()
			loadAnalysisConfiguration(cfg)
			tt.validate(t, cfg)
		})
	}
}

func TestLoadOutputConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		validate func(*testing.T, *config.Config)
	}{
		{
			name: "loads use_colors from config",
			setup: func() {
				viper.Reset()
				viper.Set("output.use_colors", false)
			},
			validate: func(t *testing.T, cfg *config.Config) {
				assert.False(t, cfg.Output.UseColors)
			},
		},
		{
			name: "loads show_progress from config",
			setup: func() {
				viper.Reset()
				viper.Set("output.show_progress", false)
			},
			validate: func(t *testing.T, cfg *config.Config) {
				assert.False(t, cfg.Output.ShowProgress)
			},
		},
		{
			name: "loads include_examples from config",
			setup: func() {
				viper.Reset()
				viper.Set("output.include_examples", true)
			},
			validate: func(t *testing.T, cfg *config.Config) {
				assert.True(t, cfg.Output.IncludeExamples)
			},
		},
		{
			name: "loads all output config values",
			setup: func() {
				viper.Reset()
				viper.Set("output.format", "json")
				viper.Set("output.destination", "/tmp/report.json")
				viper.Set("output.verbose", true)
				viper.Set("output.use_colors", false)
				viper.Set("output.show_progress", false)
				viper.Set("output.include_examples", true)
			},
			validate: func(t *testing.T, cfg *config.Config) {
				assert.Equal(t, config.FormatJSON, cfg.Output.Format)
				assert.Equal(t, "/tmp/report.json", cfg.Output.Destination)
				assert.True(t, cfg.Output.Verbose)
				assert.False(t, cfg.Output.UseColors)
				assert.False(t, cfg.Output.ShowProgress)
				assert.True(t, cfg.Output.IncludeExamples)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			cfg := config.DefaultConfig()
			loadOutputConfiguration(cfg)
			tt.validate(t, cfg)
		})
	}
}

func TestLoadPerformanceConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		validate func(*testing.T, *config.Config)
	}{
		{
			name: "loads enable_cache from config",
			setup: func() {
				viper.Reset()
				viper.Set("performance.enable_cache", false)
			},
			validate: func(t *testing.T, cfg *config.Config) {
				assert.False(t, cfg.Performance.EnableCache)
			},
		},
		{
			name: "loads max_memory_mb from config",
			setup: func() {
				viper.Reset()
				viper.Set("performance.max_memory_mb", 2048)
			},
			validate: func(t *testing.T, cfg *config.Config) {
				assert.Equal(t, 2048, cfg.Performance.MaxMemoryMB)
			},
		},
		{
			name: "loads enable_profiling from config",
			setup: func() {
				viper.Reset()
				viper.Set("performance.enable_profiling", true)
			},
			validate: func(t *testing.T, cfg *config.Config) {
				assert.True(t, cfg.Performance.EnableProfiling)
			},
		},
		{
			name: "loads all performance config values",
			setup: func() {
				viper.Reset()
				viper.Set("performance.worker_count", 16)
				viper.Set("performance.timeout", "30m")
				viper.Set("performance.enable_cache", false)
				viper.Set("performance.max_memory_mb", 512)
				viper.Set("performance.enable_profiling", true)
			},
			validate: func(t *testing.T, cfg *config.Config) {
				assert.Equal(t, 16, cfg.Performance.WorkerCount)
				assert.Equal(t, 30*time.Minute, cfg.Performance.Timeout)
				assert.False(t, cfg.Performance.EnableCache)
				assert.Equal(t, 512, cfg.Performance.MaxMemoryMB)
				assert.True(t, cfg.Performance.EnableProfiling)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			cfg := config.DefaultConfig()
			loadPerformanceConfiguration(cfg)
			tt.validate(t, cfg)
		})
	}
}

func TestConfigurationLoadingIntegration(t *testing.T) {
	// This test verifies that all configuration loaders work together
	viper.Reset()

	// Set all previously missing config values
	viper.Set("analysis.include_functions", false)
	viper.Set("analysis.include_structs", false)
	viper.Set("analysis.include_interfaces", false)
	viper.Set("analysis.duplication.min_block_lines", 12)
	viper.Set("analysis.duplication.similarity_threshold", 0.75)
	viper.Set("analysis.duplication.ignore_test_files", true)

	viper.Set("output.use_colors", false)
	viper.Set("output.show_progress", false)
	viper.Set("output.include_examples", true)

	viper.Set("performance.enable_cache", false)
	viper.Set("performance.max_memory_mb", 768)
	viper.Set("performance.enable_profiling", true)

	cfg := config.DefaultConfig()

	loadAnalysisConfiguration(cfg)
	loadOutputConfiguration(cfg)
	loadPerformanceConfiguration(cfg)

	// Verify all values were loaded correctly
	require.False(t, cfg.Analysis.IncludeFunctions)
	require.False(t, cfg.Analysis.IncludeStructs)
	require.False(t, cfg.Analysis.IncludeInterfaces)
	require.Equal(t, 12, cfg.Analysis.Duplication.MinBlockLines)
	require.Equal(t, 0.75, cfg.Analysis.Duplication.SimilarityThreshold)
	require.True(t, cfg.Analysis.Duplication.IgnoreTestFiles)

	require.False(t, cfg.Output.UseColors)
	require.False(t, cfg.Output.ShowProgress)
	require.True(t, cfg.Output.IncludeExamples)

	require.False(t, cfg.Performance.EnableCache)
	require.Equal(t, 768, cfg.Performance.MaxMemoryMB)
	require.True(t, cfg.Performance.EnableProfiling)
}
