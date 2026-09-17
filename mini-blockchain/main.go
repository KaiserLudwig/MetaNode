// mini-blockchain —— 一个教学用迷你区块链演示程序。
//
// 运行方式（WSL 中）：
//   cd /mnt/f/学习/Web3/MetaNode/mini-blockchain
//   export PATH=/usr/local/go/bin:$PATH
//   go run .
//
// 通过命令行演示：创建钱包、挖矿（PoW）、转账（UTXO + ECDSA 签名）、
// 查询余额、打印整条链、篡改演示。用纯标准库实现，零外部依赖。
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"metanode-bc/blockchain"
)

func main() {
	fmt.Println("==================================================")
	fmt.Println("  Mini Blockchain —— 迷你区块链教学演示")
	fmt.Println("==================================================")

	// 配置：难度越小挖得越快。教学用难度 3（前导 3 个 0），演示友好。
	const difficulty = 3
	const reward = 50 // 每个区块给矿工的币

	bc := blockchain.NewBlockchain(difficulty, reward)

	// 演示用钱包。
	alice := blockchain.NewWallet()
	bob := blockchain.NewWallet()
	miner := blockchain.NewWallet()

	fmt.Printf("Alice 地址: %s（.%s...）\n", alice.Address(), alice.Address()[:8])
	fmt.Printf("Bob   地址: %s（.%s...）\n", bob.Address(), bob.Address()[:8])
	fmt.Printf("矿工  地址: %s（.%s...）\n", miner.Address(), miner.Address()[:8])
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	const menu = `
命令:
  1)  挖一个区块（给矿工 50 币奖励，空交易）
  2)  Alice -> Bob 转账 25
  3)  查询 Alice / Bob / 矿工 余额
  4)  打印整条链
  5)  篡改演示（篡改一个区块的交易，看校验是否被拦截）
  0)  退出
`
	for {
		fmt.Print(menu)
		fmt.Print("> 输入编号: ")
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		switch line {
		case "1":
			minerMines(bc, miner, reader)
		case "2":
			transferDemo(bc, alice, bob, reader)
		case "3":
			showBalances(bc, alice, bob, miner)
		case "4":
			printChain(bc)
		case "5":
			tamperDemo(bc)
		case "0", "q", "quit":
			fmt.Println("再见！")
			return
		default:
			fmt.Println("无效输入，请重试。")
		}
	}
}

func minerMines(bc *blockchain.Blockchain, miner *blockchain.Wallet, reader *bufio.Reader) {
	txs := []*blockchain.Transaction{blockchain.NewCoinbaseTx(miner.Address(), bc.GetCoinbaseReward())}
	block, err := bc.MineBlock(txs)
	if err != nil {
		fmt.Println("挖矿失败:", err)
		return
	}
	fmt.Printf("✔ 已挖到并上链区块 #%d（难度 %d, nonce=%d）\n", block.Index, bc.GetDifficulty(), block.Nonce)
}

func transferDemo(bc *blockchain.Blockchain, alice, bob *blockchain.Wallet, reader *bufio.Reader) {
	// 若 Alice 尚无余额，先给她挖一个 coinbase 奖励区块，使其有钱可转。
	if bc.BalanceOf(alice.Address()) < 25 {
		fmt.Println("Alice 余额不足，先给 Alice 挖一个奖励区块……")
		rewardTx := blockchain.NewCoinbaseTx(alice.Address(), bc.GetCoinbaseReward())
		blk, err := bc.MineBlock([]*blockchain.Transaction{rewardTx})
		if err != nil {
			fmt.Println("给 Alice 发奖励失败:", err)
			return
		}
		fmt.Printf("✔ Alice 已获得奖励并上链到区块 #%d\n", blk.Index)
	}

	tx, err := bc.SignAndBuildTransfer(alice, bob.Address(), 25)
	if err != nil {
		fmt.Println("构建交易失败:", err)
		return
	}
	fmt.Printf("✔ 交易已构建: %s\n", tx.String())
	txs := []*blockchain.Transaction{tx}
	block, err := bc.MineBlock(txs)
	if err != nil {
		fmt.Println("打包上链失败:", err)
		return
	}
	fmt.Printf("✔ 交易已上链到区块 #%d（nonce=%d）\n", block.Index, block.Nonce)
}

func showBalances(bc *blockchain.Blockchain, alice, bob, miner *blockchain.Wallet) {
	fmt.Println("---------------- 余额（未花费输出） ----------------")
	fmt.Printf("Alice: %d\n", bc.BalanceOf(alice.Address()))
	fmt.Printf("Bob:   %d\n", bc.BalanceOf(bob.Address()))
	fmt.Printf("矿工:  %d\n", bc.BalanceOf(miner.Address()))
}

func printChain(bc *blockchain.Blockchain) {
	fmt.Println("---------------- 整条链 ----------------")
	for _, b := range bc.GetBlocks() {
		fmt.Println(b.String())
		for _, tx := range b.Transactions {
			fmt.Printf("    └─ %s（输入=%d 输出=%d）\n", tx.String(), len(tx.Inputs), len(tx.Outputs))
		}
	}
	if bc.IsValid() {
		fmt.Println("校验: OK，链自洽。")
	} else {
		fmt.Println("校验: FAIL，链已损坏！")
	}
}

func tamperDemo(bc *blockchain.Blockchain) {
	blocks := bc.GetBlocks()
	if len(blocks) < 1 {
		fmt.Println("链还没有区块。")
		return
	}
	// 复制最后一个区块（深拷贝交易），只篡改副本，不影响真实链。
	target := cloneBlock(blocks[len(blocks)-1])
	fmt.Println("⚠ 尝试篡改区块内容（改动副本，不影响真实链）……")
	if len(target.Transactions) > 0 {
		target.Transactions[0].Outputs[0].Value = 99999
	} else {
		target.Nonce = 54321
	}
	// 校验副本哈希是否自洽（应被拦截）。
	if err := target.Validate(blocks[len(blocks)-1].PrevHash); err != nil {
		fmt.Println("✔ 篡改检测生效:", err)
	} else {
		fmt.Println("（未检测到篡改，异常！）")
	}
	// 校验原链仍然完好。
	if bc.IsValid() {
		fmt.Println("✔ 真实链未受影响，依然自洽。")
	}
}

// cloneBlock 深拷贝一个区块及其交易，用于安全的篡改演示。
func cloneBlock(b *blockchain.Block) *blockchain.Block {
	cp := *b
	cp.Transactions = make([]*blockchain.Transaction, len(b.Transactions))
	for i, tx := range b.Transactions {
		txcp := *tx
		// 深拷贝输出，避免共享底层数组。
		txcp.Outputs = append([]blockchain.TXOutput(nil), tx.Outputs...)
		txcp.Inputs = append([]blockchain.TXInput(nil), tx.Inputs...)
		cp.Transactions[i] = &txcp
	}
	return &cp
}
