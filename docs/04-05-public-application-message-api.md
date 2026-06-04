# 04-05. 家庭公开申请与游客留言接口完整规格 v1

## 1. 模块范围

覆盖：

```text
家庭申请公开展示
平台管理员代家庭申请公开
后台审核公开申请
后台下架公开家庭
公开主页
游客 / 用户留言
留言审核
留言公开展示
留言删除
```

涉及表：

```text
families
family_public_applications
visitor_messages
operation_logs
admin_users
family_member_user_links
```

## 2. 核心规则

```text
搜索和公开展示分离
公开申请需要审核
PLATFORM_ADMIN 可以审核公开申请，但不能审核自己发起的申请
同一 family 同时只能有一个 PENDING 公开申请
公开资料修改后直接生效并写日志
下架公开家庭只更新 families.public_display_status = TAKEN_DOWN，不单独建下架表
游客可以匿名留言，留言需审核后展示
```

## 3. 枚举

```text
families.public_display_status: PRIVATE / PENDING / APPROVED / REJECTED / TAKEN_DOWN
family_public_applications.application_status: PENDING / APPROVED / REJECTED / CANCELLED
visitor_messages.status: PENDING / APPROVED / REJECTED / DELETED
```

## 4. POST /api/families/{familyId}/public-applications

提交公开申请。

### 权限

```text
FOUNDER
FAMILY_ADMIN
ROOT_ADMIN
SUPER_ADMIN
PLATFORM_ADMIN 可代申请
```

### 请求

```json
{
  "applicationReason": "希望公开展示家族资料，方便族人查找"
}
```

### 影响

```text
INSERT family_public_applications
families.public_display_status = PENDING
families.public_applied_at = 当前时间
写 operation_logs
不触发 graph_version
```

## 5. GET /api/families/{familyId}/public-applications

家庭公开申请列表。FOUNDER / FAMILY_ADMIN / 后台管理员可看。

## 6. GET /api/admin/family-public-applications

后台公开申请列表。

筛选：

```text
status
keyword
applicantType
reviewedByAdminId
createdStart
createdEnd
page
pageSize
```

## 7. POST /api/admin/family-public-applications/{applicationId}/approve

审核通过公开申请。

### 权限

```text
ROOT_ADMIN
SUPER_ADMIN
PLATFORM_ADMIN
```

限制：

```text
PLATFORM_ADMIN 不能审核自己发起的申请。
```

影响：

```text
application_status = APPROVED
review_result = APPROVED
families.public_display_status = APPROVED
families.public_approved_at = 当前时间
写 operation_logs
```

## 8. POST /api/admin/family-public-applications/{applicationId}/reject

拒绝公开申请。

影响：

```text
application_status = REJECTED
review_result = REJECTED
families.public_display_status = REJECTED
写 operation_logs
```

## 9. POST /api/families/{familyId}/public-applications/{applicationId}/cancel

撤销公开申请。

权限：

```text
申请人本人
FOUNDER
```

影响：

```text
application_status = CANCELLED
families.public_display_status = PRIVATE
写 operation_logs
```

## 10. POST /api/admin/families/{familyId}/take-down-public

下架公开家庭。

权限：

```text
ROOT_ADMIN
SUPER_ADMIN
PLATFORM_ADMIN
```

影响：

```text
families.public_display_status = TAKEN_DOWN
families.public_taken_down_at = 当前时间
写 operation_logs
```

## 11. GET /api/public/families/{familyId}/profile

公开家庭主页。

条件：

```text
families.status = NORMAL
families.public_display_status = APPROVED
```

联系方式为空时返回 contactTip。

## 12. POST /api/public/families/{familyId}/visitor-messages

提交游客留言。

### 请求

```json
{
  "visitorName": "王先生",
  "visitorPhone": "13900000000",
  "visitorWechat": "wang_wechat",
  "messageContent": "您好，我想了解这个家族的信息。"
}
```

规则：

```text
允许匿名
联系方式非必填
messageContent 必填
长度 <= 1000
同一 IP 1 分钟最多 1 条
同一 IP 1 天最多 20 条
留言状态 PENDING
```

## 13. GET /api/public/families/{familyId}/visitor-messages

公开留言列表。

```text
只展示 APPROVED
created_at DESC
不公开 visitorPhone / visitorWechat
```

## 14. GET /api/admin/visitor-messages

后台留言列表，支持状态、关键词、家庭、留言人、审核人、时间筛选。

## 15. 留言审核接口

```http
POST   /api/admin/visitor-messages/{messageId}/approve
POST   /api/admin/visitor-messages/{messageId}/reject
DELETE /api/admin/visitor-messages/{messageId}
```

权限：

```text
ROOT_ADMIN
SUPER_ADMIN
PLATFORM_ADMIN
```

## 16. 错误码范围

```text
45000-45999
```
