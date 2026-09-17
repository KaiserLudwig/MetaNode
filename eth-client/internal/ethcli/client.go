// Package ethcli 封装基于 go-ethereum 的以太坊交互逻辑。
//
// 本阶段（B）连接真实的 Sepolia 测试网，实现：
//   - 查询账户余额
//   - 查询交易详情
//   - 查询最新区块信息
//   - 查询当前 gas 价格
//   - 构造、签名并广播一笔真实的以太坊转账
package ethcli

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Client 包装 *ethclient.Client，向外暴露教学友好的方法。
type Client struct {
	rpc  *ethclient.Client
	ctx  context.Context
	chainID *big.Int
}

// New 通过 RPC URL 连接以太坊节点。
//
// 注意：RPC URL 只是与节点对话的「接口」，转账还需要私钥签名。
func New(rpcURL string) (*Client, error) {
	conn, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("连接 RPC 失败: %w", err)
	}
	c := &Client{rpc: conn, ctx: context.Background()}
	// 建立连接时顺手取一次 chainID，便于后续查询交易发送方 & 签名。
	if id, err := conn.NetworkID(c.ctx); err == nil {
		c.chainID = id
	}
	return c, nil
}

// Close 关闭底层 RPC 连接。
func (c *Client) Close() { c.rpc.Close() }

// ChainID 查询当前链的网络 ID（如 Sepolia = 11155111）。
// 网络 ID 用于区分不同的链（主网/测试网/私有链）。
func (c *Client) ChainID() (*big.Int, error) {
	id, err := c.rpc.NetworkID(c.ctx)
	if err != nil {
		return nil, fmt.Errorf("查询网络ID失败: %w", err)
	}
	return id, nil
}

// BalanceAt 查询某个地址在指定区块高度的余额（单位：wei）。
//
// 解析步骤：
//   地址字符串 -> common.Address（以太坊地址类型，校验 0x + 40 位十六进制）
//   blockNumber = nil 表示查询「最新已确认」区块的高度
func (c *Client) BalanceAt(address string) (*big.Int, error) {
	addr := common.HexToAddress(address)
	bal, err := c.rpc.BalanceAt(c.ctx, addr, nil)
	if err != nil {
		return nil, fmt.Errorf("查询余额失败: %w", err)
	}
	return bal, nil
}

// GasPrice 查询当前 gas 基础价（单位：wei/gas）。
// gas 是执行交易/调用的「燃料费」，按 gas 量 × gas 单价付费。
func (c *Client) GasPrice() (*big.Int, error) {
	p, err := c.rpc.SuggestGasPrice(c.ctx)
	if err != nil {
		return nil, fmt.Errorf("查询gas价格失败: %w", err)
	}
	return p, nil
}

// BlockNumber 查询最新区块高度。
func (c *Client) BlockNumber() (uint64, error) {
	n, err := c.rpc.BlockNumber(c.ctx)
	if err != nil {
		return 0, fmt.Errorf("查询区块高度失败: %w", err)
	}
	return n, nil
}

// BlockByNumber 按高度查询区块信息（时间戳、交易数、哈希、交易详情）。
func (c *Client) BlockByNumber(h uint64) (map[string]interface{}, error) {
	block, err := c.rpc.BlockByNumber(c.ctx, new(big.Int).SetUint64(h))
	if err != nil {
		return nil, fmt.Errorf("查询区块失败: %w", err)
	}
	// 用 map 返回便于打印，同时避免暴露过多内部类型。
	info := map[string]interface{}{
		"number":    block.Number(),
		"timestamp": block.Time(),
		"hash":      block.Hash().Hex(),
		"tx_count":  len(block.Transactions()),
	}
	// 附属作用：若有交易，返回第一笔的哈希，方便用来测试 tx 命令。
	if txs := block.Transactions(); len(txs) > 0 {
		info["first_tx_hash"] = txs[0].Hash().Hex()
	}
	return info, nil
}

// TransactionByHash 按交易哈希查询交易详情。
func (c *Client) TransactionByHash(txHash string) (map[string]interface{}, error) {
	hash := common.HexToHash(txHash)
	tx, isPending, err := c.rpc.TransactionByHash(c.ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("查询交易失败: %w", err)
	}
	if tx == nil {
		return nil, fmt.Errorf("未找到交易: %s", txHash)
	}
	h := tx.Hash().Hex()
	from, _ := types.Sender(types.LatestSignerForChainID(c.chainID), tx)
	info := map[string]interface{}{
		"hash":       h,
		"is_pending": isPending,
		"to":         tx.To(),
		"value":      tx.Value(),
		"gas":        tx.Gas(),
		"gas_price":  tx.GasPrice(),
		"from":       from.Hex(),
		"nonce":      tx.Nonce(),
	}
	return info, nil
}

// WeiToEth 把 wei 转成 ETH（big.Float），用于展示。
func WeiToEth(w *big.Int) *big.Float {
	return new(big.Float).SetInt(w).Quo(
		new(big.Float).SetInt(w),
		new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)),
	)
}

// WeiToGwei 把 wei gas 单价转成 gwei，用于展示。
func WeiToGwei(w *big.Int) *big.Float {
	return new(big.Float).SetInt(w).Quo(
		new(big.Float).SetInt(w),
		new(big.Float).SetInt(big.NewInt(1_000_000_000)),
	)
}

// SendWei 构造、签名并广播一笔转账，把 value wei 从 private key 对应的账户转给 to。
//
//     关键步骤（真实以太坊转账）：
//       1) 用私钥派生发送方地址
//       2) 查询发送方的当前 nonce（账户已发送的交易数，防止重放）
//       3) 估算 gas 上限 / 读取 gas 单价
//       4) 组装交易（这里用 Legacy 交易 type 0），用私钥签名
//       5) 广播到节点，拿到交易哈希
//
// 返回交易哈希。注意：广播成功 ≠ 已确认，需等矿工打包（通常数秒到一分钟）。
func (c *Client) SendWei(privateKeyHex, to string, valueWei *big.Int) (string, error) {
	// 1) 私钥 -> ECDSA 私钥对象 -> 发送方地址
	priv, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("解析私钥失败: %w", err)
	}
	fromAddr := crypto.PubkeyToAddress(priv.PublicKey)

	// 2) 查询 nonce（用 pending 状态，避免边发边查导致的重放）
	nonce, err := c.rpc.PendingNonceAt(c.ctx, fromAddr)
	if err != nil {
		return "", fmt.Errorf("查询nonce失败: %w", err)
	}

	// 3) 估算 gas 上限与 gas 单价
	//    ethereum.CallMsg.To 是 *common.Address（nil 表示合约创建）。
	toAddr := common.HexToAddress(to)
	gasLimit, err := c.rpc.EstimateGas(c.ctx, ethereum.CallMsg{
		From:  fromAddr,
		To:    &toAddr,
		Value: valueWei,
	})
	if err != nil {
		// 简单转账估算失败时用兜底值。
		gasLimit = 21000 // 标准以太坊转账的固定 gas 下限（transfer 不触发合约逻辑）
	}
	gasPrice, err := c.rpc.SuggestGasPrice(c.ctx)
	if err != nil {
		return "", fmt.Errorf("查询gas价格失败: %w", err)
	}

	// 4) 组装并签名交易。这里用 Legacy 交易（type 0），最简单直观。
	//    types.NewTransaction 的 to 参数是 common.Address（值类型）。
	tx := types.NewTransaction(nonce, toAddr, valueWei, gasLimit, gasPrice, nil)

	// chainID 用于防止交易在不同链上重放（等价于复用了 Nonce 原则）。
	if c.chainID == nil {
		id, err := c.ChainID()
		if err != nil {
			return "", err
		}
		c.chainID = id
	}
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(c.chainID), priv)
	if err != nil {
		return "", fmt.Errorf("签名交易失败: %w", err)
	}

	// 5) 广播交易。
	if err := c.rpc.SendTransaction(c.ctx, signedTx); err != nil {
		return "", fmt.Errorf("广播交易失败: %w", err)
	}
	return signedTx.Hash().Hex(), nil
}
