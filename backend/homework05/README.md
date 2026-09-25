# 后端作业 05：区块链读写 + abigen 合约绑定

对应 MetaNodeAcademy `LearningRoadmap` 的 `backend/homework05.md`。全部代码用 Go + go-ethereum 编写，
已在 **Sepolia 测试网**完成真实运行（区块查询、ETH 转账、合约部署与调用）。

## 任务对照

| 作业要求 | 实现位置 | 验证结果 |
|---|---|---|
| 任务 1-1 环境搭建（Go、go-ethereum、Infura） | `go.mod`（go-ethereum v1.15.0）、`internal/ethutil` | Go 1.23.4 + go-ethereum v1.15.0；支持 `INFURA_API_KEY` 或自定义 `SEPOLIA_RPC_URL`；Infura 查询证据见 `evidence/04-infura-blockquery.log` |
| 任务 1-2 查询区块（哈希/时间戳/交易数等） | [`cmd/blockquery/main.go`](cmd/blockquery/main.go) | 查询区块 11727291，输出哈希、父哈希、时间戳、交易数 164、Gas、BaseFee 等；另用 Infura 复查，见 `evidence/04-infura-blockquery.log` |
| 任务 1-3 构造、签名、广播转账交易 | [`cmd/sendtx/main.go`](cmd/sendtx/main.go) | 转账 0.0001 ETH，交易 `0x2b3821ff…c79a`，区块 11744112，gas 21000，状态成功 |
| 任务 2-1 编写并编译计数器合约 | [`contracts/Counter.sol`](contracts/Counter.sol)、`contracts/Counter.abi`、`contracts/Counter.bin` | solc 0.8.24 编译通过，生成 ABI + 字节码 |
| 任务 2-2 安装 abigen 并生成 Go 绑定 | [`scripts/compile.sh`](scripts/compile.sh)、[`bindings/counter/counter.go`](bindings/counter/counter.go) | abigen v1.15.0 生成绑定代码 |
| 任务 2-3 用绑定代码调用链上合约 | [`cmd/counter/main.go`](cmd/counter/main.go) | 部署 `0x77176Ed6…ca68`，`inc()` 后计数 100→101，`incBy(5)` 后 101→106 |

## 项目结构

```
backend/homework05/
├── cmd/
│   ├── blockquery/main.go      # 任务 1-2：查询指定区块
│   ├── sendtx/main.go          # 任务 1-3：构造/签名/广播 ETH 转账
│   └── counter/main.go         # 任务 2-3：用 abigen 绑定部署并调用 Counter
├── contracts/
│   ├── Counter.sol             # 计数器合约源码
│   ├── Counter.abi             # solc 生成的 ABI
│   └── Counter.bin             # solc 生成的字节码
├── bindings/counter/counter.go # abigen 生成的 Go 绑定（自动生成，勿手改）
├── internal/ethutil/ethutil.go # RPC 连接与私钥加载的公共封装
├── scripts/
│   ├── compile.sh              # solc + abigen 一键生成 ABI/字节码/绑定
│   └── demo.sh                 # 依次跑完三个任务并输出日志
├── evidence/                   # 本次 Sepolia 真实运行日志
└── go.mod / go.sum
```

## 环境准备

```bash
# 1) Go 与依赖
go version                                             # go1.23.4
GOPROXY=https://goproxy.cn,direct go mod tidy           # 拉取 go-ethereum v1.15.0

# 2) 合约编译工具
solc --version                                          # 0.8.24

# 3) abigen（go-ethereum 自带工具）
GOPROXY=https://goproxy.cn,direct go install github.com/ethereum/go-ethereum/cmd/abigen@v1.15.0
abigen --version                                        # abigen version 1.15.0-stable

# 4) RPC：注册 Infura 后把 API Key 写进环境变量即可（也可用其它 Sepolia 节点）
export INFURA_API_KEY=your_infura_project_id
```

## 运行方式

```bash
# 私钥通过环境变量注入，代码中不保存任何私钥
export PRIVATE_KEY=0x...                 # Sepolia 测试账户私钥
export SEPOLIA_RPC_URL=https://sepolia.infura.io/v3/$INFURA_API_KEY   # 可选，优先级最高

# 任务 1-2：查询区块（不带参数则查最新区块）
go run ./cmd/blockquery 11727291

# 任务 1-3：发送 0.0001 ETH
go run ./cmd/sendtx 0xe0bc8f3970b0fb8F0B53aaeA58c7d905494cf355 0.0001

# 任务 2-1 / 2-2：重新生成 ABI、字节码与 Go 绑定
sh scripts/compile.sh

# 任务 2-3：部署新合约并调用（初值 100）
go run ./cmd/counter 100
# 也可以连接已部署的合约：
COUNTER_ADDRESS=0x77176Ed6b75165d53bdca916F2FAbf51AcFEca68 go run ./cmd/counter

# 一次性跑完三个任务
sh scripts/demo.sh
```

## 本次 Sepolia 实测结果

运行日志见 [`evidence/`](evidence/)。

另使用 `INFURA_API_KEY` 通过 Infura Sepolia endpoint 只读复查区块 11727291；返回值与上表一致。新增日志只记录提供商标签，不记录 endpoint 中的 API Key：[`evidence/04-infura-blockquery.log`](evidence/04-infura-blockquery.log)。

| 任务 | 结果 | 交易哈希 | 区块 | gas |
|---|---|---|---|---|
| 查询区块 11727291 | 交易数 164，时间戳 1789692384，BaseFee 1.1 gwei | — | — | — |
| ETH 转账 0.0001 ETH | 状态成功 | `0x2b3821ff541de4f673bf28faf49d9d00f84eeb4c4e7cf9b68e2ff813eb80c79a` | 11744112 | 21000 |
| 部署 `Counter(100)` | 合约 `0x77176Ed6b75165d53bdca916F2FAbf51AcFEca68` | `0x8c5bb54cd51db4fa263764204e52aba35f409d9f1f3cfc61e5665105bbeb43a3` | 11744114 | 221460 |
| `inc()` | 计数 100 → 101 | `0xecfefaecbc0f98c1bb7462e182742655ce2d76e49cd361e8da2b1855caaefb6c` | 11744115 | 27824 |
| `incBy(5)` | 计数 101 → 106 | `0x005121f8a8d3da34e7c2016b21e15a22ed76fca3b4b5428b1e4a46b2a803c729` | 11744116 | 28085 |

浏览器核对（合约地址与交易哈希均可查）：

- 合约：https://sepolia.etherscan.io/address/0x77176Ed6b75165d53bdca916F2FAbf51AcFEca68
- 转账：https://sepolia.etherscan.io/tx/0x2b3821ff541de4f673bf28faf49d9d00f84eeb4c4e7cf9b68e2ff813eb80c79a
- `incBy(5)`：https://sepolia.etherscan.io/tx/0x005121f8a8d3da34e7c2016b21e15a22ed76fca3b4b5428b1e4a46b2a803c729

## 实现要点

- **连接与私钥**：统一走 `internal/ethutil`；RPC 优先级为 `SEPOLIA_RPC_URL` → `INFURA_API_KEY`（自动拼 Infura 地址）→ 公共节点；私钥只从 `PRIVATE_KEY` 环境变量读取，仓库内不含任何私钥。
- **交易签名**：使用 EIP-1559 动态手续费（`GasTipCap` + `2*BaseFee` 作为 `GasFeeCap`），`types.LatestSignerForChainID` 绑定 chainId 防重放，广播后轮询交易回执确认上链。
- **abigen 绑定**：`DeployCounter` 由绑定代码自动完成「构造函数编码 → 部署交易 → 返回实例」，调用写方法时自动处理 nonce 与签名，读方法直接 `eth_call`。
- **可重复性**：`scripts/compile.sh` 一条命令重建 ABI、字节码与绑定代码；`scripts/demo.sh` 复现全部三个任务的运行输出。
