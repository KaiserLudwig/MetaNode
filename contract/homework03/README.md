# 合约大作业：NFT 拍卖市场（Hardhat + Chainlink + UUPS）

对应 MetaNodeAcademy `LearningRoadmap` 的 `contract/homework03.md`。项目用 Hardhat 开发，
实现了 ERC721 拍卖市场、Chainlink 价格预言机换算美元、UUPS 可升级代理，以及覆盖全部功能的测试。

## 作业要求对照

| 作业要求 | 实现位置 | 说明 |
|---|---|---|
| 使用 Hardhat 开发 | `hardhat.config.js`、`package.json` | Hardhat 2.29 + ethers v6 + `@openzeppelin/hardhat-upgrades` |
| ERC721 NFT，支持铸造与转移 | [`contracts/AuctionNFT.sol`](contracts/AuctionNFT.sol) | ERC721 + ERC721URIStorage，`mint(to, uri)`，转移由标准实现提供 |
| 创建拍卖 | [`contracts/NFTAuction.sol`](contracts/NFTAuction.sol) 的 `createAuction` | NFT 转入合约托管，记录卖家/起拍价/结束时间 |
| 出价（ERC20 或 ETH） | `bidEth` / `bidErc20` | 两种支付方式各自独立，被超越的出价立即原路退回 |
| 结束拍卖 | `endAuction` | NFT 归最高出价者，资金扣手续费后归卖家；无人出价则退回 NFT |
| Chainlink 预言机取价 | [`contracts/PriceOracle.sol`](contracts/PriceOracle.sol) | `latestRoundData` 读取喂价，做新鲜度/有效性校验，统一成 8 位小数 |
| 出价换算美元 | `quoteUsd` / `highestBidUsd` | 前端可直接比较 ETH 与 ERC20 出价的美元价值 |
| UUPS / 透明代理升级 | `NFTAuction`（UUPS）+ [`contracts/upgrades/NFTAuctionV2.sol`](contracts/upgrades/NFTAuctionV2.sol) | 代理 + 实现分离，`upgradeProxy` 升级并保留状态 |
| 单元测试与集成测试 | `test/` | 60 个用例，语句覆盖率 97.25%（核心合约） |
| 部署脚本（测试网） | [`scripts/deploy.js`](scripts/deploy.js)、[`scripts/upgrade.js`](scripts/upgrade.js) | 一条命令部署/升级，自动写入 `deployments/<network>.json` |
| 额外挑战：动态手续费 | `feeBpsForUsd` | 按成交价美元金额分三档：0.5% / 1% / 2% |

## 项目结构

```
homework03/
├── contracts/
│   ├── AuctionNFT.sol              # ERC721（可铸造、带 tokenURI）
│   ├── PriceOracle.sol             # Chainlink 喂价封装，输出统一 8 位小数的美元价格
│   ├── NFTAuction.sol              # 拍卖市场（UUPS 可升级实现，V1）
│   ├── upgrades/NFTAuctionV2.sol   # 升级版本：新增最小加价比例规则
│   └── mocks/                      # 测试用 MockERC20 与 MockV3Aggregator
├── scripts/
│   ├── deploy.js                   # 部署 NFT / 预言机 / 拍卖代理，自动配置喂价
│   └── upgrade.js                  # 升级代理到 V2 并初始化新状态
├── test/
│   ├── AuctionNFT.test.js          # NFT 铸造 / 转移 / 权限
│   ├── PriceOracle.test.js         # 喂价读取 / 精度缩放 / 过期与异常
│   └── NFTAuction.test.js          # 拍卖全流程 + 动态手续费 + UUPS 升级
├── hardhat.config.js
├── package.json
└── deployments/                    # 部署结果（Sepolia 记录会提交）
```

## 快速开始

```bash
cd contract/homework03
npm install          # 已配置 npmmirror 源（.npmrc）

npx hardhat compile
npx hardhat test              # 60 个用例
npx hardhat coverage          # 生成 coverage/ 与 coverage.json

npx hardhat run scripts/deploy.js                    # 本地内存链部署演练
npx hardhat node                                     # 启动本地链（另开终端）
npx hardhat run scripts/deploy.js --network localhost
npx hardhat run scripts/upgrade.js --network localhost
```

## 核心设计

### 拍卖状态机

```
createAuction ──► Created ──(时间到 + endAuction)──► Ended
                     │
                     └──(无出价 + cancelAuction)────► Cancelled
```

- `createAuction` 会把 NFT 通过 `transferFrom` 托管到拍卖合约，卖家必须提前 `setApprovalForAll`。
- `bidEth` 直接接收 ETH；`bidErc20` 通过 `safeTransferFrom` 托管 ERC20。两者都走同一套内部记账逻辑。
- 新出价必须严格高于当前最高价（V1），被超越的出价**立即原路退回**，避免用户手动提取。
- 结算时先转移 NFT 再转账资金，并使用重入锁保护。

### 价格与手续费

- `PriceOracle` 通过 `latestRoundData()` 读取 Chainlink 喂价，要求：`answer > 0`、`updatedAt` 不超过 1 小时、`answeredInRound >= roundId`。
- 价格统一缩放为 **8 位小数**，再按代币精度把数量换算成美元：`usd = amount * price / 10^decimals`。
- 动态手续费按成交价美元金额分档：

| 成交价（USD） | 费率 | 基点 |
|---|---|---|
| < 100 | 0.5% | 50 |
| 100 ~ 1000 | 1% | 100 |
| >= 1000 | 2% | 200 |

- **兜底策略**：预言机取价失败或价格过期时，手续费回落最低档（0.5%）而不是让结算失败，保证拍卖一定能结束。`quoteUsd` 会返回 `available=false` 供前端提示。

### 可升级性

- `NFTAuction` 继承 `Initializable` / `OwnableUpgradeable` / `UUPSUpgradeable`，构造函数中调用 `_disableInitializers()` 防止实现合约被初始化。
- `_authorizeUpgrade` 限定 owner，`initialize(owner, oracle, feeRecipient)` 只在代理中执行一次。
- `NFTAuctionV2` 继承 V1 并新增 `minBidIncrementBps`，通过 `initializeV2` + `reinitializer(2)` 初始化，并重写 `_validateBid` 强制最小加价比例（默认 5%）。
- 重入锁使用 OZ `ReentrancyGuardTransient`（EIP-1153 瞬态存储，无构造函数、`@custom:stateless`），因此 `@openzeppelin/hardhat-upgrades` 的存储布局与升级安全检查全部通过。

### 安全考量

- `SafeERC20` 处理非标准 ERC20 返回值；ETH 转账检查 `call` 返回值。
- 所有涉及外部调用的入口都加 `nonReentrant`；资金先记账后转账。
- `cancelAuction` 仅卖家或平台 owner 可用，且要求无人出价。
- 平台配置（预言机、收款地址、最小加价比例）均 `onlyOwner`。
- NFT 托管、ERC20 托管都集中在拍卖合约，结算前合约持有资产，避免卖家跑路。

## 测试报告

执行 `npx hardhat test`：

```
60 passing (1s)
```

执行 `npx hardhat coverage`：

| 文件 | % Stmts | % Branch | % Funcs | % Lines |
|---|---|---|---|---|
| contracts/AuctionNFT.sol | 100 | 100 | 100 | 100 |
| contracts/NFTAuction.sol | 97.59 | 84.09 | 100 | 97.89 |
| contracts/PriceOracle.sol | 95 | 88.89 | 100 | 95.24 |
| contracts/upgrades/NFTAuctionV2.sol | 100 | 83.33 | 100 | 100 |
| **contracts/ 合计** | **97.25** | **85.19** | **100** | **97.56** |
| 全部（含 mocks） | 95.16 | 85 | 91.49 | 95.24 |

覆盖的关键场景：

- NFT：铸造、tokenURI、转移、授权、接口支持、权限校验
- 预言机：喂价设置权限、8 位小数归一、代币精度换算、价格过期、非正价格、未配置喂价
- 拍卖：ETH/ERC20 两条出价链路、退款、余额与授权异常、起拍价与时长边界、未知拍卖、取消规则
- 结算：无人出价退回、ETH 与 ERC20 分账、重复结算拦截、手续费结算到账
- 动态手续费：三档边界、预言机涨价后档位变化、价格过期与缺失时的兜底、收款地址为零免手续费
- 升级：升级后版本与状态保留、新规则生效、`initializeV2` 只能一次、非 owner 升级被拒

## 部署到 Sepolia

```bash
export SEPOLIA_RPC_URL=https://ethereum-sepolia-rpc.publicnode.com
export PRIVATE_KEY=0x...            # 有测试币的账户私钥
export ETHERSCAN_API_KEY=...        # 可选，用于源码验证

npx hardhat run scripts/deploy.js --network sepolia
npx hardhat run scripts/upgrade.js --network sepolia   # 演示升级

# 源码验证
npx hardhat verify --network sepolia <代理地址>        # 代理
npx hardhat verify --network sepolia <实现地址>        # 实现合约
```

部署脚本会自动使用下列**已核实**的 Sepolia Chainlink 喂价（脚本内已内置，调用时读取实时价格）：

| 交易对 | 地址 |
|---|---|
| ETH/USD | `0x694AA1769357215DE4FAC081bf1f309aDC325306` |
| LINK/USD | `0xc59E3633BAAC79493d908e63626716e204A45EdF` |
| USDC/USD | `0xA2F78ab2355fe2f984D808B5CeE7FD0A93D5270E` |
| DAI/USD | `0x14866185B1962B63C3Ea9E03Bc1da838bab34C19` |

本地演练结果（`npx hardhat node` + `--network localhost`，用于验证脚本正确性）：

```
PriceOracle:                  0x5FbDB2315678afecb367f032d93F642f64180aa3
AuctionNFT:                   0xCf7Ed3AccA5a467e9e704C703E8D87F634fB0Fc9
NFTAuction 代理:               0x5FC8d32690cc91D4c39d9d3abcBD16989F875707
NFTAuction 实现(V1):           0xDc64a140Aa3E981100a9becA4E685f962f0cF6C9
升级后实现(V2):                0x610178dA211FEF7D417cB0e6FeD39F05609AD788
版本变化: 1.0.0 → 2.0.0（拍卖数量与状态保留）
```

### Sepolia 部署地址

| 合约 | 地址 |
|---|---|
| PriceOracle | _待部署_ |
| AuctionNFT | _待部署_ |
| NFTAuction（代理） | _待部署_ |
| NFTAuctionV2（升级后实现） | _待部署_ |

## 已知限制

- 拍卖时长限制在 1 分钟 ~ 30 天，未实现自动结算（需要外部 keeper 调用 `endAuction`）。
- 退款采用 push 模式，若收款方是恶意合约可能拒绝收款导致出价失败；生产环境可改为 pull 模式（`pendingReturns`）。
- 预言机只支持 Chainlink USD 喂价，未接入 TWAP 或链下签名价格。
- 升级仅演示了 V2，未包含 timelock / 多签治理流程。
