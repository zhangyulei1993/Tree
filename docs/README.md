# Tree 家脉亲缘平台文档包 API v2

本包基于 `tree-docs-current-v1` 整理，并新增截至当前讨论已经形成的 API 完整规格文档。

## 新增重点

```text
04-api-overview-v2.md
04-01-auth-account-api.md
04-02-family-basic-api.md
04-03-family-member-relationship-api.md
04-04-invitation-join-api.md
04-05-public-application-message-api.md
04-06-family-role-transfer-dissolution-api.md
04-07-admin-management-api.md
04-08-operation-logs-api.md
```

## 当前状态

| 模块 | 状态 |
|---|---|
| 认证与账号接口 | 已细化到开发规格 v1 |
| 家庭基础接口 | 已细化到开发规格 v1 |
| 家庭成员与关系接口 | 已细化到开发规格 v1 |
| 邀请与加入家庭接口 | 已细化到开发规格 v1 |
| 家庭公开申请与游客留言接口 | 已细化到开发规格 v1 |
| 家庭角色 / 创始人转让 / 解散接口 | 已细化到开发规格 v1 |
| 后台管理接口 | 已细化到开发规格 v1 |
| 操作日志接口 | 已细化到开发规格 v1 |
| 测试用例与验收标准 | 见增强包 v3 |

## v3 增强内容

新增：`11-test-cases-and-acceptance.md`，覆盖一期测试用例、验收标准、上线前检查清单。


---

# v4 Codex Ready 增强说明

本版本在 v3 文档包基础上新增：

```text
TECH_STACK_DECISION.md
CODEX_TASKS.md
CODEX_HANDOFF_README.md
```

建议把本包直接放入项目仓库 `docs/` 目录，之后让 Codex 严格按 `CODEX_TASKS.md` 分阶段实现。
---

# v5 技术栈补充说明

本版本新增并明确固化项目一期技术栈与开发语言：

```text
00-TECH-STACK-FINAL.md
01-system-architecture-and-tech-stack.md
```

当前默认技术栈锁定为：

```text
后端：Go + Gin + GORM
数据库：MySQL 8
缓存：Redis 7
管理后台：Vue 3 + TypeScript + Vite + Element Plus
PC/H5 前台：Vue 3 + TypeScript + Vite
微信小程序：uni-app + Vue 3 + TypeScript
部署：Nginx + Docker Compose 可选
```

Codex 或开发团队必须优先阅读：

```text
00-TECH-STACK-FINAL.md
01-system-architecture-and-tech-stack.md
TECH_STACK_DECISION.md
CODEX_TASKS.md
CODEX_HANDOFF_README.md
```
---

# v6 Go 后端技术栈更新说明

本版本将后端技术栈从 Java / Spring Boot 调整为 Go 技术栈，并新增：

```text
02-GO-BACKEND-ARCHITECTURE.md
```

当前默认技术栈锁定为：

```text
后端：Go + Gin + GORM
数据库：MySQL 8
缓存：Redis 7 + go-redis
数据库迁移：golang-migrate
参数校验：go-playground/validator
日志：zap
认证：JWT
API 文档：swaggo / gin-swagger
管理后台：Vue 3 + TypeScript + Vite + Element Plus
PC/H5 前台：Vue 3 + TypeScript + Vite
微信小程序：uni-app + Vue 3 + TypeScript
部署：Nginx + Docker Compose 可选
```

Codex 或开发团队必须优先阅读：

```text
00-TECH-STACK-FINAL.md
01-system-architecture-and-tech-stack.md
02-GO-BACKEND-ARCHITECTURE.md
TECH_STACK_DECISION.md
CODEX_TASKS.md
CODEX_HANDOFF_README.md
```

---

# v7 开发落地补充说明

本版本在 v6 Go 后端版基础上新增：

```text
12-local-dev-setup.md
13-ui-page-spec.md
14-database-migration-ddl.md
15-error-code-summary.md
16-development-milestones.md
17-api-debug-guide.md
```

新增内容覆盖本地开发环境、Docker Compose、UI 页面规格、数据库 DDL、统一错误码、开发里程碑和 API 联调指南。当前 v7 包已经可以作为 Codex / 开发团队正式开工的资料基础。

---

# v8 上线运维与小程序发布补充说明

本版本在 v7 开发落地版基础上新增：

```text
18-design-language.md
19-core-wireframes.md
20-deployment-guide.md
21-mini-program-launch-checklist.md
22-security-and-privacy-checklist.md
23-backup-and-monitoring-guide.md
24-production-readiness-checklist.md
25-agents-md-template.md
```

新增内容覆盖设计语言、核心线框图、服务器部署、HTTPS、小程序发布、隐私安全、备份监控、生产上线检查和 AGENTS.md 模板。

---

# v9 M1 启动补充说明

本版本在 v8 上线运维版基础上新增：

```text
AGENTS.md
CODEX_M1_START_PROMPT.md
M1_ACCEPTANCE_CHECKLIST.md
```

用途：

```text
AGENTS.md：放到项目根目录，用于约束 Codex 的技术栈、业务红线、目录规范和测试要求。
CODEX_M1_START_PROMPT.md：直接复制给 Codex，启动 M1 项目骨架与本地环境任务。
M1_ACCEPTANCE_CHECKLIST.md：用于检查 Codex 完成 M1 后是否可以进入 M2。
```

建议使用方式：

```text
1. 将本包中的 docs 文档放到项目根目录 /docs。
2. 将 AGENTS.md 放到项目根目录。
3. 将 CODEX_M1_START_PROMPT.md 的内容复制给 Codex。
4. 按 M1_ACCEPTANCE_CHECKLIST.md 验收结果。
```


---

# v10 全阶段执行补充说明

本版本在 v9 M1 启动版基础上新增：

```text
26-CODEX-FULL-EXECUTION-GUIDE.md
27-CODEX-STAGE-PROMPTS-M1-M16.md
28-CODEX-STAGE-ACCEPTANCE-CHECKLISTS.md
29-CODEX-PR-REVIEW-PROMPTS.md
30-PROJECT-REPO-SETUP-GUIDE.md
```

新增内容覆盖：

```text
从 M1 到 M16 的完整 Codex 执行路线
每阶段可直接复制的 Codex 提示词
每阶段验收清单
PR Review 提示词
项目仓库初始化指南
```

建议：不要一次性让 Codex 执行全部阶段。应每次只执行一个阶段，验收通过后再继续。

---

# v11 项目命名统一说明

本版本将项目正式名称统一为：

```text
Tree
```

并新增：

```text
00-PROJECT-NAMING-CONVENTION.md
```

命名约定：

```text
项目名称：Tree
本地目录：~/Projects/Tree
后端服务：tree-api
管理后台：tree-admin-web
PC/H5 前台：tree-web
微信小程序：tree-miniapp
数据库名：tree_platform
MySQL 容器：tree-mysql
Redis 容器：tree-redis
```

注意：

```text
family / families / family_member / family_relationships 等仍然是业务领域命名，不应替换为 tree。
```
