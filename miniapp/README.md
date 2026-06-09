# Tree Miniapp

M15 微信小程序静态原型。

## 技术栈

- uni-app
- Vue 3
- TypeScript
- Pinia

## 启动方式

```bash
pnpm install
pnpm dev:mp-weixin
```

## 构建方式

```bash
pnpm build:mp-weixin
```

构建产物用于微信开发者工具导入。

## Mock 数据说明

当前所有页面均使用 `src/mock/data.ts` 中的本地 mock data。

mock token 使用 `example_user_token`，邀请 token 使用 `mock_invite_token`，手机号均为脱敏样例。

## 当前不接真实 API

本阶段不请求真实后端，不调用真实微信登录，不配置真实 AppID 或 AppSecret。

## 页面清单

- `pages/home/index`
- `pages/auth/wechat-login`
- `pages/auth/bind-phone`
- `pages/family/search`
- `pages/family/public-profile`
- `pages/family/public-tree`
- `pages/family/my`
- `pages/family/detail`
- `pages/invite/detail`
- `pages/join/apply`
- `pages/me/index`
- `pages/legal/user-agreement`
- `pages/legal/privacy-policy`
- `pages/account/cancel`

## 后续真实 API 对接计划

1. 微信登录对接 `POST /api/auth/wechat-mini/login`。
2. 手机号绑定对接 `POST /api/auth/wechat-mini/bind-phone`。
3. 公开家庭页对接 `GET /api/families/{familyId}/public`。
4. 公开树对接 `GET /api/public/families/{familyId}/tree`。
5. 邀请确认对接 M11 invitation API。
6. 加入申请对接 M11 join request API。
