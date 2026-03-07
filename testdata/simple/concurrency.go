package concurrency

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// WorkerPoolExample demonstrates the worker pool concurrency pattern with 5 workers processing jobs from a channel.
// Creates a fixed number of worker goroutines that consume jobs from a shared channel and send results to an output channel.
func WorkerPoolExample() {
	const numWorkers = 5
	jobs := make(chan int, 100)
	results := make(chan int, 100)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for job := range jobs {
				results <- job * 2
			}
		}(i)
	}

	// Send jobs
	go func() {
		for i := 0; i < 20; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	// Wait for workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	for result := range results {
		_ = result
	}
}

// PipelineExample demonstrates a multi-stage pipeline pattern where data flows through sequential processing stages.
// Numbers are generated in stage 1, squared in stage 2, and collected in stage 3 using channels to connect stages.
func PipelineExample() {
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
	for s := range squares {
		_ = s
	}
}

// FanOutExample demonstrates the fan-out pattern where a single input channel distributes work to multiple output channels.
// One goroutine sends data to an input channel while three worker goroutines consume from separate output channels.
func FanOutExample() {
	input := make(chan int)
	output1 := make(chan int)
	output2 := make(chan int)
	output3 := make(chan int)

	// Fan-out: distribute input to multiple outputs
	go func() {
		defer close(output1)
		defer close(output2)
		defer close(output3)

		for data := range input {
			select {
			case output1 <- data:
			case output2 <- data:
			case output3 <- data:
			}
		}
	}()

	// Send data
	go func() {
		defer close(input)
		for i := 0; i < 10; i++ {
			input <- i
		}
	}()

	// Consume from outputs
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		for range output1 {
		}
	}()

	go func() {
		defer wg.Done()
		for range output2 {
		}
	}()

	go func() {
		defer wg.Done()
		for range output3 {
		}
	}()

	wg.Wait()
}

// FanInExample demonstrates the fan-in pattern where multiple input channels merge into a single output channel.
// Three goroutines send data to separate input channels which are merged into one output channel for consumption.
func FanInExample() {
	input1 := make(chan int)
	input2 := make(chan int)
	input3 := make(chan int)
	output := make(chan int)

	// Fan-in: merge multiple inputs into one output
	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		for data := range input1 {
			output <- data
		}
	}()

	go func() {
		defer wg.Done()
		for data := range input2 {
			output <- data
		}
	}()

	go func() {
		defer wg.Done()
		for data := range input3 {
			output <- data
		}
	}()

	// Close output when all inputs are done
	go func() {
		wg.Wait()
		close(output)
	}()

	// Send data to inputs
	go func() {
		defer close(input1)
		for i := 0; i < 5; i++ {
			input1 <- i
		}
	}()

	go func() {
		defer close(input2)
		for i := 5; i < 10; i++ {
			input2 <- i
		}
	}()

	go func() {
		defer close(input3)
		for i := 10; i < 15; i++ {
			input3 <- i
		}
	}()

	// Consume merged output
	for data := range output {
		_ = data
	}
}

// SemaphoreExample demonstrates using a buffered channel as a semaphore to limit concurrent operations.
// Restricts execution to a maximum of 3 concurrent goroutines using a buffered channel with capacity 3.
func SemaphoreExample() {
	semaphore := make(chan struct{}, 3) // Allow max 3 concurrent operations
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }() // Release semaphore

			// Do work (simulate with sleep)
			time.Sleep(100 * time.Millisecond)
		}(i)
	}

	wg.Wait()
}

// SyncPrimitivesExample demonstrates various synchronization primitives including Mutex, RWMutex, Once, WaitGroup, Cond, and atomic operations.
// Shows proper usage patterns for Go's synchronization mechanisms to coordinate concurrent access and execution.
func SyncPrimitivesExample() {
	var mu sync.Mutex
	var rwmu sync.RWMutex
	var once sync.Once
	var wg sync.WaitGroup
	var cond = sync.NewCond(&sync.Mutex{})
	var counter int64

	// Mutex usage
	mu.Lock()
	defer mu.Unlock()

	// RWMutex usage
	rwmu.RLock()
	defer rwmu.RUnlock()

	// Once usage
	once.Do(func() {
		// Initialize something
	})

	// WaitGroup usage
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Do work
	}()
	wg.Wait()

	// Condition variable usage
	go func() {
		cond.L.Lock()
		cond.Wait()
		cond.L.Unlock()
	}()

	// Atomic operations
	atomic.AddInt64(&counter, 1)
	atomic.LoadInt64(&counter)
	atomic.StoreInt64(&counter, 42)
}

// PotentialLeakExample demonstrates common goroutine leak patterns where goroutines have no exit mechanism.
// Contains two anti-patterns: a goroutine blocked on channel read without close, and an infinite loop without cancellation.
func PotentialLeakExample() {
	ch := make(chan int)

	// This goroutine will leak if ch is never closed
	go func() {
		for {
			select {
			case data := <-ch:
				_ = data
			}
		}
	}()

	// Another potential leak - infinite loop without proper exit
	go func() {
		for {
			time.Sleep(time.Second)
			// No way to stop this goroutine
		}
	}()
}

// ContextCancellationExample demonstrates proper goroutine lifecycle management using context for cancellation.
// Shows the recommended pattern of using context.WithTimeout to ensure goroutines can be properly terminated.
func ContextCancellationExample() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch := make(chan int)

	go func() {
		for {
			select {
			case data := <-ch:
				_ = data
			case <-ctx.Done():
				return // Proper cleanup
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Do periodic work
			case <-ctx.Done():
				return // Proper cleanup
			}
		}
	}()
}
