#!/bin/sh
# 依次运行作业 05 的三个任务，并把输出保存到 evidence/login
#
# 前置：
#   export PRIVATE_KEY=0x...              # Sepolia 账户私钥
#   export SEPOLIA_RPC_URL=...            # 或 export INFURA_API_KEY=...
set -e

cd "$(dirname "$0")/.." || exit 1
mkdir -p evidence

if [ -z "$PRIVATE_KEY" ]; then
  echo "请先设置 PRIVATE_KEY 环境变量" >&2
  exit 1
fi

echo "########## 任务 1-2：查询区块 ##########"
go run ./cmd/blockquery "${1:-11727291}" | tee evidence/01-blockquery.log

echo
echo "########## 任务 1-3：发送交易 ##########"
go run ./cmd/sendtx "${2:-0xe0bc8f3970b0fb8F0B53aaeA58c7d905494cf355}" "${3:-0.0001}" | tee evidence/02-sendtx.log

echo
echo "########## 任务 2-3：abigen 绑定与合约交互 ##########"
go run ./cmd/counter "${4:-0}" | tee evidence/03-counter.log

echo
echo "全部任务执行完成，日志见 evidence/"
