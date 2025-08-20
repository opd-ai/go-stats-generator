package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeCommandFileVsDirectoryHandling(t *testing.T) {
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
	expectedSubstring := "does not exist"
	if !containsSubstring(err.Error(), expectedSubstring) {
		t.Errorf("Expected error to contain '%s', but got: %v", expectedSubstring, err)
	}
}

// Helper function to check if a string contains a substring
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    (len(s) > len(substr) && 
		     (s[:len(substr)] == substr || 
		      s[len(s)-len(substr):] == substr || 
		      containsSubstringRecursive(s, substr))))
}

func containsSubstringRecursive(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	if s[:len(substr)] == substr {
		return true
	}
	return containsSubstringRecursive(s[1:], substr)
}
