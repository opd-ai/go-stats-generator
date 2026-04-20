package generator

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
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

	return a.buildReport(ctx, results, absPath, len(files))
}

// AnalyzeFile analyzes a single Go source file and produces a comprehensive metrics report for that file only.
// It parses the file AST, runs all configured analyzers (functions, structs, documentation, complexity, patterns),
// and aggregates results into a full metrics report structure. This method is used for file-scoped analysis workflows
// and enables integration with editors/IDEs for real-time code quality feedback on individual files.
func (a *Analyzer) AnalyzeFile(ctx context.Context, filePath string) (*metrics.Report, error) {
	fileInfo := createFileInfo(filePath)

	// Parse using a dedicated per-file FileSet so that position lookups in per-file analyzers
	// work correctly without depending on the discoverer's shared FileSet.
	localFset := token.NewFileSet()
	file, err := parseFileForAnalysis(localFset, filePath, fileInfo.Src)
	if err != nil {
		return nil, err
	}
	fileInfo.Src = nil // release bytes after parsing

	result := scanner.Result{FileInfo: fileInfo, File: file, FileSet: localFset, Error: nil}
	results := make(chan scanner.Result, 1)
	results <- result
	close(results)

	return a.buildReport(ctx, results, filePath, 1)
}

// createFileInfo creates FileInfo for a single file analysis, reading and caching
// the file bytes so that the caller can reuse them for parsing without a second disk read.
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
	if src, err := os.ReadFile(filePath); err == nil {
		info.Src = src
		info.FileLines = bytes.Count(src, []byte{'\n'}) + 1
	}
	return info
}

// parseFileForAnalysis parses a Go source file using the provided FileSet.
// It uses src bytes when available to avoid a second disk read.
func parseFileForAnalysis(fset *token.FileSet, filePath string, src []byte) (*ast.File, error) {
	var srcArg interface{}
	if len(src) > 0 {
		srcArg = src
	}
	file, err := parser.ParseFile(fset, filePath, srcArg, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}
	return file, nil
}
