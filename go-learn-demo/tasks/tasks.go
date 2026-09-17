// Package tasks 汇集 Go 学习演示任务：指针、Goroutine、面向对象、Channel、锁机制。
package tasks

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed *.go
var sources embed.FS

// Task 描述一个演示任务。
type Task struct {
	ID     int
	Topic  string // 所属主题
	Title  string // 题目
	Points string // 考察点
	Run    func() // 演示函数
	File   string // 对应源码文件名（用于展示）
}

// All 按菜单顺序排列的全部任务。
var All = []Task{
	{
		ID: 1, Topic: "指针", Title: "指针加10（引用传递）",
		Points: "指针的使用、值传递与引用传递的区别",
		Run:    RunPointerDemo, File: "pointer.go",
	},
	{
		ID: 2, Topic: "指针", Title: "切片指针，每个元素×2",
		Points: "指针运算、切片操作",
		Run:    RunPointerSliceDemo, File: "pointer_slice.go",
	},
	{
		ID: 3, Topic: "Goroutine", Title: "两个协程打印奇偶数",
		Points: "go 关键字的使用、协程的并发执行",
		Run:    RunOddEvenDemo, File: "goroutine.go",
	},
	{
		ID: 4, Topic: "Goroutine", Title: "并发任务调度器（统计耗时）",
		Points: "协程原理、并发任务调度",
		Run:    RunSchedulerDemo, File: "scheduler.go",
	},
	{
		ID: 5, Topic: "面向对象", Title: "Shape 接口：Rectangle / Circle",
		Points: "接口的定义与实现、面向对象编程风格",
		Run:    RunShapeDemo, File: "shape.go",
	},
	{
		ID: 6, Topic: "面向对象", Title: "组合：Person + Employee",
		Points: "组合的使用、方法接收者",
		Run:    RunEmployeeDemo, File: "employee.go",
	},
	{
		ID: 7, Topic: "Channel", Title: "通道通信：发送 1~10 并接收打印",
		Points: "通道的基本使用、协程间通信",
		Run:    RunChannelDemo, File: "channel.go",
	},
	{
		ID: 8, Topic: "Channel", Title: "缓冲通道：生产者发 100 个整数",
		Points: "通道的缓冲机制",
		Run:    RunBufferedChannelDemo, File: "buffered_channel.go",
	},
	{
		ID: 9, Topic: "锁机制", Title: "sync.Mutex 保护计数器",
		Points: "sync.Mutex 的使用、并发数据安全",
		Run:    RunMutexCounterDemo, File: "mutex_counter.go",
	},
	{
		ID: 10, Topic: "锁机制", Title: "sync/atomic 无锁计数器",
		Points: "原子操作、并发数据安全",
		Run:    RunAtomicCounterDemo, File: "atomic_counter.go",
	},
	{
		ID: 11, Topic: "进阶 GORM", Title: "模型定义：User/Post/Comment 建表",
		Points: "GORM 模型定义、一对多关系、AutoMigrate 建表",
		Run:    RunGormModelDemo, File: "gorm_models.go",
	},
	{
		ID: 12, Topic: "进阶 GORM", Title: "关联查询：用户文章+评论 / 评论最多文章",
		Points: "Preload 预加载、关联查询、聚合统计（COUNT/GROUP BY）",
		Run:    RunGormQueryDemo, File: "gorm_query.go",
	},
	{
		ID: 13, Topic: "进阶 GORM", Title: "钩子函数：自动更新统计与状态",
		Points: "AfterCreate/AfterDelete 钩子、事务内更新",
		Run:    RunGormHooksDemo, File: "gorm_hooks.go",
	},
}

// Source 返回任务对应源码（带行号），便于代码展示。
func (t Task) Source() string {
	data, err := sources.ReadFile(t.File)
	if err != nil {
		return fmt.Sprintf("(无法读取源码: %v)\n", err)
	}
	var b strings.Builder
	for i, line := range strings.Split(string(data), "\n") {
		fmt.Fprintf(&b, "%3d | %s\n", i+1, line)
	}
	return b.String()
}
