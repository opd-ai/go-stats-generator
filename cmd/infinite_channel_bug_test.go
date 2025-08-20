package cmd

import (
	"context"
	"go/parser"
	"go/token"
	"testing"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/analyzer"
	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/opd-ai/go-stats-generator/internal/scanner"
)

func TestProcessAnalysisResultsWithContextCancellation(t *testing.T) {
	// Regression test: Ensure processAnalysisResults respects context cancellation
	// Previously, the function would hang indefinitely if the results channel wasn't closed

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
	// Note: We intentionally don't close this channel to test context cancellation

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

	// Test that context cancellation works properly
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, _, err = processAnalysisResults(ctx, unclosedResults, analyzers, report, cfg)
	duration := time.Since(start)

	// Should return with context cancelled error
	if err == nil {
		t.Error("Expected context cancellation error, but got nil")
	}

	// Should complete quickly due to context timeout, not hang indefinitely
	if duration > 100*time.Millisecond {
		t.Errorf("Function took too long (%v), suggesting it may still be hanging", duration)
	}

	// Verify it's a context cancellation error
	if !isContextError(err) {
		t.Errorf("Expected context cancellation error, got: %v", err)
	}
}

// isContextError checks if an error is a context cancellation error
func isContextError(err error) bool {
	return err != nil && (err.Error() == "analysis cancelled: context deadline exceeded" ||
		err.Error() == "analysis cancelled: context canceled")
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
		ctx := context.Background()
		result, _, err = processAnalysisResults(ctx, properResults, analyzers, report, cfg)
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
