# MetaNode：Web3 学习作业集

本仓库是「MetaNode 学习 Web3」路线的作业合集，按阶段从底层原理逐步过渡到真实链交互与 Go 后端工程。
全部代码为 Go 实现，包含 4 个可独立编译运行的项目。

## 项目一览

| 项目 | 对应阶段 | 内容 | 验证方式 |
|---|---|---|---|
| [mini-blockchain](mini-blockchain/) | 阶段 A | 纯 Go 标准库从零实现迷你区块链：区块哈希链、PoW、UTXO 交易、ECDSA 钱包、链校验 | `sh scripts/dev.sh test` |
| [eth-client](eth-client/) | 阶段 B / D | 用 go-ethereum 连接 Sepolia：查余额/区块/交易/gas、构造并签名广播转账；ERC-20 / ERC-721 / Uniswap V2 只读查询 | `sh scripts/eth.sh test`、`sh scripts/eth.sh dryrun` |
| [blog-backend](blog-backend/) | Go 后端作业 | Gin + GORM + JWT 个人博客后端：注册登录、文章 CRUD、评论、作者权限、统一错误码与结构化日志，附零依赖前端 SPA | `sh scripts/build.sh`、`sh scripts/api_test.sh` |
| [go-learn-demo](go-learn-demo/) | Go 语言练习 | 指针、Goroutine、面向对象、Channel、锁机制、GORM 共 13 道题，CLI + Web 双形态，附过程动画 | `sh scripts/build_check.sh` |
| [contract/homework01](contract/homework01/) | 合约作业 01 | Solidity 六道题：Voting 投票合约、字符串反转、整数/罗马数字互转、合并有序数组、二分查找 | `cd contract/homework01 && forge test` |

## 快速开始

各项目的详细说明、API 文档和命令示例见对应目录下的 `README.md`。

```bash
# 阶段 A：迷你区块链
cd mini-blockchain && sh scripts/dev.sh test

# 阶段 B/D：连接 Sepolia（只读演示，无需私钥）
cd eth-client && sh scripts/eth.sh build && sh scripts/eth.sh dryrun

# Go 后端作业：编译并启动，然后另开终端跑全流程测试
cd blog-backend && sh scripts/build.sh && ./blog-server
sh scripts/api_test.sh

# Go 语言练习：编译检查 + 启动 Web 演示
cd go-learn-demo && sh scripts/build_check.sh && sh scripts/start_web.sh

# 合约作业 01：Foundry 测试
cd contract/homework01 && forge test
```

## 环境说明

- Go：`mini-blockchain` 与 `eth-client` 使用 Go 1.23；`blog-backend` 与 `go-learn-demo` 的
  `go.mod` 声明了更高的工具链版本，Go 1.21+ 会在需要时自动下载对应工具链。
- 依赖代理：国内网络建议 `export GOPROXY=https://goproxy.cn,direct`，
  各项目的构建脚本已内置该设置。
- 链上写操作需要自行注入 `PRIVATE_KEY` 环境变量并提供测试币，仓库内不含任何私钥。

## 安全说明

- 代码与文档中不含私钥、助记词或 GitHub Token；示例密码仅用于本地演示账号。
- 运行期产物（SQLite 数据库、日志、编译出的二进制）均已通过 `.gitignore` 排除。
- 所有漏洞与攻击类实验仅限本地开发链、测试网或明确授权的目标。
