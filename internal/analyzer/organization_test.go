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
		want      metrics.SeverityLevel
	}{
		{
			name:      "medium severity",
			lines:     600,
			funcCount: 10,
			typeCount: 3,
			want:      metrics.SeverityLevelWarning,
		},
		{
			name:      "high severity",
			lines:     1200,
			funcCount: 10,
			typeCount: 3,
			want:      metrics.SeverityLevelViolation,
		},
		{
			name:      "critical severity",
			lines:     1200,
			funcCount: 50,
			typeCount: 3,
			want:      metrics.SeverityLevelCritical,
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

// Tests for package size analysis

func TestAnalyzePackageSizes(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)
	config := DefaultOrganizationConfig()

	tests := []struct {
		name      string
		packages  map[string]*PackageInfo
		wantCount int
	}{
		{
			name: "no oversized packages",
			packages: map[string]*PackageInfo{
				"small": {
					Name:            "small",
					Files:           []string{"a.go", "b.go"},
					ExportedSymbols: 10,
					TotalFunctions:  5,
					CohesionScore:   0.8,
				},
			},
			wantCount: 0,
		},
		{
			name: "oversized by file count",
			packages: map[string]*PackageInfo{
				"large": {
					Name:            "large",
					Files:           make([]string, 25),
					ExportedSymbols: 10,
					TotalFunctions:  30,
					CohesionScore:   0.8,
				},
			},
			wantCount: 1,
		},
		{
			name: "oversized by exported symbols",
			packages: map[string]*PackageInfo{
				"exports": {
					Name:            "exports",
					Files:           []string{"a.go"},
					ExportedSymbols: 60,
					TotalFunctions:  70,
					CohesionScore:   0.8,
				},
			},
			wantCount: 1,
		},
		{
			name: "mega package",
			packages: map[string]*PackageInfo{
				"mega": {
					Name:            "mega",
					Files:           []string{"a.go"},
					ExportedSymbols: 35,
					TotalFunctions:  50,
					CohesionScore:   0.3,
				},
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := analyzer.AnalyzePackageSizes(tt.packages, config)
			assert.Len(t, results, tt.wantCount)

			if tt.wantCount > 0 {
				result := results[0]
				assert.NotEmpty(t, result.Package)
				assert.NotEmpty(t, result.Severity)
				if tt.name == "mega package" {
					assert.True(t, result.IsMegaPackage)
				}
			}
		})
	}
}

func TestIsPackageOversized(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)
	config := DefaultOrganizationConfig()

	tests := []struct {
		name string
		pkg  *PackageInfo
		want bool
	}{
		{
			name: "within limits",
			pkg: &PackageInfo{
				Files:           []string{"a.go"},
				ExportedSymbols: 10,
			},
			want: false,
		},
		{
			name: "too many files",
			pkg: &PackageInfo{
				Files:           make([]string, 25),
				ExportedSymbols: 10,
			},
			want: true,
		},
		{
			name: "too many exports",
			pkg: &PackageInfo{
				Files:           []string{"a.go"},
				ExportedSymbols: 60,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.isPackageOversized(tt.pkg, config)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestIsMegaPackage(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)

	tests := []struct {
		name string
		pkg  *PackageInfo
		want bool
	}{
		{
			name: "not mega package",
			pkg: &PackageInfo{
				ExportedSymbols: 25,
				CohesionScore:   0.6,
			},
			want: false,
		},
		{
			name: "low cohesion but few exports",
			pkg: &PackageInfo{
				ExportedSymbols: 25,
				CohesionScore:   0.3,
			},
			want: false,
		},
		{
			name: "many exports but good cohesion",
			pkg: &PackageInfo{
				ExportedSymbols: 40,
				CohesionScore:   0.7,
			},
			want: false,
		},
		{
			name: "mega package",
			pkg: &PackageInfo{
				ExportedSymbols: 40,
				CohesionScore:   0.3,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.isMegaPackage(tt.pkg)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGetPackageSeverity(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)
	config := DefaultOrganizationConfig()

	tests := []struct {
		name string
		pkg  *PackageInfo
		want metrics.SeverityLevel
	}{
		{
			name: "medium severity",
			pkg: &PackageInfo{
				Files:           make([]string, 25),
				ExportedSymbols: 10,
				CohesionScore:   0.8,
			},
			want: metrics.SeverityLevelWarning,
		},
		{
			name: "high severity",
			pkg: &PackageInfo{
				Files:           make([]string, 45),
				ExportedSymbols: 10,
				CohesionScore:   0.8,
			},
			want: metrics.SeverityLevelViolation,
		},
		{
			name: "critical severity - violations",
			pkg: &PackageInfo{
				Files:           make([]string, 45),
				ExportedSymbols: 110,
				CohesionScore:   0.8,
			},
			want: metrics.SeverityLevelCritical,
		},
		{
			name: "critical severity - mega package",
			pkg: &PackageInfo{
				Files:           []string{"a.go"},
				ExportedSymbols: 40,
				CohesionScore:   0.3,
			},
			want: metrics.SeverityLevelCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.getPackageSeverity(tt.pkg, config)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGetPackageSuggestions(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)
	config := DefaultOrganizationConfig()

	tests := []struct {
		name      string
		pkg       *PackageInfo
		wantCount int
		wantMega  bool
	}{
		{
			name: "no suggestions",
			pkg: &PackageInfo{
				Files:           []string{"a.go"},
				ExportedSymbols: 10,
				CohesionScore:   0.8,
			},
			wantCount: 0,
			wantMega:  false,
		},
		{
			name: "file count suggestion",
			pkg: &PackageInfo{
				Files:           make([]string, 25),
				ExportedSymbols: 10,
				CohesionScore:   0.8,
			},
			wantCount: 1,
			wantMega:  false,
		},
		{
			name: "mega package suggestion",
			pkg: &PackageInfo{
				Files:           make([]string, 25),
				ExportedSymbols: 60,
				CohesionScore:   0.3,
			},
			wantCount: 3,
			wantMega:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := analyzer.getPackageSuggestions(tt.pkg, config)
			assert.Len(t, suggestions, tt.wantCount)
			if tt.wantMega {
				found := false
				for _, s := range suggestions {
					if strings.Contains(s, "Mega-package") {
						found = true
						break
					}
				}
				assert.True(t, found, "should have mega package suggestion")
			}
		})
	}
}

// Tests for directory depth analysis

func TestAnalyzeDirectoryDepth(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)
	config := DefaultOrganizationConfig()

	tests := []struct {
		name      string
		paths     []string
		rootPath  string
		wantCount int
	}{
		{
			name: "no deep directories",
			paths: []string{
				"/project/file1.go",
				"/project/pkg/file2.go",
			},
			rootPath:  "/project",
			wantCount: 0,
		},
		{
			name: "deep directory",
			paths: []string{
				"/project/a/b/c/d/e/f/file.go",
			},
			rootPath:  "/project",
			wantCount: 1,
		},
		{
			name: "multiple files in deep directory",
			paths: []string{
				"/project/a/b/c/d/e/f/file1.go",
				"/project/a/b/c/d/e/f/file2.go",
			},
			rootPath:  "/project",
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := analyzer.AnalyzeDirectoryDepth(tt.paths, tt.rootPath, config)
			assert.Len(t, results, tt.wantCount)

			if tt.wantCount > 0 {
				result := results[0]
				assert.NotEmpty(t, result.Path)
				assert.Greater(t, result.Depth, config.MaxDirectoryDepth)
				assert.NotEmpty(t, result.Severity)
				assert.NotEmpty(t, result.Suggestion)
			}
		})
	}
}

func TestCalculateDepth(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)

	tests := []struct {
		name     string
		dirPath  string
		rootPath string
		want     int
	}{
		{
			name:     "root level",
			dirPath:  "/project",
			rootPath: "/project",
			want:     0,
		},
		{
			name:     "one level",
			dirPath:  "/project/pkg",
			rootPath: "/project",
			want:     1,
		},
		{
			name:     "multiple levels",
			dirPath:  "/project/a/b/c",
			rootPath: "/project",
			want:     3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.calculateDepth(tt.dirPath, tt.rootPath)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGetDirectoryPath(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "file in root",
			path: "file.go",
			want: ".",
		},
		{
			name: "file in directory",
			path: "/project/pkg/file.go",
			want: "/project/pkg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.getDirectoryPath(tt.path)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGetDepthSeverity(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)
	config := DefaultOrganizationConfig()

	tests := []struct {
		name  string
		depth int
		want  metrics.SeverityLevel
	}{
		{
			name:  "medium severity",
			depth: 6,
			want:  metrics.SeverityLevelWarning,
		},
		{
			name:  "high severity",
			depth: 8,
			want:  metrics.SeverityLevelViolation,
		},
		{
			name:  "critical severity",
			depth: 11,
			want:  metrics.SeverityLevelCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.getDepthSeverity(tt.depth, config)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGetDepthSuggestion(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)
	config := DefaultOrganizationConfig()

	tests := []struct {
		name  string
		depth int
		want  string
	}{
		{
			name:  "moderate depth",
			depth: 6,
			want:  "Consider flattening directory structure",
		},
		{
			name:  "excessive depth",
			depth: 10,
			want:  "Restructure directory hierarchy - nesting is excessively deep",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.getDepthSuggestion(tt.depth, config)
			assert.Equal(t, tt.want, result)
		})
	}
}

// Tests for import graph analysis

func TestAnalyzeImportGraph(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)
	config := DefaultOrganizationConfig()

	tests := []struct {
		name           string
		graphData      *ImportGraphData
		wantViolations int
		wantFanIn      int
		wantFanOut     int
	}{
		{
			name:           "nil graph data",
			graphData:      nil,
			wantViolations: 0,
			wantFanIn:      0,
			wantFanOut:     0,
		},
		{
			name: "no violations",
			graphData: &ImportGraphData{
				FileImports:    map[string]int{"file1.go": 5},
				PackageFanIn:   map[string][]string{},
				PackageFanOut:  map[string][]string{},
				FilePackageMap: map[string]string{"file1.go": "pkg"},
			},
			wantViolations: 0,
			wantFanIn:      0,
			wantFanOut:     0,
		},
		{
			name: "file import violations",
			graphData: &ImportGraphData{
				FileImports:    map[string]int{"file1.go": 20},
				PackageFanIn:   map[string][]string{},
				PackageFanOut:  map[string][]string{},
				FilePackageMap: map[string]string{"file1.go": "pkg"},
			},
			wantViolations: 1,
			wantFanIn:      0,
			wantFanOut:     0,
		},
		{
			name: "high fan-in",
			graphData: &ImportGraphData{
				FileImports: map[string]int{},
				PackageFanIn: map[string][]string{
					"hub": {"pkg1", "pkg2", "pkg3", "pkg4"},
				},
				PackageFanOut:  map[string][]string{},
				FilePackageMap: map[string]string{},
			},
			wantViolations: 0,
			wantFanIn:      1,
			wantFanOut:     0,
		},
		{
			name: "high fan-out",
			graphData: &ImportGraphData{
				FileImports:  map[string]int{},
				PackageFanIn: map[string][]string{},
				PackageFanOut: map[string][]string{
					"authority": {"dep1", "dep2", "dep3", "dep4", "dep5", "dep6"},
				},
				FilePackageMap: map[string]string{},
			},
			wantViolations: 0,
			wantFanIn:      0,
			wantFanOut:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.AnalyzeImportGraph(tt.graphData, config)
			require.NoError(t, err)

			if tt.graphData == nil {
				assert.Nil(t, result)
				return
			}

			assert.NotNil(t, result)
			assert.Len(t, result.FileImportViolations, tt.wantViolations)
			assert.Len(t, result.HighFanInPackages, tt.wantFanIn)
			assert.Len(t, result.HighFanOutPackages, tt.wantFanOut)
		})
	}
}

func TestFindFileImportViolations(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)
	config := DefaultOrganizationConfig()

	graphData := &ImportGraphData{
		FileImports: map[string]int{
			"normal.go":   10,
			"high.go":     20,
			"critical.go": 35,
		},
		FilePackageMap: map[string]string{
			"normal.go":   "pkg",
			"high.go":     "pkg",
			"critical.go": "pkg",
		},
	}

	violations := analyzer.findFileImportViolations(graphData, config)
	assert.Len(t, violations, 2)

	for _, v := range violations {
		assert.Greater(t, v.ImportCount, config.MaxFileImports)
		assert.NotEmpty(t, v.Severity)
		assert.NotEmpty(t, v.Suggestion)
	}
}

func TestGetImportSeverity(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)
	config := DefaultOrganizationConfig()

	tests := []struct {
		name  string
		count int
		want  string
	}{
		{
			name:  "medium severity",
			count: 18,
			want:  string(metrics.SeverityLevelWarning),
		},
		{
			name:  "high severity",
			count: 25,
			want:  string(metrics.SeverityLevelViolation),
		},
		{
			name:  "critical severity",
			count: 35,
			want:  string(metrics.SeverityLevelCritical),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.getImportSeverity(tt.count, config)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGetImportSuggestion(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)
	config := DefaultOrganizationConfig()

	tests := []struct {
		name  string
		count int
		want  string
	}{
		{
			name:  "moderate excess",
			count: 18,
			want:  "Too many imports - consider refactoring to reduce dependencies",
		},
		{
			name:  "excessive imports",
			count: 30,
			want:  "Excessive imports - split into multiple focused files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.getImportSuggestion(tt.count, config)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestFindHighFanIn(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)

	graphData := &ImportGraphData{
		PackageFanIn: map[string][]string{
			"low":      {"dep1", "dep2"},
			"medium":   {"dep1", "dep2", "dep3"},
			"high":     {"dep1", "dep2", "dep3", "dep4", "dep5"},
			"critical": {"d1", "d2", "d3", "d4", "d5", "d6", "d7", "d8", "d9", "d10"},
		},
	}

	results := analyzer.findHighFanIn(graphData)
	assert.GreaterOrEqual(t, len(results), 2)

	for _, r := range results {
		assert.GreaterOrEqual(t, r.FanIn, 3)
		assert.NotEmpty(t, r.RiskLevel)
		assert.NotEmpty(t, r.Suggestion)
		if r.FanIn >= 5 {
			assert.True(t, r.IsBottleneck)
		}
	}
}

func TestGetFanInRisk(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)

	tests := []struct {
		name  string
		fanIn int
		want  string
	}{
		{
			name:  "medium risk",
			fanIn: 3,
			want:  string(metrics.SeverityLevelWarning),
		},
		{
			name:  "high risk",
			fanIn: 6,
			want:  string(metrics.SeverityLevelViolation),
		},
		{
			name:  "critical risk",
			fanIn: 12,
			want:  string(metrics.SeverityLevelCritical),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.getFanInRisk(tt.fanIn)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGetFanInSuggestion(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)

	tests := []struct {
		name  string
		fanIn int
		want  string
	}{
		{
			name:  "moderate fan-in",
			fanIn: 4,
			want:  "Hub package - ensure stability and comprehensive testing",
		},
		{
			name:  "critical fan-in",
			fanIn: 12,
			want:  "Critical bottleneck - changes to this package affect many dependents",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.getFanInSuggestion(tt.fanIn)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestFindHighFanOut(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)

	graphData := &ImportGraphData{
		PackageFanIn: map[string][]string{
			"low":    {},
			"medium": {"dep1"},
		},
		PackageFanOut: map[string][]string{
			"low":    {"dep1", "dep2"},
			"medium": {"dep1", "dep2", "dep3", "dep4", "dep5"},
			"high":   {"d1", "d2", "d3", "d4", "d5", "d6", "d7"},
		},
	}

	results := analyzer.findHighFanOut(graphData)
	assert.GreaterOrEqual(t, len(results), 1)

	for _, r := range results {
		assert.GreaterOrEqual(t, r.FanOut, 5)
		assert.NotEmpty(t, r.CouplingRisk)
		assert.NotEmpty(t, r.Suggestion)
		assert.GreaterOrEqual(t, r.Instability, 0.0)
		assert.LessOrEqual(t, r.Instability, 1.0)
	}
}

func TestCalculateInstability(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)

	tests := []struct {
		name    string
		fanIn   int
		fanOut  int
		want    float64
		wantMin float64
		wantMax float64
	}{
		{
			name:    "no dependencies",
			fanIn:   0,
			fanOut:  0,
			want:    0.0,
			wantMin: 0.0,
			wantMax: 0.0,
		},
		{
			name:    "only incoming",
			fanIn:   5,
			fanOut:  0,
			want:    0.0,
			wantMin: 0.0,
			wantMax: 0.0,
		},
		{
			name:    "only outgoing",
			fanIn:   0,
			fanOut:  5,
			want:    1.0,
			wantMin: 1.0,
			wantMax: 1.0,
		},
		{
			name:    "balanced",
			fanIn:   5,
			fanOut:  5,
			wantMin: 0.45,
			wantMax: 0.55,
		},
		{
			name:    "more outgoing",
			fanIn:   2,
			fanOut:  8,
			wantMin: 0.75,
			wantMax: 0.85,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.calculateInstability(tt.fanIn, tt.fanOut)
			if tt.want > 0 {
				assert.Equal(t, tt.want, result)
			} else {
				assert.GreaterOrEqual(t, result, tt.wantMin)
				assert.LessOrEqual(t, result, tt.wantMax)
			}
		})
	}
}

func TestGetCouplingRisk(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)

	tests := []struct {
		name        string
		fanOut      int
		instability float64
		want        string
	}{
		{
			name:        "medium risk",
			fanOut:      5,
			instability: 0.5,
			want:        string(metrics.SeverityLevelWarning),
		},
		{
			name:        "high risk - high instability",
			fanOut:      6,
			instability: 0.7,
			want:        string(metrics.SeverityLevelViolation),
		},
		{
			name:        "high risk - many deps",
			fanOut:      8,
			instability: 0.5,
			want:        string(metrics.SeverityLevelViolation),
		},
		{
			name:        "critical risk",
			fanOut:      12,
			instability: 0.8,
			want:        string(metrics.SeverityLevelCritical),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.getCouplingRisk(tt.fanOut, tt.instability)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestGetFanOutSuggestion(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)

	tests := []struct {
		name   string
		fanOut int
		want   string
	}{
		{
			name:   "moderate fan-out",
			fanOut: 6,
			want:   "High coupling - consider dependency injection or interfaces",
		},
		{
			name:   "excessive fan-out",
			fanOut: 12,
			want:   "Excessive coupling - refactor to reduce dependencies",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.getFanOutSuggestion(tt.fanOut)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestCalculateAvgInstability(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)

	tests := []struct {
		name       string
		graphData  *ImportGraphData
		wantMin    float64
		wantMax    float64
		expectZero bool
	}{
		{
			name: "no packages",
			graphData: &ImportGraphData{
				PackageFanIn:  map[string][]string{},
				PackageFanOut: map[string][]string{},
			},
			expectZero: true,
		},
		{
			name: "mixed instability",
			graphData: &ImportGraphData{
				PackageFanIn: map[string][]string{
					"pkg1": {"dep1"},
					"pkg2": {},
				},
				PackageFanOut: map[string][]string{
					"pkg1": {"dep1"},
					"pkg2": {"dep1", "dep2", "dep3"},
				},
			},
			wantMin: 0.4,
			wantMax: 0.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.calculateAvgInstability(tt.graphData)
			if tt.expectZero {
				assert.Equal(t, 0.0, result)
			} else {
				assert.GreaterOrEqual(t, result, tt.wantMin)
				assert.LessOrEqual(t, result, tt.wantMax)
			}
		})
	}
}

func TestDefaultOrganizationConfig_ImportsField(t *testing.T) {
	config := DefaultOrganizationConfig()
	assert.Equal(t, 15, config.MaxFileImports, "default max file imports should be 15")
}

// TestCountFileLines_EdgeCases tests edge cases for line counting
func TestCountFileLines_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantCode  int
		wantComm  int
		wantBlank int
	}{
		{
			name:      "mixed comment on code line",
			content:   "x := 5 // inline comment",
			wantCode:  1,
			wantComm:  0,
			wantBlank: 0,
		},
		{
			name:      "block comment with code on same line",
			content:   "/* comment */ package main",
			wantCode:  0,
			wantComm:  1,
			wantBlank: 0,
		},
		{
			name:      "multi-line block comment",
			content:   "/*\nline1\nline2\n*/\npackage main",
			wantCode:  1,
			wantComm:  4,
			wantBlank: 0,
		},
		{
			name:      "nested block comment markers",
			content:   "/* start /* nested */ end */\ncode",
			wantCode:  1,
			wantComm:  1,
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

// TestCountFileLines_ReadError tests file read error handling
func TestCountFileLines_ReadError(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)

	lines := analyzer.countFileLines("/nonexistent/path/file.go")
	assert.Equal(t, 0, lines.Code)
	assert.Equal(t, 0, lines.Comments)
	assert.Equal(t, 0, lines.Blank)
	assert.Equal(t, 0, lines.Total)
}

// TestAnalyzeLinesInFile_AllLineTypes tests all line classification types
func TestAnalyzeLinesInFile_AllLineTypes(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)

	lines := []string{
		"",
		"// single line comment",
		"x := 5",
		"y := 10 // inline comment",
		"/*",
		"block comment line 1",
		"block comment line 2",
		"*/",
		"z := 15",
	}

	result := analyzer.analyzeLinesInFile(lines)
	assert.Greater(t, result.Code, 0, "should have code lines")
	assert.Greater(t, result.Comments, 0, "should have comment lines")
	assert.Greater(t, result.Blank, 0, "should have blank lines")
	assert.Equal(t, result.Total, result.Code+result.Comments+result.Blank)
}

// BenchmarkAnalyzeFileSizes benchmarks file size analysis
func BenchmarkAnalyzeFileSizes(b *testing.B) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)
	config := DefaultOrganizationConfig()

	code := `package main

import "fmt"

type User struct {
	ID   int
	Name string
}

func (u *User) String() string {
	return fmt.Sprintf("User %d: %s", u.ID, u.Name)
}

func main() {
	u := &User{ID: 1, Name: "Alice"}
	fmt.Println(u)
}
`
	tmpFile := filepath.Join(b.TempDir(), "test.go")
	err := os.WriteFile(tmpFile, []byte(code), 0o644)
	require.NoError(b, err)

	file, err := parser.ParseFile(fset, tmpFile, nil, 0)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.AnalyzeFileSizes(file, tmpFile, config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkAnalyzePackageSizes benchmarks package size analysis
func BenchmarkAnalyzePackageSizes(b *testing.B) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)
	config := DefaultOrganizationConfig()

	packageData := map[string]*PackageInfo{
		"main": {
			Name:            "main",
			Files:           []string{"file1.go", "file2.go", "file3.go"},
			ExportedSymbols: 15,
			TotalFunctions:  20,
			CohesionScore:   0.8,
		},
		"util": {
			Name:            "util",
			Files:           []string{"util1.go", "util2.go"},
			ExportedSymbols: 10,
			TotalFunctions:  12,
			CohesionScore:   0.9,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.AnalyzePackageSizes(packageData, config)
	}
}

// BenchmarkAnalyzeDirectoryDepth benchmarks directory depth analysis
func BenchmarkAnalyzeDirectoryDepth(b *testing.B) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)
	config := DefaultOrganizationConfig()

	files := []string{
		"/project/cmd/app/main.go",
		"/project/internal/service/user/handler.go",
		"/project/internal/service/user/repository.go",
		"/project/pkg/util/string.go",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.AnalyzeDirectoryDepth(files, "/project", config)
	}
}

// BenchmarkAnalyzeImportGraph benchmarks import graph analysis
func BenchmarkAnalyzeImportGraph(b *testing.B) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)
	config := DefaultOrganizationConfig()

	graphData := &ImportGraphData{
		PackageFanIn: map[string][]string{
			"main":    {"service", "util"},
			"service": {"repository", "model"},
			"util":    {},
		},
		PackageFanOut: map[string][]string{
			"main":    {"service", "util"},
			"service": {"repository", "model", "util"},
			"util":    {},
		},
		FileImports: map[string]int{
			"/project/cmd/main.go":            3,
			"/project/internal/service.go":    3,
			"/project/internal/repository.go": 1,
		},
		FilePackageMap: map[string]string{
			"/project/cmd/main.go":            "main",
			"/project/internal/service.go":    "service",
			"/project/internal/repository.go": "repository",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.AnalyzeImportGraph(graphData, config)
	}
}

// TestIntegration_RealTestdata tests against actual testdata files
func TestIntegration_RealTestdata(t *testing.T) {
	testdataPath := filepath.Join("testdata", "simple")
	if _, err := os.Stat(testdataPath); os.IsNotExist(err) {
		t.Skip("testdata/simple not found")
	}

	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)
	config := DefaultOrganizationConfig()

	files, err := filepath.Glob(filepath.Join(testdataPath, "*.go"))
	require.NoError(t, err)
	require.NotEmpty(t, files, "should find test files")

	for _, filePath := range files {
		t.Run(filepath.Base(filePath), func(t *testing.T) {
			file, err := parser.ParseFile(fset, filePath, nil, 0)
			if err != nil {
				t.Logf("Skipping unparseable file: %v", err)
				return
			}

			result, err := analyzer.AnalyzeFileSizes(file, filePath, config)
			require.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}

// TestIntegration_EmptyPackage tests edge case of empty package
func TestIntegration_EmptyPackage(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)
	config := DefaultOrganizationConfig()

	packageData := map[string]*PackageInfo{}
	result := analyzer.AnalyzePackageSizes(packageData, config)

	assert.Equal(t, 0, len(result), "empty package data should return empty result")
}

// TestIntegration_SingleFilePackage tests single-file package
func TestIntegration_SingleFilePackage(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)
	config := DefaultOrganizationConfig()

	packageData := map[string]*PackageInfo{
		"main": {
			Name:            "main",
			Files:           []string{"main.go"},
			ExportedSymbols: 1,
			TotalFunctions:  1,
			CohesionScore:   1.0,
		},
	}

	result := analyzer.AnalyzePackageSizes(packageData, config)
	assert.Equal(t, 0, len(result), "single file package should not be flagged")
}

// TestGetSeverity_AllPaths tests all severity calculation paths
func TestGetSeverity_AllPaths(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)
	config := DefaultOrganizationConfig()

	tests := []struct {
		name      string
		lines     int
		funcCount int
		typeCount int
		want      metrics.SeverityLevel
	}{
		{
			name:      "no violations - warning",
			lines:     300,
			funcCount: 5,
			typeCount: 2,
			want:      metrics.SeverityLevelWarning,
		},
		{
			name:      "one critical violation - violation",
			lines:     1200,
			funcCount: 5,
			typeCount: 2,
			want:      metrics.SeverityLevelViolation,
		},
		{
			name:      "one critical func violation - violation",
			lines:     300,
			funcCount: 50,
			typeCount: 2,
			want:      metrics.SeverityLevelViolation,
		},
		{
			name:      "one critical type violation - violation",
			lines:     300,
			funcCount: 5,
			typeCount: 15,
			want:      metrics.SeverityLevelViolation,
		},
		{
			name:      "two critical violations - critical",
			lines:     1200,
			funcCount: 50,
			typeCount: 2,
			want:      metrics.SeverityLevelCritical,
		},
		{
			name:      "three critical violations - critical",
			lines:     1200,
			funcCount: 50,
			typeCount: 15,
			want:      metrics.SeverityLevelCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lineMetrics := createLineMetrics(tt.lines)
			result := analyzer.getSeverity(lineMetrics, tt.funcCount, tt.typeCount, config)
			assert.Equal(t, tt.want, result)
		})
	}
}

// TestCalculateAvgInstability_EdgeCases tests edge cases for instability calculation
func TestCalculateAvgInstability_EdgeCases(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewOrganizationAnalyzer(fset)

	tests := []struct {
		name      string
		graphData *ImportGraphData
		want      float64
	}{
		{
			name: "empty fan-out",
			graphData: &ImportGraphData{
				PackageFanOut: map[string][]string{},
				PackageFanIn:  map[string][]string{},
			},
			want: 0.0,
		},
		{
			name: "single package no deps",
			graphData: &ImportGraphData{
				PackageFanOut: map[string][]string{
					"pkg1": {},
				},
				PackageFanIn: map[string][]string{
					"pkg1": {},
				},
			},
			want: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.calculateAvgInstability(tt.graphData)
			assert.Equal(t, tt.want, result)
		})
	}
}
