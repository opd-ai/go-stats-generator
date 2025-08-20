package reporter

import (
	"bytes"
	"testing"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// TestCSVReporter_BugReproduction tests that CSV reporter currently returns an error
// This test reproduces the bug and will be converted to a negative test after the fix
func TestCSVReporter_BugReproduction(t *testing.T) {
	// Create a minimal report for testing
	report := &metrics.Report{
		Metadata: metrics.Metadata{
			RepositoryPath: "/test/path",
			AnalyzedAt:     time.Now(),
			Version:        "1.0.0",
			Duration:       time.Millisecond * 100,
		},
		Summary: metrics.Summary{
			FilesProcessed: 1,
			TotalLines:     100,
			TotalFunctions: 5,
		},
		Functions: []metrics.FunctionMetrics{
			{
				Name:       "TestFunction",
				Package:    "main",
				Lines:      metrics.LineMetrics{Code: 10, Comments: 2, Blank: 1},
				Complexity: 3.5,
				IsExported: true,
			},
		},
	}

	// Test that CSV reporter currently returns an error
	csvReporter := &CSVReporter{}
	var buf bytes.Buffer
	
	err := csvReporter.Generate(report, &buf)
	
	// This should fail with "not yet implemented" error - reproducing the bug
	if err == nil {
		t.Fatal("Expected CSV reporter to return error, but got nil")
	}
	
	expectedError := "CSV reporter not yet implemented"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
	
	// Buffer should be empty since generation failed
	if buf.Len() != 0 {
		t.Errorf("Expected empty buffer, got %d bytes", buf.Len())
	}
}

func TestCSVReporter_DiffBugReproduction(t *testing.T) {
	// Test that CSV diff reporter also returns an error
	csvReporter := &CSVReporter{}
	var buf bytes.Buffer
	
	// Create a minimal diff for testing
	diff := &metrics.ComplexityDiff{
		Old: &metrics.Report{},
		New: &metrics.Report{},
	}
	
	err := csvReporter.WriteDiff(&buf, diff)
	
	// This should fail with "not yet implemented" error
	if err == nil {
		t.Fatal("Expected CSV diff reporter to return error, but got nil")
	}
	
	expectedError := "CSV diff reporter not yet implemented"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}
