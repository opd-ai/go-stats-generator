// Package go-stats-generator provides a programmatic API for analyzing Go source code.
package go_stats_generator

import (
	"context"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"
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

// NewAnalyzer creates a new analyzer with default configuration
func NewAnalyzer() *Analyzer {
	return &Analyzer{
		config: config.DefaultConfig(),
	}
}

// NewAnalyzerWithConfig creates a new analyzer with custom configuration
func NewAnalyzerWithConfig(cfg *config.Config) *Analyzer {
	return &Analyzer{
		config: cfg,
	}
}

// AnalyzeDirectory analyzes all Go files in the specified directory
func (a *Analyzer) AnalyzeDirectory(ctx context.Context, dir string) (*metrics.Report, error) {
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	discoverer := scanner.NewDiscoverer(&a.config.Filters)
	files, err := discoverer.DiscoverFiles(absPath)
	if err != nil {
		return nil, err
	}

	workerPool := scanner.NewWorkerPool(&a.config.Performance, discoverer)
	results, err := workerPool.ProcessFiles(ctx, files, nil)
	if err != nil {
		return nil, err
	}

	return a.analyzeResults(ctx, results, discoverer.GetFileSet(), absPath, len(files))
}

// AnalyzeFile analyzes a single Go file
func (a *Analyzer) AnalyzeFile(ctx context.Context, filePath string) (*metrics.Report, error) {
	discoverer := scanner.NewDiscoverer(&a.config.Filters)
	file, err := discoverer.ParseFile(filePath)
	if err != nil {
		return nil, err
	}

	fileInfo := createFileInfo(filePath)
	result := scanner.Result{FileInfo: fileInfo, File: file, Error: nil}
	results := make(chan scanner.Result, 1)
	results <- result
	close(results)

	return a.analyzeResults(ctx, results, discoverer.GetFileSet(), filePath, 1)
}

// createFileInfo creates FileInfo for a single file analysis
func createFileInfo(filePath string) scanner.FileInfo {
	info := scanner.FileInfo{
		Path:        filePath,
		RelPath:     filepath.Base(filePath),
		IsTestFile:  strings.HasSuffix(filePath, "_test.go"),
		IsGenerated: false,
	}
	if fileInfo, err := os.Stat(filePath); err == nil {
		info.Size = fileInfo.Size()
	}
	return info
}

// analyzeResults performs comprehensive analysis on scanner results
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

// collectedMetrics holds metrics collected during analysis
type collectedMetrics struct {
	functions  []metrics.FunctionMetrics
	structs    []metrics.StructMetrics
	interfaces []metrics.InterfaceMetrics
	files      map[string]*ast.File
}

// analyzerSet holds all analyzers for comprehensive analysis
type analyzerSet struct {
	function    *analyzer.FunctionAnalyzer
	structure   *analyzer.StructAnalyzer
	iface       *analyzer.InterfaceAnalyzer
	pkg         *analyzer.PackageAnalyzer
	concurrency *analyzer.ConcurrencyAnalyzer
	duplication *analyzer.DuplicationAnalyzer
}

// createAnalyzers initializes all analyzers for comprehensive analysis
func createAnalyzers(fset *token.FileSet, cfg *config.Config) *analyzerSet {
	return &analyzerSet{
		function:    analyzer.NewFunctionAnalyzer(fset),
		structure:   analyzer.NewStructAnalyzer(fset),
		iface:       analyzer.NewInterfaceAnalyzer(fset),
		pkg:         analyzer.NewPackageAnalyzer(fset),
		concurrency: analyzer.NewConcurrencyAnalyzer(fset),
		duplication: analyzer.NewDuplicationAnalyzer(fset),
	}
}

// createReport creates initial report structure
func createReport(rootPath string, fileCount int) *metrics.Report {
	return &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:     rootPath,
			GeneratedAt:    time.Now(),
			FilesProcessed: fileCount,
			ToolVersion:    "1.0.0",
		},
		Patterns: metrics.PatternMetrics{
			ConcurrencyPatterns: metrics.ConcurrencyPatternMetrics{},
		},
	}
}

// processFile performs all analysis types on a single file
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

	analyzers.pkg.AnalyzePackage(result.File, result.FileInfo.Path)
	analyzeConcurrency(result, analyzers.concurrency, report)
}

// analyzeConcurrency analyzes concurrency patterns and aggregates to report
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

// finalizeReport aggregates all collected metrics into the report
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

	report.Overview = metrics.OverviewMetrics{
		TotalFiles:      len(collected.files),
		TotalFunctions:  len(collected.functions),
		TotalStructs:    len(collected.structs),
		TotalInterfaces: len(collected.interfaces),
	}
}
