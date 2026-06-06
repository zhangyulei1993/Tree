# User Phone Authentication API

验证码 scene 真实支持值：

`REGISTER`、`LOGIN`、`BIND_PHONE`、`CHANGE_PHONE_OLD`、`CHANGE_PHONE_NEW`、`CANCEL_ACCOUNT`。

代码定义的 clientType 值为 `PC_WEB`、`H5_WEB`、`WECHAT_MINI_PROGRAM`，但 DTO 当前只校验非空，没有枚举校验。

## POST /api/auth/send-code

- Authorization：否
- Token：`NONE`

| 字段 | 类型 | 必填 |
|---|---|---|
| `phone` | string | 是 |
| `scene` | string | 是 |
| `clientType` | string | 是 |

请求：

```json
{"phone":"13800138000","scene":"REGISTER","clientType":"H5_WEB"}
```

成功：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "expireSeconds": 300,
    "cooldownSeconds": 60,
    "devCode": "123456"
  }
}
```

`devCode` 只在 `APP_ENV != prod` 时出现。

常见错误：

```json
{"code":40002,"message":"验证码场景错误","data":null}
```

还可能返回 `40001` 手机号格式错误、`40003` 发送过于频繁、`40004` 手机号已注册、`40005` 手机号未注册、`40703` 新手机号已存在。

## POST /api/auth/register-phone

- Authorization：否
- Token：`NONE`

| 字段 | 类型 | 必填 |
|---|---|---|
| `phone` | string | 是 |
| `code` | string | 是 |
| `password` | string | 是 |
| `nickname` | string/null | 否 |
| `clientType` | string | 是 |

请求：

```json
{"phone":"13800138000","code":"123456","password":"Password123!","nickname":"张三","clientType":"H5_WEB"}
```

成功：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "accessToken": "<user-jwt>",
    "tokenType": "Bearer",
    "user": {
      "id": 8,
      "phone": "13800138000",
      "phoneVerified": true,
      "nickname": "张三",
      "status": "ACTIVE"
    }
  }
}
```

常见错误：

```json
{"code":40101,"message":"验证码错误","data":null}
```

还可能返回 `40001`、`40102`、`40103`、`40104`、`40106`。当前密码强度实现只检查长度至少 6。

## POST /api/auth/login-phone

- Authorization：否
- Token：`NONE`

| 字段 | 类型 | 必填 |
|---|---|---|
| `phone` | string | 是 |
| `password` | string | 是 |
| `clientType` | string | 是 |

请求：

```json
{"phone":"13800138000","password":"Password123!","clientType":"H5_WEB"}
```

成功响应结构与注册接口的 `LoginResponse` 相同。

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "accessToken": "<user-jwt>",
    "tokenType": "Bearer",
    "user": {
      "id": 8,
      "phone": "13800138000",
      "phoneVerified": true,
      "status": "ACTIVE"
    }
  }
}
```

常见错误：

```json
{"code":40201,"message":"手机号或密码错误","data":null}
```

状态错误可能返回 `40202`、`40203`、`40204`、`40205` 或 `10010`。

## POST /api/auth/logout

- Authorization：是
- Token：`USER`
- Request JSON：无

成功：

```json
{"code":0,"message":"success","data":{"status":"ok"}}
```

常见错误：

```json
{"code":10008,"message":"登录已过期","data":null}
```

成功时当前 token 的 `jti` 被加入 Redis blacklist。
