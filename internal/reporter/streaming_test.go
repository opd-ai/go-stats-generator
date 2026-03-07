package reporter

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONReporter_StreamingMode(t *testing.T) {
	reporter := NewJSONReporter()
	var buf bytes.Buffer

	// Create test metadata
	metadata := &metrics.ReportMetadata{
		Repository:     "test-repo",
		GeneratedAt:    time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC),
		AnalysisTime:   5 * time.Second,
		FilesProcessed: 10,
		ToolVersion:    "1.0.0",
		GoVersion:      "1.24.0",
	}

	// Begin report
	err := reporter.BeginReport(&buf, metadata)
	require.NoError(t, err)

	// Write overview section
	overview := metrics.OverviewMetrics{
		TotalLinesOfCode: 1000,
		TotalFunctions:   50,
		TotalMethods:     25,
	}
	err = reporter.WriteSection(&buf, "overview", overview)
	require.NoError(t, err)

	// Write functions section
	functions := []metrics.FunctionMetrics{
		{
			Name:  "TestFunction",
			File:  "test.go",
			Lines: metrics.LineMetrics{Total: 10, Code: 8, Comments: 1, Blank: 1},
			Complexity: metrics.ComplexityScore{
				Cyclomatic: 3,
			},
		},
	}
	err = reporter.WriteSection(&buf, "functions", functions)
	require.NoError(t, err)

	// End report
	err = reporter.EndReport(&buf)
	require.NoError(t, err)

	// Verify output is valid JSON
	var result map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err, "Output should be valid JSON")

	// Verify metadata
	assert.Contains(t, result, "metadata")
	metadataMap := result["metadata"].(map[string]interface{})
	assert.Equal(t, "test-repo", metadataMap["repository"])
	assert.Equal(t, float64(10), metadataMap["files_processed"])

	// Verify overview
	assert.Contains(t, result, "overview")
	overviewMap := result["overview"].(map[string]interface{})
	assert.Equal(t, float64(1000), overviewMap["total_lines_of_code"])
	assert.Equal(t, float64(50), overviewMap["total_functions"])

	// Verify functions
	assert.Contains(t, result, "functions")
	functionsArray := result["functions"].([]interface{})
	assert.Len(t, functionsArray, 1)
	functionMap := functionsArray[0].(map[string]interface{})
	assert.Equal(t, "TestFunction", functionMap["name"])
}

func TestJSONReporter_StreamingMatchesNonStreaming(t *testing.T) {
	// Create a complete report
	report := &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:     "test-repo",
			GeneratedAt:    time.Date(2026, 3, 7, 12, 0, 0, 0, time.UTC),
			AnalysisTime:   5 * time.Second,
			FilesProcessed: 10,
			ToolVersion:    "1.0.0",
			GoVersion:      "1.24.0",
		},
		Overview: metrics.OverviewMetrics{
			TotalLinesOfCode: 1000,
			TotalFunctions:   50,
			TotalMethods:     25,
		},
		Functions: []metrics.FunctionMetrics{
			{
				Name:  "TestFunction",
				File:  "test.go",
				Lines: metrics.LineMetrics{Total: 10, Code: 8, Comments: 1, Blank: 1},
				Complexity: metrics.ComplexityScore{
					Cyclomatic: 3,
				},
			},
		},
		Packages:             []metrics.PackageMetrics{},
		CircularDependencies: []metrics.CircularDependency{},
	}

	// Generate using non-streaming mode
	reporter1 := NewJSONReporter()
	var buf1 bytes.Buffer
	err := reporter1.Generate(report, &buf1)
	require.NoError(t, err)

	// Generate using streaming mode
	reporter2 := NewJSONReporter()
	var buf2 bytes.Buffer
	err = reporter2.BeginReport(&buf2, &report.Metadata)
	require.NoError(t, err)
	err = reporter2.WriteSection(&buf2, "overview", report.Overview)
	require.NoError(t, err)
	err = reporter2.WriteSection(&buf2, "functions", report.Functions)
	require.NoError(t, err)
	err = reporter2.WriteSection(&buf2, "packages", report.Packages)
	require.NoError(t, err)
	err = reporter2.WriteSection(&buf2, "circular_dependencies", report.CircularDependencies)
	require.NoError(t, err)
	err = reporter2.EndReport(&buf2)
	require.NoError(t, err)

	// Parse both outputs
	var result1, result2 map[string]interface{}
	err = json.Unmarshal(buf1.Bytes(), &result1)
	require.NoError(t, err)
	err = json.Unmarshal(buf2.Bytes(), &result2)
	require.NoError(t, err)

	// Compare critical fields (metadata, overview, functions)
	// Note: Non-streaming includes ALL fields, streaming only includes written sections
	assert.Equal(t, result1["metadata"], result2["metadata"])
	assert.Equal(t, result1["overview"], result2["overview"])
	assert.Equal(t, result1["functions"], result2["functions"])
}

func TestJSONReporter_StreamingEmptySections(t *testing.T) {
	reporter := NewJSONReporter()
	var buf bytes.Buffer

	metadata := &metrics.ReportMetadata{
		Repository:  "test",
		GeneratedAt: time.Now(),
	}

	// Begin and end with no sections
	err := reporter.BeginReport(&buf, metadata)
	require.NoError(t, err)
	err = reporter.EndReport(&buf)
	require.NoError(t, err)

	// Should still be valid JSON
	var result map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	// Should only have metadata
	assert.Contains(t, result, "metadata")
	assert.Len(t, result, 1)
}

func TestJSONReporter_StreamingMultipleSections(t *testing.T) {
	reporter := NewJSONReporter()
	var buf bytes.Buffer

	metadata := &metrics.ReportMetadata{
		Repository:  "test",
		GeneratedAt: time.Now(),
	}

	err := reporter.BeginReport(&buf, metadata)
	require.NoError(t, err)

	// Write 5 different sections
	sections := []string{"section1", "section2", "section3", "section4", "section5"}
	for i, name := range sections {
		data := map[string]int{"value": i}
		err = reporter.WriteSection(&buf, name, data)
		require.NoError(t, err)
	}

	err = reporter.EndReport(&buf)
	require.NoError(t, err)

	// Verify all sections present
	var result map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)

	assert.Len(t, result, 6) // metadata + 5 sections
	for _, name := range sections {
		assert.Contains(t, result, name)
	}
}

func TestJSONReporter_StreamingProperCommas(t *testing.T) {
	reporter := NewJSONReporter()
	var buf bytes.Buffer

	metadata := &metrics.ReportMetadata{Repository: "test", GeneratedAt: time.Now()}

	err := reporter.BeginReport(&buf, metadata)
	require.NoError(t, err)

	err = reporter.WriteSection(&buf, "section1", map[string]int{"a": 1})
	require.NoError(t, err)

	err = reporter.WriteSection(&buf, "section2", map[string]int{"b": 2})
	require.NoError(t, err)

	err = reporter.EndReport(&buf)
	require.NoError(t, err)

	output := buf.String()

	// Check for proper JSON structure (no duplicate commas, no trailing commas)
	assert.NotContains(t, output, ",,")
	assert.NotContains(t, output, ",\n}")

	// Verify valid JSON
	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)
}

func TestJSONReporter_StreamingIndentation(t *testing.T) {
	reporter := NewJSONReporter()
	var buf bytes.Buffer

	metadata := &metrics.ReportMetadata{Repository: "test", GeneratedAt: time.Now()}

	err := reporter.BeginReport(&buf, metadata)
	require.NoError(t, err)

	err = reporter.WriteSection(&buf, "overview", metrics.OverviewMetrics{TotalFunctions: 10})
	require.NoError(t, err)

	err = reporter.EndReport(&buf)
	require.NoError(t, err)

	output := buf.String()

	// Check indentation is present
	assert.True(t, strings.Contains(output, "  \"metadata\":"), "Should have indented metadata")
	assert.True(t, strings.Contains(output, "  \"overview\":"), "Should have indented overview")
	assert.True(t, strings.Contains(output, "    \"repository\":"), "Should have nested indentation")
}
