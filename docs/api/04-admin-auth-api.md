# Admin Authentication API

## POST /api/admin/auth/login

- Authorization：否
- Token：`NONE`

| 字段 | 类型 | 必填 |
|---|---|---|
| `username` | string | 是 |
| `password` | string | 是 |

请求：

```json
{"username":"admin","password":"<password>"}
```

成功：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "accessToken": "<admin-jwt>",
    "tokenType": "Bearer",
    "admin": {
      "id": 1,
      "username": "admin",
      "displayName": "ROOT_ADMIN",
      "role": "ROOT_ADMIN",
      "status": "ACTIVE"
    }
  }
}
```

常见错误：

```json
{"code":47001,"message":"用户名或密码错误","data":null}
```

还可能返回 `47002` 管理员禁用、`47003` 管理员锁定、`47004` 管理员删除。ROOT_ADMIN 密码失败会记录失败次数但不会设置自动锁定时间。

## POST /api/admin/auth/logout

- Authorization：是
- Token：`ADMIN`
- Request JSON：无

成功：

```json
{"code":0,"message":"success","data":{"status":"ok"}}
```

常见错误：

```json
{"code":10008,"message":"登录已过期","data":null}
```

## GET /api/admin/me

- Authorization：是
- Token：`ADMIN`
- Request JSON：无

成功：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1,
    "username": "admin",
    "displayName": "ROOT_ADMIN",
    "phone": null,
    "email": null,
    "role": "ROOT_ADMIN",
    "status": "ACTIVE"
  }
}
```

`phone` 和 `email` 使用 `omitempty`，为空时可能不出现在 JSON 中；`displayName` 没有 `omitempty`，可返回 `null`。

常见错误：

```json
{"code":47002,"message":"管理员账号已禁用","data":null}
```

## PUT /api/admin/me/password

- Authorization：是
- Token：`ADMIN`

| 字段 | 类型 | 必填 |
|---|---|---|
| `oldPassword` | string | 是 |
| `newPassword` | string | 是 |

请求：

```json
{"oldPassword":"<old>","newPassword":"<new>"}
```

成功：

```json
{"code":0,"message":"success","data":{"status":"ok"}}
```

当前 token 会写入 blacklist。代码尚未实现“撤销该管理员全部旧 token”。

常见错误：

```json
{"code":47001,"message":"用户名或密码错误","data":null}
```
