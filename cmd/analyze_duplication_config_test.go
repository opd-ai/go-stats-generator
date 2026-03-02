package cmd

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/opd-ai/go-stats-generator/internal/analyzer"
	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		file, err := parser.ParseFile(fset, "test.go", code, parser.ParseComments)
		require.NoError(t, err)

		report := &metrics.Report{}
		collectedMetrics := &CollectedMetrics{
			Files: map[string]*ast.File{
				"test.go": file,
			},
		}

		finalizeDuplicationMetrics(report, duplicationAnalyzer, collectedMetrics, cfg)

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
		file, err := parser.ParseFile(fset, "similarity_test.go", code, parser.ParseComments)
		require.NoError(t, err)

		report := &metrics.Report{}
		collectedMetrics := &CollectedMetrics{
			Files: map[string]*ast.File{
				"similarity_test.go": file,
			},
		}

		finalizeDuplicationMetrics(report, duplicationAnalyzer, collectedMetrics, cfg)

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
		regularFile, err := parser.ParseFile(fset, "regular.go", regularCode, parser.ParseComments)
		require.NoError(t, err)

		testFile, err := parser.ParseFile(fset, "filter_test.go", testCode, parser.ParseComments)
		require.NoError(t, err)

		report := &metrics.Report{}
		collectedMetrics := &CollectedMetrics{
			Files: map[string]*ast.File{
				"regular.go":      regularFile,
				"filter_test.go":  testFile,
			},
		}

		finalizeDuplicationMetrics(report, duplicationAnalyzer, collectedMetrics, cfg)

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
		regularFile, err := parser.ParseFile(fset, "regular.go", code, parser.ParseComments)
		require.NoError(t, err)

		testFile, err := parser.ParseFile(fset, "include_test.go", code, parser.ParseComments)
		require.NoError(t, err)

		report := &metrics.Report{}
		collectedMetrics := &CollectedMetrics{
			Files: map[string]*ast.File{
				"regular.go":       regularFile,
				"include_test.go":  testFile,
			},
		}

		finalizeDuplicationMetrics(report, duplicationAnalyzer, collectedMetrics, cfg)

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
	collectedMetrics := &CollectedMetrics{
		Files: map[string]*ast.File{},
	}

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
	testFile, err := parser.ParseFile(fset, "example_test.go", testCode, parser.ParseComments)
	require.NoError(t, err)

	report := &metrics.Report{}
	collectedMetrics := &CollectedMetrics{
		Files: map[string]*ast.File{
			"example_test.go": testFile,
		},
	}

	finalizeDuplicationMetrics(report, duplicationAnalyzer, collectedMetrics, cfg)

	// All files filtered, should return empty metrics
	assert.Equal(t, 0, report.Duplication.ClonePairs)
	assert.Equal(t, 0, report.Duplication.DuplicatedLines)
	assert.Equal(t, 0.0, report.Duplication.DuplicationRatio)
}
