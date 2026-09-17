# 阶段 B：用 Go 连接真实的以太坊测试网（Sepolia）

本阶段是「MetaNode 学习 Web3」路线图的第二步：**用 Go 连接真实的以太坊 Sepolia 测试网**，
实现查余额 / 查交易 / 查区块 / 查 gas 价格 / 构造并广播一笔真实转账。
相比阶段 A 的「从零写链」，这一步真正接入**真实的区块链网络节点**，体会与真实世界交互。

## 你将学到什么

| 功能 | 对应代码 | 说明 |
|---|---|---|
| 连接 RPC 节点 | `internal/ethcli/client.go` `New()` | 通过 RPC URL 与节点对话，读取链上数据 |
| 查余额 | `BalanceAt()` | wei 单位，可指定区块高度（`nil` = 最新） |
| 查区块 | `BlockByNumber()` | 高度、时间戳、哈希、交易数 |
| 查交易 | `TransactionByHash()` | 发送方/接收方/金额/gas，用 `types.Sender` 还原签名者 |
| 查 gas 价 | `GasPrice()` | `SuggestGasPrice`，单位 gwei |
| 转账 | `SendWei()` | 私钥派生地址 → 查 nonce → 估算 gas → 签名 → 广播 |

## 运行方式（在 WSL 中）

```bash
cd /mnt/f/学习/Web3/MetaNode/eth-client

# 编译检查
sh scripts/eth.sh build

# 只读演示（连接 Sepolia，查链ID/区块/gas/余额，无需私钥）
sh scripts/eth.sh dryrun

# 交互运行 / 测试
sh scripts/eth.sh run
sh scripts/eth.sh test
```

> 首次构建需下载 go-ethereum 依赖（已在 `go.mod` 中固定为 v1.15.0，兼容 Go 1.23）。
> 若下载超时，`scripts/go_cmd.sh` 已内置国内代理 `goproxy.cn`。

## 子命令

| 命令 | 说明 |
|---|---|
| `balance <地址>` | 查账户余额（wei / ETH 双单位） |
| `block   [高度]` | 查区块信息（默认最新），并给出首笔交易哈希 |
| `gas` | 查当前 gas 单价（gwei） |
| `tx <交易哈希>` | 查交易详情（发送方/接收方/金额/gas） |
| `send <收款地址> <ETH金额>` | 构造并广播一笔真实 Sepolia 转账 |
| `--demo` | 只读综合演示 |

例如：

```bash
# 查一个地址的余额
go run . balance 0x000000000000000000000000000000000000dEaD

# 查最新区块与首笔交易
go run . block

# 查指定交易详情
go run . tx 0x0d4681f5c6af9a568c2e1ea4892ef5ca672da205310014c1832089b536dd40ba
```

## 转账（send）

转账需要两样东西：**收款地址** 与 **你的私钥**。私钥通过环境变量注入，绝不写死在代码里：

```bash
# 导出你的 Sepolia 私钥（0x 开头或纯 64 位 hex）
export PRIVATE_KEY=你的私钥

# 转 0.01 ETH 给某个地址
go run . send 0xAb72d1510B4D756C34E27e82bE4E81CCec0dBF65 0.01
```

- 广播后返回交易哈希，可在区块浏览器查看：
  `https://sepolia.etherscan.io/tx/<交易哈希>`
- **注意**：广播成功 ≠ 已确认，需等矿工打包（通常数秒到一分钟）。且需要账户有足够测试币支付 gas。

### 安全约定
`send` 若未检测到 `PRIVATE_KEY` 环境变量，会**直接报错退出**，不会广播交易，避免误操作。

## 关键概念（真实以太坊转账的 5 步）

1. **私钥派生地址**：`crypto.PubkeyToAddress(priv.PublicKey)` 得到发送方。
2. **查询 nonce**：`PendingNonceAt` 拿到该账户已发送的交易数（nonce 从 0 递增），
   用于防止**交易重放**（同一 nonce 只能被确认一次）。
3. **估算 gas**：`EstimateGas` 估算本次调用消耗的 gas 上限（简单转账固定 21000 起步）。
4. **签名交易**：用私钥对交易签名（EIP-155 签名器，绑定 chainID 防止跨链重放）。
5. **广播**：`SendTransaction` 把已签名交易发给节点，等待矿工打包进区块。

## 单位约定

| 单位 | 换算 |
|---|---|
| wei | 最小单位（整数，1 ETH = 10^18 wei） |
| gwei | 常用于 gas 单价（1 gwei = 10^9 wei） |

Go 用 `*big.Int` 存 wei（大整数），避免浮点误差——这点与阶段 A 用整数存金额一致。

## 项目结构

```
eth-client/
├── go.mod / go.sum             # 依赖（go-ethereum v1.15.0）
├── main.go                     # CLI 命令入口（balance/block/gas/tx/send/token/nft/pool/demo）
├── scripts/
│   ├── eth.sh                  # 运行/测试/演示脚本
│   ├── go_cmd.sh               # WSL 内执行 go 命令（含国内代理）
│   └── find_sig.sh             # 辅助：确认 go-ethereum API（教学用）
└── internal/
    └── ethcli/
        ├── client.go           # 以太坊交互核心（查/转封装）
        ├── contract.go         # 合约 ABI 交互（ERC-20 / ERC-721 / Uniswap V2，阶段 D）
        └── ethcli_test.go      # 单元测试（无需真实网络）
```

## 已知差异与下一步

- 本实现用 **Legacy 交易（type 0）** 演示，最简单直观；真实场景常改用 **EIP-1559**（type 2，含 `maxFeePerGas`/`maxPriorityFeePerGas`）。
- 只读查询全部真实可用；`send` 的签名/广播链已实现并通过单元测试与安全测试，真正广播需你提供测试币私钥。

---

# 阶段 D：NFT 与 DeFi 实战

本阶段在阶段 B 的基础上，进一步用 go-ethereum 的 **ABI 编码** 读取真实链上的
**ERC-20 代币（DeFi）** 与 **ERC-721 NFT**，并演示用相同能力读取 **Uniswap V2 池**的储备与定价。

## 你将学到什么

| 能力 | 对应代码 | 说明 |
|---|---|---|
| ABI 编码调用 | `contract.go` `call()` | 函数选择器 + 参数 → calldata → `eth_call` → 解码返回值 |
| ERC-20 查询 | `ERC20Name/Symbol/Decimals/BalanceOf/TotalSupply` | 读 name/symbol/decimals/余额/总供应量 |
| ERC-721 NFT | `NFTName/Symbol/OwnerOf/BalanceOf/TokenURI` | 读 NFT 名称/所有者/元数据 URI |
| Uniswap V2 池 | `PoolToken0/Token1/Reserves/TotalSupply` | 读池子的两个代币与流动性储备 |

核心概念：**ABI（应用二进制接口）** 用 JSON 描述合约的函数签名与返回类型。
调用方用 ABI 把「函数名 + 参数」编码成 calldata 发给节点，节点执行 `eth_call`（**只读、不耗 gas**），
返回编码结果，调用方再用 ABI 解码成 Go 类型。

## 运行演示（WSL 中）

```bash
cd /mnt/f/学习/Web3/MetaNode/eth-client

# 阶段 D 完整只读演示（ERC-20 + NFT + Uniswap 池）
sh scripts/eth.sh defidemo
```

也可单独运行每个命令：

```bash
# ERC-20 代币（Sepolia USDC，第 3 个可选参数是 owner 地址）
go run . token 0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238 0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf

# ERC-721 NFT（Sepolia 上的一个 NFT 收藏，tokenId=11954）
go run . nft 0x937d178f593561c9db59564aa9b835d4d67593e9 11954

# Uniswap V2 池（主网 WETH/USDC，只读示例；用主网公共 RPC）
ETH_RPC_URL=https://ethereum-rpc.publicnode.com go run . pool 0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc
```

## 已验证的真实输出（Sepolia / 主网）

**ERC-20（Sepolia USDC，精度 6）：**
```
名称: USDC / 符号: USDC / 精度: 6 decimals
0x7E5F...5Bdf 持有: 187910000（最小单位，≈187.91 USDC）
```

**ERC-721 NFT（Backchain Reward Booster，tokenId=11954）：**
```
NFT 名称: Backchain Reward Booster / 符号: BKCB
tokenId 11954 所有者: 0x94Fb446508ff2001909A5ae053e019f4C5Db056b
元数据 URI: ipfs://...crystal_booster.json
```

**Uniswap V2 池（主网 WETH/USDC）：**
```
token0: USDC / token1: WETH
reserve0 / reserve1 = 10062644232905 : 4087794102892100876009
```

### 解读：储备比例即价格
Uniswap V2 是 **恒定乘积做市商**：`reserve0 × reserve1 = k`（常数）。
因此代币的相对价格 = 两者储备之比。上面 `reserve1/reserve0` 即 1 个 WETH 对应的 USDC 数量——
这正是 DEX 无需订单簿也能定价的原理。

## 关键概念

1. **`eth_call`**：只读调用合约函数，不改变链上状态、不耗 gas、不需要私钥。适合读数据。
2. **函数选择器（selector）**：`keccak256("balanceOf(address)")` 的前 4 字节，标识调用哪个函数。
3. **ABI 编码规则**：地址是 32 字节右对齐，`uint256` 是 32 字节大端整数等。go-ethereum 的 `abi` 包封装了这些细节。
4. **单位**：ERC-20 用 `decimals` 表示精度（USDC=6，绝大多数代币=18）。读到的 `balanceOf` 是最小单位，需除以 `10^decimals` 才是人类可读数量。
5. **写操作 vs 读操作**：读用 `eth_call`（免费）；写（mint/transfer/swap）需构造交易、签名、广播并付 gas。

## 为什么「写操作」留白

阶段 D 目前聚焦**只读查询**（无需私钥、无资金风险、立即可验证）。
若要做铸币/转账/swap 等**写操作**，需要：`PRIVATE_KEY` 私钥 + 测试币支付 gas，逻辑复用阶段 B 的 `SendWei`
（构造交易 → 签名 → 广播）。这部分代码结构已就绪，你拿到测试币后即可补全实测。

## 下一步方向

- **写操作扩展**：ERC-20 `approve` + `transferFrom`、Uniswap `swapExactTokensForTokens`、NFT `mint`/`transferFrom`
- **本地链**：用 Hardhat/Anvil 自部署 NFT 与 Uniswap 合约，完全可控地测试写操作
- **进阶 DeFi**：读取 Compound/Aave 借贷利率、Uniswap V3 的 `slot0`、代币价格预言机
