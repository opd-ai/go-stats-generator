package analyzer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// TestCoverageAnalyzer correlates code metrics with test coverage
type TestCoverageAnalyzer struct {
	coverageData map[string]map[int]int // file -> line -> hit count
}

// NewTestCoverageAnalyzer creates coverage analyzer
func NewTestCoverageAnalyzer() *TestCoverageAnalyzer {
	return &TestCoverageAnalyzer{
		coverageData: make(map[string]map[int]int),
	}
}

// LoadCoverageProfile parses Go coverage profile
func (a *TestCoverageAnalyzer) LoadCoverageProfile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open coverage: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "mode:") {
			continue
		}
		if err := a.parseCoverageLine(line); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func (a *TestCoverageAnalyzer) parseCoverageLine(line string) error {
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return nil
	}

	fileAndRange := strings.Split(parts[0], ":")
	if len(fileAndRange) != 2 {
		return nil
	}

	file := fileAndRange[0]
	ranges := strings.Split(fileAndRange[1], ",")
	if len(ranges) != 2 {
		return nil
	}

	startLine, err := parseLineNum(ranges[0])
	if err != nil {
		return err
	}
	endLine, err := parseLineNum(ranges[1])
	if err != nil {
		return err
	}

	count, err := strconv.Atoi(parts[2])
	if err != nil {
		return err
	}

	if a.coverageData[file] == nil {
		a.coverageData[file] = make(map[int]int)
	}

	for i := startLine; i <= endLine; i++ {
		a.coverageData[file][i] += count
	}

	return nil
}

func parseLineNum(s string) (int, error) {
	parts := strings.Split(s, ".")
	return strconv.Atoi(parts[0])
}

// AnalyzeCorrelation generates coverage metrics
func (a *TestCoverageAnalyzer) AnalyzeCorrelation(functions []metrics.FunctionMetrics) metrics.TestCoverageMetrics {
	result := metrics.TestCoverageMetrics{
		HighRiskFunctions: []metrics.HighRiskFunction{},
		CoverageGaps:      []metrics.CoverageGap{},
	}

	var totalFunctions, coveredFunctions int
	var totalComplexity, coveredComplexity float64

	for _, fn := range functions {
		coverage := a.calculateFunctionCoverage(fn)
		totalFunctions++
		totalComplexity += fn.Complexity.Overall

		if coverage > 0 {
			coveredFunctions++
			coveredComplexity += fn.Complexity.Overall
		}

		if a.isHighRisk(fn, coverage) {
			result.HighRiskFunctions = append(result.HighRiskFunctions, metrics.HighRiskFunction{
				Name:       fn.Name,
				File:       fn.File,
				Line:       fn.Line,
				Complexity: fn.Complexity.Cyclomatic,
				Coverage:   coverage,
				RiskScore:  a.calculateRiskScore(fn, coverage),
			})
		}

		if a.isCoverageGap(fn, coverage) {
			result.CoverageGaps = append(result.CoverageGaps, metrics.CoverageGap{
				Name:        fn.Name,
				File:        fn.File,
				Line:        fn.Line,
				Complexity:  fn.Complexity.Cyclomatic,
				Coverage:    coverage,
				GapSeverity: a.gapSeverity(fn, coverage),
			})
		}
	}

	if totalFunctions > 0 {
		result.FunctionCoverageRate = float64(coveredFunctions) / float64(totalFunctions)
	}
	if totalComplexity > 0 {
		result.ComplexityCoverageRate = coveredComplexity / totalComplexity
	}

	sort.Slice(result.HighRiskFunctions, func(i, j int) bool {
		return result.HighRiskFunctions[i].RiskScore > result.HighRiskFunctions[j].RiskScore
	})

	return result
}

func (a *TestCoverageAnalyzer) calculateFunctionCoverage(fn metrics.FunctionMetrics) float64 {
	file := fn.File
	startLine := fn.Line
	endLine := fn.Line + fn.Lines.Total

	if a.coverageData[file] == nil {
		return 0.0
	}

	var covered, total int
	for line := startLine; line <= endLine; line++ {
		total++
		if a.coverageData[file][line] > 0 {
			covered++
		}
	}

	if total == 0 {
		return 0.0
	}
	return float64(covered) / float64(total)
}

func (a *TestCoverageAnalyzer) isHighRisk(fn metrics.FunctionMetrics, coverage float64) bool {
	return (fn.Complexity.Cyclomatic > 5 && coverage < 0.5) ||
		(fn.Complexity.Cyclomatic > 10 && coverage < 0.8)
}

func (a *TestCoverageAnalyzer) isCoverageGap(fn metrics.FunctionMetrics, coverage float64) bool {
	return coverage < 0.7 && fn.IsExported
}

func (a *TestCoverageAnalyzer) calculateRiskScore(fn metrics.FunctionMetrics, coverage float64) float64 {
	score := float64(fn.Complexity.Cyclomatic) * (1.0 - coverage)
	if fn.Lines.Total > 50 {
		score *= 1.5
	}
	return score
}

func (a *TestCoverageAnalyzer) gapSeverity(fn metrics.FunctionMetrics, coverage float64) string {
	if coverage < 0.3 && fn.Complexity.Cyclomatic > 5 {
		return "critical"
	}
	if coverage < 0.5 || fn.Complexity.Cyclomatic > 8 {
		return "high"
	}
	if coverage < 0.7 {
		return "medium"
	}
	return "low"
}

// AnalyzeTestQuality assesses test suite quality
func AnalyzeTestQuality(repoPath string) (metrics.TestQualityMetrics, error) {
	result := metrics.TestQualityMetrics{
		TestFiles: []metrics.TestFileInfo{},
	}

	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}

		relPath, _ := filepath.Rel(repoPath, path)
		testInfo := analyzeTestFile(path, relPath)
		result.TestFiles = append(result.TestFiles, testInfo)
		result.TotalTests += testInfo.TestCount

		return nil
	})
	if err != nil {
		return result, err
	}

	if len(result.TestFiles) > 0 {
		var totalAssertion float64
		for _, tf := range result.TestFiles {
			totalAssertion += tf.AssertionRatio
		}
		result.AvgAssertionsPerTest = totalAssertion / float64(len(result.TestFiles))
	}

	return result, nil
}

func analyzeTestFile(path, relPath string) metrics.TestFileInfo {
	info := metrics.TestFileInfo{
		File: relPath,
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return info
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.Contains(line, "func Test") {
			info.TestCount++
		}
		if strings.Contains(line, "t.Run(") {
			info.SubtestCount++
		}
		if containsAssertion(line) {
			info.AssertionCount++
		}
	}

	if info.TestCount > 0 {
		info.AssertionRatio = float64(info.AssertionCount) / float64(info.TestCount)
	}

	return info
}

func containsAssertion(line string) bool {
	assertions := []string{
		"assert.", "require.", "t.Error", "t.Fatal",
		"!= nil", "== nil", "if err",
	}
	for _, a := range assertions {
		if strings.Contains(line, a) {
			return true
		}
	}
	return false
}
