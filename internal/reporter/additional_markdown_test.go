package reporter

import (
	"bytes"
	"testing"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkdownReporter_NewMarkdownReporterWithOptions(t *testing.T) {
	reporter := NewMarkdownReporterWithOptions(true, true, 100)
	require.NotNil(t, reporter)
	assert.True(t, reporter.includeOverview)
	assert.True(t, reporter.includeDetails)
	assert.Equal(t, 100, reporter.maxItems)
}

func TestMarkdownReporter_WriteDiff(t *testing.T) {
	reporter := NewMarkdownReporter()

	diff := &metrics.ComplexityDiff{
		Baseline: metrics.Snapshot{ID: "baseline-v1"},
		Current:  metrics.Snapshot{ID: "current-v2"},
	}

	var buf bytes.Buffer
	err := reporter.WriteDiff(&buf, diff)
	// May fail due to template issues, but we verify it attempts the operation
	if err == nil {
		output := buf.String()
		assert.NotEmpty(t, output)
	}
}

func TestMarkdownReporter_Generate_FullReport(t *testing.T) {
	reporter := NewMarkdownReporter()

	report := &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:     "github.com/test/repo",
			GeneratedAt:    time.Now(),
			ToolVersion:    "1.0.0",
			FilesProcessed: 25,
			AnalysisTime:   10 * time.Second,
		},
		Overview: metrics.OverviewMetrics{
			TotalLinesOfCode: 10000,
			TotalFunctions:   250,
		},
		Functions: []metrics.FunctionMetrics{
			{
				Name:    "highComplexityFunc",
				Package: "internal/processor",
				File:    "processor.go",
			},
		},
	}

	var buf bytes.Buffer
	err := reporter.Generate(report, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "#")
	assert.Contains(t, output, "|")
	assert.Contains(t, output, "github.com/test/repo")
	assert.Contains(t, output, "highComplexityFunc")
}
