# Tree 小程序 H5 真实接口 E2E 测试

> 本文档记录 H5 真实接口全流程测试过程。
> **Playwright 端到端结果由 Codex 填写；代码修复、数据准备与复测结论见下文。**

## 1. 测试目标

- 验证 miniapp H5 在 `VITE_API_MODE=real` 下连接 staging API 的完整用户流程。
- 覆盖登录、家庭、成员、家谱、邀请、加入申请、权限与公开相关场景。
- 使用可重复、可审计的真实基线数据，而非 mock API。

## 2. 环境与版本

| 项 | 值 |
|---|---|
| API | `https://tapi.bigbigboy.cn/api` |
| H5 模式 | `VITE_API_MODE=real` |
| H5 Base URL | `VITE_API_BASE_URL=/api` |
| H5 Proxy | `VITE_API_PROXY_TARGET=https://tapi.bigbigboy.cn` |
| 状态文件 | `/tmp/tree-h5-real-e2e-state.json` |
| 用户凭据 | `/tmp/tree-h5-real-e2e-credentials.json` |
| 管理员凭据 | `/tmp/tree-h5-real-e2e-admin-credentials.json`（需先配置 `/tmp/tree-h5-e2e-admin.env`） |
| 截图目录 | `/tmp/tree-h5-e2e-screenshots` |
| API 版本 | `0.1.0` / commit `b633e450155f` / staging |
| 主家庭 familyId | `20`（graphVersion 约 42，17 成员） |
| 隔离家庭 familyId | `21` |
| 长期演示 familyId | `12` |

### H5 启动命令

```bash
cd miniapp
pnpm dev:h5:staging
```

## 3. 数据准备与维护

```bash
node scripts/tree-h5-real-e2e-setup.mjs audit
node scripts/tree-h5-real-e2e-setup.mjs setup
node scripts/tree-h5-real-e2e-setup.mjs repair-tree      # 修正家谱关系（幂等）
node scripts/tree-h5-real-e2e-setup.mjs cleanup-staging    # 清理 staging 遗留
node scripts/tree-h5-real-e2e-setup.mjs setup-admin        # 初始化管理员 env 模板
node scripts/tree-h5-real-e2e-setup.mjs status
node scripts/tree-h5-real-e2e-setup.mjs cleanup            # 仅取消 pending 邀请/申请
```

管理员凭据注入（**禁止提交仓库**）：

```bash
# 编辑 /tmp/tree-h5-e2e-admin.env
TREE_E2E_ADMIN_USERNAME=admin
TREE_E2E_ADMIN_PASSWORD=<staging ROOT_ADMIN 或 SUPER_ADMIN 密码>

node scripts/tree-h5-real-e2e-setup.mjs setup-admin
```

## 4. 测试角色

| 角色 | 手机号 | 成员节点 | userId |
|---|---|---|---|
| 创建者 FOUNDER | 16600020101 | 张承志 (105) | 46 |
| 家庭管理员 | 16600020102 | 张承浩 (113) | 47 |
| 普通成员 | 16600020103 | 张承雅 (114) | 48 |
| 家庭外申请人 | 16600020104 | — | 49 |

密码见 `/tmp/tree-h5-real-e2e-credentials.json`。

## 5. 基线数据变更记录

| 时间 | 变更 |
|---|---|
| 2026-06-30 | 创建 familyId=20/21 与四角色账号 |
| 2026-07-01 | 修复 family 20：张启国改为张启明兄弟；张待安恢复未定位 |
| 2026-07-01 | family 12：驳回 join request 5（张玉磊）；解绑 member 75 林思琪 ↔ user 48 |
| 2026-07-01 | pending：invitationId=24（张文昌）；joinRequestId=20（周远航） |
| 2026-07-01 | `repair-tree` 幂等修复：按共同父母判断兄弟关系，避免重复落位 |
| 2026-07-01 | `cleanup-staging` 修正 family 8 下架条件（非保留家庭应执行 take-down） |
| 2026-07-01 | admin `cleanup-staging`：reject 武汉访客留言 **6**；family **8** `take-down-public` |

## 6. 代码修复记录

| 项 | 状态 | 说明 |
|---|---|---|
| 更换手机号 P0 | **已修复 + 复测 PASS** | `change-phone` 返回 `{status:'ok'}`；保留 token 并 `refreshMe()` |
| 加入申请基准成员 picker | **复测 PASS** | 成员列表按姓名排序；选中成员与关系展示一致 |
| 家谱代际错误 | **已修复 + 复测 PASS** | setup + `repair-tree`；张启国与张启明共享父母 |
| E2E 脚本 repair-tree | **已修复** | 判断共同父母；不再调用完整 setup；updateMember 仅字段变化时执行 |
| E2E 脚本 cleanup-staging | **已修复** | family 8 不在保留列表时应执行 `take-down-public` |
| settings 角色标签 | **已修复** | 普通成员不再误显示「管理员」 |
| settings 已解散状态 | **已修复** | 显示「家庭已解散，等待恢复」，隐藏空白 Tab |
| admin 操作日志列宽 | **已修复** | 模块/动作 `min-width` + `show-overflow-tooltip` |

## 7. 测试场景清单（复测 2026-07-01）

| # | 场景 | 执行角色 | 结果 | 备注 |
|---|---|---|---|---|
| 1 | 手机号登录 / 退出 | 各角色 | _Codex 待填_ | |
| 2 | 更换手机号 | 创建者 | **PASS** | 真实接口；token 保持；资料刷新正确；临时换号账号已注销 |
| 3 | 个人资料更新 | 创建者 | _Codex 待填_ | |
| 4 | 我的家庭 / 家庭详情 | 创建者 | _Codex 待填_ | |
| 5 | 成员 / 家谱管理 | 创建者 | **PASS** | family 20 家谱结构正确 |
| 6 | 邀请流程 | 创建者 | _Codex 待填_ | invitationId=24 |
| 7 | 加入申请审批 | 创建者 | **PASS** | picker 选中成员与关系一致；joinRequestId=20 |
| 8 | 绑定已有成员（family 12） | 创建者 | _Codex 待填_ | 林思琪已解绑 |
| 9 | 普通成员无管理权限 | 张承雅 | _Codex 待填_ | |
| 10 | 公开申请 / 审核 | 隔离家庭 + 管理员 | **PASS** | admin 登录 + cleanup-staging；武汉访客留言 reject（messageId=6）；family 8 下架 |
| 11 | 转让 / 解散审核 | 隔离家庭 + 管理员 | **PASS** | admin-web 后台审核页真实接口冒烟；操作日志页 PASS |
| 12 | 公开搜索 / 留言 | 游客 | _Codex 待填_ | family 12 公开 |

### 6 个主要页面健康检查（复测 PASS）

覆盖首页、家庭详情、成员列表、家谱、加入申请审批、账号设置等主路径：

- 无控制台 error
- 无请求失败（4xx/5xx）
- 无横向溢出（mobile viewport）

截图目录：`/tmp/tree-h5-e2e-screenshots`

## 8. 数据清理清单

### 已执行

| ID | 操作 | 原因 |
|---|---|---|
| family 20 rel 107/108 | 删除 PARENT_CHILD | 张启国不应为张启明之子 |
| family 20 rel 111/112 | 删除 PARENT_CHILD | 张待安应保持未定位 |
| family 20 | place-existing 张启国 ↔ 张启明 兄弟 | 修正代际（共享张礼安/周淑贞） |
| family 12 join request **5** | reject | 张玉磊遗留待审 |
| family 12 member **75** 林思琪 | unbind user **48** | 恢复张承雅绑定测试 |
| 换号临时账号 | 注销 | 换手机号 E2E 后清理 |
| visitor message **6** 武汉访客 | reject | admin cleanup-staging |
| family **8** Staging公开验收189912 | take-down-public | admin cleanup-staging |

### 保留

| ID | 名称 | 原因 |
|---|---|---|
| 12 | 张氏家族测试数据-M22 | 长期演示 |
| 20 | 张氏家族·武汉江夏支系 | H5 E2E 主家庭 |
| 21 | 江夏张氏·流程验证专户 | 高风险流程隔离 |

### 待管理员凭据后处理

| ID | 名称 | 建议操作 |
|---|---|---|
| 3–7 等 | 历史回归家庭 | 审计后 admin 解散（已标记，需人工确认） |

## 9. Playwright / 复测结果摘要

| 项 | 结果 |
|---|---|
| 执行时间 | 2026-07-01（Codex 复测） |
| 换手机号 | **PASS** |
| picker 基准成员 | **PASS** |
| family 20 家谱 | **PASS** |
| 6 主页面健康 | **PASS** |
| 截图目录 | `/tmp/tree-h5-e2e-screenshots` |
| 管理员审核流 | **PASS**（凭据仅写 `/tmp/tree-h5-e2e-admin.env`，未提交仓库） |

## 9.1 管理员 E2E 最终结果（2026-07-01）

| 项 | 结果 | 备注 |
|---|---|---|
| `setup-admin` 登录 | **PASS** | ROOT_ADMIN；凭据文件 `/tmp/tree-h5-real-e2e-admin-credentials.json` |
| `cleanup-staging` | **PASS** | 见 `/tmp/tree-h5-real-e2e-cleanup-staging.json` |
| 武汉访客留言审核 | **PASS** | reject messageId=**6** |
| family 8 公开下架 | **PASS** | `take-down-public` |
| admin-web 操作日志页 | **PASS** | 截图 `admin-operation-logs.png` |
| 公开申请 / 转让 / 解散审核页 | **PASS** | 后台真实接口冒烟（Codex Playwright） |
| 历史回归家庭解散 | **跳过** | 仅审计标记，禁止模糊名称批量删除 |

**未伪造**：以上结果来自 staging 真实 API 与 cleanup 报告；密码不入库、不入聊天。

## 10. 本地构建与契约测试

| 命令 | 结果 |
|---|---|
| `git diff --check` | PASS |
| `npm run test:contract`（miniapp） | PASS（1/1） |
| `npm run build:h5:real` | PASS |
| `npm run build:mp-weixin:staging` | PASS |

## 11. 缺陷跟踪

| ID | 严重度 | 描述 | 状态 |
|---|---|---|---|
| H5-001 | P0 | change-phone 误用 LoginResult | **已修复，复测 PASS** |
| H5-002 | P1 | family 20 张启国代际错误 | **已修复，复测 PASS** |
| H5-003 | P2 | H5 picker 选中与展示不一致 | **复测 PASS**（排序后一致） |
| H5-004 | P1 | 缺少 staging admin 凭据 | **已解决**（本地 env + cleanup PASS） |
| H5-005 | P2 | repair-tree 重复落位 / graphVersion 漂移 | **已修复**（幂等：actions=[] 且 graphVersion 不变） |
| H5-006 | P2 | cleanup-staging family 8 条件写反 | **已修复** |
| H5-007 | P2 | settings 普通成员显示为管理员 | **已修复** |
| H5-008 | P2 | settings 已解散家庭空白 Tab | **已修复** |
| H5-009 | P3 | admin 操作日志模块/动作列过窄 | **已修复** |
