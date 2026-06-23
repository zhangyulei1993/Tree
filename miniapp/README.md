# Tree Miniapp

uni-app 微信小程序 + H5。

## 技术栈

- uni-app
- Vue 3
- TypeScript
- Pinia

## API 地址配置

| 场景 | 命令 | `VITE_API_BASE_URL` |
|------|------|---------------------|
| H5 本地开发 | `pnpm dev:h5` | `/api`（Vite proxy → `127.0.0.1:8080`） |
| H5 real 构建 | `pnpm build:h5:real` | `/api` |
| 微信小程序 staging | `pnpm dev:mp-weixin:staging` / `pnpm build:mp-weixin:staging` | `https://tapi.bigbigboy.cn/api` |

**重要：** 微信小程序真机/体验版不能使用 `/api`、`localhost` 或 `127.0.0.1`。必须使用完整 HTTPS 域名，并在微信公众平台配置 request 合法域名：

```text
https://tapi.bigbigboy.cn
```

复制 `/.env.staging.example` 为本地 `.env.production.local` 时，请确保使用 staging HTTPS 地址，不要写 localhost。

若 MP-WEIXIN real 模式 baseURL 配置错误，`src/api/client.ts` 会在请求前给出可读错误提示。

## 启动方式

```bash
pnpm install

# H5 本地（proxy）
pnpm dev:h5

# 微信小程序 + staging API（推荐）
pnpm dev:mp-weixin:staging
```

## 构建方式

```bash
# 微信小程序 staging（上传体验版前使用）
pnpm build:mp-weixin:staging

# H5 real
pnpm build:h5:real

# 等价于 staging 小程序构建的手动命令：
# VITE_API_MODE=real VITE_API_BASE_URL=https://tapi.bigbigboy.cn/api pnpm build:mp-weixin
```

构建产物目录：`dist/build/mp-weixin`（微信开发者工具导入）。

## Mock 模式

未设置 `VITE_API_MODE=real` 时走 `src/mock/data.ts` 本地数据。

## 主要页面

- `pages/auth/wechat-login` — 微信登录
- `pages/auth/bind-phone` — 绑定手机号并设置密码
- `pages/auth/phone-login` — 手机号密码登录
- `pages/family/my` — 我的家庭
- `pages/invite/detail` — 邀请确认
- `pages/join/apply` — 加入申请
- `pages/me/index` — 个人中心
