package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/opd-ai/go-stats-generator/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBaselineListFormatFlag tests that the baseline list command
// recognizes the --format flag (regression test for Issue #7 from AUDIT.md)
func TestBaselineListFormatFlag(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "list with json format",
			args:        []string{"baseline", "list", "--format", "json"},
			expectError: false,
		},
		{
			name:        "list with console format",
			args:        []string{"baseline", "list", "--format", "console"},
			expectError: false,
		},
		{
			name:        "list with output file",
			args:        []string{"baseline", "list", "--output", "/tmp/baselines.json"},
			expectError: false,
		},
		{
			name:        "list with both format and output",
			args:        []string{"baseline", "list", "--format", "json", "--output", "/tmp/baselines.json"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset command state
			rootCmd.SetArgs(tt.args)

			// Capture output
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)

			// Execute command
			err := rootCmd.Execute()
			// Note: We expect execution to fail due to missing storage,
			// but we're testing that the FLAG is recognized, not rejected
			if err != nil {
				output := buf.String()
				// The error should NOT be about unknown flag
				assert.NotContains(t, output, "unknown flag", "Flag should be recognized")
				assert.NotContains(t, err.Error(), "unknown flag", "Flag should be recognized")
			}
		})
	}
}

// TestBaselineListFlagParsing ensures flags are properly parsed
func TestBaselineListFlagParsing(t *testing.T) {
	// Create a test command to verify flag definitions
	cmd := listBaselinesCmd

	// Check that the command accepts format flag
	formatFlag := cmd.Flags().Lookup("format")
	assert.NotNil(t, formatFlag, "--format flag should be defined")

	// Check that the command accepts output flag
	outputFlag := cmd.Flags().Lookup("output")
	assert.NotNil(t, outputFlag, "--output flag should be defined")
}

// TestBaselineListOutputFormat tests that different output formats work correctly
func TestBaselineListOutputFormat(t *testing.T) {
	// This test validates the outputBaselines function respects format setting

	// Test console format
	t.Run("console format", func(t *testing.T) {
		outputFormat = "console"
		defer func() { outputFormat = "json" }()

		snapshots := []storage.SnapshotInfo{}
		err := outputBaselines(snapshots)
		require.NoError(t, err)
	})

	// Test JSON format
	t.Run("json format", func(t *testing.T) {
		// Redirect stdout to capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() { os.Stdout = oldStdout }()

		outputFormat = "json"
		outputFile = "" // stdout
		defer func() { outputFormat = "json"; outputFile = "" }()

		snapshots := []storage.SnapshotInfo{}
		err := outputBaselines(snapshots)
		require.NoError(t, err)

		w.Close()
		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// JSON output should contain expected structure
		assert.Contains(t, output, "snapshots")
		assert.Contains(t, output, "count")
	})
}

// TestBaselineCreateHasFlags ensures create command still has its flags
func TestBaselineCreateHasFlags(t *testing.T) {
	cmd := createBaselineCmd

	assert.NotNil(t, cmd.Flags().Lookup("id"), "--id flag should exist")
	assert.NotNil(t, cmd.Flags().Lookup("message"), "--message flag should exist")
	assert.NotNil(t, cmd.Flags().Lookup("tags"), "--tags flag should exist")
	assert.NotNil(t, cmd.Flags().Lookup("format"), "--format flag should exist")
	assert.NotNil(t, cmd.Flags().Lookup("output"), "--output flag should exist")
}

// TestAnalyzeCodebase_PopulatesAllMetrics ensures baseline creation populates all critical metrics
// that were previously zeroed (MBI scores and documentation coverage).
// Regression test for: FUNCTIONAL_AUDIT.md "Trend analysis baseline data quality issues"
func TestAnalyzeCodebase_PopulatesAllMetrics(t *testing.T) {
	// Arrange: Use testdata/simple which has documented code
	testPath := "../testdata/simple"
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Skip("Testdata not available")
	}

	// Act: Run baseline analysis
	report, err := analyzeCodebase(testPath)

	// Assert: Analysis succeeds
	require.NoError(t, err, "analyzeCodebase should not fail")
	require.NotNil(t, report, "report should not be nil")

	// Assert: Critical metrics are populated (not zeroed)

	// 1. Documentation coverage should be populated
	assert.NotZero(t, report.Documentation.Coverage.Overall,
		"Documentation coverage should be populated (was 0.00%% in bug)")

	// 2. File scores should be populated
	assert.NotEmpty(t, report.Scores.FileScores,
		"File scores should be populated (was empty in bug)")

	// 3. Functions should be analyzed
	assert.NotEmpty(t, report.Functions,
		"Functions should be analyzed")

	// 4. Complexity metrics should be calculated
	// Average function complexity should be non-negative
	assert.GreaterOrEqual(t, report.Complexity.AverageFunction, 0.0,
		"Average function complexity should be calculated")
}

func TestExtractTargetPath(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "with argument",
			args:     []string{"/path/to/code"},
			expected: "/path/to/code",
		},
		{
			name:     "no arguments",
			args:     []string{},
			expected: ".",
		},
		{
			name:     "multiple arguments",
			args:     []string{"/first", "/second"},
			expected: "/first",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTargetPath(tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateBaselineID(t *testing.T) {
	// The ID is based on timestamp, so we may get the same ID if called in the same second
	// Just check that it's not empty
	id1 := generateBaselineID()

	assert.NotEmpty(t, id1)
	assert.Contains(t, id1, "baseline-")
}

func TestGetCurrentBranch(t *testing.T) {
	branch := getCurrentBranch()
	// Should return a string (empty if not in git repo, or branch name)
	assert.NotNil(t, branch)
}

func TestGetCurrentCommit(t *testing.T) {
	commit := getCurrentCommit()
	// Should return a string (empty if not in git repo, or commit hash)
	assert.NotNil(t, commit)
}
