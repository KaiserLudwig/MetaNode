// 任务 2-3：使用 abigen 生成的 Go 绑定与 Sepolia 上的 Counter 合约交互。
//
// 用法：
//
//	go run ./cmd/counter              # 部署一个新合约（初值 0），再调用 inc / incBy / get
//	go run ./cmd/counter 10           # 部署时把初值设为 10
//	COUNTER_ADDRESS=0x... go run ./cmd/counter   # 直接连接已部署的合约
package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"metanode-homework05/bindings/counter"
	"metanode-homework05/internal/ethutil"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
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

	// 构造交易签名器，统一使用 EIP-1559 动态手续费
	auth, err := bind.NewKeyedTransactorWithChainID(priv, chainID)
	if err != nil {
		return fmt.Errorf("构造签名器失败: %w", err)
	}
	auth.Context = ctx
	tipCap, err := client.SuggestGasTipCap(ctx)
	if err != nil {
		return fmt.Errorf("获取 gasTipCap 失败: %w", err)
	}
	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		return fmt.Errorf("获取最新区块头失败: %w", err)
	}
	auth.GasTipCap = tipCap
	auth.GasFeeCap = new(big.Int).Add(new(big.Int).Mul(tipCap, big.NewInt(2)), new(big.Int).Mul(header.BaseFee, big.NewInt(2)))
	auth.GasLimit = 400000

	fmt.Println("========== 任务 2-3：用 abigen 绑定调用链上合约 ==========")
	fmt.Println("RPC 节点 :", rpc)
	fmt.Println("调用账户 :", from.Hex())

	var (
		contract *counter.Counter
		address  common.Address
	)

	if envAddr := os.Getenv("COUNTER_ADDRESS"); envAddr != "" {
		address = common.HexToAddress(envAddr)
		contract, err = counter.NewCounter(address, client)
		if err != nil {
			return fmt.Errorf("连接已部署合约失败: %w", err)
		}
		fmt.Println("复用已部署合约:", address.Hex())
	} else {
		initial := big.NewInt(0)
		if len(os.Args) > 1 {
			n, err := strconv.ParseUint(os.Args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("初值 %q 不合法: %w", os.Args[1], err)
			}
			initial = new(big.Int).SetUint64(n)
		}

		deployedAddr, tx, deployed, err := counter.DeployCounter(auth, client, initial)
		if err != nil {
			return fmt.Errorf("部署合约失败: %w", err)
		}
		receipt, err := bind.WaitMined(ctx, client, tx)
		if err != nil {
			return fmt.Errorf("等待部署确认失败: %w", err)
		}
		address = deployedAddr
		contract = deployed
		fmt.Printf("合约已部署: %s\n", address.Hex())
		fmt.Printf(" 部署交易 : %s（区块 %d, gas %d）\n", tx.Hash().Hex(), receipt.BlockNumber.Uint64(), receipt.GasUsed)
		if err := printCount(ctx, contract, "部署后初值"); err != nil {
			return err
		}
	}

	// 调用 inc()：计数 +1
	txInc, err := contract.Inc(auth)
	if err != nil {
		return fmt.Errorf("调用 inc 失败: %w", err)
	}
	receiptInc, err := bind.WaitMined(ctx, client, txInc)
	if err != nil {
		return fmt.Errorf("等待 inc 确认失败: %w", err)
	}
	fmt.Printf("调用 inc() 交易: %s（区块 %d, gas %d）\n", txInc.Hash().Hex(), receiptInc.BlockNumber.Uint64(), receiptInc.GasUsed)
	if err := printCount(ctx, contract, "inc() 之后"); err != nil {
		return err
	}

	// 调用 incBy(5)：计数 +5
	txIncBy, err := contract.IncBy(auth, big.NewInt(5))
	if err != nil {
		return fmt.Errorf("调用 incBy 失败: %w", err)
	}
	receiptIncBy, err := bind.WaitMined(ctx, client, txIncBy)
	if err != nil {
		return fmt.Errorf("等待 incBy 确认失败: %w", err)
	}
	fmt.Printf("调用 incBy(5) 交易: %s（区块 %d, gas %d）\n", txIncBy.Hash().Hex(), receiptIncBy.BlockNumber.Uint64(), receiptIncBy.GasUsed)
	if err := printCount(ctx, contract, "incBy(5) 之后"); err != nil {
		return err
	}

	owner, err := contract.Owner(&bind.CallOpts{Context: ctx})
	if err == nil {
		fmt.Println("合约 owner:", owner.Hex())
	}
	fmt.Println("=========================================================")
	return nil
}

func printCount(ctx context.Context, contract *counter.Counter, label string) error {
	count, err := contract.Get(&bind.CallOpts{Context: ctx})
	if err != nil {
		return fmt.Errorf("读取计数失败: %w", err)
	}
	fmt.Printf("%-16s 计数 = %s\n", label, count.String())
	return nil
}
