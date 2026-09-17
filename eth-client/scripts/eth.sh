#!/usr/bin/env bash
# 阶段 B：以太坊 Sepolia 工具的开发脚本。
# 用法: sh scripts/eth.sh <命令>
#   tidy    初始化依赖（下载 go-ethereum 等）
#   build   编译检查
#   run     交互式运行
#   test    运行单元测试
#   fmt     gofmt 格式化
set -euo pipefail

export PATH="/usr/local/go/bin:$PATH"
# 国内容器若下载依赖超时，取消下面注释改用国内代理：
# export GOPROXY=https://goproxy.cn,direct

DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$DIR"

case "${1:-build}" in
  tidy)
    echo ">>> go mod tidy"
    go mod tidy
    ;;
  build)
    echo ">>> go build ./..."
    go build ./...
    ;;
  run)
    echo ">>> go run ."
    go run .
    ;;
  test)
    echo ">>> go test ./..."
    go test ./... -v
    ;;
  dryrun)
    # 只读演示：查余额、查交易、最新区块、gas 价格（不需要私钥）
    echo ">>> go run . --demo"
    go run . --demo
    ;;
  sendtest)
    # 测试 send 在未设 PRIVATE_KEY 时不应广播（安全行为）
    echo ">>> 测试 send 未设私钥（应给出提示并退出）"
    env -u PRIVATE_KEY go run . send 0xAb72d1510B4D756C34E27e82bE4E81CCec0dBF65 0.01
    ;;
  defidemo)
    # 阶段 D 只读演示：ERC-20(USDC on Sepolia) + ERC-721(NFT on Sepolia) + Uniswap V2 池(主网只读)
    echo ">>> 阶段 D：DeFi / NFT 只读演示"
    echo "[1] ERC-20 代币（Sepolia USDC）"
    go run . token 0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238 0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf || true
    echo
    echo "[2] ERC-721 NFT（Sepolia Backchain Reward Booster, tokenId=11954）"
    go run . nft 0x937d178f593561c9db59564aa9b835d4d67593e9 11954 || true
    echo
    echo "[3] Uniswap V2 池（主网 WETH/USDC，只读示例）"
    ETH_RPC_URL=https://ethereum-rpc.publicnode.com go run . pool 0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc || true
    ;;
  *)
    echo "未知命令: ${1:-}" >&2
    exit 1
    ;;
esac
