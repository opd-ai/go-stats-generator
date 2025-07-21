package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

// Helper function to create a temporary test file
func createTestFile(t *testing.T, content string) string {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "function_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tempDir) })

	filepath := filepath.Join(tempDir, "test.go")
	err = os.WriteFile(filepath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	return filepath
}

// Helper function to parse a function from test content
func parseTestFunction(t *testing.T, content string) (*ast.FuncDecl, *token.FileSet) {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", content, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse test content: %v", err)
	}

	// Find the first function
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			return funcDecl, fset
		}
	}

	t.Fatal("No function found in test content")
	return nil, nil
}

func TestNewFunctionAnalyzer(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewFunctionAnalyzer(fset)

	if analyzer == nil {
		t.Fatal("NewFunctionAnalyzer returned nil")
	}

	if analyzer.fset != fset {
		t.Error("FunctionAnalyzer fset not set correctly")
	}
}

func TestCountLines_SimpleFunction(t *testing.T) {
	content := `package main

func simpleFunction() {
	x := 1
	y := 2
	return x + y
}`

	funcDecl, fset := parseTestFunction(t, content)
	analyzer := NewFunctionAnalyzer(fset)

	// Create temporary file to get accurate line counting
	filepath := createTestFile(t, content)

	// Parse the file again with the actual file path
	file, err := parser.ParseFile(fset, filepath, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	// Get the function from the parsed file
	funcDecl = file.Decls[0].(*ast.FuncDecl)

	lineMetrics := analyzer.countLines(funcDecl)

	expectedCode := 3 // x := 1, y := 2, return x + y
	if lineMetrics.Code != expectedCode {
		t.Errorf("Expected %d code lines, got %d", expectedCode, lineMetrics.Code)
	}

	if lineMetrics.Comments != 0 {
		t.Errorf("Expected 0 comment lines, got %d", lineMetrics.Comments)
	}

	if lineMetrics.Blank != 0 {
		t.Errorf("Expected 0 blank lines, got %d", lineMetrics.Blank)
	}

	expectedTotal := expectedCode
	if lineMetrics.Total != expectedTotal {
		t.Errorf("Expected %d total lines, got %d", expectedTotal, lineMetrics.Total)
	}
}

func TestCountLines_FunctionWithComments(t *testing.T) {
	content := `package main

func functionWithComments() {
	// This is a comment
	x := 1
	/* Block comment
	   spanning multiple lines */
	y := 2
	z := x + y // Inline comment
	
	return z
}`

	filepath := createTestFile(t, content)
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filepath, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	analyzer := NewFunctionAnalyzer(fset)
	funcDecl := file.Decls[0].(*ast.FuncDecl)

	lineMetrics := analyzer.countLines(funcDecl)

	expectedCode := 4     // x := 1, y := 2, z := x + y, return z
	expectedComments := 3 // single line + 2 block comment lines
	expectedBlank := 1    // One blank line

	if lineMetrics.Code != expectedCode {
		t.Errorf("Expected %d code lines, got %d", expectedCode, lineMetrics.Code)
	}

	if lineMetrics.Comments != expectedComments {
		t.Errorf("Expected %d comment lines, got %d", expectedComments, lineMetrics.Comments)
	}

	if lineMetrics.Blank != expectedBlank {
		t.Errorf("Expected %d blank lines, got %d", expectedBlank, lineMetrics.Blank)
	}

	expectedTotal := expectedCode + expectedComments + expectedBlank
	if lineMetrics.Total != expectedTotal {
		t.Errorf("Expected %d total lines, got %d", expectedTotal, lineMetrics.Total)
	}
}

func TestCountLines_EmptyFunction(t *testing.T) {
	content := `package main

func emptyFunction() {
}`

	filepath := createTestFile(t, content)
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filepath, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	analyzer := NewFunctionAnalyzer(fset)
	funcDecl := file.Decls[0].(*ast.FuncDecl)

	lineMetrics := analyzer.countLines(funcDecl)

	if lineMetrics.Code != 0 {
		t.Errorf("Expected 0 code lines, got %d", lineMetrics.Code)
	}

	if lineMetrics.Comments != 0 {
		t.Errorf("Expected 0 comment lines, got %d", lineMetrics.Comments)
	}

	if lineMetrics.Blank != 0 {
		t.Errorf("Expected 0 blank lines, got %d", lineMetrics.Blank)
	}

	if lineMetrics.Total != 0 {
		t.Errorf("Expected 0 total lines, got %d", lineMetrics.Total)
	}
}

func TestCountLines_FunctionDeclarationOnly(t *testing.T) {
	content := `package main

func declarationOnly()`

	funcDecl, fset := parseTestFunction(t, content)
	analyzer := NewFunctionAnalyzer(fset)

	lineMetrics := analyzer.countLines(funcDecl)

	// Function declarations without body should return zero metrics
	if lineMetrics.Total != 0 || lineMetrics.Code != 0 || lineMetrics.Comments != 0 || lineMetrics.Blank != 0 {
		t.Errorf("Function declaration without body should return zero metrics, got %+v", lineMetrics)
	}
}

func TestCountLines_ComplexFunction(t *testing.T) {
	content := `package main

func complexFunction() error {
	// Initialize variables
	var result int
	
	/*
	 * This is a complex function that demonstrates
	 * various line counting scenarios
	 */
	if true {
		result = 42 // Set magic number
		
		// Another comment
	}
	
	/* Single line block comment */
	return nil
}`

	filepath := createTestFile(t, content)
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filepath, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	analyzer := NewFunctionAnalyzer(fset)
	funcDecl := file.Decls[0].(*ast.FuncDecl)

	lineMetrics := analyzer.countLines(funcDecl)

	// Expected: var result int, if true {, result = 42, }, return nil = 5 code lines
	expectedCode := 5
	// Expected: // Initialize variables, block comment (4 lines: /*, *, *, */), // Another comment, single line block = 7 comment lines
	expectedComments := 7
	// Expected: 3 blank lines
	expectedBlank := 3

	if lineMetrics.Code != expectedCode {
		t.Errorf("Expected %d code lines, got %d", expectedCode, lineMetrics.Code)
	}

	if lineMetrics.Comments != expectedComments {
		t.Errorf("Expected %d comment lines, got %d", expectedComments, lineMetrics.Comments)
	}

	if lineMetrics.Blank != expectedBlank {
		t.Errorf("Expected %d blank lines, got %d", expectedBlank, lineMetrics.Blank)
	}

	expectedTotal := expectedCode + expectedComments + expectedBlank
	if lineMetrics.Total != expectedTotal {
		t.Errorf("Expected %d total lines, got %d", expectedTotal, lineMetrics.Total)
	}
}

func TestAnalyzeFunctions_Integration(t *testing.T) {
	content := `package test

// TestFunction demonstrates function analysis
func TestFunction(param string) error {
	// Validate input
	if param == "" {
		return fmt.Errorf("param cannot be empty")
	}
	
	// Process the parameter
	result := strings.ToUpper(param)
	
	return nil
}

// SimpleFunction has minimal complexity
func SimpleFunction() {
	x := 1
}`

	filepath := createTestFile(t, content)
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filepath, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	analyzer := NewFunctionAnalyzer(fset)
	functions, err := analyzer.AnalyzeFunctions(file, "test")
	if err != nil {
		t.Fatalf("AnalyzeFunctions failed: %v", err)
	}

	if len(functions) != 2 {
		t.Fatalf("Expected 2 functions, got %d", len(functions))
	}

	// Check first function
	testFunc := functions[0]
	if testFunc.Name != "TestFunction" {
		t.Errorf("Expected function name 'TestFunction', got '%s'", testFunc.Name)
	}

	if testFunc.Lines.Code == 0 {
		t.Error("TestFunction should have code lines counted")
	}

	if testFunc.Lines.Comments == 0 {
		t.Error("TestFunction should have comment lines counted")
	}

	// Check second function
	simpleFunc := functions[1]
	if simpleFunc.Name != "SimpleFunction" {
		t.Errorf("Expected function name 'SimpleFunction', got '%s'", simpleFunc.Name)
	}

	if simpleFunc.Lines.Code != 1 {
		t.Errorf("SimpleFunction should have 1 code line, got %d", simpleFunc.Lines.Code)
	}
}

func TestCountLinesInRange_EdgeCases(t *testing.T) {
	content := `package main

func test() {
	x := 1
	y := 2
}`

	filepath := createTestFile(t, content)
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filepath, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	analyzer := NewFunctionAnalyzer(fset)
	tokenFile := fset.File(file.Pos())

	tests := []struct {
		name       string
		startLine  int
		endLine    int
		expectZero bool
	}{
		{"Invalid range - start > end", 5, 3, true},
		{"Invalid range - start < 1", 0, 5, true},
		{"Invalid range - end > file length", 1, 1000, false}, // Should handle gracefully
		{"Valid single line", 4, 4, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := analyzer.countLinesInRange(tokenFile, tt.startLine, tt.endLine)

			if tt.expectZero {
				if metrics.Total != 0 {
					t.Errorf("Expected zero metrics for invalid range, got total=%d", metrics.Total)
				}
			}
			// For valid ranges, we just ensure it doesn't crash
		})
	}
}
