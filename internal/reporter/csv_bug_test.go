package reporter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// TestCSVReporter_Generate tests that CSV reporter now works correctly
// This test was converted from a bug reproduction test after the fix
func TestCSVReporter_Generate(t *testing.T) {
	// Create a minimal report for testing
	report := &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:     "/test/path",
			GeneratedAt:    time.Now(),
			ToolVersion:    "1.0.0",
			AnalysisTime:   time.Millisecond * 100,
			FilesProcessed: 1,
		},
		Overview: metrics.OverviewMetrics{
			TotalFunctions: 5,
			TotalFiles:     1,
		},
		Functions: []metrics.FunctionMetrics{
			{
				Name:    "TestFunction",
				Package: "main",
				Lines:   metrics.LineMetrics{Code: 10, Comments: 2, Blank: 1},
				Complexity: metrics.ComplexityScore{
					Overall: 3.5,
				},
				IsExported: true,
			},
		},
	}

	// Test that CSV reporter now works correctly
	csvReporter := &CSVReporter{}
	var buf bytes.Buffer

	err := csvReporter.Generate(report, &buf)

	// This should now succeed
	if err != nil {
		t.Fatalf("Expected CSV reporter to succeed, but got error: %v", err)
	}

	// Buffer should contain CSV data
	if buf.Len() == 0 {
		t.Error("Expected CSV output, but buffer is empty")
	}

	// Verify it contains expected CSV content
	output := buf.String()
	if !strings.Contains(output, "# METADATA") {
		t.Error("Expected CSV output to contain metadata section")
	}
	if !strings.Contains(output, "TestFunction") {
		t.Error("Expected CSV output to contain function data")
	}
}

func TestCSVReporter_WriteDiff(t *testing.T) {
	// Test that CSV diff reporter now works correctly
	csvReporter := &CSVReporter{}
	var buf bytes.Buffer

	// Create a minimal diff for testing
	diff := &metrics.ComplexityDiff{
		Baseline: metrics.MetricsSnapshot{},
		Current:  metrics.MetricsSnapshot{},
		Summary: metrics.DiffSummary{
			TotalChanges:     5,
			RegressionCount:  1,
			ImprovementCount: 2,
		},
	}

	err := csvReporter.WriteDiff(&buf, diff)

	// This should now succeed
	if err != nil {
		t.Fatalf("Expected CSV diff reporter to succeed, but got error: %v", err)
	}

	// Verify it contains expected content
	output := buf.String()
	if !strings.Contains(output, "# SUMMARY") {
		t.Error("Expected CSV diff output to contain summary section")
	}
}
