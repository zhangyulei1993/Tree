# 17. API 联调与调试指南 v1

> 用于前后端联调、Apifox/Postman 接口集合、Token 使用、错误码处理和常见问题排查。

## 1. 基础地址

```text
后端：http://127.0.0.1:8080
普通 API：http://127.0.0.1:8080/api
后台 API：http://127.0.0.1:8080/api/admin
公开 API：http://127.0.0.1:8080/api/public
```

## 2. 请求头

```http
Content-Type: application/json
Authorization: Bearer <token>
```

普通用户 token 不能访问 `/api/admin/*`；后台管理员 token 不能当作普通家庭成员 token 使用。

## 3. 联调顺序

```text
1. GET /api/health
2. POST /api/admin/auth/login
3. POST /api/auth/send-code
4. POST /api/auth/register-phone
5. POST /api/auth/login-phone
6. POST /api/families
7. GET /api/users/me/families
8. POST /api/families/{familyId}/members
9. POST /api/families/{familyId}/relationships
10. GET /api/families/{familyId}/tree
11. POST /api/families/{familyId}/members/{memberId}/invite
12. POST /api/invitations/{invitationId}/accept
13. POST /api/families/{familyId}/public-applications
14. POST /api/admin/family-public-applications/{applicationId}/approve
15. POST /api/public/families/{familyId}/visitor-messages
16. POST /api/admin/visitor-messages/{messageId}/approve
17. GET /api/admin/operation-logs
```

## 4. 后台登录

```http
POST /api/admin/auth/login
```

请求：

```json
{"username":"admin","password":"Admin@123456"}
```

成功后保存 `admin accessToken`。

## 5. 用户注册登录

发送验证码：

```http
POST /api/auth/send-code
```

```json
{"phone":"13800000001","scene":"REGISTER","clientType":"H5_WEB"}
```

注册：

```http
POST /api/auth/register-phone
```

```json
{"phone":"13800000001","code":"123456","password":"User@123456","nickname":"测试用户","clientType":"H5_WEB"}
```

登录：

```http
POST /api/auth/login-phone
```

```json
{"phone":"13800000001","password":"User@123456","clientType":"H5_WEB"}
```

## 6. 创建家庭

```http
POST /api/families
```

```json
{
  "familyName":"张氏家族",
  "familySurname":"张",
  "nativePlace":"山东济南",
  "regionCode":"370100",
  "regionText":"山东省济南市",
  "description":"张氏家族资料整理",
  "founderMember":{
    "surname":"张",
    "generationCharacter":"明",
    "givenName":"远",
    "displayName":"张明远",
    "gender":"MALE",
    "birthDate":"1990-01-01"
  }
}
```

检查：families、family_members、family_member_user_links、operation_logs。

## 7. 添加成员与关系

创建成员后应检查 `graph_version + 1`。

添加兄弟姐妹时，如果 baseMember 无父母节点，应返回：

```json
{"code":43301,"message":"请先创建父亲或母亲节点，再添加兄弟姐妹","data":{"missingRequiredParent":true}}
```

## 8. 获取家庭树

```http
GET /api/families/{familyId}/tree
```

返回必须包含：

```text
nodes
edges
tree
graphVersion
```

不返回动态称谓。

## 9. 邀请调试

家庭管理员生成分享邀请：

```http
POST /api/families/{familyId}/members/{memberId}/invite
```

```json
{"inviteChannel":"SHARE_LINK","inviteMessage":"请确认是否绑定为该家庭成员","familyRoleAfterAccept":"MEMBER"}
```

平台管理员只能对已注册用户创建 IN_APP 邀请，不能生成 SHARE_LINK。

## 10. 公开申请与留言

公开申请审核通过后：

```text
families.public_display_status = APPROVED
GET /api/public/families/{familyId}/profile 可访问
```

游客留言审核前不展示；审核后展示；公开留言不展示 visitorPhone / visitorWechat。

## 11. 操作日志检查

```http
GET /api/admin/operation-logs?module=FAMILY&action=CREATE_FAMILY&page=1&pageSize=20
GET /api/admin/operation-logs/{logId}
```

日志详情不能出现密码、验证码、token、session_key。

## 12. graph_version 检查

必须 +1：创建成员、删除成员、添加关系、删除关系、恢复成员、修改成员关键字段。

不应变化：接受邀请、设置管理员、公开申请、留言、创始人转让、家庭解散。

## 13. 常见问题

| 问题 | 排查 |
|---|---|
| 10007 未登录 | Authorization 是否存在，Bearer 是否正确，token 是否过期 |
| 10009 请先绑定手机号 | user.phone_verified 是否为 0 |
| 后台无权访问 | 是否使用 admin token，角色是否允许 |
| 创建兄弟姐妹失败 | baseMember 是否已有父亲或母亲 |
| 删除成员失败 | 是否存在下级成员，是否为 FOUNDER |

## 14. Apifox/Postman 分组建议

```text
Auth
Admin Auth
Users
Families
Family Members
Family Relationships
Invitations
Join Requests
Public Applications
Visitor Messages
Family Roles
Founder Transfer
Dissolution
Admin Users
Admin Families
Operation Logs
```
