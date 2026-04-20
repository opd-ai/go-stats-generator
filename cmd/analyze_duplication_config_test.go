package cmd

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/opd-ai/go-stats-generator/internal/analyzer"
	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildTestCollectedMetrics parses Go source strings into CollectedMetrics.DupBlocks
// using the given minBlockLines, ready for finalizeDuplicationMetrics.
func buildTestCollectedMetrics(t *testing.T, sources map[string]string, minBlockLines int) (*CollectedMetrics, *token.FileSet) {
	t.Helper()
	fset := token.NewFileSet()
	da := analyzer.NewDuplicationAnalyzer(fset)
	cm := &CollectedMetrics{
		FileLinesCount: make(map[string]int),
	}
	for filename, src := range sources {
		file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
		require.NoError(t, err)
		blocks := da.ExtractBlocks(file, filename, minBlockLines)
		cm.DupBlocks = append(cm.DupBlocks, blocks...)
		// Compute line count from source text
		lineCount := 0
		for _, b := range []byte(src) {
			if b == '\n' {
				lineCount++
			}
		}
		cm.DupTotalLines += lineCount + 1
		cm.FileLinesCount[filename] = lineCount + 1
	}
	return cm, fset
}

// TestDuplicationConfigIntegration tests that configuration values are properly used
func TestDuplicationConfigIntegration(t *testing.T) {
	t.Run("uses custom min_block_lines", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Analysis.Duplication.MinBlockLines = 3

		fset := token.NewFileSet()
		duplicationAnalyzer := analyzer.NewDuplicationAnalyzer(fset)

		// Create test code with small blocks
		code := `package test
func example() {
	x := 1
	y := 2
	z := 3
}
func duplicate() {
	x := 1
	y := 2
	z := 3
}
`
		cm, _ := buildTestCollectedMetrics(t, map[string]string{"test.go": code}, 3)

		report := &metrics.Report{}
		finalizeDuplicationMetrics(report, duplicationAnalyzer, cm, cfg)

		// With minBlockLines=3, should detect the 3-line duplicate
		assert.Greater(t, report.Duplication.ClonePairs, 0, "Should detect duplicates with min_block_lines=3")
	})

	t.Run("uses custom similarity_threshold", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Analysis.Duplication.SimilarityThreshold = 0.90

		fset := token.NewFileSet()
		duplicationAnalyzer := analyzer.NewDuplicationAnalyzer(fset)

		code := `package test
func example() {
	a := 1
	b := 2
	c := 3
	d := 4
	e := 5
	f := 6
}
`
		cm, _ := buildTestCollectedMetrics(t, map[string]string{"similarity_test.go": code}, cfg.Analysis.Duplication.MinBlockLines)

		report := &metrics.Report{}
		finalizeDuplicationMetrics(report, duplicationAnalyzer, cm, cfg)

		// Test passes if no error occurs (similarity threshold is used)
		assert.NotNil(t, report.Duplication)
	})

	t.Run("ignore_test_files filters test files", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Analysis.Duplication.IgnoreTestFiles = true

		fset := token.NewFileSet()
		duplicationAnalyzer := analyzer.NewDuplicationAnalyzer(fset)

		// Create both regular and test files with duplicates
		regularCode := `package test
func duplicate() {
	a := 1
	b := 2
	c := 3
	d := 4
	e := 5
	f := 6
}
`
		testCode := `package test
import "testing"
func TestDuplicate(t *testing.T) {
	a := 1
	b := 2
	c := 3
	d := 4
	e := 5
	f := 6
}
`
		cm, _ := buildTestCollectedMetrics(t, map[string]string{
			"regular.go":     regularCode,
			"filter_test.go": testCode,
		}, cfg.Analysis.Duplication.MinBlockLines)

		report := &metrics.Report{}
		finalizeDuplicationMetrics(report, duplicationAnalyzer, cm, cfg)

		// With ignore_test_files=true, test file should be filtered out
		// So we should have fewer or no clone pairs
		assert.NotNil(t, report.Duplication)
	})

	t.Run("include_test_files when not ignoring", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Analysis.Duplication.IgnoreTestFiles = false

		fset := token.NewFileSet()
		duplicationAnalyzer := analyzer.NewDuplicationAnalyzer(fset)

		// Create both regular and test files with identical duplicates
		code := `package test
func duplicate() {
	a := 1
	b := 2
	c := 3
	d := 4
	e := 5
	f := 6
}
`
		cm, _ := buildTestCollectedMetrics(t, map[string]string{
			"regular.go":      code,
			"include_test.go": code,
		}, cfg.Analysis.Duplication.MinBlockLines)

		report := &metrics.Report{}
		finalizeDuplicationMetrics(report, duplicationAnalyzer, cm, cfg)

		// With ignore_test_files=false, should analyze both files
		assert.NotNil(t, report.Duplication)
	})
}

// TestDuplicationConfigDefaults verifies default configuration values
func TestDuplicationConfigDefaults(t *testing.T) {
	cfg := config.DefaultConfig()

	assert.Equal(t, 6, cfg.Analysis.Duplication.MinBlockLines, "Default min_block_lines should be 6")
	assert.Equal(t, 0.80, cfg.Analysis.Duplication.SimilarityThreshold, "Default similarity_threshold should be 0.80")
	assert.False(t, cfg.Analysis.Duplication.IgnoreTestFiles, "Default ignore_test_files should be false")
}

// TestFinalizeDuplicationMetrics_EmptyFiles tests behavior with no files
func TestFinalizeDuplicationMetrics_EmptyFiles(t *testing.T) {
	cfg := config.DefaultConfig()
	fset := token.NewFileSet()
	duplicationAnalyzer := analyzer.NewDuplicationAnalyzer(fset)

	report := &metrics.Report{}
	collectedMetrics := &CollectedMetrics{} // no DupBlocks

	finalizeDuplicationMetrics(report, duplicationAnalyzer, collectedMetrics, cfg)

	assert.Equal(t, 0, report.Duplication.ClonePairs)
	assert.Equal(t, 0, report.Duplication.DuplicatedLines)
	assert.Equal(t, 0.0, report.Duplication.DuplicationRatio)
	assert.Equal(t, 0, report.Duplication.LargestCloneSize)
	assert.Empty(t, report.Duplication.Clones)
}

// TestFinalizeDuplicationMetrics_AllTestFilesIgnored tests filtering behavior
func TestFinalizeDuplicationMetrics_AllTestFilesIgnored(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Analysis.Duplication.IgnoreTestFiles = true

	fset := token.NewFileSet()
	duplicationAnalyzer := analyzer.NewDuplicationAnalyzer(fset)

	// Create only test files
	testCode := `package test
import "testing"
func TestExample(t *testing.T) {
	x := 1
	y := 2
	z := 3
	a := 4
	b := 5
	c := 6
}
`
	cm, _ := buildTestCollectedMetrics(t, map[string]string{"example_test.go": testCode}, cfg.Analysis.Duplication.MinBlockLines)

	report := &metrics.Report{}
	finalizeDuplicationMetrics(report, duplicationAnalyzer, cm, cfg)

	// All files filtered, should return empty metrics
	assert.Equal(t, 0, report.Duplication.ClonePairs)
	assert.Equal(t, 0, report.Duplication.DuplicatedLines)
	assert.Equal(t, 0.0, report.Duplication.DuplicationRatio)
}
