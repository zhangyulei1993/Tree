# 04. API 接口分组总览 v2

## 1. 总体约定

普通用户端接口：

```http
/api
```

后台管理端接口：

```http
/api/admin
```

公开访问接口：

```http
/api/public
```

系统存在两类 token：

```text
普通用户 token：PC / H5 / 微信小程序
后台管理员 token：后台管理端
```

统一返回结构：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

分页结构：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "records": [],
    "page": 1,
    "pageSize": 20,
    "total": 100
  }
}
```

## 2. graph_version 规则

触发 `families.graph_version + 1` 的典型操作：

```text
新增成员
删除成员
恢复成员
新增关系
删除关系
修改成员关键字段
修改字辈
修改 parent_link_type
修改人工承继标记
```

通常不触发：

```text
用户绑定 / 解绑
邀请接受 / 拒绝
公开申请
留言
角色变更
账号合并
账号认领
创始人转让
家庭解散
```

## 3. 接口分组

### 3.1 认证与账号

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

### 3.2 用户个人中心

```http
GET  /api/users/me
PUT  /api/users/me
GET  /api/users/me/families
GET  /api/users/me/invitations
GET  /api/users/me/join-requests
GET  /api/users/me/operation-logs
```

### 3.3 家庭基础

```http
POST /api/families
GET  /api/families
GET  /api/families/{familyId}
PUT  /api/families/{familyId}
GET  /api/public/families/{familyId}/profile
GET  /api/families/{familyId}/tree
GET  /api/public/families/{familyId}/tree
```

### 3.4 家庭成员与关系

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

### 3.5 邀请与加入家庭

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

### 3.6 家庭公开与留言

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

### 3.7 家庭角色、转让与解散

```http
POST /api/families/{familyId}/members/{memberId}/set-admin
POST /api/families/{familyId}/members/{memberId}/unset-admin

POST /api/families/{familyId}/founder-transfer-requests
GET  /api/families/{familyId}/founder-transfer-requests
GET  /api/admin/founder-transfer-requests
POST /api/admin/founder-transfer-requests/{requestId}/approve
POST /api/admin/founder-transfer-requests/{requestId}/reject
POST /api/families/{familyId}/founder-transfer-requests/{requestId}/cancel

POST /api/families/{familyId}/dissolution-requests
GET  /api/families/{familyId}/dissolution-requests
GET  /api/admin/dissolution-requests
POST /api/admin/dissolution-requests/{requestId}/approve
POST /api/admin/dissolution-requests/{requestId}/reject
POST /api/families/{familyId}/dissolution-requests/{requestId}/cancel
POST /api/admin/families/{familyId}/restore
```

### 3.8 后台管理

```http
POST /api/admin/auth/login
POST /api/admin/auth/logout
GET  /api/admin/me
PUT  /api/admin/me/password

GET    /api/admin/admin-users
POST   /api/admin/admin-users
PUT    /api/admin/admin-users/{adminId}
POST   /api/admin/admin-users/{adminId}/disable
DELETE /api/admin/admin-users/{adminId}
POST   /api/admin/admin-users/{adminId}/unlock

GET  /api/admin/users
GET  /api/admin/users/{userId}
PUT  /api/admin/users/{userId}
POST /api/admin/users/{userId}/disable
POST /api/admin/users/{userId}/restore
POST /api/admin/users/pre-create

GET    /api/admin/families
GET    /api/admin/families/{familyId}
PUT    /api/admin/families/{familyId}
POST   /api/admin/families/{familyId}/disable
DELETE /api/admin/families/{familyId}
POST   /api/admin/families/{familyId}/restore
```

### 3.9 操作日志

```http
GET /api/admin/operation-logs
GET /api/admin/operation-logs/{logId}
```
