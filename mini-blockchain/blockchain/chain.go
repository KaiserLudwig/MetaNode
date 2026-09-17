package blockchain

// 区块链核心：持有整条链、负责校验、查询余额、构建并签名转账。
//
// 一条合法的链必须满足：
//  1. 创世块哈希正确；
//  2. 后续每个区块的 PrevHash 都等于前一块的 Hash（链式连接）；
//  3. 每个区块的哈希自洽（内容未被篡改）；
//  4. 每个区块满足 PoW 难度。
//
// 交易层面还必须满足（本实现的简化校验）：
//  5. 输入引用的上一笔输出必须真实存在且未被花费（杜绝双花）；
//  6. 每笔交易的总输入 >= 总输出（资金来源充足）；
//  7. 输入的所有者与上一笔输出的收款人是同一人，且签名通过验签（拥有证明）。

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

// Blockchain 维护整条链与已花费输出集合。
type Blockchain struct {
	blocks         []*Block
	difficulty     int
	coinbaseReward int64
	spentOutputs   map[OutPoint]bool // 记录哪些输出已被花费
	txLookup       map[string]*Transaction
}

// NewBlockchain 创建一条只含创世块的链。
func NewBlockchain(difficulty int, coinbaseReward int64) *Blockchain {
	bc := &Blockchain{
		difficulty:     difficulty,
		coinbaseReward: coinbaseReward,
		spentOutputs:   make(map[OutPoint]bool),
		txLookup:       make(map[string]*Transaction),
	}
	genesis := bc.NewBlock(nil)
	genesis.Mine(difficulty)
	bc.blocks = append(bc.blocks, genesis)
	return bc
}

// NewBlock 构造一个准备打包交易的新区块（尚未挖矿）。
func (bc *Blockchain) NewBlock(txs []*Transaction) *Block {
	var prevHash string
	if len(bc.blocks) > 0 {
		prevHash = bc.blocks[len(bc.blocks)-1].Hash
	}
	return NewBlock(len(bc.blocks), prevHash, txs, bc.difficulty)
}

// MineBlock 对打包了交易的新区块执行挖矿，校验通过后追加到链尾。
// 返回挖到的区块；若校验失败返回错误且不追加。
func (bc *Blockchain) MineBlock(txs []*Transaction) (*Block, error) {
	if err := bc.validateTransactions(txs); err != nil {
		return nil, fmt.Errorf("交易校验失败: %w", err)
	}
	block := bc.NewBlock(txs)
	block.Mine(bc.difficulty)
	if err := bc.appendBlock(block); err != nil {
		return nil, err
	}
	return block, nil
}

// appendBlock 校验并追加一个挖好的区块，同时登记交易与已花费输出。
func (bc *Blockchain) appendBlock(block *Block) error {
	var prevHash string
	if len(bc.blocks) > 0 {
		prevHash = bc.blocks[len(bc.blocks)-1].Hash
	}
	if err := block.Validate(prevHash); err != nil {
		return err
	}
	for _, tx := range block.Transactions {
		tx.ID = tx.CalculateID()
		bc.txLookup[tx.ID] = tx
	}
	for _, tx := range block.Transactions {
		for _, in := range tx.Inputs {
			if in.Prev.TxID != "COINBASE" {
				bc.spentOutputs[in.Prev] = true
			}
		}
	}
	bc.blocks = append(bc.blocks, block)
	return nil
}

// GetBlocks 返回区块列表（只读，用于展示）。
func (bc *Blockchain) GetBlocks() []*Block { return bc.blocks }

// GetDifficulty / GetCoinbaseReward 只读访问器。
func (bc *Blockchain) GetDifficulty() int       { return bc.difficulty }
func (bc *Blockchain) GetCoinbaseReward() int64 { return bc.coinbaseReward }

// BalanceOf 计算某地址的全部未花费输出金额之和。
func (bc *Blockchain) BalanceOf(address string) int64 {
	addrHash := hashPubKey([]byte(address))
	var balance int64
	for _, block := range bc.blocks {
		for _, tx := range block.Transactions {
			for outIdx, out := range tx.Outputs {
				if !bytes.Equal(out.PubKeyHash, addrHash) {
					continue
				}
				op := OutPoint{TxID: tx.ID, OutIdx: outIdx}
				if bc.spentOutputs[op] {
					continue
				}
				balance += out.Value
			}
		}
	}
	return balance
}

// SignAndBuildTransfer 构建并签名一笔普通转账：
// 从 from 的钱包里找出足够未花费输出，转给 to（金额 amount），
// 多余的作为找零返回给 from。
func (bc *Blockchain) SignAndBuildTransfer(from *Wallet, to string, amount int64) (*Transaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("金额必须为正")
	}
	fromHash := hashPubKey([]byte(from.Address()))
	toHash := hashPubKey([]byte(to))

	// 1) 收集 from 的未花费输出，凑足金额。
	var inputs []TXInput
	var collected int64
	for _, op := range bc.unspentOutpointsFor(fromHash) {
		val := bc.outputValue(op)
		if val < 0 {
			continue
		}
		inputs = append(inputs, TXInput{
			Prev:   op,
			PubKey: from.PublicKeyBytes(),
		})
		collected += val
		if collected >= amount {
			break
		}
	}
	if collected < amount {
		return nil, fmt.Errorf("余额不足: 需要 %d, 可用 %d", amount, collected)
	}

	// 2) 组装输出：给收款人，多余找零给 from。
	outputs := []TXOutput{{Value: amount, PubKeyHash: toHash}}
	if change := collected - amount; change > 0 {
		outputs = append(outputs, TXOutput{Value: change, PubKeyHash: fromHash})
	}

	tx := &Transaction{
		Inputs:  inputs,
		Outputs: outputs,
	}
	tx.ID = tx.CalculateID()

	// 3) 对每个输入签名。
	// 注意：签名发生在所有输入都填好 PubKey 之后，签名内容取交易 ID。
	sigHash := sha256.Sum256([]byte(tx.ID))
	sig := signECDSA(from.PrivateKey, sigHash[:])
	for i := range tx.Inputs {
		tx.Inputs[i].Signature = sig
		tx.Inputs[i].PubKeyHash = fromHash
	}
	return tx, nil
}

// unspentOutpointsFor 返回属于某公钥哈希的全部未花费输出。
func (bc *Blockchain) unspentOutpointsFor(pubKeyHash []byte) []OutPoint {
	var ops []OutPoint
	for _, block := range bc.blocks {
		for _, tx := range block.Transactions {
			for outIdx, out := range tx.Outputs {
				if !bytes.Equal(out.PubKeyHash, pubKeyHash) {
					continue
				}
				op := OutPoint{TxID: tx.ID, OutIdx: outIdx}
				if bc.spentOutputs[op] {
					continue
				}
				ops = append(ops, op)
			}
		}
	}
	return ops
}

// outputValue 返回某输出对应的金额；不存在返回 -1。
func (bc *Blockchain) outputValue(op OutPoint) int64 {
	tx, ok := bc.txLookup[op.TxID]
	if !ok || op.OutIdx < 0 || op.OutIdx >= len(tx.Outputs) {
		return -1
	}
	return tx.Outputs[op.OutIdx].Value
}

// outputByPoint 返回某输出；不存在返回 nil。
func (bc *Blockchain) outputByPoint(op OutPoint) *TXOutput {
	tx, ok := bc.txLookup[op.TxID]
	if !ok || op.OutIdx < 0 || op.OutIdx >= len(tx.Outputs) {
		return nil
	}
	return &tx.Outputs[op.OutIdx]
}

// validateTransactions 在打包前校验交易合法性。
func (bc *Blockchain) validateTransactions(txs []*Transaction) error {
	for _, tx := range txs {
		// coinbase 交易跳过常规校验（无真实输入）。
		if len(tx.Inputs) == 1 && tx.Inputs[0].Prev.TxID == "COINBASE" {
			continue
		}
		var inSum, outSum int64
		for _, in := range tx.Inputs {
			if _, ok := bc.txLookup[in.Prev.TxID]; !ok {
				return fmt.Errorf("引用了不存在的输出 %s:%d", in.Prev.TxID, in.Prev.OutIdx)
			}
			if bc.spentOutputs[in.Prev] {
				return fmt.Errorf("双花: 输出 %s:%d 已被花费", in.Prev.TxID, in.Prev.OutIdx)
			}
			v := bc.outputValue(in.Prev)
			if v < 0 {
				return fmt.Errorf("引用了不存在的输出 %s:%d", in.Prev.TxID, in.Prev.OutIdx)
			}
			inSum += v

			// 拥有证明：上一笔输出的收款人 hash 必须等于本次输入声明者。
			prevOut := bc.outputByPoint(in.Prev)
			if prevOut == nil || !bytes.Equal(prevOut.PubKeyHash, in.PubKeyHash) {
				return fmt.Errorf("签名所有者与上一笔输出的收款人不一致")
			}
			// 验签：用输入提供的公钥验证对交易 ID 的签名。
			pub, err := GetPubKeyFromBytes(in.PubKey)
			if err != nil {
				return err
			}
			hash := sha256.Sum256([]byte(tx.ID))
			if !verifyECDSA(pub, hash[:], in.Signature) {
				return fmt.Errorf("交易签名验证失败")
			}
		}
		for _, out := range tx.Outputs {
			outSum += out.Value
		}
		if inSum < outSum {
			return fmt.Errorf("资金来源不足: in=%d out=%d", inSum, outSum)
		}
	}
	return nil
}

// IsValid 校验整条链是否自洽（用于加载/防篡改检查）。
func (bc *Blockchain) IsValid() bool {
	for i := 1; i < len(bc.blocks); i++ {
		if err := bc.blocks[i].Validate(bc.blocks[i-1].Hash); err != nil {
			return false
		}
	}
	return true
}
