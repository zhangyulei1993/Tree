# Family Tree API

M10 提供家庭树只读组装接口。当前仅支持 `LIST_TREE`，不计算以中心人物为基准的动态称谓。

## Response Model

两个接口返回相同的公开安全结构：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "familyId": 2,
    "treeMode": "LIST_TREE",
    "graphVersion": 14,
    "nodes": [
      {
        "memberId": 6,
        "displayName": "张某",
        "gender": "MALE",
        "memberType": "LINEAGE_MEMBER",
        "userBindingState": "BOUND",
        "canExpand": true
      }
    ],
    "edges": [
      {
        "relationshipId": 1,
        "fromMemberId": 6,
        "toMemberId": 7,
        "relationshipType": "PARENT_CHILD",
        "parentLinkType": "PRIMARY"
      }
    ],
    "tree": [
      {
        "memberId": 6,
        "parentIds": [],
        "childrenIds": [7],
        "spouseIds": []
      }
    ]
  }
}
```

`nodes` 是权威节点列表，`edges` 是权威关系列表。`tree` 仅提供根节点及其 ID 引用，避免在多父母、多配偶或异常环路数据中递归复制节点。

节点可包含：

- `memberId`
- `displayName`
- `surname`
- `generationCharacter`
- `gender`
- `memberType`
- `birthDate`
- `deathDate`
- `isLiving`
- `userBindingState`
- `canExpand`
- `stopReason`

节点不返回 `userId`、`boundUserId`、手机号、token、openid、unionid、密码或家庭私有角色。边只返回 `PARENT_CHILD` 和 `SPOUSE`，绝不返回 `SIBLING`。

## GET /api/families/{familyId}/tree

- Authorization：是
- Token：`USER`
- 请求体：无
- 路径参数：`familyId`，正整数
- 权限：当前用户必须是该家庭的有效成员
- HTTP success：`200`

常见错误：

```json
{
  "code": 42401,
  "message": "无权查看家庭树",
  "data": null
}
```

未携带或携带无效 USER token 时返回认证错误。

## GET /api/public/families/{familyId}/tree

- Authorization：否
- Token：`NONE`
- 请求体：无
- 路径参数：`familyId`，正整数
- 公开条件：`family.status = NORMAL` 且 `public_display_status = APPROVED`
- HTTP success：`200`

公开条件不满足或家庭不存在时：

```json
{
  "code": 42402,
  "message": "公开家庭树不可访问",
  "data": null
}
```

公开状态会在读取缓存前校验，旧缓存不能绕过最新公开权限。

## Cache

服务先读取家庭当前 `graph_version`，再使用严格格式的 Redis key：

```text
family:{familyId}:tree:v{graphVersion}
```

例如：

```text
family:2:tree:v14
```

缓存未命中时从 MySQL 读取 ACTIVE、未软删除的成员与关系，组装后写入 Redis，TTL 为 24 小时。正确性依赖 `graph_version`，TTL 只负责回收旧版本。Redis 读取或写入失败时接口降级为 MySQL 读取。
