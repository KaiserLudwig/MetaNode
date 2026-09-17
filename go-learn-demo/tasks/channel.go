package tasks

import "fmt"

// RunChannelDemo 一个协程往通道发送 1~10，主协程接收并打印。
func RunChannelDemo() {
	ch := make(chan int) // 无缓冲通道：发送和接收必须同时准备好

	// 生产者协程
	go func() {
		for i := 1; i <= 10; i++ {
			ch <- i // 发送
		}
		close(ch) // 发完关闭，接收方用 range 能感知结束
	}()

	// 主协程（消费者）
	fmt.Print("接收到的整数: ")
	for v := range ch {
		fmt.Printf("%d ", v)
	}
	fmt.Println("\n通道已关闭，通信结束")
}
