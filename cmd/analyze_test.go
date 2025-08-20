package cmd

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/config"
)

func TestIsGoSourceFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected bool
	}{
		{
			name:     "Go source file",
			filePath: "/path/to/file.go",
			expected: true,
		},
		{
			name:     "Go test file",
			filePath: "/path/to/file_test.go",
			expected: true,
		},
		{
			name:     "Non-Go file",
			filePath: "/path/to/file.txt",
			expected: false,
		},
		{
			name:     "Markdown file",
			filePath: "/path/to/README.md",
			expected: false,
		},
		{
			name:     "JSON file",
			filePath: "/path/to/config.json",
			expected: false,
		},
		{
			name:     "No extension",
			filePath: "/path/to/Makefile",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGoSourceFile(tt.filePath)
			if result != tt.expected {
				t.Errorf("isGoSourceFile(%q) = %v, want %v", tt.filePath, result, tt.expected)
			}
		})
	}
}

func TestRunFileAnalysis(t *testing.T) {
	// Create a temporary Go file for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.go")

	testContent := `package test

import "fmt"

// TestFunction demonstrates a simple function
func TestFunction(name string) string {
	if name == "" {
		return "Hello, World!"
	}
	return fmt.Sprintf("Hello, %s!", name)
}

// AnotherFunction with more complexity
func AnotherFunction(x, y int) int {
	if x > y {
		return x
	} else if x < y {
		return y
	}
	return 0
}
`

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test single file analysis
	cfg := config.DefaultConfig()
	cfg.Output.Verbose = false // Avoid stderr output in tests

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	report, err := runFileAnalysis(ctx, testFile, cfg)
	if err != nil {
		t.Fatalf("runFileAnalysis failed: %v", err)
	}

	// Verify basic report structure
	if report == nil {
		t.Fatal("Report is nil")
	}

	if report.Metadata.FilesProcessed != 1 {
		t.Errorf("Expected 1 file processed, got %d", report.Metadata.FilesProcessed)
	}

	if len(report.Functions) == 0 {
		t.Error("Expected at least one function in the report")
	}

	if report.Overview.TotalFiles != 1 {
		t.Errorf("Expected 1 total file, got %d", report.Overview.TotalFiles)
	}

	// Verify that functions were found
	if report.Overview.TotalFunctions == 0 {
		t.Error("Expected at least one function to be found")
	}

	// Check that package information is correct
	if len(report.Packages) == 0 {
		t.Error("Expected at least one package in the report")
	} else if report.Packages[0].Name != "test" {
		t.Errorf("Expected package name 'test', got '%s'", report.Packages[0].Name)
	}
}

func TestRunFileAnalysisWithNonGoFile(t *testing.T) {
	// Create a temporary non-Go file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	err := os.WriteFile(testFile, []byte("This is not a Go file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg := config.DefaultConfig()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = runFileAnalysis(ctx, testFile, cfg)
	if err == nil {
		t.Error("Expected error for non-Go file, but got none")
	}

	if !strings.Contains(err.Error(), "is not a Go source file") {
		t.Errorf("Expected error message about non-Go file, got: %v", err)
	}
}

func TestRunFileAnalysisWithNonExistentFile(t *testing.T) {
	cfg := config.DefaultConfig()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	nonExistentFile := "/path/that/does/not/exist.go"
	_, err := runFileAnalysis(ctx, nonExistentFile, cfg)
	if err == nil {
		t.Error("Expected error for non-existent file, but got none")
	}
}

func TestRunAnalyzeCommandWithFile(t *testing.T) {
	// Create a temporary Go file for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "simple.go")

	testContent := `package simple

func SimpleFunction() {
	println("Hello from simple function")
}
`

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the analyze command
	err = runAnalyze(analyzeCmd, []string{testFile})

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout
	output, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("runAnalyze failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "=== GO SOURCE CODE STATISTICS REPORT ===") {
		t.Error("Expected report header in output")
	}

	if !strings.Contains(outputStr, "Files Processed: 1") {
		t.Error("Expected 1 file processed in output")
	}
}

func TestRunAnalyzeCommandWithDirectory(t *testing.T) {
	// Use the existing testdata directory
	testDir := "../testdata/simple"

	// Check if testdata exists
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Skip("Skipping test: testdata directory not found")
	}

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the analyze command
	err := runAnalyze(analyzeCmd, []string{testDir})

	// Restore stdout and read output
	w.Close()
	os.Stdout = oldStdout
	output, _ := io.ReadAll(r)

	if err != nil {
		t.Fatalf("runAnalyze failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "=== GO SOURCE CODE STATISTICS REPORT ===") {
		t.Error("Expected report header in output")
	}

	// Should process multiple files in directory mode
	if !strings.Contains(outputStr, "Files Processed:") {
		t.Error("Expected files processed information in output")
	}
}
