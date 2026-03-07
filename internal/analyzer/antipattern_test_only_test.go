package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckTestOnlyExports_ExportedUsedOnlyInTest(t *testing.T) {
	// Source file with exported function
	srcFile := `package mypackage

// ExportedFunc is exported but only used in tests
func ExportedFunc() string {
	return "test"
}

// unexportedFunc is not exported
func unexportedFunc() string {
	return "private"
}
`

	// Test file that uses ExportedFunc
	testFile := `package mypackage

import "testing"

func TestExportedFunc(t *testing.T) {
	result := ExportedFunc()
	if result != "test" {
		t.Error("failed")
	}
}
`

	fset := token.NewFileSet()
	
	srcAST, err := parser.ParseFile(fset, "mypackage.go", srcFile, 0)
	require.NoError(t, err)
	
	testAST, err := parser.ParseFile(fset, "mypackage_test.go", testFile, 0)
	require.NoError(t, err)

	files := map[string]*ast.File{
		"mypackage.go":      srcAST,
		"mypackage_test.go": testAST,
	}

	patterns := CheckTestOnlyExports(files, fset, "github.com/example/mypackage")

	// Should detect ExportedFunc as test-only export
	assert.Len(t, patterns, 1)
	assert.Equal(t, "test_only_export", patterns[0].Type)
	assert.Equal(t, "low", patterns[0].Severity)
	assert.Contains(t, patterns[0].Description, "ExportedFunc")
	assert.Contains(t, patterns[0].Suggestion, "export_test.go")
}

func TestCheckTestOnlyExports_ExportedUsedInAnotherPackage(t *testing.T) {
	// Source file with exported function
	srcFile := `package mypackage

// ExportedFunc is exported and used in another package
func ExportedFunc() string {
	return "public"
}
`

	// Another source file that uses ExportedFunc
	otherFile := `package mypackage

func usesExportedFunc() string {
	return ExportedFunc() + " api"
}
`

	fset := token.NewFileSet()
	
	srcAST, err := parser.ParseFile(fset, "mypackage.go", srcFile, 0)
	require.NoError(t, err)
	
	otherAST, err := parser.ParseFile(fset, "other.go", otherFile, 0)
	require.NoError(t, err)

	files := map[string]*ast.File{
		"mypackage.go": srcAST,
		"other.go":     otherAST,
	}

	patterns := CheckTestOnlyExports(files, fset, "github.com/example/mypackage")

	// Should NOT detect any violations - ExportedFunc is used in other.go
	// (usesExportedFunc is unexported so it won't be checked)
	assert.Len(t, patterns, 0)
}

func TestCheckTestOnlyExports_UnexportedSymbol(t *testing.T) {
	// Source file with unexported function
	srcFile := `package mypackage

// unexportedFunc is not exported
func unexportedFunc() string {
	return "private"
}
`

	// Test file that uses unexportedFunc
	testFile := `package mypackage

import "testing"

func TestUnexportedFunc(t *testing.T) {
	result := unexportedFunc()
	if result != "private" {
		t.Error("failed")
	}
}
`

	fset := token.NewFileSet()
	
	srcAST, err := parser.ParseFile(fset, "mypackage.go", srcFile, 0)
	require.NoError(t, err)
	
	testAST, err := parser.ParseFile(fset, "mypackage_test.go", testFile, 0)
	require.NoError(t, err)

	files := map[string]*ast.File{
		"mypackage.go":      srcAST,
		"mypackage_test.go": testAST,
	}

	patterns := CheckTestOnlyExports(files, fset, "github.com/example/mypackage")

	// Should NOT detect any violations - unexported symbols are fine
	assert.Len(t, patterns, 0)
}

func TestCheckTestOnlyExports_UnusedExport(t *testing.T) {
	// Source file with exported function that's never used
	srcFile := `package mypackage

// UnusedFunc is exported but never used anywhere
func UnusedFunc() string {
	return "unused"
}
`

	fset := token.NewFileSet()
	
	srcAST, err := parser.ParseFile(fset, "mypackage.go", srcFile, 0)
	require.NoError(t, err)

	files := map[string]*ast.File{
		"mypackage.go": srcAST,
	}

	patterns := CheckTestOnlyExports(files, fset, "github.com/example/mypackage")

	// Should detect UnusedFunc as test-only export (zero cross-package refs)
	assert.Len(t, patterns, 1)
	assert.Equal(t, "test_only_export", patterns[0].Type)
	assert.Contains(t, patterns[0].Description, "UnusedFunc")
}

func TestCheckTestOnlyExports_ExportedType(t *testing.T) {
	// Source file with exported type used only in tests
	srcFile := `package mypackage

// ExportedType is exported but only used in tests
type ExportedType struct {
	Value string
}
`

	// Test file that uses ExportedType
	testFile := `package mypackage

import "testing"

func TestExportedType(t *testing.T) {
	obj := ExportedType{Value: "test"}
	if obj.Value != "test" {
		t.Error("failed")
	}
}
`

	fset := token.NewFileSet()
	
	srcAST, err := parser.ParseFile(fset, "mypackage.go", srcFile, 0)
	require.NoError(t, err)
	
	testAST, err := parser.ParseFile(fset, "mypackage_test.go", testFile, 0)
	require.NoError(t, err)

	files := map[string]*ast.File{
		"mypackage.go":      srcAST,
		"mypackage_test.go": testAST,
	}

	patterns := CheckTestOnlyExports(files, fset, "github.com/example/mypackage")

	// Should detect ExportedType as test-only export
	assert.Len(t, patterns, 1)
	assert.Equal(t, "test_only_export", patterns[0].Type)
	assert.Contains(t, patterns[0].Description, "ExportedType")
	assert.Contains(t, patterns[0].Description, "type")
}

func TestCheckTestOnlyExports_ExportedVariable(t *testing.T) {
	// Source file with exported variable used only in tests
	srcFile := `package mypackage

// ExportedVar is exported but only used in tests
var ExportedVar = "test value"
`

	// Test file that uses ExportedVar
	testFile := `package mypackage

import "testing"

func TestExportedVar(t *testing.T) {
	if ExportedVar != "test value" {
		t.Error("failed")
	}
}
`

	fset := token.NewFileSet()
	
	srcAST, err := parser.ParseFile(fset, "mypackage.go", srcFile, 0)
	require.NoError(t, err)
	
	testAST, err := parser.ParseFile(fset, "mypackage_test.go", testFile, 0)
	require.NoError(t, err)

	files := map[string]*ast.File{
		"mypackage.go":      srcAST,
		"mypackage_test.go": testAST,
	}

	patterns := CheckTestOnlyExports(files, fset, "github.com/example/mypackage")

	// Should detect ExportedVar as test-only export
	assert.Len(t, patterns, 1)
	assert.Equal(t, "test_only_export", patterns[0].Type)
	assert.Contains(t, patterns[0].Description, "ExportedVar")
	assert.Contains(t, patterns[0].Description, "variable")
}

func TestCheckTestOnlyExports_MultipleFiles(t *testing.T) {
	// file1.go: Has ExportedA (used in file2) and ExportedB (unused)
	file1 := `package mypackage

func ExportedA() string {
	return "a"
}

func ExportedB() string {
	return "b"
}
`

	// file2.go: Uses ExportedA
	file2 := `package mypackage

func useExportedA() string {
	return ExportedA()
}
`

	// test file: Uses ExportedB
	testFile := `package mypackage

import "testing"

func TestExportedB(t *testing.T) {
	_ = ExportedB()
}
`

	fset := token.NewFileSet()
	
	file1AST, err := parser.ParseFile(fset, "file1.go", file1, 0)
	require.NoError(t, err)
	
	file2AST, err := parser.ParseFile(fset, "file2.go", file2, 0)
	require.NoError(t, err)
	
	testAST, err := parser.ParseFile(fset, "mypackage_test.go", testFile, 0)
	require.NoError(t, err)

	files := map[string]*ast.File{
		"file1.go":          file1AST,
		"file2.go":          file2AST,
		"mypackage_test.go": testAST,
	}

	patterns := CheckTestOnlyExports(files, fset, "github.com/example/mypackage")

	// Should detect only ExportedB (used only in test)
	// ExportedA is used in file2.go, so it's legitimate
	// useExportedA is unexported so won't be checked
	assert.Len(t, patterns, 1)
	assert.Contains(t, patterns[0].Description, "ExportedB")
}

func TestCheckTestOnlyExports_ConstantDetection(t *testing.T) {
	// Source file with exported constant used only in tests
	srcFile := `package mypackage

// ExportedConst is exported but only used in tests
const ExportedConst = 42
`

	// Test file that uses ExportedConst
	testFile := `package mypackage

import "testing"

func TestExportedConst(t *testing.T) {
	if ExportedConst != 42 {
		t.Error("failed")
	}
}
`

	fset := token.NewFileSet()
	
	srcAST, err := parser.ParseFile(fset, "mypackage.go", srcFile, 0)
	require.NoError(t, err)
	
	testAST, err := parser.ParseFile(fset, "mypackage_test.go", testFile, 0)
	require.NoError(t, err)

	files := map[string]*ast.File{
		"mypackage.go":      srcAST,
		"mypackage_test.go": testAST,
	}

	patterns := CheckTestOnlyExports(files, fset, "github.com/example/mypackage")

	// Should detect ExportedConst as test-only export
	assert.Len(t, patterns, 1)
	assert.Equal(t, "test_only_export", patterns[0].Type)
	assert.Contains(t, patterns[0].Description, "ExportedConst")
	assert.Contains(t, patterns[0].Description, "constant")
}
