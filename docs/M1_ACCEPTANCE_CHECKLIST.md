# M1_ACCEPTANCE_CHECKLIST.md

# M1 项目骨架与本地环境验收清单

> 本清单用于检查 Codex 完成 M1 后是否达到进入 M2 的条件。

---

# 1. 项目结构检查

必须存在：

```text
backend/
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
backend/migrations/
docker/docker-compose.yml
README.md
.gitignore
AGENTS.md
```

---

# 2. 技术栈检查

必须使用：

```text
Go
Gin
GORM
MySQL driver
go-redis
validator
zap
JWT
Viper 或 cleanenv
```

禁止出现：

```text
Java
Spring Boot
Node.js backend
NestJS
Express backend
Python backend
FastAPI
```

---

# 3. 后端启动检查

执行：

```bash
cd backend
go mod tidy
go test ./...
go run ./cmd/server
```

要求：

```text
能编译
测试通过或没有测试文件但不报错
服务能启动
默认监听 8080
```

---

# 4. 健康检查接口

请求：

```http
GET http://127.0.0.1:8080/api/health
```

期望：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "ok"
  }
}
```

---

# 5. Docker Compose 检查

执行：

```bash
docker compose -f docker/docker-compose.yml config
docker compose -f docker/docker-compose.yml up -d
docker ps
```

要求：

```text
MySQL 8 容器可启动
Redis 7 容器可启动
MySQL 端口 3306
Redis 端口 6379
数据目录有挂载
```

---

# 6. 环境变量检查

`backend/.env.example` 必须包含：

```text
APP_ENV
APP_PORT
MYSQL_HOST
MYSQL_PORT
MYSQL_USER
MYSQL_PASSWORD
MYSQL_DATABASE
REDIS_HOST
REDIS_PORT
REDIS_DB
JWT_USER_SECRET
JWT_ADMIN_SECRET
WECHAT_MINI_APP_ID
WECHAT_MINI_APP_SECRET
FILE_STORAGE_TYPE
LOG_LEVEL
```

不得包含真实密钥。

---

# 7. 统一响应检查

必须有：

```text
ApiResponse
PageResponse
Success response helper
Error response helper
```

错误码至少有：

```text
0
10000
10001
10006
10007
10008
10009
```

---

# 8. 目录职责检查

不应出现：

```text
所有代码堆在 main.go
业务逻辑写在 handler 中
数据库业务表提前实现
账号合并提前实现
家庭关系提前实现
```

M1 只允许有骨架和健康检查。

---

# 9. Git 检查

`.gitignore` 必须包含：

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

---

# 10. M1 通过标准

满足以下条件即可进入 M2：

```text
1. Go 后端骨架清晰。
2. Docker Compose 可启动 MySQL / Redis。
3. /api/health 正常。
4. .env.example 完整。
5. README 有启动说明。
6. 没有提前实现业务模块。
7. 没有违反 AGENTS.md 的技术栈规则。
```

---

# 11. M1 不通过的典型情况

```text
生成了 Java / Spring Boot 项目
生成了 Node.js 后端
没有 Docker Compose
没有 .env.example
没有 /api/health
把所有代码写在 main.go
提前实现大量业务导致结构混乱
没有统一响应结构
没有错误码结构
```
