#!/usr/bin/env bash
# 在 WSL 内进入 eth-client 目录并执行 Go 相关命令，绕开 PowerShell 引号转义。
# 用法: sh /mnt/f/学习/Web3/MetaNode/eth-client/scripts/go_cmd.sh <args...>
#   不带参数时会打印帮助；等价于在该目录下 <args...>。
set -euo pipefail
export PATH="/usr/local/go/bin:$PATH"

# 启用国内代理，避免默认代理 (proxy.golang.org) 在国内超时。
# 如需改回官方代理，注释掉下一行即可。
export GOPROXY=https://goproxy.cn,direct

DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$DIR"

exec "$@"
