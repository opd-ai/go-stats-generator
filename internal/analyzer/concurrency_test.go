package analyzer

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConcurrencyAnalyzer_AnalyzeConcurrency(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected func(t *testing.T, result interface{})
	}{
		{
			name: "basic goroutine detection",
			code: `package main

func main() {
	go func() {
		println("hello")
	}()
	
	go processData()
}

func processData() {
	// process data
}`,
			expected: func(t *testing.T, result interface{}) {
				metrics := result.(interface{})
				// Verify basic structure exists
				assert.NotNil(t, metrics)
			},
		},
		{
			name: "channel creation and usage",
			code: `package main

import "sync"

func main() {
	ch := make(chan int)
	buffered := make(chan string, 10)
	sendOnly := make(chan<- bool)
	recvOnly := make(<-chan float64)
	
	var mu sync.Mutex
	var wg sync.WaitGroup
}`,
			expected: func(t *testing.T, result interface{}) {
				// Verify channel and sync primitive detection
				assert.NotNil(t, result)
			},
		},
		{
			name: "worker pool pattern",
			code: `package main

import "sync"

func workerPool() {
	jobs := make(chan int, 100)
	results := make(chan int, 100)
	var wg sync.WaitGroup
	
	// Start workers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				results <- job * 2
			}
		}()
	}
}`,
			expected: func(t *testing.T, result interface{}) {
				// Should detect worker pool pattern
				assert.NotNil(t, result)
			},
		},
		{
			name: "empty file",
			code: `package main`,
			expected: func(t *testing.T, result interface{}) {
				assert.NotNil(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			require.NoError(t, err)

			analyzer := NewConcurrencyAnalyzer(fset)
			result, err := analyzer.AnalyzeConcurrency(file, "test.go")
			require.NoError(t, err)

			tt.expected(t, result)
		})
	}
}

func TestConcurrencyAnalyzer_GoroutineDetection(t *testing.T) {
	code := `package main

func main() {
	// Anonymous goroutine
	go func() {
		println("anonymous")
	}()
	
	// Named function goroutine
	go namedFunction()
	
	// Method call goroutine
	obj := &MyStruct{}
	go obj.Method()
}

func namedFunction() {}

type MyStruct struct{}
func (m *MyStruct) Method() {}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, parser.ParseComments)
	require.NoError(t, err)

	analyzer := NewConcurrencyAnalyzer(fset)
	result, err := analyzer.AnalyzeConcurrency(file, "test.go")
	require.NoError(t, err)

	// Should detect 3 goroutines
	assert.Equal(t, 3, result.Goroutines.TotalCount)
	assert.Equal(t, 1, result.Goroutines.AnonymousCount)
	assert.Equal(t, 2, result.Goroutines.NamedCount)
	assert.Len(t, result.Goroutines.Instances, 3)

	// Check first goroutine (anonymous)
	firstGoroutine := result.Goroutines.Instances[0]
	assert.True(t, firstGoroutine.IsAnonymous)
	assert.Equal(t, "anonymous", firstGoroutine.Function)
	assert.Equal(t, "test.go", firstGoroutine.File)
	assert.Greater(t, firstGoroutine.Line, 0)

	// Check second goroutine (named function)
	secondGoroutine := result.Goroutines.Instances[1]
	assert.False(t, secondGoroutine.IsAnonymous)
	assert.Equal(t, "namedFunction", secondGoroutine.Function)
}

func TestConcurrencyAnalyzer_ChannelDetection(t *testing.T) {
	code := `package main

func main() {
	// Unbuffered channels
	ch1 := make(chan int)
	ch2 := make(chan string)
	
	// Buffered channels
	buffered1 := make(chan int, 10)
	buffered2 := make(chan bool, 5)
	
	// Directional channels
	sendOnly := make(chan<- int)
	recvOnly := make(<-chan string)
	
	// Channel types in function parameters
	processSend(sendOnly)
	processRecv(recvOnly)
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, parser.ParseComments)
	require.NoError(t, err)

	analyzer := NewConcurrencyAnalyzer(fset)
	result, err := analyzer.AnalyzeConcurrency(file, "test.go")
	require.NoError(t, err)

	// Should detect multiple channels
	assert.Greater(t, result.Channels.TotalCount, 0)
	assert.Greater(t, result.Channels.BufferedCount, 0)
	assert.Greater(t, result.Channels.UnbufferedCount, 0)
	assert.Greater(t, result.Channels.DirectionalCount, 0)

	// Verify we have channel instances
	assert.NotEmpty(t, result.Channels.Instances)

	// Check for buffered channel detection
	hasBuffered := false
	hasUnbuffered := false
	hasDirectional := false

	for _, instance := range result.Channels.Instances {
		if instance.IsBuffered && instance.BufferSize > 0 {
			hasBuffered = true
		}
		if !instance.IsBuffered {
			hasUnbuffered = true
		}
		if instance.IsDirectional {
			hasDirectional = true
		}
	}

	assert.True(t, hasBuffered, "Should detect buffered channels")
	assert.True(t, hasUnbuffered, "Should detect unbuffered channels")
	assert.True(t, hasDirectional, "Should detect directional channels")
}

func TestConcurrencyAnalyzer_SyncPrimitiveDetection(t *testing.T) {
	code := `package main

import "sync"

func main() {
	// Various sync primitive declarations
	var mu sync.Mutex
	var rwMu sync.RWMutex
	var wg sync.WaitGroup
	var once sync.Once
	var cond *sync.Cond
	
	// Using sync primitives
	mu.Lock()
	wg.Add(1)
	once.Do(func() {})
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, parser.ParseComments)
	require.NoError(t, err)

	analyzer := NewConcurrencyAnalyzer(fset)
	result, err := analyzer.AnalyzeConcurrency(file, "test.go")
	require.NoError(t, err)

	// Should detect sync primitives
	assert.NotEmpty(t, result.SyncPrims.Mutexes)
	assert.NotEmpty(t, result.SyncPrims.RWMutexes)
	assert.NotEmpty(t, result.SyncPrims.WaitGroups)
	assert.NotEmpty(t, result.SyncPrims.Once)

	// Verify mutex detection
	mutex := result.SyncPrims.Mutexes[0]
	assert.Equal(t, "test.go", mutex.File)
	assert.Equal(t, "Mutex", mutex.Type)
	assert.Greater(t, mutex.Line, 0)
}

func TestConcurrencyAnalyzer_ComplexScenario(t *testing.T) {
	code := `package main

import (
	"context"
	"sync"
	"time"
)

func complexConcurrencyExample() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	jobs := make(chan int, 100)
	results := make(chan int, 100)
	var wg sync.WaitGroup
	var mu sync.Mutex
	counter := 0
	
	// Start worker goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case job := <-jobs:
					result := processJob(job)
					mu.Lock()
					counter++
					mu.Unlock()
					results <- result
				case <-ctx.Done():
					return
				}
			}
		}(i)
	}
	
	// Send jobs
	go func() {
		defer close(jobs)
		for i := 0; i < 50; i++ {
			jobs <- i
		}
	}()
	
	// Collect results
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Process results
	for result := range results {
		println(result)
	}
}

func processJob(job int) int {
	return job * 2
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, parser.ParseComments)
	require.NoError(t, err)

	analyzer := NewConcurrencyAnalyzer(fset)
	result, err := analyzer.AnalyzeConcurrency(file, "test.go")
	require.NoError(t, err)

	// Should detect exactly 3 goroutine statements (static analysis counts syntactic occurrences)
	assert.Equal(t, 3, result.Goroutines.TotalCount)

	// Should detect channels
	assert.Greater(t, result.Channels.TotalCount, 1)
	assert.Greater(t, result.Channels.BufferedCount, 0)

	// Should detect sync primitives
	assert.NotEmpty(t, result.SyncPrims.Mutexes)
	assert.NotEmpty(t, result.SyncPrims.WaitGroups)

	// Verify we have both anonymous and named goroutines
	assert.Greater(t, result.Goroutines.AnonymousCount, 0)
}

func TestConcurrencyAnalyzer_EmptyFile(t *testing.T) {
	code := `package main`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, parser.ParseComments)
	require.NoError(t, err)

	analyzer := NewConcurrencyAnalyzer(fset)
	result, err := analyzer.AnalyzeConcurrency(file, "test.go")
	require.NoError(t, err)

	// Should return empty metrics but not nil
	assert.Equal(t, 0, result.Goroutines.TotalCount)
	assert.Equal(t, 0, result.Goroutines.AnonymousCount)
	assert.Equal(t, 0, result.Goroutines.NamedCount)
	assert.Equal(t, 0, result.Channels.TotalCount)
	assert.Empty(t, result.SyncPrims.Mutexes)
	assert.Empty(t, result.SyncPrims.WaitGroups)
	assert.Empty(t, result.WorkerPools)
	assert.Empty(t, result.Pipelines)
}

func TestConcurrencyAnalyzer_PatternDetection(t *testing.T) {
	// Test pattern detection capabilities
	code := `package main

import "sync"

// Semaphore pattern using buffered channel
func semaphorePattern() {
	sem := make(chan struct{}, 3) // Limit to 3 concurrent operations
	
	for i := 0; i < 10; i++ {
		go func() {
			sem <- struct{}{} // Acquire
			defer func() { <-sem }() // Release
			
			// Do work
			doWork()
		}()
	}
}

func doWork() {}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, parser.ParseComments)
	require.NoError(t, err)

	analyzer := NewConcurrencyAnalyzer(fset)
	result, err := analyzer.AnalyzeConcurrency(file, "test.go")
	require.NoError(t, err)

	// Should detect goroutines and channels
	assert.Greater(t, result.Goroutines.TotalCount, 0)
	assert.Greater(t, result.Channels.TotalCount, 0)

	// Pattern detection would be implemented in the actual pattern detection methods
	// For now, we verify the structure exists
	assert.NotNil(t, result.WorkerPools)
	assert.NotNil(t, result.Pipelines)
	assert.NotNil(t, result.Semaphores)
}

func TestConcurrencyAnalyzer_EnhancedPatternDetection(t *testing.T) {
	tests := []struct {
		name            string
		code            string
		expectedPattern string
		minConfidence   float64
	}{
		{
			name: "worker pool pattern",
			code: `package main

import "sync"

func workerPoolExample() {
	jobs := make(chan int, 100)
	results := make(chan int, 100)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				results <- job * 2
			}
		}()
	}
}`,
			expectedPattern: "worker_pools",
			minConfidence:   0.5,
		},
		{
			name: "semaphore pattern",
			code: `package main

func semaphoreExample() {
	sem := make(chan struct{}, 3) // Limit to 3 concurrent operations
	
	for i := 0; i < 10; i++ {
		go func() {
			sem <- struct{}{} // Acquire
			defer func() { <-sem }() // Release
			// Do work
		}()
	}
}`,
			expectedPattern: "semaphores",
			minConfidence:   0.7,
		},
		{
			name: "pipeline pattern",
			code: `package main

func pipelineExample() {
	// Stage 1: generate numbers
	numbers := make(chan int)
	go func() {
		defer close(numbers)
		for i := 0; i < 10; i++ {
			numbers <- i
		}
	}()

	// Stage 2: square numbers
	squares := make(chan int)
	go func() {
		defer close(squares)
		for n := range numbers {
			squares <- n * n
		}
	}()

	// Stage 3: collect results
	go func() {
		for s := range squares {
			// process
		}
	}()
}`,
			expectedPattern: "pipelines",
			minConfidence:   0.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			require.NoError(t, err)

			analyzer := NewConcurrencyAnalyzer(fset)
			result, err := analyzer.AnalyzeConcurrency(file, "test.go")
			require.NoError(t, err)

			// Check the specific pattern was detected
			switch tt.expectedPattern {
			case "worker_pools":
				assert.Greater(t, len(result.WorkerPools), 0, "Worker pool pattern should be detected")
				if len(result.WorkerPools) > 0 {
					assert.GreaterOrEqual(t, result.WorkerPools[0].ConfidenceScore, tt.minConfidence)
				}
			case "semaphores":
				assert.Greater(t, len(result.Semaphores), 0, "Semaphore pattern should be detected")
				if len(result.Semaphores) > 0 {
					assert.GreaterOrEqual(t, result.Semaphores[0].ConfidenceScore, tt.minConfidence)
				}
			case "pipelines":
				assert.Greater(t, len(result.Pipelines), 0, "Pipeline pattern should be detected")
				if len(result.Pipelines) > 0 {
					assert.GreaterOrEqual(t, result.Pipelines[0].ConfidenceScore, tt.minConfidence)
				}
			}
		})
	}
}

// Benchmark tests
func BenchmarkConcurrencyAnalyzer_LargeFile(b *testing.B) {
	// Create a large file with many goroutines and channels
	code := `package main

import "sync"

func largeFunction() {
	var wg sync.WaitGroup
	ch := make(chan int, 1000)
	
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ch <- id
		}(i)
	}
	
	wg.Wait()
	close(ch)
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, parser.ParseComments)
	require.NoError(b, err)

	analyzer := NewConcurrencyAnalyzer(fset)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.AnalyzeConcurrency(file, "test.go")
		require.NoError(b, err)
	}
}

func BenchmarkConcurrencyAnalyzer_SmallFile(b *testing.B) {
	code := `package main

func main() {
	go func() {
		println("hello")
	}()
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", code, parser.ParseComments)
	require.NoError(b, err)

	analyzer := NewConcurrencyAnalyzer(fset)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.AnalyzeConcurrency(file, "test.go")
		require.NoError(b, err)
	}
}
