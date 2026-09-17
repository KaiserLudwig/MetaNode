# Go 学习任务演示程序

基于 MetaNode 学习 Web3 / 区块链 Go 后端的练习合集，把 5 大主题共 10 道题做成一个
交互式演示程序：输入编号运行对应演示，并展示**题目、考察点、运行结果、源码**。

## 运行方式（在 WSL 中）

```bash
cd /mnt/f/学习/Web3/MetaNode/go-learn-demo

# 直接运行
go run .

# 或编译成二进制再运行
go build -o demo .
./demo
```

> 如果提示 `go: command not found`，先执行：
> `export PATH=/usr/local/go/bin:$PATH`

## Web 版（推荐体验方式）

在 WSL 中启动服务，然后 **Windows 浏览器直接访问 http://localhost:8080**
（WSL2 自带 localhost 端口转发，无需额外配置）：

```bash
sh scripts/start_web.sh          # 构建并启动（端口 8080，可用 PORT=8090 改）
sh scripts/check_web.sh          # 自检：静态资源与任务页面
sh scripts/build_check.sh        # 构建 + 全部 JS 语法检查
```

停止服务：`pkill -x webdemo`。

网页上点击任务卡片进入任务页，每个任务都有两个区块：

1. **🎬 过程动画演示** —— 点击「运行动画」，以动画 + 分步教学字幕演示执行过程
   （如指针箭头、协程并发、通道小球流动、抢锁动画等），支持重播与 1x/2x/4x 调速；
   播放完成后自动高亮下方真实运行结果进行对照。
2. **真实运行结果 + 完整源码** —— 由 Go 服务端实际执行输出。

动画为纯原生 JS/CSS/SVG 实现（零外部依赖、离线可用），框架见
`web/static/demo-anim.js`，10 个任务场景在 `web/static/demos/` 下。

## 操作

| 输入 | 功能 |
|---|---|
| `1` ~ `10` | 运行对应编号的演示（含源码展示） |
| `a` | 依次运行全部演示 |
| `s` | 查看全部题目与考察点 |
| `0` / `q` | 退出 |

## 题目清单

### 指针
1. 指针加10 —— 函数接收整数指针，将指向的值 +10（值传递 vs 引用传递）
2. 切片指针×2 —— 函数接收整数切片指针，每个元素 ×2（指针运算、切片操作）

### Goroutine
3. 两个协程打印奇偶数 —— `go` 关键字、协程并发执行
4. 并发任务调度器 —— 并发执行一组任务并统计每个任务耗时（协程原理、并发调度）

### 面向对象
5. Shape 接口 —— `Area()` / `Perimeter()`，Rectangle 与 Circle 实现（接口、多态）
6. 组合 —— Person + Employee 组合，实现 `PrintInfo()`（组合、方法接收者）

### Channel
7. 通道通信 —— 协程发 1~10，另一协程接收打印（通道基本使用、协程间通信）
8. 缓冲通道 —— 生产者发 100 个整数，消费者接收（通道缓冲机制）

### 锁机制
9. `sync.Mutex` 计数器 —— 10 协程 × 1000 次递增（互斥锁、并发数据安全）
10. `sync/atomic` 无锁计数器 —— 原子操作实现（原子操作、并发数据安全）

### 进阶 GORM（内存 SQLite 真实执行，每次运行全新数据）
11. 模型定义 —— User / Post / Comment 一对多关系，AutoMigrate 建表（打印真实 DDL）
12. 关联查询 —— Preload 预加载用户文章+评论；聚合查询评论最多的文章
13. 钩子函数 —— AfterCreate 自动维护 PostCount；AfterDelete 评论删光自动更新状态

> 首次构建会通过 GOPROXY 下载 GORM 依赖（gorm.io/gorm + glebarez/sqlite 纯 Go 驱动）。
> 若下载失败，先执行：`export GOPROXY=https://goproxy.cn,direct`

## 目录结构

```
go-learn-demo/
├── go.mod
├── main.go                 # 交互式 CLI 菜单入口
├── README.md
├── scripts/
│   ├── start_web.sh        # 构建并启动 Web 服务
│   ├── check_web.sh        # 服务自检
│   └── build_check.sh      # 构建 + JS 语法检查
├── web/
│   ├── main.go             # Web 服务：任务路由、stdout 捕获、静态资源
│   ├── index.html          # 页面模板（含动画舞台）
│   └── static/
│       ├── demo-anim.css   # 动画舞台样式 + 通用零件
│       ├── demo-anim.js    # 动画框架（舞台/按钮/字幕/补间工具）
│       └── demos/          # 13 个任务动画场景
│           ├── pointer.js  # 任务 1：指针箭头 + 数值滚动
│           ├── pointer_slice.js  # 任务 2：元素翻倍
│           ├── goroutine.js      # 任务 3：双协程打字机
│           ├── scheduler.js      # 任务 4：并行进度条 + 耗时对比
│           ├── shape.js          # 任务 5：SVG 图形 + 公式
│           ├── employee.js       # 任务 6：组合字段提升
│           ├── channel.js        # 任务 7：数字小球流过管道
│           ├── buffered.js       # 任务 8：缓冲槽可视化
│           ├── mutex.js          # 任务 9：抢锁动画
│           ├── atomic.js         # 任务 10：原子递增
│           ├── gorm_models.js    # 任务 11：表卡片 + 关系线 + 建表
│           ├── gorm_query.js     # 任务 12：预加载展开评论 + 聚合统计
│           └── gorm_hooks.js     # 任务 13：统计自动 +1 / 状态自动切换
└── tasks/
    ├── tasks.go            # 任务注册表 + 源码嵌入（go:embed）
    ├── pointer.go          # 任务 1
    ├── pointer_slice.go    # 任务 2
    ├── goroutine.go        # 任务 3
    ├── scheduler.go        # 任务 4
    ├── shape.go            # 任务 5
    ├── employee.go         # 任务 6
    ├── channel.go          # 任务 7
    ├── buffered_channel.go # 任务 8
    ├── mutex_counter.go    # 任务 9
    ├── atomic_counter.go   # 任务 10
    ├── gorm_common.go      # GORM 共用：内存 SQLite、建表、种子数据、DDL 读取
    ├── gorm_models.go      # 任务 11：User/Post/Comment 模型 + 建表
    ├── gorm_query.go       # 任务 12：Preload 关联查询 + 聚合查询
    └── gorm_hooks.go       # 任务 13：AfterCreate/AfterDelete 钩子
```
