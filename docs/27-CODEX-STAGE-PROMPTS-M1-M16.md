# 27-CODEX-STAGE-PROMPTS-M1-M16.md

# Codex 全阶段提示词 v1

> 每次只复制一个阶段给 Codex。不要一次性复制全部阶段。

---

# M1：项目骨架与本地环境

请先阅读项目根目录的 `AGENTS.md`，然后阅读：

```text
docs/00-TECH-STACK-FINAL.md
docs/01-system-architecture-and-tech-stack.md
docs/02-GO-BACKEND-ARCHITECTURE.md
docs/CODEX_TASKS.md
docs/12-local-dev-setup.md
docs/15-error-code-summary.md
docs/16-development-milestones.md
```

当前只执行：

```text
M1：项目骨架与本地环境
```

允许修改：

```text
backend/
docker/
scripts/
README.md
.gitignore
AGENTS.md
```

必须完成：

```text
1. 创建 Go + Gin 后端骨架。
2. 创建 go.mod。
3. 创建 cmd/server/main.go。
4. 创建 internal/app/server.go 和 router.go。
5. 创建统一响应结构。
6. 创建基础错误码结构。
7. 创建基础 logger。
8. 创建基础 config 读取结构。
9. 创建 GET /api/health。
10. 创建 docker/docker-compose.yml，包含 MySQL 8 和 Redis 7。
11. 创建 backend/.env.example。
12. 创建 backend/migrations 目录。
```

禁止：

```text
不要实现注册登录。
不要实现家庭业务。
不要实现成员关系。
不要创建 Java / Node / Python 后端。
```

完成后运行：

```bash
cd backend
gofmt -w .
go test ./...
go run ./cmd/server
docker compose -f docker/docker-compose.yml config
```

---

# M2：数据库 migration 与基础 GORM model

请阅读：

```text
AGENTS.md
docs/14-database-migration-ddl.md
docs/15-error-code-summary.md
docs/02-GO-BACKEND-ARCHITECTURE.md
docs/16-development-milestones.md
```

当前只执行：

```text
M2：数据库 migration 与基础 GORM model
```

允许修改：

```text
backend/migrations/
backend/internal/**/model/
backend/internal/**/enum/
backend/internal/common/enums/
backend/internal/common/database/
backend/internal/common/config/
README.md
```

必须完成：

```text
1. 根据 14-database-migration-ddl.md 拆分 migration。
2. 创建 GORM model。
3. 创建核心枚举常量。
4. 创建数据库连接初始化。
5. 创建 ROOT_ADMIN seed migration。
6. 保留 Service 层强校验 TODO 注释。
```

必须覆盖表：

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

禁止：

```text
不要实现业务接口。
不要实现登录。
不要实现家庭关系业务。
```

完成后运行：

```bash
cd backend
gofmt -w .
go test ./...
```

---

# M3：基础设施、JWT、中间件、日志基础

请阅读：

```text
AGENTS.md
docs/02-GO-BACKEND-ARCHITECTURE.md
docs/15-error-code-summary.md
docs/22-security-and-privacy-checklist.md
```

当前只执行：

```text
M3：统一响应、错误码、JWT、中间件、日志基础设施
```

允许修改：

```text
backend/internal/common/
backend/internal/operationlog/
backend/internal/app/
backend/configs/
```

必须完成：

```text
1. 完善统一响应结构。
2. 完善错误码常量结构。
3. 实现 JWT 生成与解析基础工具。
4. 实现 UserAuthMiddleware 骨架。
5. 实现 AdminAuthMiddleware 骨架。
6. 实现 RequirePhoneVerified 中间件骨架。
7. 实现 zap logger 初始化。
8. 实现 OperationLogService 基础接口和空实现或基础实现。
9. 实现 TransactionManager.WithTransaction。
10. 实现 Redis client 初始化。
```

禁止：

```text
不要实现具体注册登录业务。
不要实现家庭业务。
不要写死测试 token。
不要记录敏感信息。
```

完成后运行：

```bash
cd backend
gofmt -w .
go test ./...
```

---

# M4：后台管理员认证与权限骨架

请阅读：

```text
AGENTS.md
docs/04-07-admin-management-api.md
docs/15-error-code-summary.md
docs/22-security-and-privacy-checklist.md
```

当前只执行：

```text
M4：后台管理员认证与基础权限
```

允许修改：

```text
backend/internal/admin/
backend/internal/common/security/
backend/internal/common/jwt/
backend/internal/operationlog/
backend/internal/app/router.go
```

必须实现接口：

```http
POST /api/admin/auth/login
POST /api/admin/auth/logout
GET  /api/admin/me
PUT  /api/admin/me/password
```

必须实现规则：

```text
1. ROOT_ADMIN 可登录。
2. SUPER_ADMIN / PLATFORM_ADMIN 登录失败可锁定。
3. ROOT_ADMIN 登录失败不自动锁定，但写失败日志。
4. admin token 与 user token 分离。
5. 后台登录成功 / 失败写 operation_logs。
6. 密码使用 hash。
```

禁止：

```text
不要实现普通用户认证。
不要实现后台管理员管理 CRUD。
不要跳过 operation_logs。
```

完成后运行：

```bash
cd backend
gofmt -w .
go test ./...
```

---

# M5：普通用户认证、验证码、手机号注册登录

请阅读：

```text
AGENTS.md
docs/04-01-auth-account-api.md
docs/15-error-code-summary.md
docs/22-security-and-privacy-checklist.md
```

当前只执行：

```text
M5：普通用户认证、手机号注册登录、验证码
```

允许修改：

```text
backend/internal/auth/
backend/internal/user/
backend/internal/common/jwt/
backend/internal/common/redis/
backend/internal/operationlog/
backend/internal/app/router.go
```

必须实现接口：

```http
POST /api/auth/send-code
POST /api/auth/register-phone
POST /api/auth/login-phone
POST /api/auth/logout
```

必须实现规则：

```text
1. 验证码 5 分钟有效。
2. 验证码只保存 hash。
3. 密码只保存 hash。
4. 手机号注册。
5. 手机号密码登录。
6. DISABLED / CANCELLED / MERGED 用户不能登录。
7. logout 后 token 失效。
8. 登录、注册、验证码关键操作写日志。
```

禁止：

```text
不要实现微信登录。
不要实现账号合并。
不要实现家庭业务。
```

完成后运行：

```bash
cd backend
gofmt -w .
go test ./...
```

---

# M6：微信小程序登录、手机号绑定、账号认领、账号合并

请阅读：

```text
AGENTS.md
docs/04-01-auth-account-api.md
docs/15-error-code-summary.md
docs/22-security-and-privacy-checklist.md
```

当前只执行：

```text
M6：微信小程序登录、手机号绑定、账号认领、账号合并
```

允许修改：

```text
backend/internal/auth/
backend/internal/account/
backend/internal/user/
backend/internal/common/jwt/
backend/internal/operationlog/
backend/internal/app/router.go
```

必须实现接口：

```http
POST /api/auth/wechat-mini/login
POST /api/auth/wechat-mini/bind-phone
POST /api/auth/change-phone
POST /api/auth/cancel-account
```

必须实现规则：

```text
1. 小程序首次登录生成 PENDING_PHONE_BIND user。
2. 绑定新手机号成功。
3. 绑定已有手机号触发账号合并。
4. 绑定 PENDING_CLAIM 手机号触发账号认领。
5. 合并后旧 token 失效。
6. 注销释放手机号。
7. 账号合并处理同一 family 下 member 冲突。
8. 全部关键操作写 operation_logs。
```

必须使用事务：

```text
账号合并
账号认领
手机号换绑
用户注销
```

禁止：

```text
不要把 AppSecret 写进前端。
不要日志记录 session_key、openid、unionid 完整值。
不要实现家庭业务。
```

完成后运行：

```bash
cd backend
gofmt -w .
go test ./...
```

---

# M7：家庭基础模块

请阅读：

```text
AGENTS.md
docs/04-02-family-basic-api.md
docs/14-database-migration-ddl.md
docs/15-error-code-summary.md
```

当前只执行：

```text
M7：家庭基础模块
```

允许修改：

```text
backend/internal/family/core/
backend/internal/family/tree/
backend/internal/family/role/
backend/internal/operationlog/
backend/internal/app/router.go
```

必须实现接口：

```http
POST /api/families
GET  /api/families
GET  /api/families/{familyId}
PUT  /api/families/{familyId}
GET  /api/public/families/{familyId}/profile
```

必须实现规则：

```text
1. 创建家庭时同时创建 family、founder member、founder link。
2. family_role = FOUNDER。
3. current_founder_member_id 正确写入。
4. family_surname 当前版本只允许一个姓氏。
5. 搜索返回正常状态家庭。
6. 公开主页只允许 APPROVED 家庭访问。
7. 创建家庭写 operation_logs。
```

必须使用事务：

```text
创建家庭
```

禁止：

```text
不要实现成员关系复杂逻辑。
不要实现邀请。
```

完成后运行：

```bash
cd backend
gofmt -w .
go test ./...
```

---

# M8：家庭成员模块

请阅读：

```text
AGENTS.md
docs/04-03-family-member-relationship-api.md
docs/14-database-migration-ddl.md
docs/15-error-code-summary.md
```

当前只执行：

```text
M8：家庭成员模块
```

允许修改：

```text
backend/internal/family/member/
backend/internal/family/permission/
backend/internal/family/tree/
backend/internal/operationlog/
backend/internal/app/router.go
```

必须实现接口：

```http
POST   /api/families/{familyId}/members
GET    /api/families/{familyId}/members/{memberId}
PUT    /api/families/{familyId}/members/{memberId}
DELETE /api/families/{familyId}/members/{memberId}
POST   /api/families/{familyId}/members/{memberId}/mark-not-required
POST   /api/families/{familyId}/members/{memberId}/restore
```

必须实现规则：

```text
1. 创建成员更新 graph_version。
2. 编辑成员关键字段更新 graph_version。
3. 删除成员软删除。
4. 存在下级成员不能删除。
5. FOUNDER 不能直接删除。
6. NOT_REQUIRED 成员不能绑定用户。
7. 标记 NOT_REQUIRED 前必须无 ACTIVE link 和 PENDING invitation。
8. 操作写 operation_logs。
```

必须使用事务：

```text
删除成员
恢复成员
标记 NOT_REQUIRED
```

禁止：

```text
不要实现关系添加。
不要实现 SIBLING。
```

完成后运行：

```bash
cd backend
gofmt -w .
go test ./...
```

---

# M9：家庭关系模块

请阅读：

```text
AGENTS.md
docs/04-03-family-member-relationship-api.md
docs/15-error-code-summary.md
docs/11-test-cases-and-acceptance.md
```

当前只执行：

```text
M9：家庭关系模块
```

允许修改：

```text
backend/internal/family/relationship/
backend/internal/family/member/
backend/internal/family/tree/
backend/internal/family/permission/
backend/internal/operationlog/
backend/internal/app/router.go
```

必须实现接口：

```http
POST   /api/families/{familyId}/relationships
PUT    /api/families/{familyId}/relationships/{relationshipId}
DELETE /api/families/{familyId}/relationships/{relationshipId}
```

必须实现 addType：

```text
ADD_FATHER
ADD_MOTHER
ADD_CHILD
ADD_SPOUSE
ADD_SIBLING
```

必须实现规则：

```text
1. relationship_type 只能是 PARENT_CHILD / SPOUSE。
2. 不保存 SIBLING。
3. ADD_SIBLING 必须通过共同父母创建 PARENT_CHILD。
4. baseMember 无父母时 ADD_SIBLING 必须失败。
5. PRIMARY 父亲唯一。
6. PRIMARY 母亲唯一。
7. 多个配偶允许。
8. 关系变化更新 graph_version。
9. 操作写 operation_logs。
```

必须使用事务：

```text
添加关系
删除关系
ADD_SIBLING
```

完成后运行：

```bash
cd backend
gofmt -w .
go test ./...
```

---

# M10：家庭树组装与 graph_version 缓存

请阅读：

```text
AGENTS.md
docs/04-02-family-basic-api.md
docs/04-03-family-member-relationship-api.md
docs/12-local-dev-setup.md
```

当前只执行：

```text
M10：家庭树组装与 graph_version
```

允许修改：

```text
backend/internal/family/tree/
backend/internal/family/core/
backend/internal/family/member/
backend/internal/family/relationship/
backend/internal/common/redis/
backend/internal/app/router.go
```

必须实现接口：

```http
GET /api/families/{familyId}/tree
GET /api/public/families/{familyId}/tree
```

必须实现规则：

```text
1. 返回 nodes、edges、tree。
2. 当前一期是 LIST_TREE。
3. 不显示动态称谓。
4. public tree 只允许公开家庭访问。
5. tree cache key = family:{familyId}:tree:v{graphVersion}。
6. graph_version 变化后自然失效旧缓存。
```

禁止：

```text
不要实现复杂图谱。
不要实现中心人物动态称谓。
```

完成后运行：

```bash
cd backend
gofmt -w .
go test ./...
```

---

# M11：邀请与加入家庭模块

请阅读：

```text
AGENTS.md
docs/04-04-invitation-join-api.md
docs/15-error-code-summary.md
docs/22-security-and-privacy-checklist.md
```

当前只执行：

```text
M11：邀请与加入家庭模块
```

必须实现接口：

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

必须实现规则：

```text
1. PLATFORM_ADMIN 只能站内邀请已注册用户。
2. PLATFORM_ADMIN 不能生成分享邀请链接。
3. 家庭管理员生成分享链接。
4. 邀请 7 天有效。
5. 接受邀请前必须登录并绑定手机号。
6. target member 必须无 ACTIVE 绑定。
7. 当前 user 在同一 family 无 ACTIVE 绑定。
8. 接受邀请创建 ACTIVE link。
9. 接受邀请不更新 graph_version。
10. 加入申请由家庭管理员处理为主，平台可协助。
```

必须使用事务：

```text
接受邀请
通过加入申请
```

完成后运行：

```bash
cd backend
gofmt -w .
go test ./...
```

---

# M12：公开申请与游客留言模块

请阅读：

```text
AGENTS.md
docs/04-05-public-application-message-api.md
docs/15-error-code-summary.md
docs/22-security-and-privacy-checklist.md
```

当前只执行：

```text
M12：家庭公开申请与游客留言
```

必须实现接口：

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

必须实现规则：

```text
1. 同一 family 只能有一个 PENDING 公开申请。
2. PLATFORM_ADMIN 不能审核自己发起的公开申请。
3. 审核通过后公开主页可访问。
4. 下架后公开主页不可访问。
5. 游客可匿名留言。
6. 留言审核通过后展示。
7. 公开留言不展示 visitorPhone / visitorWechat。
8. 留言提交需要限流。
9. 所有审核写 operation_logs。
```

完成后运行：

```bash
cd backend
gofmt -w .
go test ./...
```

---

# M13：家庭角色、创始人转让、家庭解散

请阅读：

```text
AGENTS.md
docs/04-06-family-role-transfer-dissolution-api.md
docs/15-error-code-summary.md
docs/22-security-and-privacy-checklist.md
```

当前只执行：

```text
M13：家庭角色、创始人转让、家庭解散
```

必须实现接口：

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

必须实现规则：

```text
1. 只有 FOUNDER 可设置 FAMILY_ADMIN。
2. NOT_REQUIRED 或无 ACTIVE link 的成员不能成为管理员。
3. 创始人转让必须 ROOT_ADMIN / SUPER_ADMIN 审核。
4. PLATFORM_ADMIN 不能审核创始人转让。
5. 家庭解散必须 ROOT_ADMIN / SUPER_ADMIN 审核。
6. PLATFORM_ADMIN 不能审核家庭解散。
7. 解散后数据保留。
8. 恢复后默认 PRIVATE。
9. 所有高风险审核写 operation_logs。
```

必须使用事务：

```text
创始人转让审核通过
家庭解散审核通过
家庭恢复
```

完成后运行：

```bash
cd backend
gofmt -w .
go test ./...
```

---

# M14：管理后台 UI

请阅读：

```text
AGENTS.md
docs/13-ui-page-spec.md
docs/18-design-language.md
docs/19-core-wireframes.md
```

当前只执行：

```text
M14：管理后台 UI 原型与页面实现
```

允许修改：

```text
admin-web/
```

第一步先做静态原型，不接真实 API。

必须实现页面：

```text
登录页
仪表盘
用户管理
用户详情
预创建账号弹窗
家庭管理
家庭详情
家庭成员管理
公开申请审核
游客留言审核
创始人转让审核
家庭解散审核
后台管理员管理
操作日志
```

必须遵守：

```text
1. 使用 Vue 3 + TypeScript + Vite + Element Plus。
2. 使用设计令牌。
3. 使用 mock data。
4. PLATFORM_ADMIN 不显示创始人转让和家庭解散审核入口。
5. 危险操作必须二次确认。
6. 不修改后端。
```

完成后运行：

```bash
cd admin-web
pnpm install
pnpm build
```

---

# M15：PC/H5 前台与微信小程序页面

请阅读：

```text
AGENTS.md
docs/13-ui-page-spec.md
docs/18-design-language.md
docs/19-core-wireframes.md
docs/21-mini-program-launch-checklist.md
```

当前只执行：

```text
M15：PC/H5 前台与微信小程序页面
```

允许修改：

```text
web/
miniapp/
```

先做静态原型，不接真实 API。

web 必须实现：

```text
首页
家庭搜索
公开家庭主页
公开家庭树
登录
注册
个人中心
我的家庭
邀请详情
加入家庭申请
```

miniapp 必须实现：

```text
首页
微信快捷登录
绑定手机号
家庭搜索
公开家庭主页
公开家庭树
我的家庭
家庭详情
邀请确认
加入家庭申请
个人中心
用户协议
隐私政策
注销账号
```

完成后运行：

```bash
cd web
pnpm install
pnpm build

cd ../miniapp
pnpm install
pnpm build:mp-weixin
```

---

# M16：测试补齐、安全检查、部署上线准备

请阅读：

```text
AGENTS.md
docs/11-test-cases-and-acceptance.md
docs/20-deployment-guide.md
docs/21-mini-program-launch-checklist.md
docs/22-security-and-privacy-checklist.md
docs/23-backup-and-monitoring-guide.md
docs/24-production-readiness-checklist.md
```

当前只执行：

```text
M16：测试补齐、安全检查、部署上线准备
```

必须完成：

```text
1. 为 P0 流程补自动化测试。
2. 检查权限越权。
3. 检查 operation_logs。
4. 检查 graph_version。
5. 检查敏感日志。
6. 检查 .env 不入库。
7. 生成 staging 部署脚本。
8. 生成生产上线前检查说明。
9. 生成小程序体验版测试清单。
```

必须运行：

```bash
cd backend
go test ./...

cd ../admin-web
pnpm build

cd ../web
pnpm build
```

如 miniapp 支持：

```bash
cd ../miniapp
pnpm build:mp-weixin
```

禁止：

```text
不要把生产密钥写入代码。
不要覆盖生产数据库。
不要绕过测试失败。
```
