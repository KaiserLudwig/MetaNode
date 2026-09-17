#!/bin/sh
# 构建并启动 Web 演示服务（在 WSL 内运行）
# 用法: sh scripts/start_web.sh   （默认端口 8080，可用 PORT=8090 指定）
cd "$(dirname "$0")/.." || exit 1
export PATH=/usr/local/go/bin:/usr/bin:/bin:$PATH
go build -o webdemo ./web || exit 1
echo "服务启动中... 请在 Windows 浏览器访问 http://localhost:${PORT:-8080}"
exec ./webdemo -port "${PORT:-8080}"
