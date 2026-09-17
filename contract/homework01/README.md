# 合约作业 01

对应 MetaNodeAcademy `LearningRoadmap` 的 `contract/homework01.md`，六道题全部用 Solidity 实现，
并配有 Foundry 测试，不依赖任何第三方库，clone 后可直接 `forge test`。

## 题目完成对照

| # | 题目 | 合约文件 | 入口函数 | 测试用例数 |
|---|---|---|---|---|
| 1 | Voting 投票合约 | [`src/Voting.sol`](src/Voting.sol) | `vote` / `getVotes` / `resetVotes` | 7 |
| 2 | 反转字符串 | [`src/ReverseString.sol`](src/ReverseString.sol) | `reverse` | 6 |
| 3 | 整数转罗马数字 | [`src/RomanNumeral.sol`](src/RomanNumeral.sol) | `intToRoman` | 3 |
| 4 | 罗马数字转整数 | [`src/RomanNumeral.sol`](src/RomanNumeral.sol) | `romanToInt` | 3 |
| 5 | 合并两个有序数组 | [`src/MergeSortedArray.sol`](src/MergeSortedArray.sol) | `mergeSorted` | 6 |
| 6 | 二分查找 | [`src/BinarySearch.sol`](src/BinarySearch.sol) | `search` / `searchIndex` | 8 |

两道罗马数字题另有 1 个覆盖 1 ~ 3999 全量互转的公共测试，测试结果：**34 passed，0 failed**。

## 运行方式

```bash
cd contract/homework01

forge build          # 编译
forge test           # 跑全部测试
forge test -vv       # 查看详细调用栈
forge test --match-contract RomanNumeralTest   # 只跑某道题
```

环境要求：Foundry（`forge`）任意近期版本；`foundry.toml` 已固定 `solc 0.8.24`，首次编译时 forge 会自动下载该版本编译器。

## 各题实现说明

### 1. Voting

- `mapping(string => uint256)` 记录每个候选人的票数，候选人首次出现时自动登记到 `string[]` 列表，`resetVotes` 依靠该列表遍历清零。
- `vote` 每次调用给对应候选人加一票，并触发 `Voted(voter, candidate, totalVotes)` 事件。
- `getVotes` 对未出现过的候选人返回 0；`resetVotes` 清零票数但保留候选人列表，触发 `VotesReset` 事件。
- 空字符串候选人会 revert。按要求未限制投票次数，同一地址可投多票。

### 2. Reverse String

- Solidity 的 `string` 底层是 UTF-8 字节序列，这里按字节反转：`"abcde" -> "edcba"`。
- 对 ASCII 完全正确；对中文、emoji 这类多字节字符会拆散字节，属于有意保留的简化，已在代码注释中说明。

### 3. 整数转罗马数字（`intToRoman`）

- 贪心法：从 1000 到 1 的 13 组「数值-符号」表中反复扣除，覆盖 `IV`、`IX`、`XL`、`CM` 等减法写法。
- 支持范围 1 ~ 3999，越界 revert；`1994 -> "MCMXCIV"`、`3888 -> "MMMDCCCLXXXVIII"`（最长 15 字符）。

### 4. 罗马数字转整数（`romanToInt`）

- 从右向左扫描：当前符号小于右侧符号则减，否则加，正确处理 `IV`、`IX`、`CM` 等。
- 末尾用 `intToRoman` 做回环规范校验，因此 `"IIII"`、`"IIV"` 这类非标准写法会被拒绝，非法字符和空串也会 revert。
- 测试包含 1 ~ 3999 全量回环校验。

### 5. 合并两个有序数组（`mergeSorted`）

- 双指针归并，时间复杂度 O(m + n)，返回新数组，保留重复元素。
- 覆盖空数组、单侧为空、区间交错等情况。

### 6. 二分查找（`search` / `searchIndex`）

- 左闭右开区间写法，`mid = low + (high - low) / 2` 避免溢出，时间复杂度 O(log n)。
- `search` 返回 `(found, index)`；`searchIndex` 未找到时返回 -1，便于调用方直接判断。
- 数组含重复元素时返回任意一个匹配下标，测试会校验 `arr[index] == target`。
