package reporter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestReportWithDuplication creates a test report with duplication metrics
func createTestReportWithDuplication() *metrics.Report {
	return &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:     "test-repo",
			GeneratedAt:    time.Date(2026, 3, 2, 15, 18, 18, 0, time.UTC),
			AnalysisTime:   1500000,
			FilesProcessed: 3,
			ToolVersion:    "1.0.0",
			GoVersion:      "go1.23.2",
		},
		Overview: metrics.OverviewMetrics{
			TotalLinesOfCode: 500,
			TotalFunctions:   10,
			TotalMethods:     5,
			TotalStructs:     3,
			TotalInterfaces:  2,
			TotalPackages:    2,
			TotalFiles:       3,
		},
		Functions: []metrics.FunctionMetrics{
			{
				Name:    "ProcessUserData",
				Package: "testpkg",
				File:    "testdata/user.go",
				Line:    10,
				Lines: metrics.LineMetrics{
					Total:    20,
					Code:     15,
					Comments: 3,
					Blank:    2,
				},
				Complexity: metrics.ComplexityScore{
					Cyclomatic:   5,
					Cognitive:    4,
					NestingDepth: 2,
					Overall:      6.5,
				},
			},
		},
		Duplication: metrics.DuplicationMetrics{
			ClonePairs:       5,
			DuplicatedLines:  120,
			DuplicationRatio: 0.24,
			LargestCloneSize: 35,
			Clones: []metrics.ClonePair{
				{
					Hash:      "abc123def456",
					Type:      metrics.CloneTypeExact,
					LineCount: 35,
					Instances: []metrics.CloneInstance{
						{
							File:      "testdata/user.go",
							StartLine: 10,
							EndLine:   45,
							NodeCount: 120,
						},
						{
							File:      "testdata/admin.go",
							StartLine: 20,
							EndLine:   55,
							NodeCount: 120,
						},
					},
				},
				{
					Hash:      "xyz789ghi012",
					Type:      metrics.CloneTypeRenamed,
					LineCount: 20,
					Instances: []metrics.CloneInstance{
						{
							File:      "testdata/validator.go",
							StartLine: 30,
							EndLine:   50,
							NodeCount: 80,
						},
						{
							File:      "testdata/checker.go",
							StartLine: 15,
							EndLine:   35,
							NodeCount: 80,
						},
					},
				},
				{
					Hash:      "near123clone",
					Type:      metrics.CloneTypeNear,
					LineCount: 15,
					Instances: []metrics.CloneInstance{
						{
							File:      "testdata/handler.go",
							StartLine: 40,
							EndLine:   55,
							NodeCount: 60,
						},
						{
							File:      "testdata/processor.go",
							StartLine: 50,
							EndLine:   65,
							NodeCount: 62,
						},
					},
				},
			},
		},
	}
}

// TestConsoleReporter_WithDuplication tests console reporter with duplication metrics
func TestConsoleReporter_WithDuplication(t *testing.T) {
	report := createTestReportWithDuplication()
	cfg := &config.OutputConfig{
		UseColors:       false,
		IncludeOverview: true,
		IncludeDetails:  true,
		Limit:           10,
	}

	reporter := NewConsoleReporter(cfg)
	var buf bytes.Buffer

	err := reporter.Generate(report, &buf)
	require.NoError(t, err)

	output := buf.String()

	// Verify duplication section exists
	assert.Contains(t, output, "=== DUPLICATION ANALYSIS ===")
	assert.Contains(t, output, "Clone Pairs Detected: 5")
	assert.Contains(t, output, "Duplicated Lines: 120")
	assert.Contains(t, output, "Duplication Ratio: 24.00%")
	assert.Contains(t, output, "Largest Clone Size: 35 lines")

	// Verify clone pairs table
	assert.Contains(t, output, "Top")
	assert.Contains(t, output, "Clone Pairs")
	assert.Contains(t, output, "exact")
	assert.Contains(t, output, "renamed")
	assert.Contains(t, output, "near")

	// Verify specific clone details
	assert.Contains(t, output, "testdata/user.go:10-45")
	assert.Contains(t, output, "testdata/validator.go:30-50")
}

// TestJSONReporter_WithDuplication tests JSON reporter with duplication metrics
func TestJSONReporter_WithDuplication(t *testing.T) {
	report := createTestReportWithDuplication()
	reporter := NewJSONReporter()
	var buf bytes.Buffer

	err := reporter.Generate(report, &buf)
	require.NoError(t, err)

	output := buf.String()

	// Verify JSON contains duplication section
	assert.Contains(t, output, `"duplication"`)
	assert.Contains(t, output, `"clone_pairs": 5`)
	assert.Contains(t, output, `"duplicated_lines": 120`)
	assert.Contains(t, output, `"duplication_ratio": 0.24`)
	assert.Contains(t, output, `"largest_clone_size": 35`)
	assert.Contains(t, output, `"clones"`)
	assert.Contains(t, output, `"hash"`)
	assert.Contains(t, output, `"type"`)
	assert.Contains(t, output, `"instances"`)
	assert.Contains(t, output, `"exact"`)
	assert.Contains(t, output, `"renamed"`)
	assert.Contains(t, output, `"near"`)
}

// TestHTMLReporter_WithDuplication tests HTML reporter with duplication metrics
func TestHTMLReporter_WithDuplication(t *testing.T) {
	report := createTestReportWithDuplication()
	reporter := NewHTMLReporter()
	var buf bytes.Buffer

	err := reporter.Generate(report, &buf)
	require.NoError(t, err)

	output := buf.String()

	// Verify HTML contains duplication tab
	assert.Contains(t, output, `data-tab="duplication"`)
	assert.Contains(t, output, `<button class="nav-tab" data-tab="duplication">Duplication</button>`)
	assert.Contains(t, output, `<section id="duplication" class="tab-content">`)
	assert.Contains(t, output, "Code Duplication Analysis")

	// Verify duplication metrics in HTML
	assert.Contains(t, output, "Clone Pairs")
	assert.Contains(t, output, "Duplicated Lines")
	assert.Contains(t, output, "Duplication Ratio")
	assert.Contains(t, output, "Largest Clone")

	// Verify clone pairs table
	assert.Contains(t, output, "Detected Clone Pairs")
	assert.Contains(t, output, "<th role=\"columnheader\">Type</th>")
	assert.Contains(t, output, "<th role=\"columnheader\">Lines</th>")
	assert.Contains(t, output, "<th role=\"columnheader\">Instances</th>")
}

// TestMarkdownReporter_WithDuplication tests Markdown reporter with duplication metrics
func TestMarkdownReporter_WithDuplication(t *testing.T) {
	report := createTestReportWithDuplication()
	reporter := NewMarkdownReporter()
	var buf bytes.Buffer

	err := reporter.Generate(report, &buf)
	require.NoError(t, err)

	output := buf.String()

	// Verify markdown contains duplication section
	assert.Contains(t, output, "## 🔄 Code Duplication")
	assert.Contains(t, output, "| **Clone Pairs** | 5 |")
	assert.Contains(t, output, "| **Duplicated Lines** | 120 |")
	assert.Contains(t, output, "| **Duplication Ratio** |")
	assert.Contains(t, output, "| **Largest Clone** | 35 lines |")

	// Verify top clone pairs table
	assert.Contains(t, output, "### Top Clone Pairs")
	assert.Contains(t, output, "| Type | Lines | Instances | First Location |")
	assert.Contains(t, output, "| exact |")
	assert.Contains(t, output, "| renamed |")
}

// TestReporters_WithNoDuplication tests that reporters handle reports with no duplication
func TestReporters_WithNoDuplication(t *testing.T) {
	report := createTestReportWithDuplication()
	report.Duplication = metrics.DuplicationMetrics{} // Empty duplication metrics

	t.Run("Console", func(t *testing.T) {
		cfg := &config.OutputConfig{
			IncludeDetails: true,
			Limit:          10,
		}
		reporter := NewConsoleReporter(cfg)
		var buf bytes.Buffer

		err := reporter.Generate(report, &buf)
		require.NoError(t, err)

		output := buf.String()
		// Should not have duplication section when ClonePairs is 0
		assert.NotContains(t, output, "=== DUPLICATION ANALYSIS ===")
	})

	t.Run("HTML", func(t *testing.T) {
		reporter := NewHTMLReporter()
		var buf bytes.Buffer

		err := reporter.Generate(report, &buf)
		require.NoError(t, err)

		output := buf.String()
		// Should not have duplication tab when ClonePairs is 0
		assert.NotContains(t, output, `data-tab="duplication"`)
	})

	t.Run("Markdown", func(t *testing.T) {
		reporter := NewMarkdownReporter()
		var buf bytes.Buffer

		err := reporter.Generate(report, &buf)
		require.NoError(t, err)

		output := buf.String()
		// Should not have duplication section when ClonePairs is 0
		assert.NotContains(t, output, "## 🔄 Code Duplication")
	})
}

// TestConsoleReporter_DuplicationSorting tests that clone pairs are sorted correctly
func TestConsoleReporter_DuplicationSorting(t *testing.T) {
	report := createTestReportWithDuplication()

	// Add more clones with different sizes to test sorting
	report.Duplication.Clones = append(report.Duplication.Clones, metrics.ClonePair{
		Hash:      "small123",
		Type:      metrics.CloneTypeExact,
		LineCount: 8,
		Instances: []metrics.CloneInstance{
			{File: "small1.go", StartLine: 1, EndLine: 9, NodeCount: 30},
			{File: "small2.go", StartLine: 1, EndLine: 9, NodeCount: 30},
		},
	})

	cfg := &config.OutputConfig{
		IncludeDetails: true,
		Limit:          10,
	}
	reporter := NewConsoleReporter(cfg)
	var buf bytes.Buffer

	err := reporter.Generate(report, &buf)
	require.NoError(t, err)

	output := buf.String()

	// Verify largest clone appears first in output
	lines := strings.Split(output, "\n")
	var cloneTableStarted bool
	var firstCloneLineCount int

	for _, line := range lines {
		if strings.Contains(line, "Top") && strings.Contains(line, "Clone Pairs") {
			cloneTableStarted = true
			continue
		}
		if cloneTableStarted && strings.Contains(line, "exact") {
			// Extract line count from the table row
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				// The line count should be the second field
				assert.Contains(t, fields[1], "35", "Largest clone should appear first")
				firstCloneLineCount = 35
				break
			}
		}
	}

	assert.Equal(t, 35, firstCloneLineCount, "First clone in table should be the largest")
}
