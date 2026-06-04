# CODEX_TASKS.md

> 重要更新：当前后端技术栈已切换为 Go。  
> 在执行任何阶段前，必须先阅读 `00-TECH-STACK-FINAL.md`、`01-system-architecture-and-tech-stack.md` 和 `02-GO-BACKEND-ARCHITECTURE.md`。  
> 当前默认技术栈已锁定为 **Go + Gin + GORM + MySQL + Redis + Vue 3 + TypeScript + uni-app**。不得自行更换技术栈。

> 重要更新：在执行任何阶段前，必须先阅读 `00-TECH-STACK-FINAL.md` 和 `01-system-architecture-and-tech-stack.md`。当前默认技术栈已锁定为 Go + Gin + GORM + MySQL + Redis + Vue 3 + TypeScript + uni-app。不得自行更换技术栈。

# Codex 分阶段开发任务清单 v1

> 本文件用于指导 Codex 按阶段实现项目。  
> 不要一次性要求 Codex “完成整个项目”。  
> 每个阶段必须先阅读 docs 文档，再完成明确范围内的任务。

---

# 0. 使用方式

将以下文档放入项目仓库：

```text
docs/tree-docs-api-tests-v3/
docs/TECH_STACK_DECISION.md
docs/CODEX_TASKS.md
docs/CODEX_HANDOFF_README.md
```

每次给 Codex 的任务都应包含：

```text
1. 当前阶段编号
2. 本阶段允许修改的目录
3. 本阶段必须完成的功能
4. 本阶段不能做的事情
5. 本阶段完成后的检查清单
```

---

# 1. 全局业务约束

Codex 在任何阶段都必须遵守：

```text
1. user 是系统账号，family_member 是家庭树节点，两者不能合并。
2. family_member 可以没有 user。
3. family_member 支持 NOT_REQUIRED 用户绑定策略。
4. relationship_type 当前只允许 PARENT_CHILD 和 SPOUSE。
5. 兄弟姐妹关系不保存 SIBLING，由共同父母推导。
6. 当前一期家庭树是列表树，不是复杂关系图谱。
7. 家庭角色存放在 family_member_user_links.family_role。
8. 家庭角色只有 FOUNDER / FAMILY_ADMIN / MEMBER。
9. ROOT_ADMIN 有且只有一个，username = admin。
10. PLATFORM_ADMIN 不能审核创始人转让和家庭解散。
11. PLATFORM_ADMIN 不能生成分享邀请链接。
12. 创始人转让和家庭解散必须由 ROOT_ADMIN / SUPER_ADMIN 审核。
13. 后台敏感操作必须写 operation_logs。
14. 影响家庭树结构的操作必须更新 graph_version。
15. 验证码只保存 hash。
16. 密码只保存 hash。
17. token、验证码、session_key 不能写入日志。
```

---

# 2. 阶段一：项目骨架

## 2.1 目标

只搭建项目骨架，不实现复杂业务。

## 2.2 允许做

```text
创建后端项目结构
创建前端 / 后台项目结构
创建统一返回结构
创建错误码结构
创建角色枚举
创建状态枚举
创建权限中间件骨架
创建数据库 migration 目录
创建 operation_logs 服务接口
创建 graph_version 服务接口
创建 API 路由分组骨架
创建 README 启动说明
```

## 2.3 不允许做

```text
不要实现家庭关系推导
不要实现账号合并细节
不要实现完整后台页面
不要写临时绕过权限的代码
不要把所有代码写在一个文件里
```

## 2.4 Codex 提示词

```text
请阅读 docs 下的全部 Markdown 文档。

当前只执行阶段一：项目骨架搭建。

要求：
1. 按 TECH_STACK_DECISION.md 的推荐技术栈创建项目结构。
2. 建立统一响应结构 ApiResponse。
3. 建立分页响应结构 PageResponse。
4. 建立错误码枚举结构。
5. 建立普通用户 token 与后台管理员 token 的分离认证骨架。
6. 建立角色枚举 ROOT_ADMIN / SUPER_ADMIN / PLATFORM_ADMIN / FOUNDER / FAMILY_ADMIN / MEMBER。
7. 建立 operation_logs 写入服务接口，但可暂时只做空实现。
8. 建立 graph_version 更新服务接口，但可暂时只做空实现。
9. 建立 Controller 路由分组骨架，不实现具体业务。
10. 完成后列出新增文件、未实现功能、下一阶段建议。

不要实现具体家庭关系业务逻辑。
```

## 2.5 验收

```text
项目能启动
基础健康检查接口可访问
统一返回结构可用
目录结构清晰
没有业务规则被硬编码到 Controller
```

---

# 3. 阶段二：数据库 Schema 与基础实体

## 3.1 目标

根据文档创建数据库迁移、实体、枚举、索引和关键约束。

## 3.2 必须覆盖表

```text
users
user_auth_identities
user_phone_history
verification_codes
user_account_merge_logs
user_account_claim_logs
families
family_members
family_relationships
family_member_user_links
family_invitations
family_join_requests
family_public_applications
visitor_messages
admin_users
operation_logs
family_founder_transfer_requests
family_dissolution_requests
```

## 3.3 关键约束

必须实现或预留：

```text
users.phone 当前有效手机号唯一
一个 openid 同一 provider_app_id 下只能对应一个 ACTIVE identity
同一 family 一个 user 只能有一个 ACTIVE member link
同一 member 只能有一个 ACTIVE user link
同一 family 同一时间只能有一个 FOUNDER
同一 member 同一时间只能有一个 PENDING invitation
同一 family 同时只能有一个 PENDING public application
同一 family 同时只能有一个 PENDING founder transfer request
同一 family 同时只能有一个 PENDING dissolution request
```

如果数据库无法直接用唯一索引表达带状态的约束，则必须在 Service 层实现。

## 3.4 Codex 提示词

```text
当前执行阶段二：数据库 Schema 与基础实体。

请根据 docs/02-database-schema.md、04-xx API 文档和 TECH_STACK_DECISION.md：
1. 创建数据库 migration。
2. 创建实体类 / Model。
3. 创建枚举。
4. 创建 Mapper / Repository。
5. 添加必要索引。
6. 对无法用数据库约束表达的规则，在注释和 Service TODO 中明确标出。
7. 不实现业务接口，只完成 schema 和基础数据访问层。

完成后请输出：
- 新增表清单
- 关键索引清单
- 未能用数据库约束实现、需要 Service 校验的规则清单
```

## 3.5 验收

```text
migration 可执行
表结构符合文档
核心状态枚举完整
基础实体可编译
```

---

# 4. 阶段三：认证与账号模块

## 4.1 目标

实现认证与账号接口。

## 4.2 接口范围

```http
POST /api/auth/send-code
POST /api/auth/register-phone
POST /api/auth/login-phone
POST /api/auth/logout
POST /api/auth/wechat-mini/login
POST /api/auth/wechat-mini/bind-phone
POST /api/auth/change-phone
POST /api/auth/cancel-account
```

## 4.3 必须实现

```text
验证码 5 分钟有效
验证码只保存 hash
手机号注册
手机号密码登录
小程序临时用户 PENDING_PHONE_BIND
小程序绑定手机号
后台预创建账号认领
账号合并内部流程
手机号换绑
用户注销
token 失效
operation_logs
```

## 4.4 不允许做

```text
不要把验证码明文存数据库
不要把 token 写入日志
不要绕过账号合并冲突
不要允许 DISABLED / CANCELLED / DELETED 用户登录
```

## 4.5 Codex 提示词

```text
当前执行阶段三：认证与账号模块。

请阅读 04-01-auth-account-api.md。
实现认证与账号接口，必须覆盖：
1. send-code
2. register-phone
3. login-phone
4. logout
5. wechat-mini/login
6. wechat-mini/bind-phone
7. change-phone
8. cancel-account
9. 账号合并内部流程
10. 后台预创建账号认领内部流程

要求：
- 验证码只保存 hash。
- 账号合并后 source_user 旧 token 失效。
- 小程序绑定已有手机号时触发合并。
- 小程序绑定 PENDING_CLAIM 手机号时触发认领。
- 用户注销释放 phone。
- 所有关键操作写 operation_logs。
- 添加基础单元测试。

不要实现家庭成员关系模块。
```

---

# 5. 阶段四：家庭基础模块

## 5.1 接口范围

```http
POST /api/families
GET  /api/families
GET  /api/families/{familyId}
PUT  /api/families/{familyId}
GET  /api/public/families/{familyId}/profile
GET  /api/families/{familyId}/tree
GET  /api/public/families/{familyId}/tree
```

## 5.2 必须实现

```text
创建家庭时创建 FOUNDER member 与 link
搜索 NORMAL + searchable 家庭
公开主页只展示 APPROVED 家庭
家庭树返回 nodes + edges + tree
不显示动态称谓
公开联系方式可以为空
```

## 5.3 Codex 提示词

```text
当前执行阶段四：家庭基础模块。

请阅读 04-02-family-basic-api.md。
实现家庭创建、搜索、详情、编辑、公开主页、家庭树基础接口。

要求：
1. 创建家庭时必须同时创建 family、founder member、founder link。
2. family_role = FOUNDER。
3. families.current_founder_member_id 正确写入。
4. 游客只能搜索 NORMAL + searchable = true 的家庭。
5. 公开主页只允许 public_display_status = APPROVED。
6. 家庭树返回 nodes + edges + tree。
7. 当前不要实现动态称谓。
8. 写 operation_logs。
```

---

# 6. 阶段五：家庭成员与关系模块

## 6.1 接口范围

```http
POST   /api/families/{familyId}/members
GET    /api/families/{familyId}/members/{memberId}
PUT    /api/families/{familyId}/members/{memberId}
DELETE /api/families/{familyId}/members/{memberId}
POST   /api/families/{familyId}/members/{memberId}/mark-not-required
POST   /api/families/{familyId}/members/{memberId}/restore
POST   /api/families/{familyId}/relationships
PUT    /api/families/{familyId}/relationships/{relationshipId}
DELETE /api/families/{familyId}/relationships/{relationshipId}
```

## 6.2 必须实现

```text
NOT_REQUIRED 成员
禁止删除存在下级成员的节点
添加兄弟姐妹时必须已有父母节点
SIBLING 不落库
PRIMARY 父亲 / 母亲唯一性校验
允许多个配偶
graph_version 更新
operation_logs
```

## 6.3 Codex 提示词

```text
当前执行阶段五：家庭成员与关系模块。

请阅读 04-03-family-member-relationship-api.md。
实现成员与关系接口。

特别注意：
1. relationship_type 只能是 PARENT_CHILD / SPOUSE。
2. 不允许保存 SIBLING。
3. ADD_SIBLING 必须通过共同父母创建 PARENT_CHILD。
4. 如果当前成员无父母，ADD_SIBLING 必须失败并提示先创建父母节点。
5. 删除存在下级成员的节点必须失败。
6. member.user_binding_policy = NOT_REQUIRED 时不能创建邀请或绑定。
7. 创建 / 删除成员和关系必须更新 graph_version。
8. 所有重要操作写 operation_logs。
```

---

# 7. 阶段六：邀请与加入家庭模块

## 7.1 接口范围

```http
POST /api/families/{familyId}/members/{memberId}/invite
GET  /api/invitations/{inviteToken}
POST /api/invitations/{invitationId}/accept
POST /api/invitations/{invitationId}/reject
POST /api/invitations/{invitationId}/cancel
GET  /api/users/me/invitations
POST /api/families/{familyId}/join-requests
GET  /api/users/me/join-requests
GET  /api/families/{familyId}/join-requests
POST /api/families/{familyId}/join-requests/{requestId}/approve
POST /api/families/{familyId}/join-requests/{requestId}/reject
POST /api/families/{familyId}/join-requests/{requestId}/cancel
```

## 7.2 关键规则

```text
PLATFORM_ADMIN 只能站内邀请已注册用户
PLATFORM_ADMIN 不能生成分享邀请链接
家庭管理员生成分享链接
邀请有效期 7 天
接受邀请前必须绑定手机号
加入申请由家庭管理员处理为主，平台协助
```

---

# 8. 阶段七：公开申请与留言模块

## 8.1 接口范围

```http
POST /api/families/{familyId}/public-applications
GET  /api/families/{familyId}/public-applications
GET  /api/admin/family-public-applications
POST /api/admin/family-public-applications/{applicationId}/approve
POST /api/admin/family-public-applications/{applicationId}/reject
POST /api/families/{familyId}/public-applications/{applicationId}/cancel
POST /api/admin/families/{familyId}/take-down-public
POST /api/public/families/{familyId}/visitor-messages
GET  /api/public/families/{familyId}/visitor-messages
GET  /api/admin/visitor-messages
POST /api/admin/visitor-messages/{messageId}/approve
POST /api/admin/visitor-messages/{messageId}/reject
DELETE /api/admin/visitor-messages/{messageId}
```

## 8.2 关键规则

```text
PLATFORM_ADMIN 不能审核自己发起的公开申请
同一 family 只能有一个 PENDING 公开申请
游客可匿名留言
留言审核通过后展示
公开留言不展示 visitorPhone / visitorWechat
```

---

# 9. 阶段八：家庭角色、转让与解散模块

## 9.1 接口范围

```http
POST /api/families/{familyId}/members/{memberId}/set-admin
POST /api/families/{familyId}/members/{memberId}/unset-admin
POST /api/families/{familyId}/founder-transfer-requests
GET  /api/admin/founder-transfer-requests
POST /api/admin/founder-transfer-requests/{requestId}/approve
POST /api/admin/founder-transfer-requests/{requestId}/reject
POST /api/families/{familyId}/founder-transfer-requests/{requestId}/cancel
POST /api/families/{familyId}/dissolution-requests
GET  /api/admin/dissolution-requests
POST /api/admin/dissolution-requests/{requestId}/approve
POST /api/admin/dissolution-requests/{requestId}/reject
POST /api/families/{familyId}/dissolution-requests/{requestId}/cancel
POST /api/admin/families/{familyId}/restore
```

## 9.2 关键规则

```text
只有 FOUNDER 可设置 FAMILY_ADMIN
NOT_REQUIRED 或无 ACTIVE link 的成员不能成为管理员
创始人转让必须 ROOT_ADMIN / SUPER_ADMIN 审核
家庭解散必须 ROOT_ADMIN / SUPER_ADMIN 审核
PLATFORM_ADMIN 不能审核以上高风险操作
解散后数据保留
恢复后默认 PRIVATE
```

---

# 10. 阶段九：后台管理与操作日志

## 10.1 后台管理接口

```http
POST /api/admin/auth/login
POST /api/admin/auth/logout
GET  /api/admin/me
PUT  /api/admin/me/password
GET  /api/admin/admin-users
POST /api/admin/admin-users
PUT  /api/admin/admin-users/{adminId}
POST /api/admin/admin-users/{adminId}/disable
DELETE /api/admin/admin-users/{adminId}
POST /api/admin/admin-users/{adminId}/unlock
GET  /api/admin/users
GET  /api/admin/users/{userId}
PUT  /api/admin/users/{userId}
POST /api/admin/users/{userId}/disable
POST /api/admin/users/{userId}/restore
POST /api/admin/users/pre-create
GET  /api/admin/families
GET  /api/admin/families/{familyId}
PUT  /api/admin/families/{familyId}
POST /api/admin/families/{familyId}/disable
DELETE /api/admin/families/{familyId}
POST /api/admin/families/{familyId}/restore
```

## 10.2 操作日志接口

```http
GET /api/admin/operation-logs
GET /api/admin/operation-logs/{logId}
```

## 10.3 关键规则

```text
ROOT_ADMIN 有且只有一个
任何管理员不能管理同级管理员
ROOT_ADMIN 登录失败不自动锁定
只有 ROOT_ADMIN 可解锁管理员
PLATFORM_ADMIN 可查看完整手机号
一期不做导出
```

---

# 11. 阶段十：测试与修复

## 11.1 必须运行

```text
单元测试
接口测试
权限测试
关键业务流程测试
graph_version 测试
operation_logs 测试
```

## 11.2 必须覆盖

参考：

```text
11-test-cases-and-acceptance.md
```

## 11.3 Codex 提示词

```text
当前执行阶段十：测试与修复。

请阅读 11-test-cases-and-acceptance.md。
为 P0 和 P1 测试用例补充自动化测试。
运行测试。
修复失败问题。
输出：
1. 已覆盖测试用例
2. 未覆盖测试用例
3. 失败与修复记录
4. 仍需人工验收的部分
```

---

# 12. 每阶段交付格式

Codex 每完成一个阶段，必须输出：

```text
1. 本阶段完成内容
2. 修改文件列表
3. 新增接口列表
4. 新增表 / 字段 / 索引
5. 已运行测试
6. 未完成事项
7. 风险提醒
8. 下一阶段建议
```

---

# 13. 禁止行为

Codex 不得：

```text
1. 删除文档中已确认的核心业务规则。
2. 自行简化账号合并。
3. 将 family_member 与 user 合并成一个表。
4. 将 SIBLING 写入 relationship_type。
5. 跳过权限校验。
6. 跳过 operation_logs。
7. 跳过 graph_version。
8. 用硬编码模拟登录用户。
9. 在生产代码中留下 mock 数据。
10. 擅自引入重型依赖。
11. 擅自更换技术栈。
12. 一次性大范围重写已有代码。
```


---

# Go 后端补充要求

阶段一创建项目骨架时，必须生成 Go 后端基础结构：

```text
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
backend/internal/operationlog
backend/internal/family/tree
backend/go.mod
backend/configs/config.example.yaml
```

推荐依赖：

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

阶段一不得实现复杂业务，只搭骨架。

---

# v7 执行补充

Codex 执行各阶段时必须参考：

```text
12-local-dev-setup.md
13-ui-page-spec.md
14-database-migration-ddl.md
15-error-code-summary.md
16-development-milestones.md
17-api-debug-guide.md
```

阶段一额外要求：创建 docker/docker-compose.yml、backend/.env.example、backend/configs/config.example.yaml、GET /api/health、统一错误码结构、migrations 目录。
