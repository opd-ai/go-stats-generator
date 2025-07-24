package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"
	"strings"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// PackageAnalyzer analyzes package-level metrics including dependencies,
// cohesion, coupling, and circular dependency detection.
// It provides architectural insights for large Go codebases.
type PackageAnalyzer struct {
	fset             *token.FileSet
	packageDeps      map[string][]string // package -> imported packages
	packageFiles     map[string][]string // package -> source files
	packageFunctions map[string]int      // package -> function count
	packageTypes     map[string]int      // package -> type count
	packageLines     map[string]int      // package -> total lines of code
}

// NewPackageAnalyzer creates a new package analyzer
func NewPackageAnalyzer(fset *token.FileSet) *PackageAnalyzer {
	return &PackageAnalyzer{
		fset:             fset,
		packageDeps:      make(map[string][]string),
		packageFiles:     make(map[string][]string),
		packageFunctions: make(map[string]int),
		packageTypes:     make(map[string]int),
		packageLines:     make(map[string]int),
	}
}

// AnalyzePackage analyzes a single package file and collects metrics
func (pa *PackageAnalyzer) AnalyzePackage(file *ast.File, filePath string) error {
	if file.Name == nil {
		return fmt.Errorf("file has no package name: %s", filePath)
	}

	pkgName := file.Name.Name

	// Track file for this package
	pa.packageFiles[pkgName] = append(pa.packageFiles[pkgName], filePath)

	// Analyze imports (dependencies)
	var imports []string
	for _, imp := range file.Imports {
		if imp.Path != nil {
			// Remove quotes and get clean import path
			importPath := strings.Trim(imp.Path.Value, `"`)
			// Skip stdlib and external packages, focus on internal dependencies
			if isInternalPackage(importPath) {
				imports = append(imports, importPath)
			}
		}
	}

	// Merge imports with existing dependencies
	existing := pa.packageDeps[pkgName]
	pa.packageDeps[pkgName] = mergeUniqueStrings(existing, imports)

	// Count functions and types
	functionCount := 0
	typeCount := 0

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			functionCount++
		case *ast.GenDecl:
			if d.Tok == token.TYPE {
				typeCount += len(d.Specs)
			}
		}
	}

	pa.packageFunctions[pkgName] += functionCount
	pa.packageTypes[pkgName] += typeCount

	// Calculate lines of code for this file
	start := pa.fset.Position(file.Pos())
	end := pa.fset.Position(file.End())
	linesInFile := end.Line - start.Line + 1
	pa.packageLines[pkgName] += linesInFile

	return nil
}

// GenerateReport generates comprehensive package metrics report
func (pa *PackageAnalyzer) GenerateReport() (*metrics.PackageReport, error) {
	packages := make([]metrics.PackageMetrics, 0, len(pa.packageFiles))

	for pkgName := range pa.packageFiles {
		pkg := metrics.PackageMetrics{
			Name:         pkgName,
			Path:         pkgName, // Package path same as name for now
			Files:        pa.packageFiles[pkgName],
			Functions:    pa.packageFunctions[pkgName],
			Structs:      pa.packageTypes[pkgName], // Approximation for now
			Interfaces:   0,                        // Will be enhanced later
			Dependencies: pa.packageDeps[pkgName],
			Lines: metrics.LineMetrics{
				Total: pa.packageLines[pkgName],
				Code:  pa.packageLines[pkgName], // Approximation for now
			},
		}

		// Calculate metrics
		pkg.CohesionScore = pa.calculateCohesion(pkgName)
		pkg.CouplingScore = pa.calculateCoupling(pkgName)

		packages = append(packages, pkg)
	}

	// Sort packages by name for consistent output
	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Name < packages[j].Name
	})

	// Detect circular dependencies
	circularDeps := pa.detectCircularDependencies()

	report := &metrics.PackageReport{
		Packages:             packages,
		TotalPackages:        len(packages),
		CircularDependencies: circularDeps,
		DependencyGraph:      pa.buildDependencyGraph(),
	}

	// Calculate summary statistics
	report.AverageFilesPerPackage = pa.calculateAverageFiles()
	report.AverageFunctionsPerPackage = pa.calculateAverageInt(pa.packageFunctions)
	report.AverageTypesPerPackage = pa.calculateAverageInt(pa.packageTypes)

	return report, nil
}

// calculateCohesion measures how well elements within a package work together
// Higher scores indicate better cohesion (elements belong together)
func (pa *PackageAnalyzer) calculateCohesion(pkgName string) float64 {
	functionCount := pa.packageFunctions[pkgName]
	typeCount := pa.packageTypes[pkgName]
	fileCount := len(pa.packageFiles[pkgName])

	// Cohesion heuristic: fewer files with more functions/types = higher cohesion
	if fileCount == 0 {
		return 0.0
	}

	elementsPerFile := float64(functionCount+typeCount) / float64(fileCount)

	// Normalize to 0-10 scale
	cohesion := elementsPerFile / 5.0 // Assuming 5 elements per file is average
	if cohesion > 10.0 {
		cohesion = 10.0
	}

	return cohesion
}

// calculateCoupling measures dependencies between packages
// Lower scores indicate better design (fewer dependencies)
func (pa *PackageAnalyzer) calculateCoupling(pkgName string) float64 {
	depCount := len(pa.packageDeps[pkgName])

	// Simple coupling metric: number of dependencies
	// Normalize to 0-10 scale (0 = no deps, 10 = many deps)
	coupling := float64(depCount) / 2.0 // Assuming 2 deps is average
	if coupling > 10.0 {
		coupling = 10.0
	}

	return coupling
}

// calculateComplexity combines multiple factors into an overall complexity score
func (pa *PackageAnalyzer) calculateComplexity(pkgName string) float64 {
	functionCount := pa.packageFunctions[pkgName]
	typeCount := pa.packageTypes[pkgName]
	depCount := len(pa.packageDeps[pkgName])

	// Weight different factors
	complexity := float64(functionCount)*0.1 + float64(typeCount)*0.2 + float64(depCount)*0.3

	// Normalize to 0-10 scale
	if complexity > 10.0 {
		complexity = 10.0
	}

	return complexity
}

// detectCircularDependencies finds cycles in the package dependency graph
func (pa *PackageAnalyzer) detectCircularDependencies() []metrics.CircularDependency {
	var cycles []metrics.CircularDependency
	visited := make(map[string]bool)

	// DFS-based cycle detection
	for pkg := range pa.packageDeps {
		if !visited[pkg] {
			recStack := make(map[string]bool)
			if cycle := pa.dfsCircular(pkg, visited, recStack, []string{}); len(cycle) > 0 {
				cycles = append(cycles, metrics.CircularDependency{
					Packages: cycle,
					Severity: pa.calculateCycleSeverity(cycle),
				})
			}
		}
	}

	return cycles
}

// dfsCircular performs depth-first search to detect cycles
func (pa *PackageAnalyzer) dfsCircular(pkg string, visited, recStack map[string]bool, path []string) []string {
	visited[pkg] = true
	recStack[pkg] = true
	path = append(path, pkg)

	for _, dep := range pa.packageDeps[pkg] {
		if !visited[dep] {
			if cycle := pa.dfsCircular(dep, visited, recStack, path); len(cycle) > 0 {
				return cycle
			}
		} else if recStack[dep] {
			// Found cycle - return the cycle path starting from dep
			cycleStart := -1
			for i, p := range path {
				if p == dep {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				return append(path[cycleStart:], dep) // Close the cycle
			}
		}
	}

	recStack[pkg] = false
	return nil
}

// calculateCycleSeverity determines how problematic a circular dependency is
func (pa *PackageAnalyzer) calculateCycleSeverity(cycle []string) string {
	if len(cycle) <= 2 {
		return "low"
	} else if len(cycle) <= 4 {
		return "medium"
	}
	return "high"
}

// buildDependencyGraph creates a representation of the dependency relationships
func (pa *PackageAnalyzer) buildDependencyGraph() map[string][]string {
	graph := make(map[string][]string)

	for pkg, deps := range pa.packageDeps {
		graph[pkg] = make([]string, len(deps))
		copy(graph[pkg], deps)
		sort.Strings(graph[pkg]) // Consistent ordering
	}

	return graph
}

// Helper functions

// isInternalPackage determines if an import is an internal package
// (not stdlib or external dependency)
func isInternalPackage(importPath string) bool {
	// Consider relative paths as internal
	if strings.HasPrefix(importPath, "./") || strings.HasPrefix(importPath, "../") {
		return true
	}

	// Consider paths starting with known internal prefixes as internal
	internalPrefixes := []string{
		"internal/",
		"pkg/",
	}

	for _, prefix := range internalPrefixes {
		if strings.HasPrefix(importPath, prefix) {
			return true
		}
	}

	// Skip standard library packages (no dots and not internal prefixes)
	if !strings.Contains(importPath, ".") {
		return false
	}

	// Skip common external packages
	external := []string{
		"github.com/spf13/",
		"github.com/stretchr/",
		"github.com/jedib0t/",
		"github.com/olekukonko/",
	}

	for _, ext := range external {
		if strings.HasPrefix(importPath, ext) {
			return false
		}
	}

	return true
}

// mergeUniqueStrings merges two string slices, removing duplicates
func mergeUniqueStrings(existing, new []string) []string {
	seen := make(map[string]bool)

	// Add existing items
	for _, item := range existing {
		seen[item] = true
	}

	// Add new items if not already present
	result := make([]string, len(existing))
	copy(result, existing)

	for _, item := range new {
		if !seen[item] {
			result = append(result, item)
			seen[item] = true
		}
	}

	return result
}

// calculateAverageFiles calculates average number of files per package
func (pa *PackageAnalyzer) calculateAverageFiles() float64 {
	if len(pa.packageFiles) == 0 {
		return 0.0
	}

	total := 0
	for _, files := range pa.packageFiles {
		total += len(files)
	}

	return float64(total) / float64(len(pa.packageFiles))
}

// calculateAverageFloat calculates average for map[string][]string (file counts)
func (pa *PackageAnalyzer) calculateAverageFloat(data map[string][]string) float64 {
	if len(data) == 0 {
		return 0.0
	}

	total := 0
	for _, items := range data {
		total += len(items)
	}

	return float64(total) / float64(len(data))
}

// calculateAverageInt calculates average for map[string]int
func (pa *PackageAnalyzer) calculateAverageInt(data map[string]int) float64 {
	if len(data) == 0 {
		return 0.0
	}

	total := 0
	for _, count := range data {
		total += count
	}

	return float64(total) / float64(len(data))
}
