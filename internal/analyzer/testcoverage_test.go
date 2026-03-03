package analyzer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTestCoverageAnalyzer(t *testing.T) {
	analyzer := NewTestCoverageAnalyzer()
	assert.NotNil(t, analyzer)
	assert.NotNil(t, analyzer.coverageData)
}

func TestLoadCoverageProfile(t *testing.T) {
	dir := t.TempDir()
	coverageFile := filepath.Join(dir, "coverage.out")

	coverageData := `mode: set
github.com/example/pkg/file.go:10.2,12.3 2 1
github.com/example/pkg/file.go:15.5,20.10 3 0
github.com/example/pkg/other.go:5.1,7.2 2 5
`
	err := os.WriteFile(coverageFile, []byte(coverageData), 0o644)
	require.NoError(t, err)

	analyzer := NewTestCoverageAnalyzer()
	err = analyzer.LoadCoverageProfile(coverageFile)
	require.NoError(t, err)

	assert.NotEmpty(t, analyzer.coverageData)
	assert.Contains(t, analyzer.coverageData, "github.com/example/pkg/file.go")
	assert.Contains(t, analyzer.coverageData, "github.com/example/pkg/other.go")
}

func TestLoadCoverageProfile_InvalidFile(t *testing.T) {
	analyzer := NewTestCoverageAnalyzer()
	err := analyzer.LoadCoverageProfile("nonexistent.out")
	assert.Error(t, err)
}

func TestParseCoverageLine(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantErr bool
		checkFn func(*TestCoverageAnalyzer) bool
	}{
		{
			name:    "valid coverage line",
			line:    "file.go:10.2,15.5 3 1",
			wantErr: false,
			checkFn: func(a *TestCoverageAnalyzer) bool {
				return len(a.coverageData["file.go"]) > 0
			},
		},
		{
			name:    "mode line skipped",
			line:    "mode: set",
			wantErr: false,
			checkFn: func(a *TestCoverageAnalyzer) bool {
				return len(a.coverageData) == 0
			},
		},
		{
			name:    "empty line",
			line:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewTestCoverageAnalyzer()
			err := analyzer.parseCoverageLine(tt.line)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.checkFn != nil {
				assert.True(t, tt.checkFn(analyzer))
			}
		})
	}
}

func TestCalculateFunctionCoverage(t *testing.T) {
	analyzer := NewTestCoverageAnalyzer()
	analyzer.coverageData = map[string]map[int]int{
		"test.go": {
			10: 1, 11: 1, 12: 1, 13: 0, 14: 0,
		},
	}

	fn := metrics.FunctionMetrics{
		File: "test.go",
		Line: 10,
		Lines: metrics.LineMetrics{
			Total: 5,
		},
	}

	coverage := analyzer.calculateFunctionCoverage(fn)
	assert.InDelta(t, 0.5, coverage, 0.01)
}

func TestIsHighRisk(t *testing.T) {
	analyzer := NewTestCoverageAnalyzer()

	tests := []struct {
		name       string
		complexity int
		coverage   float64
		want       bool
	}{
		{"high complexity, low coverage", 10, 0.3, true},
		{"high complexity, high coverage", 10, 0.9, false},
		{"low complexity, low coverage", 3, 0.3, false},
		{"medium complexity, medium coverage", 6, 0.4, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := metrics.FunctionMetrics{
				Complexity: metrics.ComplexityScore{
					Cyclomatic: tt.complexity,
				},
			}
			result := analyzer.isHighRisk(fn, tt.coverage)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestCalculateRiskScore(t *testing.T) {
	analyzer := NewTestCoverageAnalyzer()

	tests := []struct {
		name       string
		complexity int
		lines      int
		coverage   float64
		minScore   float64
	}{
		{"simple function", 5, 20, 0.5, 2.0},
		{"complex function", 10, 30, 0.3, 6.0},
		{"long function", 8, 60, 0.4, 7.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := metrics.FunctionMetrics{
				Complexity: metrics.ComplexityScore{
					Cyclomatic: tt.complexity,
				},
				Lines: metrics.LineMetrics{
					Total: tt.lines,
				},
			}
			score := analyzer.calculateRiskScore(fn, tt.coverage)
			assert.GreaterOrEqual(t, score, tt.minScore)
		})
	}
}

func TestAnalyzeCorrelation(t *testing.T) {
	analyzer := NewTestCoverageAnalyzer()
	analyzer.coverageData = map[string]map[int]int{
		"test.go": {
			10: 1, 11: 1, 12: 0, 13: 0,
			20: 0, 21: 0, 22: 0,
		},
	}

	functions := []metrics.FunctionMetrics{
		{
			Name:  "HighRisk",
			File:  "test.go",
			Line:  10,
			Lines: metrics.LineMetrics{Total: 4},
			Complexity: metrics.ComplexityScore{
				Cyclomatic: 8,
				Overall:    10.0,
			},
			IsExported: true,
		},
		{
			Name:  "LowRisk",
			File:  "test.go",
			Line:  20,
			Lines: metrics.LineMetrics{Total: 3},
			Complexity: metrics.ComplexityScore{
				Cyclomatic: 2,
				Overall:    3.0,
			},
			IsExported: true,
		},
	}

	result := analyzer.AnalyzeCorrelation(functions)

	assert.Greater(t, result.FunctionCoverageRate, 0.0)
	assert.Greater(t, result.ComplexityCoverageRate, 0.0)
	assert.NotEmpty(t, result.HighRiskFunctions)
	assert.NotEmpty(t, result.CoverageGaps)
}

func TestAnalyzeTestQuality(t *testing.T) {
	dir := t.TempDir()

	testFile := filepath.Join(dir, "example_test.go")
	testContent := `package example

import "testing"

func TestExample(t *testing.T) {
	result := Add(1, 2)
	if result != 3 {
		t.Error("expected 3")
	}
}

func TestSubtests(t *testing.T) {
	t.Run("case1", func(t *testing.T) {
		assert.Equal(t, 1, 1)
	})
}
`
	err := os.WriteFile(testFile, []byte(testContent), 0o644)
	require.NoError(t, err)

	result, err := AnalyzeTestQuality(dir)
	require.NoError(t, err)

	assert.Equal(t, 2, result.TotalTests)
	assert.Len(t, result.TestFiles, 1)
	assert.Greater(t, result.AvgAssertionsPerTest, 0.0)
}

func TestGapSeverity(t *testing.T) {
	analyzer := NewTestCoverageAnalyzer()

	tests := []struct {
		name       string
		complexity int
		coverage   float64
		want       string
	}{
		{"critical gap", 10, 0.2, "critical"},
		{"high gap", 8, 0.4, "high"},
		{"medium gap", 3, 0.6, "medium"},
		{"low gap", 2, 0.75, "low"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := metrics.FunctionMetrics{
				Complexity: metrics.ComplexityScore{
					Cyclomatic: tt.complexity,
				},
			}
			result := analyzer.gapSeverity(fn, tt.coverage)
			assert.Equal(t, tt.want, result)
		})
	}
}
