#!/bin/sh
# 验证 Web 演示服务（静态资源 + 任务页面 + 真实运行结果）
export PATH=/usr/local/go/bin:/usr/bin:/bin:$PATH
BASE=http://127.0.0.1:8080
check() {
  CODE=$(curl -s -o /dev/null -w '%{http_code}' "$1")
  echo "HTTP $CODE  $1"
}
echo "--- 静态资源（新增 GORM 动画） ---"
check $BASE/static/demo-anim.css
check $BASE/static/demo-anim.js
for f in gorm_models gorm_query gorm_hooks; do
  check $BASE/static/demos/$f.js
done
echo "--- 任务页 ---"
for id in 11 12 13; do
  check "$BASE/task?id=$id"
done
echo "--- 任务11 真实运行结果 ---"
curl -s "$BASE/task?id=11" | grep -aoE 'AutoMigrate|CREATE TABLE (users|posts|comments)|users / posts / comments 三张表' | sort -u
echo "--- 任务12 真实运行结果 ---"
curl -s "$BASE/task?id=12" | grep -aoE '张三|《Go 并发模型入门》|评论数量最多|3 条' | sort -u | head -n 6
echo "--- 任务13 真实运行结果 ---"
curl -s "$BASE/task?id=13" | grep -aoE 'PostCount 自动变为 3|无评论|AfterDelete 钩子触发' | sort -u
echo "--- 任务页 boot 脚本 ---"
for id in 11 12 13; do
  curl -s "$BASE/task?id=$id" | grep -aoE "MND\.boot\($id\)" || echo "BOOT_FAIL $id"
done
echo "DONE"
