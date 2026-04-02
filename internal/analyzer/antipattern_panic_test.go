package analyzer

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckPanicInLibraryCode_PanicDetected(t *testing.T) {
	src := `package mylib

func ProcessData(data string) string {
	if data == "" {
		panic("empty data")
	}
	return data
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	require.NoError(t, err)

	analyzer := NewAntipatternAnalyzer(fset)
	patterns := analyzer.Analyze(file)

	// Should detect panic in library code
	hasPanicPattern := false
	for _, p := range patterns {
		if p.Type == "panic_in_library" {
			hasPanicPattern = true
			assert.Equal(t, metrics.SeverityLevelViolation, p.Severity)
			assert.Contains(t, p.Description, "panic()")
			assert.Contains(t, p.Suggestion, "Return error instead")
			break
		}
	}
	assert.True(t, hasPanicPattern, "Should detect panic in library code")
}

func TestCheckPanicInLibraryCode_LogFatalDetected(t *testing.T) {
	src := `package mylib

import "log"

func ProcessData(data string) string {
	if data == "" {
		log.Fatal("empty data")
	}
	return data
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	require.NoError(t, err)

	analyzer := NewAntipatternAnalyzer(fset)
	patterns := analyzer.Analyze(file)

	// Should detect log.Fatal in library code
	hasLogFatalPattern := false
	for _, p := range patterns {
		if p.Type == "log_fatal_in_library" {
			hasLogFatalPattern = true
			assert.Equal(t, metrics.SeverityLevelCritical, p.Severity)
			assert.Contains(t, p.Description, "log.Fatal()")
			assert.Contains(t, p.Suggestion, "Return error instead")
			break
		}
	}
	assert.True(t, hasLogFatalPattern, "Should detect log.Fatal in library code")
}

func TestCheckPanicInLibraryCode_LogFatalfDetected(t *testing.T) {
	src := `package mylib

import "log"

func ProcessData(data string) string {
	if data == "" {
		log.Fatalf("empty data: %s", data)
	}
	return data
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	require.NoError(t, err)

	analyzer := NewAntipatternAnalyzer(fset)
	patterns := analyzer.Analyze(file)

	// Should detect log.Fatalf in library code
	hasLogFatalPattern := false
	for _, p := range patterns {
		if p.Type == "log_fatal_in_library" {
			hasLogFatalPattern = true
			assert.Equal(t, metrics.SeverityLevelCritical, p.Severity)
			break
		}
	}
	assert.True(t, hasLogFatalPattern, "Should detect log.Fatalf in library code")
}

func TestCheckPanicInLibraryCode_MainPackageAllowed(t *testing.T) {
	src := `package main

func main() {
	if false {
		panic("application error")
	}
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "main.go", src, 0)
	require.NoError(t, err)

	analyzer := NewAntipatternAnalyzer(fset)
	patterns := analyzer.Analyze(file)

	// Should NOT detect panic in main package
	for _, p := range patterns {
		assert.NotEqual(t, "panic_in_library", p.Type, "Should not flag panic in main package")
		assert.NotEqual(t, "log_fatal_in_library", p.Type, "Should not flag log.Fatal in main package")
	}
}

func TestCheckPanicInLibraryCode_InitFunctionAllowed(t *testing.T) {
	src := `package mylib

func init() {
	if false {
		panic("initialization error")
	}
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	require.NoError(t, err)

	analyzer := NewAntipatternAnalyzer(fset)
	patterns := analyzer.Analyze(file)

	// Should NOT detect panic in init() function
	for _, p := range patterns {
		assert.NotEqual(t, "panic_in_library", p.Type, "Should not flag panic in init() function")
	}
}

func TestCheckPanicInLibraryCode_MultiplePanics(t *testing.T) {
	src := `package mylib

import "log"

func ProcessA(data string) {
	if data == "" {
		panic("empty data")
	}
}

func ProcessB(data string) {
	if data == "" {
		log.Fatal("empty data")
	}
}

func ProcessC(data string) {
	if len(data) > 100 {
		panic("data too long")
	}
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	require.NoError(t, err)

	analyzer := NewAntipatternAnalyzer(fset)
	patterns := analyzer.Analyze(file)

	// Should detect both panic and log.Fatal patterns
	panicCount := 0
	logFatalCount := 0
	for _, p := range patterns {
		if p.Type == "panic_in_library" {
			panicCount++
		} else if p.Type == "log_fatal_in_library" {
			logFatalCount++
		}
	}
	assert.Equal(t, 2, panicCount, "Should detect 2 panic calls")
	assert.Equal(t, 1, logFatalCount, "Should detect 1 log.Fatal call")
}

func TestCheckPanicInLibraryCode_NoFalsePositives(t *testing.T) {
	src := `package mylib

import "errors"

func ProcessData(data string) error {
	if data == "" {
		return errors.New("empty data")
	}
	return nil
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	require.NoError(t, err)

	analyzer := NewAntipatternAnalyzer(fset)
	patterns := analyzer.Analyze(file)

	// Should NOT detect any panic patterns in clean code
	for _, p := range patterns {
		assert.NotEqual(t, "panic_in_library", p.Type, "Should not have false positives for panic")
		assert.NotEqual(t, "log_fatal_in_library", p.Type, "Should not have false positives for log.Fatal")
	}
}
