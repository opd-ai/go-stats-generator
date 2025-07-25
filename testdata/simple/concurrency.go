package concurrency

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Worker pool pattern example
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

// Pipeline pattern example
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

// Fan-out pattern example
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

// Fan-in pattern example
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

// Semaphore pattern using buffered channel
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

// Various sync primitives examples
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

// Potential goroutine leak example
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

// Context-based cancellation (good pattern)
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
