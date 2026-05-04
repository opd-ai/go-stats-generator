package scanner

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
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
	// FileSet is the token.FileSet into which File was parsed.
	// Callers must use this FileSet (not any shared one) for all position lookups
	// on nodes belonging to File, enabling concurrent workers to each use their
	// own FileSet and eliminating contention on a single shared mutex.
	FileSet *token.FileSet
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

// createChannels creates job and result channels with appropriate buffer sizes.
// Buffers are capped at workerCount*2 to avoid allocating channel descriptor memory proportional
// to the total file count. A small buffer keeps workers fed without holding all file metadata
// (and their cached source bytes) in memory simultaneously.
func (wp *WorkerPool) createChannels(fileCount int) (chan FileInfo, chan Result) {
	bufSize := wp.workerCount * 2
	if bufSize > fileCount {
		bufSize = fileCount
	}
	jobChan := make(chan FileInfo, bufSize)
	resultChan := make(chan Result, bufSize)
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

// processFile processes a single file using a per-file token.FileSet.
// Each invocation creates its own FileSet so that concurrent workers do not
// contend on the shared FileSet mutex in token.FileSet.AddFile.  The per-file
// FileSet is stored in the returned Result so that downstream analyzers can
// call fset.Position on AST nodes without needing a shared FileSet.
// After parsing, Src is cleared to release the cached bytes and reduce peak memory pressure.
func (wp *WorkerPool) processFile(fileInfo FileInfo) Result {
	// Give each file its own FileSet to eliminate shared-mutex contention during parsing.
	localFset := token.NewFileSet()

	var (
		file *ast.File
		err  error
	)

	if fileInfo.Src != nil {
		file, err = parser.ParseFile(localFset, fileInfo.Path, fileInfo.Src, parser.ParseComments)
		if err != nil {
			err = fmt.Errorf("failed to parse file %s: %w", fileInfo.Path, err)
		}
	} else {
		file, err = parser.ParseFile(localFset, fileInfo.Path, nil, parser.ParseComments)
		if err != nil {
			err = fmt.Errorf("failed to parse file %s: %w", fileInfo.Path, err)
		}
	}

	// Release the cached bytes now that parsing is done.
	fileInfo.Src = nil

	if err != nil {
		return Result{
			FileInfo: fileInfo,
			FileSet:  localFset,
			Error:    err,
		}
	}

	return Result{
		FileInfo: fileInfo,
		File:     file,
		FileSet:  localFset,
		Error:    nil,
	}
}

// trackProgress monitors processing progress and calls the callback
func (wp *WorkerPool) trackProgress(ctx context.Context, resultChan <-chan Result, total int, progressCb ProgressCallback) <-chan Result {
	completed := 0

	// Use the same bounded buffer size as createChannels to avoid pre-allocating total slots.
	bufSize := wp.workerCount * 2
	if bufSize > total {
		bufSize = total
	}
	forwardChan := make(chan Result, bufSize)

	// Forward results while tracking progress.
	// ctx.Done() is wired into both the receive and the send so this goroutine
	// exits promptly when the context is cancelled even if forwardChan is full.
	go func() {
		defer close(forwardChan)

		for {
			select {
			case result, ok := <-resultChan:
				if !ok {
					// Upstream channel closed; emit the final progress update and exit.
					progressCb(total, total)
					return
				}
				select {
				case forwardChan <- result:
				case <-ctx.Done():
					return
				}
				completed++
				progressCb(completed, total)
			case <-ctx.Done():
				return
			}
		}
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

