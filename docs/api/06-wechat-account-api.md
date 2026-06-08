# WeChat And Account API

## POST /api/auth/wechat-mini/login

- Authorization：否
- Token：`NONE`

| 字段 | 类型 | 必填 |
|---|---|---|
| `code` | string | 是 |
| `clientType` | string | 是 |

请求：

```json
{"code":"wx-login-code","clientType":"WECHAT_MINI_PROGRAM"}
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
      "id": 10,
      "phoneVerified": false,
      "status": "PENDING_PHONE_BIND"
    }
  }
}
```

首次登录创建 `PENDING_PHONE_BIND` 用户及微信 identity。已有 identity 会返回对应用户。开发环境可通过 `WECHAT_MOCK_ENABLED=true` 使用确定性 mock identity。

常见错误：

```json
{"code":40302,"message":"微信 openid 获取失败","data":null}
```

还可能返回 `40303` 微信身份失效、`40304` 微信账号状态异常。

## POST /api/auth/wechat-mini/bind-phone

- Authorization：是
- Token：`USER`

| 字段 | 类型 | 必填 |
|---|---|---|
| `phone` | string | 是 |
| `code` | string | 是 |

请求：

```json
{"phone":"13800138000","code":"123456"}
```

成功返回新的 USER token：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "accessToken": "<new-user-jwt>",
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

行为：

- 手机号未被占用：绑定当前账号并转为 `ACTIVE`。
- 手机号属于 `ACTIVE` 用户：合并到该用户，临时账号转为 `MERGED`。
- 手机号属于 `PENDING_CLAIM` 用户：认领该账号，临时账号转为 `MERGED`。
- 两个账号在同一家庭绑定不同 member：返回冲突，不自动合并。

常见错误：

```json
{"code":40504,"message":"同一家庭下存在多个成员绑定冲突","data":null}
```

还可能返回 `40001`、`40401`、`40402`、`40403`、`40405`、`40406`。

## POST /api/auth/change-phone

- Authorization：是
- Token：`USER`

| 字段 | 类型 | 必填 |
|---|---|---|
| `oldPhoneCode` | string | 是 |
| `newPhone` | string | 是 |
| `newPhoneCode` | string | 是 |

请求：

```json
{"oldPhoneCode":"123456","newPhone":"13900139000","newPhoneCode":"123456"}
```

成功：

```json
{"code":0,"message":"success","data":{"status":"ok"}}
```

常见错误：

```json
{"code":40701,"message":"旧手机号验证码错误","data":null}
```

还可能返回 `40001`、`10010`、`40702`、`40703`。

## POST /api/auth/cancel-account

- Authorization：是
- Token：`USER`

| 字段 | 类型 | 必填 |
|---|---|---|
| `phoneCode` | string | 是 |
| `cancelReason` | string | 否 |

请求：

```json
{"phoneCode":"123456","cancelReason":"不再使用"}
```

成功：

```json
{"code":0,"message":"success","data":{"status":"ok"}}
```

注销会将用户状态设为 `CANCELLED`，释放手机号，清除密码及部分资料，取消微信 identity，并撤销当前及旧用户 token。

常见错误：

```json
{"code":41003,"message":"家庭创始人必须先转让创始人身份","data":null}
```

还可能返回 `41001`、`41002`、`41004`、`41005`、`41006`。
