# BeggingContract 部署与测试记录（Sepolia）

> 全部数据来自 Sepolia 真实链上交易，可用交易哈希在区块浏览器自行核对。
> 部署时间：2026-09-18（北京时间）

## 合约地址

| 项目 | 值 |
|---|---|
| 网络 | Ethereum Sepolia（chainId 11155111） |
| 合约地址 | `0x5c0eCb22f0D55965aC49D29A0f00862D45B3096d` |
| 部署账户 | `0x22d2ad2336d958c5bF224cc346ce8f5Fb6741F5C` |
| 编译设置 | solc 0.8.24，optimizer runs 200，evmVersion cancun |
| 源码验证 | Sourcify `exact_match`（creation + runtime 均匹配） |
| 浏览器 | https://sepolia.etherscan.io/address/0x5c0eCb22f0D55965aC49D29A0f00862D45B3096d |
| Sourcify | https://repo.sourcify.dev/contracts/full_match/11155111/0x5c0eCb22f0D55965aC49D29A0f00862D45B3096d/ |

## 链上操作与结果

| 步骤 | 函数 / 操作 | 参数 | 交易哈希 | 区块 | gas |
|---|---|---|---|---|---|
| 1 | 部署 | 无构造参数 | `0x308c5bb848b31213f0c3b3af0c010bc9449491829cbe8a8aa9a8cf50c16158eb` | 11727291 | 570842 |
| 2 | `donate()` | 0.001 ETH（部署账户） | `0x29d94dce7c9eba391849adf06002a01c464aee93ca09a9641c666f44b1b00a1f` | 11727292 | 139160 |
| 3 | 给第二个捐赠者转账 | 0.003 ETH（gas 与捐赠额） | `0x3fbf039ce84378100f199ac8a3f8c343f51f96fb313331e53f14d31a88fa91ae` | 11727295 | 21000 |
| 4 | 直接转账到合约（走 `receive()`） | 0.0005 ETH（第二账户） | `0xe30052cf09ec20030a505dc1e2bbaa5e4ad4289b59e2f50889702f30ec99ce5b` | 11727296 | 104722 |
| 5 | `withdraw()` | 仅 owner | `0x7274d239f66dc34065fcab168a3cb813642af2c547f900a1c004e3e6c00a2848` | 11727298 | 29715 |

## 读取结果（`cast call`）

```
getDonation(0x22d2...1F5C) = 1000000000000000   （0.001 ETH）
getDonation(0xe0bc...f355) =  500000000000000   （0.0005 ETH，来自 receive()）
totalDonated               = 1500000000000000   （0.0015 ETH）
donorCount                 = 2
topDonors                  = [0x22d2...1F5C, 0xe0bc...f355, 0x0] / [1e15, 5e14, 0]
提款前合约余额              = 1500000000000000
提款后合约余额              = 0
提款后 getDonation 记录保留 = 1000000000000000
```

结论：`donate` / `receive` 记账、多地址累加、`topDonors` 排序、`getDonation` 查询、
`withdraw` 提空并保留记录，均在 Sepolia 真链上验证通过。

## 源码验证

```
$ forge verify-contract 0x5c0eCb22f0D55965aC49D29A0f00862D45B3096d \
    src/BeggingContract.sol:BeggingContract --chain sepolia --verifier sourcify
creationMatch: exact_match
runtimeMatch:  exact_match
```

> 说明：Etherscan 的 API 在本机网络下不可达（连接超时），因此改用同样公开、可自行核对的
> Sourcify 完成源码验证。若需要在 Etherscan 上显示已验证源码，在能访问 etherscan.io 的网络下执行：
> `forge verify-contract <地址> src/BeggingContract.sol:BeggingContract --chain sepolia --etherscan-api-key <KEY>`

## 测试截图清单（作业要求「测试截图」）

在自己浏览器打开下列页面截图即可，都对应上表中的真实交易：

1. 合约地址页（含余额、合约创建交易）
   https://sepolia.etherscan.io/address/0x5c0eCb22f0D55965aC49D29A0f00862D45B3096d
2. 部署交易
   https://sepolia.etherscan.io/tx/0x308c5bb848b31213f0c3b3af0c010bc9449491829cbe8a8aa9a8cf50c16158eb
3. `donate` 交易（Logs 标签能看到 `Donation` 事件）
   https://sepolia.etherscan.io/tx/0x29d94dce7c9eba391849adf06002a01c464aee93ca09a9641c666f44b1b00a1f
4. 第二笔捐赠（直接转账，`Donation` 事件同样触发）
   https://sepolia.etherscan.io/tx/0xe30052cf09ec20030a505dc1e2bbaa5e4ad4289b59e2f50889702f30ec99ce5b
5. `withdraw` 交易（Internal Txns 里能看到资金转给 owner）
   https://sepolia.etherscan.io/tx/0x7274d239f66dc34065fcab168a3cb813642af2c547f900a1c004e3e6c00a2848
6. 读写合约页（Read Contract / Write Contract 里调用 `getDonation`、`totalDonated`、`topDonors`）
   https://sepolia.etherscan.io/address/0x5c0eCb22f0D55965aC49D29A0f00862D45B3096d#readContract
