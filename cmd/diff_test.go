package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiffCommand(t *testing.T) {
	tempDir := t.TempDir()

	baselineReport := createTestReport("baseline", "v1.0.0", "abc123")
	comparisonReport := createTestReport("comparison", "v1.1.0", "def456")
	comparisonReport.Functions = append(comparisonReport.Functions, metrics.FunctionMetrics{
		Name:       "NewFunction",
		File:       "new.go",
		Lines:      metrics.LineMetrics{Total: 10, Code: 8},
		Complexity: metrics.ComplexityScore{Cyclomatic: 5},
	})

	baselineFile := filepath.Join(tempDir, "baseline.json")
	comparisonFile := filepath.Join(tempDir, "comparison.json")

	require.NoError(t, writeReportToFile(baselineReport, baselineFile))
	require.NoError(t, writeReportToFile(comparisonReport, comparisonFile))

	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "console output",
			args:        []string{"diff", baselineFile, comparisonFile},
			expectError: false,
		},
		{
			name:        "json output",
			args:        []string{"diff", baselineFile, comparisonFile, "--format", "json"},
			expectError: false,
		},
		{
			name:        "with output file",
			args:        []string{"diff", baselineFile, comparisonFile, "--output", filepath.Join(tempDir, "diff.txt")},
			expectError: false,
		},
		{
			name:        "changes only",
			args:        []string{"diff", baselineFile, comparisonFile, "--changes-only"},
			expectError: false,
		},
		{
			name:        "with threshold",
			args:        []string{"diff", baselineFile, comparisonFile, "--threshold", "10"},
			expectError: false,
		},
		{
			name:        "missing baseline file",
			args:        []string{"diff", "nonexistent.json", comparisonFile},
			expectError: true,
		},
		{
			name:        "missing comparison file",
			args:        []string{"diff", baselineFile, "nonexistent.json"},
			expectError: true,
		},
		{
			name:        "insufficient args",
			args:        []string{"diff", baselineFile},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd.SetArgs(tt.args)
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)

			err := rootCmd.Execute()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoadReport(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("valid report", func(t *testing.T) {
		report := createTestReport("test", "v1.0.0", "abc123")
		file := filepath.Join(tempDir, "valid.json")
		require.NoError(t, writeReportToFile(report, file))

		loaded, err := loadReport(file)
		require.NoError(t, err)
		assert.NotNil(t, loaded)
		assert.Equal(t, "test", loaded.Metadata.Repository)
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := loadReport(filepath.Join(tempDir, "nonexistent.json"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read file")
	})

	t.Run("invalid json", func(t *testing.T) {
		invalidFile := filepath.Join(tempDir, "invalid.json")
		require.NoError(t, os.WriteFile(invalidFile, []byte("not valid json"), 0o644))

		_, err := loadReport(invalidFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse JSON")
	})
}

func TestLoadBothReports(t *testing.T) {
	tempDir := t.TempDir()

	baselineReport := createTestReport("baseline", "v1.0.0", "abc123")
	comparisonReport := createTestReport("comparison", "v1.1.0", "def456")

	baselineFile := filepath.Join(tempDir, "baseline.json")
	comparisonFile := filepath.Join(tempDir, "comparison.json")

	require.NoError(t, writeReportToFile(baselineReport, baselineFile))
	require.NoError(t, writeReportToFile(comparisonReport, comparisonFile))

	t.Run("both valid", func(t *testing.T) {
		baseline, comparison, err := loadBothReports(baselineFile, comparisonFile)
		require.NoError(t, err)
		assert.NotNil(t, baseline)
		assert.NotNil(t, comparison)
		assert.Equal(t, "baseline", baseline.Metadata.Repository)
		assert.Equal(t, "comparison", comparison.Metadata.Repository)
	})

	t.Run("invalid baseline", func(t *testing.T) {
		_, _, err := loadBothReports("nonexistent.json", comparisonFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load baseline report")
	})

	t.Run("invalid comparison", func(t *testing.T) {
		_, _, err := loadBothReports(baselineFile, "nonexistent.json")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load comparison report")
	})
}

func TestGenerateDiffReport(t *testing.T) {
	baselineReport := createTestReport("baseline", "v1.0.0", "abc123")
	comparisonReport := createTestReport("comparison", "v1.1.0", "def456")

	comparisonReport.Functions = append(comparisonReport.Functions, metrics.FunctionMetrics{
		Name:       "AddedFunction",
		File:       "new.go",
		Lines:      metrics.LineMetrics{Total: 15, Code: 12},
		Complexity: metrics.ComplexityScore{Cyclomatic: 7},
	})

	diffReport, err := generateDiffReport(baselineReport, comparisonReport)
	require.NoError(t, err)
	assert.NotNil(t, diffReport)
}

func TestWriteDiffOutput(t *testing.T) {
	tempDir := t.TempDir()

	baselineReport := createTestReport("baseline", "v1.0.0", "abc123")
	comparisonReport := createTestReport("comparison", "v1.1.0", "def456")

	diffReport, err := generateDiffReport(baselineReport, comparisonReport)
	require.NoError(t, err)

	t.Run("console output to stdout", func(t *testing.T) {
		diffOutputFormat = "console"
		diffOutputFile = ""
		err := writeDiffOutput(diffReport)
		assert.NoError(t, err)
	})

	t.Run("json output to file", func(t *testing.T) {
		diffOutputFormat = "json"
		diffOutputFile = filepath.Join(tempDir, "diff.json")
		err := writeDiffOutput(diffReport)
		assert.NoError(t, err)
		assert.FileExists(t, diffOutputFile)
	})

	t.Run("invalid format", func(t *testing.T) {
		diffOutputFormat = "invalid-format"
		diffOutputFile = ""
		err := writeDiffOutput(diffReport)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create reporter")
	})

	diffOutputFormat = "console"
	diffOutputFile = ""
}

func TestOpenOutputFile(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("stdout when no filename", func(t *testing.T) {
		file, err := openOutputFile("")
		require.NoError(t, err)
		assert.Equal(t, os.Stdout, file)
	})

	t.Run("create new file", func(t *testing.T) {
		filename := filepath.Join(tempDir, "output.txt")
		file, err := openOutputFile(filename)
		require.NoError(t, err)
		assert.NotNil(t, file)
		assert.NotEqual(t, os.Stdout, file)
		file.Close()
		assert.FileExists(t, filename)
	})

	t.Run("invalid path", func(t *testing.T) {
		_, err := openOutputFile("/nonexistent/directory/file.txt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create output file")
	})
}

func createTestReport(targetPath, version, commit string) *metrics.Report {
	now := time.Now()
	return &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:     targetPath,
			ToolVersion:    version,
			GeneratedAt:    now,
			FilesProcessed: 1,
		},
		Functions: []metrics.FunctionMetrics{
			{
				Name:       "TestFunc",
				File:       "test.go",
				Lines:      metrics.LineMetrics{Total: 20, Code: 15},
				Complexity: metrics.ComplexityScore{Cyclomatic: 3},
				Package:    "main",
			},
		},
		Structs:    []metrics.StructMetrics{},
		Interfaces: []metrics.InterfaceMetrics{},
		Packages:   []metrics.PackageMetrics{},
	}
}

func writeReportToFile(report *metrics.Report, filename string) error {
	data, err := json.Marshal(report)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0o644)
}

func TestDiffFlagParsing(t *testing.T) {
	cmd := diffCmd

	formatFlag := cmd.Flags().Lookup("format")
	assert.NotNil(t, formatFlag)
	assert.Equal(t, "console", formatFlag.DefValue)

	outputFlag := cmd.Flags().Lookup("output")
	assert.NotNil(t, outputFlag)

	changesOnlyFlag := cmd.Flags().Lookup("changes-only")
	assert.NotNil(t, changesOnlyFlag)

	thresholdFlag := cmd.Flags().Lookup("threshold")
	assert.NotNil(t, thresholdFlag)
	assert.Equal(t, "5", thresholdFlag.DefValue)
}

func TestRunDiff(t *testing.T) {
	tempDir := t.TempDir()

	baselineReport := createTestReport("baseline", "v1.0.0", "abc123")
	comparisonReport := createTestReport("comparison", "v1.1.0", "def456")

	baselineFile := filepath.Join(tempDir, "baseline.json")
	comparisonFile := filepath.Join(tempDir, "comparison.json")

	require.NoError(t, writeReportToFile(baselineReport, baselineFile))
	require.NoError(t, writeReportToFile(comparisonReport, comparisonFile))

	t.Run("successful diff", func(t *testing.T) {
		diffOutputFormat = "console"
		diffOutputFile = ""

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := runDiff(diffCmd, []string{baselineFile, comparisonFile})

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)

		assert.NoError(t, err)
	})

	t.Run("invalid baseline", func(t *testing.T) {
		err := runDiff(diffCmd, []string{"nonexistent.json", comparisonFile})
		assert.Error(t, err)
	})
}
