// Package gostats provides a programmatic API for analyzing Go source code.
package gostats

import (
	"context"
	"path/filepath"

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
	// Convert to absolute path
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	// Discover files
	discoverer := scanner.NewDiscoverer(&a.config.Filters)
	files, err := discoverer.DiscoverFiles(absPath)
	if err != nil {
		return nil, err
	}

	// Create worker pool and process files
	workerPool := scanner.NewWorkerPool(&a.config.Performance, discoverer)
	results, err := workerPool.ProcessFiles(ctx, files, nil)
	if err != nil {
		return nil, err
	}

	// Analyze results
	functionAnalyzer := analyzer.NewFunctionAnalyzer(discoverer.GetFileSet())

	report := &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:     absPath,
			FilesProcessed: len(files),
			ToolVersion:    "1.0.0",
		},
	}

	var allFunctions []metrics.FunctionMetrics

	// Process analysis results
	for result := range results {
		if result.Error != nil {
			continue
		}

		// Analyze functions in this file
		functions, err := functionAnalyzer.AnalyzeFunctions(result.File, result.FileInfo.Package)
		if err != nil {
			continue
		}

		allFunctions = append(allFunctions, functions...)
	}

	// Populate report
	report.Functions = allFunctions
	report.Overview = metrics.OverviewMetrics{
		TotalFiles:     len(files),
		TotalFunctions: len(allFunctions),
	}

	return report, nil
}

// AnalyzeFile analyzes a single Go file
func (a *Analyzer) AnalyzeFile(ctx context.Context, filePath string) (*metrics.Report, error) {
	discoverer := scanner.NewDiscoverer(&a.config.Filters)

	file, err := discoverer.ParseFile(filePath)
	if err != nil {
		return nil, err
	}

	functionAnalyzer := analyzer.NewFunctionAnalyzer(discoverer.GetFileSet())

	// Extract package name from file
	pkgName := ""
	if file.Name != nil {
		pkgName = file.Name.Name
	}

	functions, err := functionAnalyzer.AnalyzeFunctions(file, pkgName)
	if err != nil {
		return nil, err
	}

	report := &metrics.Report{
		Metadata: metrics.ReportMetadata{
			Repository:     filePath,
			FilesProcessed: 1,
			ToolVersion:    "1.0.0",
		},
		Functions: functions,
		Overview: metrics.OverviewMetrics{
			TotalFiles:     1,
			TotalFunctions: len(functions),
		},
	}

	return report, nil
}
