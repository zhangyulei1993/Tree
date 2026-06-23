# Tree Staging 公网全业务回归测试报告

- 测试时间：2026-06-23 12:30 CST
- 分支：m21-content-center
- Backend：http://114.55.128.149:18083
- Admin：http://114.55.128.149:18084
- 数据库：远程 staging Docker MySQL `tree_platform_staging`
- 测试性质：公网 staging API 级全业务回归；不包含微信真实 code/openid 环境；不包含人工真机截图验收。
- 结果统计：API PASS 129 / FAIL 0 / SKIP 3；Web/Admin 页面 smoke PASS。
- 覆盖接口/检查项：132 项。

## 测试账号

| 类型 | 数量 | 说明 |
|---|---:|---|
| 后台管理员 | 1 | ROOT_ADMIN，本地测试账号；报告不记录密码。 |
| 普通用户 | 7 | 本轮动态注册：founder、invitee、joiner、spare、transfer、solo、cancelOnly；报告记录本地 staging 测试手机号与 userId，不记录密码。 |
| 游客 | 1 | 无账号，验证公开家庭、公开树、游客留言、内容阅读。 |

### 动态用户

| key | 测试手机号 | userId |
|---|---|---:|
| founder | 16912294201 | 3 |
| invitee | 16912294202 | 4 |
| joiner | 16912294203 | 5 |
| spare | 16912294204 | 6 |
| transfer | 16912294205 | 7 |
| solo | 16912294206 | 8 |
| cancelOnly | 16912294207，后变更为 16912294266 并完成注销 | 9 |

## 关键测试数据

| 数据 | ID / 值 | 说明 |
|---|---|---|
| mainFamilyId | 3 | 本轮生成/使用 |
| founderMemberId | 5 | 本轮生成/使用 |
| inviteMemberId | 7 | 本轮生成/使用 |
| transferMemberId | 8 | 本轮生成/使用 |
| publicApplicationId | 1 | 本轮生成/使用 |
| visitorMessageId | 1 | 本轮生成/使用 |
| transferFamilyId | 6 | 本轮生成/使用 |
| dissolutionFamilyId | 7 | 本轮生成/使用 |
| uiSmokePublicFamilyId | 8 | 页面 smoke 追加生成的隔离公开家庭 |
| uiSmokePublicApplicationId | 4 | 页面 smoke 追加生成并审核通过 |

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
| PASS | Family | family detail | GET | `/api/families/3` | 200 | 0 | private detail |
| PASS | Family | update family | PUT | `/api/families/3` | 200 | 0 | update metadata |
| PASS | Family | public detail before approval | GET | `/api/families/3/public` | 400 | 42103 | not public yet |
| PASS | Family Member | create member temp | POST | `/api/families/3/members` | 201 | 0 | member create/list/detail/update/delete |
| PASS | Family Member | list members | GET | `/api/families/3/members` | 200 | 0 | member list |
| PASS | Family Member | member detail | GET | `/api/families/3/members/6` | 200 | 0 | member detail |
| PASS | Family Member | update member | PUT | `/api/families/3/members/6` | 200 | 0 | member update |
| PASS | Family Member | delete member | DELETE | `/api/families/3/members/6` | 200 | 0 | member delete no descendants |
| PASS | Family Member | create invite target member | POST | `/api/families/3/members` | 201 | 0 | for invitation accept |
| PASS | Family Member | create transfer target member | POST | `/api/families/3/members` | 201 | 0 | for founder transfer |
| PASS | Family Member | create join existing member | POST | `/api/families/3/members` | 201 | 0 | for approve join bind existing |
| PASS | Family Relationship | add father | POST | `/api/families/3/relationships` | 201 | 0 | PARENT_CHILD father |
| PASS | Family Relationship | duplicate father fails | POST | `/api/families/3/relationships` | 400 | 43304 | negative duplicate parent |
| PASS | Family Relationship | add mother | POST | `/api/families/3/relationships` | 201 | 0 | PARENT_CHILD mother |
| PASS | Family Relationship | add child | POST | `/api/families/3/relationships` | 201 | 0 | PARENT_CHILD child |
| PASS | Family Relationship | add spouse | POST | `/api/families/3/relationships` | 201 | 0 | SPOUSE |
| PASS | Family Relationship | add sibling | POST | `/api/families/3/relationships` | 201 | 0 | ADD_SIBLING via shared parents |
| PASS | Family Relationship | update relationship | PUT | `/api/families/3/relationships/3` | 200 | 0 | relationship update |
| PASS | Family Relationship | delete relationship | DELETE | `/api/families/3/relationships/4` | 200 | 0 | relationship delete |
| PASS | Family Tree | private tree | GET | `/api/families/3/tree` | 200 | 0 | private tree nodes/edges |
| PASS | Family Invitation | create share invitation | POST | `/api/families/3/members/7/invite` | 201 | 0 | share link |
| PASS | Family Invitation | invitation detail by token | GET | `/api/invitations/[REDACTED_INVITE_TOKEN]` | 200 | 0 | anonymous detail; token not persisted in report |
| PASS | Family Invitation | accept invitation | POST | `/api/invitations/1/accept` | 200 | 0 | accept creates ACTIVE link |
| PASS | Family Invitation | my invitations invitee | GET | `/api/users/me/invitations` | 200 | 0 | my received invitations |
| PASS | Family Invitation | create transfer member invite | POST | `/api/families/3/members/8/invite` | 201 | 0 | prepare transfer target |
| PASS | Family Invitation | accept transfer target invite | POST | `/api/invitations/2/accept` | 200 | 0 | transfer target link |
| PASS | Family Member | create in-app invite member | POST | `/api/families/3/members` | 201 | 0 | in-app target |
| PASS | Family Invitation | create in-app invitation | POST | `/api/families/3/members/15/invite` | 201 | 0 | in-app invitation |
| PASS | Family Invitation | reject invitation | POST | `/api/invitations/3/reject` | 200 | 0 | reject pending invitation |
| PASS | Family Member | create cancel invite member | POST | `/api/families/3/members` | 201 | 0 | cancel invite target |
| PASS | Family Invitation | create cancelable invitation | POST | `/api/families/3/members/16/invite` | 201 | 0 | cancel route |
| PASS | Family Invitation | cancel invitation | POST | `/api/invitations/4/cancel` | 200 | 0 | cancel invitation |
| PASS | Family Role | set family admin | POST | `/api/families/3/members/7/set-admin` | 200 | 0 | founder sets FAMILY_ADMIN |
| PASS | Family Role | unset family admin | POST | `/api/families/3/members/7/unset-admin` | 200 | 0 | founder unsets FAMILY_ADMIN |
| PASS | Family Public Display | submit public application | POST | `/api/families/3/public-applications` | 201 | 0 | submit public display |
| PASS | Family Public Display | list family public applications | GET | `/api/families/3/public-applications` | 200 | 0 | family public app list |
| PASS | Family Public Display | admin list public applications | GET | `/api/admin/family-public-applications?status=PENDING&page=1&pageSize=20` | 200 | 0 | admin public app list |
| PASS | Family Public Display | admin approve public application | POST | `/api/admin/family-public-applications/1/approve` | 200 | 0 | approve public |
| PASS | Family | public detail after approval | GET | `/api/families/3/public` | 200 | 0 | public detail |
| PASS | Family Tree | public tree after approval | GET | `/api/public/families/3/tree` | 200 | 0 | public tree |
| PASS | Visitor Message | create visitor message | POST | `/api/public/families/3/visitor-messages` | 201 | 0 | pending visitor message |
| PASS | Visitor Message | public visitor messages before approve | GET | `/api/public/families/3/visitor-messages?page=1&pageSize=20` | 200 | 0 | pending should not be visible |
| PASS | Visitor Message | admin list visitor messages | GET | `/api/admin/visitor-messages?familyId=3&page=1&pageSize=20` | 200 | 0 | admin visitor message list |
| PASS | Visitor Message | admin approve visitor message | POST | `/api/admin/visitor-messages/1/approve` | 200 | 0 | approve visitor message |
| PASS | Visitor Message | public visitor messages after approve | GET | `/api/public/families/3/visitor-messages?page=1&pageSize=20` | 200 | 0 | approved visible without phone/wechat |
| PASS | Visitor Message | create second visitor message | POST | `/api/public/families/3/visitor-messages` | 429 | 45105 | may hit same-IP rate limit |
| SKIP | Visitor Message | admin reject/delete second message | POST/DELETE | `/api/admin/visitor-messages/{messageId}` | - | - | Skipped because DB rate limit correctly blocked same-IP second message |
| PASS | Family Join Request | create join request cancel | POST | `/api/families/3/join-requests` | 201 | 0 | join request cancel |
| PASS | Family Join Request | my join requests | GET | `/api/users/me/join-requests` | 200 | 0 | my join requests |
| PASS | Family Join Request | cancel join request | POST | `/api/families/3/join-requests/1/cancel` | 200 | 0 | applicant cancel |
| PASS | Family Join Request | create join request approve | POST | `/api/families/3/join-requests` | 201 | 0 | join request approve existing |
| PASS | Family Join Request | family join requests | GET | `/api/families/3/join-requests` | 200 | 0 | family admin join list |
| PASS | Family Join Request | approve join existing member | POST | `/api/families/3/join-requests/2/approve` | 200 | 0 | approve join existing |
| PASS | Family Join Request | create join request reject | POST | `/api/families/3/join-requests` | 201 | 0 | join request reject |
| PASS | Family Join Request | reject join request | POST | `/api/families/3/join-requests/3/reject` | 200 | 0 | reject join |
| PASS | Family | create family public cancel | POST | `/api/families` | 201 | 0 | for public app cancel |
| PASS | Family Public Display | submit public app cancel | POST | `/api/families/4/public-applications` | 201 | 0 | submit then cancel |
| PASS | Family Public Display | cancel public application | POST | `/api/families/4/public-applications/2/cancel` | 200 | 0 | cancel public app |
| PASS | Family | create family public reject | POST | `/api/families` | 201 | 0 | for public app reject |
| PASS | Family Public Display | submit public app reject | POST | `/api/families/5/public-applications` | 201 | 0 | submit then reject |
| PASS | Family Public Display | admin reject public application | POST | `/api/admin/family-public-applications/3/reject` | 200 | 0 | reject public app |
| PASS | Family Public Display | admin take down public family | POST | `/api/admin/families/3/take-down-public` | 200 | 0 | take down after public tests |
| PASS | Family | public detail after take down | GET | `/api/families/3/public` | 400 | 42103 | public unavailable after take-down |
| PASS | Family | create transfer family | POST | `/api/families` | 201 | 0 | for founder transfer |
| PASS | Family Member | create transfer family member | POST | `/api/families/6/members` | 201 | 0 | transfer target member |
| PASS | Family Invitation | invite transfer family target | POST | `/api/families/6/members/20/invite` | 201 | 0 | bind transfer target |
| PASS | Family Invitation | accept transfer family target | POST | `/api/invitations/5/accept` | 200 | 0 | bind target |
| PASS | Founder Transfer | create founder transfer reject | POST | `/api/families/6/founder-transfer-requests` | 201 | 0 | create transfer request |
| PASS | Founder Transfer | admin list founder transfers | GET | `/api/admin/founder-transfer-requests?familyId=6&page=1&pageSize=20` | 200 | 0 | admin transfer list |
| PASS | Founder Transfer | admin reject founder transfer | POST | `/api/admin/founder-transfer-requests/1/reject` | 200 | 0 | reject transfer |
| PASS | Founder Transfer | create founder transfer approve | POST | `/api/families/6/founder-transfer-requests` | 201 | 0 | create transfer request approve |
| PASS | Founder Transfer | admin approve founder transfer | POST | `/api/admin/founder-transfer-requests/2/approve` | 200 | 0 | approve transfer |
| PASS | Family | create dissolution family | POST | `/api/families` | 201 | 0 | for dissolution |
| PASS | Family | create dissolution request cancel | POST | `/api/families/7/dissolution-requests` | 200 | 0 | user dissolution create |
| PASS | Family | current dissolution request | GET | `/api/families/7/dissolution-requests/current` | 200 | 0 | current dissolution |
| PASS | Family | cancel dissolution request | POST | `/api/families/7/dissolution-requests/1/cancel` | 200 | 0 | cancel dissolution |
| PASS | Family | create dissolution request reject | POST | `/api/families/7/dissolution-requests` | 200 | 0 | create dissolution reject |
| PASS | Family Dissolution | admin list dissolution | GET | `/api/admin/dissolution-requests?familyId=7&page=1&pageSize=20` | 200 | 0 | admin dissolution list |
| PASS | Family Dissolution | admin reject dissolution | POST | `/api/admin/dissolution-requests/2/reject` | 200 | 0 | reject dissolution |
| PASS | Family | create dissolution request approve | POST | `/api/families/7/dissolution-requests` | 200 | 0 | create dissolution approve |
| PASS | Family Dissolution | admin approve dissolution | POST | `/api/admin/dissolution-requests/3/approve` | 200 | 0 | approve dissolution |
| PASS | Family Restore | admin restore family | POST | `/api/admin/families/7/restore` | 200 | 0 | restore dissolved family |
| PASS | Content Public | public categories | GET | `/api/content/categories` | 200 | 0 | content categories |
| PASS | Content Public | public articles | GET | `/api/content/articles?page=1&pageSize=5` | 200 | 0 | content public list |
| PASS | Content Public | public article detail | GET | `/api/content/articles/tutorial-create-family` | 200 | 0 | content detail |
| PASS | Content Admin | admin content categories | GET | `/api/admin/content/categories` | 200 | 0 | admin content category list |
| PASS | Content Admin | create category | POST | `/api/admin/content/categories` | 201 | 0 | create content category |
| PASS | Content Admin | update category | PUT | `/api/admin/content/categories/5` | 200 | 0 | update content category |
| PASS | Content Admin | create article | POST | `/api/admin/content/articles` | 201 | 0 | create content article |
| PASS | Content Admin | admin article detail | GET | `/api/admin/content/articles/16` | 200 | 0 | admin content detail |
| PASS | Content Admin | update article | PUT | `/api/admin/content/articles/16` | 200 | 0 | update content article |
| PASS | Content Admin | publish article | POST | `/api/admin/content/articles/16/publish` | 200 | 0 | publish content article |
| PASS | Content Public | public detail new article | GET | `/api/content/articles/regression-content-122942` | 200 | 0 | published visible |
| PASS | Content Admin | unpublish article | POST | `/api/admin/content/articles/16/unpublish` | 200 | 0 | unpublish content article |
| PASS | Content Admin | delete article | DELETE | `/api/admin/content/articles/16` | 200 | 0 | delete content article |
| PASS | Content Admin | delete category | DELETE | `/api/admin/content/categories/5` | 200 | 0 | delete content category |
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

## 安全检查结论

- 报告未记录完整 accessToken、admin token、验证码、密码、inviteToken、openid、unionid、AppSecret。
- `operation_logs.detail_json` 敏感关键字 SQL 扫描结果为 0。
- 主测试家庭 ACTIVE FOUNDER 唯一性 SQL 检查通过。
- 当前公网安全风险：`3306` 仍可从公网连接，建议从阿里云安全组和服务器 firewalld 移除。

## Web / Admin 页面级 Smoke

测试方式：

- 通过 SSH 隧道访问公网 staging Nginx 入口：
  - Web：`http://127.0.0.1:19083` -> `http://114.55.128.149:18083`
  - Admin：`http://127.0.0.1:19084` -> `http://114.55.128.149:18084`
- 浏览器自动化只验证页面加载、真实登录、关键路由、控制台 error。
- 不记录 token、密码、验证码。

| 状态 | 端 | 页面 / 路由 | 说明 |
|---|---|---|---|
| PASS | Web | `/` | 首页加载正常 |
| PASS | Web | `/content` | 内容中心加载正常 |
| PASS | Web | `/login` -> `/me` | 用户真实登录成功 |
| PASS | Web | `/me/families` | 我的家庭加载正常 |
| PASS | Web | `/families/3` | 家庭详情加载正常 |
| PASS | Web | `/families/3/members` | 成员列表加载正常 |
| PASS | Web | `/families/3/tree` | 私有家谱加载正常 |
| PASS | Web | `/families/8/public` | 公开家庭主页加载正常 |
| PASS | Web | `/families/8/tree/public` | 公开家谱加载正常 |
| PASS | Admin | `/admin/login` -> `/admin/dashboard` | ROOT_ADMIN 真实登录成功 |
| PASS | Admin | `/admin/public-applications` | 公开申请审核页加载正常 |
| PASS | Admin | `/admin/visitor-messages` | 游客留言审核页加载正常 |
| PASS | Admin | `/admin/founder-transfer-requests` | 创始人转让审核页加载正常 |
| PASS | Admin | `/admin/dissolution-requests` | 家庭解散审核页加载正常 |
| PASS | Admin | `/admin/content` | 内容管理页加载正常 |

页面 smoke 备注：

- `familyId=3` 已在 API 回归中执行公开下架，公开主页按设计不可访问。
- 为公开页面正向 smoke 追加创建隔离家庭 `familyId=8`，并将公开申请 `applicationId=4` 审核通过。

## 公网入口与域名状态

| 项 | 状态 | 说明 |
|---|---|---|
| `http://114.55.128.149:18083/api/health` | PASS | Web 入口反代 API 正常 |
| `http://114.55.128.149:18084/api/health` | PASS | Admin 入口反代 API 正常 |
| `http://114.55.128.149:18081/api/health` | PASS（不可公网访问） | 后端直连端口未暴露公网，API 通过 Nginx 反代访问 |
| `tree.bigbigboy.cn` / `tapi.bigbigboy.cn` / `tadmin.bigbigboy.cn` | SKIP | 当前未解析到 ECS，无法配置域名 HTTPS 验收 |
| 小程序合法 request 域名 | SKIP | 需要 HTTPS API 域名后才能配置和验收 |

## 结论

公网 staging API 级全业务回归通过，Web/Admin 页面级 smoke 通过。SKIP 项均为真实微信环境、同 IP 留言限流、域名/HTTPS 未配置相关，未标记为 PASS。

下一步应配置 staging 域名 DNS、HTTPS 证书、小程序合法 request 域名，然后做小程序体验版人工验收。
