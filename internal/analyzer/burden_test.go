package analyzer

import (
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

func TestDetectDeadCode_Placeholder(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewBurdenAnalyzer(fset)

	result := analyzer.DetectDeadCode(nil, "test")
	if result != nil {
		t.Error("Expected nil result from placeholder implementation")
	}
}

func TestAnalyzeSignatureComplexity_Placeholder(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewBurdenAnalyzer(fset)

	result := analyzer.AnalyzeSignatureComplexity(nil, 5, 3)
	if result != nil {
		t.Error("Expected nil result from placeholder implementation")
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
