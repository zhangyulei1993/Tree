# Tree API Overview

本文档基于当前 `backend/internal/app/router.go`、DTO、VO、handler 和 service 实现整理，覆盖 M1-M9 已注册的 31 个接口。

## 基础约定

- 基础路径：`/api`
- 本地地址：`http://127.0.0.1:8080`
- 请求与响应：`application/json`
- 路径参数 ID：无符号整数
- 时间字段：Go `time.Time` 输出的 RFC 3339 字符串

## 统一响应

成功：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

失败：

```json
{
  "code": 10001,
  "message": "参数错误",
  "data": null
}
```

大部分成功接口返回 HTTP `200`。`POST /api/families` 返回 HTTP `201`。

## 认证类型

| 标记 | 含义 |
|---|---|
| `NONE` | 不需要 Authorization |
| `USER` | 需要用户 JWT，由 `JWT_USER_SECRET` 签名 |
| `ADMIN` | 需要管理员 JWT，由 `JWT_ADMIN_SECRET` 签名 |

USER 与 ADMIN token 不可互换。JWT payload 的 `typ` 必须与路由中间件期望类型一致。

## 已实现接口

| 模块 | Method | Path | Token |
|---|---|---|---|
| Health | GET | `/api/health` | NONE |
| Admin Auth | POST | `/api/admin/auth/login` | NONE |
| Admin Auth | POST | `/api/admin/auth/logout` | ADMIN |
| Admin Auth | GET | `/api/admin/me` | ADMIN |
| Admin Auth | PUT | `/api/admin/me/password` | ADMIN |
| User Auth | POST | `/api/auth/send-code` | NONE |
| User Auth | POST | `/api/auth/register-phone` | NONE |
| User Auth | POST | `/api/auth/login-phone` | NONE |
| User Auth | POST | `/api/auth/logout` | USER |
| WeChat Account | POST | `/api/auth/wechat-mini/login` | NONE |
| WeChat Account | POST | `/api/auth/wechat-mini/bind-phone` | USER |
| WeChat Account | POST | `/api/auth/change-phone` | USER |
| WeChat Account | POST | `/api/auth/cancel-account` | USER |
| Family | POST | `/api/families` | USER |
| Family | GET | `/api/families` | USER |
| Family | GET | `/api/families/{familyId}` | USER |
| Family | GET | `/api/families/{familyId}/public` | NONE |
| Family | PUT | `/api/families/{familyId}` | USER |
| Family | POST | `/api/families/{familyId}/dissolution-requests` | USER |
| Family | GET | `/api/families/{familyId}/dissolution-requests/current` | USER |
| Family | POST | `/api/families/{familyId}/dissolution-requests/{requestId}/cancel` | USER |
| Family Member | POST | `/api/families/{familyId}/members` | USER |
| Family Member | GET | `/api/families/{familyId}/members` | USER |
| Family Member | GET | `/api/families/{familyId}/members/{memberId}` | USER |
| Family Member | PUT | `/api/families/{familyId}/members/{memberId}` | USER |
| Family Member | DELETE | `/api/families/{familyId}/members/{memberId}` | USER |
| Family Member | POST | `/api/families/{familyId}/members/{memberId}/bind-user` | USER |
| Family Member | POST | `/api/families/{familyId}/members/{memberId}/unbind-user` | USER |
| Family Relationship | POST | `/api/families/{familyId}/relationships` | USER |
| Family Relationship | PUT | `/api/families/{familyId}/relationships/{relationshipId}` | USER |
| Family Relationship | DELETE | `/api/families/{familyId}/relationships/{relationshipId}` | USER |

详细契约见同目录其他文档及 [openapi.yaml](./openapi.yaml)。
