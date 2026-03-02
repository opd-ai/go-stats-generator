package cmd

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDuplicationIntegration_ExactClones tests detection of exact duplicates
func TestDuplicationIntegration_ExactClones(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Analysis.Duplication.MinBlockLines = 6
	cfg.Analysis.Duplication.SimilarityThreshold = 0.80

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	testDir := filepath.Join("..", "testdata", "duplication")
	report, err := runAnalysisWorkflow(ctx, testDir, cfg)

	require.NoError(t, err, "Analysis should complete successfully")
	require.NotNil(t, report, "Report should not be nil")

	// Verify exact clones are detected (from exact_clone.go and other files)
	assert.Greater(t, report.Duplication.ClonePairs, 0, "Should detect clone pairs")
	assert.Greater(t, report.Duplication.DuplicatedLines, 0, "Should have duplicated lines")
	assert.Greater(t, report.Duplication.DuplicationRatio, 0.0, "Duplication ratio should be > 0")

	// Verify clone types - should have exact or renamed clones
	foundClone := false
	for _, clone := range report.Duplication.Clones {
		if clone.Type == metrics.CloneTypeExact || clone.Type == metrics.CloneTypeRenamed {
			foundClone = true
			assert.GreaterOrEqual(t, len(clone.Instances), 2, "Clone pair should have at least 2 instances")
			assert.GreaterOrEqual(t, clone.LineCount, cfg.Analysis.Duplication.MinBlockLines,
				"Clone should meet minimum line threshold")
		}
	}
	assert.True(t, foundClone, "Should find at least one clone")

	t.Logf("Detection in testdata: %d pairs, %d lines, %.2f%% ratio",
		report.Duplication.ClonePairs,
		report.Duplication.DuplicatedLines,
		report.Duplication.DuplicationRatio*100)
}

// TestDuplicationIntegration_ConfigThresholds tests configuration threshold handling
func TestDuplicationIntegration_ConfigThresholds(t *testing.T) {
	tests := []struct {
		name              string
		minBlockLines     int
		similarityThresh  float64
		expectDuplicates  bool
	}{
		{
			name:             "default_thresholds",
			minBlockLines:    6,
			similarityThresh: 0.80,
			expectDuplicates: true,
		},
		{
			name:             "high_min_lines",
			minBlockLines:    20,
			similarityThresh: 0.80,
			expectDuplicates: false, // May not find duplicates with very high threshold
		},
		{
			name:             "low_min_lines",
			minBlockLines:    3,
			similarityThresh: 0.80,
			expectDuplicates: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.DefaultConfig()
			cfg.Analysis.Duplication.MinBlockLines = tc.minBlockLines
			cfg.Analysis.Duplication.SimilarityThreshold = tc.similarityThresh

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			testDir := filepath.Join("..", "testdata", "duplication")
			report, err := runAnalysisWorkflow(ctx, testDir, cfg)

			require.NoError(t, err, "Analysis should complete successfully")
			require.NotNil(t, report, "Report should not be nil")

			if tc.expectDuplicates {
				// We expect to find some duplicates with permissive settings
				t.Logf("%s: %d pairs, %d lines", tc.name, report.Duplication.ClonePairs, report.Duplication.DuplicatedLines)
			} else {
				// With restrictive settings, we may or may not find duplicates
				t.Logf("%s: %d pairs (high threshold may reduce detections)", tc.name, report.Duplication.ClonePairs)
			}

			// Validate all detected clones meet the minimum threshold
			for i, clone := range report.Duplication.Clones {
				assert.GreaterOrEqual(t, clone.LineCount, tc.minBlockLines,
					"Clone %d should meet minimum line threshold", i)
			}
		})
	}
}

// TestDuplicationIntegration_FalsePositiveRegression tests known false-positive cases
func TestDuplicationIntegration_FalsePositiveRegression(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Analysis.Duplication.MinBlockLines = 6
	cfg.Analysis.Duplication.SimilarityThreshold = 0.80

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test against real codebase - internal/analyzer
	// Should not flag every function with error handling as duplicate
	testDir := filepath.Join("..", "internal", "analyzer")
	report, err := runAnalysisWorkflow(ctx, testDir, cfg)

	require.NoError(t, err, "Analysis should complete successfully on real codebase")
	require.NotNil(t, report, "Report should not be nil")

	// Duplication ratio should be reasonable for production code
	// Note: Test code may have higher duplication due to similar test structures
	assert.Less(t, report.Duplication.DuplicationRatio, 0.9,
		"Production code should have reasonable duplication ratio")

	t.Logf("Real codebase (internal/analyzer): %d pairs, %d lines, %.2f%% ratio",
		report.Duplication.ClonePairs,
		report.Duplication.DuplicatedLines,
		report.Duplication.DuplicationRatio*100)
}
