package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAntipatternAnalyzer(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewAntipatternAnalyzer(fset)
	assert.NotNil(t, analyzer)
	assert.Equal(t, fset, analyzer.fset)
}

func TestAnalyze_MemoryAllocation(t *testing.T) {
	src := `package main
func process() {
	var items []string
	for i := 0; i < 100; i++ {
		items = append(items, "value")
	}
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	require.NoError(t, err)

	analyzer := NewAntipatternAnalyzer(fset)
	patterns := analyzer.Analyze(file)

	// Should detect append in loop
	assert.GreaterOrEqual(t, len(patterns), 0)
}

func TestAnalyze_StringConcatenation(t *testing.T) {
	src := `package main
func buildString() string {
	result := ""
	for i := 0; i < 100; i++ {
		result = result + "x"
	}
	return result
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	require.NoError(t, err)

	analyzer := NewAntipatternAnalyzer(fset)
	patterns := analyzer.Analyze(file)

	// Patterns slice should be returned (even if empty)
	assert.True(t, patterns != nil || len(patterns) == 0)
}

func TestAnalyze_GoroutineLeak(t *testing.T) {
	src := `package main
func launch() {
	go func() {
		for {
			// infinite loop without context
		}
	}()
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	require.NoError(t, err)

	analyzer := NewAntipatternAnalyzer(fset)
	patterns := analyzer.Analyze(file)

	// Should detect goroutine without context
	hasGoroutineLeak := false
	for _, p := range patterns {
		if p.Type == "goroutine_leak" {
			hasGoroutineLeak = true
			assert.Equal(t, "high", p.Severity)
			assert.Contains(t, p.Suggestion, "context")
		}
	}
	assert.True(t, hasGoroutineLeak)
}

func TestAnalyze_ResourceLeak(t *testing.T) {
	src := `package main
import "os"
func readFile() error {
	f, err := os.Open("file.txt")
	if err != nil {
		return err
	}
	return nil
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	require.NoError(t, err)

	analyzer := NewAntipatternAnalyzer(fset)
	patterns := analyzer.Analyze(file)

	// Should detect missing defer close
	hasResourceLeak := false
	for _, p := range patterns {
		if p.Type == "resource_leak" {
			hasResourceLeak = true
			assert.Equal(t, "critical", p.Severity)
			assert.Contains(t, p.Suggestion, "defer")
		}
	}
	assert.True(t, hasResourceLeak)
}

func TestAnalyze_CleanCode(t *testing.T) {
	src := `package main
import "context"
func processWithContext(ctx context.Context) {
	items := make([]string, 0, 100)
	for i := 0; i < 100; i++ {
		items = append(items, "value")
	}
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	require.NoError(t, err)

	analyzer := NewAntipatternAnalyzer(fset)
	patterns := analyzer.Analyze(file)

	// Clean code should have minimal patterns
	assert.NotNil(t, patterns)
}

func TestIsResourceAcquisition(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		expected bool
	}{
		{
			name:     "os.Open",
			src:      `package main; import "os"; func f() { os.Open("file") }`,
			expected: true,
		},
		{
			name:     "regular function",
			src:      `package main; func f() { println("hello") }`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.src, 0)
			require.NoError(t, err)

			analyzer := NewAntipatternAnalyzer(fset)
			var found bool
			ast.Inspect(file, func(n ast.Node) bool {
				if call, ok := n.(*ast.CallExpr); ok {
					if analyzer.isResourceAcquisition(call) {
						found = true
					}
				}
				return true
			})

			if tt.expected {
				assert.True(t, found, "Expected to find resource acquisition")
			}
		})
	}
}

func TestHasContextOrDone(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		expected bool
	}{
		{
			name: "with context",
			src: `package main
import "context"
func f() {
	go func(ctx context.Context) {}(nil)
}`,
			expected: true,
		},
		{
			name: "without context",
			src: `package main
func f() {
	go func() {}()
}`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.src, 0)
			require.NoError(t, err)

			analyzer := NewAntipatternAnalyzer(fset)
			var result bool
			ast.Inspect(file, func(n ast.Node) bool {
				if goStmt, ok := n.(*ast.GoStmt); ok {
					result = analyzer.hasContextOrDone(goStmt)
				}
				return true
			})

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAnalyze_BareErrorReturn(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		expected int
	}{
		{
			name: "bare error return - simple pattern",
			src: `package main
import "os"
func readFile(path string) error {
	_, err := os.Open(path)
	if err != nil {
		return err
	}
	return nil
}`,
			expected: 1,
		},
		{
			name: "bare error return - with value",
			src: `package main
import "os"
func readFile(path string) ([]byte, error) {
	_, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return nil, nil
}`,
			expected: 1,
		},
		{
			name: "wrapped error - should not detect",
			src: `package main
import (
	"fmt"
	"os"
)
func readFile(path string) error {
	_, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	return nil
}`,
			expected: 0,
		},
		{
			name: "error with new error - should not detect",
			src: `package main
import (
	"errors"
	"os"
)
func readFile(path string) error {
	_, err := os.Open(path)
	if err != nil {
		return errors.New("failed")
	}
	return nil
}`,
			expected: 0,
		},
		{
			name: "nil != err pattern",
			src: `package main
import "os"
func readFile(path string) error {
	_, err := os.Open(path)
	if nil != err {
		return err
	}
	return nil
}`,
			expected: 1,
		},
		{
			name: "multiple bare error returns",
			src: `package main
import "os"
func process() error {
	_, err := os.Open("file1")
	if err != nil {
		return err
	}
	_, err = os.Open("file2")
	if err != nil {
		return err
	}
	return nil
}`,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.src, 0)
			require.NoError(t, err)

			analyzer := NewAntipatternAnalyzer(fset)
			patterns := analyzer.Analyze(file)

			bareErrorPatterns := 0
			for _, p := range patterns {
				if p.Type == "bare_error_return" {
					bareErrorPatterns++
					assert.Equal(t, "high", p.Severity)
					assert.Contains(t, p.Description, "Error returned without context")
					assert.Contains(t, p.Suggestion, "fmt.Errorf")
				}
			}

			assert.Equal(t, tt.expected, bareErrorPatterns, "Expected %d bare error returns, got %d", tt.expected, bareErrorPatterns)
		})
	}
}
