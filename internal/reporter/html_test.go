package reporter

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTMLReporter_TemplateFieldsCorrect(t *testing.T) {
	// Create test data that exercises the corrected template fields
	report := &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:     "test-repo",
			GeneratedAt:    time.Now(),
			AnalysisTime:   time.Second,
			FilesProcessed: 5,
		},
		Overview: metrics.OverviewMetrics{
			TotalLinesOfCode: 1000, // Fix #1: Correct field name
			TotalFunctions:   25,
			TotalFiles:       5,
		},
		Functions: []metrics.FunctionMetrics{
			{
				Name:    "TestFunction",
				Package: "test",
				Lines: metrics.LineMetrics{ // Fix #3: Nested structure
					Code: 20,
				},
				Signature: metrics.FunctionSignature{ // Fix #5: Nested structure
					ParameterCount: 2,
					ReturnCount:    1,
				},
				Complexity: metrics.ComplexityScore{ // Fix #4: Nested structure
					Cyclomatic: 5,
				},
			},
		},
	}

	reporter := NewHTMLReporterWithConfig(&config.OutputConfig{
		IncludeOverview: true,
		IncludeDetails:  true,
	})

	var buf bytes.Buffer
	err := reporter.Generate(report, &buf)
	if err != nil {
		t.Fatalf("HTML generation failed: %v", err)
	}

	html := buf.String()

	// Verify template fixes work (data should appear in output)
	if !strings.Contains(html, "1000") { // TotalLinesOfCode
		t.Error("Fix #1 failed: TotalLinesOfCode not rendered")
	}
	if !strings.Contains(html, "20") { // Lines.Code
		t.Error("Fix #3 failed: Lines.Code not rendered")
	}
	if !strings.Contains(html, "5") { // Complexity.Cyclomatic
		t.Error("Fix #4 failed: Complexity.Cyclomatic not rendered")
	}
	if !strings.Contains(html, "2") { // Signature.ParameterCount
		t.Error("Fix #5 failed: Signature.ParameterCount not rendered")
	}
	if !strings.Contains(html, "1") { // Signature.ReturnCount
		t.Error("Fix #5 failed: Signature.ReturnCount not rendered")
	}

	// Verify Fix #2: AverageComplexity should not appear
	if strings.Contains(html, "Average Complexity") {
		t.Error("Fix #2 failed: AverageComplexity section should be removed")
	}
}

func TestHTMLReporter_DiffTemplateFieldsCorrect(t *testing.T) {
	// Create test diff data that exercises the corrected template fields
	diff := &metrics.ComplexityDiff{
		Baseline: metrics.MetricsSnapshot{
			ID: "baseline",
			Metadata: metrics.SnapshotMetadata{
				Timestamp: time.Now().Add(-time.Hour),
			},
		},
		Current: metrics.MetricsSnapshot{
			ID: "current",
			Metadata: metrics.SnapshotMetadata{
				Timestamp: time.Now(),
			},
		},
		Summary: metrics.DiffSummary{
			TotalChanges:     5,
			RegressionCount:  2,
			ImprovementCount: 1,
			OverallTrend:     metrics.TrendImproving, // Fix #6: Correct field
		},
		Regressions: []metrics.Regression{
			{
				Type:        metrics.ComplexityRegression, // Fix #7: Correct field
				Location:    "test.TestFunc",              // Fix #7: Correct field
				Description: "Complexity increased",       // Fix #7: Correct field
				OldValue:    5.0,
				NewValue:    8.0,
				Delta: metrics.Delta{ // Fix #7: Nested structure
					Percentage: 60.0,
				},
				Severity: metrics.SeverityLevelWarning,
			},
		},
		Improvements: []metrics.Improvement{
			{
				Type:        metrics.ComplexityImprovement, // Fix #7: Correct field
				Location:    "test.AnotherFunc",            // Fix #7: Correct field
				Description: "Complexity decreased",        // Fix #7: Correct field
				OldValue:    10.0,
				NewValue:    7.0,
				Delta: metrics.Delta{ // Fix #7: Nested structure
					Percentage: -30.0,
				},
				Impact: metrics.ImpactLevelMedium,
			},
		},
		Changes: []metrics.MetricChange{
			{
				Category:    "function",      // Fix #8: Correct field
				Name:        "TestFunc",      // Fix #8: Correct field
				Description: "Lines changed", // Fix #8: Correct field
				OldValue:    20.0,            // Use float64
				NewValue:    25.0,            // Use float64
				Delta: metrics.Delta{ // Fix #8: Nested structure
					Percentage: 25.0,
				},
			},
		},
	}

	reporter := NewHTMLReporterWithConfig(&config.OutputConfig{
		IncludeDetails: true,
	})

	var buf bytes.Buffer
	err := reporter.WriteDiff(&buf, diff)
	if err != nil {
		t.Fatalf("HTML diff generation failed: %v", err)
	}

	html := buf.String()

	// Verify Fix #6: OverallTrend renders correctly
	if !strings.Contains(html, "improving") {
		t.Error("Fix #6 failed: OverallTrend not rendered")
	}

	// Verify Fix #7: Regression fields render correctly
	if !strings.Contains(html, "complexity_increase") {
		t.Error("Fix #7 failed: Regression.Type not rendered")
	}
	if !strings.Contains(html, "test.TestFunc") {
		t.Error("Fix #7 failed: Regression.Location not rendered")
	}
	if !strings.Contains(html, "Complexity increased") {
		t.Error("Fix #7 failed: Regression.Description not rendered")
	}

	// Verify Fix #8: Change fields render correctly
	if !strings.Contains(html, "function") {
		t.Error("Fix #8 failed: MetricChange.Category not rendered")
	}
	if !strings.Contains(html, "TestFunc") {
		t.Error("Fix #8 failed: MetricChange.Name not rendered")
	}
	if !strings.Contains(html, "Lines changed") {
		t.Error("Fix #8 failed: MetricChange.Description not rendered")
	}
}

// TestEnhancedHTMLReporter tests the enhanced interactive HTML reporter
func TestEnhancedHTMLReporter(t *testing.T) {
	tests := []struct {
		name         string
		report       *metrics.Report
		config       *config.OutputConfig
		expectCharts bool
		expectTabs   bool
	}{
		{
			name:         "Full report with all features",
			report:       createComprehensiveTestReport(),
			config:       &config.OutputConfig{IncludeOverview: true, IncludeDetails: true, Limit: 50},
			expectCharts: true,
			expectTabs:   true,
		},
		{
			name:         "Minimal report",
			report:       createMinimalTestReport(),
			config:       &config.OutputConfig{IncludeOverview: false, IncludeDetails: false, Limit: 10},
			expectCharts: false,
			expectTabs:   true,
		},
		{
			name:         "Report with concurrency metrics",
			report:       createConcurrencyTestReport(),
			config:       &config.OutputConfig{IncludeOverview: true, IncludeDetails: true, Limit: 25},
			expectCharts: true,
			expectTabs:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := NewHTMLReporterWithConfig(tt.config)
			var output bytes.Buffer

			err := reporter.Generate(tt.report, &output)
			require.NoError(t, err, "HTML generation should not fail")

			html := output.String()

			// Test basic HTML structure
			assert.Contains(t, html, "<!DOCTYPE html>", "Should contain HTML doctype")
			assert.Contains(t, html, "<html lang=\"en\">", "Should contain HTML tag with language")
			assert.Contains(t, html, "Chart.js", "Should include Chart.js CDN")

			// Test responsive design
			assert.Contains(t, html, "viewport", "Should include viewport meta tag")
			assert.Contains(t, html, "grid-template-columns", "Should use CSS Grid for responsive design")

			// Test interactive features
			if tt.expectTabs {
				assert.Contains(t, html, "nav-tabs", "Should contain navigation tabs")
				assert.Contains(t, html, "data-tab=", "Should have tab data attributes")
				assert.Contains(t, html, "tab-content", "Should have tab content sections")
			}

			if tt.expectCharts {
				assert.Contains(t, html, "canvas id=\"complexityChart\"", "Should contain complexity chart canvas")
				assert.Contains(t, html, "canvas id=\"lengthChart\"", "Should contain length chart canvas")
				assert.Contains(t, html, "canvas id=\"packageChart\"", "Should contain package chart canvas")
				assert.Contains(t, html, "canvas id=\"qualityChart\"", "Should contain quality chart canvas")
			}

			// Test filtering and sorting
			assert.Contains(t, html, "functionFilter", "Should contain function filter input")
			assert.Contains(t, html, "complexityFilter", "Should contain complexity filter")
			assert.Contains(t, html, "data-sort=", "Should have sortable table headers")

			// Test modal functionality
			assert.Contains(t, html, "functionModal", "Should contain function details modal")
			assert.Contains(t, html, "showFunctionDetails", "Should have function to show details")

			// Test CSS variables and modern styling
			assert.Contains(t, html, ":root", "Should use CSS custom properties")
			assert.Contains(t, html, "var(--", "Should use CSS variables")

			// Test JavaScript functionality
			assert.Contains(t, html, "initializeTabs", "Should have tab initialization function")
			assert.Contains(t, html, "initializeCharts", "Should have chart initialization function")
			assert.Contains(t, html, "filterFunctions", "Should have function filtering")
			assert.Contains(t, html, "sortTable", "Should have table sorting")

			// Test accessibility features
			assert.Contains(t, html, "role=", "Should include ARIA roles") // This might not be present yet

			// Test specific content based on report data
			if len(tt.report.Functions) > 0 {
				assert.Contains(t, html, tt.report.Functions[0].Name, "Should display function names")
			}

			if len(tt.report.Packages) > 0 {
				assert.Contains(t, html, tt.report.Packages[0].Name, "Should display package names")
			}
		})
	}
}

// TestHTMLReporterResponsiveDesign tests responsive design features
func TestHTMLReporterResponsiveDesign(t *testing.T) {
	reporter := NewHTMLReporterWithConfig(&config.OutputConfig{
		IncludeOverview: true,
		IncludeDetails:  true,
		Limit:           50,
	})

	report := createComprehensiveTestReport()
	var output bytes.Buffer

	err := reporter.Generate(report, &output)
	require.NoError(t, err)

	html := output.String()

	// Test responsive breakpoints
	assert.Contains(t, html, "@media (max-width: 768px)", "Should have tablet breakpoint")
	assert.Contains(t, html, "@media (max-width: 480px)", "Should have mobile breakpoint")

	// Test grid responsiveness
	assert.Contains(t, html, "auto-fit", "Should use auto-fit for responsive grids")
	assert.Contains(t, html, "minmax(", "Should use minmax for flexible sizing")

	// Test print styles
	assert.Contains(t, html, "@media print", "Should include print styles")
}

// TestHTMLReporterChartIntegration tests Chart.js integration
func TestHTMLReporterChartIntegration(t *testing.T) {
	reporter := NewHTMLReporterWithConfig(&config.OutputConfig{
		IncludeOverview: true,
		IncludeDetails:  true,
	})

	report := createConcurrencyTestReport()
	var output bytes.Buffer

	err := reporter.Generate(report, &output)
	require.NoError(t, err)

	html := output.String()

	// Test Chart.js CDN inclusion
	assert.Contains(t, html, "chart.js@4.4.0", "Should include Chart.js v4.4.0")

	// Test chart creation functions
	chartFunctions := []string{
		"createComplexityChart",
		"createLengthChart",
		"createPackageChart",
		"createQualityChart",
		"createConcurrencyChart",
	}

	for _, fn := range chartFunctions {
		assert.Contains(t, html, fn, "Should contain chart creation function: "+fn)
	}

	// Test chart configurations
	assert.Contains(t, html, "type: 'doughnut'", "Should have doughnut chart")
	assert.Contains(t, html, "type: 'bar'", "Should have bar chart")
	assert.Contains(t, html, "type: 'scatter'", "Should have scatter chart")
	assert.Contains(t, html, "type: 'radar'", "Should have radar chart")

	// Test responsive chart options
	assert.Contains(t, html, "responsive: true", "Charts should be responsive")
	assert.Contains(t, html, "maintainAspectRatio: false", "Charts should adapt to container")
}

// TestHTMLReporterInteractivity tests interactive features
func TestHTMLReporterInteractivity(t *testing.T) {
	reporter := NewHTMLReporterWithConfig(&config.OutputConfig{
		IncludeOverview: true,
		IncludeDetails:  true,
	})

	report := createComprehensiveTestReport()
	var output bytes.Buffer

	err := reporter.Generate(report, &output)
	require.NoError(t, err)

	html := output.String()

	// Test event listeners
	assert.Contains(t, html, "addEventListener", "Should add event listeners")
	assert.Contains(t, html, "click", "Should handle click events")
	assert.Contains(t, html, "input", "Should handle input events")

	// Test filtering functionality
	assert.Contains(t, html, "filterFunctions", "Should have function filtering")
	assert.Contains(t, html, "showRow = true", "Should control row visibility")

	// Test sorting functionality
	assert.Contains(t, html, "sortTable", "Should have table sorting")
	assert.Contains(t, html, "sort-asc", "Should handle ascending sort")
	assert.Contains(t, html, "sort-desc", "Should handle descending sort")

	// Test modal functionality
	assert.Contains(t, html, "modal.style.display", "Should control modal visibility")
	assert.Contains(t, html, "getComplexityAnalysis", "Should analyze complexity")
	assert.Contains(t, html, "getRecommendations", "Should provide recommendations")
}

// createComprehensiveTestReport creates a test report with all metrics
func createComprehensiveTestReport() *metrics.Report {
	return &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:     "test-repo",
			GeneratedAt:    time.Now(),
			AnalysisTime:   2 * time.Second,
			FilesProcessed: 10,
		},
		Overview: metrics.OverviewMetrics{
			TotalFiles:       10,
			TotalLinesOfCode: 1000,
			TotalFunctions:   25,
		},
		Functions: []metrics.FunctionMetrics{
			{
				Name:       "TestFunction",
				Package:    "main",
				Lines:      metrics.LineMetrics{Code: 15, Comments: 3, Blank: 2},
				Complexity: metrics.ComplexityScore{Cyclomatic: 3},
				Signature: metrics.FunctionSignature{
					ParameterCount: 2,
					ReturnCount:    1,
				},
			},
			{
				Name:       "ComplexFunction",
				Package:    "utils",
				Lines:      metrics.LineMetrics{Code: 45, Comments: 8, Blank: 5},
				Complexity: metrics.ComplexityScore{Cyclomatic: 12},
				Signature: metrics.FunctionSignature{
					ParameterCount: 6,
					ReturnCount:    2,
				},
			},
		},
		Structs: []metrics.StructMetrics{
			{
				Name:        "TestStruct",
				Package:     "main",
				TotalFields: 5,
				Methods: []metrics.MethodInfo{
					{Name: "Method1"},
					{Name: "Method2"},
					{Name: "Method3"},
				},
				Complexity: metrics.ComplexityScore{Overall: 8.5},
			},
		},
		Packages: []metrics.PackageMetrics{
			{
				Name:          "main",
				Files:         []string{"main.go", "util.go", "helper.go"},
				Functions:     10,
				Structs:       2,
				Dependencies:  []string{"fmt", "os"},
				CohesionScore: 0.8,
				CouplingScore: 0.3,
			},
			{
				Name:          "utils",
				Files:         []string{"utils.go", "helper.go"},
				Functions:     8,
				Structs:       1,
				Dependencies:  []string{"fmt"},
				CohesionScore: 0.9,
				CouplingScore: 0.2,
			},
		},
	}
}

// createMinimalTestReport creates a minimal test report
func createMinimalTestReport() *metrics.Report {
	return &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:     "minimal-repo",
			GeneratedAt:    time.Now(),
			AnalysisTime:   100 * time.Millisecond,
			FilesProcessed: 1,
		},
		Overview: metrics.OverviewMetrics{
			TotalFiles:       1,
			TotalLinesOfCode: 50,
			TotalFunctions:   2,
		},
		Functions: []metrics.FunctionMetrics{
			{
				Name:       "SimpleFunction",
				Package:    "main",
				Lines:      metrics.LineMetrics{Code: 5, Comments: 1, Blank: 1},
				Complexity: metrics.ComplexityScore{Cyclomatic: 1},
				Signature: metrics.FunctionSignature{
					ParameterCount: 0,
					ReturnCount:    0,
				},
			},
		},
	}
}

// createConcurrencyTestReport creates a test report with concurrency metrics
func createConcurrencyTestReport() *metrics.Report {
	report := createComprehensiveTestReport()

	// Add concurrency patterns to the patterns field
	report.Patterns = metrics.PatternMetrics{
		ConcurrencyPatterns: metrics.ConcurrencyPatternMetrics{
			Goroutines: metrics.GoroutineMetrics{
				TotalCount: 2,
				Instances: []metrics.GoroutineInstance{
					{
						File:     "main.go",
						Line:     25,
						Function: "main.worker",
					},
					{
						File:     "utils.go",
						Line:     45,
						Function: "utils.processor",
					},
				},
			},
			Channels: metrics.ChannelMetrics{
				TotalCount:      2,
				BufferedCount:   1,
				UnbufferedCount: 1,
				Instances: []metrics.ChannelInstance{
					{
						File:       "main.go",
						Line:       15,
						Function:   "main",
						Type:       "chan Job",
						IsBuffered: true,
						BufferSize: 10,
					},
					{
						File:       "main.go",
						Line:       16,
						Function:   "main",
						Type:       "chan Result",
						IsBuffered: false,
						BufferSize: 0,
					},
				},
			},
			WorkerPools: []metrics.PatternInstance{
				{
					Name:            "JobProcessor",
					File:            "main.go",
					Line:            20,
					ConfidenceScore: 0.95,
				},
			},
			Pipelines: []metrics.PatternInstance{
				{
					Name:            "DataPipeline",
					File:            "utils.go",
					Line:            40,
					ConfidenceScore: 0.88,
				},
			},
			Semaphores: []metrics.PatternInstance{
				{
					Name:            "ConnectionPool",
					File:            "db.go",
					Line:            30,
					ConfidenceScore: 0.92,
				},
			},
			FanOut: []metrics.PatternInstance{
				{
					Name:            "RequestFanOut",
					File:            "server.go",
					Line:            50,
					ConfidenceScore: 0.85,
				},
			},
			FanIn: []metrics.PatternInstance{
				{
					Name:            "ResponseFanIn",
					File:            "server.go",
					Line:            75,
					ConfidenceScore: 0.90,
				},
			},
			SyncPrims: metrics.SyncPrimitives{
				Mutexes: []metrics.SyncPrimitiveInstance{
					{
						File:     "cache.go",
						Line:     15,
						Function: "Cache.Get",
						Type:     "sync.Mutex",
						Variable: "mu",
					},
				},
				WaitGroups: []metrics.SyncPrimitiveInstance{
					{
						File:     "main.go",
						Line:     35,
						Function: "main.main",
						Type:     "sync.WaitGroup",
						Variable: "wg",
					},
				},
			},
		},
	}

	return report
}

// TestHTMLReporterPerformance tests performance with large datasets
func TestHTMLReporterPerformance(t *testing.T) {
	// Create a large report to test performance
	report := createLargeTestReport(1000) // 1000 functions

	reporter := NewHTMLReporterWithConfig(&config.OutputConfig{
		IncludeOverview: true,
		IncludeDetails:  true,
		Limit:           1000,
	})

	start := time.Now()
	var output bytes.Buffer

	err := reporter.Generate(report, &output)
	duration := time.Since(start)

	require.NoError(t, err, "Large report generation should not fail")
	assert.Less(t, duration, 5*time.Second, "Report generation should complete within 5 seconds")

	// Verify the output contains expected content
	html := output.String()
	assert.Greater(t, len(html), 50000, "Large report should generate substantial HTML")
	assert.Contains(t, html, "TestFunction999", "Should contain last function")
}

// createLargeTestReport creates a test report with many functions for performance testing
func createLargeTestReport(functionCount int) *metrics.Report {
	functions := make([]metrics.FunctionMetrics, functionCount)
	for i := 0; i < functionCount; i++ {
		functions[i] = metrics.FunctionMetrics{
			Name:       fmt.Sprintf("TestFunction%d", i),
			Package:    "test",
			Lines:      metrics.LineMetrics{Code: 10 + i%40, Comments: 2, Blank: 1},
			Complexity: metrics.ComplexityScore{Cyclomatic: 1 + i%15},
			Signature: metrics.FunctionSignature{
				ParameterCount: i % 8,
				ReturnCount:    (i % 3) + 1,
			},
		}
	}

	return &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:     "large-test-repo",
			GeneratedAt:    time.Now(),
			AnalysisTime:   10 * time.Second,
			FilesProcessed: 100,
		},
		Overview: metrics.OverviewMetrics{
			TotalFiles:       100,
			TotalLinesOfCode: functionCount * 15,
			TotalFunctions:   functionCount,
		},
		Functions: functions,
	}
}

// TestHTMLReporter_WithPlacement tests HTML output with placement metrics
func TestHTMLReporter_WithPlacement(t *testing.T) {
	report := &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:     "test-repo",
			GeneratedAt:    time.Now(),
			AnalysisTime:   time.Second,
			FilesProcessed: 3,
		},
		Overview: metrics.OverviewMetrics{
			TotalFiles:       3,
			TotalLinesOfCode: 500,
			TotalFunctions:   10,
		},
		Placement: metrics.PlacementMetrics{
			MisplacedFunctions: 2,
			MisplacedMethods:   1,
			LowCohesionFiles:   1,
			AvgFileCohesion:    0.65,
			FunctionIssues: []metrics.MisplacedFunctionIssue{
				{
					Name:              "HelperFunc",
					CurrentFile:       "main.go",
					SuggestedFile:     "helper.go",
					CurrentAffinity:   0.2,
					SuggestedAffinity: 0.8,
					ReferencedSymbols: []string{"HelperStruct", "HelperMethod"},
					Severity:          "high",
				},
				{
					Name:              "UtilFunc",
					CurrentFile:       "controller.go",
					SuggestedFile:     "utils.go",
					CurrentAffinity:   0.3,
					SuggestedAffinity: 0.7,
					ReferencedSymbols: []string{"StringUtils"},
					Severity:          "medium",
				},
			},
			MethodIssues: []metrics.MisplacedMethodIssue{
				{
					MethodName:   "Process",
					ReceiverType: "Handler",
					CurrentFile:  "utils.go",
					ReceiverFile: "handler.go",
					Distance:     "same_package",
					Severity:     "medium",
				},
			},
			CohesionIssues: []metrics.FileCohesionIssue{
				{
					File:            "mixed.go",
					CohesionScore:   0.25,
					IntraFileRefs:   5,
					TotalRefs:       20,
					SuggestedSplits: []string{"handlers.go", "models.go"},
					Severity:        "high",
				},
			},
		},
	}

	reporter := NewHTMLReporterWithConfig(&config.OutputConfig{
		IncludeOverview: true,
		IncludeDetails:  true,
		Limit:           10,
	})

	var output bytes.Buffer
	err := reporter.Generate(report, &output)

	require.NoError(t, err, "HTML generation should not fail")

	html := output.String()

	// Verify placement tab exists
	require.Contains(t, html, `data-tab="placement"`, "HTML should include placement tab button")
	require.Contains(t, html, `id="placement"`, "HTML should include placement section")

	// Verify summary metrics
	require.Contains(t, html, "Misplaced Functions", "Should show misplaced functions label")
	require.Contains(t, html, "Misplaced Methods", "Should show misplaced methods label")
	require.Contains(t, html, "Low Cohesion Files", "Should show low cohesion files label")

	// Verify function issues table
	require.Contains(t, html, "HelperFunc", "Should include misplaced function name")
	require.Contains(t, html, "helper.go", "Should include suggested file")

	// Verify method issues table
	require.Contains(t, html, "Process", "Should include method name")
	require.Contains(t, html, "Handler", "Should include receiver type")

	// Verify cohesion issues table
	require.Contains(t, html, "mixed.go", "Should include low cohesion file")
	require.Contains(t, html, "handlers.go", "Should include suggested split")
}
