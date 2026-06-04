# 04-07. 后台用户管理、后台家庭管理与后台管理员接口完整规格 v1

## 1. 模块范围

覆盖：

```text
后台登录 / 退出
后台管理员管理
普通用户管理
后台预创建用户
家庭后台管理
管理员锁定 / 解锁
```

涉及表：

```text
admin_users
users
families
family_members
family_member_user_links
user_auth_identities
user_phone_history
user_account_claim_logs
operation_logs
```

## 2. 后台角色

```text
ROOT_ADMIN
SUPER_ADMIN
PLATFORM_ADMIN
```

ROOT_ADMIN：

```text
username = admin
有且只有一个
不能删除 / 禁用 / 降级
登录失败不自动锁定
可管理 SUPER_ADMIN / PLATFORM_ADMIN
可解除管理员锁定
可配置管理员锁定策略
```

SUPER_ADMIN：

```text
可管理 PLATFORM_ADMIN
可处理高风险业务
不能管理 ROOT_ADMIN 或同级 SUPER_ADMIN
```

PLATFORM_ADMIN：

```text
可信任运营人员
可查看完整手机号
可编辑普通用户资料、禁用普通用户、预创建用户
可编辑家庭、成员、审核公开申请和留言、下架公开家庭
不能管理后台管理员
不能审核创始人转让 / 家庭解散
不能恢复家庭
不能导出数据
```

## 3. POST /api/admin/auth/login

后台登录。

请求：

```json
{
  "username": "admin",
  "password": "Password123!"
}
```

规则：

```text
登录成功 / 失败都写 operation_logs
ROOT_ADMIN 失败不自动锁定
SUPER_ADMIN / PLATFORM_ADMIN 失败达到阈值可锁定
锁定时间由 ROOT_ADMIN 配置
```

## 4. POST /api/admin/auth/logout

后台 token 失效，写 operation_logs。

## 5. GET /api/admin/me

当前后台管理员信息。

## 6. PUT /api/admin/me/password

修改自己密码，校验旧密码，写 operation_logs。

## 7. GET /api/admin/admin-users

后台管理员列表。

权限：

```text
ROOT_ADMIN
SUPER_ADMIN
```

PLATFORM_ADMIN 不可访问。

## 8. POST /api/admin/admin-users

创建后台管理员。

权限：

```text
ROOT_ADMIN 可创建 SUPER_ADMIN / PLATFORM_ADMIN
SUPER_ADMIN 可创建 PLATFORM_ADMIN
PLATFORM_ADMIN 不可创建
```

不允许创建第二个 ROOT_ADMIN。

## 9. PUT /api/admin/admin-users/{adminId}

编辑后台管理员基础资料。

规则：

```text
不能管理 ROOT_ADMIN
不能管理同级管理员
不能通过该接口修改 role / password / status / is_root_admin
```

## 10. POST /api/admin/admin-users/{adminId}/disable

禁用后台管理员。

规则：

```text
ROOT_ADMIN 不可禁用
不能禁用同级管理员
不能禁用自己
```

## 11. DELETE /api/admin/admin-users/{adminId}

软删除后台管理员。

## 12. POST /api/admin/admin-users/{adminId}/unlock

解除管理员锁定，仅 ROOT_ADMIN 可操作。

## 13. GET /api/admin/users

普通用户列表。

权限：

```text
ROOT_ADMIN
SUPER_ADMIN
PLATFORM_ADMIN
```

支持状态、关键词、来源、注册端、手机号验证、时间筛选。

## 14. GET /api/admin/users/{userId}

普通用户详情，包含账号、微信身份摘要、加入家庭、绑定成员、合并认领摘要。

## 15. PUT /api/admin/users/{userId}

编辑普通用户普通资料。

可编辑：

```text
nickname
avatar_url
real_name
remark
```

不能编辑：

```text
phone
phone_verified
openid
unionid
account_origin
register_client
merged_to_user_id
claimed_at
账号合并记录
账号认领记录
```

## 16. POST /api/admin/users/{userId}/disable

禁用普通用户。

注意：

```text
如果该用户是家庭 FOUNDER，应提示高风险处理。
前面规则：家庭创始人被平台禁用时，家庭正常情况下平台会设立新创始人。
建议由 ROOT_ADMIN / SUPER_ADMIN 处理此场景。
```

## 17. POST /api/admin/users/{userId}/restore

恢复普通用户。

建议仅 ROOT_ADMIN / SUPER_ADMIN 支持。

## 18. POST /api/admin/users/pre-create

后台预创建账号。

请求：

```json
{
  "phone": "13800000000",
  "nickname": "张三",
  "realName": "张明远",
  "remark": "平台预创建账号，等待用户认领"
}
```

规则：

```text
phone 必填
phone 不能被 ACTIVE user 占用
已有 PENDING_CLAIM 不重复创建
创建 users.status = PENDING_CLAIM
account_origin = ADMIN_PRE_CREATED
register_client = ADMIN_WEB
写 user_phone_history
写 operation_logs
```

## 19. GET /api/admin/families

后台家庭列表。支持状态、公开状态、关键词、姓氏、地区、时间筛选。

## 20. GET /api/admin/families/{familyId}

后台家庭详情。

## 21. PUT /api/admin/families/{familyId}

后台编辑家庭资料。

PLATFORM_ADMIN 可编辑家庭基础资料。修改公开资料直接生效并写日志。

如果修改 familySurname 影响主线判断，建议触发 graph_version + 1。

## 22. POST /api/admin/families/{familyId}/disable

禁用家庭。

影响：

```text
families.status = DISABLED
families.public_display_status = PRIVATE
families.searchable = 0
```

## 23. DELETE /api/admin/families/{familyId}

删除家庭，建议仅 ROOT_ADMIN / SUPER_ADMIN 支持。软删除，历史数据保留。

## 24. POST /api/admin/families/{familyId}/restore

恢复家庭，仅 ROOT_ADMIN / SUPER_ADMIN 支持。

## 25. 错误码范围

```text
47000-47999
```
