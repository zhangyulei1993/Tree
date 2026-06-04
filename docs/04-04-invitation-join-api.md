# 04-04. 邀请与加入家庭接口完整规格 v1

## 1. 模块范围

覆盖：

```text
平台管理员站内邀请
家庭创始人 / 家族管理员分享链接邀请
查看邀请详情
接受邀请
拒绝邀请
撤销邀请
用户主动申请加入家庭
处理加入申请
```

涉及表：

```text
family_invitations
family_join_requests
family_members
family_member_user_links
families
users
operation_logs
```

## 2. 核心规则

```text
family_invitations = 家庭侧主动邀请用户绑定成员节点
family_join_requests = 用户主动申请加入家庭
邀请必须绑定到具体 target_member_id
邀请接受后默认 family_role = MEMBER
邀请有效期 7 天，访问时实时判断过期
```

平台管理员：

```text
手机号搜索用户池
找到已注册 user -> 站内邀请
找不到 user -> 不能发邀请，只能选择无需绑定用户
不能生成分享邀请链接
```

家庭创始人 / 家族管理员：

```text
生成分享链接
不要求提前找到 user
用户接受前必须登录 / 注册 / 绑定手机号
```

## 3. 枚举

```text
invite_type: CLAIM_EXISTING_MEMBER / CREATE_AND_BIND_MEMBER / JOIN_FAMILY
invite_channel: IN_APP / SHARE_LINK
invite_actor_type: PLATFORM_ADMIN / FAMILY_FOUNDER / FAMILY_ADMIN
invitation.status: PENDING / ACCEPTED / REJECTED / CANCELLED / EXPIRED / INVALID
```

当前一期主要使用：

```text
invite_type = CLAIM_EXISTING_MEMBER
family_role_after_accept = MEMBER
```

## 4. POST /api/families/{familyId}/members/{memberId}/invite

创建邀请。

### 权限

```text
FOUNDER
FAMILY_ADMIN
ROOT_ADMIN
SUPER_ADMIN
PLATFORM_ADMIN
```

### 家庭端请求

```json
{
  "inviteChannel": "SHARE_LINK",
  "inviteMessage": "请确认是否绑定为该家庭成员",
  "familyRoleAfterAccept": "MEMBER"
}
```

### 平台端请求

```json
{
  "inviteChannel": "IN_APP",
  "targetPhone": "13800000000",
  "targetUserId": 10001,
  "inviteMessage": "平台协助邀请用户认领成员节点",
  "familyRoleAfterAccept": "MEMBER"
}
```

### 通用校验

```text
family.status = NORMAL
target member ACTIVE
target member.user_binding_policy != NOT_REQUIRED
target member 无 ACTIVE user 绑定
target member 不存在 PENDING invitation
操作者有权限
```

## 5. GET /api/invitations/{inviteToken}

未登录也可查看基础邀请详情。

接受前必须登录并绑定手机号。

## 6. POST /api/invitations/{invitationId}/accept

接受邀请。

### 校验

```text
invitation.status = PENDING
未过期
family.status = NORMAL
target_member ACTIVE
target_member 无 ACTIVE 绑定
当前 user 在该 family 无 ACTIVE 绑定
当前 user.status = ACTIVE
当前 user.phone_verified = 1
IN_APP 邀请必须由 target_user_id 指定用户接受
```

### 影响

```text
invitation.status = ACCEPTED
创建 family_member_user_links
link_status = ACTIVE
link_source = INVITATION_ACCEPTED
family_role = MEMBER
写 operation_logs
不触发 graph_version
```

## 7. POST /api/invitations/{invitationId}/reject

拒绝邀请，拒绝后可重新邀请。

## 8. POST /api/invitations/{invitationId}/cancel

撤销邀请。

```text
FAMILY_ADMIN 只能撤销自己发出的邀请
FOUNDER 可以撤销本家庭所有未接受邀请
后台管理员可协助撤销
已 ACCEPTED 邀请不能撤销，只能走解绑流程
```

## 9. GET /api/users/me/invitations

查询我的邀请。站内邀请可直接出现在收到的邀请中；分享链接邀请通常通过 inviteToken 查看。

## 10. POST /api/families/{familyId}/join-requests

用户主动申请加入家庭。

### 权限

```text
已登录
users.status = ACTIVE
phone_verified = 1
```

### 请求

```json
{
  "applicantMessage": "我是张某某，申请加入本家庭",
  "applicantRealName": "张明远"
}
```

申请说明非必填。

### 校验

```text
family.status = NORMAL
当前 user 未加入该 family
不存在 PENDING join request
```

## 11. GET /api/families/{familyId}/join-requests

家庭加入申请列表。

权限：

```text
FOUNDER
FAMILY_ADMIN
ROOT_ADMIN
SUPER_ADMIN
PLATFORM_ADMIN
```

家庭管理员处理为主，平台协助。

## 12. POST /api/families/{familyId}/join-requests/{requestId}/approve

通过加入申请。

### 模式

```text
BIND_EXISTING_MEMBER
CREATE_NEW_MEMBER
```

绑定已有成员不触发 graph_version；创建新成员触发 graph_version + 1。

通过后创建 family_member_user_links，family_role = MEMBER。

## 13. POST /api/families/{familyId}/join-requests/{requestId}/reject

拒绝加入申请，拒绝后允许重新申请。

## 14. POST /api/families/{familyId}/join-requests/{requestId}/cancel

申请人撤销加入申请。

## 15. 错误码范围

```text
44000-44999
```
