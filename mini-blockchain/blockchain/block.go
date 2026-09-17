// Package blockchain 是一个教学用的迷你区块链实现。
//
// 本阶段（A）用纯 Go 标准库实现：区块、SHA-256 哈希链、PoW 共识、
// 交易与 UTXO 账本、ECDSA 钱包签名。不依赖任何外部库，方便你理解
// 比特币/以太坊底层的核心原理。
package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Block 区块链中的一个区块。
//
// 一个区块把「一段时间内发生的交易」打包起来，并持有指向前一个区块的
// 哈希（PrevHash），从而把所有区块串成一条链——这就是「区块链」名字的由来。
type Block struct {
	Index        int     // 区块高度（创世块为 0，之后递增）
	Timestamp    int64   // 出块时间戳（Unix 秒）
	PrevHash     string  // 前一个区块的哈希，形成链式连接
	Hash         string  // 本区块自身的哈希
	Nonce        int     // 工作量证明随机数（挖矿时不断调整）
	Difficulty   int     // 目标难度：要求哈希以 Difficulty 个前导十六进制 0 开头
	Transactions []*Transaction // 打包在本区块里的交易列表
}

// NewBlock 构造一个新区块（尚未挖矿，Hash/Nonce 留空，等待 Mine 填充）。
func NewBlock(index int, prevHash string, txs []*Transaction, difficulty int) *Block {
	return &Block{
		Index:        index,
		Timestamp:    time.Now().Unix(),
		PrevHash:     prevHash,
		Difficulty:   difficulty,
		Transactions: txs,
	}
}

// CalculateHash 计算区块头的 SHA-256 哈希。
//
// 比特币实际是「双 SHA-256」且包含 Merkle 树根，这里为教学简化为单次 SHA-256，
// 并直接对交易列表的序列化拼接进行哈希。重点是理解：只要有任意一个字段（交易）
// 被改动，哈希就会变化——这正是区块链「防篡改」的本质。
func (b *Block) CalculateHash() string {
	var sb strings.Builder
	sb.WriteString(strconv.Itoa(b.Index))
	sb.WriteByte('|')
	sb.WriteString(strconv.FormatInt(b.Timestamp, 10))
	sb.WriteByte('|')
	sb.WriteString(b.PrevHash)
	sb.WriteByte('|')
	sb.WriteString(strconv.Itoa(b.Nonce))
	sb.WriteByte('|')
	for _, tx := range b.Transactions {
		// 关键：这里重新计算交易的内容哈希（而非缓存字段），
		// 这样一旦交易内容被篡改，区块哈希也会随之改变，从而被校验发现。
		sb.WriteString(tx.CalculateID())
		sb.WriteByte('|')
	}
	sum := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(sum[:])
}

// HashMatchesDifficulty 判断一个哈希是否满足目标难度（前导 Difficulty 个 "0"）。
//
// 例如 Difficulty=4 时，哈希必须以 "0000" 开头。难度越高，找到合法哈希的
// 期望算力开销越大，这正是 PoW 让「出块」变得昂贵的机制。
func HashMatchesDifficulty(hash string, difficulty int) bool {
	if difficulty <= 0 {
		return true
	}
	if len(hash) < difficulty {
		return false
	}
	return strings.Repeat("0", difficulty) == hash[:difficulty]
}

// String 便于打印区块信息。
func (b *Block) String() string {
	return fmt.Sprintf(
		"Block #%d | time=%d | prev=%.10s... | nonce=%d | diff=%d | hash=%.10s... | txs=%d",
		b.Index, b.Timestamp, b.PrevHash, b.Nonce, b.Difficulty, b.Hash, len(b.Transactions),
	)
}
