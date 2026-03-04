//go:build !js || !wasm

// Package go-stats-generator provides a programmatic API for analyzing Go source code.
package go_stats_generator

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
