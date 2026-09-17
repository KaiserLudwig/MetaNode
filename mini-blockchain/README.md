# 阶段 A：从零实现一个迷你区块链（Go）

本阶段是「MetaNode 学习 Web3」路线图的第一步：**用纯 Go 标准库从零写一个迷你区块链**，
理解比特币/以太坊底层的核心原理。**零外部依赖**，便于专注概念本身。

## 你将学到什么

| 概念 | 对应代码文件 | 说明 |
|---|---|---|
| 区块与哈希链 | `blockchain/block.go` | 区块头（索引/时间/前一哈希/随机数/难度）+ SHA-256；链式连接防篡改 |
| 工作量证明 PoW | `blockchain/proof.go` | 挖矿：枚举 Nonce 直到哈希满足目标难度；验证容易、生成困难 |
| 交易与 UTXO | `blockchain/transaction.go` | 比特币式「未花费输出」账本；输入/输出、找零、杜绝双花 |
| ECDSA 钱包 | `blockchain/wallet.go` | 公私钥生成、地址派生、签名/验签、公钥哈希锁定 |
| 链核心与校验 | `blockchain/chain.go` | 添加区块、校验链自洽、余额查询、构建并签名转账 |

## 运行方式（在 WSL 中）

```bash
cd /mnt/f/学习/Web3/MetaNode/mini-blockchain

# 交互式运行
sh scripts/dev.sh run

# 自动演示流程（挖矿->转账->余额->打印链->退出）
sh scripts/dev.sh runauto

# 篡改检测演示
sh scripts/dev.sh runtamper

# 运行全部单元测试
sh scripts/dev.sh test

# 编译检查
sh scripts/dev.sh build
```

> 若提示 `go: command not found`，先执行：
> `export PATH=/usr/local/go/bin:$PATH`

## CLI 演示功能

```
命令:
  1)  挖一个区块（给矿工 50 币奖励，空交易）
  2)  Alice -> Bob 转账 25
  3)  查询 Alice / Bob / 矿工 余额
  4)  打印整条链
  5)  篡改演示（篡改一个区块的交易，看校验是否被拦截）
  0)  退出
```

## 关键设计决策（教学向的简化）

为了让概念清晰，本实现相对真实区块链做了这些简化，并保留了最核心的规则：

1. **哈希**：用单次 SHA-256（真实比特币是双 SHA-256）+ 直接拼接交易，而非 Merkle 树。
   核心仍成立：**任意字段被改，哈希即变**。
2. **共识难度**：用「哈希前导 N 个 0」表示难度（真实比特币用目标值/比特位编码）。
   难度每 +1，期望尝试次数约 ×16，直观展示 PoW 的算力成本。
3. **UTXO 账本**：保留「输入必须未被花费 + 资金来源充足 + 签名验签」三条铁律，
   省略了脚本、手续费、确认数等。
4. **ECDSA**：用 Go 标准库内置的 P-256 曲线（真实用 SECP256k1）；地址直接用公钥哈希，
   保留「私钥签名、公钥验签、地址 ≠ 私钥」三个关键概念。
5. **金额**：用整数（真实用最小单位 satoshi/wei），避免浮点误差。

## 三个你一定要做的实验

1. **改难度**：把 `main.go` 里的 `const difficulty = 3` 改成 `2` 和 `4`，
   观察挖矿耗时（Nonce 尝试次数）的变化——直观体会「难度与算力成正比」。
2. **改交易金额**：在 `runtamper` 演示中手动改 `Outputs[0].Value`，
   观察区块哈希改变导致校验失败——直观体会「内容不可篡改」。
3. **打印链**：运行 `4` 观察每个区块的 `prev` 与前一块 `hash` 的衔接——
   直观体会「链」的连接方式。

## 项目结构

```
mini-blockchain/
├── go.mod
├── main.go                      # CLI 交互入口
├── scripts/
│   └── dev.sh                   # 运行/测试/演示脚本
└── blockchain/
    ├── block.go                 # 区块结构与哈希链接
    ├── proof.go                 # PoW 挖矿与校验
    ├── transaction.go           # 交易与 UTXO 模型
    ├── wallet.go                # ECDSA 钱包
    ├── chain.go                 # 区块链核心（校验/余额/转账）
    └── blockchain_test.go       # 单元测试（6 个用例）
```

## 下一阶段

阶段 B：用 Go 连接真实的以太坊测试网（Sepolia），实现查余额 / 查交易 / 转账的小工具。
届时会引入 `go-ethereum` 依赖，也会真正用到「地址、私钥、签名、广播交易」这些真实概念。
