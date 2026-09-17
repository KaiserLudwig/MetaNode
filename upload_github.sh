#!/bin/sh
# 一键创建 GitHub 公开仓库并推送两个项目
# 用法: GITHUB_TOKEN=ghp_xxx sh upload_github.sh
set -e
TOKEN="$GITHUB_TOKEN"
[ -n "$TOKEN" ] || { echo "缺少 GITHUB_TOKEN 环境变量"; exit 1; }
USER="KaiserLudwig"
API="https://api.github.com"
ROOT="/mnt/f/学习/Web3/MetaNode"

echo "== 1. 验证 token =="
curl -s -H "Authorization: token $TOKEN" "$API/user" -o /tmp/ghuser.json
grep -q '"login"' /tmp/ghuser.json || { echo "TOKEN 无效:"; cat /tmp/ghuser.json; exit 1; }
echo "OK: $(grep -o '"login":"[^"]*"' /tmp/ghuser.json)"

echo "== 2. 创建公开仓库 =="
create_repo() {
  NAME="$1"; DESC="$2"
  curl -s -X POST -H "Authorization: token $TOKEN" -H "Content-Type: application/json" \
    -d "{\"name\":\"$NAME\",\"description\":\"$DESC\",\"public\":true}" "$API/user/repos" -o /tmp/ghrepo.json
  if grep -q '"full_name"' /tmp/ghrepo.json; then
    echo "已创建: $(grep -o '"full_name":"[^"]*"' /tmp/ghrepo.json)"
  elif grep -q 'already_exists' /tmp/ghrepo.json; then
    echo "仓库已存在，跳过创建: $NAME"
  else
    echo "创建失败: $NAME"; cat /tmp/ghrepo.json; exit 1
  fi
}
create_repo "blog-backend" "个人博客系统后端：Go + Gin + GORM + JWT（文章 CRUD / 用户认证 / 评论 / 前端 SPA）"
create_repo "go-learn-demo" "Go 学习演示程序：指针 / Goroutine / 面向对象 / Channel / 锁机制 / 进阶 GORM，CLI 与 Web 双形态"

echo "== 3. 提交并推送 =="
push_repo() {
  NAME="$1"; MSG="$2"
  cd "$ROOT/$NAME" || exit 1
  echo "--- $NAME ---"
  git init -b main 2>/dev/null || git init
  git config user.name "$USER"
  git config user.email "$USER@users.noreply.github.com"
  git add -A
  git commit -m "$MSG" || echo "（无新提交）"
  git remote remove origin 2>/dev/null || true
  git remote add origin "https://github.com/$USER/$NAME.git"
  git push "https://$USER:$TOKEN@github.com/$USER/$NAME.git" main:main 2>&1 | tail -n 3
  git remote set-url origin "https://github.com/$USER/$NAME.git"
  echo "推送完成: https://github.com/$USER/$NAME"
}
push_repo "blog-backend" "个人博客系统后端：Gin + GORM + JWT（文章 CRUD / 用户认证 / 评论 / 前端 SPA）"
push_repo "go-learn-demo" "Go 学习演示程序：指针 / Goroutine / 面向对象 / Channel / 锁机制 / 进阶 GORM，CLI 与 Web 双形态"

echo "== 全部完成 =="
