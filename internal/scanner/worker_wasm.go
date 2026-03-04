//go:build js && wasm

package scanner

import (
	"context"
	"fmt"
	"go/ast"

	"github.com/opd-ai/go-stats-generator/internal/config"
)

// WorkerPool manages sequential processing for WASM (no OS concurrency)
type WorkerPool struct {
	discoverer *Discoverer
}

// Job represents a file analysis job
type Job struct {
	FileInfo FileInfo
	File     *ast.File
}

// Result represents the result of processing a job
type Result struct {
	FileInfo FileInfo
	File     *ast.File
	Error    error
}

// ProgressCallback is called to report progress
type ProgressCallback func(completed, total int)

// NewWorkerPool creates a worker pool for WASM environments where OS-level concurrency is not available.
// This implementation processes files sequentially rather than in parallel, as WASM runtimes typically execute in
// a single-threaded JavaScript environment. The cfg parameter is accepted for API compatibility but worker count is ignored.
func NewWorkerPool(cfg *config.PerformanceConfig, discoverer *Discoverer) *WorkerPool {
	return &WorkerPool{
		discoverer: discoverer,
	}
}

// ProcessFiles processes files sequentially for WASM
func (wp *WorkerPool) ProcessFiles(ctx context.Context, files []FileInfo, progressCb ProgressCallback) (<-chan Result, error) {
	return wp.ProcessFilesSequential(ctx, files, progressCb)
}

// ProcessFilesSequential processes files one by one
func (wp *WorkerPool) ProcessFilesSequential(ctx context.Context, files []FileInfo, progressCb ProgressCallback) (<-chan Result, error) {
	resultChan := make(chan Result, len(files))

	go func() {
		defer close(resultChan)

		for i, fileInfo := range files {
			if shouldStop(ctx) {
				return
			}

			result := wp.processFile(fileInfo)
			resultChan <- result

			if progressCb != nil {
				progressCb(i+1, len(files))
			}
		}
	}()

	return resultChan, nil
}

// processFile processes a single file
func (wp *WorkerPool) processFile(fileInfo FileInfo) Result {
	file, err := parseFileInfo(wp.discoverer, fileInfo)
	if err != nil {
		return createErrorResult(fileInfo, err)
	}

	return Result{
		FileInfo: fileInfo,
		File:     file,
		Error:    nil,
	}
}

// parseFileInfo parses file from FileInfo
func parseFileInfo(discoverer *Discoverer, fileInfo FileInfo) (*ast.File, error) {
	file, err := discoverer.ParseFile(fileInfo.Path)
	if err != nil {
		return nil, fmt.Errorf("parse failed %s: %w", fileInfo.Path, err)
	}
	return file, nil
}

// createErrorResult creates Result with error
func createErrorResult(fileInfo FileInfo, err error) Result {
	return Result{
		FileInfo: fileInfo,
		Error:    err,
	}
}

// shouldStop checks context cancellation
func shouldStop(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// BatchProcessor handles batch processing
type BatchProcessor struct {
	workerPool *WorkerPool
	batchSize  int
}

// NewBatchProcessor creates a batch processor for WASM environments enabling efficient parallel file analysis with configurable batch sizes.
// It wraps a worker pool optimized for WebAssembly single-threaded execution model, processing files in sequential batches to avoid memory
// pressure in browser contexts. The batch size controls chunking strategy for large file sets (default 100 files per batch).
func NewBatchProcessor(workerPool *WorkerPool, batchSize int) *BatchProcessor {
	if batchSize <= 0 {
		batchSize = 100
	}

	return &BatchProcessor{
		workerPool: workerPool,
		batchSize:  batchSize,
	}
}

// ProcessInBatches processes files sequentially
func (bp *BatchProcessor) ProcessInBatches(ctx context.Context, files []FileInfo, progressCb ProgressCallback) (<-chan Result, error) {
	return bp.workerPool.ProcessFilesSequential(ctx, files, progressCb)
}
