package tasks

import (
	"fmt"
	"sync"
	"time"
)

// Job 描述一个待调度的任务。
type Job struct {
	Name string
	Fn   func()
}

// RunSchedulerDemo 并发执行一组任务并统计每个任务耗时，
// 再与串行执行对比，直观感受并发调度的收益。
func RunSchedulerDemo() {
	jobs := []Job{
		{"任务A: 模拟下载", func() { time.Sleep(300 * time.Millisecond) }},
		{"任务B: 模拟解析", func() { time.Sleep(150 * time.Millisecond) }},
		{"任务C: 模拟渲染", func() { time.Sleep(450 * time.Millisecond) }},
		{"任务D: 模拟上传", func() { time.Sleep(200 * time.Millisecond) }},
	}

	// 串行执行作为对照
	seqStart := time.Now()
	for _, j := range jobs {
		j.Fn()
	}
	seqCost := time.Since(seqStart)
	fmt.Printf("串行总耗时: %s\n", seqCost.Round(time.Millisecond))

	// 并发执行（每个任务一个协程）
	var wg sync.WaitGroup
	wg.Add(len(jobs))
	concStart := time.Now()
	for _, j := range jobs {
		go func(j Job) { // 必须把 j 作为参数传入，避免循环变量捕获问题
			defer wg.Done()
			t0 := time.Now()
			j.Fn()
			fmt.Printf("  [完成] %s 耗时 %s\n", j.Name, time.Since(t0).Round(time.Millisecond))
		}(j)
	}
	wg.Wait()
	concCost := time.Since(concStart)

	fmt.Printf("并发总耗时: %s（约等于最慢的单个任务）\n", concCost.Round(time.Millisecond))
}
