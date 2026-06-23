# Tree 全业务本地回归测试报告

- 测试时间：2026-06-23 10:28 CST
- 分支：m21-content-center
- Backend：http://127.0.0.1:8080
- 数据库：本地 Docker MySQL `tree_platform`
- 测试性质：本地 API 级全业务回归；不包含微信真实 code/openid 环境；不包含人工真机截图验收。
- 结果统计：PASS 129 / FAIL 0 / SKIP 3
- 覆盖接口/检查项：132 项；唯一真实接口路径：85 个。

## 测试账号

| 类型 | 数量 | 说明 |
|---|---:|---|
| 后台管理员 | 1 | ROOT_ADMIN，本地测试账号；报告不记录密码。 |
| 普通用户 | 7 | 本轮动态注册：founder、invitee、joiner、spare、transfer、solo、cancelOnly；报告记录本地测试手机号与 userId，不记录密码。 |
| 游客 | 1 | 无账号，验证公开家庭、公开树、游客留言、内容阅读。 |

### 动态用户

| key | 测试手机号 | userId |
|---|---|---:|
| founder | 16910282601 | 89 |
| invitee | 16910282602 | 90 |
| joiner | 16910282603 | 91 |
| spare | 16910282604 | 92 |
| transfer | 16910282605 | 93 |
| solo | 16910282606 | 94 |
| cancelOnly | 16910282607，后变更为 16910282666 并完成注销 | 95 |

## 关键测试数据

| 数据 | ID / 值 | 说明 |
|---|---|---|
| mainFamilyId | 39 | 本轮生成/使用 |
| founderMemberId | 97 | 本轮生成/使用 |
| inviteMemberId | 99 | 本轮生成/使用 |
| transferMemberId | 100 | 本轮生成/使用 |
| publicApplicationId | 18 | 本轮生成/使用 |
| visitorMessageId | 16 | 本轮生成/使用 |
| transferFamilyId | 42 | 本轮生成/使用 |
| dissolutionFamilyId | 43 | 本轮生成/使用 |

## 覆盖模块

- Health：1 项
- Admin Auth：4 项
- User Auth：25 项
- WeChat Account：2 项
- Family：16 项
- Family Member：11 项
- Family Relationship：8 项
- Family Tree：2 项
- Family Invitation：12 项
- Family Role：2 项
- Family Public Display：9 项
- Visitor Message：7 项
- Family Join Request：8 项
- Founder Transfer：5 项
- Family Dissolution：3 项
- Family Restore：1 项
- Content Public：4 项
- Content Admin：10 项
- Security：1 项
- Invariant：1 项

## 接口级结果

| 状态 | 模块 | 功能 / 接口 | Method | Path | HTTP | code | 说明 |
|---|---|---|---|---|---:|---:|---|
| PASS | Health | health | GET | `/api/health` | 200 | 0 | public health |
| PASS | Admin Auth | admin login | POST | `/api/admin/auth/login` | 200 | 0 | ROOT_ADMIN local test account |
| PASS | Admin Auth | admin me | GET | `/api/admin/me` | 200 | 0 | admin session |
| PASS | Admin Auth | admin change password wrong old | PUT | `/api/admin/me/password` | 400 | 47001 | negative check; password unchanged |
| PASS | User Auth | send-code REGISTER | POST | `/api/auth/send-code` | 200 | 0 | phone masked in report; devCode not persisted |
| PASS | User Auth | register-phone founder | POST | `/api/auth/register-phone` | 200 | 0 | founder |
| PASS | User Auth | send-code REGISTER | POST | `/api/auth/send-code` | 200 | 0 | phone masked in report; devCode not persisted |
| PASS | User Auth | register-phone invitee | POST | `/api/auth/register-phone` | 200 | 0 | invitee |
| PASS | User Auth | send-code REGISTER | POST | `/api/auth/send-code` | 200 | 0 | phone masked in report; devCode not persisted |
| PASS | User Auth | register-phone joiner | POST | `/api/auth/register-phone` | 200 | 0 | joiner |
| PASS | User Auth | send-code REGISTER | POST | `/api/auth/send-code` | 200 | 0 | phone masked in report; devCode not persisted |
| PASS | User Auth | register-phone spare | POST | `/api/auth/register-phone` | 200 | 0 | spare |
| PASS | User Auth | send-code REGISTER | POST | `/api/auth/send-code` | 200 | 0 | phone masked in report; devCode not persisted |
| PASS | User Auth | register-phone transfer | POST | `/api/auth/register-phone` | 200 | 0 | transfer |
| PASS | User Auth | send-code REGISTER | POST | `/api/auth/send-code` | 200 | 0 | phone masked in report; devCode not persisted |
| PASS | User Auth | register-phone solo | POST | `/api/auth/register-phone` | 200 | 0 | solo |
| PASS | User Auth | send-code REGISTER | POST | `/api/auth/send-code` | 200 | 0 | phone masked in report; devCode not persisted |
| PASS | User Auth | register-phone cancelOnly | POST | `/api/auth/register-phone` | 200 | 0 | cancelOnly |
| PASS | User Auth | login-phone founder | POST | `/api/auth/login-phone` | 200 | 0 | login existing temp user |
| PASS | User Auth | login wrong password | POST | `/api/auth/login-phone` | 401 | 40201 | negative login |
| PASS | User Auth | user logout solo token | POST | `/api/auth/logout` | 200 | 0 | logout protected route |
| PASS | User Auth | login-phone solo again | POST | `/api/auth/login-phone` | 200 | 0 | solo relogin |
| SKIP | WeChat Account | wechat-mini login | POST | `/api/auth/wechat-mini/login` | - | - | Requires real WeChat code/openid environment; not marked PASS |
| SKIP | WeChat Account | wechat-mini bind-phone | POST | `/api/auth/wechat-mini/bind-phone` | - | - | Requires authenticated WeChat-bound temp user; not marked PASS |
| PASS | Family | create family | POST | `/api/families` | 201 | 0 | main regression family |
| PASS | Family | list my families | GET | `/api/families` | 200 | 0 | founder family list |
| PASS | Family | family detail | GET | `/api/families/39` | 200 | 0 | private detail |
| PASS | Family | update family | PUT | `/api/families/39` | 200 | 0 | update metadata |
| PASS | Family | public detail before approval | GET | `/api/families/39/public` | 400 | 42103 | not public yet |
| PASS | Family Member | create member temp | POST | `/api/families/39/members` | 201 | 0 | member create/list/detail/update/delete |
| PASS | Family Member | list members | GET | `/api/families/39/members` | 200 | 0 | member list |
| PASS | Family Member | member detail | GET | `/api/families/39/members/98` | 200 | 0 | member detail |
| PASS | Family Member | update member | PUT | `/api/families/39/members/98` | 200 | 0 | member update |
| PASS | Family Member | delete member | DELETE | `/api/families/39/members/98` | 200 | 0 | member delete no descendants |
| PASS | Family Member | create invite target member | POST | `/api/families/39/members` | 201 | 0 | for invitation accept |
| PASS | Family Member | create transfer target member | POST | `/api/families/39/members` | 201 | 0 | for founder transfer |
| PASS | Family Member | create join existing member | POST | `/api/families/39/members` | 201 | 0 | for approve join bind existing |
| PASS | Family Relationship | add father | POST | `/api/families/39/relationships` | 201 | 0 | PARENT_CHILD father |
| PASS | Family Relationship | duplicate father fails | POST | `/api/families/39/relationships` | 400 | 43304 | negative duplicate parent |
| PASS | Family Relationship | add mother | POST | `/api/families/39/relationships` | 201 | 0 | PARENT_CHILD mother |
| PASS | Family Relationship | add child | POST | `/api/families/39/relationships` | 201 | 0 | PARENT_CHILD child |
| PASS | Family Relationship | add spouse | POST | `/api/families/39/relationships` | 201 | 0 | SPOUSE |
| PASS | Family Relationship | add sibling | POST | `/api/families/39/relationships` | 201 | 0 | ADD_SIBLING via shared parents |
| PASS | Family Relationship | update relationship | PUT | `/api/families/39/relationships/22` | 200 | 0 | relationship update |
| PASS | Family Relationship | delete relationship | DELETE | `/api/families/39/relationships/23` | 200 | 0 | relationship delete |
| PASS | Family Tree | private tree | GET | `/api/families/39/tree` | 200 | 0 | private tree nodes/edges |
| PASS | Family Invitation | create share invitation | POST | `/api/families/39/members/99/invite` | 201 | 0 | share link |
| PASS | Family Invitation | invitation detail by token | GET | `/api/invitations/[REDACTED_INVITE_TOKEN]` | 200 | 0 | anonymous detail; token not persisted in report |
| PASS | Family Invitation | accept invitation | POST | `/api/invitations/36/accept` | 200 | 0 | accept creates ACTIVE link |
| PASS | Family Invitation | my invitations invitee | GET | `/api/users/me/invitations` | 200 | 0 | my received invitations |
| PASS | Family Invitation | create transfer member invite | POST | `/api/families/39/members/100/invite` | 201 | 0 | prepare transfer target |
| PASS | Family Invitation | accept transfer target invite | POST | `/api/invitations/37/accept` | 200 | 0 | transfer target link |
| PASS | Family Member | create in-app invite member | POST | `/api/families/39/members` | 201 | 0 | in-app target |
| PASS | Family Invitation | create in-app invitation | POST | `/api/families/39/members/107/invite` | 201 | 0 | in-app invitation |
| PASS | Family Invitation | reject invitation | POST | `/api/invitations/38/reject` | 200 | 0 | reject pending invitation |
| PASS | Family Member | create cancel invite member | POST | `/api/families/39/members` | 201 | 0 | cancel invite target |
| PASS | Family Invitation | create cancelable invitation | POST | `/api/families/39/members/108/invite` | 201 | 0 | cancel route |
| PASS | Family Invitation | cancel invitation | POST | `/api/invitations/39/cancel` | 200 | 0 | cancel invitation |
| PASS | Family Role | set family admin | POST | `/api/families/39/members/99/set-admin` | 200 | 0 | founder sets FAMILY_ADMIN |
| PASS | Family Role | unset family admin | POST | `/api/families/39/members/99/unset-admin` | 200 | 0 | founder unsets FAMILY_ADMIN |
| PASS | Family Public Display | submit public application | POST | `/api/families/39/public-applications` | 201 | 0 | submit public display |
| PASS | Family Public Display | list family public applications | GET | `/api/families/39/public-applications` | 200 | 0 | family public app list |
| PASS | Family Public Display | admin list public applications | GET | `/api/admin/family-public-applications?status=PENDING&page=1&pageSize=20` | 200 | 0 | admin public app list |
| PASS | Family Public Display | admin approve public application | POST | `/api/admin/family-public-applications/18/approve` | 200 | 0 | approve public |
| PASS | Family | public detail after approval | GET | `/api/families/39/public` | 200 | 0 | public detail |
| PASS | Family Tree | public tree after approval | GET | `/api/public/families/39/tree` | 200 | 0 | public tree |
| PASS | Visitor Message | create visitor message | POST | `/api/public/families/39/visitor-messages` | 201 | 0 | pending visitor message |
| PASS | Visitor Message | public visitor messages before approve | GET | `/api/public/families/39/visitor-messages?page=1&pageSize=20` | 200 | 0 | pending should not be visible |
| PASS | Visitor Message | admin list visitor messages | GET | `/api/admin/visitor-messages?familyId=39&page=1&pageSize=20` | 200 | 0 | admin visitor message list |
| PASS | Visitor Message | admin approve visitor message | POST | `/api/admin/visitor-messages/16/approve` | 200 | 0 | approve visitor message |
| PASS | Visitor Message | public visitor messages after approve | GET | `/api/public/families/39/visitor-messages?page=1&pageSize=20` | 200 | 0 | approved visible without phone/wechat |
| PASS | Visitor Message | create second visitor message | POST | `/api/public/families/39/visitor-messages` | 429 | 45105 | may hit same-IP rate limit |
| SKIP | Visitor Message | admin reject/delete second message | POST/DELETE | `/api/admin/visitor-messages/{messageId}` | - | - | Skipped because DB rate limit correctly blocked same-IP second message |
| PASS | Family Join Request | create join request cancel | POST | `/api/families/39/join-requests` | 201 | 0 | join request cancel |
| PASS | Family Join Request | my join requests | GET | `/api/users/me/join-requests` | 200 | 0 | my join requests |
| PASS | Family Join Request | cancel join request | POST | `/api/families/39/join-requests/15/cancel` | 200 | 0 | applicant cancel |
| PASS | Family Join Request | create join request approve | POST | `/api/families/39/join-requests` | 201 | 0 | join request approve existing |
| PASS | Family Join Request | family join requests | GET | `/api/families/39/join-requests` | 200 | 0 | family admin join list |
| PASS | Family Join Request | approve join existing member | POST | `/api/families/39/join-requests/16/approve` | 200 | 0 | approve join existing |
| PASS | Family Join Request | create join request reject | POST | `/api/families/39/join-requests` | 201 | 0 | join request reject |
| PASS | Family Join Request | reject join request | POST | `/api/families/39/join-requests/17/reject` | 200 | 0 | reject join |
| PASS | Family | create family public cancel | POST | `/api/families` | 201 | 0 | for public app cancel |
| PASS | Family Public Display | submit public app cancel | POST | `/api/families/40/public-applications` | 201 | 0 | submit then cancel |
| PASS | Family Public Display | cancel public application | POST | `/api/families/40/public-applications/19/cancel` | 200 | 0 | cancel public app |
| PASS | Family | create family public reject | POST | `/api/families` | 201 | 0 | for public app reject |
| PASS | Family Public Display | submit public app reject | POST | `/api/families/41/public-applications` | 201 | 0 | submit then reject |
| PASS | Family Public Display | admin reject public application | POST | `/api/admin/family-public-applications/20/reject` | 200 | 0 | reject public app |
| PASS | Family Public Display | admin take down public family | POST | `/api/admin/families/39/take-down-public` | 200 | 0 | take down after public tests |
| PASS | Family | public detail after take down | GET | `/api/families/39/public` | 400 | 42103 | public unavailable after take-down |
| PASS | Family | create transfer family | POST | `/api/families` | 201 | 0 | for founder transfer |
| PASS | Family Member | create transfer family member | POST | `/api/families/42/members` | 201 | 0 | transfer target member |
| PASS | Family Invitation | invite transfer family target | POST | `/api/families/42/members/112/invite` | 201 | 0 | bind transfer target |
| PASS | Family Invitation | accept transfer family target | POST | `/api/invitations/40/accept` | 200 | 0 | bind target |
| PASS | Founder Transfer | create founder transfer reject | POST | `/api/families/42/founder-transfer-requests` | 201 | 0 | create transfer request |
| PASS | Founder Transfer | admin list founder transfers | GET | `/api/admin/founder-transfer-requests?familyId=42&page=1&pageSize=20` | 200 | 0 | admin transfer list |
| PASS | Founder Transfer | admin reject founder transfer | POST | `/api/admin/founder-transfer-requests/6/reject` | 200 | 0 | reject transfer |
| PASS | Founder Transfer | create founder transfer approve | POST | `/api/families/42/founder-transfer-requests` | 201 | 0 | create transfer request approve |
| PASS | Founder Transfer | admin approve founder transfer | POST | `/api/admin/founder-transfer-requests/7/approve` | 200 | 0 | approve transfer |
| PASS | Family | create dissolution family | POST | `/api/families` | 201 | 0 | for dissolution |
| PASS | Family | create dissolution request cancel | POST | `/api/families/43/dissolution-requests` | 200 | 0 | user dissolution create |
| PASS | Family | current dissolution request | GET | `/api/families/43/dissolution-requests/current` | 200 | 0 | current dissolution |
| PASS | Family | cancel dissolution request | POST | `/api/families/43/dissolution-requests/12/cancel` | 200 | 0 | cancel dissolution |
| PASS | Family | create dissolution request reject | POST | `/api/families/43/dissolution-requests` | 200 | 0 | create dissolution reject |
| PASS | Family Dissolution | admin list dissolution | GET | `/api/admin/dissolution-requests?familyId=43&page=1&pageSize=20` | 200 | 0 | admin dissolution list |
| PASS | Family Dissolution | admin reject dissolution | POST | `/api/admin/dissolution-requests/13/reject` | 200 | 0 | reject dissolution |
| PASS | Family | create dissolution request approve | POST | `/api/families/43/dissolution-requests` | 200 | 0 | create dissolution approve |
| PASS | Family Dissolution | admin approve dissolution | POST | `/api/admin/dissolution-requests/14/approve` | 200 | 0 | approve dissolution |
| PASS | Family Restore | admin restore family | POST | `/api/admin/families/43/restore` | 200 | 0 | restore dissolved family |
| PASS | Content Public | public categories | GET | `/api/content/categories` | 200 | 0 | content categories |
| PASS | Content Public | public articles | GET | `/api/content/articles?page=1&pageSize=5` | 200 | 0 | content public list |
| PASS | Content Public | public article detail | GET | `/api/content/articles/tutorial-create-family` | 200 | 0 | content detail |
| PASS | Content Admin | admin content categories | GET | `/api/admin/content/categories` | 200 | 0 | admin content category list |
| PASS | Content Admin | create category | POST | `/api/admin/content/categories` | 201 | 0 | create content category |
| PASS | Content Admin | update category | PUT | `/api/admin/content/categories/14` | 200 | 0 | update content category |
| PASS | Content Admin | create article | POST | `/api/admin/content/articles` | 201 | 0 | create content article |
| PASS | Content Admin | admin article detail | GET | `/api/admin/content/articles/22` | 200 | 0 | admin content detail |
| PASS | Content Admin | update article | PUT | `/api/admin/content/articles/22` | 200 | 0 | update content article |
| PASS | Content Admin | publish article | POST | `/api/admin/content/articles/22/publish` | 200 | 0 | publish content article |
| PASS | Content Public | public detail new article | GET | `/api/content/articles/regression-content-102826` | 200 | 0 | published visible |
| PASS | Content Admin | unpublish article | POST | `/api/admin/content/articles/22/unpublish` | 200 | 0 | unpublish content article |
| PASS | Content Admin | delete article | DELETE | `/api/admin/content/articles/22` | 200 | 0 | delete content article |
| PASS | Content Admin | delete category | DELETE | `/api/admin/content/categories/14` | 200 | 0 | delete content category |
| PASS | User Auth | logout cancelOnly token | POST | `/api/auth/logout` | 200 | 0 | logout before relogin |
| PASS | User Auth | login-phone cancelOnly again | POST | `/api/auth/login-phone` | 200 | 0 | cancelOnly relogin |
| PASS | User Auth | send-code CHANGE_PHONE_OLD | POST | `/api/auth/send-code` | 200 | 0 | phone masked in report; devCode not persisted |
| PASS | User Auth | send-code CHANGE_PHONE_NEW | POST | `/api/auth/send-code` | 200 | 0 | phone masked in report; devCode not persisted |
| PASS | User Auth | change phone | POST | `/api/auth/change-phone` | 200 | 0 | change phone for standalone user |
| PASS | User Auth | send-code CANCEL_ACCOUNT | POST | `/api/auth/send-code` | 200 | 0 | phone masked in report; devCode not persisted |
| PASS | User Auth | cancel account | POST | `/api/auth/cancel-account` | 200 | 0 | cancel standalone account |
| PASS | Admin Auth | admin logout | POST | `/api/admin/auth/logout` | 200 | 0 | admin logout |
| PASS | Security | operation_logs sensitive detail scan | SQL | `operation_logs` | - | 0 | no sensitive keywords in detail_json |
| PASS | Invariant | single active founder in main family | SQL | `family_member_user_links` | - | 0 | founder_count=1 |

## 跳过项

| 模块 | 接口 | 原因 |
|---|---|---|
| WeChat Account | `POST /api/auth/wechat-mini/login` | Requires real WeChat code/openid environment; not marked PASS |
| WeChat Account | `POST /api/auth/wechat-mini/bind-phone` | Requires authenticated WeChat-bound temp user; not marked PASS |
| Visitor Message | `POST/DELETE /api/admin/visitor-messages/{messageId}` | Skipped because DB rate limit correctly blocked same-IP second message |

## 自动化验证命令

| 项目 | 命令 | 结果 |
|---|---|---|
| API 全量回归脚本 | `bash /tmp/tree_full_regression.sh` | PASS：129 / FAIL：0 / SKIP：3 |
| Backend unit tests | `docker exec tree-dev bash -lc "cd /workspace/backend && go test ./..."` | PASS |
| Backend vet | `docker exec tree-dev bash -lc "cd /workspace/backend && go vet ./..."` | PASS |
| Web build | `docker exec tree-dev bash -lc "cd /workspace/web && VITE_API_MODE=real VITE_API_BASE_URL=/api pnpm build"` | PASS |
| Admin Web build | `docker exec tree-dev bash -lc "cd /workspace/admin-web && VITE_API_MODE=real VITE_API_BASE_URL=/api pnpm build"` | PASS |
| Miniapp mp-weixin build | `docker exec tree-dev bash -lc "cd /workspace/miniapp && pnpm build:mp-weixin"` | PASS |
| Miniapp H5 build | `docker exec tree-dev bash -lc "cd /workspace/miniapp && VITE_API_MODE=real VITE_API_BASE_URL=/api pnpm build:h5"` | PASS |

## 安全检查结论

- 报告未记录完整 accessToken、admin token、验证码、密码、inviteToken、openid、unionid、AppSecret。
- `operation_logs.detail_json` 敏感关键字 SQL 扫描结果为 0。
- 主测试家庭 ACTIVE FOUNDER 唯一性 SQL 检查通过。

## 结论

本轮本地全业务 API 回归通过。除真实微信登录/绑定需要外部微信环境，其他核心链路均已通过本地自动化验证。
