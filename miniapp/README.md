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
| H5 本地开发 | `pnpm dev:h5:local` | `.env.development`：`/api`（Vite proxy → `127.0.0.1:8080`） |
| H5 staging | `pnpm dev:h5:staging` / `pnpm build:h5:staging` | `.env.staging-h5`：`/api`（Vite proxy → `tapi.bigbigboy.cn`） |
| 微信小程序本地开发 | `pnpm dev:mp-weixin:local` / `pnpm build:mp-weixin:local` | `.env.mp-local`：`http://127.0.0.1:8080/api` |
| 微信小程序 staging | `pnpm dev:mp-weixin:staging` / `pnpm build:mp-weixin:staging` | `.env.staging-mp`：`https://tapi.bigbigboy.cn/api` |

**重要：** 微信小程序真机/体验版不能使用 `/api`、`localhost` 或 `127.0.0.1`。必须使用完整 HTTPS 域名，并在微信公众平台配置以下合法域名（域名填 `tapi.bigbigboy.cn`，不要带路径）：

| 类型 | 配置项 | 域名 |
|------|--------|------|
| 普通请求 | request 合法域名 | `https://tapi.bigbigboy.cn` |
| 头像上传 | uploadFile 合法域名 | `https://tapi.bigbigboy.cn` |
| 头像展示 | downloadFile 合法域名 | `https://tapi.bigbigboy.cn` |

仅配置 request 域名时，登录接口可用，但头像上传会报 `uploadFile:fail url not in domain list`。

以后切接口环境只改命令或 `.env.*` 文件，微信开发者工具固定打开同一个小程序产物目录，不需要来回切项目目录，也不需要手改源码。

若 MP-WEIXIN real 模式 baseURL 配置错误，`src/api/client.ts` 会在请求前给出可读错误提示。

## 启动方式

```bash
pnpm install

# H5 本地（proxy）
pnpm dev:h5:local

# 微信小程序 + 本地 API
pnpm dev:mp-weixin:local

# 微信小程序 + staging API
pnpm dev:mp-weixin:staging
```

## 构建方式

```bash
# 微信小程序本地接口构建
pnpm build:mp-weixin:local

# 微信小程序 staging（上传体验版前使用）
pnpm build:mp-weixin:staging

# H5 本地
pnpm build:h5:local

# H5 staging
pnpm build:h5:staging
```

`pnpm build:mp-weixin` 保留为 staging 构建兼容命令；新流程请优先使用带环境后缀的命令。

微信开发者工具固定打开 `miniapp/dist/mp-weixin`。开发、联调、上传前构建都写入这个目录。切换接口环境或上传前，只需要停止当前命令并重新运行对应命令，工具里重新编译或上传即可，不需要手动同步目录。

## 看不到界面更新？

源码改完后，**微信开发者工具不会自动读 `src/`**，必须重新编译并导入正确目录。

### 1. 确认导入目录（最常见问题）

| 命令 | 应导入的目录 |
|------|----------------|
| `pnpm dev:mp-weixin:local` | `miniapp/dist/mp-weixin` |
| `pnpm dev:mp-weixin:staging` | `miniapp/dist/mp-weixin` |
| `pnpm build:mp-weixin:local` | `miniapp/dist/mp-weixin` |
| `pnpm build:mp-weixin:staging` | `miniapp/dist/mp-weixin` |

**不要**导入 `miniapp` 根目录或 `miniapp/src`。

操作：微信开发者工具 → **项目** → **重新打开** → 选择上表对应目录。

### 2. 重新编译

```bash
cd miniapp
pnpm install
pnpm dev:mp-weixin:staging   # 开发：保持终端运行，微信开发者工具打开 dist/mp-weixin
# 或
pnpm build:mp-weixin:staging # 构建：上传前仍使用 dist/mp-weixin
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
- `pages/family/my` — 我的家庭
- `pages/invite/detail` — 邀请确认
- `pages/join/apply` — 加入申请
- `pages/me/index` — 个人中心
