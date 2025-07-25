package simple

import "sync"

func SimpleGoroutineTest() {
	go func() {
		println("simple test")
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		println("with waitgroup")
	}()

	wg.Wait()
}
