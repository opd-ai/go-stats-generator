package analyzer

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPackageAnalyzer(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewPackageAnalyzer(fset)

	assert.NotNil(t, analyzer)
	assert.Equal(t, fset, analyzer.fset)
	assert.NotNil(t, analyzer.packageDeps)
	assert.NotNil(t, analyzer.packageFiles)
	assert.NotNil(t, analyzer.packageFunctions)
	assert.NotNil(t, analyzer.packageTypes)
	assert.NotNil(t, analyzer.packageLines)
}

func TestAnalyzePackage(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		packageName string
		filePath    string
		wantFuncs   int
		wantTypes   int
		wantImports []string
	}{
		{
			name: "simple package with function and type",
			source: `package main

import (
	"fmt"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

type User struct {
	Name string
	Age  int
}

func main() {
	fmt.Println("Hello")
}

func processUser(u User) {
	// Process user
}`,
			packageName: "main",
			filePath:    "main.go",
			wantFuncs:   2,
			wantTypes:   1,
			wantImports: []string{"github.com/opd-ai/go-stats-generator/internal/metrics"},
		},
		{
			name: "package with no internal imports",
			source: `package utils

import (
	"fmt"
	"strings"
)

func UtilFunc() string {
	return strings.ToUpper("test")
}`,
			packageName: "utils",
			filePath:    "utils/utils.go",
			wantFuncs:   1,
			wantTypes:   0,
			wantImports: []string{}, // No internal imports
		},
		{
			name: "package with multiple types",
			source: `package models

type User struct {
	Name string
}

type Product struct {
	Title string
	Price float64
}

type Category struct {
	Name string
}`,
			packageName: "models",
			filePath:    "models/types.go",
			wantFuncs:   0,
			wantTypes:   3,
			wantImports: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			analyzer := NewPackageAnalyzer(fset)

			// Parse the source code
			file, err := parser.ParseFile(fset, tt.filePath, tt.source, parser.ParseComments)
			require.NoError(t, err)

			// Analyze the package
			err = analyzer.AnalyzePackage(file, tt.filePath)
			require.NoError(t, err)

			// Check function count
			assert.Equal(t, tt.wantFuncs, analyzer.packageFunctions[tt.packageName],
				"Function count mismatch for package %s", tt.packageName)

			// Check type count
			assert.Equal(t, tt.wantTypes, analyzer.packageTypes[tt.packageName],
				"Type count mismatch for package %s", tt.packageName)

			// Check imports (only internal ones)
			deps := analyzer.packageDeps[tt.packageName]
			assert.ElementsMatch(t, tt.wantImports, deps,
				"Import dependencies mismatch for package %s", tt.packageName)

			// Check file tracking
			files := analyzer.packageFiles[tt.packageName]
			assert.Contains(t, files, tt.filePath, "File path should be tracked")

			// Check that lines were counted
			assert.Greater(t, analyzer.packageLines[tt.packageName], 0,
				"Line count should be greater than 0")
		})
	}
}

func TestAnalyzePackageErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		filePath    string
		expectError bool
	}{
		{
			name: "file with no package name",
			source: `import "fmt"

func main() {
	fmt.Println("Hello")
}`,
			filePath:    "invalid.go",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			analyzer := NewPackageAnalyzer(fset)

			// Parse the source code (this may fail)
			file, err := parser.ParseFile(fset, tt.filePath, tt.source, parser.ParseComments)
			if err != nil && tt.expectError {
				return // Expected parsing error
			}
			require.NoError(t, err)

			// Analyze the package
			err = analyzer.AnalyzePackage(file, tt.filePath)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGenerateReport(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewPackageAnalyzer(fset)

	// Add some test data
	sources := map[string]string{
		"main.go": `package main

import (
	"fmt"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

type User struct {
	Name string
}

func main() {
	fmt.Println("Hello")
}`,
		"utils/helper.go": `package utils

import (
	"github.com/opd-ai/go-stats-generator/internal/analyzer"
)

func Helper() {
	// Helper function
}

type Config struct {
	Value string
}`,
	}

	// Parse and analyze all files
	for filePath, source := range sources {
		file, err := parser.ParseFile(fset, filePath, source, parser.ParseComments)
		require.NoError(t, err)

		err = analyzer.AnalyzePackage(file, filePath)
		require.NoError(t, err)
	}

	// Generate report
	report, err := analyzer.GenerateReport()
	require.NoError(t, err)
	assert.NotNil(t, report)

	// Check basic report structure
	assert.Equal(t, 2, report.TotalPackages)
	assert.Len(t, report.Packages, 2)
	assert.NotNil(t, report.DependencyGraph)

	// Check package details
	var mainPkg, utilsPkg *metrics.PackageMetrics
	for i := range report.Packages {
		pkg := &report.Packages[i]
		switch pkg.Name {
		case "main":
			mainPkg = pkg
		case "utils":
			utilsPkg = pkg
		}
	}

	require.NotNil(t, mainPkg, "main package should exist")
	require.NotNil(t, utilsPkg, "utils package should exist")

	// Check main package
	assert.Equal(t, "main", mainPkg.Name)
	assert.Equal(t, 1, mainPkg.Functions)
	assert.Equal(t, 1, mainPkg.Structs)
	assert.Contains(t, mainPkg.Files, "main.go")
	assert.Contains(t, mainPkg.Dependencies, "github.com/opd-ai/go-stats-generator/internal/metrics")

	// Check utils package
	assert.Equal(t, "utils", utilsPkg.Name)
	assert.Equal(t, 1, utilsPkg.Functions)
	assert.Equal(t, 1, utilsPkg.Structs)
	assert.Contains(t, utilsPkg.Files, "utils/helper.go")
	assert.Contains(t, utilsPkg.Dependencies, "github.com/opd-ai/go-stats-generator/internal/analyzer")

	// Check averages
	assert.Equal(t, 1.0, report.AverageFilesPerPackage)
	assert.Equal(t, 1.0, report.AverageFunctionsPerPackage)
	assert.Equal(t, 1.0, report.AverageTypesPerPackage)
}

func TestCohesionCalculation(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewPackageAnalyzer(fset)

	// Add test data for cohesion calculation
	analyzer.packageFunctions["high_cohesion"] = 10
	analyzer.packageTypes["high_cohesion"] = 5
	analyzer.packageFiles["high_cohesion"] = []string{"file1.go", "file2.go"}

	analyzer.packageFunctions["low_cohesion"] = 2
	analyzer.packageTypes["low_cohesion"] = 1
	analyzer.packageFiles["low_cohesion"] = []string{"f1.go", "f2.go", "f3.go", "f4.go", "f5.go"}

	highCohesion := analyzer.calculateCohesion("high_cohesion")
	lowCohesion := analyzer.calculateCohesion("low_cohesion")

	// High cohesion should have more elements per file
	assert.Greater(t, highCohesion, lowCohesion, "High cohesion package should score higher")

	// Test edge case - no files
	emptyCohesion := analyzer.calculateCohesion("nonexistent")
	assert.Equal(t, 0.0, emptyCohesion)
}

func TestCouplingCalculation(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewPackageAnalyzer(fset)

	// Add test data for coupling calculation
	analyzer.packageDeps["high_coupling"] = []string{"dep1", "dep2", "dep3", "dep4", "dep5"}
	analyzer.packageDeps["low_coupling"] = []string{"dep1"}
	analyzer.packageDeps["no_coupling"] = []string{}

	highCoupling := analyzer.calculateCoupling("high_coupling")
	lowCoupling := analyzer.calculateCoupling("low_coupling")
	noCoupling := analyzer.calculateCoupling("no_coupling")

	// More dependencies should result in higher coupling
	assert.Greater(t, highCoupling, lowCoupling, "High coupling package should score higher")
	assert.Greater(t, lowCoupling, noCoupling, "Low coupling should be higher than no coupling")
	assert.Equal(t, 0.0, noCoupling, "No dependencies should result in 0 coupling")
}

func TestCircularDependencyDetection(t *testing.T) {
	tests := []struct {
		name         string
		dependencies map[string][]string
		wantCycles   int
		wantSeverity string
	}{
		{
			name: "simple two-package cycle",
			dependencies: map[string][]string{
				"pkg1": {"pkg2"},
				"pkg2": {"pkg1"},
			},
			wantCycles:   1,
			wantSeverity: "low",
		},
		{
			name: "three-package cycle",
			dependencies: map[string][]string{
				"pkg1": {"pkg2"},
				"pkg2": {"pkg3"},
				"pkg3": {"pkg1"},
			},
			wantCycles:   1,
			wantSeverity: "medium",
		},
		{
			name: "complex multi-cycle",
			dependencies: map[string][]string{
				"pkg1": {"pkg2"},
				"pkg2": {"pkg3"},
				"pkg3": {"pkg1"},
				"pkg4": {"pkg5"},
				"pkg5": {"pkg6"},
				"pkg6": {"pkg7"},
				"pkg7": {"pkg8"},
				"pkg8": {"pkg4"},
			},
			wantCycles:   2,
			wantSeverity: "high",
		},
		{
			name: "no cycles",
			dependencies: map[string][]string{
				"pkg1": {"pkg2"},
				"pkg2": {"pkg3"},
				"pkg3": {},
			},
			wantCycles: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			analyzer := NewPackageAnalyzer(fset)
			analyzer.packageDeps = tt.dependencies

			// Initialize other required maps to avoid nil panics
			for pkg := range tt.dependencies {
				analyzer.packageFiles[pkg] = []string{pkg + ".go"}
				analyzer.packageFunctions[pkg] = 1
				analyzer.packageTypes[pkg] = 1
				analyzer.packageLines[pkg] = 10
			}

			cycles := analyzer.detectCircularDependencies()

			assert.Len(t, cycles, tt.wantCycles, "Number of cycles should match")

			if tt.wantCycles > 0 && tt.wantSeverity != "" {
				// Check that at least one cycle has the expected severity
				foundSeverity := false
				for _, cycle := range cycles {
					if cycle.Severity == tt.wantSeverity {
						foundSeverity = true
						break
					}
				}
				assert.True(t, foundSeverity, "Should find cycle with severity %s", tt.wantSeverity)
			}
		})
	}
}

func TestIsInternalPackage(t *testing.T) {
	tests := []struct {
		importPath string
		want       bool
	}{
		{"fmt", false},                                                  // stdlib
		{"strings", false},                                              // stdlib
		{"encoding/json", false},                                        // stdlib
		{"github.com/spf13/cobra", false},                               // external
		{"github.com/stretchr/testify", false},                          // external
		{"github.com/jedib0t/go-pretty", false},                         // external
		{"github.com/olekukonko/tablewriter", false},                    // external
		{"github.com/mycompany/myproject", true},                        // internal
		{"internal/analyzer", true},                                     // internal
		{"pkg/gostats", true},                                           // internal
		{"github.com/opd-ai/go-stats-generator/internal/metrics", true}, // internal
	}

	for _, tt := range tests {
		t.Run(tt.importPath, func(t *testing.T) {
			got := isInternalPackage(tt.importPath)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMergeUniqueStrings(t *testing.T) {
	tests := []struct {
		name     string
		existing []string
		new      []string
		want     []string
	}{
		{
			name:     "no duplicates",
			existing: []string{"a", "b"},
			new:      []string{"c", "d"},
			want:     []string{"a", "b", "c", "d"},
		},
		{
			name:     "with duplicates",
			existing: []string{"a", "b"},
			new:      []string{"b", "c"},
			want:     []string{"a", "b", "c"},
		},
		{
			name:     "empty existing",
			existing: []string{},
			new:      []string{"a", "b"},
			want:     []string{"a", "b"},
		},
		{
			name:     "empty new",
			existing: []string{"a", "b"},
			new:      []string{},
			want:     []string{"a", "b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeUniqueStrings(tt.existing, tt.new)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

// Integration test with real testdata
func TestPackageAnalyzerIntegration(t *testing.T) {
	fset := token.NewFileSet()
	analyzer := NewPackageAnalyzer(fset)

	// Use testdata from the project
	testFiles := []string{
		"../../../testdata/simple/calculator.go",
		"../../../testdata/simple/user.go",
		"../../../testdata/simple/interfaces.go",
	}

	for _, filePath := range testFiles {
		// Parse file
		file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if err != nil {
			// Skip files that don't exist or can't be parsed
			t.Logf("Skipping %s: %v", filePath, err)
			continue
		}

		// Analyze package
		err = analyzer.AnalyzePackage(file, filePath)
		assert.NoError(t, err)
	}

	// Generate report if we analyzed any files
	if len(analyzer.packageFiles) > 0 {
		report, err := analyzer.GenerateReport()
		assert.NoError(t, err)
		assert.NotNil(t, report)
		assert.Greater(t, report.TotalPackages, 0)
	}
}
