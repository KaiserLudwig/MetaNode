// 交互式 Go 学习演示程序。
// 输入编号运行对应演示，并展示题目、考察点、运行结果与源码。
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"metanode-demo/tasks"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("==================================================")
	fmt.Println("  Go 学习任务演示（指针 / Goroutine / 面向对象 / Channel / 锁机制）")
	fmt.Println("==================================================")

	for {
		showMenu()
		fmt.Print("> 请输入编号 (0 退出 / a 全部演示 / s 查看题目列表): ")
		line, err := reader.ReadString('\n')
		if err != nil && len(line) == 0 { // 输入流结束（如管道）
			fmt.Println("输入结束，再见！")
			return
		}
		line = strings.TrimSpace(strings.ToLower(line))
		switch line {
		case "0", "q", "quit", "exit":
			fmt.Println("再见！")
			return
		case "a", "all":
			runAll(reader)
			continue
		case "s":
			showQuestions()
			continue
		case "":
			continue
		}
		n, err := strconv.Atoi(line)
		if err != nil || n < 1 || n > len(tasks.All) {
			fmt.Println("无效输入，请输入菜单中的编号。")
			continue
		}
		runTask(reader, tasks.All[n-1])
	}
}

// showMenu 打印菜单。
func showMenu() {
	var currentTopic string
	fmt.Println("--------------------------------------------------")
	for _, t := range tasks.All {
		if t.Topic != currentTopic {
			currentTopic = t.Topic
			fmt.Printf("[%s]\n", currentTopic)
		}
		fmt.Printf("  %2d. %s\n", t.ID, t.Title)
	}
	fmt.Println("--------------------------------------------------")
}

// showQuestions 打印全部题目与考察点。
func showQuestions() {
	fmt.Println("================ 全部题目与考察点 ================")
	for _, t := range tasks.All {
		fmt.Printf("【%d】[%s] %s\n", t.ID, t.Topic, t.Title)
		fmt.Printf("    考察点: %s\n", t.Points)
	}
	fmt.Println("==================================================")
}

// runTask 运行单个演示：题目、考察点、运行结果、源码。
func runTask(reader *bufio.Reader, t tasks.Task) {
	fmt.Println("================ 题目 ================")
	fmt.Printf("【%d】[%s] %s\n", t.ID, t.Topic, t.Title)
	fmt.Printf("考察点: %s\n", t.Points)
	fmt.Println("---------------- 运行结果 ----------------")
	t.Run()
	fmt.Println("---------------- 源码展示 ----------------")
	fmt.Print(t.Source())
	fmt.Println("==================================================")
	pause(reader)
}

// runAll 依次运行全部演示。
func runAll(reader *bufio.Reader) {
	for _, t := range tasks.All {
		runTask(reader, t)
	}
}

// pause 交互终端下按回车继续；管道输入时自动跳过。
func pause(reader *bufio.Reader) {
	if !isTerminal() {
		return
	}
	fmt.Print("按回车键继续...")
	reader.ReadString('\n')
}

func isTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
