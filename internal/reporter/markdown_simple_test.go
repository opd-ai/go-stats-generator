package reporter

import (
	"bytes"
	"testing"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

func TestMarkdownReporter_SimpleGenerate(t *testing.T) {
	reporter := NewMarkdownReporter()

	// Create the simplest possible report data
	testReport := &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:     "test-repo",
			GeneratedAt:    time.Now(),
			AnalysisTime:   time.Second,
			FilesProcessed: 1,
			ToolVersion:    "v1.0.0",
			GoVersion:      "go1.21.0",
		},
		Overview: metrics.OverviewMetrics{
			TotalLinesOfCode: 100,
			TotalFunctions:   10,
			TotalMethods:     5,
			TotalStructs:     3,
			TotalInterfaces:  2,
			TotalPackages:    1,
			TotalFiles:       1,
		},
		Functions:  []metrics.FunctionMetrics{},
		Structs:    []metrics.StructMetrics{},
		Interfaces: []metrics.InterfaceMetrics{},
		Packages:   []metrics.PackageMetrics{},
		Patterns:   metrics.PatternMetrics{},
		Complexity: metrics.ComplexityMetrics{
			AverageFunction: 5.0,
		},
		Documentation: metrics.DocumentationMetrics{
			Coverage: metrics.DocumentationCoverage{
				Overall: 0.8,
			},
		},
	}

	var buf bytes.Buffer
	err := reporter.Generate(testReport, &buf)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("Generated report is empty")
	}

	// Basic checks
	if output == "" {
		t.Error("Expected non-empty output")
	}

	t.Logf("Generated report length: %d characters", len(output))
	t.Logf("First 200 characters: %s", output[:min(200, len(output))])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
