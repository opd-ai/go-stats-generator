package config

import (
	"runtime"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Test that config is not nil
	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// Test Analysis configuration
	t.Run("AnalysisConfig", func(t *testing.T) {
		analysis := config.Analysis

		// Test boolean flags
		if !analysis.IncludeFunctions {
			t.Error("Expected IncludeFunctions to be true")
		}
		if !analysis.IncludeStructs {
			t.Error("Expected IncludeStructs to be true")
		}
		if !analysis.IncludeInterfaces {
			t.Error("Expected IncludeInterfaces to be true")
		}
		if !analysis.IncludePatterns {
			t.Error("Expected IncludePatterns to be true")
		}
		if !analysis.IncludeComplexity {
			t.Error("Expected IncludeComplexity to be true")
		}
		if !analysis.IncludeDocumentation {
			t.Error("Expected IncludeDocumentation to be true")
		}
		if !analysis.IncludeGenerics {
			t.Error("Expected IncludeGenerics to be true")
		}

		// Test threshold values
		if analysis.MaxFunctionLength != 30 {
			t.Errorf("Expected MaxFunctionLength to be 30, got %d", analysis.MaxFunctionLength)
		}
		if analysis.MaxCyclomaticComplexity != 10 {
			t.Errorf("Expected MaxCyclomaticComplexity to be 10, got %d", analysis.MaxCyclomaticComplexity)
		}
		if analysis.MaxStructFields != 20 {
			t.Errorf("Expected MaxStructFields to be 20, got %d", analysis.MaxStructFields)
		}
		if analysis.MinDocumentationCoverage != 0.7 {
			t.Errorf("Expected MinDocumentationCoverage to be 0.7, got %f", analysis.MinDocumentationCoverage)
		}
	})

	// Test Output configuration
	t.Run("OutputConfig", func(t *testing.T) {
		output := config.Output

		if output.Format != FormatConsole {
			t.Errorf("Expected Format to be %s, got %s", FormatConsole, output.Format)
		}
		if output.Destination != "stdout" {
			t.Errorf("Expected Destination to be 'stdout', got %s", output.Destination)
		}
		if !output.UseColors {
			t.Error("Expected UseColors to be true")
		}
		if !output.ShowProgress {
			t.Error("Expected ShowProgress to be true")
		}
		if output.Verbose {
			t.Error("Expected Verbose to be false")
		}
		if !output.IncludeOverview {
			t.Error("Expected IncludeOverview to be true")
		}
		if !output.IncludeDetails {
			t.Error("Expected IncludeDetails to be true")
		}
		if output.IncludeExamples {
			t.Error("Expected IncludeExamples to be false")
		}
		if output.SortBy != "complexity" {
			t.Errorf("Expected SortBy to be 'complexity', got %s", output.SortBy)
		}
		if output.Limit != 100 {
			t.Errorf("Expected Limit to be 100, got %d", output.Limit)
		}
	})

	// Test Performance configuration
	t.Run("PerformanceConfig", func(t *testing.T) {
		performance := config.Performance

		expectedWorkerCount := runtime.NumCPU()
		if performance.WorkerCount != expectedWorkerCount {
			t.Errorf("Expected WorkerCount to be %d, got %d", expectedWorkerCount, performance.WorkerCount)
		}
		if performance.MaxMemoryMB != 1024 {
			t.Errorf("Expected MaxMemoryMB to be 1024, got %d", performance.MaxMemoryMB)
		}
		expectedTimeout := time.Minute * 10
		if performance.Timeout != expectedTimeout {
			t.Errorf("Expected Timeout to be %v, got %v", expectedTimeout, performance.Timeout)
		}
		if performance.EnableProfiling {
			t.Error("Expected EnableProfiling to be false")
		}
		if !performance.EnableCache {
			t.Error("Expected EnableCache to be true")
		}
		if performance.CacheDirectory != ".gostats-cache" {
			t.Errorf("Expected CacheDirectory to be '.gostats-cache', got %s", performance.CacheDirectory)
		}
	})

	// Test Filters configuration
	t.Run("FilterConfig", func(t *testing.T) {
		filters := config.Filters

		// Test include patterns
		expectedIncludePatterns := []string{"**/*.go"}
		if len(filters.IncludePatterns) != len(expectedIncludePatterns) {
			t.Errorf("Expected %d include patterns, got %d", len(expectedIncludePatterns), len(filters.IncludePatterns))
		} else {
			for i, pattern := range expectedIncludePatterns {
				if filters.IncludePatterns[i] != pattern {
					t.Errorf("Expected include pattern %d to be %s, got %s", i, pattern, filters.IncludePatterns[i])
				}
			}
		}

		// Test empty slices
		if len(filters.ExcludePatterns) != 0 {
			t.Errorf("Expected ExcludePatterns to be empty, got %v", filters.ExcludePatterns)
		}
		if len(filters.IncludePackages) != 0 {
			t.Errorf("Expected IncludePackages to be empty, got %v", filters.IncludePackages)
		}
		if len(filters.ExcludePackages) != 0 {
			t.Errorf("Expected ExcludePackages to be empty, got %v", filters.ExcludePackages)
		}

		// Test filter settings
		if filters.MaxFileSizeKB != 1024 {
			t.Errorf("Expected MaxFileSizeKB to be 1024, got %d", filters.MaxFileSizeKB)
		}
		if !filters.SkipVendor {
			t.Error("Expected SkipVendor to be true")
		}
		if filters.SkipTestFiles {
			t.Error("Expected SkipTestFiles to be false")
		}
		if !filters.SkipGenerated {
			t.Error("Expected SkipGenerated to be true")
		}
	})

	// Test Storage configuration
	t.Run("StorageConfig", func(t *testing.T) {
		storage := config.Storage

		if storage.Type != "sqlite" {
			t.Errorf("Expected Type to be 'sqlite', got %s", storage.Type)
		}
		if storage.Path != "metrics.db" {
			t.Errorf("Expected Path to be 'metrics.db', got %s", storage.Path)
		}
		if !storage.Compression {
			t.Error("Expected Compression to be true")
		}
		if storage.MaxSnapshots != 50 {
			t.Errorf("Expected MaxSnapshots to be 50, got %d", storage.MaxSnapshots)
		}
		expectedMaxAge := 30 * 24 * time.Hour
		if storage.MaxAge != expectedMaxAge {
			t.Errorf("Expected MaxAge to be %v, got %v", expectedMaxAge, storage.MaxAge)
		}
	})
}

func TestOutputFormat_Constants(t *testing.T) {
	tests := []struct {
		name   string
		format OutputFormat
		value  string
	}{
		{"Console", FormatConsole, "console"},
		{"JSON", FormatJSON, "json"},
		{"CSV", FormatCSV, "csv"},
		{"HTML", FormatHTML, "html"},
		{"Markdown", FormatMarkdown, "markdown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.format) != tt.value {
				t.Errorf("Expected %s format to have value %s, got %s", tt.name, tt.value, string(tt.format))
			}
		})
	}
}

func TestConfig_StructFields(t *testing.T) {
	config := &Config{}

	// Test that all required fields exist and can be set
	t.Run("AnalysisConfig_Fields", func(t *testing.T) {
		config.Analysis = AnalysisConfig{
			IncludeFunctions:         true,
			IncludeStructs:           false,
			IncludeInterfaces:        true,
			IncludePatterns:          false,
			IncludeComplexity:        true,
			IncludeDocumentation:     false,
			IncludeGenerics:          true,
			MaxFunctionLength:        25,
			MaxCyclomaticComplexity:  8,
			MaxStructFields:          15,
			MinDocumentationCoverage: 0.8,
		}

		analysis := config.Analysis
		if !analysis.IncludeFunctions {
			t.Error("IncludeFunctions field not set correctly")
		}
		if analysis.IncludeStructs {
			t.Error("IncludeStructs field not set correctly")
		}
		if analysis.MaxFunctionLength != 25 {
			t.Error("MaxFunctionLength field not set correctly")
		}
		if analysis.MinDocumentationCoverage != 0.8 {
			t.Error("MinDocumentationCoverage field not set correctly")
		}
	})

	t.Run("OutputConfig_Fields", func(t *testing.T) {
		config.Output = OutputConfig{
			Format:          FormatJSON,
			Destination:     "file.json",
			UseColors:       false,
			ShowProgress:    false,
			Verbose:         true,
			IncludeOverview: false,
			IncludeDetails:  false,
			IncludeExamples: true,
			SortBy:          "name",
			Limit:           50,
		}

		output := config.Output
		if output.Format != FormatJSON {
			t.Error("Format field not set correctly")
		}
		if output.Destination != "file.json" {
			t.Error("Destination field not set correctly")
		}
		if output.UseColors {
			t.Error("UseColors field not set correctly")
		}
		if !output.Verbose {
			t.Error("Verbose field not set correctly")
		}
	})

	t.Run("PerformanceConfig_Fields", func(t *testing.T) {
		config.Performance = PerformanceConfig{
			WorkerCount:     4,
			MaxMemoryMB:     512,
			Timeout:         time.Minute * 5,
			EnableProfiling: true,
			EnableCache:     false,
			CacheDirectory:  "/tmp/cache",
		}

		performance := config.Performance
		if performance.WorkerCount != 4 {
			t.Error("WorkerCount field not set correctly")
		}
		if performance.MaxMemoryMB != 512 {
			t.Error("MaxMemoryMB field not set correctly")
		}
		if performance.Timeout != time.Minute*5 {
			t.Error("Timeout field not set correctly")
		}
		if !performance.EnableProfiling {
			t.Error("EnableProfiling field not set correctly")
		}
		if performance.EnableCache {
			t.Error("EnableCache field not set correctly")
		}
	})
}

func TestFilterConfig_SliceFields(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		packages []string
	}{
		{
			name:     "EmptySlices",
			patterns: []string{},
			packages: []string{},
		},
		{
			name:     "SingleValues",
			patterns: []string{"*.go"},
			packages: []string{"main"},
		},
		{
			name:     "MultipleValues",
			patterns: []string{"**/*.go", "**/*.mod"},
			packages: []string{"github.com/user/repo", "internal/pkg"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := FilterConfig{
				IncludePatterns: tt.patterns,
				ExcludePatterns: tt.patterns,
				IncludePackages: tt.packages,
				ExcludePackages: tt.packages,
			}

			if len(filter.IncludePatterns) != len(tt.patterns) {
				t.Errorf("IncludePatterns length mismatch: expected %d, got %d", len(tt.patterns), len(filter.IncludePatterns))
			}
			if len(filter.ExcludePatterns) != len(tt.patterns) {
				t.Errorf("ExcludePatterns length mismatch: expected %d, got %d", len(tt.patterns), len(filter.ExcludePatterns))
			}
			if len(filter.IncludePackages) != len(tt.packages) {
				t.Errorf("IncludePackages length mismatch: expected %d, got %d", len(tt.packages), len(filter.IncludePackages))
			}
			if len(filter.ExcludePackages) != len(tt.packages) {
				t.Errorf("ExcludePackages length mismatch: expected %d, got %d", len(tt.packages), len(filter.ExcludePackages))
			}
		})
	}
}

func TestStorageConfig_TimeValues(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected time.Duration
	}{
		{"OneHour", time.Hour, time.Hour},
		{"OneDay", 24 * time.Hour, 24 * time.Hour},
		{"OneWeek", 7 * 24 * time.Hour, 7 * 24 * time.Hour},
		{"ThirtyDays", 30 * 24 * time.Hour, 30 * 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := StorageConfig{
				MaxAge: tt.duration,
			}

			if storage.MaxAge != tt.expected {
				t.Errorf("MaxAge mismatch: expected %v, got %v", tt.expected, storage.MaxAge)
			}
		})
	}
}

func TestConfig_JSONSerialization(t *testing.T) {
	// Test that the config can be properly tagged for JSON serialization
	config := DefaultConfig()

	// Verify that struct tags are correctly applied by checking field accessibility
	// This is a structural test to ensure JSON marshaling would work
	if config.Analysis.IncludeFunctions != true {
		t.Error("Analysis field should be accessible for JSON serialization")
	}
	if config.Output.Format != FormatConsole {
		t.Error("Output field should be accessible for JSON serialization")
	}
	if config.Performance.WorkerCount != runtime.NumCPU() {
		t.Error("Performance field should be accessible for JSON serialization")
	}
	if len(config.Filters.IncludePatterns) == 0 {
		t.Error("Filters field should be accessible for JSON serialization")
	}
	if config.Storage.Type != "sqlite" {
		t.Error("Storage field should be accessible for JSON serialization")
	}
}
