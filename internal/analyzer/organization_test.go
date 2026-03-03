package analyzer

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOrganizationAnalyzer(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)
	assert.NotNil(t, analyzer)
	assert.Equal(t, fset, analyzer.fset)
}

func TestDefaultOrganizationConfig(t *testing.T) {
	config := DefaultOrganizationConfig()
	assert.Equal(t, 500, config.MaxFileLines)
	assert.Equal(t, 20, config.MaxFileFunctions)
	assert.Equal(t, 5, config.MaxFileTypes)
}

func TestCountFileLines(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantCode  int
		wantComm  int
		wantBlank int
	}{
		{
			name:      "empty file",
			content:   "",
			wantCode:  0,
			wantComm:  0,
			wantBlank: 1,
		},
		{
			name:      "code only",
			content:   "package main\n\nfunc main() {}",
			wantCode:  2,
			wantComm:  0,
			wantBlank: 1,
		},
		{
			name:      "with comments",
			content:   "// comment\npackage main",
			wantCode:  1,
			wantComm:  1,
			wantBlank: 0,
		},
		{
			name:      "block comment",
			content:   "/*\nblock\n*/\npackage main",
			wantCode:  1,
			wantComm:  3,
			wantBlank: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := createTempFile(t, tt.content)
			defer os.Remove(tmpFile)

			fset := token.NewFileSet()
			analyzer := NewOrganizationAnalyzer(fset)

			lines := analyzer.countFileLines(tmpFile)
			assert.Equal(t, tt.wantCode, lines.Code, "code lines")
			assert.Equal(t, tt.wantComm, lines.Comments, "comment lines")
			assert.Equal(t, tt.wantBlank, lines.Blank, "blank lines")
		})
	}
}

func TestCountFunctions(t *testing.T) {
	tests := []struct {
		name string
		code string
		want int
	}{
		{
			name: "no functions",
			code: "package main\n\nvar x int",
			want: 0,
		},
		{
			name: "single function",
			code: "package main\n\nfunc foo() {}",
			want: 1,
		},
		{
			name: "multiple functions",
			code: "package main\n\nfunc foo() {}\nfunc bar() {}",
			want: 2,
		},
		{
			name: "with methods",
			code: "package main\n\ntype T struct{}\nfunc (t T) method() {}",
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, 0)
			require.NoError(t, err)

			analyzer := NewOrganizationAnalyzer(fset)
			count := analyzer.countFunctions(file)
			assert.Equal(t, tt.want, count)
		})
	}
}

func TestCountTypes(t *testing.T) {
	tests := []struct {
		name string
		code string
		want int
	}{
		{
			name: "no types",
			code: "package main\n\nfunc foo() {}",
			want: 0,
		},
		{
			name: "single type",
			code: "package main\n\ntype T struct{}",
			want: 1,
		},
		{
			name: "multiple types",
			code: "package main\n\ntype T1 struct{}\ntype T2 int",
			want: 2,
		},
		{
			name: "type block",
			code: "package main\n\ntype (\n\tT1 struct{}\n\tT2 int\n)",
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, 0)
			require.NoError(t, err)

			analyzer := NewOrganizationAnalyzer(fset)
			count := analyzer.countTypes(file)
			assert.Equal(t, tt.want, count)
		})
	}
}

func TestCalculateBurden(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)

	tests := []struct {
		name      string
		lines     int
		funcCount int
		typeCount int
		wantMin   float64
		wantMax   float64
	}{
		{
			name:      "no burden",
			lines:     100,
			funcCount: 5,
			typeCount: 2,
			wantMin:   0.0,
			wantMax:   0.5,
		},
		{
			name:      "high burden",
			lines:     1000,
			funcCount: 40,
			typeCount: 10,
			wantMin:   1.5,
			wantMax:   3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := createLineMetrics(tt.lines)
			burden := analyzer.calculateBurden(lines, tt.funcCount, tt.typeCount)
			assert.GreaterOrEqual(t, burden, tt.wantMin)
			assert.LessOrEqual(t, burden, tt.wantMax)
		})
	}
}

func TestIsOversized(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)
	config := DefaultOrganizationConfig()

	tests := []struct {
		name      string
		lines     int
		funcCount int
		typeCount int
		want      bool
	}{
		{
			name:      "within limits",
			lines:     100,
			funcCount: 5,
			typeCount: 2,
			want:      false,
		},
		{
			name:      "exceeds line limit",
			lines:     600,
			funcCount: 5,
			typeCount: 2,
			want:      true,
		},
		{
			name:      "exceeds function limit",
			lines:     100,
			funcCount: 25,
			typeCount: 2,
			want:      true,
		},
		{
			name:      "exceeds type limit",
			lines:     100,
			funcCount: 5,
			typeCount: 7,
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := createLineMetrics(tt.lines)
			result := analyzer.isOversized(lines, tt.funcCount, tt.typeCount, config)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestOrganizationGetSeverity(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)
	config := DefaultOrganizationConfig()

	tests := []struct {
		name      string
		lines     int
		funcCount int
		typeCount int
		want      string
	}{
		{
			name:      "medium severity",
			lines:     600,
			funcCount: 10,
			typeCount: 3,
			want:      "medium",
		},
		{
			name:      "high severity",
			lines:     1200,
			funcCount: 10,
			typeCount: 3,
			want:      "high",
		},
		{
			name:      "critical severity",
			lines:     1200,
			funcCount: 50,
			typeCount: 3,
			want:      "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := createLineMetrics(tt.lines)
			severity := analyzer.getSeverity(lines, tt.funcCount, tt.typeCount, config)
			assert.Equal(t, tt.want, severity)
		})
	}
}

func TestOrganizationGetSuggestions(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)
	config := DefaultOrganizationConfig()

	tests := []struct {
		name      string
		lines     int
		funcCount int
		typeCount int
		wantCount int
	}{
		{
			name:      "no suggestions",
			lines:     100,
			funcCount: 5,
			typeCount: 2,
			wantCount: 0,
		},
		{
			name:      "all suggestions",
			lines:     600,
			funcCount: 25,
			typeCount: 7,
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := createLineMetrics(tt.lines)
			suggestions := analyzer.getSuggestions(lines, tt.funcCount, tt.typeCount, config)
			assert.Len(t, suggestions, tt.wantCount)
		})
	}
}

func TestAnalyzeFileSizes(t *testing.T) {
	code := `package main

import "fmt"

type User struct {
	Name string
	Age  int
}

func main() {
	fmt.Println("hello")
}

func foo() {}
func bar() {}
`
	tmpFile := createTempFile(t, code)
	defer os.Remove(tmpFile)

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, tmpFile, nil, 0)
	require.NoError(t, err)

	analyzer := NewOrganizationAnalyzer(fset)
	config := DefaultOrganizationConfig()

	result, err := analyzer.AnalyzeFileSizes(file, tmpFile, config)
	require.NoError(t, err)
	assert.Nil(t, result, "small file should not be flagged")
}

func TestAnalyzeFileSizes_Oversized(t *testing.T) {
	// Generate large file
	var builder strings.Builder
	builder.WriteString("package main\n\n")

	for i := 0; i < 30; i++ {
		builder.WriteString("func f")
		builder.WriteString(string(rune('0' + i%10)))
		builder.WriteString("() {}\n")
	}

	tmpFile := createTempFile(t, builder.String())
	defer os.Remove(tmpFile)

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, tmpFile, nil, 0)
	require.NoError(t, err)

	analyzer := NewOrganizationAnalyzer(fset)
	config := DefaultOrganizationConfig()

	result, err := analyzer.AnalyzeFileSizes(file, tmpFile, config)
	require.NoError(t, err)
	require.NotNil(t, result, "oversized file should be flagged")
	assert.Equal(t, tmpFile, result.File)
	assert.Greater(t, result.FunctionCount, config.MaxFileFunctions)
}

// Helper functions

func createTempFile(t *testing.T, content string) string {
	t.Helper()
	tmpFile := filepath.Join(t.TempDir(), "test.go")
	err := os.WriteFile(tmpFile, []byte(content), 0o644)
	require.NoError(t, err)
	return tmpFile
}

func createLineMetrics(total int) metrics.LineMetrics {
	return metrics.LineMetrics{
		Total:    total,
		Code:     total * 7 / 10,
		Comments: total * 2 / 10,
		Blank:    total * 1 / 10,
	}
}
