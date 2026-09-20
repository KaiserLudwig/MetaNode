// 任务 1-3：构造、签名并广播一笔 Sepolia 上的 ETH 转账交易。
//
// 用法：
//
//	go run ./cmd/sendtx <收款地址> <ETH 金额>       # 例如: go run ./cmd/sendtx 0xe0bc... 0.0001
package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"metanode-homework05/internal/ethutil"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 3 {
		return fmt.Errorf("用法: go run ./cmd/sendtx <收款地址> <ETH 金额>")
	}
	to := common.HexToAddress(os.Args[1])
	if to == (common.Address{}) {
		return fmt.Errorf("收款地址 %q 不合法", os.Args[1])
	}
	amount, ok := new(big.Float).SetString(os.Args[2])
	if !ok {
		return fmt.Errorf("金额 %q 不合法", os.Args[2])
	}
	value, _ := amount.Mul(amount, big.NewFloat(1e18)).Int(nil)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	client, rpc, err := ethutil.Dial(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	priv, from, err := ethutil.PrivateKey()
	if err != nil {
		return err
	}
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("获取 chainID 失败: %w", err)
	}

	balance, err := client.BalanceAt(ctx, from, nil)
	if err != nil {
		return fmt.Errorf("查询余额失败: %w", err)
	}

	fmt.Println("========== 任务 1-3：发送 ETH 转账 ==========")
	fmt.Println("RPC 节点 :", rpc)
	fmt.Println("发送方   :", from.Hex())
	fmt.Println("接收方   :", to.Hex())
	fmt.Printf("转账金额 : %s wei (%s ETH)\n", value.String(), os.Args[2])
	fmt.Printf("发送方余额: %s wei\n", balance.String())

	// 1) nonce
	nonce, err := client.PendingNonceAt(ctx, from)
	if err != nil {
		return fmt.Errorf("获取 nonce 失败: %w", err)
	}
	// 2) 动态手续费（EIP-1559）
	tipCap, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		return fmt.Errorf("获取 gasTipCap 失败: %w", err)
	}
	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		return fmt.Errorf("获取最新区块头失败: %w", err)
	}
	feeCap := new(big.Int).Add(tipCap, new(big.Int).Mul(header.BaseFee, big.NewInt(2)))

	// 3) 构造并签名交易（普通转账固定 21000 gas）
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: tipCap,
		GasFeeCap: feeCap,
		Gas:       21000,
		To:        &to,
		Value:     value,
	})
	signed, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), priv)
	if err != nil {
		return fmt.Errorf("签名失败: %w", err)
	}
	raw, err := signed.MarshalBinary()
	if err != nil {
		return fmt.Errorf("序列化失败: %w", err)
	}
	fmt.Printf("交易已签名（nonce=%d, tip=%s, feeCap=%s, 原始长度=%d 字节）\n", nonce, tipCap, feeCap, len(raw))

	// 4) 广播
	if err := client.SendTransaction(ctx, signed); err != nil {
		return fmt.Errorf("广播失败: %w", err)
	}
	fmt.Println("已广播，交易哈希:", signed.Hash().Hex())

	// 5) 等待打包
	receipt, err := waitMined(ctx, client, signed.Hash())
	if err != nil {
		return fmt.Errorf("等待交易回执失败: %w", err)
	}
	fmt.Printf("打包完成: 区块 %d, gas 使用 %d, 状态 %d\n", receipt.BlockNumber.Uint64(), receipt.GasUsed, receipt.Status)
	if receipt.Status == types.ReceiptStatusSuccessful {
		fmt.Println("交易执行成功 ✅")
	} else {
		fmt.Println("交易执行失败 ❌")
	}
	fmt.Println("=============================================")
	return nil
}

func waitMined(ctx context.Context, client *ethclient.Client, hash common.Hash) (*types.Receipt, error) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		receipt, err := client.TransactionReceipt(ctx, hash)
		if err == nil {
			return receipt, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}
