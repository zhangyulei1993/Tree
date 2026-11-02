# Family Public Display And Visitor Message API

M12 实现家庭公开展示申请与游客留言。公开申请接口基于真实 router、DTO、VO 和 service 整理。

公开状态由 `families.public_display_status` 驱动：

- `APPROVED`：公开主页与公开树可访问。
- `PRIVATE` / `PENDING` / `REJECTED` / `CANCELLED` / `TAKEN_DOWN`：公开主页与公开树不可访问。

M12 不增加 `families.graph_version`。

## Public Application

### POST /api/families/{familyId}/public-applications

- Token：`USER`
- 权限：`FOUNDER` / `FAMILY_ADMIN`
- 成功状态：`201`
- Request JSON：

| Field | Type | Required | Description |
|---|---|---:|---|
| applicationReason | string | No | 公开展示申请说明 |

成功示例：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "applicationId": 1,
    "familyId": 2,
    "status": "PENDING",
    "createdAt": "2026-06-08T10:00:00Z"
  }
}
```

常见错误：`45003` 已有待审核公开申请，`45004` 家庭状态不允许公开申请，`45005` 无权操作公开申请，`45006` 家庭已公开。

### GET /api/families/{familyId}/public-applications

- Token：`USER`
- 权限：`FOUNDER` / `FAMILY_ADMIN`
- Query：`status`、`page`、`pageSize`
- Request JSON：无

返回当前家庭公开申请记录，按 `created_at` 倒序。

### POST /api/families/{familyId}/public-applications/{applicationId}/cancel

- Token：`USER`
- 权限：申请人本人或 `FOUNDER` / `FAMILY_ADMIN`
- Request JSON：

| Field | Type | Required | Description |
|---|---|---:|---|
| cancelReason | string | No | 取消原因 |

只能取消 `PENDING` 申请。成功后 `families.public_display_status` 回到 `PRIVATE`。

### GET /api/admin/family-public-applications

- Token：`ADMIN`
- 角色：`ROOT_ADMIN` / `SUPER_ADMIN` / `PLATFORM_ADMIN`
- Query：`status`、`page`、`pageSize`
- Request JSON：无

后台按 `created_at` 倒序查看公开申请列表。

### POST /api/admin/family-public-applications/{applicationId}/approve

- Token：`ADMIN`
- 角色：`ROOT_ADMIN` / `SUPER_ADMIN` / `PLATFORM_ADMIN`
- Request JSON：

| Field | Type | Required | Description |
|---|---|---:|---|
| reviewComment | string | No | 审核备注 |

只能审核 `PENDING` 申请。成功后 application 为 `APPROVED`，family 为 `APPROVED`，并写入 `public_approved_at`。

### POST /api/admin/family-public-applications/{applicationId}/reject

- Token：`ADMIN`
- 角色：`ROOT_ADMIN` / `SUPER_ADMIN` / `PLATFORM_ADMIN`
- Request JSON：

| Field | Type | Required | Description |
|---|---|---:|---|
| reviewComment | string | No | 审核备注 |

只能审核 `PENDING` 申请。成功后 application 为 `REJECTED`，family 为 `REJECTED`。

### POST /api/admin/families/{familyId}/take-down-public

- Token：`ADMIN`
- 角色：`ROOT_ADMIN` / `SUPER_ADMIN` / `PLATFORM_ADMIN`
- Request JSON：

| Field | Type | Required | Description |
|---|---|---:|---|
| reason | string | No | 下架原因；当前 DTO 接收但 service 不写入 family 字段 |

仅当前 `public_display_status=APPROVED` 的家庭可下架。成功后 family 为 `TAKEN_DOWN`，并写入 `public_taken_down_at`。

## Visitor Message

### POST /api/public/families/{familyId}/visitor-messages

- Token：`NONE`
- Request JSON：

| Field | Type | Required | Description |
|---|---|---:|---|
| visitorName | string | No | 游客称呼 |
| visitorPhone | string | No | 游客电话；保存到 `visitor_messages`，不写入 operation log |
| visitorWechat | string | No | 游客微信；保存到 `visitor_messages`，不写入 operation log |
| messageContent | string | Yes | 留言内容，1-1000 字符 |

要求 family `status=NORMAL` 且 `public_display_status=APPROVED`。同 IP 同 family 1 分钟最多 1 条，1 天最多 20 条。新留言状态为 `PENDING`。

### GET /api/public/families/{familyId}/visitor-messages

- Token：`NONE`
- Query：`page`、`pageSize`
- Request JSON：无

只返回 `APPROVED` 留言，不返回 `visitorPhone` / `visitorWechat`。

### GET /api/admin/visitor-messages

- Token：`ADMIN`
- 角色：`ROOT_ADMIN` / `SUPER_ADMIN` / `PLATFORM_ADMIN`
- Query：`status`、`familyId`、`page`、`pageSize`

后台可查看 `PENDING` / `APPROVED` / `REJECTED` / `DELETED` 留言。

### POST /api/admin/visitor-messages/{messageId}/approve

- Token：`ADMIN`
- Request JSON：`{"reviewComment":"内容合规"}`

`PENDING` 或 `REJECTED` 留言可通过，写入 `reviewed_by_admin_id` 和 `reviewed_at`。

### POST /api/admin/visitor-messages/{messageId}/reject

- Token：`ADMIN`
- Request JSON：`{"reviewComment":"内容不合适"}`

将留言状态更新为 `REJECTED`，写入审核字段。

### DELETE /api/admin/visitor-messages/{messageId}

- Token：`ADMIN`
- Request JSON：`{"deleteReason":"管理员删除"}`

将留言状态改为 `DELETED`，写入 `deleted_at` 和 `deleted_by_admin_id`。公开列表不再返回。

## Operation Logs

已写入：

- `SUBMIT_PUBLIC_APPLICATION`
- `CANCEL_PUBLIC_APPLICATION`
- `APPROVE_PUBLIC_APPLICATION`
- `REJECT_PUBLIC_APPLICATION`
- `TAKE_DOWN_PUBLIC_FAMILY`
- `APPROVE_VISITOR_MESSAGE`
- `REJECT_VISITOR_MESSAGE`
- `DELETE_VISITOR_MESSAGE`

游客提交留言当前不写 operation log，避免 visitorPhone / visitorWechat 进入日志。

## Error Codes

见 [error-codes.md](./error-codes.md) 的 `45001-45007` 和 `45101-45107`。
