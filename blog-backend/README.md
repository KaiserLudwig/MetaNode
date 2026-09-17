# 个人博客系统后端（Go + Gin + GORM）

基于 **Go / Gin / GORM** 的个人博客系统后端，实现文章 CRUD、用户认证（bcrypt + JWT）与评论功能，
并具备统一错误处理、结构化日志、分页查询等工程化能力。

## 技术栈

| 组件 | 选型 | 说明 |
|---|---|---|
| Web 框架 | [gin-gonic/gin](https://github.com/gin-gonic/gin) | 高性能 HTTP 框架 |
| ORM | [gorm.io/gorm](https://gorm.io) | 模型、迁移、关联查询 |
| 数据库 | SQLite（[glebarez/sqlite](https://github.com/glebarez/sqlite)） | 纯 Go 驱动，零配置，无需 CGO |
| 认证 | JWT（[golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt)）+ bcrypt | HS256，24 小时有效 |
| 日志 | 标准库 `log/slog` | 结构化 JSON 日志，控制台 + 文件双输出 |
| 校验 | gin 内置 validator | 请求参数校验（binding 标签） |

## 功能清单

- ✅ **前端交互界面（SPA）**：浏览器打开 `http://localhost:8080` 即可操作——
  注册/登录、浏览文章、发布/编辑/删除文章、发表评论（hash 路由，无框架零依赖）
- ✅ 用户注册（bcrypt 加密存储，用户名/邮箱唯一性校验）
- ✅ 用户登录（校验密码，签发 JWT；统一 401 提示，不泄露用户是否存在）
- ✅ 文章 CRUD：列表（分页+作者+评论数）、详情、创建（需认证）、
  更新/删除（**仅作者本人**，删除时事务级联删除评论）
- ✅ 评论：发表（需认证）、按文章查询评论列表（含评论作者）
- ✅ 统一响应格式 `{code, message, data}` + 业务错误码（40000/40100/40300/40400/40900/50000）
- ✅ 统一错误处理：GORM 错误映射、全局 panic 恢复、4xx/5xx 分级日志
- ✅ 结构化日志：请求日志（方法/路径/状态/耗时/IP）、启动与关停日志
- ✅ 优雅关停（SIGINT/SIGTERM + 10s 超时）
- ✅ 分页查询、Preload 预加载、事务、启动种子数据（demo 账号）
- ✅ 内嵌 API 文档页：`http://localhost:8080/api-doc`

## 项目结构

```
blog-backend/
├── cmd/
│   └── api/main.go          # 入口：配置、日志、路由、优雅关停
├── internal/
│   ├── config/config.go     # 环境变量配置
│   ├── database/database.go # SQLite 连接、自动迁移、种子数据
│   ├── models/models.go     # GORM 模型 + 请求/响应 DTO
│   ├── web/
│   │   ├── embed.go         # 内嵌资源声明
│   │   ├── index.html       # API 文档页（/api-doc）
│   │   └── app/             # 前端 SPA（/ 与 /static/*）
│   │       ├── index.html   # 应用骨架
│   │       ├── app.css      # 样式
│   │       └── app.js       # hash 路由 + API 调用 + 页面渲染
│   ├── handlers/
│   │   ├── auth.go          # 注册 / 登录 / 当前用户
│   │   ├── post.go          # 文章列表/详情/创建/更新/删除
│   │   └── comment.go       # 评论列表 / 发表评论
│   ├── middleware/
│   │   ├── auth.go          # JWT 认证中间件
│   │   └── logger.go        # 请求日志 + panic 恢复
│   └── utils/
│       ├── errors.go        # 业务错误类型与归一化
│       ├── response.go      # 统一响应
│       └── jwt.go           # JWT 签发与解析
├── scripts/
│   ├── build.sh             # 构建脚本（tidy/gofmt/vet/build）
│   └── api_test.sh          # API 全流程自动化测试（30 个用例）
├── go.mod / go.sum
└── blog.db                  # SQLite 数据库（运行时生成）
```

## 快速开始（WSL / Linux）

```bash
cd /mnt/f/学习/Web3/MetaNode/blog-backend

# 构建（自动下载依赖，需联网）
sh scripts/build.sh

# 启动（默认 0.0.0.0:8080）
./blog-server

# 全流程自动化测试（另开终端）
sh scripts/api_test.sh
```

启动后可用演示账号体验：**demo / demo123456**（空库时自动写入 3 篇文章）。

浏览器访问 **http://localhost:8080** 进入前端应用（注册/登录、发文章、评论）；
API 文档见 **http://localhost:8080/api-doc**。

## 配置（环境变量）

| 变量 | 默认值 | 说明 |
|---|---|---|
| `PORT` | `8080` | 监听端口 |
| `DB_PATH` | `blog.db` | SQLite 数据库文件 |
| `JWT_SECRET` | 开发默认值 | JWT 签名密钥（**生产环境必须注入**） |
| `LOG_FILE` | `logs/app.log` | 日志文件 |
| `SEED_DEMO` | `true` | 是否写入演示数据 |
| `DEBUG` | - | 设为 `1` 开启 Debug 日志 |

## API 文档

统一响应：`{"code": 0, "message": "ok", "data": ...}`，`code=0` 为成功。

### 认证

| 方法 | 路径 | 认证 | 说明 |
|---|---|---|---|
| POST | `/api/v1/auth/register` | - | 注册 `{username, password, email}` |
| POST | `/api/v1/auth/login` | - | 登录，返回 `{token, user}` |
| GET | `/api/v1/auth/me` | ✅ | 当前登录用户信息 |

```jsonc
// POST /api/v1/auth/login
{ "username": "demo", "password": "demo123456" }

// 200
{
  "code": 0, "message": "ok",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "user": { "id": 1, "username": "demo", "email": "demo@example.com", "created_at": "..." }
  }
}
```

### 文章

| 方法 | 路径 | 认证 | 说明 |
|---|---|---|---|
| GET | `/api/v1/posts?page=1&size=10` | - | 文章列表（分页，含作者与评论数） |
| GET | `/api/v1/posts/:id` | - | 文章详情 |
| POST | `/api/v1/posts` | ✅ | 创建文章 `{title, content}` |
| PUT | `/api/v1/posts/:id` | ✅ 作者 | 更新文章（部分字段更新） |
| DELETE | `/api/v1/posts/:id` | ✅ 作者 | 删除文章（级联删除评论） |

认证请求示例：

```bash
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"title":"我的新文章","content":"正文内容"}'
```

### 评论

| 方法 | 路径 | 认证 | 说明 |
|---|---|---|---|
| GET | `/api/v1/posts/:id/comments` | - | 文章评论列表（含评论作者） |
| POST | `/api/v1/posts/:id/comments` | ✅ | 发表评论 `{content}` |

### 错误码

| code | HTTP | 含义 |
|---|---|---|
| 0 | 200/201 | 成功 |
| 40000 | 400 | 请求参数错误 |
| 40100 | 401 | 未登录 / 令牌无效或过期 |
| 40300 | 403 | 无权限（只能操作自己的资源） |
| 40400 | 404 | 资源不存在 |
| 40900 | 409 | 资源冲突（用户名/邮箱已占用） |
| 50000 | 500 | 服务器内部错误 |

## 测试

`scripts/api_test.sh` 覆盖 30 个用例：健康检查、注册（含冲突/非法参数）、登录（含错误密码）、
JWT 认证（未认证 401）、文章 CRUD（含越权 403、404）、评论（含未认证、404）、删除级联等。

```bash
sh scripts/api_test.sh
```

完整测试结果见 [TEST_RESULTS.md](./TEST_RESULTS.md)。也可以在 Windows 下直接访问
`http://localhost:8080/api/v1/healthz`（WSL2 localhost 端口转发）。

## 备注

- 数据库为 SQLite 单文件，删除 `blog.db` 即可重置数据
- 参考代码中的 `dgrijalva/jwt-go` 已停止维护，本项目使用官方推荐的 `golang-jwt/jwt/v5`
- 作业要求的 MySQL 亦可支持：替换驱动与 DSN 即可（`gorm.io/driver/mysql`）
