#!/bin/sh
# 博客后端 API 全流程测试（覆盖作业全部要求 + 错误场景）
# 用法: sh scripts/api_test.sh   （默认访问 http://127.0.0.1:8080，可用 BASE 环境变量覆盖）
# 依赖: curl、sed（标准工具）
BASE=${BASE:-http://127.0.0.1:8080}
PASS=0
FAIL=0

check() { # 名称 期望状态码 实际状态码
  if [ "$2" = "$3" ]; then
    PASS=$((PASS + 1))
    echo "PASS  $1  (HTTP $3)"
  else
    FAIL=$((FAIL + 1))
    echo "FAIL  $1  (期望 $2, 实际 $3)"
    [ -n "$RESP" ] && echo "     响应: $(echo "$RESP" | head -c 200)"
  fi
}

call() { # 方法 路径 [data] [token] -> 设置 RESP / CODE
  RESP=$(curl -s -w '\n%{http_code}' -X "$1" \
    ${4:+-H "Authorization: Bearer $4"} \
    ${3:+-H 'Content-Type: application/json' -d "$3"} \
    "$BASE$2")
  CODE=$(echo "$RESP" | tail -n 1)
  RESP=$(echo "$RESP" | sed '$d')
}

echo "=============================================="
echo " 博客后端 API 全流程测试  BASE=$BASE"
echo "=============================================="

# ---------- 1. 健康检查 ----------
call GET /api/v1/healthz
check "健康检查 /healthz" 200 "$CODE"
echo "$RESP" | grep -q '"status":"ok"' && echo "PASS  健康检查返回 status=ok" || echo "FAIL  健康检查返回内容异常"

# ---------- 2. 注册 ----------
TS=$(date +%s)
U1="alice${TS}"; U2="bob${TS}"
P1="Passw0rd!${TS}"; P2="Passw0rd@${TS}"

call POST /api/v1/auth/register "{\"username\":\"$U1\",\"password\":\"$P1\",\"email\":\"$U1@test.com\"}"
check "注册用户A $U1" 201 "$CODE"

call POST /api/v1/auth/register "{\"username\":\"$U2\",\"password\":\"$P2\",\"email\":\"$U2@test.com\"}"
check "注册用户B $U2" 201 "$CODE"

call POST /api/v1/auth/register "{\"username\":\"$U1\",\"password\":\"$P1\",\"email\":\"$U1@test.com\"}"
check "重复注册（用户名已存在）" 409 "$CODE"

call POST /api/v1/auth/register "{\"username\":\"x\",\"password\":\"123\",\"email\":\"bad-email\"}"
check "非法注册（弱密码/非法邮箱）" 400 "$CODE"

# ---------- 3. 登录 ----------
call POST /api/v1/auth/login "{\"username\":\"$U1\",\"password\":\"wrong-pass\"}"
check "登录错误密码" 401 "$CODE"

call POST /api/v1/auth/login "{\"username\":\"$U1\",\"password\":\"$P1\"}"
check "登录成功" 200 "$CODE"
TOKEN_A=$(echo "$RESP" | sed -E 's/.*"token":"([^"]+)".*/\1/')
[ ${#TOKEN_A} -gt 20 ] && echo "PASS  获得 JWT（长度 ${#TOKEN_A}）" || echo "FAIL  未能解析出 JWT"

call POST /api/v1/auth/login "{\"username\":\"$U2\",\"password\":\"$P2\"}"
TOKEN_B=$(echo "$RESP" | sed -E 's/.*"token":"([^"]+)".*/\1/')

call GET /api/v1/auth/me "" "$TOKEN_A"
check "获取当前用户信息 /auth/me" 200 "$CODE"
echo "$RESP" | grep -q "$U1" && echo "PASS  /auth/me 返回用户A" || echo "FAIL  /auth/me 内容异常"

call GET /api/v1/auth/me ""
check "未认证访问 /auth/me" 401 "$CODE"

# ---------- 4. 文章创建（认证/权限） ----------
call POST /api/v1/posts "{\"title\":\"未认证创建\",\"content\":\"x\"}"
check "未认证创建文章" 401 "$CODE"

call POST /api/v1/posts "{\"title\":\"Go 并发实战\",\"content\":\"goroutine 与 channel 是核心。\"}" "$TOKEN_A"
check "用户A创建文章1" 201 "$CODE"
PID_A1=$(echo "$RESP" | sed -E 's/.*"id":([0-9]+).*/\1/')

call POST /api/v1/posts "{\"title\":\"GORM 高级技巧\",\"content\":\"预加载、钩子、事务。\"}" "$TOKEN_A"
check "用户A创建文章2" 201 "$CODE"
PID_A2=$(echo "$RESP" | sed -E 's/.*"id":([0-9]+).*/\1/')

call POST /api/v1/posts "{\"title\":\"区块链入门\",\"content\":\"从零理解去中心化。\"}" "$TOKEN_B"
check "用户B创建文章" 201 "$CODE"
PID_B=$(echo "$RESP" | sed -E 's/.*"id":([0-9]+).*/\1/')

call POST /api/v1/posts "{\"title\":\"\",\"content\":\"\"}" "$TOKEN_A"
check "空标题/空内容创建" 400 "$CODE"

# ---------- 5. 文章读取 ----------
call GET "/api/v1/posts?page=1&size=10"
check "文章列表（分页）" 200 "$CODE"
echo "$RESP" | grep -q '"total"' && echo "PASS  列表含 total/分页字段" || echo "FAIL  列表字段异常"

call GET "/api/v1/posts/$PID_A1"
check "文章详情" 200 "$CODE"
echo "$RESP" | grep -q "$U1" && echo "PASS  详情含作者信息" || echo "FAIL  详情缺少作者"

call GET /api/v1/posts/999999
check "不存在的文章详情" 404 "$CODE"

# ---------- 6. 文章更新（仅作者） ----------
call PUT "/api/v1/posts/$PID_A1" "{\"title\":\"Go 并发实战（修订版）\"}" "$TOKEN_B"
check "用户B修改用户A文章（越权）" 403 "$CODE"

call PUT "/api/v1/posts/$PID_A1" "{\"title\":\"Go 并发实战（修订版）\",\"content\":\"补充了 channel 示例。\"}" "$TOKEN_A"
check "作者修改自己的文章" 200 "$CODE"

call GET "/api/v1/posts/$PID_A1"
echo "$RESP" | grep -q "修订版" && echo "PASS  更新已生效" || echo "FAIL  更新未生效"

call PUT "/api/v1/posts/$PID_A1" "{}" "$TOKEN_A"
check "空更新（无字段）" 400 "$CODE"

# ---------- 7. 评论 ----------
call POST "/api/v1/posts/$PID_A1/comments" "{\"content\":\"好文！\"}" "$TOKEN_B"
check "用户B评论文章A1" 201 "$CODE"
CID=$(echo "$RESP" | sed -E 's/.*"id":([0-9]+).*/\1/')

call POST "/api/v1/posts/$PID_A1/comments" "{\"content\":\"自己顶一下\"}" "$TOKEN_A"
check "作者评论自己的文章" 201 "$CODE"

call POST "/api/v1/posts/$PID_A1/comments" "{\"content\":\"匿名评论\"}"
check "未认证发表评论" 401 "$CODE"

call POST "/api/v1/posts/999999/comments" "{\"content\":\"评论不存在的文章\"}" "$TOKEN_A"
check "评论不存在的文章" 404 "$CODE"

call GET "/api/v1/posts/$PID_A1/comments"
check "文章评论列表" 200 "$CODE"
echo "$RESP" | grep -q '"total":2' && echo "PASS  评论列表共 2 条" || echo "FAIL  评论数量异常"

# ---------- 8. 文章删除（仅作者，级联删评论） ----------
call DELETE "/api/v1/posts/$PID_B" "" "$TOKEN_A"
check "用户A删除用户B文章（越权）" 403 "$CODE"

call DELETE "/api/v1/posts/$PID_B" "" "$TOKEN_B"
check "作者删除自己的文章" 200 "$CODE"

call GET "/api/v1/posts/$PID_B"
check "删除后文章详情" 404 "$CODE"

call DELETE "/api/v1/posts/$PID_A2" "" "$TOKEN_A"
call DELETE "/api/v1/posts/$PID_A1" "" "$TOKEN_A"
check "清理：删除剩余测试文章" 200 "$CODE"

call GET "/api/v1/posts/$PID_A1/comments"
check "文章删除后评论列表" 404 "$CODE"

# ---------- 汇总 ----------
echo "=============================================="
echo " 结果: PASS=$PASS  FAIL=$FAIL"
echo "=============================================="
[ "$FAIL" = "0" ] && echo "ALL_TESTS_PASSED" || echo "SOME_TESTS_FAILED"
exit "$FAIL"
