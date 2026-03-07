//go:build js && wasm

package generator

import (
	"context"
	"fmt"

	"github.com/opd-ai/go-stats-generator/internal/metrics"
	"github.com/opd-ai/go-stats-generator/internal/scanner"
)

// MemoryFile represents a file in memory for WASM
type MemoryFile struct {
	Path    string
	Content string
}

// AnalyzeMemoryFiles analyzes files from memory (WASM only)
func (a *Analyzer) AnalyzeMemoryFiles(ctx context.Context, files []MemoryFile, rootDir string) (*metrics.Report, error) {
	memFiles := convertToMemoryFiles(files)

	discoverer := scanner.NewDiscoverer(&a.config.Filters)
	fileInfos, err := discoverer.DiscoverFilesFromMemory(memFiles, rootDir)
	if err != nil {
		return nil, err
	}

	results, err := processMemoryFiles(ctx, discoverer, fileInfos, files)
	if err != nil {
		return nil, err
	}

	return a.buildReport(ctx, results, discoverer.GetFileSet(), rootDir, len(fileInfos))
}

// convertToMemoryFiles converts API MemoryFile to scanner MemoryFile
func convertToMemoryFiles(files []MemoryFile) []scanner.MemoryFile {
	result := make([]scanner.MemoryFile, len(files))
	for i, f := range files {
		result[i] = scanner.MemoryFile{
			Path:    f.Path,
			Content: f.Content,
		}
	}
	return result
}

// processMemoryFiles parses and processes memory files
func processMemoryFiles(ctx context.Context, discoverer *scanner.Discoverer, fileInfos []scanner.FileInfo, files []MemoryFile) (<-chan scanner.Result, error) {
	resultChan := make(chan scanner.Result, len(fileInfos))

	go func() {
		defer close(resultChan)
		processFilesSequentially(ctx, discoverer, fileInfos, files, resultChan)
	}()

	return resultChan, nil
}

// processFilesSequentially processes files one by one
func processFilesSequentially(ctx context.Context, discoverer *scanner.Discoverer, fileInfos []scanner.FileInfo, files []MemoryFile, resultChan chan<- scanner.Result) {
	contentMap := createContentMap(files)

	for _, fileInfo := range fileInfos {
		if shouldStopProcessing(ctx) {
			return
		}

		result := parseSingleFile(discoverer, fileInfo, contentMap)
		resultChan <- result
	}
}

// createContentMap creates path to content mapping
func createContentMap(files []MemoryFile) map[string]string {
	contentMap := make(map[string]string)
	for _, f := range files {
		contentMap[f.Path] = f.Content
	}
	return contentMap
}

// parseSingleFile parses a single memory file
func parseSingleFile(discoverer *scanner.Discoverer, fileInfo scanner.FileInfo, contentMap map[string]string) scanner.Result {
	content, exists := contentMap[fileInfo.Path]
	if !exists {
		return createParseError(fileInfo, "content not found")
	}

	file, err := discoverer.ParseMemoryFile(fileInfo.Path, content)
	if err != nil {
		return createParseError(fileInfo, err.Error())
	}

	return scanner.Result{
		FileInfo: fileInfo,
		File:     file,
		Error:    nil,
	}
}

// createParseError creates Result with error
func createParseError(fileInfo scanner.FileInfo, msg string) scanner.Result {
	return scanner.Result{
		FileInfo: fileInfo,
		Error:    fmt.Errorf("%s", msg),
	}
}

// shouldStopProcessing checks for context cancellation
func shouldStopProcessing(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
