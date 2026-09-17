package blockchain

// Proof of Work（工作量证明）共识。
//
// 核心思想：想让一个区块被接纳，必须「付出算力成本」找到一个 Nonce，
// 使得区块哈希满足目标难度（前导若干个 0）。找出这个 Nonce 的过程就叫
// 挖矿。验证方只需重新计算一次哈希即可低成本确认，完美体现了
// 「验证容易、生成困难」的非对称性。
//
// 本实现采用最简单的难度模型：哈希前导 0 的个数。真实的比特币还用
// 目标值/比特位编码，并含挖矿 reward 交易、Merkle 树，这里为教学简化。

import (
	"fmt"
	"strings"
)

// Mine 通过暴力枚举 Nonce 直到 Block 哈希满足 Difficulty 要求。
//
// 返回挖到的 Nonce，并就地更新 b.Hash / b.Nonce。difficulty 越高，尝试
// 次数通常越多（每增高一位难度，期望尝试次数约 ×16）。
func (b *Block) Mine(difficulty int) int {
	if difficulty <= 0 {
		// 无难度时也把哈希算出来，方便演示。
		b.Nonce = 0
		b.Hash = b.CalculateHash()
		return 0
	}
	var attempts int64
	for {
		b.Hash = b.CalculateHash()
		if HashMatchesDifficulty(b.Hash, difficulty) {
			return b.Nonce
		}
		b.Nonce++
		attempts++
		if attempts%100_000 == 0 {
			// 周期性提示，避免长时间无明显反馈。可用 strings 为了将来扩展。
			_ = fmt.Sprintf("  ...已尝试 %d 次 (nonce=%d)\n", attempts, b.Nonce)
		}
	}
}

// Validate 校验区块是否合法：哈希正确、Nonce 满足难度、链接正确。
func (b *Block) Validate(prevHash string) error {
	if b.PrevHash != prevHash {
		return fmt.Errorf("区块 %d 的前块哈希不匹配", b.Index)
	}
	if b.Hash != b.CalculateHash() {
		return fmt.Errorf("区块 %d 的哈希不自洽（内容被篡改？）", b.Index)
	}
	if !HashMatchesDifficulty(b.Hash, b.Difficulty) {
		return fmt.Errorf("区块 %d 未满足难度要求", b.Index)
	}
	return nil
}

// nextLeadingZeros 统计哈希前导 0 个数（辅助诊断用）。
func nextLeadingZeros(hash string) int {
	return len(hash) - len(strings.TrimLeft(hash, "0"))
}
