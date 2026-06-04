# 04-02. 家庭基础接口完整规格 v1

## 1. 模块范围

覆盖：

```text
创建家庭
搜索家庭
查看家庭详情
编辑家庭基础资料
查看公开家庭主页
查看家庭树入口
```

涉及表：

```text
families
family_members
family_member_user_links
family_relationships
visitor_messages
operation_logs
```

## 2. 核心规则

```text
当前版本一个家庭只允许一个核心姓氏 family_surname。
游客可搜索 NORMAL + searchable = 1 的家庭。
只有 public_display_status = APPROVED 的家庭可展示公开主页和公开树。
一期家庭树为列表树 / 层级树，不使用 centerMemberId。
家庭公开联系方式可以为空。
```

## 3. POST /api/families

创建家庭。

### 权限

```text
用户已登录
users.status = ACTIVE
users.phone_verified = 1
```

### 请求

```json
{
  "familyName": "张氏家族",
  "familySurname": "张",
  "nativePlace": "山东济南",
  "regionCode": "370100",
  "regionText": "山东省济南市",
  "description": "张氏家族资料整理",
  "avatarUrl": "https://example.com/avatar.png",
  "founderMember": {
    "surname": "张",
    "generationCharacter": "明",
    "givenName": "明远",
    "displayName": "张明远",
    "gender": "MALE",
    "birthDate": "1990-01-01"
  }
}
```

### 流程

```text
创建 families
创建创建者 family_members
创建 family_member_user_links
family_role = FOUNDER
families.current_founder_member_id = 创建者 member_id
写 operation_logs
```

## 4. GET /api/families

搜索家庭。

### 权限

无需登录，游客可访问。

### 查询参数

```text
keyword
familySurname
nativePlace
regionCode
regionText
publicOnly
page
pageSize
```

### 条件

```text
status = NORMAL
searchable = 1
publicOnly=true 时 additionally public_display_status = APPROVED
```

## 5. GET /api/families/{familyId}

查看家庭详情。

### 规则

```text
游客 / 非成员：只能看基础公开信息
家庭成员：可看完整家庭资料
后台管理员：建议使用 /api/admin/families/{familyId}
```

## 6. PUT /api/families/{familyId}

编辑家庭资料。

### 权限

```text
FOUNDER
FAMILY_ADMIN
ROOT_ADMIN
SUPER_ADMIN
PLATFORM_ADMIN
```

### 可编辑字段

```text
familyName
nativePlace
regionCode
regionText
description
avatarUrl
searchable
publicContactName
publicContactPhone
publicContactWechat
publicContactNote
publicContactVisible
```

### 不建议前台修改

```text
familySurname
status
publicDisplayStatus
currentFounderMemberId
graphVersion
```

公开资料修改后直接生效，必须写 operation_logs。

## 7. GET /api/public/families/{familyId}/profile

公开家庭主页。

### 条件

```text
families.status = NORMAL
families.public_display_status = APPROVED
```

返回公开联系方式、公开树入口、留言数量等。

## 8. GET /api/families/{familyId}/tree

私有家庭树。

### 权限

```text
该 family 的 MEMBER / FAMILY_ADMIN / FOUNDER
后台管理员
```

返回：

```text
nodes
edges
tree
```

## 9. GET /api/public/families/{familyId}/tree

公开家庭树。

### 条件

```text
family.status = NORMAL
public_display_status = APPROVED
```

游客可看当前公开树；手动切换其他公开树时，需要登录并绑定手机号。

## 10. 家庭树节点字段

```json
{
  "memberId": 20001,
  "displayName": "张明远",
  "surname": "张",
  "generationCharacter": "明",
  "gender": "MALE",
  "memberType": "LINEAGE_MEMBER",
  "userBindingState": "BOUND",
  "canExpand": true,
  "stopReason": null,
  "spouses": []
}
```

`userBindingState`：

```text
BOUND
INVITING
NOT_REQUIRED
UNBOUND
```

## 11. 错误码范围

```text
42000-42999
```
