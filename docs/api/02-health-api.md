# Health API

## GET /api/health

- Authorization：否
- Token：`NONE`
- Request JSON：无
- HTTP success：`200`

成功响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "ok"
  }
}
```

该接口只反映 Gin 服务可响应，不检查 MySQL 或 Redis 健康状态。
