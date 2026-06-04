# TECH_STACK_DECISION.md

# Tree 家脉亲缘平台技术栈决策文档 v2 - Go 后端版

> 本文档用于指导 Codex 或开发团队在正式编码前统一技术栈。  
> 当前版本已经明确：后端使用 **Go + Gin + GORM**。  
> 如果已有代码仓库，请以现有仓库技术栈为准；如果尚未建仓库，按本文档启动一期开发。

---

# 1. 项目端划分

项目包含以下端：

```text
1. 后端 API 服务
2. 管理后台
3. PC / H5 前台网站
4. 微信小程序
5. 数据库
6. Redis 缓存
7. 文件 / 图片存储
```

---

# 2. 最终推荐技术栈

## 2.1 后端

```text
Go
Gin
GORM
golang-migrate
go-playground/validator
go-redis
zap
JWT
swaggo/gin-swagger
```

理由：

```text
适合模块化单体
部署简单
性能足够
业务逻辑可控
开发效率高
生态成熟
```

---

## 2.2 数据库

```text
MySQL 8
```

要求：

```text
utf8mb4
BIGINT AUTO_INCREMENT
DATETIME
VARCHAR 状态字段
migration 管理 schema
```

---

## 2.3 Redis

```text
Redis 7
go-redis
```

用途：

```text
验证码
token 黑名单
家庭树缓存
游客留言限流
接口限流
```

---

## 2.4 管理后台

```text
Vue 3
TypeScript
Vite
Element Plus
Pinia
Axios
Vue Router
```

---

## 2.5 PC / H5 前台

```text
Vue 3
TypeScript
Vite
Pinia
Axios
Vue Router
```

---

## 2.6 微信小程序

默认：

```text
uni-app
Vue 3
TypeScript
```

可选：

```text
原生微信小程序 + TypeScript
```

---

# 3. 当前锁定技术栈

```text
后端：Go + Gin + GORM
数据库：MySQL 8
缓存：Redis 7
后台：Vue 3 + TypeScript + Vite + Element Plus
前台：Vue 3 + TypeScript + Vite
小程序：uni-app + Vue 3 + TypeScript
部署：Docker Compose + Nginx
```

---

# 4. 为什么不用 Java / Spring Boot

当前用户已明确计划：

```text
后端用 Go 语言
```

因此文档不再推荐 Java / Spring Boot 作为默认实现。

除非用户后续明确变更，否则 Codex 不应生成 Java 项目。

---

# 5. 为什么不用 Node.js / NestJS

虽然 Node.js / NestJS 适合 TypeScript 全栈，但当前项目后端已决定采用 Go。

Codex 不应自行改成：

```text
Node.js
NestJS
Express
Fastify
```

---

# 6. 为什么不用微服务

当前一期不建议微服务。

原因：

```text
1. 业务规则仍在快速演化。
2. 家庭、成员、关系、账号、权限之间事务强相关。
3. 微服务会增加部署、事务、联调复杂度。
4. 当前阶段模块化单体更稳。
```

采用：

```text
模块化单体
```

---

# 7. Go 项目架构原则

```text
handler 只处理 HTTP
service 写业务规则
repository 写数据库查询
model 定义 GORM 模型
dto 定义请求结构
vo 定义返回结构
middleware 处理认证权限
operation log service 统一写日志
graph version service 统一更新版本
```

---

# 8. Codex 执行约束

Codex 开发时必须遵守：

```text
1. 不要擅自更换技术栈。
2. 不要生成 Java / Spring Boot 项目。
3. 不要生成 Node.js / NestJS 项目。
4. 不要擅自合并 user 和 family_member。
5. 不要删除 NOT_REQUIRED 成员节点设计。
6. 不要把 SIBLING 保存为 relationship_type。
7. 不要让 PLATFORM_ADMIN 审核创始人转让或家庭解散。
8. 不要跳过 operation_logs。
9. 不要跳过 graph_version。
10. 不要把普通用户 token 和后台管理员 token 混用。
11. 不要把验证码、密码、token 明文写入日志。
12. 不要一次性实现所有模块，应按 CODEX_TASKS.md 分阶段执行。
```

---

# 9. 待确认项

正式开工前还需要确认：

```text
1. Go 具体版本。
2. 配置库使用 Viper 还是 cleanenv。
3. 是否一期启用 Redis。
4. 文件存储一期使用本地还是云 OSS。
5. 是否需要 Docker Compose 本地开发环境。
6. 是否需要 Swagger 自动生成 OpenAPI 文档。
7. 后台权限是否一期使用 Casbin，还是自定义 RBAC。
```

默认建议：

```text
配置库：Viper
Redis：一期启用
文件存储：先本地，抽象接口
Docker Compose：需要
Swagger：需要
后台权限：一期自定义 RBAC，后续可引入 Casbin
```
