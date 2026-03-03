package analyzer

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectFunctions(t *testing.T) {
	src := `package test
func Foo() {}
func Bar() {}
type T struct{}
func (t T) Method() {}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	require.NoError(t, err)

	funcs := CollectFunctions(file)
	assert.Len(t, funcs, 3, "should collect 3 functions (2 funcs + 1 method)")

	names := make([]string, len(funcs))
	for i, fn := range funcs {
		names[i] = fn.Name.Name
	}
	assert.Contains(t, names, "Foo")
	assert.Contains(t, names, "Bar")
	assert.Contains(t, names, "Method")
}

func TestCollectTypes(t *testing.T) {
	src := `package test
type Foo struct{}
type Bar interface{}
const C = 1
var V int
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	require.NoError(t, err)

	types := CollectTypes(file)
	assert.Len(t, types, 2, "should collect 2 types")

	names := make([]string, len(types))
	for i, ts := range types {
		names[i] = ts.Name.Name
	}
	assert.Contains(t, names, "Foo")
	assert.Contains(t, names, "Bar")
}

func TestExtractReceiverType(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		expected string
	}{
		{
			name:     "value receiver",
			src:      `package test; type T struct{}; func (t T) M() {}`,
			expected: "T",
		},
		{
			name:     "pointer receiver",
			src:      `package test; type T struct{}; func (t *T) M() {}`,
			expected: "T",
		},
		{
			name:     "generic receiver",
			src:      `package test; type T[K any] struct{}; func (t T[K]) M() {}`,
			expected: "T",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.src, 0)
			require.NoError(t, err)

			funcs := CollectFunctions(file)
			require.Len(t, funcs, 1)
			require.True(t, IsMethod(funcs[0]))

			result := ExtractReceiverType(funcs[0].Recv.List[0].Type)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsMethod(t *testing.T) {
	src := `package test
type T struct{}
func Foo() {}
func (t T) Bar() {}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	require.NoError(t, err)

	funcs := CollectFunctions(file)
	require.Len(t, funcs, 2)

	// First is Foo (function)
	assert.False(t, IsMethod(funcs[0]), "Foo should not be a method")

	// Second is Bar (method)
	assert.True(t, IsMethod(funcs[1]), "Bar should be a method")
}

func TestGetMethodReceiverType(t *testing.T) {
	src := `package test
type T struct{}
func Foo() {}
func (t T) Bar() {}
func (t *T) Baz() {}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	require.NoError(t, err)

	funcs := CollectFunctions(file)
	require.Len(t, funcs, 3)

	assert.Equal(t, "", GetMethodReceiverType(funcs[0]), "Foo should have no receiver")
	assert.Equal(t, "T", GetMethodReceiverType(funcs[1]), "Bar should have T receiver")
	assert.Equal(t, "T", GetMethodReceiverType(funcs[2]), "Baz should have T receiver")
}

func TestCountNodes(t *testing.T) {
	src := `package test
func Foo() {
	x := 1
	y := 2
	z := x + y
}
`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	require.NoError(t, err)

	funcs := CollectFunctions(file)
	require.Len(t, funcs, 1)
	require.NotNil(t, funcs[0].Body)

	count := CountNodes(funcs[0].Body.List)
	assert.Greater(t, count, 0, "should count nodes in statement list")
}
