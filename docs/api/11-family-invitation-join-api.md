# Family Invitation And Join Request API

M11 使用 USER token 完成家庭邀请和加入申请。接受邀请与批准申请均在事务中创建 ACTIVE `family_member_user_links`。

## Invitation

### POST /api/families/{familyId}/members/{memberId}/invite

- Token：`USER`
- 权限：`FOUNDER` / `FAMILY_ADMIN`
- 成功状态：`201`

分享邀请：

```json
{"inviteChannel":"SHARE_LINK","inviteMessage":"请确认家庭成员身份"}
```

站内邀请：

```json
{"inviteChannel":"IN_APP","targetUserId":9,"inviteMessage":"请确认家庭成员身份"}
```

原始 `inviteToken` 仅在本接口成功响应中返回一次。数据库保存 SHA-256 摘要，operation log 不记录原始 token。

### GET /api/invitations/{inviteToken}

- Token：`NONE`
- 请求体：无

返回家庭名称、目标成员显示名、邀请渠道、状态和过期时间，不返回手机号或账号敏感信息。

### POST /api/invitations/{invitationId}/accept

- Token：`USER`
- 请求体：无

当前用户必须为 ACTIVE 且已验证手机号。站内邀请只能由指定用户接受。成功后创建：

```text
link_status=ACTIVE
link_source=INVITATION_ACCEPTED
family_role=MEMBER
```

接受邀请不增加 `graph_version`。

### POST /api/invitations/{invitationId}/reject

- Token：`USER`
- 可选请求：`{"reason":"暂不加入"}`

### POST /api/invitations/{invitationId}/cancel

- Token：`USER`
- 权限：邀请创建者或当前家庭管理员
- 可选请求：`{"reason":"邀请信息有误"}`

### GET /api/users/me/invitations

- Token：`USER`
- 请求体：无
- 返回当前用户作为 `targetUserId` 的站内邀请

## Join Request

### POST /api/families/{familyId}/join-requests

- Token：`USER`
- 成功状态：`201`

```json
{"applicantRealName":"张某","applicantMessage":"申请加入家庭"}
```

用户必须 ACTIVE、已验证手机号、尚未绑定该家庭，且不能存在重复 PENDING 申请。

### GET /api/users/me/join-requests

- Token：`USER`
- 返回当前用户提交的申请

### GET /api/families/{familyId}/join-requests

- Token：`USER`
- 权限：`FOUNDER` / `FAMILY_ADMIN`

### POST /api/families/{familyId}/join-requests/{requestId}/approve

- Token：`USER`
- 权限：`FOUNDER` / `FAMILY_ADMIN`

绑定已有成员：

```json
{"approveMode":"BIND_EXISTING_MEMBER","memberId":6,"handleComment":"身份已确认"}
```

不增加 `graph_version`。

创建新成员：

```json
{
  "approveMode":"CREATE_NEW_MEMBER",
  "newMember":{"name":"张某","gender":"UNKNOWN","isAlive":true},
  "handleComment":"创建独立成员节点"
}
```

创建 ACTIVE member 和 ACTIVE link，不创建 relationship，`graph_version` 增加一次。

### POST /api/families/{familyId}/join-requests/{requestId}/reject

- Token：`USER`
- 权限：`FOUNDER` / `FAMILY_ADMIN`
- 可选请求：`{"handleComment":"身份信息不足"}`

### POST /api/families/{familyId}/join-requests/{requestId}/cancel

- Token：`USER`
- 权限：仅申请人本人
- 可选请求：`{"cancelReason":"不再申请"}`

## Shared Response

```json
{"code":0,"message":"success","data":{}}
```

错误码见 [error-codes.md](./error-codes.md) 的 44001-44107。
