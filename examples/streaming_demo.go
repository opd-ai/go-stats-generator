package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/opd-ai/go-stats-generator/internal/reporter"
)

func main() {
	// Create a streaming reporter
	rep := reporter.NewJSONReporter()

	// Create test metadata
	metadata := &metrics.ReportMetadata{
		Repository:     "test-repo",
		GeneratedAt:    time.Now(),
		AnalysisTime:   5 * time.Second,
		FilesProcessed: 10,
		ToolVersion:    "1.0.0",
		GoVersion:      "1.24.0",
	}

	// Write using streaming API
	var buf bytes.Buffer
	
	if err := rep.BeginReport(&buf, metadata); err != nil {
		panic(err)
	}

	overview := metrics.OverviewMetrics{
		TotalLinesOfCode: 1000,
		TotalFunctions:   50,
		TotalMethods:     25,
	}
	if err := rep.WriteSection(&buf, "overview", overview); err != nil {
		panic(err)
	}

	functions := []metrics.FunctionMetrics{
		{
			Name: "TestFunc1",
			File: "test.go",
			Lines: metrics.LineMetrics{Total: 10, Code: 8, Comments: 1, Blank: 1},
		},
		{
			Name: "TestFunc2",
			File: "test.go",
			Lines: metrics.LineMetrics{Total: 20, Code: 18, Comments: 1, Blank: 1},
		},
	}
	if err := rep.WriteSection(&buf, "functions", functions); err != nil {
		panic(err)
	}

	if err := rep.EndReport(&buf); err != nil {
		panic(err)
	}

	// Verify output is valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		panic(fmt.Sprintf("Invalid JSON: %v", err))
	}

	fmt.Println("✓ Streaming reporter produces valid JSON")
	fmt.Printf("✓ Sections: metadata, %d additional sections\n", len(result)-1)
	fmt.Println("✓ Output sample:")
	fmt.Println(buf.String()[:min(500, len(buf.String()))])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
