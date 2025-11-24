package reporter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/stretchr/testify/assert"
)

func TestConsoleReporter_ComplexitySorting_TieBreakByLength(t *testing.T) {
	// Create test data with functions that have the same complexity but different lengths
	report := &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:     "test-repo",
			GeneratedAt:    time.Now(),
			AnalysisTime:   time.Second,
			FilesProcessed: 5,
			ToolVersion:    "1.0.0",
		},
		Overview: metrics.OverviewMetrics{
			TotalLinesOfCode: 1000,
			TotalFunctions:   5,
		},
		Functions: []metrics.FunctionMetrics{
			{
				Name:    "ShortFunc",
				Package: "test",
				File:    "test.go",
				Lines: metrics.LineMetrics{
					Total: 10,
					Code:  8,
				},
				Complexity: metrics.ComplexityScore{
					Cyclomatic: 5,
					Overall:    5.0,
				},
			},
			{
				Name:    "MediumFunc",
				Package: "test",
				File:    "test.go",
				Lines: metrics.LineMetrics{
					Total: 50,
					Code:  45,
				},
				Complexity: metrics.ComplexityScore{
					Cyclomatic: 5,
					Overall:    5.0,
				},
			},
			{
				Name:    "LongFunc",
				Package: "test",
				File:    "test.go",
				Lines: metrics.LineMetrics{
					Total: 100,
					Code:  90,
				},
				Complexity: metrics.ComplexityScore{
					Cyclomatic: 5,
					Overall:    5.0,
				},
			},
			{
				Name:    "VeryShortFunc",
				Package: "test",
				File:    "test.go",
				Lines: metrics.LineMetrics{
					Total: 5,
					Code:  4,
				},
				Complexity: metrics.ComplexityScore{
					Cyclomatic: 5,
					Overall:    5.0,
				},
			},
			{
				Name:    "VeryLongFunc",
				Package: "test",
				File:    "test.go",
				Lines: metrics.LineMetrics{
					Total: 150,
					Code:  140,
				},
				Complexity: metrics.ComplexityScore{
					Cyclomatic: 5,
					Overall:    5.0,
				},
			},
		},
	}

	cfg := &config.OutputConfig{
		IncludeOverview: false,
		IncludeDetails:  true,
		Limit:           5,
	}

	reporter := NewConsoleReporter(cfg)
	var buf bytes.Buffer
	err := reporter.Generate(report, &buf)

	assert.NoError(t, err)
	output := buf.String()

	// Find the complexity analysis section
	lines := strings.Split(output, "\n")
	var complexitySection []string
	inComplexitySection := false

	for _, line := range lines {
		if strings.Contains(line, "=== COMPLEXITY ANALYSIS ===") {
			inComplexitySection = true
			continue
		}
		if inComplexitySection && strings.Contains(line, "===") {
			break
		}
		if inComplexitySection && strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "Top") && !strings.HasPrefix(line, "Function") && !strings.HasPrefix(line, "---") {
			complexitySection = append(complexitySection, line)
		}
	}

	// Verify that functions are sorted by length when complexity is tied
	// Expected order: VeryLongFunc (150), LongFunc (100), MediumFunc (50), ShortFunc (10), VeryShortFunc (5)
	assert.True(t, len(complexitySection) >= 5, "Should have at least 5 function entries")

	// Extract function names from the output
	var functionOrder []string
	for _, line := range complexitySection {
		fields := strings.Fields(line)
		if len(fields) > 0 {
			functionOrder = append(functionOrder, fields[0])
		}
	}

	assert.Equal(t, "VeryLongFunc", functionOrder[0], "VeryLongFunc (150 lines) should be first")
	assert.Equal(t, "LongFunc", functionOrder[1], "LongFunc (100 lines) should be second")
	assert.Equal(t, "MediumFunc", functionOrder[2], "MediumFunc (50 lines) should be third")
	assert.Equal(t, "ShortFunc", functionOrder[3], "ShortFunc (10 lines) should be fourth")
	assert.Equal(t, "VeryShortFunc", functionOrder[4], "VeryShortFunc (5 lines) should be fifth")
}

func TestConsoleReporter_ComplexitySorting_DifferentComplexities(t *testing.T) {
	// Create test data with functions that have different complexities
	report := &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:     "test-repo",
			GeneratedAt:    time.Now(),
			AnalysisTime:   time.Second,
			FilesProcessed: 3,
			ToolVersion:    "1.0.0",
		},
		Overview: metrics.OverviewMetrics{
			TotalLinesOfCode: 500,
			TotalFunctions:   3,
		},
		Functions: []metrics.FunctionMetrics{
			{
				Name:    "LowComplexityFunc",
				Package: "test",
				File:    "test.go",
				Lines: metrics.LineMetrics{
					Total: 100,
					Code:  90,
				},
				Complexity: metrics.ComplexityScore{
					Cyclomatic: 2,
					Overall:    2.0,
				},
			},
			{
				Name:    "HighComplexityFunc",
				Package: "test",
				File:    "test.go",
				Lines: metrics.LineMetrics{
					Total: 20,
					Code:  18,
				},
				Complexity: metrics.ComplexityScore{
					Cyclomatic: 15,
					Overall:    15.0,
				},
			},
			{
				Name:    "MediumComplexityFunc",
				Package: "test",
				File:    "test.go",
				Lines: metrics.LineMetrics{
					Total: 50,
					Code:  45,
				},
				Complexity: metrics.ComplexityScore{
					Cyclomatic: 8,
					Overall:    8.0,
				},
			},
		},
	}

	cfg := &config.OutputConfig{
		IncludeOverview: false,
		IncludeDetails:  true,
		Limit:           10,
	}

	reporter := NewConsoleReporter(cfg)
	var buf bytes.Buffer
	err := reporter.Generate(report, &buf)

	assert.NoError(t, err)
	output := buf.String()

	// Find the complexity analysis section
	lines := strings.Split(output, "\n")
	var functionOrder []string
	inComplexitySection := false

	for _, line := range lines {
		if strings.Contains(line, "=== COMPLEXITY ANALYSIS ===") {
			inComplexitySection = true
			continue
		}
		if inComplexitySection && strings.Contains(line, "===") {
			break
		}
		if inComplexitySection && strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "Top") && !strings.HasPrefix(line, "Function") && !strings.HasPrefix(line, "---") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				functionOrder = append(functionOrder, fields[0])
			}
		}
	}

	// When complexities are different, order should be by complexity descending
	assert.Equal(t, "HighComplexityFunc", functionOrder[0], "HighComplexityFunc (15) should be first")
	assert.Equal(t, "MediumComplexityFunc", functionOrder[1], "MediumComplexityFunc (8) should be second")
	assert.Equal(t, "LowComplexityFunc", functionOrder[2], "LowComplexityFunc (2) should be third")
}

func TestConsoleReporter_ComplexitySorting_MixedTiesAndDifferent(t *testing.T) {
	// Create test data with a mix of same and different complexities
	report := &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:     "test-repo",
			GeneratedAt:    time.Now(),
			AnalysisTime:   time.Second,
			FilesProcessed: 6,
			ToolVersion:    "1.0.0",
		},
		Overview: metrics.OverviewMetrics{
			TotalLinesOfCode: 600,
			TotalFunctions:   6,
		},
		Functions: []metrics.FunctionMetrics{
			{
				Name:    "Complexity10_Short",
				Package: "test",
				File:    "test.go",
				Lines:   metrics.LineMetrics{Total: 20},
				Complexity: metrics.ComplexityScore{
					Cyclomatic: 10,
					Overall:    10.0,
				},
			},
			{
				Name:    "Complexity10_Long",
				Package: "test",
				File:    "test.go",
				Lines:   metrics.LineMetrics{Total: 80},
				Complexity: metrics.ComplexityScore{
					Cyclomatic: 10,
					Overall:    10.0,
				},
			},
			{
				Name:    "Complexity15",
				Package: "test",
				File:    "test.go",
				Lines:   metrics.LineMetrics{Total: 30},
				Complexity: metrics.ComplexityScore{
					Cyclomatic: 15,
					Overall:    15.0,
				},
			},
			{
				Name:    "Complexity5_Medium",
				Package: "test",
				File:    "test.go",
				Lines:   metrics.LineMetrics{Total: 40},
				Complexity: metrics.ComplexityScore{
					Cyclomatic: 5,
					Overall:    5.0,
				},
			},
			{
				Name:    "Complexity5_Short",
				Package: "test",
				File:    "test.go",
				Lines:   metrics.LineMetrics{Total: 15},
				Complexity: metrics.ComplexityScore{
					Cyclomatic: 5,
					Overall:    5.0,
				},
			},
			{
				Name:    "Complexity10_Medium",
				Package: "test",
				File:    "test.go",
				Lines:   metrics.LineMetrics{Total: 50},
				Complexity: metrics.ComplexityScore{
					Cyclomatic: 10,
					Overall:    10.0,
				},
			},
		},
	}

	cfg := &config.OutputConfig{
		IncludeOverview: false,
		IncludeDetails:  true,
		Limit:           10,
	}

	reporter := NewConsoleReporter(cfg)
	var buf bytes.Buffer
	err := reporter.Generate(report, &buf)

	assert.NoError(t, err)
	output := buf.String()

	// Find the complexity analysis section
	lines := strings.Split(output, "\n")
	var functionOrder []string
	inComplexitySection := false

	for _, line := range lines {
		if strings.Contains(line, "=== COMPLEXITY ANALYSIS ===") {
			inComplexitySection = true
			continue
		}
		if inComplexitySection && strings.Contains(line, "===") {
			break
		}
		if inComplexitySection && strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "Top") && !strings.HasPrefix(line, "Function") && !strings.HasPrefix(line, "---") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				functionOrder = append(functionOrder, fields[0])
			}
		}
	}

	// Expected order:
	// 1. Complexity15 (complexity 15, 30 lines)
	// 2. Complexity10_Long (complexity 10, 80 lines) - longest of complexity 10
	// 3. Complexity10_Medium (complexity 10, 50 lines)
	// 4. Complexity10_Short (complexity 10, 20 lines) - shortest of complexity 10
	// 5. Complexity5_Medium (complexity 5, 40 lines) - longest of complexity 5
	// 6. Complexity5_Short (complexity 5, 15 lines) - shortest of complexity 5

	assert.Equal(t, 6, len(functionOrder), "Should have 6 functions")
	assert.Equal(t, "Complexity15", functionOrder[0], "Complexity15 should be first")
	assert.Equal(t, "Complexity10_Long", functionOrder[1], "Complexity10_Long (80 lines) should be second")
	assert.Equal(t, "Complexity10_Medium", functionOrder[2], "Complexity10_Medium (50 lines) should be third")
	assert.Equal(t, "Complexity10_Short", functionOrder[3], "Complexity10_Short (20 lines) should be fourth")
	assert.Equal(t, "Complexity5_Medium", functionOrder[4], "Complexity5_Medium (40 lines) should be fifth")
	assert.Equal(t, "Complexity5_Short", functionOrder[5], "Complexity5_Short (15 lines) should be sixth")
}
