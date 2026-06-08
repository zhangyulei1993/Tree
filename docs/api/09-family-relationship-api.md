# Family Relationship API

M9 家庭关系接口使用 USER token。只有家庭 `FOUNDER` 或
`FAMILY_ADMIN` 可以创建、更新或删除关系。

数据库 `relationship_type` 只保存 `PARENT_CHILD` 和 `SPOUSE`。
`ADD_SIBLING` 是操作类型，不是数据库关系类型。

## POST /api/families/{familyId}/relationships

- Authorization：是
- Token：`USER`
- 权限：`FOUNDER` / `FAMILY_ADMIN`
- HTTP success：`201`

请求字段：

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `baseMemberId` | integer | 是 | 已存在且属于当前家庭的 ACTIVE member |
| `addType` | string | 是 | `ADD_FATHER` / `ADD_MOTHER` / `ADD_CHILD` / `ADD_SPOUSE` / `ADD_SIBLING` |
| `newMember` | object | 是 | 要创建的新成员 |
| `newMember.name` | string | 是 | 新成员姓名 |
| `newMember.gender` | string | 否 | `MALE` / `FEMALE` / `UNKNOWN` |
| `newMember.birthDate` | string | 否 | `YYYY-MM-DD` |
| `newMember.birthYear` | integer | 否 | 未传 birthDate 时转换为当年 `01-01` |
| `newMember.deathDate` | string | 否 | `YYYY-MM-DD` |
| `newMember.deathYear` | integer | 否 | 未传 deathDate 时转换为当年 `01-01` |
| `newMember.isAlive` | boolean | 否 | 是否健在 |
| `newMember.userBindingPolicy` | string | 否 | `OPTIONAL` / `REQUIRED` / `NOT_REQUIRED` |
| `relationship.relationshipType` | string | 否 | 必须与 addType 对应 |
| `relationship.parentLinkType` | string | 否 | 仅 PARENT_CHILD；默认 `PRIMARY` |
| `relationship.relationNoteType` | string | 否 | 最长 80 字节 |
| `relationship.relationNote` | string | 否 | 最长 500 字节 |

`parentLinkType` 支持：

```text
PRIMARY
STEP
ADOPTIVE
SUCCESSION
NOTE_ONLY
OTHER
```

示例：

```json
{
  "baseMemberId": 3,
  "addType": "ADD_CHILD",
  "newMember": {
    "name": "张小明",
    "gender": "MALE",
    "birthYear": 2000,
    "isAlive": true
  },
  "relationship": {
    "relationshipType": "PARENT_CHILD",
    "parentLinkType": "PRIMARY",
    "relationNote": "亲生子女"
  }
}
```

成功响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "createdMember": {
      "memberId": 21,
      "familyId": 2,
      "name": "张小明",
      "gender": "MALE",
      "status": "ACTIVE"
    },
    "relationships": [
      {
        "relationshipId": 31,
        "familyId": 2,
        "fromMemberId": 3,
        "toMemberId": 21,
        "relationshipType": "PARENT_CHILD",
        "parentLinkType": "PRIMARY",
        "relationNote": "亲生子女",
        "status": "ACTIVE",
        "createdAt": "2026-06-07T12:00:00Z",
        "updatedAt": "2026-06-07T12:00:00Z"
      }
    ],
    "graphVersion": 5
  }
}
```

方向规则：

| addType | 落库方向 |
|---|---|
| `ADD_FATHER` | 新父亲 `PARENT_CHILD` -> base member |
| `ADD_MOTHER` | 新母亲 `PARENT_CHILD` -> base member |
| `ADD_CHILD` | base member `PARENT_CHILD` -> 新子女 |
| `ADD_SPOUSE` | base member `SPOUSE` -> 新配偶 |
| `ADD_SIBLING` | base member 的每个父母 `PARENT_CHILD` -> 新 sibling |

`ADD_SIBLING` 没有父母时返回 `43301`，且事务不会创建成员或关系。
成功时即使创建多条共同父母关系，`graph_version` 也只增加一次。

## PUT /api/families/{familyId}/relationships/{relationshipId}

- Authorization：是
- Token：`USER`
- 权限：`FOUNDER` / `FAMILY_ADMIN`

允许修改：

```json
{
  "parentLinkType": "STEP",
  "relationNoteType": "OTHER",
  "relationNote": "关系备注"
}
```

不允许传入或修改：

```text
fromMemberId
toMemberId
relationshipType
```

`SPOUSE` 关系不支持 `parentLinkType`。成功响应使用与创建接口相同的
mutation response，但不返回 `createdMember`，并将 `graphVersion + 1`。

## DELETE /api/families/{familyId}/relationships/{relationshipId}

- Authorization：是
- Token：`USER`
- 权限：`FOUNDER` / `FAMILY_ADMIN`

请求体可省略，也可传：

```json
{"reason":"关系录入错误"}
```

删除采用软删除：

```text
status = DELETED
deleted_at 设置当前时间
deleted_by_user_id 设置当前用户
graph_version + 1
```

成功响应中的 relationship 状态为 `DELETED`。

## Error Codes

| Code | Message |
|---:|---|
| 43301 | 请先创建父亲或母亲节点，再添加兄弟姐妹 |
| 43302 | 关系成员不存在或不属于该家庭 |
| 43303 | 当前关系已存在 |
| 43304 | PRIMARY 父亲已存在 |
| 43305 | PRIMARY 母亲已存在 |
| 43306 | 不支持的关系类型或修改 |
| 43307 | 无权修改家庭关系 |
| 43308 | 关系不存在 |
| 43309 | 不能与自己建立关系 |
| 43310 | 家庭不存在或状态不允许操作 |
