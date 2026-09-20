#!/bin/sh
# 任务 2-1 / 2-2：编译 Counter.sol 生成 ABI 与字节码，再用 abigen 生成 Go 绑定
#
# 依赖：solc（可用 forge 安装的版本）、abigen（go install github.com/ethereum/go-ethereum/cmd/abigen@v1.15.0）
# 可用环境变量覆盖：SOLC=/path/to/solc ABIGEN=/path/to/abigen
set -e

cd "$(dirname "$0")/.." || exit 1

SOLC="${SOLC:-solc}"
ABIGEN="${ABIGEN:-abigen}"

echo "=== 1. 编译合约：solc -> ABI + 字节码 ==="
"$SOLC" --abi --bin --optimize --optimize-runs 200 --overwrite -o contracts contracts/Counter.sol
ls -l contracts/Counter.abi contracts/Counter.bin

echo
echo "=== 2. 生成 Go 绑定：abigen ==="
mkdir -p bindings/counter
"$ABIGEN" --abi contracts/Counter.abi --bin contracts/Counter.bin --pkg counter --out bindings/counter/counter.go
head -20 bindings/counter/counter.go
echo "已生成 bindings/counter/counter.go"
