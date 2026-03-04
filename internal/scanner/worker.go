//go:build !js || !wasm

package scanner

import (
	"context"
	"fmt"
	"go/ast"
	"sync"

	"github.com/opd-ai/go-stats-generator/internal/config"
)

// WorkerPool manages concurrent processing of Go files
type WorkerPool struct {
	workerCount int
	discoverer  *Discoverer
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

// NewWorkerPool creates a new worker pool for concurrent file processing with configurable parallelism.
// The worker count is determined by cfg.WorkerCount (defaults to 1 if <= 0). Each worker processes Go source files
// independently, enabling high-throughput analysis of large codebases. Uses the provided discoverer for file discovery.
func NewWorkerPool(cfg *config.PerformanceConfig, discoverer *Discoverer) *WorkerPool {
	workerCount := cfg.WorkerCount
	if workerCount <= 0 {
		workerCount = 1
	}

	return &WorkerPool{
		workerCount: workerCount,
		discoverer:  discoverer,
	}
}

// ProcessFiles processes a list of files concurrently
func (wp *WorkerPool) ProcessFiles(ctx context.Context, files []FileInfo, progressCb ProgressCallback) (<-chan Result, error) {
	if len(files) == 0 {
		return wp.createEmptyChannel(), nil
	}

	jobChan, resultChan := wp.createChannels(len(files))
	wg := wp.startWorkers(ctx, jobChan, resultChan)
	wp.distributeJobs(ctx, jobChan, files)
	wp.closeResultOnCompletion(wg, resultChan)

	return wp.applyProgressTracking(ctx, resultChan, len(files), progressCb), nil
}

// createEmptyChannel returns a closed channel for empty file lists
func (wp *WorkerPool) createEmptyChannel() <-chan Result {
	resultChan := make(chan Result)
	close(resultChan)
	return resultChan
}

// createChannels creates job and result channels with appropriate buffer sizes
func (wp *WorkerPool) createChannels(fileCount int) (chan FileInfo, chan Result) {
	jobChan := make(chan FileInfo, fileCount)
	resultChan := make(chan Result, fileCount)
	return jobChan, resultChan
}

// startWorkers launches worker goroutines and returns the WaitGroup
func (wp *WorkerPool) startWorkers(ctx context.Context, jobChan <-chan FileInfo, resultChan chan<- Result) *sync.WaitGroup {
	var wg sync.WaitGroup
	for i := 0; i < wp.workerCount; i++ {
		wg.Add(1)
		go wp.worker(ctx, &wg, jobChan, resultChan)
	}
	return &wg
}

// distributeJobs sends files to the job channel asynchronously
func (wp *WorkerPool) distributeJobs(ctx context.Context, jobChan chan<- FileInfo, files []FileInfo) {
	go func() {
		defer close(jobChan)
		for _, file := range files {
			select {
			case jobChan <- file:
			case <-ctx.Done():
				return
			}
		}
	}()
}

// closeResultOnCompletion closes the result channel after all workers finish
func (wp *WorkerPool) closeResultOnCompletion(wg *sync.WaitGroup, resultChan chan Result) {
	go func() {
		wg.Wait()
		close(resultChan)
	}()
}

// applyProgressTracking wraps the result channel with progress tracking if callback provided
func (wp *WorkerPool) applyProgressTracking(ctx context.Context, resultChan <-chan Result, fileCount int, progressCb ProgressCallback) <-chan Result {
	if progressCb != nil {
		return wp.trackProgress(ctx, resultChan, fileCount, progressCb)
	}
	return resultChan
}

// worker processes jobs from the job channel
func (wp *WorkerPool) worker(ctx context.Context, wg *sync.WaitGroup, jobChan <-chan FileInfo, resultChan chan<- Result) {
	defer wg.Done()

	for {
		if wp.shouldStopWorker(ctx, jobChan, resultChan) {
			return
		}
	}
}

// shouldStopWorker processes one job or checks for cancellation; returns true if worker should stop
func (wp *WorkerPool) shouldStopWorker(ctx context.Context, jobChan <-chan FileInfo, resultChan chan<- Result) bool {
	select {
	case fileInfo, ok := <-jobChan:
		if !ok {
			return true
		}
		return wp.processAndSendResult(ctx, fileInfo, resultChan)

	case <-ctx.Done():
		return true
	}
}

// processAndSendResult processes a file and sends the result; returns true if worker should stop
func (wp *WorkerPool) processAndSendResult(ctx context.Context, fileInfo FileInfo, resultChan chan<- Result) bool {
	result := wp.processFile(fileInfo)
	return wp.sendResultOrCancel(ctx, result, resultChan)
}

// sendResultOrCancel sends a result or stops on context cancellation; returns true if worker should stop
func (wp *WorkerPool) sendResultOrCancel(ctx context.Context, result Result, resultChan chan<- Result) bool {
	select {
	case resultChan <- result:
		return false
	case <-ctx.Done():
		return true
	}
}

// processFile processes a single file
func (wp *WorkerPool) processFile(fileInfo FileInfo) Result {
	file, err := wp.discoverer.ParseFile(fileInfo.Path)
	if err != nil {
		return Result{
			FileInfo: fileInfo,
			Error:    fmt.Errorf("failed to parse file %s: %w", fileInfo.Path, err),
		}
	}

	return Result{
		FileInfo: fileInfo,
		File:     file,
		Error:    nil,
	}
}

// trackProgress monitors processing progress and calls the callback
func (wp *WorkerPool) trackProgress(ctx context.Context, resultChan <-chan Result, total int, progressCb ProgressCallback) <-chan Result {
	completed := 0

	// Create a new channel to forward results
	forwardChan := make(chan Result, total)

	// Forward results while tracking progress
	go func() {
		defer close(forwardChan)

		for result := range resultChan {
			// Forward the result
			forwardChan <- result

			// Update progress
			completed++
			progressCb(completed, total)
		}

		// Final progress update
		progressCb(total, total)
	}()

	return forwardChan
}

// ProcessFilesSequential processes files one by one (useful for debugging)
func (wp *WorkerPool) ProcessFilesSequential(ctx context.Context, files []FileInfo, progressCb ProgressCallback) (<-chan Result, error) {
	resultChan := make(chan Result, len(files))

	go func() {
		defer close(resultChan)

		for i, fileInfo := range files {
			select {
			case <-ctx.Done():
				return
			default:
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

// BatchProcessor handles batch processing with memory management
type BatchProcessor struct {
	workerPool *WorkerPool
	batchSize  int
}

// NewBatchProcessor creates a batch processor for efficient parallel file analysis with configurable batch sizes.
// It wraps a worker pool to enable chunked processing of large file sets, reducing memory pressure and improving throughput
// for repositories with thousands of files. The batch size controls memory vs. parallelism tradeoff (default 100 files per batch).
// Used by the analyzer to optimize performance when processing enterprise-scale codebases with 10,000+ source files.
func NewBatchProcessor(workerPool *WorkerPool, batchSize int) *BatchProcessor {
	if batchSize <= 0 {
		batchSize = 100
	}

	return &BatchProcessor{
		workerPool: workerPool,
		batchSize:  batchSize,
	}
}

// ProcessInBatches processes files in batches to manage memory usage
// ProcessInBatches processes files in batches using worker pools with progress tracking
func (bp *BatchProcessor) ProcessInBatches(ctx context.Context, files []FileInfo, progressCb ProgressCallback) (<-chan Result, error) {
	totalFiles := len(files)
	if totalFiles == 0 {
		return bp.createEmptyResultChannel(), nil
	}

	resultChan := make(chan Result, totalFiles)
	go bp.processBatchesAsync(ctx, files, progressCb, resultChan)
	return resultChan, nil
}

// createEmptyResultChannel creates and immediately closes a result channel for empty file sets
func (bp *BatchProcessor) createEmptyResultChannel() <-chan Result {
	resultChan := make(chan Result)
	close(resultChan)
	return resultChan
}

// processBatchesAsync handles the asynchronous batch processing workflow
func (bp *BatchProcessor) processBatchesAsync(ctx context.Context, files []FileInfo, progressCb ProgressCallback, resultChan chan<- Result) {
	defer close(resultChan)

	processed := 0
	totalFiles := len(files)

	for i := 0; i < totalFiles; i += bp.batchSize {
		if bp.shouldStopProcessing(ctx) {
			return
		}

		batch := bp.createBatch(files, i, totalFiles)
		if err := bp.processSingleBatch(ctx, batch, &processed, totalFiles, progressCb, resultChan); err != nil {
			return
		}
	}
}

// shouldStopProcessing checks if processing should be stopped due to context cancellation
func (bp *BatchProcessor) shouldStopProcessing(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// createBatch creates a file batch slice from the given range
func (bp *BatchProcessor) createBatch(files []FileInfo, start, totalFiles int) []FileInfo {
	end := start + bp.batchSize
	if end > totalFiles {
		end = totalFiles
	}
	return files[start:end]
}

// processSingleBatch processes a single batch and handles the results
func (bp *BatchProcessor) processSingleBatch(ctx context.Context, batch []FileInfo, processed *int, totalFiles int, progressCb ProgressCallback, resultChan chan<- Result) error {
	// Process batch through worker pool
	batchResults, err := bp.workerPool.ProcessFiles(ctx, batch, nil)
	if err != nil {
		resultChan <- Result{Error: fmt.Errorf("batch processing failed: %w", err)}
		return err
	}

	// Collect and forward batch results
	return bp.collectBatchResults(ctx, batchResults, processed, totalFiles, progressCb, resultChan)
}

// collectBatchResults collects results from a single batch and forwards them
func (bp *BatchProcessor) collectBatchResults(ctx context.Context, batchResults <-chan Result, processed *int, totalFiles int, progressCb ProgressCallback, resultChan chan<- Result) error {
	for result := range batchResults {
		if err := bp.forwardResult(ctx, result, resultChan); err != nil {
			return err
		}
		bp.updateProgress(processed, totalFiles, progressCb)
	}
	return nil
}

// forwardResult forwards a result to the output channel with context cancellation
func (bp *BatchProcessor) forwardResult(ctx context.Context, result Result, resultChan chan<- Result) error {
	select {
	case resultChan <- result:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// updateProgress updates the progress counter and calls the progress callback
func (bp *BatchProcessor) updateProgress(processed *int, totalFiles int, progressCb ProgressCallback) {
	*processed++
	if progressCb != nil {
		progressCb(*processed, totalFiles)
	}
}
