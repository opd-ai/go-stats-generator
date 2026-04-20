package generator

import (
	"context"
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

// NewAnalyzer creates a new analyzer with default configuration settings for comprehensive Go codebase analysis.
// It initializes all internal analyzers (complexity, documentation, naming, etc.) and uses default thresholds.
// Returns an Analyzer ready for immediate use with Analyze() or AnalyzeWithContext() methods.
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		config: config.DefaultConfig(),
	}
}

// NewAnalyzerWithConfig creates analyzer with custom configuration, allowing fine-grained control over
// analysis thresholds (complexity, function length, documentation coverage), performance settings (worker count),
// and feature toggles. Use this when default settings don't match your project's requirements.
func NewAnalyzerWithConfig(cfg *config.Config) *Analyzer {
	return &Analyzer{
		config: cfg,
	}
}

// buildReport processes parsed files and builds a comprehensive metrics report.
// Each result carries its own token.FileSet so that per-file analyzers use the correct
// position data without contending on a shared FileSet mutex.  The PackageAnalyzer and
// DuplicationAnalyzer are shared across all files because they accumulate cross-file state.
func (a *Analyzer) buildReport(ctx context.Context, results <-chan scanner.Result, rootPath string, fileCount int) (*metrics.Report, error) {
	// Shared cross-file analyzers; per-file analyzers are created per result below.
	sharedPkg := analyzer.NewPackageAnalyzer(token.NewFileSet())

	report := createReport(rootPath, fileCount)
	collected := &collectedMetrics{}

	for result := range results {
		if result.Error != nil {
			continue
		}
		processFile(result, sharedPkg, collected, report, a.config)
	}

	finalizeReport(report, collected, sharedPkg)
	return report, nil
}

// collectedMetrics holds cross-file metrics accumulated during the streaming phase.
// Statement blocks are accumulated here and passed to AnalyzeDuplicationFromBlocks
// after all files are processed, avoiding the need to retain any *ast.File in memory.
type collectedMetrics struct {
	functions    []metrics.FunctionMetrics
	structs      []metrics.StructMetrics
	interfaces   []metrics.InterfaceMetrics
	generics     []metrics.GenericMetrics
	dupBlocks    []analyzer.StatementBlock
	dupTotalLines int
	fileCount    int
}

// analysisSet holds per-file analyzers initialized with the result's own token.FileSet.
type analysisSet struct {
	function    *analyzer.FunctionAnalyzer
	structure   *analyzer.StructAnalyzer
	iface       *analyzer.InterfaceAnalyzer
	concurrency *analyzer.ConcurrencyAnalyzer
	pattern     *analyzer.PatternAnalyzer
	duplication *analyzer.DuplicationAnalyzer
	generic     *analyzer.GenericAnalyzer
}

// createPerFileAnalyzers initializes analyzers using the provided per-file FileSet so that
// all position lookups operate on the correct token.File entries without a shared FileSet.
func createPerFileAnalyzers(fset *token.FileSet, cfg *config.Config) *analysisSet {
	return &analysisSet{
		function:    analyzer.NewFunctionAnalyzer(fset),
		structure:   analyzer.NewStructAnalyzer(fset),
		iface:       analyzer.NewInterfaceAnalyzer(fset),
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

// processFile performs per-file analysis using the result's own token.FileSet,
// extracts duplication blocks immediately (so the AST can be reclaimed by the GC),
// and accumulates cross-file metrics via the shared PackageAnalyzer.
func processFile(result scanner.Result, sharedPkg *analyzer.PackageAnalyzer, collected *collectedMetrics, report *metrics.Report, cfg *config.Config) {
	collected.fileCount++
	collected.dupTotalLines += result.FileInfo.FileLines

	// Create per-file analyzers bound to this result's FileSet.
	analyzers := createPerFileAnalyzers(result.FileSet, cfg)

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

	// Extract duplication blocks now and discard the AST reference; the GC can
	// reclaim *ast.File memory once processFile returns instead of waiting until
	// the entire analysis is complete (cf. the former collected.files map approach).
	blocks := analyzers.duplication.ExtractBlocks(result.File, result.FileInfo.RelPath, 6)
	collected.dupBlocks = append(collected.dupBlocks, blocks...)

	sharedPkg.AnalyzePackageWithFileLines(result.File, result.FileInfo.Path, result.FileInfo.FileLines)
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

// finalizeReport aggregates metrics after all files have been processed.
// Duplication analysis operates on the pre-extracted block slices rather than
// the full AST map, keeping peak memory proportional to block count rather
// than to total AST size.
func finalizeReport(report *metrics.Report, collected *collectedMetrics, sharedPkg *analyzer.PackageAnalyzer) {
	report.Functions = collected.functions
	report.Structs = collected.structs
	report.Interfaces = collected.interfaces

	if pkgReport, err := sharedPkg.GenerateReport(); err == nil {
		report.Packages = pkgReport.Packages
		report.CircularDependencies = pkgReport.CircularDependencies
	}

	if len(collected.dupBlocks) > 0 {
		dupAnalyzer := analyzer.NewDuplicationAnalyzer(token.NewFileSet())
		report.Duplication = dupAnalyzer.AnalyzeDuplicationFromBlocks(collected.dupBlocks, collected.dupTotalLines, 0.8)
	}

	aggregateGenerics(report, collected)
	calculateComplexityMetrics(report, collected)

	report.Overview = metrics.OverviewMetrics{
		TotalFiles:      collected.fileCount,
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

// calculateGenericComplexity calculates score
func calculateGenericComplexity(complexities []metrics.GenericComplexity) float64 {
	total := 0.0
	for _, c := range complexities {
		total += c.ComplexityScore
	}
	return total / float64(len(complexities))
}
