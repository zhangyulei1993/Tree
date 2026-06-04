# 12. 本地开发环境搭建说明 v1

> 目标：让 Codex 或开发团队能够在本地一次性启动 MySQL、Redis、Go 后端、管理后台、PC/H5 前台和小程序开发环境。

## 1. 推荐工具版本

| 工具 | 推荐版本 | 用途 |
|---|---:|---|
| Go | 1.22+ | 后端开发 |
| Node.js | 20 LTS / 22 LTS | 前端与 uni-app |
| pnpm | 9+ | 前端包管理 |
| Docker Desktop | 最新稳定版 | MySQL / Redis |
| Docker Compose | v2+ | 本地服务编排 |
| MySQL | 8.x | 主数据库 |
| Redis | 7.x | 缓存、验证码、限流 |
| Apifox / Postman | 最新版 | API 联调 |
| 微信开发者工具 | 最新版 | 小程序调试 |

## 2. 推荐项目目录

```text
Tree
├── docs
├── backend
├── admin-web
├── web
├── miniapp
├── docker
├── scripts
└── README.md
```

## 3. Docker Compose

建议创建 `docker/docker-compose.yml`：

```yaml
version: "3.9"
services:
  mysql:
    image: mysql:8.0
    container_name: tree-mysql
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: 123456
      MYSQL_DATABASE: tree_platform
      MYSQL_USER: tree_user
      MYSQL_PASSWORD: tree_pass
      TZ: Asia/Shanghai
    ports:
      - "3306:3306"
    volumes:
      - ./mysql/data:/var/lib/mysql
      - ./mysql/init:/docker-entrypoint-initdb.d
    command:
      ["--character-set-server=utf8mb4", "--collation-server=utf8mb4_0900_ai_ci", "--default-time-zone=+08:00"]

  redis:
    image: redis:7
    container_name: tree-redis
    restart: unless-stopped
    ports:
      - "6379:6379"
    volumes:
      - ./redis/data:/data
    command: ["redis-server", "--appendonly", "yes"]
```

启动：

```bash
cd docker
docker compose up -d
```

## 4. 后端 `.env.example`

建议创建 `backend/.env.example`：

```env
APP_ENV=dev
APP_NAME=Tree
APP_PORT=8080

MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_USER=tree_user
MYSQL_PASSWORD=tree_pass
MYSQL_DATABASE=tree_platform
MYSQL_CHARSET=utf8mb4

REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

JWT_USER_SECRET=change_me_user_secret
JWT_ADMIN_SECRET=change_me_admin_secret
JWT_ACCESS_TOKEN_EXPIRE_MINUTES=120
JWT_REFRESH_TOKEN_EXPIRE_DAYS=7

VERIFY_CODE_EXPIRE_SECONDS=300
VERIFY_CODE_COOLDOWN_SECONDS=60

WECHAT_MINI_APP_ID=
WECHAT_MINI_APP_SECRET=

FILE_STORAGE_TYPE=local
FILE_STORAGE_LOCAL_PATH=./uploads
LOG_LEVEL=debug
```

## 5. Go 后端启动

推荐依赖：

```bash
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/mysql
go get github.com/redis/go-redis/v9
go get github.com/go-playground/validator/v10
go get go.uber.org/zap
go get github.com/golang-jwt/jwt/v5
go get github.com/spf13/viper
go get github.com/swaggo/gin-swagger
go get github.com/swaggo/files
```

启动：

```bash
cd backend
go mod tidy
go run ./cmd/server
```

健康检查：

```http
GET /api/health
```

期望返回：

```json
{"code":0,"message":"success","data":{"status":"ok"}}
```

## 6. 数据库 migration

推荐工具：`golang-migrate`。

迁移目录：

```text
backend/migrations
```

执行：

```bash
migrate -path ./migrations -database "mysql://tree_user:tree_pass@tcp(127.0.0.1:3306)/tree_platform?multiStatements=true" up
```

回滚一版：

```bash
migrate -path ./migrations -database "mysql://tree_user:tree_pass@tcp(127.0.0.1:3306)/tree_platform?multiStatements=true" down 1
```

## 7. 前端启动

管理后台：

```bash
cd admin-web
pnpm install
pnpm dev
```

`admin-web/.env.development`：

```env
VITE_APP_TITLE=Tree 家脉亲缘平台管理后台
VITE_API_BASE_URL=http://127.0.0.1:8080
```

PC/H5 前台：

```bash
cd web
pnpm install
pnpm dev
```

`web/.env.development`：

```env
VITE_APP_TITLE=Tree 家脉亲缘平台
VITE_API_BASE_URL=http://127.0.0.1:8080
```

小程序：

```bash
cd miniapp
pnpm install
pnpm dev:mp-weixin
```

微信开发者工具中开启本地调试时可临时勾选“不校验合法域名、web-view、TLS 版本以及 HTTPS 证书”。上线必须配置 HTTPS 合法域名。

## 8. 测试账号建议

后台默认账号：

```text
username: admin
password: Admin@123456
role: ROOT_ADMIN
```

平台管理员测试账号：

```text
username: platform01
password: Platform@123456
role: PLATFORM_ADMIN
```

普通用户测试账号：

```text
phone: 13800000001
password: User@123456
```

上线前必须强制修改默认密码。

## 9. 本地启动顺序

```text
1. docker compose up -d
2. 执行 migration
3. 启动 backend
4. 启动 admin-web
5. 启动 web
6. 启动 miniapp
7. 使用 Apifox / Postman 联调接口
```

## 10. 常见问题

### 10.1 MySQL 连接失败

检查容器、端口、用户名密码、数据库名和 `.env`。

### 10.2 Redis 连接失败

检查 Redis 容器、端口、密码和 DB 配置。

### 10.3 前端跨域

后端 Gin 需要配置 CORS，允许本地前端端口：`5173`、`5174`。

### 10.4 小程序请求本地接口失败

本地调试开启域名校验豁免；生产必须使用 HTTPS。
