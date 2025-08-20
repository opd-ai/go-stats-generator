package cmd

import (
	"go/parser"
	"go/token"
	"testing"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/analyzer"
	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/opd-ai/go-stats-generator/internal/scanner"
)

func TestProcessAnalysisResultsHangBug(t *testing.T) {
	// Create a test scenario where the channel reading could hang
	// This demonstrates the lack of timeout/context handling

	// Create a minimal valid AST file for testing
	src := `package main
func test() {
	println("hello")
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse test file: %v", err)
	}

	// Create a channel that will send one result but never close
	unclosedResults := make(chan scanner.Result, 1)
	unclosedResults <- scanner.Result{
		FileInfo: scanner.FileInfo{Path: "test.go", Package: "main"},
		File:     file,
		Error:    nil,
	}
	// Note: We intentionally don't close this channel to simulate the bug

	// Create test analyzers and report
	analyzers := &AnalyzerSet{
		Function:    analyzer.NewFunctionAnalyzer(fset),
		Struct:      analyzer.NewStructAnalyzer(fset),
		Interface:   analyzer.NewInterfaceAnalyzer(fset),
		Package:     analyzer.NewPackageAnalyzer(fset),
		Concurrency: analyzer.NewConcurrencyAnalyzer(fset),
	}

	report := &metrics.Report{}
	cfg := &config.Config{}

	// Test with a timeout to detect infinite hanging
	done := make(chan bool, 1)

	go func() {
		// This should hang indefinitely after processing the first result
		// because the channel is never closed
		_, _, _ = processAnalysisResults(unclosedResults, analyzers, report, cfg)
		done <- true
	}()

	// Wait for either completion or timeout
	select {
	case <-done:
		// If we get here quickly, the function completed (which means it didn't hang)
		t.Log("processAnalysisResults completed - either the bug is fixed or channel was somehow closed")
	case <-time.After(200 * time.Millisecond):
		// Expected: timeout because processAnalysisResults hangs waiting for more results
		t.Log("processAnalysisResults hangs waiting for channel closure (demonstrates the bug)")
		// This confirms the bug exists - the function waits indefinitely for channel closure
	}
}

func TestProcessAnalysisResultsProperChannelClosure(t *testing.T) {
	// Test the normal case where the channel is properly closed

	// Create a minimal valid AST file for testing
	src := `package main
func test() {
	println("hello")
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse test file: %v", err)
	}

	// Create a properly closed channel
	properResults := make(chan scanner.Result, 1)
	properResults <- scanner.Result{
		FileInfo: scanner.FileInfo{Path: "test.go", Package: "main"},
		File:     file,
		Error:    nil,
	}
	close(properResults) // This is the key difference

	// Create test analyzers and report
	analyzers := &AnalyzerSet{
		Function:    analyzer.NewFunctionAnalyzer(fset),
		Struct:      analyzer.NewStructAnalyzer(fset),
		Interface:   analyzer.NewInterfaceAnalyzer(fset),
		Package:     analyzer.NewPackageAnalyzer(fset),
		Concurrency: analyzer.NewConcurrencyAnalyzer(fset),
	}

	report := &metrics.Report{}
	cfg := &config.Config{}

	// Test with a timeout - this should complete quickly
	done := make(chan bool, 1)
	var result *CollectedMetrics

	go func() {
		var err error
		result, _, err = processAnalysisResults(properResults, analyzers, report, cfg)
		if err != nil {
			t.Errorf("processAnalysisResults failed: %v", err)
		}
		done <- true
	}()

	// Wait for completion or timeout
	select {
	case <-done:
		// This should happen quickly with a properly closed channel
		if result == nil {
			t.Error("Expected result but got nil")
		}
		t.Log("processAnalysisResults completed successfully with closed channel")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("processAnalysisResults took too long even with closed channel")
	}
}
