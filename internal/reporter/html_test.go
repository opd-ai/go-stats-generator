package reporter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

func TestHTMLReporter_TemplateFieldsCorrect(t *testing.T) {
	// Create test data that exercises the corrected template fields
	report := &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:     "test-repo",
			GeneratedAt:    time.Now(),
			AnalysisTime:   time.Second,
			FilesProcessed: 5,
		},
		Overview: metrics.OverviewMetrics{
			TotalLinesOfCode: 1000, // Fix #1: Correct field name
			TotalFunctions:   25,
			TotalFiles:       5,
		},
		Functions: []metrics.FunctionMetrics{
			{
				Name:    "TestFunction",
				Package: "test",
				Lines: metrics.LineMetrics{ // Fix #3: Nested structure
					Code: 20,
				},
				Signature: metrics.FunctionSignature{ // Fix #5: Nested structure
					ParameterCount: 2,
					ReturnCount:    1,
				},
				Complexity: metrics.ComplexityScore{ // Fix #4: Nested structure
					Cyclomatic: 5,
				},
			},
		},
	}

	reporter := NewHTMLReporterWithConfig(&config.OutputConfig{
		IncludeOverview: true,
		IncludeDetails:  true,
	})

	var buf bytes.Buffer
	err := reporter.Generate(report, &buf)
	if err != nil {
		t.Fatalf("HTML generation failed: %v", err)
	}

	html := buf.String()

	// Verify template fixes work (data should appear in output)
	if !strings.Contains(html, "1000") { // TotalLinesOfCode
		t.Error("Fix #1 failed: TotalLinesOfCode not rendered")
	}
	if !strings.Contains(html, "20") { // Lines.Code
		t.Error("Fix #3 failed: Lines.Code not rendered")
	}
	if !strings.Contains(html, "5") { // Complexity.Cyclomatic
		t.Error("Fix #4 failed: Complexity.Cyclomatic not rendered")
	}
	if !strings.Contains(html, "2") { // Signature.ParameterCount
		t.Error("Fix #5 failed: Signature.ParameterCount not rendered")
	}
	if !strings.Contains(html, "1") { // Signature.ReturnCount
		t.Error("Fix #5 failed: Signature.ReturnCount not rendered")
	}

	// Verify Fix #2: AverageComplexity should not appear
	if strings.Contains(html, "Average Complexity") {
		t.Error("Fix #2 failed: AverageComplexity section should be removed")
	}
}

func TestHTMLReporter_DiffTemplateFieldsCorrect(t *testing.T) {
	// Create test diff data that exercises the corrected template fields
	diff := &metrics.ComplexityDiff{
		Baseline: metrics.MetricsSnapshot{
			ID: "baseline",
			Metadata: metrics.SnapshotMetadata{
				Timestamp: time.Now().Add(-time.Hour),
			},
		},
		Current: metrics.MetricsSnapshot{
			ID: "current",
			Metadata: metrics.SnapshotMetadata{
				Timestamp: time.Now(),
			},
		},
		Summary: metrics.DiffSummary{
			TotalChanges:     5,
			RegressionCount:  2,
			ImprovementCount: 1,
			OverallTrend:     metrics.TrendImproving, // Fix #6: Correct field
		},
		Regressions: []metrics.Regression{
			{
				Type:        metrics.ComplexityRegression, // Fix #7: Correct field
				Location:    "test.TestFunc",              // Fix #7: Correct field
				Description: "Complexity increased",       // Fix #7: Correct field
				OldValue:    5.0,
				NewValue:    8.0,
				Delta: metrics.Delta{ // Fix #7: Nested structure
					Percentage: 60.0,
				},
				Severity: metrics.SeverityLevelWarning,
			},
		},
		Improvements: []metrics.Improvement{
			{
				Type:        metrics.ComplexityImprovement, // Fix #7: Correct field
				Location:    "test.AnotherFunc",            // Fix #7: Correct field
				Description: "Complexity decreased",        // Fix #7: Correct field
				OldValue:    10.0,
				NewValue:    7.0,
				Delta: metrics.Delta{ // Fix #7: Nested structure
					Percentage: -30.0,
				},
				Impact: metrics.ImpactLevelMedium,
			},
		},
		Changes: []metrics.MetricChange{
			{
				Category:    "function",       // Fix #8: Correct field
				Name:        "TestFunc",       // Fix #8: Correct field
				Description: "Lines changed",  // Fix #8: Correct field
				OldValue:    20.0,             // Use float64
				NewValue:    25.0,             // Use float64
				Delta: metrics.Delta{ // Fix #8: Nested structure
					Percentage: 25.0,
				},
			},
		},
	}

	reporter := NewHTMLReporterWithConfig(&config.OutputConfig{
		IncludeDetails: true,
	})

	var buf bytes.Buffer
	err := reporter.WriteDiff(&buf, diff)
	if err != nil {
		t.Fatalf("HTML diff generation failed: %v", err)
	}

	html := buf.String()

	// Verify Fix #6: OverallTrend renders correctly
	if !strings.Contains(html, "improving") {
		t.Error("Fix #6 failed: OverallTrend not rendered")
	}

	// Verify Fix #7: Regression fields render correctly
	if !strings.Contains(html, "complexity_increase") {
		t.Error("Fix #7 failed: Regression.Type not rendered")
	}
	if !strings.Contains(html, "test.TestFunc") {
		t.Error("Fix #7 failed: Regression.Location not rendered")
	}
	if !strings.Contains(html, "Complexity increased") {
		t.Error("Fix #7 failed: Regression.Description not rendered")
	}

	// Verify Fix #8: Change fields render correctly
	if !strings.Contains(html, "function") {
		t.Error("Fix #8 failed: MetricChange.Category not rendered")
	}
	if !strings.Contains(html, "TestFunc") {
		t.Error("Fix #8 failed: MetricChange.Name not rendered")
	}
	if !strings.Contains(html, "Lines changed") {
		t.Error("Fix #8 failed: MetricChange.Description not rendered")
	}
}
