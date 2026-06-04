# 26-CODEX-FULL-EXECUTION-GUIDE.md

# Codex 全阶段执行总手册 v1

> 本文件用于指导 Codex 从项目初始化到上线准备的完整执行流程。  
> 注意：这是完整路线图，不代表要一次性让 Codex 实现全部功能。  
> 正确方式是：每次只执行一个阶段，每个阶段完成后验收，再进入下一阶段。

---

# 1. 总体原则

## 1.1 不要一次性执行全部项目

禁止直接对 Codex 说：

```text
请根据文档完成整个项目。
```

原因：

```text
1. 容易漏业务规则。
2. 容易跳过测试。
3. 容易生成不可维护代码。
4. 容易把复杂业务写进 handler。
5. 容易提前实现后续阶段导致结构混乱。
```

正确方式：

```text
M1 -> 验收 -> M2 -> 验收 -> M3 -> 验收 ...
```

---

# 2. 推荐执行顺序

```text
M1  项目骨架与本地环境
M2  数据库 migration 与基础 GORM model
M3  统一响应、错误码、JWT、中间件、日志基础设施
M4  后台管理员认证与后台权限骨架
M5  普通用户认证、手机号注册登录、验证码
M6  微信小程序登录、手机号绑定、账号认领、账号合并
M7  家庭基础模块
M8  家庭成员模块
M9  家庭关系模块
M10 家庭树组装与 graph_version
M11 邀请与加入家庭模块
M12 家庭公开申请与游客留言模块
M13 家庭角色、创始人转让、家庭解散
M14 管理后台 UI 原型与页面实现
M15 PC/H5 前台与微信小程序页面实现
M16 测试补齐、安全检查、部署上线准备
```

---

# 3. 每阶段固定工作法

每个阶段都要求 Codex 输出：

```text
1. 已完成内容
2. 修改文件列表
3. 新增接口
4. 新增数据库表 / 字段
5. 已运行命令
6. 测试结果
7. 未完成事项
8. 风险点
9. 下一步建议
```

每阶段都要求：

```text
1. 先读 AGENTS.md。
2. 只读本阶段相关文档。
3. 只修改允许目录。
4. 不提前实现下一阶段。
5. 完成后运行测试。
```

---

# 4. 分支建议

建议每阶段一个分支：

```bash
git checkout -b m1-project-skeleton
git checkout -b m2-database-migration
git checkout -b m3-infra-auth-base
git checkout -b m4-admin-auth
git checkout -b m5-user-auth
git checkout -b m6-wechat-account-merge
git checkout -b m7-family-core
git checkout -b m8-family-member
git checkout -b m9-family-relationship
git checkout -b m10-family-tree
git checkout -b m11-invitation-join
git checkout -b m12-public-message
git checkout -b m13-role-transfer-dissolution
git checkout -b m14-admin-web
git checkout -b m15-web-miniapp
git checkout -b m16-test-release
```

---

# 5. 最重要的业务红线

任何阶段都不能违反：

```text
1. user 和 family_member 不能合并。
2. family_member 可以没有 user。
3. NOT_REQUIRED 成员必须支持。
4. relationship_type 只能是 PARENT_CHILD / SPOUSE。
5. SIBLING 不落库。
6. ADD_SIBLING 必须通过共同父母推导。
7. PLATFORM_ADMIN 不能审核创始人转让。
8. PLATFORM_ADMIN 不能审核家庭解散。
9. PLATFORM_ADMIN 不能生成分享邀请链接。
10. 操作日志 operation_logs 必须写。
11. 家庭树结构变化必须更新 graph_version。
12. 普通 user token 和 admin token 必须分离。
13. 验证码、密码、token、session_key、AppSecret 不得明文记录。
```

---

# 6. 进入下一阶段的条件

每阶段进入下一阶段前必须满足：

```text
1. 当前阶段验收清单通过。
2. 当前阶段代码能编译。
3. 关键测试通过。
4. 没有明显违反 AGENTS.md。
5. 没有把后续阶段提前做乱。
6. 修改文件列表清楚。
```

---

# 7. 推荐 Codex 使用方式

## 7.1 每次复制单阶段 prompt

使用：

```text
27-CODEX-STAGE-PROMPTS-M1-M16.md
```

每次只复制一个阶段的 prompt。

## 7.2 每阶段完成后验收

使用：

```text
28-CODEX-STAGE-ACCEPTANCE-CHECKLISTS.md
```

## 7.3 PR 前审查

使用：

```text
29-CODEX-PR-REVIEW-PROMPTS.md
```

---

# 8. 不建议的方式

不要：

```text
1. 一次性让 Codex 完成所有后端。
2. 一次性让 Codex 完成所有前端。
3. 让 Codex 自己决定技术栈。
4. 让 Codex 自己改业务规则。
5. 没有测试就进入下一阶段。
6. 不看 diff 就合并代码。
```

---

# 9. 当前最推荐下一步

执行：

```text
M1：项目骨架与本地环境
```

使用：

```text
CODEX_M1_START_PROMPT.md
```

M1 通过后再执行：

```text
M2：数据库 migration 与基础 GORM model
```
