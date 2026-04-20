package analyzer

import (
	"fmt"
	"go/ast"
	"go/token"
	"sort"
	"strings"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// knownInternalPrefixes is evaluated at package init time to avoid repeated slice allocation
// inside the per-import hot path of isKnownInternalPrefix.
var knownInternalPrefixes = []string{
	"internal/",
	"pkg/",
}

// knownExternalPrefixes is evaluated at package init time to avoid repeated slice allocation
// inside the per-import hot path of isKnownExternalPackage.
var knownExternalPrefixes = []string{
	"github.com/spf13/",
	"github.com/stretchr/",
	"github.com/jedib0t/",
	"github.com/olekukonko/",
}

// PackageAnalyzer analyzes package-level metrics including dependencies,
// PackageAnalyzer computes cohesion, coupling, and circular dependency detection.
// PackageAnalyzer provides architectural insights for large Go codebases.
type PackageAnalyzer struct {
	fset             *token.FileSet
	packageDeps      map[string][]string // package -> imported packages
	packageFiles     map[string][]string // package -> source files
	packageFunctions map[string]int      // package -> function count
	packageTypes     map[string]int      // package -> type count
	packageLines     map[string]int      // package -> total lines of code
}

// NewPackageAnalyzer creates a new package analyzer for architectural analysis including dependency
// tracking, cohesion and coupling metrics, circular dependency detection, and package organization
// assessment. Analyzes package-level structure to identify architectural issues, high coupling, and
// design patterns. Essential for large codebase refactoring and architecture review.
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

// AnalyzePackage analyzes a single source file within a package, collecting dependency imports,
// function counts, type definitions, and lines of code for cohesion/coupling analysis. Multiple
// files per package are aggregated to compute package-level metrics. Returns error if the file
// lacks a package name declaration (malformed Go file).
func (pa *PackageAnalyzer) AnalyzePackage(file *ast.File, filePath string) error {
	if file.Name == nil {
		return fmt.Errorf("file has no package name: %s", filePath)
	}

	pkgName := file.Name.Name
	pa.trackPackageFile(pkgName, filePath)
	pa.analyzePackageImports(file, pkgName)
	pa.countDeclarations(file, pkgName)
	pa.trackFileLines(file, pkgName)

	return nil
}

// trackPackageFile records a file as belonging to the specified package.
func (pa *PackageAnalyzer) trackPackageFile(pkgName, filePath string) {
	pa.packageFiles[pkgName] = append(pa.packageFiles[pkgName], filePath)
}

// analyzePackageImports extracts internal dependencies from file imports.
func (pa *PackageAnalyzer) analyzePackageImports(file *ast.File, pkgName string) {
	imports := pa.extractInternalImports(file)
	existing := pa.packageDeps[pkgName]
	pa.packageDeps[pkgName] = mergeUniqueStrings(existing, imports)
}

// extractInternalImports collects internal package imports, excluding stdlib and external packages.
func (pa *PackageAnalyzer) extractInternalImports(file *ast.File) []string {
	var imports []string
	for _, imp := range file.Imports {
		if imp.Path != nil {
			importPath := strings.Trim(imp.Path.Value, `"`)
			if isInternalPackage(importPath) {
				imports = append(imports, importPath)
			}
		}
	}
	return imports
}

// countDeclarations counts functions and type declarations in the file.
func (pa *PackageAnalyzer) countDeclarations(file *ast.File, pkgName string) {
	functionCount, typeCount := pa.extractDeclCounts(file)
	pa.packageFunctions[pkgName] += functionCount
	pa.packageTypes[pkgName] += typeCount
}

// extractDeclCounts extracts function and type declaration counts from file.
func (pa *PackageAnalyzer) extractDeclCounts(file *ast.File) (functionCount, typeCount int) {
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
	return functionCount, typeCount
}

// trackFileLines calculates and records the line count for the file.
func (pa *PackageAnalyzer) trackFileLines(file *ast.File, pkgName string) {
	linesInFile := pa.calculateFileLines(file)
	pa.packageLines[pkgName] += linesInFile
}

// calculateFileLines computes the total line count for a file.
func (pa *PackageAnalyzer) calculateFileLines(file *ast.File) int {
	start := pa.fset.Position(file.Pos())
	end := pa.fset.Position(file.End())
	return end.Line - start.Line + 1
}

// GenerateReport generates comprehensive package metrics report including
// GenerateReport computes cohesion, coupling, and dependency analysis for all analyzed packages.
func (pa *PackageAnalyzer) GenerateReport() (*metrics.PackageReport, error) {
	packages := pa.buildPackageMetrics()
	sortPackagesByName(packages)

	report := &metrics.PackageReport{
		Packages:             packages,
		TotalPackages:        len(packages),
		CircularDependencies: pa.detectCircularDependencies(),
		DependencyGraph:      pa.buildDependencyGraph(),
	}

	report.AverageFilesPerPackage = pa.calculateAverageFiles()
	report.AverageFunctionsPerPackage = pa.calculateAverageInt(pa.packageFunctions)
	report.AverageTypesPerPackage = pa.calculateAverageInt(pa.packageTypes)

	return report, nil
}

// buildPackageMetrics creates PackageMetrics for all analyzed packages.
func (pa *PackageAnalyzer) buildPackageMetrics() []metrics.PackageMetrics {
	packages := make([]metrics.PackageMetrics, 0, len(pa.packageFiles))
	for pkgName := range pa.packageFiles {
		packages = append(packages, pa.createPackageMetrics(pkgName))
	}
	return packages
}

// createPackageMetrics builds metrics for a single package.
func (pa *PackageAnalyzer) createPackageMetrics(pkgName string) metrics.PackageMetrics {
	pkg := metrics.PackageMetrics{
		Name:         pkgName,
		Path:         pkgName,
		Files:        pa.packageFiles[pkgName],
		Functions:    pa.packageFunctions[pkgName],
		Structs:      pa.packageTypes[pkgName],
		Interfaces:   0,
		Dependencies: pa.packageDeps[pkgName],
		Lines: metrics.LineMetrics{
			Total: pa.packageLines[pkgName],
			Code:  pa.packageLines[pkgName],
		},
	}
	pkg.CohesionScore = pa.calculateCohesion(pkgName)
	pkg.CouplingScore = pa.calculateCoupling(pkgName)
	return pkg
}

// sortPackagesByName sorts packages alphabetically by name.
func sortPackagesByName(packages []metrics.PackageMetrics) {
	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Name < packages[j].Name
	})
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
	cycles := make([]metrics.CircularDependency, 0)
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
		if cycle := pa.checkDependencyForCycle(dep, visited, recStack, path); len(cycle) > 0 {
			return cycle
		}
	}

	recStack[pkg] = false
	return nil
}

// checkDependencyForCycle checks a single dependency for cycles
func (pa *PackageAnalyzer) checkDependencyForCycle(dep string, visited, recStack map[string]bool, path []string) []string {
	if !visited[dep] {
		return pa.checkUnvisitedDependency(dep, visited, recStack, path)
	}
	if recStack[dep] {
		return pa.extractCyclePath(dep, path)
	}
	return nil
}

// checkUnvisitedDependency recursively explores unvisited dependencies
func (pa *PackageAnalyzer) checkUnvisitedDependency(dep string, visited, recStack map[string]bool, path []string) []string {
	if cycle := pa.dfsCircular(dep, visited, recStack, path); len(cycle) > 0 {
		return cycle
	}
	return nil
}

// extractCyclePath finds and returns the cycle path from the recursion stack
func (pa *PackageAnalyzer) extractCyclePath(dep string, path []string) []string {
	cycleStart := pa.findCycleStart(dep, path)
	if cycleStart >= 0 {
		return append(path[cycleStart:], dep)
	}
	return nil
}

// findCycleStart locates the starting index of the cycle in the path
func (pa *PackageAnalyzer) findCycleStart(dep string, path []string) int {
	for i, p := range path {
		if p == dep {
			return i
		}
	}
	return -1
}

// calculateCycleSeverity determines how problematic a circular dependency is
func (pa *PackageAnalyzer) calculateCycleSeverity(cycle []string) metrics.SeverityLevel {
	// Count unique packages in the cycle (exclude the closing duplicate)
	uniquePackages := len(cycle) - 1
	if uniquePackages <= 2 {
		return metrics.SeverityLevelInfo
	} else if uniquePackages <= 4 {
		return metrics.SeverityLevelWarning
	}
	return metrics.SeverityLevelCritical
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
	if isKnownInternalPrefix(importPath) {
		return true
	}

	// Skip standard library packages (no dots and not internal prefixes)
	if !strings.Contains(importPath, ".") {
		return false
	}

	// Skip common external packages
	if isKnownExternalPackage(importPath) {
		return false
	}

	return true
}

// isKnownInternalPrefix checks if the import path starts with known internal prefixes
func isKnownInternalPrefix(importPath string) bool {
	for _, prefix := range knownInternalPrefixes {
		if strings.HasPrefix(importPath, prefix) {
			return true
		}
	}
	return false
}

// isKnownExternalPackage checks if the import path is a known external package
func isKnownExternalPackage(importPath string) bool {
	for _, ext := range knownExternalPrefixes {
		if strings.HasPrefix(importPath, ext) {
			return true
		}
	}
	return false
}

// mergeUniqueStrings merges two string slices, removing duplicates.
// It avoids allocating a map for the common case where neither slice is large by
// appending and then deduplicating via sort, which avoids heap allocation for short lists.
func mergeUniqueStrings(existing, newItems []string) []string {
	if len(newItems) == 0 {
		return existing
	}

	combined := make([]string, len(existing)+len(newItems))
	copy(combined, existing)
	copy(combined[len(existing):], newItems)

	sort.Strings(combined)

	deduped := combined[:0]
	for i, s := range combined {
		if i == 0 || s != combined[i-1] {
			deduped = append(deduped, s)
		}
	}
	return deduped
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
