package analyzer

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
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
		expected metrics.SeverityLevel
	}{
		{"FIXME", metrics.SeverityLevelCritical},
		{"BUG", metrics.SeverityLevelCritical},
		{"HACK", metrics.SeverityLevelViolation},
		{"TODO", metrics.SeverityLevelWarning},
		{"XXX", metrics.SeverityLevelWarning},
		{"DEPRECATED", metrics.SeverityLevelInfo},
		{"NOTE", metrics.SeverityLevelInfo},
		{"UNKNOWN", metrics.SeverityLevelInfo},
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

func TestAnalyzePackageDocs(t *testing.T) {
	tests := []struct {
		name           string
		sources        []string
		expectedPkgCov float64
	}{
		{
			name: "all packages documented",
			sources: []string{
				`// Package test provides testing utilities.
package test

// MyFunc performs an operation.
func MyFunc() {}`,
				`// Package helper contains helper functions.
package helper

// HelperFunc assists with operations.
func HelperFunc() {}`,
			},
			expectedPkgCov: 100.0,
		},
		{
			name: "no packages documented",
			sources: []string{
				`package test

func MyFunc() {}`,
				`package helper

func HelperFunc() {}`,
			},
			expectedPkgCov: 0.0,
		},
		{
			name: "partially documented",
			sources: []string{
				`// Package test provides testing utilities.
package test

func MyFunc() {}`,
				`package helper

func HelperFunc() {}`,
			},
			expectedPkgCov: 50.0,
		},
		{
			name: "single package with doc",
			sources: []string{
				`// Package test provides testing utilities and helpers.
package test

func Func1() {}`,
			},
			expectedPkgCov: 100.0,
		},
		{
			name: "single package without doc",
			sources: []string{
				`package test

func Func1() {}`,
			},
			expectedPkgCov: 0.0,
		},
		{
			name: "multiple files same package",
			sources: []string{
				`// Package test provides testing utilities.
package test

func Func1() {}`,
				`package test

func Func2() {}`,
			},
			expectedPkgCov: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			analyzer := NewDocumentationAnalyzer(fset, nil)

			var files []*ast.File
			for i, src := range tt.sources {
				file, err := parser.ParseFile(fset, "test"+string(rune('0'+i))+".go", src, parser.ParseComments)
				require.NoError(t, err)
				files = append(files, file)
			}

			result := analyzer.Analyze(files, nil)
			require.NotNil(t, result)

			assert.Equal(t, tt.expectedPkgCov, result.Coverage.Packages, "Package coverage mismatch")
		})
	}
}

func TestAnalyzeAnnotations(t *testing.T) {
	tests := []struct {
		name                  string
		source                string
		expectedTODO          int
		expectedFIXME         int
		expectedHACK          int
		expectedBUG           int
		expectedXXX           int
		expectedDEPR          int
		expectedNOTE          int
		expectedCategoryCount int
	}{
		{
			name: "TODO annotation",
			source: `package test
// TODO: implement this function
func Example() {}`,
			expectedTODO:          1,
			expectedCategoryCount: 1,
		},
		{
			name: "FIXME annotation",
			source: `package test
// FIXME: fix the bug here
func Example() {}`,
			expectedFIXME:         1,
			expectedCategoryCount: 1,
		},
		{
			name: "HACK annotation",
			source: `package test
// HACK: temporary workaround
func Example() {}`,
			expectedHACK:          1,
			expectedCategoryCount: 1,
		},
		{
			name: "BUG annotation",
			source: `package test
// BUG: this causes issues
func Example() {}`,
			expectedBUG:           1,
			expectedCategoryCount: 1,
		},
		{
			name: "XXX annotation",
			source: `package test
// XXX: review this code
func Example() {}`,
			expectedXXX:           1,
			expectedCategoryCount: 1,
		},
		{
			name: "DEPRECATED annotation",
			source: `package test
// DEPRECATED: use NewExample instead
func Example() {}`,
			expectedDEPR:          1,
			expectedCategoryCount: 1,
		},
		{
			name: "NOTE annotation",
			source: `package test
// NOTE: this is important
func Example() {}`,
			expectedNOTE:          1,
			expectedCategoryCount: 1,
		},
		{
			name: "multiple annotations",
			source: `package test
// TODO: implement this
func Example1() {}

// FIXME: fix bug
func Example2() {}

// HACK: workaround
func Example3() {}`,
			expectedTODO:          1,
			expectedFIXME:         1,
			expectedHACK:          1,
			expectedCategoryCount: 3,
		},
		{
			name: "case insensitive",
			source: `package test
// todo: lowercase
func Example1() {}

// ToDo: mixed case
func Example2() {}`,
			expectedTODO:          2,
			expectedCategoryCount: 1,
		},
		{
			name: "annotations with colon",
			source: `package test
// TODO: with colon
func Example() {}`,
			expectedTODO:          1,
			expectedCategoryCount: 1,
		},
		{
			name: "annotations without colon",
			source: `package test
// TODO without colon
func Example() {}`,
			expectedTODO:          1,
			expectedCategoryCount: 1,
		},
		{
			name: "no annotations",
			source: `package test
// This is a regular comment
func Example() {}`,
			expectedCategoryCount: 0,
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

			assert.Len(t, result.TODOComments, tt.expectedTODO, "TODO count mismatch")
			assert.Len(t, result.FIXMEComments, tt.expectedFIXME, "FIXME count mismatch")
			assert.Len(t, result.HACKComments, tt.expectedHACK, "HACK count mismatch")
			assert.Len(t, result.BUGComments, tt.expectedBUG, "BUG count mismatch")
			assert.Len(t, result.XXXComments, tt.expectedXXX, "XXX count mismatch")
			assert.Len(t, result.DEPRECATEDComments, tt.expectedDEPR, "DEPRECATED count mismatch")
			assert.Len(t, result.NOTEComments, tt.expectedNOTE, "NOTE count mismatch")
			assert.Len(t, result.AnnotationsByCategory, tt.expectedCategoryCount, "Category count mismatch")
		})
	}
}

func TestAnnotationDetails(t *testing.T) {
	source := `package test

// TODO: implement feature X
func Example1() {}

// FIXME: critical bug in logic
func Example2() {}

// HACK: temporary fix for issue #123
func Example3() {}`

	fset := token.NewFileSet()
	analyzer := NewDocumentationAnalyzer(fset, nil)

	file, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	require.NoError(t, err)

	result := analyzer.Analyze([]*ast.File{file}, nil)
	require.NotNil(t, result)

	// Verify TODO details
	require.Len(t, result.TODOComments, 1)
	assert.Equal(t, "implement feature X", result.TODOComments[0].Description)
	assert.Equal(t, "test.go", result.TODOComments[0].File)
	assert.Greater(t, result.TODOComments[0].Line, 0)

	// Verify FIXME details
	require.Len(t, result.FIXMEComments, 1)
	assert.Equal(t, "critical bug in logic", result.FIXMEComments[0].Description)
	assert.Equal(t, metrics.SeverityLevelCritical, result.FIXMEComments[0].Severity)
	assert.Greater(t, result.FIXMEComments[0].Line, 0)

	// Verify HACK details
	require.Len(t, result.HACKComments, 1)
	assert.Equal(t, "temporary fix for issue #123", result.HACKComments[0].Description)
	assert.Greater(t, result.HACKComments[0].Line, 0)

	// Verify category counts
	assert.Equal(t, 1, result.AnnotationsByCategory["TODO"])
	assert.Equal(t, 1, result.AnnotationsByCategory["FIXME"])
	assert.Equal(t, 1, result.AnnotationsByCategory["HACK"])
}

func TestSeverityClassification(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewDocumentationAnalyzer(fset, nil)

	tests := []struct {
		category string
		expected metrics.SeverityLevel
	}{
		{"FIXME", metrics.SeverityLevelCritical},
		{"BUG", metrics.SeverityLevelCritical},
		{"HACK", metrics.SeverityLevelViolation},
		{"TODO", metrics.SeverityLevelWarning},
		{"XXX", metrics.SeverityLevelWarning},
		{"DEPRECATED", metrics.SeverityLevelInfo},
		{"NOTE", metrics.SeverityLevelInfo},
		{"UNKNOWN", metrics.SeverityLevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			severity := analyzer.getSeverity(tt.category)
			assert.Equal(t, tt.expected, severity)
		})
	}
}
