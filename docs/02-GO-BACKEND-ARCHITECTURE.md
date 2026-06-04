# 02-GO-BACKEND-ARCHITECTURE.md

# Go 后端项目架构规范 v1

> 本文档专门约束 Go 后端代码结构、模块分层、命名规范、事务规范、日志规范和禁止事项。  
> Codex 实现后端时必须优先遵守本文档。

---

# 1. 后端总目录

```text
backend
├── cmd
│   └── server
│       └── main.go
├── configs
├── deployments
├── migrations
├── internal
├── pkg
├── scripts
├── docs
├── go.mod
└── go.sum
```

---

# 2. internal 目录

```text
internal
├── app
├── common
├── auth
├── user
├── account
├── family
├── admin
└── operationlog
```

---

# 3. common 目录

```text
common
├── response
├── errors
├── enums
├── middleware
├── validator
├── jwt
├── redis
├── logger
├── transaction
└── config
```

说明：

| 目录 | 职责 |
|---|---|
| `response` | 统一返回结构 |
| `errors` | 错误码和业务错误 |
| `enums` | 全局枚举 |
| `middleware` | Gin 中间件 |
| `validator` | 参数校验 |
| `jwt` | token 生成与校验 |
| `redis` | Redis 客户端封装 |
| `logger` | zap 日志 |
| `transaction` | GORM 事务封装 |
| `config` | 配置读取 |

---

# 4. 标准模块结构

每个业务模块采用：

```text
module
├── handler
├── service
├── repository
├── model
├── dto
├── vo
└── enum
```

示例：

```text
family/member
├── handler
│   └── member_handler.go
├── service
│   └── member_service.go
├── repository
│   └── member_repository.go
├── model
│   └── family_member.go
├── dto
│   ├── create_member_request.go
│   └── update_member_request.go
├── vo
│   └── member_detail_vo.go
└── enum
    └── member_enum.go
```

---

# 5. Handler 规范

handler 只允许做：

```text
1. 绑定请求参数。
2. 调用 validator。
3. 获取当前登录 user/admin。
4. 调用 service。
5. 返回 response。
```

handler 禁止：

```text
1. 直接写复杂业务规则。
2. 直接开启事务。
3. 直接操作多个 repository。
4. 直接写 operation_logs。
5. 直接更新 graph_version。
```

---

# 6. Service 规范

service 负责：

```text
1. 权限校验。
2. 业务规则判断。
3. 事务控制。
4. 调用 repository。
5. 调用 operation log service。
6. 调用 graph version service。
7. 返回业务结果。
```

复杂流程必须放在 service：

```text
账号合并
账号认领
创建家庭
删除成员
添加兄弟姐妹
接受邀请
通过加入申请
创始人转让
家庭解散
```

---

# 7. Repository 规范

repository 只负责数据访问：

```text
Create
Update
FindByID
FindList
FindActiveLink
FindPendingRequest
```

repository 不负责：

```text
权限判断
业务状态机
operation_logs
graph_version
```

---

# 8. 事务规范

以下操作必须使用数据库事务：

```text
创建家庭
账号合并
账号认领
接受邀请
通过加入申请
删除成员
添加关系
创始人转让审核通过
家庭解散审核通过
恢复家庭
```

Go 实现建议：

```go
func (tm *TransactionManager) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error
```

---

# 9. OperationLogService 规范

所有敏感操作必须通过统一服务写日志：

```go
OperationLogService.WriteSuccess(ctx, input)
OperationLogService.WriteFailed(ctx, input)
```

不要在业务模块里手写散乱日志插入。

---

# 10. GraphVersionService 规范

影响家庭树结构的操作必须调用：

```go
GraphVersionService.Increment(ctx, familyID)
```

触发操作：

```text
新增成员
删除成员
恢复成员
新增关系
删除关系
修改成员关键字段
修改关系性质
```

---

# 11. FamilyPermissionService 规范

家庭权限不能只靠路由角色判断。

必须提供：

```go
CanViewFamily(ctx, userID, familyID)
CanEditFamily(ctx, userID, familyID)
CanCreateMember(ctx, userID, familyID)
CanEditMember(ctx, userID, familyID, memberID)
CanDeleteMember(ctx, userID, familyID, memberID)
CanInviteMember(ctx, userID, familyID, memberID)
CanSetFamilyAdmin(ctx, userID, familyID)
CanRequestFounderTransfer(ctx, userID, familyID)
CanRequestDissolution(ctx, userID, familyID)
```

---

# 12. 重要业务红线

Go 后端不得违反：

```text
1. family_member 与 user 分离。
2. family_member 可以没有 user。
3. relationship_type 只能是 PARENT_CHILD / SPOUSE。
4. 不保存 SIBLING。
5. ADD_SIBLING 必须通过共同父母创建 PARENT_CHILD。
6. 删除存在下级成员的节点必须失败。
7. NOT_REQUIRED 成员不能绑定用户，不能成为管理员。
8. 一个 user 在同一 family 只能有一个 ACTIVE member link。
9. 一个 member 只能有一个 ACTIVE user link。
10. 同一 family 只能有一个 FOUNDER。
11. 账号合并后旧 token 必须失效。
12. PLATFORM_ADMIN 不能审核创始人转让。
13. PLATFORM_ADMIN 不能审核家庭解散。
14. 所有高风险操作必须写 operation_logs。
```

---

# 13. 测试规范

每个 service 至少应有核心单元测试。

必须覆盖：

```text
账号合并
账号认领
创建家庭
删除成员
添加兄弟姐妹
邀请接受
加入申请通过
公开申请审核
创始人转让
家庭解散
权限越权
operation_logs
graph_version
```

---

# 14. Codex 阶段一必须生成

阶段一只搭骨架时，Go 后端至少生成：

```text
cmd/server/main.go
internal/app/server.go
internal/app/router.go
internal/common/response
internal/common/errors
internal/common/enums
internal/common/middleware
internal/common/logger
internal/common/config
internal/common/transaction
internal/operationlog/service
internal/family/tree/service
go.mod
configs/config.example.yaml
```

不实现具体复杂业务。
