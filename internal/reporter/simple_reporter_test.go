package reporter

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONReporter_WriteDiff_Simple(t *testing.T) {
	reporter := NewJSONReporter()

	diff := &metrics.ComplexityDiff{
		Baseline: metrics.Snapshot{ID: "baseline-v1"},
		Current:  metrics.Snapshot{ID: "current-v2"},
	}

	var buf bytes.Buffer
	err := reporter.WriteDiff(&buf, diff)
	require.NoError(t, err)

	var decoded metrics.ComplexityDiff
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)
	assert.Equal(t, "baseline-v1", decoded.Baseline.ID)
	assert.Equal(t, "current-v2", decoded.Current.ID)
}

func TestJSONReporter_Generate_Simple(t *testing.T) {
	reporter := NewJSONReporter()

	report := &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:  "github.com/test/repo",
			GeneratedAt: time.Now(),
			ToolVersion: "1.0.0",
		},
	}

	var buf bytes.Buffer
	err := reporter.Generate(report, &buf)
	require.NoError(t, err)

	var decoded metrics.Report
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)
	assert.Equal(t, "github.com/test/repo", decoded.Metadata.Repository)
}

func TestNewReporter_AllTypes(t *testing.T) {
	tests := []struct {
		name         string
		reporterType string
		wantErr      bool
	}{
		{"JSON", "json", false},
		{"CSV", "csv", false},
		{"HTML", "html", false},
		{"Markdown", "markdown", false},
		{"Console", "console", false},
		{"Invalid", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter, err := NewReporter(tt.reporterType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, reporter)
			}
		})
	}
}

func TestCreateReporter_AllTypes(t *testing.T) {
	types := []ReporterType{TypeJSON, TypeCSV, TypeHTML, TypeMarkdown, TypeConsole}
	for _, rtype := range types {
		reporter := CreateReporter(rtype, nil)
		assert.NotNil(t, reporter)
	}
}
