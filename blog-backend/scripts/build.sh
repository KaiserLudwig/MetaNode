#!/bin/sh
# 构建博客后端：go mod tidy + 格式检查 + vet + 编译
cd "$(dirname "$0")/.." || exit 1
export PATH=/usr/local/go/bin:/usr/bin:/bin:$PATH
export GOPROXY=https://goproxy.cn,direct

echo "=== go mod tidy ==="
go mod tidy || exit 1

echo "=== gofmt check ==="
if gofmt -l cmd internal | grep -q .; then
  echo "FMT_ISSUES:"; gofmt -l cmd internal
  exit 1
fi
echo "FMT_OK"

echo "=== go vet ==="
go vet ./... || exit 1

echo "=== build ==="
go build -o blog-server ./cmd/api || exit 1
echo "BUILD_OK"
