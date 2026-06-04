# 04-03. 家庭成员与关系接口完整规格 v1

## 1. 模块范围

覆盖：

```text
创建成员
编辑成员
删除成员
恢复成员
标记无需绑定用户
添加父亲 / 母亲 / 子女 / 兄弟姐妹 / 配偶
删除关系
修改关系备注
```

涉及表：

```text
families
family_members
family_relationships
family_member_user_links
family_invitations
operation_logs
```

## 2. 核心原则

```text
family_member 不等于 user。
family_member 可以没有 user。
relationship_type 只保存 PARENT_CHILD 和 SPOUSE。
兄弟姐妹由共同父母推导，不保存 SIBLING。
addType 是用户操作入口；relationship_type 是数据库结构关系。
```

## 3. 枚举

```text
member_type: LINEAGE_MEMBER / SPOUSE / EXTERNAL_MEMBER
gender: MALE / FEMALE / UNKNOWN
user_binding_policy: OPTIONAL / REQUIRED / NOT_REQUIRED / DISABLED
unbound_reason: DECEASED / MINOR / ELDERLY_UNABLE_TO_USE / UNREACHABLE / NOT_WILLING_TO_REGISTER / FORCE_MAJEURE / HISTORICAL_ANCESTOR / OTHER
relationship_type: PARENT_CHILD / SPOUSE
parent_link_type: PRIMARY / STEP / ADOPTIVE / SUCCESSION / NOTE_ONLY / OTHER
addType: ADD_FATHER / ADD_MOTHER / ADD_CHILD / ADD_SIBLING / ADD_SPOUSE
userBindingState: BOUND / INVITING / NOT_REQUIRED / UNBOUND
```

## 4. POST /api/families/{familyId}/members

创建成员。

### 权限

```text
FOUNDER
FAMILY_ADMIN
ROOT_ADMIN
SUPER_ADMIN
PLATFORM_ADMIN
```

### 请求

```json
{
  "memberType": "LINEAGE_MEMBER",
  "surname": "张",
  "generationCharacter": "明",
  "givenName": "明远",
  "displayName": "张明远",
  "gender": "MALE",
  "birthDate": "1990-01-01",
  "birthOrder": 1,
  "nativePlace": "山东济南",
  "regionText": "山东省济南市",
  "isLiving": true,
  "userBindingPolicy": "OPTIONAL",
  "unboundReason": null,
  "unboundNote": null
}
```

### 影响

```text
INSERT family_members
families.graph_version + 1
写 operation_logs: FAMILY_MEMBER / CREATE_MEMBER
```

## 5. GET /api/families/{familyId}/members/{memberId}

查看成员详情。

权限：

```text
家庭成员
家庭管理员
后台管理员
```

游客不能查看私有成员详情。

## 6. PUT /api/families/{familyId}/members/{memberId}

编辑成员。

### 权限

```text
MEMBER 可编辑本人资料
FAMILY_ADMIN / FOUNDER 可编辑家庭成员
后台管理员可协助编辑
```

普通 MEMBER 不能编辑：

```text
memberType
userBindingPolicy
unboundReason
unboundNote
familyRole
status
关系
绑定状态
```

修改关键字段触发 graph_version：

```text
surname
generationCharacter
givenName
displayName
gender
birthDate
birthOrder
manualOrder
memberType
manualLineageOverride
lineageNoteType
lineageNote
```

## 7. DELETE /api/families/{familyId}/members/{memberId}

删除成员。

### 权限

```text
FAMILY_ADMIN
FOUNDER
ROOT_ADMIN
SUPER_ADMIN
PLATFORM_ADMIN
```

### 规则

```text
存在下级成员 -> 禁止删除
不能直接删除家庭当前 FOUNDER
删除主成员且有配偶 -> 需要 confirmDeleteSpouse = true
删除配偶节点 -> 只删除该配偶
```

### 影响

```text
软删除 member
相关关系软删除
ACTIVE link 置 REVOKED / UNLINKED
families.graph_version + 1
写 operation_logs
```

## 8. POST /api/families/{familyId}/members/{memberId}/mark-not-required

标记无需绑定用户。

### 请求

```json
{
  "unboundReason": "DECEASED",
  "unboundNote": "该成员已故，无需绑定系统账号"
}
```

### 规则

```text
已有 ACTIVE 绑定 -> 不能直接标记
存在 PENDING invitation -> 需先撤销邀请
通常不触发 graph_version
```

## 9. POST /api/families/{familyId}/members/{memberId}/restore

恢复成员。

建议一期仅 ROOT_ADMIN / SUPER_ADMIN 支持。

## 10. POST /api/families/{familyId}/relationships

通过 addType 创建成员与关系。

### 请求

```json
{
  "baseMemberId": 20001,
  "addType": "ADD_CHILD",
  "newMember": {
    "memberType": "LINEAGE_MEMBER",
    "surname": "张",
    "generationCharacter": "明",
    "givenName": "小明",
    "displayName": "张小明",
    "gender": "MALE"
  },
  "relationship": {
    "relationshipType": "PARENT_CHILD",
    "parentLinkType": "PRIMARY",
    "relationNoteType": null,
    "relationNote": null
  }
}
```

### ADD_FATHER

落库：

```text
newFather -- PARENT_CHILD --> baseMember
```

同一 child 的 PRIMARY 父亲最多一个。

### ADD_MOTHER

落库：

```text
newMother -- PARENT_CHILD --> baseMember
```

同一 child 的 PRIMARY 母亲最多一个。

### ADD_CHILD

落库：

```text
baseMember -- PARENT_CHILD --> newChild
```

### ADD_SIBLING

```text
查询 baseMember 的父母
没有父母 -> 返回错误：请先创建父亲或母亲节点
有父母 -> 创建新成员，并为每个父母创建 PARENT_CHILD 到新成员
```

数据库不写 SIBLING。

### ADD_SPOUSE

落库：

```text
baseMember -- SPOUSE --> newSpouse
newSpouse.member_type = SPOUSE
```

允许多个配偶。

## 11. DELETE /api/families/{familyId}/relationships/{relationshipId}

删除关系，软删除，触发 graph_version + 1。

## 12. PUT /api/families/{familyId}/relationships/{relationshipId}

修改关系备注。

可修改：

```text
parent_link_type
relation_note_type
relation_note
```

不建议修改 from_member_id / to_member_id / relationship_type。关系方向错了应删除后重建。

## 13. 错误码范围

```text
43000-43999
```
