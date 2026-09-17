#!/usr/bin/env bash
# 开发/运行辅助脚本。用法:
#   sh scripts/dev.sh build    # 编译检查
#   sh scripts/dev.sh run      # 编译并运行
#   sh scripts/dev.sh clean    # 清理编译产物
# 在 WSL 中执行。若 go 不在 PATH，脚本会自动补 /usr/local/go/bin。
set -euo pipefail

export PATH="/usr/local/go/bin:$PATH"
if ! command -v go >/dev/null 2>&1; then
  echo "未找到 go，请安装并确认 /usr/local/go/bin 存在。" >&2
  exit 1
fi

DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$DIR"

case "${1:-build}" in
  build)
    echo ">>> go build ./..."
    go build ./...
    ;;
  run)
    echo ">>> go run ."
    go run .
    ;;
  runauto)
    # 自动演示：挖矿 -> 转账 -> 余额 -> 打印链 -> 退出
    printf '1\n2\n3\n4\n0\n' | go run .
    ;;
  runtamper)
    # 篡改演示
    printf '5\n0\n' | go run .
    ;;
  test)
    echo ">>> go test ./..."
    go test ./... -v
    ;;
  clean)
    echo ">>> 清理"
    rm -f "$DIR/mini-chain"
    ;;
  *)
    echo "未知命令: ${1:-}" >&2
    exit 1
    ;;
esac
