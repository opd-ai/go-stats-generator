package analyzer

import (
	"go/parser"
	"go/token"
	"testing"
)

func TestInitComplexityDetection(t *testing.T) {
	src := `package test

import "fmt"

// Simple init - should NOT be flagged
func init() {
fmt.Println("simple init")
}

// Complex init - should be flagged (cyclomatic complexity > 5)
func init() {
for i := 0; i < 10; i++ {
if i%2 == 0 {
if i > 5 {
fmt.Println(i)
} else {
fmt.Println("low")
}
} else {
switch i {
case 1:
fmt.Println("one")
case 3:
fmt.Println("three")
case 7:
fmt.Println("seven")
}
}
}
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse test source: %v", err)
	}

	analyzer := NewAntipatternAnalyzer(fset)
	patterns := analyzer.Analyze(file)

	// Should detect one init_complexity pattern
	var initComplexityFound bool
	for _, p := range patterns {
		t.Logf("Found pattern: Type=%s, Severity=%s, Line=%d, Description=%s",
			p.Type, p.Severity, p.Line, p.Description)
		if p.Type == "init_complexity" {
			initComplexityFound = true
		}
	}

	if !initComplexityFound {
		t.Error("Expected to find init_complexity pattern but did not")
	}
}
