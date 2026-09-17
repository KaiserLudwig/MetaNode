#!/usr/bin/env bash
# 确认 go-ethereum v1.15.0 中 abi.Unpack / abi.Pack 的签名。
set -euo pipefail
echo "==== Unpack/Pack 签名 ===="
grep -nE 'Unpack\(|Pack\(' "/home/dev/go/pkg/mod/github.com/ethereum/go-ethereum@v1.15.0/accounts/abi/abi.go" | head
echo "==== methods.go Unpack 相关 ===="
grep -nE 'func \(method Method\) UnpackOutput|UnpackInput|Unpack\(' "/home/dev/go/pkg/mod/github.com/ethereum/go-ethereum@v1.15.0/accounts/abi/method.go" 2>/dev/null | head
