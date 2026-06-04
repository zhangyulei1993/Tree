# 00-CHANGELOG-v6-GO.md

# v6 Go 后端技术栈变更记录

## 1. 变更原因

用户明确决定：

```text
后端使用 Go 语言
```

因此项目文档从 v5 的 Java / Spring Boot 方案调整为 Go 后端方案。

---

## 2. 主要变更

| 项目 | v5 | v6 |
|---|---|---|
| 后端语言 | Java 17 | Go |
| 后端框架 | Spring Boot 3.x | Gin |
| ORM | MyBatis-Plus | GORM |
| 迁移工具 | 未强制 | golang-migrate |
| Redis 客户端 | Java Redis Client | go-redis |
| 日志 | Java 日志体系 | zap |
| 参数校验 | Java Bean Validation | go-playground/validator |
| 构建 | Maven | Go Modules |

---

## 3. 保持不变

```text
数据库：MySQL 8
缓存：Redis 7
管理后台：Vue 3 + TypeScript + Vite + Element Plus
PC/H5 前台：Vue 3 + TypeScript + Vite
微信小程序：uni-app + Vue 3 + TypeScript
部署：Nginx + Docker Compose 可选
```

---

## 4. 业务规则不变

以下业务规则完全不变：

```text
user 与 family_member 分离
family_member 可以没有 user
NOT_REQUIRED 成员节点
relationship_type 只允许 PARENT_CHILD / SPOUSE
SIBLING 不落库
PLATFORM_ADMIN 不能审核创始人转让和家庭解散
操作日志 operation_logs
graph_version
账号合并
账号认领
家庭公开申请
游客留言审核
```

---

## 5. Codex 执行提醒

Codex 后续必须按 Go 后端实现，不得再生成 Java / Spring Boot 项目。
