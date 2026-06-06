# Family Core API

## POST /api/families

- Authorization：是
- Token：`USER`
- HTTP success：`201`

| 字段 | 类型 | 必填 |
|---|---|---|
| `surname` | string | 是 |
| `familyName` | string/null | 否 |
| `name` | string/null | 否，`familyName` 的兼容别名 |
| `nativePlace` | string/null | 否 |
| `regionCode` | string/null | 否 |
| `regionText` | string/null | 否 |
| `description` | string/null | 否 |
| `avatarUrl` | string/null | 否 |
| `publicContact` | object/null | 否 |
| `publicContact.name` | string/null | 否 |
| `publicContact.phone` | string/null | 否 |
| `publicContact.wechat` | string/null | 否 |
| `publicContact.note` | string/null | 否 |
| `publicContact.visible` | boolean/null | 否 |

请求：

```json
{
  "surname": "张",
  "familyName": "张氏家族",
  "nativePlace": "山东济南",
  "publicContact": {
    "name": "联系人",
    "phone": "13800138000",
    "visible": true
  }
}
```

成功：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 2,
    "familyName": "张氏家族",
    "familySurname": "张",
    "nativePlace": "山东济南",
    "status": "NORMAL",
    "searchable": true,
    "publicDisplayStatus": "PRIVATE",
    "publicContactName": "联系人",
    "publicContactPhone": "13800138000",
    "publicContactVisible": true,
    "currentFounderMemberId": 3,
    "graphVersion": 1,
    "role": "FOUNDER"
  }
}
```

只允许数据库中 `ACTIVE` 且 `phone_verified=true` 的用户创建。家庭、创建者占位 member、FOUNDER link 和操作日志在同一事务中创建。

常见错误：

```json
{"code":42006,"message":"当前账号状态不允许创建家庭","data":null}
```

## GET /api/families

- Authorization：是
- Token：`USER`
- Request JSON：无

返回当前用户拥有 ACTIVE member link 的家庭，不是公开搜索接口。

成功：

```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "id": 2,
      "familyName": "张氏家族",
      "familySurname": "张",
      "status": "NORMAL",
      "publicDisplayStatus": "PRIVATE",
      "role": "FOUNDER"
    }
  ]
}
```

常见错误：

```json
{"code":10008,"message":"登录已过期","data":null}
```

## GET /api/families/{familyId}

- Authorization：是
- Token：`USER`
- Path `familyId`：integer，必填

需要当前用户是该家庭 ACTIVE member。

成功：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 2,
    "familyName": "张氏家族",
    "familySurname": "张",
    "status": "NORMAL",
    "searchable": true,
    "publicDisplayStatus": "PRIVATE",
    "publicContactVisible": true,
    "currentFounderMemberId": 3,
    "graphVersion": 1,
    "role": "FOUNDER"
  }
}
```

常见错误：

```json
{"code":42104,"message":"无权查看家庭详情","data":null}
```

## GET /api/families/{familyId}/public

- Authorization：否
- Token：`NONE`
- Path `familyId`：integer，必填

仅查询 `status=NORMAL` 且 `public_display_status=APPROVED` 的家庭。联系方式仅在 `publicContactVisible=true` 时返回。

成功：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 2,
    "familyName": "张氏家族",
    "familySurname": "张",
    "description": "家族介绍",
    "publicContactName": "联系人",
    "publicContactVisible": true
  }
}
```

PRIVATE 或其他不可公开状态：

```json
{"code":42103,"message":"家庭公开信息不可访问","data":null}
```

## PUT /api/families/{familyId}

- Authorization：是
- Token：`USER`
- Path `familyId`：integer，必填
- 权限：`FOUNDER` 或 `FAMILY_ADMIN`

所有请求字段均可选：

| 字段 | 类型 |
|---|---|
| `familyName` | string/null |
| `nativePlace` | string/null |
| `regionCode` | string/null |
| `regionText` | string/null |
| `description` | string/null |
| `avatarUrl` | string/null |
| `searchable` | boolean/null |
| `publicContactName` | string/null |
| `publicContactPhone` | string/null |
| `publicContactWechat` | string/null |
| `publicContactNote` | string/null |
| `publicContactVisible` | boolean/null |

请求：

```json
{"familyName":"张氏家族新名称","publicContactVisible":false}
```

成功返回更新后的 `FamilyDetail`：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 2,
    "familyName": "张氏家族新名称",
    "familySurname": "张",
    "status": "NORMAL",
    "searchable": true,
    "publicDisplayStatus": "PRIVATE",
    "publicContactVisible": false,
    "currentFounderMemberId": 3,
    "graphVersion": 1,
    "role": "FOUNDER"
  }
}
```

`familySurname`、`status`、`publicDisplayStatus`、`currentFounderMemberId`、`graphVersion` 当前均不能通过该接口修改。

常见错误：

```json
{"code":42201,"message":"无权修改家庭信息","data":null}
```

## POST /api/families/{familyId}/dissolution-requests

- Authorization：是
- Token：`USER`
- 权限：`FOUNDER` 或 `FAMILY_ADMIN`

| 字段 | 类型 | 必填 |
|---|---|---|
| `requestReason` | string/null | 否 |

请求：

```json
{"requestReason":"停止维护"}
```

成功：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "familyId": 2,
    "requesterMemberId": 3,
    "requesterUserId": 8,
    "requestStatus": "PENDING",
    "requestReason": "停止维护",
    "createdAt": "2026-06-06T06:13:38Z"
  }
}
```

创建后家庭状态改为 `DISSOLUTION_PENDING`。同一家庭只能有一条 PENDING 申请。

常见错误：

```json
{"code":46303,"message":"家庭已有待处理的解散申请","data":null}
```

## GET /api/families/{familyId}/dissolution-requests/current

- Authorization：是
- Token：`USER`
- 权限：家庭 ACTIVE member

有申请时返回 `DissolutionRequest`。没有 PENDING 申请时：

```json
{"code":0,"message":"success","data":null}
```

存在申请时：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "familyId": 2,
    "requesterMemberId": 3,
    "requesterUserId": 8,
    "requestStatus": "PENDING",
    "createdAt": "2026-06-06T06:13:38Z"
  }
}
```

常见错误：

```json
{"code":42104,"message":"无权查看家庭详情","data":null}
```

## POST /api/families/{familyId}/dissolution-requests/{requestId}/cancel

- Authorization：是
- Token：`USER`
- 权限：申请人或家庭管理员

| 字段 | 类型 | 必填 |
|---|---|---|
| `cancelReason` | string/null | 否（DTO 未标 required） |

请求：

```json
{"cancelReason":"继续维护家庭"}
```

成功：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "familyId": 2,
    "requesterMemberId": 3,
    "requesterUserId": 8,
    "requestStatus": "CANCELLED",
    "cancelReason": "继续维护家庭",
    "cancelledAt": "2026-06-06T06:17:14Z",
    "createdAt": "2026-06-06T06:13:38Z"
  }
}
```

取消后家庭状态恢复 `NORMAL`。

常见错误：

```json
{"code":10004,"message":"数据状态不允许操作","data":null}
```
