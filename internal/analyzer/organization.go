package analyzer

import (
	"go/ast"
	"go/token"
	"os"
	"strings"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
)

// OrganizationAnalyzer analyzes file and package organization
type OrganizationAnalyzer struct {
	fset *token.FileSet
}

// NewOrganizationAnalyzer creates a new organization analyzer for evaluating
// NewOrganizationAnalyzer analyzes file sizes, package structure, and directory depth against thresholds.
func NewOrganizationAnalyzer(fset *token.FileSet) *OrganizationAnalyzer {
	return &OrganizationAnalyzer{
		fset: fset,
	}
}

// OrganizationConfig holds threshold configuration
type OrganizationConfig struct {
	MaxFileLines       int
	MaxFileFunctions   int
	MaxFileTypes       int
	MaxPackageFiles    int
	MaxExportedSymbols int
	MaxDirectoryDepth  int
	MaxFileImports     int
}

// DefaultOrganizationConfig returns default configuration values for organization
// DefaultOrganizationConfig sets analysis thresholds including file lines, functions, and package limits.
func DefaultOrganizationConfig() OrganizationConfig {
	return OrganizationConfig{
		MaxFileLines:       500,
		MaxFileFunctions:   20,
		MaxFileTypes:       5,
		MaxPackageFiles:    20,
		MaxExportedSymbols: 50,
		MaxDirectoryDepth:  5,
		MaxFileImports:     15,
	}
}

// AnalyzeFileSizes analyzes file sizes and complexity against configured limits.
// AnalyzeFileSizes returns nil if the file is within acceptable thresholds.
func (oa *OrganizationAnalyzer) AnalyzeFileSizes(file *ast.File, filePath string, config OrganizationConfig) (*metrics.OversizedFile, error) {
	lines := oa.countFileLines(filePath)
	funcCount := oa.countFunctions(file)
	typeCount := oa.countTypes(file)

	burden := oa.calculateBurden(lines, funcCount, typeCount)

	if !oa.isOversized(lines, funcCount, typeCount, config) {
		return nil, nil
	}

	return &metrics.OversizedFile{
		File:              filePath,
		Lines:             lines,
		FunctionCount:     funcCount,
		TypeCount:         typeCount,
		MaintenanceBurden: burden,
		Severity:          oa.getSeverity(lines, funcCount, typeCount, config),
		Suggestions:       oa.getSuggestions(lines, funcCount, typeCount, config),
	}, nil
}

// countFileLines counts lines in an entire file
func (oa *OrganizationAnalyzer) countFileLines(filePath string) metrics.LineMetrics {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return metrics.LineMetrics{}
	}

	lines := strings.Split(string(src), "\n")
	return oa.analyzeLinesInFile(lines)
}

// analyzeLinesInFile categorizes lines
func (oa *OrganizationAnalyzer) analyzeLinesInFile(lines []string) metrics.LineMetrics {
	var codeLines, commentLines, blankLines int
	inBlockComment := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			blankLines++
			continue
		}

		lineType := oa.classifyLine(trimmed, &inBlockComment)
		switch lineType {
		case "code":
			codeLines++
		case "comment":
			commentLines++
		case "mixed":
			codeLines++
		}
	}

	total := len(lines)
	return metrics.LineMetrics{
		Total:    total,
		Code:     codeLines,
		Comments: commentLines,
		Blank:    blankLines,
	}
}

// classifyLine determines line type
func (oa *OrganizationAnalyzer) classifyLine(line string, inBlock *bool) string {
	if oa.isBlockCommentStart(line, inBlock) {
		return "comment"
	}

	if *inBlock {
		return oa.handleBlockComment(line, inBlock)
	}

	if strings.HasPrefix(line, "//") {
		return "comment"
	}

	if strings.Contains(line, "//") {
		return "mixed"
	}

	return "code"
}

// isBlockCommentStart checks for block comment start
func (oa *OrganizationAnalyzer) isBlockCommentStart(line string, inBlock *bool) bool {
	if !strings.HasPrefix(line, "/*") {
		return false
	}

	*inBlock = true
	if strings.HasSuffix(line, "*/") {
		*inBlock = false
	}
	return true
}

// handleBlockComment processes lines in block comment
func (oa *OrganizationAnalyzer) handleBlockComment(line string, inBlock *bool) string {
	if strings.Contains(line, "*/") {
		*inBlock = false
	}
	return "comment"
}

// countFunctions counts functions and methods
func (oa *OrganizationAnalyzer) countFunctions(file *ast.File) int {
	count := 0
	for _, decl := range file.Decls {
		if _, ok := decl.(*ast.FuncDecl); ok {
			count++
		}
	}
	return count
}

// countTypes counts type declarations
func (oa *OrganizationAnalyzer) countTypes(file *ast.File) int {
	count := 0
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			if genDecl.Tok == token.TYPE {
				count += len(genDecl.Specs)
			}
		}
	}
	return count
}

// calculateBurden computes composite burden score
func (oa *OrganizationAnalyzer) calculateBurden(lines metrics.LineMetrics, funcCount, typeCount int) float64 {
	lineScore := float64(lines.Total) / 500.0
	funcScore := float64(funcCount) / 20.0
	typeScore := float64(typeCount) / 5.0

	return (lineScore + funcScore + typeScore) / 3.0
}

// isOversized checks if file exceeds thresholds
func (oa *OrganizationAnalyzer) isOversized(lines metrics.LineMetrics, funcCount, typeCount int, config OrganizationConfig) bool {
	return lines.Total > config.MaxFileLines ||
		funcCount > config.MaxFileFunctions ||
		typeCount > config.MaxFileTypes
}

// getSeverity determines severity level
func (oa *OrganizationAnalyzer) getSeverity(lines metrics.LineMetrics, funcCount, typeCount int, config OrganizationConfig) metrics.SeverityLevel {
	criticalCount := 0

	if lines.Total > config.MaxFileLines*2 {
		criticalCount++
	}
	if funcCount > config.MaxFileFunctions*2 {
		criticalCount++
	}
	if typeCount > config.MaxFileTypes*2 {
		criticalCount++
	}

	if criticalCount >= 2 {
		return metrics.SeverityLevelCritical
	}
	if criticalCount == 1 {
		return metrics.SeverityLevelViolation
	}
	return metrics.SeverityLevelWarning
}

// getSuggestions generates improvement suggestions
func (oa *OrganizationAnalyzer) getSuggestions(lines metrics.LineMetrics, funcCount, typeCount int, config OrganizationConfig) []string {
	var suggestions []string

	if lines.Total > config.MaxFileLines {
		suggestions = append(suggestions, "Consider splitting file - exceeds maximum line count")
	}
	if funcCount > config.MaxFileFunctions {
		suggestions = append(suggestions, "Too many functions - group related functions into separate files")
	}
	if typeCount > config.MaxFileTypes {
		suggestions = append(suggestions, "Too many types - separate type definitions into focused files")
	}

	return suggestions
}

// PackageInfo holds aggregated package data for organization analysis, tracking the package name,
// associated files, exported symbol count, total functions, and cohesion score for package size analysis.
type PackageInfo struct {
	Name            string
	Files           []string
	ExportedSymbols int
	TotalFunctions  int
	CohesionScore   float64
}

// AnalyzePackageSizes analyzes package size metrics against configured limits.
// AnalyzePackageSizes returns a list of packages exceeding thresholds or exhibiting mega-package signs.
func (oa *OrganizationAnalyzer) AnalyzePackageSizes(pkgs map[string]*PackageInfo, config OrganizationConfig) []metrics.OversizedPackage {
	var results []metrics.OversizedPackage

	for name, pkg := range pkgs {
		if oa.shouldReportPackage(pkg, config) {
			results = append(results, metrics.OversizedPackage{
				Package:         name,
				FileCount:       len(pkg.Files),
				ExportedSymbols: pkg.ExportedSymbols,
				TotalFunctions:  pkg.TotalFunctions,
				CohesionScore:   pkg.CohesionScore,
				IsMegaPackage:   oa.isMegaPackage(pkg),
				Severity:        oa.getPackageSeverity(pkg, config),
				Suggestions:     oa.getPackageSuggestions(pkg, config),
			})
		}
	}

	return results
}

// shouldReportPackage checks if package should be reported
func (oa *OrganizationAnalyzer) shouldReportPackage(pkg *PackageInfo, config OrganizationConfig) bool {
	return oa.isPackageOversized(pkg, config) || oa.isMegaPackage(pkg)
}

// isPackageOversized checks if package exceeds thresholds
func (oa *OrganizationAnalyzer) isPackageOversized(pkg *PackageInfo, config OrganizationConfig) bool {
	return len(pkg.Files) > config.MaxPackageFiles ||
		pkg.ExportedSymbols > config.MaxExportedSymbols
}

// isMegaPackage checks for low cohesion + high symbol count
func (oa *OrganizationAnalyzer) isMegaPackage(pkg *PackageInfo) bool {
	return pkg.CohesionScore < 0.5 && pkg.ExportedSymbols > 30
}

// getPackageSeverity determines severity level
func (oa *OrganizationAnalyzer) getPackageSeverity(pkg *PackageInfo, config OrganizationConfig) metrics.SeverityLevel {
	if oa.isMegaPackage(pkg) {
		return metrics.SeverityLevelCritical
	}

	violations := 0
	if len(pkg.Files) > config.MaxPackageFiles*2 {
		violations++
	}
	if pkg.ExportedSymbols > config.MaxExportedSymbols*2 {
		violations++
	}

	if violations >= 2 {
		return metrics.SeverityLevelCritical
	}
	if violations == 1 {
		return metrics.SeverityLevelViolation
	}
	return metrics.SeverityLevelWarning
}

// getPackageSuggestions generates package improvement suggestions
func (oa *OrganizationAnalyzer) getPackageSuggestions(pkg *PackageInfo, config OrganizationConfig) []string {
	var suggestions []string

	if len(pkg.Files) > config.MaxPackageFiles {
		suggestions = append(suggestions, "Too many files - consider splitting into sub-packages")
	}
	if pkg.ExportedSymbols > config.MaxExportedSymbols {
		suggestions = append(suggestions, "Too many exported symbols - reduce public API surface")
	}
	if oa.isMegaPackage(pkg) {
		suggestions = append(suggestions, "Mega-package detected - refactor into cohesive modules")
	}

	return suggestions
}

// AnalyzeDirectoryDepth analyzes directory nesting depth against threshold.
// AnalyzeDirectoryDepth returns directories exceeding the maximum allowed depth with file counts.
func (oa *OrganizationAnalyzer) AnalyzeDirectoryDepth(paths []string, rootPath string, config OrganizationConfig) []metrics.DeepDirectory {
	depthMap := make(map[string]*directoryStats)

	for _, path := range paths {
		dir := oa.getDirectoryPath(path)
		depth := oa.calculateDepth(dir, rootPath)
		if depth > config.MaxDirectoryDepth {
			if stats, exists := depthMap[dir]; exists {
				stats.fileCount++
			} else {
				depthMap[dir] = &directoryStats{
					depth:     depth,
					fileCount: 1,
				}
			}
		}
	}

	return oa.buildDeepDirectories(depthMap, config)
}

// directoryStats holds directory statistics
type directoryStats struct {
	depth     int
	fileCount int
}

// calculateDepth calculates directory nesting depth
func (oa *OrganizationAnalyzer) calculateDepth(dirPath, rootPath string) int {
	rel := strings.TrimPrefix(dirPath, rootPath)
	rel = strings.TrimPrefix(rel, "/")
	if rel == "" || rel == "." {
		return 0
	}
	return strings.Count(rel, "/") + 1
}

// getDirectoryPath extracts directory from file path
func (oa *OrganizationAnalyzer) getDirectoryPath(filePath string) string {
	lastSlash := strings.LastIndex(filePath, "/")
	if lastSlash == -1 {
		return "."
	}
	return filePath[:lastSlash]
}

// buildDeepDirectories converts depth map to results
func (oa *OrganizationAnalyzer) buildDeepDirectories(depthMap map[string]*directoryStats, config OrganizationConfig) []metrics.DeepDirectory {
	var results []metrics.DeepDirectory

	for path, stats := range depthMap {
		results = append(results, metrics.DeepDirectory{
			Path:       path,
			Depth:      stats.depth,
			FileCount:  stats.fileCount,
			Severity:   oa.getDepthSeverity(stats.depth, config),
			Suggestion: oa.getDepthSuggestion(stats.depth, config),
		})
	}

	return results
}

// getDepthSeverity determines depth severity
func (oa *OrganizationAnalyzer) getDepthSeverity(depth int, config OrganizationConfig) metrics.SeverityLevel {
	threshold := float64(config.MaxDirectoryDepth)
	if float64(depth) > threshold*2 {
		return metrics.SeverityLevelCritical
	}
	if float64(depth) > threshold*1.5 {
		return metrics.SeverityLevelViolation
	}
	return metrics.SeverityLevelWarning
}

// getDepthSuggestion generates depth suggestion
func (oa *OrganizationAnalyzer) getDepthSuggestion(depth int, config OrganizationConfig) string {
	excess := depth - config.MaxDirectoryDepth
	if excess > 3 {
		return "Restructure directory hierarchy - nesting is excessively deep"
	}
	return "Consider flattening directory structure"
}

// ImportGraphData tracks import relationships across the codebase
type ImportGraphData struct {
	FileImports    map[string]int      // file -> import count
	PackageFanIn   map[string][]string // package -> packages that import it
	PackageFanOut  map[string][]string // package -> packages it imports
	FilePackageMap map[string]string   // file -> package name
}

// AnalyzeImportGraph analyzes import relationships and identifies problematic
// AnalyzeImportGraph detects patterns such as excessive imports, high fan-in/fan-out, and instability.
func (oa *OrganizationAnalyzer) AnalyzeImportGraph(graphData *ImportGraphData, config OrganizationConfig) (*ImportGraphMetrics, error) {
	if graphData == nil {
		return nil, nil
	}

	metrics := &ImportGraphMetrics{
		FileImportViolations: oa.findFileImportViolations(graphData, config),
		HighFanInPackages:    oa.findHighFanIn(graphData),
		HighFanOutPackages:   oa.findHighFanOut(graphData),
	}

	metrics.AvgInstability = oa.calculateAvgInstability(graphData)

	return metrics, nil
}

// ImportGraphMetrics holds import graph analysis results
type ImportGraphMetrics struct {
	FileImportViolations []FileImportViolation
	HighFanInPackages    []metrics.FanInPackage
	HighFanOutPackages   []metrics.FanOutPackage
	AvgInstability       float64
}

// FileImportViolation represents a file with too many imports
type FileImportViolation struct {
	File        string
	Package     string
	ImportCount int
	Severity    string
	Suggestion  string
}

// findFileImportViolations identifies files with excessive imports
func (oa *OrganizationAnalyzer) findFileImportViolations(graphData *ImportGraphData, config OrganizationConfig) []FileImportViolation {
	var violations []FileImportViolation

	for file, count := range graphData.FileImports {
		if count > config.MaxFileImports {
			violations = append(violations, FileImportViolation{
				File:        file,
				Package:     graphData.FilePackageMap[file],
				ImportCount: count,
				Severity:    oa.getImportSeverity(count, config),
				Suggestion:  oa.getImportSuggestion(count, config),
			})
		}
	}

	return violations
}

// getImportSeverity determines severity of import count violation
func (oa *OrganizationAnalyzer) getImportSeverity(count int, config OrganizationConfig) string {
	threshold := float64(config.MaxFileImports)
	if float64(count) > threshold*2 {
		return string(metrics.SeverityLevelCritical)
	}
	if float64(count) > threshold*1.5 {
		return string(metrics.SeverityLevelViolation)
	}
	return string(metrics.SeverityLevelWarning)
}

// getImportSuggestion generates import violation suggestion
func (oa *OrganizationAnalyzer) getImportSuggestion(count int, config OrganizationConfig) string {
	excess := count - config.MaxFileImports
	if excess > 10 {
		return "Excessive imports - split into multiple focused files"
	}
	return "Too many imports - consider refactoring to reduce dependencies"
}

// findHighFanIn identifies hub packages with high incoming dependencies
func (oa *OrganizationAnalyzer) findHighFanIn(graphData *ImportGraphData) []metrics.FanInPackage {
	var results []metrics.FanInPackage
	threshold := 3

	for pkg, dependents := range graphData.PackageFanIn {
		fanIn := len(dependents)
		if fanIn >= threshold {
			results = append(results, metrics.FanInPackage{
				Package:      pkg,
				FanIn:        fanIn,
				Dependents:   dependents,
				IsBottleneck: fanIn >= 5,
				RiskLevel:    oa.getFanInRisk(fanIn),
				Suggestion:   oa.getFanInSuggestion(fanIn),
			})
		}
	}

	return results
}

// getFanInRisk determines risk level for high fan-in
func (oa *OrganizationAnalyzer) getFanInRisk(fanIn int) string {
	if fanIn >= 10 {
		return string(metrics.SeverityLevelCritical)
	}
	if fanIn >= 5 {
		return string(metrics.SeverityLevelViolation)
	}
	return string(metrics.SeverityLevelWarning)
}

// getFanInSuggestion generates suggestion for high fan-in
func (oa *OrganizationAnalyzer) getFanInSuggestion(fanIn int) string {
	if fanIn >= 10 {
		return "Critical bottleneck - changes to this package affect many dependents"
	}
	return "Hub package - ensure stability and comprehensive testing"
}

// findHighFanOut identifies authority packages with high outgoing dependencies
func (oa *OrganizationAnalyzer) findHighFanOut(graphData *ImportGraphData) []metrics.FanOutPackage {
	var results []metrics.FanOutPackage
	threshold := 5

	for pkg, dependencies := range graphData.PackageFanOut {
		fanOut := len(dependencies)
		if fanOut >= threshold {
			fanIn := len(graphData.PackageFanIn[pkg])
			instability := oa.calculateInstability(fanIn, fanOut)

			results = append(results, metrics.FanOutPackage{
				Package:      pkg,
				FanOut:       fanOut,
				Dependencies: dependencies,
				Instability:  instability,
				CouplingRisk: oa.getCouplingRisk(fanOut, instability),
				Suggestion:   oa.getFanOutSuggestion(fanOut),
			})
		}
	}

	return results
}

// calculateInstability computes Martin's instability metric: Ce / (Ca + Ce)
func (oa *OrganizationAnalyzer) calculateInstability(fanIn, fanOut int) float64 {
	total := fanIn + fanOut
	if total == 0 {
		return 0.0
	}
	return float64(fanOut) / float64(total)
}

// getCouplingRisk determines coupling risk level
func (oa *OrganizationAnalyzer) getCouplingRisk(fanOut int, instability float64) string {
	if fanOut >= 10 && instability > 0.7 {
		return string(metrics.SeverityLevelCritical)
	}
	if fanOut >= 7 || instability > 0.6 {
		return string(metrics.SeverityLevelViolation)
	}
	return string(metrics.SeverityLevelWarning)
}

// getFanOutSuggestion generates suggestion for high fan-out
func (oa *OrganizationAnalyzer) getFanOutSuggestion(fanOut int) string {
	if fanOut >= 10 {
		return "Excessive coupling - refactor to reduce dependencies"
	}
	return "High coupling - consider dependency injection or interfaces"
}

// calculateAvgInstability computes average instability across all packages
func (oa *OrganizationAnalyzer) calculateAvgInstability(graphData *ImportGraphData) float64 {
	if len(graphData.PackageFanOut) == 0 {
		return 0.0
	}

	total := 0.0
	count := 0

	for pkg, fanOutDeps := range graphData.PackageFanOut {
		fanOut := len(fanOutDeps)
		fanIn := len(graphData.PackageFanIn[pkg])
		instability := oa.calculateInstability(fanIn, fanOut)
		total += instability
		count++
	}

	if count == 0 {
		return 0.0
	}

	return total / float64(count)
}
