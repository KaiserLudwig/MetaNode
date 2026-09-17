#!/bin/sh
# 构建 Web 服务与 CLI，检查格式与 JS 语法
cd "$(dirname "$0")/.." || exit 1
export PATH=/usr/local/go/bin:/usr/local/node/bin:/usr/bin:/bin:$PATH
export GOPROXY=https://goproxy.cn,direct

echo "=== gofmt check ==="
if ! gofmt -l tasks/ web/ | grep -q .; then
  echo "FMT_OK"
else
  echo "FMT_ISSUES:"; gofmt -l tasks/ web/
  exit 1
fi

echo "=== go vet + build ==="
go vet ./... || exit 1
go build -o webdemo ./web || exit 1
go build -o demo . || exit 1
echo "BUILD_OK"

echo "=== JS syntax check ==="
FAIL=0
for f in web/static/demo-anim.js web/static/demos/*.js; do
  if ! node --check "$f"; then
    echo "SYNTAX_FAIL: $f"
    FAIL=1
  fi
done
if [ "$FAIL" = "0" ]; then
  echo "JS_CHECK_DONE"
fi
exit "$FAIL"
