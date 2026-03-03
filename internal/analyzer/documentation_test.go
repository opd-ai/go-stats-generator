package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDocumentationAnalyzer(t *testing.T) {
	fset := token.NewFileSet()

	t.Run("with nil config uses defaults", func(t *testing.T) {
		analyzer := NewDocumentationAnalyzer(fset, nil)
		require.NotNil(t, analyzer)
		assert.NotNil(t, analyzer.cfg)
		assert.True(t, analyzer.cfg.RequireExportedDoc)
		assert.True(t, analyzer.cfg.RequirePackageDoc)
		assert.Equal(t, 180, analyzer.cfg.StaleAnnotationDays)
		assert.Equal(t, 5, analyzer.cfg.MinCommentWords)
	})

	t.Run("with custom config", func(t *testing.T) {
		cfg := &DocumentationConfig{
			RequireExportedDoc:  false,
			RequirePackageDoc:   false,
			StaleAnnotationDays: 90,
			MinCommentWords:     3,
		}
		analyzer := NewDocumentationAnalyzer(fset, cfg)
		require.NotNil(t, analyzer)
		assert.Equal(t, cfg, analyzer.cfg)
	})
}

func TestCheckExportedDoc(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewDocumentationAnalyzer(fset, nil)

	tests := []struct {
		name     string
		symbol   string
		comment  string
		expected bool
	}{
		{
			name:     "valid GoDoc",
			symbol:   "MyFunc",
			comment:  "// MyFunc performs some operation and returns a result.",
			expected: true,
		},
		{
			name:     "no comment",
			symbol:   "MyFunc",
			comment:  "",
			expected: false,
		},
		{
			name:     "comment without symbol name",
			symbol:   "MyFunc",
			comment:  "// This function does something cool.",
			expected: false,
		},
		{
			name:     "too short comment",
			symbol:   "MyFunc",
			comment:  "// MyFunc does it.",
			expected: false,
		},
		{
			name:     "multiline valid comment",
			symbol:   "Process",
			comment:  "// Process handles the input data by validating it\n// and then transforming it into the desired format.",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var doc *ast.CommentGroup
			if tt.comment != "" {
				src := "package test\n\n" + tt.comment + "\nfunc " + tt.symbol + "() {}"
				file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
				require.NoError(t, err)
				if len(file.Decls) > 0 {
					if fn, ok := file.Decls[0].(*ast.FuncDecl); ok {
						doc = fn.Doc
					}
				}
			}

			result := analyzer.checkExportedDoc(tt.symbol, doc)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractAnnotation(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewDocumentationAnalyzer(fset, nil)

	tests := []struct {
		name         string
		comment      string
		expectedCat  string
		expectedDesc string
	}{
		{
			name:         "TODO annotation",
			comment:      "// TODO: refactor this function",
			expectedCat:  "TODO",
			expectedDesc: "refactor this function",
		},
		{
			name:         "FIXME annotation",
			comment:      "// FIXME handle edge case",
			expectedCat:  "FIXME",
			expectedDesc: "handle edge case",
		},
		{
			name:         "HACK annotation",
			comment:      "// HACK: workaround for upstream bug",
			expectedCat:  "HACK",
			expectedDesc: "workaround for upstream bug",
		},
		{
			name:         "BUG annotation",
			comment:      "// BUG: nil pointer panic",
			expectedCat:  "BUG",
			expectedDesc: "nil pointer panic",
		},
		{
			name:         "XXX annotation",
			comment:      "// XXX: needs review",
			expectedCat:  "XXX",
			expectedDesc: "needs review",
		},
		{
			name:         "DEPRECATED annotation",
			comment:      "// DEPRECATED: use NewAPI instead",
			expectedCat:  "DEPRECATED",
			expectedDesc: "use NewAPI instead",
		},
		{
			name:         "NOTE annotation",
			comment:      "// NOTE important context here",
			expectedCat:  "NOTE",
			expectedDesc: "important context here",
		},
		{
			name:         "lowercase todo",
			comment:      "// todo: this should work too",
			expectedCat:  "TODO",
			expectedDesc: "this should work too",
		},
		{
			name:         "no annotation",
			comment:      "// regular comment",
			expectedCat:  "",
			expectedDesc: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			category, description := analyzer.extractAnnotation(tt.comment)
			assert.Equal(t, tt.expectedCat, category)
			assert.Equal(t, tt.expectedDesc, description)
		})
	}
}

func TestGetSeverity(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewDocumentationAnalyzer(fset, nil)

	tests := []struct {
		category string
		expected string
	}{
		{"FIXME", "critical"},
		{"BUG", "critical"},
		{"HACK", "high"},
		{"TODO", "medium"},
		{"XXX", "medium"},
		{"DEPRECATED", "low"},
		{"NOTE", "low"},
		{"UNKNOWN", "low"},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			severity := analyzer.getSeverity(tt.category)
			assert.Equal(t, tt.expected, severity)
		})
	}
}

func TestAnalyzeBasic(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewDocumentationAnalyzer(fset, nil)

	src := `package test

// MyFunc does something useful and important.
func MyFunc() {}
`

	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	require.NoError(t, err)

	result := analyzer.Analyze([]*ast.File{file}, nil)
	require.NotNil(t, result)
	assert.NotNil(t, result.Coverage)
	assert.NotNil(t, result.Quality)
}

func TestAnalyzeExportedSymbols(t *testing.T) {
	tests := []struct {
		name               string
		source             string
		expectedFuncCov    float64
		expectedTypeCov    float64
		expectedMethodCov  float64
		expectedOverallCov float64
	}{
		{
			name: "all documented",
			source: `package test

// MyFunc performs an important operation successfully.
func MyFunc() {}

// MyType represents something important in our system.
type MyType struct{}

// Process handles the processing logic efficiently.
func (m MyType) Process() {}
`,
			expectedFuncCov:    100.0,
			expectedTypeCov:    100.0,
			expectedMethodCov:  100.0,
			expectedOverallCov: 100.0,
		},
		{
			name: "none documented",
			source: `package test

func MyFunc() {}

type MyType struct{}

func (m MyType) Process() {}
`,
			expectedFuncCov:    0.0,
			expectedTypeCov:    0.0,
			expectedMethodCov:  0.0,
			expectedOverallCov: 0.0,
		},
		{
			name: "partially documented",
			source: `package test

// MyFunc performs an important operation successfully.
func MyFunc() {}

func AnotherFunc() {}

// MyType represents something important in our system.
type MyType struct{}

type UndocType struct{}

// Process handles the processing logic efficiently.
func (m MyType) Process() {}

func (u UndocType) Run() {}
`,
			expectedFuncCov:    50.0,
			expectedTypeCov:    50.0,
			expectedMethodCov:  50.0,
			expectedOverallCov: 50.0,
		},
		{
			name: "unexported symbols ignored",
			source: `package test

// MyFunc performs an important operation successfully.
func MyFunc() {}

func privateFunc() {}

// MyType represents something important in our system.
type MyType struct{}

type privateType struct{}
`,
			expectedFuncCov:    100.0,
			expectedTypeCov:    100.0,
			expectedMethodCov:  0.0,
			expectedOverallCov: 100.0,
		},
		{
			name: "inadequate documentation",
			source: `package test

// MyFunc does it.
func MyFunc() {}

// MyType is.
type MyType struct{}
`,
			expectedFuncCov:    0.0,
			expectedTypeCov:    0.0,
			expectedMethodCov:  0.0,
			expectedOverallCov: 0.0,
		},
		{
			name: "comment without symbol name",
			source: `package test

// This function does something.
func MyFunc() {}

// A type for testing.
type MyType struct{}
`,
			expectedFuncCov:    0.0,
			expectedTypeCov:    0.0,
			expectedMethodCov:  0.0,
			expectedOverallCov: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			analyzer := NewDocumentationAnalyzer(fset, nil)

			file, err := parser.ParseFile(fset, "test.go", tt.source, parser.ParseComments)
			require.NoError(t, err)

			result := analyzer.Analyze([]*ast.File{file}, nil)
			require.NotNil(t, result)

			assert.Equal(t, tt.expectedFuncCov, result.Coverage.Functions, "Functions coverage mismatch")
			assert.Equal(t, tt.expectedTypeCov, result.Coverage.Types, "Types coverage mismatch")
			assert.Equal(t, tt.expectedMethodCov, result.Coverage.Methods, "Methods coverage mismatch")
			assert.Equal(t, tt.expectedOverallCov, result.Coverage.Overall, "Overall coverage mismatch")
		})
	}
}

func TestAnalyzeMultipleFiles(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewDocumentationAnalyzer(fset, nil)

	src1 := `package test

// Func1 does something important and useful for testing.
func Func1() {}

// Type1 represents an important concept in the system.
type Type1 struct{}
`

	src2 := `package test

func Func2() {}

// Type2 represents another important concept in the codebase.
type Type2 struct{}
`

	file1, err := parser.ParseFile(fset, "test1.go", src1, parser.ParseComments)
	require.NoError(t, err)

	file2, err := parser.ParseFile(fset, "test2.go", src2, parser.ParseComments)
	require.NoError(t, err)

	result := analyzer.Analyze([]*ast.File{file1, file2}, nil)
	require.NotNil(t, result)

	// 1 of 2 functions documented = 50%
	assert.Equal(t, 50.0, result.Coverage.Functions)
	// 2 of 2 types documented = 100%
	assert.Equal(t, 100.0, result.Coverage.Types)
	// 3 of 4 total documented = 75%
	assert.Equal(t, 75.0, result.Coverage.Overall)
}
