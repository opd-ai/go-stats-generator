package generator

import (
	"context"
	"go/ast"
	"go/token"
	"time"

	"github.com/opd-ai/go-stats-generator/internal/analyzer"
	"github.com/opd-ai/go-stats-generator/internal/config"
	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/opd-ai/go-stats-generator/internal/scanner"
)

// Analyzer provides programmatic access to Go code analysis
type Analyzer struct {
	config *config.Config
}

// NewAnalyzer creates a new analyzer
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		config: config.DefaultConfig(),
	}
}

// NewAnalyzerWithConfig creates analyzer with config
func NewAnalyzerWithConfig(cfg *config.Config) *Analyzer {
	return &Analyzer{
		config: cfg,
	}
}

// analyzeResults processes parsed files
func (a *Analyzer) analyzeResults(ctx context.Context, results <-chan scanner.Result, fset *token.FileSet, rootPath string, fileCount int) (*metrics.Report, error) {
	analyzers := createAnalyzers(fset, a.config)
	report := createReport(rootPath, fileCount)
	collected := &collectedMetrics{files: make(map[string]*ast.File)}

	for result := range results {
		if result.Error != nil {
			continue
		}
		processFile(result, analyzers, collected, report, a.config)
	}

	finalizeReport(report, collected, analyzers)
	return report, nil
}

// collectedMetrics holds metrics
type collectedMetrics struct {
	functions  []metrics.FunctionMetrics
	structs    []metrics.StructMetrics
	interfaces []metrics.InterfaceMetrics
	files      map[string]*ast.File
	generics   []metrics.GenericMetrics
}

// analyzerSet holds all analyzers
type analyzerSet struct {
	function    *analyzer.FunctionAnalyzer
	structure   *analyzer.StructAnalyzer
	iface       *analyzer.InterfaceAnalyzer
	pkg         *analyzer.PackageAnalyzer
	concurrency *analyzer.ConcurrencyAnalyzer
	pattern     *analyzer.PatternAnalyzer
	duplication *analyzer.DuplicationAnalyzer
	generic     *analyzer.GenericAnalyzer
}

// createAnalyzers initializes analyzers
func createAnalyzers(fset *token.FileSet, cfg *config.Config) *analyzerSet {
	return &analyzerSet{
		function:    analyzer.NewFunctionAnalyzer(fset),
		structure:   analyzer.NewStructAnalyzer(fset),
		iface:       analyzer.NewInterfaceAnalyzer(fset),
		pkg:         analyzer.NewPackageAnalyzer(fset),
		concurrency: analyzer.NewConcurrencyAnalyzer(fset),
		pattern:     analyzer.NewPatternAnalyzer(fset),
		duplication: analyzer.NewDuplicationAnalyzer(fset),
		generic:     analyzer.NewGenericAnalyzer(fset),
	}
}

// createReport creates initial report
func createReport(rootPath string, fileCount int) *metrics.Report {
	return &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:     rootPath,
			GeneratedAt:    time.Now(),
			FilesProcessed: fileCount,
			ToolVersion:    "1.0.0",
		},
		Patterns: metrics.PatternMetrics{
			DesignPatterns: metrics.DesignPatternMetrics{
				Singleton: []metrics.PatternInstance{},
				Factory:   []metrics.PatternInstance{},
				Builder:   []metrics.PatternInstance{},
				Observer:  []metrics.PatternInstance{},
				Strategy:  []metrics.PatternInstance{},
			},
			ConcurrencyPatterns: metrics.ConcurrencyPatternMetrics{},
		},
	}
}

// processFile performs analysis on file
func processFile(result scanner.Result, analyzers *analyzerSet, collected *collectedMetrics, report *metrics.Report, cfg *config.Config) {
	collected.files[result.FileInfo.RelPath] = result.File

	if funcs, err := analyzers.function.AnalyzeFunctionsWithPath(result.File, result.FileInfo.Package, result.FileInfo.RelPath); err == nil {
		collected.functions = append(collected.functions, funcs...)
	}

	if structs, err := analyzers.structure.AnalyzeStructsWithPath(result.File, result.FileInfo.Package, result.FileInfo.RelPath); err == nil {
		collected.structs = append(collected.structs, structs...)
	}

	if ifaces, err := analyzers.iface.AnalyzeInterfacesWithPath(result.File, result.FileInfo.Package, result.FileInfo.RelPath); err == nil {
		collected.interfaces = append(collected.interfaces, ifaces...)
	}

	if generics, err := analyzers.generic.AnalyzeGenerics(result.File, result.FileInfo.Package, result.FileInfo.RelPath); err == nil {
		collected.generics = append(collected.generics, generics)
	}

	analyzers.pkg.AnalyzePackage(result.File, result.FileInfo.Path)
	analyzeConcurrency(result, analyzers.concurrency, report)
	analyzePatterns(result, analyzers.pattern, report)
}

// analyzeConcurrency analyzes concurrency
func analyzeConcurrency(result scanner.Result, concurrencyAnalyzer *analyzer.ConcurrencyAnalyzer, report *metrics.Report) {
	concurrency, err := concurrencyAnalyzer.AnalyzeConcurrency(result.File, result.FileInfo.Package)
	if err != nil {
		return
	}

	report.Patterns.ConcurrencyPatterns.Goroutines.Instances = append(
		report.Patterns.ConcurrencyPatterns.Goroutines.Instances,
		concurrency.Goroutines.Instances...)
	report.Patterns.ConcurrencyPatterns.Channels.Instances = append(
		report.Patterns.ConcurrencyPatterns.Channels.Instances,
		concurrency.Channels.Instances...)
}

// analyzePatterns analyzes patterns
func analyzePatterns(result scanner.Result, patternAnalyzer *analyzer.PatternAnalyzer, report *metrics.Report) {
	patterns, err := patternAnalyzer.AnalyzePatterns(result.File, result.FileInfo.Package, result.FileInfo.RelPath)
	if err != nil {
		return
	}

	report.Patterns.DesignPatterns.Singleton = append(
		report.Patterns.DesignPatterns.Singleton,
		patterns.Singleton...)
	report.Patterns.DesignPatterns.Factory = append(
		report.Patterns.DesignPatterns.Factory,
		patterns.Factory...)
	report.Patterns.DesignPatterns.Builder = append(
		report.Patterns.DesignPatterns.Builder,
		patterns.Builder...)
	report.Patterns.DesignPatterns.Observer = append(
		report.Patterns.DesignPatterns.Observer,
		patterns.Observer...)
	report.Patterns.DesignPatterns.Strategy = append(
		report.Patterns.DesignPatterns.Strategy,
		patterns.Strategy...)
}

// finalizeReport aggregates metrics
func finalizeReport(report *metrics.Report, collected *collectedMetrics, analyzers *analyzerSet) {
	report.Functions = collected.functions
	report.Structs = collected.structs
	report.Interfaces = collected.interfaces

	if pkgReport, err := analyzers.pkg.GenerateReport(); err == nil {
		report.Packages = pkgReport.Packages
		report.CircularDependencies = pkgReport.CircularDependencies
	}

	if len(collected.files) > 0 {
		dupReport := analyzers.duplication.AnalyzeDuplication(collected.files, 6, 0.8)
		report.Duplication = dupReport
	}

	aggregateGenerics(report, collected)
	calculateComplexityMetrics(report, collected)

	report.Overview = metrics.OverviewMetrics{
		TotalFiles:      len(collected.files),
		TotalFunctions:  len(collected.functions),
		TotalStructs:    len(collected.structs),
		TotalInterfaces: len(collected.interfaces),
	}
}

// calculateComplexityMetrics computes averages
func calculateComplexityMetrics(report *metrics.Report, collected *collectedMetrics) {
	var totalFunctionComplexity float64
	for _, fn := range collected.functions {
		totalFunctionComplexity += fn.Complexity.Overall
	}
	if len(collected.functions) > 0 {
		report.Complexity.AverageFunction = totalFunctionComplexity / float64(len(collected.functions))
	}

	var totalStructComplexity float64
	for _, s := range collected.structs {
		totalStructComplexity += s.Complexity.Overall
	}
	if len(collected.structs) > 0 {
		report.Complexity.AverageStruct = totalStructComplexity / float64(len(collected.structs))
	}
}

// aggregateGenerics merges generic metrics
func aggregateGenerics(report *metrics.Report, collected *collectedMetrics) {
	if len(collected.generics) == 0 {
		return
	}

	merged := createMergedGenerics(collected.generics)
	if len(merged.TypeParameters.Complexity) > 0 {
		merged.ComplexityScore = calculateGenericComplexity(merged.TypeParameters.Complexity)
	}

	report.Generics = merged
}

// createMergedGenerics creates merged result
func createMergedGenerics(generics []metrics.GenericMetrics) metrics.GenericMetrics {
	merged := metrics.GenericMetrics{
		TypeParameters: metrics.GenericTypeParameters{
			Constraints: make(map[string]int),
		},
		ConstraintUsage: make(map[string]int),
	}

	for _, gen := range generics {
		metrics.MergeGenericsData(&merged, gen)
	}
	return merged
}

// mergeGenericsData is deprecated - use metrics.MergeGenericsData instead
// Kept for backward compatibility
func mergeGenericsData(merged *metrics.GenericMetrics, gen metrics.GenericMetrics) {
	metrics.MergeGenericsData(merged, gen)
}

// calculateGenericComplexity calculates score
func calculateGenericComplexity(complexities []metrics.GenericComplexity) float64 {
	total := 0.0
	for _, c := range complexities {
		total += c.ComplexityScore
	}
	return total / float64(len(complexities))
}
