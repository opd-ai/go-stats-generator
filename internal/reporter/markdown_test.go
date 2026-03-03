package reporter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

func TestMarkdownReporter_NewMarkdownReporter(t *testing.T) {
	reporter := NewMarkdownReporter()
	if reporter == nil {
		t.Fatal("NewMarkdownReporter() returned nil")
	}

	// Test type assertion to ensure it's the right type
	mdReporter, ok := reporter.(*MarkdownReporter)
	if !ok {
		t.Fatal("NewMarkdownReporter() did not return *MarkdownReporter")
	}

	// Test default values
	if !mdReporter.includeOverview {
		t.Error("Expected includeOverview to be true by default")
	}
	if !mdReporter.includeDetails {
		t.Error("Expected includeDetails to be true by default")
	}
	if mdReporter.maxItems != 50 {
		t.Errorf("Expected maxItems to be 50, got %d", mdReporter.maxItems)
	}
}

func TestMarkdownReporter_Generate_BasicReport(t *testing.T) {
	reporter := NewMarkdownReporter()

	// Create minimal test report data
	testReport := &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:     "github.com/test/repo",
			GeneratedAt:    time.Date(2025, 7, 25, 10, 30, 0, 0, time.UTC),
			AnalysisTime:   time.Millisecond * 150,
			FilesProcessed: 42,
			ToolVersion:    "v1.2.0",
			GoVersion:      "go1.21.0",
		},
		Overview: metrics.OverviewMetrics{
			TotalLinesOfCode: 1500,
			TotalFunctions:   85,
			TotalMethods:     45,
			TotalStructs:     25,
			TotalInterfaces:  8,
			TotalPackages:    12,
			TotalFiles:       42,
		},
	}

	var buf bytes.Buffer
	err := reporter.Generate(testReport, &buf)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	output := buf.String()

	// Test key sections are present
	expectedSections := []string{
		"# Go Code Analysis Report",
		"## 📊 Overview",
		"## 📈 Analysis Summary",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Expected section '%s' not found in output", section)
		}
	}

	// Test specific data is present
	expectedData := []string{
		"github.com/test/repo",
		"v1.2.0",
		"42",   // Files processed
		"1500", // Total lines of code
	}

	for _, data := range expectedData {
		if !strings.Contains(output, data) {
			t.Errorf("Expected data '%s' not found in output", data)
		}
	}
}

func TestMarkdownReporter_FormatHelpers(t *testing.T) {
	reporter := &MarkdownReporter{}

	// Test formatDuration
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"microseconds", 500 * time.Nanosecond, "0.50μs"},
		{"milliseconds", 250 * time.Millisecond, "250.00ms"},
		{"seconds", 2 * time.Second, "2.00s"},
	}

	for _, tt := range tests {
		t.Run("formatDuration_"+tt.name, func(t *testing.T) {
			result := reporter.formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("formatDuration(%v) = %s, expected %s", tt.duration, result, tt.expected)
			}
		})
	}

	// Test formatFloat
	floatTests := []struct {
		input    float64
		expected string
	}{
		{5.0, "5"},
		{5.25, "5.25"},
		{10.123456, "10.12"},
	}

	for _, tt := range floatTests {
		t.Run("formatFloat", func(t *testing.T) {
			result := reporter.formatFloat(tt.input)
			if result != tt.expected {
				t.Errorf("formatFloat(%v) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}

	// Test escapeMarkdown
	escapeTests := []struct {
		input    string
		expected string
	}{
		{"*bold*", "\\*bold\\*"},
		{"_italic_", "\\_italic\\_"},
		{"`code`", "\\`code\\`"},
		{"# Header", "\\# Header"},
		{"[link](url)", "\\[link\\]\\(url\\)"},
		{"|table|", "\\|table\\|"},
		{"normal text", "normal text"},
	}

	for _, tt := range escapeTests {
		t.Run("escapeMarkdown", func(t *testing.T) {
			result := reporter.escapeMarkdown(tt.input)
			if result != tt.expected {
				t.Errorf("escapeMarkdown(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMarkdownReporter_SimpleGenerate(t *testing.T) {
	reporter := NewMarkdownReporter()

	// Create the simplest possible report data
	testReport := &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:     "test-repo",
			GeneratedAt:    time.Now(),
			AnalysisTime:   time.Second,
			FilesProcessed: 1,
			ToolVersion:    "v1.0.0",
			GoVersion:      "go1.21.0",
		},
		Overview: metrics.OverviewMetrics{
			TotalLinesOfCode: 100,
			TotalFunctions:   10,
			TotalMethods:     5,
			TotalStructs:     3,
			TotalInterfaces:  2,
			TotalPackages:    1,
			TotalFiles:       1,
		},
		Functions:  []metrics.FunctionMetrics{},
		Structs:    []metrics.StructMetrics{},
		Interfaces: []metrics.InterfaceMetrics{},
		Packages:   []metrics.PackageMetrics{},
		Patterns:   metrics.PatternMetrics{},
		Complexity: metrics.ComplexityMetrics{
			AverageFunction: 5.0,
		},
		Documentation: metrics.DocumentationMetrics{
			Coverage: metrics.DocumentationCoverage{
				Overall: 0.8,
			},
		},
	}

	var buf bytes.Buffer
	err := reporter.Generate(testReport, &buf)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("Generated report is empty")
	}

	t.Logf("Generated report length: %d characters", len(output))
	t.Logf("First 200 characters: %s", output[:min(200, len(output))])
}

func TestMarkdownReporter_WithPlacement(t *testing.T) {
	reporter := NewMarkdownReporter()

	// Create test report with placement metrics
	testReport := &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:     "github.com/test/repo",
			GeneratedAt:    time.Date(2025, 7, 25, 10, 30, 0, 0, time.UTC),
			AnalysisTime:   time.Millisecond * 150,
			FilesProcessed: 10,
			ToolVersion:    "v1.2.0",
			GoVersion:      "go1.21.0",
		},
		Overview: metrics.OverviewMetrics{
			TotalLinesOfCode: 500,
			TotalFunctions:   20,
			TotalMethods:     10,
			TotalStructs:     5,
			TotalInterfaces:  2,
			TotalPackages:    3,
			TotalFiles:       10,
		},
		Complexity: metrics.ComplexityMetrics{
			AverageFunction: 5.0,
		},
		Documentation: metrics.DocumentationMetrics{
			Coverage: metrics.DocumentationCoverage{
				Overall: 0.8,
			},
		},
		Placement: metrics.PlacementMetrics{
			MisplacedFunctions: 2,
			MisplacedMethods:   1,
			LowCohesionFiles:   1,
			AvgFileCohesion:    0.65,
			FunctionIssues: []metrics.MisplacedFunctionIssue{
				{
					Name:              "HelperFunc",
					CurrentFile:       "util.go",
					SuggestedFile:     "helpers.go",
					CurrentAffinity:   0.25,
					SuggestedAffinity: 0.85,
					ReferencedSymbols: []string{"Helper1", "Helper2"},
					Severity:          "high",
				},
				{
					Name:              "ProcessData",
					CurrentFile:       "main.go",
					SuggestedFile:     "processor.go",
					CurrentAffinity:   0.30,
					SuggestedAffinity: 0.75,
					ReferencedSymbols: []string{"DataStruct", "ProcessConfig"},
					Severity:          "medium",
				},
			},
			MethodIssues: []metrics.MisplacedMethodIssue{
				{
					MethodName:   "String",
					ReceiverType: "User",
					CurrentFile:  "helpers.go",
					ReceiverFile: "user.go",
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
					SuggestedSplits: []string{"user_ops.go", "config_ops.go"},
					Severity:        "high",
				},
			},
		},
	}

	var buf bytes.Buffer
	err := reporter.Generate(testReport, &buf)
	if err != nil {
		t.Fatalf("Generate() failed: %v", err)
	}

	output := buf.String()

	// Test placement section is present
	expectedSections := []string{
		"## 📍 Placement Analysis",
		"### Misplaced Functions",
		"### Misplaced Methods",
		"### Low Cohesion Files",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Expected section '%s' not found in output", section)
		}
	}

	// Test specific placement data is present
	expectedData := []string{
		"HelperFunc",
		"util.go",
		"helpers.go",
		"ProcessData",
		"main.go",
		"processor.go",
		"String",
		"User",
		"user.go",
		"mixed.go",
		"0.25", // cohesion score
		"user\\_ops.go", // Escaped by markdown
		"config\\_ops.go", // Escaped by markdown
	}

	for _, data := range expectedData {
		if !strings.Contains(output, data) {
			t.Errorf("Expected data '%s' not found in output", data)
		}
	}

	// Test summary metrics
	if !strings.Contains(output, "**Misplaced Functions** | 2") {
		t.Error("Misplaced functions count not found in summary")
	}
	if !strings.Contains(output, "**Misplaced Methods** | 1") {
		t.Error("Misplaced methods count not found in summary")
	}
	if !strings.Contains(output, "**Low Cohesion Files** | 1") {
		t.Error("Low cohesion files count not found in summary")
	}
}

