# Tree Web

M15 PC/H5 前台静态原型。

## 技术栈

- Vue 3
- TypeScript
- Vite
- Pinia
- Vue Router
- Axios

## 启动方式

```bash
pnpm install
pnpm dev --host 0.0.0.0 --port 5173
```

访问：

```text
http://127.0.0.1:5173/
```

## 构建方式

```bash
pnpm build
```

## Mock 数据说明

当前所有页面均使用 `src/mock/data.ts` 中的本地 mock data。

mock token 使用 `example_user_token`，邀请 token 使用 `mock_invite_token`，手机号均为脱敏样例。

## 当前不接真实 API

`src/api/client.ts` 仅保留未来 API 对接结构。本阶段不请求真实后端。

## 页面清单

- `/`
- `/families`
- `/families/:familyId/public`
- `/families/:familyId/tree/public`
- `/login`
- `/register`
- `/me`
- `/me/families`
- `/invite/:inviteToken`
- `/families/:familyId/join`

## 后续真实 API 对接计划

1. 用户登录注册对接 `/api/auth/*`。
2. 公开家庭主页对接 `GET /api/families/{familyId}/public`。
3. 公开树对接 `GET /api/public/families/{familyId}/tree`。
4. 留言对接 `/api/public/families/{familyId}/visitor-messages`。
5. 邀请和加入申请对接 M11 已实现接口。
6. 如果需要公开家庭搜索，应由后端补充公开搜索接口。
