# Family Member API

本模块所有接口均复用 USER token。`family_members` 是家谱节点，`users` 是平台账号，两者通过 `family_member_user_links` 绑定。

## Member Response

```json
{
  "memberId": 5,
  "familyId": 2,
  "name": "张明远",
  "gender": "MALE",
  "birthDate": "1992-01-01",
  "birthYear": 1992,
  "deathDate": null,
  "deathYear": null,
  "isAlive": true,
  "avatarUrl": "https://example.com/avatar.png",
  "description": "成员简介",
  "status": "ACTIVE",
  "userBindingPolicy": "OPTIONAL",
  "boundUserId": 10,
  "boundFamilyRole": "MEMBER",
  "createdAt": "2026-06-06T16:39:14Z",
  "updatedAt": "2026-06-06T16:39:14Z"
}
```

可选字段为空时因 `omitempty` 可能不出现在响应中。响应不会返回手机号、密码 hash、openid 或 unionid。

M2 表没有独立的 member `avatar_url` 和 `description` 列。当前 M8 将这两个 API 字段编码为 JSON 存入 `lineage_note`，并将 `lineage_note_type` 标记为 `M8_PROFILE_JSON`。两项合计编码后不能超过现有列的 500 字节限制。

## POST /api/families/{familyId}/members

- Authorization：是
- Token：`USER`
- 权限：`FOUNDER` 或 `FAMILY_ADMIN`
- HTTP success：`201`

| 字段 | 类型 | 必填 |
|---|---|---|
| `name` | string | 是 |
| `gender` | string | 否，`MALE` / `FEMALE` / `UNKNOWN` |
| `birthDate` | string | 否，`YYYY-MM-DD` |
| `birthYear` | integer | 否，未传 birthDate 时转为当年 `01-01` |
| `deathDate` | string | 否，`YYYY-MM-DD` |
| `deathYear` | integer | 否，未传 deathDate 时转为当年 `01-01` |
| `isAlive` | boolean | 否 |
| `avatarUrl` | string | 否 |
| `description` | string | 否 |
| `userBindingPolicy` | string | 否，`OPTIONAL` / `REQUIRED` / `NOT_REQUIRED` |

请求：

```json
{
  "name": "张明远",
  "gender": "MALE",
  "birthYear": 1992,
  "isAlive": true,
  "avatarUrl": "https://example.com/avatar.png",
  "description": "成员简介"
}
```

成功：

```json
{"code":0,"message":"success","data":{"memberId":5,"familyId":2,"name":"张明远","gender":"MALE","birthDate":"1992-01-01","birthYear":1992,"isAlive":true,"status":"ACTIVE","userBindingPolicy":"OPTIONAL","createdAt":"2026-06-06T16:39:14Z","updatedAt":"2026-06-06T16:39:14Z"}}
```

创建成功后 `families.graph_version + 1`。

常见错误：

```json
{"code":43003,"message":"无权创建家庭成员","data":null}
```

## GET /api/families/{familyId}/members

- Authorization：是
- Token：`USER`
- 权限：该家庭 ACTIVE member
- Request JSON：无

成功：

```json
{"code":0,"message":"success","data":[{"memberId":3,"familyId":2,"name":"创建者","gender":"UNKNOWN","status":"ACTIVE","userBindingPolicy":"OPTIONAL","boundUserId":8,"boundFamilyRole":"FOUNDER","createdAt":"2026-06-06T06:13:38Z","updatedAt":"2026-06-06T06:13:38Z"}]}
```

常见错误：

```json
{"code":43104,"message":"无权查看家庭成员","data":null}
```

## GET /api/families/{familyId}/members/{memberId}

- Authorization：是
- Token：`USER`
- 权限：该家庭 ACTIVE member
- Request JSON：无

成功响应为单个 Member Response。

常见错误：

```json
{"code":43101,"message":"成员不存在","data":null}
```

## PUT /api/families/{familyId}/members/{memberId}

- Authorization：是
- Token：`USER`
- 权限：`FOUNDER` 或 `FAMILY_ADMIN`

字段与创建接口相同，但全部可选。传入空 `name` 会返回 `43004`。将 `userBindingPolicy` 改为 `NOT_REQUIRED` 时，如 member 已有 ACTIVE user link，会返回 `43401`。

请求：

```json
{"name":"张明远（更新）","description":"更新后的简介"}
```

成功返回更新后的 Member Response。存在实际更新字段时 `families.graph_version + 1`。

常见错误：

```json
{"code":43103,"message":"无权编辑成员","data":null}
```

## DELETE /api/families/{familyId}/members/{memberId}

- Authorization：是
- Token：`USER`
- 权限：`FOUNDER` 或 `FAMILY_ADMIN`

可不传请求体，也可传：

| 字段 | 类型 | 必填 |
|---|---|---|
| `reason` | string | 否 |

```json
{"reason":"重复节点"}
```

成功：

```json
{"code":0,"message":"success","data":{"status":"ok"}}
```

删除采用软删除：member `status=DELETED` 并写入 `deleted_at`。如普通 MEMBER link 存在，会同时标记为 `UNLINKED`。当前实现对任何 ACTIVE relationship 都拒绝删除，不执行关系级联。FOUNDER 和 FAMILY_ADMIN 节点不能直接删除。成功后 `graph_version + 1`。

常见错误：

```json
{"code":43201,"message":"该成员存在关联关系，不能删除","data":null}
```

## POST /api/families/{familyId}/members/{memberId}/bind-user

- Authorization：是
- Token：`USER`
- 权限：`FOUNDER` 或 `FAMILY_ADMIN`

| 字段 | 类型 | 必填 |
|---|---|---|
| `userId` | integer | 是 |

请求：

```json
{"userId":10}
```

成功返回 Member Response，并包含 `boundUserId` 和 `boundFamilyRole=MEMBER`。

目标 user 必须存在且为 `ACTIVE`。同一个 member 只能有一个 ACTIVE link；同一个 user 在同一 family 只能绑定一个 ACTIVE member；`NOT_REQUIRED` member 不能绑定。绑定不更新 graph version。

常见错误：

```json
{"code":43404,"message":"当前成员已绑定用户","data":null}
```

## POST /api/families/{familyId}/members/{memberId}/unbind-user

- Authorization：是
- Token：`USER`
- 权限：`FOUNDER` 或 `FAMILY_ADMIN`

可不传请求体，也可传：

| 字段 | 类型 | 必填 |
|---|---|---|
| `reason` | string | 否 |

```json
{"reason":"绑定错误"}
```

成功返回解绑后的 Member Response，不再包含 `boundUserId`。FOUNDER 或 FAMILY_ADMIN link 不能直接解绑。解绑不更新 graph version。

常见错误：

```json
{"code":43407,"message":"创始人或家庭管理员不能直接解绑","data":null}
```

## M8 Error Codes

| Code | Message |
|---:|---|
| 43002 | 家庭不存在或状态不允许操作 |
| 43003 | 无权创建家庭成员 |
| 43004 | 成员姓名不能为空 |
| 43005 | 成员字段错误 |
| 43101 | 成员不存在 |
| 43103 | 无权编辑成员 |
| 43104 | 无权查看家庭成员 |
| 43201 | 该成员存在关联关系，不能删除 |
| 43203 | 家庭创始人或管理员成员不能直接删除 |
| 43204 | 无权删除成员 |
| 43401 | 成员绑定策略冲突 |
| 43403 | 目标用户不存在或状态不可用 |
| 43404 | 当前成员已绑定用户 |
| 43405 | 目标用户已在该家庭绑定成员 |
| 43406 | 当前成员未绑定用户 |
| 43407 | 创始人或家庭管理员不能直接解绑 |
