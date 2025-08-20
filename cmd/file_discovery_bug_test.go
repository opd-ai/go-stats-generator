package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnalyzeCommandFileVsDirectoryHandling(t *testing.T) {
	// Regression test: Ensure analyze command properly handles both files and directories
	// Previously, the command might not properly distinguish between files and directories

	// Create a temporary Go file for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	testContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}

func add(a, b int) int {
	return a + b
}
`

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test that analyze command can handle a single file
	err = runAnalyze(nil, []string{testFile})
	if err != nil {
		t.Errorf("Analyze command failed on single file: %v", err)
	}

	// Test that analyze command can handle a directory
	err = runAnalyze(nil, []string{tmpDir})
	if err != nil {
		t.Errorf("Analyze command failed on directory: %v", err)
	}
}

func TestAnalyzeCommandNonExistentPath(t *testing.T) {
	// Test that analyze command properly handles non-existent paths
	err := runAnalyze(nil, []string{"/nonexistent/path"})

	if err == nil {
		t.Error("Expected error for non-existent path, but got nil")
	}

	// Should get a clear error message about the path not existing
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Expected error to contain 'does not exist', but got: %v", err)
	}
}
