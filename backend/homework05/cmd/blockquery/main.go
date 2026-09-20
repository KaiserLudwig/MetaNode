// 任务 1-2：查询指定区块号的区块信息（哈希、时间戳、交易数量等）。
//
// 用法：
//
//	go run ./cmd/blockquery            # 查询最新区块
//	go run ./cmd/blockquery 11727291   # 查询指定区块号
package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"time"

	"metanode-homework05/internal/ethutil"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, rpc, err := ethutil.Dial(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("获取 chainID 失败: %w", err)
	}
	fmt.Println("========== 任务 1-2：查询 Sepolia 区块 ==========")
	fmt.Println("RPC 节点 :", rpc)
	fmt.Println("链 ID    :", chainID)

	var number *big.Int
	if len(os.Args) > 1 {
		n, err := strconv.ParseUint(os.Args[1], 10, 64)
		if err != nil {
			return fmt.Errorf("区块号 %q 不合法: %w", os.Args[1], err)
		}
		number = new(big.Int).SetUint64(n)
		fmt.Println("查询区块 :", n, "（指定）")
	} else {
		fmt.Println("查询区块 : 最新（latest）")
	}

	block, err := client.BlockByNumber(ctx, number)
	if err != nil {
		return fmt.Errorf("查询区块失败: %w", err)
	}

	fmt.Println("--------------------------------------------------")
	fmt.Printf("区块号     : %d\n", block.NumberU64())
	fmt.Printf("区块哈希   : %s\n", block.Hash().Hex())
	fmt.Printf("父区块哈希 : %s\n", block.ParentHash().Hex())
	fmt.Printf("时间戳     : %d（%s）\n", block.Time(), time.Unix(int64(block.Time()), 0).Format("2006-01-02 15:04:05"))
	fmt.Printf("交易数量   : %d\n", len(block.Transactions()))
	fmt.Printf("Gas 使用   : %d / %d\n", block.GasUsed(), block.GasLimit())
	fmt.Printf("出块者     : %s\n", block.Coinbase().Hex())
	if baseFee := block.BaseFee(); baseFee != nil {
		fmt.Printf("BaseFee    : %s wei\n", baseFee.String())
	}

	txs := block.Transactions()
	if len(txs) > 0 {
		fmt.Println("前几笔交易:")
		for i, tx := range txs {
			if i >= 3 {
				break
			}
			fmt.Printf("  [%d] %s\n", i, tx.Hash().Hex())
		}
	}
	fmt.Println("==================================================")
	return nil
}
