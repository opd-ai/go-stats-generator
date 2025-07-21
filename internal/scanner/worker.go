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

// NewWorkerPool creates a new worker pool
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
		resultChan := make(chan Result)
		close(resultChan)
		return resultChan, nil
	}

	// Create job channel
	jobChan := make(chan FileInfo, len(files))
	resultChan := make(chan Result, len(files))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < wp.workerCount; i++ {
		wg.Add(1)
		go wp.worker(ctx, &wg, jobChan, resultChan)
	}

	// Start progress tracker if callback provided
	if progressCb != nil {
		go wp.trackProgress(ctx, resultChan, len(files), progressCb)
	}

	// Send jobs
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

	// Close result channel when all workers finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	return resultChan, nil
}

// worker processes jobs from the job channel
func (wp *WorkerPool) worker(ctx context.Context, wg *sync.WaitGroup, jobChan <-chan FileInfo, resultChan chan<- Result) {
	defer wg.Done()

	for {
		select {
		case fileInfo, ok := <-jobChan:
			if !ok {
				return
			}

			result := wp.processFile(fileInfo)

			select {
			case resultChan <- result:
			case <-ctx.Done():
				return
			}

		case <-ctx.Done():
			return
		}
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

// NewBatchProcessor creates a new batch processor
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
func (bp *BatchProcessor) ProcessInBatches(ctx context.Context, files []FileInfo, progressCb ProgressCallback) (<-chan Result, error) {
	totalFiles := len(files)
	if totalFiles == 0 {
		resultChan := make(chan Result)
		close(resultChan)
		return resultChan, nil
	}

	resultChan := make(chan Result, totalFiles)

	go func() {
		defer close(resultChan)

		processed := 0

		for i := 0; i < totalFiles; i += bp.batchSize {
			end := i + bp.batchSize
			if end > totalFiles {
				end = totalFiles
			}

			batch := files[i:end]

			// Process batch
			batchResults, err := bp.workerPool.ProcessFiles(ctx, batch, nil)
			if err != nil {
				// Send error result
				resultChan <- Result{Error: fmt.Errorf("batch processing failed: %w", err)}
				return
			}

			// Collect batch results
			for result := range batchResults {
				select {
				case resultChan <- result:
					processed++
					if progressCb != nil {
						progressCb(processed, totalFiles)
					}
				case <-ctx.Done():
					return
				}
			}

			// Check for cancellation between batches
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}()

	return resultChan, nil
}
