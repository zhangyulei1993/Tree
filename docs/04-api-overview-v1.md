# 04. API 接口分组总览 v1

> 当前文档是接口分组总览，不是最终完整接口规格。  
> 下一阶段需要逐个接口补充：权限、请求参数、请求示例、返回示例、错误码、数据库影响、是否触发 graph_version。

## 1. 认证与账号接口

```http
POST /api/auth/register-phone
POST /api/auth/login-phone
POST /api/auth/logout
POST /api/auth/send-code
POST /api/auth/verify-code
POST /api/auth/wechat-mini/login
POST /api/auth/wechat-mini/bind-phone
POST /api/auth/change-phone
POST /api/auth/cancel-account
```

## 2. 用户账号接口

```http
GET  /api/users/me
PUT  /api/users/me
GET  /api/users/me/families
GET  /api/users/me/invitations
GET  /api/users/me/join-requests
GET  /api/users/me/operation-logs
```

## 3. 家庭基础接口

```http
POST /api/families
GET  /api/families
GET  /api/families/{familyId}
PUT  /api/families/{familyId}
POST /api/families/{familyId}/apply-public
POST /api/families/{familyId}/apply-dissolution
GET  /api/families/{familyId}/public-profile
GET  /api/families/{familyId}/tree
```

## 4. 家庭树接口

```http
GET /api/families/{familyId}/tree
```

一期规则：

```text
只做列表树 / 层级树
不做复杂图谱
不显示动态称谓
返回 nodes + edges + tree
配偶作为附属信息显示
删除节点默认隐藏
保留 canExpand / stopReason
返回 userBindingState
```

`userBindingState`：

```text
BOUND
INVITING
NOT_REQUIRED
UNBOUND
```

## 5. 家庭成员接口

```http
POST   /api/families/{familyId}/members
GET    /api/families/{familyId}/members/{memberId}
PUT    /api/families/{familyId}/members/{memberId}
DELETE /api/families/{familyId}/members/{memberId}
POST   /api/families/{familyId}/members/{memberId}/mark-not-required
POST   /api/families/{familyId}/members/{memberId}/restore
```

## 6. 成员关系接口

```http
POST   /api/families/{familyId}/relationships
DELETE /api/families/{familyId}/relationships/{relationshipId}
PUT    /api/families/{familyId}/relationships/{relationshipId}
```

一期关系类型：

```text
PARENT_CHILD
SPOUSE
```

用户操作类型：

```text
ADD_FATHER
ADD_MOTHER
ADD_CHILD
ADD_SIBLING
ADD_SPOUSE
```

## 7. 邀请接口

```http
POST /api/families/{familyId}/members/{memberId}/invite
GET  /api/invitations/{inviteToken}
POST /api/invitations/{invitationId}/accept
POST /api/invitations/{invitationId}/reject
POST /api/invitations/{invitationId}/cancel
GET  /api/users/me/invitations
```

## 8. 申请加入家庭接口

```http
POST /api/families/{familyId}/join-requests
GET  /api/users/me/join-requests
GET  /api/families/{familyId}/join-requests
POST /api/families/{familyId}/join-requests/{requestId}/approve
POST /api/families/{familyId}/join-requests/{requestId}/reject
POST /api/families/{familyId}/join-requests/{requestId}/cancel
```

## 9. 家庭公开申请接口

```http
POST /api/families/{familyId}/public-applications
GET  /api/families/{familyId}/public-applications
POST /api/admin/family-public-applications/{applicationId}/approve
POST /api/admin/family-public-applications/{applicationId}/reject
POST /api/families/{familyId}/public-applications/{applicationId}/cancel
POST /api/admin/families/{familyId}/take-down-public
```

## 10. 游客留言接口

```http
POST /api/families/{familyId}/visitor-messages
GET  /api/families/{familyId}/visitor-messages
GET  /api/admin/visitor-messages
POST /api/admin/visitor-messages/{messageId}/approve
POST /api/admin/visitor-messages/{messageId}/reject
DELETE /api/admin/visitor-messages/{messageId}
```

## 11. 家庭角色接口

```http
POST /api/families/{familyId}/members/{memberId}/set-admin
POST /api/families/{familyId}/members/{memberId}/unset-admin
POST /api/families/{familyId}/founder-transfer-requests
GET  /api/families/{familyId}/founder-transfer-requests
POST /api/admin/founder-transfer-requests/{requestId}/approve
POST /api/admin/founder-transfer-requests/{requestId}/reject
POST /api/families/{familyId}/founder-transfer-requests/{requestId}/cancel
```

## 12. 家庭解散接口

```http
POST /api/families/{familyId}/dissolution-requests
GET  /api/families/{familyId}/dissolution-requests
POST /api/admin/dissolution-requests/{requestId}/approve
POST /api/admin/dissolution-requests/{requestId}/reject
POST /api/families/{familyId}/dissolution-requests/{requestId}/cancel
POST /api/admin/families/{familyId}/restore
```

## 13. 后台用户管理接口

```http
GET  /api/admin/users
GET  /api/admin/users/{userId}
PUT  /api/admin/users/{userId}
POST /api/admin/users/{userId}/disable
POST /api/admin/users/{userId}/restore
POST /api/admin/users/pre-create
```

## 14. 后台家庭管理接口

```http
GET  /api/admin/families
GET  /api/admin/families/{familyId}
PUT  /api/admin/families/{familyId}
POST /api/admin/families/{familyId}/disable
POST /api/admin/families/{familyId}/restore
DELETE /api/admin/families/{familyId}
```

## 15. 后台管理员接口

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
```

## 16. 操作日志接口

```http
GET /api/admin/operation-logs
```

一期支持筛选、搜索、分页、详情查看，不支持导出、删除日志、清空日志。
