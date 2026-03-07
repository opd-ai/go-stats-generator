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
			wantLen:  1,
			wantVals: []string{"42"},
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
		wantSeverity   metrics.SeverityLevel
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

func TestDetectDeepNesting(t *testing.T) {
	tests := []struct {
		name       string
		src        string
		maxNesting int
		wantIssue  bool
		wantDepth  int
	}{
		{
			name: "shallow nesting - no issue",
			src: `package test
func ShallowNesting() {
	if true {
		x := 1
		if x > 0 {
			y := 2
		}
	}
}`,
			maxNesting: 4,
			wantIssue:  false,
		},
		{
			name: "deep nesting - exceeds threshold",
			src: `package test
func DeepNesting() {
	if true {
		if true {
			if true {
				if true {
					if true {
						x := 1
					}
				}
			}
		}
	}
}`,
			maxNesting: 4,
			wantIssue:  true,
			wantDepth:  5,
		},
		{
			name: "mixed control structures",
			src: `package test
func MixedNesting() {
	for i := 0; i < 10; i++ {
		if i > 5 {
			switch i {
				case 6:
					for j := 0; j < i; j++ {
						if j > 3 {
							x := 1
						}
					}
			}
		}
	}
}`,
			maxNesting: 4,
			wantIssue:  true,
			wantDepth:  5,
		},
		{
			name: "range statements",
			src: `package test
func RangeNesting() {
	for _, v := range []int{1, 2, 3} {
		if v > 0 {
			for _, w := range []int{4, 5} {
				if w > 4 {
					if v == w {
						x := 1
					}
				}
			}
		}
	}
}`,
			maxNesting: 4,
			wantIssue:  true,
			wantDepth:  5,
		},
		{
			name: "select statement nesting",
			src: `package test
func SelectNesting() {
	ch := make(chan int)
	for {
		select {
		case v := <-ch:
			if v > 0 {
				for i := 0; i < v; i++ {
					if i > 5 {
						x := 1
					}
				}
			}
		}
	}
}`,
			maxNesting: 4,
			wantIssue:  true,
			wantDepth:  5,
		},
		{
			name: "exactly at threshold",
			src: `package test
func AtThreshold() {
	if true {
		if true {
			if true {
				if true {
					x := 1
				}
			}
		}
	}
}`,
			maxNesting: 4,
			wantIssue:  false,
		},
		{
			name: "nil function",
			src: `package test
// empty`,
			maxNesting: 4,
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

			result := analyzer.DetectDeepNesting(fn, tt.maxNesting)

			if tt.wantIssue {
				if result == nil {
					t.Fatal("expected nesting issue but got nil")
				}

				if result.MaxDepth != tt.wantDepth {
					t.Errorf("max depth = %d, want %d", result.MaxDepth, tt.wantDepth)
				}

				if result.Function == "" {
					t.Error("expected non-empty function name")
				}

				if result.File == "" {
					t.Error("expected non-empty file name")
				}

				if result.Line == 0 {
					t.Error("expected non-zero line number")
				}

				if result.Location == "" {
					t.Error("expected non-empty location")
				}

				if result.Suggestion == "" {
					t.Error("expected non-empty suggestion")
				}
			} else {
				if result != nil {
					t.Errorf("expected no issue but got: %+v", result)
				}
			}
		})
	}
}

func TestDetectFeatureEnvy(t *testing.T) {
	tests := []struct {
		name         string
		src          string
		ratio        float64
		wantIssue    bool
		wantExtType  string
		wantMinRatio float64
	}{
		{
			name: "detects feature envy when external refs exceed threshold",
			src: `package test
type Person struct{}
type Address struct{ City string }

func (p *Person) GetCity(addr Address) string {
	// 5 external references to Address
	c1 := addr.City
	c2 := addr.City
	c3 := addr.City
	c4 := addr.City
	return addr.City
}`,
			ratio:        2.0,
			wantIssue:    true,
			wantExtType:  "addr",
			wantMinRatio: 2.0,
		},
		{
			name: "no issue when self-references dominate",
			src: `package test
type Person struct{ Name string; Age int }

func (p *Person) GetInfo() string {
	return p.Name + string(p.Age)
}`,
			ratio:     2.0,
			wantIssue: false,
		},
		{
			name: "no issue for non-methods",
			src: `package test
func standaloneFunc() {}`,
			ratio:     2.0,
			wantIssue: false,
		},
		{
			name: "no issue when ratio not met",
			src: `package test
type Person struct{ Name string }
type Address struct{ City string }

func (p *Person) Format(addr Address) string {
	return p.Name + addr.City
}`,
			ratio:     5.0,
			wantIssue: false,
		},
		{
			name:      "handles nil function gracefully",
			src:       `package test`,
			ratio:     2.0,
			wantIssue: false,
		},
		{
			name: "detects with pointer receivers",
			src: `package test
type Handler struct{}
type Logger struct{ Level string }

func (h *Handler) Log(logger Logger) {
	_ = logger.Level
	_ = logger.Level
	_ = logger.Level
}`,
			ratio:        2.0,
			wantIssue:    true,
			wantExtType:  "logger",
			wantMinRatio: 2.0,
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

			var fnDecl *ast.FuncDecl
			ast.Inspect(file, func(n ast.Node) bool {
				if fd, ok := n.(*ast.FuncDecl); ok && fd.Recv != nil {
					fnDecl = fd
					return false
				}
				return true
			})

			if fnDecl == nil && tt.wantIssue {
				t.Fatal("Expected to find method declaration")
			}

			result := analyzer.DetectFeatureEnvy(fnDecl, file, tt.ratio)

			if tt.wantIssue {
				if result == nil {
					t.Error("Expected feature envy issue but got nil")
					return
				}

				if result.ExternalType != tt.wantExtType {
					t.Errorf("External type = %q, want %q", result.ExternalType, tt.wantExtType)
				}

				if result.Ratio < tt.wantMinRatio {
					t.Errorf("Ratio = %.2f, want >= %.2f", result.Ratio, tt.wantMinRatio)
				}

				if result.SuggestedMove == "" {
					t.Error("Expected suggestion but got empty string")
				}
			} else {
				if result != nil {
					t.Errorf("Expected no issue but got: %+v", result)
				}
			}
		})
	}
}
