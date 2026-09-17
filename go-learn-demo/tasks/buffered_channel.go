package tasks

import (
	"fmt"
	"sync"
)

// RunBufferedChannelDemo 缓冲通道：生产者发 100 个整数，消费者接收打印。
func RunBufferedChannelDemo() {
	const total = 100
	ch := make(chan int, 10) // 容量 10 的缓冲通道：未满时发送不会阻塞

	var wg sync.WaitGroup
	wg.Add(2)

	// 生产者
	go func() {
		defer wg.Done()
		for i := 1; i <= total; i++ {
			ch <- i
		}
		close(ch)
		fmt.Println("[生产者] 发送完毕（100 个整数）")
	}()

	// 消费者
	go func() {
		defer wg.Done()
		count := 0
		for range ch {
			count++
		}
		fmt.Printf("[消费者] 共接收 %d 个整数\n", count)
	}()

	wg.Wait()
	fmt.Println("缓冲通道演示结束")
}
