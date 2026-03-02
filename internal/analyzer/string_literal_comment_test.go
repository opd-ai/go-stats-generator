package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCountLines_StringLiteralWithCommentMarkers tests that comment markers
// inside string literals are not treated as actual comments
func TestCountLines_StringLiteralWithCommentMarkers(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		expectedCode   int
		expectedComments int
		expectedBlank  int
	}{
		{
			name: "URL with // in string literal",
			code: `package main
func example() {
	url := "https://example.com"
}`,
			expectedCode:   1, // Just the url line
			expectedComments: 0,
			expectedBlank:  0,
		},
		{
			name: "URL with // and actual comment",
			code: `package main
func example() {
	url := "https://example.com" // This is a URL
}`,
			expectedCode:   1, // url line (mixed counted as code)
			expectedComments: 0,
			expectedBlank:  0,
		},
		{
			name: "String with /* inside",
			code: `package main
func example() {
	pattern := "/* This is not a comment */"
}`,
			expectedCode:   1, // pattern line
			expectedComments: 0,
			expectedBlank:  0,
		},
		{
			name: "Multiple strings with comment markers",
			code: `package main
func example() {
	url1 := "https://example.com/path"
	url2 := "http://test.com"
	comment := "/* not a comment */"
	slash := "//"
}`,
			expectedCode:   4, // 4 var lines
			expectedComments: 0,
			expectedBlank:  0,
		},
		{
			name: "String literal with // followed by real comment",
			code: `package main
func example() {
	// This is a comment
	url := "https://example.com" // URL comment
	path := "//" // Double slash
}`,
			expectedCode:   2, // url line (mixed), path line (mixed)
			expectedComments: 1, // The standalone comment line
			expectedBlank:  0,
		},
		{
			name: "Backtick string with // and /*",
			code: `package main
func example() {
	text := ` + "`" + `This // is /* all */ text` + "`" + `
}`,
			expectedCode:   1,
			expectedComments: 0,
			expectedBlank:  0,
		},
		{
			name: "Escaped quotes in string",
			code: `package main
func example() {
	s := "She said \"https://example.com\" to me"
}`,
			expectedCode:   1,
			expectedComments: 0,
			expectedBlank:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file for accurate position information
			filepath := createTestFile(t, tt.code)
			
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, filepath, nil, parser.ParseComments)
			require.NoError(t, err)
			require.NotNil(t, file)

			// Find the function named "example"
			var funcDecl *ast.FuncDecl
			for _, decl := range file.Decls {
				if fd, ok := decl.(*ast.FuncDecl); ok && fd.Name.Name == "example" {
					funcDecl = fd
					break
				}
			}
			require.NotNil(t, funcDecl, "Function 'example' should exist")

			analyzer := NewFunctionAnalyzer(fset)
			result := analyzer.countLines(funcDecl)

			assert.Equal(t, tt.expectedCode, result.Code, 
				"Code lines mismatch in %s", tt.name)
			assert.Equal(t, tt.expectedComments, result.Comments, 
				"Comment lines mismatch in %s", tt.name)
			assert.Equal(t, tt.expectedBlank, result.Blank, 
				"Blank lines mismatch in %s", tt.name)
		})
	}
}

// TestClassifyLine_StringLiterals tests the low-level line classification
// with various string literal scenarios
func TestClassifyLine_StringLiterals(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "URL in string",
			line:     `url := "https://example.com"`,
			expected: "code",
		},
		{
			name:     "URL with comment after",
			line:     `url := "https://example.com" // comment`,
			expected: "mixed",
		},
		{
			name:     "Block comment in string",
			line:     `s := "/* not a comment */"`,
			expected: "code",
		},
		{
			name:     "Double slash in string",
			line:     `slash := "//"`,
			expected: "code",
		},
		{
			name:     "Backtick string with //",
			line:     "text := `https://example.com`",
			expected: "code",
		},
		{
			name:     "Escaped quote with //",
			line:     `s := "\"https://test.com\""`,
			expected: "code",
		},
		{
			name:     "Multiple strings on one line",
			line:     `a, b := "https://one.com", "https://two.com"`,
			expected: "code",
		},
		{
			name:     "Empty string before comment",
			line:     `s := "" // comment`,
			expected: "mixed",
		},
	}

	fset := token.NewFileSet()
	analyzer := NewFunctionAnalyzer(fset)
	inBlockComment := false

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.classifyLine(tt.line, &inBlockComment)
			assert.Equal(t, tt.expected, result,
				"Line classification mismatch for: %s", tt.line)
		})
	}
}
