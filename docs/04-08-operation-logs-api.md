# 04-08. 操作日志接口完整规格 v1

## 1. 模块范围

覆盖：

```text
记录后台管理员敏感操作
记录普通用户重要操作
记录系统自动处理操作
后台日志筛选
后台日志分页
后台日志详情
失败操作追踪
before_json / after_json 对比
```

涉及表：

```text
operation_logs
admin_users
users
families
family_members
```

## 2. 核心规则

```text
普通用户和后台管理员日志使用同一张 operation_logs
operator_type = ADMIN / USER / SYSTEM
后台敏感操作必须记录
普通用户重要操作建议记录
系统自动操作也应记录
一期不做导出、删除、清空、批量归档
```

## 3. 主要枚举

operator_type：

```text
ADMIN
USER
SYSTEM
```

result：

```text
SUCCESS
FAILED
PARTIAL_SUCCESS
```

module：

```text
AUTH
ADMIN_USER
USER
ACCOUNT
FAMILY
FAMILY_MEMBER
FAMILY_RELATIONSHIP
FAMILY_INVITATION
FAMILY_JOIN_REQUEST
PUBLIC_DISPLAY
MESSAGE
FAMILY_ROLE
FOUNDER_TRANSFER
FAMILY_DISSOLUTION
FAMILY_RESTORE
SYSTEM_SETTING
```

## 4. 常用 action

AUTH：

```text
ADMIN_LOGIN_SUCCESS
ADMIN_LOGIN_FAILED
ADMIN_LOCKED
ADMIN_UNLOCKED
USER_LOGIN
USER_LOGOUT
USER_BIND_PHONE
USER_CHANGE_PHONE
USER_CANCEL_ACCOUNT
SEND_VERIFICATION_CODE
VERIFY_CODE_SUCCESS
VERIFY_CODE_FAILED
```

ACCOUNT：

```text
MERGE_USER
CLAIM_PRE_CREATED_USER
RELEASE_PHONE
MIGRATE_AUTH_IDENTITY
INVALIDATE_TOKEN
```

FAMILY_MEMBER：

```text
CREATE_MEMBER
UPDATE_MEMBER
DELETE_MEMBER
RESTORE_MEMBER
MARK_MEMBER_NOT_REQUIRED
BIND_MEMBER_USER
UNLINK_MEMBER_USER
```

FAMILY_RELATIONSHIP：

```text
ADD_FATHER
ADD_MOTHER
ADD_CHILD
ADD_SIBLING
ADD_SPOUSE
CREATE_RELATIONSHIP
UPDATE_RELATIONSHIP
DELETE_RELATIONSHIP
```

## 5. GET /api/admin/operation-logs

后台操作日志列表。

### 权限

```text
ROOT_ADMIN
SUPER_ADMIN
PLATFORM_ADMIN
```

一期 PLATFORM_ADMIN 可查看完整日志。后续可限制高风险模块日志。

### 查询参数

```text
operatorType
operatorRole
operatorAdminId
operatorUserId
module
action
targetType
targetId
familyId
memberId
userId
result
keyword
createdStart
createdEnd
page
pageSize
```

keyword 搜索范围：

```text
action
module
target_type
error_message
operator_admin.username
operator_admin.display_name
operator_user.nickname
operator_user.real_name
family.family_name
family.family_surname
family_member.display_name
```

默认排序：

```sql
ORDER BY created_at DESC
```

## 6. GET /api/admin/operation-logs/{logId}

后台操作日志详情。

返回：

```text
操作者信息
目标对象信息
before_json
after_json
detail_json
result
error_message
ip
user_agent
created_at
```

敏感字段不返回：

```text
password_hash
验证码
accessToken
refreshToken
完整 openid
完整 unionid
session_key
```

## 7. 日志写入规范

成功操作记录：

```text
operator_type
operator_admin_id / operator_user_id
operator_role
module
action
target_type
target_id
before_json
after_json
result = SUCCESS
ip
user_agent
```

失败操作也记录，尤其是：

```text
后台登录失败
权限不足
删除存在下级成员失败
审核权限不足
验证码验证失败
非法关系创建失败
```

## 8. before_json / after_json

编辑类操作建议保存关键字段差异，不记录密码、验证码、token、session_key 等敏感数据。

## 9. 错误码范围

```text
48000-48999
```
