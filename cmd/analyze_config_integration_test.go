package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigFileIntegration verifies that config file values are properly loaded and used
func TestConfigFileIntegration(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".go-stats-generator.yaml")
	
	configContent := `analysis:
  include_functions: false
  include_structs: false
  include_interfaces: false
  duplication:
    min_block_lines: 12
    similarity_threshold: 0.75
    ignore_test_files: true

output:
  use_colors: false
  show_progress: false
  include_examples: true

performance:
  enable_cache: false
  max_memory_mb: 768
  enable_profiling: false
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)
	
	// Reset viper and load the config file
	viper.Reset()
	viper.SetConfigFile(configFile)
	err = viper.ReadInConfig()
	require.NoError(t, err)
	
	// Create config and load from viper
	cfg := config.DefaultConfig()
	loadAnalysisConfiguration(cfg)
	loadOutputConfiguration(cfg)
	loadPerformanceConfiguration(cfg)
	
	// Verify all config values were loaded correctly
	assert.False(t, cfg.Analysis.IncludeFunctions, "include_functions should be false")
	assert.False(t, cfg.Analysis.IncludeStructs, "include_structs should be false")
	assert.False(t, cfg.Analysis.IncludeInterfaces, "include_interfaces should be false")
	
	assert.Equal(t, 12, cfg.Analysis.Duplication.MinBlockLines, "duplication.min_block_lines should be 12")
	assert.Equal(t, 0.75, cfg.Analysis.Duplication.SimilarityThreshold, "duplication.similarity_threshold should be 0.75")
	assert.True(t, cfg.Analysis.Duplication.IgnoreTestFiles, "duplication.ignore_test_files should be true")
	
	assert.False(t, cfg.Output.UseColors, "output.use_colors should be false")
	assert.False(t, cfg.Output.ShowProgress, "output.show_progress should be false")
	assert.True(t, cfg.Output.IncludeExamples, "output.include_examples should be true")
	
	assert.False(t, cfg.Performance.EnableCache, "performance.enable_cache should be false")
	assert.Equal(t, 768, cfg.Performance.MaxMemoryMB, "performance.max_memory_mb should be 768")
	assert.False(t, cfg.Performance.EnableProfiling, "performance.enable_profiling should be false")
}

// TestConfigDefaultValuesIntact verifies that config loaders don't break default values
func TestConfigDefaultValuesIntact(t *testing.T) {
	viper.Reset()
	
	// Load config without setting any values
	cfg := config.DefaultConfig()
	loadAnalysisConfiguration(cfg)
	loadOutputConfiguration(cfg)
	loadPerformanceConfiguration(cfg)
	
	// Verify defaults remain intact
	assert.True(t, cfg.Analysis.IncludeFunctions, "default include_functions should be true")
	assert.True(t, cfg.Analysis.IncludeStructs, "default include_structs should be true")
	assert.True(t, cfg.Analysis.IncludeInterfaces, "default include_interfaces should be true")
	
	assert.Equal(t, 6, cfg.Analysis.Duplication.MinBlockLines, "default duplication.min_block_lines should be 6")
	assert.Equal(t, 0.80, cfg.Analysis.Duplication.SimilarityThreshold, "default duplication.similarity_threshold should be 0.80")
	assert.False(t, cfg.Analysis.Duplication.IgnoreTestFiles, "default duplication.ignore_test_files should be false")
	
	assert.True(t, cfg.Output.UseColors, "default output.use_colors should be true")
	assert.True(t, cfg.Output.ShowProgress, "default output.show_progress should be true")
	assert.False(t, cfg.Output.IncludeExamples, "default output.include_examples should be false")
	
	assert.True(t, cfg.Performance.EnableCache, "default performance.enable_cache should be true")
	assert.Equal(t, 1024, cfg.Performance.MaxMemoryMB, "default performance.max_memory_mb should be 1024")
	assert.False(t, cfg.Performance.EnableProfiling, "default performance.enable_profiling should be false")
}

// TestPartialConfigOverride verifies that partial config files work correctly
func TestPartialConfigOverride(t *testing.T) {
	viper.Reset()
	
	// Only set some values
	viper.Set("analysis.duplication.min_block_lines", 15)
	viper.Set("output.use_colors", false)
	viper.Set("performance.max_memory_mb", 2048)
	
	cfg := config.DefaultConfig()
	loadAnalysisConfiguration(cfg)
	loadOutputConfiguration(cfg)
	loadPerformanceConfiguration(cfg)
	
	// Verify overridden values
	assert.Equal(t, 15, cfg.Analysis.Duplication.MinBlockLines)
	assert.False(t, cfg.Output.UseColors)
	assert.Equal(t, 2048, cfg.Performance.MaxMemoryMB)
	
	// Verify defaults for non-overridden values
	assert.True(t, cfg.Analysis.IncludeFunctions)
	assert.Equal(t, 0.80, cfg.Analysis.Duplication.SimilarityThreshold)
	assert.True(t, cfg.Output.ShowProgress)
	assert.True(t, cfg.Performance.EnableCache)
}

// TestConfigurationPrecedence verifies that CLI flags override config file values
func TestConfigurationPrecedence(t *testing.T) {
	// This test documents the intended precedence: CLI flags > config file > defaults
	// The actual flag binding happens in the analyze command, so we just test
	// that the config loaders respect viper values
	
	viper.Reset()
	
	// Simulate config file setting
	viper.Set("analysis.duplication.min_block_lines", 10)
	
	cfg := config.DefaultConfig()
	loadAnalysisConfiguration(cfg)
	
	// Config loader should pick up viper value
	assert.Equal(t, 10, cfg.Analysis.Duplication.MinBlockLines)
	
	// Now simulate CLI flag override
	viper.Set("analysis.duplication.min_block_lines", 20)
	
	cfg2 := config.DefaultConfig()
	loadAnalysisConfiguration(cfg2)
	
	// Config loader should pick up new viper value (CLI flag wins)
	assert.Equal(t, 20, cfg2.Analysis.Duplication.MinBlockLines)
}

// TestLoadConfigurationWithNonExistentFile verifies graceful handling of missing config
func TestLoadConfigurationWithNonExistentFile(t *testing.T) {
	viper.Reset()
	viper.SetConfigFile("/nonexistent/config.yaml")
	
	// Viper will error on ReadInConfig, but our loaders should work with empty viper
	cfg := config.DefaultConfig()
	loadAnalysisConfiguration(cfg)
	loadOutputConfiguration(cfg)
	loadPerformanceConfiguration(cfg)
	
	// Should still have defaults
	assert.True(t, cfg.Analysis.IncludeFunctions)
	assert.True(t, cfg.Output.UseColors)
	assert.True(t, cfg.Performance.EnableCache)
}

// BenchmarkConfigurationLoading benchmarks config loading performance
func BenchmarkConfigurationLoading(b *testing.B) {
	viper.Reset()
	viper.Set("analysis.include_functions", false)
	viper.Set("analysis.duplication.min_block_lines", 12)
	viper.Set("output.use_colors", false)
	viper.Set("performance.enable_cache", false)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg := config.DefaultConfig()
		loadAnalysisConfiguration(cfg)
		loadOutputConfiguration(cfg)
		loadPerformanceConfiguration(cfg)
	}
}

// TestConfigValidationAfterLoad verifies config values are sensible after loading
func TestConfigValidationAfterLoad(t *testing.T) {
	tests := []struct {
		name      string
		setup     func()
		validator func(*testing.T, *config.Config)
	}{
		{
			name: "similarity threshold in valid range",
			setup: func() {
				viper.Reset()
				viper.Set("analysis.duplication.similarity_threshold", 0.5)
			},
			validator: func(t *testing.T, cfg *config.Config) {
				threshold := cfg.Analysis.Duplication.SimilarityThreshold
				assert.GreaterOrEqual(t, threshold, 0.0, "threshold should be >= 0.0")
				assert.LessOrEqual(t, threshold, 1.0, "threshold should be <= 1.0")
			},
		},
		{
			name: "min_block_lines is positive",
			setup: func() {
				viper.Reset()
				viper.Set("analysis.duplication.min_block_lines", 5)
			},
			validator: func(t *testing.T, cfg *config.Config) {
				assert.Greater(t, cfg.Analysis.Duplication.MinBlockLines, 0, "min_block_lines should be positive")
			},
		},
		{
			name: "max_memory_mb is reasonable",
			setup: func() {
				viper.Reset()
				viper.Set("performance.max_memory_mb", 512)
			},
			validator: func(t *testing.T, cfg *config.Config) {
				assert.Greater(t, cfg.Performance.MaxMemoryMB, 0, "max_memory_mb should be positive")
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			cfg := config.DefaultConfig()
			loadAnalysisConfiguration(cfg)
			loadOutputConfiguration(cfg)
			loadPerformanceConfiguration(cfg)
			tt.validator(t, cfg)
		})
	}
}
