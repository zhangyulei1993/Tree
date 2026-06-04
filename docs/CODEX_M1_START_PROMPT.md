# CODEX_M1_START_PROMPT.md

# Codex M1 启动提示词

将下面这段完整复制给 Codex，作为第一阶段开发任务。

---

请先阅读项目根目录的 `AGENTS.md`，然后阅读 `docs/` 目录中的以下文档：

- docs/00-TECH-STACK-FINAL.md
- docs/01-system-architecture-and-tech-stack.md
- docs/02-GO-BACKEND-ARCHITECTURE.md
- docs/CODEX_TASKS.md
- docs/12-local-dev-setup.md
- docs/15-error-code-summary.md
- docs/16-development-milestones.md
- docs/24-production-readiness-checklist.md

当前只执行：

```text
M1：项目骨架与本地环境
```

## 本阶段目标

搭建可启动的项目骨架和本地开发环境。

本阶段只做基础设施，不实现具体业务功能。

## 允许修改

只允许修改或创建：

```text
backend/
docker/
scripts/
README.md
.gitignore
AGENTS.md
```

如果当前仓库还没有前端目录，可以创建空目录或基础 README：

```text
admin-web/
web/
miniapp/
```

但不要实现具体前端页面。

## 禁止修改

```text
docs/
```

除非文档中明显存在路径错误且需要记录说明，否则不要修改文档。

## 必须完成

### 1. Go 后端骨架

创建：

```text
backend/go.mod
backend/cmd/server/main.go
backend/internal/app/server.go
backend/internal/app/router.go
backend/internal/common/response
backend/internal/common/errors
backend/internal/common/enums
backend/internal/common/middleware
backend/internal/common/logger
backend/internal/common/config
backend/internal/common/transaction
backend/internal/common/validator
backend/internal/common/jwt
backend/internal/common/redis
backend/internal/operationlog
backend/internal/family/tree
backend/configs/config.example.yaml
backend/.env.example
backend/migrations/.gitkeep
```

### 2. 基础依赖

使用 Go 技术栈：

```text
Gin
GORM
MySQL driver
go-redis
go-playground/validator
zap
JWT
Viper 或 cleanenv
```

优先使用：

```text
github.com/gin-gonic/gin
gorm.io/gorm
gorm.io/driver/mysql
github.com/redis/go-redis/v9
github.com/go-playground/validator/v10
go.uber.org/zap
github.com/golang-jwt/jwt/v5
github.com/spf13/viper
```

### 3. 统一响应结构

实现统一响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

实现分页结构：

```json
{
  "records": [],
  "page": 1,
  "pageSize": 20,
  "total": 0
}
```

### 4. 错误码结构

根据 `docs/15-error-code-summary.md` 创建错误码基础结构。

本阶段不需要实现所有错误码，但至少要有：

```text
Success = 0
SystemError = 10000
InvalidParams = 10001
Unauthorized = 10007
TokenExpired = 10008
PhoneBindRequired = 10009
Forbidden = 10006
```

### 5. 健康检查接口

实现：

```http
GET /api/health
```

返回：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "ok"
  }
}
```

### 6. Docker Compose

创建：

```text
docker/docker-compose.yml
```

包含：

```text
MySQL 8
Redis 7
```

要求：

```text
MySQL 暴露 3306
Redis 暴露 6379
MySQL 默认数据库 tree_platform
数据目录挂载到 docker/mysql/data
Redis 数据目录挂载到 docker/redis/data
```

### 7. 环境变量

创建：

```text
backend/.env.example
```

至少包含：

```env
APP_ENV=dev
APP_NAME=Tree
APP_PORT=8080

MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_USER=tree_user
MYSQL_PASSWORD=tree_pass
MYSQL_DATABASE=tree_platform

REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

JWT_USER_SECRET=change_me_user_secret
JWT_ADMIN_SECRET=change_me_admin_secret

WECHAT_MINI_APP_ID=
WECHAT_MINI_APP_SECRET=

FILE_STORAGE_TYPE=local
FILE_STORAGE_LOCAL_PATH=./uploads

LOG_LEVEL=debug
```

### 8. README 启动说明

更新根目录 README，写明：

```text
启动 MySQL / Redis
启动 Go 后端
访问 /api/health
常用端口
```

### 9. .gitignore

必须包含：

```gitignore
.env
.env.*
!.env.example
docker/mysql/data
docker/redis/data
uploads
logs
node_modules
dist
.DS_Store
```

## 本阶段禁止事项

本阶段禁止：

```text
1. 不要实现账号注册登录。
2. 不要实现微信小程序登录。
3. 不要实现家庭创建。
4. 不要实现成员关系。
5. 不要实现数据库业务表 migration。
6. 不要实现后台页面。
7. 不要实现小程序页面。
8. 不要生成 Java / Spring Boot 项目。
9. 不要生成 Node.js / NestJS 后端。
10. 不要把所有代码写在 main.go。
```

## 运行检查

完成后请运行：

```bash
cd backend
gofmt -w .
go test ./...
go run ./cmd/server
```

如果因为环境缺少 MySQL / Redis 无法完全运行，请明确说明。

也请检查：

```bash
docker compose -f docker/docker-compose.yml config
```

## 完成后输出格式

请按以下格式输出：

```text
1. 已完成内容
2. 修改文件列表
3. 新增目录结构
4. 新增接口
5. 启动命令
6. 已运行命令
7. 测试结果
8. 未完成事项
9. 风险点
10. 下一步建议
```

## 下一阶段预告

M1 完成后，下一阶段是：

```text
M2：数据库 migration 与基础 GORM model
```

不要提前实现 M2。
