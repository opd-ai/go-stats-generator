package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestNewBurdenAnalyzer(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewBurdenAnalyzer(fset)

	if analyzer == nil {
		t.Fatal("NewBurdenAnalyzer returned nil")
	}

	if analyzer.FileSet() != fset {
		t.Error("FileSet() did not return the expected token.FileSet")
	}
}

func TestDetectMagicNumbers(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		wantLen  int
		wantVals []string
	}{
		{
			name: "detects numeric literals",
			src: `package test
func example() {
	x := 42
	y := 3.14
}`,
			wantLen:  2,
			wantVals: []string{"42", "3.14"},
		},
		{
			name: "detects string literals",
			src: `package test
func example() {
	s := "hardcoded"
}`,
			wantLen:  1,
			wantVals: []string{`"hardcoded"`},
		},
		{
			name: "ignores benign numbers",
			src: `package test
func example() {
	x := 0
	y := 1
	z := -1
}`,
			wantLen: 0,
		},
		{
			name: "ignores empty strings",
			src: `package test
func example() {
	s := ""
}`,
			wantLen: 0,
		},
		{
			name: "ignores const declarations",
			src: `package test
const MaxRetries = 42
const Timeout = 3.14
func example() {
	x := MaxRetries
}`,
			wantLen: 0,
		},
		{
			name: "detects in various contexts",
			src: `package test
func example() int {
	x := 42
	return 100
}`,
			wantLen:  2,
			wantVals: []string{"42", "100"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			analyzer := NewBurdenAnalyzer(fset)

			file, err := parser.ParseFile(fset, "test.go", tt.src, 0)
			if err != nil {
				t.Fatal(err)
			}

			result := analyzer.DetectMagicNumbers(file, "test")

			if len(result) != tt.wantLen {
				t.Errorf("got %d magic numbers, want %d", len(result), tt.wantLen)
			}

			if tt.wantVals != nil {
				for i, want := range tt.wantVals {
					if i >= len(result) {
						t.Errorf("missing magic number at index %d", i)
						continue
					}
					if result[i].Value != want {
						t.Errorf("magic number[%d] = %s, want %s", i, result[i].Value, want)
					}
				}
			}

			// Validate structure of magic numbers
			for _, mn := range result {
				if mn.Function != "example" {
					t.Errorf("expected function='example', got '%s'", mn.Function)
				}
				if mn.Line == 0 {
					t.Error("expected non-zero line number")
				}
				if mn.Type != "numeric" && mn.Type != "string" {
					t.Errorf("unexpected type: %s", mn.Type)
				}
			}
		})
	}
}

func TestDetectDeadCode(t *testing.T) {
	tests := []struct {
		name                  string
		src                   string
		wantUnreferenced      int
		wantUnreachable       int
		wantUnreferencedFn    string
		wantUnreachableReason string
	}{
		{
			name: "detects unreferenced unexported function",
			src: `package test
func Exported() {
	helper()
}
func helper() {}
func unusedHelper() {}`,
			wantUnreferenced:   1,
			wantUnreferencedFn: "unusedHelper",
		},
		{
			name: "ignores exported functions",
			src: `package test
func ExportedUnused() {}`,
			wantUnreferenced: 0,
		},
		{
			name: "detects unreachable after return",
			src: `package test
func Example() {
	return
	x := 42
	y := 100
}`,
			wantUnreachable:       1,
			wantUnreachableReason: "return statement",
		},
		{
			name: "detects unreachable after panic",
			src: `package test
func Example() {
	panic("error")
	x := 42
}`,
			wantUnreachable:       1,
			wantUnreachableReason: "panic call",
		},
		{
			name: "detects unreachable after os.Exit",
			src: `package test
import "os"
func Example() {
	os.Exit(1)
	x := 42
}`,
			wantUnreachable:       1,
			wantUnreachableReason: "os.Exit call",
		},
		{
			name: "no dead code",
			src: `package test
func Used() {
	helper()
}
func helper() {}`,
			wantUnreferenced: 0,
			wantUnreachable:  0,
		},
		{
			name: "unreachable in nested blocks",
			src: `package test
func Example() {
	if true {
		return
		x := 42
	}
}`,
			wantUnreachable: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			analyzer := NewBurdenAnalyzer(fset)

			file, err := parser.ParseFile(fset, "test.go", tt.src, 0)
			if err != nil {
				t.Fatal(err)
			}

			result := analyzer.DetectDeadCode([]*ast.File{file}, "test")
			if result == nil {
				t.Fatal("DetectDeadCode returned nil")
			}

			if len(result.UnreferencedFunctions) != tt.wantUnreferenced {
				t.Errorf("got %d unreferenced functions, want %d",
					len(result.UnreferencedFunctions), tt.wantUnreferenced)
			}

			if tt.wantUnreferencedFn != "" && len(result.UnreferencedFunctions) > 0 {
				if result.UnreferencedFunctions[0].Name != tt.wantUnreferencedFn {
					t.Errorf("unreferenced function = %s, want %s",
						result.UnreferencedFunctions[0].Name, tt.wantUnreferencedFn)
				}
				if result.UnreferencedFunctions[0].Type != "function" {
					t.Errorf("expected type='function', got '%s'", result.UnreferencedFunctions[0].Type)
				}
			}

			if len(result.UnreachableCode) != tt.wantUnreachable {
				t.Errorf("got %d unreachable blocks, want %d",
					len(result.UnreachableCode), tt.wantUnreachable)
			}

			if tt.wantUnreachableReason != "" && len(result.UnreachableCode) > 0 {
				if result.UnreachableCode[0].Reason != tt.wantUnreachableReason {
					t.Errorf("unreachable reason = %s, want %s",
						result.UnreachableCode[0].Reason, tt.wantUnreachableReason)
				}
				if result.UnreachableCode[0].Lines == 0 {
					t.Error("expected non-zero line count for unreachable block")
				}
			}
		})
	}
}

func TestAnalyzeSignatureComplexity(t *testing.T) {
	tests := []struct {
		name           string
		src            string
		maxParams      int
		maxReturns     int
		wantIssue      bool
		wantSeverity   string
		wantParamCount int
		wantRetCount   int
		wantBoolParams int
	}{
		{
			name: "function under thresholds",
			src: `package test
func Simple(a, b int) int {
	return a + b
}`,
			maxParams:      5,
			maxReturns:     3,
			wantIssue:      false,
			wantParamCount: 2,
			wantRetCount:   1,
		},
		{
			name: "function with too many params",
			src: `package test
func TooManyParams(a, b, c, d, e, f int) int {
	return a + b
}`,
			maxParams:      5,
			maxReturns:     3,
			wantIssue:      true,
			wantSeverity:   "medium",
			wantParamCount: 6,
			wantRetCount:   1,
		},
		{
			name: "function with too many returns",
			src: `package test
func TooManyReturns(a int) (int, int, int, int) {
	return a, a, a, a
}`,
			maxParams:      5,
			maxReturns:     3,
			wantIssue:      true,
			wantSeverity:   "medium",
			wantParamCount: 1,
			wantRetCount:   4,
		},
		{
			name: "function with bool params",
			src: `package test
func WithBoolParam(value int, flag bool) int {
	return value
}`,
			maxParams:      5,
			maxReturns:     3,
			wantIssue:      true,
			wantSeverity:   "low",
			wantParamCount: 2,
			wantRetCount:   1,
			wantBoolParams: 1,
		},
		{
			name: "high severity - double threshold",
			src: `package test
func Extreme(a, b, c, d, e, f, g, h, i, j, k int) int {
	return a
}`,
			maxParams:      5,
			maxReturns:     3,
			wantIssue:      true,
			wantSeverity:   "high",
			wantParamCount: 11,
			wantRetCount:   1,
		},
		{
			name: "nil function",
			src: `package test
// empty file`,
			maxParams:  5,
			maxReturns: 3,
			wantIssue:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			analyzer := NewBurdenAnalyzer(fset)

			file, err := parser.ParseFile(fset, "test.go", tt.src, 0)
			if err != nil {
				t.Fatal(err)
			}

			var fn *ast.FuncDecl
			ast.Inspect(file, func(n ast.Node) bool {
				if f, ok := n.(*ast.FuncDecl); ok {
					fn = f
					return false
				}
				return true
			})

			result := analyzer.AnalyzeSignatureComplexity(fn, tt.maxParams, tt.maxReturns)

			if tt.wantIssue {
				if result == nil {
					t.Fatal("expected issue but got nil")
				}

				if result.Severity != tt.wantSeverity {
					t.Errorf("severity = %s, want %s", result.Severity, tt.wantSeverity)
				}

				if result.ParameterCount != tt.wantParamCount {
					t.Errorf("parameter count = %d, want %d", result.ParameterCount, tt.wantParamCount)
				}

				if result.ReturnCount != tt.wantRetCount {
					t.Errorf("return count = %d, want %d", result.ReturnCount, tt.wantRetCount)
				}

				if len(result.BoolParams) != tt.wantBoolParams {
					t.Errorf("bool params count = %d, want %d", len(result.BoolParams), tt.wantBoolParams)
				}

				if result.File == "" {
					t.Error("expected non-empty file name")
				}

				if result.Line == 0 {
					t.Error("expected non-zero line number")
				}
			} else {
				if result != nil {
					t.Errorf("expected no issue but got: %+v", result)
				}
			}
		})
	}
}

func TestDetectDeepNesting_Placeholder(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewBurdenAnalyzer(fset)

	result := analyzer.DetectDeepNesting(nil, 4)
	if result != nil {
		t.Error("Expected nil result from placeholder implementation")
	}
}

func TestDetectFeatureEnvy_Placeholder(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewBurdenAnalyzer(fset)

	result := analyzer.DetectFeatureEnvy(nil, 2.0)
	if result != nil {
		t.Error("Expected nil result from placeholder implementation")
	}
}
