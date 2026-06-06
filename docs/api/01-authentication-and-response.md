# Authentication And Response

## Authorization Header

受保护接口必须发送：

```http
Authorization: Bearer <accessToken>
```

Header 缺失、不是精确的 `Bearer ` 前缀或 token 为空时：

```json
{"code":10007,"message":"未登录","data":null}
```

签名无效、过期、token 类型不匹配、token 已撤销或 Redis 撤销检查失败时：

```json
{"code":10008,"message":"登录已过期","data":null}
```

## JWT 类型

JWT 使用 HS256，主要 claims：

| Claim | 类型 | 说明 |
|---|---|---|
| `sub` | integer | user ID 或 admin ID |
| `typ` | string | `USER` 或 `ADMIN` |
| `iss` | string | 默认 `Tree` |
| `iat` | integer | 签发 Unix 时间 |
| `exp` | integer | 过期 Unix 时间 |
| `jti` | string | token 唯一 ID |
| `role` | string | 管理员 token 包含角色；用户 token 为空 |

用户和管理员使用不同 secret。配置相同会导致 JWT manager 初始化失败。

## Logout 与撤销

- 用户与管理员 logout 都将当前 `jti` 写入 Redis blacklist，TTL 为 token 剩余有效期。
- 用户账号合并和注销还使用 user-level revoked timestamp，使该用户指定时间之前签发的 token 失效。
- Redis 不可用时，router 当前会降级为 `NoopTokenBlacklist`；此时 logout 无法持久化撤销状态。

## 统一响应

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

`code` 是业务码，不等于 HTTP 状态码。失败时 `data` 为 `null`。

常见 HTTP 与业务码组合：

| HTTP | 常见业务码 |
|---|---|
| 200 | `0` |
| 201 | `0`，仅创建家庭 |
| 400 | `10001` 或模块业务错误 |
| 401 | `10007`、`10008`、登录错误 |
| 403 | 权限或账号状态错误 |
| 404 | 家庭/请求不存在 |
| 500 | `10000` |

## Phone Verified 中间件状态

代码中存在 `RequirePhoneVerified`，但 M1-M7 router 当前没有挂载它。家庭创建由 service 直接检查数据库中的 `users.status = ACTIVE` 和 `phone_verified = true`。
