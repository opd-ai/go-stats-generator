package reporter

import (
	"bytes"
	"testing"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/stretchr/testify/assert"
)

func TestConsoleReporter_WriteDiff_Basic(t *testing.T) {
	reporter := NewConsoleReporter(nil)

	diff := &metrics.ComplexityDiff{
		Baseline: metrics.Snapshot{ID: "baseline-v1"},
		Current:  metrics.Snapshot{ID: "current-v2"},
	}

	var buf bytes.Buffer
	err := reporter.WriteDiff(&buf, diff)
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Diff Report")
	assert.Contains(t, output, "baseline-v1")
	assert.Contains(t, output, "current-v2")
}

func TestConsoleReporter_WriteDiff_WithRegressions(t *testing.T) {
	reporter := NewConsoleReporter(nil)

	diff := &metrics.ComplexityDiff{
		Baseline: metrics.Snapshot{ID: "baseline"},
		Current:  metrics.Snapshot{ID: "current"},
		Regressions: []metrics.Regression{
			{
				Type:     metrics.ComplexityRegression,
				Function: "processData",
				File:     "processor.go",
				Impact:   metrics.ImpactLevelHigh,
			},
		},
	}

	var buf bytes.Buffer
	err := reporter.WriteDiff(&buf, diff)
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "REGRESSIONS")
	assert.Contains(t, output, "processor.go")
}

func TestConsoleReporter_WriteDiff_WithImprovements(t *testing.T) {
	reporter := NewConsoleReporter(nil)

	diff := &metrics.ComplexityDiff{
		Baseline: metrics.Snapshot{ID: "baseline"},
		Current:  metrics.Snapshot{ID: "current"},
		Improvements: []metrics.Improvement{
			{
				Type:     metrics.ComplexityImprovement,
				Function: "validateInput",
				File:     "validator.go",
				Impact:   metrics.ImpactLevelHigh,
			},
		},
	}

	var buf bytes.Buffer
	err := reporter.WriteDiff(&buf, diff)
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "IMPROVEMENTS")
	assert.Contains(t, output, "validator.go")
}

func TestConsoleReporter_WriteDiff_WithChanges(t *testing.T) {
	cfg := &config.OutputConfig{
		IncludeDetails: true,
	}
	reporter := NewConsoleReporter(cfg)

	diff := &metrics.ComplexityDiff{
		Baseline: metrics.Snapshot{ID: "baseline"},
		Current:  metrics.Snapshot{ID: "current"},
		Changes: []metrics.MetricChange{
			{
				Name: "modifiedFunc",
				File: "modified.go",
			},
		},
	}

	var buf bytes.Buffer
	err := reporter.WriteDiff(&buf, diff)
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "modifiedFunc")
}

func TestConsoleReporter_WithRefactoringSuggestions(t *testing.T) {
	cfg := &config.OutputConfig{
		IncludeDetails: true,
	}
	reporter := NewConsoleReporter(cfg)

	report := &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:  "test-repo",
			GeneratedAt: time.Now(),
			ToolVersion: "1.0.0",
		},
		Suggestions: []metrics.SuggestionInfo{
			{
				Category:      "complexity",
				Description:   "Reduce complexity",
				Target:        "processData",
				Location:      "processor.go:45",
				Effort:        "medium",
				AffectedLines: 50,
				MBIImpact:     25.5,
				ImpactEffort:  0.85,
			},
		},
	}

	var buf bytes.Buffer
	err := reporter.Generate(report, &buf)
	assert.NoError(t, err)

	output := buf.String()
	assert.NotEmpty(t, output)
}

func TestConsoleReporter_WithEmptyRefactoringSuggestions(t *testing.T) {
	cfg := &config.OutputConfig{
		IncludeDetails: true,
	}
	reporter := NewConsoleReporter(cfg)

	report := &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:  "test-repo",
			GeneratedAt: time.Now(),
			ToolVersion: "1.0.0",
		},
		Suggestions: []metrics.SuggestionInfo{},
	}

	var buf bytes.Buffer
	err := reporter.Generate(report, &buf)
	assert.NoError(t, err)

	// With empty suggestions and IncludeDetails, the section might not appear
	// Just verify no error occurred
	output := buf.String()
	assert.NotEmpty(t, output)
}

func TestConsoleReporter_WithManyRefactoringSuggestions(t *testing.T) {
	cfg := &config.OutputConfig{
		IncludeDetails: true,
	}
	reporter := NewConsoleReporter(cfg)

	suggestions := make([]metrics.SuggestionInfo, 25)
	for i := 0; i < 25; i++ {
		suggestions[i] = metrics.SuggestionInfo{
			Category:      "complexity",
			Description:   "Refactor",
			Target:        "func",
			Location:      "file.go:100",
			Effort:        "low",
			AffectedLines: 10,
			MBIImpact:     5.0,
			ImpactEffort:  0.5,
		}
	}

	report := &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:  "test-repo",
			GeneratedAt: time.Now(),
			ToolVersion: "1.0.0",
		},
		Suggestions: suggestions,
	}

	var buf bytes.Buffer
	err := reporter.Generate(report, &buf)
	assert.NoError(t, err)

	output := buf.String()
	assert.NotEmpty(t, output)
}
