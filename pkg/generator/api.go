package generator

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/opd-ai/go-stats-generator/internal/scanner"
)

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

	return a.buildReport(ctx, results, discoverer.GetFileSet(), absPath, len(files))
}

// AnalyzeFile analyzes a single Go source file and produces a comprehensive metrics report for that file only.
// It parses the file AST, runs all configured analyzers (functions, structs, documentation, complexity, patterns),
// and aggregates results into a full metrics report structure. This method is used for file-scoped analysis workflows
// and enables integration with editors/IDEs for real-time code quality feedback on individual files.
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

	return a.buildReport(ctx, results, discoverer.GetFileSet(), filePath, 1)
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
