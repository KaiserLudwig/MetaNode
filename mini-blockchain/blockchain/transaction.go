package blockchain

// 交易与 UTXO（Unspent Transaction Output，未花费交易输出）账本模型。
//
// 比特币采用 UTXO 模型：每笔交易把「上一个输出」当作输入，并产生「新输出」。
// 每个输出都绑定到一个收款地址（公钥哈希）。钱包的余额 = 所有「未被花费」
// 的、属于该地址的输出金额之和。转账时把若干输入消耗掉，生成给收款人的输出，
// 差额作为找零回到自己账户。
//
// 本实现为教学做了大量简化（真实比特币有脚本、签名、手续费、找零、确认数等），
// 但保留最关键的「输入必须未被花费 + 资金来源充足 + 签名校验」三条铁律。

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// OutPoint 指向某笔交易的一个输出，用于标识「钱从哪里来」。
// 这里用交易 ID + 输出序号定位，比链内序号更直观、也更贴近真实的
// (交易哈希, 输出索引) 引用方式。
type OutPoint struct {
	TxID   string // 被引用交易的哈希
	OutIdx int    // 该交易的哪个输出
}

// TXInput 交易输入：声明我要花掉哪一笔输出，并附带签名证明我是所有者。
type TXInput struct {
	Prev       OutPoint   // 被花费的上一笔输出
	Signature  []byte     // 对签名内容的签名
	PubKeyHash []byte     // 公钥哈希（用于定位所有者）
	PubKey     []byte     // 本次交易的发送方公钥（完整）
}

// TXOutput 交易输出：把一笔金额锁定给某个地址。
type TXOutput struct {
	Value      int64  // 金额（单位：最小单位，这里用整数避免浮点误差）
	PubKeyHash []byte // 收款人公钥哈希（谁才能花这笔钱）
}

// Transaction 一笔交易。
type Transaction struct {
	ID       string     // 交易哈希（可当作交易 ID）
	Inputs   []TXInput  // 输入列表
	Outputs  []TXOutput // 输出列表
	LockTime int64      // 附加元信息（简化，可忽略）
}

// NewCoinbaseTx 构造创世/挖矿奖励交易：没有真实输入，直接把奖励给 miner 地址。
func NewCoinbaseTx(minerAddress string, reward int64) *Transaction {
	rewardHash := hashPubKey([]byte(minerAddress))
	tx := &Transaction{
		ID: "",
		Inputs: []TXInput{
			{
				Prev:       OutPoint{TxID: "COINBASE", OutIdx: -1},
				Signature:  []byte("COINBASE"),
				PubKeyHash: rewardHash,
			},
		},
		Outputs: []TXOutput{
			{Value: reward, PubKeyHash: rewardHash},
		},
		LockTime: time.Now().Unix(),
	}
	tx.ID = tx.CalculateID()
	return tx
}

// CalculateID 计算交易哈希（去掉签名后的内容，避免签名本身参与哈希造成循环）。
func (tx *Transaction) CalculateID() string {
	var sb strings.Builder
	for _, in := range tx.Inputs {
		sb.WriteString(in.Prev.TxID)
		sb.WriteByte(':')
		sb.WriteString(strconv.Itoa(in.Prev.OutIdx))
		sb.WriteByte('|')
	}
	for _, out := range tx.Outputs {
		sb.WriteString(strconv.FormatInt(out.Value, 10))
		sb.WriteByte('|')
		sb.WriteString(hex.EncodeToString(out.PubKeyHash))
		sb.WriteByte('|')
	}
	sb.WriteString(strconv.FormatInt(tx.LockTime, 10))
	sum := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(sum[:])
}

// Sign 用发送方私钥对「交易去签名后的内容」签名，写入每个输入的 Signature。
//
// 真实的钱包会对「去除脚本签名的整笔交易」做签名，且输入输出都用到了
// SIGHASH 机制。这里为教学负责：对 CalculateID 的字节做 ECDSA 签名。
func (tx *Transaction) Sign(privateKey *ecdsa.PrivateKey, pubKeyHash []byte) error {
	// 简单起见：对交易 ID 做签名。
	hash := sha256.Sum256([]byte(tx.ID))
	sig := signECDSA(privateKey, hash[:])
	for i := range tx.Inputs {
		tx.Inputs[i].Signature = sig
		tx.Inputs[i].PubKeyHash = pubKeyHash
	}
	return nil
}

// Hash 返回交易 ID。
func (tx *Transaction) Hash() string { return tx.ID }

// String 便于打印。
func (tx *Transaction) String() string {
	return fmt.Sprintf("Tx %s | ins=%d | outs=%d", tx.ID[:8], len(tx.Inputs), len(tx.Outputs))
}
