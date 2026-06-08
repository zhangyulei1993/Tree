# Tree Admin Web

M14 管理后台 UI 静态原型。

## 技术栈

- Vue 3
- TypeScript
- Vite
- Element Plus
- Pinia
- Vue Router
- Axios

## 启动方式

```bash
pnpm install
pnpm dev --host 0.0.0.0 --port 5174
```

访问：

```text
http://127.0.0.1:5174/admin/login
```

## 构建方式

```bash
pnpm build
```

## Mock 登录说明

登录页使用本地 mock 登录，不请求真实后端。可选择：

- ROOT_ADMIN
- SUPER_ADMIN
- PLATFORM_ADMIN

mock token 使用明显的示例值 `example_admin_token`，不是生产 token。

## 页面清单

- `/admin/login`
- `/admin/dashboard`
- `/admin/users`
- `/admin/users/:userId`
- `/admin/families`
- `/admin/families/:familyId`
- `/admin/families/:familyId/members`
- `/admin/public-applications`
- `/admin/visitor-messages`
- `/admin/founder-transfer-requests`
- `/admin/dissolution-requests`
- `/admin/admin-users`
- `/admin/operation-logs`

## 当前 API 状态

当前阶段不接真实 API，所有页面均使用 `src/mock/data.ts` 中的本地 mock data。

`src/api/client.ts` 仅保留未来对接结构。

## 后续真实 API 对接计划

1. 使用 `POST /api/admin/auth/login` 替换 mock 登录。
2. 使用 `GET /api/admin/me` 初始化当前管理员。
3. 将审核类页面对接已实现的 `/api/admin/...` 审核接口。
4. 等后端补齐用户管理、家庭管理、管理员管理、操作日志 API 后替换对应 mock data。
5. 保持 ADMIN token 与 USER token 分离，不在前端日志中输出 token。
