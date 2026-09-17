package tasks

import (
	"fmt"
	"strings"
	"sync"
)

// RunOddEvenDemo 用两个协程分别打印 1~10 的奇数和偶数。
func RunOddEvenDemo() {
	var wg sync.WaitGroup
	wg.Add(2)

	// 协程 1：打印奇数
	go func() {
		defer wg.Done()
		var sb strings.Builder
		sb.WriteString("奇数协程: ")
		for i := 1; i <= 10; i += 2 {
			fmt.Fprintf(&sb, "%d ", i)
		}
		fmt.Println(sb.String())
	}()

	// 协程 2：打印偶数
	go func() {
		defer wg.Done()
		var sb strings.Builder
		sb.WriteString("偶数协程: ")
		for i := 2; i <= 10; i += 2 {
			fmt.Fprintf(&sb, "%d ", i)
		}
		fmt.Println(sb.String())
	}()

	wg.Wait() // 等待两个协程都结束
	fmt.Println("两个协程并发执行完毕（谁先打印每次可能不同）")
}
