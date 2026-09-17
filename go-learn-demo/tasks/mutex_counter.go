package tasks

import (
	"fmt"
	"sync"
)

// RunMutexCounterDemo 用 sync.Mutex 保护共享计数器。
func RunMutexCounterDemo() {
	var (
		mu      sync.Mutex
		counter int
		wg      sync.WaitGroup
	)

	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				mu.Lock()
				counter++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()
	fmt.Printf("Mutex 计数器最终值: %d（期望 10000，加锁保证并发安全）\n", counter)
}
