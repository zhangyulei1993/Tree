# Family Role Transfer And Dissolution API

M13 实现家庭角色管理、创始人转让、家庭解散后台审核和家庭恢复。

M13 不增加 `families.graph_version`，不物理删除 family、member、relationship 或 link 数据。

## Family Role

### POST /api/families/{familyId}/members/{memberId}/set-admin

- Token：`USER`
- 权限：当前用户必须是该 family 的 `FOUNDER`
- Request JSON：可为空，或 `{"reason":"协助管理"}`

成功后目标成员 ACTIVE link 的 `family_role` 变为 `FAMILY_ADMIN`，写入 `role_granted_at` 和 `role_granted_by_user_id`。

### POST /api/families/{familyId}/members/{memberId}/unset-admin

- Token：`USER`
- 权限：当前用户必须是该 family 的 `FOUNDER`
- Request JSON：可为空，或 `{"reason":"取消管理员"}`

目标成员当前必须是 `FAMILY_ADMIN`。不能 unset `FOUNDER`。

## Founder Transfer

### POST /api/families/{familyId}/founder-transfer-requests

- Token：`USER`
- 权限：当前用户必须是 `FOUNDER`

```json
{"toMemberId": 4, "requestReason": "转让给新的维护人"}
```

目标成员必须 ACTIVE、非 `NOT_REQUIRED`，且存在 ACTIVE `family_member_user_links`。同一 family 不允许重复 `PENDING` 转让申请。

### POST /api/families/{familyId}/founder-transfer-requests/{requestId}/cancel

- Token：`USER`
- 权限：申请发起人或当前 `FOUNDER`
- Request JSON：可为空，或 `{"cancelReason":"暂不转让"}`

### GET /api/admin/founder-transfer-requests

- Token：`ADMIN`
- 角色：`ROOT_ADMIN` / `SUPER_ADMIN`
- Query：`status`、`familyId`、`page`、`pageSize`

### POST /api/admin/founder-transfer-requests/{requestId}/approve

- Token：`ADMIN`
- 角色：`ROOT_ADMIN` / `SUPER_ADMIN`
- 明确禁止：`PLATFORM_ADMIN`

事务内完成：

- 原 `FOUNDER` link 降为 `MEMBER`
- 目标 link 升为 `FOUNDER`
- `families.current_founder_member_id = to_member_id`
- request 更新为 `APPROVED`
- 保证同一 family 只有一个 ACTIVE `FOUNDER`

### POST /api/admin/founder-transfer-requests/{requestId}/reject

- Token：`ADMIN`
- 角色：`ROOT_ADMIN` / `SUPER_ADMIN`
- Request JSON：可为空，或 `{"reviewComment":"原因"}`。

拒绝不修改当前 founder。

## Family Dissolution

用户侧 M7 接口保留：

- `POST /api/families/{familyId}/dissolution-requests`
- `GET /api/families/{familyId}/dissolution-requests/current`
- `POST /api/families/{familyId}/dissolution-requests/{requestId}/cancel`

M13 将创建解散申请限制为当前用户必须是 `FOUNDER`。创建后 family 进入 `DISSOLUTION_PENDING`，取消后回到 `NORMAL`。

### GET /api/admin/dissolution-requests

- Token：`ADMIN`
- 角色：`ROOT_ADMIN` / `SUPER_ADMIN`
- Query：`status`、`familyId`、`page`、`pageSize`

### POST /api/admin/dissolution-requests/{requestId}/approve

- Token：`ADMIN`
- 角色：`ROOT_ADMIN` / `SUPER_ADMIN`
- 明确禁止：`PLATFORM_ADMIN`

事务内完成：

- request 更新为 `APPROVED`
- `families.status = DISSOLVED`
- `families.dissolved_at = now`
- `families.searchable = false`
- `families.public_display_status = PRIVATE`

### POST /api/admin/dissolution-requests/{requestId}/reject

- Token：`ADMIN`
- 角色：`ROOT_ADMIN` / `SUPER_ADMIN`

拒绝后 request 为 `REJECTED`，family 回到 `NORMAL`。

## Family Restore

### POST /api/admin/families/{familyId}/restore

- Token：`ADMIN`
- 角色：`ROOT_ADMIN` / `SUPER_ADMIN`
- 明确禁止：`PLATFORM_ADMIN`

Request JSON 可为空：

```json
{"searchable": true, "restoreReason": "确认恢复"}
```

事务内完成：

- `families.status = NORMAL`
- `families.public_display_status = PRIVATE`
- `families.restored_at = now`
- `families.searchable = true`，除非请求显式传 `false`

不会重新激活历史 invitation、join request、public application 或 visitor message。

## Operation Logs

必须记录：

- `SET_FAMILY_ADMIN`
- `UNSET_FAMILY_ADMIN`
- `CREATE_FOUNDER_TRANSFER_REQUEST`
- `CANCEL_FOUNDER_TRANSFER_REQUEST`
- `APPROVE_FOUNDER_TRANSFER_REQUEST`
- `REJECT_FOUNDER_TRANSFER_REQUEST`
- `CREATE_DISSOLUTION_REQUEST`
- `CANCEL_DISSOLUTION_REQUEST`
- `APPROVE_DISSOLUTION_REQUEST`
- `REJECT_DISSOLUTION_REQUEST`
- `RESTORE_FAMILY`

错误码见 [error-codes.md](./error-codes.md) 的 `46001-46207`，以及兼容保留的 M7 `46301-46303`。
