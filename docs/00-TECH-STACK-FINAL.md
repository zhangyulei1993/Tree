# 00-TECH-STACK-FINAL.md

# Tree 家脉亲缘平台最终技术栈与开发语言说明 v2 - Go 后端版

> 本文件用于明确项目一期实际采用的技术栈、开发语言、数据库、中间件和端侧框架。  
> Codex 或开发团队必须优先遵守本文件。  
> 当前版本已经将后端技术栈正式切换为 **Go + Gin + GORM**。  
> 如果实际已有代码仓库，则以现有仓库为准；如果尚未建仓库，则按本文档启动一期开发。

---

# 1. 技术栈最终选择

## 1.1 后端 API 服务

```text
开发语言：Go
Go 版本：使用项目启动时最新稳定版，建议 Go 1.22+
Web 框架：Gin
ORM / 数据访问：GORM
数据库迁移：golang-migrate
认证方式：JWT
参数校验：go-playground/validator
日志：zap
配置：Viper 或 cleanenv
API 文档：swaggo / gin-swagger
构建方式：Go Modules
```

选择理由：

```text
1. Go 部署简单，编译成单一二进制文件，适合中小型平台快速上线。
2. Gin 路由和中间件生态成熟，适合 RESTful API。
3. GORM 对动态查询、事务、分页、软删除和后台筛选较友好。
4. Go 的并发模型适合验证码、缓存、日志、限流等服务。
5. 项目一期不需要微服务，Go 模块化单体架构更稳。
```

---

## 1.2 数据库

```text
数据库：MySQL 8.x
字符集：utf8mb4
排序规则：utf8mb4_0900_ai_ci 或 utf8mb4_general_ci
主键类型：BIGINT AUTO_INCREMENT
时间字段：DATETIME
状态字段：VARCHAR
迁移工具：golang-migrate
```

设计要求：

```text
1. 不使用 MySQL ENUM，状态统一使用 VARCHAR。
2. 所有核心表必须包含 created_at / updated_at。
3. 需要软删除的表必须包含 deleted_at。
4. 手机号、openid、family-member-user 绑定关系需要唯一性约束或 Service 层强校验。
5. 数据库迁移脚本必须纳入版本管理。
6. 所有迁移文件放在 backend/migrations 目录。
```

---

## 1.3 Redis

```text
缓存 / 中间件：Redis 7.x
Go 客户端：go-redis
```

一期用途：

```text
1. 验证码缓存。
2. token 黑名单。
3. 家庭树缓存。
4. 游客留言 IP 频率限制。
5. 简单接口限流。
```

缓存 key 建议：

```text
verify_code:{scene}:{phoneHash}
token:blacklist:{tokenId}
family:{familyId}:tree:v{graphVersion}
rate_limit:visitor_message:{ip}:{minute}
rate_limit:visitor_message:{ip}:{date}
```

---

## 1.4 管理后台

```text
开发语言：TypeScript
前端框架：Vue 3
构建工具：Vite
UI 框架：Element Plus
状态管理：Pinia
路由：Vue Router
HTTP 客户端：Axios
```

选择理由：

```text
1. Element Plus 适合后台表格、筛选、表单、弹窗、审核页面。
2. Vue 3 + TypeScript 便于维护权限按钮和页面状态。
3. 后台需要大量列表、详情、审核、筛选页面，Element Plus 开发效率较高。
```

---

## 1.5 PC / H5 前台网站

```text
开发语言：TypeScript
前端框架：Vue 3
构建工具：Vite
状态管理：Pinia
路由：Vue Router
HTTP 客户端：Axios
UI：自定义组件 + 移动端适配
```

用途：

```text
1. 家庭搜索。
2. 公开家庭主页。
3. 公开家庭树展示。
4. 用户注册 / 登录。
5. 我的家庭。
6. 邀请确认。
7. 加入申请。
8. 个人中心。
```

---

## 1.6 微信小程序

当前推荐：

```text
开发语言：TypeScript
框架：uni-app
语法：Vue 3
```

选择理由：

```text
1. 和 PC / H5 / 后台共用 Vue 技术栈。
2. 团队维护成本更低。
3. 后续如需 H5、小程序多端复用更方便。
```

如果团队更熟悉原生微信小程序，也可以采用：

```text
微信小程序原生 + TypeScript
```

但当前文档默认：

```text
uni-app + Vue 3 + TypeScript
```

---

## 1.7 文件 / 图片存储

一期建议先做抽象接口：

```text
FileStorageService
```

默认实现：

```text
本地存储
```

后续可替换为：

```text
阿里云 OSS
腾讯云 COS
七牛云
```

涉及字段：

```text
users.avatar_url
families.avatar_url
家庭封面图
用户头像
成员头像，后续预留
```

---

## 1.8 部署与运行环境

建议：

```text
后端运行环境：Go 编译后的二进制程序
数据库：MySQL 8
缓存：Redis 7
反向代理：Nginx
部署方式：Docker Compose 或传统服务器部署
```

开发环境建议：

```text
Docker Compose 启动 MySQL + Redis
backend 本地 go run ./cmd/server 启动
admin-web / web / miniapp 本地 Vite / uni-app 启动
```

---

# 2. 项目代码仓库建议结构

```text
Tree
├── docs
├── backend
├── admin-web
├── web
├── miniapp
├── docker
└── README.md
```

说明：

| 目录 | 说明 |
|---|---|
| `docs` | 项目文档 |
| `backend` | Go + Gin + GORM 后端 |
| `admin-web` | Vue 3 管理后台 |
| `web` | PC / H5 前台 |
| `miniapp` | uni-app 微信小程序 |
| `docker` | Docker Compose、Nginx、MySQL 初始化脚本 |

---

# 3. Go 后端目录结构建议

```text
backend
├── cmd
│   └── server
│       └── main.go
├── configs
│   ├── config.dev.yaml
│   ├── config.prod.yaml
│   └── config.example.yaml
├── deployments
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── nginx.conf
├── migrations
│   ├── 000001_create_users.up.sql
│   ├── 000001_create_users.down.sql
│   ├── 000002_create_families.up.sql
│   └── ...
├── internal
│   ├── app
│   │   ├── server.go
│   │   └── router.go
│   ├── common
│   │   ├── response
│   │   ├── errors
│   │   ├── enums
│   │   ├── middleware
│   │   ├── validator
│   │   ├── jwt
│   │   ├── redis
│   │   ├── logger
│   │   ├── transaction
│   │   └── config
│   ├── auth
│   ├── user
│   ├── account
│   ├── family
│   │   ├── core
│   │   ├── member
│   │   ├── relationship
│   │   ├── tree
│   │   ├── invitation
│   │   ├── joinrequest
│   │   ├── publicdisplay
│   │   ├── message
│   │   ├── role
│   │   ├── transfer
│   │   └── dissolution
│   ├── admin
│   └── operationlog
├── pkg
│   └── utils
├── scripts
├── docs
├── go.mod
├── go.sum
└── README.md
```

---

# 4. Go 后端分层规范

每个业务模块建议采用：

```text
handler
service
repository
model
dto
vo
enum
```

说明：

| 层 | 职责 |
|---|---|
| `handler` | HTTP 入参、参数绑定、调用 service、返回响应 |
| `service` | 业务规则、权限校验、事务、日志、graph_version |
| `repository` | 数据库查询与持久化 |
| `model` | GORM 数据模型 |
| `dto` | 请求结构 |
| `vo` | 返回结构 |
| `enum` | 状态枚举与常量 |

禁止：

```text
1. 把复杂业务写在 handler。
2. 在 handler 中直接操作多个 repository 拼业务流程。
3. 跳过 service 层权限校验。
4. 跳过 operation_logs。
5. 跳过 graph_version。
6. 把所有模块都堆在一个 service 文件里。
```

---

# 5. 必须单独拆出的 Go 服务

以下服务必须单独存在，不能散落在 handler 中：

```text
AuthService
TokenService
VerificationCodeService
AccountMergeService
AccountClaimService
FamilyPermissionService
FamilyMemberService
FamilyRelationshipService
FamilyTreeService
FamilyInvitationService
FamilyJoinRequestService
PublicDisplayService
VisitorMessageService
FamilyRoleService
FounderTransferService
FamilyDissolutionService
OperationLogService
GraphVersionService
TransactionManager
```

尤其重要：

```text
FamilyPermissionService
GraphVersionService
OperationLogService
AccountMergeService
FamilyRelationshipService
```

---

# 6. 管理后台页面

必须支持：

```text
1. 登录页。
2. 用户管理。
3. 家庭管理。
4. 成员管理。
5. 公开申请审核。
6. 游客留言审核。
7. 创始人转让审核。
8. 家庭解散审核。
9. 后台管理员管理。
10. 操作日志。
```

---

# 7. PC / H5 前台页面

必须支持：

```text
1. 首页。
2. 家庭搜索。
3. 公开家庭主页。
4. 公开家庭树。
5. 登录 / 注册。
6. 我的家庭。
7. 邀请详情。
8. 加入申请。
9. 个人中心。
```

---

# 8. 微信小程序页面

必须支持：

```text
1. 小程序快捷登录。
2. 绑定手机号。
3. 家庭搜索。
4. 公开家庭主页。
5. 公开家庭树。
6. 我的家庭。
7. 邀请确认。
8. 加入家庭申请。
9. 个人中心。
```

---

# 9. 必须统一的接口规范

## 9.1 统一成功返回

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

## 9.2 统一失败返回

```json
{
  "code": 40001,
  "message": "验证码错误或已过期",
  "data": null
}
```

## 9.3 统一分页返回

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "records": [],
    "page": 1,
    "pageSize": 20,
    "total": 100
  }
}
```

Go 结构建议：

```go
type ApiResponse[T any] struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    T      `json:"data"`
}

type PageResponse[T any] struct {
    Records  []T   `json:"records"`
    Page     int   `json:"page"`
    PageSize int   `json:"pageSize"`
    Total    int64 `json:"total"`
}
```

---

# 10. 必须统一的认证规范

## 10.1 普通用户 token

用于：

```text
PC / H5
微信小程序
```

## 10.2 后台管理员 token

用于：

```text
管理后台
```

## 10.3 禁止混用

```text
普通用户 token 不能访问 /api/admin/*
后台管理员 token 不能作为普通家庭成员身份访问家庭端接口
```

---

# 11. 开发语言汇总

| 端 | 开发语言 | 框架 |
|---|---|---|
| 后端 API | Go | Gin + GORM |
| 管理后台 | TypeScript | Vue 3 + Vite + Element Plus |
| PC / H5 前台 | TypeScript | Vue 3 + Vite |
| 微信小程序 | TypeScript | uni-app + Vue 3 |
| 数据库脚本 | SQL | MySQL 8 |
| 缓存 | Redis 命令 / Go Redis Client | Redis 7 |
| 部署 | Shell / YAML | Docker Compose / Nginx |

---

# 12. Codex 必须遵守

Codex 不得自行改用：

```text
Java / Spring Boot
Node.js / NestJS
Python / FastAPI
React
PostgreSQL
MongoDB
```

除非用户明确要求变更技术栈。

当前默认技术栈已经锁定为：

```text
Go + Gin + GORM + MySQL + Redis + Vue 3 + TypeScript + uni-app
```

---

# 13. 后续可选升级

一期后可考虑：

```text
1. OpenAPI / Swagger 文档自动生成。
2. Redis 缓存家庭树。
3. 操作日志异步写入队列。
4. 对象存储替换本地文件。
5. CI/CD 自动部署。
6. 权限按钮动态配置。
7. 小程序订阅消息。
8. 家庭树可视化图谱。
9. 在极复杂查询中局部引入 sqlc。
10. 后台权限可选引入 Casbin。
```
