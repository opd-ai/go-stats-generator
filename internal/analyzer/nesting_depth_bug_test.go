package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateNestingDepthRegression(t *testing.T) {
	// Regression test to ensure nesting depth is calculated correctly
	src := `
package main

func nestedFunction() {
	if true {           // depth 1
		for i := 0; i < 10; i++ {  // depth 2
			if i > 5 {  // depth 3
				println(i)
			}
		}
	}
	// After exiting all blocks, we're back at depth 0
	println("done")
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	assert.NoError(t, err)

	analyzer := NewFunctionAnalyzer(fset)

	var funcDecl *ast.FuncDecl
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "nestedFunction" {
			funcDecl = fn
			return false
		}
		return true
	})

	assert.NotNil(t, funcDecl)

	// The maximum nesting depth should be 3 (if->for->if)
	actualDepth := analyzer.calculateNestingDepth(funcDecl.Body)
	assert.Equal(t, 3, actualDepth, "Maximum nesting depth should be 3 (if->for->if)")
}

// TestCalculateNestingDepthFixed - Additional tests to verify the fix works for various cases
func TestCalculateNestingDepthFixed(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected int
	}{
		{
			name: "no nesting",
			code: `
func simple() {
		x := 1
		y := 2
	}`,
			expected: 0,
		},
		{
			name: "single if",
			code: `
func singleIf() {
		if true {
			x := 1
		}
	}`,
			expected: 1,
		},
		{
			name: "nested if statements",
			code: `
func nestedIfs() {
		if true {      // depth 1
			if false { // depth 2
				if true { // depth 3
					x := 1
				}
			}
		}
	}`,
			expected: 3,
		},
		{
			name: "mixed control structures",
			code: `
func mixedNesting() {
		for i := 0; i < 10; i++ {  // depth 1
			switch i {             // depth 2
			case 1:
				if i > 0 {         // depth 3
					println(i)
				}
			}
		}
	}`,
			expected: 3,
		},
		{
			name: "select statement",
			code: `
func selectStmt() {
		select {           // depth 1
		case <-ch:
			if true {      // depth 2
				println("received")
			}
		}
	}`,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := "package main\n" + tt.code
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
			assert.NoError(t, err)

			analyzer := NewFunctionAnalyzer(fset)

			var funcDecl *ast.FuncDecl
			ast.Inspect(file, func(n ast.Node) bool {
				if fn, ok := n.(*ast.FuncDecl); ok {
					funcDecl = fn
					return false
				}
				return true
			})

			assert.NotNil(t, funcDecl)
			actualDepth := analyzer.calculateNestingDepth(funcDecl.Body)
			assert.Equal(t, tt.expected, actualDepth, "Nesting depth mismatch for %s", tt.name)
		})
	}
}

func TestCalculateNestingDepthSimple(t *testing.T) {
	// Test with no nesting
	src := `
package main

func simpleFunction() {
	println("hello")
	x := 42
	println(x)
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	assert.NoError(t, err)

	analyzer := NewFunctionAnalyzer(fset)

	var funcDecl *ast.FuncDecl
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "simpleFunction" {
			funcDecl = fn
			return false
		}
		return true
	})

	assert.NotNil(t, funcDecl)

	// Simple function should have nesting depth of 0 (no nested control structures)
	actualDepth := analyzer.calculateNestingDepth(funcDecl.Body)
	assert.Equal(t, 0, actualDepth, "Simple function should have nesting depth of 0")
}

func TestCalculateNestingDepthComplexCases(t *testing.T) {
	tests := []struct {
		name          string
		src           string
		expectedDepth int
	}{
		{
			name: "switch statement",
			src: `
package main
func switchFunc() {
	switch x {
	case 1:
		if true {
			println("nested")
		}
	}
}`,
			expectedDepth: 2, // switch -> if
		},
		{
			name: "select statement",
			src: `
package main
func selectFunc() {
	select {
	case <-ch:
		for i := 0; i < 10; i++ {
			println(i)
		}
	}
}`,
			expectedDepth: 2, // select -> for
		},
		{
			name: "deeply nested",
			src: `
package main
func deeplyNested() {
	if true {
		for i := 0; i < 10; i++ {
			switch i {
			case 1:
				if i > 0 {
					for j := 0; j < 5; j++ {
						println(j)
					}
				}
			}
		}
	}
}`,
			expectedDepth: 5, // if -> for -> switch -> if -> for
		},
	}

	fset := token.NewFileSet()
	analyzer := NewFunctionAnalyzer(fset)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := parser.ParseFile(fset, "test.go", tt.src, parser.ParseComments)
			assert.NoError(t, err)

			var funcDecl *ast.FuncDecl
			ast.Inspect(file, func(n ast.Node) bool {
				if fn, ok := n.(*ast.FuncDecl); ok {
					funcDecl = fn
					return false
				}
				return true
			})

			assert.NotNil(t, funcDecl)
			actualDepth := analyzer.calculateNestingDepth(funcDecl.Body)
			assert.Equal(t, tt.expectedDepth, actualDepth, "Nesting depth mismatch for %s", tt.name)
		})
	}
}
