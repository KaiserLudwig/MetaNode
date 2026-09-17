#!/bin/sh
# 一键把 MetaNode 作业集推送到 GitHub（仓库不存在则自动创建）
#
# 用法：
#   sh upload_metanode.sh                          # 自动读取 ../.secrets/github_token.env
#   GITHUB_TOKEN=ghp_xxx sh upload_metanode.sh     # 或直接传环境变量
#   PUBLIC=false sh upload_metanode.sh             # 默认公开，设为 false 可建私有仓库
set -e

OWNER="${OWNER:-KaiserLudwig}"
REPO="${REPO:-MetaNode}"
API="https://api.github.com"
PUBLIC="${PUBLIC:-true}"
DESC="MetaNode Web3 学习作业集：迷你区块链 / Sepolia 客户端 / 博客后端 / Go 语言练习"

DIR="$(cd "$(dirname "$0")" && pwd)"
SECRET_FILE="$(cd "$DIR/.." && pwd)/.secrets/github_token.env"

if [ -z "${GITHUB_TOKEN:-}" ] && [ -f "$SECRET_FILE" ]; then
  # shellcheck disable=SC1090
  . "$SECRET_FILE"
fi
if [ -z "${GITHUB_TOKEN:-}" ]; then
  echo "缺少 token：请把 token 写入 $SECRET_FILE ，或设置 GITHUB_TOKEN 环境变量" >&2
  exit 1
fi

TMP="$(mktemp -d)"

echo "== 1. 验证 token =="
curl -sS -H "Authorization: token $GITHUB_TOKEN" "$API/user" -o "$TMP/user.json"
grep -q '"login"' "$TMP/user.json" || { echo "TOKEN 无效或已过期"; exit 1; }
echo "OK: $(grep -o '"login":"[^"]*"' "$TMP/user.json")"

echo "== 2. 创建仓库 $OWNER/$REPO (public=$PUBLIC) =="
curl -sS -X POST -H "Authorization: token $GITHUB_TOKEN" -H "Content-Type: application/json" \
  -d "{\"name\":\"$REPO\",\"description\":\"$DESC\",\"public\":$PUBLIC}" \
  "$API/user/repos" -o "$TMP/repo.json"
if grep -q '"full_name"' "$TMP/repo.json"; then
  echo "已创建: $(grep -o '"full_name":"[^"]*"' "$TMP/repo.json")"
elif grep -q 'already_exists' "$TMP/repo.json"; then
  echo "仓库已存在，跳过创建"
else
  echo "创建失败:"; cat "$TMP/repo.json"; exit 1
fi

echo "== 3. 推送 main =="
cd "$DIR"
git remote remove origin 2>/dev/null || true
git remote add origin "https://github.com/$OWNER/$REPO.git"
git push "https://$OWNER:$GITHUB_TOKEN@github.com/$OWNER/$REPO.git" main:main
git remote set-url origin "https://github.com/$OWNER/$REPO.git"
echo "完成: https://github.com/$OWNER/$REPO"
