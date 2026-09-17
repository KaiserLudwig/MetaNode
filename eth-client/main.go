// 阶段 B：以太坊 Sepolia 测试网交互工具。
//
// 运行（WSL 中）：
//   cd /mnt/f/学习/Web3/MetaNode/eth-client
//   export PATH=/usr/local/go/bin:$PATH
//   sh scripts/eth.sh build         # 编译
//   sh scripts/eth.sh dryrun        # 只读演示（无需私钥）
//   sh scripts/eth.sh run           # 交互式
//
// 子命令：
//   balance <地址>      查询账户余额（wei 与 ETH）
//   block   [高度]      查询区块信息（默认最新）
//   gas                 查询当前 gas 价格
//   tx <交易哈希>       查询交易详情
//   send <收款地址> <ETH金额>   从环境变量指定账户转账
//
// 环境变量：
//   ETH_RPC_URL   RPC 节点地址（默认 Sepolia 公共 RPC）
//   PRIVATE_KEY   转账账户私钥（0x... 或裸 64 位 hex），仅在 send 时需要
package main

import (
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"eth-client/internal/ethcli"
)

// 默认 RPC：Sepolia 公共节点。你也可通过环境变量 ETH_RPC_URL 覆盖。
const defaultRPC = "https://ethereum-sepolia-rpc.publicnode.com"

func main() {
	rpcURL := os.Getenv("ETH_RPC_URL")
	if rpcURL == "" {
		rpcURL = defaultRPC
	}

	cli, err := ethcli.New(rpcURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "初始化客户端失败:", err)
		os.Exit(1)
	}
	defer cli.Close()

	if len(os.Args) < 2 {
		usage()
		os.Exit(0)
	}
	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "balance":
		requireArgs(cmd, args, 1)
		cmdBalance(cli, args[0])
	case "block":
		cmdBlock(cli, args)
	case "gas":
		cmdGas(cli)
	case "tx":
		requireArgs(cmd, args, 1)
		cmdTx(cli, args[0])
	case "send":
		requireArgs(cmd, args, 2)
		cmdSend(cli, args[0], args[1])
	case "token":
		// token <代币地址> [owner 地址]：查询 ERC-20 信息；若给 owner 则多显示其持有量。
		requireArgs(cmd, args, 1)
		cmdToken(cli, args[0], optionalArg(args, 1))
	case "nft":
		requireArgs(cmd, args, 2)
		cmdNFT(cli, args[0], args[1])
	case "pool":
		requireArgs(cmd, args, 1)
		cmdPool(cli, args[0])
	case "--demo":
		demo(cli)
	default:
		usage()
	}
}

func usage() {
	fmt.Println(`用法: eth-client <命令> [参数]
命令:
  balance <地址>      查询账户余额（wei / ETH）
  block   [高度]      查询区块信息（默认最新）
  gas                 查询当前 gas 价格（gwei）
  tx <交易哈希>       查询交易详情
  send <收款地址> <ETH金额>   从 PRIVATE_KEY 环境变量账户转账
  token <代币地址>    查询 ERC-20 代币信息（name/symbol/decimals/总供应量）
  nft <NFT合约> <tokenId>   查询 ERC-721 NFT 信息（所有者/元数据URI）
  pool <池地址>       查询 Uniswap V2 池（token0/token1/储备/流动性）
  --demo              只读演示：查询链ID、最新区块、gas 价格
环境变量:
  ETH_RPC_URL  RPC 节点地址（默认 Sepolia 公共节点）
  PRIVATE_KEY  转账私钥（仅在 send 时使用）`)
}

// requireArgs 校验命令参数个数。
func requireArgs(cmd string, args []string, n int) {
	if len(args) < n {
		fmt.Fprintf(os.Stderr, "%s 需要 %d 个参数\n", cmd, n)
		os.Exit(1)
	}
}

func cmdBalance(cli *ethcli.Client, address string) {
	bal, err := cli.BalanceAt(address)
	if err != nil {
		fmt.Fprintln(os.Stderr, "查询余额失败:", err)
		return
	}
	fmt.Printf("地址:   %s\n", address)
	fmt.Printf("余额:   %v wei\n", bal)
	fmt.Printf("        %v ETH\n", ethcli.WeiToEth(bal))
}

func cmdBlock(cli *ethcli.Client, args []string) {
	var h uint64
	if len(args) >= 1 {
		n, err := strconv.ParseUint(args[0], 10, 64)
		if err != nil {
			fmt.Fprintln(os.Stderr, "无效的区块高度:", args[0])
			return
		}
		h = n
	} else {
		n, err := cli.BlockNumber()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		h = n
	}
	info, err := cli.BlockByNumber(h)
	if err != nil {
		fmt.Fprintln(os.Stderr, "查询区块失败:", err)
		return
	}
	fmt.Printf("高度:     #%v\n", info["number"])
	fmt.Printf("时间戳:   %v\n", info["timestamp"])
	fmt.Printf("区块哈希: %v\n", info["hash"])
	fmt.Printf("交易数:   %v\n", info["tx_count"])
	if ftx, ok := info["first_tx_hash"]; ok {
		fmt.Printf("首笔交易: %v\n", ftx)
		fmt.Println("（复制上面的哈希，用 `tx <哈希>` 查询完整交易详情）")
	}
}

func cmdGas(cli *ethcli.Client) {
	p, err := cli.GasPrice()
	if err != nil {
		fmt.Fprintln(os.Stderr, "查询gas价格失败:", err)
		return
	}
	fmt.Printf("当前 gas 单价: %v wei\n", p)
	fmt.Printf("              %.4f gwei\n", ethcli.WeiToGwei(p))
}

func cmdTx(cli *ethcli.Client, txHash string) {
	info, err := cli.TransactionByHash(txHash)
	if err != nil {
		fmt.Fprintln(os.Stderr, "查询交易失败:", err)
		return
	}
	fmt.Printf("交易哈希:   %v\n", info["hash"])
	fmt.Printf("是否待确认: %v\n", info["is_pending"])
	fmt.Printf("发送方:     %v\n", info["from"])
	fmt.Printf("接收方:     %v\n", info["to"])
	fmt.Printf("金额:       %v wei\n", info["value"])
	if v := info["value"].(*big.Int); v != nil {
		fmt.Printf("= %v ETH\n", ethcli.WeiToEth(v))
	}
	fmt.Printf("gas:        %v\n", info["gas"])
	fmt.Printf("gas_price:  %v\n", info["gas_price"])
}

func cmdSend(cli *ethcli.Client, to, ethAmount string) {
	priv := os.Getenv("PRIVATE_KEY")
	if priv == "" {
		fmt.Fprintln(os.Stderr, "缺少环境变量 PRIVATE_KEY（转账账户私钥）")
		os.Exit(1)
	}
	// 把 ETH 金额转成 wei。
	ethF, ok := new(big.Float).SetString(ethAmount)
	if !ok {
		fmt.Fprintln(os.Stderr, "无效的 ETH 金额:", ethAmount)
		return
	}
	weiF := new(big.Float).Mul(ethF, new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)))
	wei, _ := weiF.Int(nil)

	txHash, err := cli.SendWei(strings.TrimPrefix(priv, "0x"), to, wei)
	if err != nil {
		fmt.Fprintln(os.Stderr, "转账失败:", err)
		return
	}
	fmt.Println("✔ 交易已广播!")
	fmt.Println("  交易哈希:", txHash)
	fmt.Println("  可在区块浏览器查看:", "https://sepolia.etherscan.io/tx/"+txHash)
}

// demo 只读演示：链ID、最新区块、gas 价格、以及一个热门地址余额。
func demo(cli *ethcli.Client) {
	fmt.Println("========== Sepolia 只读演示 ==========")
	if id, err := cli.ChainID(); err == nil {
		fmt.Printf("链 ID: %v\n", id)
	}
	if n, err := cli.BlockNumber(); err == nil {
		fmt.Printf("最新区块: #%v\n", n)
	}
	if p, err := cli.GasPrice(); err == nil {
		fmt.Printf("gas 单价: %v wei (%.4f gwei)\n", p, ethcli.WeiToGwei(p))
	}
	// 用一个公开的测试地址演示余额查询（dead 地址）。
	demoAddr := "0x000000000000000000000000000000000000dEaD"
	if bal, err := cli.BalanceAt(demoAddr); err == nil {
		fmt.Printf("%s 余额: %v wei (%v ETH)\n", demoAddr, bal, ethcli.WeiToEth(bal))
	}
	fmt.Println("======================================")
}

// cmdToken 查询 ERC-20 代币信息。owner 为非空时额外查询其持有量。
func cmdToken(cli *ethcli.Client, tokenAddr, owner string) {
	token := common.HexToAddress(tokenAddr)
	name, err := cli.ERC20Name(token)
	if err != nil {
		fmt.Fprintln(os.Stderr, "查 name 失败:", err)
	} else {
		fmt.Printf("名称:      %s\n", name)
	}
	sym, err := cli.ERC20Symbol(token)
	if err != nil {
		fmt.Fprintln(os.Stderr, "查 symbol 失败:", err)
	} else {
		fmt.Printf("符号:      %s\n", sym)
	}
	dec, err := cli.ERC20Decimals(token)
	if err != nil {
		fmt.Fprintln(os.Stderr, "查 decimals 失败:", err)
	} else {
		fmt.Printf("精度:      %d decimals\n", dec)
	}
	supply, err := cli.ERC20TotalSupply(token)
	if err != nil {
		fmt.Fprintln(os.Stderr, "查 totalSupply 失败:", err)
	} else {
		fmt.Printf("总供应量:  %v（最小单位）\n", supply)
		// 按 decimals 折算为可读单位。
		if dec > 0 {
			unit := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(dec)), nil)
			fmt.Printf("           ≈ %v（10^%d 折算后）\n", new(big.Float).SetInt(supply).Quo(new(big.Float).SetInt(supply), new(big.Float).SetInt(unit)), dec)
		}
	}
	if owner != "" {
		ownerAddr := common.HexToAddress(owner)
		bal, err := cli.ERC20BalanceOf(token, ownerAddr)
		if err != nil {
			fmt.Fprintln(os.Stderr, "查 balanceOf 失败:", err)
		} else {
			fmt.Printf("%s 持有: %v（最小单位）\n", ownerAddr.Hex(), bal)
			if dec > 0 {
				unit := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(dec)), nil)
				fmt.Printf("               ≈ %v\n", new(big.Float).SetInt(bal).Quo(new(big.Float).SetInt(bal), new(big.Float).SetInt(unit)))
			}
		}
	}
}

// optionalArg 返回 args 的第 idx 项；越界时返回空串，用于可选参数。
func optionalArg(args []string, idx int) string {
	if idx < len(args) {
		return args[idx]
	}
	return ""
}

// cmdNFT 查询 ERC-721 NFT 信息。
func cmdNFT(cli *ethcli.Client, nftAddr, tokenIDStr string) {
	contract := common.HexToAddress(nftAddr)
	tokenID, ok := new(big.Int).SetString(tokenIDStr, 10)
	if !ok {
		fmt.Fprintln(os.Stderr, "无效的 tokenId:", tokenIDStr)
		return
	}
	name, err := cli.NFTName(contract)
	if err != nil {
		fmt.Fprintln(os.Stderr, "查 name 失败:", err)
	} else {
		fmt.Printf("NFT 名称:  %s\n", name)
	}
	sym, err := cli.NFTSymbol(contract)
	if err != nil {
		fmt.Fprintln(os.Stderr, "查 symbol 失败:", err)
	} else {
		fmt.Printf("符号:      %s\n", sym)
	}
	owner, err := cli.NFTOwnerOf(contract, tokenID)
	if err != nil {
		fmt.Fprintln(os.Stderr, "查 ownerOf 失败:", err)
	} else {
		fmt.Printf("tokenId %s 所有者: %s\n", tokenIDStr, owner.Hex())
	}
	uri, err := cli.NFTTokenURI(contract, tokenID)
	if err != nil {
		fmt.Fprintln(os.Stderr, "查 tokenURI 失败:", err)
	} else {
		fmt.Printf("元数据 URI: %s\n", uri)
	}
}

// cmdPool 查询 Uniswap V2 池信息。
func cmdPool(cli *ethcli.Client, poolAddr string) {
	pool := common.HexToAddress(poolAddr)
	t0, err := cli.PoolToken0(pool)
	if err != nil {
		fmt.Fprintln(os.Stderr, "查 token0 失败:", err)
	} else {
		fmt.Printf("token0:    %s\n", t0.Hex())
	}
	t1, err := cli.PoolToken1(pool)
	if err != nil {
		fmt.Fprintln(os.Stderr, "查 token1 失败:", err)
	} else {
		fmt.Printf("token1:    %s\n", t1.Hex())
	}
	r0, r1, err := cli.PoolReserves(pool)
	if err != nil {
		fmt.Fprintln(os.Stderr, "查 getReserves 失败:", err)
	} else {
		fmt.Printf("reserve0:  %v\n", r0)
		fmt.Printf("reserve1:  %v\n", r1)
		fmt.Printf("≈ 储备比例 %s : %s\n", r0.String(), r1.String())
	}
	supply, err := cli.PoolTotalSupply(pool)
	if err != nil {
		fmt.Fprintln(os.Stderr, "查 totalSupply 失败:", err)
	} else {
		fmt.Printf("LP 供应量: %v\n", supply)
	}
}
