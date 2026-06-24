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

**重要：** 微信小程序真机/体验版不能使用 `/api`、`localhost` 或 `127.0.0.1`。必须使用完整 HTTPS 域名，并在微信公众平台配置以下合法域名（域名填 `tapi.bigbigboy.cn`，不要带路径）：

| 类型 | 配置项 | 域名 |
|------|--------|------|
| 普通请求 | request 合法域名 | `https://tapi.bigbigboy.cn` |
| 头像上传 | uploadFile 合法域名 | `https://tapi.bigbigboy.cn` |
| 头像展示 | downloadFile 合法域名 | `https://tapi.bigbigboy.cn` |

仅配置 request 域名时，登录接口可用，但头像上传会报 `uploadFile:fail url not in domain list`。

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

## 看不到界面更新？

源码改完后，**微信开发者工具不会自动读 `src/`**，必须重新编译并导入正确目录。

### 1. 确认导入目录（最常见问题）

| 命令 | 应导入的目录 |
|------|----------------|
| `pnpm dev:mp-weixin:staging` | `miniapp/dist/dev/mp-weixin` |
| `pnpm build:mp-weixin:staging` | `miniapp/dist/build/mp-weixin` |

**不要**导入 `miniapp` 根目录或 `miniapp/src`。

操作：微信开发者工具 → **项目** → **重新打开** → 选择上表对应目录。

### 2. 重新编译

```bash
cd miniapp
pnpm install
pnpm dev:mp-weixin:staging   # 开发：保持终端运行，改代码后点开发者工具「编译」
# 或
pnpm build:mp-weixin:staging # 构建：再导入 dist/build/mp-weixin
```

### 3. 清缓存

开发者工具 → **工具** → **清缓存** → **全部清除** → 再点 **编译**。

### 4. 如何确认已是新版本

| 页面 | 旧版特征 | 新版特征 |
|------|----------|----------|
| 私有家谱 | 导航栏「私有家庭树」；上半部分有「家族成员」列表 | 导航栏「私有家谱」；**只有**「亲属关系」列表 |
| 成员列表 | 四格「年龄/性别/是否健在/是否绑定」 | 姓名右侧一行小标签：`98岁` `男` `已故` |

若导航栏仍是「私有家庭树」，说明加载的仍是旧构建产物。

### 5. 真机 / 体验版

手机上打开的是**已上传**的旧包，本地改代码不会影响。需在本机构建后，在开发者工具 **上传** 新版本，再在公众平台设为体验版。

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
