package tasks

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// RunAtomicCounterDemo 用原子操作实现无锁计数器。
func RunAtomicCounterDemo() {
	var counter int64
	var wg sync.WaitGroup

	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				atomic.AddInt64(&counter, 1) // CPU 级原子指令，无需加锁
			}
		}()
	}

	wg.Wait()
	fmt.Printf("原子计数器最终值: %d（期望 10000）\n", counter)
}
