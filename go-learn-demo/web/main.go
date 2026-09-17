// Web 版演示服务：把 10 个 Go 学习任务做成可点击的网页，
// 点击卡片即可看到运行结果与源码。监听 0.0.0.0:PORT，
// 通过 WSL2 的 localhost 转发，Windows 浏览器直接访问 http://localhost:PORT。
package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"metanode-demo/tasks"
)

//go:embed index.html
var pageHTML string

//go:embed static
var staticFS embed.FS

var tmpl = template.Must(template.New("page").Parse(pageHTML))

// Group 按主题分组后的任务列表。
type Group struct {
	Topic string
	Tasks []tasks.Task
}

// PageData 模板数据：Task 为空时渲染菜单页。
type PageData struct {
	Groups     []Group
	Task       *tasks.Task
	Output     string
	Source     string
	BootScript template.JS // 任务页初始化脚本（html/template 的 JS 上下文转义需要）
}

// runMu 串行化任务执行，避免并发请求时 os.Stdout 重定向互相干扰。
var runMu sync.Mutex

// runCapture 运行演示函数，捕获其标准输出；函数 panic 时返回错误信息。
func runCapture(fn func()) string {
	runMu.Lock()
	defer runMu.Unlock()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return "无法创建输出管道: " + err.Error()
	}
	os.Stdout = w

	var out string
	func() {
		defer func() {
			w.Close()
			os.Stdout = old
			if p := recover(); p != nil {
				out = fmt.Sprintf("运行出错 (panic): %v", p)
			}
		}()
		fn()
	}()
	data, _ := io.ReadAll(r)
	if out == "" {
		out = string(data)
	}
	return out
}

// buildGroups 按主题顺序分组。
func buildGroups() []Group {
	var groups []Group
	var cur *Group
	for _, t := range tasks.All {
		if cur == nil || cur.Topic != t.Topic {
			groups = append(groups, Group{Topic: t.Topic})
			cur = &groups[len(groups)-1]
		}
		cur.Tasks = append(cur.Tasks, t)
	}
	return groups
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	data := PageData{Groups: buildGroups()}
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("渲染失败: %v", err)
	}
}

func handleTask(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	n, err := strconv.Atoi(idStr)
	if err != nil || n < 1 || n > len(tasks.All) {
		http.Error(w, "任务不存在，请返回菜单重新选择。", http.StatusNotFound)
		return
	}
	t := tasks.All[n-1]
	data := PageData{
		Groups:     buildGroups(),
		Task:       &t,
		Output:     runCapture(t.Run),
		Source:     t.Source(),
		BootScript: template.JS(fmt.Sprintf("MND.boot(%d);", t.ID)),
	}
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("渲染失败: %v", err)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	port := flag.String("port", envOr("PORT", "8080"), "监听端口")
	flag.Parse()
	addr := "0.0.0.0:" + *port

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/task", handleTask)
	mux.Handle("/static/", http.FileServer(http.FS(staticFS)))

	log.Printf("Web 演示服务已启动: http://%s", addr)
	log.Printf("Windows 浏览器访问: http://localhost:%s", *port)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
