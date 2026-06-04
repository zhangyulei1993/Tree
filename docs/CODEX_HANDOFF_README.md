# CODEX_HANDOFF_README.md

# Tree 家脉亲缘平台 Codex 交接说明 v1

> 本文件用于说明如何把当前项目文档交给 Codex 或开发团队继续实现。

---

# 1. 当前文档状态

当前已经形成：

```text
tree-docs-api-tests-v3.zip
```

该文档包已经覆盖：

```text
项目一期功能边界
核心业务规则
数据库核心表草案
权限矩阵
认证与账号接口
家庭基础接口
成员与关系接口
邀请与加入家庭接口
公开申请与留言接口
家庭角色 / 创始人转让 / 解散接口
后台管理接口
操作日志接口
测试用例与验收标准
```

本次新增 Codex 交接文件：

```text
TECH_STACK_DECISION.md
CODEX_TASKS.md
CODEX_HANDOFF_README.md
```

---

# 2. 推荐放置目录

在代码仓库中建议这样放：

```text
project-root
├── docs
│   ├── tree-docs-api-tests-v3
│   ├── TECH_STACK_DECISION.md
│   ├── CODEX_TASKS.md
│   └── CODEX_HANDOFF_README.md
├── backend
├── admin-web
├── web
└── miniapp
```

如果还没有仓库，可以先创建：

```text
Tree
├── docs
├── backend
├── admin-web
├── web
└── miniapp
```

---

# 3. 给 Codex 的第一条提示词

建议第一条不要让 Codex 直接完成全部项目，而是这样写：

```text
请先阅读 docs 目录下的所有 Markdown 文档，尤其是：
1. TECH_STACK_DECISION.md
2. CODEX_TASKS.md
3. tree-docs-api-tests-v3 中的全部文档

当前任务只执行 CODEX_TASKS.md 的阶段一：项目骨架搭建。

要求：
- 不要实现具体业务逻辑。
- 不要擅自更换技术栈。
- 不要合并 user 和 family_member。
- 不要删除 NOT_REQUIRED 成员节点设计。
- 不要把 SIBLING 保存为 relationship_type。
- 建立统一响应结构、错误码结构、权限枚举、路由分组、operation_logs 服务接口、graph_version 服务接口。
- 完成后列出新增文件、未实现事项、下一阶段建议。
```

---

# 4. 后续给 Codex 的节奏

建议每次只执行一个阶段：

```text
阶段一：项目骨架
阶段二：数据库 Schema
阶段三：认证与账号
阶段四：家庭基础
阶段五：成员与关系
阶段六：邀请与加入
阶段七：公开申请与留言
阶段八：角色、转让、解散
阶段九：后台管理与日志
阶段十：测试与修复
```

不要一次性发：

```text
请完成整个项目
```

这样容易导致业务规则丢失、代码混乱、测试缺失。

---

# 5. Codex 修改代码后的检查方式

每次 Codex 完成后，你应该要求它输出：

```text
1. 修改了哪些文件
2. 新增了哪些接口
3. 新增了哪些表
4. 哪些规则已经实现
5. 哪些规则还没实现
6. 运行了哪些测试
7. 哪些测试失败
8. 下一步建议
```

---

# 6. 最重要的业务红线

以下规则不能被 Codex 改掉：

```text
1. family_member 与 user 必须分离。
2. family_member 可以没有 user。
3. 无需绑定用户的家庭成员必须支持。
4. 一个 user 在同一 family 只能绑定一个 ACTIVE member。
5. 一个 member 只能被一个 ACTIVE user 绑定。
6. 同一 family 只能有一个 FOUNDER。
7. PLATFORM_ADMIN 不能审核创始人转让。
8. PLATFORM_ADMIN 不能审核家庭解散。
9. PLATFORM_ADMIN 不能生成分享邀请链接。
10. relationship_type 不能出现 SIBLING。
11. 添加兄弟姐妹必须通过共同父母转化为 PARENT_CHILD。
12. 删除存在下级成员的节点必须失败。
13. 高风险操作必须写 operation_logs。
14. 家庭树结构变化必须更新 graph_version。
15. 账号合并后旧 token 必须失效。
```

---

# 7. 人工确认节点

以下内容 Codex 实现前最好人工确认：

```text
1. 最终后端技术栈。
2. 小程序采用原生还是 uni-app。
3. 是否一期启用 Redis。
4. 对象存储方案。
5. 后台 UI 框架。
6. 数据库唯一约束实现方式。
7. 账号合并冲突选择页面。
8. 家庭树展示页面 UI。
9. 管理后台菜单结构。
```

---

# 8. 当前建议

当前最稳妥的方式是：

```text
先让 Codex 做阶段一和阶段二。
你检查项目结构和数据库。
确认无误后，再让 Codex 进入认证模块。
```

不要直接进入家庭关系模块。

家庭关系模块是本项目最复杂、最容易出错的部分，必须在账号、数据库、权限、日志基础稳定后再做。
---

# 技术栈强制说明

v5 已新增：

```text
00-TECH-STACK-FINAL.md
01-system-architecture-and-tech-stack.md
```

Codex 第一优先级应读取这两个文件。当前默认技术栈是：

```text
后端：Java 17 + Gin 3.x
数据库：MySQL 8
缓存：Redis 7
管理后台：Vue 3 + TypeScript + Element Plus
前台网站：Vue 3 + TypeScript
微信小程序：uni-app + Vue 3 + TypeScript
```

不要让 Codex 自行选择 Node.js、Python、Go、React 或 PostgreSQL，除非用户明确变更技术栈。
---

# v6 Go 后端强制说明

当前版本已明确切换为 Go 后端。

Codex 第一优先级应读取：

```text
00-TECH-STACK-FINAL.md
01-system-architecture-and-tech-stack.md
02-GO-BACKEND-ARCHITECTURE.md
TECH_STACK_DECISION.md
CODEX_TASKS.md
```

当前默认技术栈是：

```text
后端：Go + Gin + GORM
数据库：MySQL 8
缓存：Redis 7
管理后台：Vue 3 + TypeScript + Element Plus
前台网站：Vue 3 + TypeScript
微信小程序：uni-app + Vue 3 + TypeScript
```

不要让 Codex 自行选择 Java、Spring Boot、Node.js、Python、React、PostgreSQL 或 MongoDB，除非用户明确变更技术栈。

---

# v7 开发落地文件说明

Codex 在正式执行阶段一前，除技术栈文档外，还必须阅读：

```text
12-local-dev-setup.md
13-ui-page-spec.md
14-database-migration-ddl.md
15-error-code-summary.md
16-development-milestones.md
17-api-debug-guide.md
```

阶段一优先实现：Go 后端骨架、Docker Compose、.env.example、健康检查接口、统一响应、错误码结构、日志结构、migration 目录。
