package analyzer

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenericAnalyzer_BasicTypeParameter(t *testing.T) {
	src := `package test
func Identity[T any](x T) T { return x }
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	require.NoError(t, err)

	analyzer := NewGenericAnalyzer(fset)
	result, err := analyzer.AnalyzeGenerics(file, "test", "test.go")
	require.NoError(t, err)

	assert.Equal(t, 1, result.TypeParameters.Count)
	assert.Equal(t, 1, result.TypeParameters.Constraints["any"])
}

func TestGenericAnalyzer_ComparableConstraint(t *testing.T) {
	src := `package test
func Equal[T comparable](a, b T) bool { return a == b }
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	require.NoError(t, err)

	analyzer := NewGenericAnalyzer(fset)
	result, err := analyzer.AnalyzeGenerics(file, "test", "test.go")
	require.NoError(t, err)

	assert.Equal(t, 1, result.TypeParameters.Count)
	assert.Equal(t, 1, result.TypeParameters.Constraints["comparable"])
}

func TestGenericAnalyzer_MultipleTypeParameters(t *testing.T) {
	src := `package test
func Map[T any, U any](items []T, fn func(T) U) []U {
	result := make([]U, len(items))
	return result
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	require.NoError(t, err)

	analyzer := NewGenericAnalyzer(fset)
	result, err := analyzer.AnalyzeGenerics(file, "test", "test.go")
	require.NoError(t, err)

	assert.Equal(t, 2, result.TypeParameters.Count)
	assert.Equal(t, 2, result.TypeParameters.Constraints["any"])
}

func TestGenericAnalyzer_GenericType(t *testing.T) {
	src := `package test
type Stack[T any] struct {
	items []T
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	require.NoError(t, err)

	analyzer := NewGenericAnalyzer(fset)
	result, err := analyzer.AnalyzeGenerics(file, "test", "test.go")
	require.NoError(t, err)

	assert.Equal(t, 1, result.TypeParameters.Count)
	assert.Equal(t, 1, result.TypeParameters.Constraints["any"])
}

func TestGenericAnalyzer_ComplexityScore(t *testing.T) {
	src := `package test
func Process[T any, U comparable](x T, y U) {}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	require.NoError(t, err)

	analyzer := NewGenericAnalyzer(fset)
	result, err := analyzer.AnalyzeGenerics(file, "test", "test.go")
	require.NoError(t, err)

	assert.Greater(t, result.ComplexityScore, 0.0)
}

func TestGenericAnalyzer_NoGenerics(t *testing.T) {
	src := `package test
func Regular(x int) int { return x }
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	require.NoError(t, err)

	analyzer := NewGenericAnalyzer(fset)
	result, err := analyzer.AnalyzeGenerics(file, "test", "test.go")
	require.NoError(t, err)

	assert.Equal(t, 0, result.TypeParameters.Count)
	assert.Equal(t, 0.0, result.ComplexityScore)
}

func TestGenericAnalyzer_GenericInstantiation(t *testing.T) {
	src := `package test
func Use() {
	x := Identity[int](42)
	_ = x
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	require.NoError(t, err)

	analyzer := NewGenericAnalyzer(fset)
	result, err := analyzer.AnalyzeGenerics(file, "test", "test.go")
	require.NoError(t, err)

	assert.NotEmpty(t, result.Instantiations.Functions)
	assert.Equal(t, "Identity", result.Instantiations.Functions[0].GenericName)
	assert.Equal(t, []string{"int"}, result.Instantiations.Functions[0].TypeArgs)
}
